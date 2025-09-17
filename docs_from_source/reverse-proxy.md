
# Reverse Proxy

**Source-driven** for `lokstra-0.2.1` (from `core/router/*` and `modules/coreservice/router_engine/*`).  
Lokstra can mount a **reverse proxy** under a URL prefix and pass requests to an upstream server, while still letting you run **middleware** before/after the proxy call.

---

## API

```go
// Router (core/router/router.go)
MountReverseProxy mounts a reverse proxy at the specified prefix, targeting the given URL, with optional middleware and override option
	MountReverseProxy(prefix string, target string, overrideMiddlew overrideMiddleware bool, mw ...any) Router
```

- `prefix`: path prefix to intercept (e.g., `/api/`)
- `target`: upstream base URL (e.g., `https://api.github.com`)
- `overrideMiddleware`: if `true`, do **not** inherit parent middleware stack
- `mw ...any`: additional middlewares at this mount (string, `midware.Func`, or `*midware.Execution`)

**Metadata (build-time carrier):**
```go
type ReverseProxyMeta struct {
	Prefix             string
	Target             string
	OverrideMiddleware bool
	Middleware         []*midware.Execution
}

type RPCServiceMeta struct {
	Path               string
	Service  
```

---

## Runtime Wiring

At build time, for each reverse proxy mount the router composes middleware and registers a handler into the **router engine**:

```go
// router_impl.go
andlerMeta, rpc.Middleware...)
		}
	}

	for _, rp := range router.ReverseProxies {
		handler := composeReverseProxyMw(rp, mwh)
		r.r_engine.ServeReverseProxy(rp.Prefix, handler)
	}

	for _, sdf := range router.StaticMounts {
		r.r_engine.ServeStatic(sdf.Prefix, sdf.Spa, sdf.Sources...)
	}

	for _, htmx := range router.HTMXPages {
		r.r_engine.ServeHtmxPage(r.r_engine, htmx.Prefix, htmx.Script, htmx.Sources...)
	}
}


```

The engine exposes a `ServeReverseProxy` hook (supported by standard engines):

```go
// httprouter_engine.go
ServeReverseProxy(prefix string, handler http.HandlerFunc) {
	h.getServeMux().ServeReverseProxy(prefix, handler)
}

// RawHandle implements RouterEngine.
func (h *HttpRouterEngine) RawHandle(pattern string, handler http.

// servemux_engine.go
ServeReverseProxy(prefix string, handler http.HandlerFunc) {
	cleanPrefix := cleanPrefix(prefix)
	if cleanPrefix == "/" {
		m.mux.Handle("/", handler)
	} else {
		m.mux.Handle(cleanPrefix+"/", http.StripPrefix(cleanPrefix, handler))
	}
}

// RawHandle implements RouterEngine.
func (m *ServeMuxEngine) RawHandle(pattern string, handler http.Handler) {
	m.mux.Handle(pattern, handler)
}

// RawHandleFunc implements Route
```

The engine implementation uses the underlying mux (`http.ServeMux` wrapper) to route requests whose path starts with `prefix` to your proxy handler.

---

## Middleware Semantics

Reverse proxy mounts use a **special composition** so that:
- On **success**, the upstream writes **directly** to `http.ResponseWriter` (Lokstra does not re-encode JSON).
- On **error** (your middleware returns non-nil error or sets a response status ≥ 400), Lokstra writes a **standard JSON** response using the `Response` object.

```go
// composeReverseProxyMw (core/router/helper.go)
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
					// Return error to stop middleware chain and prevent "after" logic
					return errors.New("previous middleware set error response")
				}
				return innerHandler(ctx)
			})
		}(handler, currentMw)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, deferFunc := request.NewContext(w, r)
		defer deferFunc()

		// Execute the wrapped middleware chain
		err := handler(ctx)

		// Handle response writing
		if err != nil || ctx.Response.StatusCode >= 400 {
			// Error case - write lokstra response
			ctx.Response.WriteHttp(ctx.Writer)
		}
		// Success case - proxy has already written response directly to http.ResponseWriter
		// No need to call WriteHttp for successful proxy responses
	})
}

```

**Ordering:** same as regular routes → the chain is ordered by `Priority + ExecutionOrder`.  
Use `overrideMiddleware` to isolate the proxy from parent middlewares.

Typical uses:
- `auth` / `rate-limit` / `audit` before proxying
- Transform headers in a pre-proxy middleware (e.g., inject `Authorization`)

---

## YAML

You can mount proxies from YAML:

```go
type MountReverseProxyConfig struct {
	Prefix string `yaml:"prefix"`
	Target string `yaml:"target"`
}
```

In an app or group config:

```yaml
apps:
  - name: web
    address: ":8080"
    mount_reverse_proxy:
      - prefix: /api/
        target: https://httpbin.org
      - prefix: /github/
        target: https://api.github.com
```

---

## Example

From `cmd/examples/02_router_features/05_mount_reverse_proxy`:

```go
app.MountReverseProxy("/api/external/", "https://jsonplaceholder.typicode.com", false)
app.MountReverseProxy("/api/secure/",   "https://httpbin.org", false, "auth")
app.MountReverseProxy("/api/github/",   "https://api.github.com", false, "audit")
```

Then hit, e.g.:

```
curl http://localhost:8080/api/external/posts/1
curl http://localhost:8080/api/secure/get
curl http://localhost:8080/api/github/users/octocat
```

---

## Tips & Gotchas

- **Prefix matching**: use a **trailing slash** on the prefix (e.g., `/api/`) when proxying subpaths.
- **Error handling**: to return JSON errors from your middlewares, call helpers like `ctx.ErrorUnauthorized("...")` or return a non-nil error — the proxy wrapper handles it and **stops** calling the upstream.
- **Headers**: transform, add, or strip headers in middleware before the proxy runs. The proxy writes upstream headers to the response untouched on success.
- **Performance**: proxy path avoids extra allocations by streaming directly to `ResponseWriter` on success.
- **Testing locally**: combine with `MountStatic` for assets and reverse proxy for `/api/` to a dev backend.

