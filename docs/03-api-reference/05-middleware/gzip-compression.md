# Gzip Compression Middleware

> Response compression with gzip encoding

## Overview

Gzip Compression middleware compresses HTTP responses using gzip encoding to reduce bandwidth usage. It intelligently decides when to compress based on response size, content type, and client capabilities.

## Import Path

```go
import "github.com/primadi/lokstra/middleware/gzipcompression"
```

---

## Configuration

### Config Type

```go
type Config struct {
    MinSize              int      // Minimum response size to compress (bytes)
    CompressionLevel     int      // Gzip compression level (1-9, or -1 for default)
    ExcludedContentTypes []string // Content types that should not be compressed
}
```

**Fields:**
- `MinSize` - Minimum response size in bytes to compress (default: 1024 = 1KB)
- `CompressionLevel` - Compression level:
  - `-1`: Default compression (recommended)
  - `0`: No compression
  - `1`: Best speed
  - `9`: Best compression
- `ExcludedContentTypes` - List of content types to skip compression

---

### DefaultConfig

```go
func DefaultConfig() *Config
```

**Returns:**
```go
&Config{
    MinSize:          1024, // 1KB minimum
    CompressionLevel: gzip.DefaultCompression, // -1
    ExcludedContentTypes: []string{
        "image/jpeg",
        "image/png",
        "image/gif",
        "image/webp",
        "video/mp4",
        "video/webm",
        "application/zip",
        "application/gzip",
    },
}
```

---

## Usage

### Basic Usage

```go
router.Use(gzipcompression.Middleware(&gzipcompression.Config{
    MinSize:          1024, // 1KB
    CompressionLevel: gzip.DefaultCompression,
}))
```

---

### Fast Compression

```go
// Prioritize speed over compression ratio
router.Use(gzipcompression.Middleware(&gzipcompression.Config{
    MinSize:          1024,
    CompressionLevel: gzip.BestSpeed, // Level 1
}))
```

---

### Best Compression

```go
// Prioritize compression ratio over speed
router.Use(gzipcompression.Middleware(&gzipcompression.Config{
    MinSize:          1024,
    CompressionLevel: gzip.BestCompression, // Level 9
}))
```

---

### Custom Exclusions

```go
router.Use(gzipcompression.Middleware(&gzipcompression.Config{
    MinSize:          1024,
    CompressionLevel: gzip.DefaultCompression,
    ExcludedContentTypes: []string{
        // Images (already compressed)
        "image/jpeg",
        "image/png",
        "image/gif",
        "image/webp",
        
        // Videos (already compressed)
        "video/mp4",
        "video/webm",
        "video/mpeg",
        
        // Archives (already compressed)
        "application/zip",
        "application/gzip",
        "application/x-rar-compressed",
        
        // Fonts (consider compression)
        "font/woff",
        "font/woff2",
    },
}))
```

---

## YAML Configuration

```yaml
middlewares:
  - type: gzip_compression
    params:
      min_size: 1024
      compression_level: -1  # Default
      excluded_content_types:
        - "image/jpeg"
        - "image/png"
        - "video/mp4"
        - "application/zip"
```

---

## Compression Levels

| Level | Name | Speed | Ratio | Use Case |
|-------|------|-------|-------|----------|
| -1 | Default | Balanced | Balanced | Most cases (recommended) |
| 0 | No Compression | Fastest | None | Debugging |
| 1 | Best Speed | Very Fast | Low | Real-time APIs |
| 6 | Default | Medium | Good | General use |
| 9 | Best Compression | Slow | High | Static assets |

---

## Examples

### API Server

```go
// Fast compression for API responses
router.Use(gzipcompression.Middleware(&gzipcompression.Config{
    MinSize:          512, // Compress responses > 512 bytes
    CompressionLevel: gzip.BestSpeed,
}))
```

---

### Static File Server

```go
// Best compression for static assets
staticRouter := router.Group("/static")
staticRouter.Use(gzipcompression.Middleware(&gzipcompression.Config{
    MinSize:          1024,
    CompressionLevel: gzip.BestCompression,
    ExcludedContentTypes: []string{
        "image/jpeg",
        "image/png",
        "video/mp4",
    },
}))
```

---

### Mixed Content

```go
// Different compression for different routes
apiRouter := router.Group("/api")
apiRouter.Use(gzipcompression.Middleware(&gzipcompression.Config{
    MinSize:          512,
    CompressionLevel: gzip.BestSpeed, // Fast for API
}))

contentRouter := router.Group("/content")
contentRouter.Use(gzipcompression.Middleware(&gzipcompression.Config{
    MinSize:          2048,
    CompressionLevel: gzip.BestCompression, // Best for content
}))
```

---

### Environment-Based Configuration

```go
var compressionLevel int

if os.Getenv("ENV") == "production" {
    compressionLevel = gzip.BestCompression
} else {
    compressionLevel = gzip.BestSpeed // Faster development
}

router.Use(gzipcompression.Middleware(&gzipcompression.Config{
    MinSize:          1024,
    CompressionLevel: compressionLevel,
}))
```

