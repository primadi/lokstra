---
layout: default
title: Quick Reference - Lokstra Framework
description: Quick reference cheatsheet for Lokstra Framework
---

# Lokstra Quick Reference

Fast lookup for common patterns, imports, and configurations.

---

## Installation

```bash
# Install CLI
go install github.com/primadi/lokstra/cmd/lokstra@latest

# Install framework
go get github.com/primadi/lokstra

# Create new project
lokstra new myapp
lokstra new myapp -template 02_app_framework/01_medium_system

# Generate code from annotations
lokstra autogen .
```

---

## Common Imports

```go
// Core framework
import "github.com/primadi/lokstra"
import "github.com/primadi/lokstra/core/request"
import "github.com/primadi/lokstra/core/service"
import "github.com/primadi/lokstra/core/deploy"
import "github.com/primadi/lokstra/lokstra_registry"

// Middleware
import "github.com/primadi/lokstra/middleware/recovery"
import "github.com/primadi/lokstra/middleware/request_logger"
import "github.com/primadi/lokstra/middleware/cors"
import "github.com/primadi/lokstra/middleware/body_limit"

// Built-in services
import "github.com/primadi/lokstra/services/dbpool_pg"
import "github.com/primadi/lokstra/services/redis"
```

---

## Router Basics

### Simple Router

```go
r := lokstra.NewRouter("api")

// Basic routes
r.GET("/", func() string { return "Hello" })
r.POST("/users", createUser)
r.PUT("/users/{id}", updateUser)
r.DELETE("/users/{id}", deleteUser)

// Groups
v1 := r.Group("/v1")
v1.GET("/users", listUsers)

// Middleware
r.Use(recovery.Middleware(nil))
r.Use(cors.Middleware([]string{"*"}))

// Run
app := lokstra.NewApp("myapp", ":8080", r)
app.Run(30 * time.Second)
```

### Handler Signatures

```go
// 1. Simple return
func() string { return "Hello" }

// 2. With error
func(id string) (string, error) { return "User", nil }

// 3. Struct response
func(id string) (*User, error) { return &User{}, nil }

// 4. Context access
func(ctx *request.Context) error { return ctx.Api.Ok("data") }

// 5. Request body
func(ctx *request.Context, params *CreateUserParams) error {
    return ctx.Api.Created(params)
}

// 6. Path + body
func(ctx *request.Context, id string, params *UpdateParams) error {
    return ctx.Api.Ok(params)
}

// 7. Query parameters
type SearchParams struct {
    Q     string `query:"q" validate:"required"`
    Page  int    `query:"page" validate:"min=1"`
}
func(params *SearchParams) ([]Result, error) { return results, nil }
```

---

## Service Patterns

### Service Factory

```go
type UserService struct {
    UserRepo *service.Cached[UserRepository]
}

func (s *UserService) GetByID(id string) (*User, error) {
    return s.UserRepo.MustGet().GetByID(id)
}

// Local factory
func UserServiceFactory(deps map[string]any, config map[string]any) any {
    return &UserService{
        UserRepo: service.Cast[UserRepository](deps["user-repository"]),
    }
}

// Remote factory (microservices)
func UserServiceRemoteFactory(deps map[string]any, config map[string]any) any {
    proxyService := config["remote"].(*proxy.Service)
    return NewUserServiceRemote(proxyService)
}
```

### Service Registration

```go
import "github.com/primadi/lokstra/lokstra_registry"

func registerServiceTypes() {
    lokstra_registry.RegisterServiceType(
        "user-service-factory",      // Type name
        UserServiceFactory,           // Local factory
        UserServiceRemoteFactory,     // Remote factory (optional)
    )
}
```

### Lazy Loading

```go
import "github.com/primadi/lokstra/core/service"

// Define lazy reference
var userService = service.LazyLoad[*UserService]("user-service")

func handler() {
    // First call loads service (thread-safe)
    users := userService.MustGet().GetAll()
    
    // Cached thereafter
    user := userService.MustGet().GetByID("123")
}
```

---

## Domain Models

