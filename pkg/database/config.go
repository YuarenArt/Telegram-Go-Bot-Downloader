package database

import (
	"fmt"
	"net/url"
	"os"
)

// Config holds the database configuration parameters
type Config struct {
	Host         string
	Port         string
	User         string
	Password     string
	DBName       string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	CookiesPath  string `yaml:"cookies_path" env:"COOKIES_PATH"`
}

// GetCookiesPath returns the path to the cookies.txt file
// If the path is not set or file doesn't exist, returns an empty string
func (c *Config) GetCookiesPath() string {
	if c == nil || c.CookiesPath == "" {
		return ""
	}
	// Check if the file exists
	if _, err := os.Stat(c.CookiesPath); err == nil {
		return c.CookiesPath
	}
	return ""
}

// DefaultConfig returns a default database configuration
func DefaultConfig() *Config {
	return &Config{
		Host:         "localhost",
		Port:         "5432",
		User:         "postgres",
		Password:     "postgres",
		DBName:       "users",
		SSLMode:      "disable",
		MaxOpenConns: 25,
		MaxIdleConns: 5,
		CookiesPath:  "cookies.txt",
	}
}

// DSN returns the Data Source Name for the database connection.
// Uses url.QueryEscape for password and database name safety.
func (c *Config) DSN() string {
	// lib/pq connection string format
	// host=... port=... user=... password=... dbname=... sslmode=...
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host,
		c.Port,
		c.User,
		url.QueryEscape(c.Password),
		url.QueryEscape(c.DBName),
		c.SSLMode,
	)
}
