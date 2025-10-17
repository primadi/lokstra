# What is Lokstra?

> **A Go framework for building REST APIs with convention over configuration, powerful dependency injection, and flexible deployment options.**

---

## 🎯 The Big Idea

Lokstra helps you build Go REST APIs that are:
- **Easy to start** - Simple, intuitive API
- **Easy to scale** - From monolith to microservices without code changes
- **Easy to maintain** - Clean separation of concerns

---

## 📖 Core Concepts (5-Minute Overview)

Lokstra has **6 core components** that work together:

### 1. **Router** - HTTP Routing
```go
r := lokstra.NewRouter("api")
r.GET("/users", func() ([]User, error) {
    return db.GetAllUsers()
})
```

**Key Feature**: Flexible handler signatures - many forms supported!

---

### 2. **Service** - Business Logic
```go
type UserService struct {}

func (s *UserService) GetAll() ([]User, error) {
    return db.Query("SELECT * FROM users")
}

// Register service
lokstra_registry.RegisterServiceFactory("users", func() any {
    return &UserService{}
})
```

**Key Feature**: Service methods can become HTTP endpoints automatically!

---

### 3. **Middleware** - Request Processing
```go
r.Use(logging.Middleware(), auth.Middleware())
```

**Key Feature**: Apply middleware globally or per-route

---

### 4. **Configuration** - YAML or Code
```yaml
servers:
  - name: web-server
    apps:
      - addr: ":8080"
        routers: [api]
```

**Key Feature**: One config file for multiple deployments

---

### 5. **App** - Application Container
```go
app := lokstra.NewApp("my-app", ":8080", router)
```

**Key Feature**: Combine multiple routers into one app

---

### 6. **Server** - Server Manager
```go
server := lokstra.NewServer("main", app)
server.Run(30 * time.Second) // Graceful shutdown
```

**Key Feature**: Manage multiple apps, automatic graceful shutdown

---

## 🚀 Your First Lokstra App

**Complete working example in 20 lines:**

```go
package main

import (
    "github.com/primadi/lokstra"
    "time"
)

func main() {
    // 1. Create router
    r := lokstra.NewRouter("api")
    
    // 2. Register routes
    r.GET("/ping", func() string {
        return "pong"
    })
    
    r.GET("/users", func() []string {
        return []string{"Alice", "Bob", "Charlie"}
    })
    
    // 3. Create app & run
    app := lokstra.NewApp("demo", ":8080", r)
    app.Run(30 * time.Second)
}
```

**Test it:**
```bash
curl http://localhost:8080/ping    # → "pong"
curl http://localhost:8080/users   # → ["Alice","Bob","Charlie"]
```

**💭 Note**: This is a simplified introduction example. For complete runnable examples, see the [Quick Start](quick-start.md) guide and [Examples](../01-essentials/06-putting-it-together/README.md) section.

---

## 💡 What Makes Lokstra Different?

### 1. **Flexible Handler Signatures**
You're not locked into one pattern. Write handlers that make sense:

```go
// Simple
r.GET("/ping", func() string { return "pong" })

// With error
r.GET("/users", func() ([]User, error) { 
    return db.GetUsers() 
})

// With context
r.POST("/users", func(ctx *request.Context, user *User) error {
    return db.Save(user)
})

// With full control
r.GET("/complex", func(ctx *request.Context) (*response.Response, error) {
    return response.Success(data), nil
})
```

**29 different handler forms supported!**

---

### 2. **Service as Router** (Game Changer!)
Your service methods automatically become HTTP endpoints:

```go
type UserService struct {}

type GetAllParams struct {}
type GetByIDParams struct {
    ID string `path:"id"`
}
type CreateUserParams struct {
    User *User `json:"user"`
}

func (s *UserService) GetAll(p *GetAllParams) ([]User, error) { ... }
func (s *UserService) GetByID(p *GetByIDParams) (*User, error) { ... }
func (s *UserService) Create(p *CreateUserParams) error { ... }

// Automatically creates:
// GET  /users
// GET  /users/{id}
// POST /users
```

**Zero boilerplate routing code!**

---

### 3. **One Binary, Multiple Deployments**
Configure once, deploy anywhere:

```yaml
servers:
  - name: monolith
    deployment-id: monolith
    apps:
      - addr: ":8080"
        services: [users, orders, payments]
  
  - name: user-service
    deployment-id: multi-server
    base-url: http://user-service
    apps:
      - addr: ":8001"
        services: [users]
  
  - name: order-service
    deployment-id: multi-server
    base-url: http://order-payment-service
    apps:
      - addr: ":8002"
        services: [orders, payments]
```

**Same code, different architectures** - just change config!

```bash
# Run as monolith
./app --server=monolith

# Run as microservice
./app --server=user-service
```

---

### 4. **Built-in Dependency Injection**
No external DI framework needed:

```go
import "github.com/primadi/lokstra/core/service"

type UserService struct {
    DB *service.Cached[*Database]
}

// Register factories
lokstra_registry.RegisterServiceFactory("db", createDB)
lokstra_registry.RegisterServiceFactory("users", func() any {
    return &UserService{
        DB: service.LazyLoad[*Database]("db"),
    }
})

// Use anywhere
users := lokstra_registry.GetService[*UserService]("users")

// Inside service method - lazy load DB
func (u *UserService) GetUsers() ([]User, error) {
    db := u.DB.Get()  // Lazy loaded on first access
    return db.Query("SELECT * FROM users")
}
```

