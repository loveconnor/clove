package gitservice

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"clove/apps/git-http/internal/auth"
	"clove/apps/git-http/internal/config"
)

const (
	serviceUploadPack  = "git-upload-pack"
	serviceReceivePack = "git-receive-pack"
)

var (
	errNotFound        = errors.New("repository not found")
	errAuthRequired    = errors.New("authentication required")
	errAuthUnavailable = errors.New("authentication is unavailable")
	errForbidden       = errors.New("forbidden")
)

type Authenticator interface {
	Authenticate(ctx context.Context, token string) (auth.Principal, error)
	Configured() bool
}

type RepositoryStore interface {
	FindRepository(ctx context.Context, owner, name, userID string) (Repository, error)
}

type Server struct {
	cfg    config.Config
	store  RepositoryStore
	auth   Authenticator
	logger *slog.Logger
	gitBin string
	mux    *http.ServeMux
}

type Dependencies struct {
	Config config.Config
	Store  RepositoryStore
	Auth   Authenticator
	Logger *slog.Logger
}

type Repository struct {
	ID         string
	OwnerType  string
	OwnerID    string
	Owner      string
	Name       string
	Visibility string
	GitPath    string
	UserRole   string
}

type route struct {
	Owner   string
	Repo    string
	Service string
	RPC     bool
}

func New(deps Dependencies) http.Handler {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}
	gitBin := strings.TrimSpace(deps.Config.GitBin)
	if gitBin == "" {
		gitBin = "git"
	}

	s := &Server{
		cfg:    deps.Config,
		store:  deps.Store,
		auth:   deps.Auth,
		logger: logger,
		gitBin: gitBin,
		mux:    http.NewServeMux(),
	}
	s.routes()
	return s.requestLogger(s.mux)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("/", s.handleGit)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, fmt.Sprintf(`{"status":"ok","service":%q}`, s.cfg.AppName))
}

