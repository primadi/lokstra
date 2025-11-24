#!/bin/bash
# Build script for Linux/Mac - ensures code generation before build
# Usage: ./build.sh [linux|windows|darwin]

set -e

TARGET_OS=${1:-$(uname -s | tr '[:upper:]' '[:lower:]')}
TARGET_OS=${TARGET_OS//darwin/darwin}
TARGET_OS=${TARGET_OS//linux/linux}

echo ""
echo "╔════════════════════════════════════════════════════╗"
echo "║  Lokstra Build Script - Code Gen + Build          ║"
echo "╚════════════════════════════════════════════════════╝"
echo ""

# Step 1: Generate code
echo "Step 1/4: Generating code (forced rebuild)..."
echo "  Running 'go run . --generate-only'..."

if go run . --generate-only ; then
    echo "  ✓ Code generation completed"
else
    echo "  ✗ Code generation failed!"
    exit 1
fi

echo ""

# Step 2: Tidy dependencies
echo "Step 2/4: Tidying dependencies..."
echo "  Running 'go mod tidy'..."

if go mod tidy ; then
    echo "  ✓ Dependencies tidied"
else
    echo "  ⚠ Warning: go mod tidy failed (continuing anyway)"
fi

echo ""

# Step 3: Build binary
echo "Step 3/4: Building binary for $TARGET_OS..."

case $TARGET_OS in
    linux)
        echo "  Building for Linux..."
        BINARY_NAME="app-linux"
        if GOOS=linux GOARCH=amd64 go build -o $BINARY_NAME . ; then
            echo "  ✓ Build successful: $BINARY_NAME"
        else
            echo "  ✗ Build failed!"
            exit 1
        fi
        ;;
    windows)
        echo "  Building for Windows..."
        BINARY_NAME="app-windows.exe"
        if GOOS=windows GOARCH=amd64 go build -o $BINARY_NAME . ; then
            echo "  ✓ Build successful: $BINARY_NAME"
        else
            echo "  ✗ Build failed!"
            exit 1
        fi
        ;;
    darwin)
        echo "  Building for macOS..."
        BINARY_NAME="app-darwin"
        if GOOS=darwin GOARCH=amd64 go build -o $BINARY_NAME . ; then
            echo "  ✓ Build successful: $BINARY_NAME"
        else
            echo "  ✗ Build failed!"
            exit 1
        fi
        ;;
    *)
        echo "  Building for current platform..."
        BINARY_NAME="app"
        if go build -o $BINARY_NAME . ; then
            echo "  ✓ Build successful: $BINARY_NAME"
        else
            echo "  ✗ Build failed!"
            exit 1
        fi
        ;;
esac

echo ""
echo "╔════════════════════════════════════════════════════╗"
echo "║  Build Complete!                                   ║"
echo "╚════════════════════════════════════════════════════╝"
echo ""
echo "Binary created: ./$BINARY_NAME"
echo ""
echo "Usage:"
echo "  ./build.sh         - Build for current platform"
echo "  ./build.sh linux   - Build for Linux"
echo "  ./build.sh windows - Build for Windows"
echo "  ./build.sh darwin  - Build for macOS"
echo ""
