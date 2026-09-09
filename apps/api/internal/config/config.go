package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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

// SecurityConfig holds encryption keys, provider secrets, and the
// admin-portal hardening knobs: cookie security, login-attempt throttling,
// and the proxy hops trusted to set X-Forwarded-For.
type SecurityConfig struct {
	BankEncryptionKey string

	// CookieSecure controls the Secure flag on the sun_admin_token cookie.
	// It defaults to true whenever APP_ENV != "development" so a missing env
	// var fails safe (secure) rather than open; COOKIE_SECURE overrides it
	// explicitly in either direction.
	CookieSecure bool

	// LoginRateLimitPerMinute and LoginRateLimitBurst bound the in-memory
	// per-IP/email login throttle (decisions.md D3).
	LoginRateLimitPerMinute int
	LoginRateLimitBurst     int

	// TrustedProxyCIDRs lists the proxy hops trusted to set
	// X-Forwarded-For; anything beyond these hops is not trusted for
	// per-IP rate limiting.
	TrustedProxyCIDRs []string
}

// Load reads configuration from environment variables and optional .env file.
func Load() (*Config, error) {
	// Try loading from .env if present, ignore if missing
	_ = godotenv.Load()

	// D2 (decisions.md): admin-portal JWT lifetime is 1h, not the
	// spec's original 24h default — there is no refresh-token table, so a
	// short expiry is the only way to bound a leaked token's usable life.
	accessHours := getEnvInt("JWT_ACCESS_EXPIRY_HOURS", 1)
	refreshDays := getEnvInt("JWT_REFRESH_EXPIRY_DAYS", 7)

	appEnv := getEnv("APP_ENV", "development")
	cookieSecureDefault := appEnv != "development"

	cfg := &Config{
		App: AppConfig{
			Name:           getEnv("APP_NAME", "sun-booking-tours-api"),
			Env:            appEnv,
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
			BankEncryptionKey:       getEnv("BANK_ENCRYPTION_KEY", "my-super-secret-encryption-key32b"),
			CookieSecure:            getEnvBool("COOKIE_SECURE", cookieSecureDefault),
			LoginRateLimitPerMinute: getEnvInt("LOGIN_RATE_LIMIT_PER_MINUTE", 5),
			LoginRateLimitBurst:     getEnvInt("LOGIN_RATE_LIMIT_BURST", 5),
			TrustedProxyCIDRs:       getEnvList("TRUSTED_PROXY_CIDRS", "127.0.0.1/32,::1/128"),
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

// getEnvBool reads a boolean env var, falling back to fallback when the var
// is unset, empty, or not a valid bool — so a malformed value fails safe to
// the caller-supplied default rather than silently becoming false.
func getEnvBool(key string, fallback bool) bool {
	valStr := os.Getenv(key)
	if valStr == "" {
		return fallback
	}
	val, err := strconv.ParseBool(valStr)
	if err != nil {
		return fallback
	}
	return val
}

// getEnvList reads a comma-separated env var into a trimmed, non-empty
// string slice, falling back to a comma-separated default when unset.
func getEnvList(key, fallback string) []string {
	raw := getEnv(key, fallback)
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
