package config

import (
	"os"
	"testing"
	"time"
)

func TestConfigDefaults(t *testing.T) {
	// Clear relevant environment variables to test defaults
	vars := []string{
		"APP_NAME", "APP_ENV", "APP_PORT",
		"POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_USER",
		"POSTGRES_PASSWORD", "POSTGRES_DB", "POSTGRES_SSLMODE",
		"DATABASE_URL", "JWT_ACCESS_EXPIRY_HOURS", "JWT_REFRESH_EXPIRY_DAYS",
		"COOKIE_SECURE", "LOGIN_RATE_LIMIT_PER_MINUTE", "LOGIN_RATE_LIMIT_BURST",
		"TRUSTED_PROXY_CIDRS",
	}
	for _, v := range vars {
		os.Unsetenv(v)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error loading config, got: %v", err)
	}

	if cfg.App.Name != "sun-booking-tours-api" {
		t.Errorf("expected default App.Name 'sun-booking-tours-api', got '%s'", cfg.App.Name)
	}
	if cfg.App.Port != "8080" {
		t.Errorf("expected default App.Port '8080', got '%s'", cfg.App.Port)
	}
	if cfg.Database.Host != "localhost" {
		t.Errorf("expected default DB Host 'localhost', got '%s'", cfg.Database.Host)
	}
	if cfg.Database.Port != "5434" {
		t.Errorf("expected default DB Port '5434', got '%s'", cfg.Database.Port)
	}
	// D2: admin-portal override — 1h, not the spec's original 24h default.
	if cfg.JWT.AccessExpiryHours != 1*time.Hour {
		t.Errorf("expected 1h JWT access expiry (decisions.md D2), got %v", cfg.JWT.AccessExpiryHours)
	}
}

func TestSecurityConfigDefaults(t *testing.T) {
	vars := []string{"APP_ENV", "COOKIE_SECURE", "LOGIN_RATE_LIMIT_PER_MINUTE", "LOGIN_RATE_LIMIT_BURST", "TRUSTED_PROXY_CIDRS"}
	for _, v := range vars {
		os.Unsetenv(v)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error loading config, got: %v", err)
	}

	// APP_ENV unset -> defaults to "development" -> cookies not forced secure.
	if cfg.Security.CookieSecure != false {
		t.Errorf("expected CookieSecure=false in development, got %v", cfg.Security.CookieSecure)
	}
	if cfg.Security.LoginRateLimitPerMinute != 5 {
		t.Errorf("expected LoginRateLimitPerMinute=5, got %d", cfg.Security.LoginRateLimitPerMinute)
	}
	if cfg.Security.LoginRateLimitBurst != 5 {
		t.Errorf("expected LoginRateLimitBurst=5, got %d", cfg.Security.LoginRateLimitBurst)
	}
	wantCIDRs := []string{"127.0.0.1/32", "::1/128"}
	if len(cfg.Security.TrustedProxyCIDRs) != len(wantCIDRs) {
		t.Fatalf("expected %d trusted proxy CIDRs, got %d (%v)", len(wantCIDRs), len(cfg.Security.TrustedProxyCIDRs), cfg.Security.TrustedProxyCIDRs)
	}
	for i, want := range wantCIDRs {
		if cfg.Security.TrustedProxyCIDRs[i] != want {
			t.Errorf("TrustedProxyCIDRs[%d] = %q, want %q", i, cfg.Security.TrustedProxyCIDRs[i], want)
		}
	}
}

func TestCookieSecureFailsSafeOutsideDevelopment(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	os.Unsetenv("COOKIE_SECURE")
	defer os.Unsetenv("APP_ENV")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error loading config, got: %v", err)
	}
	if !cfg.Security.CookieSecure {
		t.Error("expected CookieSecure=true when APP_ENV != development and COOKIE_SECURE unset")
	}
}

func TestCookieSecureExplicitOverride(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	os.Setenv("COOKIE_SECURE", "false")
	defer os.Unsetenv("APP_ENV")
	defer os.Unsetenv("COOKIE_SECURE")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error loading config, got: %v", err)
	}
	if cfg.Security.CookieSecure {
		t.Error("expected explicit COOKIE_SECURE=false to override the production default")
	}
}

func TestDBConfigDSN(t *testing.T) {
	tests := []struct {
		name     string
		dbConfig DBConfig
		expected string
	}{
		{
			name: "custom connection parameters",
			dbConfig: DBConfig{
				Host:     "db.prod.internal",
				Port:     "5432",
				User:     "prod_user",
				Password: "secret_password",
				DBName:   "sun_tours",
				SSLMode:  "require",
			},
			expected: "postgres://prod_user:secret_password@db.prod.internal:5432/sun_tours?sslmode=require",
		},
		{
			name: "explicit DATABASE_URL takes priority",
			dbConfig: DBConfig{
				URL:  "postgres://override_user:pass@remote:5432/override_db?sslmode=verify-full",
				Host: "ignored_host",
			},
			expected: "postgres://override_user:pass@remote:5432/override_db?sslmode=verify-full",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := tt.dbConfig.DSN()
			if dsn != tt.expected {
				t.Errorf("expected DSN '%s', got '%s'", tt.expected, dsn)
			}
		})
	}
}
