
# A1. Lokstra Core Overview (v4)

> **Simple. Scalable. Structured.**  
>
> Lokstra is a Go-based application framework designed to scale from a single-file monolith into a full microservice system — all without changing your core code.  
> Its architecture is built around five composable core components that define every part of the system lifecycle — from request handling to deployment orchestration.

---

## 📂 Lokstra Folder Structure

```
lokstra/
├── api_client/         # Remote call system for distributed services (client-side)
├── cmd/                # Runnable learning examples (no CLI yet)
├── common/             # Foundation utilities: cast, customtype, json, response_writer, utils, validator
├── core/               # Runtime engine: app, config, request, response, route, router, server, service
├── docs/               # Official documentation & learning guides
├── lokstra_handler/    # Mount helpers: reverse_proxies, spa, static (next: htmx)
├── lokstra_registry/   # Unified registry: servicefactory, service, middleware, router, app, server, client_router
├── middleware/         # Built-in middleware modules (CORS, recovery, logger, auth, etc.)
├── serviceapi/         # Shared interface contracts for inter-service communication
├── services/           # Default service implementations (dbpool, logger, metrics, session, etc.)
└── lokstra.go          # Root helper exporting common types & shortcuts
```

### Design Principles
- **Discoverable:** each folder has one clear responsibility.  
- **Extensible:** developers can register new services, middleware, or routers without touching the core.  
- **Unified:** all runtime components are discoverable via `lokstra_registry`.  

---

## 🧩 The 5 Core Components

| # | Component | Description | Analogy |
|---|------------|--------------|----------|
| 1️⃣ | **Router** | Defines endpoints and routes requests to handlers. | **Entrance door** — directs requests where they belong. |
| 2️⃣ | **Middleware** | Executes logic before, after, or around a handler. | **Security scanner** — can intercept, inspect, or modify. |
| 3️⃣ | **Service** | Reusable singleton providing business logic or infrastructure (DB, logger, cache). | **Engine room** — provides power and data. |
| 4️⃣ | **App** | Hosts routers and middleware on one listener (port/socket). | **Logical application** — e.g., an API or dashboard. |
| 5️⃣ | **Server** | Manages multiple Apps and shared Services. | **Control center** — the home of Apps, Routers, and Registry. |

---

## ⚙️ Build & Boot Phase

Lokstra builds components lazily but deterministically.  
Routers are built automatically on the first request, and services are initialized when needed.

```
Server.Start()
   ↓
App.Start()
   ↓
Router.Build() → constructs routes and middleware chains
   ↓
Listener active and ready to serve requests
```

Both **App** and **Server** have `Start()` (blocking) and `Run(timeout)` methods.  
`Run()` also waits for system signals (`SIGINT`, `SIGTERM`) and performs graceful shutdown automatically.

---

## 🚀 Request Execution Phase

When a request arrives, Lokstra executes a **chain of middleware and route handler**.

```
Incoming Request
   ↓
App Listener
   ↓
Router Lookup
   ↓
┌───────────────────────────────────────────┐
│ [Middleware #1: before]                   │
│   ↓                                       │
│ [Middleware #2: before]                   │
│   ↓                                       │
│ [Route Handler]                           │
│   ↓                                       │
│ [Middleware #2: after]                    │
│   ↓                                       │
│ [Middleware #1: after]                    │
└───────────────────────────────────────────┘
   ↓
Response Writer
```

### Notes
- Middleware can execute **before**, **after**, or **around** `Next()`.  
- Services can be called **anywhere**: middleware, handler, or another service.  
- Registry is used only for lookup, not execution.  
- Server only supervises Apps; it does not participate in per-request logic.

---

## 🧱 Minimal Example

```go
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/primadi/lokstra"
)

func main() {
	r := lokstra.NewRouter("basic-router")

	r.Use(func(c *lokstra.RequestContext) error {
		start := time.Now()
		err := c.Next()
		fmt.Println("Request", c.R.URL.Path, "took:", time.Since(start))
		return err
	})

	r.GET("/ping", func() (string, error) {
		return "pong", nil
	})

	app := lokstra.NewApp(":8080")
	app.AddRouter(r)
	app.Run(0)
}
```

📘 **See also:** [`cmd/examples/01_basic_router`](../../cmd/examples/01_basic_router)

---

## 🧭 App and Server Lifecycle

Both App and Server provide two main execution methods:

| Method | Behavior |
|--------|-----------|
| **Start()** | Starts the listener(s) and blocks until stopped. |
| **Run(timeout)** | Starts, waits for shutdown signals, and performs graceful shutdown within the given timeout. |

#### Example

```go
app := lokstra.NewApp(":8080")
app.AddRouter(r)

// Blocking start
app.Start()

// Or blocking with graceful shutdown
app.Run(5 * time.Second)
```

Multiple apps can be managed by a single server:

```go
server := lokstra.NewServer("main")
server.AddApp(app)
server.Run(10 * time.Second)
```

> Lokstra automatically handles signal listening and graceful shutdown.  
> Developers only need to choose whether to call `Start()` or `Run(timeout)`.

---

## 🧩 Conceptual Separation

| Layer | When it runs | What it does |
|-------|---------------|---------------|
| **Server** | Startup & shutdown | Runs apps concurrently and manages lifecycle |
| **App** | Startup & runtime | Owns routers, middleware, and listeners |
| **Router** | Build & runtime | Constructs route tree and executes handler chains |
| **Middleware** | Per request | Filters, wraps, or extends handler execution |
| **Service** | Anytime | Provides logic, data, or integration; may be lazy |
| **Registry** | Initialization & lookup | Stores and resolves global definitions |

---

## 🧠 Philosophy in Practice

| Principle | Applied As |
|------------|-------------|
| **Simple** | Define routes and responses in a few lines of Go. |
| **Scalable** | Run as one binary or distributed microservices using YAML config. |
| **Structured** | Each component (Router, App, Server, Service) has a clear lifecycle. |

> Lokstra separates **build-time orchestration** (Server, App, Registry)  
> from **runtime execution** (Router, Middleware, Handler Chain).  
> This ensures instant startup, lazy initialization, and predictable scaling.

---

## 🔗 Related Learning Examples

| Example | Description |
|----------|-------------|
| [`01_basic_router`](../../cmd/examples/01_basic_router) | Minimal router and middleware usage |
| [`02_app_server`](../../cmd/examples/02_app_server) | Running multiple apps on one server |
| [`03_service_usage`](../../cmd/examples/03_service_usage) | Registering and accessing services |
| [`04_response_hooks`](../../cmd/examples/04_response_hooks) | Handling API responses and hooks |
| [`05_yaml_deployment`](../../cmd/examples/05_yaml_deployment) | YAML-based configuration and orchestration |
