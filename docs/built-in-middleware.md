# Built-in Middleware

Lokstra provides a comprehensive set of built-in middleware components located in the `/middleware` directory. These middleware components are production-ready and can be easily configured through YAML configuration or programmatically.

## Available Middleware

### 1. Body Limit Middleware (`body_limit`)

The Body Limit middleware restricts the size of request bodies to prevent memory exhaustion attacks and large payload abuse.

**Location**: `/middleware/body_limit/`

#### Features
- Request body size limiting based on configuration
- Content-Length header validation for early detection
- Customizable error messages and status codes
- Path pattern matching for selective enforcement
- Option to skip large payloads instead of rejecting them
- Convenience functions for common size limits

#### Configuration

```yaml
middleware:
  - name: "body_limit"
    enabled: true
    config:
      max_size: 10485760        # 10MB in bytes
      skip_large_payloads: false # Return error when limit exceeded
      message: "Request body too large"
      status_code: 413
      skip_on_path: 
        - "/uploads/*"          # Skip limit for upload endpoints
        - "/api/bulk/*"         # Skip for bulk operations
```

#### Configuration Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `max_size` | int64 | 10485760 (10MB) | Maximum request body size in bytes |
| `skip_large_payloads` | bool | false | Skip reading body instead of returning error |
| `message` | string | "Request body too large" | Custom error message |
| `status_code` | int | 413 | HTTP status code for errors |
| `skip_on_path` | []string | [] | Array of path patterns to skip |

#### Usage Examples

**Basic Usage:**
```yaml
middleware:
  - name: "body_limit"
    config:
      max_size: 5242880  # 5MB
```

**Custom Error Response:**
```yaml
middleware:
  - name: "body_limit"
    config:
      max_size: 1048576  # 1MB
      message: "File too large. Maximum size is 1MB"
      status_code: 400
```

**Skip Specific Paths:**
```yaml
middleware:
  - name: "body_limit"
    config:
      max_size: 2097152  # 2MB
      skip_on_path:
        - "/api/v1/files/*"
        - "/upload"
```

### 2. CORS Middleware (`cors`)

Cross-Origin Resource Sharing middleware for handling browser cross-origin requests.

**Location**: `/middleware/cors/`

#### Features
- Configurable allowed origins, methods, and headers
- Support for credentials and preflight requests
- Automatic OPTIONS request handling
- Wildcard and pattern-based origin matching

#### Configuration

```yaml
middleware:
  - name: "cors"
    enabled: true
    config:
      allowed_origins: ["*"]
      allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
      allowed_headers: ["*"]
      exposed_headers: ["Content-Length"]
      allow_credentials: false
      max_age: 86400
```

### 3. Gzip Compression Middleware (`gzipcompression`)

Response compression middleware to reduce bandwidth usage and improve performance.

**Location**: `/middleware/gzipcompression/`

#### Features
- Automatic gzip compression for responses
- Configurable compression levels
- Content-type filtering
- Size threshold configuration

#### Configuration

```yaml
middleware:
  - name: "gzipcompression"
    enabled: true
    config:
      level: 6                    # Compression level 1-9
      min_length: 1024           # Minimum response size to compress
      types: ["text/*", "application/json", "application/javascript"]
```

### 4. Recovery Middleware (`recovery`)

Panic recovery middleware that catches panics and returns proper HTTP error responses.

**Location**: `/middleware/recovery/`

#### Features
- Panic capture and recovery
- Stack trace logging (configurable)
- Custom error responses
- Environment-based configuration

#### Configuration

```yaml
middleware:
  - name: "recovery"
    enabled: true
    config:
      enable_stack_trace: true    # Include stack trace in logs
      print_stack: false         # Print stack trace to stdout
      log_stack: true            # Log stack trace
```

#### Environment-based Setup

```yaml
# Development
middleware:
  - name: "recovery"
    config:
      enable_stack_trace: true
      print_stack: true

# Production
middleware:
  - name: "recovery"
    config:
      enable_stack_trace: false
      print_stack: false
```

### 5. Request Logger Middleware (`request_logger`)

Comprehensive request logging middleware for monitoring and debugging.

**Location**: `/middleware/request_logger/`

#### Features
- Request and response metadata logging
- Optional request body logging
- Optional response body logging (planned)
- JSON and text body parsing
- Automatic body truncation
- Status-based log levels

#### Configuration

```yaml
middleware:
  - name: "request_logger"
    enabled: true
    config:
      include_request_body: false   # Log request body
      include_response_body: false  # Log response body (TODO)
```

#### Log Fields

**Request Log:**
- `method`: HTTP method
- `path`: Request path
- `query`: Query string
- `remote_ip`: Client IP address
- `user_agent`: User-Agent header
- `request_body`: Request body (if enabled)

**Response Log:**
- `duration`: Request duration (human readable)
- `duration_ms`: Duration in milliseconds
- `status`: HTTP status code
- `response_body`: Response body (if enabled)

#### Usage Examples

**Basic Logging:**
```yaml
middleware:
  - name: "request_logger"
    enabled: true
```

