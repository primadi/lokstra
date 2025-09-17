
# RPC

**Source-driven:** This page documents Lokstra RPC from the real sources in `lokstra-0.2.1`:
- `serviceapi/rpc_server.go` (interface)
- `core/router/*` (mount & runtime wiring)
- `modules/rpc_service/*` (default implementation)
- `core/service/rpcservice_meta.go` (handler extension)

Lokstra RPC is a **HTTP + MessagePack** reflection-based RPC:
- You *mount a service* at a base path.
- Each method is invoked by **POST** to {base}/:method with **msgpack**-encoded `[]any` as the request body.
- The response is **msgpack** for successful returns, or a structured Lokstra error JSON on failure.

---

## Interfaces & Types

### RPC server interface
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

### Handler/Extension metadata
```go
package request

// HandlerMeta represents a named handler.
// Can be a direct function or resolved later by name.
type HandlerMeta struct {
	Name        string
	HandlerFunc HandlerFunc
	Extension   any // Optional extension for the handler, currently used for *RpcServiceMeta
}

// Used as handler extension when mounting RPC:
package service

type RpcServiceMeta struct {
	MethodParam string // default "method"
	ServiceName string
	ServiceInst Service
}
```

---

## Mounting RPC

You mount an RPC service at a path on any `Router` (root, group, or app-level):

```go
// path: base URL; svc: service instance OR name OR *service.RpcServiceMeta
// overrideMiddleware: ignore inherited middleware at this node
// mw ...any: additional middleware (string, midware.Func, or *midware.Execution)
Router.MountRpcService(path string, svc any, overrideMiddleware bool, mw ...any)
```

Examples:
```go
// 1) With a concrete service instance
app.MountRpcService("/rpc", myService, false)

// 2) Refer to a named service (resolved from Registration Context)
app.MountRpcService("/rpc", "svc.greeting", false)

// 3) Advanced: custom method param key (default is "method")
app.MountRpcService("/rpc", &service.RpcServiceMeta{
    ServiceName: "svc.greeting",
    MethodParam: "action",
}, false)
```

The router will create a **POST route** at {clean(path)}/:method (or `:action` if you set `MethodParam`).

Wiring (from source):
```go
for _, rpc := range router.RPCHandles {
		cleanPath := r.cleanPrefix(rpc.Path)
		if strings.HasSuffix(cleanPath, "/") {
			cleanPath += ":method"
		} else {
			cleanPath += "/:method"
		}

		rpcMeta := &service.RpcServiceMeta{
			MethodParam: "method",
		}
		switch s := rpc.Service.(type) {
		case string:
			rpcMeta.ServiceName = s
		case *service.RpcServiceMeta:
			rpcMeta = s
		case service.Service:
			rpcMeta.ServiceInst = s
		default:
			fmt.Printf("Service type: %T\n", rpc.Service)
			panic("Invalid service type, must be a string, *RpcServiceMeta, or iface.Service")
		}

		handlerMeta := &request.HandlerMeta{
			HandlerFunc: func(ctx *request.Context) error {
				return ctx.ErrorInternal("RpcService not yet resolved")
			},
			Extension: rpcMeta,
		}

		if rpc.OverrideMiddleware {
			r.meta.Handle("POST", cleanPath, handlerMeta, true, rpc.Middleware...)
		} else {
			r.Handle("POST", cleanPath, handlerMeta, rpc.Middleware...)
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

func (r *RouterImpl) BuildRouter() {
	r.buildRouter(r.meta, nil)
}

func (r *RouterImpl) GetEngine() serviceapi.RouterEngine {
	return r.r_engine
}

func (r *RouterImpl) SetEngine(engine serviceapi.RouterEngine) {
	r.r_engine = engine
}

func composeMiddleware(mw []
```

At build-time (`ResolveAllNamed`), Lokstra resolves the service and the RPC server, then sets the concrete handler:

```go
ar err error
				svc, err = ctx.GetService(rpcServiceMeta.ServiceName)
				if err != nil {
					panic(fmt.Sprintf("Rpc Service '%s' not found", rpcServiceMeta.ServiceName))
				}
				r.Routes[i].Handler.Extension.(*service.RpcServiceMeta).ServiceInst = svc
			}
			rpcSvr, err := getOrCreateService[serviceapi.RpcServer](ctx, "rpc_server.default",
				"rpc_service.rpc_server")
			if err != nil {
				panic(fmt.Sprintf("Failed to get RPC server: %s", err.Error()))
			}

			r.Routes[i].Handler.HandlerFunc = func(ctx *re
```

> **Note:** The code attempts to **get or create** a service named `"rpc_server.default"`.
> If it's missing, it will try to `CreateService` using factory name **`"rpc_service.rpc_server"`**.
> You can either:
> - **Pre-create** the service `"rpc_server.default"` yourself (e.g., factory `"rpc_service"` or your own).
> - Or **register** a factory alias under `"rpc_service.rpc_server"` that returns a `serviceapi.RpcServer` (default implementation is in `modules/rpc_service`).

