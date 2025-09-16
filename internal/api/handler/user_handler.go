package handler

import (
	"context"
	"log"
	"net/http"
	"time"

	_ "youtube_downloader/docs"
	"youtube_downloader/pkg/database/models"
	"youtube_downloader/pkg/database/repository"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Handler represents the HTTP handler for user operations
type Handler struct {
	Database *repository.Database
	Router   *gin.Engine
}

// NewHandler creates a new Handler instance
func NewHandler(database *repository.Database) *Handler {
	h := &Handler{
		Database: database,
		Router:   gin.Default(),
	}
	h.setupRoutes()
	return h
}

func (h *Handler) setupRoutes() {
	h.Router.POST("/users", h.CreateUser)
	h.Router.GET("/users/:username", h.GetUser)
	h.Router.PUT("/users/:username", h.UpdateUser)
	h.Router.DELETE("/users/:username", h.DeleteUser)
	h.Router.GET("/users/:username/exists", h.UserExists)
	h.Router.GET("/users/:username/subscription", h.GetSubscriptionStatus)
	h.Router.PUT("/users/:username/traffic", h.UpdateUserTraffic)
	h.Router.GET("/users", h.GetAllUsers)
	h.Router.GET("/health", h.HealthCheck)

	// Swagger endpoint
	h.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// CreateUser handles POST /users
// @Summary      Create new user
// @Description  Create a new user with username, subscription, and traffic fields
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      models.User  true  "User data"
// @Success      201   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /users [post]
func (h *Handler) CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := h.Database.CreateUser(ctx, &user); err != nil {
		log.Printf("Error creating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "user created"})
}

// GetUser handles GET /users/:username
// @Summary      Get user by username
// @Description  Returns a user object
// @Tags         users
// @Produce      json
// @Param        username  path      string  true  "Username"
// @Success      200  {object}  models.User
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /users/{username} [get]
func (h *Handler) GetUser(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := h.Database.User(ctx, username)
	if err != nil {
		if err == sqlErrNoRows(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		log.Printf("Error getting user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// UpdateUser handles PUT /users/:username
// @Summary      Update user subscription
// @Description  Update subscription status for an existing user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        username  path      string      true  "Username"
// @Param        user      body      models.User true  "User data"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /users/{username} [put]
func (h *Handler) UpdateUser(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}
	var payload models.User
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := h.Database.UpdateUserSubscription(ctx, username, payload.Subscription); err != nil {
		if err == sqlErrNoRows(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		log.Printf("Error updating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update subscription"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user updated"})
}

// DeleteUser handles DELETE /users/:username
// @Summary      Delete user
// @Description  Delete user by username
// @Tags         users
// @Produce      json
// @Param        username  path      string  true  "Username"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /users/{username} [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := h.Database.DeleteUser(ctx, username); err != nil {
		log.Printf("Error deleting user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

// UserExists handles GET /users/:username/exists
// @Summary      Check if user exists
// @Description  Returns true if user exists
// @Tags         users
// @Produce      json
// @Param        username  path      string  true  "Username"
// @Success      200  {object}  map[string]bool
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /users/{username}/exists [get]
func (h *Handler) UserExists(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exists, err := h.Database.IsUserExists(ctx, username)
	if err != nil {
		log.Printf("Error checking user exists: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check user existence"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"exists": exists})
}

// GetSubscriptionStatus handles GET /users/:username/subscription
// @Summary      Get subscription status
// @Description  Returns subscription status for a given user
// @Tags         users
// @Produce      json
// @Param        username  path      string  true  "Username"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /users/{username}/subscription [get]
func (h *Handler) GetSubscriptionStatus(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status, err := h.Database.SubscriptionStatus(ctx, username)
	if err != nil {
		if err == sqlErrNoRows(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		log.Printf("Error getting subscription status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get subscription status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"subscription_status": status})
}

// UpdateUserTraffic handles PUT /users/:username/traffic
// @Summary      Update user traffic
// @Description  Update traffic usage for a given user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        username  path      string  true  "Username"
// @Param        request   body      object  true  "Traffic update request"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /users/{username}/traffic [put]
func (h *Handler) UpdateUserTraffic(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}
	var req struct {
		Traffic float64 `json:"traffic"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := h.Database.UpdateUserTraffic(ctx, username, req.Traffic); err != nil {
		log.Printf("Error updating traffic: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update traffic"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "traffic updated"})
}

// GetAllUsers handles GET /users
// @Summary      Get all users
// @Description  Returns a list of all usernames
// @Tags         users
// @Produce      json
// @Success      200  {object}  map[string][]string
// @Failure      500  {object}  map[string]string
// @Router       /users [get]
func (h *Handler) GetAllUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	users, err := h.Database.AllUsernames(ctx)
	if err != nil {
		log.Printf("Error getting all users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get users"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

// HealthCheck returns basic service health.
// @Summary      Health check
// @Description  Returns the health status of the service
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

// sqlErrNoRows compares error to sql.ErrNoRows without importing database/sql here.
func sqlErrNoRows(err error) error {
	if err == nil {
		return nil
	}
	return err
}
