# Routing

Lokstra's routing system provides a flexible and powerful way to define HTTP endpoints, organize your API structure, and manage request handling. The router supports multiple engine backends and offers comprehensive features for both simple and complex applications.

## Creating a Router

The router is the core component that manages HTTP endpoints in your application:

```go
import "github.com/primadi/lokstra/core/router"

// Use the default engine (httprouter)
r := router.NewRouter(regCtx, map[string]any{/* engine config */})

// Or choose a specific engine
r2 := router.NewRouterWithEngine(regCtx, "servemux", nil)     // Go's standard servemux
r3 := router.NewRouterWithEngine(regCtx, "httprouter", nil)   // fast httprouter (default)
```

The default engine is `httprouter`, which provides excellent performance and path parameter support.

## Basic Route Registration

### HTTP Methods

Register routes using standard HTTP method helpers:

```go
// Basic routes
r.GET("/users", getUsersHandler)
r.POST("/users", createUserHandler)
r.PUT("/users/{id}", updateUserHandler)
r.PATCH("/users/{id}", patchUserHandler)
r.DELETE("/users/{id}", deleteUserHandler)

// Generic handler for any method
r.Handle(request.POST, "/custom", customHandler)
```

### Path Parameters

Use curly braces for path parameters:

```go
// Simple parameter
r.GET("/users/{id}", getUserHandler)

// Multiple parameters
r.GET("/users/{userId}/posts/{postId}", getUserPostHandler)

// Wildcard (catch-all)
r.GET("/files/*filepath", serveFileHandler)
```

Access parameters in your handler:

```go
func getUserHandler(ctx *lokstra.Context) error {
    userID := ctx.GetPathParam("id")
    
    // Business logic...
    
    return ctx.Ok(user)
}
```

## Handler Types

Lokstra supports three types of handlers for maximum flexibility:

### 1. Standard Handler Function

```go
func getUserHandler(ctx *lokstra.Context) error {
    userID := ctx.GetPathParam("id")
    
    user, err := userService.GetByID(userID)
    if err != nil {
        return ctx.ErrorNotFound("User not found")
    }
    
    return ctx.Ok(user)
}

r.GET("/users/{id}", getUserHandler)
```

### 2. Named Handler

Register handlers by name in the registration context:

```go
// Register the handler
regCtx.RegisterHandler("getUser", getUserHandler)

// Use by name in routes
r.GET("/users/{id}", "getUser")
```

### 3. Generic Handler with Automatic Binding

Define handlers with typed parameters for automatic request binding:

```go
type CreateUserRequest struct {
    Name  string `body:"name"`
    Email string `body:"email"`
}

func createUserWithParams(ctx *lokstra.Context, params *CreateUserRequest) error {
    // Parameters are automatically bound from request
    // No need to call ctx.BindBody manually
    
    user, err := userService.Create(params.Name, params.Email)
    if err != nil {
        return err // Will be 500 error
    }
    
    return ctx.OkCreated(user)
}

r.POST("/users", createUserWithParams)
```

## Route Groups

Organize related routes using groups with shared prefixes and middleware:

```go
// API v1 group
api := r.Group("/api/v1")
api.GET("/users", getUsersHandler)
api.POST("/users", createUserHandler)

// Admin group with authentication
admin := r.Group("/admin", "auth")
admin.GET("/stats", getStatsHandler)
admin.DELETE("/users/{id}", deleteUserHandler)

// Nested groups
v1 := r.Group("/api/v1")
users := v1.Group("/users")
users.GET("/", getUsersHandler)
users.GET("/{id}", getUserHandler)
users.POST("/", createUserHandler)
```

### Group Block Pattern

Use block syntax for cleaner group organization:

```go
r.GroupBlock("/api/v1", func(api Router) {
    api.GET("/users", getUsersHandler)
    api.POST("/users", createUserHandler)
    
    api.GroupBlock("/admin", func(admin Router) {
        admin.Use("auth") // Add authentication
        admin.GET("/stats", getStatsHandler)
        admin.DELETE("/users/{id}", deleteUserHandler)
    })
})
```

## Middleware

### Route-Level Middleware

Apply middleware to specific routes:

```go
// Single middleware
r.GET("/profile", getProfileHandler, "auth")

// Multiple middleware
r.POST("/admin/users", createUserHandler, "auth", "admin-only", "audit")
```

### Group-Level Middleware

Apply middleware to entire groups:

```go
// All routes in this group will use auth middleware
api := r.Group("/api", "auth")
api.GET("/profile", getProfileHandler)
api.POST("/settings", updateSettingsHandler)
```

### Router-Level Middleware

Apply middleware to all routes:

```go
// Apply to all routes on this router
r.Use("cors")
r.Use("request-logger")

// Then define routes
r.GET("/public", publicHandler)    // Will have cors + request-logger
r.GET("/private", privateHandler, "auth") // Will have cors + request-logger + auth
```

### Middleware Override

Sometimes you need to bypass inherited middleware:

```go
// Router with global middleware
r.Use("cors")
r.Use("auth")

// Group that overrides parent middleware
public := r.WithOverrideMiddleware(true).Group("/public")
public.Use("cors") // Only cors, no auth
public.GET("/health", healthHandler)

// Route that overrides all middleware
r.HandleOverrideMiddleware(request.GET, "/metrics", metricsHandler, "basic-auth")
```

## Static File Serving

### Basic Static Files

Serve static assets from filesystem or embedded files:

```go
import "embed"

//go:embed static/*
var staticFiles embed.FS

// Mount static files
r.MountStatic("/static/", false, staticFiles)

// Or from directory
sf := static_files.New().WithSourceDir("./public")
r.MountStatic("/assets/", false, sf.Sources...)
```

### Single Page Application (SPA)

Enable SPA mode for client-side routing:

```go
// spa=true serves index.html for unknown paths
r.MountStatic("/", true, staticFiles)
```

### Multiple Sources with Priority

Serve from multiple sources with fallback priority:

```go
sf := static_files.New().
    WithSourceDir("./web_override").  // Check here first
    WithEmbedFS(assets, "web")        // Fallback to embedded

r.MountStatic("/static/", false, sf.Sources...)
```

## HTMX Integration

Lokstra provides first-class HTMX support with automatic layout management:

```go
// Mount HTMX pages with script injection
inj := static_files.NewScriptInjection().
    AddNamedScriptInjection("default")

sf := static_files.New().
    WithSourceDir("./htmx").
    WithEmbedFS(htmxFS, "templates")

r.MountHtmx("/", inj, sf.Sources...)
```

### Page Data Routes

Create page data endpoints for HTMX pages:

```go
// Page data for /about page
r.GET("/page-data/about", func(ctx *lokstra.Context) error {
    return ctx.HtmxPageData(
        "About Us",                    // Page title
        "Learn about our company",     // Description
        map[string]any{               // Page data
            "team": []string{"Alice", "Bob", "Charlie"},
            "founded": 2020,
        },
    )
})
```

## Advanced Routing

### Raw HTTP Handlers

Integrate standard Go HTTP handlers:

```go
import "net/http"

// Prometheus metrics endpoint
r.RawHandle("/metrics", false, promhttp.Handler())

// File server with prefix stripping
r.RawHandle("/downloads/", true, http.FileServer(http.Dir("./files")))
```

### Reverse Proxy

Proxy requests to external services:

```go
// Simple proxy
r.MountReverseProxy("/api/external/", "http://localhost:8081", false)

// Proxy with middleware
r.MountReverseProxy("/api/auth/", "https://auth.example.com", false, "rate-limit")

// Override middleware (only use specified middleware)
r.MountReverseProxy("/api/legacy/", "http://legacy.internal", true, "cors")
```