Default module & server:
```go
// modules/rpc_service/module.go
package rpc_service

import (
	"github.com/primadi/lokstra/core/registration"
)

const NAME = "rpc_service"

type module struct{}

// Description implements registration.Module.
func (r *module) Description() string {
	return "RPC Service Module provides RPC service functionality"
}

// Name implements registration.Module.
func (r *module) Name() string {
	return NAME
}

// Register implements registration.Module.
func (r *module) Register(regCtx registration.Context) error {
	// Register the RPC service factory
	regCtx.RegisterServiceFactory(r.Name(), NewRpcServer)

	return nil
}

var _ registration.Module = (*module)(nil)

func GetModule() registration.Module {
	return &module{}
}

// modules/rpc_service/rpc_server.go
package rpc_service

import (
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/serviceapi"
)

type RpcServerImpl struct{}

func NewRpcServer(_ any) (service.Service, error) {
	return &RpcServerImpl{}, nil
}

// HandleRequest implements serviceapi.RpcServer.
func (r *RpcServerImpl) HandleRequest(ctx *request.Context,
	svc service.Service, MethodName string) error {
	return HandleRpcRequest(ctx, svc, MethodName)
}

var _ service.Service = (*RpcServerImpl)(nil)
var _ serviceapi.RpcServer = (*RpcServerImpl)(nil)
```

---

## Method Binding & I/O

Lokstra discovers methods on your **service instance** via reflection. Method names are matched **case-insensitively** (the path param is lower-cased internally).

**Request body:** msgpack-encoded array `[]any` for arguments, in order.  
**Return types supported:** either
- `(error)` — only error; success → no payload
- `(T, error)` — any serializable value `T` (struct, map, slice, primitive, time, etc.)

Server-side execution (key fragment):
```go
// Step 3: Call actual method
	result := mm.Func.Call(in)

	// Step 4: Encode response
	switch len(result) {
	case 1:
		if !result[0].IsNil() {
			errVal := result[0].Interface().(error)
			return ctx.ErrorInternal(errVal.Error())
		}
		ctx.WithHeader("Content-Type", "application/octet-stream")
		return ctx.Ok(nil) // No content
	case 2:
		if !result[1].IsNil() {
			errVal := result[1].Interface().(error)
			return ctx.ErrorInternal(errVal.Error())
		}
		respData, err := msgpack.Marshal(result[0].Interface())
		if err != nil {
			return ctx.ErrorInternal("encoding error")
		}

		return ctx.WriteRaw("application/octet-stream", 200, respData)
	default:
		return ctx.ErrorInternal("unexpected number of return values")
	}
}

func (sm *serviceMeta) HandleRpcRequest(ctx *request.Context, methodName string) error {
	method := sm.Methods[strings.ToLower(methodName)]
	if method == nil {
		return c
```

Behavior:
- **Decode errors / arg count mismatch** → `400 Bad Request`.
- **Method not found** → `400 Bad Request`.
- **Method returned error** → `500 Internal` via `ctx.ErrorInternal(...)`.
- **Success with data** → `200 OK` and `application/octet-stream` body containing msgpack-encoded `T`.
- **Success without data** → `200 OK` and `application/octet-stream` with empty payload.

---

## Middleware & Override

`MountRpcService(..., overrideMiddleware, mw...)` follows the same rules as normal routes:
- Attach per-call **middleware** (`string`, `midware.Func`, or `*midware.Execution`).
- If `overrideMiddleware == true`, the RPC endpoint **does not inherit** parent middlewares.
- Otherwise, it inherits and composes according to priority + insertion order (see `middleware.md`).

---

## YAML (Config)

You can declare RPC mounts in **App** or **Group** config:

```yaml
apps:
  - name: hello-service-app
    address: ":8080"
    routes: []
    mount_rpc_service:
      - base_path: /rpc
        service_name: svc.greeting     # must exist or be resolvable at build
```

Also ensure an RPC server exists (pick **one** approach):

**A) Pre-create the default by name**
```yaml
services:
  - name: rpc_server.default
    type: rpc_service           # or your own factory that returns serviceapi.RpcServer
    # config: {{}}              # optional factory config
```

**B) Provide a factory under the expected name (advanced)**
```go
// during bootstrap (module or code)
reg.RegisterServiceFactory("rpc_service.rpc_server", rpc_service.NewRpcServer)
```

If `"rpc_server.default"` is present, the router will **use it** and skip auto-creation.

---

## Client Helpers

The default module provides a small HTTP client using msgpack (`modules/rpc_service/rpc_client.go`). Typical usage pattern:

```go
c := rpc_service.NewRpcClient("http://localhost:8080/rpc")

// No return value
if err := rpc_service.CallReturnVoid(c, "Ping"); err != nil { /* ... */ }

// Typed return
sys, err := rpc_service.CallReturnType[SystemInfo](c, "GetSystemInfo")
```

Additional helpers exist for **interfaces** and **slices** (see examples under `cmd/examples/01_basic_overview/05_client_rpc`).

---

## Practical Checklist

- [ ] Mount your service via `MountRpcService("/rpc", ...)`.
- [ ] Ensure an `RpcServer` service exists (e.g., `"rpc_server.default"`).
- [ ] Send **POST** with **msgpack** array body: `[]any` of arguments.
- [ ] Expect **msgpack** response on success, Lokstra JSON on errors.
- [ ] Use middleware/override as needed (auth, rate-limit, audit).
