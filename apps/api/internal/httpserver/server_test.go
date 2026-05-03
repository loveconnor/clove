package httpserver

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"clove/apps/api/internal/auth"
	"clove/apps/api/internal/users"
	"clove/apps/api/pkg/apierror"
	"clove/apps/api/pkg/config"
)

func TestHealth(t *testing.T) {
	t.Parallel()

	handler := New(Dependencies{
		Config: config.Config{AppName: "clove-api", Environment: "test"},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Auth:   fakeAuth{},
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-Request-ID", "test-request-id")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if got := rec.Header().Get("X-Request-ID"); got != "test-request-id" {
		t.Fatalf("expected request id header to round trip, got %q", got)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body["status"] != "ok" {
		t.Fatalf("expected ok status, got %#v", body["status"])
	}
	if body["database"] != "not_configured" {
		t.Fatalf("expected not_configured database, got %#v", body["database"])
	}
	if body["request_id"] != "test-request-id" {
		t.Fatalf("expected body request_id to round trip, got %#v", body["request_id"])
	}
}

func TestMe(t *testing.T) {
	t.Parallel()

	handler := New(Dependencies{
		Config: config.Config{
			AppName:           "clove-api",
			Environment:       "test",
			AccessCookieName:  "access",
			RefreshCookieName: "refresh",
			StateCookieName:   "state",
		},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Auth:   fakeAuth{},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.AddCookie(&http.Cookie{Name: "access", Value: "valid-token"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body struct {
		User struct {
			ID          string `json:"id"`
			Username    string `json:"username"`
			Email       string `json:"email"`
			DisplayName string `json:"display_name"`
		} `json:"user"`
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.User.Username != "tester" {
		t.Fatalf("expected local dev user, got %#v", body.User)
	}
	if body.RequestID == "" {
		t.Fatal("expected generated request id")
	}
}

func TestMeRequiresSession(t *testing.T) {
	t.Parallel()

	handler := New(Dependencies{
		Config: config.Config{
			AppName:           "clove-api",
			Environment:       "test",
			AccessCookieName:  "access",
			RefreshCookieName: "refresh",
			StateCookieName:   "state",
		},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Auth:   fakeAuth{},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestLoginSetsSessionCookies(t *testing.T) {
	t.Parallel()

	handler := New(Dependencies{
		Config: config.Config{
			AppName:           "clove-api",
			Environment:       "test",
			AccessCookieName:  "access",
			RefreshCookieName: "refresh",
			StateCookieName:   "state",
		},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Auth:   fakeAuth{},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"email":"test@example.com","password":"secret"}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	cookies := rec.Result().Cookies()
	if !hasCookie(cookies, "access", "valid-token") {
		t.Fatalf("expected access cookie, got %#v", cookies)
	}
	if !hasCookie(cookies, "refresh", "refresh-token") {
		t.Fatalf("expected refresh cookie, got %#v", cookies)
	}
}

type fakeAuth struct{}

func (fakeAuth) Register(context.Context, auth.RegisterRequest, auth.RequestMetadata) (auth.Session, error) {
	return fakeSession(), nil
}

func (fakeAuth) Login(context.Context, auth.LoginRequest, auth.RequestMetadata) (auth.Session, error) {
	return fakeSession(), nil
}

func (fakeAuth) AuthenticateCode(context.Context, string, auth.RequestMetadata) (auth.Session, error) {
	return fakeSession(), nil
}

func (fakeAuth) Refresh(context.Context, string, auth.RequestMetadata) (auth.Session, error) {
	return fakeSession(), nil
}

func (fakeAuth) ValidateAccessToken(_ context.Context, accessToken string) (auth.Principal, error) {
	if accessToken != "valid-token" {
		return auth.Principal{}, apierror.Unauthorized("bad token")
	}
	return auth.Principal{
		User:      fakeUser(),
		SessionID: "session_123",
	}, nil
}

func (fakeAuth) RevokeSession(context.Context, string) error {
	return nil
}

func (fakeAuth) AuthorizationURL(string, string, string) (string, error) {
	return "https://api.workos.com/user_management/authorize", nil
}

func (fakeAuth) Enabled() bool {
	return true
}

func fakeSession() auth.Session {
	return auth.Session{
		AccessToken:  "valid-token",
		RefreshToken: "refresh-token",
		User:         fakeUser(),
		SessionID:    "session_123",
	}
}

func fakeUser() users.User {
	return users.User{
		ID:          "user_123",
		Username:    "tester",
		Email:       "test@example.com",
		DisplayName: "Test User",
	}
}

func hasCookie(cookies []*http.Cookie, name, value string) bool {
	for _, cookie := range cookies {
		if cookie.Name == name && cookie.Value == value {
			return true
		}
	}
	return false
}
