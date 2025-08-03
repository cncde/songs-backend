# Execute Go API PRP

Execute a comprehensive Go REST API PRP to implement AWS Lambda containerized APIs with JWT authentication, DynamoDB integration, and S3 file handling.

## PRP File: PRPs/{filename}.md

## Execution Process

1. **Load Go API PRP**
   - Read the specified Go API PRP file completely
   - Understand all Go framework requirements and AWS integrations
   - Review all web research findings and implementation patterns
   - Follow all instructions for Go API development

2. **ULTRATHINK - Go API Architecture Design**
   - Create comprehensive Go API implementation plan
   - Design project structure following Go conventions (cmd/, internal/, pkg/)
   - Plan HTTP routing, middleware chains, and handler organization
   - Design AWS Lambda container deployment strategy
   - Plan JWT authentication integration with AWS Cognito
   - Design DynamoDB data access patterns and error handling
   - Plan S3 file upload workflows and security measures

3. **Implement Go API Foundation**
   - Create Go module with proper dependencies
   - Set up standard Go project structure (cmd/, internal/, pkg/)
   - Implement HTTP server with chosen framework (Gin/Echo/Gorilla Mux)
   - Create middleware for authentication, logging, CORS, error handling
   - Set up AWS SDK v2 clients for DynamoDB and S3

4. **Implement Core Features**
   - Build JWT authentication middleware with AWS Cognito integration
   - Create DynamoDB repository layer with proper error handling
   - Implement S3 file upload handlers with security validation
   - Add comprehensive input validation and sanitization
   - Implement proper Go error handling and context propagation

5. **AWS Lambda Integration**
   - Create multi-stage Dockerfile for Lambda container deployment
   - Configure Lambda handler with proper context management
   - Set up AWS IAM roles with least-privilege permissions
   - Implement Lambda-specific logging and error handling
   - Configure environment-based configuration management

6. **Testing & Quality Assurance**
   - Implement comprehensive unit tests with table-driven patterns
   - Create integration tests for AWS service interactions
   - Add HTTP handler tests using httptest package
   - Implement security tests for authentication and authorization
   - Set up mocking for external dependencies

7. **CI/CD & Deployment**
   - Create GitHub Actions workflow for Go testing and building
   - Set up Docker image building and pushing to ECR
   - Configure Lambda deployment automation
   - Implement post-deployment smoke tests
   - Set up monitoring and logging

## Go API Implementation Requirements

### Project Structure
```
cmd/
  api/
    main.go                 # Application entry point and Lambda handler
internal/
  api/
    handlers/              # HTTP request handlers
      auth.go             # Authentication handlers
      users.go            # User management handlers
      files.go            # File upload handlers
    middleware/            # Custom middleware
      jwt.go              # JWT authentication middleware
      cors.go             # CORS middleware
      logging.go          # Request logging middleware
    models/               # Data models and structs
      user.go             # User data models
      response.go         # API response models
    services/             # Business logic services
      user_service.go     # User business logic
      file_service.go     # File handling logic
  config/                 # Configuration management
    config.go             # Application configuration
  db/                     # Database layer
    dynamodb.go           # DynamoDB client and operations
  storage/                # File storage layer
    s3.go                 # S3 client and operations
pkg/
  auth/                   # JWT validation utilities
    cognito.go            # AWS Cognito JWT verification
  utils/                  # Shared utilities
    validator.go          # Input validation helpers
api/
  openapi.yaml            # API specification (OpenAPI/Swagger)
configs/
  config.yaml             # Configuration templates
Dockerfile                # Multi-stage Lambda container build
go.mod                    # Go module definition
go.sum                    # Go module checksums
```

### Core Implementation Patterns

**HTTP Server Setup (with Gin):**
```go
func main() {
    cfg := config.Load()
    
    // Initialize AWS clients
    awsCfg, _ := config.LoadDefaultConfig(context.TODO())
    dynamoClient := dynamodb.NewFromConfig(awsCfg)
    s3Client := s3.NewFromConfig(awsCfg)
    
    // Initialize services
    userService := services.NewUserService(dynamoClient, cfg.UserTable)
    
    // Setup router and middleware
    r := gin.New()
    r.Use(middleware.Logger())
    r.Use(middleware.CORS())
    r.Use(middleware.Recovery())
    
    // Setup routes with JWT middleware
    api := r.Group("/api/v1")
    api.Use(middleware.JWTAuth(cfg.CognitoUserPoolID))
    
    // Register handlers
    handlers.RegisterUserHandlers(api, userService)
    handlers.RegisterFileHandlers(api, s3Client, cfg.S3Bucket)
    
    // Lambda or local server
    if cfg.IsLambda {
        lambda.Start(ginLambda.New(r).ProxyWithContext)
    } else {
        r.Run(":" + cfg.Port)
    }
}
```

**JWT Authentication Middleware:**
```go
func JWTAuth(userPoolID string) gin.HandlerFunc {
    keyFunc, err := keyfunc.NewCognito(context.Background(), keyfunc.CognitoOptions{
        Region:     "us-east-1",
        UserPoolID: userPoolID,
    })
    
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, gin.H{"error": "authorization header required"})
            c.Abort()
            return
        }
        
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        token, err := jwt.Parse(tokenString, keyFunc.Keyfunc)
        
        if err != nil || !token.Valid {
            c.JSON(401, gin.H{"error": "invalid token"})
            c.Abort()
            return
        }
        
        // Store claims in context
        if claims, ok := token.Claims.(jwt.MapClaims); ok {
            c.Set("user_id", claims["sub"])
            c.Set("username", claims["cognito:username"])
        }
        
        c.Next()
    }
}
```

