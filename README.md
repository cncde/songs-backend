# Songs Backend API

A production-ready REST API for managing songs and comments, built with Go 1.22
and designed for AWS Lambda deployment. This API provides comprehensive song
management capabilities, comment systems, secure media handling, and JWT
authentication via AWS Cognito.

## Features

- **Song Management**: Full CRUD operations with stage support
- **Comment System**: Comments on song audio files with user attribution
- **Media Handling**: Secure S3-based file upload/download for images and audio
- **JWT Authentication**: AWS Cognito integration with JWKS validation
- **Serverless Architecture**: Containerized Lambda deployment with Docker
- **Database**: DynamoDB for scalable NoSQL storage
- **Production Ready**: Comprehensive error handling, validation, and security

## Architecture Overview

```text
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   AWS Lambda    │────│  Songs Backend   │────│   DynamoDB      │
│   (Container)   │    │   (Go + Gin)     │    │   (Songs +      │
└─────────────────┘    └──────────────────┘    │   Comments)     │
                              │                └─────────────────┘
                              │                 
                       ┌──────────────────┐    ┌─────────────────┐
                       │  AWS Cognito     │    │      S3         │
                       │  (JWT Auth)      │    │  (Media Files)  │
                       └──────────────────┘    └─────────────────┘
```

## Technology Stack

- **Language**: Go 1.22
- **Framework**: Gin HTTP framework
- **Authentication**: JWT with AWS Cognito
- **Database**: AWS DynamoDB
- **Storage**: AWS S3
- **Deployment**: AWS Lambda with Container images
- **Validation**: go-playground/validator
- **Documentation**: OpenAPI/Swagger compatible

## Quick Start

### Prerequisites

- Go 1.22+
- Docker
- AWS CLI configured
- AWS Account with appropriate permissions

### Local Development

1. **Clone the repository**:

   ```bash
   git clone <repository-url>
   cd songs-backend
   ```

2. **Install dependencies**:

   ```bash
   go mod download
   go mod tidy
   ```

3. **Set environment variables**:

   ```bash
   export ENVIRONMENT=development
   export PORT=8080
   export AWS_REGION=us-east-1
   export SONGS_TABLE=songs-dev
   export COMMENTS_TABLE=comments-dev
   export S3_BUCKET=songs-media-dev
   export COGNITO_USER_POOL_ID=your-user-pool-id
   ```

4. **Run locally**:

   ```bash
   go run cmd/api/main.go
   ```

5. **Test the API**:

   ```bash
   curl http://localhost:8080/health
   ```

### Docker Development

1. **Build the container**:

   ```bash
   docker build -t songs-api .
   ```

2. **Run with environment variables**:

   ```bash
   docker run -p 8080:8080 \
     -e ENVIRONMENT=development \
     -e PORT=8080 \
     -e AWS_REGION=us-east-1 \
     -e SONGS_TABLE=songs-dev \
     -e COMMENTS_TABLE=comments-dev \
     -e S3_BUCKET=songs-media-dev \
     -e COGNITO_USER_POOL_ID=your-user-pool-id \
     songs-api
   ```

## Environment Variables

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `SONGS_TABLE` | DynamoDB table name for songs | `songs-prod` |
| `COMMENTS_TABLE` | DynamoDB table name for comments | `comments-prod` |
| `S3_BUCKET` | S3 bucket name for media files | `songs-media-prod` |
| `COGNITO_USER_POOL_ID` | AWS Cognito User Pool ID | `us-east-1_ABC123DEF` |

### Optional Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `ENVIRONMENT` | Application environment | `development` |
| `PORT` | HTTP server port | `8080` |
| `AWS_REGION` | AWS region | `us-east-1` |
| `COGNITO_REGION` | Cognito region (if different) | Same as `AWS_REGION` |
| `API_VERSION` | API version prefix | `v1` |
| `LOG_LEVEL` | Logging level | `info` |
| `ENABLE_METRICS` | Enable metrics collection | `true` |

## API Endpoints

### Health Check

- `GET /health` - API health status

### Public Endpoints (No Authentication)

- `GET /api/v1/public/songs` - List published songs
- `GET /api/v1/public/songs/{id}` - Get published song details
- `GET /api/v1/public/songs/search` - Search published songs

### Protected Endpoints (JWT Required)

#### Songs

- `POST /api/v1/songs` - Create new song
- `GET /api/v1/songs` - List user's songs
- `GET /api/v1/songs/{id}` - Get song details
- `PUT /api/v1/songs/{id}` - Update song
- `DELETE /api/v1/songs/{id}` - Delete song

#### Song Media

- `POST /api/v1/songs/{id}/images` - Add image to song
- `DELETE /api/v1/songs/{id}/images` - Remove image from song
- `POST /api/v1/songs/{id}/audio` - Add audio to song
- `DELETE /api/v1/songs/{id}/audio` - Remove audio from song

#### Comments

- `GET /api/v1/songs/{id}/comments` - List song comments
- `POST /api/v1/songs/{id}/comments` - Create comment on song
- `GET /api/v1/comments/{id}` - Get comment details
- `PUT /api/v1/comments/{id}` - Update comment
- `DELETE /api/v1/comments/{id}` - Delete comment
- `GET /api/v1/comments` - List user's comments

#### Media Management

- `POST /api/v1/media/upload-url` - Generate presigned upload URL
- `GET /api/v1/media/download-url/{fileKey}` - Generate presigned download URL
- `DELETE /api/v1/media/{fileKey}` - Delete media file
- `GET /api/v1/media/files` - List user's files

