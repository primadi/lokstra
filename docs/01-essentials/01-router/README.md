# Router - Essential Guide

> **HTTP routing made flexible and intuitive**  
> **Time**: 30-40 minutes • **Level**: Beginner

---

## 📖 What You'll Learn

- ✅ Create routers and register routes
- ✅ Write handlers in 4 essential forms (out of 29 total!)
- ✅ Handle path parameters and query strings
- ✅ Organize routes with groups
- ✅ Apply middleware to routes

---

## 🎯 What is a Router?

A **Router** is Lokstra's HTTP request matcher. It:
- Matches incoming requests to handlers
- Extracts path parameters
- Applies middleware
- Invokes your handler function

**Key Insight**: Router implements `http.Handler`, so you can use it directly:
```go
r := lokstra.NewRouter("api")
r.GET("/ping", func() string { return "pong" })

// Use directly with Go's http package!
http.ListenAndServe(":8080", r)
```

---

## 🚀 Quick Start (2 Minutes)

```go
package main

import (
    "github.com/primadi/lokstra"
    "time"
)

func main() {
    // 1. Create router
    r := lokstra.NewRouter("api")
    
    // 2. Register routes
    r.GET("/ping", func() string {
        return "pong"
    })
    
    r.GET("/users", func() []string {
        return []string{"Alice", "Bob"}
    })
    
    // 3. Create app and run
    app := lokstra.NewApp("demo", ":3000", r)
    app.Run(30 * time.Second)
}
```

**Test it:**
```bash
curl http://localhost:3000/ping   # → "pong"
curl http://localhost:3000/users  # → ["Alice","Bob"]
```

---

## 📝 Basic Concepts

### 1. Creating a Router

```go
// Simple router
r := lokstra.NewRouter("my-api")

// Router with specific engine (advanced)
r := lokstra.NewRouterWithEngine("my-api", "httprouter")
```

**💭 Tip**: Use descriptive names. They appear in logs and debugging output.

---

### 2. HTTP Methods

Lokstra supports all standard HTTP methods:

```go
r.GET("/users", getUsersHandler)
r.POST("/users", createUserHandler)
r.PUT("/users/{id}", updateUserHandler)
r.PATCH("/users/{id}", patchUserHandler)
r.DELETE("/users/{id}", deleteUserHandler)
```

**Special method**:
```go
// ANY matches all HTTP methods
r.ANY("/webhook", webhookHandler)
```

---

### 3. Path Parameters

Extract dynamic values from URLs:

```go
type UserRequest struct {
    ID string `path:"id"`  // Auto-extracted from path
}

r.GET("/users/{id}", func(req *UserRequest) (string, error) {
    return "User ID: " + req.ID, nil
})
```

**Test it:**
```bash
curl http://localhost:3000/users/123
# → "User ID: 123"
```

---

### 4. Query Parameters

Extract from query string:

```go
type SearchRequest struct {
    Query string `query:"q"`
    Page  int    `query:"page"`
}

r.GET("/search", func(req *SearchRequest) (string, error) {
    return fmt.Sprintf("Searching for: %s (page %d)", req.Query, req.Page), nil
})
```

**Test it:**
```bash
curl "http://localhost:3000/search?q=lokstra&page=2"
# → "Searching for: lokstra (page 2)"
```

---

## 🎨 Handler Forms (The Essential 4)

Lokstra supports **29 handler forms**, but you'll use these 4 most often:

### Form 1: Simple Return Value
**Use when**: Simple data, no errors

```go
r.GET("/ping", func() string {
    return "pong"
})

r.GET("/status", func() map[string]string {
    return map[string]string{"status": "ok"}
})
```

---

### Form 2: Return with Error
**Use when**: Operations that can fail (most common!)

```go
r.GET("/users", func() ([]User, error) {
    users, err := db.GetUsers()
    if err != nil {
        return nil, err  // Lokstra handles error response
    }
    return users, nil
})
```

---

### Form 3: Request Binding with Error
**Use when**: Need request data (POST/PUT)

```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

r.POST("/users", func(req *CreateUserRequest) (User, error) {
    // req is auto-bound from JSON body
    user, err := db.CreateUser(req.Name, req.Email)
    return user, err
})
```

---

### Form 4: Context + Request with Error
**Use when**: Need full control (headers, status codes, etc)

```go
r.GET("/users/{id}", func(ctx *request.Context, req *UserRequest) (*User, error) {
    // Access request context
    authHeader := ctx.R.Header.Get("Authorization")
    
    // req.ID auto-extracted from path
    user, err := db.GetUser(req.ID)
    return user, err
})
```

---

**💭 Which form to use?**
- 90% of the time: **Form 2** (return with error)
- POST/PUT endpoints: **Form 3** (request binding)
- Need headers/cookies: **Form 4** (with context)
- Ultra-simple: **Form 1** (no errors possible)

