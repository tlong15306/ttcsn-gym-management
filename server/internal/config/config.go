package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv      string
	HTTPAddr    string
	DatabaseURL string
	JWTSecret   string
	JWTTTL      time.Duration
	BcryptCost  int
	CORSOrigins []string
	UploadDir   string
}

func Load() (Config, error) {
	if err := loadLocalEnvironment(); err != nil {
		return Config{}, err
	}

	ttl, err := time.ParseDuration(getEnv("JWT_TTL", "15m"))
	if err != nil || ttl <= 0 {
		return Config{}, fmt.Errorf("JWT_TTL must be a positive duration")
	}

	cost, err := strconv.Atoi(getEnv("BCRYPT_COST", "10"))
	if err != nil || cost < 10 || cost > 31 {
		return Config{}, fmt.Errorf("BCRYPT_COST must be between 10 and 31")
	}

	cfg := Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		HTTPAddr:    getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		JWTTTL:      ttl,
		BcryptCost:  cost,
		CORSOrigins: splitCSV(getEnv("CORS_ORIGINS", "http://127.0.0.1:5173,http://localhost:5173")),
		UploadDir:   getEnv("UPLOAD_DIR", "../uploads/private"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must contain at least 32 characters")
	}
	if len(cfg.CORSOrigins) == 0 {
		return Config{}, fmt.Errorf("CORS_ORIGINS must contain at least one origin")
	}

	return cfg, nil
}

// loadLocalEnvironment applies .env.local over .env while preserving any
// variables supplied by the host process. This keeps personal credentials out
// of version control and lets a developer retain an unrelated legacy .env.
func loadLocalEnvironment() error {
	values := make(map[string]string)
	for _, path := range []string{".env", ".env.local"} {
		fileValues, err := godotenv.Read(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("read %s: %w", path, err)
		}
		for key, value := range fileValues {
			values[key] = value
		}
	}
	for key, value := range values {
		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("set %s from local config: %w", key, err)
			}
		}
	}
	return nil
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
