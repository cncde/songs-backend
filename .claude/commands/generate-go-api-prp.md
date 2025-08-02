# Generate Go API PRP

## Feature file: PRPs/INITIAL.md

Generate a comprehensive PRP for Go REST API development with AWS Lambda containerized deployment, JWT authentication, DynamoDB data persistence, and S3 file handling based on the detailed requirements in the INITIAL.md file.

**CRITICAL: Web search and Go ecosystem research is essential. Use it extensively throughout this process.**

## Research Process

1. **Read and Understand Go API Requirements**
   - Read the specified INITIAL.md file thoroughly
   - Understand the Go API features, authentication requirements, and AWS integrations
   - Note specific Go frameworks, libraries, or patterns mentioned
   - Identify the scope and complexity of the Go API needed

2. **Extensive Go Ecosystem Research (CRITICAL)**
   - **WEB SEARCH Go REST API frameworks extensively** - Gin, Echo, Gorilla Mux, Fiber
   - Study Go HTTP server patterns, middleware chains, and routing approaches
   - Research AWS Lambda Go containerized deployment patterns and best practices
   - Find Go JWT authentication examples with AWS Cognito integration
   - Research AWS SDK for Go v2 patterns for DynamoDB and S3 operations
   - Identify Go project structure conventions and dependency management
   - Look for Go API security patterns and input validation approaches

3. **AWS Integration Pattern Analysis**
   - Examine Go Lambda container deployment examples and Dockerfiles
   - Research Go DynamoDB client patterns, query optimization, and error handling
   - Find Go S3 upload patterns, multipart uploads, and file handling examples
   - Study AWS IAM patterns for Go Lambda functions and least-privilege access
   - Research GitHub Actions workflows for Go testing, building, and Lambda deployment
   - Document AWS service integration gotchas and performance considerations

4. **Go Development Workflow Research**
   - Research Go testing frameworks, table-driven test patterns, and mocking strategies
   - Study Go linting tools (golangci-lint, go vet, gosec) and static analysis
   - Find Go build patterns, cross-compilation, and optimization techniques
   - Research Go dependency management with modules and security scanning
   - Look for Go development environment setup and debugging approaches

## PRP Generation

Using PRPs/templates/prp_go_api_base.md as the foundation:

### Critical Go Context to Include from Web Research

**Go Framework Documentation (from web search)**:
- Official Go documentation URLs and effective Go patterns
- Go REST framework docs (Gin, Echo, Gorilla Mux) with usage examples
- AWS Lambda Go documentation and container deployment guides
- Go JWT library documentation and authentication patterns
- AWS SDK for Go v2 documentation and service client examples

**Implementation Patterns (from research)**:
- Go project structure patterns (cmd/, internal/, pkg/) with real examples
- HTTP handler patterns, middleware chains, and request/response handling
- Go error handling patterns and context propagation approaches
- AWS service integration patterns with proper error handling
- Go testing patterns including table-driven tests and mocking

**Real-World Go API Examples**:
- Working Go API code snippets with proper error handling
- Lambda container Dockerfiles with multi-stage builds
- JWT middleware implementations with AWS Cognito verification
- DynamoDB operation examples with proper pagination and error handling
- S3 upload implementations with metadata and security considerations

### Implementation Blueprint

Based on Go ecosystem research findings:
- **Go API Architecture**: HTTP server patterns, middleware design, handler organization
- **AWS Lambda Integration**: Container deployment, runtime configuration, performance optimization
- **Authentication System**: JWT verification, middleware implementation, security validation
- **Data Layer Design**: DynamoDB client patterns, query optimization, error handling strategies
- **File Upload System**: S3 integration patterns, multipart uploads, security considerations

### Validation Gates (Must be Executable)

```bash
# Go Project Structure Validation
find . -name "*.go" | head -10  # Should show Go source files
ls cmd/ internal/ pkg/ 2>/dev/null || echo "Standard Go project structure expected"
test -f go.mod && echo "Go module found" || echo "go.mod required"

# Go Code Quality Validation
go mod tidy                      # Clean dependencies
go vet ./...                    # Static analysis
golangci-lint run || echo "golangci-lint recommended"
go test ./...                   # Run tests

# Go Build Validation
go build ./...                  # Verify compilation
GOOS=linux GOARCH=amd64 go build -o bootstrap cmd/api/main.go  # Lambda build

# AWS Lambda Container Validation
test -f Dockerfile && echo "Dockerfile found" || echo "Dockerfile needed for Lambda"
docker build -t test-api . 2>/dev/null || echo "Docker build should succeed"

# Go Security Validation
gosec ./... || echo "Security scan recommended"
go mod verify || echo "Module verification failed"
```

*** CRITICAL: Do extensive Go ecosystem web research before writing the PRP ***
*** Use WebSearch tool extensively to understand Go patterns and AWS integration ***

## Output

Save as: `PRPs/go-api-{feature-name}.md`

## Quality Checklist

- [ ] Extensive Go ecosystem research completed
- [ ] Go REST framework patterns thoroughly reviewed  
- [ ] AWS Lambda Go container patterns documented
- [ ] Real-world Go API examples identified
- [ ] Go testing and validation patterns researched
- [ ] All web research findings documented in PRP
- [ ] Go-specific gotchas and best practices captured
- [ ] AWS integration patterns and security considerations included

Score the PRP on a scale of 1-10 (confidence level for creating production-ready, secure Go APIs with proper AWS integration).

Remember: The goal is creating comprehensive, battle-tested Go API implementations that leverage Go's strengths while integrating seamlessly with AWS services through proper containerized Lambda deployment.