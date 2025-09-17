
# Services

**Source-driven:** This page documents the service system in `lokstra-0.2.1` from `core/service`, `core/registration`, `serviceapi/*`, and `core/config`.

Lokstra's **services** are named, runtime-resolvable objects (any type) stored in the **Registration Context**.  
You register **factories** by name, then create named service instances consumed by routers, apps, modules, and handlers.

---

## Core Types (from source)

```go
package service

import (
	"fmt"
)

type ServiceFactory = func(config any) (Service, error)

type Service = any

func ErrUnsupportedConfig(config any) error {
	return fmt.Errorf("unsupported config type: %T", config)
}

func ErrInvalidServiceType(serviceName, expectedType string) error {
	return fmt.Errorf("invalid service type for %s, expected %s", serviceName, expectedType)
}

func ErrServiceNotFound(serviceName string) error {
	return fmt.Errorf("service %s not found", serviceName)
}
```

Notes:
- `Service` is an alias to `any` — services can be any Go value.
- `ServiceFactory` standardizes construction from a `config` payload (any type).
- Error helpers for common failure modes are provided.

RPC helper:
```go
package service

type RpcServiceMeta struct {
	MethodParam string // default "method"
	ServiceName string
	ServiceInst Service
}
```

---

## Service APIs you can implement (serviceapi/*)

Lokstra defines common **interfaces** for services used by the core (listeners, routers, RPC, DB, etc.).
Factories you register can return concrete implementations of these interfaces.

### HTTP Listener
```go
package serviceapi

import (
	"net/http"
	"time"
)

const HTTP_LISTENER_PREFIX string = "lokstra.http_listener."

type HttpListener interface {
	// ListenAndServe starts the HTTP server on the specified address.
	// It returns an error if the server fails to start.
	ListenAndServe(addr string, handler http.Handler) error
	// Shutdown gracefully stops the HTTP server.
	// It waits for all active requests to finish before shutting down.
	Shutdown(shutdownTimeout time.Duration) error
	// IsRunning checks if the HTTP server is currently running.
	IsRunning() bool
	// ActiveRequest returns the number of currently active requests.
	ActiveRequest() int

	// GetStartMessage returns a message indicating where the server is listening.
	GetStartMessage(addr string) string
}
```

### Router Engine
```go
package serviceapi

import (
	"io/fs"
	"net/http"

	"github.com/primadi/lokstra/common/static_files"
	"github.com/primadi/lokstra/core/request"
)

const HTTP_ROUTER_PREFIX string = "lokstra.http_router."

// RouterEngine defines the interface for a router engine that can handle HTTP methods,
// serve static files, HTMX Page, and reverse proxies.
type RouterEngine interface {
	// HandleMethod registers a handler for a specific HTTP method and path.
	HandleMethod(method request.HTTPMethod, path string, handler http.Handler)

	ServeHTTP(w http.ResponseWriter, r *http.Request)

	RawHandle(pattern string, handler http.Handler)
	RawHandleFunc(pattern string, handlerFunc http.HandlerFunc)

	ServeStatic(prefix string, spa bool, sources ...fs.FS)
	ServeReverseProxy(prefix string, handler http.HandlerFunc)

	// Assume sources has:
	//   - "/layouts" for HTML layout templates
	//   - "/pages" for HTML page templates
	//
	// All Request paths will be treated as page requests,
	ServeHtmxPage(pageDataRouter http.Handler, prefix string,
		si *static_files.ScriptInjection, sources ...fs.FS)
}
```

### RPC Server
```go
package serviceapi

import (
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
)

type RpcServer interface {
	HandleRequest(ctx *request.Context, service service.Service, MethodName string) error
}
```

### DB Pool
```go
package serviceapi

import "context"

// DbPool defines a connection pool interface
// supporting schema-aware connection acquisition
// and future multi-backend support.
type DbPool interface {
	Acquire(ctx context.Context, schema string) (DbConn, error)
}

type RowMap = map[string]any

// DbConn represents a live DB connection (e.g. from pgxpool)
type DbConn interface {
	Begin(ctx context.Context) (DbTx, error)
	Transaction(ctx context.Context, fn func(tx DbExecutor) error) error

	Release() error
	DbExecutor
}

// DbTx represents an ongoing transaction
type DbTx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	DbExecutor
}

type DbExecutor interface {
	Exec(ctx context.Context, query string, args ...any) (CommandResult, error)
	Query(ctx context.Context, query string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) Row

	SelectOne(ctx context.Context, query string, args []any, dest ...any) error
	SelectMustOne(ctx context.Context, query string, args []any, dest ...any) error

	SelectOneRowMap(ctx context.Context, query string, args ...any) (RowMap, error)
	SelectManyRowMap(ctx context.Context, query string,
...
```

