
# Bootstrap Project

A practical quickstart for **Lokstra 0.2.1**, based on the real APIs in the repo (`lokstra.go`, `core/config/*`, `core/server/*`, `core/router/*`). It shows **two paths**:

1) **Code‑first** — now documented in **two styles**:
   - **Manual binding** (use `BindXXX` / `BindAllSmart`)
   - **Auto binding** (generic handler: `func(ctx *request.Context, params *T) error`)
2) **Config‑first** (YAML + code loader)

Both end with a production‑style server that supports routes, static files, HTMX pages, reverse proxy, and RPC.

---

## 0) Prereqs

- Go 1.21+
- Module import path: `github.com/primadi/lokstra`

Initialize a module:

```bash
mkdir hello-lokstra && cd hello-lokstra
go mod init example.com/hello
go get github.com/primadi/lokstra@v0.2.1
```

> If you vendor or use a monorepo, just make sure imports point to `github.com/primadi/lokstra`.

---

## 1) Code‑first (two styles)

### A) Manual binding (BindAllSmart)

```go
package main

import (
    "time"
    "github.com/primadi/lokstra"
    request "github.com/primadi/lokstra/core/request"
)

type HelloReq struct {
    Name string ` + "`query:"name"`" + `
}

func hello(ctx *request.Context) error {
    var req HelloReq
    if err := ctx.BindAllSmart(&req); err != nil {
        // IMPORTANT: return a *handled* error to avoid 500
        return ctx.ErrorBadRequest(err.Error())
    }
    if req.Name == "" { req.Name = "World" }
    return ctx.Ok(map[string]any{"message": "Hello " + req.Name})
}

func main() {
    reg := lokstra.NewGlobalRegistrationContext()

    svr := lokstra.NewServer(reg, "hello-server")
    app := lokstra.NewApp(reg, "web-app").WithAddress(":8080")

    app.GET("/api/hello", hello)

    svr.AddApp(app)
    _ = svr.StartAndWaitForShutdown(5 * time.Second)
}
```

> Cheatsheet:  
> - `BindBody` → JSON only, **or** use `BindBodySmart` (auto-detect content type).  
> - `BindAll` → bind path+query+header+body (JSON)  
> - `BindAllSmart` → bind path+query+header+body **smart** (content‑type aware).  
> - Tags are **required** for auto‑bind: `path:""`, `query:""`, `header:""`, `body:""` (no fallback).

### B) Auto binding (generic handler param)

Let Lokstra bind your parameters **automatically** by declaring the handler as:

```go
func hello(ctx *request.Context, req *HelloReq) error {
    if req.Name == "" { req.Name = "World" }
    return ctx.Ok(map[string]any{"message": "Hello " + req.Name})
}
```

where:

```go
type HelloReq struct {
    Name string ` + "`query:"name"`" + ` // from URL query
}
```

You can mix multiple sources in one struct with tags:

```go
type UserRequest struct {
    ID    string ` + "`path:"id"`" + `
    Token string ` + "`header:"Authorization"`" + `
    Name  string ` + "`body:"name"`" + `
    Q     string ` + "`query:"q"`" + `
}

func createUser(ctx *request.Context, req *UserRequest) error {
    // No manual BindXXX needed; Lokstra bound `req` already.
    return ctx.Ok("Created user " + req.Name)
}

// Route example:
app.POST("/users/:id", createUser)
```

> Notes:  
> - If binding fails, the framework will surface the error; when doing manual binding, always wrap with `ctx.ErrorBadRequest(...)`.  
> - `ctx.Ok(...)` returns **nil** (success already written).  
> - Keep tags accurate; no tag → field won’t be auto‑bound.

---

## 2) Config‑first (YAML + loader)

