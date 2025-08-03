# Songs Backend API - Development Guide

This guide provides comprehensive instructions for local development,
testing, and contributing to the Songs Backend API. The guide covers setup,
development workflows, testing strategies, and contribution guidelines.

## Table of Contents

- [Development Environment Setup](#development-environment-setup)
- [Local Development](#local-development)
- [Testing](#testing)
- [Code Quality & Standards](#code-quality--standards)
- [Debugging](#debugging)
- [Contributing Guidelines](#contributing-guidelines)
- [Development Workflows](#development-workflows)
- [Tools & IDE Setup](#tools--ide-setup)

## Development Environment Setup

### Prerequisites

Ensure you have the following tools installed:

```bash
# Go 1.22 or later
go version
# Should return: go version go1.22.x

# Docker and Docker Compose
docker --version
docker-compose --version

# AWS CLI v2 (for local AWS services simulation)
aws --version

# Git
git --version

# Make (optional but recommended)
make --version
```

### IDE Recommendations

**VS Code Extensions:**

- Go extension (golang.go)
- AWS Toolkit (aws-toolkit-vscode)
- REST Client (humao.rest-client)
- Docker (ms-azuretools.vscode-docker)
- YAML (redhat.vscode-yaml)

**GoLand/IntelliJ IDEA:**

- Go plugin
- AWS Toolkit
- Docker plugin

### Project Setup

1. **Clone the repository:**

```bash
git clone https://github.com/cncde/songs-backend.git
cd songs-backend
```

<!-- markdownlint-disable MD029 -->
2. **Install Go dependencies:**

```bash
go mod download
go mod tidy
```

3. **Install development tools:**

```bash
# Install golangci-lint for linting
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install gosec for security scanning
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Install air for live reloading (optional)
go install github.com/cosmtrek/air@latest

# Install mockgen for generating mocks
go install github.com/golang/mock/mockgen@latest
```

4. **Set up pre-commit hooks (optional):**
<!-- markdownlint-enable MD029 -->

```bash
# Install pre-commit
pip install pre-commit

# Create .pre-commit-config.yaml
cat > .pre-commit-config.yaml << EOF
repos:
  - repo: local
    hooks:
      - id: go-fmt
        name: Go Format
        entry: gofmt
        language: golang
        types: [go]
        args: [-w]
      - id: go-lint
        name: Go Lint
        entry: golangci-lint
        language: golang
        types: [go]
        args: [run]
      - id: go-test
        name: Go Test
        entry: go
        language: golang
        types: [go]
        args: [test, ./...]
EOF

# Install the hooks
pre-commit install
```

## Local Development

### Environment Configuration

Create a `.env.local` file for local development:

```bash
# Application Settings
ENVIRONMENT=development
PORT=8080
AWS_REGION=us-east-1
API_VERSION=v1
LOG_LEVEL=debug
ENABLE_METRICS=false

# AWS LocalStack Configuration (for local AWS services)
AWS_ENDPOINT_URL=http://localhost:4566
AWS_ACCESS_KEY_ID=test
AWS_SECRET_ACCESS_KEY=test

# DynamoDB Configuration (LocalStack)
SONGS_TABLE=songs-local
COMMENTS_TABLE=comments-local

# S3 Configuration (LocalStack)
S3_BUCKET=songs-media-local

# Cognito Configuration (you can use a real Cognito pool or mock)
COGNITO_USER_POOL_ID=local_pool_id
COGNITO_REGION=us-east-1
# For development, you might want to disable JWT validation
JWT_VALIDATION_DISABLED=true
```

### Using LocalStack for Local AWS Services

Create `docker-compose.yml` for local AWS services:

```yaml
version: '3.8'
services:
  localstack:
    image: localstack/localstack:latest
    container_name: songs-localstack
    ports:
      - "4566:4566"
    environment:
      - SERVICES=dynamodb,s3,iam
      - DEBUG=1
      - DATA_DIR=/tmp/localstack/data
      - DOCKER_HOST=unix:///var/run/docker.sock
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock"
      - "./localstack-data:/tmp/localstack"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:4566/health"]
      interval: 30s
      timeout: 10s
      retries: 5

  # Optional: DynamoDB Admin UI
  dynamodb-admin:
    image: aaronshaf/dynamodb-admin:latest
    container_name: songs-dynamodb-admin
    ports:
      - "8001:8001"
    environment:
      - DYNAMO_ENDPOINT=http://localstack:4566
    depends_on:
      - localstack
```

### Initialize Local AWS Resources

Create `scripts/setup-local.sh`:

```bash
#!/bin/bash
set -e

echo "Setting up local AWS resources..."

# Wait for LocalStack to be ready
echo "Waiting for LocalStack..."
until curl -f http://localhost:4566/health > /dev/null 2>&1; do
  echo "Waiting for LocalStack to start..."
  sleep 2
done

export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=us-east-1
export AWS_ENDPOINT_URL=http://localhost:4566

# Create DynamoDB tables
echo "Creating DynamoDB tables..."

# Songs table
aws dynamodb create-table \
    --table-name songs-local \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
        AttributeName=user_id,AttributeType=S \
        AttributeName=stage,AttributeType=S \
        AttributeName=created_at,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
    --global-secondary-indexes \
        "IndexName=UserIndex,KeySchema=[{AttributeName=user_id,KeyType=HASH},\
            {AttributeName=created_at,KeyType=RANGE}],Projection={ProjectionType=ALL},\
            ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5}" \
        "IndexName=StageIndex,KeySchema=[{AttributeName=stage,KeyType=HASH},\
            {AttributeName=created_at,KeyType=RANGE}],Projection={ProjectionType=ALL},\
            ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5}" \
    --provisioned-throughput \
        ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --endpoint-url http://localhost:4566

# Comments table
aws dynamodb create-table \
    --table-name comments-local \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
        AttributeName=song_id,AttributeType=S \
        AttributeName=user_id,AttributeType=S \
        AttributeName=audio_url,AttributeType=S \
        AttributeName=created_at,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
    --global-secondary-indexes \
        "IndexName=SongIndex,KeySchema=[{AttributeName=song_id,KeyType=HASH},\
            {AttributeName=created_at,KeyType=RANGE}],Projection={ProjectionType=ALL},\
            ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5}" \
        "IndexName=UserIndex,KeySchema=[{AttributeName=user_id,KeyType=HASH},\
            {AttributeName=created_at,KeyType=RANGE}],Projection={ProjectionType=ALL},\
            ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5}" \
        "IndexName=AudioIndex,KeySchema=[{AttributeName=audio_url,KeyType=HASH},\
            {AttributeName=created_at,KeyType=RANGE}],Projection={ProjectionType=ALL},\
            ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5}" \
    --provisioned-throughput \
        ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --endpoint-url http://localhost:4566

# Create S3 bucket
echo "Creating S3 bucket..."
aws s3 mb s3://songs-media-local --endpoint-url http://localhost:4566

# Configure bucket CORS
aws s3api put-bucket-cors \
    --bucket songs-media-local \
    --cors-configuration file://cors-config.json \
    --endpoint-url http://localhost:4566

echo "Local AWS resources created successfully!"
echo "DynamoDB Admin UI: http://localhost:8001"
echo "LocalStack Health: http://localhost:4566/health"
```

Create `cors-config.json`:

```json
{
    "CORSRules": [
        {
            "AllowedHeaders": ["*"],
            "AllowedMethods": ["GET", "PUT", "POST", "DELETE"],
            "AllowedOrigins": ["*"],
            "ExposeHeaders": ["ETag"]
        }
    ]
}
```

### Running the Application

1. **Start LocalStack:**

```bash
docker-compose up -d
```

1. **Initialize local resources:**

```bash
chmod +x scripts/setup-local.sh
./scripts/setup-local.sh
```

1. **Run the application:**

```bash
# Load environment variables
export $(cat .env.local | xargs)

# Run the application
go run cmd/api/main.go
```

1. **Or use air for live reloading:**

```bash
# Create .air.toml configuration
cat > .air.toml << EOF
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main cmd/api/main.go"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
EOF

# Load environment and run with air
export $(cat .env.local | xargs)
air
```

1. **Test the API:**

```bash
# Health check
curl http://localhost:8080/health

# Create a song (mock JWT)
curl -X POST http://localhost:8080/api/v1/songs \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer mock-jwt-token" \
  -d '{
    "title": "Test Song",
    "description": "A test song for development",
    "stage": "draft",
    "tags": ["test", "development"]
  }'
```

## Testing

### Test Structure

```sh
tests/
├── unit/           # Unit tests for individual components
│   ├── models/
│   ├── service/
│   └── repository/
├── integration/    # Integration tests with dependencies
│   ├── api/
│   └── database/
├── e2e/           # End-to-end API tests
│   └── scenarios/
├── testdata/      # Test fixtures and mock data
└── mocks/         # Generated mocks
```

### Unit Testing

**Example unit test for Song model:**

```go
// tests/unit/models/song_test.go
package models_test

import (
    "testing"
    
    "github.com/cncde/songs-backend/internal/models"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestNewSong(t *testing.T) {
    tests := []struct {
        name     string
        request  models.CreateSongRequest
        userID   string
        expected func(*testing.T, *models.Song)
    }{
        {
            name: "creates song with valid data",
            request: models.CreateSongRequest{
                Title:       "Test Song",
                Description: "A test song",
                Stage:       "draft",
                Tags:        []string{"rock", "acoustic"},
            },
            userID: "user-123",
            expected: func(t *testing.T, song *models.Song) {
                assert.Equal(t, "Test Song", song.Title)
                assert.Equal(t, "A test song", song.Description)
                assert.Equal(t, "draft", song.Stage)
                assert.Equal(t, []string{"rock", "acoustic"}, song.Tags)
                assert.Equal(t, "user-123", song.UserID)
                assert.NotEmpty(t, song.ID)
                assert.NotEmpty(t, song.CreatedAt)
                assert.NotEmpty(t, song.UpdatedAt)
            },
        },
        {
            name: "defaults to draft stage when empty",
            request: models.CreateSongRequest{
                Title: "Test Song",
                Stage: "",
            },
            userID: "user-123",
            expected: func(t *testing.T, song *models.Song) {
                assert.Equal(t, "draft", song.Stage)
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            song := models.NewSong(tt.request, tt.userID)
            require.NotNil(t, song)
            tt.expected(t, song)
        })
    }
}

func TestSong_Update(t *testing.T) {
    song := &models.Song{
        ID:     "song-123",
        Title:  "Original Title",
        Stage:  "draft",
        UserID: "user-123",
    }
    
    updateReq := models.UpdateSongRequest{
        Title: stringPtr("Updated Title"),
        Stage: stringPtr("review"),
    }
    
    song.Update(updateReq)
    
    assert.Equal(t, "Updated Title", song.Title)
    assert.Equal(t, "review", song.Stage)
}

func stringPtr(s string) *string {
    return &s
}
```

### Integration Testing

**Example integration test for Song Service:**

```go
// tests/integration/service/song_service_test.go
package service_test

import (
    "context"
    "testing"
    
    "github.com/cncde/songs-backend/internal/models"
    "github.com/cncde/songs-backend/internal/service"
    "github.com/cncde/songs-backend/tests/testhelpers"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestSongService_CreateSong(t *testing.T) {
    // Setup test dependencies
    dynamoClient := testhelpers.SetupTestDynamoDB(t)
    songRepo := testhelpers.NewTestSongRepository(dynamoClient)
    commentRepo := testhelpers.NewTestCommentRepository(dynamoClient)
    
    songService := service.NewSongService(songRepo, commentRepo)
    
    ctx := context.Background()
    userID := "test-user-123"
    
    req := models.CreateSongRequest{
        Title:       "Integration Test Song",
        Description: "Testing song creation",
        Stage:       "draft",
        Tags:        []string{"test"},
    }
    
    // Create song
    song, err := songService.CreateSong(ctx, req, userID)
    require.NoError(t, err)
    require.NotNil(t, song)
    
    assert.Equal(t, req.Title, song.Title)
    assert.Equal(t, req.Description, song.Description)
    assert.Equal(t, userID, song.UserID)
    assert.NotEmpty(t, song.ID)
    
    // Verify song can be retrieved
    retrievedSong, err := songService.GetSong(ctx, song.ID, userID)
    require.NoError(t, err)
    assert.Equal(t, song.ID, retrievedSong.ID)
    assert.Equal(t, song.Title, retrievedSong.Title)
}
```

### End-to-End Testing

**Example E2E test:**

```go
// tests/e2e/songs_api_test.go
package e2e_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/cncde/songs-backend/internal/models"
    "github.com/cncde/songs-backend/internal/server"
    "github.com/cncde/songs-backend/tests/testhelpers"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestSongsAPI_CreateAndRetrieve(t *testing.T) {
    // Setup test server
    deps := testhelpers.SetupTestDependencies(t)
    srv, err := server.New(deps)
    require.NoError(t, err)
    
    router := srv.Handler()
    
    // Create song request
    createReq := models.CreateSongRequest{
        Title:       "E2E Test Song",
        Description: "End-to-end test",
        Stage:       "draft",
        Tags:        []string{"e2e", "test"},
    }
    
    reqBody, err := json.Marshal(createReq)
    require.NoError(t, err)
    
    // Create song
    req := httptest.NewRequest("POST", "/api/v1/songs", bytes.NewBuffer(reqBody))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+testhelpers.GenerateMockJWT("test-user"))
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusCreated, w.Code)
    
    var createResp models.SongResponse
    err = json.Unmarshal(w.Body.Bytes(), &createResp)
    require.NoError(t, err)
    require.NotNil(t, createResp.Song)
    
    songID := createResp.Song.ID
    
    // Retrieve song
    req = httptest.NewRequest("GET", "/api/v1/songs/"+songID, nil)
    req.Header.Set("Authorization", "Bearer "+testhelpers.GenerateMockJWT("test-user"))
    
    w = httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
    
    var getResp models.SongResponse
    err = json.Unmarshal(w.Body.Bytes(), &getResp)
    require.NoError(t, err)
    require.NotNil(t, getResp.Song)
    
    assert.Equal(t, songID, getResp.Song.ID)
    assert.Equal(t, createReq.Title, getResp.Song.Title)
}
```

### Test Helpers

Create `tests/testhelpers/setup.go`:

```go
package testhelpers

import (
    "context"
    "testing"
    
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/cncde/songs-backend/internal/models"
    "github.com/cncde/songs-backend/internal/repository"
    "github.com/cncde/songs-backend/internal/server"
    "github.com/golang-jwt/jwt/v5"
)

func SetupTestDependencies(t *testing.T) *server.Dependencies {
    t.Helper()
    
    // Load AWS config for LocalStack
    cfg, err := config.LoadDefaultConfig(context.TODO(),
        config.WithEndpointResolver(config.EndpointResolverFunc(
            func(service, region string) (aws.Endpoint, error) {
                return aws.Endpoint{
                    URL: "http://localhost:4566",
                }, nil
            })),
    )
    if err != nil {
        t.Fatalf("Failed to load AWS config: %v", err)
    }
    
    return &server.Dependencies{
        Config: &config.Config{
            Environment:   "test",
            SongsTable:    "songs-test",
            CommentsTable: "comments-test",
            S3Bucket:      "songs-media-test",
        },
        DynamoClient: dynamodb.NewFromConfig(cfg),
        S3Client:     s3.NewFromConfig(cfg),
    }
}

func GenerateMockJWT(userID string) string {
    claims := jwt.MapClaims{
        "sub":              userID,
        "cognito:username": userID,
        "email":            userID + "@example.com",
        "token_use":        "access",
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, _ := token.SignedString([]byte("test-secret"))
    
    return tokenString
}

func CreateTestSong(t *testing.T, repo repository.SongRepository, \
    userID string) *models.Song {
    t.Helper()
    
    req := models.CreateSongRequest{
        Title:       "Test Song",
        Description: "Test description",
        Stage:       "draft",
        Tags:        []string{"test"},
    }
    
    song := models.NewSong(req, userID)
    err := repo.Create(context.Background(), song)
    if err != nil {
        t.Fatalf("Failed to create test song: %v", err)
    }
    
    return song
}
```

### Running Tests

Create a `Makefile` for test automation:

```makefile
.PHONY: test test-unit test-integration test-e2e test-coverage \
    lint fmt vet security clean

# Test commands
test: test-unit test-integration test-e2e

test-unit:
 @echo "Running unit tests..."
 go test -v ./tests/unit/...

test-integration:
 @echo "Running integration tests..."
 docker-compose up -d localstack
 ./scripts/setup-local.sh
 go test -v ./tests/integration/...
 docker-compose down

test-e2e:
 @echo "Running end-to-end tests..."
 docker-compose up -d localstack
 ./scripts/setup-local.sh
 go test -v ./tests/e2e/...
 docker-compose down

test-coverage:
 @echo "Running tests with coverage..."
 go test -v -race -coverprofile=coverage.out ./...
 go tool cover -html=coverage.out -o coverage.html
 @echo "Coverage report generated: coverage.html"

# Quality checks
lint:
 @echo "Running linter..."
 golangci-lint run

fmt:
 @echo "Formatting code..."
 go fmt ./...

vet:
 @echo "Running go vet..."
 go vet ./...

security:
 @echo "Running security scan..."
 gosec ./...

# Development commands
run:
 @echo "Starting local development server..."
 docker-compose up -d localstack
 ./scripts/setup-local.sh
 export $$(cat .env.local | xargs) && go run cmd/api/main.go

clean:
 @echo "Cleaning up..."
 docker-compose down -v
 rm -rf tmp/
 rm -f coverage.out coverage.html

# Build commands
build:
 @echo "Building application..."
 go build -o bin/songs-backend cmd/api/main.go

docker-build:
 @echo "Building Docker image..."
 docker build -t songs-backend:latest .

# Help
help:
 @echo "Available commands:"
 @echo "  test           - Run all tests"
 @echo "  test-unit      - Run unit tests"
 @echo "  test-integration - Run integration tests"
 @echo "  test-e2e       - Run end-to-end tests"
 @echo "  test-coverage  - Run tests with coverage report"
 @echo "  lint           - Run linter"
 @echo "  fmt            - Format code"
 @echo "  vet            - Run go vet"
 @echo "  security       - Run security scan"
 @echo "  run            - Start local development server"
 @echo "  build          - Build application binary"
 @echo "  docker-build   - Build Docker image"
 @echo "  clean          - Clean up resources"
```

## Code Quality & Standards

### Linting Configuration

Create `.golangci.yml`:

```yaml
linters-settings:
  errcheck:
    check-type-assertions: true
  goconst:
    min-len: 2
    min-occurrences: 3
  gocyclo:
    min-complexity: 15
  gofmt:
    simplify: true
  goimports:
    local-prefixes: github.com/cncde/songs-backend
  golint:
    min-confidence: 0
  govet:
    check-shadowing: true
  misspell:
    locale: US
  nolintlint:
    allow-leading-space: true
    require-explanation: false
    require-specific: false

linters:
  disable-all: true
  enable:
    - bodyclose
    - deadcode
    - depguard
    - dogsled
    - errcheck
    - exportloopref
    - exhaustive
    - gochecknoinits
    - goconst
    - gocritic
    - gocyclo
    - gofmt
    - goimports
    - golint
    - gomnd
    - goprintffuncname
    - gosec
    - gosimple
    - govet
    - ineffassign
    - misspell
    - noctx
    - nolintlint
    - rowserrcheck
    - staticcheck
    - structcheck
    - stylecheck
    - typecheck
    - unconvert
    - unparam
    - unused
    - varcheck
    - whitespace

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gomnd
        - goconst
        - gosec

run:
  timeout: 5m
```

### Code Formatting

Use `gofmt` and `goimports` for consistent formatting:

```bash
# Format code
go fmt ./...

# Organize imports
goimports -w .
```

### Git Hooks

Create `.githooks/pre-commit`:

```bash
#!/bin/sh
set -e

echo "Running pre-commit checks..."

# Format code
echo "Formatting code..."
go fmt ./...

# Run linter
echo "Running linter..."
golangci-lint run

# Run tests
echo "Running tests..."
go test ./...

# Run security scan
echo "Running security scan..."
gosec ./...

echo "All checks passed!"
```

Make it executable:

```bash
chmod +x .githooks/pre-commit
git config core.hooksPath .githooks
```

## Debugging

### Local Debugging with VS Code

Create `.vscode/launch.json`:

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug API",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/cmd/api/main.go",
            "envFile": "${workspaceFolder}/.env.local",
            "console": "integratedTerminal",
            "preLaunchTask": "setup-local-env"
        },
        {
            "name": "Debug Tests",
            "type": "go",
            "request": "launch",
            "mode": "test",
            "program": "${workspaceFolder}",
            "envFile": "${workspaceFolder}/.env.local"
        }
    ]
}
```

Create `.vscode/tasks.json`:

```json
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "setup-local-env",
            "type": "shell",
            "command": "docker-compose up -d && ./scripts/setup-local.sh",
            "group": "build",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared"
            }
        }
    ]
}
```

### Logging and Monitoring

Use structured logging for better debugging:

```go
package main

import (
    "log/slog"
    "os"
)

func setupLogger(level string) *slog.Logger {
    var logLevel slog.Level
    switch level {
    case "debug":
        logLevel = slog.LevelDebug
    case "info":
        logLevel = slog.LevelInfo
    case "warn":
        logLevel = slog.LevelWarn
    case "error":
        logLevel = slog.LevelError
    default:
        logLevel = slog.LevelInfo
    }
    
    handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: logLevel,
    })
    
    return slog.New(handler)
}

