# Body Limit Middleware

Middleware to limit request body size in Lokstra framework.

## Features

### 1. **Size Limiting**
- Limits body size based on `MaxSize` (default: 10MB)
- Two-layer checking:
  - `ContentLength` header check (early validation)
  - Actual reading with `limitedReadCloser` (runtime enforcement)

### 2. **Skip Large Payloads**
- `SkipLargePayloads: true` → read body up to limit, then EOF
- `SkipLargePayloads: false` → return error if exceeds limit

### 3. **Path Skipping**
- Skip body limit check for specific paths
- Support wildcard patterns:
  - `*` → single segment (e.g., `/api/*`)
  - `**` → multi segments (e.g., `/public/**`)

### 4. **Custom Error Response**
- Configurable status code (default: 413)
- Configurable error message

## Usage

### Basic Usage

```go
import "github.com/primadi/lokstra/middleware"

// Default config (10MB)
r.Use(middleware.BodyLimit(&middleware.Config{}))
```

### Custom Configuration

```go
r.Use(middleware.BodyLimit(&middleware.Config{
    MaxSize:    5 * 1024 * 1024, // 5MB
    Message:    "File too large",
    StatusCode: http.StatusBadRequest,
}))
```

### Skip Large Payloads

```go
// Read body sampai limit, lalu stop (tidak error)
r.Use(middleware.BodyLimit(&middleware.Config{
    MaxSize:           1024 * 1024, // 1MB
    SkipLargePayloads: true,
}))
```

### Skip on Specific Paths

```go
r.Use(middleware.BodyLimit(&middleware.Config{
    MaxSize: 1024 * 1024, // 1MB
    SkipOnPath: []string{
        "/upload/*",      // Skip /upload/file, /upload/image, etc
        "/public/**",     // Skip all paths under /public
        "/api/*/large",   // Skip /api/v1/large, /api/v2/large, etc
    },
}))
```

## Config Options

```go
type Config struct {
    // Maximum allowed body size in bytes
    // Default: 10MB (10 * 1024 * 1024)
    MaxSize int64

    // If true, skip reading body when exceeds limit (return EOF)
    // If false, return error immediately
    // Default: false
    SkipLargePayloads bool

    // Custom error message for oversized bodies
    // Default: "Request body too large"
    Message string

    // HTTP status code to return for oversized bodies
    // Default: 413 (StatusRequestEntityTooLarge)
    StatusCode int

    // List of path patterns to skip body limit check
    // Supports wildcards: * (single segment), ** (multi segments)
    // Default: []
    SkipOnPath []string
}
```

## How It Works

### 1. Optional Early Validation (ContentLength Header)

**Before handler execution:**
- HTTP `ContentLength` header is checked if available (value > 0)
- If exceeds limit:
  - `SkipLargePayloads: false` → reject immediately (413 error)
  - `SkipLargePayloads: true` → allow, but body will be limited by reader

**Important Limitations:**
- ⚠️ **Only works if client sets ContentLength header**
- ⚠️ **If ContentLength = -1 (unknown/chunked), this check is skipped**
- ⚠️ **Cannot detect if handler/middleware manually sets ContentLength**
- ⚠️ **This is an optimization, NOT reliable primary protection**

**When it works:**
- Standard POST/PUT requests with `Content-Length` header
- Client properly declares body size

**When it doesn't work:**
- Chunked transfer encoding (ContentLength = -1)
- Streaming requests
- Clients that don't set ContentLength

### 2. Primary Enforcement (limitedReadCloser)

**This is the REAL protection:**
- Body reader is wrapped with `limitedReadCloser` **before** `c.Next()`
- When handler calls `io.Read()` on body:
  - Tracks bytes read via `remaining` counter
  - Stops at `MaxSize` bytes
  - Returns error or EOF based on `SkipLargePayloads`

**Why this is reliable:**
- ✅ Works regardless of ContentLength header
- ✅ Works at any middleware position
- ✅ Enforces limit during actual read operations
- ✅ Tracks bytes read in real-time
- ✅ Cannot be bypassed by incorrect headers

### 3. Protection Summary

| Scenario | ContentLength Check | limitedReadCloser | Result |
|----------|-------------------|-------------------|---------|
| Client sends Content-Length: 100MB (limit 10MB) | ✅ Rejects early | N/A (already rejected) | Protected |
| Client sends Content-Length: 5MB, actual 50MB | ⚠️ Passes (wrong header) | ✅ Stops at 10MB | Protected |
| Chunked encoding (no Content-Length) | ❌ Skipped (value = -1) | ✅ Stops at 10MB | Protected |
| Streaming (ContentLength = -1) | ❌ Skipped | ✅ Stops at 10MB | Protected |

**Conclusion:** 
- ContentLength check = **optimization** for early rejection when header is available
- limitedReadCloser = **mandatory protection** that always enforces the limit

### 4. Path Matching

Pattern matching for skip paths:
- `/api/*` → match `/api/test` but not `/api/test/sub`
- `/public/**` → match all paths under `/public` including nested
- Exact match: `/upload/file`

## Why Not Check After c.Next()?

**Checking after c.Next() is ineffective because:**
1. Handler has already executed
2. Body may have been fully read
3. Response may have been sent
4. Error would occur too late to prevent damage

