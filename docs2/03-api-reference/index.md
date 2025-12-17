---
layout: docs
title: 03 – API Reference (Cheatsheet)
description: Quick reference for Lokstra essentials: init, YAML keys, annotations.
---

## lokstra_init Cheatsheet

- **Basic bootstrap**

```go
lokstra_init.BootstrapAndRun()
```

- **With options**

```go
lokstra_init.BootstrapAndRun(
    lokstra_init.WithAnnotations(true),
    lokstra_init.WithYAMLConfigPath(true, "config"),
    lokstra_init.WithDbPoolManager(true, true),
    lokstra_init.WithDbMigrations(true, "migrations"),
    lokstra_init.WithServerInitFunc(func() error { return nil }),
)
```

## YAML Config Keys (Core)

```yaml
configs:
  store:
    tenant-store: postgres-tenant-store

dbpool-manager:
  db_auth:
    dsn: postgres://...
    schema: lokstra_auth

deployments:        # (enterprise-style configs)
  development:
    servers:
      api-server:
        base-url: http://localhost
        addr: :3000
        published-services: [tenant-service]

servers:            # (simple configs)
  api-server:
    base-url: http://localhost
    addr: :3000
    published-services: [tenant-service]
```

**Rules:**

- Each `dbpool-manager.<name>` → lazy `DbPool` service with name `<name>`.  
- Each `published-services` entry → a service defined by `@RouterService`.  
- `configs.store.<key>` can be referenced via `@store.<key>` in `@Inject`.

## Annotations Cheatsheet

- **Router service**

```go
// @RouterService name="tenant-service",
//   prefix="/api/auth/core/tenants",
//   middlewares=["recovery", "request_logger"]
type TenantService struct { ... }
```

- **Inject by name**

```go
// direct name
// @Inject "db_auth"

// via configs.store.<key>
// @Inject "@store.tenant-store"
```

- **Route**

```go
// @Route "GET /{id}"
func (s *TenantService) GetTenant(
    ctx *request.Context,
    req *domain.GetTenantRequest,
) (*domain.Tenant, error) { ... }
```

## CLI Commands

- Install:

```bash
go install github.com/primadi/lokstra/cmd/lokstra@latest
```

- Create project:

```bash
lokstra new myapp
lokstra new myapp -template 01_router/01_router_only
lokstra new myapp -template 02_app_framework/03_tenant_management
```

- Generate code from annotations:

```bash
lokstra autogen ./myproject
```


