# Mount Reverse Proxy Example

This example demonstrates how to mount reverse proxies using Lokstra's `MountReverseProxy` functionality, enabling you to forward requests to external services or microservices.

## Features Demonstrated

1. **Basic Reverse Proxy** - Forward requests to external APIs
2. **Proxy with Middleware** - Add authentication, logging, and other middleware
3. **Multiple Proxy Targets** - Route different paths to different services
4. **Middleware Override** - Option to bypass global middleware for specific proxies
5. **Request Forwarding** - Preserve headers, query parameters, and request bodies

## How Reverse Proxy Works

```go
// Basic proxy: /api/external/* → https://jsonplaceholder.typicode.com/*
app.MountReverseProxy("/api/external", "https://jsonplaceholder.typicode.com", false)

// Proxy with middleware: Authentication + Logging
app.MountReverseProxy("/api/secure", "https://httpbin.org", false, "auth", "logger")

// Proxy with middleware override (bypasses global middleware)
app.MountReverseProxy("/api/fast", "https://httpbin.org", true, "custom_middleware")
```

## Directory Structure

```
05_mount_reverse_proxy/
├── main.go              # Main application with proxy configuration
└── README.md            # This file
```

## Proxy Configuration

### 1. External API Proxy
- **Route**: `/api/external/*`
- **Target**: `https://jsonplaceholder.typicode.com`
- **Middleware**: `proxy_logger`
- **Purpose**: Demonstrate basic proxy functionality

### 2. Secure API Proxy  
- **Route**: `/api/secure/*`
- **Target**: `https://httpbin.org`
- **Middleware**: `proxy_auth`, `proxy_logger`
- **Purpose**: Show authenticated proxy access

### 3. GitHub API Proxy
- **Route**: `/api/github/*`
- **Target**: `https://api.github.com`
- **Middleware**: Global middleware only
- **Purpose**: Demonstrate public API proxying

## How to Run

```bash
go run main.go
```

The server will start on port 8080.

## Test the Reverse Proxy

### Basic Proxy (JSONPlaceholder API)
```bash
# Get posts (proxied to jsonplaceholder.typicode.com/posts)
curl http://localhost:8080/api/external/posts

# Get specific post
curl http://localhost:8080/api/external/posts/1

# Get users
curl http://localhost:8080/api/external/users

# Get user details
curl http://localhost:8080/api/external/users/1
```

### Secure Proxy (HTTPBin - requires API key)
```bash
# Get request info (requires X-API-Key header)
curl -H "X-API-Key: secret-key" http://localhost:8080/api/secure/get

# JSON response
curl -H "X-API-Key: secret-key" http://localhost:8080/api/secure/json

# Without API key (should fail)
curl http://localhost:8080/api/secure/get
```

### GitHub API Proxy
```bash
# Get user info
curl http://localhost:8080/api/github/users/octocat

# Get repository info
curl http://localhost:8080/api/github/repos/microsoft/vscode

# Search repositories
curl http://localhost:8080/api/github/search/repositories?q=lokstra
```

### Local API Endpoints
```bash
# Server information
curl http://localhost:8080/

# Health check
curl http://localhost:8080/health

# Local API status
curl http://localhost:8080/api/local/status
```

## Middleware Examples

### Proxy Logger Middleware
```go
ctx.RegisterMiddlewareFunc("proxy_logger", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
    return func(ctx *lokstra.Context) error {
        lokstra.Logger.Infof("[PROXY] Forwarding %s %s", ctx.Request.Method, ctx.Request.URL.Path)
        return next(ctx)
    }
})
```

### Proxy Authentication Middleware
```go
ctx.RegisterMiddlewareFunc("proxy_auth", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
    return func(ctx *lokstra.Context) error {
        apiKey := ctx.Request.Header.Get("X-API-Key")
        if apiKey != "secret-key" {
            return ctx.ErrorBadRequest("API key required")
        }
        return next(ctx)
    }
})
```

## Key Features

### 1. Path Forwarding
- Request to `/api/external/posts/1`
- Gets forwarded to `https://jsonplaceholder.typicode.com/posts/1`
- Path prefix is automatically stripped

### 2. Header Preservation
- All request headers are forwarded to the target
- Response headers are returned to the client
- Custom headers can be added via middleware

### 3. Method Support
- Supports all HTTP methods (GET, POST, PUT, DELETE, etc.)
- Request body is forwarded for POST/PUT requests
- Query parameters are preserved

### 4. Middleware Integration
- Apply authentication, logging, rate limiting
- Transform requests/responses
- Add custom headers or modify behavior

## Real-World Use Cases

### 1. Microservices Gateway
```go
// User service
app.MountReverseProxy("/api/users", "http://user-service:8001", false, "auth")

// Order service  
app.MountReverseProxy("/api/orders", "http://order-service:8002", false, "auth")

// Notification service
app.MountReverseProxy("/api/notifications", "http://notification-service:8003", false, "auth")
```

### 2. API Aggregation
```go
// External payment provider
app.MountReverseProxy("/api/payments", "https://api.stripe.com/v1", false, "stripe_auth")

// External email service
app.MountReverseProxy("/api/email", "https://api.sendgrid.com/v3", false, "sendgrid_auth")
```

### 3. Legacy System Integration
```go
// Legacy SOAP service (with transformation middleware)
app.MountReverseProxy("/api/legacy", "http://legacy-system:8080", false, "soap_transform")

// Legacy database API
app.MountReverseProxy("/api/db", "http://legacy-db-api:3000", false, "db_auth")
```

## Advanced Configuration

### Load Balancing (Multiple Targets)
While `MountReverseProxy` supports single targets, you can implement load balancing through middleware:

```go
ctx.RegisterMiddlewareFunc("load_balancer", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
    targets := []string{
        "http://service1:8001",
        "http://service2:8001", 
        "http://service3:8001",
    }
    
    return func(ctx *lokstra.Context) error {
        // Select target based on load balancing algorithm
        target := selectTarget(targets)
        // Modify request to use selected target
        return next(ctx)
    }
})
```

### Request/Response Transformation
```go
ctx.RegisterMiddlewareFunc("transform", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
    return func(ctx *lokstra.Context) error {
        // Transform request before forwarding
        transformRequest(ctx)
        
        err := next(ctx)
        
        // Transform response before returning
        transformResponse(ctx)
        return err
    }
})
```

## Security Considerations

1. **API Key Management**: Store API keys securely, not in code
2. **Rate Limiting**: Implement rate limiting to prevent abuse
3. **Request Validation**: Validate requests before forwarding
4. **Response Filtering**: Filter sensitive data from responses
5. **Logging**: Log all proxy requests for monitoring and debugging

## Expected Behavior

1. **Transparent Forwarding**: Requests are seamlessly forwarded to target services
2. **Error Handling**: Errors from target services are properly returned
3. **Header Preservation**: All relevant headers are maintained
4. **Middleware Execution**: Middleware runs before forwarding requests
5. **Path Transformation**: URL paths are correctly transformed for target services

This setup is perfect for building API gateways, microservice aggregators, and integration layers.
