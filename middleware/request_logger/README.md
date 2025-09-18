# Request Logger Middleware

Request Logger middleware for logging incoming HTTP requests and their metadata. This middleware can optionally log request and response bodies for debugging and monitoring purposes.

## Features

- Logs incoming requests with method, path, query, remote IP, and user agent
- Logs request completion with duration and status code
- Configurable request and response body logging
- Different log levels based on status codes (info/warn/error)
- Automatic body truncation for large payloads
- JSON detection and formatting
- Typed middleware and config recommended

## Configuration

```go
type Config struct {
    IncludeRequestBody  bool `json:"include_request_body" yaml:"include_request_body"`
    IncludeResponseBody bool `json:"include_response_body" yaml:"include_response_body"`
}
```

### Configuration Parameters

- `include_request_body` (bool): Include request body in log (default: `false`)
- `include_response_body` (bool): Include response body in log (default: `false`)

## Usage

### 1. Typed Middleware (Recommended)

It is recommended to use typed middleware with typed config:

```go
import "github.com/primadi/lokstra/middleware/request_logger"

config := &request_logger.Config{
    IncludeRequestBody:  true,
    IncludeResponseBody: false,
}
router.Use(request_logger.GetMidware(config))
```

### 2. Basic Usage (default config)

```go
// Use default request logger config
router.Use("request_logger")
```

### 3. With Map Configuration

```go
config := map[string]any{
    "include_request_body": true,
    "include_response_body": true,
}
router.Use("request_logger", config)
```

### 4. With Struct Config

```go
config := &request_logger.Config{
    IncludeRequestBody:  false,
    IncludeResponseBody: true,
}
router.Use("request_logger", config)
```

## Example Log Output

### Incoming Request Log
```json
{
  "level": "info",
  "method": "POST",
  "path": "/api/users",
  "query": "include=profile",
  "remote_ip": "192.168.1.100",
  "user_agent": "curl/7.68.0",
  "request_body": {
    "name": "John Doe",
    "email": "john@example.com"
  },
  "msg": "Incoming request"
}
```

### Request Completion Logs
```json
{
  "level": "info",
  "duration": "150ms",
  "duration_ms": 150,
  "status": 200,
  "msg": "Request completed successfully"
}

{
  "level": "warn",
  "duration": "80ms", 
  "status": 404,
  "msg": "Request completed with client error"
}

{
  "level": "error",
  "duration": "300ms",
  "status": 500,
  "msg": "Request completed with server error"
}
```

## Implementation Example

```go
package main

import (
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/middleware/request_logger"
)

func main() {
    regCtx := lokstra.NewGlobalRegistrationContext()
    
    // Create server
    server := lokstra.NewServer(regCtx, "my-app")
    
    // For development with request body logging
    config := &request_logger.Config{
        IncludeRequestBody:  true,
        IncludeResponseBody: false,
    }
    server.Use(request_logger.GetMidware(config))

    // Handler
    server.POST("/api/users", func(ctx *lokstra.Context) error {
        // Handle user creation
        return ctx.JSON(200, map[string]string{"status": "created"})
    })
    
    server.Listen(":8080")
}
```

## Best Practices

### 1. Environment-based Configuration

```go
config := &request_logger.Config{
    IncludeRequestBody:  os.Getenv("ENV") != "prod",
    IncludeResponseBody: false,
}
```

### 2. Selective Body Logging

For production, consider disabling body logging for sensitive endpoints:

```go
// Production - basic logging only
config := &request_logger.Config{
    IncludeRequestBody:  false,
    IncludeResponseBody: false,
}

// Development - full logging
config := &request_logger.Config{
    IncludeRequestBody:  true,
    IncludeResponseBody: true,
}
```

### 3. Security Considerations

Avoid logging request bodies for endpoints containing sensitive data such as:
- Authentication credentials
- Personal information
- Payment data
- API keys

## Testing

Request logger middleware comes with a comprehensive test suite:

```bash
go test ./middleware/request_logger -v
```

Tests cover:
- Configuration parsing in various formats
- Request and response logging
- Body inclusion/exclusion
- Different status code handling
- Error scenarios

(See request_logger_test.go for details.)
