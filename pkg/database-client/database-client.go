package database_client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"youtube_downloader/internal/api/types"
	"youtube_downloader/pkg/database/models"
)

const (
	timeout = 20 * time.Second
)

var durations = [3]string{"month", "year", "forever"}

// Client is a structure that contains the HTTP client and the base URL of the server.
type Client struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

// NewClient initializes and returns a new Client.
func NewClient(token string) *Client {

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
		RootCAs:            certPool,
		InsecureSkipVerify: true, // Disable TLS verification
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: timeout,
	}

	err = godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading env.example file: %v", err)
	}
	baseURL := os.Getenv("DB_URL")

	return &Client{
		httpClient: client,
		baseURL:    baseURL,
		token:      token,
	}
}

// NewUser initializes and returns a new User.
func NewUser(username string, ChatID int64) *models.User {
	return &models.User{
		Username: username,
		Traffic:  0,
		ChatID:   ChatID,
		Subscription: models.Subscription{
			StartSubscription: time.Now(),
			EndSubscription:   time.Now(),
			Duration:          durations[0],
		},
	}
}

// CreateUser sends a request to create a new user.
func (c *Client) CreateUser(ctx context.Context, username string, chatID int64) (*types.UserResponse, error) {
	url := fmt.Sprintf("%s/users", c.baseURL)

	reqBody := types.UserRequest{
		Username: username,
		ChatID:   chatID,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusCreated:
		var userResp types.UserResponse
		if err := json.Unmarshal(respBody, &userResp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		return &userResp, nil

	case http.StatusBadRequest, http.StatusConflict:
		var errResp types.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			return nil, fmt.Errorf("invalid error response: %w", err)
		}
		return nil, errors.New(errResp.Error)

	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

// GetUser sends a request to retrieve a user by username.
func (c *Client) GetUser(ctx context.Context, username string) (*types.UserResponse, error) {
	url := fmt.Sprintf("%s/users/%s", c.baseURL, username)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var userResp types.UserResponse
		if err := json.Unmarshal(respBody, &userResp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		return &userResp, nil

	case http.StatusNotFound:
		return nil, fmt.Errorf("user not found")

	case http.StatusBadRequest:
		var errResp types.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			return nil, fmt.Errorf("invalid error response: %w", err)
		}
		return nil, errors.New(errResp.Error)

	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

// GetSubscriptionStatus sends a request to retrieve a user's subscription status.
func (c *Client) GetSubscriptionStatus(ctx context.Context, username string) (*types.SubscriptionResponse, error) {
	url := fmt.Sprintf("%s/users/%s/subscription", c.baseURL, username)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var subResp types.SubscriptionResponse
		if err := json.Unmarshal(respBody, &subResp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		return &subResp, nil

	case http.StatusNotFound:
		return nil, fmt.Errorf("user not found")

	case http.StatusBadRequest:
		var errResp types.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			return nil, fmt.Errorf("invalid error response: %w", err)
		}
		return nil, errors.New(errResp.Error)

	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (c *Client) IsUserExist(ctx context.Context, username string) (bool, error) {
	url := fmt.Sprintf("%s/users/%s/exists", c.baseURL, username)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil

	case http.StatusNotFound:
		return false, nil

	default:
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

// UpdateTraffic sends a request to update a user's traffic.
func (c *Client) UpdateTraffic(ctx context.Context, username string, traffic int64) error {
	url := fmt.Sprintf("%s/users/%s/traffic", c.baseURL, username)

	trafficUpdate := types.TrafficUpdateRequest{
		Traffic: traffic,
	}

	body, err := json.Marshal(trafficUpdate)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return nil

	case http.StatusNotFound:
		return fmt.Errorf("user not found")

	case http.StatusBadRequest:
		var errResp types.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			return fmt.Errorf("invalid error response: %w", err)
		}
		return errors.New(errResp.Error)

	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (c *Client) UpdateSubscription(ctx context.Context, user *models.User) error {

	if user == nil {
		return fmt.Errorf("user is nil")
	}

	url := fmt.Sprintf("%s/users/%s", c.baseURL, user.Username)

	body, err := json.Marshal(user)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create user: status %d", resp.StatusCode)
	}
	return nil
}
