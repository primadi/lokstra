---
layout: docs
title: Router - Essential Guide
---

# Router - Essential Guide

> **HTTP routing made flexible and intuitive**  
> **Time**: 45 minutes (with examples) • **Level**: Beginner

---

## 📖 What You'll Learn

- ✅ Create routers and register routes
- ✅ Write handlers in 5 essential forms (out of 29 total!)
- ✅ Handle path parameters and query strings
- ✅ Organize routes with groups
- ✅ Master 3 response patterns (Manual, Generic, Opinionated)
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

**Total learning time**: ~45 minutes

### [01 - Basic Routes](examples/01-basic-routes/) ⏱️ 5 min
**Learn**: Router basics, GET/POST, auto JSON conversion

```go
r.GET("/ping", func() string { return "pong" })
r.GET("/users", func() []User { return users })
r.POST("/users", func(req *CreateUserRequest) (*User, error) { ... })
```

**Key Concepts**: Router creation, HTTP methods, automatic JSON, request binding, validation

---

### [02 - Route Parameters](examples/02-route-parameters/) ⏱️ 7 min
**Learn**: Path params, query params, type conversion

```go
// Path parameter
r.GET("/users/{id}", func(req *GetUserRequest) (*User, error) {
    // req.ID extracted from path
})

// Query parameters
r.GET("/products", func(req *SearchRequest) ([]Product, error) {
    // req.Category, req.MinPrice from ?category=x&min_price=y
})
```

**Key Concepts**: `path:"id"`, `query:"name"`, default values, automatic type conversion

---

### [03 - Route Groups](examples/03-route-groups/) ⏱️ 7 min
**Learn**: API versioning, route organization, nested groups

```go
// API v1
v1 := r.AddGroup("/v1")
v1.GET("/users", getUsersV1)

// API v2
v2 := r.AddGroup("/v2")
v2.GET("/users", getUsersV2)  // Enhanced version

// Nested groups
admin := r.AddGroup("/admin")
adminUsers := admin.AddGroup("/users")
```

**Key Concepts**: Route groups, prefixes, API versioning, `PrintRoutes()` debugging

---

### [04 - Handler Forms](examples/04-handler-forms/) ⏱️ 10 min
**Learn**: 5 essential handler patterns, when to use each

```go
// Form 1: Simple return
func() string { return "pong" }

// Form 2: With error (most common!)
func() ([]User, error) { return users, nil }

// Form 3: Request binding
func(req *CreateUserRequest) (*User, error) { ... }

// Form 4: Full control with context
func(ctx *request.Context, req *GetUserRequest) (*User, error) { ... }

// Form 5: Custom response
func(ctx *request.Context) (*response.Response, error) { ... }
```

**Key Concepts**: Handler signatures, flexibility, context access, decision guide

---

### [05 - Response Patterns](examples/05-response-patterns/) ⏱️ 15 min ⭐
**Learn**: 3 response types, 2 response paths, when to use each

**3 Response Types:**
```go
// 1. Manual (http.ResponseWriter)
func(ctx *request.Context) error {
    ctx.W.Write([]byte(`{"message":"hello"}`))
}

// 2. Generic (response.Response) - JSON, HTML, text
func() (*response.Response, error) {
    resp := response.NewResponse()
    resp.Json(data)  // or .Html(), .Text()
    return resp, nil
}

// 3. Opinionated (response.ApiHelper) - Structured JSON
func() (*response.ApiHelper, error) {
    api := response.NewApiHelper()
    api.Ok(data)  // Standard format
    return api, nil
}
```

**Key Concepts**: 
- Manual vs Generic vs Opinionated responses
- When to use each type
- ApiHelper standard JSON format (success, error, validation)
- PagingRequest for list endpoints
- Decision guide for REST APIs

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

### 4. **Use ApiHelper for REST APIs**
```go
// ✅ Good - Consistent JSON structure
r.GET("/users", func() (*response.ApiHelper, error) {
    api := response.NewApiHelper()
    users, err := db.GetUsers()
    if err != nil {
        api.InternalError("Database error")
        return api, nil
    }
    api.Ok(users)
    return api, nil
})

// 🚫 Inconsistent - no standard format
r.GET("/users", func() (map[string]any, error) {
    users, err := db.GetUsers()
    return map[string]any{"data": users}, err
})
```

---

### 5. **Use PagingRequest for Lists**
```go
// ✅ Good - Standard pagination
type ListUsersRequest struct {
    request.PagingRequest  // Embeds page, page_size, order_by, etc
    Status string `query:"status"`
}

r.GET("/users", func(req *ListUsersRequest) (*response.ApiHelper, error) {
    req.SetDefaults()  // Apply default page, page_size
    users, total := db.GetUsers(req.GetOffset(), req.GetLimit())
    
    api := response.NewApiHelper()
    api.OkList(users, &api_formatter.ListMeta{
        Page:      req.Page,
        PageSize:  req.PageSize,
        TotalRows: total,
    })
    return api, nil
})

// 🚫 Bad - Reinventing pagination
type CustomPaging struct {
    P  int `query:"p"`
    Sz int `query:"sz"`
}
```

---

## 📚 What's Next?

You now understand:
- ✅ Creating routers and registering routes
- ✅ 5 essential handler forms (simple, error, binding, context, response)
- ✅ Path and query parameters with automatic binding
- ✅ Route groups for organization
- ✅ 3 response patterns (Manual, Generic, Opinionated)
- ✅ PagingRequest for list endpoints
- ✅ Basic middleware usage

### Next Steps:

**Continue Learning (Recommended order)**:  
1. 👉 **[02 - Service](../02-service/README.md)** - Service patterns and dependency injection ⭐ CRITICAL
2. 👉 **[03 - Middleware](../03-middleware/README.md)** - Deep dive into middleware patterns
3. 👉 **[04 - Configuration](../04-configuration/README.md)** - Config-driven development
4. 👉 **[05 - App and Server](../05-app-and-server/README.md)** - Application lifecycle
5. 👉 **[06 - Complete API](../06-putting-it-together/README.md)** - Build a real TODO API

**Deep Dive Topics**:
- [All 29 Handler Forms](../../02-deep-dive/router/handler-forms.md) (coming soon)
- [Router Lifecycle](../../02-deep-dive/router/lifecycle.md) (coming soon)
- [Advanced Routing](../../02-deep-dive/router/advanced.md) (coming soon)

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
