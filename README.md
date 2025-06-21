# Lokstra ⚡

> **Simple. Scalable. Structured.**  
> Lightweight Go backend framework for monoliths and microservices.

---

## ✨ Overview

**Lokstra** is a modular backend framework written in Go, designed for building scalable APIs and backend services with minimal boilerplate. Lokstra supports monolithic and microservice architectures out of the box, with a focus on fast development, clean structure, and runtime flexibility.

Whether you're building a SaaS platform, internal tools, or event-driven systems, Lokstra adapts to your structure — not the other way around.

---

## 🚀 Features

- ✅ Simple `Server → App → Router` structure with clean lifecycle
- ✅ Supports **multi-binary** and **multi-config** deployment
- ✅ Lightweight & fast routing (uses `httprouter`)
- ✅ **Built-in services**: Logger, DB pool, Redis, Metrics, JWT Auth, etc.
- ✅ Battery-included middleware: recovery, CORS, request logging, etc.
- ✅ Middleware at global, group, and handler levels
- ✅ **Service registry and lifecycle hooks**
- ✅ Extensible: add your own service or middleware easily
- ✅ **Multi-tenant ready**
- ✅ Configurable via **YAML** or **pure code**
- ✅ Graceful shutdown built-in

---

## 🧱 Directory Structure

```
lokstra/
├── core/           # Core: server, app, router, context
├── middleware/     # Built-in middleware
├── services/       # Built-in services
├── loader/         # YAML/config loader
├── internal/       # Internal helpers (non-exported)
├── cmd/examples/   # Example apps using Lokstra
├── docs/           # Documentation & tutorials
├── go.mod
├── LICENSE
└── README.md
```

---

## 📂 Example: Minimal App

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/primadi/lokstra/core"
)

func main() {
	srv := core.NewServer("lokstra-dev")

	app := &core.App{
		Name: "hello-app",
		Port: 8080,
		Router: func() *core.Router {
			r := core.NewRouter()
			r.Handle("/hello", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "Hello from Lokstra!")
			}))
			return r
		}(),
	}

	srv.AddApp(app)
	_ = srv.Start()
}
```

---

## 🧩 Planned Services

Lokstra includes pluggable services with minimal setup:

- [x] Logger (zero-dependency `zerolog`)
- [x] Redis connection pool
- [x] PostgreSQL connection pool (via `pgx`)
- [x] Prometheus metrics (built-in + custom)
- [x] JWT Authenticator
- [ ] Email sender
- [ ] WebSocket pub/sub engine
- [ ] Background worker engine
- [ ] RBAC + Permission manager

---

## 🧪 Examples

Explore runnable examples in:
```
cmd/examples/
├── simple/
├── multiapp/
└── yaml-config/
```

---

## 🧭 Roadmap

- [ ] Full middleware stack with YAML loader
- [ ] Service lifecycle & dependency injection
- [ ] Web UI helper (React + Mantine)
- [ ] Plugin architecture for modules
- [ ] CLI code generators
- [ ] Multi-tenant dashboard
- [ ] RBAC UI + User management

---

## 📜 License

Lokstra is licensed under the [Apache License 2.0](LICENSE).

---

## 🙌 Contributing

Lokstra is currently in active development and will be opened to contributors soon.  
Stay tuned for public release announcements and contribution guides.
