# Multi-stage Docker build for ZeroUI
# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags "-w -s \
    -X 'github.com/mrtkrcm/ZeroUI/internal/version.Version=${VERSION}' \
    -X 'github.com/mrtkrcm/ZeroUI/internal/version.Commit=${COMMIT}' \
    -X 'github.com/mrtkrcm/ZeroUI/internal/version.BuildTime=${BUILD_TIME}'" \
    -o zeroui .

# Final stage - minimal image
FROM alpine:3.19

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -S appgroup && \
    adduser -S appuser -G appgroup

# Set timezone
ENV TZ=UTC

# Create app directory
WORKDIR /app

# Copy binary and set permissions
COPY --from=builder /app/zeroui .
RUN chmod +x zeroui

# Copy configuration directory structure
RUN mkdir -p /home/appuser/.config/zeroui/{apps,presets,backups} && \
    chown -R appuser:appgroup /home/appuser/.config

# Create sample configuration
COPY configs/* /home/appuser/.config/zeroui/apps/
RUN chown -R appuser:appgroup /home/appuser/.config

# Switch to non-root user
USER appuser:appgroup

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ./zeroui --version || exit 1

# Set default command
ENTRYPOINT ["./zeroui"]
CMD ["--help"]

# Metadata
LABEL maintainer="ZeroUI Team" \
      description="Zero-configuration UI toolkit manager for developers" \
      version="${VERSION}" \
      org.opencontainers.image.title="ZeroUI" \
      org.opencontainers.image.description="Zero-configuration UI toolkit manager" \
      org.opencontainers.image.url="https://github.com/mrtkrcm/ZeroUI" \
      org.opencontainers.image.source="https://github.com/mrtkrcm/ZeroUI" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_TIME}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.vendor="ZeroUI" \
      org.opencontainers.image.licenses="MIT"