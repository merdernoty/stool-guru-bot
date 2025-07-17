package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken string
	WebhookURL    string
	Port          string
	Debug         bool
	Timeout       time.Duration
	GeminiAPIKey  string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	cfg := &Config{
		TelegramToken: getEnv("TELEGRAM_TOKEN", ""),
		WebhookURL:    getEnv("WEBHOOK_URL", ""),
		Port:          getEnv("PORT", "8080"),
		GeminiAPIKey:  getEnv("GEMINI_API_KEY", ""),
		Debug:         getEnvAsBool("DEBUG", false),
		Timeout:       time.Duration(getEnvAsInt("TIMEOUT_SECONDS", 60)) * time.Second,
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.TelegramToken == "" {
		return fmt.Errorf("TELEGRAM_TOKEN is required")
	}

	if !c.Debug && c.WebhookURL == "" {
		return fmt.Errorf("WEBHOOK_URL is required in production mode (DEBUG=false)")
	}

	return nil
}

func (c *Config) String() string {
	tokenDisplay := "not_set"
	if c.TelegramToken != "" {
		tokenDisplay = "set"
	}

	return fmt.Sprintf("Config{Port: %s, Debug: %t, WebhookURL: %s, Token: %s, Timeout: %v}",
		c.Port, c.Debug, c.WebhookURL, tokenDisplay, c.Timeout)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
