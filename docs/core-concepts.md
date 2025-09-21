# Core Concepts

Lokstra is built around structured and flexible building blocks that work together to create scalable web applications. This document provides an overview of the key concepts that make Lokstra powerful yet easy to use.

## Registration Context

The **Registration Context** is the foundation of Lokstra's dependency injection system. It manages all services, handlers, middleware, and modules in your application.

### Key Features

- **Service Management**: Register, create, and retrieve services by name
- **Handler Registry**: Named handlers with priority ordering
- **Middleware Pipeline**: Configurable middleware chain
- **Module System**: Load reusable feature packages

### Basic Usage

```go
// Create global context with default services
regCtx := lokstra.NewGlobalRegistrationContext()

// Register a custom service
regCtx.RegisterServiceFactory("my-db", func(config map[string]any) (service.Service, error) {
    return &MyDatabase{}, nil
})

// Register handlers
regCtx.RegisterHandler("getUser", func(ctx *lokstra.Context) error {
    return ctx.Ok("User data")
})
```

## Request Context

Every HTTP request is handled through a **Request Context** (`*lokstra.Context`), which provides a unified interface for request processing and response generation.

### Core Features

- Wraps Go's standard `context.Context`
- Built-in request binding with struct tags
- Structured response helpers
- Type-safe parameter extraction

### Request Binding

Lokstra supports comprehensive request binding using struct tags:

```go
type UserRequest struct {
    ID    string `path:"id"`           // From URL path
    Token string `header:"Authorization"` // From HTTP headers
    Name  string `body:"name"`         // From request body
    Page  int    `query:"page"`        // From query parameters
}

func getUserHandler(ctx *lokstra.Context) error {
    var req UserRequest
    
    // Bind all sources automatically
    if err := ctx.BindAllSmart(&req); err != nil {
        return ctx.ErrorBadRequest(err.Error())
    }
    
    // Use the bound data
    return ctx.Ok(map[string]any{
        "user_id": req.ID,
        "name":    req.Name,
    })
}

// Alternative: Auto-bind smart pattern (automatic binding)
func getUserHandlerSmart(ctx *lokstra.Context, req *UserRequest) error {    
    // Request automatically bound - use the data directly
    return ctx.Ok(map[string]any{
        "user_id": req.ID,
        "name":    req.Name,
    })
}
```

### Binding Methods

- `BindPath(&dto)` - URL path parameters
- `BindQuery(&dto)` - Query string parameters  
- `BindHeader(&dto)` - HTTP headers
- `BindBody(&dto)` - JSON request body
- `BindAll(&dto)` - All sources (JSON body only)
- `BindBodySmart(&dto)` - Auto-detect body format (JSON, form, multipart)
- `BindAllSmart(&dto)` - All sources with smart body detection

## Handlers

Lokstra handlers follow a consistent pattern - they always return an `error`. This provides clear error handling and response management.

### Basic Handler

```go
func simpleHandler(ctx *lokstra.Context) error {
    return ctx.Ok("Hello, World!")
}
```

### Handler with Binding

```go
type CreateUserRequest struct {
    Name  string `body:"name"`
    Email string `body:"email"`
}

func createUserHandler(ctx *lokstra.Context) error {
    var req CreateUserRequest
    if err := ctx.BindBodySmart(&req); err != nil {
        return ctx.ErrorBadRequest(err.Error())
    }
    
    // Business logic here
    user := createUser(req.Name, req.Email)
    
    return ctx.OkCreated(user)
}
```

### Error Handling

Return appropriate HTTP status codes using response helpers:

```go
func getUserHandler(ctx *lokstra.Context) error {
    userID := ctx.GetPathParam("id")
    
    user, err := findUser(userID)
    if err != nil {
        if err == ErrUserNotFound {
            return ctx.ErrorNotFound("User not found")
        }
        return err // Returns 500 Internal Server Error
    }
    
    return ctx.Ok(user)
}
```

