---
name: lokstra-overview
description: Provides foundational understanding of Lokstra Framework architecture, design principles, annotation-based code generation, and decision-making guidelines for AI agents building Go web applications. Use when starting any Lokstra project or when agents need framework context.
license: MIT
metadata:
  author: lokstra-framework
  version: "1.0"
  framework: lokstra
  framework-version: "0.1.x"
  phase: design
  order: 1
compatibility: Designed for GitHub Copilot, Cursor, Claude Code, and similar AI coding assistants
---

# Lokstra Framework - Overview & Philosophy

## When to use this skill

Use this skill when:
- Starting any new Lokstra project with AI agents
- Need to understand annotation system (`@Handler`, `@Service`, `@Route`, `@Inject`)
- Understanding code generation and dependency injection
- Determining module structure (DDD bounded contexts)
- Need framework-specific best practices

## Framework Mode Overview

Lokstra is a **Go web framework** with annotation-based code generation.

**Key Features:**
- Annotation-based code generation (`@Handler`, `@Service`, `@Route`, `@Inject`)
- Full dependency injection (like NestJS, Spring Boot)
- Type-safe with compile-time verification
- Modular architecture (DDD bounded contexts)
- Auto-generates service registration, routes, middleware

---

## Core Design Philosophy

### 1. Document-Driven Development

**Problem:** Code-first approaches lead to inconsistencies, rework cycles, and technical debt.

**Solution:** Enforce design-first workflow:
```
BRD → Module Requirements → API Spec → Schema → Implementation
```

**Benefits:**
- 10x productivity improvement (estimated)
- Eliminates "code-then-fix" cycles
- Built-in compliance documentation
- Stakeholder alignment before coding

### 2. Bounded Context (DDD)

**Module Granularity:**
- ❌ Wrong: Monolithic `user` module handling auth + profile + notifications
- ✅ Right: Separate modules - `auth`, `user_profile`, `notification`

**Cross-Module Communication:**
- Via repository injection (not direct service calls)
- Shared domain models in `/modules/shared/domain/`
- Event-driven for async operations

### 3. Project Structure Consistency

**Universal Pattern:**
```
myapp/
├── .github/
│   ├── skills/                    # AI Agent instructions (auto-copied)
│   └── copilot-instructions.md
├── configs/                       # YAML config files (auto-merged)
│   ├── config.yaml
│   ├── database-config.yaml
│   └── <module-name>-config.yaml
├── migrations/                    # Database migrations per module
│   ├── shared/
│   │   ├── 001_init.up.sql
│   │   └── 001_init.down.sql
│   └── <module_name>/
│       ├── 001_create_table.up.sql
│       └── 001_create_table.down.sql
├── docs/
│   ├── drafts/                    # BRD/requirements draft versions
│   └── modules/                   # Published BRD/requirements
│       └── <module_name>/
│           ├── BRD-*.md
│           ├── REQUIREMENTS-*.md
│           ├── API_SPEC.md
│           └── SCHEMA.md
└── modules/
    ├── shared/
    │   └── domain/                # Cross-module models
    └── <module_name>/
        ├── handler/               # @Handler files (1 per entity)
        │   └── user_handler.go
        ├── repository/            # Data access (1 per entity)
        │   └── user_repository.go
        └── domain/                # Business logic & DTOs
            ├── user.go
            └── user_dto.go
```

**Key Rules:**
- Multi-file YAML config in `/configs/` - all files auto-merged
- Migrations organized by module in `/migrations/{module}/`
- 1 handler file per entity (user_handler.go, not handlers.go)
- 1 repository file per entity
- DTOs with `validate` tags in domain layer

---

## Annotation System

### @Handler (Business Services)

```go
// @Handler name="user-handler", prefix="/api/users"
type UserHandler struct {
    // @Inject "user-repository"
    UserRepo UserRepository
    
    // @Inject "@repository.implementation"  // From config
    Repository Repository
    
    // @Inject "cfg:app.timeout"             // Config value
    Timeout time.Duration
}

// @Route "GET /{id}"
func (h *UserHandler) GetByID(id string) (*User, error) {
    return h.UserRepo.GetByID(id)
}

// @Route "POST /", middlewares=["auth"]
func (h *UserHandler) Create(req *CreateUserRequest) (*User, error) {
    // Auto-validated via struct tags
}
```

### @Service (Infrastructure Services)

```go
// @Service "postgres-user-repository"
type PostgresUserRepository struct {
    // @Inject "db-pool"
    DB *sql.DB
}

func (r *PostgresUserRepository) GetByID(id string) (*User, error) {
    // Implementation
}
```

### Injection Patterns

