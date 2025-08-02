package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

// User represents a user in our system
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// In-memory user storage (replace with DynamoDB in production)
var users = map[string]User{
	"1": {ID: "1", Username: "lambda_user", Email: "lambda@example.com"},
	"2": {ID: "2", Username: "container_user", Email: "container@example.com"},
}

var ginLambda *ginlambda.GinLambda

func init() {
	// Initialize Gin router
	r := setupRouter()
	
	// Create Lambda adapter
	ginLambda = ginlambda.New(r)
}

func setupRouter() *gin.Engine {
	// Set Gin to release mode for Lambda
	gin.SetMode(gin.ReleaseMode)
	
	// Create Gin router
	r := gin.New()
	
	// Add middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	
	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "healthy",
			"service":     "go-lambda-api",
			"environment": os.Getenv("ENVIRONMENT"),
			"version":     os.Getenv("VERSION"),
		})
	})
	
	// API routes
	v1 := r.Group("/api/v1")
	{
		v1.GET("/users", getUsers)
		v1.GET("/users/:id", getUser)
		v1.POST("/users", createUser)
	}
	
	return r
}

func main() {
	// Check if running in Lambda environment
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		// Running in AWS Lambda
		log.Println("Starting Lambda function")
		lambda.Start(Handler)
	} else {
		// Running locally
		log.Println("Starting local server on :8080")
		r := setupRouter()
		if err := r.Run(":8080"); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}
}

// Handler is the Lambda function handler
func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Use the Gin Lambda adapter to handle the request
	return ginLambda.ProxyWithContext(ctx, req)
}

// getUsers returns all users
func getUsers(c *gin.Context) {
	userList := make([]User, 0, len(users))
	for _, user := range users {
		userList = append(userList, user)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"users": userList,
		"count": len(userList),
	})
}

// getUser returns a specific user by ID
func getUser(c *gin.Context) {
	id := c.Param("id")
	user, exists := users[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
			"id":    id,
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// createUser creates a new user
func createUser(c *gin.Context) {
	var newUser User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON format",
			"details": err.Error(),
		})
		return
	}
	
	// Simple validation
	if newUser.Username == "" || newUser.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Username and email are required",
		})
		return
	}
	
	// Generate simple ID
	newUser.ID = generateID()
	users[newUser.ID] = newUser
	
	c.JSON(http.StatusCreated, gin.H{
		"user": newUser,
		"message": "User created successfully",
	})
}

// generateID generates a simple ID (replace with UUID in production)
func generateID() string {
	return fmt.Sprintf("lambda_user_%d", len(users)+1)
}