---
name: "Go REST API PRP"
description: "Production-ready Go REST API with AWS Lambda, JWT authentication, DynamoDB, and S3 integration"
---

## Purpose

Develop a secure, scalable Go REST API with containerized AWS Lambda deployment, JWT authentication via AWS Cognito, DynamoDB data persistence, and S3 file handling capabilities.

## Core Principles

1. **Go Idiomatic Code**: Follow effective Go patterns and conventions
2. **Security First**: Implement comprehensive authentication, authorization, and input validation
3. **AWS Native**: Leverage AWS services with proper integration patterns
4. **Production Ready**: Include comprehensive testing, monitoring, and deployment automation
5. **Performance Optimized**: Design for scalability and efficient resource utilization

---

## Goal

**[DESCRIBE YOUR SPECIFIC GO API GOAL HERE]**

Example: Build a user management API that handles user registration, profile management, and file uploads with secure authentication and efficient data storage.

## Why

**[EXPLAIN THE BUSINESS VALUE AND TECHNICAL REQUIREMENTS]**

Example:
- **Scalability**: Handle variable load with serverless architecture
- **Security**: Protect user data with enterprise-grade authentication
- **Cost Efficiency**: Pay only for actual usage with Lambda pricing
- **Developer Experience**: Provide clean, well-documented APIs
- **Maintainability**: Use Go's strong typing and testing capabilities

## What

### Technical Requirements

**API Capabilities:**
- [ ] RESTful API design following OpenAPI standards
- [ ] JWT authentication with AWS Cognito integration
- [ ] CRUD operations with DynamoDB backend
- [ ] File upload/download with S3 storage
- [ ] Input validation and error handling
- [ ] Rate limiting and security headers
- [ ] Comprehensive logging and monitoring

**AWS Integration:**
- [ ] Lambda containerized deployment
- [ ] DynamoDB table design and operations
- [ ] S3 bucket configuration and access
- [ ] IAM roles with least-privilege permissions
- [ ] VPC configuration (if required)
- [ ] CloudWatch logging and monitoring

**Development Workflow:**
- [ ] Go modules dependency management
- [ ] Comprehensive test suite (unit + integration)
- [ ] GitHub Actions CI/CD pipeline
- [ ] Docker multi-stage builds
- [ ] Local development environment
- [ ] API documentation generation

## All Needed Context

### Go Framework Research (WEB SEARCH REQUIRED)

**HTTP Framework Selection:**
```yaml
Research Areas:
- Gin Framework: Performance benchmarks, middleware patterns, JSON handling
- Echo Framework: Feature comparison, routing capabilities, middleware ecosystem
- Gorilla Mux: Flexibility analysis, subrouting patterns, compatibility
- Standard net/http: When to use, performance characteristics, middleware patterns

Web Search Queries:
- "Go REST API framework comparison 2024"
- "Gin vs Echo vs Gorilla Mux performance benchmarks"
- "Go HTTP middleware patterns and best practices"
- "Production Go API framework recommendations"
```

**Project Structure Patterns:**
```yaml
Standard Go Layout Research:
- /cmd: Application executables and entry points
- /internal: Private application code (compiler-enforced)
- /pkg: Public library code safe for external use
- /api: OpenAPI/Swagger specifications
- /configs: Configuration templates and defaults

Additional Directories:
- /build: CI/CD and packaging configurations
- /deployments: Infrastructure deployment configs
- /scripts: Build and utility scripts
- /test: Additional test data and test applications
```

### AWS Integration Patterns (WEB SEARCH REQUIRED)

**Lambda Container Deployment:**
```yaml
Research Requirements:
- Multi-stage Dockerfile patterns for Go Lambda
- AWS Lambda Runtime Interface Emulator usage
- Container optimization techniques for Lambda
- Cold start optimization strategies
- Lambda function configuration best practices

Implementation Patterns:
- Handler function design with context management
- Environment-based configuration loading
- Logging integration with CloudWatch
- Error handling and recovery patterns
```

