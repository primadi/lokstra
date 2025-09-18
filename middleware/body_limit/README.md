# Body Limit Middleware

The body limit middleware restricts the maximum size of incoming request bodies, protecting your server from memory exhaustion and large payload attacks.

## Features

- Enforces request body size limit via configuration
- Early detection using Content-Length header
- Customizable error message and status code
- Supports skip paths with pattern matching (`*`, `**`)
- Option to skip large payloads instead of returning error
- Convenience functions for common size limits

## Configuration

```go
type Config struct {
    MaxSize           int64    `json:"max_size" yaml:"max_size"`
    SkipLargePayloads bool     `json:"skip_large_payloads" yaml:"skip_large_payloads"`
    Message           string   `json:"message" yaml:"message"`
    StatusCode        int      `json:"status_code" yaml:"status_code"`
    SkipOnPath        []string `json:"skip_on_path" yaml:"skip_on_path"`
}
```

### Configuration Parameters

- `max_size` (int64): Maximum allowed request body size in bytes  
  Default: `10485760` (10MB)
- `skip_large_payloads` (bool): If true, skips reading body exceeding the limit  
  Default: `false`
- `message` (string): Custom error message for oversized payloads  
  Default: `"Request body too large"`
- `status_code` (int): HTTP status code to return  
  Default: `413` (Request Entity Too Large)
- `skip_on_path` ([]string): Array of path patterns to skip limit check  
  Default: `[]` (empty), supports wildcards: `*`, `**`

## Usage

### 1. Basic Usage

```go
// Use default 10MB limit
router.Use(body_limit.GetMidware(nil))

// Or use convenience function
router.Use(body_limit.BodyLimit10MB())
```

### 2. Configuration with Map

```yaml
# config.yaml
middleware:
  - name: "body_limit"
    config:
      max_size: 5242880      # 5MB
      status_code: 413
      message: "Payload too large"
      skip_on_path:
        - "/upload/*"        # Skip all upload paths
        - "/webhook"         # Skip webhook endpoint
        - "/api/files/**"    # Skip all file operations
```

```go
// In code
config := map[string]any{
    "max_size":     int64(5 * 1024 * 1024), // 5MB
    "status_code":  413,
    "message":      "Request body too large",
    "skip_on_path": []string{"/upload/*", "/webhook"},
}
router.Use("body_limit", config)
```

### 3. Configuration with Struct

```go
config := &body_limit.Config{
    MaxSize:    1024 * 1024, // 1MB
    StatusCode: 400,
    Message:    "File too large",
    SkipOnPath: []string{
        "/api/webhooks/*",     // Skip webhook endpoints
        "/upload/large/**",    // Skip large file uploads
    },
}

// Recommended for type safety
router.Use(body_limit.GetMidware(config))
```

## Pattern Matching for Skip Paths

The middleware supports several patterns for `skip_on_path`:

- Exact match: `/webhook`, `/api/status`
- Single wildcard (`*`): `/upload/*`, `/api/*/status`
- Double wildcard (`**`): `/static/**`, `/api/**/upload`

## Convenience Functions

```go
router.Use(body_limit.BodyLimit1MB())   // 1MB limit
router.Use(body_limit.BodyLimit5MB())   // 5MB limit
router.Use(body_limit.BodyLimit10MB())  // 10MB limit (default)
router.Use(body_limit.BodyLimit50MB())  // 50MB limit (for file uploads)
router.Use(body_limit.BodyLimit(2 * 1024 * 1024)) // 2MB
router.Use(body_limit.BodyLimitWithSkip(1024 * 1024)) // 1MB with skip
```

## Example Response

When limit exceeded:
```json
{
    "success": false,
    "message": "Request body too large",
    "data": {
        "maxSize": 1048576,
        "actual": 2097152
    }
}
```

Error log:
```
Request body too large (maxSize: 1048576, actual: 2097152)
```

## Testing

Body limit middleware includes a comprehensive test suite:

```bash
go test ./middleware/body_limit -v
```

Tests cover:
- Basic body limit enforcement
- Content-Length header checking
- Skip large payloads functionality
- Custom configuration
- Skip path pattern matching
- Factory function parsing
- Error scenarios
