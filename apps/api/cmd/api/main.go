package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"clove/apps/api/internal/db"
	"clove/apps/api/internal/httpserver"
	"clove/apps/api/pkg/config"
	"clove/apps/api/pkg/logger"
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

	log := logger.New(cfg)

	database, err := db.Open(ctx, cfg, log)
	if err != nil {
		log.Error("failed to open database", slog.Any("error", err))
		os.Exit(1)
	}
	if database != nil {
		defer database.Close()

		if err := db.RunMigrations(ctx, database, log); err != nil {
			log.Error("failed to run migrations", slog.Any("error", err))
			os.Exit(1)
		}
	}

	server := &http.Server{
		Addr:              cfg.HTTPAddr(),
		Handler:           httpserver.New(httpserver.Dependencies{Config: cfg, DB: database, Logger: log}),
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	errs := make(chan error, 1)
	go func() {
		log.Info("api server listening", slog.String("addr", server.Addr), slog.String("env", cfg.Environment))
		errs <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error("api server shutdown failed", slog.Any("error", err))
			os.Exit(1)
		}

		log.Info("api server stopped")
	case err := <-errs:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("api server failed", slog.Any("error", err))
			os.Exit(1)
		}
	}

	time.Sleep(50 * time.Millisecond)
}

func loadDotEnv() {
	paths := []string{".env", "apps/api/.env"}
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			if err := godotenv.Load(path); err != nil {
				slog.Warn("failed to load env file", slog.String("path", path), slog.Any("error", err))
			}
			return
		}
	}
}
