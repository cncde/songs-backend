---
name: "Go Songs API PRP"
description: "Production-ready Go REST API for songs and comments with AWS Lambda, JWT authentication, DynamoDB, and S3 integration"
---

## Purpose

Develop a secure, scalable Go REST API for managing songs and comments with containerized AWS Lambda deployment, JWT authentication via AWS Cognito, DynamoDB data persistence, and S3 file handling for audio and image uploads.

## Core Principles

1. **Go Idiomatic Code**: Follow effective Go patterns and Mat Ryer's HTTP service best practices
2. **Security First**: Implement comprehensive authentication, authorization, and input validation
3. **AWS Native**: Leverage AWS services with proper integration patterns
4. **Production Ready**: Include comprehensive testing, monitoring, and deployment automation
5. **Performance Optimized**: Design for scalability and efficient Lambda cold starts

---

## Goal

Build a songs management API that handles:
- **Song Management**: CRUD operations for songs with metadata (title, subtitle, description, lyrics, stage, tags)
- **Media Storage**: Image and audio file uploads via S3 with secure URLs
- **Comment System**: Comments associated with specific audio files
- **Authentication**: JWT-based authentication with AWS Cognito integration

## Why

**Business Value:**
- **Content Management**: Centralized system for managing song content and metadata
- **User Engagement**: Comment system enables user interaction and feedback
- **Media Organization**: Structured storage and retrieval of audio/image assets
- **Scalability**: Serverless architecture handles variable traffic efficiently
- **Cost Efficiency**: Pay-per-use Lambda pricing model

**Technical Requirements:**
- **Data Integrity**: Strong typing and validation prevent data corruption
- **Security**: Enterprise-grade authentication protects content
- **Performance**: Fast response times with optimized queries
- **Maintainability**: Clean Go architecture enables easy modifications

## What

### Technical Requirements

**API Capabilities:**
- [ ] RESTful song management endpoints (GET, POST, PUT, DELETE)
- [ ] Comment CRUD operations linked to specific audio files
- [ ] JWT authentication with AWS Cognito integration
- [ ] File upload/download for images and audio with S3
- [ ] Input validation and comprehensive error handling
- [ ] Rate limiting and security headers
- [ ] Health check and metrics endpoints

**Data Models:**
```go
type Song struct {
    ID          string   `json:"id" dynamodbav:"id"`
    Title       string   `json:"title" dynamodbav:"title"`
    Subtitle    string   `json:"subtitle" dynamodbav:"subtitle"`
    Description string   `json:"description" dynamodbav:"description"`
    ImageURLs   []string `json:"image_urls" dynamodbav:"image_urls"`
    AudioURLs   []string `json:"audio_urls" dynamodbav:"audio_urls"`
    Lyrics      string   `json:"lyrics" dynamodbav:"lyrics"`
    Stage       string   `json:"stage" dynamodbav:"stage"`
    Tags        []string `json:"tags" dynamodbav:"tags"`
    CreatedAt   string   `json:"created_at" dynamodbav:"created_at"`
    UpdatedAt   string   `json:"updated_at" dynamodbav:"updated_at"`
}

type Comment struct {
    ID        string `json:"id" dynamodbav:"id"`
    SongID    string `json:"song_id" dynamodbav:"song_id"`
    AudioURL  string `json:"audio_url" dynamodbav:"audio_url"`
    Content   string `json:"content" dynamodbav:"content"`
    UserID    string `json:"user_id" dynamodbav:"user_id"`
    CreatedAt string `json:"created_at" dynamodbav:"created_at"`
    UpdatedAt string `json:"updated_at" dynamodbav:"updated_at"`
}
```

**AWS Integration:**
- [ ] Lambda containerized deployment with optimized cold starts
- [ ] DynamoDB tables: songs, comments with proper indexing
- [ ] S3 bucket for media storage with presigned URLs
- [ ] IAM roles with least-privilege permissions
- [ ] CloudWatch logging and monitoring
- [ ] Terraform infrastructure as code

**Development Workflow:**
- [ ] Go modules dependency management
- [ ] Table-driven test suite (unit + integration)
- [ ] GitHub Actions CI/CD pipeline
- [ ] Docker multi-stage builds for Lambda
- [ ] Local development with Lambda emulator

