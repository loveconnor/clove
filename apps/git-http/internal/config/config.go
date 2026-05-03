package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppName           string
	Environment       string
	Host              string
	Port              int
	LogLevel          string
	DatabaseURL       string
	DatabaseDriver    string
	DatabasePingTime  time.Duration
	RepositoryRoot    string
	GitBin            string
	WorkOSClientID    string
	WorkOSBaseURL     string
	WorkOSIssuer      string
	AccessCookieName  string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
}

func Load() (Config, error) {
	port, err := envInt("GIT_HTTP_PORT", 8081)
	if err != nil {
		return Config{}, err
	}
	databasePingTime, err := envDuration("DATABASE_PING_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}
	readHeaderTimeout, err := envDuration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}
	readTimeout, err := envDuration("HTTP_READ_TIMEOUT", 0)
	if err != nil {
		return Config{}, err
	}
	writeTimeout, err := envDuration("HTTP_WRITE_TIMEOUT", 0)
	if err != nil {
		return Config{}, err
	}
	idleTimeout, err := envDuration("HTTP_IDLE_TIMEOUT", 60*time.Second)
	if err != nil {
		return Config{}, err
	}
	shutdownTimeout, err := envDuration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		AppName:           envString("GIT_HTTP_APP_NAME", "clove-git-http"),
		Environment:       envString("APP_ENV", "development"),
		Host:              envString("GIT_HTTP_HOST", envString("HOST", "0.0.0.0")),
		Port:              port,
		LogLevel:          envString("LOG_LEVEL", "info"),
		DatabaseURL:       strings.TrimSpace(os.Getenv("DATABASE_URL")),
		DatabaseDriver:    envString("DATABASE_DRIVER", "pgx"),
		DatabasePingTime:  databasePingTime,
		RepositoryRoot:    envString("REPOSITORY_ROOT", "/data/repos"),
		GitBin:            envString("GIT_BIN", "git"),
		WorkOSClientID:    strings.TrimSpace(os.Getenv("WORKOS_CLIENT_ID")),
		WorkOSBaseURL:     envString("WORKOS_BASE_URL", "https://api.workos.com"),
		WorkOSIssuer:      envString("WORKOS_ISSUER", "https://api.workos.com/"),
		AccessCookieName:  envString("AUTH_ACCESS_COOKIE", "clove_access_token"),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ShutdownTimeout:   shutdownTimeout,
	}

	if cfg.Port <= 0 || cfg.Port > 65535 {
		return Config{}, fmt.Errorf("GIT_HTTP_PORT must be between 1 and 65535, got %d", cfg.Port)
	}
	if strings.TrimSpace(cfg.AppName) == "" {
		return Config{}, errors.New("GIT_HTTP_APP_NAME must not be empty")
	}
	if strings.TrimSpace(cfg.DatabaseDriver) == "" {
		return Config{}, errors.New("DATABASE_DRIVER must not be empty")
	}
	if strings.TrimSpace(cfg.GitBin) == "" {
		return Config{}, errors.New("GIT_BIN must not be empty")
	}

	return cfg, nil
}

func (c Config) HTTPAddr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

func envString(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return parsed, nil
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration like 5s or 1m: %w", key, err)
	}
	return parsed, nil
}