**DynamoDB Service Pattern:**
```go
type UserService struct {
    client *dynamodb.Client
    table  string
}

func (s *UserService) GetUser(ctx context.Context, id string) (*models.User, error) {
    result, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
        TableName: aws.String(s.table),
        Key: map[string]types.AttributeValue{
            "id": &types.AttributeValueMemberS{Value: id},
        },
    })
    
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    
    if result.Item == nil {
        return nil, fmt.Errorf("user not found: %s", id)
    }
    
    var user models.User
    if err := attributevalue.UnmarshalMap(result.Item, &user); err != nil {
        return nil, fmt.Errorf("failed to unmarshal user: %w", err)
    }
    
    return &user, nil
}
```

## Validation Loop

### Level 1: Go Code Quality Validation

```bash
# Go module and dependencies
go mod tidy                          # Clean up dependencies
go mod verify                        # Verify module integrity
go mod download                      # Download dependencies

# Static analysis and linting
go vet ./...                        # Go's built-in static analysis
golangci-lint run                   # Comprehensive linting
gosec ./...                         # Security vulnerability scan

# Expected: All checks pass without errors
# If failing: Fix code issues, update dependencies, address security concerns
```

### Level 2: Testing Validation

```bash
# Unit and integration testing
go test ./...                       # Run all tests
go test -race ./...                 # Test with race detection
go test -cover ./...                # Test with coverage analysis
go test -v -run TestJWTMiddleware    # Test specific components

# HTTP handler testing
go test ./internal/api/handlers/... -v

# Expected: All tests pass, good coverage (>80%), no race conditions
# If failing: Fix failing tests, add missing tests, resolve race conditions
```

### Level 3: Build and Container Validation

```bash
# Go build validation
go build ./...                                          # Build all packages
go build -o api cmd/api/main.go                        # Build main binary
GOOS=linux GOARCH=amd64 go build -o main cmd/api/main.go  # Lambda build

# Docker container validation
docker build -t go-api:test .                          # Build container
docker run --rm -p 8080:8080 go-api:test &            # Run container
curl -f http://localhost:8080/health || echo "Health check failed"
docker stop $(docker ps -q --filter ancestor=go-api:test)

# Expected: Clean builds, container runs successfully, health checks pass
# If failing: Fix build errors, resolve Docker issues, check port configuration
```

### Level 4: AWS Integration Testing

```bash
# AWS CLI and credentials validation
aws sts get-caller-identity                             # Verify AWS credentials
aws ecr describe-repositories --repository-names go-api 2>/dev/null || echo "ECR repo needed"

# Lambda deployment test (if configured)
sam build && sam local start-api &                     # Local Lambda testing
curl -f http://localhost:3000/api/v1/health || echo "Lambda local test failed"

# DynamoDB local testing (if available)
aws dynamodb list-tables --endpoint-url http://localhost:8000 2>/dev/null || echo "DynamoDB local not running"

# Expected: AWS integration works, Lambda functions locally, DynamoDB accessible
# If failing: Check AWS credentials, configure services, fix integration issues
```

### Level 5: Security and Performance Testing

```bash
# Security validation
gosec ./... | tee security-report.txt                  # Security scan
go list -json -m all | nancy sleuth                    # Dependency vulnerability scan

# API security testing (if server running)
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"test": "data"}' | grep -q "401\|403" && echo "Auth protection working"

# Performance baseline
go test -bench=. ./internal/api/handlers/...           # Benchmark API handlers
ab -n 1000 -c 10 http://localhost:8080/health || echo "Load test needs running server"

# Expected: No security issues, proper authentication, acceptable performance
# If failing: Fix security issues, implement proper auth, optimize performance
```

## Success Criteria

### Implementation Completeness
- [ ] Complete Go project structure following standards
- [ ] HTTP server implemented with proper routing and middleware
- [ ] JWT authentication integrated with AWS Cognito
- [ ] DynamoDB integration with proper error handling
- [ ] S3 file upload functionality implemented
- [ ] Lambda containerized deployment configured
- [ ] Comprehensive test suite with good coverage

### Code Quality
- [ ] All Go validation checks pass (go vet, golangci-lint, gosec)
- [ ] Test coverage above 80% with no race conditions
- [ ] Proper error handling and context propagation
- [ ] Input validation and security measures implemented
- [ ] Code follows Go idioms and best practices

### AWS Integration
- [ ] Lambda container builds and deploys successfully
- [ ] AWS SDK v2 integration working correctly
- [ ] IAM roles configured with least-privilege access
- [ ] Environment-based configuration working
- [ ] Logging and monitoring properly configured

### Production Readiness
- [ ] GitHub Actions CI/CD pipeline implemented
- [ ] Docker multi-stage build optimized for Lambda
- [ ] API documentation (OpenAPI/Swagger) complete
- [ ] Error handling comprehensive and user-friendly
- [ ] Security headers and CORS properly configured

Remember: The goal is production-ready, secure, and maintainable Go REST APIs that leverage AWS services effectively while following Go best practices and context engineering principles.