# Lokstra Framework - Overview & Philosophy

**Purpose:** Provide AI Agents with foundational understanding of Lokstra Framework architecture, design principles, and decision-making guidelines.

---

## Framework Identity

Lokstra is a **Go web framework** with dual operational modes:

1. **Router Mode** - Lightweight HTTP routing (similar to Echo, Gin, Chi)
2. **Framework Mode** - Full dependency injection framework (similar to NestJS, Spring Boot)

**Key Feature:** Annotation-based code generation (`@Handler`, `@Service`, `@Route`, `@Inject`) eliminates boilerplate while maintaining type safety.

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

**Universal Pattern (simple or enterprise):**
```
myapp/
├── .github/
│   ├── skills/                    # AI Agent instructions (auto-copied)
│   └── copilot-instructions.md
├── config.yaml
├── docs/
│   ├── BRD.md                     # Business requirements
│   └── templates/                  # Document templates
└── modules/
    ├── shared/
    │   └── domain/                 # Cross-module domain models
    └── <module_name>/
        ├── handler/                # @Handler (1 file per entity)
        │   └── user_handler.go
        ├── repository/             # Data access (1 file per entity)
        │   └── user_repository.go
        └── domain/                 # Business logic
            ├── user.go
            └── user_service.go
```

**Rules:**
- Always use `/modules` structure (even for 1-2 entities)
- 1 file per entity (no `handlers.go` with multiple entities)
- DTOs live in `domain/` layer with `validate` tags
- Infrastructure services use `@Service` annotation

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
- **Full Docs:** https://primadi.github.io/lokstra/
- **Templates:** [docs/templates/](../../docs/templates/)
- **Examples:** https://primadi.github.io/lokstra/00-introduction/examples/

---

**Next:** Read [01-document-workflow.md](01-document-workflow.md) for BRD creation and document-driven development process.