```go
// Entity
type User struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// Request DTOs
type CreateUserParams struct {
    Name  string `json:"name" validate:"required,min=3,max=50"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"min=18,max=120"`
}

type GetUserParams struct {
    ID string `path:"id" validate:"required"`
}

type ListUsersParams struct {
    Page  int `query:"page" validate:"min=1"`
    Limit int `query:"limit" validate:"min=1,max=100"`
}

// Repository interface
type UserRepository interface {
    GetByID(id string) (*User, error)
    List() ([]*User, error)
    Create(user *User) (*User, error)
    Update(user *User) (*User, error)
    Delete(id string) error
}

// Service interface
type UserService interface {
    GetByID(p *GetUserParams) (*User, error)
    List(p *ListUsersParams) ([]*User, error)
    Create(p *CreateUserParams) (*User, error)
    Update(p *UpdateUserParams) (*User, error)
    Delete(p *DeleteUserParams) error
}
```

---

## Configuration YAML

### Minimal Config

```yaml
service-definitions:
  user-service:
    type: user-service-factory

deployments:
  development:
    servers:
      api:
        addr: ":8080"
        published-services:
          - user-service
```

### Complete Config

```yaml
# Middleware
middleware-definitions:
  recovery:
    type: recovery
    config:
      enable_stack_trace: false
  
  request-logger:
    type: request-logger
    config:
      prefix: "API"
      skip_paths: ["/health"]

# Services
service-definitions:
  user-repository:
    type: user-repository-factory
    config:
      dsn: "postgres://localhost/mydb"
  
  user-service:
    type: user-service-factory
    depends-on:
      - user-repository
    router:
      path-prefix: /api
      middlewares:
        - recovery
        - request-logger
      hidden:
        - InternalMethod

# Deployments
deployments:
  development:
    servers:
      api:
        base-url: "http://localhost:8080"
        addr: ":8080"
        published-services:
          - user-service
  
  production:
    servers:
      user-api:
        base-url: "https://user-api.example.com"
        addr: ":8001"
        published-services:
          - user-service
```

---

## Annotations

```go
// @RouterService name="user-service", prefix="/api", middlewares=["recovery"]
type UserService struct {
    // @Inject "user-repository"
    UserRepo *service.Cached[UserRepository]
}

// @Route "GET /users/{id}"
func (s *UserService) GetByID(p *GetUserParams) (*User, error) {
    return s.UserRepo.MustGet().GetByID(p.ID)
}

// @Route "GET /users"
func (s *UserService) List(p *ListUsersParams) ([]*User, error) {
    return s.UserRepo.MustGet().List()
}

// @Route "POST /users", middlewares=["auth"]
func (s *UserService) Create(p *CreateUserParams) (*User, error) {
    // ...
}

// @Route "PUT /users/{id}", middlewares=["auth", "admin"]
func (s *UserService) Update(p *UpdateUserParams) (*User, error) {
    // ...
}

// @Route "DELETE /users/{id}", middlewares=["auth", "admin"]
func (s *UserService) Delete(p *DeleteUserParams) error {
    // ...
}
```

**Generate code:**
```bash
lokstra autogen ./path/to/service
```

**Per-route middleware:**
- Add `middlewares=["mw1", "mw2"]` to `@Route` annotation
- Route-specific middleware applies only to that route
- Service-level middleware applies to all routes

---

## Middleware

### Built-in Middleware

```go
import (
    "github.com/primadi/lokstra/middleware/recovery"
    "github.com/primadi/lokstra/middleware/request_logger"
    "github.com/primadi/lokstra/middleware/slow_request_logger"
    "github.com/primadi/lokstra/middleware/cors"
    "github.com/primadi/lokstra/middleware/body_limit"
    "github.com/primadi/lokstra/middleware/gzipcompression"
)

r := lokstra.NewRouter("api")

// Recommended order
r.Use(recovery.Middleware(nil))
r.Use(request_logger.Middleware(nil))
r.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold: 500 * time.Millisecond,
}))
r.Use(cors.Middleware([]string{"*"}))
r.Use(body_limit.Middleware(&body_limit.Config{
    MaxSize: 10 * 1024 * 1024, // 10MB
}))
r.Use(gzipcompression.Middleware(nil))
```

### Custom Middleware

```go
func CustomMiddleware(cfg *Config) request.HandlerFunc {
    return request.HandlerFunc(func(ctx *request.Context) error {
        // Pre-processing
        
        err := ctx.Next() // Call next handler
        
        // Post-processing
        
        return err
    })
}
```

---

## Response Helpers