func (s *Server) handleGit(w http.ResponseWriter, r *http.Request) {
	rt, err := parseRoute(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	principal, authenticated, authErr := s.authenticateRequest(r)
	if authErr != nil {
		if errors.Is(authErr, errAuthUnavailable) {
			http.Error(w, "authentication is not configured", http.StatusServiceUnavailable)
			return
		}
		s.challenge(w)
		return
	}

	if s.store == nil {
		http.Error(w, "repository lookup is not configured", http.StatusServiceUnavailable)
		return
	}
	repository, err := s.store.FindRepository(r.Context(), rt.Owner, rt.Repo, principal.UserID)
	if err != nil {
		if errors.Is(err, errNotFound) {
			http.NotFound(w, r)
			return
		}
		s.logger.Error("repository lookup failed", slog.Any("error", err), slog.String("owner", rt.Owner), slog.String("repo", rt.Repo))
		http.Error(w, "repository lookup failed", http.StatusInternalServerError)
		return
	}

	if err := s.authorize(rt.Service, repository, principal, authenticated); err != nil {
		switch {
		case errors.Is(err, errAuthRequired):
			s.challenge(w)
		case errors.Is(err, errAuthUnavailable):
			http.Error(w, "authentication is not configured", http.StatusServiceUnavailable)
		case errors.Is(err, errForbidden):
			http.Error(w, "repository access denied", http.StatusForbidden)
		default:
			http.Error(w, "repository access denied", http.StatusForbidden)
		}
		return
	}

	gitPath, err := resolveGitPath(repository.GitPath, s.cfg.RepositoryRoot)
	if err != nil {
		s.logger.Error("repository path is unavailable",
			slog.Any("error", err),
			slog.String("repository_id", repository.ID),
			slog.String("git_path", repository.GitPath),
		)
		http.Error(w, "repository storage unavailable", http.StatusInternalServerError)
		return
	}
	repository.GitPath = gitPath

	if rt.RPC {
		s.handleRPC(w, r, rt.Service, repository)
		return
	}
	s.handleInfoRefs(w, r, rt.Service, repository)
}

func (s *Server) authenticateRequest(r *http.Request) (auth.Principal, bool, error) {
	token := tokenFromRequest(r, s.cfg.AccessCookieName)
	if token == "" {
		return auth.Principal{}, false, nil
	}
	if s.auth == nil || !s.auth.Configured() {
		return auth.Principal{}, false, errAuthUnavailable
	}
	principal, err := s.auth.Authenticate(r.Context(), token)
	if err != nil {
		return auth.Principal{}, false, err
	}
	return principal, true, nil
}

func (s *Server) authorize(service string, repo Repository, principal auth.Principal, authenticated bool) error {
	if service == serviceReceivePack {
		if !authenticated {
			if s.auth == nil || !s.auth.Configured() {
				return errAuthUnavailable
			}
			return errAuthRequired
		}
		if canPush(repo, principal) {
			return nil
		}
		return errForbidden
	}

	if canView(repo, principal, authenticated) {
		return nil
	}
	if !authenticated {
		if s.auth == nil || !s.auth.Configured() {
			return errAuthUnavailable
		}
		return errAuthRequired
	}
	return errForbidden
}

func (s *Server) handleInfoRefs(w http.ResponseWriter, r *http.Request, service string, repo Repository) {
	setNoCache(w)
	w.Header().Set("Content-Type", fmt.Sprintf("application/x-%s-advertisement", service))

	if _, err := io.WriteString(w, packetLine("# service="+service+"\n")+"0000"); err != nil {
		return
	}

	cmd := s.gitCommand(r, service, "--stateless-rpc", "--advertise-refs", repo.GitPath)
	if err := runGit(cmd, nil, w); err != nil {
		s.logger.Warn("git advertise-refs failed",
			slog.Any("error", err),
			slog.String("service", service),
			slog.String("repository_id", repo.ID),
		)
	}
}

func (s *Server) handleRPC(w http.ResponseWriter, r *http.Request, service string, repo Repository) {
	defer r.Body.Close()

	setNoCache(w)
	w.Header().Set("Content-Type", fmt.Sprintf("application/x-%s-result", service))

	cmd := s.gitCommand(r, service, "--stateless-rpc", repo.GitPath)
	if err := runGit(cmd, r.Body, w); err != nil {
		s.logger.Warn("git rpc failed",
			slog.Any("error", err),
			slog.String("service", service),
			slog.String("repository_id", repo.ID),
		)
	}
}

func (s *Server) gitCommand(r *http.Request, service string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(r.Context(), s.gitBin, append([]string{strings.TrimPrefix(service, "git-")}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	if protocol := strings.TrimSpace(r.Header.Get("Git-Protocol")); protocol != "" {
		cmd.Env = append(cmd.Env, "GIT_PROTOCOL="+protocol)
	}
	return cmd
}

func runGit(cmd *exec.Cmd, stdin io.Reader, stdout io.Writer) error {
	cmd.Stdin = stdin
	cmd.Stdout = stdout

	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			return err
		}
		return fmt.Errorf("%w: %s", err, message)
	}
	return nil
}

func parseRoute(r *http.Request) (route, error) {
	path := strings.Trim(r.URL.EscapedPath(), "/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return route{}, errNotFound
	}

	owner, err := url.PathUnescape(parts[0])
	if err != nil {
		return route{}, err
	}
	repoSegment, err := url.PathUnescape(parts[1])
	if err != nil {
		return route{}, err
	}
	if !strings.HasSuffix(repoSegment, ".git") {
		return route{}, errNotFound
	}
	repo := strings.TrimSuffix(repoSegment, ".git")
	if owner == "" || repo == "" || strings.Contains(owner, "/") || strings.Contains(repo, "/") {
		return route{}, errNotFound
	}

	if r.Method == http.MethodGet && len(parts) == 4 && parts[2] == "info" && parts[3] == "refs" {
		service := strings.TrimSpace(r.URL.Query().Get("service"))
		if service != serviceUploadPack && service != serviceReceivePack {
			return route{}, errNotFound
		}
		return route{Owner: owner, Repo: repo, Service: service}, nil
	}

	if r.Method == http.MethodPost && len(parts) == 3 {
		service := parts[2]
		if service != serviceUploadPack && service != serviceReceivePack {
			return route{}, errNotFound
		}
		return route{Owner: owner, Repo: repo, Service: service, RPC: true}, nil
	}

	return route{}, errNotFound
}

func tokenFromRequest(r *http.Request, accessCookieName string) string {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return strings.TrimSpace(header[len("bearer "):])
	}
	if strings.HasPrefix(strings.ToLower(header), "basic ") {
		encoded := strings.TrimSpace(header[len("basic "):])
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return ""
		}
		username, password, ok := strings.Cut(string(decoded), ":")
		if ok && strings.TrimSpace(password) != "" {
			return strings.TrimSpace(password)
		}
		return strings.TrimSpace(username)
	}

	if accessCookieName != "" {
		if cookie, err := r.Cookie(accessCookieName); err == nil {
			return cookie.Value
		}
	}
	return ""
}

