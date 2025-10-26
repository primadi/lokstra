# Lokstra v0.3.0 Architecture (Updated)

**Tagline:** *Simple. Scalable. Structured.*

Lokstra v0.3.0 focuses on an engine-first design, clear separation of concerns, and a DX that scales from “single file demo” to “multi‑module systems”. This update captures our latest decisions around router/listener engines, middleware execution, handler interop, YAML config, and app chaining.

---

## 0) Core Principles

- **Engine-first:** Listener and Router are first-class, pluggable engines (not “services”).
- **DX-first router API:** One Router interface for direct apps and AppModules.
- **Std interop:** Internally normalize everything to **`http.Handler`**.
- **Deterministic middleware:** Single linear chain with `Next()` semantics.
- **Doc-first:** Docs are source of truth; code and YAML mirror the docs.
- **Phased lifecycle:** Configuration → Build → Run (config freezes after build).

---

## 1) Lifecycle Phases

1. **Configuration Phase**
   - Load YAML, expand variables, resolve module paths, validate schema.
   - Compose *engines* (listener/router) & *registries*.
2. **Build Phase**
   - Resolve/merge app chains (same `addr` share one listener).
   - Build middleware chains (server → app → group → route) with priorities.
   - Register routes and named handlers into the chosen Router engine.
3. **Run Phase**
   - Start listeners; serve requests through the resolved Router engine.
   - No config mutation allowed (frozen).

---

## 2) Engines: Listener vs Router

Listener and Router are **pluggable engines** managed by their own registries:

- **ListenerEngine** examples: `net/http` (HTTP/1), HTTP/2 (TLS), HTTP/3 (QUIC), `fasthttp`, Unix socket, Windows named pipe.
- **RouterEngine** examples: `servemux` (std), `httprouter`, `fasthttprouter`.

> Engines are **not** part of the Service Registry. They live in dedicated registries: `ListenerRegistry` and `RouterRegistry`.

### Defaults
- **Listener:** `net/http` (HTTP/1)  
- **Router:** `servemux` (standard library `http.ServeMux`)

You can swap engines via YAML without touching application code.

---

## 3) Router Interface (v3)

One interface used by direct apps. Internally, every handler is normalized into `http.Handler`.

```go
type Router interface {
    // High-level (Lokstra) handlers
    GET(name, path string, h HandlerFunc, middleware ...string)
    POST(name, path string, h HandlerFunc, middleware ...string)
    PUT(name, path string, h HandlerFunc, middleware ...string)
    DELETE(name, path string, h HandlerFunc, middleware ...string)
    PATCH(name, path string, h HandlerFunc, middleware ...string)

    // Standard library interop
    GETStd(name, path string, h http.Handler, middleware ...string)
    POSTStd(name, path string, h http.Handler, middleware ...string)
    // ... other methods

    // Grouping
    Group(prefix string, fn func(r Router))
    AddGroup(prefix string) Router

    // Middleware by name (resolved from YAML)
    Use(middleware ...string)

    // Raw registration without method (e.g., reverse proxy)
    // This bypasses method matching in the Router engine.
    RawHandler(path string, h http.Handler)
}
```

**Notes**  
- `name` is required; if empty it auto-generates as `METHOD:/path`.  
- Route-level `middleware` are **string names**. Their configs live in YAML.  
- Internally, Lokstra adapts `HandlerFunc` → `http.Handler` once at registration.  
- `RawHandler` registers a path that delegates fully to a provided `http.Handler` (no method check).

---

## 4) Middleware Model

- **Single linear chain**; middleware is just a function that may call `Next()`.
- If a middleware **does not call** `Next()`, the chain **stops** (short‑circuit).
- “Before/After” behavior is done naturally via `Next()` and `defer`:

```go
func mw(c *RequestContext) error {
    // before
    defer func(){ /* after */ }()
    return c.Next()
}
```

