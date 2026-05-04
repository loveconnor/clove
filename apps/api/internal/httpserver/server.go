package httpserver

import (
	"database/sql"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"clove/apps/api/internal/auth"
	"clove/apps/api/pkg/apierror"
	"clove/apps/api/pkg/config"
)

type Dependencies struct {
	Config config.Config
	DB     *sql.DB
	Logger *slog.Logger
	Auth   auth.Authenticator
}

type Server struct {
	cfg    config.Config
	db     *sql.DB
	logger *slog.Logger
	auth   auth.Authenticator
	mux    *http.ServeMux
}

func New(deps Dependencies) http.Handler {
	authenticator := deps.Auth
	if authenticator == nil {
		authenticator = auth.NewService(deps.Config)
	}

	s := &Server{
		cfg:    deps.Config,
		db:     deps.DB,
		logger: deps.Logger,
		auth:   authenticator,
		mux:    http.NewServeMux(),
	}

	s.routes()

	return s.requestID(s.recoverer(s.requestLogger(s.mux)))
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /api/auth/register", s.handleRegisterRedirect)
	s.mux.HandleFunc("POST /api/auth/register", s.handleRegister)
	s.mux.HandleFunc("GET /api/auth/login", s.handleLoginRedirect)
	s.mux.HandleFunc("POST /api/auth/login", s.handleLogin)
	s.mux.HandleFunc("GET /api/auth/callback", s.handleAuthCallback)
	s.mux.HandleFunc("POST /api/auth/callback", s.handleAuthCallback)
	s.mux.HandleFunc("GET /api/auth/logout", s.handleLogout)
	s.mux.HandleFunc("POST /api/auth/logout", s.handleLogout)
	s.mux.Handle("GET /api/me", s.requireSession(http.HandlerFunc(s.handleMe)))
	s.mux.Handle("GET /api/organizations", s.requireSession(http.HandlerFunc(s.handleOrganizations)))
	s.mux.Handle("POST /api/organizations", s.requireSession(http.HandlerFunc(s.handleCreateOrganization)))
	s.mux.Handle("GET /api/organizations/{owner}", s.requireSession(http.HandlerFunc(s.handleOrganization)))
	s.mux.Handle("PATCH /api/organizations/{owner}", s.requireSession(http.HandlerFunc(s.handleUpdateOrganization)))
	s.mux.Handle("GET /api/organizations/{owner}/members", s.requireSession(http.HandlerFunc(s.handleOrganizationMembers)))
	s.mux.Handle("POST /api/organizations/{owner}/invitations", s.requireSession(http.HandlerFunc(s.handleCreateOrganizationInvitation)))
	s.mux.Handle("GET /api/repositories", s.requireSession(http.HandlerFunc(s.handleRepositories)))
	s.mux.Handle("POST /api/repositories", s.requireSession(http.HandlerFunc(s.handleCreateRepository)))
	s.mux.Handle("GET /api/repositories/{owner}/{repo}", s.requireSession(http.HandlerFunc(s.handleRepository)))
	s.mux.Handle("GET /api/repositories/{owner}/{repo}/tree", s.requireSession(http.HandlerFunc(s.handleRepositoryTree)))
	s.mux.Handle("GET /api/repositories/{owner}/{repo}/blob", s.requireSession(http.HandlerFunc(s.handleRepositoryBlob)))
	s.mux.Handle("GET /api/personal-access-tokens", s.requireSession(http.HandlerFunc(s.handlePersonalAccessTokens)))
	s.mux.Handle("POST /api/personal-access-tokens", s.requireSession(http.HandlerFunc(s.handleCreatePersonalAccessToken)))
	s.mux.Handle("DELETE /api/personal-access-tokens/{id}", s.requireSession(http.HandlerFunc(s.handleDeletePersonalAccessToken)))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	dbStatus := "not_configured"

	if s.db != nil {
		ctx := r.Context()
		if err := s.db.PingContext(ctx); err != nil {
			status = "degraded"
			dbStatus = "unavailable"
		} else {
			dbStatus = "ok"
		}
	}

	apierror.Respond(w, r, http.StatusOK, map[string]any{
		"status":     status,
		"service":    s.cfg.AppName,
		"env":        s.cfg.Environment,
		"database":   dbStatus,
		"request_id": RequestID(r.Context()),
	})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		apierror.Write(w, r, apierror.Unauthorized("authentication required"))
		return
	}
	if err := s.upsertUser(r, principal); err != nil {
		apierror.Write(w, r, err)
		return
	}

	apierror.Respond(w, r, http.StatusOK, map[string]any{
		"user":            principal.User,
		"session_id":      principal.SessionID,
		"organization_id": principal.OrganizationID,
		"role":            principal.Role,
		"permissions":     principal.Permissions,
		"request_id":      RequestID(r.Context()),
	})
}

func (s *Server) requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = newRequestID()
		}

		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, withRequestID(r, requestID))
	})
}

func (s *Server) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(recorder, r)

		s.logger.Info("request completed",
			slog.String("request_id", RequestID(r.Context())),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", recorder.status),
			slog.Duration("duration", time.Since(start)),
		)
	})
}

func (s *Server) recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if value := recover(); value != nil {
				s.logger.Error("request panic recovered",
					slog.String("request_id", RequestID(r.Context())),
					slog.Any("panic", value),
					slog.String("stack", string(debug.Stack())),
				)
				apierror.Write(w, r, apierror.Internal("Something went wrong."))
			}
		}()

		next.ServeHTTP(w, r)
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
