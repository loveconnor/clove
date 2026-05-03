package gitservice

import (
	"context"
	"encoding/base64"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"clove/apps/git-http/internal/auth"
	"clove/apps/git-http/internal/config"
)

func TestParseInfoRefsRoute(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/clove/api.git/info/refs?service=git-upload-pack", nil)
	rt, err := parseRoute(req)
	if err != nil {
		t.Fatalf("parse route: %v", err)
	}
	if rt.Owner != "clove" || rt.Repo != "api" || rt.Service != serviceUploadPack || rt.RPC {
		t.Fatalf("unexpected route: %#v", rt)
	}
}

func TestTokenFromBasicPassword(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	encoded := base64.StdEncoding.EncodeToString([]byte("anything:access-token"))
	req.Header.Set("Authorization", "Basic "+encoded)

	if got := tokenFromRequest(req, "access"); got != "access-token" {
		t.Fatalf("expected access-token, got %q", got)
	}
}

func TestPrivateFetchRequiresAuthentication(t *testing.T) {
	t.Parallel()

	handler := New(Dependencies{
		Config: config.Config{AppName: "clove-git-http", GitBin: "git"},
		Store: fakeStore{repo: Repository{
			ID:         "repo_123",
			OwnerType:  "user",
			OwnerID:    "user_owner",
			Owner:      "owner",
			Name:       "repo",
			Visibility: "private",
			GitPath:    filepath.Join(t.TempDir(), "missing.git"),
		}},
		Auth:   fakeAuth{configured: true},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	req := httptest.NewRequest(http.MethodGet, "/owner/repo.git/info/refs?service=git-upload-pack", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if rec.Header().Get("WWW-Authenticate") == "" {
		t.Fatal("expected basic challenge")
	}
}

func TestPushRequiresWriteAuthorization(t *testing.T) {
	t.Parallel()

	gitPath := initBareRepo(t)
	handler := New(Dependencies{
		Config: config.Config{AppName: "clove-git-http", GitBin: "git"},
		Store: fakeStore{repo: Repository{
			ID:         "repo_123",
			OwnerType:  "organization",
			OwnerID:    "org_123",
			Owner:      "owner",
			Name:       "repo",
			Visibility: "private",
			GitPath:    gitPath,
			UserRole:   "member",
		}},
		Auth:   fakeAuth{configured: true},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	req := httptest.NewRequest(http.MethodPost, "/owner/repo.git/git-receive-pack", strings.NewReader(""))
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestPublicInfoRefsUsesNativeGit(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary is not available")
	}

	gitPath := initBareRepo(t)
	handler := New(Dependencies{
		Config: config.Config{AppName: "clove-git-http", GitBin: "git"},
		Store: fakeStore{repo: Repository{
			ID:         "repo_123",
			OwnerType:  "user",
			OwnerID:    "user_owner",
			Owner:      "owner",
			Name:       "repo",
			Visibility: "public",
			GitPath:    gitPath,
		}},
		Auth:   fakeAuth{},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	req := httptest.NewRequest(http.MethodGet, "/owner/repo.git/info/refs?service=git-upload-pack", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "application/x-git-upload-pack-advertisement" {
		t.Fatalf("unexpected content-type %q", got)
	}
	if !strings.HasPrefix(rec.Body.String(), packetLine("# service=git-upload-pack\n")+"0000") {
		t.Fatalf("response does not start with service advertisement: %q", rec.Body.String())
	}
}

func TestCloneCommitAndPushOverHTTP(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary is not available")
	}

	gitPath := initBareRepo(t)
	handler := New(Dependencies{
		Config: config.Config{AppName: "clove-git-http", GitBin: "git"},
		Store: fakeStore{repo: Repository{
			ID:         "repo_123",
			OwnerType:  "organization",
			OwnerID:    "org_123",
			Owner:      "owner",
			Name:       "repo",
			Visibility: "private",
			GitPath:    gitPath,
			UserRole:   "admin",
		}},
		Auth:   fakeAuth{configured: true},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	remoteURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse test server url: %v", err)
	}
	remoteURL.User = url.UserPassword("tester", "valid-token")
	remoteURL.Path = "/owner/repo.git"

	workdir := t.TempDir()
	cloneDir := filepath.Join(workdir, "clone")
	runGitForTest(t, workdir, "clone", remoteURL.String(), cloneDir)
	runGitForTest(t, cloneDir, "config", "user.email", "test@example.com")
	runGitForTest(t, cloneDir, "config", "user.name", "Test User")
	if err := os.WriteFile(filepath.Join(cloneDir, "README.md"), []byte("# test\n"), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}
	runGitForTest(t, cloneDir, "add", "README.md")
	runGitForTest(t, cloneDir, "commit", "-m", "Initial commit")
	runGitForTest(t, cloneDir, "push", "origin", "HEAD:main")

	output := runGitForTest(t, workdir, "--git-dir", gitPath, "rev-parse", "--verify", "refs/heads/main")
	if strings.TrimSpace(output) == "" {
		t.Fatal("expected pushed main ref")
	}
}

func initBareRepo(t *testing.T) string {
	t.Helper()
	gitPath := filepath.Join(t.TempDir(), "repo.git")
	cmd := exec.Command("git", "init", "--bare", gitPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init --bare: %v: %s", err, output)
	}
	return gitPath
}

func runGitForTest(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, output)
	}
	return string(output)
}

type fakeStore struct {
	repo Repository
	err  error
}

func (s fakeStore) FindRepository(context.Context, string, string, string) (Repository, error) {
	if s.err != nil {
		return Repository{}, s.err
	}
	return s.repo, nil
}

type fakeAuth struct {
	configured bool
}

func (a fakeAuth) Authenticate(_ context.Context, token string) (auth.Principal, error) {
	if token != "valid-token" {
		return auth.Principal{}, errAuthRequired
	}
	return auth.Principal{UserID: "user_123", Username: "tester"}, nil
}

func (a fakeAuth) Configured() bool {
	return a.configured
}