**Key Feature**: Lazy loading - services created only when needed!

---

## 🏗️ Architecture at a Glance

```
┌─────────────────────────────────────────┐
│           HTTP Request                  │
└─────────────┬───────────────────────────┘
              │
┌─────────────▼───────────────────────────┐
│          APP (with Listener)            │
│   (ServeMux / FastHTTP / HTTP2 / etc)   │
│  ┌────────────────────────────────────┐ │
│  │       ROUTER (http.Handler)        │ │
│  │  ┌──────────────────────────────┐  │ │
│  │  │       ROUTE                  │  │ │
│  │  │  ┌───────────────────────┐   │  │ │
│  │  │  │ MidWare 1 (before) ───┼───┼──┼─┼──┐
│  │  │  └────────┬──────────────┘   │  │ │  │
│  │  │  ┌────────▼──────────────┐   │  │ │  │
│  │  │  │ MidWare 2 (before) ───┼───┼──┼─┼──┤
│  │  │  └────────┬──────────────┘   │  │ │  │
│  │  │  ┌────────▼──────────────┐   │  │ │  │
│  │  │  │    HANDLER        ────┼───┼──┼─┼──┤
│  │  │  └────────┬──────────────┘   │  │ │  │
│  │  │  ┌───────────────────────┐   │  │ │  │
│  │  │  │ MidWare 2 (after)  ───┼───┼──┼─┼──┤
│  │  │  └────────┬──────────────┘   │  │ │  │
│  │  │  ┌────────▼──────────────┐   │  │ │  │
│  │  │  │ MidWare 1 (after)  ───┼───┼──┼─┼──┤
│  │  │  └────────┬──────────────┘   │  │ │  │
│  │  └───────────┼──────────────────┘  │ │  │
│  └──────────────┼─────────────────────┘ │  │
└─────────────────┼───────────────────────┘  │
                  │              ┌───────────┘
                  │     ┌────────▼───────────┐
                  │     │    SERVICES        │◄───┐
                  │     │  (business logic)  │    │
                  │     │                    │    │
                  │     │  Service can call: |    │
                  │     │  - Other Services ─┼────┘
                  │     │  - Database        |    
                  │     │  - External APIs   |    
                  │     └────────────────────┘
                  │
         ┌────────▼─────────┐
         │ RESPONSE OBJECT  │
         │  (formatting)    │
         └──────────────────┘
```

**Request Flow:**
1. HTTP request arrives at App's listener
2. Listener routes to Router (http.Handler)
3. Router finds matching Route
4. Route executes Middleware chain (via ctx.Next())
   - Middleware can call Services (e.g., auth checks)
5. Handler executes
   - Handler can call Services
6. Services contain business logic
   - Services can call other Services
   - Services can call databases, external APIs, etc.
7. Response Object formats the response
8. Response flows back to client

**Key Points:**
- **Services are accessible everywhere**: Middleware, Handlers, and other Services can all call Services
- **Server** is just a container for Apps, Services, and configuration - not part of request flow
- **Middleware** can use Services for auth, logging, etc.
- **Services** can depend on other Services (via Dependency Injection)

---

## 🎓 What You'll Learn

After working through this documentation:

### Essentials (2-3 hours)
✅ Create routers and register routes  
✅ Write handlers in multiple styles  
✅ Organize code with services  
✅ Apply middleware  
✅ Configure via YAML  
✅ Build complete REST APIs  

### Deep Dive (4-6 hours)
✅ Master all 29 handler forms  
✅ Auto-generate routes from services  
✅ Create custom middleware  
✅ Multi-deployment strategies  
✅ Remote service communication  
✅ Performance optimization  

---

## 🚦 Where to Go Next?

### I want to understand the "why"
👉 [Why Lokstra?](why-lokstra.md) - Compare with other frameworks

### I want to see the big picture
👉 [Architecture](architecture.md) - Deep dive into design

### I want to code NOW
👉 [Quick Start](quick-start.md) - Build your first API in 5 minutes

### I want to learn systematically
👉 [Essentials](../01-essentials/README.md) - Step-by-step tutorial

---

## 🎯 Key Takeaways

Before moving on, remember:

1. **Lokstra is flexible** - Multiple ways to do things, pick what works
2. **Convention over configuration** - Smart defaults, configure when needed
3. **Service-oriented** - Services are first-class citizens
4. **Deployment-agnostic** - Same code, monolith or microservices
5. **Production-ready** - Built for real applications

---

## 🚀 Roadmap

### Next Release
We're actively working on:

- **🎨 HTMX Support** - Build modern web apps easier
  - Template rendering integration
  - Partial page updates
  - Form handling patterns
  - Example applications

- **🛠️ CLI Tools** - Developer productivity
  - Project scaffolding (`lokstra new`)
  - Code generation (`lokstra generate`)
  - Migration helpers
  - Development server

- **📦 Complete Standard Library** - Production essentials
  - **Middleware**: Metrics, Auth (JWT, OAuth), Rate limiting
  - **Services**: Monitoring, Tracing, Health checks
  - **Utilities**: Common patterns and helpers

Want to contribute or suggest features? Visit our [GitHub repository](https://github.com/primadi/lokstra)

---

**Ready?** → Continue to [Why Lokstra?](why-lokstra.md) or jump straight to [Quick Start](quick-start.md)
