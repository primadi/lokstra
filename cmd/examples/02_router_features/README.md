# Router Features Examples

This directory contains comprehensive examples demonstrating Lokstra's advanced router features. Each example focuses on specific routing capabilities and can be run independently.

## 📁 Examples Overview

### [01_group_and_nested_routes](./01_group_and_nested_routes/)
Demonstrates router grouping and nesting capabilities:
- Basic route groups with prefixes
- Nested groups (groups within groups)
- Group-level middleware application
- Route organization patterns

### [02_middleware_usage](./02_middleware_usage/)
Comprehensive middleware usage patterns:
- Global middleware (applied to all routes)
- Route-specific middleware
- Group-level middleware
- Multiple middleware chaining
- Context value storage and retrieval
- Middleware override capabilities

### [03_mount_static](./03_mount_static/)
Serving static files with Lokstra:
- Multiple static directories mounting
- Custom URL prefixes for static content
- Mixed content types (HTML, CSS, text)
- API endpoints coexisting with static files

### [04_mount_spa](./04_mount_spa/)
Single Page Application (SPA) serving:
- SPA mounting with fallback routing
- Client-side routing support
- API coexistence with SPA routes
- Browser history and navigation handling

### [05_mount_reverse_proxy](./05_mount_reverse_proxy/)
Reverse proxy functionality:
- Basic reverse proxy mounting
- Proxy with middleware (auth, logging)
- Multiple proxy targets
- External API integration examples

## 🚀 Quick Start

Navigate to any example directory and run:

```bash
cd 01_group_and_nested_routes
go run main.go
```

Each example includes:
- Complete working code
- Detailed README with explanations
- Test commands and endpoints
- Real-world usage patterns

## 🎯 Learning Path

### Beginner
1. **01_group_and_nested_routes** - Learn basic routing organization
2. **03_mount_static** - Understand static file serving

### Intermediate  
3. **02_middleware_usage** - Master middleware patterns
4. **04_mount_spa** - Build single page applications

### Advanced
5. **05_mount_reverse_proxy** - Implement microservice gateways

## 📚 Key Concepts Covered

### Route Organization
- Grouping related routes with prefixes
- Nested group structures
- Clean URL hierarchy design

### Middleware Patterns
- Request/response lifecycle hooks
- Authentication and authorization
- Logging and monitoring
- Error handling and recovery

### Static Content Serving
- File system mapping to URLs
- MIME type handling
- Cache headers and optimization

### Modern Web Applications
- SPA fallback routing
- Client-side navigation support
- API and frontend integration

### Service Integration
- Reverse proxy configuration
- Microservice routing
- External API aggregation

## 🛠️ Common Patterns

### API Gateway Pattern
```go
// Group API routes
api := app.Group("/api/v1", "cors", "auth")

// Mount microservices
api.MountReverseProxy("/users", "http://user-service:8001", false)
api.MountReverseProxy("/orders", "http://order-service:8002", false)

// Serve frontend
app.MountSPA("/", "./dist/index.html")
```

### Content Management Pattern
```go
// Static assets
app.MountStatic("/assets", http.Dir("./assets"))

// API endpoints  
app.Group("/api", "auth").GET("/content", getContentHandler)

// Admin interface
app.Group("/admin", "admin_auth").GET("/dashboard", adminHandler)

// Public site
app.MountSPA("/", "./public/index.html")
```

### Development Server Pattern
```go
// API development
app.Group("/api", "cors", "dev_logging").Handle("*", apiHandler)

// Static development files
app.MountStatic("/static", http.Dir("./src"))

// Hot reload support
app.MountSPA("/", "./src/index.html")
```

## 🔧 Testing All Examples

Run this script to test all examples:

```bash
#!/bin/bash
examples=("01_group_and_nested_routes" "02_middleware_usage" "03_mount_static" "04_mount_spa" "05_mount_reverse_proxy")

for example in "${examples[@]}"; do
    echo "Testing $example..."
    cd "$example"
    go build -o test.exe .
    if [ $? -eq 0 ]; then
        echo "✅ $example builds successfully"
        rm test.exe
    else
        echo "❌ $example build failed"
    fi
    cd ..
done
```

## 📖 Further Reading

- Check the main Lokstra documentation for detailed API reference
- Review `cmd/examples/01_basic_overview/` for fundamental concepts
- Explore `middleware/` directory for built-in middleware options
- See `modules/` directory for advanced service integrations

## 🤝 Contributing

When adding new router feature examples:

1. Create a new numbered directory (e.g., `06_custom_feature`)
2. Include complete working code in `main.go`
3. Add comprehensive `README.md` with explanations
4. Update this main README with the new example
5. Include sample files/directories if needed
6. Test the example thoroughly

Each example should be self-contained and demonstrate a specific router feature or pattern clearly.
