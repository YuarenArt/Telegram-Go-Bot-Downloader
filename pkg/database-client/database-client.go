package database_client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/YuarenArt/tg-users-database/pkg/db"
	"net/http"
)

const (
	timeout = 20 * time.Second
	baseURL = "https://localhost:8082" // https://localhost:8082 or https://tg-database:8082]
)

// Client is a structure that contains the HTTP client and the base URL of the server.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient initializes and returns a new Client.
func NewClient() *Client {

	certFile := "cert.pem"
	cert, err := os.ReadFile(certFile)
	if err != nil {
		log.Printf("Failed to read certificate file: %v", err)
	}

	// Create a certificate pool from the self-signed certificate
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(cert) {
		log.Println("failed to append certificate")
	}

	tlsConfig := &tls.Config{
		RootCAs: certPool,
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: timeout,
	}

	return &Client{
		httpClient: client,
		baseURL:    baseURL,
	}
}

// NewUser initializes and returns a new User.
func NewUser(username string, ChatID int64) *db.User {
	return &db.User{
		Username: username,
		Traffic:  0,
		ChatID:   ChatID,
	}
}

// CreateUser sends a request to create a new user.
func (c *Client) CreateUser(ctx context.Context, newUser *db.User) error {
	url := fmt.Sprintf("%s/users", c.baseURL)
	body, err := json.Marshal(newUser)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create user: status %d", resp.StatusCode)
	}

	return nil
}

// GetUser sends a request to retrieve a user by username.
func (c *Client) GetUser(ctx context.Context, username string) (*db.User, error) {
	url := fmt.Sprintf("%s/users/%s", c.baseURL, username)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var retrievedUser db.User
	if err := json.NewDecoder(resp.Body).Decode(&retrievedUser); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return &retrievedUser, nil
}

// GetSubscriptionStatus sends a request to retrieve a user's subscription status.
func (c *Client) GetSubscriptionStatus(ctx context.Context, username string) (string, error) {
	url := fmt.Sprintf("%s/users/%s/subscription", c.baseURL, username)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode the response body into a user struct
	var subscriptionStatus string
	if err := json.NewDecoder(resp.Body).Decode(&subscriptionStatus); err != nil {
		return "", fmt.Errorf("failed to decode response body: %w", err)
	}

	return subscriptionStatus, nil
}

// IsUserExist sends a request to check if a user exists.
func (c *Client) IsUserExist(ctx context.Context, username string) (bool, error) {
	url := fmt.Sprintf("%s/users/%s/exists", c.baseURL, username)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
}