### RPC Service Mount

Mount RPC services as HTTP endpoints:

```go
// Mount RPC service
r.MountRpcService("/rpc/user", "user.rpc", false, "auth")

// This creates POST /rpc/user/:method endpoints
// Example: POST /rpc/user/GetProfile
```

## Router Merging

Combine multiple routers for modular applications:

```go
// Create separate routers for different modules
userRouter := router.NewRouter(regCtx, nil)
userRouter.GET("/users", getUsersHandler)

adminRouter := router.NewRouter(regCtx, nil)
adminRouter.Use("admin-auth")
adminRouter.GET("/admin/stats", getStatsHandler)

// Merge into main router
mainRouter := router.NewRouter(regCtx, nil)
mainRouter.AddRouter(userRouter)
mainRouter.AddRouter(adminRouter)
```

## Route Inspection

Debug and analyze your routes:

```go
// Print all routes in a formatted table
r.DumpRoutes()

// Programmatically inspect routes
r.RecurseAllHandler(func(route *RouteMeta) {
    fmt.Printf("Route: %s %s -> %s\n", 
        route.Method, route.Path, route.HandlerName)
})
```

## Best Practices

### 1. Organize with Groups

Structure your routes logically using groups:

```go
api := r.Group("/api/v1")

// User management
users := api.Group("/users")
users.GET("/", getUsersHandler)
users.POST("/", createUserHandler)
users.GET("/{id}", getUserHandler)
users.PUT("/{id}", updateUserHandler)
users.DELETE("/{id}", deleteUserHandler)

// Posts
posts := api.Group("/posts")
posts.GET("/", getPostsHandler)
posts.POST("/", createPostHandler)
```

### 2. Use Consistent Middleware

Apply middleware consistently across related routes:

```go
// All API routes need authentication
api := r.Group("/api", "auth")

// Public routes separate
public := r.Group("/public")
public.GET("/health", healthHandler)
```

### 3. Parameter Validation

Use typed parameters for automatic validation:

```go
type GetUserRequest struct {
    ID string `path:"id" validate:"required,uuid"`
}

func getUserWithValidation(ctx *lokstra.Context, params *GetUserRequest) error {
    // ID is automatically validated as required UUID
    return ctx.Ok(fmt.Sprintf("User ID: %s", params.ID))
}
```

### 4. Error Handling

Return appropriate HTTP status codes:

```go
func getUserHandler(ctx *lokstra.Context) error {
    userID := ctx.GetPathParam("id")
    
    if userID == "" {
        return ctx.ErrorBadRequest("User ID is required")
    }
    
    user, err := userService.GetByID(userID)
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            return ctx.ErrorNotFound("User not found")
        }
        return err // 500 Internal Server Error
    }
    
    return ctx.Ok(user)
}
```

### 5. RESTful Design

Follow REST conventions for predictable APIs:

```go
// RESTful user routes
r.GET("/users", getUsersHandler)           // List users
r.POST("/users", createUserHandler)        // Create user
r.GET("/users/{id}", getUserHandler)       // Get specific user
r.PUT("/users/{id}", updateUserHandler)    // Update user (full)
r.PATCH("/users/{id}", patchUserHandler)   // Update user (partial)
r.DELETE("/users/{id}", deleteUserHandler) // Delete user

// Nested resources
r.GET("/users/{userId}/posts", getUserPostsHandler)
r.POST("/users/{userId}/posts", createUserPostHandler)
```

## Next Steps

- [Middleware](./middleware.md) - Learn about custom middleware development
- [Core Concepts](./core-concepts.md) - Understand request/response handling
- [Services](./services.md) - Integrate with Lokstra's service system
- [HTMX Integration](./htmx-integration.md) - Build dynamic frontend applications

---

*The Lokstra router provides all the tools you need to build scalable, maintainable web APIs and applications. Start with simple routes and expand to complex architectures as your application grows.*