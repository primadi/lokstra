# Middleware - Essential Guide

> **Request/response processing made easy**  
> **Time**: 25-30 minutes • **Level**: Beginner

---

## 📖 What You'll Learn

- ✅ What middleware is and why you need it
- ✅ Using middleware in 2 ways (direct & by name)
- ✅ Global vs per-route middleware
- ✅ Middleware registration and factories
- ✅ Built-in middleware (logging, CORS, auth)

---

## 🎯 What is Middleware?

**Middleware** is a function that runs **before** (or after) your handler. It's perfect for:
- 📝 **Logging** - Log every request
- 🔒 **Authentication** - Verify user identity
- 🛡️ **Authorization** - Check permissions
- 🌐 **CORS** - Handle cross-origin requests
- ⏱️ **Rate Limiting** - Prevent abuse
- 📊 **Metrics** - Collect statistics

**Request Flow with Middleware:**
```
HTTP Request
   ↓
Middleware 1 (logging)
   ↓
Middleware 2 (auth)
   ↓
Your Handler
   ↓
Response
```

---

## 🚀 Quick Start (2 Minutes)

```go
package main

import (
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/middleware/request_logger"
)

func main() {
    r := lokstra.NewRouter("api")
    
    // Add logging middleware
    r.Use(request_logger.Middleware(nil))
    
    r.GET("/users", func() []string {
        return []string{"Alice", "Bob"}
    })
    
    app := lokstra.NewApp("demo", ":3000", r)
    app.Run(30 * time.Second)
}
```

**Test it:**
```bash
curl http://localhost:3000/users
# Console will show: [LOG] GET /users 200 OK (2ms)
```

---

## 📝 Two Ways to Use Middleware

Lokstra supports **2 methods** for using middleware:

### Method 1: Direct Function Call
**Use when**: Simple apps, code-only configuration

```go
import (
    "github.com/primadi/lokstra/middleware/request_logger"
    "github.com/primadi/lokstra/middleware/cors"
)

r := lokstra.NewRouter("api")

// Use middleware directly
r.Use(
    request_logger.Middleware(nil),
    cors.Middleware(corsConfig),
)
```

**Pros**:
- ✅ Simple and direct
- ✅ Type-safe
- ✅ No registration needed

**Cons**:
- ❌ Hardcoded in code
- ❌ Can't configure via YAML

---

### Method 2: By Name (Registry Pattern)
**Use when**: Config-driven apps, reusable middleware

```go
import (
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/middleware/request_logger"
)

// Step 1: Register middleware factory (once, usually in main.go)
lokstra_registry.RegisterMiddlewareFactory("logger", 
    request_logger.MiddlewareFactory)

// Step 2: Register named instances with config
lokstra_registry.RegisterMiddlewareName("logger_std", "logger", map[string]any{
    "show_body": false,
    "show_headers": true,
})

lokstra_registry.RegisterMiddlewareName("logger_verbose", "logger", map[string]any{
    "show_body": true,
    "show_headers": true,
})

// Step 3: Use by name
r.Use("logger_std")  // Uses standard config

// Different router with verbose logging
adminRouter := lokstra.NewRouter("admin")
adminRouter.Use("logger_verbose")  // Uses verbose config
```

**Pros**:
- ✅ Reusable with different configs
- ✅ Can be configured via YAML
- ✅ Environment-specific configs

**Cons**:
- ❌ Requires registration setup
- ❌ String-based (not type-safe)

---

### When to Use Which?

| Scenario | Method | Reason |
|----------|--------|--------|
| Simple app, few middleware | Direct | Less code, simpler |
| Multiple environments | By Name | Different configs per env |
| YAML-driven config | By Name | Can configure externally |
| Prototype/learning | Direct | Faster to write |
| Production app | By Name | More flexible |

---

## 🔧 Middleware Scopes

### 1. Global Middleware
Applied to **all routes** in the router:

```go
r := lokstra.NewRouter("api")

// These run for EVERY route
r.Use(loggingMiddleware, corsMiddleware)

r.GET("/users", getUsers)      // Has logging + CORS
r.POST("/products", addProduct) // Has logging + CORS
```

---

### 2. Per-Route Middleware
Applied to **specific routes** only:

```go
r.GET("/public", publicHandler)  // No auth

// Method 1: Direct
r.GET("/private", privateHandler, authMiddleware)

// Method 2: By name
r.GET("/admin", adminHandler, "auth_jwt", "admin_check")
```

---

### 3. Group Middleware
Applied to **all routes in a group**:

```go
// API v1 - basic auth
v1 := r.AddGroup("/v1")
v1.Use(basicAuthMiddleware)
v1.GET("/users", getUsersV1)

// API v2 - JWT auth
v2 := r.AddGroup("/v2")
v2.Use(jwtAuthMiddleware)
v2.GET("/users", getUsersV2)
```

