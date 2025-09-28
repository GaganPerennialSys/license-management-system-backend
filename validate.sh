#!/bin/bash

# License Management System Validation Script

echo "🔍 Validating License Management System"
echo "======================================="

# Check if required files exist
REQUIRED_FILES=(
    "main.go"
    "go.mod"
    "openapi.yaml"
    "Dockerfile"
    "docker-compose.yml"
    "README.md"
    "internal/config/config.go"
    "internal/database/database.go"
    "internal/models/user.go"
    "internal/models/customer.go"
    "internal/models/subscription_pack.go"
    "internal/models/subscription.go"
    "internal/middleware/auth.go"
    "internal/handlers/base.go"
    "internal/handlers/user.go"
    "internal/handlers/customer.go"
    "internal/handlers/subscription_pack.go"
    "internal/handlers/subscription.go"
    "internal/handlers/sdk.go"
)

echo "📁 Checking required files..."
MISSING_FILES=()

for file in "${REQUIRED_FILES[@]}"; do
    if [ -f "$file" ]; then
        echo "✅ $file"
    else
        echo "❌ $file (missing)"
        MISSING_FILES+=("$file")
    fi
done

if [ ${#MISSING_FILES[@]} -eq 0 ]; then
    echo ""
    echo "✅ All required files are present"
else
    echo ""
    echo "❌ Missing files:"
    for file in "${MISSING_FILES[@]}"; do
        echo "   - $file"
    done
    exit 1
fi

# Check Go module structure
echo ""
echo "📦 Checking Go module structure..."

if [ -f "go.mod" ]; then
    echo "✅ go.mod exists"
    
    # Check if module name is set
    if grep -q "module cursor-ai-backend" go.mod; then
        echo "✅ Module name is set correctly"
    else
        echo "❌ Module name is not set correctly"
    fi
    
    # Check for required dependencies
    REQUIRED_DEPS=(
        "github.com/gin-gonic/gin"
        "github.com/golang-jwt/jwt/v5"
        "gorm.io/gorm"
        "gorm.io/driver/sqlite"
    )
    
    for dep in "${REQUIRED_DEPS[@]}"; do
        if grep -q "$dep" go.mod; then
            echo "✅ $dep"
        else
            echo "❌ $dep (missing)"
        fi
    done
else
    echo "❌ go.mod not found"
fi

# Check OpenAPI specification
echo ""
echo "📋 Checking OpenAPI specification..."

if [ -f "openapi.yaml" ]; then
    echo "✅ openapi.yaml exists"
    
    # Check for basic OpenAPI structure
    if grep -q "openapi: 3.0" openapi.yaml; then
        echo "✅ OpenAPI 3.0 specification"
    else
        echo "❌ Not a valid OpenAPI 3.0 specification"
    fi
    
    if grep -q "License Management System API" openapi.yaml; then
        echo "✅ API title is set"
    else
        echo "❌ API title is not set"
    fi
else
    echo "❌ openapi.yaml not found"
fi

# Check Docker configuration
echo ""
echo "🐳 Checking Docker configuration..."

if [ -f "Dockerfile" ]; then
    echo "✅ Dockerfile exists"
    
    if grep -q "FROM golang:1.21-alpine" Dockerfile; then
        echo "✅ Uses Go 1.21 Alpine base image"
    else
        echo "❌ Does not use Go 1.21 Alpine base image"
    fi
else
    echo "❌ Dockerfile not found"
fi

if [ -f "docker-compose.yml" ]; then
    echo "✅ docker-compose.yml exists"
else
    echo "❌ docker-compose.yml not found"
fi

# Check internal package structure
echo ""
echo "🏗️  Checking internal package structure..."

INTERNAL_DIRS=(
    "internal"
    "internal/config"
    "internal/database"
    "internal/models"
    "internal/middleware"
    "internal/handlers"
)

for dir in "${INTERNAL_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        echo "✅ $dir/"
    else
        echo "❌ $dir/ (missing)"
    fi
done

echo ""
echo "🎉 Validation completed!"
echo ""
echo "Next steps:"
echo "1. Install Go 1.21 or higher if not already installed"
echo "2. Run: ./install.sh"
echo "3. Start the application: ./license-management-api"
echo "4. Access the API at: http://localhost:8080"
echo "5. View documentation at: http://localhost:8080/swagger/index.html"
