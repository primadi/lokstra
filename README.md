# Lokstra ⚡

<p align="center">
	<img src="docs/asset/logo.png" alt="Logo" style="max-width: 100%; width: 300px;">
</p>

> Modern Go web framework with smart request binding, HTMX integration, and type-safe dependency injection.

📘 [Positioning Statement](./POSITIONING.md) — What Lokstra *is* and *is not*  
📈 [Milestone & Roadmap](./MILESTONE.md) — Development status and upcoming features  
📚 [Full Documentation](./docs/README.md) — Complete framework documentation

---

## ✨ Overview

**Lokstra** is a modern Go web framework designed for building scalable web applications with minimal boilerplate. It combines Go's simplicity with structured conventions, featuring smart request binding, first-class HTMX support, and a type-safe service container.

Whether you're building REST APIs, HTMX-powered web applications, or microservices, Lokstra provides the structure and tools you need without sacrificing Go's performance and simplicity.

---

## 🧭 Philosophy

> **Simple. Smart. Structured.**

Lokstra follows the Rails philosophy applied to Go: convention over configuration, but with escape hatches. It provides sensible defaults and clear patterns while remaining flexible enough to customize when needed. The framework emphasizes developer productivity through smart request binding, structured responses, and type-safe dependency injection.

---

## 🚀 Key Features

### 🎯 Smart Request Binding
```go
type UserRequest struct {
    ID    string `path:"id"`           // From URL path
    Token string `header:"Authorization"` // From HTTP headers  
    Name  string `body:"name"`         // From request body
    Page  int    `query:"page"`        // From query parameters
}

// Auto-bind smart pattern
func getUserHandler(ctx *lokstra.Context, req *UserRequest) error {
    // Request automatically bound - use data directly
    return ctx.Ok(req)
}
```

### 🏗️ Clean Architecture
- ✅ **Registration Context** - Type-safe dependency injection container
- ✅ **App/Server Separation** - Multiple apps on same/different ports
- ✅ **Modular Structure** - Clear separation of concerns
- ✅ **Service Container** - Factory-based service management with configuration

### 🌐 HTMX Integration
- ✅ **First-class HTMX support** - Built-in page serving with script injection
- ✅ **Static file serving** - Efficient serving with SPA support
- ✅ **HTMX helpers** - Response methods for dynamic content
- ✅ **Modern web apps** - Build interactive UIs without complex frontend builds

### 📝 Configuration System
- ✅ **YAML schema validation** - IDE support with auto-completion
- ✅ **Directory loading** - Merge multiple config files
- ✅ **Environment overrides** - Type-safe environment variable support
- ✅ **Declarative setup** - Define apps, routes, and services in YAML

### 🔧 Developer Experience
- ✅ **Minimal boilerplate** - Get started quickly with sensible defaults
- ✅ **Structured responses** - Consistent JSON API responses with method chaining
- ✅ **Type-safe services** - Compile-time checked service retrieval
- ✅ **Built-in middleware** - CORS, recovery, logging, compression, and more

---

## 🧱 Directory Structure

```
## 📂 Quick Start

### Simple REST API

```go
package main

import (
    "github.com/primadi/lokstra"
    "time"
)

