# Middleware Error Handling Best Practices

## 📋 **ShouldStopMiddlewareChain Helper**

The `ShouldStopMiddlewareChain` helper method provides **consistent error checking** across all middleware implementations in Lokstra.

### ✅ **Correct Usage**

```go
ctx.RegisterMiddlewareFunc("example_middleware", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
    return func(ctx *lokstra.Context) error {
        // Pre-processing logic
        lokstra.Logger.Infof("Before processing")
        
        // Call next middleware/handler
        err := next(ctx)
        
        // Use helper for consistent error checking
        if ctx.ShouldStopMiddlewareChain(err) {
            return err
        }
        
        // Post-processing logic (only if no error and status < 400)
        lokstra.Logger.Infof("After processing")
        return nil
    }
})
```

### ❌ **Incorrect Usage (Manual Checking)**

```go
// Don't do this - prone to errors and inconsistent
if err := next(ctx); err != nil || ctx.StatusCode >= 400 {
    return err
}
```

## 🏗️ **Framework Integration**

The Lokstra framework automatically uses this error handling pattern in its **middleware composition**:

### **Regular Routes** (`composeMiddleware`)
- Used for all standard HTTP routes (GET, POST, PUT, etc.)
- Automatically detects error status and stops middleware chain
- Ensures consistent behavior across all route handlers

### **Reverse Proxy Routes** (`composeReverseProxyMw`) 
- Used for reverse proxy mounted routes
- Same error detection logic for proxy middleware
- Consistent behavior between proxy and regular routes

This means **all middleware** in Lokstra benefits from the same error handling pattern, whether you're using regular routes or reverse proxy routes.

## 🔍 **Logic Explanation**

The helper returns `true` (stop middleware chain) when:

1. **`err != nil`**: Any error occurred in next middleware/handler
2. **`ctx.StatusCode >= 400`**: HTTP error status (400-499 client errors, 500+ server errors)

## 🎯 **Benefits**

1. **Consistency**: Same error logic across all middleware
2. **Maintainability**: Single place to update error handling logic
3. **Readability**: Clear intent with descriptive method name
4. **Less Error-Prone**: No manual status code comparisons

## 📚 **Common HTTP Status Codes**

| Status | Range | Description | Helper Result |
|--------|-------|-------------|---------------|
| 200-299 | Success | Request succeeded | `false` (continue) |
| 300-399 | Redirection | Further action needed | `false` (continue) |
| 400-499 | Client Error | Bad request | `true` (stop) |
| 500-599 | Server Error | Server failed | `true` (stop) |

## 🔧 **Usage Examples**

### Authentication Middleware
```go
ctx.RegisterMiddlewareFunc("auth", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
    return func(ctx *lokstra.Context) error {
        token := ctx.Request.Header.Get("Authorization")
        if token == "" {
            return ctx.ErrorBadRequest("Authorization required")
        }
        
        err := next(ctx)
        if ctx.ShouldStopMiddlewareChain(err) {
            return err
        }
        
        // Log successful authenticated request
        lokstra.Logger.Infof("Authenticated request completed")
        return nil
    }
})
```

### Logging Middleware
```go
ctx.RegisterMiddlewareFunc("request_logger", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
    return func(ctx *lokstra.Context) error {
        start := time.Now()
        lokstra.Logger.Infof("Request started: %s %s", ctx.Request.Method, ctx.Request.URL.Path)
        
        err := next(ctx)
        if ctx.ShouldStopMiddlewareChain(err) {
            lokstra.Logger.Errorf("Request failed: %v", err)
            return err
        }
        
        duration := time.Since(start)
        lokstra.Logger.Infof("Request completed in %v", duration)
        return nil
    }
})
```

### Rate Limiting Middleware
```go
ctx.RegisterMiddlewareFunc("rate_limit", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
    return func(ctx *lokstra.Context) error {
        if !rateLimiter.Allow() {
            return ctx.ErrorBadRequest("Rate limit exceeded")
        }
        
        err := next(ctx)
        if ctx.ShouldStopMiddlewareChain(err) {
            return err
        }
        
        // Update rate limiting metrics on success
        rateLimiter.RecordSuccess()
        return nil
    }
})
```

## ⚠️ **Important Notes**

1. **Always call `next(ctx)` first** before checking with helper
2. **Pre-processing logic** goes before `next(ctx)` call
3. **Post-processing logic** goes after helper check
4. **Return early** if helper returns `true`
5. **Helper is read-only** - doesn't modify context state

## 🧪 **Testing**

The helper method is thoroughly tested with various scenarios:
- No error, success status (200, 201) → continue
- No error, error status (400+) → stop  
- Error present, any status → stop
- Boundary conditions (399 vs 400) → proper behavior

This ensures **reliable middleware behavior** across all use cases.
