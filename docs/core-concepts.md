# Core Concepts

Lokstra is designed around **structured but flexible building blocks**.  
This document provides a **high-level overview** of the core concepts.  
For detailed specifications, see the linked documents in this guide.

---

## 🔑 Registration Context

The **Registration Context** is the heart of Lokstra.  
It manages services, handlers, middleware, and modules.

- Acts as a **dependency injection container**.  
- Services can be **registered, created via factory, or retrieved by name**.  
- Handlers and middleware are registered by **name** with priority ordering.  
- Modules (Go code or `.so` plugin) can register multiple components at once.  

> Quick start:
>
> ```go
> ctx := lokstra.NewGlobalRegistrationContext()
> ```
>
> This creates a global context, registers default modules, and prepares the default logger.

👉 See **[Registration](registration.md)** for full API details.

---

## 🛠 Request Context

Every HTTP request is handled with a **RequestContext** (abbrev. `ctx`). It wraps Go’s context and carries helpers for binding and responding.

- Wraps `context.Context`
- Embeds a `Response` object for unified response handling
- Provides **binding** from path, query, header, and body using **struct tags**  
  *(tags are **required** — fields without tags won’t be auto-bound).*

### Binding overview
- Standard:
  - `BindQuery(&dto)` → query string
  - `BindPath(&dto)` → route params
  - `BindHeader(&dto)` → HTTP headers
  - `BindBody(&dto)` → **JSON only**
- Combined:
  - `BindAll(&dto)` → Path + Query + Header + **Body (JSON only)**
- Smart:
  - `BindBodySmart(&dto)` → auto-detect: JSON, form-urlencoded, multipart, text
  - `BindAllSmart(&dto)` → Path + Query + Header + **Body (smart)**

Struct tag rules: `path:"..."`, `query:"..."`, `header:"..."`, `body:"..."`.  
Type-aware conversion supports primitives, `time.Time`, decimals, slices, and nested structs.

👉 See **[Request Context](request-context.md)** for binding rules and type conversion.

---

## 🧰 Handlers (Return `error`)

Lokstra handlers consistently **return `error`**. You can write handlers in two styles:

### 1) Manual binding
```go
type UserRequest struct {
    ID    string `path:"id"`
    Token string `header:"Authorization"`
    Name  string `body:"name"`
}

func createUser(ctx *lokstra.RequestContext) error {
    var req UserRequest
    if err := ctx.BindAllSmart(&req); err != nil {
        // Returning raw err would become 500.
        // For 400 Bad Request, use a response helper:
        return ctx.ErrorBadRequest(err.Error())
    }
    return ctx.Ok("Created user " + req.Name) // Ok(...) returns nil (already handled)
}
```

### 2) Generic handler with auto-binding
```go
func createUser(ctx *lokstra.RequestContext, req *UserRequest) error {
    // Framework auto-binds *req from path/query/header/body based on tags.
    // On bind/validation error it prepares an appropriate response.
    return ctx.Ok("Created user " + req.Name)
}
```

> **Note**  
> Returning a plain `error` signals an internal failure and results in **HTTP 500**.  
> Use response helpers (e.g., `ctx.ErrorBadRequest(...)`) to return domain/validation errors.

👉 See **[Request Context](request-context.md)** for generic handler details.

---

## 📤 Response

Responses are **structured objects**, not just raw JSON:

- Include `ResponseCode`, `Message`, and `Success` flag
- Helpers for success, error, **pagination**, **HTMX page-data**, and **raw** responses
- Method chaining: `.WithMessage`, `.WithHeader`, `.WithResponseCode`

Common helpers:
- Success: `Ok`, `OkCreated`, `OkNoContent`, `OkPagination(items, total, page, pageSize)`
- Errors: `ErrorBadRequest`, `ErrorUnauthorized`, `ErrorForbidden`, `ErrorNotFound`, `ErrorConflict`, `ErrorInternal`
- HTMX: `HtmxPageData(data)`
- Raw: `RawResponse([]byte, contentType)`, `RawStream(io.Reader, contentType)`

👉 See **[Response](response.md)** for the complete reference.

---

## 🌐 Server, App, Router

- **Server** → runs multiple Apps in a single process  
- **App** → an HTTP application bound to a port, router, and middleware stack  
- **Router** → manages routes, groups, middleware, and mounts  

Deployment models:
- **Monolithic** → one App handles all features
- **Microservice** → each App runs separately
- **Hybrid** → mix of both

👉 See **[Server & App](server-app.md)** and **[Router](router.md)**.

---

## ⚙️ Services, Middleware, Modules

- **Service** → reusable building block (e.g., dbpool, logger, metrics)  
- **Middleware** → request pipeline, ordered by priority (1–100)  
- **Module** → a bundle of services, handlers, and middleware as a feature package  

👉 See **[Services](services.md)**, **[Middleware](middleware.md)**, and **[Modules](modules.md)**.

---

## 📦 Configuration (YAML)

Lokstra applications can be fully described in YAML:

- Start services & register middleware
- Load feature modules
- Mount static directories, HTMX apps, or reverse proxies
- Create servers & apps

👉 See **[YAML Configuration](yaml-config.md)**.

---

## 🚏 Router Mounts

The router supports multiple mounting strategies:

- `MountStatic(dir, prefix)` → serve static files  
- `MountHtmx(dir)` → HTMX pages with layout + page-data convention  
- `MountReverseProxy(target)` → forward requests to another backend  

👉 See **[Router](router.md)**.

---

## 🔄 Flow DSL

The **Flow DSL** reduces boilerplate for CRUD endpoints:

```go
flow.NewHandler(ctx).
    RequiredService("db.main").
    BindBodySmart(&req).
    Validate().
    ExecSQL("INSERT INTO users ...").
    RespondJSON()
```

👉 See **[Flow DSL](flow-dsl.md)**.

---

## ✅ Summary

Lokstra’s building blocks:
- **Registration Context** → central DI for services, handlers, middleware, modules  
- **RequestContext** → unified request handling with standard & smart binding (tags required)  
- **Handlers** → always `error`-returning; use response helpers for non-500 outcomes  
- **Response** → structured helpers incl. Pagination, HTMX, and RawResponse  
- **Server–App–Router** → flexible deployment architecture  
- **Services, Middleware, Modules** → reusable features  
- **YAML Config** → declarative application setup  
- **Router Mounts** → static, HTMX, reverse proxy support  
- **Flow DSL** → faster CRUD without losing flexibility
