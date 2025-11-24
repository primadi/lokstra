---
layout: docs
title: Framework Guide - Lokstra as Business Application Framework
description: Complete guide to using Lokstra as a full enterprise framework with dependency injection, service architecture, and configuration-driven deployment
---

# Framework Guide - Lokstra as Business Application Framework

> **Use Lokstra like NestJS or Spring Boot** - Full enterprise framework with DI, services, and zero-code deployment topology changes

💡 **New to framework comparison?** See detailed comparisons:
- **[Lokstra vs NestJS](./lokstra-vs-nestjs)** - TypeScript framework comparison
- **[Lokstra vs Spring Boot](./lokstra-vs-spring-boot)** - Java framework comparison

Welcome to **Track 2** of Lokstra! Here you'll learn to use Lokstra as a **complete business application framework** with:

- ✅ **Lazy Dependency Injection** - Type-safe with Go generics
- ✅ **Service Architecture** - Clean separation of concerns
- ✅ **Auto-Generated Routers** - REST APIs from service definitions
- ✅ **Configuration-Driven Deployment** - Monolith ↔ Microservices without code changes
- ✅ **Environment Management** - Dev, staging, production configs

---

## Framework Comparison

| Feature | Lokstra Framework | NestJS | Spring Boot |
|---------|------------------|--------|-------------|
| **Language** | Go | TypeScript | Java |
| **DI Pattern** | Lazy + Generics | Decorators + Reflection | Annotations + Reflection |
| **Performance** | Compiled, fast startup | Runtime, slower startup | JVM, medium startup |
| **Auto Router** | ✅ From service methods | ✅ From decorators | ✅ From annotations |
| **Deployment Flexibility** | ✅ Zero-code topology change | ❌ Requires code changes | ❌ Requires code changes |
| **Type Safety** | ✅ Compile-time | ✅ TypeScript | ✅ Java |

**Detailed Comparisons:**
- **[👉 Lokstra vs NestJS](./lokstra-vs-nestjs)** - TypeScript framework analysis
- **[👉 Lokstra vs Spring Boot](./lokstra-vs-spring-boot)** - Java framework analysis

---

## Quick Framework Example

**1. Define Service** (`user_service.go`):
```go
type UserService struct {
    db *Database
}

func (s *UserService) GetAll(p *GetAllParams) ([]User, error) {
    return s.db.Query("SELECT * FROM users")
}

func (s *UserService) GetByID(p *GetByIDParams) (*User, error) {
    return s.db.QueryOne("SELECT * FROM users WHERE id = ?", p.ID)
}

func (s *UserService) Create(p *CreateParams) (*User, error) {
    user := &User{Name: p.Name, Email: p.Email}
    return s.db.Insert(user)
}

func UserServiceFactory(deps map[string]any, config map[string]any) any {
    return &UserService{
        db: deps["database"].(*Database),
    }
}
```

**2. Register Service Factory** (one-time setup):
```go
lokstra_registry.RegisterServiceType(
    "user-service-factory",
    UserServiceFactory,
    nil,
    deploy.WithResource("user", "users"),
    deploy.WithConvention("rest"),
)
```

**3. Configure Deployment** (`config.yaml`):
```yaml
service-definitions:
  user-service:
    type: user-service-factory
    depends-on: [database]
  
  database:
    type: database-factory

deployments:
  production:
    servers:
      api:
        addr: ":8080"
        published-services: [user-service]  # Auto-generates REST router!
```

**4. Run Application**:
```go
func main() {
    RegisterServiceAndMiddlewareTypes()

    lokstra_registry.RunServerFromConfig()
}
```

**Generated Routes:**
```
GET    /users       → UserService.GetAll()
GET    /users/{id}  → UserService.GetByID()
POST   /users       → UserService.Create()
PUT    /users/{id}  → UserService.Update() (if defined)
DELETE /users/{id}  → UserService.Delete() (if defined)
```

---

## Learning Path

### 🎯 1. [Service Management](./02-service/)
Learn the heart of Lokstra Framework - lazy dependency injection and service architecture.

**What you'll learn:**
- ✅ Type-safe lazy loading with `service.LazyLoad[T]`
- ✅ Service registration and factory patterns
- ✅ Dependency resolution and caching
- ✅ Testing strategies with mocked dependencies

**Key concepts:**
```go
// Service-level lazy loading - services created on first access
var userService = service.LazyLoad[*UserService]("user-service")
var database = service.LazyLoad[*Database]("database")

func handler() {
    // Thread-safe, dependencies eagerly loaded when service created
    users := userService.MustGet().GetAll()
}
```

