# GitHub Copilot Instructions for Lokstra Framework

This file provides quick syntax reference for Lokstra Framework development.

**For comprehensive guides:** Use AI skills in `.github/skills/` or run `lokstra update-skills`

## Quick Reference

### Annotations Syntax

```
@Handler name="service-name", prefix="/api/path"
@Service "service-name"
@Route "METHOD /path", middlewares=["auth"]
@Inject "service-name" or "cfg:config.key" or "@config.impl"
```

### Injection Patterns

| Pattern         | Syntax                   | Example                      |
| --------------- | ------------------------ | ---------------------------- |
| Direct service  | `@Inject "service-name"` | `@Inject "user-repo"`        |
| From config     | `@Inject "@config.key"`  | `@Inject "@repository.impl"` |
| Config value    | `@Inject "cfg:key"`      | `@Inject "cfg:app.timeout"`  |
| Indirect config | `@Inject "cfg:@key"`     | `@Inject "cfg:@jwt.path"`    |

### Handler Signatures (29+ supported)

```go
func() string { return "Hello" }
func(id string) (*User, error) { /* ... */ }
func(ctx *request.Context, params *CreateUserParams) error { /* ... */ }
func(ctx *request.Context, id string, params *UpdateParams) error { /* ... */ }
```

### Common Imports

```go
import "github.com/primadi/lokstra"
import "github.com/primadi/lokstra/core/request"
import "github.com/primadi/lokstra/core/service"
import "github.com/primadi/lokstra/lokstra_registry"
```

### Response Helpers

```go
ctx.Api.Ok(data)                    // 200
ctx.Api.Created(data)               // 201
ctx.Api.BadRequest("message")       // 400
ctx.Api.Unauthorized("message")     // 401
ctx.Api.NotFound("message")         // 404
ctx.Api.InternalServerError("msg")  // 500
```

### DTO Validation Tags

```go
type CreateUserParams struct {
    Name  string `json:"name" validate:"required,min=3,max=50"`
    Email string `json:"email" validate:"required,email"`
}
```

## CLI Commands

```bash
lokstra new myapp                    # Create new project
lokstra update-skills                # Download latest skills
lokstra autogen .                    # Manual code generation
go run . --generate-only             # Force rebuild
```

---

## For Complete Guidance

This file provides quick syntax reference only. For comprehensive guides on:

- **Framework fundamentals & philosophy** → Use `lokstra-overview` skill
- **Complete project setup** → Use `lokstra-brd-generation` skill
- **Building handlers, services, configs** → Use relevant Phase 2 implementation skills
- **Testing & validation** → Use Phase 3 advanced skills

Run `lokstra update-skills` in your project to download all skills locally.
