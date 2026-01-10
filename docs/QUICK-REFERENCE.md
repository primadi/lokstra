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
lokstra autogen .           # Manual generation
go run . --generate-only    # Force rebuild all

# Recommended: Use lokstra.Bootstrap() in main() for auto-generation
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

## Database & Transactions

### Inject DB Pool

```go
// @Service "user-repository"
type UserRepository struct {
    // @Inject "main-db"
    DB serviceapi.DbPool
}

func (r *UserRepository) GetUser(ctx context.Context, id string) (*User, error) {
    conn, err := r.DB.Acquire(ctx)
    if err != nil {
        return nil, err
    }
    defer conn.Release()
    
    var user User
    err = conn.QueryRow(ctx,
        "SELECT id, name, email FROM users WHERE id = $1", id,
    ).Scan(&user.ID, &user.Name, &user.Email)
    
    return &user, err
}
```

### Transactions (Auto-Finalized)

```go
// ✅ RECOMMENDED: Automatic management
func (s *Service) Create(ctx *request.Context, req *Request) error {
    ctx.BeginTransaction("main-db")  // No defer needed!
    
    s.repo1.Create(ctx, data1)
    s.repo2.Create(ctx, data2)
    
    return ctx.Api.Ok(data)  // Auto-commit (status 200)
}

// ❌ Error or 400+ status → Auto-rollback
func (s *Service) Update(ctx *request.Context, req *Request) error {
    ctx.BeginTransaction("main-db")
    
    if err := s.repo.Update(ctx, data); err != nil {
        return err  // ← Rollback
    }
    
    return ctx.Api.BadRequest("Invalid") // ← Also rollback (status 400)
}
```

**Auto-finalization rules:**
- **Commit:** `nil` error + status < 400
- **Rollback:** Error returned OR status >= 400

### Manual Transaction Control

```go
// Dry-run: Execute but rollback
func (s *Service) DryRun(ctx *request.Context) error {
    ctx.BeginTransaction("main-db")
    
    result, _ := s.repo.Create(ctx, data)
    
    ctx.RollbackTransaction("main-db")  // ← Force rollback
    return ctx.Api.Ok(result)           // 200 OK, no data saved
}

// Conditional commit
func (s *Service) Batch(ctx *request.Context, items []Item) error {
    ctx.BeginTransaction("main-db")
    
    successCount := s.process(ctx, items)
    
    if successCount < len(items)*0.8 {
        ctx.RollbackTransaction("main-db")  // < 80% success
    } else {
        ctx.CommitTransaction("main-db")    // >= 80% success
    }
    
    return ctx.Api.Ok(map[string]any{"success": successCount})
}
```

### Multiple Pools

```go
func (s *Service) CrossDatabase(ctx *request.Context) error {
    ctx.BeginTransaction("db-auth")
    ctx.BeginTransaction("db-tenant")  // Independent
    
    s.authRepo.Create(ctx, authData)    // Uses db-auth tx
    s.tenantRepo.Create(ctx, tenantData) // Uses db-tenant tx
    
    return ctx.Api.Ok("done")  // Both auto-commit
}
```

### Service Layer (Without Request Context)

```go
import "github.com/primadi/lokstra/serviceapi"

func (s *Service) DoWork(ctx context.Context) (err error) {
    ctx, finish := serviceapi.BeginTransaction(ctx, "main-db")
    defer finish(&err)  // ← Manual defer needed
    
    s.repo1.Create(ctx, ...)
    s.repo2.Update(ctx, ...)
    
    return nil  // Auto-commit
}
```

---

## Service Patterns (Recommended: Use Annotations)

### Annotation-Based Service (Recommended)

```go
// @EndpointService name="user-service", prefix="/api/users"
type UserService struct {
    // @Inject "user-repository"
    UserRepo UserRepository
}

// @Route "GET /{id}"
func (s *UserService) GetByID(p *GetUserParams) (*User, error) {
    return s.UserRepo.GetByID(p.ID)
}

// @Route "POST /", middlewares=["auth"]
func (s *UserService) Create(p *CreateUserParams) (*User, error) {
    u := &User{Name: p.Name, Email: p.Email}
    return s.UserRepo.Create(u)
}

// @Route "DELETE /{id}", middlewares=["auth", "admin"]
func (s *UserService) Delete(p *DeleteUserParams) error {
    return s.UserRepo.Delete(p.ID)
}

// Generate code
// lokstra autogen .
```

