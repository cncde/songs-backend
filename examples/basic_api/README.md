# Basic Go API Example

A simple REST API built with Gin framework demonstrating basic CRUD operations.

## Features

- **HTTP Router**: Gin framework for high-performance HTTP handling
- **REST Endpoints**: Full CRUD operations for user management
- **JSON Handling**: Automatic JSON marshaling/unmarshaling
- **Error Handling**: Proper HTTP status codes and error responses
- **Health Check**: Basic health check endpoint for monitoring

## API Endpoints

- `GET /health` - Health check endpoint
- `GET /api/v1/users` - Get all users
- `GET /api/v1/users/:id` - Get user by ID
- `POST /api/v1/users` - Create new user
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

## Usage

### Run the API

```bash
# Install dependencies
go mod tidy

# Run the server
go run main.go
```

The server will start on `http://localhost:8080`

### Test the API

```bash
# Health check
curl http://localhost:8080/health

# Get all users
curl http://localhost:8080/api/v1/users

# Get specific user
curl http://localhost:8080/api/v1/users/1

# Create new user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"username":"newuser","email":"newuser@example.com"}'

# Update user
curl -X PUT http://localhost:8080/api/v1/users/1 \
  -H "Content-Type: application/json" \
  -d '{"username":"updated_user","email":"updated@example.com"}'

# Delete user
curl -X DELETE http://localhost:8080/api/v1/users/1
```

## Key Go Patterns Demonstrated

- **Gin Router**: High-performance HTTP router with middleware support
- **JSON Binding**: Automatic request/response JSON handling
- **Error Handling**: Proper HTTP status codes and error responses
- **Route Grouping**: Organized API versioning with route groups
- **Handler Functions**: Clean separation of route handling logic

## Production Considerations

This is a basic example for learning. For production use, consider:

- Database integration instead of in-memory storage
- Authentication and authorization middleware
- Input validation with proper validation library
- Logging and monitoring integration
- CORS configuration for web applications
- Rate limiting and security headers
- Environment-based configuration management