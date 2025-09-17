
# Middleware

**Source-driven:** This page reflects the real API & behavior in `lokstra-0.2.1` (`core/midware`, `core/router`).  
Middleware can be **registered** globally (by name) or **inlined** per router/group/route.

---

## Core Types (from source)

```go
package midware

type Execution struct {
	Name   string
	Config any // Configuration for the middleware

	MiddlewareFn   Func // Function to create the middleware
	Priority       int  // Lower number means higher priority (1-100)
	ExecutionOrder int  // Order of execution, lower number means earlier execution
}

func NewExecution(fn Func) *Execution {
	return &Execution{MiddlewareFn: fn, Priority: 5000}
}
```

```go
package midware

import "github.com/primadi/lokstra/core/request"

type Func = func(next request.HandlerFunc) request.HandlerFunc
type Factory = func(config any) Func

func Named(name string, config ...any) *Execution {
	var cfg any
	switch len(config) {
	case 0:
		cfg = nil // No config provided, use nil
	case 1:
		cfg = config[0] // Single config provided
	default:
		cfg = config // Multiple configs provided, use as is
	}
	return &Execution{Name: name, Config: cfg, Priority: 5000}
}
```

- `midware.Func` is the standard middleware signature: it *wraps* a `request.HandlerFunc` and returns a new one.
- `midware.Factory` builds a `midware.Func` from a `config` payload (`any`).
- `Execution` is an internal carrier used by the router:
  - `Name` (for named factories), `Config` (passed into the factory)
  - `MiddlewareFn` (the actual function), `Priority` (lower runs earlier), `ExecutionOrder` (insertion order)

Helpers:
- `midware.Named(name, config...)` → make an `*Execution` referencing a **registered factory** by `name`.  
  If multiple `config` args are given, they are passed to the factory as a slice (kept as-is).
- `midware.NewExecution(fn midware.Func)` → wrap a function directly (no factory).

> Defaults: `RegisterMiddlewareFactory(...)` uses **priority 50**; `Named(...)` and `NewExecution(...)` start with a **placeholder** priority `5000` and will be assigned the real priority on resolve (for named) or stay `5000` (for raw functions).

---

## Registering Middleware (global)

Register global **middleware factories** in the Registration Context (see `registration.md`). Factories may be **prioritized**.

```go
// Default priority 50
reg.RegisterMiddlewareFactory("cors", func(cfg any) midware.Func {
    return func(next request.HandlerFunc) request.HandlerFunc {
        return func(ctx *request.Context) error {
            // ... set CORS headers using cfg ...
            return next(ctx)
        }
    }
})

// Or with explicit priority (1 = earliest)
reg.RegisterMiddlewareFactoryWithPriority("auth", authFactory, 10)

// You can also register a plain function (wrapped into a factory) with default priority 50
reg.RegisterMiddlewareFunc("audit", func(next request.HandlerFunc) request.HandlerFunc {
    return func(ctx *request.Context) error { /* ... */; return next(ctx) }
})
```

Lookup during build:
```go
func resolveMiddleware(ctx registration.Context, mw *midware.Execution) {
	if mw.MiddlewareFn == nil {
		mwFactory, priority, found := ctx.GetMiddlewareFactory(mw.Name)
		if !found {
			panic(fmt.Sprintf("Middleware factory '%s' not found", mw.Name))
		}
		mw.MiddlewareFn = mwFactory(mw.Config)
		mw.Priority = priority
	}
}

func getOrCreateService[T any](ctx registration.Context,
	serviceName str
```

If a named factory is not found → **panic** with a clear message.

---

## Attaching Middleware (router/group/route)

Every `Use(...)`, `Group(..., mw...)`, or `Handle(..., mw...)` call accepts:
- `string` (named factory)
- `midware.Func`
- `*midware.Execution` (e.g., `midware.Named("name", cfg)`)

Example:
```go
api := r.Group("/api", "auth")            // string -> named factory
api.Use("audit")
api.GET("/profile", getProfile, "audit")  // attach to a route
api.Use(func(next request.HandlerFunc) request.HandlerFunc {
    return func(ctx *request.Context) error { /* inline function */; return next(ctx) }
})
```

Route builder converts the inputs:
```go
// from core/router/router_meta.go::Handle(...)
case midware.Func:        mw = midware.NewExecution(m)
case string:              mw = midware.Named(m)
case *midware.Execution:  mw = m
default:                  panic("Invalid middleware type")
```

---

## Priority & Execution Order

When building the router, Lokstra sorts middleware by **`Priority + ExecutionOrder`** ascending and **wraps from inside → out**:

```go
func composeMiddleware(mw []*midware.Execution,
	finalHandler request.HandlerFunc) request.HandlerFunc {
	// Update execution order based on order of addition
	execOrder := 0
	for _, m := range mw {
		m.ExecutionOrder = execOrder
		execOrder++
	}

	// Sort middleware by priority and execution order
	slices.SortStableFunc(mw, func(a, b *midware.Execution) int {
		aOrder := a.Priority + a.ExecutionOrder
		bOrder := b.Priority + b.ExecutionOrder

		if aOrder < bOrder {
			return -1
		} else if aOrder > bOrder {
			return 1
		}

		return 0
	})

	// Compose middleware functions in reverse order
	handler := finalHandler
	for i := len(mw) - 1; i >= 0; i-- {
		handler = mw[i].MiddlewareFn(handler)
	}
	return handler
}

// ComposeMiddlewareForTest exposes composeMiddleware for testing
func ComposeM
```

