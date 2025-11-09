---
layout: docs
title: Router Guide - Lokstra as HTTP Router
description: Complete guide to using Lokstra as a fast, flexible HTTP router like Echo, Gin, or Chi
---

# Router Guide - Lokstra as HTTP Router

> **Use Lokstra like Echo, Gin, or Chi** - Fast, flexible HTTP routing without the complexity of dependency injection frameworks

Welcome to **Track 1** of Lokstra! Here you'll learn to use Lokstra as a **simple, powerful HTTP router** with:

- ✅ **Flexible Handler Signatures** - 29+ different handler forms
- ✅ **Clean Middleware Support** - Composable, reusable middleware
- ✅ **Group Routing** - API versioning and organization  
- ✅ **Type-Safe Parameters** - Automatic parameter binding and validation
- ✅ **Zero Framework Lock-in** - Use as much or as little as you need

---

## Router Comparison

| Feature | Lokstra Router | Echo | Gin | Chi |
|---------|---------------|------|-----|-----|
| **Handler Forms** | 29+ signatures | Fixed signatures | Fixed signatures | Fixed signatures |
| **Middleware** | ✅ Composable | ✅ Good | ✅ Good | ✅ Excellent |
| **Parameter Binding** | ✅ Auto-validation | Manual | Manual | Manual |
| **Type Safety** | ✅ Struct-based | Manual casting | Manual casting | Manual casting |
| **Learning Curve** | Low | Medium | Low | Low |

