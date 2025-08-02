# Go Lambda Container Example

A Go REST API designed for AWS Lambda containerized deployment with local development support.

## Features

- **Dual Mode**: Runs both as Lambda function and local server
- **Container Optimized**: Multi-stage Docker build for minimal image size
- **Lambda Integration**: Uses AWS Lambda Go API Proxy for seamless HTTP handling
- **Environment Detection**: Automatically detects Lambda vs local execution
- **Production Ready**: Optimized for cold start performance

## Architecture

```
Local Development    AWS Lambda
     |                   |
     v                   v
  Gin Router  →  Lambda Adapter  →  API Gateway
     |                   |
     v                   v
  HTTP Server      Lambda Runtime
```

## Usage

### Local Development

```bash
# Install dependencies
go mod tidy

# Run locally
go run main.go

# Test the API
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/users
```

### Docker Build and Test

```bash
# Build the container
docker build -t go-lambda-api .

# Run container locally (simulates Lambda environment)
docker run -p 9000:8080 -e AWS_LAMBDA_FUNCTION_NAME=test go-lambda-api

# Test with Lambda Runtime Interface Emulator
curl -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" -d '{
  "httpMethod": "GET",
  "path": "/health",
  "headers": {},
  "body": ""
}'
```

### AWS Lambda Deployment

```bash
# Tag for ECR
docker tag go-lambda-api:latest 123456789012.dkr.ecr.us-east-1.amazonaws.com/go-lambda-api:latest

# Push to ECR
docker push 123456789012.dkr.ecr.us-east-1.amazonaws.com/go-lambda-api:latest

# Update Lambda function
aws lambda update-function-code \
    --function-name go-lambda-api \
    --image-uri 123456789012.dkr.ecr.us-east-1.amazonaws.com/go-lambda-api:latest
```

## Key Lambda Patterns Demonstrated

### Environment Detection
```go
if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
    // Running in AWS Lambda
    lambda.Start(Handler)
} else {
    // Running locally
    r.Run(":8080")
}
```

### Lambda Handler
```go
func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    return ginLambda.ProxyWithContext(ctx, req)
}
```

### Multi-Stage Docker Build
- **Builder Stage**: Compiles Go binary with static linking
- **Runtime Stage**: Uses AWS Lambda base image with minimal footprint
- **Optimizations**: CGO disabled, static binary, stripped symbols

## Container Optimizations

- **Static Binary**: `CGO_ENABLED=0` for no external dependencies
- **Size Optimization**: `-ldflags='-w -s'` removes debug info
- **Lambda Base**: Uses `public.ecr.aws/lambda/provided:al2-x86_64`
- **CA Certificates**: Included for HTTPS calls
- **Timezone Data**: Available if needed

## Environment Variables

- `AWS_LAMBDA_FUNCTION_NAME`: Auto-set by Lambda, used for environment detection
- `ENVIRONMENT`: Custom environment identifier (dev/staging/prod)
- `VERSION`: Application version for health checks

## Performance Considerations

- **Cold Start**: Minimal initialization in `init()` function
- **Memory Usage**: Optimized for Lambda memory tiers
- **Response Time**: Gin framework for high-performance HTTP handling
- **Static Assets**: No static file serving (use CloudFront/S3)

## Production Checklist

- [ ] Configure proper IAM role for Lambda execution
- [ ] Set appropriate memory and timeout limits
- [ ] Configure CloudWatch logging and monitoring
- [ ] Set up API Gateway integration
- [ ] Configure environment variables securely
- [ ] Implement health checks and error handling
- [ ] Set up container registry (ECR) with proper permissions
- [ ] Configure CI/CD pipeline for automated deployments