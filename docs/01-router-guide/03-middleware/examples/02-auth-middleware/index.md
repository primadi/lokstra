# Authentication Middleware Example

Demonstrates authentication and authorization middleware patterns.

## What You'll Learn

- Create custom authentication middleware
- Create custom authorization middleware
- Per-route middleware (protect specific routes)
- Route group middleware (protect multiple routes)
- Middleware chain execution order

## Running

```bash
cd docs/01-router-guide/03-middleware/examples/02-auth-middleware
go run main.go
```

## Testing

Use `test.http` or curl:

**Public endpoint (no auth):**
```bash
curl http://localhost:3000/public
```

**Protected endpoint (requires API key):**
```bash
# Without key - should fail
curl http://localhost:3000/protected

# With valid key - should work
curl -H "X-API-Key: secret-key-123" http://localhost:3000/protected

# With invalid key - should fail
curl -H "X-API-Key: wrong-key" http://localhost:3000/protected
```

**Admin endpoint (requires API key + admin role):**
```bash
# Without admin role - should fail
curl -H "X-API-Key: secret-key-123" http://localhost:3000/admin

# With admin role - should work
curl -H "X-API-Key: secret-key-123" -H "X-User-Role: admin" http://localhost:3000/admin
```

## How It Works

### 1. Auth Middleware
```go
func authMiddleware(ctx *request.Context) error {
    apiKey := ctx.GetHeader("X-API-Key")
    
    if apiKey == "" {
        return ctx.Api.Unauthorized("API key required")
    }
    
    if apiKey != "secret-key-123" {
        return ctx.Api.Forbidden("Invalid API key")
    }
    
    // Continue to next middleware or handler
    return ctx.Next()
}
```

### 2. Admin Middleware
```go
func adminMiddleware(ctx *request.Context) error {
    role := ctx.GetHeader("X-User-Role")
    
    if role != "admin" {
        return ctx.Api.Forbidden("Admin access required")
    }
    
    return ctx.Next()
}
```

### 3. Per-Route Middleware
```go
// Single route with middleware
router.GET("/protected", handler, authMiddleware)

// Multiple middleware (executed in order)
router.GET("/admin", handler, authMiddleware, adminMiddleware)
```

### 4. Group Middleware
```go
// All routes in group use middleware
apiGroup := router.AddGroup("/api")
apiGroup.Use(authMiddleware)

apiGroup.GET("/users", handler)    // Protected
apiGroup.GET("/orders", handler)   // Protected

// Nested group with additional middleware
adminGroup := router.AddGroup("/api/admin")
adminGroup.Use(authMiddleware, adminMiddleware)

adminGroup.GET("/stats", handler)  // Protected + Admin only
adminGroup.GET("/logs", handler)   // Protected + Admin only
```

**Alternative syntax with callback:**
```go
router.Group("/api", func(g lokstra.Router) {
    g.Use(authMiddleware)
    g.GET("/users", handler)
    g.GET("/orders", handler)
})
```

## Middleware Execution Flow

```
Request: GET /admin

1. authMiddleware runs
   ├─ Checks X-API-Key
   └─ Calls ctx.Next()

2. adminMiddleware runs
   ├─ Checks X-User-Role
   └─ Calls ctx.Next()

3. Handler runs
   └─ Returns response

Response sent back to client
```

## Expected Responses

**No API key:**
```json
{
  "status": "error",
  "error": {
    "code": "UNAUTHORIZED",
    "message": "API key required"
  }
}
```

**Invalid API key:**
```json
{
  "status": "error",
  "error": {
    "code": "FORBIDDEN",
    "message": "Invalid API key"
  }
}
```

**No admin role:**
```json
{
  "status": "error",
  "error": {
    "code": "FORBIDDEN",
    "message": "Admin access required"
  }
}
```

**Success:**
```json
{
  "status": "success",
  "data": {
    "message": "This is a protected endpoint",
    "access": "authenticated users only"
  }
}
```

## Key Takeaways

- ✅ Middleware can check headers, query params, body, etc.
- ✅ Return early with error response to stop execution
- ✅ Call `ctx.Next()` to continue to next middleware/handler
- ✅ Multiple middleware execute in order (auth → admin → handler)
- ✅ Group middleware applies to all routes in group
- ✅ Per-route middleware only applies to that specific route

## Production Tips

In real applications:
- Use JWT tokens instead of simple API keys
- Store user info in context: `ctx.Set("user", user)`
- Use database/cache to validate tokens
- Implement rate limiting
- Log failed auth attempts
- Use HTTPS in production
