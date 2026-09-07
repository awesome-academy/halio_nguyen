package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the API server.
type Config struct {
	App      AppConfig
	Database DBConfig
	JWT      JWTConfig
	Security SecurityConfig
}

// AppConfig contains general application settings.
type AppConfig struct {
	Name            string
	Env             string
	Port            string
	BaseURL         string
	FrontendURL     string
	AllowedOrigins  string
}

// DBConfig holds PostgreSQL connection parameters.
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	URL      string
}

// DSN returns the PostgreSQL connection string.
func (c DBConfig) DSN() string {
	if c.URL != "" {
		return c.URL
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode,
	)
}

// JWTConfig holds JWT authentication configuration.
type JWTConfig struct {
	Secret            string
	AccessExpiryHours time.Duration
	RefreshExpiryDays time.Duration
}

// SecurityConfig holds encryption keys and provider secrets.
type SecurityConfig struct {
	BankEncryptionKey string
}

// Load reads configuration from environment variables and optional .env file.
func Load() (*Config, error) {
	// Try loading from .env if present, ignore if missing
	_ = godotenv.Load()

	accessHours := getEnvInt("JWT_ACCESS_EXPIRY_HOURS", 24)
	refreshDays := getEnvInt("JWT_REFRESH_EXPIRY_DAYS", 7)

	cfg := &Config{
		App: AppConfig{
			Name:           getEnv("APP_NAME", "sun-booking-tours-api"),
			Env:            getEnv("APP_ENV", "development"),
			Port:           getEnv("APP_PORT", "8080"),
			BaseURL:        getEnv("API_BASE_URL", "http://localhost:8080"),
			FrontendURL:    getEnv("NEXT_PUBLIC_SITE_URL", "http://localhost:3000"),
			AllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
		},
		Database: DBConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnv("POSTGRES_PORT", "5434"),
			User:     getEnv("POSTGRES_USER", "sun_booking"),
			Password: getEnv("POSTGRES_PASSWORD", "sun_booking_secret"),
			DBName:   getEnv("POSTGRES_DB", "sun_booking_tours"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
			URL:      getEnv("DATABASE_URL", ""),
		},
		JWT: JWTConfig{
			Secret:            getEnv("JWT_SECRET", "super-secret-jwt-key-min-32-chars-default"),
			AccessExpiryHours: time.Duration(accessHours) * time.Hour,
			RefreshExpiryDays: time.Duration(refreshDays) * 24 * time.Hour,
		},
		Security: SecurityConfig{
			BankEncryptionKey: getEnv("BANK_ENCRYPTION_KEY", "my-super-secret-encryption-key32b"),
		},
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return fallback
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return fallback
	}
	return val
}
