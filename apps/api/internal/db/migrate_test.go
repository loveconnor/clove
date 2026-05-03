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
