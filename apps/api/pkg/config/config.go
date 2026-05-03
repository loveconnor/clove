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
	WorkOSAPIKey      string
	WorkOSClientID    string
	WorkOSRedirectURI string
	WorkOSBaseURL     string
	WorkOSIssuer      string
	AuthSuccessURL    string
	AuthLogoutURL     string
	AccessCookieName  string
	RefreshCookieName string
	StateCookieName   string
	CookieDomain      string
	CookieSecure      bool
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
}

func Load() (Config, error) {
	port, err := envInt("PORT", 8080)
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
	readTimeout, err := envDuration("HTTP_READ_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	writeTimeout, err := envDuration("HTTP_WRITE_TIMEOUT", 15*time.Second)
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
	cookieSecure, err := envBool("AUTH_COOKIE_SECURE", envString("APP_ENV", "development") != "development")
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		AppName:           envString("APP_NAME", "clove-api"),
		Environment:       envString("APP_ENV", "development"),
		Host:              envString("HOST", "0.0.0.0"),
		Port:              port,
		LogLevel:          envString("LOG_LEVEL", "info"),
		DatabaseURL:       strings.TrimSpace(os.Getenv("DATABASE_URL")),
		DatabaseDriver:    envString("DATABASE_DRIVER", "pgx"),
		DatabasePingTime:  databasePingTime,
		WorkOSAPIKey:      strings.TrimSpace(os.Getenv("WORKOS_API_KEY")),
		WorkOSClientID:    strings.TrimSpace(os.Getenv("WORKOS_CLIENT_ID")),
		WorkOSRedirectURI: envString("WORKOS_REDIRECT_URI", "http://localhost:8080/api/auth/callback"),
		WorkOSBaseURL:     envString("WORKOS_BASE_URL", ""),
		WorkOSIssuer:      envString("WORKOS_ISSUER", "https://api.workos.com/"),
		AuthSuccessURL:    envString("AUTH_SUCCESS_REDIRECT_URL", "/"),
		AuthLogoutURL:     envString("AUTH_LOGOUT_REDIRECT_URL", "/"),
		AccessCookieName:  envString("AUTH_ACCESS_COOKIE", "clove_access_token"),
		RefreshCookieName: envString("AUTH_REFRESH_COOKIE", "clove_refresh_token"),
		StateCookieName:   envString("AUTH_STATE_COOKIE", "clove_auth_state"),
		CookieDomain:      strings.TrimSpace(os.Getenv("AUTH_COOKIE_DOMAIN")),
		CookieSecure:      cookieSecure,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ShutdownTimeout:   shutdownTimeout,
	}

	if cfg.Port <= 0 || cfg.Port > 65535 {
		return Config{}, fmt.Errorf("PORT must be between 1 and 65535, got %d", cfg.Port)
	}
	if strings.TrimSpace(cfg.AppName) == "" {
		return Config{}, errors.New("APP_NAME must not be empty")
	}
	if strings.TrimSpace(cfg.DatabaseDriver) == "" {
		return Config{}, errors.New("DATABASE_DRIVER must not be empty")
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

func envBool(key string, fallback bool) (bool, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean: %w", key, err)
	}
	return parsed, nil
}