---

### Conditional Compression

```go
// Only compress JSON and HTML
router.Use(gzipcompression.Middleware(&gzipcompression.Config{
    MinSize:          512,
    CompressionLevel: gzip.DefaultCompression,
    ExcludedContentTypes: []string{
        // Exclude everything except JSON and HTML
        "image/*",
        "video/*",
        "audio/*",
        "application/pdf",
        "application/zip",
    },
}))
```

---

## Behavior

### Compression Decision Flow

1. **Check client support**
   - If `Accept-Encoding` header doesn't contain "gzip" → skip

2. **Check content type**
   - If content type in `ExcludedContentTypes` → skip

3. **Check response size**
   - If response < `MinSize` → skip (not worth compressing)

4. **Compress response**
   - Set `Content-Encoding: gzip` header
   - Remove `Content-Length` header (will change)
   - Compress response body

---

### Headers

**Client Request:**
```http
GET /api/users HTTP/1.1
Accept-Encoding: gzip, deflate, br
```

**Server Response (Compressed):**
```http
HTTP/1.1 200 OK
Content-Type: application/json
Content-Encoding: gzip
Vary: Accept-Encoding
Transfer-Encoding: chunked

[compressed data]
```

**Server Response (Not Compressed):**
```http
HTTP/1.1 200 OK
Content-Type: image/jpeg
Content-Length: 45678

[raw data]
```

---

## Compression Ratios

Typical compression ratios by content type:

| Content Type | Typical Ratio | Example |
|--------------|---------------|---------|
| JSON | 70-90% | 100KB → 10-30KB |
| HTML | 60-80% | 100KB → 20-40KB |
| CSS | 60-80% | 100KB → 20-40KB |
| JavaScript | 60-75% | 100KB → 25-40KB |
| Plain Text | 60-80% | 100KB → 20-40KB |
| XML | 70-85% | 100KB → 15-30KB |
| JPEG | 0-5% | Already compressed |
| PNG | 0-10% | Already compressed |
| Video | 0-5% | Already compressed |

---

## Best Practices

### 1. Place Last in Middleware Chain

```go
// ✅ Good - compress final response
router.Use(
    recovery.Middleware(&recovery.Config{}),
    request_logger.Middleware(&request_logger.Config{}),
    // ... other middleware
    gzipcompression.Middleware(&gzipcompression.Config{}), // Last
)

// 🚫 Bad - compresses intermediate responses
router.Use(
    gzipcompression.Middleware(&gzipcompression.Config{}), // Too early
    request_logger.Middleware(&request_logger.Config{}),
)
```

---

### 2. Set Appropriate Minimum Size

```go
// ✅ Good - don't compress tiny responses
gzipcompression.Middleware(&gzipcompression.Config{
    MinSize: 1024, // 1KB minimum
})

// 🚫 Bad - wastes CPU on small responses
gzipcompression.Middleware(&gzipcompression.Config{
    MinSize: 0, // Compresses everything
})
```

---

### 3. Exclude Pre-Compressed Content

```go
// ✅ Good - skip already compressed content
ExcludedContentTypes: []string{
    "image/jpeg",
    "video/mp4",
    "application/zip",
    "font/woff2",
}

// 🚫 Bad - wastes CPU trying to compress compressed data
ExcludedContentTypes: []string{}
```

---

### 4. Use Default Compression Level

```go
// ✅ Good - balanced speed/ratio
gzipcompression.Middleware(&gzipcompression.Config{
    CompressionLevel: gzip.DefaultCompression, // -1
})

// 🚫 Bad - slow for marginal gains
gzipcompression.Middleware(&gzipcompression.Config{
    CompressionLevel: gzip.BestCompression, // Too slow
})
```

---

### 5. Consider Content Type

```go
// ✅ Good - compress text-based content only
ExcludedContentTypes: []string{
    "image/*",
    "video/*",
    "audio/*",
    "application/zip",
    "application/gzip",
}

// 🚫 Bad - tries to compress everything
ExcludedContentTypes: []string{}
```

---

## Performance

### Overhead

| Compression Level | Speed | CPU Usage | Ratio |
|-------------------|-------|-----------|-------|
| 0 (No compression) | 0μs | 0% | 0% |
| 1 (Best speed) | 50-100μs | Low | 40-60% |
| -1 (Default) | 100-200μs | Medium | 60-70% |
| 9 (Best compression) | 200-500μs | High | 70-80% |

**For 10KB response:**
- Level 1: ~50μs
- Level 6 (default): ~150μs
- Level 9: ~300μs

**Savings vs Cost:**
```
10KB JSON response:
- Uncompressed: 10KB, 0μs CPU
- Compressed (level 1): 4KB, 50μs CPU
- Network savings: 6KB (60%)
- Time saved: ~60ms on 1Mbps connection

Trade-off: 50μs CPU for 60ms network time → 1200x benefit
```

