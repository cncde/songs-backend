package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// User represents a user in our system
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// In-memory user storage (replace with database in production)
var users = map[string]User{
	"1": {ID: "1", Username: "john_doe", Email: "john@example.com"},
	"2": {ID: "2", Username: "jane_doe", Email: "jane@example.com"},
}

func main() {
	// Create Gin router with default middleware (logger and recovery)
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "go-api-example",
		})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Get all users
		v1.GET("/users", getUsers)
		// Get user by ID
		v1.GET("/users/:id", getUser)
		// Create new user
		v1.POST("/users", createUser)
		// Update user
		v1.PUT("/users/:id", updateUser)
		// Delete user
		v1.DELETE("/users/:id", deleteUser)
	}

	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// getUsers returns all users
func getUsers(c *gin.Context) {
	userList := make([]User, 0, len(users))
	for _, user := range users {
		userList = append(userList, user)
	}
	c.JSON(http.StatusOK, gin.H{"users": userList})
}

// getUser returns a specific user by ID
func getUser(c *gin.Context) {
	id := c.Param("id")
	user, exists := users[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// createUser creates a new user
func createUser(c *gin.Context) {
	var newUser User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Simple validation
	if newUser.Username == "" || newUser.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and email are required"})
		return
	}

	// Generate simple ID (use UUID in production)
	newUser.ID = generateID()
	users[newUser.ID] = newUser

	c.JSON(http.StatusCreated, gin.H{"user": newUser})
}

// updateUser updates an existing user
func updateUser(c *gin.Context) {
	id := c.Param("id")
	_, exists := users[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var updatedUser User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedUser.ID = id
	users[id] = updatedUser

	c.JSON(http.StatusOK, gin.H{"user": updatedUser})
}

// deleteUser removes a user
func deleteUser(c *gin.Context) {
	id := c.Param("id")
	_, exists := users[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	delete(users, id)
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// generateID generates a simple ID (replace with proper UUID generation)
func generateID() string {
	return "user_" + string(rune(len(users)+3)) // Simple incrementing ID
}