## Response System

Lokstra provides a structured response system with built-in helpers for common HTTP patterns.

### Response Structure

All responses include:
- `success` - Boolean indicating success/failure
- `code` - Response code (can be custom)
- `message` - Human-readable message
- `data` - Response payload

### Success Responses

```go
// Simple success
ctx.Ok("Hello")

// Created resource
ctx.OkCreated(newUser)

// No content
ctx.OkNoContent()

// Paginated data
ctx.OkPagination(users, totalCount, currentPage, pageSize)
```

### Error Responses

```go
// Client errors
ctx.ErrorBadRequest("Invalid input")
ctx.ErrorUnauthorized("Login required")
ctx.ErrorForbidden("Access denied")
ctx.ErrorNotFound("Resource not found")
ctx.ErrorConflict("Resource already exists")

// Server error
ctx.ErrorInternal("Something went wrong")
```

### Method Chaining

Responses support method chaining for customization:

```go
return ctx.Ok(data).
    WithMessage("Custom success message").
    WithHeader("X-Custom-Header", "value").
    WithResponseCode("CUSTOM_CODE")
```

## Architecture Components

### Server

The **Server** manages multiple applications in a single process:

```go
server := lokstra.NewServer(regCtx, "my-server")
server.AddApp(apiApp)
server.AddApp(adminApp)
server.StartWithGracefulShutdown(30 * time.Second)
```

### App

An **App** represents a single HTTP application with its own router and middleware:

```go
app := lokstra.NewApp(regCtx, "api-app", ":8080")
app.GET("/users", getUsersHandler)
app.POST("/users", createUserHandler)
```

### Router

The **Router** manages routes, groups, and middleware within an app:

```go
// Route groups
api := app.Group("/api/v1")
api.GET("/users", getUsersHandler)

// Middleware
api.Use(authMiddleware)

// Static file serving
app.MountStatic("/static", false, staticFS)
```

## Services and Dependency Injection

Lokstra includes a built-in service container for dependency management:

```go
// Get a service
dbPool, err := lokstra.GetService[serviceapi.DbPool](regCtx, "db.main")

// Create or get service
logger, err := lokstra.GetOrCreateService[serviceapi.Logger](regCtx, "logger", "default-logger")
```

## Middleware

Middleware functions process requests in a pipeline with configurable priority:

```go
func authMiddleware(next lokstra.HandlerFunc) lokstra.HandlerFunc {
    return func(ctx *lokstra.Context) error {
        token := ctx.GetHeader("Authorization")
        if token == "" {
            return ctx.ErrorUnauthorized("Token required")
        }
        
        // Validate token...
        
        return next(ctx)
    }
}

// Register middleware
app.Use(lokstra.NamedMiddleware("auth", authMiddleware))
```

## Configuration

Lokstra supports YAML-based configuration for declarative application setup:

```yaml
server:
  name: my-server

apps:
  - name: api-app
    addr: ":8080"
    middleware:
      - name: cors
      - name: logger
    routes:
      - method: GET
        path: /users
        handler: getUsers
      - method: POST
        path: /users
        handler: createUser
```

## HTMX Integration

Lokstra has first-class support for HTMX applications:

```go
// Mount HTMX pages
app.MountHtmx("/", htmxFS)

// HTMX page data handler
func dashboardData(ctx *lokstra.Context) error {
    return ctx.HtmxPageData("Dashboard", "Welcome to dashboard", map[string]any{
        "user": getCurrentUser(),
        "stats": getDashboardStats(),
    })
}
```

## Next Steps

- [Getting Started](./getting-started.md) - Build your first Lokstra app
- [Routing](./routing.md) - Advanced routing features
- [Middleware](./middleware.md) - Custom middleware development
- [Services](./services.md) - Service management and DI
- [Configuration](./configuration.md) - YAML configuration guide

---

*These core concepts form the foundation of all Lokstra applications. Understanding them will help you build robust and maintainable web applications.*