---

### 4. Mixed Middleware
Combine global, group, and route-level:

```go
// Global: All routes
r.Use(loggingMiddleware)

// Group: /admin/* routes
admin := r.AddGroup("/admin")
admin.Use(authMiddleware)

// Route-specific: Only this route
admin.GET("/danger", dangerHandler, confirmMiddleware)

// Final middleware chain for /admin/danger:
// 1. loggingMiddleware (global)
// 2. authMiddleware (group)
// 3. confirmMiddleware (route)
// 4. dangerHandler
```

---

## 🏭 Middleware Factory Pattern

### Understanding the Pattern

**Factory** = Function that creates middleware with config

```go
// Middleware Factory Type
type MiddlewareFactory = func(config map[string]any) request.HandlerFunc

// Example: Logger Factory
func LoggerFactory(config map[string]any) request.HandlerFunc {
    showBody := config["show_body"].(bool)
    showHeaders := config["show_headers"].(bool)
    
    // Return the actual middleware
    return func(ctx *request.Context) error {
        if showBody {
            // Log body
        }
        if showHeaders {
            // Log headers
        }
        return ctx.Next()
    }
}
```

### Step-by-Step Usage

```go
// 1. Register the factory (once)
lokstra_registry.RegisterMiddlewareFactory("logger", LoggerFactory)

// 2. Create named instances (different configs)
lokstra_registry.RegisterMiddlewareName("logger_dev", "logger", map[string]any{
    "show_body": true,
    "show_headers": true,
})

lokstra_registry.RegisterMiddlewareName("logger_prod", "logger", map[string]any{
    "show_body": false,
    "show_headers": false,
})

// 3. Use by name
devRouter := lokstra.NewRouter("dev-api")
devRouter.Use("logger_dev")  // Verbose logging

prodRouter := lokstra.NewRouter("prod-api")
prodRouter.Use("logger_prod")  // Minimal logging
```

---

## 🎨 Built-in Middleware

Lokstra provides several ready-to-use middleware:

### 1. Request Logger
```go
import "github.com/primadi/lokstra/middleware/request_logger"

// Direct usage
r.Use(request_logger.Middleware(nil))

// Or with config
r.Use(request_logger.Middleware(map[string]any{
    "show_body": true,
}))

// Or by name
lokstra_registry.RegisterMiddlewareFactory("logger", 
    request_logger.MiddlewareFactory)
lokstra_registry.RegisterMiddlewareName("logger_std", "logger", nil)
r.Use("logger_std")
```

**Output**:
```
[INFO] GET /users 200 OK (5ms)
[INFO] POST /users 201 Created (15ms)
```

---

### 2. CORS
```go
import "github.com/primadi/lokstra/middleware/cors"

corsConfig := map[string]any{
    "allow_origins": []string{"*"},
    "allow_methods": []string{"GET", "POST", "PUT", "DELETE"},
    "allow_headers": []string{"Content-Type", "Authorization"},
}

// Direct
r.Use(cors.Middleware(corsConfig))

// By name
lokstra_registry.RegisterMiddlewareFactory("cors", cors.MiddlewareFactory)
lokstra_registry.RegisterMiddlewareName("cors_all", "cors", corsConfig)
r.Use("cors_all")
```

---

### 3. JWT Authentication
```go
import "github.com/primadi/lokstra/middleware/jwtauth"

jwtConfig := map[string]any{
    "validator_service_name": "auth_validator",
}

// Per-route
r.GET("/public", publicHandler)
r.GET("/private", privateHandler, jwtauth.MiddlewareFactory(jwtConfig))

// Or by name
lokstra_registry.RegisterMiddlewareFactory("jwt", jwtauth.MiddlewareFactory)
lokstra_registry.RegisterMiddlewareName("jwt_std", "jwt", jwtConfig)
r.GET("/private", privateHandler, "jwt_std")
```

---

## 🧪 Examples

All examples are runnable!

### [01 - Logging Middleware](examples/01-logging/)
**Learn**: Request logging setup  
**Time**: 5 minutes

```go
r.Use(request_logger.Middleware(nil))
```

---

### [02 - Authentication](examples/02-authentication/)
**Learn**: JWT authentication, protected routes  
**Time**: 10 minutes

```go
r.GET("/public", handler)
r.GET("/private", handler, "jwt_auth")
```

---

### [03 - CORS Configuration](examples/03-cors/)
**Learn**: Cross-origin setup for APIs  
**Time**: 7 minutes

```go
r.Use(cors.Middleware(corsConfig))
```

---

