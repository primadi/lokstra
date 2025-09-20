
# Modules

**Source-driven:** This page reflects the actual module system in `lokstra-0.2.1` (`core/registration`, `core/config`, `defaults/*`, `middleware/*`).

A **Module** is a small plugin that registers things during **bootstrap**:
- services / service factories
- handlers
- middleware

Modules can live **in-process** (compiled together) or as **Go plugins** (`.so`) loaded at runtime.

---

## Module Interface

```go
type Module interface {
	Name() string
	Description() string
	Register(regCtx Context) error
}

// registration.Context is used only during startup phase
// to
```

Implement `Name()`, `Description()`, and `Register(regCtx)` to perform registrations.

Example (from built-ins):
```go
// middleware/body_limit/module.go
type BodyLimitModule struct{}

func (b *BodyLimitModule) Name() string        { return "body_limit" }
func (b *BodyLimitModule) Description() string { return "Body limit middleware" }

func (b *BodyLimitModule) Register(regCtx registration.Context) error {
    return regCtx.RegisterMiddlewareFactory("body_limit", factory)
}

func GetModule() registration.Module { return &BodyLimitModule{} }
```

Another example with service usage inside Register:
```go
// middleware/request_logger/request_logger.go
type RequestLogger struct{}

func (r *RequestLogger) Name() string        { return "request_logger" }
func (r *RequestLogger) Description() string { return "Logs incoming requests..." }

func (r *RequestLogger) Register(regCtx registration.Context) error {
    if svc, err := regCtx.GetService("logger"); err == nil {
        logger = svc.(serviceapi.Logger) // cache for middleware usage
    }
    return regCtx.RegisterMiddlewareFactoryWithPriority("request_logger", factory, 20)
}

func GetModule() registration.Module { return &RequestLogger{} }
```

> `lokstra.NewGlobalRegistrationContext()` registers **all default modules** (`defaults.RegisterAll(ctx)`), then retrieves the default `logger` service for you.

---

## Loading Modules

You can register modules in three ways via the **Registration Context**:

### 1) In-process (no plugin)
```go
reg.RegisterModule(body_limit.GetModule) // getModuleFunc: func() Module
```

### 2) Compiled plugin with default entry
```go
// Looks up symbol 'GetModule' in the .so
err := reg.RegisterCompiledModule("/path/to/plugin.so")
```

### 3) Compiled plugin with custom entry
```go
// Use a different exported function name to obtain Module
err := reg.RegisterCompiledModuleWithFuncName("/path/to/plugin.so", "MyEntry")
```

**Under the hood** (`plugin.Open` / `Lookup`):
- The loader loads the shared object.
- Resolves the entry function (`GetModule` by default).
- Calls it to obtain a `registration.Module`, then calls `Module.Register(regCtx)`.

---

## Permissioned Contexts

When loading modules from YAML, Lokstra creates a **permissioned sub-context** for each module:

```go
// signature on registration.Context
NewPermissionContextFromConfig(settings map[string]any, permission map[string]any) Context
```

Permission model (from `context_permission.go`):
```go
package registration

import "maps"

type PermissionRequest struct {
	WhitelistGetService []string

	AllowRegisterHandler    bool
	AllowRegisterMiddleware bool
	AllowRegisterService    bool

	ContextSettings map[string]any
}

type PermissionGranted struct {
	whitelistGetService []string

	allowRegisterHandler    bool
	allowRegisterMiddleware bool
	allowRegisterService    bool

	contextSettings map[string]any
}

func newPermissionGranted(req *PermissionRequest) *PermissionGranted {
	return &PermissionGranted{
		whitelistGetService: req.WhitelistGetService,

		allowRegisterHandler:    req.AllowRegisterHandler,
		allowRegisterMiddleware: req.AllowRegisterMiddleware,
		allowRegisterService:    req.AllowRegisterService,

		contextSettings: maps.Clone(req.ContextSettings),
	}
}

func (p *PermissionGranted) IsAllowedGetService(name string) bool {
	// allow full match or prefix match with wildcard
	for _, whitelisted := range p.whitelistGetService {
		if whitelisted == name {
			return true
		}
		lw := len(whitelisted)
		if lw > 0 && whitelisted[lw-1] == '*' {
			prefix := whitelisted[:lw-1]
			if len(name) >= len(prefix) && name[:len(prefix)] == prefix {
				return true
			}
		}
	}
	return false
}

func (p *PermissionGranted) IsAllowedRegisterHandler() bool {
	return p.allowRegisterHandler
}

func (p *PermissionGranted) IsAllowedRegisterMiddleware() bool {
	return p.allowRegisterMiddleware
}

func (p *PermissionGranted) IsAllowedRegisterService() bool {
	return p.allowRegisterService
}

func (p *PermissionGranted) GetContextSettings() map[string]any {
	return p.contextSettings
}
```

