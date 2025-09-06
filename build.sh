#!/bin/bash

# Build script for TucanBIT with network fallback options

set -e

echo "🚀 Starting TucanBIT build process..."

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo " Error: go.mod not found. Please run this script from the project root."
    exit 1
fi

# Function to check network connectivity
check_network() {
    echo "🔍 Checking network connectivity..."
    
    # Test Go proxy
    if curl -s --connect-timeout 5 https://proxy.golang.org > /dev/null 2>&1; then
        echo "Go proxy is accessible"
        return 0
    else
        echo "⚠️  Go proxy is not accessible"
        return 1
    fi
}

# Function to build locally
build_local() {
    echo "🔨 Building application locally..."
    
    # Set Go environment variables
    export GOPROXY=direct
    export GOSUMDB=off
    
    # Clean previous builds
    rm -f tucanbit
    
    # Download modules
    echo "📥 Downloading Go modules..."
    go mod download
    
    # Build the application
    echo "🔨 Building application..."
    go build -o tucanbit cmd/main.go
    
    if [ -f "tucanbit" ]; then
        echo "Local build successful!"
        echo "📁 Binary created: ./tucanbit"
    else
        echo " Local build failed!"
        exit 1
    fi
}

# Function to build with Docker
build_docker() {
    echo "🐳 Building with Docker..."
    
    # Check if Docker is running
    if ! docker info > /dev/null 2>&1; then
        echo " Docker is not running. Please start Docker and try again."
        exit 1
    fi
    
    # Build the Docker image
    echo "🔨 Building Docker image..."
    docker build -t tucanbit:latest .
    
    echo "Docker build successful!"
    echo "🐳 Image created: tucanbit:latest"
}

# Function to run with Docker Compose
run_docker_compose() {
    echo "🚀 Starting services with Docker Compose..."
    
    # Check if docker-compose is available
    if ! command -v docker-compose > /dev/null 2>&1; then
        echo " docker-compose not found. Please install it and try again."
        exit 1
    fi
    
    # Start the services
    docker-compose up --build -d
    
    echo "Services started successfully!"
    echo "Application should be available at http://localhost:8080"
    echo "📊 Check status with: docker-compose ps"
}

# Main execution
echo "🔧 TucanBIT Build Script"
echo "========================"

# Check network first
if check_network; then
    echo "Network is accessible, proceeding with normal build..."
else
    echo "⚠️  Network issues detected, using fallback options..."
fi

# Build locally first (this helps with dependency issues)
build_local

# Ask user what they want to do next
echo ""
echo "🎯 What would you like to do next?"
echo "1) Run the application locally (./tucanbit)"
echo "2) Build Docker image"
echo "3) Start all services with Docker Compose"
echo "4) Exit"

read -p "Enter your choice (1-4): " choice

case $choice in
    1)
        echo "🚀 Starting application locally..."
        ./tucanbit
        ;;
    2)
        build_docker
        ;;
    3)
        build_docker
        run_docker_compose
        ;;
    4)
        echo "👋 Goodbye!"
        exit 0
        ;;
    *)
        echo " Invalid choice. Exiting."
        exit 1
        ;;
esac 