---

## Testing

### Test Compression

```go
func TestGzipCompression(t *testing.T) {
    router := lokstra.NewRouter()
    router.Use(gzipcompression.Middleware(&gzipcompression.Config{
        MinSize: 100,
    }))
    
    router.GET("/data", func(c *request.Context) error {
        // Large response that should be compressed
        largeData := strings.Repeat("test", 1000)
        return c.Api.Ok(largeData)
    })
    
    req := httptest.NewRequest("GET", "/data", nil)
    req.Header.Set("Accept-Encoding", "gzip")
    rec := httptest.NewRecorder()
    
    router.ServeHTTP(rec, req)
    
    assert.Equal(t, 200, rec.Code)
    assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
}
```

---

### Test Exclusions

```go
func TestExcludedContentType(t *testing.T) {
    router := lokstra.NewRouter()
    router.Use(gzipcompression.Middleware(&gzipcompression.Config{
        MinSize: 100,
        ExcludedContentTypes: []string{"image/jpeg"},
    }))
    
    router.GET("/image", func(c *request.Context) error {
        c.W.Header().Set("Content-Type", "image/jpeg")
        return c.Api.Ok([]byte("fake image data"))
    })
    
    req := httptest.NewRequest("GET", "/image", nil)
    req.Header.Set("Accept-Encoding", "gzip")
    rec := httptest.NewRecorder()
    
    router.ServeHTTP(rec, req)
    
    assert.Equal(t, "", rec.Header().Get("Content-Encoding"))
}
```

---

### Test Minimum Size

```go
func TestMinSize(t *testing.T) {
    router := lokstra.NewRouter()
    router.Use(gzipcompression.Middleware(&gzipcompression.Config{
        MinSize: 1024, // 1KB
    }))
    
    router.GET("/small", func(c *request.Context) error {
        return c.Api.Ok("small") // < 1KB, not compressed
    })
    
    req := httptest.NewRequest("GET", "/small", nil)
    req.Header.Set("Accept-Encoding", "gzip")
    rec := httptest.NewRecorder()
    
    router.ServeHTTP(rec, req)
    
    assert.Equal(t, "", rec.Header().Get("Content-Encoding"))
}
```

---

## Common Issues

### Issue: Response Not Compressed

**Possible Causes:**
1. Client doesn't send `Accept-Encoding: gzip`
2. Response smaller than `MinSize`
3. Content type in `ExcludedContentTypes`
4. Middleware not applied

**Solution:**
```go
// Verify client supports gzip
curl -H "Accept-Encoding: gzip" http://localhost:8080/api/data -v

// Check middleware is applied
router.Use(gzipcompression.Middleware(&gzipcompression.Config{
    MinSize: 100, // Lower threshold for testing
}))
```

---

### Issue: Images Being Compressed

**Problem:** Wasting CPU on already-compressed images

**Solution:**
```go
// Exclude all images
ExcludedContentTypes: []string{
    "image/*", // Wildcard not supported, list explicitly
    "image/jpeg",
    "image/png",
    "image/gif",
    "image/webp",
}
```

---

### Issue: Slow Response Times

**Problem:** High compression level causing delays

**Solution:**
```go
// Use faster compression
gzipcompression.Middleware(&gzipcompression.Config{
    CompressionLevel: gzip.BestSpeed, // Level 1
})
```

---

## Advanced Usage

### Dynamic Compression Level

```go
func getCompressionLevel(path string) int {
    if strings.HasPrefix(path, "/api/") {
        return gzip.BestSpeed // Fast for API
    }
    return gzip.BestCompression // Best for content
}

// Note: Config is static, use different routers instead
apiRouter.Use(gzipcompression.Middleware(&gzipcompression.Config{
    CompressionLevel: gzip.BestSpeed,
}))

contentRouter.Use(gzipcompression.Middleware(&gzipcompression.Config{
    CompressionLevel: gzip.BestCompression,
}))
```

---

### Pre-Compressed Static Files

```go
// Serve pre-compressed files if available
router.GET("/static/*", func(c *request.Context) error {
    path := c.Params.Get("*")
    
    // Check for .gz version
    if strings.Contains(c.R.Header.Get("Accept-Encoding"), "gzip") {
        gzPath := path + ".gz"
        if fileExists(gzPath) {
            c.W.Header().Set("Content-Encoding", "gzip")
            return c.File(gzPath)
        }
    }
    
    return c.File(path)
})
```

---

## See Also

- **[Body Limit](./body-limit.md)** - Request size protection
- **[Request Logger](./request-logger.md)** - Request logging
- **[Response](../01-core-packages/response.md)** - Response formatting

---

## Related Guides

- **[Performance Optimization](../../04-guides/performance/)** - Performance tips
- **[Static Assets](../../04-guides/static-assets/)** - Static file serving
- **[CDN Integration](../../04-guides/cdn/)** - CDN setup
