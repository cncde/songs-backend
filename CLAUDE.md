# Go REST API - Global Rules for Context Engineering

This file contains the global rules and principles for Go REST API development with AWS Lambda, JWT authentication, DynamoDB, and S3 integration. These rules apply to ALL Go API development work and should be followed consistently.

## 🔄 Go API Core Principles

**IMPORTANT: These principles apply to ALL Go API development work:**

### PRP Framework Workflow
- **Always start with INITIAL.md** - Define API requirements before generating PRPs
- **Use the PRP pattern**: INITIAL.md → `/generate-go-api-prp INITIAL.md` → `/execute-go-api-prp PRPs/filename.md`
- **Follow validation loops** - Each PRP must include executable Go validation steps
- **Context is King** - Include ALL necessary Go documentation, examples, and AWS patterns

### Go Development Methodology
- **Follow Go idioms** - Use effective Go patterns and conventions
- **Web search extensively** - Always research Go frameworks, AWS patterns, and best practices
- **Documentation deep dive** - Study official Go docs, AWS SDK docs, and framework patterns
- **Pattern extraction** - Identify reusable Go API patterns and AWS integration approaches
- **Gotcha documentation** - Document common Go and AWS pitfalls

## 📚 Go Project Awareness & Context

**Project Structure Patterns (from golang-standards/project-layout):**
- **`/cmd`** - Main application executables, each subdirectory matches executable name
- **`/internal`** - Private application code enforced by Go compiler (no external imports)
- **`/pkg`** - Publicly importable library code safe for external use
- **`/api`** - OpenAPI/Swagger specs, JSON schema files, protocol definitions
- **`/configs`** - Configuration file templates and default configurations

**Go Naming Conventions:**
- **MixedCaps** for exported functions and types
- **mixedCaps** for unexported functions and types
- **Package names** should be lowercase, single words when possible
- **Interface names** should be descriptive and end with 'er' when possible

## 🧱 Go Code Structure & Modularity

**Never create files longer than 500 lines** - Split into modules when approaching limit:
```
internal/
  api/
    handlers/     # HTTP request handlers
    middleware/   # Custom middleware functions
    models/       # Data models and structs
    services/     # Business logic services
  config/         # Configuration management
  db/            # Database layer abstractions
```

**Use Go modules consistently:**
- Always use `go mod init` for new projects
- Run `go mod tidy` regularly to clean dependencies
- Use semantic versioning for module releases
- Prefer relative imports within packages

**Follow Go concurrency patterns:**
- "Do not communicate by sharing memory; instead, share memory by communicating"
- Use goroutines for handling concurrent requests
- Implement proper context cancellation patterns
- Use channels for safe communication between components

## 🔧 Go Tooling & Development Workflow

**Required Go Commands:**
- **`go mod tidy`** - Clean up module dependencies
- **`go build ./...`** - Build all packages
- **`go test ./...`** - Run all tests
- **`go test -race ./...`** - Test with race detection
- **`go test -cover ./...`** - Test with coverage analysis
- **`go vet ./...`** - Static analysis for common errors
- **`golangci-lint run`** - Comprehensive linting
- **`gosec ./...`** - Security vulnerability scanning

**AWS Lambda Container Development:**
- **Build for Linux**: `GOOS=linux GOARCH=amd64 go build -o main cmd/api/main.go`
- **Docker build**: Multi-stage builds with Go 1.22+ base images
- **Local testing**: Use AWS Lambda Runtime Interface Emulator
- **Entry point**: Executable name is specified in ENTRYPOINT directive

**Development Environment:**
- Use `CGO_ENABLED=0` for most Lambda deployments
- Always test with both `linux/amd64` and `linux/arm64` architectures
- Use `go build -tags lambda.norpc` for optimized Lambda builds
- Include `.air.toml` for live reload during development

## 🧪 Testing & Reliability

**Always create comprehensive tests for Go APIs:**
- **Table-driven tests** for multiple test cases
- **HTTP handler testing** with `httptest` package
- **Mock interfaces** for external dependencies (AWS services)
- **Integration tests** for database and AWS service interactions

**Test Structure:**
```
tests/
  unit/           # Unit tests for individual functions
  integration/    # Integration tests with dependencies
  e2e/           # End-to-end API tests
```

**Required Test Types:**
- 1 test for expected API behavior
- 1 test for edge cases (empty inputs, boundary values)
- 1 test for failure cases (error conditions)
- 1 test for authentication/authorization scenarios

## ✅ Task Management

**Mark completed tasks immediately** after finishing them
**Update task status in real-time** as work progresses
**Break complex API features into smaller steps** with clear completion criteria

