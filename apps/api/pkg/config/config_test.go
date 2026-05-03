package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	clearEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.AppName != "clove-api" {
		t.Fatalf("expected default app name, got %q", cfg.AppName)
	}
	if cfg.Port != 8080 {
		t.Fatalf("expected default port, got %d", cfg.Port)
	}
}

func TestLoadInvalidPort(t *testing.T) {
	clearEnv(t)
	t.Setenv("PORT", "not-a-port")

	if _, err := Load(); err == nil {
		t.Fatal("expected invalid port to return an error")
	}
}

func TestLoadInvalidDuration(t *testing.T) {
	clearEnv(t)
	t.Setenv("HTTP_READ_TIMEOUT", "slow")

	if _, err := Load(); err == nil {
		t.Fatal("expected invalid duration to return an error")
	}
}

func clearEnv(t *testing.T) {
	t.Helper()

	keys := []string{
		"APP_NAME",
		"APP_ENV",
		"HOST",
		"PORT",
		"LOG_LEVEL",
		"DATABASE_URL",
		"DATABASE_DRIVER",
		"DATABASE_PING_TIMEOUT",
		"WORKOS_API_KEY",
		"WORKOS_CLIENT_ID",
		"WORKOS_REDIRECT_URI",
		"WORKOS_BASE_URL",
		"WORKOS_ISSUER",
		"AUTH_SUCCESS_REDIRECT_URL",
		"AUTH_LOGOUT_REDIRECT_URL",
		"AUTH_ACCESS_COOKIE",
		"AUTH_REFRESH_COOKIE",
		"AUTH_STATE_COOKIE",
		"AUTH_COOKIE_DOMAIN",
		"AUTH_COOKIE_SECURE",
		"HTTP_READ_HEADER_TIMEOUT",
		"HTTP_READ_TIMEOUT",
		"HTTP_WRITE_TIMEOUT",
		"HTTP_IDLE_TIMEOUT",
		"HTTP_SHUTDOWN_TIMEOUT",
	}

	for _, key := range keys {
		t.Setenv(key, "")
	}
}
