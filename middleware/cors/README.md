# CORS Middleware

CORS (Cross-Origin Resource Sharing) middleware for handling cross-origin HTTP requests. This middleware sets appropriate CORS headers to allow or restrict access from different origins.

## Features

- Configurable allowed origins, methods, and headers
- Supports credentials and exposed headers
- Handles preflight (OPTIONS) requests
- Customizable max age for preflight cache
- Flexible configuration via map, struct, or direct param

## Configuration

```go
type Config struct {
    AllowOrigins     []string `json:"allow_origins" yaml:"allow_origins"`
    AllowMethods     []string `json:"allow_methods" yaml:"allow_methods"`
    AllowHeaders     []string `json:"allow_headers" yaml:"allow_headers"`
    ExposeHeaders    []string `json:"expose_headers" yaml:"expose_headers"`
    AllowCredentials bool     `json:"allow_credentials" yaml:"allow_credentials"`
    MaxAge           int      `json:"max_age" yaml:"max_age"`
}
```

### Configuration Parameters

- `allow_origins` ([]string): List of allowed origins (default: `["*"]`)
- `allow_methods` ([]string): List of allowed HTTP methods (default: `["GET", "POST", "PUT", "DELETE", "OPTIONS"]`)
- `allow_headers` ([]string): List of allowed headers (default: `["*"]`)
- `expose_headers` ([]string): List of headers exposed to the browser (default: `[]`)
- `allow_credentials` (bool): Allow credentials (default: `false`)
- `max_age` (int): Max age for preflight cache in seconds (default: `86400`)

## Usage



### 1. Typed Middleware (Recommended)

It is recommended to use typed middleware with typed config:

```go
import "github.com/primadi/lokstra/middleware/cors"

config := &cors.Config{
    AllowOrigins:     []string{"https://mydomain.com"},
    AllowMethods:     []string{"GET", "POST"},
    AllowHeaders:     []string{"Authorization", "Content-Type"},
    ExposeHeaders:    []string{"X-Custom-Header"},
    AllowCredentials: true,
    MaxAge:           600,
}
router.Use(cors.GetMidware(config))
```

### 2. Basic Usage (default config)

```go
// Use default CORS config
router.Use("cors")
```

### 3. With Map Configuration

```go
config := map[string]any{
    "allow_origins": []string{"https://example.com", "https://another.com"},
    "allow_methods": []string{"GET", "POST"},
    "allow_headers": []string{"Authorization", "Content-Type"},
    "allow_credentials": true,
    "max_age": 3600,
}
router.Use("cors", config)
```

### 4. With Struct Config

```go
config := &cors.Config{
    AllowOrigins:     []string{"https://mydomain.com"},
    AllowMethods:     []string{"GET", "POST"},
    AllowHeaders:     []string{"Authorization", "Content-Type"},
    ExposeHeaders:    []string{"X-Custom-Header"},
    AllowCredentials: true,
    MaxAge:           600,
}
router.Use("cors", config)
```

## Example Response Headers

When a request matches the allowed origin, the following headers are set:

```
Access-Control-Allow-Origin: https://example.com
Access-Control-Allow-Methods: GET, POST
Access-Control-Allow-Headers: Authorization, Content-Type
Access-Control-Allow-Credentials: true
Access-Control-Max-Age: 3600
```

## Preflight (OPTIONS) Request Handling

For OPTIONS requests, the middleware responds with status 204 (No Content) and sets the appropriate CORS headers.

## Best Practices

- Use specific origins in production for better security
- Set `allow_credentials` to `true` only if needed
- Limit allowed methods and headers to only those required
- Use environment-based configuration for flexibility

## Testing

Run tests for the CORS middleware:

```bash
go test ./middleware/cors -v
```

Tests cover:
- Configuration parsing in various formats
- Handling of allowed origins, methods, and headers
- Preflight request handling
- Credentials and exposed headers
- Edge cases and error handling

(See cors_test.go for details.)
