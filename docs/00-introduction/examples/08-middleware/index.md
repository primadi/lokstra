# Example 08: Middleware

⏱️ **30 minutes** • 🎯 **Intermediate**

**Master middleware patterns in Lokstra**

Learn how to use middleware for cross-cutting concerns like authentication, logging, recovery, and rate limiting.

---

## 📚 What You'll Learn

- ✅ **Built-in middlewares**: CORS, Recovery, Request Logger
- ✅ **Custom middlewares**: Auth, Rate limiting, Custom logging
- ✅ **Global middlewares**: Apply to all routes
- ✅ **Route-specific middlewares**: Apply to specific endpoints
- ✅ **Middleware chaining**: Multiple middlewares in sequence
- ✅ **Middleware factory pattern**: Create configurable middlewares
- ✅ **Context access**: Share data between middlewares and handlers
- ✅ **Config-based middleware**: Define middleware in YAML config
- ✅ **Lazy middleware resolution**: Use middleware by name before config loads

---

## 🏗️ Architecture

```
Request Flow:
┌─────────────────────────────────────────────────────────┐
│ Client Request                                          │
└───────────────┬─────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────┐
│ Global Middlewares (Applied to ALL routes)             │
│  1. Recovery      → Catch panics                        │
│  2. CORS          → Cross-origin handling               │
│  3. Logger        → Request/response logging            │
│  4. Custom Logger → Additional logging                  │
│  5. Rate Limiter  → Prevent abuse                       │
└───────────────┬─────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────┐
│ Route-Specific Middlewares                              │
│  - Auth          → Check API key                        │
│  - Admin Check   → Verify admin role                    │
│  - Custom        → Any route-specific logic             │
└───────────────┬─────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────┐
│ Handler → Business Logic                                │
└───────────────┬─────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────┐
│ Response                                                │
└─────────────────────────────────────────────────────────┘
```

---

## 🚀 Running the Example

```bash
cd docs/00-introduction/examples/08-middleware
go run main.go
```

The server starts on `http://localhost:3000`

---

## 📝 Testing

Use the provided `test.http` file with VS Code REST Client extension, or use curl:

### Public Endpoints (No Auth)

```bash
# Basic endpoint
curl http://localhost:3000/

# Public endpoint
curl http://localhost:3000/public

# Health check
curl http://localhost:3000/health
```

### Protected Endpoints (Requires Auth)

```bash
# Without API key → 401 Unauthorized
curl http://localhost:3000/protected

# With wrong key → 403 Forbidden
curl -H "X-API-Key: wrong-key" http://localhost:3000/protected

# With valid key → 200 OK
curl -H "X-API-Key: secret-key-123" http://localhost:3000/protected

# Profile endpoint
curl -H "X-API-Key: secret-key-123" http://localhost:3000/api/profile
```

### Admin Endpoints (Requires Admin Key)

```bash
# With regular key → 403 Forbidden
curl -H "X-API-Key: secret-key-123" http://localhost:3000/api/admin/dashboard

# With admin key → 200 OK
curl -H "X-API-Key: admin-key-456" http://localhost:3000/api/admin/dashboard
```

### Test Middlewares

```bash
# Test panic recovery
curl http://localhost:3000/panic

# Test slow request logging
curl http://localhost:3000/slow

# Test middleware chain
curl http://localhost:3000/chain
```

---

## 🔑 Key Concepts

### 1. **Config-Based Middleware (Recommended)**

Define middlewares in `config.yaml` for easy configuration management:

```yaml
middleware-definitions:
  recovery-prod:
    type: recovery
    config:
      log-stack-trace: true
      include-stack-in-response: false
  
  cors-api:
    type: cors
    config:
      allowed-origins: ["*"]
      allowed-methods: ["GET", "POST", "PUT", "DELETE"]
```

Then use them by name in code:

```go
// IMPORTANT: Middleware resolution is LAZY!
// You can Use() middleware names BEFORE loading config

// Step 1: Create router and register routes
r := lokstra.NewRouter("api")
r.Use("recovery-prod")  // ✓ OK! String stored, not resolved yet
r.Use("cors-api")       // ✓ OK! Lazy resolution

// Step 2: Load config (can be done AFTER router setup!)
lokstra_registry.LoadAndBuild([]string{"config.yaml"})

// Step 3: Build router - all middleware names resolved here
r.Build() // or app.Run() which calls Build()

// Or route-specific
r.GET("/api/users", handler, "jwt-auth", "cors-api")
```

**Benefits:**
- ✅ Change configuration without rebuilding
- ✅ Cleaner code (no CreateMiddleware calls)
- ✅ Centralized middleware management
- ✅ Easy to swap configs per environment (dev/prod)
- ✅ **Lazy resolution** - define routes before loading config!

### 2. **Built-in Middlewares**

Lokstra provides ready-to-use middlewares:

```go
// Register factories
cors.Register()
recovery.Register()
request_logger.Register()

// Programmatic registration (alternative to config)
lokstra_registry.RegisterMiddlewareName("cors-all", cors.CORS_TYPE, map[string]any{
    "allow_origins": []string{"*"},
})

// Use them
r.Use("cors-all")  // String name (if registered or in config)
// OR
r.Use(lokstra_registry.CreateMiddleware("cors-all"))  // Old way
```

**Available Built-in Middlewares:**
- `cors` - CORS handling
- `recovery` - Panic recovery
- `request_logger` - HTTP request/response logging
- `slow_request_logger` - Slow request detection
- `gzipcompression` - Response compression
- `body_limit` - Request body size limits
- `jwtauth` - JWT authentication
- `accesscontrol` - Role-based access control

### 3. **Custom Middleware**

Simple middleware signature: `func(*request.Context) error`

```go
func CustomAuthMiddleware(ctx *request.Context) error {
    apiKey := ctx.R.Header.Get("X-API-Key")
    if apiKey == "" {
        return ctx.Api.Unauthorized("Missing API key")
    }
    
    // Store data for later use
    ctx.Set("api_key", apiKey)
    return nil // Continue to next middleware/handler
}
```

### 4. **Middleware Factory Pattern**

Create configurable middleware:

```go
func RateLimitMiddleware(maxRequests int, window time.Duration) request.HandlerFunc {
    requests := make(map[string][]time.Time)
    
    return func(ctx *request.Context) error {
        ip := ctx.R.RemoteAddr
        // ... rate limit logic
        return nil
    }
}

// Use it
r.Use(RateLimitMiddleware(10, time.Minute)) // 10 requests per minute
```

### 5. **Global vs Route-Specific**

```go
// Global - Applied to ALL routes
r.Use(RecoveryMiddleware)
r.Use(CORSMiddleware)

// Route-specific - Only for this endpoint
r.GET("/protected", ProtectedHandler, AuthMiddleware)

// Multiple middlewares for one route (executed in order)
r.GET("/admin", AdminHandler, AuthMiddleware, AdminCheckMiddleware)
```

### 5. **Middleware Chaining**

Middlewares execute in order:

```go
r.GET("/endpoint", 
    Handler,
    Middleware1,  // Runs first
    Middleware2,  // Runs second
    Middleware3,  // Runs third
)
```

Each middleware must call `ctx.Next()` or return to continue the chain:

```go
func LoggingMiddleware(ctx *request.Context) error {
    start := time.Now()
    
    // Before handler
    log.Println("Before:", ctx.R.URL.Path)
    
    // Execute next middleware/handler
    err := ctx.Next()
    
    // After handler
    duration := time.Since(start)
    log.Println("After:", duration)
    
    return err
}
```

### 6. **Context Sharing**

Share data between middlewares and handlers:

```go
// In middleware
func AuthMiddleware(ctx *request.Context) error {
    user := authenticate(ctx)
    ctx.Set("user", user)      // Store in context
    return nil
}

// In handler
func ProfileHandler(ctx *request.Context) map[string]any {
    user := ctx.Get("user")    // Retrieve from context
    return map[string]any{
        "user": user,
    }
}
```

---

## 🎯 Common Middleware Patterns

### Authentication

```go
func AuthMiddleware(ctx *request.Context) error {
    token := ctx.R.Header.Get("Authorization")
    if token == "" {
        return ctx.Api.Unauthorized("Missing token")
    }
    
    user, err := validateToken(token)
    if err != nil {
        return ctx.Api.Forbidden("Invalid token")
    }
    
    ctx.Set("user", user)
    return nil
}
```

### Authorization

```go
func AdminOnlyMiddleware(ctx *request.Context) error {
    user := ctx.Get("user").(*User)
    if !user.IsAdmin {
        return ctx.Api.Forbidden("Admin access required")
    }
    return nil
}
```

### Rate Limiting

```go
func RateLimitMiddleware(max int, window time.Duration) request.HandlerFunc {
    limiter := rate.NewLimiter(rate.Every(window/time.Duration(max)), max)
    
    return func(ctx *request.Context) error {
        if !limiter.Allow() {
            return ctx.Api.Error(429, "RATE_LIMIT", "Too many requests")
        }
        return nil
    }
}
```

### Logging

```go
func LoggingMiddleware(ctx *request.Context) error {
    start := time.Now()
    
    err := ctx.Next()
    
    log.Printf("%s %s - %d (%v)",
        ctx.R.Method,
        ctx.R.URL.Path,
        ctx.W.StatusCode(),
        time.Since(start))
    
    return err
}
```

### Error Recovery

```go
func RecoveryMiddleware(ctx *request.Context) error {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("PANIC: %v\n%s", r, debug.Stack())
            ctx.Api.Error(500, "INTERNAL_ERROR", "Internal server error")
        }
    }()
    
    return ctx.Next()
}
```

---

## 📊 Execution Order

```
Request → Global MW 1 → Global MW 2 → Route MW 1 → Route MW 2 → Handler
                ↓             ↓             ↓             ↓          ↓
            ctx.Next()    ctx.Next()    ctx.Next()    ctx.Next()  return
                ↓             ↓             ↓             ↓          ↓
Response ← Global MW 1 ← Global MW 2 ← Route MW 1 ← Route MW 2 ← Handler
```

Each middleware can:
1. Execute code **before** the handler (`ctx.Next()`)
2. Execute code **after** the handler (after `ctx.Next()` returns)
3. Short-circuit the chain by returning early
4. Modify the request or response
5. Store data in context

---

## 💡 Best Practices

1. **Order Matters**: Recovery should be first, logging second
2. **Global vs Specific**: Use global for cross-cutting concerns, route-specific for targeted logic
3. **Fail Fast**: Auth/validation middlewares should fail early
4. **Share Data**: Use `ctx.Set()` / `ctx.Get()` to share between middlewares
5. **Error Handling**: Return errors properly, don't panic
6. **Performance**: Keep middlewares lightweight, avoid heavy operations
7. **Reusability**: Create factory functions for configurable middlewares

---

## 🔄 Comparison: Global vs Route-Specific

### Global Middleware

```go
// Applied to ALL routes
r.Use(RecoveryMiddleware)
r.Use(LoggingMiddleware)

r.GET("/public", PublicHandler)    // Has Recovery + Logging
r.GET("/private", PrivateHandler)  // Has Recovery + Logging
```

**Use for:**
- Recovery from panics
- CORS headers
- Request logging
- Rate limiting (global)
- Response compression

### Route-Specific Middleware

```go
// Applied only to specific routes
r.GET("/public", PublicHandler)                          // No auth
r.GET("/private", PrivateHandler, AuthMiddleware)        // Has auth
r.GET("/admin", AdminHandler, AuthMiddleware, AdminMW)   // Has auth + admin
```

