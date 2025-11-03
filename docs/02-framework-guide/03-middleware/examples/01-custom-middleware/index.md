# Custom Middleware

> # Custom Middleware Example

This example demonstrates how to create custom middleware in Lokstra. Learn about middleware signatures, patterns, and best practices.

## What You'll Learn

- Middleware function signatures
- Request/response interception
- Context data storage and retrieval
- Authorization patterns
- Rate limiting implementations
- Logging and timing middleware

## Running the Example

```bash
go run main.go
```

The server will start on `http://localhost:3000`

## Example Middleware Implementations

### 1. Logging Middleware

Logs every request with method, path, and processing time:

```go
func LoggingMiddleware(c *request.Context) error {
    start := time.Now()
    log.Printf("[%s] %s - Started", c.R.Method, c.R.URL.Path)
    
    err := c.Next()
    
    duration := time.Since(start)
    log.Printf("[%s] %s - Completed in %v", c.R.Method, c.R.URL.Path, duration)
    
    return err
}
```

**Pattern**: Measure before/after calling `c.Next()`

### 2. Request ID Middleware

Generates unique IDs for request tracking:

```go
func RequestIDMiddleware() func(c *request.Context) error {
    counter := 0
    mu := sync.Mutex{}
    
    return func(c *request.Context) error {
        mu.Lock()
        counter++
        requestID := fmt.Sprintf("req-%d-%d", time.Now().Unix(), counter)
        mu.Unlock()
        
        c.Set("request_id", requestID)
        return c.Next()
    }
}
```

**Pattern**: Generate metadata and store in context with `c.Set()`

### 3. Authorization Middleware

Validates authorization headers:

```go
func AuthMiddleware(c *request.Context) error {
    authHeader := c.R.Header.Get("Authorization")
    
    if authHeader == "" {
        return fmt.Errorf("missing authorization header")
    }
    
    if !strings.HasPrefix(authHeader, "Bearer ") {
        return fmt.Errorf("invalid authorization format")
    }
    
    token := strings.TrimPrefix(authHeader, "Bearer ")
    if token != "secret-token-123" {
        return fmt.Errorf("invalid token")
    }
    
    c.Set("user_id", "user-123")
    return c.Next()
}
```

**Pattern**: Early return on validation failure

### 4. Rate Limiter Middleware

Limits requests per IP address:

```go
func RateLimiterMiddleware(limit int, window time.Duration) func(c *request.Context) error {
    requests := make(map[string][]time.Time)
    mu := sync.Mutex{}
    
    return func(c *request.Context) error {
        ip := c.R.RemoteAddr
        now := time.Now()
        
        mu.Lock()
        defer mu.Unlock()
        
        // Clean old entries
        times := requests[ip]
        validTimes := []time.Time{}
        for _, t := range times {
            if now.Sub(t) < window {
                validTimes = append(validTimes, t)
            }
        }
        
        if len(validTimes) >= limit {
            return fmt.Errorf("rate limit exceeded")
        }
        
        validTimes = append(validTimes, now)
        requests[ip] = validTimes
        
        mu.Unlock()
        err := c.Next()
        mu.Lock()
        
        return err
    }
}
```

**Pattern**: Stateful middleware with in-memory storage

## Testing

Use the included `test.http` file to test all endpoints:

- `GET /` - Home page with links
- `GET /public` - Public endpoint (no auth)
- `GET /protected` - Protected endpoint (auth required)
- `GET /limited` - Rate limited endpoint (5 req/min)

## Key Concepts

### Middleware Signature

All middleware must follow this signature:

```go
func(c *request.Context) error
```

### Calling Next Handler

Use `c.Next()` to continue the middleware chain:

```go
err := c.Next()
return err
```

### Accessing Request Data

Use `c.R` to access the standard `*http.Request`:

```go
method := c.R.Method
path := c.R.URL.Path
header := c.R.Header.Get("Authorization")
ip := c.R.RemoteAddr
```

### Storing Context Data

Use `c.Set()` and `c.Get()` for request-scoped data:

```go
// Store
c.Set("user_id", "123")

// Retrieve
if userID, ok := c.Get("user_id"); ok {
    log.Println("User:", userID)
}
```

### Error Handling

Return errors from middleware - handlers will deal with responses:

```go
if authHeader == "" {
    return fmt.Errorf("missing authorization")
}
```

### Middleware Registration

```go
// Global middleware (applies to all routes)
router.Use(LoggingMiddleware)

// Route-specific middleware
router.GET("/protected", handler, AuthMiddleware)

// Multiple middleware (execute in order)
router.Use(middleware1, middleware2, middleware3)
```

## Common Patterns

### Pre/Post Processing

```go
func TimingMiddleware(c *request.Context) error {
    start := time.Now()
    
    err := c.Next() // Process request
    
    duration := time.Since(start)
    log.Printf("Took: %v", duration)
    return err
}
```

### Early Exit

```go
func AuthMiddleware(c *request.Context) error {
    if !isAuthorized(c) {
        return fmt.Errorf("unauthorized")
    }
    return c.Next()
}
```

### State Management

```go
func CounterMiddleware() func(c *request.Context) error {
    count := 0
    mu := sync.Mutex{}
    
    return func(c *request.Context) error {
        mu.Lock()
        count++
        c.Set("request_number", count)
        mu.Unlock()
        
        return c.Next()
    }
}
```

## Next Steps

- Explore [02-composition](../02-composition/) for middleware chaining patterns
- See [03-context-management](../03-context-management/) for advanced context usage
- Check [04-error-recovery](../04-error-recovery/) for panic recovery patterns - This example is being prepared.

## Topics Covered

Creation patterns, context handling

## Placeholder

This example is being prepared. Check back soon for:
- Working code examples
- Comprehensive documentation
- Test files
- Best practices guide

---

**Status**: 📝 In Development
