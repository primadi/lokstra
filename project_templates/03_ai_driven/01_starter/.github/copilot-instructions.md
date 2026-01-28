# GitHub Copilot Instructions for Lokstra Framework

This file provides instructions to GitHub Copilot when working with the Lokstra Framework.

## Framework Overview

Lokstra is a Go web framework with two modes:

1. **Router Mode** - Simple HTTP routing (like Echo, Gin, Chi)
2. **Framework Mode** - Full DI framework (like NestJS, Spring Boot)

## Key Documentation

When helping with Lokstra:

- **Complete AI Guide:** [AI-AGENT-GUIDE.md](https://primadi.github.io/lokstra/AI-AGENT-GUIDE)
- **Quick Reference:** [QUICK-REFERENCE.md](https://primadi.github.io/lokstra/QUICK-REFERENCE)
- **Full Documentation:** https://primadi.github.io/lokstra/

## Common Patterns

### 1. Simple Router (No DI)

```go
r := lokstra.NewRouter("api")
r.GET("/users", func() []User { return getUsers() })
app := lokstra.NewApp("myapp", ":8080", r)
app.Run(30 * time.Second)
```

### 2. Framework Mode with Annotations (Recommended)

**Use annotations for business services - minimal boilerplate:**

```go
// main.go
func main() {
	// TODO: register any middleware, services, or database pools if needed
	recovery.Register()
	request_logger.Register()
	dbpool_pg.Register()

	// Auto-generate code from @Handler, @Service annotations
	// This will detect changes in files with @Handler, @Service and regenerate code
	lokstra_init.BootstrapAndRun()
}
```

**Service with annotations:**

```go
// @Handler name="user-service", prefix="/api/users"
type UserService struct {
    // @Inject "user-repository"               - Direct service injection
    UserRepo UserRepository

    // @Inject "cfg:app.name"                  - Config value injection
    AppName string
}

// @Route "GET /{id}"
func (s *UserService) GetByID(ctx *request.Context, p *GetUserParams) (*User, error) {
    return s.UserRepo.GetByID(p.ID)
}

// @Route "POST /", ["auth"]
func (s *UserService) Create(ctx *request.Context, p *CreateUserParams) (*User, error) {
    // ...
}

// @Route "DELETE /{id}", ["auth", "admin"]
func (s *UserService) Delete(ctx *request.Context, p *DeleteUserParams) error {
    // ...
}
```

**Code generation (automatic):**

```go
func main() {
    lokstra.Bootstrap() // Auto-generates when changes detected
    // ...
}
```

**Manual generation (before build/deploy):**

```bash
lokstra autogen .        # Manual generation
go run . --generate-only # Force rebuild all (optional)
```

**Per-route middleware:** Add `middlewares=["mw1", "mw2"]` to `@Route`

### 3. Interface Injection Pattern (Config-Based Selection)

**Use case:** Multiple implementations of an interface, selectable via config.

**Domain interface:**

```go
// domain/repository.go
type Repository interface {
    GetUser(id string) (*User, error)
    SaveUser(user *User) error
}
```

**Implementations with @Service:**

```go
// infrastructure/postgres_repository.go
// @Service "postgres-repository"
type PostgresRepository struct {
    // @Inject "db-pool"
    DB *sql.DB
}

var _ Repository = (*PostgresRepository)(nil)

func (s *PostgresRepository) GetUser(id string) (*User, error) { /* ... */ }

// infrastructure/mysql_repository.go
// @Service "mysql-repository"
type MySQLRepository struct {
    // @Inject "db-pool"
    DB *sql.DB
}

var _ Repository = (*MySQLRepository)(nil)

func (s *MySQLRepository) GetUser(id string) (*User, error) { /* ... */ }
```

**Business service using config-based injection:**

```go
// application/user_service.go
// @Handler name="user-service", prefix="/api/users"
type UserService struct {
    // @Inject "@repository.implementation"
    Repository Repository  // Actual service injected based on config!

    // @Inject "cfg:app.timeout"  // Config value injection
    Timeout time.Duration

    // @Inject "cfg:@jwt.key-path"  // Indirect config
    JWTSecret string  // Key resolved from another config value
}

// @Route "GET /{id}"
func (s *UserService) GetUser(id string) (*User, error) {
    return s.Repository.GetUser(id)
}
```

**config.yaml:**

```yaml
configs:
  repository:
    implementation: "postgres-repository" # Switch to "mysql-repository" here!

  app:
    timeout: "30s"  # Direct config value

  jwt:
    key-path: "app.production-jwt-secret"  # Points to actual key

  app:
    production-jwt-secret: "super-secret-key"

service-definitions:
  postgres-repository:
    type: postgres-repository

  mysql-repository:
    type: mysql-repository

deployments:
  development:
    servers:
      api:
        addr: ":8080"
        published-services: [user-service]
```

**Injection Patterns Summary:**

| Annotation                  | Syntax              | Purpose                        | Example                      |
| --------------------------- | ------------------- | ------------------------------ | ---------------------------- |
| `@Inject "service-name"`    | Direct service      | Inject service                 | `@Inject "user-repo"`        |
| `@Inject "@config.key"`     | Service from config | Service name from config       | `@Inject "@repository.impl"` |
| `@Inject "cfg:config.key"`  | Config value        | Config value injection         | `@Inject "cfg:app.timeout"`  |
| `@Inject "cfg:@config.key"` | Indirect config     | Config key from another config | `@Inject "cfg:@jwt.path"`    |

**Benefits:**

- Switch implementation by changing ONE line in config
- No code changes needed
- Type-safe (compile-time interface checking)
- Perfect for: database drivers, cache providers, storage backends

### 4. Manual Registration (Advanced/Infrastructure Services)

**For infrastructure services or custom factories:**

```go
func UserServiceFactory(deps map[string]any, config map[string]any) any {
    // Access global config from YAML
    dbDSN := lokstra_registry.GetConfig("database.dsn", "postgres://localhost/mydb")

    return &UserService{
        UserRepo: deps["user-repository"].(UserRepository),
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
# Global config (optional)
configs:
  database:
    dsn: "postgres://localhost:5432/mydb"

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

### 5. Handler Signatures (29+ supported)

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

### 6. Domain Models

```go
type CreateUserParams struct {
    Name  string `json:"name" validate:"required,min=3,max=50"`
    Email string `json:"email" validate:"required,email"`
}
```

## Project Structure

### Simple (Router Only)

```
myapp/
├── main.go
└── handlers.go
```

### Medium (DDD with Annotations)

```
myapp/
├── main.go
├── config.yaml
├── domain/
│   └── user/
│       ├── models.go
│       ├── repository.go
│       └── service.go
├── infrastructure/
│   └── user_repository.go
└── application/
    ├── user_service.go              # Contains @Handler
    └── zz_generated.lokstra.go      # Auto-generated
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
lokstra autogen .        # Manual generation
go run . --generate-only # Force rebuild all

# Recommended: Use lokstra.Bootstrap() in main() for auto-generation
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
5. **Type-safe DI**: Use direct type assertions and `service.LazyLoad[T]` for lazy service loading
6. **Prefer annotations** for business services: Use `@Handler` + `@Route` instead of manual registration

## When Suggesting Code

1. **Ask about scale first:**
   - Small → Router mode
   - Medium → Framework mode with annotations
   - Large → Enterprise modular with annotations

2. **Provide complete code:**
   - Include imports
   - Include error handling
   - Include validation tags
   - Include config.yaml if using framework mode
   - Use `@Handler` annotations for business services

3. **Follow project structure:**
   - Separate domain/application/infrastructure
   - Use interfaces in domain layer
   - Business logic in application layer with `@Handler`
   - Data access in infrastructure layer

## Resources

- AI Agent Guide: [AI-AGENT-GUIDE.md](https://primadi.github.io/lokstra/AI-AGENT-GUIDE) - **READ THIS FIRST**
- Quick Reference: [QUICK-REFERENCE.md](https://primadi.github.io/lokstra/QUICK-REFERENCE)
- Full Docs: https://primadi.github.io/lokstra/
- Examples: https://primadi.github.io/lokstra/00-introduction/examples/
- Templates: [project_templates/](https://github.com/primadi/lokstra/tree/main/project_templates)