**Use for:**
- Authentication
- Authorization
- Input validation
- Request-specific logging
- Feature flags

---

## 🎓 Learning Path

1. **Start Simple**: Global recovery and logging
2. **Add Auth**: Implement authentication middleware
3. **Authorization**: Add role-based access control
4. **Rate Limiting**: Prevent abuse
5. **Custom Logic**: Build your own middlewares

---

## 📚 Next Steps

- [Example 06: Auto-Router](../06-auto-router/) (coming soon) - Combine middleware with auto-generated routes
- [Built-in Middlewares Documentation](../../../../01-essentials/middleware/) - Deep dive into all built-in middlewares
- [Custom Middleware Guide](../../../../02-deep-dive/middleware/) - Advanced patterns

---

## 📖 Additional Documentation

- [NAMING-CONVENTIONS.md](./NAMING-CONVENTIONS.md) - Middleware naming best practices
- [RECOVERY-ANALYSIS.md](./RECOVERY-ANALYSIS.md) - Why recovery middleware is critical
- [RATE-LIMIT-PERFORMANCE.md](./RATE-LIMIT-PERFORMANCE.md) - Performance analysis

---

## ⚠️ Important Notes

### Recovery Middleware is CRITICAL

**Neither Go's ServeMux nor Chi router auto-recover from panics!**

Without recovery middleware, a single panic will **crash your entire server**. All users will be disconnected and the server must be manually restarted.

✅ **Always use recovery middleware in production:**

```go
// Development
r.Use("recovery-dev")  // Shows stack traces

// Production  
r.Use("recovery-prod") // Hides stack traces from clients
```

See [RECOVERY-ANALYSIS.md](./RECOVERY-ANALYSIS.md) for detailed explanation.

### Naming Conventions

For consistency across the framework:

- **Factory types**: Use `snake_case` (e.g., `request_logger`, `cors`)
- **Instance names**: Use `kebab-case` (e.g., `request-logger-verbose`, `cors-api`)

See [NAMING-CONVENTIONS.md](./NAMING-CONVENTIONS.md) for complete guide.

### Performance Notes

#### Endpoint Response Times

- **Fast endpoints** (`/`, `/public`, `/health`): ~10-20ms
  - Middleware overhead: < 2ms
  - Handler: < 1ms
  - Network latency: ~10ms

- **Slow endpoint** (`/slow`): ~2000ms
  - **Intentionally slow** (simulates long-running operation)
  - Uses `time.Sleep(2 * time.Second)` for testing

#### Middleware Overhead

Each middleware adds minimal overhead:
- RateLimitMiddleware: ~0.1-0.2ms
- LoggingMiddleware: ~0.05ms
- CORS: ~0.01ms
- Recovery: ~0.01ms

**Total middleware overhead: < 2ms**

Rate limiting does NOT slow down successful requests - it's just an in-memory map check.

See [RATE-LIMIT-PERFORMANCE.md](./RATE-LIMIT-PERFORMANCE.md) for benchmark details.

### ctx.Next() is MANDATORY

**CRITICAL:** In Lokstra, middleware MUST call `ctx.Next()` to continue the chain:

```go
func MyMiddleware(ctx *request.Context) error {
    // Pre-processing
    log.Println("Before handler")
    
    // ✅ MUST call ctx.Next() to continue
    err := ctx.Next()
    
    // Post-processing
    log.Println("After handler")
    
    return err
}
```

**Without `ctx.Next()`**, the chain stops and handlers never execute!

```go
func BrokenMiddleware(ctx *request.Context) error {
    log.Println("Processing...")
    return nil  // ❌ Chain stops here! Handler never runs!
}
```

---

**Key Takeaway**: Middlewares are the backbone of cross-cutting concerns in Lokstra. Master them for clean, maintainable APIs! 🎯
