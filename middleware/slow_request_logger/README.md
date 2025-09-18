# Slow Request Logger Middleware

Slow Request Logger middleware for logging requests that exceed a specified duration threshold. This helps identify performance bottlenecks and monitor slow endpoints in your application.

## Features

- Logs requests that take longer than the configured threshold
- Configurable log level and threshold duration
- Typed middleware and config recommended
- Easy integration with router

## Configuration

```go
type Config struct {
    IncludeRequestBody  bool          `json:"include_request_body" yaml:"include_request_body"`
    IncludeResponseBody bool          `json:"include_response_body" yaml:"include_response_body"`
    Threshold           time.Duration `json:"threshold" yaml:"threshold"`
}
```

### Configuration Parameters

- `threshold` (time.Duration): Minimum duration to consider a request as slow (e.g., "500ms", "2s"; default: `500ms`)
- `include_request_body` (bool): Include request body in log (default: `false`)
- `include_response_body` (bool): Include response body in log (default: `false`)

## Usage

### 1. Typed Middleware (Recommended)

It is recommended to use typed middleware with typed config:

```go
import "github.com/primadi/lokstra/middleware/slow_request_logger"

config := &slow_request_logger.Config{
    Threshold:           1200 * time.Millisecond, // Log requests slower than 1.2 seconds
    IncludeRequestBody:  true,
    IncludeResponseBody: false,
}
router.Use(slow_request_logger.GetMidware(config))
```

### 2. Basic Usage (default config)

```go
// Use default slow request logger config
router.Use("slow_request_logger")
```

### 3. With Map Configuration

```go
config := map[string]any{
    "threshold": "2s",
    "include_request_body": true,
    "include_response_body": true,
}
router.Use("slow_request_logger", config)
```

### 4. With Struct Config

```go
config := &slow_request_logger.Config{
    Threshold:           750 * time.Millisecond,
    IncludeRequestBody:  false,
    IncludeResponseBody: true,
}
router.Use("slow_request_logger", config)
```

## Example Log Output

```
INFO Slow request detected {"method":"GET","path":"/api/data","duration":"1200ms"}
```

## Best Practices

- Use typed middleware and config for type safety and clarity
- Adjust `threshold` based on your application's performance requirements
- Monitor slow request logs regularly to identify bottlenecks

## Testing

Run tests for the slow request logger middleware:

```bash
go test ./middleware/slow_request_logger -v
```

Tests cover:
- Configuration parsing in various formats
- Logging of slow requests
- Threshold and log level effects
- Edge cases and error handling

(See slow_request_logger_test.go for details.)
