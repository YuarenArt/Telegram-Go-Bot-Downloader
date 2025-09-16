package database

import (
	"fmt"
	"net/url"
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