### Resolution & Ordering
All middleware (server, app, group, route) are collected into one list:
- Each item has `Name`, `Priority`, `RegisterOrder`, `Source` (server/app/group/route), and `Func`.
- Sorted by `(Priority, RegisterOrder)` to guarantee deterministic execution.
- Inline/anonymous middleware get a default priority and emit a DX warning.

### No-Op
No explicit “null middleware” is needed. Disabled or unresolved middleware are **skipped** during chain build.

---

## 5) Handler Interop & Response Rules

Lokstra supports **three** registration styles:

1. **Lokstra handler**: `func(*lokstra.RequestContext) error`  
   - Return `nil` → success; `error` → converted to an HTTP error by response mapping.
2. **Stdlib handler**: `http.Handler` / `http.HandlerFunc`
3. **Raw path**: `RawHandler(path, http.Handler)` (no method; useful for reverse proxy)

**Direct Writer Access**  
If a handler writes directly to `http.ResponseWriter`, Lokstra **steps aside**:
- First status write wins (subsequent writes are ignored by the writer, silently).
- Use Raw/Std handlers for fully manual responses.

---

## 6) Modules (Server‑Scoped, no AppModule)

There is **no AppModule**. We removed AppModule to simplify mental model and avoid two places to register things.
All modules are **server-scoped** and are loaded during the **Configuration/Build** phases.

A **Module** may:
- Register **service factories** and/or create **service instances** (to Server DI).
- Register **named middleware** (with default config/priority) into the **middleware registry**.
- Register **named handlers** (no path/method) into the **handler registry** so Apps can reference them by name in code or YAML routes.

A **Module** may **not**:
- Mount routes or groups directly.
- Access App settings.
- Mutate runtime after Run phase starts.

**Why:** Apps should only decide *composition* (addr, prefix, which handlers to wire, which middleware to require). 
Modules provide capabilities (services, middleware, handlers) but do not attach to an App. This avoids duplication between AppModule vs ServerModule and keeps DX focused: “capabilities come from Server, wiring happens in App.”

---

## 7) Services & DI


- Access services in handlers via `GetService[T](name)` (with lazy caching).  
- **Required services**: missing → runtime error.  
- **Optional services**: missing → fallback to a **Null** implementation.

```go
type NullLogger struct{}
func (n *NullLogger) Debug(string, ...any) {}
func (n *NullLogger) Info(string, ...any)  {}
func (n *NullLogger) Warn(string, ...any)  {}
func (n *NullLogger) Error(string, ...any) {}
```

### Variable Resolver Service
Interpolation supports multiple resolvers:
- `${ENV:KEY:default}` (environment variables)
- `${SVC:LOADER:KEY:default}` (from a named service/loader)  
The system is extensible via a `VariableResolverService` interface.

---

## 8) YAML Configuration (v3)

### Server section
Defines engines, services, and default/named middleware.

```yaml
servers:
  - name: monolith
    listener: http          # default: http (net/http)
    router: servemux        # default: servemux
    apps:
      - name: blog-app
        addr: ":8080"
      - name: admin-app
        addr: ":8080"       # same addr → chained into the same listener

    # Load modules at SERVER level (no AppModule):
    modules:
      - name: AuthService
      - name: UserManagement
      - name: PermissionEvaluator

    services:
      - type: lokstra.logger
        name: logger
      - type: lokstra.dbpool
        name: db&main
        config:
          dsn: ${ENV:DB_DSN:postgres://...}

    middleware:
      - name: cors&default
        type: lokstra.cors
        config:
          allow-origins: ["*"]
```

### App section
Links modules to apps and declares per-app middleware and dependencies.

```yaml
apps:
  - name: blog-app
    required-services: [db&main]
    required-middleware: [cors&default]
    middleware:             # app-level chain augmentation
      - name: request-logger
        enabled: true
        config: { level: info }

  - name: admin-app
    prefix: /admin
```

### Routes (optional explicit form)
For YAML-driven routes without code:

```yaml
routes:
  - app: blog-app
    name: health
    method: GET
    path: /healthz
    handler: healthHandler         # refers to a named handler registered in code or a module
    middleware:
      - name: rate-limit
        enabled: true
```

