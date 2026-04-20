package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type AppConfig struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Auth        AuthConfig
	TrustedLLM  LLMConfig
	CloudLLM    LLMConfig
	Masking     MaskingConfig
	ProjectRoot string
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Type string
	DSN  string
}

type AuthConfig struct {
	InitialToken  string
	SessionCookie string
	SessionTTL    time.Duration
	SecureCookie  bool
}

type LLMConfig struct {
	Provider     string
	Model        string
	BaseURL      string
	APIKey       string
	SystemPrompt string
}

type MaskingConfig struct {
	MaxRetryTimes int
	RecentContext int
}

func Load() AppConfig {
	return AppConfig{
		Server: ServerConfig{
			Port: getEnv("APP_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Type: getEnv("DB_TYPE", "sqlite"),
			DSN:  getEnv("DB_DSN", "llm_guard.db"),
		},
		Auth: AuthConfig{
			InitialToken:  os.Getenv("APP_AUTHTOKEN"),
			SessionCookie: "llm_guard_session",
			SessionTTL:    12 * time.Hour,
			SecureCookie:  false,
		},
		Masking: MaskingConfig{
			MaxRetryTimes: 3,
			RecentContext: 4,
		},
		ProjectRoot: resolveProjectRoot(getEnv("APP_PROJECT_ROOT", "./llm-privacy-gaurd")),
	}
}

func (c AppConfig) Address() string {
	return fmt.Sprintf(":%s", c.Server.Port)
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func resolveProjectRoot(value string) string {
	if filepath.IsAbs(value) {
		return filepath.Clean(value)
	}
	abs, err := filepath.Abs(value)
	if err != nil {
		return value
	}
	return abs
}