// Usage in handlers
func (h *Handler) handleCreateSong(c *gin.Context) {
    logger := h.logger.With(
        "method", c.Request.Method,
        "path", c.Request.URL.Path,
        "user_id", c.GetString("user_id"),
    )
    
    logger.Info("Creating song")
    
    // Handle request...
    
    logger.Info("Song created successfully", "song_id", song.ID)
}
```

## Contributing Guidelines

### Branch Naming Convention

Use descriptive branch names:

- `feature/add-song-search` - New features
- `fix/jwt-validation-bug` - Bug fixes
- `refactor/repository-layer` - Code refactoring
- `docs/api-documentation` - Documentation updates
- `test/integration-tests` - Test improvements

### Commit Message Format

Follow conventional commits:

```text
type(scope): description

body (optional)

footer (optional)
```

Types:

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Test changes
- `chore`: Maintenance tasks

Examples:

```text
feat(songs): add song search functionality

Implement full-text search for songs with support for:
- Title and description search
- Tag-based filtering
- Pagination with cursor

Closes #123

fix(auth): resolve JWT token validation issue

The JWKS refresh mechanism was not handling expired keys correctly.
Updated to properly refresh keys when validation fails.

Fixes #456
```

### Pull Request Process

1. **Create feature branch:**

```bash
git checkout -b feature/your-feature-name
```

1. **Make changes following coding standards**

1. **Add/update tests:**

```bash
# Ensure all tests pass
make test

