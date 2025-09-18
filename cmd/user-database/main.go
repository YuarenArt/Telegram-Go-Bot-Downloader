package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"youtube_downloader/internal/api/handler"
	"youtube_downloader/internal/config"
	"youtube_downloader/pkg/database"
	dbrepo "youtube_downloader/pkg/database/repository"
)

// @title user Database API
// @version 2.3
// @description This is a server for managing user subscriptions over HTTPS.
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @host localhost:8082
// @BasePath /
// @schemes https

func main() {
	// Root context with signal handling for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Load configuration
	cfg := config.NewConfig()

	// Validate database port is a valid integer
	if _, err := strconv.Atoi(cfg.DBPort); err != nil {
		log.Fatalf("invalid database port: %v", err)
	}

	// Create database configuration from the loaded config
	dbCfg := &database.Config{
		Host:         cfg.DBHost,
		Port:         cfg.DBPort,
		User:         cfg.DBUser,
		Password:     cfg.DBPassword,
		DBName:       cfg.DBName,
		SSLMode:      cfg.DBSSLMode,
		MaxOpenConns: cfg.MaxOpenConns,
		MaxIdleConns: cfg.MaxIdleConns,
	}

	// Initialize DB with retries (helps when DB container starts slowly).
	db, err := dbrepo.NewDatabase(ctx, dbCfg)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	// Ensure DB is closed on exit.
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("error closing database: %v", err)
		}
	}()

	// TLS certs (if absent, we start plain HTTP but log it explicitly).
	certFile := "cert.pem"
	keyFile := "key.pem"
	useTLS := fileExists(certFile) && fileExists(keyFile)
	if !useTLS {
		log.Println("TLS certificates not found; starting HTTP (non-TLS). In production provide cert.pem and key.pem.")
	}

	// HTTP server configuration.
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler.NewHandler(db).Router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Server lifecycle: listen and graceful shutdown.
	serverCtx, serverStop := context.WithCancel(context.Background())

	// Shutdown goroutine on OS signal.
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		log.Println("Shutting down server gracefully...")
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
		serverStop()
	}()

	// Start server.
	go func() {
		log.Printf("Starting server on %s (TLS=%v)", srv.Addr, useTLS)
		var err error
		if useTLS {
			err = srv.ListenAndServeTLS(certFile, keyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-serverCtx.Done()
	log.Println("Server stopped")
}

// fileExists checks file presence.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
