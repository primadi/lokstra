# Lokstra Framework Documentation

**Tagline:** *Simple. Scalable. Structured.*

Lokstra is a lightweight Go framework for building modular and scalable backend applications.  
This documentation is generated directly from the source code and runnable examples,  
making it the **single source of truth** for developers and AI agents (e.g., Copilot).

---

## 📖 Table of Contents

- [Getting Started](getting-started.md)
- [Core Concepts](core-concepts.md)
- [Bootstrap Project](bootstrap.md)
- [Router](router.md)
- [YAML Configuration](yaml-config.md)
- [Services](services.md)
- [Modules](modules.md)
- [Examples](examples/README.md)
- [Advanced Topics](advanced.md)

---

## Core Concepts

- **App** → HTTP application bound to a port, router, and middleware.  
- **Server** → Runs multiple Apps in a single process.  
- **Router** → Named handlers, route groups, and multi-level middleware. Supports `MountStatic` (serve files), `MountHtmx` (HTMX-driven HTML), and `MountReverseProxy` (proxy to backend services).
- **Registration / Dependency Injection** → Defined in `core/registration`. The `Context` interface (see `core/registration/context.go`) acts as a lightweight DI container:
  - `RegisterService(type, name, factory)` to provide service instances.
  - `GetService(typeOrName)` to resolve a service.
  - `RegisterHandler(name, handler)` and `GetHandler(name)` for named HTTP handlers.
  - A **GlobalContext** is provided for most apps.
- **Service (optional lifecycle)** → Services are plain Go structs created/resolved via DI. Lokstra does not mandate lifecycle hooks for all services, but any service that implements `service.Shutdownable` will be gracefully shut down when the server stops. Service factories remain responsible for constructing ready-to-use instances.
- **RequestContext** → See `core/request/context.go`. A handler receives `*RequestContext` which:
  - Embeds `context.Context` (from the incoming request).
  - Exposes `Writer http.ResponseWriter` and `Req *http.Request`.
  - Embeds `*response.Response` helpers so you can call `ctx.Ok(data)` etc.
- **Response** → See `core/response/response.go`. Unified response object:
  - Fields: `ResponseCode string`, `Message string`, `Data any`, `Success bool`, `Headers map[string][]string`.
  - `ResponseCode` maps to an HTTP status (computed), keeping handlers protocol-agnostic.
  - Writers: `WriteHttp`, `WriteStdout`, `WriteBuffer`, `WriteWebSocket` (where applicable).
  - Fluent helpers like `.WithMessage()`, `.WithHeader()`, and success/error shortcuts.
- **Helper Package (`lokstra`)** → Defined in `lokstra.go`. Provides aliases and factory functions to simplify the API surface:  
  - **Aliases**: `RegistrationContext`, `RequestContext`, `Response`, `App`, `Server`, `HandlerFunc`, `MiddlewareFunc`, `Service`, `ServiceFactory`, etc.  
  - **Functions**: `NewApp`, `NewServer`, `NewServerFromConfig`, `LoadConfigFile`, `LoadConfigDir`, `NamedMiddleware`, etc.  
  This allows developers to use Lokstra without importing multiple internal packages.
- **Flow** → Defined in `core/flow`.  
  A builder-style DSL to create request handlers:  
  - Bind request body/query/path to structs  
  - Validate input  
  - Access services (db, logger, redis, etc.)  
  - Execute SQL or business logic  
  - Respond with JSON or custom response  
  Flow is optional: you can mix it with manually written handlers.
- **Module** → Static or plugin code that can register handlers, middleware, and services.



## About Lokstra

Lokstra was created to solve the problem of balancing **simplicity, scalability, and structure** in Go projects.  
It enables developers to choose between **monolith**, **microservices**, or **hybrid deployment**, while reusing the same codebase.

### Key Features
- 🟢 **Apps & Server** → Run multiple HTTP apps in one server.  
- 🟢 **Router** → Named handlers, groups, middleware, static mount, HTMX, reverse proxy.  
- 🟢 **Config** → YAML-based, with variable interpolation (`${ENV:default}`).  
- 🟢 **Services** → Reusable components (logger, metrics, worker, session, dbpool, redis).  
- 🟢 **Modules** → Static or plugin-based, register handlers/middleware/services.  
- 🟢 **Developer Experience** → Runnable examples in `/cmd/examples`.  

---

## Relation to Code

**Key files:**  
- `core/registration/context.go` — DI context and registration APIs.  
- `core/request/context.go` — RequestContext used by handlers.  
- `core/response/response.go` — Response object and writers.

Documentation always follows **code as source of truth**:  
- `/core/router` → Router features.  
- `/core/config` → Configuration & YAML mapping.  
- `/core/service` → Service lifecycle.  
- `/middleware` → Built-in middleware modules.  
- `/cmd/examples` → Usage examples.  

---

## Next Steps

- Start with [Getting Started](getting-started.md) to create your first Lokstra app.  
- Or jump into [Examples](examples/README.md) to see runnable demos.  

---

## Repository

GitHub: [github.com/primadi/lokstra](https://github.com/primadi/lokstra)

---
