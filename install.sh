#!/bin/bash

# License Management System Installation Script

echo "🚀 License Management System Installation"
echo "=========================================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or higher."
    echo ""
    echo "Installation instructions:"
    echo "1. Visit https://golang.org/dl/"
    echo "2. Download Go 1.21 or higher for your platform"
    echo "3. Follow the installation instructions"
    echo "4. Verify installation with: go version"
    echo ""
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
REQUIRED_VERSION="1.21"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo "❌ Go version $GO_VERSION is too old. Please install Go 1.21 or higher."
    exit 1
fi

echo "✅ Go version $GO_VERSION is installed"

# Install dependencies
echo "📦 Installing dependencies..."
go mod download

if [ $? -eq 0 ]; then
    echo "✅ Dependencies installed successfully"
else
    echo "❌ Failed to install dependencies"
    exit 1
fi

# Build the application
echo "🔨 Building application..."
go build -o license-management-api main.go

if [ $? -eq 0 ]; then
    echo "✅ Application built successfully"
else
    echo "❌ Failed to build application"
    exit 1
fi

# Create data directory
mkdir -p data

echo ""
echo "🎉 Installation completed successfully!"
echo ""
echo "To run the application:"
echo "  ./license-management-api"
echo ""
echo "Or with Go:"
echo "  go run main.go"
echo ""
echo "The API will be available at: http://localhost:8080"
echo "Swagger documentation: http://localhost:8080/swagger/index.html"
echo ""
echo "Default admin credentials:"
echo "  Email: admin@example.com"
echo "  Password: admin123"
echo ""
