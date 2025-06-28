package scheduler

import (
	"context"
	"log"
	"time"

	"youtube_downloader/pkg/database/repository"
)

// Scheduler represents a task scheduler for database operations
type Scheduler struct {
	Database *repository.Database
	stopChan chan bool
}

// NewScheduler creates a new Scheduler instance
func NewScheduler(database *repository.Database) *Scheduler {
	return &Scheduler{
		Database: database,
		stopChan: make(chan bool),
	}
}

// Start starts the scheduler with periodic tasks
func (s *Scheduler) Start() {
	log.Println("Starting scheduler...")

	// Start subscription checker
	go s.checkSubscriptions()

	// Start traffic resetter
	go s.resetTraffic()

	log.Println("Scheduler started successfully")
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	log.Println("Stopping scheduler...")
	close(s.stopChan)
}

// checkSubscriptions runs periodically to check and update subscription statuses
func (s *Scheduler) checkSubscriptions() {
	ticker := time.NewTicker(24 * time.Hour) // Check every 24 hours
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.performSubscriptionCheck()
		case <-s.stopChan:
			return
		}
	}
}

// resetTraffic runs monthly to reset user traffic
func (s *Scheduler) resetTraffic() {
	ticker := time.NewTicker(30 * 24 * time.Hour) // Reset every 30 days
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.performTrafficReset()
		case <-s.stopChan:
			return
		}
	}
}

// performSubscriptionCheck checks all user subscriptions and updates expired ones
func (s *Scheduler) performSubscriptionCheck() {
	log.Println("Performing subscription check...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	usernames, err := s.Database.AllUsername(ctx)
	if err != nil {
		log.Printf("Error getting usernames for subscription check: %v", err)
		return
	}

	for _, username := range usernames {
		user, err := s.Database.User(ctx, username)
		if err != nil {
			log.Printf("Error getting user %s for subscription check: %v", username, err)
			continue
		}

		// Check if subscription has expired
		if user.Subscription.SubscriptionStatus == "active" && time.Now().After(user.Subscription.EndSubscription) {
			log.Printf("Subscription expired for user: %s", username)

			// Update subscription status to inactive
			user.Subscription.SubscriptionStatus = "inactive"
			err = s.Database.UpdateUserSubscription(ctx, username, user.Subscription)
			if err != nil {
				log.Printf("Error updating expired subscription for user %s: %v", username, err)
			}
		}
	}

	log.Println("Subscription check completed")
}

// performTrafficReset resets traffic for all users
func (s *Scheduler) performTrafficReset() {
	log.Println("Performing traffic reset...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	usernames, err := s.Database.AllUsername(ctx)
	if err != nil {
		log.Printf("Error getting usernames for traffic reset: %v", err)
		return
	}

	for _, username := range usernames {
		err := s.Database.ResetUserTraffic(ctx, username)
		if err != nil {
			log.Printf("Error resetting traffic for user %s: %v", username, err)
		} else {
			log.Printf("Traffic reset for user: %s", username)
		}
	}

	log.Println("Traffic reset completed")
}
