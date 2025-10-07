# 04. Middleware Patterns

Middleware in Lokstra wraps handlers to add cross-cutting concerns like logging, authentication, rate limiting, and recovery.

## Key Concepts

### 1. Middleware Signature
```go
func myMiddleware(c *lokstra.RequestContext) error {
    // Before handler
    
    err := c.Next() // Continue to next middleware or handler
    
    // After handler (runs in reverse order)
    return err
}
```

### 2. Applying Middleware
**Important:** Handler FIRST, then middleware(s)

```go
r.GET("/path", 
    handlerFunc,      // Handler first
    middleware1,      // Then middleware
    middleware2,      // More middleware
)
```

### 3. Execution Order (Onion Layers)
```
Request → MW2 → MW1 → Handler → MW1 → MW2 → Response
          ↓     ↓       ↓        ↑     ↑
        Before Before Process  After After
```

Code after `c.Next()` runs in **reverse order**.

## Middleware Examples

### 1. Logging Middleware
```go
loggingMiddleware := func(c *lokstra.RequestContext) error {
    start := time.Now()
    log.Printf("→ %s %s", c.R.Method, c.R.URL.Path)
    
    err := c.Next()
    
    log.Printf("← %s %s [%v]", c.R.Method, c.R.URL.Path, time.Since(start))
    return err
}
```

**Use case:** Request/response logging, timing, debugging

### 2. Authentication Middleware
```go
authMiddleware := func(c *lokstra.RequestContext) error {
    token := c.Req.HeaderParam("Authorization", "")
    
    if token == "" {
        return c.Api.Unauthorized("Missing token")
    }
    
    // Validate token...
    
    // Store user info
    c.Set("user_id", "user-123")
    c.Set("username", "john")
    
    return c.Next()
}
```

**Use case:** Protect endpoints, verify tokens, store user context

**Stop chain:** Return error without calling `c.Next()`

### 3. Authorization Middleware
```go
adminOnlyMiddleware := func(c *lokstra.RequestContext) error {
    role := c.Get("role")
    
    if role != "admin" {
        return c.Api.Forbidden("Admin access required")
    }
    
    return c.Next()
}
```

**Use case:** Role-based access control

### 4. Recovery Middleware
```go
recoveryMiddleware := func(c *lokstra.RequestContext) error {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("PANIC: %v", r)
            c.Api.InternalError("Internal server error")
        }
    }()
    
    return c.Next()
}
```

**Use case:** Catch panics, prevent server crashes

### 5. Rate Limiting Middleware
```go
rateLimitMiddleware := func(c *lokstra.RequestContext) error {
    ip := c.R.RemoteAddr
    
    if tooManyRequests(ip) {
        return c.Api.Error(429, "RATE_LIMIT", "Too many requests")
    }
    
    recordRequest(ip)
    return c.Next()
}
```

**Use case:** Prevent abuse, protect resources

### 6. CORS Middleware
```go
corsMiddleware := func(c *lokstra.RequestContext) error {
    c.W.Header().Set("Access-Control-Allow-Origin", "*")
    c.W.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
    
    if c.R.Method == "OPTIONS" {
        c.W.WriteHeader(200)
        return nil // Stop here for preflight
    }
    
    return c.Next()
}
```

**Use case:** Enable cross-origin requests

## Middleware Patterns

### Pattern 1: Single Middleware
```go
r.GET("/api/data", handler, loggingMiddleware)
```

### Pattern 2: Multiple Middleware
```go
r.GET("/api/protected", 
    handler,
    loggingMiddleware,
    authMiddleware,
)
```

### Pattern 3: Stacked Middleware
```go
r.POST("/api/secure",
    handler,
    recoveryMiddleware,   // Outermost - catches all panics
    loggingMiddleware,    // Logs timing
    corsMiddleware,       // CORS headers
    rateLimitMiddleware,  // Rate limit
    authMiddleware,       // Authentication
)

// Execution order:
// Request → recovery → logging → cors → rateLimit → auth → handler
//         ← recovery ← logging ← cors ← rateLimit ← auth ← Response
```

### Pattern 4: Conditional Middleware
```go
// Middleware factory
func roleMiddleware(role string) func(*lokstra.RequestContext) error {
    return func(c *lokstra.RequestContext) error {
        c.Set("role", role)
        return c.Next()
    }
}

// Usage
r.GET("/admin", handler, authMiddleware, roleMiddleware("admin"), adminOnlyMiddleware)
r.GET("/user", handler, authMiddleware, roleMiddleware("user"))
```

## Context Storage for Middleware Communication

Middleware can store data for handlers:

```go
// Middleware sets data
authMiddleware := func(c *lokstra.RequestContext) error {
    c.Set("user_id", "123")
    c.Set("username", "john")
    c.Set("role", "admin")
    return c.Next()
}

// Handler reads data
handler := func(c *lokstra.RequestContext) error {
    userID := c.Get("user_id")
    username := c.Get("username")
    // Use the data...
}
```

## Important Rules

✅ **DO:**
- Call `c.Next()` to continue the chain
- Return errors to stop execution
- Use `c.Set()` / `c.Get()` to share data
- Put recovery middleware outermost
- Put logging middleware early

❌ **DON'T:**
- Call `c.Next()` in handlers (only in middleware!)
- Forget to return `c.Next()`'s error
- Rely on middleware order accidentally
- Mix handler and middleware concepts

## Running the Example

```bash
go run .
```

Then test with `test.http` or curl commands shown in the server output.

## Middleware vs Handler

| Feature | Middleware | Handler |
|---------|-----------|---------|
| Calls `c.Next()` | ✅ Yes | ❌ No |
| Can stop chain | ✅ Yes (return error) | N/A |
| Runs before/after | ✅ Both | ❌ Only once |
| Purpose | Cross-cutting concerns | Business logic |

## Next Steps

- **[05-services](../05-services/)** - Use services in handlers
- **[02-architecture](../../02-architecture/)** - Config-driven middleware
- **[03-best-practices](../../03-best-practices/)** - Production middleware patterns
