# Body Limit Middleware

> Request body size protection

## Overview

Body Limit middleware enforces maximum request body size to prevent memory exhaustion attacks. It provides two-layer protection: ContentLength header check and runtime enforcement during reading.

## Import Path

```go
import "github.com/primadi/lokstra/middleware/body_limit"
```

---

## Configuration

### Config Type

```go
type Config struct {
    MaxSize           int64    // Maximum body size in bytes
    SkipLargePayloads bool     // Skip reading if exceeds limit (vs error)
    Message           string   // Custom error message
    StatusCode        int      // HTTP status code for error
    SkipOnPath        []string // Path patterns to skip limit check
}
```

**Fields:**
- `MaxSize` - Maximum allowed body size in bytes (default: 10MB)
- `SkipLargePayloads` - If `true`, skip reading oversized bodies; if `false`, return error
- `Message` - Custom error message (default: "Request body too large")
- `StatusCode` - HTTP status code (default: 413 Request Entity Too Large)
- `SkipOnPath` - Path patterns to skip (supports `*` and `**` wildcards)

---

### DefaultConfig

```go
func DefaultConfig() *Config
```

**Returns:**
```go
&Config{
    MaxSize:           10 * 1024 * 1024, // 10MB
    SkipLargePayloads: false,
    Message:           "Request body too large",
    StatusCode:        http.StatusRequestEntityTooLarge, // 413
    SkipOnPath:        []string{},
}
```

---

## Usage

### Basic Usage

```go
// 10MB limit (default)
router.Use(body_limit.Middleware(&body_limit.Config{
    MaxSize: 10 * 1024 * 1024,
}))
```

---

### Custom Limit

```go
// 5MB limit
router.Use(body_limit.Middleware(&body_limit.Config{
    MaxSize: 5 * 1024 * 1024,
}))

// 100KB limit for API
router.Use(body_limit.Middleware(&body_limit.Config{
    MaxSize: 100 * 1024,
}))
```

---

### Skip Specific Paths

```go
router.Use(body_limit.Middleware(&body_limit.Config{
    MaxSize: 1 * 1024 * 1024, // 1MB default
    SkipOnPath: []string{
        "/upload/**",   // Skip all upload paths
        "/import/**",   // Skip import endpoints
        "/webhook/*",   // Skip webhooks (single level)
    },
}))
```

---

### Custom Error Message

```go
router.Use(body_limit.Middleware(&body_limit.Config{
    MaxSize:    5 * 1024 * 1024,
    Message:    "File too large. Maximum size is 5MB",
    StatusCode: http.StatusRequestEntityTooLarge,
}))
```

---

## YAML Configuration

```yaml
middlewares:
  - type: body_limit
    params:
      max_size: 10485760  # 10MB in bytes
      message: "Request body too large"
      status_code: 413
      skip_on_path:
        - "/upload/**"
        - "/import/**"
```

---

## Path Patterns

### Exact Match

```go
SkipOnPath: []string{
    "/upload",        // Only matches /upload
    "/api/import",    // Only matches /api/import
}
```

---

### Single Wildcard (`*`)

Matches any characters **within a single path segment**:

```go
SkipOnPath: []string{
    "/api/*",         // Matches /api/upload, /api/import
                      // Does NOT match /api/v1/upload
    
    "/files/*.json",  // Matches /files/data.json
                      // Does NOT match /files/2024/data.json
}
```

---

### Double Wildcard (`**`)

Matches **any number of path segments**:

```go
SkipOnPath: []string{
    "/upload/**",     // Matches /upload/file
                      // Matches /upload/images/photo.jpg
                      // Matches /upload/a/b/c/file.pdf
    
    "/api/**/files",  // Matches /api/v1/files
                      // Matches /api/v2/admin/files
}
```

---

## Examples

### Different Limits for Different Endpoints