#### Statistics

- `GET /api/v1/stats/songs` - Get song statistics
- `GET /api/v1/stats/comments` - Get comment statistics
- `GET /api/v1/stats/media` - Get media statistics

## Authentication

The API uses JWT authentication via AWS Cognito. Include the JWT token in the
Authorization header:

```bash
Authorization: Bearer <jwt-token>
```

### Required JWT Claims

- `sub`: User ID (subject)
- `cognito:username`: Username
- `email`: User email
- `token_use`: Must be "access" or "id"
- `iss`: Must match Cognito issuer
- `exp`: Expiration time
- `iat`: Issued at time

## Data Models

### Song Model

```json
{
  "id": "uuid",
  "title": "Song Title",
  "subtitle": "Optional Subtitle",
  "description": "Song description",
  "image_urls": ["https://s3.../image1.jpg"],
  "audio_urls": ["https://s3.../audio1.mp3"],
  "lyrics": "Song lyrics text",
  "stage": "draft|review|published|archived",
  "tags": ["rock", "acoustic"],
  "user_id": "user-uuid",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Comment Model

```json
{
  "id": "uuid",
  "song_id": "song-uuid",
  "audio_url": "https://s3.../audio.mp3",
  "content": "Comment text",
  "user_id": "user-uuid",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```


## Validation Rules

### Song Validation

- Title: Required, 1-255 characters
- Subtitle: Optional, max 255 characters  
- Description: Optional, max 2000 characters
- Lyrics: Optional, max 10000 characters
- Tags: Max 20 tags, each 1-50 characters
- Stage: Must be valid stage value

### Comment Validation

- Content: Required, 1-2000 characters
- Audio URL: Required, valid URL format
- Song ID: Must reference existing song

### Media Validation

- Images: JPEG, PNG, WebP, GIF (max 10MB)
- Audio: MP3, WAV, MP4, AAC, OGG, FLAC (max 100MB)

## Error Handling

The API returns consistent error responses:

```json
{
  "error": "Error message",
  "status": 400,
  "path": "/api/v1/songs",
  "method": "POST",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### Common HTTP Status Codes

- `200` - Success
- `201` - Created
- `400` - Bad Request (validation error)
- `401` - Unauthorized (missing/invalid JWT)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found
- `409` - Conflict (duplicate resource)
- `500` - Internal Server Error

## Development

### Project Structure

```text
songs-backend/
├── cmd/api/main.go              # Application entry point
├── internal/
│   ├── config/config.go         # Configuration management
│   ├── middleware/auth.go       # JWT authentication middleware
│   ├── models/                  # Data models and validation
│   │   ├── song.go
│   │   └── comment.go
│   ├── repository/              # DynamoDB operations
│   │   ├── dynamodb.go
│   │   ├── songs.go
│   │   └── comments.go
│   ├── service/                 # Business logic
│   │   ├── songs.go
│   │   ├── comments.go
│   │   └── media.go
│   ├── storage/s3.go           # S3 file operations
│   └── server/server.go        # HTTP server setup
├── Dockerfile                  # Container build configuration
├── go.mod                      # Go module definition
└── go.sum                      # Go module checksums
```

### Code Style

This project follows Go best practices and conventions:

- Use `gofmt` for code formatting
- Follow effective Go idioms
- Handle errors explicitly
- Use context for cancellation
- Write comprehensive tests
- Document exported functions

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run benchmarks
go test -bench=. ./...
```

### Linting

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run

# Run security checks
gosec ./...

# Vet code
go vet ./...
```

## Production Deployment

See [DEPLOYMENT.md](DEPLOYMENT.md) for detailed AWS deployment instructions.

### Build for Production

```bash
# Build binary for Linux Lambda
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -tags lambda.norpc \
  -ldflags="-s -w" \
  -o main \
  cmd/api/main.go

# Build Docker image
docker build -t songs-api:latest .

# Tag for ECR
docker tag songs-api:latest <account>.dkr.ecr.<region>.amazonaws.com/songs-api:latest
```

## Monitoring and Observability

### Metrics

The API exposes metrics for monitoring:

- Request count and duration
- Error rates by endpoint
- Database operation metrics
- S3 operation metrics

### Logging

Structured logging with JSON format in production:

```json
{
  "level": "info",
  "timestamp": "2024-01-01T00:00:00Z",
  "method": "POST",
  "path": "/api/v1/songs",
  "status": 201,
  "duration": "45ms",
  "user_id": "user-123"
}
```

### Health Checks

- `GET /health` - Basic health status
- Lambda runtime health monitoring
- DynamoDB and S3 connectivity checks

## Security Considerations

- JWT validation with JWKS refresh
- Input validation and sanitization
- SQL injection prevention (N/A for DynamoDB)
- Rate limiting (configured at AWS API Gateway level)
- CORS configuration
- File upload restrictions
- User isolation (data segregation by user ID)

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make changes following Go conventions
4. Add tests for new functionality
5. Run linting and tests (`make test lint`)
6. Commit changes (`git commit -m 'Add amazing feature'`)
7. Push to branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Development Workflow

1. Write failing tests
2. Implement functionality
3. Ensure tests pass
4. Update documentation
5. Submit PR with clear description

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE)
file for details.

## Support

- Create an issue for bug reports
- Start a discussion for feature requests
- Check existing issues before creating new ones

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history and release notes.