### Metrics
```go
package serviceapi

type Metrics interface {
	IncCounter(name string, labels Labels)
	ObserveHistogram(name string, value float64, labels Labels)
	SetGauge(name string, value float64, labels Labels)
}

type Labels = map[string]string
```

### Redis
```go
package serviceapi

import "github.com/redis/go-redis/v9"

type Redis interface {
	Client() *redis.Client
}
```

> Implementations live in modules (e.g., `modules/coreservice/listener/*` for listeners, `modules/coreservice/router_engine/*` for routers).

---

## Registering & Creating Services

Use the **Registration Context** (see `registration.md`) to register factories and create named services.

**Register a factory**:
```go
reg.RegisterServiceFactory("lokstra.http_listener.net_http",
    func(cfg any) (service.Service, error) {{ /* return serviceapi.HttpListener */ }})
```

**Create a service instance** (named):
```go
svc, err := reg.CreateService("lokstra.http_listener.net_http", "listener.main", map[string]any{{
    "read_timeout":  "30s",
    "write_timeout": "30s",
}})
```

**Get / GetOrCreate**:
```go
inst, err := reg.GetService("listener.main")                      // returns service.Service
inst2, err := reg.GetOrCreateService("factoryName", "svc.name")   // idempotent create
```

**Typed retrieval**:
```go
// Convert from 'any' safely into a known interface/type
listener, err := serviceapi.GetService[serviceapi.HttpListener](reg, "listener.main")
```

**From config** (string name or map field → resolves to a typed service):
```go
// Given: cfg = "listener.main" or cfg = map[string]string{{"listener_name": "listener.main"}}
listener, err := registration.GetServiceFromConfig[serviceapi.HttpListener](reg, cfg, "listener_name")
```

---

## Naming Conventions

Factories use conventional **prefixes** (see `serviceapi` constants):

- `lokstra.http_listener.*` — HTTP listener engines (`net_http`, `fast_http`, `secure_net_http`, `http3`)
- `lokstra.http_router.*`   — Router engines (`httprouter`, `servemux`)

Defaults are registered by `defaults.RegisterAllHTTPListeners` and `defaults.RegisterAllHTTPRouters` when you call `lokstra.NewGlobalRegistrationContext()`.

You can define **your own** factory names for custom services (e.g., `"acme.redis"`, `"acme.metrics"`, etc.).

---

## YAML: Declaring Services

From `core/config/types.go`:

```go
type ServiceConfig struct {
	Name      string         `yaml:"name"`
	Type      string         `yaml:"type"`
	Config    map[string]any `yaml:"config"`
	DependsOn []string       `yaml:"depends_on,omitempty"`
}
```

The loader will **topologically order** services by `depends_on` and create them in dependency order:
```go
// in (cfg *LokstraConfig) StartAllServices(regCtx)
if _, err := regCtx.CreateService(svc.Type, svc.Name, svc.Config); err != nil {{ /* ... */ }}
```
(Detects cycles and errors out.)

Modules can also **create services** or **register factories** via `ModuleConfig` (when loading plugins).

---

## Error Handling

Use the provided helpers (from `core/service/service.go`) when writing factories or resolving services:

- `service.ErrUnsupportedConfig(config any)`
- `service.ErrInvalidServiceType(serviceName, expectedType string)`
- `service.ErrServiceNotFound(serviceName string)`

Examples:
```go
func NewRedisFactory(cfg any) (service.Service, error) {{
    m, ok := cfg.(map[string]any)
    if !ok {{ return nil, service.ErrUnsupportedConfig(cfg) }}

    url, _ := m["url"].(string)
    if url == "" {{ return nil, fmt.Errorf("missing 'url'") }}

    return myredis.New(url), nil // should satisfy serviceapi.Redis
}}

func UseMetrics(reg registration.Context) error {{
    m, err := reg.GetService("metrics.main")
    if err != nil {{ return service.ErrServiceNotFound("metrics.main") }}

    metrics, ok := m.(serviceapi.Metrics)
    if !ok {{ return service.ErrInvalidServiceType("metrics.main", "serviceapi.Metrics") }}

    metrics.IncCounter("boot.success", nil)
    return nil
}}
```

---

## Patterns & Tips

- Prefer **factories** for all services so they can be created from YAML or code uniformly.
- Keep **service names** unique and descriptive (e.g., `db.main`, `redis.cache`, `listener.admin`).
- When exposing an engine-like service (listener/router), follow the **prefix** conventions for discoverability.
- Use `serviceapi.GetService[T]` for **type safety** at call sites.
- When reading service references from mixed config values, use `registration.GetServiceFromConfig[T]`.
- For RPC services, pass either a **service name** or a concrete instance to `Router.MountRpcService`; when using a name, the router will resolve to `service.Service` at build time and mount an appropriate `RpcServer`.