**📖 Want all 29 forms?** See [Deep Dive: Handler Forms](../../02-deep-dive/router/handler-forms.md)

---

## 🗂️ Route Groups

Organize routes with shared prefixes:

### Method 1: Inline Groups
```go
r := lokstra.NewRouter("api")

// API v1
r.Group("/v1", func(v1 Router) {
    v1.GET("/users", getUsersV1)
    v1.GET("/products", getProductsV1)
})

// API v2
r.Group("/v2", func(v2 Router) {
    v2.GET("/users", getUsersV2)
    v2.GET("/products", getProductsV2)
})
```

**Result:**
```
GET /v1/users
GET /v1/products
GET /v2/users
GET /v2/products
```

---

### Method 2: Stored Groups
```go
v1 := r.AddGroup("/v1")
v1.GET("/users", getUsersV1)
v1.GET("/products", getProductsV1)

v2 := r.AddGroup("/v2")
v2.GET("/users", getUsersV2)
v2.GET("/products", getProductsV2)
```

**💭 Tip**: Use inline for simple cases, stored for complex routing logic.

---

## 🛡️ Middleware Basics

Lokstra supports **2 ways** to use middleware:

### Method 1: Direct Middleware Function
```go
r := lokstra.NewRouter("api")

// Use middleware functions directly
r.Use(logging.Middleware(), auth.Middleware())

r.GET("/users", getUsersHandler)
r.POST("/users", createUserHandler)
// Both routes get logging + auth
```

---

### Method 2: By Name (Registry-Based)
```go
// First, register middleware factories (usually in main.go or setup)
lokstra_registry.RegisterMiddlewareFactory("logger", loggerFactory)
lokstra_registry.RegisterMiddlewareFactory("auth", authFactory)

// Then register named instances with config
lokstra_registry.RegisterMiddlewareName("logger_std", "logger", loggerStdConfig)
lokstra_registry.RegisterMiddlewareName("auth_jwt", "auth", jwtConfig)

// Use by name in router
r.Use("logger_std", "auth_jwt")

r.GET("/users", getUsersHandler)
```

**💭 When to use which?**
- **Method 1**: Simple apps, few middleware, code-only setup
- **Method 2**: Config-driven apps, reusable middleware with different configs

**Example - Multiple auth configurations**:
```go
// Register factory once
lokstra_registry.RegisterMiddlewareFactory("auth", authFactory)

// Create named instances with different configs
lokstra_registry.RegisterMiddlewareName("auth_basic", "auth", basicConfig)
lokstra_registry.RegisterMiddlewareName("auth_jwt", "auth", jwtConfig)
lokstra_registry.RegisterMiddlewareName("auth_oauth", "auth", oauthConfig)

// Use different auth per router
publicAPI := lokstra.NewRouter("public")
publicAPI.Use("auth_basic")

adminAPI := lokstra.NewRouter("admin")
adminAPI.Use("auth_jwt")
```

---

### Global Middleware
```go
r := lokstra.NewRouter("api")

// Applied to ALL routes
r.Use(loggingMiddleware, corsMiddleware)
// Or by name:
r.Use("logger_std", "cors_default")

r.GET("/users", getUsersHandler)
r.POST("/users", createUserHandler)
// Both routes get logging + CORS
```

---

### Per-Route Middleware
```go
r.GET("/public", publicHandler)  // No auth

// Method 1: Direct function
r.GET("/private", privateHandler, authMiddleware)

// Method 2: By name
r.GET("/private", privateHandler, "auth_jwt")
```

---

### Group Middleware
```go
admin := r.AddGroup("/admin")
admin.Use(authMiddleware, adminMiddleware)
// Or by name:
admin.Use("auth_jwt", "admin_check")

admin.GET("/users", getAllUsers)      // Requires auth + admin
admin.DELETE("/users", deleteUser)    // Requires auth + admin

r.GET("/public", publicEndpoint)      // No middleware
```

**📖 More on middleware**: See [03 - Middleware](../03-middleware/README.md)

---

## 🧪 Examples

All examples are runnable! Navigate to each folder and `go run main.go`

### [01 - Basic Routes](examples/01-basic-routes/)
**Learn**: GET, POST, simple handlers  
**Time**: 5 minutes

```go
r.GET("/ping", func() string { return "pong" })
r.GET("/users", func() []User { return users })
r.POST("/users", func(user *User) error { ... })
```

---

### [02 - Route Parameters](examples/02-route-parameters/)
**Learn**: Path params, query params, request binding  
**Time**: 7 minutes

```go
r.GET("/users/{id}", getUserHandler)
r.GET("/search", searchHandler)  // ?q=term&page=1
```

---

### [03 - Route Groups](examples/03-route-groups/)
**Learn**: API versioning, grouped routes  
**Time**: 7 minutes

