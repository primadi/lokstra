# GitHub Copilot Instructions for Lokstra Framework

This file provides instructions to GitHub Copilot when working with the Lokstra Framework.

## Framework Overview

Lokstra is a Go web framework with two modes:

1. **Router Mode** - Simple HTTP routing (like Echo, Gin, Chi)
2. **Framework Mode** - Full DI framework (like NestJS, Spring Boot)

## Key Documentation

When helping with Lokstra:

- **Complete AI Guide:** [AI-AGENT-GUIDE.md](../docs/AI-AGENT-GUIDE.md)
- **Quick Reference:** [QUICK-REFERENCE.md](../docs/QUICK-REFERENCE.md)
- **Full Documentation:** https://primadi.github.io/lokstra/

## Common Patterns

### 1. Simple Router (No DI)

```go
r := lokstra.NewRouter("api")
r.GET("/users", func() []User { return getUsers() })
app := lokstra.NewApp("myapp", ":8080", r)
app.Run(30 * time.Second)
```

### 2. Framework Mode (With DI)

**Service:**

```go
type UserService struct {
    UserRepo *service.Cached[UserRepository]
}

func UserServiceFactory(deps map[string]any, config map[string]any) any {
    // Get user repository from dependencies
    return &UserService{
        UserRepo: service.Cast[UserRepository]\(deps["user-repository"]),
    }
}
```

**Registration:**

```go
lokstra_registry.RegisterServiceType(
    "user-service-factory",
    UserServiceFactory,
    nil, // Remote factory (optional)
)
```

**config.yaml:**

```yaml
service-definitions:
  user-service:
    type: user-service-factory
    depends-on: [user-repository]

deployments:
  development:
    servers:
      api:
        addr: ":8080"
        published-services: [user-service]
```

### 3. Handler Signatures (29+ supported)

```go
// Simple
func() string { return "Hello" }

// With error
func(id string) (*User, error) { return user, nil }

// Request body (auto-validated)
func(ctx *request.Context, params *CreateUserParams) error {
    return ctx.Api.Created(params)
}

// Path + body
func(ctx *request.Context, id string, params *UpdateParams) error {
    return ctx.Api.Ok(params)
}
```

### 4. Domain Models

```go
type CreateUserParams struct {
    Name  string `json:"name" validate:"required,min=3,max=50"`
    Email string `json:"email" validate:"required,email"`
}
```

### 5. Annotations (Enterprise)

```go
// @RouterService name="user-service", prefix="/api"
type UserService struct {
    // @Inject "user-repository"
    UserRepo *service.Cached[UserRepository]
}

// @Route "GET /users/{id}"
func (s *UserService) GetByID(p *GetUserParams) (*User, error) {
    return s.UserRepo.MustGet().GetByID(p.ID)
}

// @Route "POST /users", middlewares=["auth"]
func (s *UserService) Create(p *CreateUserParams) (*User, error) {
    // ...
}

// @Route "DELETE /users/{id}", middlewares=["auth", "admin"]
func (s *UserService) Delete(p *DeleteUserParams) error {
    // ...
}
```

Generate code:

```bash
lokstra autogen .
```

**Per-route middleware:** Add `middlewares=["mw1", "mw2"]` to `@Route`

## Project Structure

### Simple (Router Only)

```
myapp/
├── main.go
└── handlers.go
```

### Medium (DDD)

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

### Enterprise (Modular)

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

## Common Imports

```go
import "github.com/primadi/lokstra"
import "github.com/primadi/lokstra/core/request"
import "github.com/primadi/lokstra/core/service"
import "github.com/primadi/lokstra/lokstra_registry"
import "github.com/primadi/lokstra/middleware/recovery"
import "github.com/primadi/lokstra/middleware/cors"
```

## Middleware Order (Best Practice)

```go
r.Use(recovery.Middleware(nil))           // 1. Catch panics
r.Use(request_logger.Middleware(nil))     // 2. Log requests
r.Use(cors.Middleware([]string{"*"}))     // 3. CORS
r.Use(body_limit.Middleware(nil))         // 4. Limit body size
```

## Response Helpers

```go
ctx.Api.Ok(data)                    // 200
ctx.Api.Created(data)               // 201
ctx.Api.BadRequest("message")       // 400
ctx.Api.Unauthorized("message")     // 401
ctx.Api.NotFound("message")         // 404
ctx.Api.InternalServerError("msg")  // 500
```

## CLI Commands

```bash
# Create new project
lokstra new myapp
lokstra new myapp -template 02_app_framework/01_medium_system

# Generate code from annotations
lokstra autogen .
```

## Templates

- `01_router/01_router_only` - Learning, simple APIs
- `02_app_framework/01_medium_system` - Production apps (2-10 entities)
- `02_app_framework/03_enterprise_router_service` - Enterprise with annotations

## Best Practices

1. **Always include error handling** in handlers
2. **Use validation tags** on request structs: `validate:"required,email"`
3. **Use pointer parameters** for request binding: `*CreateUserParams`
4. **Follow domain-driven design**: domain → repository → service
5. **Type-safe DI**: Use `service.Cached[T]` and `service.LazyLoad[T]`

## When Suggesting Code

1. **Ask about scale first:**

   - Small → Router mode
   - Medium → Framework mode with DDD
   - Large → Enterprise modular

2. **Provide complete code:**

   - Include imports
   - Include error handling
   - Include validation tags
   - Include config.yaml if using framework mode

3. **Follow project structure:**
   - Separate domain/repository/service
   - Use interfaces in domain layer
   - Implement in repository/service layers

## Resources

- AI Agent Guide: [AI-AGENT-GUIDE.md](../docs/AI-AGENT-GUIDE.md) - **READ THIS FIRST**
- Quick Reference: [QUICK-REFERENCE.md](../docs/QUICK-REFERENCE.md)
- Full Docs: https://primadi.github.io/lokstra/
- Examples: https://primadi.github.io/lokstra/00-introduction/examples/
- Templates: [project_templates/](../project_templates/)
