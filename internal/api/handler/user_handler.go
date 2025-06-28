package handler

import (
	"context"
	"log"
	"net/http"
	"time"

	"youtube_downloader/pkg/database/models"
	"youtube_downloader/pkg/database/repository"

	"github.com/gin-gonic/gin"
)

// Handler represents the HTTP handler for user operations
type Handler struct {
	Database *repository.Database
	Router   *gin.Engine
}

// NewHandler creates a new Handler instance
func NewHandler(database *repository.Database) *Handler {
	handler := &Handler{
		Database: database,
		Router:   gin.Default(),
	}

	handler.setupRoutes()
	return handler
}

// setupRoutes configures all the routes for the API
func (h *Handler) setupRoutes() {
	// User routes
	h.Router.POST("/users", h.CreateUser)
	h.Router.GET("/users/:username", h.GetUser)
	h.Router.PUT("/users/:username", h.UpdateUser)
	h.Router.DELETE("/users/:username", h.DeleteUser)
	h.Router.GET("/users/:username/exists", h.UserExists)
	h.Router.GET("/users/:username/subscription", h.GetSubscriptionStatus)
	h.Router.PUT("/users/:username/traffic", h.UpdateUserTraffic)
	h.Router.GET("/users", h.GetAllUsers)

	// Health check
	h.Router.GET("/health", h.HealthCheck)
}

// CreateUser handles POST /users
func (h *Handler) CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := h.Database.CreateUser(ctx, &user)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

// GetUser handles GET /users/:username
func (h *Handler) GetUser(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	user, err := h.Database.User(ctx, username)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser handles PUT /users/:username
func (h *Handler) UpdateUser(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := h.Database.UpdateUserSubscription(ctx, username, user.Subscription)
	if err != nil {
		log.Printf("Error updating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// DeleteUser handles DELETE /users/:username
func (h *Handler) DeleteUser(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := h.Database.DeleteUser(ctx, username)
	if err != nil {
		log.Printf("Error deleting user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// UserExists handles GET /users/:username/exists
func (h *Handler) UserExists(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	exists, err := h.Database.IsUserExists(ctx, username)
	if err != nil {
		log.Printf("Error checking if user exists: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user existence"})
		return
	}

	if exists {
		c.JSON(http.StatusOK, gin.H{"exists": true})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"exists": false})
	}
}

// GetSubscriptionStatus handles GET /users/:username/subscription
func (h *Handler) GetSubscriptionStatus(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	status, err := h.Database.SubscriptionStatus(ctx, username)
	if err != nil {
		log.Printf("Error getting subscription status: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subscription_status": status})
}

// UpdateUserTraffic handles PUT /users/:username/traffic
func (h *Handler) UpdateUserTraffic(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	var request struct {
		Traffic float64 `json:"traffic"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := h.Database.UpdateUserTraffic(ctx, username, request.Traffic)
	if err != nil {
		log.Printf("Error updating user traffic: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user traffic"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User traffic updated successfully"})
}

// GetAllUsers handles GET /users
func (h *Handler) GetAllUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	usernames, err := h.Database.AllUsername(ctx)
	if err != nil {
		log.Printf("Error getting all users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": usernames})
}

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
