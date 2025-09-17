
# YAML Configuration

**Source-driven:** This page mirrors the actual config types in `lokstra-0.2.1/core/config/types.go`.  
Use YAML to declare servers, apps, routes, mounts, services, and modules. The App will build routers and wire everything at startup.

---

## Top-Level Schema

```yaml
server:        # -> ServerConfig
apps:          # -> []AppConfig
services:      # -> []ServiceConfig
modules:       # -> []ModuleConfig
```

### `ServerConfig`
```yaml
server:
  name: demo
  global_setting:        # map[string]any
    env: prod
    trace: true
```

### `AppConfig`
```yaml
apps:
  - name: api
    address: ":8080"
    listener_type: "net"            # e.g., "net"
    router_engine_type: "httprouter"  # or "servemux"

    setting:                        # map[string]any (app-local)
      cors: true

    # Optional: middleware at app level
    middleware:
      - name: cors
        enabled: true
        config: { allow_origins: ["*"] }
      - auth                         # string form is also accepted

    # Mounts (any subset)
    mount_static:
      - prefix: "/static/"
        spa: false
        folder: ["./public", "./theme"]

    mount_htmx:
      - prefix: "/"
        sources: ["./htmx", "./modules/blog/htmx"]

    mount_reverse_proxy:
      - prefix: "/ext/"
        target: "https://example.org"

    mount_rpc_service:
      - base_path: "/rpc/user"
        service_name: "user.rpc"

    # Routes declared directly
    routes:
      - method: GET
        path: "/health"
        handler: "healthz"           # named handler registered in Registration
        middleware: ["audit"]
      - method: POST
        path: "/users"
        handler: "user.create"
        override_middleware: false   # if true, only route-level middleware is used

    # Router groups (nesting supported)
    groups:
      - prefix: "/api"
        middleware: ["auth"]
        routes:
          - method: GET
            path: "/users/{id}"
            handler: "user.get"
            middleware: ["audit"]
        groups:
          - prefix: "/admin"
            middleware:
              - name: audit
                enabled: true
            routes:
              - method: DELETE
                path: "/users/{id}"
                handler: "user.delete"
```

> `override_middleware`:  
> - At **route** level → ignore parent middlewares and use only the route list.  
> - At **app/group** level → children start from an empty chain (if enabled).

---

## RouteConfig

```yaml
method: GET | POST | PUT | PATCH | DELETE
path: "/users/{id}"
handler: "<named handler>"   # string name resolved from the Registration Context
override_middleware: false
middleware:                   # either an array of strings or objects
  - audit
  - name: cors
    enabled: true
    config: { allow_origins: ["*"] }
```

**Handler resolution** (from source): you may register handlers by name, or use the generic function form `func(ctx, *T) error` when wiring imperatively. In YAML, `handler` is always a **name** that must exist in the Registration Context at build time.

---


### Group Includes (`load_from`)

**Schema reference:** `\schema` (uses the **GroupConfig** schema).

Each file referenced by `groups[].load_from` **must** follow the *GroupConfig* shape.
Relative paths are resolved against the parent YAML’s directory. Environment variables are expanded.

**Allowed top-level keys in included files:**
- `routes` (array of `RouteConfig`)
- `groups` (array of nested `GroupConfig`)
- `mount_static`, `mount_htmx`, `mount_reverse_proxy`, `mount_rpc_service`

**Forbidden at the root of an included file** *(enforced by loader)*:
- `prefix`
- `override_middleware`
- `middleware`

> If these forbidden keys appear at the root of an included file, the loader returns an error
> (see `expandGroupIncludes` in `core/config/config_loader.go`).

**Minimal example** (`routes/user.yaml`):
```yaml
routes:
  - method: GET
    path: "/users/{id}"
    handler: "user.get"
  - method: POST
    path: "/users"
    handler: "user.create"
```

