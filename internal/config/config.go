package config

import (
	"flag"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
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
	CookiesPath      string `yaml:"cookies_path" env:"COOKIES_PATH"`

	// Database configuration
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	DBSSLMode     string
	DatabaseToken string `yaml:"database_token" env:"DATABASE_TOKEN"`
	MaxOpenConns  int
	MaxIdleConns  int
}

var (
	instance *Config
	once     sync.Once
	parsed   bool
)

// NewConfig loads configuration from .env file, environment variables, or flags
func NewConfig() *Config {
	once.Do(func() {
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found, using environment variables or defaults")
		}

		// Parse command line flags first
		if !parsed {
			// Server flags
			flag.String("port", "8080", "HTTP server port")
			flag.String("host", "", "Server host")
			flag.String("env", "development", "Application environment (development/production)")
			flag.String("shutdown-timeout", "30", "Graceful shutdown timeout in seconds")

			// Database flags
			flag.String("db-host", "localhost", "Database host")
			flag.String("db-port", "5432", "Database port")
			flag.String("db-user", "postgres", "Database user")
			flag.String("db-password", "yourpassword", "Database password")
			flag.String("db-name", "users", "Database name")
			flag.String("db-sslmode", "disable", "Database SSL mode")
			flag.Int("db-max-open-conns", 25, "Maximum number of open connections to the database")
			flag.Int("db-max-idle-conns", 5, "Maximum number of idle connections in the connection pool")

			flag.Parse()
			parsed = true
		}

		instance = &Config{
			// Server configuration
			Port:            configValue("PORT", "port", "8080", "HTTP server port"),
			Host:            configValue("HOST", "host", "", "Server host"),
			Environment:     configValue("ENV", "env", "development", "Application environment (development/production)"),
			ShutdownTimeout: configValue("SHUTDOWN_TIMEOUT", "shutdown-timeout", "30", "Graceful shutdown timeout in seconds"),

			// Telegram bot configuration
			TelegramBotToken: configValue("TELEGRAM_BOT_TOKEN", "telegram-bot-token", "", "Telegram Bot API token"),
			APIEndpoint:      configValue("TELEGRAM_API_ENDPOINT", "api-endpoint", "", "Custom Telegram API endpoint (optional)"),
			CookiesPath:      configValue("COOKIES_PATH", "cookies-path", "cookies.txt", "Path to cookies.txt file for yt-dlp (optional)"),

			// Database configuration
			DBHost:        configValue("DB_HOST", "db-host", "localhost", "Database host"),
			DBPort:        configValue("DB_PORT", "db-port", "5432", "Database port"),
			DBUser:        configValue("DB_USER", "db-user", "postgres", "Database user"),
			DBPassword:    configValue("DB_PASSWORD", "db-password", "postgres", "Database password"),
			DBName:        configValue("DB_NAME", "db-name", "youtube_downloader", "Database name"),
			DBSSLMode:     configValue("DB_SSLMODE", "db-sslmode", "disable", "Database SSL mode"),
			DatabaseToken: configValue("DATABASE_TOKEN", "database-token", "", "Database access token"),
			MaxOpenConns:  getIntConfigValue("DB_MAX_OPEN_CONNS", "db-max-open-conns", 25, "Maximum number of open connections to the database"),
			MaxIdleConns:  getIntConfigValue("DB_MAX_IDLE_CONNS", "db-max-idle-conns", 25, "Maximum number of idle connections in the connection pool"),
		}
	})

	return instance
}

// configValue returns the value of a parameter based on the following priority:
// 1. Environment variable.
// 2. Command-line flag.
// 3. Default value.
func configValue(envVar, flagName, defaultValue, description string) string {
	// Check environment variable first
	if envValue := os.Getenv(envVar); envValue != "" {
		return envValue
	}

	// Then check command-line flag
	if f := flag.Lookup(flagName); f != nil {
		return f.Value.String()
	}

	// Return default value if neither is set
	return defaultValue
}

// getIntConfigValue returns an integer configuration value using the same priority as configValue
func getIntConfigValue(envVar, flagName string, defaultValue int, description string) int {
	// First check environment variables
	if val, exists := os.LookupEnv(envVar); exists {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}

	// Then check command line flags
	if !parsed {
		flag.Parse()
		parsed = true
	}

	// If flag was provided, return its value
	if flag.Lookup(flagName) != nil && flag.Lookup(flagName).Value.String() != "" {
		return flag.Lookup(flagName).Value.(flag.Getter).Get().(int)
	}

	// Otherwise return the default value
	return defaultValue
}
