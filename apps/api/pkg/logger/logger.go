package logger

import (
	"log/slog"
	"os"
	"strings"

	"clove/apps/api/pkg/config"
)

func New(cfg config.Config) *slog.Logger {
	opts := &slog.HandlerOptions{Level: parseLevel(cfg.LogLevel)}

	var handler slog.Handler
	if cfg.Environment == "development" || cfg.Environment == "local" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler).With(
		slog.String("app", cfg.AppName),
		slog.String("env", cfg.Environment),
	)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