**Correct approach (current implementation):**
- Check `ContentLength` **before** c.Next() → early rejection (if header available)
- Wrap body **before** c.Next() → runtime protection (always works)
- Handler reads from wrapped reader → automatic enforcement

## Middleware Ordering Considerations

### ❌ Common Mistake: Body Limit as Last Middleware

```go
// BAD: Body limit runs last
router.Use(middleware.Logger())
router.Use(middleware.CORS())
router.Use(middleware.BodyLimit(&middleware.Config{MaxSize: 1024 * 1024}))
```

**Problem:** If previous middleware reads the body, it's already consumed.

### ✅ Recommended: Body Limit as First or Early Middleware

```go
// GOOD: Body limit runs first
router.Use(middleware.BodyLimit(&middleware.Config{MaxSize: 1024 * 1024}))
router.Use(middleware.Logger())
router.Use(middleware.CORS())
```

**Benefits:**
- Body is wrapped before any middleware can read it
- All subsequent middleware and handlers get the limited reader
- Protection is guaranteed regardless of what other middleware do

### Important Notes:

1. **limitedReadCloser works at any position** - it will limit reads whenever they happen
2. **ContentLength check only works early** - if body is already read, ContentLength may be stale
3. **Best practice: Place BodyLimit early** - ensures consistent protection across all handlers
4. **Handler cannot bypass** - once body is wrapped, all reads go through the limiter

### Edge Case: Handler Sets ContentLength Manually

```go
// This does NOT bypass the limit
func handler(c *request.Context) error {
    c.R.ContentLength = 999999999999  // Attempt to bypass
    
    // Body reading still limited by limitedReadCloser
    body, err := io.ReadAll(c.R.Body)  // Will stop at MaxSize
    return err
}
```

**Why it doesn't work:**
- ContentLength check happens before handler runs
- limitedReadCloser wraps the body reader before handler
- Changing ContentLength field doesn't change the wrapped reader
- Actual read limit is enforced by limitedReadCloser, not ContentLength header

## Examples

### Example 1: API with Different Limits

```go
router := router.New("api")

// Default limit for all routes
router.Use(middleware.BodyLimit(&middleware.Config{
    MaxSize: 1024 * 1024, // 1MB
}))

// Routes
router.POST("/api/data", dataHandler)
router.POST("/api/file", fileHandler)

// Upload endpoint with larger limit (skip default limit)
uploadRouter := router.Group("/upload")
uploadRouter.Use(middleware.BodyLimit(&middleware.Config{
    MaxSize: 50 * 1024 * 1024, // 50MB for uploads
}))
uploadRouter.POST("/file", uploadHandler)
```

### Example 2: Public Assets (No Limit)

```go
router.Use(middleware.BodyLimit(&middleware.Config{
    MaxSize: 1024 * 1024, // 1MB default
    SkipOnPath: []string{
        "/public/**",  // No limit for public assets
        "/webhook/*",  // No limit for webhooks
    },
}))
```

### Example 3: Streaming with SkipLargePayloads

```go
// For streaming: read up to limit then stop
router.Use(middleware.BodyLimit(&middleware.Config{
    MaxSize:           10 * 1024 * 1024, // 10MB
    SkipLargePayloads: true, // Don't error, just stop reading
}))

router.POST("/stream", func(c *request.Context) error {
    // Will read maximum 10MB, then EOF
    data, _ := io.ReadAll(c.R.Body)
    // Process data...
})
```

## Testing

Run tests:
```bash
go test -v ./middleware -run "TestBodyLimit|TestLimited"
```

Test coverage:
- ✅ Body within limit
- ✅ Body exceeds limit (ContentLength)
- ✅ Body exceeds limit (actual reading)
- ✅ SkipLargePayloads behavior
- ✅ Path skipping with wildcards
- ✅ Custom status code and message
- ✅ Default config values
- ✅ limitedReadCloser unit tests

## Best Practices

1. **Set appropriate limits**: Adjust based on endpoint requirements
2. **Use path skipping**: For uploads or webhooks that need different limits
3. **SkipLargePayloads**: Use for streaming or partial reading scenarios
4. **Layer limits**: Use multiple middleware with different limits per route group
5. **Monitor**: Log or track requests exceeding limits

## Implementation Details

### limitedReadCloser
- Wraps `io.ReadCloser` to enforce limit
- Tracks `remaining` bytes during reading
- Returns error or EOF based on config
- Transparent to handler (no changes needed)

### Path Matching
- Path normalization with `path.Clean()`
- Supports exact match and wildcard patterns
- Pattern processing order: exact → ** → * → fallback

## Performance

- ✅ **Early check**: ContentLength validated before reading
- ✅ **Zero-copy**: limitedReadCloser only wraps, doesn't copy data
- ✅ **Lazy evaluation**: Path matching only if SkipOnPath configured
- ✅ **No reflection**: Pure interface implementation

## Production Ready

- ✅ Comprehensive tests (15+ test cases)
- ✅ Edge cases handled (unknown ContentLength, EOF, etc)
- ✅ Configurable defaults
- ✅ Backward compatible
- ✅ Well documented
