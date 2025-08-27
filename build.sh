#!/bin/bash
# Cross-platform build script for Lokstra

set -e

APP_NAME="lokstra"
VERSION=${VERSION:-"1.0.0"}
BUILD_DIR="./build"

# Create build directory
mkdir -p $BUILD_DIR

echo "Building Lokstra v$VERSION for multiple platforms..."

# Windows AMD64
echo "Building for Windows AMD64..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $BUILD_DIR/${APP_NAME}-${VERSION}-windows-amd64.exe .

# Linux AMD64
echo "Building for Linux AMD64..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $BUILD_DIR/${APP_NAME}-${VERSION}-linux-amd64 .

# Linux ARM64
echo "Building for Linux ARM64..."
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $BUILD_DIR/${APP_NAME}-${VERSION}-linux-arm64 .

# macOS AMD64
echo "Building for macOS AMD64..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $BUILD_DIR/${APP_NAME}-${VERSION}-darwin-amd64 .

# macOS ARM64 (Apple Silicon)
echo "Building for macOS ARM64..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $BUILD_DIR/${APP_NAME}-${VERSION}-darwin-arm64 .

echo "✓ All builds completed successfully!"
echo "Build artifacts are in: $BUILD_DIR"
ls -la $BUILD_DIR/
