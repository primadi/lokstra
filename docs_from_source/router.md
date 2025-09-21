
# Router

**Source-driven**: This document reflects the real API in `lokstra-0.2.1` (`core/router`, `serviceapi/router_engine.go`, `common/static_files/*`).  
The Router wires **handlers**, **middleware**, and **mounts** (static, HTMX, reverse proxy, RPC) onto an underlying **router engine**.

> The Router is built during app startup. You typically compose routes/groups, then the App will resolve names and build the engine.

---

## Create a Router

```go
import "github.com/primadi/lokstra/core/router"

// Use the default engine (alias to httprouter)
r := router.NewRouter(regCtx, map[string]any{/* engine config */})

// Or choose an engine explicitly
r2 := router.NewRouterWithEngine(regCtx, "servemux", nil)     // maps to "lokstra.http_router.servemux"
r3 := router.NewRouterWithEngine(regCtx, "httprouter", nil)   // maps to "lokstra.http_router.httprouter"
```

- Engine names are normalized to `serviceapi.HTTP_ROUTER_PREFIX + <name>` (e.g. `lokstra.http_router.httprouter`).  
- **Default** engine: `httprouter` (registered by `defaults.RegisterAllHTTPRouters`).

---

## Interface (essentials)

```go
type Router interface {
    // Routing
    Handle(method request.HTTPMethod, path string, handler any, mw ...any) Router
    HandleOverrideMiddleware(method request.HTTPMethod, path string, handler any, mw ...any) Router
    GET(path string, handler any, mw ...any) Router
    POST(path string, handler any, mw ...any) Router
    PUT(path string, handler any, mw ...any) Router
    PATCH(path string, handler any, mw ...any) Router
    DELETE(path string, handler any, mw ...any) Router

    // Middleware & prefix
    Use(any) Router
    WithOverrideMiddleware(enable bool) Router
    WithPrefix(prefix string) Router

    // Mounts
    RawHandle(prefix string, stripPrefix bool, h http.Handler) Router
    MountStatic(prefix string, spa bool, sources ...fs.FS) Router
    MountHtmx(prefix string, si *static_files.ScriptInjection, sources ...fs.FS) Router
    MountReverseProxy(prefix string, target string, overrideMiddleware bool, mw ...any) Router
    MountRpcService(path string, service any, overrideMiddleware bool, mw ...any) Router

    // Grouping & merging
    Group(prefix string, mw ...any) Router
    GroupBlock(prefix string, fn func(gr Router)) Router
    AddRouter(r Router) Router

    // Introspection & engines
    RecurseAllHandler(func(rt *RouteMeta))
    DumpRoutes()
    ServeHTTP(w http.ResponseWriter, r *http.Request)
    FastHttpHandler() fasthttp.RequestHandler
    OverrideMiddleware() bool
    GetMeta() *RouterMeta
}
```

---

## Handlers (3 ways)

The `handler` parameter in `Handle` / `GET` / ... accepts:

1) **Function**: `request.HandlerFunc`  
   ```go
   func(ctx *request.Context) error
   ```

2) **Named handler**: `string` — name previously registered in the **registration context**.

3) **Generic form with params**:  
   ```go
   func(ctx *request.Context, params *T) error
   ```
   The router auto-wraps this form and performs **`ctx.BindAll(params)`** (JSON body for body-part).  
   On bind error, it returns `ctx.ErrorBadRequest(err.Error())` for you.

> If a handler returns a raw `error`, Lokstra treats it as **500**. Use response helpers for 4xx/5xx.

---

## Middleware

You can attach middleware at the **router**, **group**, or **route** level:

```go
// Accepts: midware.Func, string (named factory), or *midware.Execution
r.Use("cors")
api := r.Group("/api", "auth")
api.GET("/info", getInfo, "audit")
```

Resolution & ordering:

- Named middleware is resolved via `regCtx.GetMiddlewareFactory(name)` during **build** time.
- Each middleware carries a **Priority** (from the factory registration). Lower number ⇒ higher priority.
- Within the same priority, **insertion order** is preserved (`ExecutionOrder`).
- Composition sorts by `(Priority + ExecutionOrder)`, then wraps handler from **innermost → outermost**.

Override rules:

- At **router/group**: `WithOverrideMiddleware(true)` means children start with an empty chain (they only get what you add under that node).  
  Otherwise, children inherit parent middleware chain.
- At **route**: `HandleOverrideMiddleware(...)` uses only the route-level middleware (no parent chain).

---

## Paths & Prefixes

- `WithPrefix("/v1")` sets an absolute prefix; `"users"` becomes `/v1/users`.
- If you call `WithPrefix("v1")` (no leading slash), it is **relative** to the current prefix.
- `Group("/admin")` nests a sub-router with its own prefix and optional middleware.

Utility methods:

- `DumpRoutes()` prints a comprehensive view of routes, mounts, and middleware names.
- `RecurseAllHandler(cb)` lets you enumerate all registered routes programmatically.

---

## Static Files

```go
// Highest to lowest priority lookup order
sf := static_files.New().
    WithSourceDir("./web_override").
    WithEmbedFS(assets, "web") // compiled fallback

r.MountStatic("/static/", false, sf.Sources...)
```

- `MountStatic(prefix, spa, sources...)` serves static assets from multiple `fs.FS` sources.  
  The first source that contains the file **wins**.  
- `spa=true` ⇒ for unknown paths, serves `/index.html` (client-side routing).

---

## HTMX Pages (Layouts + Page-Data)

```go
// Provide layout & page templates via FS sources
sf := static_files.New().
    WithSourceDir("./htmx").       // contains layouts/ and pages/
    WithEmbedFS(appFS, "htmx_app") // fallback

// Optionally inject scripts into layouts (head/body)
inj := static_files.NewScriptInjection().AddNamedScriptInjection("default")

r.MountHtmx("/", inj, sf.Sources...)
```

### How it works

- Expected directories in your FS sources:
  - `/layouts` — HTML layout templates
  - `/pages`   — HTML page templates
- A request `GET /about` loads `pages/about.html`. The page may declare its layout via a comment:  
  `<!-- layout: base.html -->` (default: `base.html`)
- The HTMX handler **internal-calls** your page-data endpoint at `/page-data/about` to get data:
  - Expects JSON body `{ "code": "...", "data": { "title": "...", "description": "...", "data": {...} } }`
  - Build this via `ctx.HtmxPageData(title, description, dataMap)`
- **Partial render** (HTMX) if request has `HX-Request: true` and `LS-Layout` matches: only the page template is rendered, with headers:
  - `HX-Partial: true`, `LS-Title`, `LS-Description`
- **Full render** otherwise: layout + page are combined; `title`/`description` are injected into `<head>`.

Example page-data routes:
```go
pg := r.Group("/page-data")
pg.GET("/", func(ctx *lokstra.Context) error {
    return ctx.HtmxPageData("Home Page", "", map[string]any{ "message": "Welcome" })
})
pg.GET("/about", func(ctx *lokstra.Context) error {
    return ctx.HtmxPageData("About Us", "Team and story", map[string]any{ "team": []string{"A","B"} })
})
```

---

## Reverse Proxy

```go
// Simple proxy (no middleware)
r.MountReverseProxy("/api/", "http://localhost:8081", false)

// Proxy with middleware & override (use ONLY these middlewares for this proxy)
r.MountReverseProxy("/ext/", "https://example.org", true, "auth", "audit")
```

Behavior:

- Uses `httputil.ReverseProxy` under the hood (engine decides strip-prefix behavior).
- Middleware chain for proxy requests is composed like normal routes (including **override** rules).
- **Success**: the upstream writes directly to `http.ResponseWriter` (Lokstra does **not** re-encode JSON).
- **Error**: if a middleware sets an error (or handler returns error), Lokstra writes the structured JSON response.

---

## RPC Service Mount

```go
// Service can be a string (service name in registry), a *service.RpcServiceMeta, or a concrete service instance
r.MountRpcService("/rpc/user", "user.rpc", false, "auth")
// Internally it registers POST /rpc/user/:method and wires it to an RpcServer
```

At build time, named services are resolved; an `RpcServer` is created or reused (`rpc_service.rpc_server`) and receives calls.

---

## Raw Handlers

```go
r.RawHandle("/metrics", false, promhttp.Handler())
r.RawHandle("/assets/", true, http.FileServer(os.DirFS("./public"))) // stripPrefix=true
```

Raw handlers bypass Lokstra’s `request.Context` and write directly to `http.ResponseWriter`.

---

## Putting It Together

```go
api := r.Group("/api", "auth")
api.GET("/users/{id}", getUser, "audit")
api.POST("/users", createUser, "audit")

assets := static_files.New().WithSourceDir("./public")
r.MountStatic("/static/", false, assets.Sources...)

inj := static_files.NewScriptInjection().AddNamedScriptInjection("default")
r.MountHtmx("/", inj, assets.Sources...)

r.MountReverseProxy("/ext/", "https://example.org", true, "auth")
```

When the App starts, it will:

1. Resolve named handlers/services/middleware (`ResolveAllNamed`).
2. Build the router engine (`BuildRouter`) and attach everything.
3. Print `DumpRoutes()` for visibility.

---

## Notes & Tips

- Prefer **generic handlers with params** (`func(ctx, *T) error`) to get automatic binding & 400 on bind errors.
- Use **override** thoughtfully to isolate middleware stacks per subtree or specific routes.
- For HTMX, ensure you mount `page-data` routes that return `HtmxPageData(...)`.
- `AddRouter` merges routes/mounts from another router **without** merging its middleware stack; call `Use(...)` explicitly if you need to unify middleware.
