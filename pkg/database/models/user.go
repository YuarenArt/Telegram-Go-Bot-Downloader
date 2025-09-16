package models

import "time"

// User represents a user in the system
type User struct {
	Username     string       `json:"username"`
	Subscription Subscription `json:"subscription"`
	Traffic      float64      `json:"traffic"`
	ChatID       int64        `json:"chat_id"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// Subscription represents a user's subscription
type Subscription struct {
	ID                 int64     `json:"id"`
	SubscriptionStatus string    `json:"subscription_status"` // active, inactive
	Duration           string    `json:"duration"`            // month, year, forever
	StartSubscription  time.Time `json:"start_subscription"`
	EndSubscription    time.Time `json:"end_subscription"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