### [04 - Multiple Middleware](examples/04-multiple/)
**Learn**: Combining middleware, execution order  
**Time**: 8 minutes

```go
r.Use("logger", "cors")
admin := r.AddGroup("/admin")
admin.Use("auth", "admin_check")
```

---

## 🎯 Common Patterns

### Pattern 1: Standard API Setup
```go
r := lokstra.NewRouter("api")

// Every API needs these
r.Use("logger", "cors")

// Public routes
r.GET("/health", healthCheck)

// Protected routes
auth := r.AddGroup("/api")
auth.Use("jwt_auth")
auth.GET("/users", getUsers)
```

---

### Pattern 2: Multi-Environment Config
```go
// Development - verbose
if env == "dev" {
    r.Use("logger_verbose", "cors_dev")
}

// Production - minimal
if env == "prod" {
    r.Use("logger_minimal", "cors_prod")
}
```

---

### Pattern 3: Progressive Authentication
```go
// Public
r.GET("/products", listProducts)

// Basic auth
registered := r.AddGroup("/registered")
registered.Use("basic_auth")
registered.GET("/profile", getProfile)

// Admin
admin := r.AddGroup("/admin")
admin.Use("jwt_auth", "admin_check")
admin.DELETE("/users", deleteUser)
```

---

## 🚫 Common Mistakes

### ❌ Don't: Add same middleware twice
```go
r.Use(loggingMiddleware)
r.Use(loggingMiddleware)  // ❌ Duplicate!
```

**✅ Do**: Add once at appropriate level
```go
r.Use(loggingMiddleware)  // Global - runs once
```

---

### ❌ Don't: Forget to call Next()
```go
func badMiddleware(ctx *request.Context) error {
    // Do something
    return nil  // ❌ Didn't call ctx.Next()!
}
```

**✅ Do**: Always call Next() unless stopping
```go
func goodMiddleware(ctx *request.Context) error {
    // Do something before handler
    err := ctx.Next()  // ✅ Continue chain
    // Do something after handler
    return err
}
```

---

### ❌ Don't: Mix named and unnamed inconsistently
```go
r.Use("logger")
r.Use(corsMiddleware)  // ❌ Mixing styles in same router
```

**✅ Do**: Pick one style per router
```go
// All by name
r.Use("logger", "cors")

// Or all direct
r.Use(loggingMiddleware, corsMiddleware)
```

---

## 🎓 Best Practices

### 1. **Register Factories Early**
```go
// ✅ Good - in main() or init()
func main() {
    registerMiddleware()  // All registrations
    setupRouters()        // Use middleware
    startServer()
}
```

---

### 2. **Use Descriptive Names**
```go
// ✅ Good
lokstra_registry.RegisterMiddlewareName("logger_verbose", "logger", config)
lokstra_registry.RegisterMiddlewareName("auth_jwt_admin", "jwt", config)

// 🚫 Bad
lokstra_registry.RegisterMiddlewareName("mw1", "logger", config)
lokstra_registry.RegisterMiddlewareName("a", "jwt", config)
```

---

### 3. **Order Matters**
```go
// ✅ Good - logical order
r.Use(
    "logger",     // 1. Log first
    "cors",       // 2. Handle CORS
    "rate_limit", // 3. Rate limit
    "auth",       // 4. Authenticate
)

// 🚫 Bad - auth before rate limit?
r.Use("auth", "rate_limit")  // Attacker can spam auth!
```

---

## 📚 What's Next?

You now understand:
- ✅ What middleware is and why it's useful
- ✅ Two ways to use middleware (direct & by name)
- ✅ Global, per-route, and group middleware
- ✅ Middleware factory pattern
- ✅ Built-in middleware (logging, CORS, auth)

### Next Steps:

**Immediate**: 
👉 [04 - Configuration](../04-configuration/README.md) - Learn config patterns

**Related**:
- [Deep Dive: Custom Middleware](../../02-deep-dive/middleware/custom-middleware.md)
- [Guide: Authentication Strategies](../../04-guides/authentication.md)

---

## 🔍 Quick Reference

### Direct Usage
```go
r.Use(middleware1, middleware2)
r.GET("/path", handler, middleware3)
```

### By Name Usage
```go
// Register
lokstra_registry.RegisterMiddlewareFactory(type, factory)
lokstra_registry.RegisterMiddlewareName(name, type, config)

// Use
r.Use("name1", "name2")
r.GET("/path", handler, "name3")
```

### Built-in Middleware
- `request_logger.Middleware(config)` - Request logging
- `cors.Middleware(config)` - CORS handling
- `jwtauth.MiddlewareFactory(config)` - JWT auth
- `accesscontrol.MiddlewareFactory(config)` - Access control

---

**Continue learning** → [04 - Configuration](../04-configuration/README.md)
