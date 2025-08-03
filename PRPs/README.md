# Go REST API Template

🚀 **Production-ready Go REST API template with AWS Lambda containerized deployment, JWT authentication, DynamoDB integration, and S3 file handling.**

This template enables rapid development of secure, scalable Go APIs with comprehensive AWS service integration, following Go best practices and context engineering principles.

## 🚀 Quick Start - Copy Template First

**Start by copying this template to your project directory:**

```bash
# Copy template to your project location
python copy_template.py ~/projects/my-go-api

# Navigate to your new project
cd ~/projects/my-go-api

# You now have a complete Go API development environment!
```

The copy script includes:

- ✅ Complete project structure with Go conventions
- ✅ Working code examples and patterns
- ✅ Specialized context engineering commands
- ✅ Comprehensive documentation and guides
- ✅ AWS integration patterns and configurations

## 📋 PRP Framework Workflow

This template follows the **3-step PRP (Problem-Research-Plan) framework** for systematic API development:

### Step 1: Define Requirements (INITIAL.md)

```bash
# Create your API specification
cp PRPs/INITIAL.md PRPs/my-api-INITIAL.md
# Edit with your specific API requirements, features, and constraints
```

### Step 2: Generate Research Plan

```bash
# Generate comprehensive PRP with Go-specific research
/generate-go-api-prp PRPs/my-api-INITIAL.md
# → Creates PRPs/go-api-my-feature.md with research and implementation plan
```

### Step 3: Execute Implementation

```bash
# Execute the PRP to build your production-ready API
/execute-go-api-prp PRPs/go-api-my-feature.md
# → Implements complete Go API with testing, deployment, and documentation
```

## 📁 Template Structure

```sh
go-api/
├── 📜 CLAUDE.md                           # Go-specific global rules and patterns
├── 🔧 .claude/commands/
│   ├── generate-go-api-prp.md             # Go API PRP generation command
│   └── execute-go-api-prp.md              # Go API PRP execution command
├── 📋 PRPs/
│   ├── templates/
│   │   └── prp_go_api_base.md             # Go API base PRP template
│   └── INITIAL.md                         # Example API feature request
├── 💡 examples/
│   ├── basic_api/                         # Simple Gin REST API
│   ├── lambda_container/                  # AWS Lambda containerized API
│   ├── jwt_auth/                          # JWT authentication with Cognito
│   ├── dynamodb_crud/                     # DynamoDB data operations
│   ├── s3_upload/                         # S3 file upload handling
│   └── github_workflow/                   # CI/CD pipeline configurations
├── 🐍 copy_template.py                    # Template deployment script
└── 📖 README.md                           # This comprehensive guide
```

## 🎯 What You Can Build

This template is optimized for building:

### 🔐 **Secure APIs**

- JWT authentication with AWS Cognito integration
- Role-based access control and authorization
- Input validation and sanitization
- Security headers and CORS configuration

### 📊 **Data-Driven APIs**

- DynamoDB integration with query optimization
- S3 file storage and retrieval
- Efficient pagination and filtering
- Proper error handling and recovery

### ☁️ **Cloud-Native Architecture**

- AWS Lambda containerized deployment
- API Gateway integration
- CloudWatch logging and monitoring
- Infrastructure as Code patterns

### 🧪 **Production-Ready Quality**

- Comprehensive test suites (unit + integration)
- GitHub Actions CI/CD pipelines
- Docker multi-stage builds
- Performance monitoring and alerting

## 📚 Key Features

### 🏗️ **Go Framework Integration**

