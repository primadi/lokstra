
# App

**Source-driven:** Mirrors `core/app/app.go` in `lokstra-0.2.1`.  
An **App** binds a **Router** and an **HTTP listener** to an address, and controls the lifecycle (build, start, shutdown).

---

## Struct


```go
type App struct {
    router.Router

    ctx      registration.Context
    listener serviceapi.HttpListener

    name              string
    addr              string
    listenerType      string
    routingEngineType string
    settings          map[string]any
    merged            bool
    mergedRoutes      []router.Router
}
```


- Embeds `router.Router` → the high-level router you compose (routes, groups, mounts).
- Holds a `serviceapi.HttpListener` → the underlying HTTP server (net/http, fasthttp, secure, http3).
- Keeps metadata: name, address, listener & router engine type strings, settings map, and merge state.

---

## Constructors


```go
func NewApp(ctx registration.Context, name string, addr string) *App
func NewAppCustom(ctx registration.Context, name string, addr string,
    listenerType string, routerEngineType string, settings map[string]any) *App

// Lifecycle
func (a *App) Start() error
func (a *App) BuildRouter() error
func (a *App) ListenAndServe() error
func (a *App) StartWithGracefulShutdown(shutdownTimeout time.Duration) error
func (a *App) Shutdown(shutdownTimeout time.Duration) error
func (a *App) PrintStartMessage(dumpRoutes bool)

// Composition
func (a *App) MergeOtherApp(otherApp *App)

// Introspection & settings
func (a *App) GetName() string
func (a *App) GetAddr() string
func (a *App) GetSettings() map[string]any
func (a *App) GetSetting(key string) (any, bool)
func (a *App) SetSetting(key string, value any)
func (a *App) IsMerged() bool
```


### Normalization & defaults
`NewAppCustom(...)` calls:

- `router.NormalizeListenerType(listenerType)`  
- `router.NormalizeRouterType(routerEngineType)`

Then it creates concrete engine instances via **service factories** (registered in the **Registration Context**):

```go
listener := router.NewListenerWithEngine(ctx, listenerType, settings)
router   := router.NewRouterWithEngine(ctx, routerEngineType, settings)
```

**Factory name prefixes** (from `serviceapi`):

- `HTTP_LISTENER_PREFIX = "lokstra.http_listener."`
- `HTTP_ROUTER_PREFIX   = "lokstra.http_router."`

**Defaults** (from `defaults/*`):

- Listeners: `default` → **net_http**, also available: `fast_http`, `secure_net_http`, `http3`.
- Routers:   `default` → **httprouter**, also available: `servemux`.

So, for example, `"httprouter"` is normalized to `"lokstra.http_router.httprouter"` before factory lookup.

---

## Lifecycle

### `Start()`
- Calls `BuildRouter()` (unless the app is **merged**, see below).
- Delegates to `ListenAndServe()`.

### `BuildRouter()`
- Resolves all named handlers/middleware/services:  
  `router.ResolveAllNamed(regCtx, app.Router.GetMeta())`
- Builds the underlying engine (`RouterImpl.BuildRouter()`).
- If the app has **merged** routes from other apps (`MergeOtherApp`), those routers:
  - **Share the same engine**: `otherRouterImpl.SetEngine(r_engine)`
  - Are then built on the shared engine.

### `ListenAndServe()`
- Prints a start banner via `listener.GetStartMessage(addr)` and dumps routes:
  `app.GetMeta().DumpRoutes()` (and for merged routers too).
- Starts the HTTP listener: `listener.ListenAndServe(addr, app.Router)`

### `StartWithGracefulShutdown(timeout)`
- Runs `Start()` in a goroutine.
- Waits for OS signals: `os.Interrupt`, `syscall.SIGTERM`.
- On signal, calls `Shutdown(timeout)` and returns.

### `Shutdown(timeout)`
- Delegates to `listener.Shutdown(timeout)`.

---

## Merging Apps

Use `MergeOtherApp(other *App)` to **join multiple routers** under a **single listener** and address:

```go
api := lokstra.NewApp(reg, "api", ":8080")
admin := lokstra.NewApp(reg, "admin", ":8080")

// Build admin under the same engine as api
api.MergeOtherApp(admin)

// Now only start 'api'; it builds both routers on one engine and listener
_ = api.Start()
```

Notes:
- The merged app is marked `merged=true` and won’t start its own listener.
- During `BuildRouter`, the base app’s engine is reused for all merged routers.

---

## Settings

`settings map[string]any` is passed into both the **listener** and **router** factories.  
Use it for engine-specific options (timeouts, TLS files, etc.).

### Examples (listeners)

From the default listener factories (in `modules/coreservice/listener/*`):

- **net_http** / **secure_net_http**:
  - `read_timeout`, `write_timeout`, `idle_timeout` (durations, e.g., `"30s"`)
  - TLS files can be provided as `["cert_file", "key_file"]` array or as keys in a map (for secure listener).

- **fast_http**:
  - Same timeout keys are supported.

### Examples (router engines)

- **httprouter** and **servemux** engines are factory-created.  
  (Most router behavior is configured at the `router.Router` level via routes/middleware/mounts.)

---

## Typical Boot Code

```go
reg := lokstra.NewGlobalRegistrationContext()

// Build Router imperatively (or via YAML → see yaml-config.md)
r := router.NewRouter(reg, nil).
    WithPrefix("/api").
    Use("auth").
    GET("/health", "healthz")

app := app.NewAppCustom(reg, "api", ":8080", "net_http", "httprouter", map[string]any{
    "read_timeout":  "30s",
    "write_timeout": "30s",
})

if err := app.StartWithGracefulShutdown(5 * time.Second); err != nil {
    panic(err)
}
```

Or use YAML to construct the router and app settings, then call `BuildRouter()` followed by `Start()`.

---

## What to Remember

- **App** = Router + Listener + Address + Lifecycle.
- **Factories** (registered in the **Registration Context**) create listeners & router engines by name.
- **BuildRouter** resolves names and shares the engine when merging apps.
- Use **StartWithGracefulShutdown** for graceful exits on SIGINT/SIGTERM.
