---
layout: docs
title: 02 – Framework Guide
description: Using Lokstra as a full application framework (like NestJS / Spring Boot).
---

## Big Picture

Track 2 turns Lokstra into a **business application framework**:

- `lokstra_init.BootstrapAndRun()` for bootstrapping.
- YAML config (`configs`, `dbpool-manager`, `deployments`/`servers`).
- Annotation-based services: `@RouterService`, `@Inject`, `@Route`.

We will use **`03_tenant_management`** as the hero example.

---

## 1. Bootstrap with `lokstra_init`

`main.go`:

```go
recovery.Register()
request_logger.Register()

lokstra_init.BootstrapAndRun()
```

What `BootstrapAndRun()` does (simplified):

- Loads config (by default from `./config` next to the binary).  
- Initializes registries, db pools, and annotation-generated services.  
- Starts all servers defined in the config.

If you need initialization that depends on config values, use:

```go
lokstra_init.BootstrapAndRun(
    lokstra_init.WithServerInitFunc(func() error {
        // custom init that can read configs
        return nil
    }),
)
```

---

## 2. YAML Config (Tenant Example)

`config/config.yaml`:

```yaml
configs:
  store:
    tenant-store: postgres-tenant-store

dbpool-manager:
  db_auth:
    dsn: ${GLOBAL_DB_DSN:postgres://postgres:adm1n@localhost:5432/lokstra_db}
    schema: ${GLOBAL_DB_SCHEMA:lokstra_auth}

servers:
  api-server:
    base-url: "http://localhost"
    addr: ":3000"
    published-services: [tenant-service]
```

Key ideas:

- **`configs.store.tenant-store`** – logical name `tenant-store` → implementation `postgres-tenant-store`.  
- **`dbpool-manager.db_auth`** – defines a lazy DbPool service named `db_auth`.  
- **`servers.api-server.published-services`** – tells Lokstra to expose `tenant-service` over HTTP on `:3000`.

Lokstra supports multiple YAML files (e.g. `deployment.yaml`, `user.yaml`, `order.yaml`);
they are merged based on deployment/server names.

---

## 3. Annotations: Router + DI

`application/tenant_service.go`:

```go
// @RouterService name="tenant-service",
//   prefix="${api-auth-prefix:/api/auth}/core/tenants",
//   middlewares=["recovery", "request_logger"]
type TenantService struct {
    // @Inject "@store.tenant-store"
    Store repository.TenantStore
}
```

How it works:

- `@RouterService`:
  - Registers a service named `tenant-service`.
  - Generates a router with base path `/api/auth/core/tenants`
    (overridable via `${api-auth-prefix:...}`).
  - Attaches middlewares `recovery` and `request_logger`.

- `@Inject "@store.tenant-store"`:
  - `@store.<key>` looks up `configs.store.<key>` in YAML.
  - Here `configs.store.tenant-store = "postgres-tenant-store"`,
    so it resolves to `@Inject "postgres-tenant-store"`.

`repository/tenant_store.go`:

```go
type PostgresTenantStore struct {
    // @Inject "db_auth"
    dbPool serviceapi.DbPool
}
```

- `db_auth` comes from `dbpool-manager.db_auth`, so Lokstra injects the
  configured `DbPool` into `dbPool`.

---

## 4. Routes via `@Route`

Still in `tenant_service.go`:

```go
// @Route "POST /"
func (s *TenantService) CreateTenant(ctx *request.Context,
    req *domain.CreateTenantRequest) (*domain.Tenant, error) { ... }

// @Route "GET /{id}"
func (s *TenantService) GetTenant(ctx *request.Context,
    req *domain.GetTenantRequest) (*domain.Tenant, error) { ... }
```

Each `@Route`:

- Defines HTTP method + relative path under the `@RouterService` prefix.  
- Automatically binds path/query/body to the `req` struct (with validation).  
- Serializes the return value to JSON and maps errors to HTTP status codes.

The generated `zz_generated.lokstra.go` file wires all of this together.

---

## 5. When to Use Track 2

Use Lokstra as an application framework when:

- You would normally reach for **NestJS / Spring Boot / Fx**.  
- You have **non-trivial domain logic** and multiple services.  
- You care about **deployment topology** (monolith now, microservices later).  
- You like the idea of:
  - Declarative services in YAML.  
  - Annotation-based routers.  
  - Type-safe DI with minimal boilerplate.

If your service is mostly simple handlers, stay with Track 1 (Router) for now
and move to Track 2 when the domain and deployment start to get complex.