- **Priority**: lower runs earlier (e.g., `10` runs before `50`).  
  - Named factories use the priority set at registration (default `50`).
  - **Inline functions** (not from a factory) keep `5000`, so they run **after** any named middleware by default.
- **ExecutionOrder**: the insertion order under the same node (router/group/route). Earlier added ⇒ earlier.

**Override rules**:
- At **router/group** level: `WithOverrideMiddleware(true)` makes children **not inherit** parent middleware.  
- At **route** level: `HandleOverrideMiddleware(...)` ignores parent chain and uses **only** the route's middlewares.

---

## Reverse Proxy Specifics

Reverse proxy mounts run middleware with a slightly different composition so that **errors short-circuit the chain** and write a Lokstra JSON response, while **success** lets the upstream write directly to `http.ResponseWriter`:

```go
func composeReverseProxyMw(rp *ReverseProxyMeta, mwParent []*midware.Execution) http.HandlerFunc {
	var mw []*midware.Execution

	if rp.OverrideMiddleware {
		mw = make([]*midware.Execution, len(rp.Middleware))
		copy(mw, rp.Middleware)
	} else {
		mw = utils.SlicesConcat(mwParent, rp.Middleware)
	}

	// Update execution order based on order of addition
	execOrder := 0
	for _, m := range mw {
		m.ExecutionOrder = execOrder
		execOrder++
	}

	// Sort middleware by priority and execution order
	slices.SortStableFunc(mw, func(a, b *midware.Execution) int {
		aOrder := a.Priority + a.ExecutionOrder
		bOrder := b.Priority + b.ExecutionOrder

		if aOrder < bOrder {
			return -1
		} else if aOrder > bOrder {
			return 1
		}

		return 0
	})

	// Create the final proxy handler
	proxyHandler := createReverseProxyHandler(rp.Target)

	// Wrap proxy handler with error-aware middleware composition
	// Start from the innermost handler (proxy) and wrap outward
	handler := proxyHandler
	for i := len(mw) - 1; i >= 0; i-- {
		currentMw := mw[i]
		handler = func(innerHandler request.HandlerFunc, middleware *midware.Execution) request.HandlerFunc {
			return middleware.MiddlewareFn(func(ctx *request.Context) error {
				// Check if previous middleware already set an error response (4xx or 5xx)
				if ctx.Response.StatusCode >= 400 {
					// Return error to stop middleware chain and prevent "after"
```

This ensures:
- If a middleware sets an error (e.g., `ctx.ErrorUnauthorized(...)`), the chain stops and the structured error is sent.
- On success, the proxy handler streams the upstream response; Lokstra **does not** re-encode JSON.

---

## Common Patterns

### 1) Configurable named middleware
```go
type CorsCfg struct {{ AllowOrigins []string }}

func corsFactory(cfg any) midware.Func {{
    c, _ := cfg.(CorsCfg)  // or map[string]any
    return func(next request.HandlerFunc) request.HandlerFunc {{
        return func(ctx *request.Context) error {{
            // apply headers based on c
            return next(ctx)
        }}
    }}
}}

reg.RegisterMiddlewareFactoryWithPriority("cors", corsFactory, 30)

r.Use(midware.Named("cors", CorsCfg{{AllowOrigins: []string{{"*" }} }}))
// or r.Use("cors") if the factory uses defaults
```

### 2) Before/after hooks
```go
reg.RegisterMiddlewareFunc("audit", func(next request.HandlerFunc) request.HandlerFunc {{
    return func(ctx *request.Context) error {{
        start := time.Now()
        err := next(ctx)
        duration := time.Since(start)

        // record duration regardless of err
        return err
    }}
}})
```

### 3) Role-based auth
```go
reg.RegisterMiddlewareFactoryWithPriority("require_role", func(cfg any) midware.Func {{
    role := cfg.(string)
    return func(next request.HandlerFunc) request.HandlerFunc {{
        return func(ctx *request.Context) error {{
            if !hasRole(ctx, role) {{
                return ctx.ErrorForbidden("forbidden")
            }}
            return next(ctx)
        }}
    }}
}}, 10)

api.Use(midware.Named("require_role", "admin"))
```

---

## Troubleshooting

- **"Middleware factory 'X' not found"** → ensure you registered it in the same `registration.Context` used to build the router.
- **Unexpected order** → check priorities and where you attach `Use(...)` (router vs group vs route). Also remember inline functions default to priority **5000**.
- **Reverse proxy not writing JSON on error** → ensure your middleware sets a proper error via `ctx.ErrorBadRequest(...)`/`ErrorUnauthorized(...)`, etc., or returns a non-nil error; the proxy wrapper checks both.

---

## TL;DR

- Register **named factories** with priorities (1 = earliest).  
- Attach middleware by **name**, **function**, or **Execution**.  
- Order = `Priority + ExecutionOrder`; inline functions (5000) typically run **last**.  
- Use **override** to isolate stacks per subtree or route.  
- Reverse proxy composition gracefully handles **success passthrough** vs **error JSON**.
