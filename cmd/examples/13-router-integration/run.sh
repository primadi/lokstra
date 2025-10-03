#!/bin/bash

# Lokstra Router Integration Test Script
# This script helps you run different deployment modes easily

DEPLOYMENT_ID=""
SERVER_NAME="all"

# Function to show usage
show_usage() {
    echo ""
    echo "🚀 Lokstra Router Integration Test Runner"
    echo ""
    echo "Usage:"
    echo "  ./run.sh [deployment-id] [server-name]"
    echo ""
    echo "Available Deployment Types:"
    echo "  monolith-single-port   - All routers in one server on one port (:8081)"
    echo "  monolith-multi-port    - All routers in one server on different ports (:8081, :8082)"
    echo "  microservice           - Each router on its own server (:8082, :8083)"
    echo ""
    echo "Examples:"
    echo "  ./run.sh monolith-single-port"
    echo "  ./run.sh monolith-multi-port"
    echo "  ./run.sh microservice"
    echo "  ./run.sh microservice product-service"
    echo ""
}

# Parse arguments
if [ $# -ge 1 ]; then
    DEPLOYMENT_ID="$1"
fi

if [ $# -ge 2 ]; then
    SERVER_NAME="$2"
fi

# Show usage if no deployment ID provided
if [ -z "$DEPLOYMENT_ID" ]; then
    show_usage
    read -p "Enter deployment ID: " DEPLOYMENT_ID
    if [ -z "$DEPLOYMENT_ID" ]; then
        echo "❌ No deployment ID provided. Exiting."
        exit 1
    fi
fi

# Validate deployment ID
case $DEPLOYMENT_ID in
    "monolith-single-port"|"monolith-multi-port"|"microservice")
        ;;
    *)
        echo "❌ Invalid deployment ID: $DEPLOYMENT_ID"
        echo "Valid options: monolith-single-port, monolith-multi-port, microservice"
        exit 1
        ;;
esac

echo ""
echo "🎯 Starting deployment: $DEPLOYMENT_ID"
echo "🖥️  Server filter: $SERVER_NAME"
echo ""

# Set environment variables
export DEPLOYMENT_ID="$DEPLOYMENT_ID"
export SERVER_NAME="$SERVER_NAME"

# Change to the correct directory
cd "$(dirname "$0")"

# Build and run
echo "🔨 Building application..."
go build -o router-integration .

if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi

echo "✅ Build successful!"
echo ""
echo "🚀 Starting server(s)..."

# Run the application
./router-integration