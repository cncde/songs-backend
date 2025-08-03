package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/cncde/songs-backend/internal/config"
	"github.com/cncde/songs-backend/internal/server"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize AWS configuration
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(cfg.AWSRegion),
	)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	// Initialize AWS service clients
	dynamoClient := dynamodb.NewFromConfig(awsCfg)
	s3Client := s3.NewFromConfig(awsCfg)

	// Create dependencies
	deps := &server.Dependencies{
		Config:       cfg,
		DynamoClient: dynamoClient,
		S3Client:     s3Client,
	}

	// Run the application
	if err := run(deps); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}

// run initializes and starts the application
func run(deps *server.Dependencies) error {
	// Create server
	srv, err := server.New(deps)
	if err != nil {
		return err
	}

	// Start server based on environment
	if deps.Config.IsLambda {
		// Lambda environment - use AWS Lambda Go API Proxy
		ginLambda := ginadapter.New(srv.Handler())
		lambda.Start(ginLambda.ProxyWithContext)
		return nil
	} else {
		// Local development environment - start HTTP server
		log.Printf("Starting server on port %s", deps.Config.Port)
		return srv.Handler().Run(":" + deps.Config.Port)
	}
}