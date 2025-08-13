#!/bin/bash

echo "=== EnDeCode Web UI Build Script ==="
echo ""

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "Error: Node.js is not installed. Please install Node.js first."
    exit 1
fi

# Check if Go is installed  
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go first."
    exit 1
fi

echo "✅ Node.js version: $(node --version)"
echo "✅ Go version: $(go version)"
echo ""

# Build React frontend
echo "📦 Building React frontend..."
cd web/frontend

# Install dependencies if node_modules doesn't exist
if [ ! -d "node_modules" ]; then
    echo "📥 Installing npm dependencies..."
    npm install
fi

# Build for production
echo "🏗️ Building React app..."
npm run build

if [ $? -ne 0 ]; then
    echo "❌ React build failed!"
    exit 1
fi

echo "✅ React app built successfully!"
echo ""

# Go back to main directory
cd ../..

# Download Go dependencies
echo "📥 Downloading Go dependencies..."
go mod tidy

if [ $? -ne 0 ]; then
    echo "❌ Failed to download Go dependencies!"
    exit 1
fi

# Build Go server
echo "🏗️ Building Go web server..."
go build -o bin/endecode-web-server cmd/web-server/main.go

if [ $? -ne 0 ]; then
    echo "❌ Go build failed!"
    exit 1
fi

echo "✅ Go server built successfully!"
echo ""

echo "🎉 Build completed successfully!"
echo ""
echo "To run the web server:"
echo "  ./bin/endecode-web-server"
echo ""
echo "Then open: http://localhost:8080"
echo "API: http://localhost:8080/api"
echo "WebSocket: ws://localhost:8080/ws"
echo ""
echo "=== Build Complete ==="