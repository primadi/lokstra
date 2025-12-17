---
layout: docs
title: Quick Start
description: Minimal steps to run Lokstra in Router and Framework modes.
---

## 1. Install

```bash
go get github.com/primadi/lokstra
go install github.com/primadi/lokstra/cmd/lokstra@latest
```

---

## 2. Track 1 – Router (like Gin/Echo)

Run the router‑only example:

```bash
go run ./project_templates/01_router/01_router_only
```

Open in browser:

- `http://localhost:3000/users`  
- `http://localhost:3000/roles`

Or use the included `test.http` file with VS Code REST Client / Cursor.

Key code (simplified):

```go
r := lokstra.NewRouter("demo_router")

r.Use(recovery.Middleware(recovery.DefaultConfig()))

users := r.AddGroup("/users")
users.GET("", handleGetUsers)
users.POST("", handleCreateUser)

http.ListenAndServe(":3000", r)
```

---

## 3. Track 2 – Framework (Tenant Service)

Run the tenant management example:

```bash
go run ./project_templates/02_app_framework/03_tenant_management
```

Requirements:

- PostgreSQL running with a DSN compatible with `config/config.yaml`.

Key pieces:

- `main.go`:

```go
recovery.Register()
request_logger.Register()
lokstra_init.BootstrapAndRun()
```

- `config/config.yaml`:

```yaml
servers:
  api-server:
    base-url: "http://localhost"
    addr: ":3000"
    published-services: [tenant-service]
```

Test with `tenant-service.http` – create, get, list, update, delete tenants.

Once you are comfortable with these two examples, dive into:

- [01 – Router Guide](../01-router-guide/)  
- [02 – Framework Guide](../02-framework-guide/)