## 📎 Go API Style & Conventions

**Use Go as the primary language** with these patterns:
- **HTTP Router**: Gin framework for performance or Gorilla Mux for flexibility
- **JSON handling**: Use struct tags for JSON marshaling/unmarshaling
- **Error handling**: Always return explicit errors with context
- **Middleware patterns**: Chain middleware for authentication, logging, CORS
- **Context usage**: Pass `context.Context` through all functions for cancellation

**Write docstrings for every exported function:**
```go
// GetUser retrieves a user by ID from DynamoDB.
//
// It validates the user ID format, queries DynamoDB using the SDK v2,
// and returns the user data or an error if not found.
func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    // Implementation...
}
```

**AWS SDK v2 Patterns:**
- **Configuration**: Use `config.LoadDefaultConfig(ctx)` for AWS credentials
- **Service Clients**: Create service-specific clients (DynamoDB, S3)
- **Error Handling**: Check for AWS-specific error types
- **Context**: Always pass context to AWS SDK operations

## 🔐 Security & AWS Best Practices

**JWT Authentication with AWS Cognito:**
- **Validate JWT structure** - Verify header.payload.signature format
- **Download JWKS** from Cognito user pool and cache public keys
- **Verify signature** using RSA256 with public keys
- **Check claims**: exp (expiration), aud (audience), iss (issuer), token_use
- **Use libraries**: `golang-jwt/jwt/v5` and `MicahParks/keyfunc` for JWKS

**Input Validation Patterns:**
```go
type CreateUserRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Username string `json:"username" binding:"required,min=3,max=20,alphanum"`
    Age      int    `json:"age" binding:"min=13,max=120"`
}
```

**AWS Security:**
- **IAM Roles**: Use least-privilege IAM roles for Lambda execution
- **Environment Variables**: Store sensitive configuration in environment variables
- **VPC**: Configure VPC settings for DynamoDB and S3 access when needed
- **Encryption**: Use AWS KMS for encrypting sensitive data

**API Security:**
- **Rate limiting**: Implement request rate limiting
- **CORS**: Configure appropriate CORS policies
- **Input sanitization**: Validate and sanitize all user inputs
- **SQL injection prevention**: Use parameterized queries (not applicable to DynamoDB)

## 📚 AWS Integration Patterns

**DynamoDB with Go SDK v2:**
```go
// Service initialization
cfg, err := config.LoadDefaultConfig(ctx)
client := dynamodb.NewFromConfig(cfg)

// Query pattern
result, err := client.Query(ctx, &dynamodb.QueryInput{
    TableName: aws.String("users"),
    KeyConditionExpression: aws.String("id = :id"),
    ExpressionAttributeValues: map[string]types.AttributeValue{
        ":id": &types.AttributeValueMemberS{Value: userID},
    },
})
```

**S3 File Upload Pattern:**
```go
// Upload with metadata
_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
    Bucket: aws.String(bucketName),
    Key:    aws.String(objectKey),
    Body:   file,
    ContentType: aws.String("image/jpeg"),
    Metadata: map[string]string{
        "uploaded-by": userID,
        "upload-time": time.Now().Format(time.RFC3339),
    },
})
```

## 🚫 Go API Anti-Patterns to Always Avoid

- ❌ Don't ignore errors - Always handle Go errors explicitly
- ❌ Don't use `panic()` for normal error conditions - Return errors instead  
- ❌ Don't share memory between goroutines - Use channels for communication
- ❌ Don't forget to close resources - Use `defer` for cleanup
- ❌ Don't ignore context cancellation - Always respect context.Context
- ❌ Don't hardcode AWS credentials - Use IAM roles and environment config
- ❌ Don't skip input validation - Always validate user inputs
- ❌ Don't forget JWT validation - Always verify tokens before processing requests
- ❌ Don't ignore AWS service limits - Implement proper error handling and retries
- ❌ Don't skip comprehensive testing - Include unit, integration, and security tests

## 🔧 GitHub Actions & CI/CD

**Standard Go GitHub Actions Workflow:**
```yaml
name: Go API CI/CD

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.22'
    - run: go mod download
    - run: go vet ./...
    - run: golangci-lint run
    - run: go test -race -cover ./...
    - run: gosec ./...
    
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - run: docker build -t api:latest .
    - run: docker run --rm api:latest
```

**Lambda Deployment Integration:**
- Build Docker images in CI/CD pipeline
- Push to Amazon ECR (Elastic Container Registry)
- Update Lambda function with new image URI
- Run post-deployment smoke tests

These global rules ensure that all Go API development follows best practices, maintains security, and integrates properly with AWS services while leveraging the context engineering framework effectively.