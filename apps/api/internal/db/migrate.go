package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"path"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(ctx context.Context, database *sql.DB, log *slog.Logger) error {
	if database == nil {
		log.Warn("database is not configured; skipping migrations")
		return nil
	}

	if _, err := database.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		files = append(files, entry.Name())
	}
	sort.Strings(files)

	for _, file := range files {
		applied, err := migrationApplied(ctx, database, file)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		if err := applyMigration(ctx, database, file); err != nil {
			return err
		}
		log.Info("migration applied", slog.String("version", file))
	}

	return nil
}

func migrationApplied(ctx context.Context, database *sql.DB, version string) (bool, error) {
	var exists bool
	if err := database.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists); err != nil {
		return false, fmt.Errorf("check migration %s: %w", version, err)
	}
	return exists, nil
}

func applyMigration(ctx context.Context, database *sql.DB, file string) error {
	sqlBytes, err := migrationsFS.ReadFile(path.Join("migrations", file))
	if err != nil {
		return fmt.Errorf("read migration %s: %w", file, err)
	}

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", file, err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, string(sqlBytes)); err != nil {
		return fmt.Errorf("execute migration %s: %w", file, err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", file); err != nil {
		return fmt.Errorf("record migration %s: %w", file, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", file, err)
	}

	return nil
}
