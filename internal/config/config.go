package config

import (
	"log"
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	// Application settings
	Environment string
	Port        string
	IsLambda    bool

	// AWS configuration
	AWSRegion string

	// DynamoDB configuration
	SongsTable    string
	CommentsTable string

	// S3 configuration
	S3Bucket string

	// JWT/Cognito configuration
	CognitoUserPoolID string
	CognitoRegion     string
	JWKSUrl           string

	// API configuration
	APIVersion    string
	EnableMetrics bool
	LogLevel      string
}

// Load creates and loads configuration from environment variables
func Load() *Config {
	cfg := &Config{
		// Default values
		Environment:   getEnvOrDefault("ENVIRONMENT", "development"),
		Port:          getEnvOrDefault("PORT", "8080"),
		IsLambda:      getEnvBool("AWS_LAMBDA_RUNTIME_API", false),
		AWSRegion:     getEnvOrDefault("AWS_REGION", "us-east-1"),
		APIVersion:    getEnvOrDefault("API_VERSION", "v1"),
		EnableMetrics: getEnvBool("ENABLE_METRICS", true),
		LogLevel:      getEnvOrDefault("LOG_LEVEL", "info"),
	}

	// Required environment variables - fail fast if missing
	cfg.SongsTable = getRequiredEnv("SONGS_TABLE")
	cfg.CommentsTable = getRequiredEnv("COMMENTS_TABLE")
	cfg.S3Bucket = getRequiredEnv("S3_BUCKET")
	cfg.CognitoUserPoolID = getRequiredEnv("COGNITO_USER_POOL_ID")

	// Derive other configuration
	cfg.CognitoRegion = getEnvOrDefault("COGNITO_REGION", cfg.AWSRegion)
	cfg.JWKSUrl = buildJWKSUrl(cfg.CognitoRegion, cfg.CognitoUserPoolID)

	return cfg
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getRequiredEnv returns environment variable value or panics if not set
func getRequiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return value
}

// getEnvBool returns environment variable as boolean or default if not set
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		log.Printf("Invalid boolean value for %s: %s, using default: %t", key, value, defaultValue)
		return defaultValue
	}

	return boolValue
}

// buildJWKSUrl constructs the JWKS URL for AWS Cognito
func buildJWKSUrl(region, userPoolID string) string {
	return "https://cognito-idp." + region + ".amazonaws.com/" + userPoolID + "/.well-known/jwks.json"
}