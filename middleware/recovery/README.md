# Recovery Middleware

Recovery middleware for handling panics and preventing application crashes. This middleware catches panics that occur in handlers and returns appropriate error responses.

## Features

- Catches panics from handlers
- Logs detailed error information
- Returns HTTP 500 Internal Server Error
- Configurable stack trace enable/disable
- Maintains request ID for tracing

## Configuration

```go
type Config struct {
    EnableStackTrace bool `json:"enable_stack_trace" yaml:"enable_stack_trace"`
}
```

### Configuration Parameters

- `enable_stack_trace` (bool): Controls whether stack trace is included in error logs
  - Default: `false`
  - `true`: Stack trace will be logged for debugging
  - `false`: Only panic message is logged, without stack trace

## Usage

### 1. Basic Usage (with/without stack trace)

```go
// enable stack trace : true/ false

// 1. typed middleware:
router.Use(recovery.GetMidware(enableStackTrace))

// 2. use middleware by name, parameter by direct param:
router.Use("recovery", enableStackTrace)

// recommended to use the first form, because of typed middleware and typed parameter
```

### 2. With Map Configuration

```go
// Enable stack trace
config := map[string]any{
    "enable_stack_trace": true,
}
// middleware by name, config by map
router.Use("recovery", config)
```

### 3. With Struct Config

```go
config := &recovery.Config{
    EnableStackTrace: false, // for production environment
}
// middleware by name, config by struct
router.Use("recovery", config)
```

## Example Response

When a panic occurs, the middleware will return the following response:

```json
{
    "success": false,
    "message": "Internal Server Error",
    "code": "INTERNAL"
}
```

## Log Output

### With Stack Trace Active (Development)

```json
{
    "level": "error",
    "msg": "Recovered from panic in middleware",
    "error": "division by zero",
    "request_id": "req-12345",
    "url": "/api/users",
    "method": "GET",
    "stack": "goroutine 1 [running]:\nruntime/debug.Stack()..."
}
```

### Without Stack Trace (Production)

```json
{
    "level": "error", 
    "msg": "Recovered from panic in middleware",
    "error": "division by zero",
    "request_id": "req-12345", 
    "url": "/api/users",
    "method": "GET"
}
```

## Implementation Example

```go
package main

import (
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/middleware/recovery"
)

func main() {
    regCtx := lokstra.NewGlobalRegistrationContext()

    
    // Create server
    server := lokstra.NewServer(regCtx, "my-app")
    
    // For development with stack trace
    devConfig := os.Getenv("ENV") != "prod"
    server.Use("recovery", devConfig)

    // Handler that might panic
    server.GET("/test", func(ctx *lokstra.Context) error {
        panic("something went wrong!")
        return nil
    })
    
    server.Listen(":8080")
}
```

## Best Practices

### 1. Environment-based Configuration

```go
config := map[string]any{
    "enable_stack_trace": os.Getenv("ENV") != "prod",
}
```

### 2. Selective Stack Trace

For production, consider disabling stack trace to reduce log size and save performance:

```go
// Production
config := map[string]any{
    "enable_stack_trace": false,
}

// Development/Staging
config := map[string]any{
    "enable_stack_trace": true,
}
```

### 3. Monitoring Integration

Recovery middleware works well with monitoring systems. Error logs can be forwarded using a recovery hook:

### Integration Example: Using `recovery.SetRecoverHook`

You can forward errors to monitoring systems by setting a custom hook function:

```go
import (
    "github.com/primadi/lokstra/middleware/recovery"
    "github.com/getsentry/sentry-go"
    "fmt"
)

func main() {
    // Set a hook to forward errors to Sentry
    recovery.SetRecoverHook(func(ctx *request.Context, err any, stack string) {
        sentry.CaptureException(fmt.Errorf("%v\n%s", err, stack))
    })
    // ...existing code...
}
```

You can use similar hooks for Rollbar, New Relic, CloudWatch, or any other monitoring system.

## Testing

Recovery middleware comes with a comprehensive test suite:

```bash
go test ./middleware/recovery -v
```

Tests cover:
- Configuration parsing in various formats
- Recovery from panic with stack trace
- Recovery from panic without stack trace
- Normal execution without panic
- Configuration edge cases
- Custom recovery hook integration

(See recovery_test.go for details.)
