package config

import (
	"flag"
	"os"
	"sync"
)

// Config holds all configuration parameters for the application
type Config struct {
	// Server configuration
	Port            string
	Host            string
	Environment     string
	ShutdownTimeout string

	// Telegram bot configuration
	TelegramBotToken string
	APIEndpoint      string
}

var (
	instance *Config
	once     sync.Once
	parsed   bool
)

// NewConfig loads configuration from environment variables or flags
func NewConfig() *Config {
	once.Do(func() {
		instance = &Config{
			Port:             configValue("PORT", "port", "8080", "HTTP server port"),
			Host:             configValue("HOST", "host", "", "Server host"),
			Environment:      configValue("ENV", "env", "development", "Application environment (development/production)"),
			ShutdownTimeout:  configValue("SHUTDOWN_TIMEOUT", "shutdown-timeout", "30", "Graceful shutdown timeout in seconds"),
			TelegramBotToken: configValue("TELEGRAM_BOT_TOKEN", "telegram-token", "", "Telegram bot token"),
			APIEndpoint:      configValue("API_ENDPOINT", "api-endpoint", "", "Custom API endpoint"),
		}
	})

	return instance
}

// configValue returns the value of a parameter based on the following priority:
// 1. Environment variable.
// 2. Command-line flag.
// 3. Default value.
func configValue(envVar, flagName, defaultValue, description string) string {
	envValue := os.Getenv(envVar)
	if envValue != "" {
		return envValue
	}

	// Create command-line flag only once
	if !parsed {
		flag.String(flagName, defaultValue, description)
		parsed = true
		flag.Parse()
	}

	// Get the flag value
	flagValue := flag.Lookup(flagName)
	if flagValue != nil {
		return flagValue.Value.String()
	}

	return defaultValue
}
