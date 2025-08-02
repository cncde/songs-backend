package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// User represents a user in our system
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	CognitoSub string `json:"cognito_sub,omitempty"`
}

// JWTClaims represents the structure of JWT claims from AWS Cognito
type JWTClaims struct {
	Sub           string `json:"sub"`
	CognitoGroups []string `json:"cognito:groups"`
	CognitoUsername string `json:"cognito:username"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Aud           string `json:"aud"`
	EventID       string `json:"event_id"`
	TokenUse      string `json:"token_use"`
	AuthTime      int64  `json:"auth_time"`
	Iss           string `json:"iss"`
	Exp           int64  `json:"exp"`
	Iat           int64  `json:"iat"`
	jwt.RegisteredClaims
}

// In-memory user storage (replace with database in production)
var users = map[string]User{
	"user-123": {ID: "user-123", Username: "john_doe", Email: "john@example.com"},
	"user-456": {ID: "user-456", Username: "jane_doe", Email: "jane@example.com"},
}

// Global key function for JWT verification
var keyFunc jwt.Keyfunc

func main() {
	// Initialize JWT key function for AWS Cognito
	if err := initJWTKeyFunc(); err != nil {
		log.Fatal("Failed to initialize JWT key function:", err)
	}

	// Create Gin router
	r := gin.Default()

	// Public endpoints (no authentication required)
	public := r.Group("/public")
	{
		public.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "healthy",
				"service": "jwt-auth-api",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
		})
		public.GET("/info", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "This is a public endpoint",
				"auth_required": false,
			})
		})
	}

	// Protected endpoints (authentication required)
	protected := r.Group("/api/v1")
	protected.Use(JWTAuthMiddleware())
	{
		protected.GET("/profile", getUserProfile)
		protected.PUT("/profile", updateUserProfile)
		protected.GET("/users", getUsers)
		protected.GET("/protected", func(c *gin.Context) {
			// Get user info from context (set by middleware)
			userSub, exists := c.Get("user_sub")
			if !exists {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "User context not found"})
				return
			}
			
			username, _ := c.Get("username")
			email, _ := c.Get("email")
			
			c.JSON(http.StatusOK, gin.H{
				"message": "This is a protected endpoint",
				"user_sub": userSub,
				"username": username,
				"email": email,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
		})
	}

	log.Println("Starting JWT Auth API server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// initJWTKeyFunc initializes the JWT key function for AWS Cognito
func initJWTKeyFunc() error {
	// Get Cognito configuration from environment variables
	userPoolID := os.Getenv("COGNITO_USER_POOL_ID")
	region := os.Getenv("AWS_REGION")
	
	if userPoolID == "" {
		userPoolID = "us-east-1_EXAMPLE123" // Default for demo
	}
	if region == "" {
		region = "us-east-1" // Default region
	}

	// Construct the JWKS URL for AWS Cognito
	jwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", region, userPoolID)
	
	// Create the key function options
	options := keyfunc.Options{
		RefreshErrorHandler: func(err error) {
			log.Printf("JWKS refresh error: %v", err)
		},
		RefreshInterval: time.Hour, // Refresh keys every hour
		RefreshTimeout:  10 * time.Second,
	}

	// Create the key function
	kf, err := keyfunc.NewDefault([]string{jwksURL}, options)
	if err != nil {
		return fmt.Errorf("failed to create key function: %w", err)
	}

	keyFunc = kf.Keyfunc
	log.Printf("Initialized JWT key function with JWKS URL: %s", jwksURL)
	return nil
}

// JWTAuthMiddleware validates JWT tokens from AWS Cognito
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
				"code": "MISSING_AUTH_HEADER",
			})
			c.Abort()
			return
		}

		// Check Bearer prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header must use Bearer scheme",
				"code": "INVALID_AUTH_FORMAT",
			})
			c.Abort()
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token is required",
				"code": "MISSING_TOKEN",
			})
			c.Abort()
			return
		}

		// Parse and validate the token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, keyFunc)
		if err != nil {
			log.Printf("JWT parsing error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
				"code": "INVALID_TOKEN",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Check if token is valid
		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token is not valid",
				"code": "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
				"code": "INVALID_CLAIMS",
			})
			c.Abort()
			return
		}

		// Validate token use (should be "access" for API access)
		if claims.TokenUse != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token use, expected access token",
				"code": "INVALID_TOKEN_USE",
			})
			c.Abort()
			return
		}

		// Store user information in context for use in handlers
		c.Set("user_sub", claims.Sub)
		c.Set("username", claims.CognitoUsername)
		c.Set("email", claims.Email)
		c.Set("cognito_groups", claims.CognitoGroups)
		c.Set("jwt_claims", claims)

		// Continue to next handler
		c.Next()
	}
}

// getUserProfile returns the current user's profile
func getUserProfile(c *gin.Context) {
	userSub, exists := c.Get("user_sub")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User context not found"})
		return
	}

	// In a real app, you would look up the user in your database
	// using the Cognito sub (subject) identifier
	username, _ := c.Get("username")
	email, _ := c.Get("email")
	groups, _ := c.Get("cognito_groups")

	c.JSON(http.StatusOK, gin.H{
		"profile": gin.H{
			"sub": userSub,
			"username": username,
			"email": email,
			"groups": groups,
		},
		"message": "Profile retrieved successfully",
	})
}

// updateUserProfile updates the current user's profile
func updateUserProfile(c *gin.Context) {
	userSub, exists := c.Get("user_sub")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User context not found"})
		return
	}

	var updateData struct {
		DisplayName string `json:"display_name"`
		Bio         string `json:"bio"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON format",
			"details": err.Error(),
		})
		return
	}

	// In a real app, you would update the user in your database
	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user_sub": userSub,
		"updated_data": updateData,
	})
}

// getUsers returns all users (admin function)
func getUsers(c *gin.Context) {
	// Check if user has admin privileges
	groups, exists := c.Get("cognito_groups")
	if exists {
		if groupList, ok := groups.([]string); ok {
			hasAdminAccess := false
			for _, group := range groupList {
				if group == "admin" || group == "moderator" {
					hasAdminAccess = true
					break
				}
			}
			if !hasAdminAccess {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Insufficient permissions",
					"code": "FORBIDDEN",
				})
				return
			}
		}
	}

	// Return all users
	userList := make([]User, 0, len(users))
	for _, user := range users {
		userList = append(userList, user)
	}

	c.JSON(http.StatusOK, gin.H{
		"users": userList,
		"count": len(userList),
	})
}