```go
func handler(ctx *request.Context) error {
    // Success
    ctx.Api.Ok(data)                    // 200
    ctx.Api.Created(data)               // 201
    ctx.Api.NoContent()                 // 204
    
    // Client errors
    ctx.Api.BadRequest("message")       // 400
    ctx.Api.Unauthorized("message")     // 401
    ctx.Api.Forbidden("message")        // 403
    ctx.Api.NotFound("message")         // 404
    
    // Server errors
    ctx.Api.InternalServerError("msg")  // 500
    
    // Custom
    ctx.Api.ErrorWithCode(422, "message", data)
}
```

---

## Validation Tags

```go
type Params struct {
    // String validation
    Name  string `validate:"required"`
    Email string `validate:"required,email"`
    URL   string `validate:"url"`
    UUID  string `validate:"uuid"`
    
    // Numeric validation
    Age   int `validate:"min=18,max=120"`
    Price int `validate:"gt=0"`
    
    // String length
    Title string `validate:"min=3,max=100"`
    Code  string `validate:"len=6"`
    
    // Options
    Status string `validate:"oneof=active inactive pending"`
    
    // Nested struct
    Address *Address `validate:"required"`
    
    // Slice
    Tags []string `validate:"min=1,max=10,dive,min=2"`
}
```

**Common validators:**
- `required` - Must not be empty
- `email` - Valid email
- `url` - Valid URL
- `uuid` - Valid UUID
- `min=N` - Minimum value/length
- `max=N` - Maximum value/length
- `len=N` - Exact length
- `gt=N` - Greater than
- `gte=N` - Greater than or equal
- `lt=N` - Less than
- `lte=N` - Less than or equal
- `oneof=val1 val2` - One of values
- `dive` - Validate slice/map elements

---

## Environment Variables

```bash
# Deployment selection
LOKSTRA_DEPLOYMENT=production

# Server selection (multi-server)
LOKSTRA_SERVER=api-server

# Log level
LOKSTRA_LOG_LEVEL=debug  # silent, error, warn, info, debug

# Config file
LOKSTRA_CONFIG=./config.yaml
```

---

## Project Structure

### Simple Router

```
myapp/
├── main.go
├── handlers.go
└── go.mod
```

### Medium System (DDD)

```
myapp/
├── main.go
├── register.go
├── config.yaml
├── domain/
│   └── user/
│       ├── models.go
│       ├── repository.go
│       └── service.go
├── repository/
│   └── user_repository.go
└── service/
    └── user_service.go
```

### Enterprise Modular

```
myapp/
├── main.go
├── register.go
├── config.yaml
└── modules/
    ├── user/
    │   ├── domain/
    │   ├── application/
    │   └── infrastructure/
    └── order/
        ├── domain/
        ├── application/
        └── infrastructure/
```

---

## CLI Commands

```bash
# Create project
lokstra new myapp
lokstra new myapp -template 01_router/01_router_only
lokstra new myapp -template 02_app_framework/01_medium_system

# Generate code
lokstra autogen .
lokstra autogen ./modules/user/application

# List templates
lokstra new --help
```

---

## Common Issues

### Service Not Found
```
panic: service 'user-service' not found
```
**Fix:** Register service type in `register.go`

### Import Cycle
```
import cycle not allowed
```
**Fix:** Use domain interfaces, separate packages properly

### Validation Not Working
```
validation tags ignored
```
**Fix:** Use pointer to struct: `*CreateUserParams`

### Handler Not Recognized
```
unsupported handler signature
```
**Fix:** Return error or supported type, use `*request.Context`

---

## Resources

- **Documentation:** https://primadi.github.io/lokstra/
- **AI Agent Guide:** https://primadi.github.io/lokstra/AI-AGENT-GUIDE
- **GitHub:** https://github.com/primadi/lokstra
- **Examples:** https://primadi.github.io/lokstra/00-introduction/examples/
- **Schema:** https://primadi.github.io/lokstra/schema/lokstra.schema.json

---

## Templates

| Template | Use Case | Complexity |
|----------|----------|------------|
| `01_router/01_router_only` | Learning, simple APIs | ⭐ |
| `01_router/02_single_app` | Production single app | ⭐⭐ |
| `01_router/03_multi_app` | Multi-app servers | ⭐⭐⭐ |
| `02_app_framework/01_medium_system` | 2-10 entities, DDD | ⭐⭐⭐ |
| `02_app_framework/02_enterprise_modular` | 10+ entities, bounded contexts | ⭐⭐⭐⭐⭐ |
| `02_app_framework/03_enterprise_router_service` | Annotation-based, enterprise | ⭐⭐⭐⭐⭐ |