Create a **config/** folder. Lokstra can load **a single file** or **merge all files** in a directory (`core/config/config_loader.go`). It also supports `${VAR}` expansion.

**config/app.yaml**

```yaml
server:
  name: hello-server
  global_setting:
    log_level: debug           # affects default logger service
    # optional flow defaults (if you use flow package):
    # flow_logger: logger
    # flow_dbschema: public

apps:
  - name: web-app
    address: ":${PORT:8080}"   # env expansion supported (${VAR} or ${VAR:default})
    routes:
      - method: GET
        path: /api/hello
        handler: hello         # resolved from Registration Context (code below)
    mount_static:
      - prefix: /static/
        spa: false
        folder: ["./public"]
    mount_htmx:
      - prefix: /
        sources: ["./web_htmx"]
    # mount_reverse_proxy:
    #   - prefix: /api/github/
    #     target: https://api.github.com
```

**main.go**

```go
package main

import (
    "log"
    "time"
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/core/config"
    request "github.com/primadi/lokstra/core/request"
)

// === Handlers referenced by name in YAML ===
func hello(ctx *request.Context) error {
    name, _ := ctx.Query("name")
    if name == "" { name = "World" }
    return ctx.Ok(map[string]any{"message": "Hello " + name})
}

func main() {
    reg := lokstra.NewGlobalRegistrationContext()

    // Register handlers by name so YAML can find them
    reg.RegisterHandler("hello", hello)

    // Load all YAMLs in ./config (LoadConfigDir)
    cfg, err := config.LoadConfigDir("./config")
    if err != nil { log.Fatal(err) }

    // Build server from config (starts modules, services, apps)
    svr, err := lokstra.NewServerFromConfig(reg, cfg)
    if err != nil { log.Fatal(err) }

    _ = svr.StartAndWaitForShutdown(5 * time.Second)
}
```

Run:

```bash
PORT=9090 go run .
```

### Multiple files & includes

You can split config across files; `LoadConfigDir` merges them. Inside **group** entries you can **include** other YAMLs with `load_from` (see `core/config/types.go` and loader).

```yaml
# config/routes.yaml
apps:
  - name: web-app
    groups:
      - prefix: /api
        load_from:
          - routes/users.yaml
          - routes/admin.yaml
```

The included files must contain a **GroupConfig** schema (routes, nested groups, mounts), not app-level fields.

### Middleware in YAML

Middleware accepts either simple names or objects:

```yaml
apps:
  - name: web-app
    address: ":8080"
    middleware:
      - cors                 # simple
      - name: auth           # detailed
        enabled: true
        config:
          secret: ${AUTH_SECRET:dev}
    routes:
      - method: GET
        path: /secure
        handler: hello
        middleware:
          - audit
```

---

## 3) Add common mounts quickly

### Static files

```go
app.MountStatic("/static/", false, os.DirFS("./public"))
```

- `spa=true` for client‑side routers (fallback to `index.html`), see `static-files.md`.

### HTMX pages

```go
app.MountHtmx("/", nil, os.DirFS("./web_htmx"))
pd := app.Group("/page-data")
pd.GET("/", func(c *request.Context) error {
    return c.HtmxPageData("Home", "", map[string]any{"now": time.Now().Format(time.RFC3339)})
})
```

See `htmx.md` for the full layout/page‑data model.

### Reverse proxy

```go
app.MountReverseProxy("/api/github/", "https://api.github.com", false, "auth")
```

See `reverse-proxy.md` for middleware & error semantics.

### RPC

Mount a service instance and expose its exported methods via `POST {base}/:method` (MessagePack body):

```go
type Greeter struct{}
func (g *Greeter) Hello(name string) (string, error) { return "Hello " + name, nil }
app.MountRpcService("/rpc", &Greeter{}, false)
```

If you configure via YAML, ensure an RPC server service exists with **name** `rpc_server.default` and **type** `rpc_service`:

```yaml
services:
  - name: rpc_server.default
    type: rpc_service
```

(When present, the router will use it and skip auto‑creation.)

---

## 4) Services & modules

`lokstra.NewGlobalRegistrationContext()` calls `defaults.RegisterAll` internally; you already get common modules (CORS, recovery, gzip, request loggers) and service factories (logger, metrics, etc.).

Add your own services in code:

```go
reg.RegisterServiceFactory("my.db", func(cfg any) (service.Service, error) { /* ... */ })
_, _ = reg.CreateService("my.db", "db.primary", map[string]any{"dsn":"postgres://..."})
```

Or via YAML:

```yaml
services:
  - name: db.primary
    type: my.db
    config:
      dsn: ${ENV:PG_DSN:postgres://localhost:5432/app}
```

You can also load **plugin modules** (`.so`) via YAML `modules` (see `modules.md`).

---

## 5) Environment variables in YAML

The loader supports `${NAME}` and `${NAME:default}`. It also supports **resolver prefixes** like `${ENV:NAME:default}` (see `core/config/var_resolver.go`).

Example:

```yaml
apps:
  - name: web
    address: ":${PORT:8080}"
```

### Useful server‑level settings

Under `server.global_setting` you can set:

- `log_level` / `log_format` / `log_output` → affects logger service (`serviceapi/logger.go`)
- `flow_logger`, `flow_dbschema` → default settings for the `flow` package (if you use it)

---

## 6) Project layout (suggested)

```
hello-lokstra/
  main.go
  config/
    app.yaml
    routes/
      users.yaml
  public/
    app.css
    app.js
  web_htmx/
    layouts/base.html
    pages/index.html
```

---

## 7) Common pitfalls

- **Handler signature** must return `error`. Use helpers to set status/JSON and return the result:
  ```go
  if err := ctx.BindAllSmart(&req); err != nil { return ctx.ErrorBadRequest(err.Error()) }
  return ctx.Ok(data) // ctx.Ok returns nil on success
  ```
- **Auto‑binding requires tags**: `path:""`, `query:""`, `header:""`, `body:""` (no tag → not bound).
- **RPC server factory name**: if you don’t pre‑create `rpc_server.default` (type `rpc_service`), auto‑creation expects a factory named `rpc_service.rpc_server`. Easiest fix: just define the service in YAML as shown above.
- **Middleware override**: for mounts and routes, `override_middleware: true` isolates from parent stacks.
- **Use trailing slash** for static mounts (e.g., `/static/`) to avoid odd path joins.
- **Signals**: use `StartAndWaitForShutdown` for graceful exit on Ctrl‑C; apps shut down concurrently.

---

## What to read next

- `router.md` – groups, route metadata, middleware ordering
- `static-files.md` – SPA fallback and FS layering
- `htmx.md` – server‑side layouts + partial renders
- `reverse-proxy.md` – gateway pattern with error semantics
- `rpc.md` – MessagePack RPC (server and client helpers)
