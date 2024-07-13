package database_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	timeout = 15 * time.Second
	baseURL = "http://telegram-bot-api:8082"
)

// Client is a structure that contains the HTTP client and the base URL of the server.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

type User struct {
	ID                 int64  `json:"id"`
	TelegramUsername   string `json:"username"`
	SubscriptionStatus string `json:"subscription_status"`
}

// NewUser initializes and returns a new User.
func NewUser(username string) *User {
	return &User{
		TelegramUsername:   username,
		SubscriptionStatus: "inactive",
	}
}

// NewClient initializes and returns a new Client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: timeout},
		baseURL:    baseURL,
	}
}

// CreateUser sends a request to create a new user.
func (c *Client) CreateUser(ctx context.Context, newUser *User) error {
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
func (c *Client) GetUser(ctx context.Context, username string) (*User, error) {
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

	// Decode the response body into a user struct
	var retrievedUser User
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