- **[Gin](https://gin-gonic.com/)**: High-performance HTTP framework (40x faster than Martini)
- **[AWS SDK for Go v2](https://aws.github.io/aws-sdk-go-v2/)**: Modern AWS service integration
- **[golang-jwt](https://github.com/golang-jwt/jwt)**: Secure JWT token handling
- **Standard Project Layout**: Following [golang-standards/project-layout](https://github.com/golang-standards/project-layout)

### 🔒 **Security & Authentication**

- **AWS Cognito JWT Verification**: Automatic JWKS key rotation and caching
- **Middleware Patterns**: Reusable authentication and authorization middleware
- **Input Validation**: Comprehensive request validation with proper error responses
- **Security Headers**: CORS, HSTS, and other security headers implementation

### ☁️ **AWS Service Integration**

- **Lambda Containers**: Optimized Docker builds for serverless deployment
- **DynamoDB**: Efficient NoSQL data operations with pagination
- **S3**: File upload/download with metadata and security validation
- **IAM**: Least-privilege permission patterns

### 🧪 **Testing & Quality Assurance**

- **Table-Driven Tests**: Go-idiomatic testing patterns
- **HTTP Handler Testing**: Complete request/response validation
- **Mock Implementations**: Testable code architecture
- **Integration Tests**: AWS service interaction testing

### 🚀 **Development & Deployment**

- **Local Development**: Docker Compose for local AWS services
- **CI/CD Pipelines**: GitHub Actions for automated testing and deployment
- **Environment Management**: Configuration for dev/staging/production
- **Monitoring Setup**: CloudWatch integration and alerting

## 🔍 Examples Included

### 1. **Basic API** (`examples/basic_api/`)

```bash
cd examples/basic_api
go run main.go
# → Simple REST API with CRUD operations
```

**Features**: Gin router, JSON handling, error responses, health checks

### 2. **Lambda Container** (`examples/lambda_container/`)

```bash
cd examples/lambda_container
docker build -t go-lambda-api .
docker run -p 9000:8080 go-lambda-api
# → Containerized Lambda function with dual-mode operation
```

**Features**: AWS Lambda integration, container optimization, environment detection

### 3. **JWT Authentication** (`examples/jwt_auth/`)

```bash
export COGNITO_USER_POOL_ID="us-east-1_EXAMPLE123"
cd examples/jwt_auth
go run main.go
# → JWT authentication with AWS Cognito JWKS verification
```

**Features**: AWS Cognito integration, JWKS caching, role-based access

## 🚫 Common Gotchas

### **Go-Specific Pitfalls**

- ⚠️ **Error Handling**: Always check and handle errors explicitly - don't ignore them
- ⚠️ **Goroutine Leaks**: Ensure proper cleanup and context cancellation
- ⚠️ **Interface Design**: Keep interfaces small and focused ("accept interfaces, return structs")
- ⚠️ **Memory Management**: Be aware of allocations in hot paths and HTTP handlers

### **AWS Lambda Gotchas**

- ⚠️ **Cold Starts**: Minimize initialization code and container size
- ⚠️ **Timeout Management**: Design for Lambda's 15-minute maximum execution time
- ⚠️ **Memory Configuration**: Right-size memory allocation for cost optimization
- ⚠️ **Environment Variables**: Use for configuration, never hardcode credentials

### **DynamoDB Pitfalls**

- ⚠️ **Hot Partitions**: Design partition keys to distribute load evenly
- ⚠️ **Query vs Scan**: Prefer Query operations over Scan for performance
- ⚠️ **Item Size Limits**: Keep items under 400KB, design for efficient storage
- ⚠️ **Eventually Consistent**: Use strongly consistent reads when data consistency is critical

### **Security Gotchas**

- ⚠️ **JWT Validation**: Always validate ALL claims (exp, aud, iss), not just signature
- ⚠️ **Input Sanitization**: Validate all inputs including file uploads and query parameters
- ⚠️ **Error Information**: Don't leak sensitive information in error responses
- ⚠️ **CORS Configuration**: Configure CORS properly for web application integration

## 🛠️ Quick Development Commands

### **Go Project Commands**

```bash
# Initialize and manage dependencies
go mod tidy
go mod download
go mod verify

# Code quality and testing
go vet ./...
golangci-lint run
go test ./...
go test -race ./...
go test -cover ./...

# Building
go build ./...
GOOS=linux GOARCH=amd64 go build -o bootstrap cmd/api/main.go
```

### **Docker Commands**

```bash
# Build and test containers
docker build -t my-go-api .
docker run -p 8080:8080 my-go-api

# AWS Lambda local testing
docker run -p 9000:8080 -e AWS_LAMBDA_FUNCTION_NAME=test my-go-api
```

### **AWS CLI Commands**

```bash
# Deploy to Lambda
aws ecr get-login-password | docker login --username AWS --password-stdin <ecr-uri>
docker tag my-go-api:latest <ecr-uri>/my-go-api:latest
docker push <ecr-uri>/my-go-api:latest
aws lambda update-function-code --function-name my-go-api --image-uri <ecr-uri>/my-go-api:latest
```

## 🎯 Production Checklist

### **Before Deployment**

- [ ] All tests passing (`go test ./...`)
- [ ] Security scan clean (`gosec ./...`)
- [ ] Linting passing (`golangci-lint run`)
- [ ] Docker build optimized
- [ ] Environment variables configured
- [ ] IAM roles with least-privilege permissions
- [ ] Monitoring and logging configured
- [ ] Error handling and recovery tested
- [ ] Performance benchmarks acceptable
- [ ] Documentation updated

### **Post-Deployment**

- [ ] Health checks passing
- [ ] CloudWatch logs flowing
- [ ] Metrics and alarms configured
- [ ] Security headers validated
- [ ] API documentation published
- [ ] Load testing completed
- [ ] Backup and recovery procedures tested
- [ ] Incident response plan updated

---

**🌟 Happy building with Go APIs!**

This template provides everything you need to create production-ready Go REST APIs with AWS integration. Start with the examples, follow the PRP workflow, and build amazing APIs!