**Notes**
- **`addr` (not `port`)** is used to support sockets and advanced listeners.  
- If multiple apps share the same `addr`, they are **chained** (see next).  
- `enabled` in middleware defaults to `true` when omitted.

---

## 9) App Chaining (same addr)

When multiple apps declare the **same `addr`**:
- They are **chained** behind a **single listener**.
- The **first** app encountered provides the **listener** and **router** engine configuration for the chain.
- Subsequent apps on that `addr` reuse the same engines and only contribute modules/routes.
- Order of route registration follows declaration order (first wins for overlapping paths).

This mechanism enables a **single-binary** to serve monolith or multiple modular apps without changing code, merely by editing YAML.

---

## 10) Struct Snapshots (Go)

```go
type ServerConfig struct {
    Name       string              `yaml:"name"`
    Listener   string              `yaml:"listener,omitempty"` // http, http2, http3, fasthttp, unix, npipe
    Router     string              `yaml:"router,omitempty"`   // servemux, httprouter, fasthttprouter
    Modules    []ModuleDecl        `yaml:"modules,omitempty"`  // server-scoped modules
    Services   []ServiceConfig     `yaml:"services,omitempty"`
    Middleware []MiddlewareDecl    `yaml:"middleware,omitempty"`
    Apps       []AppLink           `yaml:"apps,omitempty"`
}

type AppLink struct {
    Name string `yaml:"name"`
    Addr string `yaml:"addr"` // ":8080", "unix:/tmp/app.sock", etc.
}

type AppConfig struct {
    Name               string             `yaml:"name"`
    Prefix             string             `yaml:"prefix,omitempty"`
    RequiredServices   []string           `yaml:"required-services,omitempty"`
    OptionalServices   []string           `yaml:"optional-services,omitempty"`
    RequiredMiddleware []string           `yaml:"required-middleware,omitempty"`
    OptionalMiddleware []string           `yaml:"optional-middleware,omitempty"`
    Middleware         []MiddlewareConfig `yaml:"middleware,omitempty"` // app-level add/override
    Routes             []RouteConfig      `yaml:"routes,omitempty"`
}

type RouteConfig struct {
    App        string             `yaml:"app"`
    Name       string             `yaml:"name"`
    Path       string             `yaml:"path"`
    Method     string             `yaml:"method"`
    Handler    string             `yaml:"handler"`
    Middleware []MiddlewareConfig `yaml:"middleware,omitempty"`
}

type ServiceConfig struct {
    Type   string         `yaml:"type"`
    Name   string         `yaml:"name"` // "type&name" convention applies where helpful
    Config map[string]any `yaml:"config,omitempty"`
}

type ModuleDecl struct {
    Name string `yaml:"name"` // e.g., "AuthService"
}

type MiddlewareDecl struct {
    Name   string         `yaml:"name"`   // e.g., "cors&default"
    Type   string         `yaml:"type"`   // e.g., "lokstra.cors"
    Config map[string]any `yaml:"config,omitempty"`
}

type MiddlewareConfig struct {
    Name    string         `yaml:"name"`
    Enabled bool           `yaml:"enabled,omitempty"` // default true
    Config  map[string]any `yaml:"config,omitempty"`
}
```

---

## 11) Folder Structure

```
/lokstra
  /cmd            # runnable examples (from minimal to complex)
  /core           # core engine (server, app, engines, router, context)
  /common         # shared utilities (response, logging, helpers)
  /modules        # official modules (auth, user, etc.)
  /serviceapi     # standard service contracts
  /ui             # SSR/HTMX renderers & UI integrations (if present)
  /docs           # documentation (doc-first)
```

---

**Why this design?**  
- Easy to start (std `servemux`, `net/http`) and easy to swap (fasthttp/httprouter) later.  
- Deterministic middleware semantics with `Next()` keeps code simple yet powerful.  
- Clear separation between **engines** and **services** avoids awkward DI for listeners/routers.  
- App chaining enables flexible deployment topologies without code changes.