## All Needed Context

### Go Framework Research Results

**Selected Framework: Gin**
Based on research, Gin provides:
- High performance (up to 40x faster than alternatives)
- Zero allocation router
- Built-in middleware support
- JSON binding and validation
- Extensive community support

**Key Gin Patterns:**
```go
r := gin.Default()
r.Use(middleware.CORS())
r.Use(middleware.JWTAuth())

v1 := r.Group("/api/v1")
{
    songs := v1.Group("/songs")
    songs.GET("", h.ListSongs)
    songs.POST("", h.CreateSong)
    songs.GET("/:id", h.GetSong)
    songs.PUT("/:id", h.UpdateSong)
    songs.DELETE("/:id", h.DeleteSong)
    
    songs.POST("/:id/comments", h.CreateComment)
    songs.GET("/:id/comments", h.ListComments)
}
```

**Mat Ryer's Best Practices Applied:**
- Constructor pattern: `NewServer(deps Dependencies)`
- Centralized routes.go mapping entire API surface
- Handler maker functions for dependency injection
- End-to-end testing priority over granular unit tests

### AWS Integration Patterns

**Lambda Container Deployment:**
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o main cmd/api/main.go

FROM public.ecr.aws/lambda/provided:al2023
COPY --from=builder /app/main ./main
ENTRYPOINT [ "./main" ]
```

**Lambda Handler Pattern:**
```go
func main() {
    cfg, err := config.LoadDefaultConfig(context.TODO())
    if err != nil {
        log.Fatal("failed to load AWS config:", err)
    }
    
    deps := &Dependencies{
        DynamoDB: dynamodb.NewFromConfig(cfg),
        S3:       s3.NewFromConfig(cfg),
    }
    
    server := NewServer(deps)
    lambda.Start(server.Handler())
}
```

**AWS SDK v2 Patterns:**
```go
// Client initialization
cfg, err := config.LoadDefaultConfig(ctx, 
    config.WithRegion("us-west-2"))
dynamoClient := dynamodb.NewFromConfig(cfg)
s3Client := s3.NewFromConfig(cfg)

// DynamoDB operation
result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
    TableName: aws.String("songs"),
    Key: map[string]types.AttributeValue{
        "id": &types.AttributeValueMemberS{Value: songID},
    },
})
```

### Authentication & Security

**JWT Cognito Integration:**
Research shows using these libraries:
- `github.com/golang-jwt/jwt/v5` for JWT parsing
- `github.com/MicahParks/keyfunc` for JWKS integration

**JWT Middleware Pattern:**
```go
func JWTMiddleware(jwks *keyfunc.JWKS) gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        tokenString := extractToken(c.GetHeader("Authorization"))
        
        token, err := jwt.Parse(tokenString, jwks.Keyfunc)
        if err != nil || !token.Valid {
            c.JSON(401, gin.H{"error": "unauthorized"})
            c.Abort()
            return
        }
        
        claims := token.Claims.(jwt.MapClaims)
        c.Set("user_id", claims["sub"])
        c.Next()
    })
}
```

**JWKS Setup:**
- URL: `https://cognito-idp.<region>.amazonaws.com/<user_pool_id>/.well-known/jwks.json`
- Validate: `exp`, `aud`, `iss`, `token_use` claims
- Cache public keys with periodic refresh

### Go Code Quality Standards

**From Go Code Review Comments:**
- Use short, descriptive variable names
- Handle errors explicitly, never ignore with `_`
- Return concrete types, accept interfaces
- Use `gofmt`/`goimports` for consistent formatting
- Write clear documentation comments

**Error Handling Pattern:**
```go
func (s *SongService) GetSong(ctx context.Context, id string) (*Song, error) {
    if id == "" {
        return nil, errors.New("song id required")
    }
    
    song, err := s.repo.GetSong(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get song %s: %w", id, err)
    }
    
    return song, nil
}
```

## Implementation Blueprint

### Task List (One Task Per File)

