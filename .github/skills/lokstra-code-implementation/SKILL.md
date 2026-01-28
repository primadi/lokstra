---
name: lokstra-code-implementation
description: Generate Lokstra module code from approved specifications. Creates domain models, repositories, handlers with @Handler annotations, and basic tests. Use after all specs (BRD, requirements, API, schema) are approved. SKILL 4-7 from original implementation guide.
license: MIT
metadata:
  author: lokstra-framework
  version: "1.0"
  framework: lokstra
  skill-level: basic
compatibility: Designed for GitHub Copilot, Cursor, Claude Code
---

# Lokstra Code Implementation (Basic)

## When to use this skill

Use this skill when:
- All specifications approved (BRD, requirements, API spec, schema)
- Ready to generate working code
- Creating: migrations, domain models, repositories, handlers
- Need basic implementation (SKILL 4-7)

## Implementation Order

```
1. Migrations → 2. Domain Models → 3. Repository → 4. Handler
```

## SKILL 4: Module Structure

Create folder structure:
```
modules/{module-name}/
├── domain/
│   ├── {entity}.go          # Domain model
│   ├── {entity}_dto.go      # Request/Response DTOs
│   └── {entity}_service.go  # Business logic (optional)
├── repository/
│   └── {entity}_repository.go
└── handler/
    └── {entity}_handler.go
```

## SKILL 5: Database Migrations

From SCHEMA.md, create migration files:
```sql
-- migrations/{module}/001_create_{table}.sql
-- Migration: Create {table} table
-- UP
CREATE TABLE {table_name} (...);
CREATE INDEX...;
CREATE TRIGGER...;

-- DOWN
DROP TABLE IF EXISTS {table_name} CASCADE;
```

## SKILL 6: Domain Models

```go
// modules/{module}/domain/{entity}.go
package domain

type {Entity} struct {
    ID        string    `json:"id" db:"id"`
    Name      string    `json:"name" db:"name" validate:"required,min=3,max=50"`
    Email     string    `json:"email" db:"email" validate:"required,email"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Request DTO
type Create{Entity}Request struct {
    Name  string `json:"name" validate:"required,min=3,max=50"`
    Email string `json:"email" validate:"required,email"`
}

// Response DTO
type {Entity}Response struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}
```

## SKILL 7: Repository

```go
// modules/{module}/repository/{entity}_repository.go
package repository

type {Entity}Repository interface {
    GetByID(ctx context.Context, id string) (*domain.{Entity}, error)
    Create(ctx context.Context, entity *domain.{Entity}) error
    Update(ctx context.Context, entity *domain.{Entity}) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, limit, offset int) ([]*domain.{Entity}, error)
}

// @Service "{entity}-repository"
type {Entity}RepositoryImpl struct {
    // @Inject "db-pool"
    DB serviceapi.DbPool
}

func (r *{Entity}RepositoryImpl) GetByID(ctx context.Context, id string) (*domain.{Entity}, error) {
    var entity domain.{Entity}
    query := `SELECT * FROM {table_name} WHERE id = $1 AND deleted_at IS NULL`
    
    err := r.DB.QueryRow(ctx, query, id).Scan(
        &entity.ID, &entity.Name, &entity.Email, 
        &entity.CreatedAt, &entity.UpdatedAt,
    )
    if err == sql.ErrNoRows {
        return nil, errors.New("not found")
    }
    return &entity, err
}

// Implement other methods...
```

## Handler with @Handler

```go
// modules/{module}/handler/{entity}_handler.go
package handler

// @Handler name="{entity}-handler", prefix="/api/{module}"
type {Entity}Handler struct {
    // @Inject "{entity}-repository"
    Repo repository.{Entity}Repository
}

// @Route "GET /{id}"
func (h *{Entity}Handler) GetByID(id string) (*domain.{Entity}Response, error) {
    entity, err := h.Repo.GetByID(context.Background(), id)
    if err != nil {
        return nil, err
    }
    return &domain.{Entity}Response{
        ID:        entity.ID,
        Name:      entity.Name,
        Email:     entity.Email,
        CreatedAt: entity.CreatedAt,
    }, nil
}

// @Route "POST /", middlewares=["auth"]
func (h *{Entity}Handler) Create(req *domain.Create{Entity}Request) (*domain.{Entity}Response, error) {
    entity := &domain.{Entity}{
        ID:    uuid.New().String(),
        Name:  req.Name,
        Email: req.Email,
    }
    
    if err := h.Repo.Create(context.Background(), entity); err != nil {
        return nil, err
    }
    
    return &domain.{Entity}Response{
        ID:    entity.ID,
        Name:  entity.Name,
        Email: entity.Email,
    }, nil
}
```

## Update main.go

```go
import (
    _ "{project}/modules/{module}/handler"  // Import to trigger code generation
)
```

## Update config.yaml

```yaml
service-definitions:
  {entity}-repository:
    type: {entity}-repository
    depends-on: [db-pool]
  
  {entity}-handler:
    type: {entity}-handler
    depends-on: [{entity}-repository]

deployments:
  development:
    servers:
      api:
        addr: ":8080"
        published-services: [{entity}-handler]
```

## Generate Code

```bash
# Auto-generate from annotations
go run . --generate-only

# Or use CLI
lokstra autogen .
```

## Resources

For advanced implementation (tests, config, consistency checks):
- Use `lokstra-code-advanced` skill for SKILL 8-13
