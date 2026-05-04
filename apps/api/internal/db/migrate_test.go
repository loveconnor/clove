package db

import (
	"strings"
	"testing"
)

func TestInitialMigrationContainsCoreTables(t *testing.T) {
	t.Parallel()

	sqlBytes, err := migrationsFS.ReadFile("migrations/000001_init.sql")
	if err != nil {
		t.Fatalf("read initial migration: %v", err)
	}

	migration := string(sqlBytes)
	tables := []string{
		"users",
		"sessions",
		"organizations",
		"organization_members",
		"repositories",
		"ssh_keys",
		"audit_logs",
	}

	for _, table := range tables {
		needle := "CREATE TABLE IF NOT EXISTS " + table
		if !strings.Contains(migration, needle) {
			t.Fatalf("expected initial migration to contain %q", needle)
		}
	}
}

func TestPersonalAccessTokenMigration(t *testing.T) {
	t.Parallel()

	sqlBytes, err := migrationsFS.ReadFile("migrations/000004_personal_access_tokens.sql")
	if err != nil {
		t.Fatalf("read personal access token migration: %v", err)
	}

	migration := string(sqlBytes)
	for _, needle := range []string{
		"CREATE TABLE IF NOT EXISTS personal_access_tokens",
		"user_id TEXT NOT NULL REFERENCES users(id)",
		"token_hash TEXT NOT NULL UNIQUE",
		"last_used_at TIMESTAMPTZ",
	} {
		if !strings.Contains(migration, needle) {
			t.Fatalf("expected migration to contain %q", needle)
		}
	}
}

func TestGitRefsPushEventsMigration(t *testing.T) {
	t.Parallel()

	sqlBytes, err := migrationsFS.ReadFile("migrations/000005_git_refs_push_events.sql")
	if err != nil {
		t.Fatalf("read git refs push events migration: %v", err)
	}

	migration := string(sqlBytes)
	for _, needle := range []string{
		"CREATE TABLE IF NOT EXISTS git_refs",
		"CREATE TABLE IF NOT EXISTS push_events",
		"repo_id TEXT NOT NULL REFERENCES repositories(id)",
		"ref_name TEXT NOT NULL",
		"old_sha TEXT NOT NULL",
		"new_sha TEXT NOT NULL",
		"pusher_id TEXT REFERENCES users(id)",
		"created_at TIMESTAMPTZ NOT NULL DEFAULT now()",
	} {
		if !strings.Contains(migration, needle) {
			t.Fatalf("expected migration to contain %q", needle)
		}
	}
}

func TestInitialMigrationUsesUUIDPrimaryKeys(t *testing.T) {
	t.Parallel()

	sqlBytes, err := migrationsFS.ReadFile("migrations/000001_init.sql")
	if err != nil {
		t.Fatalf("read initial migration: %v", err)
	}

	migration := string(sqlBytes)
	if strings.Contains(migration, "BIGSERIAL") {
		t.Fatal("expected initial migration to use UUID primary keys, found BIGSERIAL")
	}
	if !strings.Contains(migration, "id UUID PRIMARY KEY") {
		t.Fatal("expected initial migration to define UUID primary keys")
	}
}
