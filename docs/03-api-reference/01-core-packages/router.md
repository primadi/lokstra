# Router

> HTTP routing and request handling

## Overview

The Router interface provides a flexible, type-safe way to register HTTP routes and middleware in Lokstra applications. It supports multiple handler signatures, automatic request binding, middleware chains, and route grouping.

## Import Path

```go
import "github.com/primadi/lokstra/core/router"

// Or use the main package alias
import "github.com/primadi/lokstra"
router := lokstra.NewRouter("api")
```

---

## Interface

```go
type Router interface {
    http.Handler
    
    // Metadata
    Name() string
    EngineType() string
    PathPrefix() string
    SetPathPrefix(prefix string) Router
    Clone() Router
    
    // HTTP Methods
    GET(path string, h any, middleware ...any) Router
    POST(path string, h any, middleware ...any) Router
    PUT(path string, h any, middleware ...any) Router
    DELETE(path string, h any, middleware ...any) Router
    PATCH(path string, h any, middleware ...any) Router
    ANY(path string, h any, middleware ...any) Router
    
    // Prefix Matching
    GETPrefix(prefix string, h any, middleware ...any) Router
    POSTPrefix(prefix string, h any, middleware ...any) Router
    PUTPrefix(prefix string, h any, middleware ...any) Router
    DELETEPrefix(prefix string, h any, middleware ...any) Router
    PATCHPrefix(prefix string, h any, middleware ...any) Router
    ANYPrefix(prefix string, h any, middleware ...any) Router
    
    // Grouping
    Group(prefix string, fn func(r Router)) Router
    AddGroup(prefix string) Router
    
    // Middleware
    Use(middleware ...any) Router
    WithOverrideParentMiddleware(override bool) Router
    
    // Introspection
    Walk(fn func(rt *route.Route))
    PrintRoutes()
    
    // Lifecycle
    Build()
    IsBuilt() bool
    
    // Chaining
    IsChained() bool
    GetNextChain() Router
    SetNextChain(next Router) Router
    SetNextChainWithPrefix(next Router, prefix string) Router
}
```

---

## Methods

### Name
Returns the router identifier.

**Signature:**
```go
func (r Router) Name() string
```

**Example:**
```go
router := lokstra.NewRouter("api-v1")
fmt.Println(router.Name()) // Output: api-v1
```

---

### EngineType
Returns the underlying engine type.

**Signature:**
```go
func (r Router) EngineType() string
```

**Returns:**
- `"default"` - Lokstra's default router engine
- `"servemux"` - Go's standard http.ServeMux
- Custom engine types

**Example:**
```go
router := lokstra.NewRouter("api")
fmt.Println(router.EngineType()) // Output: default
```

---

### PathPrefix
Returns the current path prefix of the router.

**Signature:**
```go
func (r Router) PathPrefix() string
```

**Example:**
```go
api := lokstra.NewRouter("api")
v1 := api.AddGroup("/v1")
fmt.Println(v1.PathPrefix()) // Output: /v1
```

---

### SetPathPrefix
Sets the path prefix for the router.

**Signature:**
```go
func (r Router) SetPathPrefix(prefix string) Router
```

**Parameters:**
- `prefix` - Path prefix (e.g., "/api", "/v1")

**Returns:**
- `Router` - Returns self for chaining

**Example:**
```go
router := lokstra.NewRouter("api")
router.SetPathPrefix("/api/v1")
router.GET("/users", handler) // Actual path: /api/v1/users
```

---

### Clone
Creates a shallow copy of the router without routes and children.

**Signature:**
```go
func (r Router) Clone() Router
```

**Returns:**
- `Router` - New router instance with same configuration

**Use Cases:**
- Creating independent router instances
- Internal use by App when mounting routers

**Example:**
```go
template := lokstra.NewRouter("template")
template.Use("cors", "logger")

router1 := template.Clone()
router2 := template.Clone()
// Each clone has middleware but no routes
```

---

## HTTP Method Registration

