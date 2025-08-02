# JWT Authentication Example

A Go API demonstrating JWT authentication with AWS Cognito integration using JWKS (JSON Web Key Set) verification.

## Features

- **AWS Cognito Integration**: JWT token verification with automatic key rotation
- **JWKS Support**: Automatic public key fetching and caching from Cognito
- **Middleware Pattern**: Reusable authentication middleware for route protection
- **Claims Extraction**: Full JWT claims parsing and validation
- **Role-Based Access**: Group-based authorization support
- **Security Headers**: Proper HTTP status codes and error responses

## Architecture

```
Client → Bearer Token → JWT Middleware → JWKS Verification → Claims Validation → Route Handler
```

## Setup

### Environment Variables

```bash
export COGNITO_USER_POOL_ID="us-east-1_EXAMPLE123"
export AWS_REGION="us-east-1"
```

### Run the API

```bash
# Install dependencies
go mod tidy

# Run the server
go run main.go
```

## API Endpoints

### Public Endpoints (No Authentication)
- `GET /public/health` - Health check
- `GET /public/info` - Public information

### Protected Endpoints (JWT Required)
- `GET /api/v1/profile` - Get current user profile
- `PUT /api/v1/profile` - Update user profile
- `GET /api/v1/users` - Get all users (admin only)
- `GET /api/v1/protected` - Protected test endpoint

## Testing

### Public Endpoint Test
```bash
curl http://localhost:8080/public/health
```

### Protected Endpoint Test (will fail without token)
```bash
curl http://localhost:8080/api/v1/protected
# Returns: 401 Unauthorized
```

### With JWT Token
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     http://localhost:8080/api/v1/protected
```

### Get User Profile
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     http://localhost:8080/api/v1/profile
```

### Update Profile
```bash
curl -X PUT \
     -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"display_name":"John Doe","bio":"Software Engineer"}' \
     http://localhost:8080/api/v1/profile
```

## JWT Token Structure

Expected AWS Cognito access token claims:
```json
{
  "sub": "user-uuid-here",
  "cognito:groups": ["admin", "user"],
  "cognito:username": "johndoe",
  "email": "john@example.com",
  "email_verified": true,
  "aud": "client-id",
  "token_use": "access",
  "auth_time": 1672531200,
  "iss": "https://cognito-idp.us-east-1.amazonaws.com/us-east-1_EXAMPLE123",
  "exp": 1672534800,
  "iat": 1672531200
}
```

## Key Security Features

### JWT Validation Steps
1. **Token Extraction**: Extract Bearer token from Authorization header
2. **Format Validation**: Verify JWT structure (header.payload.signature)
3. **Signature Verification**: Validate using Cognito's public keys (JWKS)
4. **Claims Validation**: Verify exp, aud, iss, and token_use claims
5. **Context Storage**: Store user information for handler access

### JWKS Key Management
- **Automatic Refresh**: Keys refreshed hourly from Cognito
- **Error Handling**: Graceful handling of key refresh failures
- **Caching**: Efficient key caching to minimize external calls
- **Timeout Protection**: 10-second timeout for key fetching

### Authorization Patterns
```go
// Extract user information from JWT claims
userSub := c.Get("user_sub")
username := c.Get("username")
email := c.Get("email")
groups := c.Get("cognito_groups")

// Role-based access control
if hasAdminRole(groups) {
    // Admin-only functionality
}
```

## Error Responses

### Missing Token
```json
{
  "error": "Authorization header is required",
  "code": "MISSING_AUTH_HEADER"
}
```

### Invalid Token
```json
{
  "error": "Invalid token",
  "code": "INVALID_TOKEN",
  "details": "token is expired"
}
```

### Insufficient Permissions
```json
{
  "error": "Insufficient permissions",
  "code": "FORBIDDEN"
}
```

## Production Considerations

### Security
- [ ] Configure proper CORS policies
- [ ] Implement rate limiting to prevent token brute-force
- [ ] Add request ID tracking for audit logs
- [ ] Implement token revocation checking
- [ ] Set up monitoring for failed authentication attempts

### Performance
- [ ] Optimize JWKS caching strategy
- [ ] Implement connection pooling for external requests
- [ ] Add metrics for authentication latency
- [ ] Consider token caching for high-frequency requests

### Monitoring
- [ ] Log authentication events with user context
- [ ] Monitor JWKS refresh failures
- [ ] Track JWT validation errors
- [ ] Set up alerts for suspicious authentication patterns

## AWS Cognito Setup

To use this example with real AWS Cognito:

1. Create a Cognito User Pool
2. Configure App Client with appropriate settings
3. Set up user groups (optional)
4. Configure the correct User Pool ID and region
5. Generate tokens using Cognito authentication flow

## Integration Examples

### Frontend JavaScript
```javascript
// Store JWT token from Cognito authentication
const token = await Auth.currentSession().getAccessToken().getJwtToken();

// Make authenticated requests
fetch('/api/v1/profile', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});
```

### Mobile App Integration
- Use AWS Amplify or Cognito SDKs
- Handle token refresh automatically
- Implement proper error handling for expired tokens