| Pattern                     | Syntax                  | Example                      |
| --------------------------- | ----------------------- | ---------------------------- |
| Direct service              | `@Inject "service-name"`| `@Inject "user-repository"`  |
| Service from config         | `@Inject "@config.key"` | `@Inject "@repository.impl"` |
| Config value                | `@Inject "cfg:key"`     | `@Inject "cfg:app.timeout"`  |
| Indirect config (reference) | `@Inject "cfg:@key"`    | `@Inject "cfg:@jwt.path"`    |

---

## Code Generation

**Automatic (Development):**
```go
func main() {
    lokstra.Bootstrap()  // Auto-generates on @Handler changes
    _ "myapp/modules/user/handler"  // Import triggers scan
    lokstra_registry.RunServerFromConfig()
}
```

**Manual (Before Build/Deploy):**
```bash
lokstra autogen .        # Scan and generate
go run . --generate-only # Force rebuild all
```

**Generated Files:**
- `zz_generated.lokstra.go` in same package as `@Handler`
- `zz_lokstra_imports.go` at project root
- `zz_cache.lokstra.json` for metadata tracking
- Registers services, routes, dependencies

---

## Configuration (config.yaml)

```yaml
configs:
  # Service selection (interface injection)
  repository:
    implementation: "postgres-user-repository"  # Switch here!
  
  # Application settings
  app:
    timeout: "30s"
    jwt-secret: "super-secret"

# Service definitions
service-definitions:
  postgres-user-repository:
    type: postgres-user-repository
    depends-on: [db-pool]

# Deployment
deployments:
  development:
    servers:
      api:
        addr: ":8080"
        published-services: [user-handler]
```

---

## Handler Signatures (29+ Supported)

```go
// Simple
func() string

// With error
func(id string) (*User, error)

// Request body (auto-validated)
func(req *CreateUserRequest) (*User, error)

// Context + body
func(ctx *request.Context, req *CreateUserRequest) error {
    return ctx.Api.Created(user)
}

// Path param + body
func(ctx *request.Context, id string, req *UpdateRequest) error
```

---

## Response Helpers

```go
ctx.Api.Ok(data)                    // 200
ctx.Api.Created(data)               // 201
ctx.Api.BadRequest("message")       // 400
ctx.Api.Unauthorized("message")     // 401
ctx.Api.NotFound("message")         // 404
ctx.Api.InternalServerError("msg")  // 500
```

---

## Decision Guidelines for AI Agents

### When to use Framework Mode:
- 2+ entities/modules
- Dependency injection needed
- Interface-based design (swappable implementations)
- Enterprise requirements (testing, modularity)

### When to use Router Mode:
- Learning/prototyping
- Single-file APIs
- No DI requirements
- < 5 endpoints

### Module Granularity:
- Ask: "Can this be a separate microservice?" → If yes, separate module
- Example: Auth, User Profile, Notification = 3 modules (not 1 User module)

### File Organization:
- 1 entity = 1 handler file + 1 repository file + 1 domain file
- No `handlers.go` with 10+ entities
- DTOs in domain layer, not in handler

---

## Best Practices Checklist

- ✅ Always include error handling
- ✅ Use validation tags: `validate:"required,email"`
- ✅ Use pointer parameters: `*CreateUserRequest`
- ✅ Follow DDD: domain → repository → handler
- ✅ Type-safe DI: Direct type assertions
- ✅ Prefer annotations over manual registration
- ✅ 1 file per entity (maintainability)
- ✅ Document-driven: BRD → specs → code

---

## Common Imports

```go
import "github.com/primadi/lokstra"
import "github.com/primadi/lokstra/core/request"
import "github.com/primadi/lokstra/core/service"
import "github.com/primadi/lokstra/lokstra_registry"
import "github.com/primadi/lokstra/middleware/recovery"
import "github.com/primadi/lokstra/middleware/cors"
```

---

## Resources

- **AI Agent Guide:** [AI-AGENT-GUIDE.md](https://primadi.github.io/lokstra/AI-AGENT-GUIDE)
- **Quick Reference:** [QUICK-REFERENCE.md](https://primadi.github.io/lokstra/QUICK-REFERENCE)
- **Full Documentation:** https://primadi.github.io/lokstra/
- **Examples:** https://primadi.github.io/lokstra/00-introduction/examples/

---

## Next Steps

After understanding the framework overview:
1. Use `lokstra-brd-generation` skill to create Business Requirements
2. Use `lokstra-module-requirements` skill to define modules
3. Use `lokstra-api-specification` skill to design APIs
4. Use `lokstra-schema-design` skill to create database schema
5. Use `lokstra-code-implementation` and `lokstra-code-advanced` skills to generate code