All HTTP method functions (`GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `ANY`) share the same flexible signature.

### GET
Register a GET route.

**Signature:**
```go
func (r Router) GET(path string, h any, middleware ...any) Router
```

**Parameters:**
- `path` - Route path (supports params: `/users/:id`, wildcards: `/files/*`)
- `h` - Handler (see [Handler Signatures](#handler-signatures))
- `middleware` - Optional middleware (see [Middleware](#middleware-parameter))

**Returns:**
- `Router` - Returns self for chaining

**Example:**
```go
router.GET("/users", listUsers)
router.GET("/users/:id", getUser)
router.GET("/files/*filepath", serveFile)
```

---

### POST
Register a POST route.

**Signature:**
```go
func (r Router) POST(path string, h any, middleware ...any) Router
```

**Example:**
```go
// Simple handler
router.POST("/users", createUser)

// With request body binding
router.POST("/users", func(c *lokstra.RequestContext, input *CreateUserInput) error {
    // input automatically bound from request body
    return c.Api.Created(user)
})

// With middleware
router.POST("/users", createUser, "auth", "validate")
```

---

### PUT
Register a PUT route.

**Signature:**
```go
func (r Router) PUT(path string, h any, middleware ...any) Router
```

**Example:**
```go
router.PUT("/users/:id", updateUser)
router.PUT("/users/:id", func(c *lokstra.RequestContext, input *UpdateUserInput) error {
    id := c.Req.Param("id")
    // ... update user
    return c.Api.Success(user)
}, "auth")
```

---

### DELETE
Register a DELETE route.

**Signature:**
```go
func (r Router) DELETE(path string, h any, middleware ...any) Router
```

**Example:**
```go
router.DELETE("/users/:id", deleteUser)
router.DELETE("/users/:id", func(c *lokstra.RequestContext) error {
    id := c.Req.Param("id")
    // ... delete user
    return c.Api.NoContent()
}, "auth", "admin")
```

---

### PATCH
Register a PATCH route.

**Signature:**
```go
func (r Router) PATCH(path string, h any, middleware ...any) Router
```

**Example:**
```go
router.PATCH("/users/:id", patchUser)
router.PATCH("/users/:id/email", func(c *lokstra.RequestContext, input *UpdateEmailInput) error {
    // Partial update
    return c.Api.Success(user)
})
```

---

### ANY
Register a route for all HTTP methods.

**Signature:**
```go
func (r Router) ANY(path string, h any, middleware ...any) Router
```

**Example:**
```go
// Handle all methods
router.ANY("/webhook", handleWebhook)

// CORS preflight + actual request
router.ANY("/api/*", corsHandler)
```

---

## Prefix Matching Routes

Prefix matching routes match any path starting with the given prefix.

### GETPrefix
Register a GET route with prefix matching.

**Signature:**
```go
func (r Router) GETPrefix(prefix string, h any, middleware ...any) Router
```

**Example:**
```go
// Matches /static/*, /static/css/*, /static/js/*, etc.
router.GETPrefix("/static/", serveStaticFiles)

// Matches /docs, /docs/, /docs/api, etc.
router.GETPrefix("/docs", serveDocs)
```

---

### POSTPrefix, PUTPrefix, DELETEPrefix, PATCHPrefix
Register prefix matching routes for other HTTP methods.

**Signatures:**
```go
func (r Router) POSTPrefix(prefix string, h any, middleware ...any) Router
func (r Router) PUTPrefix(prefix string, h any, middleware ...any) Router
func (r Router) DELETEPrefix(prefix string, h any, middleware ...any) Router
func (r Router) PATCHPrefix(prefix string, h any, middleware ...any) Router
```

---

### ANYPrefix
Register a prefix matching route for all HTTP methods.

**Signature:**
```go
func (r Router) ANYPrefix(prefix string, h any, middleware ...any) Router
```

**Example:**
```go
// Catch-all for API v2
router.ANYPrefix("/api/v2/", v2Handler)
```

---

## Route Grouping

### Group
Create a sub-router with prefix and register routes in a callback.

**Signature:**
```go
func (r Router) Group(prefix string, fn func(r Router)) Router
```

**Parameters:**
- `prefix` - Path prefix for the group
- `fn` - Callback function to register routes

**Returns:**
- `Router` - Returns parent router for chaining

**Example:**
```go
router := lokstra.NewRouter("api")

// API v1
router.Group("/v1", func(v1 Router) {
    v1.GET("/users", v1ListUsers)
    v1.POST("/users", v1CreateUser)
})

// API v2
router.Group("/v2", func(v2 Router) {
    v2.GET("/users", v2ListUsers)
    v2.POST("/users", v2CreateUser)
})

// Admin group with middleware
router.Group("/admin", func(admin Router) {
    admin.Use("auth", "admin-only")
    admin.GET("/stats", getStats)
    admin.GET("/users", adminListUsers)
})

// Nested groups
router.Group("/api", func(api Router) {
    api.Group("/v1", func(v1 Router) {
        v1.GET("/health", healthCheck) // Path: /api/v1/health
    })
})
```

---

### AddGroup
Create a sub-router with prefix and return it for further registration.

**Signature:**
```go
func (r Router) AddGroup(prefix string) Router
```

**Parameters:**
- `prefix` - Path prefix for the group

**Returns:**
- `Router` - New sub-router instance

**Example:**
```go
router := lokstra.NewRouter("api")

// Create groups
v1 := router.AddGroup("/v1")
v2 := router.AddGroup("/v2")
admin := router.AddGroup("/admin")

// Register routes on each group
v1.GET("/users", v1ListUsers)
v2.GET("/users", v2ListUsers)

admin.Use("auth", "admin-only")
admin.GET("/stats", getStats)
```

**Difference from Group():**
- `Group()` - Inline callback style
- `AddGroup()` - Returns router for separate registration

---

## Middleware

### Use
Add global middleware to the router.

**Signature:**
```go
func (r Router) Use(middleware ...any) Router
```

**Parameters:**
- `middleware` - One or more middleware (see [Middleware Parameter](#middleware-parameter))

**Returns:**
- `Router` - Returns self for chaining

**Example:**
```go
router := lokstra.NewRouter("api")

// Single middleware
router.Use("cors")

// Multiple middleware (executed in order)
router.Use("cors", "logger", "recovery")

// Function middleware
router.Use(func(c *lokstra.RequestContext) error {
    log.Println("Before request")
    err := c.Next()
    log.Println("After request")
    return err
})

// Mixed
router.Use("cors", loggingMiddleware, "auth")
```

**Middleware Execution Order:**
```go
router.Use("A", "B")          // Global: A → B
router.GET("/users", h, "C")  // Route-specific: C

// Execution: A → B → C → handler
```

---

### WithOverrideParentMiddleware
Control whether child routers inherit parent middleware.

**Signature:**
```go
func (r Router) WithOverrideParentMiddleware(override bool) Router
```

**Parameters:**
- `override` - If `true`, child middleware replaces parent middleware. If `false` (default), child middleware appends to parent.

**Returns:**
- `Router` - Returns self for chaining

**Example:**
```go
parent := lokstra.NewRouter("parent")
parent.Use("A", "B")

// Default: inherit parent middleware
child1 := parent.AddGroup("/child1")
child1.Use("C")
// Middleware: A → B → C

// Override: replace parent middleware
child2 := parent.AddGroup("/child2")
child2.WithOverrideParentMiddleware(true).Use("D")
// Middleware: D (only)
```

---

## Handler Signatures

Lokstra supports flexible handler signatures with automatic type detection.

### Basic Handlers

**No parameters:**
```go
func handler() error {
    return nil
}
```

**Context only:**
```go
func handler(c *lokstra.RequestContext) error {
    return c.Api.Success("OK")
}
```

**Context + Request Body:**
```go
type CreateUserInput struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

func handler(c *lokstra.RequestContext, input *CreateUserInput) error {
    // input automatically bound and validated
    return c.Api.Created(user)
}
```

**Request Body only (context available via input):**
```go
func handler(input *CreateUserInput) error {
    // Access context if needed
    return nil
}
```

### Return Types

**Error only:**
```go
func handler(c *lokstra.RequestContext) error {
    if err := doSomething(); err != nil {
        return err // Auto-converted to 500 response
    }
    return c.Api.Success("OK")
}
```

**Response object:**
```go
func handler(c *lokstra.RequestContext) *response.Response {
    return c.Resp.WithStatus(200).Json(data)
}
```

**API helper:**
```go
func handler(c *lokstra.RequestContext) *response.ApiHelper {
    return c.Api.Success(data)
}
```

**Any data (auto-wrapped):**
```go
func handler(c *lokstra.RequestContext) any {
    return user // Auto-wrapped in ApiResponse
}
```

**Data + Error:**
```go
func handler(c *lokstra.RequestContext) (any, error) {
    user, err := getUser()
    if err != nil {
        return nil, err
    }
    return user, nil
}

func handler2(c *lokstra.RequestContext) (*User, error) {
    return getUser()
}
```

**Response + Error:**
```go
func handler(c *lokstra.RequestContext) (*response.Response, error) {
    user, err := getUser()
    if err != nil {
        return nil, err
    }
    return c.Resp.WithStatus(200).Json(user), nil
}
```

---

## Middleware Parameter

The `middleware ...any` parameter accepts multiple formats:

### String (Middleware Name)
```go
router.GET("/users", handler, "cors", "auth", "logger")
```
References middleware registered in config or registry.

### HandlerFunc
```go
func loggingMiddleware(c *lokstra.RequestContext) error {
    log.Println("Request:", c.R.URL.Path)
    return c.Next()
}

router.GET("/users", handler, loggingMiddleware)
```

### Inline Function
```go
router.GET("/users", handler, func(c *lokstra.RequestContext) error {
    // Middleware logic
    return c.Next()
})
```

### Route Options
```go
import "github.com/primadi/lokstra/core/route"

router.GET("/users", handler,
    route.WithMiddleware("auth"),
    route.WithName("list-users"),
)
```

### Mixed
```go
router.GET("/users", handler,
    "cors",                    // String
    loggingMiddleware,         // Function
    func(c *lokstra.RequestContext) error { // Inline
        return c.Next()
    },
    route.WithName("users"),   // Option
)
```

---

## Path Patterns

### Static Paths
```go
router.GET("/users", handler)
router.GET("/api/v1/health", handler)
```

### Named Parameters
```go
router.GET("/users/:id", handler)
router.GET("/posts/:postID/comments/:commentID", handler)

func handler(c *lokstra.RequestContext) error {
    id := c.Req.Param("id")
    postID := c.Req.Param("postID")
    commentID := c.Req.Param("commentID")
    return c.Api.Success(data)
}
```

### Wildcard
```go
router.GET("/files/*filepath", handler)

func handler(c *lokstra.RequestContext) error {
    filepath := c.Req.Param("filepath")
    // filepath = "css/main.css" for request "/files/css/main.css"
    return c.Api.Success(filepath)
}
```

### Prefix Matching
```go
router.GETPrefix("/static/", handler)
// Matches: /static/css/main.css, /static/js/app.js, etc.
```

---

## Introspection

### Walk
Walk through all registered routes.

**Signature:**
```go
func (r Router) Walk(fn func(rt *route.Route))
```

**Parameters:**
- `fn` - Callback function called for each route

**Example:**
```go
router.Walk(func(rt *route.Route) {
    fmt.Printf("%s %s\n", rt.Method, rt.Path)
})
```

---

### PrintRoutes
Print all routes to stdout.

**Signature:**
```go
func (r Router) PrintRoutes()
```

**Example:**
```go
router.PrintRoutes()
// Output:
// GET /users
// POST /users
// GET /users/:id
// PUT /users/:id
// DELETE /users/:id
```

---

## Lifecycle

### Build
Finalize the router and build the underlying engine.

**Signature:**
```go
func (r Router) Build()
```

**Notes:**
- Called automatically by App
- Can be called manually for introspection
- After building, no more routes can be added

**Example:**
```go
router.GET("/users", handler)
router.Build() // Finalize
```

---

### IsBuilt
Check if the router has been built.

**Signature:**
```go
func (r Router) IsBuilt() bool
```

**Example:**
```go
if !router.IsBuilt() {
    router.GET("/new-route", handler)
}
```

---

## Router Chaining

### IsChained
Check if the router is part of a chain.

**Signature:**
```go
func (r Router) IsChained() bool
```

---

### GetNextChain
Get the next router in the chain.

**Signature:**
```go
func (r Router) GetNextChain() Router
```

**Returns:**
- `Router` - Next router, or `nil` if none

---

### SetNextChain
Set the next router in the chain.

**Signature:**
```go
func (r Router) SetNextChain(next Router) Router
```

**Returns:**
- `Router` - The next router

---

### SetNextChainWithPrefix
Set the next router in the chain with a prefix.

**Signature:**
```go
func (r Router) SetNextChainWithPrefix(next Router, prefix string) Router
```

**Parameters:**
- `next` - Next router
- `prefix` - Path prefix for the next router

**Returns:**
- `Router` - The next router

**Example:**
```go
api := lokstra.NewRouter("api")
admin := lokstra.NewRouter("admin")

api.SetNextChainWithPrefix(admin, "/admin")
// Requests to /admin/* handled by admin router
// All other requests handled by api router
```

---

## Complete Examples

### Basic CRUD API
```go
router := lokstra.NewRouter("api")

// List
router.GET("/users", func(c *lokstra.RequestContext) error {
    users := getUsersFromDB()
    return c.Api.Success(users)
})

// Get
router.GET("/users/:id", func(c *lokstra.RequestContext) error {
    id := c.Req.Param("id")
    user, err := getUserByID(id)
    if err != nil {
        return c.Api.NotFound("User not found")
    }
    return c.Api.Success(user)
})

// Create
type CreateUserInput struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

router.POST("/users", func(c *lokstra.RequestContext, input *CreateUserInput) error {
    user, err := createUser(input)
    if err != nil {
        return err
    }
    return c.Api.Created(user)
})

// Update
type UpdateUserInput struct {
    Name  string `json:"name"`
    Email string `json:"email" validate:"omitempty,email"`
}

router.PUT("/users/:id", func(c *lokstra.RequestContext, input *UpdateUserInput) error {
    id := c.Req.Param("id")
    user, err := updateUser(id, input)
    if err != nil {
        return err
    }
    return c.Api.Success(user)
})

// Delete
router.DELETE("/users/:id", func(c *lokstra.RequestContext) error {
    id := c.Req.Param("id")
    if err := deleteUser(id); err != nil {
        return err
    }
    return c.Api.NoContent()
})
```

### Versioned API with Middleware
```go
router := lokstra.NewRouter("api")

// Global middleware
router.Use("cors", "recovery", "logger")

// API v1
router.Group("/v1", func(v1 Router) {
    v1.GET("/users", v1ListUsers)
    v1.POST("/users", v1CreateUser)
    
    // Protected routes
    v1.Group("/admin", func(admin Router) {
        admin.Use("auth", "admin-role")
        admin.GET("/stats", getStats)
        admin.DELETE("/users/:id", adminDeleteUser)
    })
})

// API v2
router.Group("/v2", func(v2 Router) {
    v2.GET("/users", v2ListUsers)
    v2.POST("/users", v2CreateUser, "rate-limit")
})
```

### File Server
```go
router := lokstra.NewRouter("assets")

// Serve static files
router.GETPrefix("/static/", func(c *lokstra.RequestContext) error {
    filepath := c.Req.Param("filepath")
    http.ServeFile(c.W, c.R, "./public/"+filepath)
    return nil
})

// Serve SPA (catch-all)
router.GETPrefix("/", func(c *lokstra.RequestContext) error {
    http.ServeFile(c.W, c.R, "./public/index.html")
    return nil
})
```

---

## See Also

- **[lokstra](lokstra)** - Main package functions
- **[Request Context](request)** - Handler context API
- **[Response](response)** - Response helpers
- **[Route](../08-advanced/route)** - Route options and configuration

---

## Related Guides

- **[Router Essentials](../../01-essentials/01-router/)** - Learn router basics
- **[Middleware Guide](../../01-essentials/03-middleware/)** - Working with middleware
- **[Handler Patterns](../../02-deep-dive/router/)** - Advanced handler techniques