---

### 📋 2. [Configuration Management](./04-configuration/)
Master YAML-driven configuration and environment management.

**What you'll learn:**
- ✅ Service definitions and dependencies
- ✅ Multi-environment deployments
- ✅ Environment variable overrides
- ✅ Deployment topology strategies

**Key concepts:**
```yaml
service-definitions:
  user-service:
    type: user-service-factory
    depends-on: [database, cache]

deployments:
  development:
    servers:
      dev: {addr: ":3000", published-services: [user-service]}
  
  production:
    servers:
      api: {addr: ":8080", published-services: [user-service]}
      worker: {addr: ":8081", published-services: [background-jobs]}
```

---

### 🚦 3. [Auto Router Generation](./01-router/)
Generate REST APIs automatically from service method signatures.

**What you'll learn:**
- ✅ Convention-based routing from method names
- ✅ Parameter binding and validation
- ✅ Route customization and overrides
- ✅ Middleware integration

**Key concepts:**
```go
// Method signature determines HTTP route
func (s *UserService) GetAll(p *GetAllParams) ([]User, error)     // GET /users
func (s *UserService) GetByID(p *GetByIDParams) (*User, error)    // GET /users/{id}
func (s *UserService) Create(p *CreateParams) (*User, error)      // POST /users

// Register with auto-router
lokstra_registry.RegisterServiceType("user-service-factory", NewUserService, nil,
    deploy.WithResource("user", "users"))
```

---

### 🔧 4. [Middleware & Plugins](./03-middleware/)
Add cross-cutting concerns like authentication, logging, and validation.

**What you'll learn:**
- ✅ Framework middleware integration
- ✅ Service-level middleware
- ✅ Request/Response transformation
- ✅ Error handling patterns

---

### 🚀 5. [Application & Server Management](./05-app-and-server/)
Deploy and manage your application across different environments.

**What you'll learn:**
- ✅ Multi-server deployments
- ✅ Monolith vs microservices patterns
- ✅ Health checks and monitoring
- ✅ Graceful shutdown handling

---

## Framework Benefits

### vs Manual Service Management
- ✅ **No initialization order issues** - lazy loading handles dependencies
- ✅ **Type-safe dependency injection** - compile-time guarantees
- ✅ **Declarative configuration** - YAML instead of complex bootstrap code
- ✅ **Easy testing** - mock services via registry

### vs Other Go DI Frameworks
- ✅ **Zero reflection overhead** - uses generics, not `any`
- ✅ **Lazy by default** - memory efficient, fast startup
- ✅ **Auto-generated routers** - no controller boilerplate
- ✅ **Configuration flexibility** - works with or without YAML

### vs Traditional Frameworks (NestJS/Spring)
- ✅ **Better performance** - compiled Go vs runtime interpretation
- ✅ **Simpler deployment** - single binary vs complex packaging
- ✅ **Flexible topology** - monolith ↔ microservices without code changes
- ✅ **Type safety** - compile-time errors vs runtime discovery

---

## When to Use Framework Track

Choose **Framework Track** when you're building:

- ✅ **Enterprise applications** with multiple services
- ✅ **Applications that may scale** from monolith to microservices
- ✅ **Team projects** needing consistent patterns
- ✅ **API-heavy applications** with many endpoints
- ✅ **Applications with complex dependencies** between components

If you just need simple routing, consider **[Router Track](../01-router-guide/)** instead.

---

## Next Steps

1. **[Start with Service Management](./02-service/)** - Core framework concepts
2. **[Learn Configuration](./04-configuration/)** - YAML-driven setup
3. **[Explore Auto Routers](./01-router/)** - Generate REST APIs
4. **[Add Middleware](./03-middleware/)** - Cross-cutting concerns
5. **[Deploy Applications](./05-app-and-server/)** - Production patterns

**Ready to build enterprise applications with Lokstra? Let's start with services! 🚀**

---

## Quick Links

- 🎯 **[Router Track](../01-router-guide/)** - Use as Router only (like Echo/Gin)
- 💡 **[Introduction](../00-introduction/)** - New to Lokstra?
- 📖 **[API Reference](../03-api-reference/)** - Complete technical docs
- 🧩 **[Examples](../00-introduction/examples/)** - Working code samples
- 🔄 **[Lokstra vs NestJS](./lokstra-vs-nestjs)** - TypeScript framework comparison
- ☕ **[Lokstra vs Spring Boot](./lokstra-vs-spring-boot)** - Java framework comparison