package httpserver

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"clove/apps/api/internal/auth"
	"clove/apps/api/pkg/apierror"
)

const (
	accessCookieMaxAge  = 60 * 60
	refreshCookieMaxAge = 60 * 60 * 24 * 30
	stateCookieMaxAge   = 5 * 60
)

func (s *Server) handleRegisterRedirect(w http.ResponseWriter, r *http.Request) {
	s.beginHostedAuth(w, r, "sign-up")
}

func (s *Server) handleLoginRedirect(w http.ResponseWriter, r *http.Request) {
	s.beginHostedAuth(w, r, "sign-in")
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req auth.RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		apierror.Write(w, r, err)
		return
	}

	session, err := s.auth.Register(r.Context(), req, requestMetadata(r))
	if err != nil {
		s.writeAuthError(w, r, err)
		return
	}

	s.setSessionCookies(w, session)
	apierror.Respond(w, r, http.StatusCreated, map[string]any{
		"user":       session.User,
		"request_id": RequestID(r.Context()),
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req auth.LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		apierror.Write(w, r, err)
		return
	}

	session, err := s.auth.Login(r.Context(), req, requestMetadata(r))
	if err != nil {
		s.writeAuthError(w, r, err)
		return
	}

	s.setSessionCookies(w, session)
	apierror.Respond(w, r, http.StatusOK, map[string]any{
		"user":       session.User,
		"request_id": RequestID(r.Context()),
	})
}

func (s *Server) handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.handleAuthCallbackJSON(w, r)
		return
	}

	code := strings.TrimSpace(r.URL.Query().Get("code"))
	state := strings.TrimSpace(r.URL.Query().Get("state"))
	if err := s.validateState(r, state); err != nil {
		apierror.Write(w, r, err)
		return
	}

	session, err := s.auth.AuthenticateCode(r.Context(), code, requestMetadata(r))
	if err != nil {
		s.writeAuthError(w, r, err)
		return
	}

	s.setSessionCookies(w, session)
	s.clearCookie(w, s.cfg.StateCookieName)
	http.Redirect(w, r, s.cfg.AuthSuccessURL, http.StatusFound)
}

func (s *Server) handleAuthCallbackJSON(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
	}
	if err := decodeJSON(r, &req); err != nil {
		apierror.Write(w, r, err)
		return
	}

	session, err := s.auth.AuthenticateCode(r.Context(), req.Code, requestMetadata(r))
	if err != nil {
		s.writeAuthError(w, r, err)
		return
	}

	s.setSessionCookies(w, session)
	apierror.Respond(w, r, http.StatusOK, map[string]any{
		"user":       session.User,
		"request_id": RequestID(r.Context()),
	})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	sessionID := ""
	if accessToken := s.cookieValue(r, s.cfg.AccessCookieName); accessToken != "" {
		if principal, err := s.auth.ValidateAccessToken(r.Context(), accessToken); err == nil {
			sessionID = principal.SessionID
		}
	}

	s.clearSessionCookies(w)
	if sessionID != "" {
		if err := s.auth.RevokeSession(r.Context(), sessionID); err != nil {
			s.logger.Warn("failed to revoke workos session",
				slog.String("request_id", RequestID(r.Context())),
				slog.Any("error", err),
			)
		}
	}

	if r.Method == http.MethodGet {
		http.Redirect(w, r, s.cfg.AuthLogoutURL, http.StatusFound)
		return
	}

	apierror.Respond(w, r, http.StatusOK, map[string]any{
		"ok":         true,
		"request_id": RequestID(r.Context()),
	})
}

func (s *Server) requireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken := s.cookieValue(r, s.cfg.AccessCookieName)
		principal, err := s.auth.ValidateAccessToken(r.Context(), accessToken)
		if err == nil {
			next.ServeHTTP(w, r.WithContext(auth.WithPrincipal(r.Context(), principal)))
			return
		}

		refreshToken := s.cookieValue(r, s.cfg.RefreshCookieName)
		if refreshToken != "" {
			session, refreshErr := s.auth.Refresh(r.Context(), refreshToken, requestMetadata(r))
			if refreshErr == nil {
				s.setSessionCookies(w, session)
				principal = auth.Principal{
					User:           session.User,
					SessionID:      session.SessionID,
					OrganizationID: session.OrganizationID,
					Role:           session.Role,
					Permissions:    session.Permissions,
					Entitlements:   session.Entitlements,
				}
				next.ServeHTTP(w, r.WithContext(auth.WithPrincipal(r.Context(), principal)))
				return
			}
		}

		s.clearSessionCookies(w)
		apierror.Write(w, r, apierror.Unauthorized("authentication required"))
	})
}

func (s *Server) beginHostedAuth(w http.ResponseWriter, r *http.Request, screenHint string) {
	state, err := randomState()
	if err != nil {
		apierror.Write(w, r, apierror.Internal("Could not start authentication."))
		return
	}

	loginHint := strings.TrimSpace(r.URL.Query().Get("login_hint"))
	url, err := s.auth.AuthorizationURL(screenHint, state, loginHint)
	if err != nil {
		s.writeAuthError(w, r, err)
		return
	}

	http.SetCookie(w, s.cookie(s.cfg.StateCookieName, state, stateCookieMaxAge))
	http.Redirect(w, r, url, http.StatusFound)
}

func (s *Server) validateState(r *http.Request, state string) error {
	if state == "" {
		return apierror.BadRequest("state is required")
	}

	stateCookie, err := r.Cookie(s.cfg.StateCookieName)
	if err != nil || stateCookie.Value == "" {
		return apierror.BadRequest("state cookie is missing")
	}
	if stateCookie.Value != state {
		return apierror.BadRequest("state is invalid")
	}
	return nil
}

func (s *Server) setSessionCookies(w http.ResponseWriter, session auth.Session) {
	http.SetCookie(w, s.cookie(s.cfg.AccessCookieName, session.AccessToken, accessCookieMaxAge))
	http.SetCookie(w, s.cookie(s.cfg.RefreshCookieName, session.RefreshToken, refreshCookieMaxAge))
}

func (s *Server) clearSessionCookies(w http.ResponseWriter) {
	s.clearCookie(w, s.cfg.AccessCookieName)
	s.clearCookie(w, s.cfg.RefreshCookieName)
	s.clearCookie(w, s.cfg.StateCookieName)
}

func (s *Server) clearCookie(w http.ResponseWriter, name string) {
	cookie := s.cookie(name, "", -1)
	http.SetCookie(w, cookie)
}

func (s *Server) cookie(name, value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Domain:   s.cfg.CookieDomain,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   s.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	}
}

func (s *Server) cookieValue(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (s *Server) writeAuthError(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr apierror.Error
	if errors.As(err, &apiErr) {
		apierror.Write(w, r, apiErr)
		return
	}

	s.logger.Warn("authentication request failed",
		slog.String("request_id", RequestID(r.Context())),
		slog.Any("error", err),
	)
	apierror.Write(w, r, apierror.Unauthorized("authentication failed"))
}

func decodeJSON(r *http.Request, target any) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return apierror.BadRequest("invalid JSON request body")
	}
	return nil
}

func requestMetadata(r *http.Request) auth.RequestMetadata {
	return auth.RequestMetadata{
		IPAddress: clientIP(r),
		UserAgent: r.UserAgent(),
	}
}

func clientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ip := strings.TrimSpace(strings.Split(forwarded, ",")[0])
		if ip != "" {
			return ip
		}
	}
	realIP := strings.TrimSpace(r.Header.Get("X-Real-IP"))
	if realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func randomState() (string, error) {
	var buf [32]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf[:]), nil
}
