package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env         string
	GRPCPort    string
	GatewayPort string // Port for the HTTP JSON Gateway
	Database    DatabaseConfig
	JWT         JWTConfig
	SMTP        SMTPConfig
}

type DatabaseConfig struct {
	Driver   string // "postgres" or "sqlite"
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	Secret                  string
	AccessExpiration        time.Duration
	RefreshExpiration       time.Duration
	ResetPasswordExpiration time.Duration
	VerifyEmailExpiration   time.Duration
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// LoadConfig loads environment variables
func LoadConfig() *Config {
	// Attempt to load .env, ignore if not found (e.g. Docker)
	_ = godotenv.Load()

	return &Config{
		Env:         getEnv("GO_ENV", "development"),
		GRPCPort:    getEnv("GRPC_PORT", "50051"),
		GatewayPort: getEnv("GATEWAY_PORT", "8080"),
		Database: DatabaseConfig{
			Driver:   getEnv("DB_DRIVER", "sqlite"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "starter_kit_grpc_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:                  getEnv("JWT_SECRET", "super_secret_key_change_me"),
			AccessExpiration:        time.Duration(getEnvAsInt("JWT_ACCESS_EXPIRATION_MINUTES", 30)) * time.Minute,
			RefreshExpiration:       time.Duration(getEnvAsInt("JWT_REFRESH_EXPIRATION_DAYS", 30)) * 24 * time.Hour,
			ResetPasswordExpiration: time.Duration(getEnvAsInt("JWT_RESET_PASSWORD_EXPIRATION_MINUTES", 15)) * time.Minute,
			VerifyEmailExpiration:   time.Duration(getEnvAsInt("JWT_VERIFY_EMAIL_EXPIRATION_MINUTES", 15)) * time.Minute,
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "smtp.example.com"),
			Port:     getEnvAsInt("SMTP_PORT", 587),
			Username: getEnv("SMTP_USERNAME", "user"),
			Password: getEnv("SMTP_PASSWORD", "pass"),
			From:     getEnv("EMAIL_FROM", "no-reply@example.com"),
		},
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return fallback
}