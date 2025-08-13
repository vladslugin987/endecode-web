# Photo Processing Server - Makefile

.PHONY: help build test clean run docker-build docker-run dev

# Default target
help:
	@echo "Photo Processing Server - Available Commands:"
	@echo ""
	@echo "  build         - Build the application binary"
	@echo "  test          - Run basic tests (Caesar cipher, OpenCV)"
	@echo "  test-file     - Test with specific file (set FILE=/path/to/image.jpg)"
	@echo "  test-dir      - Test with directory (set DIR=/path/to/photos)"
	@echo "  clean         - Clean build artifacts"
	@echo "  run           - Run the server"
	@echo "  dev           - Development mode (build + test)"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run in Docker container"
	@echo "  deps          - Download and tidy dependencies"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make test"
	@echo "  make test-file FILE=./test.jpg"
	@echo "  make test-dir DIR=./test-photos"

# Build the application
build:
	@echo "Building Photo Processing Server..."
	@CGO_ENABLED=1 go build -o photo-processor ./cmd/server
	@echo "✓ Build complete: ./photo-processor"

# Download and tidy dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies updated"

# Run basic tests
test: build
	@echo "Running basic tests..."
	@./photo-processor --test

# Test with specific file
test-file: build
ifndef FILE
	@echo "Error: FILE parameter required"
	@echo "Usage: make test-file FILE=/path/to/image.jpg"
	@exit 1
endif
	@echo "Testing with file: $(FILE)"
	@./photo-processor --test --test-file "$(FILE)"

# Test with directory
test-dir: build
ifndef DIR
	@echo "Error: DIR parameter required"
	@echo "Usage: make test-dir DIR=/path/to/photos"
	@exit 1
endif
	@echo "Testing with directory: $(DIR)"
	@./photo-processor --test --test-dir "$(DIR)"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f photo-processor
	@rm -rf data/
	@echo "✓ Clean complete"

# Run the server
run: build
	@echo "Starting Photo Processing Server..."
	@./photo-processor

# Development mode
dev: deps build test
	@echo "✓ Development build and test complete"

# Docker build
docker-build:
	@echo "Building Docker image..."
	@docker build -t photo-processing-server:latest .
	@echo "✓ Docker image built: photo-processing-server:latest"

# Docker run
docker-run: docker-build
	@echo "Running Docker container..."
	@docker run --rm -it \
		-p 8080:8080 \
		-v $(PWD)/data:/app/data \
		photo-processing-server:latest

# Run tests in Docker
docker-test: docker-build
	@echo "Running tests in Docker..."
	@docker run --rm -it photo-processing-server:latest ./photo-processor --test

# Format code
fmt:
	@echo "Formatting Go code..."
	@go fmt ./...
	@echo "✓ Code formatted"

# Lint code (requires golangci-lint)
lint:
	@echo "Linting Go code..."
	@golangci-lint run
	@echo "✓ Lint complete"

# Security scan (requires gosec)
security:
	@echo "Running security scan..."
	@gosec ./...
	@echo "✓ Security scan complete"

# Full quality check
quality: fmt lint security test
	@echo "✓ Full quality check complete"

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@echo "✓ Development tools installed"

# Create test data structure
create-test-data:
	@echo "Creating test data structure..."
	@mkdir -p test-data/{photos,videos,mixed}
	@echo "Test photos directory" > test-data/photos/readme.txt
	@echo "Test videos directory" > test-data/videos/readme.txt
	@echo "Mixed media directory" > test-data/mixed/readme.txt
	@echo "✓ Test data structure created in ./test-data/"

# Show project info
info:
	@echo "Photo Processing Server - Project Info"
	@echo "======================================"
	@echo "Go version: $(shell go version)"
	@echo "Project: $(shell pwd)"
	@echo "Binary: ./photo-processor"
	@echo "Status: Core logic ported from Kotlin ✓"
	@echo ""
	@echo "Quick start:"
	@echo "  make dev          # Build and test everything"
	@echo "  make test         # Run basic tests"
	@echo "  make build && ./photo-processor --help"