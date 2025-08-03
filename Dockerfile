# Multi-stage build for AWS Lambda container deployment
FROM golang:1.22-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application for Linux Lambda environment
# CGO_ENABLED=0 for static linking
# GOOS=linux for Lambda runtime
# -tags lambda.norpc for optimized Lambda builds
# -ldflags="-s -w" to strip debug info and reduce size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -tags lambda.norpc \
    -ldflags="-s -w" \
    -o main \
    cmd/api/main.go

# Final stage - AWS Lambda provided runtime
FROM public.ecr.aws/lambda/provided:al2023

# Copy the binary from builder stage
COPY --from=builder /app/main ./main

# Copy timezone data for time operations
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy CA certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Set environment variables for optimal Lambda performance
ENV GIN_MODE=release
ENV AWS_LAMBDA_RUNTIME_API=""

# Lambda runtime will execute this binary
ENTRYPOINT ["./main"]