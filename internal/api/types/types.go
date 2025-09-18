package types

import (
	"time"
)

// UserRequest represents the request body for user creation and updates
type UserRequest struct {
	Username string `json:"username"`
	ChatID   int64  `json:"chat_id"`
}

// UserResponse represents the user data in API responses
type UserResponse struct {
	Username     string               `json:"username"`
	Traffic      int64                `json:"traffic"`
	ChatID       int64                `json:"chat_id"`
	Subscription SubscriptionResponse `json:"subscription"`
}

// SubscriptionResponse represents subscription data in API responses
type SubscriptionResponse struct {
	StartSubscription time.Time `json:"start_subscription"`
	EndSubscription   time.Time `json:"end_subscription"`
	Duration          string    `json:"duration"`
}

// TrafficUpdateRequest represents the request body for updating user traffic
type TrafficUpdateRequest struct {
	Traffic int64 `json:"traffic"`
}

// SubscriptionUpdateRequest represents the request body for updating user subscription
type SubscriptionUpdateRequest struct {
	Subscription SubscriptionResponse `json:"subscription"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string `json:"message"`
}
