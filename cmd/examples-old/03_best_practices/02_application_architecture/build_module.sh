#!/bin/bash

# Build script for user management module
# This script compiles the user management module as a shared library (.so file)

echo "Building user management module..."

# Set build flags
export CGO_ENABLED=1
export GOOS=linux
export GOARCH=amd64

# Build the module as a shared library
go build -buildmode=plugin -o user_management.so ./modules/user_management/

if [ $? -eq 0 ]; then
    echo "✅ User management module built successfully: user_management.so"
    
    # Verify the plugin
    echo "📁 Plugin file information:"
    ls -la user_management.so
    
    echo "🔍 Plugin symbols:"
    nm -D user_management.so | grep -E "(RegisterModule|RequiredServices|CreateServices)"
else
    echo "❌ Failed to build user management module"
    exit 1
fi

echo ""
echo "📝 Usage:"
echo "  1. Update lokstra.yaml to reference the built plugin:"
echo "     modules:"
echo "       - name: \"user_management\""
echo "       path: \"./user_management.so\""
echo "       entry: \"RegisterModule\""
echo ""
echo "  2. Run your application:"
echo "     go run main.go"
