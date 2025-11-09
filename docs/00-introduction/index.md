---
layout: docs
title: Introduction
---

# What is Lokstra?

> **A versatile Go web framework that works in two ways: as a Router or as a complete Business Application Framework**

---

## 🎯 Two Ways to Use Lokstra

Lokstra is designed to meet you where you are:

### 🎯 Track 1: As a Router (Like Echo, Gin, Chi)
**Quick, simple HTTP routing without framework complexity**

Use Lokstra as a pure HTTP router for:
- Quick prototypes and MVPs
- Simple REST APIs
- Learning HTTP routing fundamentals
- Projects without DI needs

```go
r := lokstra.NewRouter("api")
r.GET("/users", getUsersHandler)
r.Use(cors.Middleware("*"))

app := lokstra.NewApp("api", ":8080", r)
app.Run(30 * time.Second)
```

**Compare with:** Echo, Gin, Chi, Fiber  
**Learn more:** [Router Guide](../01-router-guide/)

---

### �️ Track 2: As a Business Application Framework (Like NestJS, Spring Boot)
**Enterprise features with DI, auto-router, and configuration-driven deployment**

Use Lokstra as a complete framework for:
- Enterprise applications
- Microservices architectures
- Dependency injection and service layer
- Configuration-driven deployment

```yaml
# Define services in YAML
service-definitions:
  user-service:
    type: user-service-factory
    depends-on: [database]
```

```go
// Type-safe lazy loading
var userService = service.LazyLoad[*UserService]("user-service")

func handler() {
    users := userService.MustGet().GetAll()
}

// Auto-router generates REST endpoints from service methods!
```

**Compare with:** NestJS, Spring Boot, Uber Fx, Buffalo  
**Learn more:** [Framework Guide](../02-framework-guide/)

---

## 🚀 The Big Idea

**Start simple, scale when needed:**
1. Start with routing (Track 1) for quick results
2. Add services and DI when complexity grows (Track 2)
3. Enable auto-router to reduce boilerplate
4. Switch deployment topology without code changes

**Lokstra grows with your needs!**

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

// Register service type (factory creates instances)
lokstra_registry.RegisterServiceType("user-service", func() any {
    return &UserService{}
}, nil)
```

**Key Feature**: Service methods can become HTTP endpoints automatically!

---

### 3. **Middleware** - Request Processing
```go
r.Use(recovery.Middleware(nil), cors.Middleware("*"))
```

**Key Feature**: Apply middleware globally or per-route

---

### 4. **Configuration** - YAML or Code
```yaml
deployments:
  production:
    servers:
      web-server:
        base-url: http://localhost
        addr: ":8080"
        published-services: [user-service]
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
    app.PrintStartInfo()
    if err := app.Run(30 * time.Second); err != nil {
      log.Fatal(err)
    }
}
```

**Test it:**
```bash
curl http://localhost:8080/ping    # → "pong"
curl http://localhost:8080/users   # → ["Alice","Bob","Charlie"]
```

**💭 Note**: This is a simplified introduction example. For complete runnable examples with proper project structure, see:
- [Quick Start](quick-start) - Your first working API in 5 minutes
- [Examples](examples) - 4 progressive examples from basics to production

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
    result, err := db.Save(user)
    if err != nil {
      return ctx.Api.InternalError(err.Error())
    }
    return ctx.Api.Ok(result)
})

// With response.ApiHelper
r.GET("/complex", func(user *User) (*response.ApiHelper, error) {
    result, err := db.Save(user)
    if err != nil {
      return nil, err
    }
    return response.NewApiOk(result), nil
})
```

**29+ different handler forms supported!**

---

### 2. **Service as Router** (Game Changer!)
Your service methods automatically become HTTP endpoints:

```go
// Step 1: Define service with methods
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

func NewUserService() *UserService {
    return &UserService{}
}

// Step 2: Register service factory with metadata
lokstra_registry.RegisterServiceType(
    "user-service-factory",
    NewUserService,
    nil,
    deploy.WithResource("user", "users"),
    deploy.WithConvention("rest"),
)

// Step 3: Auto-generate router from service!
userRouter := lokstra_registry.NewRouterFromServiceType("user-service-factory")

// Automatically creates:
// GET  /users       → GetAll() method
// GET  /users/{id}  → GetByID() method  
// POST /users       → Create() method
```

**Zero boilerplate routing code!**

---

### 3. **One Binary, Multiple Deployments**
Configure once, deploy anywhere:

```yaml
deployments:
  monolith:
    servers:
      web-server:
        addr: ":8080"
        published-services: [users, orders, payments]
  
  multi-server:
    servers:
      user-service:
        base-url: http://user-service
        addr: ":8001"
        published-services: [users]
      
      order-payment-service:
        base-url: http://order-payment-service
        addr: ":8002"
        published-services: [orders, payments]
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
lokstra_registry.RegisterServiceType("db", createDB, nil)

lokstra_registry.RegisterServiceFactory("users-factory", 
    func(deps map[string]any, config map[string]any) any {
        return &UserService{
            DB: service.Cast[*Database](deps["db"]),
        }
    })

lokstra_registry.RegisterLazyService("users", "users-factory", 
    map[string]any{"depends-on": []string{"db"}})

// Use anywhere
users := lokstra_registry.GetService[*UserService]("users")

// Inside service method - DB injected, accessed lazily
func (u *UserService) GetUsers() ([]User, error) {
    db := u.DB.MustGet()  // Injected dependency, accessed on first call
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

Choose your learning path:

### Track 1: Router Only (2-3 hours)
✅ Create routers and register routes  
✅ Write handlers in 29+ different styles  
✅ Apply middleware (global, per-route, groups)  
✅ Manage app and server lifecycle  
✅ Build REST APIs without DI

**[→ Router Guide](../01-router-guide/)**

### Track 2: Full Framework (6-8 hours)
✅ Everything in Track 1, plus:  
✅ Service layer and dependency injection  
✅ Auto-generate routes from services  
✅ Configuration-driven deployment (YAML or Code)  
✅ Monolith → Microservices migration  
✅ External service integration  

**[→ Framework Guide](../02-framework-guide/)**

---

## 🚦 Where to Go Next?

### New to Lokstra? Start Here:

#### Option 1: Quick Start (Fastest)
**Just want to see it work?**
- 👉 [Quick Start](quick-start) - Build your first API in 5 minutes
- 👉 [Examples - Router Track](examples/router-only/) - 3 quick examples (2-3 hours)

#### Option 2: Understand First (Recommended)
**Want to understand before coding?**
- 👉 [Why Lokstra?](why-lokstra) - Compare with other frameworks
- 👉 [Architecture](architecture) - Deep dive into design
- 👉 [Key Features](key-features) - What makes Lokstra different

### Ready to Learn? Choose Your Track:

#### Track 1: Router Only
**For: Quick APIs, prototypes, simple projects**
1. [Examples - Router Track](examples/router-only/) - Hands-on learning (2-3 hours)
2. [Router Guide](../01-router-guide/) - Deep dive into routing
3. [API Reference - Router](../03-api-reference/01-core-packages/router) - Complete API

#### Track 2: Full Framework  
**For: Enterprise apps, microservices, DI architecture**
1. [Examples - Framework Track](examples/full-framework/) - Hands-on learning (8-12 hours)
2. [Framework Guide](../02-framework-guide/) - Services, DI, auto-router
3. [Configuration Reference](../03-api-reference/03-configuration/) - YAML schema

### Not Sure Which Track?

**Start with Track 1 (Router)** if:
- New to Lokstra
- Building MVP or prototype
- Want quick results
- Don't need DI yet

**Start with Track 2 (Framework)** if:
- Need dependency injection
- Building microservices
- Want auto-generated routers
- Familiar with NestJS/Spring Boot

**Remember:** You can always upgrade from Track 1 to Track 2 later!

---

## 🎯 Key Takeaways

Before moving on, remember:

1. **Lokstra is flexible** - Use as router-only OR full framework
2. **Start simple, scale up** - Begin with Track 1, upgrade to Track 2 when needed
3. **Convention over configuration** - Smart defaults, configure when needed
4. **Service-oriented** - Services are first-class citizens (Track 2)
5. **Deployment-agnostic** - Same code, monolith or microservices (Track 2)
6. **Production-ready** - Built for real applications

---

## 📚 Learning with Examples

We provide **two example tracks** based on how you want to use Lokstra:

### Track 1: Router-Only Examples
⏱️ **2-3 hours total** • Perfect for quick APIs

- [01: Hello World](examples/router-only/01-hello-world/) - Basic routing (15 min)
- [02: Handler Forms](examples/router-only/02-handler-forms/) - 29+ handler signatures (30 min)
- [03: Middleware](examples/router-only/03-middleware/) - Global, per-route, groups (45 min)

**What you'll build:** REST APIs with routing and middleware (no DI)

👉 **[Start Router Examples](examples/router-only/)**

---

### Track 2: Full Framework Examples
⏱️ **8-12 hours total** • For enterprise apps

- [01: CRUD with Services](examples/full-framework/01-crud-api/) - Service layer, DI (1 hour)
- [02: Multi-Deployment YAML](examples/full-framework/02-multi-deployment-yaml/) - Auto-router, microservices (2-3 hours)
- [03: Multi-Deployment Code](examples/full-framework/03-multi-deployment-pure-code/) - Type-safe config (30 min)
- [04: External Services](examples/full-framework/04-external-services/) - proxy.Service patterns (1-2 hours)
- [05: Remote Router](examples/full-framework/05-remote-router/) - Quick HTTP integration (30 min)

**What you'll build:** Enterprise apps with DI, auto-router, and microservices

👉 **[Start Framework Examples](examples/full-framework/)**

---

### 🤔 Which Track to Choose?

| | Router Track | Framework Track |
|---|---|---|
| **Time** | 2-3 hours | 8-12 hours |
| **Use Case** | Quick APIs, prototypes | Enterprise, microservices |
| **Features** | Routing, middleware | + DI, auto-router, config |
| **Compare With** | Echo, Gin, Chi | NestJS, Spring Boot |

**Not sure?** Start with Router Track (it's compatible with Framework features!)

---

## �🚀 Roadmap

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

**Ready?** → Continue to [Why Lokstra?](why-lokstra) or jump straight to [Quick Start](quick-start)