```go
// Small limit for general API
apiRouter := router.Group("/api")
apiRouter.Use(body_limit.Middleware(&body_limit.Config{
    MaxSize: 1 * 1024 * 1024, // 1MB
}))

// Large limit for uploads
uploadRouter := router.Group("/upload")
uploadRouter.Use(body_limit.Middleware(&body_limit.Config{
    MaxSize: 100 * 1024 * 1024, // 100MB
}))
```

---

### Skip File Uploads

```go
router.Use(body_limit.Middleware(&body_limit.Config{
    MaxSize: 1 * 1024 * 1024, // 1MB default
    SkipOnPath: []string{
        "/upload/image/**",
        "/upload/video/**",
        "/upload/document/**",
    },
}))

// Separate middleware for uploads with higher limit
uploadRouter := router.Group("/upload")
uploadRouter.Use(body_limit.Middleware(&body_limit.Config{
    MaxSize: 50 * 1024 * 1024, // 50MB for uploads
}))
```

---

### Webhook Endpoints

```go
router.Use(body_limit.Middleware(&body_limit.Config{
    MaxSize: 1 * 1024 * 1024,
    SkipOnPath: []string{
        "/webhook/**", // Webhooks may send large payloads
    },
}))
```

---

### Environment-Based Limits

```go
var maxSize int64 = 10 * 1024 * 1024 // 10MB default

if limit := os.Getenv("MAX_BODY_SIZE"); limit != "" {
    if size, err := strconv.ParseInt(limit, 10, 64); err == nil {
        maxSize = size
    }
}

router.Use(body_limit.Middleware(&body_limit.Config{
    MaxSize: maxSize,
}))
```

---

### Production vs Development

```go
var cfg *body_limit.Config

if os.Getenv("ENV") == "production" {
    cfg = &body_limit.Config{
        MaxSize:    5 * 1024 * 1024, // Stricter in production
        Message:    "Request too large",
        StatusCode: 413,
    }
} else {
    cfg = &body_limit.Config{
        MaxSize: 50 * 1024 * 1024, // Larger for testing
    }
}

router.Use(body_limit.Middleware(cfg))
```

---

## Protection Layers

### Layer 1: ContentLength Check

If `Content-Length` header is present and exceeds limit:
- **Immediate rejection** (no body reading)
- **Fast response** (minimal processing)

```http
POST /api/users HTTP/1.1
Content-Length: 11000000

// Rejected immediately if limit is 10MB
```

---

### Layer 2: Runtime Enforcement

Body reader is wrapped with `limitedReadCloser`:
- **Enforced during actual reading**
- **Works regardless of ContentLength header**
- **Protects against chunked encoding**

```go
// Even if ContentLength is missing or incorrect,
// reading will stop at MaxSize bytes
```

---

## Behavior

### Request Rejected

When body exceeds limit:

**Response:**
```json
{
  "error": "Request body too large",
  "code": "BODY_TOO_LARGE"
}
```

**Status Code:** 413 Request Entity Too Large

---

### SkipLargePayloads

When `SkipLargePayloads: true`:
- Body reading stops at limit
- No error returned
- Handler receives partial body (up to limit)

```go
router.Use(body_limit.Middleware(&body_limit.Config{
    MaxSize:           10 * 1024 * 1024,
    SkipLargePayloads: true, // Don't error, just truncate
}))
```

**⚠️ Warning:** Use with caution - handler may process incomplete data.

---

## Best Practices

### 1. Set Appropriate Limits

```go
// ✅ Good - different limits for different needs
router.Use(body_limit.Middleware(&body_limit.Config{
    MaxSize: 1 * 1024 * 1024, // 1MB for API
    SkipOnPath: []string{"/upload/**"},
}))

uploadRouter.Use(body_limit.Middleware(&body_limit.Config{
    MaxSize: 50 * 1024 * 1024, // 50MB for uploads
}))

// 🚫 Bad - one size for all
router.Use(body_limit.Middleware(&body_limit.Config{
    MaxSize: 100 * 1024 * 1024, // Too large for API
}))
```

---

### 2. Place Before Body Parsing

