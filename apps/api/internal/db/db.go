package db

import (
	"context"
	"database/sql"
	"log/slog"

	"clove/apps/api/pkg/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(ctx context.Context, cfg config.Config, log *slog.Logger) (*sql.DB, error) {
	if cfg.DatabaseURL == "" {
		log.Warn("DATABASE_URL is not set; database connection is disabled")
		return nil, nil
	}

	database, err := sql.Open(cfg.DatabaseDriver, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, cfg.DatabasePingTime)
	defer cancel()

	if err := database.PingContext(pingCtx); err != nil {
		_ = database.Close()
		return nil, err
	}

	log.Info("database connection established", slog.String("driver", cfg.DatabaseDriver))
	return database, nil
}