**DynamoDB Integration:**
```yaml
AWS SDK for Go v2 Patterns:
- Client initialization with context and configuration
- Query and Scan operations with pagination
- Batch operations for bulk data processing
- Error handling and retry logic
- Attribute value marshaling/unmarshaling

Data Modeling:
- Primary key design patterns
- Global Secondary Index strategies
- Item size optimization
- Query pattern optimization
```

**S3 File Handling:**
```yaml
Upload Patterns:
- Single-part uploads for small files
- Multipart uploads for large files
- Presigned URL generation for direct uploads
- File type validation and security scanning
- Metadata management and retrieval

Security Considerations:
- Bucket policy configuration
- CORS settings for web uploads
- File size and type restrictions
- Virus scanning integration
- Access logging and monitoring
```

### Authentication & Security (WEB SEARCH REQUIRED)

**AWS Cognito JWT Integration:**
```yaml
Implementation Requirements:
- JWKS (JSON Web Key Set) retrieval and caching
- JWT signature verification with RS256
- Claims validation (exp, aud, iss, token_use)
- Token refresh handling
- Error scenarios and fallback

Go Libraries:
- github.com/golang-jwt/jwt/v5: Core JWT handling
- github.com/MicahParks/keyfunc: JWKS integration
- Custom middleware implementation patterns
```

**Security Headers and Validation:**
```yaml
Input Validation:
- Request body validation with struct tags
- Query parameter sanitization
- Path parameter validation
- File upload restrictions

Security Headers:
- CORS configuration for cross-origin requests
- Security headers (HSTS, Content-Security-Policy, etc.)
- Rate limiting implementation
- Request ID tracking for debugging
```

### Testing Strategy (GO-SPECIFIC PATTERNS)

**Unit Testing Patterns:**
```yaml
Table-Driven Tests:
- Multiple test cases in single test function
- Input/output validation patterns
- Error condition testing
- Mock interface implementations

HTTP Handler Testing:
- httptest package usage for request/response testing
- Test server setup and teardown
- Authentication middleware testing
- Error response validation
```

**Integration Testing:**
```yaml
AWS Service Integration:
- DynamoDB Local for local testing
- S3 mock implementations
- Lambda local testing with SAM CLI
- End-to-end API testing patterns
```

## Implementation Blueprint

### Phase 1: Project Foundation
```yaml
Setup Tasks:
1. Initialize Go module with proper dependencies
2. Create standard Go project structure (cmd/, internal/, pkg/)
3. Set up basic HTTP server with chosen framework
4. Implement basic middleware (logging, CORS, recovery)
5. Create configuration management system

Dependencies to Add:
- HTTP framework (gin-gonic/gin or labstack/echo)
- AWS SDK v2 (github.com/aws/aws-sdk-go-v2)
- JWT libraries (golang-jwt/jwt and MicahParks/keyfunc)
- Testing libraries (testify for assertions)
- Configuration management (viper or similar)
```

### Phase 2: AWS Integration
```yaml
AWS Setup:
1. Configure AWS SDK v2 clients (DynamoDB, S3)
2. Implement DynamoDB repository layer with error handling
3. Create S3 service layer with upload/download capabilities
4. Set up IAM roles and policies for Lambda execution
5. Implement AWS-specific error handling patterns

Infrastructure:
- DynamoDB table creation and configuration
- S3 bucket setup with proper CORS and policies
- Lambda function configuration for container deployment
- CloudWatch log group setup for monitoring
```

### Phase 3: Authentication & Security
```yaml
Security Implementation:
1. JWT middleware with AWS Cognito integration
2. Input validation with proper error responses
3. Rate limiting middleware implementation
4. Security headers and CORS configuration
5. File upload security validation

Authentication Flow:
- Token extraction from Authorization header
- JWKS retrieval and caching mechanism
- Claims validation and user context setup
- Error handling for invalid/expired tokens
```

### Phase 4: API Implementation
```yaml
Core API Features:
1. User management endpoints (CRUD operations)
2. File upload/download endpoints with S3 integration
3. Health check and status endpoints
4. Comprehensive error handling and logging
5. API versioning strategy implementation

Response Patterns:
- Consistent JSON response structure
- Proper HTTP status codes
- Error message standardization
- Pagination for list endpoints
```