```go
// ✅ Good - check size before parsing
router.Use(
    recovery.Middleware(&recovery.Config{}),
    body_limit.Middleware(&body_limit.Config{
        MaxSize: 10 * 1024 * 1024,
    }), // Before handlers read body
)

// 🚫 Bad - after handlers already read body
router.Use(
    recovery.Middleware(&recovery.Config{}),
    // handlers here...
    body_limit.Middleware(&body_limit.Config{}), // Too late
)
```

---

### 3. Use Wildcards Appropriately

```go
// ✅ Good - clear intent
SkipOnPath: []string{
    "/upload/**",    // All upload paths
    "/webhook/*",    // Top-level webhooks only
}

// 🚫 Bad - overly broad
SkipOnPath: []string{
    "/**", // Disables protection everywhere!
}
```

---

### 4. Avoid SkipLargePayloads

```go
// ✅ Good - fail fast on oversized requests
body_limit.Middleware(&body_limit.Config{
    MaxSize:           10 * 1024 * 1024,
    SkipLargePayloads: false, // Error on oversized
})

// 🚫 Bad - may cause unexpected behavior
body_limit.Middleware(&body_limit.Config{
    MaxSize:           10 * 1024 * 1024,
    SkipLargePayloads: true, // Truncates silently
})
```

---

### 5. Provide Clear Error Messages

```go
// ✅ Good - helpful error message
body_limit.Middleware(&body_limit.Config{
    MaxSize: 5 * 1024 * 1024,
    Message: "File too large. Maximum size is 5MB",
})

// 🚫 Bad - generic message
body_limit.Middleware(&body_limit.Config{
    MaxSize: 5 * 1024 * 1024,
    Message: "Error", // Not helpful
})
```

---

## Common Sizes

```go
const (
    KB = 1024
    MB = 1024 * KB
    GB = 1024 * MB
)

// API requests
MaxSize: 100 * KB      // 100KB - strict API

// JSON payloads
MaxSize: 1 * MB        // 1MB - typical JSON

// Form data
MaxSize: 5 * MB        // 5MB - forms with files

// Image uploads
MaxSize: 10 * MB       // 10MB - images

// Document uploads
MaxSize: 50 * MB       // 50MB - PDFs, documents

// Video uploads
MaxSize: 500 * MB      // 500MB - videos
```

---

## Performance

**Overhead:** ~100ns per request (wrapper allocation only)

**Impact:** Minimal - protection is only enforced when body is actually read

---

## Testing

```go
func TestBodyLimit(t *testing.T) {
    router := lokstra.NewRouter()
    router.Use(body_limit.Middleware(&body_limit.Config{
        MaxSize: 1024, // 1KB limit
    }))
    
    router.POST("/test", func(c *request.Context) error {
        return c.Api.Ok("success")
    })
    
    // Test oversized body
    largeBody := make([]byte, 2048) // 2KB
    req := httptest.NewRequest("POST", "/test", bytes.NewReader(largeBody))
    rec := httptest.NewRecorder()
    
    router.ServeHTTP(rec, req)
    
    assert.Equal(t, 413, rec.Code)
}

func TestBodyLimitSkipPath(t *testing.T) {
    router := lokstra.NewRouter()
    router.Use(body_limit.Middleware(&body_limit.Config{
        MaxSize:    1024,
        SkipOnPath: []string{"/upload/**"},
    }))
    
    // Test oversized body on skipped path
    largeBody := make([]byte, 2048)
    req := httptest.NewRequest("POST", "/upload/file", bytes.NewReader(largeBody))
    rec := httptest.NewRecorder()
    
    router.ServeHTTP(rec, req)
    
    assert.Equal(t, 200, rec.Code) // Not blocked
}
```

---

## See Also

- **[Recovery](./recovery)** - Panic recovery
- **[Request Logger](./request-logger)** - Request logging
- **[Gzip Compression](./gzip-compression)** - Response compression

---

## Related Guides

- **[Security Best Practices](../../04-guides/security/)** - Security patterns
- **[File Uploads](../../04-guides/file-uploads/)** - File upload handling
- **[Performance](../../04-guides/performance/)** - Optimization tips
