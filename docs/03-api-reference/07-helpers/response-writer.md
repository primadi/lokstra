# Response Writer Package

The `response_writer` package provides utilities for HTTP response handling, including buffered response writers for middleware that need to inspect or modify response bodies.

## Table of Contents

- [Overview](#overview)
- [BufferedBodyWriter](#bufferedbodywriter)
- [Use Cases](#use-cases)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Overview

**Import Path:** `github.com/primadi/lokstra/common/response_writer`

**Key Features:**

```
✓ Buffered Response     - Capture response body before sending
✓ Status Code Capture   - Intercept HTTP status codes
✓ Middleware Support    - Enable response inspection/modification
✓ Standard Interface    - Implements http.ResponseWriter
```

## BufferedBodyWriter

Captures HTTP response in memory before sending to client.

### Structure

```go
type BufferedBodyWriter struct {
    http.ResponseWriter           // Original response writer
    Buf  bytes.Buffer             // Buffered response body
    Code int                      // HTTP status code
}
```

### Creation

```go
func NewBufferedBodyWriter(w http.ResponseWriter) *BufferedBodyWriter
```

Creates a new buffered writer wrapping the original response writer.

### Methods

#### Write

```go
func (d *BufferedBodyWriter) Write(b []byte) (int, error)
```

Writes data to internal buffer instead of sending to client:

```go
writer := response_writer.NewBufferedBodyWriter(w)
writer.Write([]byte("Hello"))  // Buffered, not sent
writer.Write([]byte(" World")) // Buffered, not sent

// Content in buffer: "Hello World"
```

#### WriteHeader

```go
func (d *BufferedBodyWriter) WriteHeader(code int)
```

Captures status code without sending to client:

```go
writer := response_writer.NewBufferedBodyWriter(w)
writer.WriteHeader(http.StatusNotFound)  // Captured, not sent

// writer.Code = 404
```

### Accessing Buffer

```go
writer := response_writer.NewBufferedBodyWriter(w)

// Write response
json.NewEncoder(writer).Encode(data)

// Read buffered content
body := writer.Buf.Bytes()
bodyString := writer.Buf.String()

// Get status code
statusCode := writer.Code
```

### Sending to Client

After inspecting/modifying, send to client:

```go
// Send status code
if writer.Code != 0 {
    w.WriteHeader(writer.Code)
}

// Send body
w.Write(writer.Buf.Bytes())
```

## Use Cases

### Logging Middleware

Capture response for logging:

```go
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Create buffered writer
        buffered := response_writer.NewBufferedBodyWriter(w)
        
        // Call next handler (response goes to buffer)
        next.ServeHTTP(buffered, r)
        
        // Log response
        log.Printf(
            "Method: %s, Path: %s, Status: %d, Body: %s",
            r.Method,
            r.URL.Path,
            buffered.Code,
            buffered.Buf.String(),
        )
        
        // Send to client
        if buffered.Code != 0 {
            w.WriteHeader(buffered.Code)
        }
        w.Write(buffered.Buf.Bytes())
    })
}
```

### Error Transformation

Transform error responses:

```go
func ErrorTransformMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        buffered := response_writer.NewBufferedBodyWriter(w)
        
        next.ServeHTTP(buffered, r)
        
        // Check if error response
        if buffered.Code >= 400 {
            // Parse original error
            var originalError map[string]any
            json.Unmarshal(buffered.Buf.Bytes(), &originalError)
            
            // Create standardized error
            standardError := map[string]any{
                "status": "error",
                "error": map[string]any{
                    "code":    buffered.Code,
                    "message": originalError["message"],
                    "details": originalError,
                },
                "timestamp": time.Now().Format(time.RFC3339),
            }
            
            // Send transformed error
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(buffered.Code)
            json.NewEncoder(w).Encode(standardError)
            return
        }
        
        // Non-error response - send as-is
        if buffered.Code != 0 {
            w.WriteHeader(buffered.Code)
        }
        w.Write(buffered.Buf.Bytes())
    })
}
```

### Response Compression

Compress large responses:

```go
func CompressionMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Check if client accepts gzip
        if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
            next.ServeHTTP(w, r)
            return
        }
        
        buffered := response_writer.NewBufferedBodyWriter(w)
        next.ServeHTTP(buffered, r)
        
        // Only compress if response is large enough
        if buffered.Buf.Len() < 1024 {
            // Too small, send uncompressed
            if buffered.Code != 0 {
                w.WriteHeader(buffered.Code)
            }
            w.Write(buffered.Buf.Bytes())
            return
        }
        
        // Compress response
        var compressed bytes.Buffer
        gzipWriter := gzip.NewWriter(&compressed)
        gzipWriter.Write(buffered.Buf.Bytes())
        gzipWriter.Close()
        
        // Send compressed
        w.Header().Set("Content-Encoding", "gzip")
        w.Header().Set("Content-Length", strconv.Itoa(compressed.Len()))
        if buffered.Code != 0 {
            w.WriteHeader(buffered.Code)
        }
        w.Write(compressed.Bytes())
    })
}
```

### Response Validation

Validate response format:

```go
func ResponseValidationMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        buffered := response_writer.NewBufferedBodyWriter(w)
        next.ServeHTTP(buffered, r)
        
        // Validate JSON response
        if strings.Contains(w.Header().Get("Content-Type"), "application/json") {
            var jsonData map[string]any
            if err := json.Unmarshal(buffered.Buf.Bytes(), &jsonData); err != nil {
                // Invalid JSON
                http.Error(w, "Invalid JSON response", http.StatusInternalServerError)
                return
            }
            
            // Check required fields
            if _, ok := jsonData["status"]; !ok {
                log.Printf("WARNING: Response missing 'status' field")
            }
        }
        
        // Send response
        if buffered.Code != 0 {
            w.WriteHeader(buffered.Code)
        }
        w.Write(buffered.Buf.Bytes())
    })
}
```

### Metrics Collection

Collect response metrics:

```go
func MetricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        buffered := response_writer.NewBufferedBodyWriter(w)
        next.ServeHTTP(buffered, r)
        
        duration := time.Since(start)
        
        // Record metrics
        metrics.RecordRequest(
            r.Method,
            r.URL.Path,
            buffered.Code,
            buffered.Buf.Len(),
            duration,
        )
        
        // Send response
        if buffered.Code != 0 {
            w.WriteHeader(buffered.Code)
        }
        w.Write(buffered.Buf.Bytes())
    })
}
```

## Best Practices

### Memory Management

```go
✓ DO: Use for middleware that needs response inspection
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        buffered := response_writer.NewBufferedBodyWriter(w)
        next.ServeHTTP(buffered, r)
        
        // Inspect and log
        log.Printf("Response: %d - %d bytes", buffered.Code, buffered.Buf.Len())
        
        // Send to client
        if buffered.Code != 0 {
            w.WriteHeader(buffered.Code)
        }
        w.Write(buffered.Buf.Bytes())
    })
}

✗ DON'T: Use for large file responses
func FileDownloadHandler(w http.ResponseWriter, r *http.Request) {
    // BAD: Buffers entire file in memory
    buffered := response_writer.NewBufferedBodyWriter(w)
    
    // This loads entire file into buffer
    http.ServeFile(buffered, r, "/path/to/large/file.zip")
    
    w.Write(buffered.Buf.Bytes())
}
```

### Status Code Handling

```go
✓ DO: Check if status code was set
if buffered.Code != 0 {
    w.WriteHeader(buffered.Code)
}
w.Write(buffered.Buf.Bytes())

✓ DO: Provide default status
statusCode := buffered.Code
if statusCode == 0 {
    statusCode = http.StatusOK
}
w.WriteHeader(statusCode)
w.Write(buffered.Buf.Bytes())

✗ DON'T: Always write status
w.WriteHeader(buffered.Code)  // BAD: Will write 0 if not set
```

### Header Handling

```go
✓ DO: Copy headers before writing status
// Copy headers from original writer
for key, values := range buffered.Header() {
    for _, value := range values {
        w.Header().Add(key, value)
    }
}

// Then write status and body
if buffered.Code != 0 {
    w.WriteHeader(buffered.Code)
}
w.Write(buffered.Buf.Bytes())

✗ DON'T: Forget to propagate headers
// Headers set by handler are lost
if buffered.Code != 0 {
    w.WriteHeader(buffered.Code)
}
w.Write(buffered.Buf.Bytes())
```

### Error Handling

```go
✓ DO: Handle write errors
if buffered.Code != 0 {
    w.WriteHeader(buffered.Code)
}
if _, err := w.Write(buffered.Buf.Bytes()); err != nil {
    log.Printf("Failed to write response: %v", err)
}

✓ DO: Validate buffer content before sending
if buffered.Buf.Len() == 0 && buffered.Code != http.StatusNoContent {
    log.Printf("WARNING: Empty response body with status %d", buffered.Code)
}
```

## Examples

### Complete Logging Middleware

```go
func LoggingMiddleware(logger *log.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            // Create buffered writer
            buffered := response_writer.NewBufferedBodyWriter(w)
            
            // Call next handler
            next.ServeHTTP(buffered, r)
            
            // Calculate duration
            duration := time.Since(start)
            
            // Determine status code
            statusCode := buffered.Code
            if statusCode == 0 {
                statusCode = http.StatusOK
            }
            
            // Log request/response
            logger.Printf(
                "[%s] %s %s - Status: %d, Size: %d bytes, Duration: %v",
                time.Now().Format(time.RFC3339),
                r.Method,
                r.URL.Path,
                statusCode,
                buffered.Buf.Len(),
                duration,
            )
            
            // Log response body for errors
            if statusCode >= 400 {
                logger.Printf("Error Response Body: %s", buffered.Buf.String())
            }
            
            // Copy headers
            for key, values := range buffered.Header() {
                for _, value := range values {
                    w.Header().Add(key, value)
                }
            }
            
            // Send response
            w.WriteHeader(statusCode)
            w.Write(buffered.Buf.Bytes())
        })
    }
}
```

### API Response Wrapper

```go
func APIResponseMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        buffered := response_writer.NewBufferedBodyWriter(w)
        next.ServeHTTP(buffered, r)
        
        statusCode := buffered.Code
        if statusCode == 0 {
            statusCode = http.StatusOK
        }
        
        // Wrap response in standard format
        var wrapper map[string]any
        
        if statusCode >= 400 {
            // Error response
            wrapper = map[string]any{
                "status": "error",
                "error":  json.RawMessage(buffered.Buf.Bytes()),
            }
        } else {
            // Success response
            wrapper = map[string]any{
                "status": "success",
                "data":   json.RawMessage(buffered.Buf.Bytes()),
            }
        }
        
        // Send wrapped response
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(statusCode)
        json.NewEncoder(w).Encode(wrapper)
    })
}
```

### Response Caching

```go
var responseCache sync.Map

func CachingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Only cache GET requests
        if r.Method != http.MethodGet {
            next.ServeHTTP(w, r)
            return
        }
        
        cacheKey := r.URL.Path
        
        // Check cache
        if cached, ok := responseCache.Load(cacheKey); ok {
            cachedResp := cached.(*CachedResponse)
            
            // Send cached response
            for key, values := range cachedResp.Headers {
                for _, value := range values {
                    w.Header().Add(key, value)
                }
            }
            w.WriteHeader(cachedResp.StatusCode)
            w.Write(cachedResp.Body)
            return
        }
        
        // Buffer response
        buffered := response_writer.NewBufferedBodyWriter(w)
        next.ServeHTTP(buffered, r)
        
        // Cache successful responses
        statusCode := buffered.Code
        if statusCode == 0 {
            statusCode = http.StatusOK
        }
        
        if statusCode >= 200 && statusCode < 300 {
            cachedResp := &CachedResponse{
                StatusCode: statusCode,
                Headers:    buffered.Header(),
                Body:       buffered.Buf.Bytes(),
            }
            responseCache.Repository(cacheKey, cachedResp)
        }
        
        // Send response
        if statusCode != 0 {
            w.WriteHeader(statusCode)
        }
        w.Write(buffered.Buf.Bytes())
    })
}

type CachedResponse struct {
    StatusCode int
    Headers    http.Header
    Body       []byte
}
```

## Related Documentation

- [Helpers Overview](index) - All helper packages
- [JSON Package](json) - JSON encoding/decoding
- [Middleware Documentation](../05-middleware) - Middleware patterns

---

**Section Complete:** All helper packages documented