**Annotation with Variables** (resolves from config.yaml):
```go
// @EndpointService name="user-service", prefix="${api-prefix}"
// @Route "GET ${api-version}/users/{id}"
```

```yaml
# config.yaml
configs:
  - name: api-prefix
    value: /api/v1
  - name: api-version
    value: v2
```

### Pure Service with @Service (Recommended)

For non-HTTP services (business logic, utilities, infrastructure):

```go
// @Service name="auth-service"
type AuthService struct {
    // Required dependency
    // @Inject "user-repository"
    UserRepo UserRepository
    
    // Optional dependency
    // @Inject service="cache-service", optional=true
    Cache CacheService
    
    // Configuration injection
    // @Inject "cfg:auth.jwt-secret"
    JwtSecret string
    
    // @Inject "cfg:auth.token-expiry", "24h"
    TokenExpiry time.Duration
    
    // @Inject "cfg:auth.max-attempts", "5"
    MaxAttempts int
}

func (s *AuthService) Login(email, password string) (string, error) {
    // Use cache if available
    if s.Cache != nil {
        // Check cache
    }
    
    user, err := s.UserRepo.GetByEmail(email)
    if err != nil {
        return "", err
    }
    
    token := s.generateToken(user.ID, s.TokenExpiry)
    return token, nil
}
```

**Config (config.yaml):**
```yaml
configs:
  auth:
    jwt-secret: "your-secret-key"
    token-expiry: "48h"
    max-attempts: 3
```

**@Service supports:**
- `@Inject` - Service dependencies (required or optional)
- `@Inject "cfg:..."` - Configuration injection (auto-typed)
- No HTTP routes (use `@EndpointService` for that)

**Generated code:**
```go
func RegisterAuthService() {
    lokstra_registry.RegisterLazyService("auth-service", func(deps map[string]any, cfg map[string]any) any {
        return &AuthService{
            UserRepo:    lokstra_registry.GetService[UserRepository]("user-repository"),
            Cache:       // Optional - returns nil if not found
            JwtSecret:   lokstra_registry.GetConfig("auth.jwt-secret", ""),
            TokenExpiry: lokstra_registry.GetConfigDuration("auth.token-expiry", 24*time.Hour),
            MaxAttempts: lokstra_registry.GetConfigInt("auth.max-attempts", 5),
        }
    }, nil)
}
```

### Annotation Summary

| Annotation | Purpose | Where |
|------------|---------|-------|
| `@EndpointService` | HTTP service with routes | Above struct |
| `@Service` | Pure service (no HTTP) | Above struct |
| `@Route` | HTTP endpoint | Above method (RouterService only) |
| `@Inject` | Dependency/config injection | Above field |

**@Inject parameters:****
- `service` (positional or named) - service name (or use `cfg:` prefix for config)
- `optional` - `true`/`false` (default: `false`)
- For config: `@Inject "cfg:config.key"` or `@Inject "cfg:config.key", "default"`
- Type auto-detected for config: `string`, `int`, `bool`, `float64`, `time.Duration`

### Manual Service Factory (Advanced)

```go
type UserService struct {
    UserRepo UserRepository
}

func (s *UserService) GetByID(id string) (*User, error) {
    return s.UserRepo.GetByID(id)
}

// Local factory
func UserServiceFactory(deps map[string]any, config map[string]any) any {
    return &UserService{
        UserRepo: deps["user-repository"].(UserRepository),
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

Services are loaded lazily (created on first access), but their dependencies are eagerly resolved when the service is created:

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
// @EndpointService name="user-service", prefix="/api", middlewares=["recovery"]
type UserService struct {
    // @Inject "user-repository"
    UserRepo UserRepository
}

// @Route "GET /users/{id}"
func (s *UserService) GetByID(p *GetUserParams) (*User, error) {
    return s.UserRepo.GetByID(p.ID)
}

// @Route "GET /users"
func (s *UserService) List(p *ListUsersParams) ([]*User, error) {
    return s.UserRepo.List()
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
```go
// Recommended: Auto-generation in main()
func main() {
    lokstra.Bootstrap() // Detects changes and regenerates
    // ...
}
```

```bash
# Manual generation (before build/deploy)
lokstra autogen ./path/to/service
go run . --generate-only
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