- `WhitelistGetService` (`permissions.get_service` in YAML): list of service names allowed (`"*"` = all).
- `AllowRegisterHandler` / `AllowRegisterMiddleware` / `AllowRegisterService`: enable those actions.
- `ContextSettings`: a map copied into the context; modules can read it via `regCtx.GetSetting(...)`.

The context **enforces** permissions:
- `GetService` → denies unless service name is whitelisted.
- `RegisterHandler` / `RegisterMiddlewareFactory` / `RegisterServiceFactory` → deny unless allowed; the code **panics** with a clear message.

---

## YAML: Declaring Modules

```go
type ModuleConfig struct {
	Name        string         `yaml:"name"`
	Path        string         `yaml:"path"`
	Entry       string         `yaml:"entry,omitempty"`
	Settings    map[string]any `yaml:"settings,omitempty"`
	Permissions map[string]any `yaml:"permissions,omitempty"`

	RequiredServices         []string        `yaml:"required_services,omitempty"`
	CreateServices           []ServiceConfig `yaml:"create_services,omitempty"`
	RegisterServiceFactories []string        `yaml:"register_service_factories,omitempty"` // list of method names
	RegisterHandlers         []string        `yaml:"register_handlers,omitempty"`          // list of method names
	RegisterMiddleware       []string        `yaml:"register_middleware,omitempty"`        // list of method names
}
```

### Loader sequence (`StartAllModules`)

For each module entry:

1) **Register the plugin** (if `path` is set) under a **permissioned context** created from `{settings, permissions}`.  
2) **Check** `required_services` exist (by name) in the **global** context.  
3) **Create services** listed in `create_services` (ordinary `ServiceConfig` entries).  
4) **Call exported methods** from the plugin:
   - `register_service_factories`: each method must be `func(registration.Context) error` and should register factories
   - `register_handlers`: same signature, should register named handlers
   - `register_middleware`: same signature, should register middleware
   Each method name is looked up via `plugin.Open(...).Lookup(name)` and then invoked.

If any step fails → loader returns an error with the module name for diagnostics.

---

## Typical Patterns

### A) Pure module (only Register)
- Expose `GetModule() registration.Module`
- In `Register`, call `reg.RegisterMiddlewareFactory(...)` or `reg.RegisterServiceFactory(...)`, etc.
- Optionally access whitelisted services via `reg.GetService("...")`

### B) Hybrid: Register + Extra Hooks
- Keep your `Register` simple (baseline factories/handlers).
- Add optional hook functions exported from the plugin and list them in YAML under the right key:
  - `register_service_factories`, `register_handlers`, `register_middleware`
- These hooks run **after** module registration and after `required_services` validation.

### C) Permission scoping
Grant the **minimum** permissions needed by the module:
```yaml
modules:
  - name: analytics
    path: ./modules/analytics/analytics.so
    permissions:
      get_service: ["logger", "metrics.*"]   # or ["*"]
      allow_register_handler: false
      allow_register_middleware: true
      allow_register_service: true
    settings:
      sample_rate: 0.1
```

Inside the module you can read settings via `regCtx.GetSetting("sample_rate")` (string/number/etc.).

---

## Defaults

`lokstra.NewGlobalRegistrationContext()` calls:
```go
ctx := registration.NewGlobalContext()
defaults.RegisterAll(ctx)
Logger, _ = lokstra.GetService[serviceapi.Logger](ctx, "logger")
```
So you start with a registry that already has common middleware modules registered (CORS, recovery, gzip, body_limit, request_logger, slow_request_logger) and default HTTP listener/router factories.

---

## Troubleshooting

- **"module with name 'X' already registered"** — you called `RegisterModule` twice for the same module name.
- **"service 'X' is not allowed to be accessed"** — check module permissions (`get_service` whitelist).
- **panic: registering handler/middleware/service is not allowed** — enable the corresponding permission in YAML.
- **"method <name> not found in plugin"** — verify the function is exported by the `.so` and signatures match `func(registration.Context) error`.
- **Plugin load errors** — ensure the plugin was built with `-buildmode=plugin` and the Go version matches the host.

---

## TL;DR

- A **Module** is `Name + Description + Register(Context)`.
- Load via `RegisterModule` (in-proc) or `RegisterCompiledModule*(.so)` (plugins).
- Use **permissioned contexts** per module (from YAML `{settings, permissions}`) to sandbox capabilities.
- Defaults are pre-registered when using `NewGlobalRegistrationContext()`.