func main() {
    regCtx := lokstra.NewGlobalRegistrationContext()
    app := lokstra.NewApp(regCtx, "api-app", ":8080")
    
    // Smart binding example
    app.POST("/users", func(ctx *lokstra.Context) error {
        var req struct {
            Name  string `body:"name"`
            Email string `body:"email"`
        }
        
        if err := ctx.BindBodySmart(&req); err != nil {
            return ctx.ErrorBadRequest(err.Error())
        }
        
        // Business logic here...
        user := createUser(req.Name, req.Email)
        
        return ctx.OkCreated(user).WithMessage("User created successfully")
    })
    
    app.StartWithGracefulShutdown(true, 30 * time.Second)
}
```

### HTMX Web Application

```go
func main() {
    regCtx := lokstra.NewGlobalRegistrationContext()
    app := lokstra.NewApp(regCtx, "web-app", ":8080")
    
    // Serve HTMX pages
    app.MountHtmx("/", htmxTemplates)
    
    // Dynamic content endpoint
    app.GET("/api/dashboard", func(ctx *lokstra.Context) error {
        return ctx.HtmxPageData("Dashboard", "Welcome!", map[string]any{
            "user":  getCurrentUser(),
            "stats": getDashboardStats(),
        })
    })
    
    app.Start()
}
```

### Multiple Applications

```go
func main() {
    regCtx := lokstra.NewGlobalRegistrationContext()
    server := lokstra.NewServer(regCtx, "multi-app-server")

    // API application
    apiApp := lokstra.NewApp(regCtx, "api", ":8080")
    apiApp.GET("/api/users", getUsersHandler)
    
    // Admin application  
    adminApp := lokstra.NewApp(regCtx, "admin", ":8081")
    adminApp.GET("/admin/dashboard", adminHandler)
    
    server.AddApp(apiApp).AddApp(adminApp)
    server.StartWithGracefulShutdown(true, 30 * time.Second)
}
```

---

## 🏗️ Architecture

```
lokstra/
├── common/         # Utilities: customtype, json, utils
├── core/           # Core framework components
│   ├── app/        # Application management
│   ├── config/     # YAML configuration system
│   ├── request/    # Smart request binding
│   ├── response/   # Structured response system
│   ├── router/     # Flexible routing engine
│   └── service/    # Type-safe service container
├── middleware/     # Built-in middleware
├── services/       # Built-in services (DB, Redis, Logger, etc.)
├── serviceapi/     # Service interface definitions
├── modules/        # Service modules
├── cmd/examples/   # Comprehensive examples
├── docs/           # Complete documentation
└── schema/         # YAML schema definitions
```

---

## 🧪 Example: Minimal App

```go
package main

import "github.com/primadi/lokstra"

func main() {
	regCtx := lokstra.NewGlobalRegistrationContext()

	srv := lokstra.NewServer(regCtx, "my-server")
	app := lokstra.NewApp(regCtx, "app1", ":8080")

	app.GET("/hello", func(ctx *lokstra.Context) error {
		return ctx.Ok("Hello From Lokstra")
	})

	srv.AddApp(app)
	_ = srv.Start()
}
```
```

---

## ⚙️ YAML Configuration with Schema Support

Lokstra provides comprehensive YAML configuration with full IntelliSense support for modern editors.

### JSON Schema for YAML Language Server

The `schema/lokstra.json` file provides complete validation and auto-completion for all configuration options:

```yaml
# yaml-language-server: $schema=./schema/lokstra.json

server:
  name: my-server
  global_setting:
    log_level: info

apps:
  - name: api-app
    address: ":8080"
    routes:
      - method: GET
        path: /health
        handler: health.check
    middleware:
      - name: cors
        enabled: true
      - name: logger
        config:
          level: debug

services:
  - name: main-db
    type: lokstra.dbpool.pg
    config:
      dsn: "postgres://user:pass@localhost/db"
```

### VS Code Setup

Add to your `.vscode/settings.json`:

```json
{
  "yaml.schemas": {
    "./schema/lokstra.json": [
      "**/configs/**/*.yaml",
      "lokstra.yaml",
      "server.yaml"
    ]
  }
}
```

---

## 🧩 Built-in Services & Middleware

### Services
- ✅ **Database Pool** - PostgreSQL connection pool with schema support
- ✅ **Redis** - Redis client with connection pooling  
- ✅ **Logger** - Structured logging with multiple output formats
- ✅ **Metrics** - Prometheus metrics integration
- ✅ **Health Check** - Application health monitoring
- ✅ **Key-Value Store** - In-memory and Redis-backed implementations