**Phase 1: Project Foundation**
1. **cmd/api/main.go** - Lambda entry point with dependency injection
2. **internal/config/config.go** - Environment configuration management
3. **internal/models/song.go** - Song data model and validation
4. **internal/models/comment.go** - Comment data model and validation
5. **go.mod** - Module initialization with required dependencies

**Phase 2: AWS Integration Layer**
6. **internal/repository/dynamodb.go** - DynamoDB client wrapper
7. **internal/repository/songs.go** - Song repository with CRUD operations
8. **internal/repository/comments.go** - Comment repository operations
9. **internal/storage/s3.go** - S3 service for file operations
10. **internal/middleware/auth.go** - JWT authentication middleware

**Phase 3: Business Logic Layer**
11. **internal/service/songs.go** - Song business logic service
12. **internal/service/comments.go** - Comment business logic service
13. **internal/service/media.go** - Media upload/download service
14. **internal/handlers/songs.go** - Song HTTP handlers
15. **internal/handlers/comments.go** - Comment HTTP handlers

**Phase 4: HTTP Server Setup**
16. **internal/server/server.go** - Server constructor with dependencies
17. **internal/server/routes.go** - Route definitions and middleware setup
18. **internal/server/middleware.go** - Custom middleware implementations
19. **internal/handlers/health.go** - Health check endpoint
20. **internal/handlers/media.go** - File upload/download handlers

**Phase 5: Testing Infrastructure**
21. **tests/integration/songs_test.go** - Song endpoint integration tests
22. **tests/integration/comments_test.go** - Comment endpoint tests
23. **tests/unit/services_test.go** - Service layer unit tests
24. **tests/mocks/repository.go** - Mock interfaces for testing
25. **tests/fixtures/data.go** - Test data fixtures

**Phase 6: Deployment & Infrastructure**
26. **Dockerfile** - Multi-stage container build for Lambda
27. **terraform/main.tf** - DynamoDB and S3 infrastructure
28. **terraform/iam.tf** - IAM roles and policies
29. **.github/workflows/deploy.yml** - CI/CD pipeline
30. **docker-compose.yml** - Local development environment

### Dependencies Required
```go
// Core framework
github.com/gin-gonic/gin v1.10.0

// AWS SDK v2
github.com/aws/aws-sdk-go-v2/config
github.com/aws/aws-sdk-go-v2/service/dynamodb
github.com/aws/aws-sdk-go-v2/service/s3
github.com/aws/aws-lambda-go/lambda
github.com/aws/aws-lambda-go/events

// Authentication
github.com/golang-jwt/jwt/v5
github.com/MicahParks/keyfunc

// Utilities
github.com/google/uuid
github.com/stretchr/testify

// Validation
github.com/go-playground/validator/v10
```

### Project Structure
```
songs-backend/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── config/
│   ├── models/
│   ├── repository/
│   ├── service/
│   ├── handlers/
│   ├── server/
│   └── middleware/
├── tests/
│   ├── integration/
│   ├── unit/
│   ├── mocks/
│   └── fixtures/
├── terraform/
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

## Validation Loop

### Level 1: Go Code Quality
```bash
# Module management
go mod tidy && go mod verify
go mod download

# Static analysis
go vet ./...
golangci-lint run
gosec ./...

# Build validation
go build ./...
GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o main cmd/api/main.go
```

### Level 2: Testing Validation
```bash
# Test execution
go test ./...
go test -race ./...
go test -cover ./...
go test -v ./tests/integration/...

# Benchmark testing
go test -bench=. ./internal/service/...
```

### Level 3: Security Validation
```bash
# Security scanning
gosec ./... | tee security-report.txt
go list -json -m all | nancy sleuth

# JWT validation test
curl -X GET http://localhost:8080/api/v1/songs \
  -H "Authorization: Bearer invalid-token" \
  | grep -q "401" && echo "✓ Auth protection working"
```

### Level 4: AWS Integration Testing
```bash
# AWS service validation
aws sts get-caller-identity
aws dynamodb describe-table --table-name songs 2>/dev/null || echo "Table setup needed"
aws s3 ls s3://songs-media-bucket 2>/dev/null || echo "S3 bucket setup needed"

