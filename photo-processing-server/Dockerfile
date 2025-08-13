# Multi-stage build for Photo Processing Server
FROM golang:1.21-alpine AS builder

# Install build dependencies for OpenCV
RUN apk add --no-cache \
    build-base \
    pkgconfig \
    opencv-dev \
    opencv-contrib-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o photo-processor ./cmd/server

# Final stage - runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    opencv \
    opencv-contrib \
    ca-certificates \
    tzdata

# Create app user
RUN addgroup -g 1001 -S appuser && \
    adduser -u 1001 -S appuser -G appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/photo-processor .

# Create data directories
RUN mkdir -p /app/data/{photos,processed,temp} && \
    chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ./photo-processor --test || exit 1

# Default command
CMD ["./photo-processor"]