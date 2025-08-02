# Multi-stage build for AWS Lambda container
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 ensures a static binary
# GOOS=linux ensures Linux compatibility
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o bootstrap .

# Final stage - use AWS Lambda base image
FROM public.ecr.aws/lambda/provided:al2023-arm64

# Copy the binary from builder stage
COPY --from=builder /app/bootstrap /bootstrap

# Copy timezone data if needed
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy CA certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Set the executable as the entry point
ENTRYPOINT ["/bootstrap"]

# Alternative: Use scratch base for minimal image size
# FROM scratch
# COPY --from=builder /app/bootstrap /bootstrap
# COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# ENTRYPOINT ["/bootstrap"]