# Lambda local testing
sam build && sam local start-api &
sleep 5 && curl -f http://localhost:3000/health
```

### Level 5: Container and Deployment
```bash
# Docker build validation
docker build -t songs-api:test .
docker run --rm -d -p 8080:8080 --name api-test songs-api:test
sleep 5 && curl -f http://localhost:8080/health
docker stop api-test && docker rm api-test

# Infrastructure validation
terraform -chdir=terraform init
terraform -chdir=terraform plan
terraform -chdir=terraform validate
```

## Success Criteria

### Functional Requirements
- [ ] All song CRUD operations implemented and tested
- [ ] Comment system fully functional with audio file association
- [ ] JWT authentication working with AWS Cognito
- [ ] S3 file upload/download for images and audio
- [ ] DynamoDB operations optimized with proper indexing
- [ ] Error handling comprehensive and user-friendly

### Quality Requirements
- [ ] Test coverage above 80%
- [ ] All linting checks pass (golangci-lint, go vet, gosec)
- [ ] Security scan shows no critical vulnerabilities
- [ ] Performance benchmarks meet <200ms response time
- [ ] Code follows Go idioms and Mat Ryer's patterns
- [ ] All functions have clear documentation

### Production Readiness
- [ ] Docker container builds and runs successfully
- [ ] Lambda deployment works without errors
- [ ] CI/CD pipeline executes end-to-end
- [ ] Terraform infrastructure deploys correctly
- [ ] Health check endpoint returns proper status
- [ ] CloudWatch logging captures all important events

---

## API Endpoints Design

### Songs Endpoints
```
GET    /api/v1/songs              # List songs with pagination
POST   /api/v1/songs              # Create new song
GET    /api/v1/songs/:id          # Get specific song
PUT    /api/v1/songs/:id          # Update song
DELETE /api/v1/songs/:id          # Delete song
POST   /api/v1/songs/:id/media    # Upload media files
```

### Comments Endpoints
```
GET    /api/v1/songs/:id/comments # List comments for song
POST   /api/v1/songs/:id/comments # Create comment
GET    /api/v1/comments/:id       # Get specific comment
PUT    /api/v1/comments/:id       # Update comment
DELETE /api/v1/comments/:id       # Delete comment
```

### Utility Endpoints
```
GET    /health                    # Health check
GET    /metrics                   # Application metrics
```

## Common Patterns & Examples

### HTTP Handler Pattern (Mat Ryer Style)
```go
func (h *SongHandler) CreateSong() gin.HandlerFunc {
    return func(c *gin.Context) {
        var req CreateSongRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        
        userID := c.GetString("user_id")
        song, err := h.service.CreateSong(c.Request.Context(), userID, req)
        if err != nil {
            log.Printf("Error creating song: %v", err)
            c.JSON(500, gin.H{"error": "internal server error"})
            return
        }
        
        c.JSON(201, gin.H{"song": song})
    }
}
```

### DynamoDB Repository Pattern
```go
func (r *SongRepository) CreateSong(ctx context.Context, song *models.Song) error {
    item, err := attributevalue.MarshalMap(song)
    if err != nil {
        return fmt.Errorf("failed to marshal song: %w", err)
    }
    
    _, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
        TableName: aws.String(r.tableName),
        Item:      item,
        ConditionExpression: aws.String("attribute_not_exists(id)"),
    })
    
    if err != nil {
        return fmt.Errorf("failed to create song: %w", err)
    }
    
    return nil
}
```

### S3 Upload Pattern with Presigned URLs
```go
func (s *MediaService) GetUploadURL(ctx context.Context, filename, contentType string) (string, error) {
    key := fmt.Sprintf("uploads/%s/%s", time.Now().Format("2006/01/02"), filename)
    
    presignClient := s3.NewPresignClient(s.s3Client)
    request, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
        Bucket:      aws.String(s.bucket),
        Key:         aws.String(key),
        ContentType: aws.String(contentType),
    }, func(opts *s3.PresignOptions) {
        opts.Expires = time.Duration(15 * time.Minute)
    })
    
    if err != nil {
        return "", fmt.Errorf("failed to generate presigned URL: %w", err)
    }
    
    return request.URL, nil
}
```

### Table-Driven Test Pattern
```go
func TestCreateSong(t *testing.T) {
    tests := []struct {
        name        string
        input       CreateSongRequest
        setupMock   func(*mocks.SongRepository)
        expectedErr error
        expectedCode int
    }{
        {
            name: "valid song creation",
            input: CreateSongRequest{
                Title:    "Test Song",
                Subtitle: "Test Subtitle",
                Tags:     []string{"rock", "indie"},
            },
            setupMock: func(m *mocks.SongRepository) {
                m.On("CreateSong", mock.Anything, mock.AnythingOfType("*models.Song")).Return(nil)
            },
            expectedErr: nil,
            expectedCode: 201,
        },
        {
            name: "invalid input - empty title",
            input: CreateSongRequest{
                Title: "",
            },
            setupMock:   func(m *mocks.SongRepository) {},
            expectedErr: nil,
            expectedCode: 400,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := new(mocks.SongRepository)
            tt.setupMock(mockRepo)
            
            service := NewSongService(mockRepo)
            handler := NewSongHandler(service)
            
            router := gin.New()
            router.POST("/songs", handler.CreateSong())
            
            body, _ := json.Marshal(tt.input)
            req := httptest.NewRequest("POST", "/songs", bytes.NewBuffer(body))
            req.Header.Set("Content-Type", "application/json")
            w := httptest.NewRecorder()
            
            router.ServeHTTP(w, req)
            
            assert.Equal(t, tt.expectedCode, w.Code)
            mockRepo.AssertExpectations(t)
        })
    }
}
```

## Gotchas & Common Issues

### Go-Specific Gotchas
- **Error Handling**: Always check and handle errors explicitly - never use `_`
- **Context Propagation**: Pass context through all layers for cancellation
- **Goroutine Leaks**: Ensure proper cleanup, especially in HTTP handlers
- **Interface Design**: Keep interfaces small and define them where used
- **Pointer vs Value**: Be explicit about receiver types and memory usage

### AWS Lambda Gotchas
- **Cold Starts**: Minimize initialization time by reusing connections
- **Memory Limits**: Right-size memory allocation (affects CPU allocation)
- **Timeout Configuration**: Set appropriate timeout based on operation complexity
- **Environment Variables**: Use for configuration, not secrets (use Parameter Store)
- **Container Size**: Optimize image size to reduce cold start time

### DynamoDB Gotchas
- **Hot Partitions**: Design partition keys to distribute load evenly
- **Query Patterns**: Optimize GSI design based on access patterns
- **Item Size**: Keep items under 400KB, use S3 for large data
- **Consistency**: Use strongly consistent reads when data consistency is critical
- **Batch Operations**: Use batch operations for bulk processing efficiency

### JWT/Security Gotchas
- **Token Expiration**: Always validate `exp` claim before processing
- **Algorithm Validation**: Ensure token uses expected algorithm (RS256)
- **JWKS Caching**: Cache public keys but refresh periodically
- **Claims Validation**: Verify all required claims (`aud`, `iss`, `token_use`)
- **Error Messages**: Don't leak sensitive information in error responses

### S3 Integration Gotchas
- **Presigned URL Expiration**: Set appropriate expiration time for upload URLs
- **File Type Validation**: Validate file types both client and server side
- **CORS Configuration**: Set proper CORS headers for browser uploads
- **File Size Limits**: Enforce file size limits to prevent abuse
- **Metadata Storage**: Use object metadata for searchable attributes

## Confidence Score: 9/10

This PRP provides comprehensive context for building a production-ready Go API with:
- ✅ Extensive Go ecosystem research completed
- ✅ Mat Ryer's HTTP service patterns integrated
- ✅ AWS Lambda container deployment patterns documented
- ✅ Real-world Go code examples with proper error handling
- ✅ JWT/Cognito authentication patterns researched
- ✅ DynamoDB and S3 integration patterns documented
- ✅ Comprehensive testing strategy with table-driven tests
- ✅ Infrastructure as code with Terraform
- ✅ CI/CD pipeline with GitHub Actions
- ✅ Security best practices and validation gates

The implementation should achieve one-pass success with this level of context and research-backed patterns.