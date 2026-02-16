#!/bin/bash

# TucanBIT Backend Deployment Script
# This script builds and deploys the backend to the server

set -e

# Configuration
SERVER_IP="13.48.56.1317"
SERVER_USER="ubuntu"
KEY_PATH="~/Developer/Upwork/Tucanbit/Tucanbit/TucanBIT.pem"
IMAGE_NAME="tucanbit-backend-dev:latest"
CONTAINER_NAME="tucanbit-backend-dev"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 TucanBIT Backend Deployment${NC}"
echo "=================================="

# Step 1: Build the Docker image
echo -e "${BLUE}📦 Building Docker image...${NC}"
docker build -t $IMAGE_NAME .
echo -e "${GREEN}✅ Docker image built successfully!${NC}"

# Step 2: Save the image to a tar file
echo -e "${BLUE}💾 Saving Docker image to tar file...${NC}"
docker save $IMAGE_NAME -o tucanbit-backend-dev.tar
echo -e "${GREEN}✅ Docker image saved to tucanbit-backend-dev.tar${NC}"

# Step 3: Upload the image to the server
echo -e "${BLUE}📤 Uploading image to server...${NC}"
scp -i $KEY_PATH tucanbit-backend-dev.tar $SERVER_USER@$SERVER_IP:~/
echo -e "${GREEN}✅ Image uploaded to server!${NC}"

# Step 4: Deploy on the server
echo -e "${BLUE}🔄 Deploying on server...${NC}"
ssh -i $KEY_PATH $SERVER_USER@$SERVER_IP << 'EOF'
    # Load the Docker image
    echo "Loading Docker image..."
    docker load -i tucanbit-backend-dev.tar
    
    # Stop and remove existing container
    echo "Stopping existing container..."
    docker stop tucanbit-backend-dev 2>/dev/null || true
    docker rm tucanbit-backend-dev 2>/dev/null || true
    
    # Logs directory already created on server
    echo "Using existing logs directory..."
    
    # Run the new container
    echo "Starting new container..."
    docker run -d \
        --name tucanbit-backend-dev \
        --restart unless-stopped \
        -p 8094:8094 \
        -v /opt/tucanbit/logs:/app/logs \
        -e LOG_LEVEL=info \
        -e APP_PORT=8094 \
        -e DB_URL="postgres://tucanbit:5kj0YmV5FKKpU9D50B7yH5A@172.31.36.46:5433/tucanbit?sslmode=disable&connect_timeout=10" \
        --network host \
        tucanbit-backend-dev:latest
    
    # Wait a moment for the container to start
    sleep 5
    
    # Check container status
    echo "Checking container status..."
    docker ps | grep tucanbit-backend-dev
    
    # Show recent logs
    echo "Recent logs:"
    docker logs --tail 20 tucanbit-backend-dev
    
    # Clean up
    echo "Cleaning up..."
    rm -f tucanbit-backend-dev.tar
EOF

echo -e "${GREEN}✅ Deployment completed!${NC}"
echo -e "${YELLOW}🌐 Backend should be available at: http://$SERVER_IP:8094${NC}"

# Clean up local tar file
rm -f tucanbit-backend-dev.tar
echo -e "${GREEN}✅ Local cleanup completed!${NC}"