### Middleware
- ✅ **CORS** - Cross-origin request handling
- ✅ **Recovery** - Panic recovery with error logging
- ✅ **Request Logger** - HTTP request/response logging
- ✅ **Body Limit** - Request body size limiting
- ✅ **Gzip Compression** - Response compression
- ✅ **Slow Request Logger** - Performance monitoring

---

## 📚 Learning & Examples

### Documentation
- 📖 [Getting Started](./docs/getting-started.md) - Installation and first steps
- 🏗️ [Core Concepts](./docs/core-concepts.md) - Understanding Lokstra's architecture  
- 🛣️ [Routing](./docs/routing.md) - Advanced routing features
- 🛡️ [Middleware](./docs/middleware.md) - Custom middleware development
- ⚙️ [Services](./docs/services.md) - Service management and DI
- 📝 [Configuration](./docs/configuration.md) - YAML configuration guide
- 🌐 [HTMX Integration](./docs/htmx-integration.md) - Building modern web apps

### Progressive Examples

📂 See full details in [`cmd/examples/README.md`](cmd/examples/README.md)

1. **01_basic_overview** – From minimal router to YAML-configured server  
2. **02_router_features** – Groups, mounting, middleware examples  
3. **03_best_practices** – Custom context, naming, config patterns  
4. **04_customization** – Override JSON, responses, router engines  
5. **05_service_lifecycle** – Service registration and management
6. **06_business_modules** – Domain-driven service examples  
7. **07_default_services** – Logger, DBPool, Redis, Metrics, etc.  
8. **08_default_middleware** – Recovery, CORS, logging, compression

> 💡 Each example is self-contained and runnable with comprehensive documentation.

---

## 🎯 Production Ready

Lokstra is **production-ready** today with:

- ✅ **Smart Request Binding** - Comprehensive struct tag binding with auto-detection
- ✅ **HTMX Integration** - First-class support for modern web applications
- ✅ **Type-Safe Services** - Dependency injection with compile-time safety
- ✅ **Configuration Schema** - YAML validation with IDE support
- ✅ **Structured Responses** - Consistent API responses with method chaining
- ✅ **Built-in Observability** - Metrics, logging, and health checks
- ✅ **Static File Serving** - Efficient serving with SPA support
- ✅ **Graceful Shutdown** - Production-ready lifecycle management

## 🔮 Roadmap

### Near Term (Next Release)
- [ ] **WebSocket Support** - Real-time communication
- [ ] **File Upload Handling** - Multipart form processing  
- [ ] **Enhanced Validation** - Custom validation rules
- [ ] **CLI Tool** - Project scaffolding and management

### Medium Term
- [ ] **Advanced Security** - JWT, rate limiting, CSRF protection
- [ ] **Background Jobs** - Task queuing and processing
- [ ] **Distributed Tracing** - OpenTelemetry integration
- [ ] **Advanced HTMX** - Enhanced patterns and helpers

### Long Term  
- [ ] **lokstra-call** - Internal RPC framework
- [ ] **Multi-Database** - MySQL, SQLite, and more
- [ ] **Plugin Architecture** - Extensible module system
- [ ] **Admin Dashboard** - Management interface

---

## 📜 License

Lokstra is licensed under the [Apache License 2.0](LICENSE).

---

## 🙌 Contributing

Lokstra is actively developed and welcomes contributions! Whether you're interested in:

- 🐛 **Bug Reports** - Help us improve by reporting issues
- 💡 **Feature Requests** - Suggest new capabilities  
- 📝 **Documentation** - Improve guides and examples
- 🔧 **Code Contributions** - Submit patches and enhancements
- 🧪 **Testing** - Help validate new features

Please open an issue or submit a pull request on GitHub. Check out our [contributing guidelines](CONTRIBUTING.md) for more details.

For roadmap discussions and major feature planning, we encourage community input through GitHub discussions.
