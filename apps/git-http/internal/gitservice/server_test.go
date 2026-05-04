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

func TestCredentialsFromBasicPassword(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	encoded := base64.StdEncoding.EncodeToString([]byte("anything:access-token"))
	req.Header.Set("Authorization", "Basic "+encoded)

	got := credentialsFromRequest(req)
	if got.Username != "anything" || got.Password != "access-token" {
		t.Fatalf("unexpected credentials: %#v", got)
	}
}

func TestReadReceivePackCommands(t *testing.T) {
	t.Parallel()

	payload := "0000000000000000000000000000000000000000 1111111111111111111111111111111111111111 refs/heads/main\x00 report-status\n"
	body := strings.NewReader(packetLine(payload) + "0000PACK")

	prefix, updates, err := readReceivePackCommands(body)
	if err != nil {
		t.Fatalf("read receive-pack commands: %v", err)
	}
	if string(prefix) != packetLine(payload)+"0000" {
		t.Fatalf("unexpected prefix %q", prefix)
	}
	if len(updates) != 1 {
		t.Fatalf("expected one update, got %#v", updates)
	}
	if updates[0].RefName != "refs/heads/main" || updates[0].OldSHA != strings.Repeat("0", 40) || updates[0].NewSHA != strings.Repeat("1", 40) {
		t.Fatalf("unexpected update: %#v", updates[0])
	}
	rest, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("read remaining body: %v", err)
	}
	if string(rest) != "PACK" {
		t.Fatalf("unexpected remaining body %q", rest)
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
	req.SetBasicAuth("tester", "valid-token")
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
	recorder := &fakeRecorder{}
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
		Recorder: recorder,
		Auth:     fakeAuth{configured: true},
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
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
	if len(recorder.updates) != 1 {
		t.Fatalf("expected one recorded update, got %#v", recorder.updates)
	}
	if recorder.repoID != "repo_123" || recorder.pusherID != "user_123" {
		t.Fatalf("unexpected recorder context: repo=%q pusher=%q", recorder.repoID, recorder.pusherID)
	}
	if recorder.updates[0].RefName != "refs/heads/main" || recorder.updates[0].OldSHA != strings.Repeat("0", 40) || recorder.updates[0].NewSHA == strings.Repeat("0", 40) {
		t.Fatalf("unexpected recorded update: %#v", recorder.updates[0])
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

func (a fakeAuth) Authenticate(_ context.Context, username, token string) (Principal, error) {
	if username != "tester" || token != "valid-token" {
		return Principal{}, errAuthRequired
	}
	return Principal{UserID: "user_123", Username: "tester"}, nil
}

type fakeRecorder struct {
	repoID   string
	pusherID string
	updates  []RefUpdate
}

func (r *fakeRecorder) RecordPush(_ context.Context, repo Repository, pusher Principal, updates []RefUpdate) error {
	r.repoID = repo.ID
	r.pusherID = pusher.UserID
	r.updates = append([]RefUpdate(nil), updates...)
	return nil
}