### Phase 5: Testing & Quality
```yaml
Testing Implementation:
1. Unit tests for all service functions
2. HTTP handler tests with httptest package
3. Integration tests for AWS service interactions
4. Mock implementations for external dependencies
5. Performance benchmarks for critical paths

Quality Assurance:
- Go linting with golangci-lint
- Security scanning with gosec
- Test coverage analysis and reporting
- Code review checklist implementation
```

### Phase 6: Deployment & Monitoring
```yaml
CI/CD Pipeline:
1. GitHub Actions workflow for testing and building
2. Docker multi-stage build for Lambda deployment
3. ECR image pushing and Lambda function updates
4. Post-deployment smoke testing
5. Rollback procedures for failed deployments

Monitoring Setup:
- CloudWatch metrics and alarms
- Structured logging with correlation IDs
- Error tracking and alerting
- Performance monitoring dashboards
```

## Validation Loop

### Level 1: Go Code Quality
```bash
# Dependencies and module validation
go mod tidy && go mod verify                    # Clean and verify dependencies
go mod download                                 # Download all dependencies

# Static analysis and linting
go vet ./...                                   # Built-in static analysis
golangci-lint run                              # Comprehensive linting
gosec ./...                                    # Security vulnerability scan

# Build validation
go build ./...                                 # Build all packages
GOOS=linux GOARCH=amd64 go build -o bootstrap cmd/api/main.go  # Lambda build
```

### Level 2: Testing Validation
```bash
# Test execution
go test ./...                                  # Run all tests
go test -race ./...                           # Race condition detection
go test -cover ./...                          # Coverage analysis
go test -v ./internal/api/handlers/...        # Verbose handler testing

# Benchmark testing
go test -bench=. ./internal/api/...           # Performance benchmarks
```

### Level 3: Security Validation
```bash
# Security scanning
gosec ./... | tee security-report.txt         # Security vulnerability report
go list -json -m all | nancy sleuth           # Dependency vulnerability scan

# Authentication testing
curl -X GET http://localhost:8080/api/v1/protected \
  -H "Authorization: Bearer invalid-token" \
  | grep -q "401" && echo "Auth protection working"
```

### Level 4: AWS Integration Testing
```bash
# AWS credentials and service validation
aws sts get-caller-identity                    # Verify AWS credentials
aws dynamodb describe-table --table-name users 2>/dev/null || echo "Table setup needed"
aws s3 ls s3://your-bucket 2>/dev/null || echo "S3 bucket setup needed"

# Lambda local testing
sam build && sam local start-api &            # Start local Lambda
curl -f http://localhost:3000/health          # Health check
```

### Level 5: Container and Deployment
```bash
# Docker build validation
docker build -t go-api:test .                 # Build container image
docker run --rm -d -p 8080:8080 --name api-test go-api:test  # Run container
sleep 5 && curl -f http://localhost:8080/health  # Test running container
docker stop api-test && docker rm api-test    # Cleanup

# Deployment readiness
test -f Dockerfile && echo "✓ Dockerfile present"
test -f .github/workflows/deploy.yml && echo "✓ GitHub Actions configured"
```

## Success Criteria

### Functional Requirements
- [ ] All API endpoints implemented and tested
- [ ] JWT authentication working with AWS Cognito
- [ ] DynamoDB operations functioning correctly
- [ ] S3 file upload/download implemented
- [ ] Error handling comprehensive and user-friendly
- [ ] API documentation complete and accurate

### Quality Requirements
- [ ] Test coverage above 80%
- [ ] All linting checks pass
- [ ] Security scan shows no critical issues
- [ ] Performance benchmarks meet requirements
- [ ] Code follows Go idioms and best practices
- [ ] Documentation is clear and complete

### Deployment Requirements
- [ ] Docker container builds successfully
- [ ] Lambda deployment works without errors
- [ ] CI/CD pipeline executes successfully
- [ ] Monitoring and logging configured
- [ ] Rollback procedures tested
- [ ] Post-deployment smoke tests pass

### Production Readiness
- [ ] Environment-specific configuration working
- [ ] Security headers and CORS configured
- [ ] Rate limiting implemented and tested
- [ ] Error monitoring and alerting set up
- [ ] Backup and recovery procedures documented
- [ ] Performance under load validated