# Add tests for new functionality
# Update existing tests if needed
```

1. **Update documentation:**

- Update README.md if needed
- Add/update API documentation
- Update architecture docs for significant changes

1. **Create pull request:**

- Use the PR template
- Include detailed description
- Link related issues
- Add screenshots for UI changes

### Pull Request Template

Create `.github/pull_request_template.md`:

```markdown
## Description
Brief description of changes made.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality
  to not work as expected)
- [ ] Documentation update

## Changes Made
- Change 1
- Change 2
- Change 3

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] E2E tests pass
- [ ] Manual testing completed

## Screenshots (if applicable)
[Add screenshots here]

## Related Issues
Fixes #[issue_number]
Relates to #[issue_number]

## Checklist
- [ ] Code follows the project's coding standards
- [ ] Self-review of the code completed
- [ ] Code is commented, particularly in hard-to-understand areas
- [ ] Corresponding changes to the documentation have been made
- [ ] Tests have been added/updated
- [ ] All tests pass locally
- [ ] Changes generate no new warnings
```

## Development Workflows

### Feature Development Workflow

1. **Planning:**
   - Create/review GitHub issue
   - Design API changes if needed
   - Update architecture docs if significant

2. **Implementation:**
   - Create feature branch
   - Implement with TDD approach
   - Add comprehensive tests
   - Update documentation

3. **Review:**
   - Self-review code
   - Create pull request
   - Address review feedback
   - Ensure CI passes

4. **Merge:**
   - Squash and merge to main
   - Delete feature branch
   - Update local main branch

### Hotfix Workflow

1. **Create hotfix branch from main:**

```bash
git checkout main
git pull origin main
git checkout -b hotfix/critical-fix
```

1. **Implement minimal fix**

1. **Test thoroughly**

1. **Create PR with hotfix label**

1. **Fast-track review and merge**

### Release Workflow

1. **Create release branch:**

```bash
git checkout -b release/v1.2.0
```

1. **Update version information**

1. **Run full test suite**

1. **Create release PR**

1. **Tag release after merge:**

```bash
git tag -a v1.2.0 -m "Release version 1.2.0"
git push origin v1.2.0
```

## Tools & IDE Setup

### Recommended VS Code Settings

Create `.vscode/settings.json`:

```json
{
    "go.formatTool": "goimports",
    "go.lintTool": "golangci-lint",
    "go.testFlags": ["-v"],
    "go.testTimeout": "60s",
    "go.coverOnSave": true,
    "go.coverageDecorator": {
        "type": "gutter"
    },
    "go.useLanguageServer": true,
    "gopls": {
        "analyses": {
            "unusedparams": true,
            "shadow": true
        },
        "staticcheck": true
    },
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
        "source.organizeImports": true
    }
}
```

### Useful VS Code Snippets

Create `.vscode/go.json`:

```json
{
    "Test Function": {
        "prefix": "test",
        "body": [
            "func Test${1:FunctionName}(t *testing.T) {",
            "\ttests := []struct {",
            "\t\tname string",
            "\t\texpected ${2:type}",
            "\t}{",
            "\t\t{",
            "\t\t\tname: \"${3:test case}\",",
            "\t\t\texpected: ${4:value},",
            "\t\t},",
            "\t}",
            "",
            "\tfor _, tt := range tests {",
            "\t\tt.Run(tt.name, func(t *testing.T) {",
            "\t\t\t${5:// test implementation}",
            "\t\t})",
            "\t}",
            "}"
        ],
        "description": "Create table-driven test function"
    },
    "HTTP Handler": {
        "prefix": "handler",
        "body": [
            "func (h *Handler) handle${1:HandlerName}(c *gin.Context) {",
            "\tlogger := h.logger.With(",
            "\t\t\"method\", c.Request.Method,",
            "\t\t\"path\", c.Request.URL.Path,",
            "\t\t\"user_id\", c.GetString(\"user_id\"),",
            "\t)",
            "",
            "\t${2:// Implementation}",
            "",
            "\tc.JSON(http.StatusOK, ${3:response})",
            "}"
        ],
        "description": "Create HTTP handler function"
    }
}
```

This development guide provides everything needed to contribute effectively to
the Songs Backend API project, from initial setup through advanced workflows
and tooling.
