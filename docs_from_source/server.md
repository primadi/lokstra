
# Server

**Source-driven**: Mirrors `core/server/server.go` and `core/config/lokstra_config.go` in `lokstra-0.2.1`.

A **Server** hosts one or more **Apps** (each App binds a Router + Listener + Address).  
It handles **parallel start**, **signal-aware shutdown**, settings, and **merging apps with the same address**.

---

## Struct

```go
type Server struct {
    ctx      registration.Context
    name     string
    apps     []*app.App
    settings map[string]any
}
```

---

## Constructor

```go
func NewServer(ctx registration.Context, name string) *Server
```
Initializes an empty server with its own settings map and app list.

---

## Methods

```go
func (s *Server) GetName() string
func (s *Server) AddApp(app *app.App)
func (s *Server) NewApp(name string, addr string) *app.App

func (s *Server) RegisterModule(module registration.Module) error

func (s *Server) SetSetting(key string, value any)
func (s *Server) SetSettingsIfAbsent(settings map[string]any)
func (s *Server) GetSetting(key string) (any, bool)

func (s *Server) MergeAppsWithSameAddress()

func (s *Server) Start() error
func (s *Server) StartAndWaitForShutdown(shutdownTimeout time.Duration) error
func (s *Server) Shutdown(shutdownTimeout time.Duration) error

func (s *Server) ListApps() []*app.App
```

### What they do

- **GetName** — returns the server's name.  
- **AddApp** — appends an `*app.App` to the server.  
- **NewApp** — shorthand: constructs a new App via `app.NewApp(ctx, name, addr)`, adds it to the server, and returns it.  
- **RegisterModule** — executes `module.Register(registration.Context)` so the module can register services/handlers/middleware during bootstrap.  
- **SetSetting / SetSettingsIfAbsent / GetSetting** — key/value server-level settings store.  
- **MergeAppsWithSameAddress** — groups apps by `.GetAddr()`; apps sharing the same address will be merged to run on a **single listener/router engine** (internally calls `base.MergeOtherApp(peer)` on duplicates).  
- **Start** — builds and starts all apps (after merging same-address apps). Uses goroutines & a wait group; returns the first error encountered (if any).  
- **StartAndWaitForShutdown** — runs `Start()` in the background, waits for `os.Interrupt`/`syscall.SIGTERM`. On signal it calls `Shutdown(timeout)` and returns.  
- **Shutdown** — calls `app.Shutdown(timeout)` for each app, aggregating errors.  
- **ListApps** — returns the slice of registered apps.

> Note: The server prints each app's start banner through the app's `PrintStartMessage()` path.

---

## Using with Apps (imperative)

```go
reg := lokstra.NewGlobalRegistrationContext()

// App 1
app1 := lokstra.NewApp(reg, "api", ":8080")
app1.GET("/ping", "pingHandler")

// App 2 (same address, will be merged onto one listener)
app2 := lokstra.NewApp(reg, "admin", ":8080")
app2.GET("/dashboard", "admin.dashboard")

// Compose a server
svr := lokstra.NewServer(reg, "my-server")
svr.AddApp(app1)
svr.AddApp(app2)

// Optional: adjust server settings
svr.SetSetting("log_level", "info")

// Start and handle SIGINT/SIGTERM gracefully (10s timeout on shutdown)
if err := svr.StartAndWaitForShutdown(10 * time.Second); err != nil {
    panic(err)
}
```

---

## Using with YAML Config

High-level helpers (from `lokstra.go` and `core/config/lokstra_config.go`):

```go
// Create & start from config
svr, err := lokstra.NewServerFromConfig(reg, cfg)              // cfg.NewServerFromConfig(reg)
svr, err = lokstra.LoadConfigToServer(reg, cfg, existingServer) // cfg.LoadConfigToServer(reg, svr)

// Inside cfg.loadAndStartAll(reg, svr):
// 1) cfg.StartAllModules(reg)
// 2) cfg.StartAllServices(reg)
// 3) cfg.NewAllApps(reg, svr) + wire routes, groups, mounts
// 4) svr.SetSettingsIfAbsent(cfg.Server.Settings)
```

With YAML you can declare multiple apps; the loader will build them under a single server and preserve the same **merge** semantics for identical addresses.

---

## Tips

- Use `NewApp` on the server when you want a convenient way to create & attach apps.
- If multiple apps share `addr`, the server **merges** them automatically; you need to start only once.
- Keep module registration at the server/bootstrap level using `RegisterModule`, or rely on YAML `modules` so the loader calls `Module.Register(...)` for you.
- Prefer `StartAndWaitForShutdown` in production to handle OS signals cleanly.