```go
r.Group("/v1", func(v1 Router) { ... })
r.Group("/v2", func(v2 Router) { ... })
```

---

### [04 - With Middleware](examples/04-with-middleware/)
**Learn**: Global, per-route, and group middleware  
**Time**: 10 minutes

```go
r.Use(loggingMiddleware)
r.GET("/private", handler, authMiddleware)
```

---

### [05 - Complete API](examples/05-complete-api/)
**Learn**: Full REST API with all concepts  
**Time**: 15 minutes

Complete user management API:
- CRUD operations
- Path & query parameters
- Request validation
- Error handling
- Middleware

---

## 🎯 Common Patterns

### Pattern 1: RESTful Resources
```go
r.GET("/users", listUsers)           // List
r.POST("/users", createUser)         // Create
r.GET("/users/{id}", getUser)        // Read
r.PUT("/users/{id}", updateUser)     // Update
r.DELETE("/users/{id}", deleteUser)  // Delete
```

---

### Pattern 2: Nested Resources
```go
r.GET("/users/{userId}/posts", getUserPosts)
r.POST("/users/{userId}/posts", createUserPost)
r.GET("/users/{userId}/posts/{postId}", getUserPost)
```

---

### Pattern 3: API Versioning
```go
// Option 1: Path-based
r.Group("/v1", func(v1 Router) { ... })
r.Group("/v2", func(v2 Router) { ... })

// Option 2: Subdomain (via multiple routers)
apiV1 := lokstra.NewRouter("api-v1")
apiV2 := lokstra.NewRouter("api-v2")
```

---

## 🚫 Common Mistakes

### ❌ Don't: Register after Build()
```go
r := lokstra.NewRouter("api")
r.GET("/first", handler1)
app := lokstra.NewApp("demo", ":8080", r)
app.Start()  // Router builds here

r.GET("/second", handler2)  // ❌ PANIC! Can't register after build
```

**✅ Do**: Register all routes before starting
```go
r.GET("/first", handler1)
r.GET("/second", handler2)
app.Start()  // Now it's safe
```

---

### ❌ Don't: Ignore errors in handlers
```go
r.GET("/users", func() []User {
    users, _ := db.GetUsers()  // ❌ Ignoring error!
    return users
})
```

**✅ Do**: Return errors
```go
r.GET("/users", func() ([]User, error) {
    return db.GetUsers()  // ✅ Error handled by Lokstra
})
```

---

## 🎓 Best Practices

### 1. **Use Meaningful Names**
```go
// ✅ Good
r := lokstra.NewRouter("user-api")
r := lokstra.NewRouter("admin-api")

// 🚫 Bad
r := lokstra.NewRouter("r1")
r := lokstra.NewRouter("temp")
```

---

### 2. **Group Related Routes**
```go
// ✅ Good
users := r.AddGroup("/users")
users.GET("", listUsers)
users.POST("", createUser)
users.GET("/{id}", getUser)

// 🚫 Bad - harder to see relationships
r.GET("/users", listUsers)
r.POST("/users", createUser)
r.GET("/users/{id}", getUser)
```

---

### 3. **Choose Right Handler Form**
```go
// ✅ Good - simple case
r.GET("/ping", func() string { return "pong" })

// ✅ Good - can fail
r.GET("/users", func() ([]User, error) { return db.GetUsers() })

// 🚫 Overkill - don't need context
r.GET("/ping", func(ctx *request.Context) (string, error) {
    return "pong", nil
})
```

---

## 📚 What's Next?

You now understand:
- ✅ Creating routers
- ✅ Registering routes with multiple handler forms
- ✅ Path and query parameters
- ✅ Route groups
- ✅ Basic middleware usage

### Next Steps:

**Immediate**: 
👉 [02 - Service](../02-service/README.md) - Learn service patterns

**Related**:
- [03 - Middleware](../03-middleware/README.md) - Deep dive middleware
- [Deep Dive: All 29 Handler Forms](../../02-deep-dive/router/handler-forms.md)
- [Deep Dive: Router Lifecycle](../../02-deep-dive/router/lifecycle.md)

---

## 🔍 Quick Reference

### Common Methods
```go
// Create
r := lokstra.NewRouter("name")

// HTTP Methods
r.GET(path, handler, middleware...)
r.POST(path, handler, middleware...)
r.PUT(path, handler, middleware...)
r.PATCH(path, handler, middleware...)
r.DELETE(path, handler, middleware...)
r.ANY(path, handler, middleware...)

// Groups
r.Group(prefix, func(Router))
g := r.AddGroup(prefix)

// Middleware
r.Use(middleware...)

// Debugging
r.PrintRoutes()  // Print all registered routes
```

---

**Continue learning** → [02 - Service](../02-service/README.md)