**With Request Body Logging:**
```yaml
middleware:
  - name: "request_logger"
    config:
      include_request_body: true
```

**Group-level Logging:**
```yaml
groups:
  - prefix: "/api/v1/auth"
    middleware:
      - name: "request_logger"
        config:
          include_request_body: true  # Log auth requests
  - prefix: "/api/v1/users"
    middleware:
      - name: "request_logger"
        config:
          include_request_body: false # Skip body for performance
```

#### Log Level Behavior

- **Info**: Status 200-399 (success responses)
- **Warn**: Status 400-499 (client errors)
- **Error**: Status 500+ (server errors)

### 6. Slow Request Logger Middleware (`slow_request_logger`)

Specialized middleware for detecting and logging slow requests that exceed performance thresholds.

**Location**: `/middleware/slow_request_logger/`

#### Features
- Configurable response time threshold
- Automatic slow request detection
- Performance monitoring
- Duration-based filtering

#### Configuration

```yaml
middleware:
  - name: "slow_request_logger"
    enabled: true
    config:
      threshold: "5s"            # Log requests slower than 5 seconds
```

#### Usage Examples

**Performance Monitoring:**
```yaml
middleware:
  - name: "slow_request_logger"
    config:
      threshold: "2s"            # Monitor requests > 2 seconds
```

**API-specific Monitoring:**
```yaml
groups:
  - prefix: "/api/v1"
    middleware:
      - name: "slow_request_logger"
        config:
          threshold: "1s"        # Stricter threshold for API
```

## Middleware Registration

All built-in middleware are automatically registered when using the defaults package:

```go
import "github.com/primadi/lokstra/defaults"

func main() {
    regCtx := lokstra.NewGlobalRegistrationContext()
    
    // Register all built-in middleware
    defaults.RegisterAllMiddleware(regCtx)
    
    // Your app configuration...
}
```

## Best Practices

### 1. Middleware Order

Order middleware carefully for optimal performance and functionality:

```yaml
middleware:
  - name: "recovery"              # First: Catch panics
  - name: "request_logger"        # Early: Log all requests
  - name: "cors"                  # Before auth: Handle preflight
  - name: "body_limit"           # Before parsing: Limit size
  - name: "gzipcompression"      # Last: Compress responses
```

### 2. Environment-specific Configuration

Use different configurations for different environments:

```yaml
# Development
middleware:
  - name: "request_logger"
    config:
      include_request_body: true
  - name: "recovery"
    config:
      enable_stack_trace: true

# Production
middleware:
  - name: "request_logger"
    config:
      include_request_body: false
  - name: "recovery"
    config:
      enable_stack_trace: false
```

### 3. Selective Application

Apply middleware selectively based on routes:

```yaml
# Global middleware
middleware:
  - name: "recovery"
  - name: "cors"

groups:
  - prefix: "/api"
    middleware:
      - name: "request_logger"    # Only log API requests
      - name: "body_limit"
        config:
          max_size: 1048576       # 1MB for API

  - prefix: "/uploads"
    middleware:
      - name: "body_limit"
        config:
          max_size: 104857600     # 100MB for uploads
```

### 4. Performance Considerations

- **Request Body Logging**: Only enable for debugging/development
- **Compression**: Configure appropriate levels and thresholds
- **Recovery Stack Traces**: Disable in production for performance
- **Body Limits**: Set appropriate limits for different endpoints

### 5. Security

- **CORS**: Configure specific origins in production
- **Body Limits**: Always set limits to prevent DoS attacks
- **Recovery**: Don't expose stack traces in production
- **Logging**: Be careful with sensitive data in request bodies

## Testing

Each middleware includes comprehensive test suites:

```bash
# Test specific middleware
go test ./middleware/request_logger -v
go test ./middleware/body_limit -v
go test ./middleware/recovery -v

# Test all middleware
go test ./middleware/... -v
```

## Error Handling

All middleware components include robust error handling:

- **Graceful degradation** when services are unavailable
- **Fallback behavior** for configuration errors
- **Proper error responses** for client-facing errors
- **Comprehensive logging** for debugging

## Custom Middleware

To create custom middleware following the same patterns:

```go
package my_middleware

import (
    "github.com/primadi/lokstra/core/midware"
    "github.com/primadi/lokstra/core/request"
)

func factory(config any) midware.Func {
    // Parse configuration
    return func(next request.HandlerFunc) request.HandlerFunc {
        return func(ctx *request.Context) error {
            // Middleware logic here
            return next(ctx)
        }
    }
}

func GetModule() registration.Module {
    return midware.NewModule("my_middleware", factory)
}
```

## Next Steps

- [Services](./built-in-services.md) - Learn about built-in services
- [Configuration](./configuration.md) - Advanced configuration patterns
- [Advanced Features](./advanced-features.md) - Production optimization
- [Schema Reference](./schema.md) - YAML schema documentation

---

*Built-in middleware in Lokstra provides production-ready functionality with comprehensive configuration options and robust error handling for building scalable web applications.*