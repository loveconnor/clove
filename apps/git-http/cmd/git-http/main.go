package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"clove/apps/git-http/internal/config"
	"clove/apps/git-http/internal/gitservice"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	loadDotEnv()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	log := logger(cfg)
	database, err := openDatabase(ctx, cfg, log)
	if err != nil {
		log.Error("failed to open database", slog.Any("error", err))
		os.Exit(1)
	}
	if database != nil {
		defer database.Close()
	}

	server := &http.Server{
		Addr:              cfg.HTTPAddr(),
		Handler:           gitservice.New(gitservice.Dependencies{Config: cfg, Store: gitservice.DBStore{DB: database}, Auth: gitservice.DBAuth{DB: database}, Logger: log}),
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	errs := make(chan error, 1)
	go func() {
		log.Info("git http server listening", slog.String("addr", server.Addr), slog.String("env", cfg.Environment))
		errs <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error("git http server shutdown failed", slog.Any("error", err))
			os.Exit(1)
		}
		log.Info("git http server stopped")
	case err := <-errs:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("git http server failed", slog.Any("error", err))
			os.Exit(1)
		}
	}

	time.Sleep(50 * time.Millisecond)
}

func openDatabase(ctx context.Context, cfg config.Config, log *slog.Logger) (*sql.DB, error) {
	if cfg.DatabaseURL == "" {
		log.Warn("DATABASE_URL is not set; repository lookup is disabled")
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

func logger(cfg config.Config) *slog.Logger {
	level := slog.LevelInfo
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}

func loadDotEnv() {
	_ = godotenv.Load(".env", "apps/git-http/.env", "apps/api/.env")
}