**Usage in parent file:**
```yaml
groups:
  - prefix: "/api"
    load_from:
      - "routes/user.yaml"      # resolved relative to this file
      - "routes/order.yaml"
```

**Merging rules:**
- `routes`, `groups`, and all `mount_*` arrays are **appended** to the parent group.
- Nested `groups` inside included files are also expanded recursively.


## Mounts

### Static
```yaml
mount_static:
  - prefix: "/static/"
    spa: true               # SPA fallback to /index.html
    folder: ["./public", "./theme"]
```
- Multiple folders are supported; the first matching file wins.

### HTMX
```yaml
mount_htmx:
  - prefix: "/"
    sources: ["./htmx", "./modules/blog/htmx"]
```
- Expected directories: `/layouts` and `/pages` inside the sources.
- Each page can declare its layout using `<!-- layout: base.html -->` (default `base.html`).
- The handler makes an **internal call** to `/page-data/<slug>` to fetch JSON:
  ```json
  {
    "code": "...",
    "data": { "title": "...", "description": "...", "data": { /* page data */ } }
  }
  ```
  Build this JSON using `ctx.HtmxPageData(title, description, dataMap)` in your page-data endpoints.

### Reverse Proxy
```yaml
mount_reverse_proxy:
  - prefix: "/ext/"
    target: "https://example.org"
```
- Requests under `prefix` are forwarded using `httputil.ReverseProxy`.
- Middleware composition follows normal rules (including override).

### RPC Service
```yaml
mount_rpc_service:
  - base_path: "/rpc/user"
    service_name: "user.rpc"     # must resolve to a registered service
```
- The router will mount RPC endpoints under the base path (e.g., POST `/rpc/user/:method`).

---

## MiddlewareConfig

```yaml
middleware:
  - name: cors
    enabled: true
    config:
      allow_origins: ["*"]
  - auth
```
- Names resolve to **middleware factories** in the Registration Context.
- Priority comes from the factory registration; order within the same priority is by insertion.

---

## Services

```yaml
services:
  - name: "db.main"
    type: "lokstra.dbpool"       # factory name
    config:
      dsn: ${DB_DSN}             # typical env-substitution via your loader
    depends_on: ["logger"]
```
- `type` maps to a registered **service factory**.  
- `depends_on` controls startup ordering (if used by your boot sequence).

---

## Modules

```yaml
modules:
  - name: user
    path: "./modules/user/user.so"   # compiled plugin
    entry: "GetModule"               # optional; default is "GetModule"

    settings:
      default_role: "member"
    permissions:
      allow_register_handler: true
      get_service: ["logger", "db.*"]

    required_services: ["logger", "db.main"]
    create_services:
      - name: "cache.main"
        type: "lokstra.cache"
        config: { url: "redis://localhost:6379" }

    register_service_factories: ["RegisterCacheFactories"]
    register_handlers: ["RegisterUserHandlers"]
    register_middleware: ["RegisterAuditMiddleware"]
```
- From source: registration will **open** the plugin, call the entry function to get a `Module`, then invoke `Module.Register(reg)`.
- If a module with the same `name` already exists, registration fails.

---

## Splitting Files & Reuse

- `groups[].load_from`: you can split route groups into separate YAML files and **include** them:
  ```yaml
  groups:
    - prefix: "/api"
      load_from: ["./routes/user.yaml", "./routes/order.yaml"]
  ```
- You can also load multiple config files or a folder in your bootstrap (e.g., iterate files and merge). The exact loader is up to your app layer; Lokstra’s config structs are designed to be merged.

---

## Tips

- Keep **handler names** and **middleware names** consistent between Registration and YAML.
- Prefer **grouping** routes by feature and mounting static/HTMX/reverse-proxy at the **app** level.
- Use **override** carefully to isolate middleware stacks for admin/public zones.
- When using HTMX, don’t forget the corresponding **page-data** endpoints that return `HtmxPageData(...)`.
