package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration for ContentBank v2.
// Values are loaded from environment variables (with .env fallback).
type Config struct {
	AppPort string
	AppEnv  string

	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string

	JWTSecret string
	JWTTTL    time.Duration

	AIBaseURL string
	AIAPIKey  string
	AIModel   string

	AdminUsername string
	AdminPassword string
}

// LoadConfig reads .env (if present) and populates Config from environment
// variables. Missing required values are returned as an error.
func LoadConfig() (*Config, error) {
	// Best-effort .env load. It is fine if the file does not exist.
	_ = godotenv.Load()

	cfg := &Config{
		AppPort: getEnv("APP_PORT", "8080"),
		AppEnv:  getEnv("APP_ENV", "development"),

		DBHost:     getEnv("DB_HOST", "/home/aril/pg-data/socket"),
		DBPort:     getEnv("DB_PORT", "5433"),
		DBName:     getEnv("DB_NAME", "contentbank"),
		DBUser:     getEnv("DB_USER", "aril"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		AIBaseURL: getEnv("AI_BASE_URL", ""),
		AIAPIKey:  getEnv("AI_API_KEY", ""),
		AIModel:   getEnv("AI_MODEL", "gpt-4o-mini"),

		AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", ""),
	}

	secret := getEnv("JWT_SECRET", "")
	if secret == "" {
		// In dev, generate a deterministic-but-obvious default to avoid
		// silent insecure defaults in production.
		if cfg.AppEnv == "production" {
			return nil, fmt.Errorf("JWT_SECRET is required in production")
		}
		secret = "dev-only-insecure-secret-change-me"
	}
	cfg.JWTSecret = secret

	ttlHours, err := strconv.Atoi(getEnv("JWT_TTL_HOURS", "24"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_TTL_HOURS: %w", err)
	}
	cfg.JWTTTL = time.Duration(ttlHours) * time.Hour

	return cfg, nil
}

// DSN returns a PostgreSQL connection string suitable for lib/pq / GORM.
// Note: empty password must be omitted from DSN, otherwise some DSN parsers
// (lib/pq, pgx) will mis-tokenize and fall back to the user name as the
// database name.
func (c *Config) DSN() string {
	parts := []string{
		fmt.Sprintf("host=%s", c.DBHost),
		fmt.Sprintf("port=%s", c.DBPort),
		fmt.Sprintf("user=%s", c.DBUser),
		fmt.Sprintf("dbname=%s", c.DBName),
		fmt.Sprintf("sslmode=%s", c.DBSSLMode),
	}
	if c.DBPassword != "" {
		parts = append(parts, fmt.Sprintf("password=%s", c.DBPassword))
	}
	return strings.Join(parts, " ")
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}