**[👉 Why choose Lokstra over other routers?](#why-lokstra-router)**

---

## Quick Router Example

**Simple API in 10 lines:**
```go
package main

import "github.com/primadi/lokstra"

func main() {
    r := lokstra.NewRouter("api")
    
    r.GET("/", func() string {
        return "Hello, Lokstra Router!"
    })
    
    r.GET("/users/{id}", func(id string) string {
        return fmt.Sprintf("User ID: %s", id)
    })
    
    app := lokstra.NewApp("my-api", ":8080", r)
    app.Run(30 * time.Second)
}
```

**What's special:**
- ✅ **No manual parameter extraction** - `id` automatically injected
- ✅ **Multiple return types** - string, JSON, error handling
- ✅ **Clean syntax** - No complex structs or handlers
- ✅ **Type safety** - Compile-time parameter validation

---

## Learning Path

### 🎯 1. [Router Basics](./01-router/)
Learn the foundation of Lokstra routing with flexible handler signatures.

**What you'll learn:**
- ✅ 29+ handler signature patterns
- ✅ Path parameters and query strings
- ✅ Request/Response handling patterns
- ✅ HTTP method routing (GET, POST, PUT, DELETE)

**Key concepts:**
```go
// Simple handlers
r.GET("/ping", func() string { return "pong" })

// With parameters
r.GET("/users/{id}", func(id int) (*User, error) {
    return getUserByID(id)
})

// With request context
r.POST("/users", func(ctx *request.Context, user *User) error {
    return ctx.Api.Created(createUser(user))
})
```

---

### 🛡️ 2. [Simple Services](./02-service/)
Add basic service layer without complex dependency injection.

**What you'll learn:**
- ✅ Simple service patterns
- ✅ Manual dependency management
- ✅ Service organization
- ✅ Testing strategies

**Key concepts:**
```go
type UserService struct {
    db *Database
}

func NewUserService(db *Database) *UserService {
    return &UserService{db: db}
}

func (s *UserService) GetByID(id int) (*User, error) {
    return s.db.QueryOne("SELECT * FROM users WHERE id = ?", id)
}

// Simple registration
userService := NewUserService(db)
r.GET("/users/{id}", userService.GetByID)
```

---

### 🔗 3. [Middleware](./03-middleware/)
Add cross-cutting concerns like logging, authentication, and CORS.

**What you'll learn:**
- ✅ Built-in middleware (CORS, logging, recovery)
- ✅ Custom middleware creation
- ✅ Middleware composition and ordering
- ✅ Request/Response transformation

**Key concepts:**
```go
// Built-in middleware
r.Use(cors.Middleware("*"))
r.Use(logging.Middleware())
r.Use(recovery.Middleware())

// Custom middleware
r.Use(func(next lokstra.Handler) lokstra.Handler {
    return func(ctx *request.Context) error {
        start := time.Now()
        err := next(ctx)
        log.Printf("Request took %v", time.Since(start))
        return err
    }
})

// Group middleware
api := r.Group("/api/v1")
api.Use(authMiddleware)
api.GET("/protected", protectedHandler)
```

---

### ⚙️ 4. [Configuration](./04-configuration/)
Simple configuration patterns without YAML complexity.

**What you'll learn:**
- ✅ Environment variables
- ✅ Configuration structs
- ✅ Development vs production setup
- ✅ Secret management

**Key concepts:**
```go
type Config struct {
    Port     string `env:"PORT" default:"8080"`
    Database string `env:"DATABASE_URL" required:"true"`
    Debug    bool   `env:"DEBUG" default:"false"`
}

func loadConfig() *Config {
    cfg := &Config{}
    env.Parse(cfg)
    return cfg
}
```

---

### 🚀 5. [App & Server](./05-app-and-server/)
Application lifecycle, graceful shutdown, and server management.

**What you'll learn:**
- ✅ Application setup and lifecycle
- ✅ Graceful shutdown handling
- ✅ Health checks and monitoring
- ✅ Production deployment patterns

---

### 🎯 6. [Putting it Together](./06-putting-it-together/)
Complete example combining all router concepts.

**What you'll learn:**
- ✅ Project structure best practices
- ✅ Testing strategies
- ✅ Performance optimization
- ✅ Production deployment

---

## Router Benefits

### vs Manual HTTP Handling
- ✅ **No manual mux setup** - Clean router abstraction
- ✅ **Automatic parameter binding** - No more `mux.Vars(r)`
- ✅ **Type-safe handlers** - Compile-time validation
- ✅ **Built-in middleware** - No need to write common functionality

### vs Other Go Routers
- ✅ **More flexible handlers** - 29+ signature patterns vs fixed signatures
- ✅ **Automatic validation** - Struct tags for parameter validation
- ✅ **Clean middleware** - Simple function composition
- ✅ **No framework lock-in** - Use incrementally

### vs Full Frameworks
- ✅ **Simpler learning curve** - Just routing, no DI complexity
- ✅ **Faster development** - Less boilerplate, more productivity
- ✅ **Smaller binary** - No unused framework features
- ✅ **Easy migration** - Can upgrade to full framework later

---

## Why Lokstra Router?

### 🎯 **Flexibility Without Complexity**
```go
// All of these work automatically:
r.GET("/simple", func() string { return "ok" })
r.GET("/with-error", func() (string, error) { return "ok", nil })
r.GET("/with-context", func(ctx *request.Context) error { 
    return ctx.Api.Ok("ok") 
})
r.GET("/with-params/{id}", func(id int) (*User, error) { 
    return getUser(id) 
})
r.POST("/with-body", func(user *CreateUserRequest) (*User, error) { 
    return createUser(user) 
})
```

### 🛡️ **Type Safety by Default**
```go
type GetUserParams struct {
    ID int `path:"id" validate:"required,min=1"`
}

// Automatic validation - no manual checking needed!
r.GET("/users/{id}", func(p *GetUserParams) (*User, error) {
    // p.ID is already validated as positive integer
    return userService.GetByID(p.ID)
})
```

### 🔗 **Composable Middleware**
```go
// Build middleware chains easily
authChain := lokstra.Chain(
    cors.Middleware("*"),
    logging.Middleware(),
    auth.Middleware("jwt-secret"),
)

api := r.Group("/api/v1")
api.Use(authChain)
```

### 📦 **Incremental Adoption**
```go
// Start simple
r.GET("/ping", func() string { return "pong" })

// Add services gradually
r.GET("/users/{id}", userService.GetByID)

// Add middleware when needed
r.Use(logging.Middleware())

// Upgrade to full framework later if needed
```

---

## When to Use Router Track

Choose **Router Track** when you're building:

- ✅ **Simple APIs** with straightforward routing needs
- ✅ **Microservices** that don't need complex DI
- ✅ **Rapid prototypes** that need fast development
- ✅ **Team projects** where not everyone knows DI patterns
- ✅ **Migration projects** from other Go routers
- ✅ **Learning projects** to understand routing fundamentals

If you need auto-generated APIs, dependency injection, or enterprise features, consider **[Framework Track](../02-framework-guide/)** instead.

---

## Router vs Framework Track

| Aspect | Router Track | Framework Track |
|--------|-------------|----------------|
| **Complexity** | Simple, straightforward | More concepts to learn |
| **Setup Time** | Minutes | Hours |
| **DI Pattern** | Manual, simple | Automatic, type-safe |
| **API Generation** | Manual routing | Auto-generated from services |
| **Configuration** | Environment variables | YAML + code |
| **Team Size** | 1-5 developers | 5+ developers |
| **Use Case** | Simple APIs, microservices | Enterprise applications |

**Not sure which track?** Start with **Router Track** - you can always upgrade to Framework Track later!

---

## Quick Links

- 🎯 **[Start Router Tutorial](./01-router/)** - Begin learning Lokstra routing
- 🏗️ **[Framework Track](../02-framework-guide/)** - Full enterprise features
- 💡 **[Examples](../00-introduction/examples/)** - Working code samples
- 📖 **[API Reference](../03-api-reference/)** - Complete technical docs

---

## Next Steps

### Continue Learning:
1. **[Router Basics](./01-router/)** - Core routing concepts
2. **[Simple Services](./02-service/)** - Service organization  
3. **[Middleware](./03-middleware/)** - Cross-cutting concerns
4. **[Configuration](./04-configuration/)** - Environment setup
5. **[Complete Example](./06-putting-it-together/)** - Production-ready app

### Build Real Projects:
- Build REST APIs with clean routing
- Add authentication and middleware
- Create microservices with simple patterns
- Deploy to production with confidence

**Ready to start routing?** → [Router Basics](./01-router/)

**Need enterprise features?** → [Framework Guide](../02-framework-guide/)

---

<div align="center">
  <p><strong>Simple routing. Clean code. Fast development.</strong></p>
  <p>Start with Router Track and grow into Framework Track when you need it! 🚀</p>
</div>