---

## Common Patterns & Examples

### HTTP Handler Pattern
```go
func (h *UserHandler) GetUser(c *gin.Context) {
    userID := c.Param("id")
    
    // Input validation
    if userID == "" {
        c.JSON(400, gin.H{"error": "user ID required"})
        return
    }
    
    // Business logic
    user, err := h.userService.GetUser(c.Request.Context(), userID)
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            c.JSON(404, gin.H{"error": "user not found"})
            return
        }
        log.Printf("Error getting user %s: %v", userID, err)
        c.JSON(500, gin.H{"error": "internal server error"})
        return
    }
    
    c.JSON(200, gin.H{"user": user})
}
```

### DynamoDB Repository Pattern
```go
func (r *UserRepository) GetUser(ctx context.Context, id string) (*models.User, error) {
    result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
        TableName: aws.String(r.table),
        Key: map[string]types.AttributeValue{
            "id": &types.AttributeValueMemberS{Value: id},
        },
    })
    
    if err != nil {
        return nil, fmt.Errorf("failed to get user from DynamoDB: %w", err)
    }
    
    if result.Item == nil {
        return nil, ErrUserNotFound
    }
    
    var user models.User
    if err := attributevalue.UnmarshalMap(result.Item, &user); err != nil {
        return nil, fmt.Errorf("failed to unmarshal user: %w", err)
    }
    
    return &user, nil
}
```

### S3 Upload Pattern
```go
func (s *FileService) UploadFile(ctx context.Context, file io.Reader, key, contentType string) error {
    _, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
        Bucket:      aws.String(s.bucket),
        Key:         aws.String(key),
        Body:        file,
        ContentType: aws.String(contentType),
        Metadata: map[string]string{
            "uploaded-at": time.Now().UTC().Format(time.RFC3339),
        },
    })
    
    if err != nil {
        return fmt.Errorf("failed to upload file to S3: %w", err)
    }
    
    return nil
}
```

### Table-Driven Test Pattern
```go
func TestUserValidation(t *testing.T) {
    tests := []struct {
        name        string
        input       models.User
        expectedErr error
    }{
        {
            name: "valid user",
            input: models.User{
                Email:    "test@example.com",
                Username: "testuser",
                Age:      25,
            },
            expectedErr: nil,
        },
        {
            name: "invalid email",
            input: models.User{
                Email:    "invalid-email",
                Username: "testuser",
                Age:      25,
            },
            expectedErr: ErrInvalidEmail,
        },
        // Add more test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateUser(tt.input)
            if !errors.Is(err, tt.expectedErr) {
                t.Errorf("expected error %v, got %v", tt.expectedErr, err)
            }
        })
    }
}
```

## Gotchas & Common Issues

### Go-Specific Gotchas
- **Error Handling**: Always check and handle errors explicitly
- **Goroutine Leaks**: Ensure proper cleanup of goroutines
- **Context Cancellation**: Respect context cancellation in long-running operations
- **Interface Usage**: Keep interfaces small and focused
- **Memory Management**: Be aware of memory allocation in hot paths

### AWS Lambda Gotchas
- **Cold Starts**: Optimize initialization code and container size
- **Timeout Limits**: Design for Lambda's maximum execution time
- **Memory Configuration**: Right-size memory allocation for cost and performance
- **Environment Variables**: Use environment variables for configuration
- **Concurrent Execution**: Understand Lambda's concurrent execution limits

### DynamoDB Gotchas
- **Hot Partitions**: Design keys to distribute load evenly
- **Query Limitations**: Understand query vs scan performance implications
- **Item Size Limits**: Keep items under 400KB limit
- **Eventually Consistent Reads**: Use strongly consistent reads when needed
- **Throttling**: Implement exponential backoff for throttled requests

### Security Gotchas
- **JWT Validation**: Always validate all claims, not just signature
- **Input Sanitization**: Validate all inputs, including file uploads
- **CORS Configuration**: Configure CORS properly for web applications
- **Error Information**: Don't leak sensitive information in error messages
- **Rate Limiting**: Implement rate limiting to prevent abuse