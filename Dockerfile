# Multi-stage Docker build for ZeroUI

# Build stage
FROM golang:1.21-alpine AS builder

# Install git and ca-certificates (needed for Go modules)
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.version=$(git describe --tags --always --dirty) -X main.commit=$(git rev-parse --short HEAD) -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o zeroui .

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates

# Create a non-root user
RUN adduser -D -s /bin/sh zeroui

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/zeroui .

# Change ownership to non-root user
RUN chown zeroui:zeroui zeroui

# Switch to non-root user
USER zeroui

# Expose port if needed (for future web interface)
# EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ./zeroui --version || exit 1

# Set the binary as the entrypoint
ENTRYPOINT ["./zeroui"]

# Default command
CMD ["--help"]
