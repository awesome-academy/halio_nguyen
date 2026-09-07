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
	if cfg.JWT.AccessExpiryHours != 24*time.Hour {
		t.Errorf("expected 24h JWT access expiry, got %v", cfg.JWT.AccessExpiryHours)
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