func canView(repo Repository, principal auth.Principal, authenticated bool) bool {
	switch strings.ToLower(strings.TrimSpace(repo.Visibility)) {
	case "public":
		return true
	case "internal":
		return authenticated
	default:
		if repo.OwnerType == "user" && repo.OwnerID != "" && repo.OwnerID == principal.UserID {
			return true
		}
		return normalizeOrgRole(repo.UserRole) != ""
	}
}

func canPush(repo Repository, principal auth.Principal) bool {
	if repo.OwnerType == "user" && repo.OwnerID != "" && repo.OwnerID == principal.UserID {
		return true
	}
	switch normalizeOrgRole(repo.UserRole) {
	case "owner", "admin":
		return true
	default:
		return false
	}
}

func normalizeOrgRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "owner", "admin", "member":
		return strings.ToLower(strings.TrimSpace(role))
	default:
		return ""
	}
}

func resolveGitPath(gitPath, root string) (string, error) {
	gitPath = strings.TrimSpace(gitPath)
	if gitPath == "" {
		return "", errors.New("empty git path")
	}
	if !filepath.IsAbs(gitPath) && strings.TrimSpace(root) != "" {
		gitPath = filepath.Join(root, gitPath)
	}
	stat, err := os.Stat(gitPath)
	if err != nil {
		return "", err
	}
	if !stat.IsDir() {
		return "", fmt.Errorf("%s is not a directory", gitPath)
	}
	if _, err := os.Stat(filepath.Join(gitPath, "HEAD")); err != nil {
		return "", err
	}
	return gitPath, nil
}

func packetLine(payload string) string {
	return fmt.Sprintf("%04x%s", len(payload)+4, payload)
}

func setNoCache(w http.ResponseWriter) {
	w.Header().Set("Expires", "Fri, 01 Jan 1980 00:00:00 GMT")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Cache-Control", "no-cache, max-age=0, must-revalidate")
}

func (s *Server) challenge(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Clove Git"`)
	http.Error(w, "authentication required", http.StatusUnauthorized)
}

func (s *Server) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		s.logger.Info("request completed",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", recorder.status),
			slog.Duration("duration", time.Since(start)),
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func writeJSON(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, body)
}

type DBStore struct {
	DB *sql.DB
}

func (s DBStore) FindRepository(ctx context.Context, owner, name, userID string) (Repository, error) {
	if s.DB == nil {
		return Repository{}, errors.New("database is not configured")
	}

	var repo Repository
	err := s.DB.QueryRowContext(ctx, `
		SELECT r.id, r.owner_type, r.owner_id,
			CASE WHEN r.owner_type = 'organization' THEN o.name ELSE u.username END AS owner,
			r.name, r.visibility, r.git_path,
			COALESCE(
				CASE
					WHEN r.owner_type = 'organization' AND o.owner_id = $3 THEN 'owner'
					WHEN r.owner_type = 'organization' THEN om.role
					ELSE ''
				END,
				''
			) AS user_role
		FROM repositories r
		LEFT JOIN organizations o ON r.owner_type = 'organization' AND r.owner_id = o.id
		LEFT JOIN users u ON r.owner_type = 'user' AND r.owner_id = u.id
		LEFT JOIN organization_members om
			ON r.owner_type = 'organization'
			AND om.organization_id = r.owner_id
			AND om.user_id = $3
		WHERE (o.name = $1 OR u.username = $1) AND r.name = $2
	`, owner, name, userID).Scan(
		&repo.ID,
		&repo.OwnerType,
		&repo.OwnerID,
		&repo.Owner,
		&repo.Name,
		&repo.Visibility,
		&repo.GitPath,
		&repo.UserRole,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Repository{}, errNotFound
	}
	if err != nil {
		return Repository{}, err
	}
	return repo, nil
}
