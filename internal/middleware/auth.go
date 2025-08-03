package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the expected JWT claims structure
type JWTClaims struct {
	Subject   string `json:"sub"`
	Username  string `json:"cognito:username"`
	Email     string `json:"email"`
	TokenUse  string `json:"token_use"`
	Audience  string `json:"aud"`
	Issuer    string `json:"iss"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
	jwt.RegisteredClaims
}

// UserContext represents authenticated user information
type UserContext struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// JWTMiddleware handles JWT authentication using AWS Cognito
type JWTMiddleware struct {
	jwks         *keyfunc.JWKS
	jwksURL      string
	userPoolID   string
	expectedAud  string
	refreshTimer *time.Timer
}

// NewJWTMiddleware creates a new JWT middleware instance
func NewJWTMiddleware(jwksURL, userPoolID, expectedAudience string) (*JWTMiddleware, error) {
	// Initialize JWKS from Cognito
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshInterval:   time.Hour,     // Refresh keys every hour
		RefreshTimeout:    10 * time.Second,
		RefreshRateLimit:  time.Minute,   // Rate limit key refresh
		RefreshErrorHandler: func(err error) {
			log.Printf("Failed to refresh JWKS: %v", err)
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize JWKS: %w", err)
	}

	return &JWTMiddleware{
		jwks:        jwks,
		jwksURL:     jwksURL,
		userPoolID:  userPoolID,
		expectedAud: expectedAudience,
	}, nil
}

// Middleware returns the Gin middleware function
func (j *JWTMiddleware) Middleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			j.respondWithError(c, http.StatusUnauthorized, "authorization header required")
			return
		}

		// Parse Bearer token
		tokenString, err := j.extractBearerToken(authHeader)
		if err != nil {
			j.respondWithError(c, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		// Parse and validate JWT
		token, err := j.parseAndValidateJWT(tokenString)
		if err != nil {
			j.respondWithError(c, http.StatusUnauthorized, "invalid token")
			return
		}

		// Extract claims
		userCtx, err := j.extractUserContext(token)
		if err != nil {
			j.respondWithError(c, http.StatusUnauthorized, "invalid token claims")
			return
		}

		// Store user context in Gin context
		c.Set("user", userCtx)
		c.Set("user_id", userCtx.UserID)
		c.Set("username", userCtx.Username)
		c.Set("email", userCtx.Email)

		c.Next()
	})
}

// OptionalMiddleware returns middleware that doesn't require authentication but extracts user if present
func (j *JWTMiddleware) OptionalMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		tokenString, err := j.extractBearerToken(authHeader)
		if err != nil {
			c.Next()
			return
		}

		token, err := j.parseAndValidateJWT(tokenString)
		if err != nil {
			c.Next()
			return
		}

		userCtx, err := j.extractUserContext(token)
		if err != nil {
			c.Next()
			return
		}

		// Store user context if valid
		c.Set("user", userCtx)
		c.Set("user_id", userCtx.UserID)
		c.Set("username", userCtx.Username)
		c.Set("email", userCtx.Email)

		c.Next()
	})
}

// extractBearerToken extracts the token from Bearer authorization header
func (j *JWTMiddleware) extractBearerToken(authHeader string) (string, error) {
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}

// parseAndValidateJWT parses and validates the JWT token
func (j *JWTMiddleware) parseAndValidateJWT(tokenString string) (*jwt.Token, error) {
	// Parse token with JWKS validation
	token, err := jwt.Parse(tokenString, j.jwks.Keyfunc, jwt.WithValidMethods([]string{"RS256"}))
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}

// extractUserContext extracts user information from JWT claims
func (j *JWTMiddleware) extractUserContext(token *jwt.Token) (*UserContext, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims format")
	}

	// Validate required claims
	if err := j.validateClaims(claims); err != nil {
		return nil, fmt.Errorf("claim validation failed: %w", err)
	}

	// Extract user information
	userCtx := &UserContext{
		UserID:   j.getStringClaim(claims, "sub"),
		Username: j.getStringClaim(claims, "cognito:username"),
		Email:    j.getStringClaim(claims, "email"),
	}

	// Validate required fields
	if userCtx.UserID == "" {
		return nil, errors.New("missing sub claim")
	}

	return userCtx, nil
}

// validateClaims validates JWT claims according to AWS Cognito standards
func (j *JWTMiddleware) validateClaims(claims jwt.MapClaims) error {
	// Validate issuer
	iss, ok := claims["iss"].(string)
	if !ok {
		return errors.New("missing or invalid iss claim")
	}
	expectedIssuer := fmt.Sprintf("https://cognito-idp.us-east-1.amazonaws.com/%s", j.userPoolID)
	if iss != expectedIssuer {
		return fmt.Errorf("invalid issuer: expected %s, got %s", expectedIssuer, iss)
	}

	// Validate token use
	tokenUse, ok := claims["token_use"].(string)
	if !ok || (tokenUse != "access" && tokenUse != "id") {
		return errors.New("invalid or missing token_use claim")
	}

	// Validate audience (for id tokens)
	if tokenUse == "id" && j.expectedAud != "" {
		aud, ok := claims["aud"].(string)
		if !ok || aud != j.expectedAud {
			return fmt.Errorf("invalid audience: expected %s, got %s", j.expectedAud, aud)
		}
	}

	// Validate expiration (additional check beyond jwt library)
	exp, ok := claims["exp"].(float64)
	if !ok {
		return errors.New("missing or invalid exp claim")
	}
	if time.Now().Unix() > int64(exp) {
		return errors.New("token has expired")
	}

	// Validate issued at time
	iat, ok := claims["iat"].(float64)
	if !ok {
		return errors.New("missing or invalid iat claim")
	}
	if time.Now().Unix() < int64(iat) {
		return errors.New("token issued in the future")
	}

	return nil
}

// getStringClaim safely extracts a string claim
func (j *JWTMiddleware) getStringClaim(claims jwt.MapClaims, key string) string {
	if value, ok := claims[key].(string); ok {
		return value
	}
	return ""
}

// respondWithError sends an error response
func (j *JWTMiddleware) respondWithError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"error":   message,
		"status":  statusCode,
		"path":    c.Request.URL.Path,
		"method":  c.Request.Method,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
	c.Abort()
}

// GetUserFromContext extracts user information from Gin context
func GetUserFromContext(c *gin.Context) (*UserContext, error) {
	user, exists := c.Get("user")
	if !exists {
		return nil, errors.New("user not found in context")
	}

	userCtx, ok := user.(*UserContext)
	if !ok {
		return nil, errors.New("invalid user context type")
	}

	return userCtx, nil
}

// GetUserIDFromContext extracts user ID from Gin context
func GetUserIDFromContext(c *gin.Context) (string, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", errors.New("user ID not found in context")
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return "", errors.New("invalid user ID type")
	}

	return userIDStr, nil
}

// RequireAuth ensures user is authenticated (helper for handlers)
func RequireAuth(c *gin.Context) (*UserContext, error) {
	user, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":     "authentication required",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		c.Abort()
		return nil, err
	}

	return user, nil
}

// Close closes the JWKS and stops background refresh
func (j *JWTMiddleware) Close() error {
	if j.refreshTimer != nil {
		j.refreshTimer.Stop()
	}
	j.jwks.EndBackground()
	return nil
}

// RefreshJWKS manually refreshes the JWKS (useful for testing or forced refresh)
func (j *JWTMiddleware) RefreshJWKS(ctx context.Context) error {
	return j.jwks.Refresh(ctx, keyfunc.RefreshOptions{})
}

// HealthCheck verifies the JWKS endpoint is accessible
func (j *JWTMiddleware) HealthCheck(ctx context.Context) error {
	// Try to refresh JWKS to verify connectivity
	return j.jwks.Refresh(ctx, keyfunc.RefreshOptions{})
}