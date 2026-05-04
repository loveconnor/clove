package gitservice

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
	Authenticate(ctx context.Context, username, token string) (Principal, error)
}

type RepositoryStore interface {
	FindRepository(ctx context.Context, owner, name, userID string) (Repository, error)
}

type PushRecorder interface {
	RecordPush(ctx context.Context, repo Repository, pusher Principal, updates []RefUpdate) error
}

type Server struct {
	cfg      config.Config
	store    RepositoryStore
	recorder PushRecorder
	auth     Authenticator
	logger   *slog.Logger
	gitBin   string
	mux      *http.ServeMux
}

type Dependencies struct {
	Config   config.Config
	Store    RepositoryStore
	Recorder PushRecorder
	Auth     Authenticator
	Logger   *slog.Logger
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

type Principal struct {
	UserID   string
	Username string
	Email    string
}

type Credentials struct {
	Username string
	Password string
}

type RefUpdate struct {
	RefName string
	OldSHA  string
	NewSHA  string
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

	recorder := deps.Recorder
	if recorder == nil {
		if storeRecorder, ok := deps.Store.(PushRecorder); ok {
			recorder = storeRecorder
		}
	}

	s := &Server{
		cfg:      deps.Config,
		store:    deps.Store,
		recorder: recorder,
		auth:     deps.Auth,
		logger:   logger,
		gitBin:   gitBin,
		mux:      http.NewServeMux(),
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
		s.handleRPC(w, r, rt.Service, repository, principal)
		return
	}
	s.handleInfoRefs(w, r, rt.Service, repository)
}

func (s *Server) authenticateRequest(r *http.Request) (Principal, bool, error) {
	credentials := credentialsFromRequest(r)
	if credentials.Username == "" && credentials.Password == "" {
		return Principal{}, false, nil
	}
	if credentials.Username == "" || credentials.Password == "" {
		return Principal{}, false, errAuthRequired
	}
	if s.auth == nil {
		return Principal{}, false, errAuthUnavailable
	}
	principal, err := s.auth.Authenticate(r.Context(), credentials.Username, credentials.Password)
	if err != nil {
		return Principal{}, false, err
	}
	return principal, true, nil
}

func (s *Server) authorize(service string, repo Repository, principal Principal, authenticated bool) error {
	if service == serviceReceivePack {
		if !authenticated {
			if s.auth == nil {
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
		if s.auth == nil {
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

func (s *Server) handleRPC(w http.ResponseWriter, r *http.Request, service string, repo Repository, principal Principal) {
	defer r.Body.Close()

	setNoCache(w)
	w.Header().Set("Content-Type", fmt.Sprintf("application/x-%s-result", service))

	stdin := io.Reader(r.Body)
	var updates []RefUpdate
	if service == serviceReceivePack {
		prefix, parsedUpdates, err := readReceivePackCommands(r.Body)
		if err != nil {
			s.logger.Warn("git receive-pack command parsing failed",
				slog.Any("error", err),
				slog.String("repository_id", repo.ID),
			)
			http.Error(w, "invalid receive-pack request", http.StatusBadRequest)
			return
		}
		stdin = io.MultiReader(bytes.NewReader(prefix), r.Body)
		updates = parsedUpdates
	}

	cmd := s.gitCommand(r, service, "--stateless-rpc", repo.GitPath)
	if err := runGit(cmd, stdin, w); err != nil {
		s.logger.Warn("git rpc failed",
			slog.Any("error", err),
			slog.String("service", service),
			slog.String("repository_id", repo.ID),
		)
		return
	}

	if service == serviceReceivePack && len(updates) > 0 && s.recorder != nil {
		acceptedUpdates := s.acceptedRefUpdates(r.Context(), repo, updates)
		if len(acceptedUpdates) == 0 {
			return
		}
		if err := s.recorder.RecordPush(r.Context(), repo, principal, acceptedUpdates); err != nil {
			s.logger.Error("push event recording failed",
				slog.Any("error", err),
				slog.String("repository_id", repo.ID),
				slog.Int("ref_updates", len(acceptedUpdates)),
			)
		}
	}
}

func (s *Server) acceptedRefUpdates(ctx context.Context, repo Repository, updates []RefUpdate) []RefUpdate {
	accepted := make([]RefUpdate, 0, len(updates))
	for _, update := range updates {
		if refUpdateApplied(ctx, s.gitBin, repo.GitPath, update) {
			accepted = append(accepted, update)
		}
	}
	return accepted
}

func refUpdateApplied(ctx context.Context, gitBin, gitPath string, update RefUpdate) bool {
	if allZeroSHA(update.NewSHA) {
		cmd := exec.CommandContext(ctx, gitBin, "--git-dir", gitPath, "rev-parse", "--verify", "--quiet", update.RefName)
		return cmd.Run() != nil
	}

	cmd := exec.CommandContext(ctx, gitBin, "--git-dir", gitPath, "rev-parse", "--verify", "--quiet", update.RefName)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(string(output)), strings.TrimSpace(update.NewSHA))
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

func readReceivePackCommands(body io.Reader) ([]byte, []RefUpdate, error) {
	var prefix bytes.Buffer
	var updates []RefUpdate

	for {
		header := make([]byte, 4)
		if _, err := io.ReadFull(body, header); err != nil {
			return nil, nil, err
		}
		prefix.Write(header)

		sizeText := string(header)
		size, err := strconv.ParseInt(sizeText, 16, 32)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid pkt-line size %q: %w", sizeText, err)
		}
		if size == 0 {
			return prefix.Bytes(), updates, nil
		}
		if size < 4 {
			return nil, nil, fmt.Errorf("invalid pkt-line size %d", size)
		}

		payload := make([]byte, int(size)-4)
		if _, err := io.ReadFull(body, payload); err != nil {
			return nil, nil, err
		}
		prefix.Write(payload)

		if update, ok := parseReceivePackCommand(payload); ok {
			updates = append(updates, update)
		}
	}
}

func parseReceivePackCommand(payload []byte) (RefUpdate, bool) {
	line := string(payload)
	if beforeCapabilities, _, ok := strings.Cut(line, "\x00"); ok {
		line = beforeCapabilities
	}
	line = strings.TrimSpace(line)
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return RefUpdate{}, false
	}
	return RefUpdate{
		OldSHA:  parts[0],
		NewSHA:  parts[1],
		RefName: parts[2],
	}, true
}

func allZeroSHA(sha string) bool {
	sha = strings.TrimSpace(sha)
	if sha == "" {
		return false
	}
	for _, char := range sha {
		if char != '0' {
			return false
		}
	}
	return true
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

func credentialsFromRequest(r *http.Request) Credentials {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(header), "basic ") {
		encoded := strings.TrimSpace(header[len("basic "):])
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return Credentials{}
		}
		username, password, ok := strings.Cut(string(decoded), ":")
		if !ok {
			return Credentials{}
		}
		return Credentials{Username: strings.TrimSpace(username), Password: strings.TrimSpace(password)}
	}
	return Credentials{}
}

func canView(repo Repository, principal Principal, authenticated bool) bool {
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

func canPush(repo Repository, principal Principal) bool {
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

func (s DBStore) RecordPush(ctx context.Context, repo Repository, pusher Principal, updates []RefUpdate) error {
	if s.DB == nil {
		return errors.New("database is not configured")
	}
	if repo.ID == "" || pusher.UserID == "" || len(updates) == 0 {
		return nil
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, update := range updates {
		update.RefName = strings.TrimSpace(update.RefName)
		update.OldSHA = strings.TrimSpace(update.OldSHA)
		update.NewSHA = strings.TrimSpace(update.NewSHA)
		if update.RefName == "" || update.OldSHA == "" || update.NewSHA == "" {
			continue
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO git_refs (repo_id, ref_name, old_sha, new_sha, pusher_id)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (repo_id, ref_name) DO UPDATE SET
				old_sha = EXCLUDED.old_sha,
				new_sha = EXCLUDED.new_sha,
				pusher_id = EXCLUDED.pusher_id,
				created_at = now()
		`, repo.ID, update.RefName, update.OldSHA, update.NewSHA, pusher.UserID); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO push_events (id, repo_id, ref_name, old_sha, new_sha, pusher_id)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, newPushEventID(), repo.ID, update.RefName, update.OldSHA, update.NewSHA, pusher.UserID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func newPushEventID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err == nil {
		return "push_" + hex.EncodeToString(b[:])
	}
	return fmt.Sprintf("push_%d", time.Now().UnixNano())
}

type DBAuth struct {
	DB *sql.DB
}

func (a DBAuth) Authenticate(ctx context.Context, username, token string) (Principal, error) {
	if a.DB == nil {
		return Principal{}, errAuthUnavailable
	}

	username = strings.TrimSpace(username)
	token = strings.TrimSpace(token)
	if username == "" || token == "" {
		return Principal{}, errAuthRequired
	}

	var principal Principal
	err := a.DB.QueryRowContext(ctx, `
		UPDATE personal_access_tokens pat
		SET last_used_at = now()
		FROM users u
		WHERE pat.user_id = u.id
			AND lower(u.username) = lower($1)
			AND pat.token_hash = $2
		RETURNING u.id, u.username, COALESCE(u.email, '')
	`, username, hashPersonalAccessToken(token)).Scan(
		&principal.UserID,
		&principal.Username,
		&principal.Email,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Principal{}, errAuthRequired
	}
	if err != nil {
		return Principal{}, err
	}
	return principal, nil
}

func hashPersonalAccessToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
