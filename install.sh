#!/bin/bash

# Photo Processing Web Server Installation Script
# This script sets up everything needed to run the photo processing web application

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is installed
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed!"
        print_status "Please install Docker first:"
        echo "  - Linux: https://docs.docker.com/engine/install/"
        echo "  - macOS: https://docs.docker.com/docker-for-mac/install/"
        echo "  - Windows: https://docs.docker.com/docker-for-windows/install/"
        exit 1
    fi
    print_success "Docker is installed: $(docker --version)"
}

# Check if Docker Compose is installed
check_docker_compose() {
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        print_error "Docker Compose is not installed!"
        print_status "Please install Docker Compose: https://docs.docker.com/compose/install/"
        exit 1
    fi
    
    if command -v docker-compose &> /dev/null; then
        DOCKER_COMPOSE="docker-compose"
        print_success "Docker Compose is installed: $(docker-compose --version)"
    else
        DOCKER_COMPOSE="docker compose"
        print_success "Docker Compose is installed: $(docker compose version)"
    fi
}

# Create data directories
create_directories() {
    print_status "Creating data directories..."
    mkdir -p data/{photos,processed,temp}
    chmod 755 data data/photos data/processed data/temp
    print_success "Data directories created"
}

# Start the application
start_application() {
    print_status "Starting Photo Processing Web Server..."
    print_status "This may take a few minutes for the first run (building images)..."
    
    $DOCKER_COMPOSE up -d --build
    
    print_success "Photo Processing Web Server is starting!"
    print_status "Waiting for the application to be ready..."
    
    # Wait for health check to pass
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if $DOCKER_COMPOSE ps | grep -q "healthy"; then
            print_success "Application is ready!"
            break
        fi
        
        if [ $attempt -eq $max_attempts ]; then
            print_warning "Application may still be starting. Check logs if needed:"
            echo "  $DOCKER_COMPOSE logs photo-processing-web"
            break
        fi
        
        sleep 5
        ((attempt++))
    done
}

# Display access information
show_access_info() {
    echo ""
    echo "=================================="
    print_success "Installation Complete!"
    echo "=================================="
    echo ""
    print_status "Your Photo Processing Web Server is now running!"
    echo ""
    echo "📱 Web Interface: http://localhost:8090"
    echo "📁 Drop photos in: ./data/photos/"
    echo "📁 Processed files: ./data/processed/"
    echo ""
    echo "Useful commands:"
    echo "  View logs:     $DOCKER_COMPOSE logs -f"
    echo "  Stop server:   $DOCKER_COMPOSE down"
    echo "  Restart:       $DOCKER_COMPOSE restart"
    echo "  Update:        $DOCKER_COMPOSE up -d --build"
    echo ""
}

# Main installation process
main() {
    echo "=========================================="
    echo "  Photo Processing Web Server Installer"
    echo "=========================================="
    echo ""
    
    check_docker
    check_docker_compose
    create_directories
    start_application
    show_access_info
}

# Run main function
main "$@"