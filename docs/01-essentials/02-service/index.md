---
layout: docs
title: Service - Essential Guide
---

# Service - Essential Guide

> **Service layer patterns and dependency injection**  
> **Time**: 45 minutes (with examples) • **Level**: Beginner to Intermediate

---

## 📖 What You'll Learn

- ✅ Service factory pattern and registration
- ✅ 3 ways to access services (and when to use each)
- ✅ LazyLoad for performance (20-100x faster!) ⭐
- ✅ Service dependencies and injection
- ✅ **Service as Router** - Auto-generate endpoints 🚀 (UNIQUE!)
- ✅ Best practices for production code

---

## 🎯 What is a Service?

A **Service** in Lokstra is a business logic container that:
- Encapsulates domain logic (users, orders, payments, etc)
- Can be registered in the global registry
- Can be accessed by handlers and other services
- Can automatically generate REST endpoints (Service as Router!)

**Key Insight**: Services are the backbone of your application architecture.

---

## 🚀 Quick Start (2 Minutes)

```go
package main

import (
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/core/service"
)

// 1. Define service
type UserService struct {
    users []User
}

func (s *UserService) GetAll() ([]User, error) {
    return s.users, nil
}

// 2. Create factory
func NewUserService() (*UserService, error) {
    return &UserService{
        users: []User{
            {ID: 1, Name: "Alice"},
            {ID: 2, Name: "Bob"},
        },
    }, nil
}

// 3. Register service
func main() {
    lokstra_registry.RegisterServiceFactory("users", NewUserService)
    
    // 4. Access in handler (LazyLoad - recommended!)
    var userService = service.LazyLoad[*UserService]("users")
    
    r := lokstra.NewRouter("api")
    r.GET("/users", func() ([]User, error) {
        return userService.MustGet().GetAll()
    })
    
    app := lokstra.NewApp("demo", ":3000", r)
    if err := app.Run(30 * time.Second); err != nil {
        fmt.Println("Error starting server:", err)
    }
}
```

---

## 📝 Service Registration Pattern

### 3 Ways to Register Services

Lokstra provides **two registration methods** with different characteristics:

| Method | When Created | Use Case | Factory Modes |
|--------|-------------|----------|---------------|
| `RegisterServiceFactory` | App startup (eager) | Simple services, always needed | 1 mode |
| `RegisterLazyService` | First access (lazy) | Complex deps, conditional use | 3 modes |

---

### Step 1: Define Service Struct

```go
type UserService struct {
    db       *Database  // Dependencies injected
    cache    *Cache
}

func (s *UserService) GetAll() ([]User, error) {
    // Business logic here
    return s.db.Query("SELECT * FROM users")
}

func (s *UserService) GetByID(id int) (*User, error) {
    return s.db.QueryOne("SELECT * FROM users WHERE id = ?", id)
}

func (s *UserService) Create(user *User) error {
    return s.db.Insert("users", user)
}
```

---

### Step 2a: Eager Registration (Simple Services)

**Use `RegisterServiceFactory`** for services that are always needed and have no complex dependencies:

```go
func NewUserService() (*UserService, error) {
    // Initialize service
    db, err := ConnectDatabase()
    if err != nil {
        return nil, err
    }
    
    cache := NewCache()
    
    return &UserService{
        db:    db,
        cache: cache,
    }, nil
}

func main() {
    // Eager: Created immediately at app startup
    lokstra_registry.RegisterServiceFactory("users", NewUserService)
    
    app := lokstra.NewApp("myapp", ":8080", routers...)
    if err := app.Run(30 * time.Second); err != nil {
        fmt.Println("Error starting server:", err)
    }
}
```

**Characteristics:**
- ✅ Created at app startup (before routes activated)
- ✅ Simpler for services without complex dependencies
- ✅ One factory signature: `func() (T, error)`
- ❌ Can't handle circular dependencies
- ❌ All services created even if unused

---

### Step 2b: Lazy Registration (Complex Dependencies) ⭐ RECOMMENDED

**Use `RegisterLazyService`** for services with dependencies or conditional usage:

```go
func main() {
    // Lazy: Created on first access, any order!
    // Supports 3 factory modes:
    
    // Mode 1: No params (simplest!)
    lokstra_registry.RegisterLazyService("cache", func() any {
        return NewCache()
    }, nil)
    
    // Mode 2: Config only
    lokstra_registry.RegisterLazyService("db", func(cfg map[string]any) any {
        dsn := cfg["dsn"].(string)
        return ConnectDatabase(dsn)
    }, map[string]any{
        "dsn": "postgresql://localhost/mydb",
    })
    
    // Mode 3: Full signature (deps + config)
    lokstra_registry.RegisterLazyService("users", func(deps, cfg map[string]any) any {
        // Get dependencies from registry
        db := lokstra_registry.MustGetService[*Database]("db")
        cache := lokstra_registry.MustGetService[*Cache]("cache")
        
        // Use config if needed
        timeout := cfg["timeout"].(int)
        
        return &UserService{db: db, cache: cache}
    }, map[string]any{
        "timeout": 30,
    })
    
    app := lokstra.NewApp("myapp", ":8080", routers...)
    if err := app.Run(30 * time.Second); err != nil {
        fmt.Println("Error starting server:", err)
    }
}
```

**Characteristics:**
- ✅ Created only when first accessed (lazy!)
- ✅ Register in any order (auto dependency resolution)
- ✅ **3 flexible factory signatures** (choose simplest that fits)
- ✅ Per-instance config (e.g., multiple DBs with different DSN)
- ✅ Thread-safe singleton
- ✅ Handles complex dependency graphs
- ⚠️ Slightly more complex API

**💭 Which to use?**
- **Simple services, no deps** → `RegisterServiceFactory`
- **Complex deps, conditional use** → `RegisterLazyService` ⭐
- **Most apps** → Mix both! (see Example 03)

---

### Step 3: Access Services (Same for Both!)

```go
// Both methods accessed the same way:
var userService = service.LazyLoad[*UserService]("users")

func handler() (*response.ApiHelper, error) {
    api := response.NewApiHelper()
    users, err := userService.MustGet().GetAll()
    // ...
}
```

**⚠️ Important**: Register services **before** `NewApp()`. Services are initialized during app creation (eager) or first access (lazy).

---

## 🔍 3 Ways to Access Services

### Method 1: GetService() - Direct Registry Lookup

**Use case**: Dynamic service names, prototypes, optional services

```go
r.GET("/users", func(ctx *request.Context) error {
    // ⚠️ Registry lookup EVERY request
    userService := lokstra_registry.GetService[*UserService]("users")
    
    // ⚠️ Must check for nil
    if userService == nil {
        return ctx.Api.InternalError("Service not found")
    }
    
    users, err := userService.GetAll()
    if err != nil {
        return ctx.Api.InternalError(err.Error())
    }
    
    return ctx.Api.Ok(users)
})
```

**Pros**:
- ✅ Simple and straightforward
- ✅ Works for dynamic service names
- ✅ Handles optional services

**Cons**:
- ❌ Slow (map lookup every request)
- ❌ Returns nil (confusing error messages)
- ❌ Verbose (need nil check)

**Performance**: ~100-200ns overhead per call

---

### Method 2: MustGetService() - Fail-Fast Lookup

**Use case**: Critical services, development, fail-fast behavior

```go
r.GET("/users", func(ctx *request.Context) error {
    // ⚠️ Panics if service not found
    userService := lokstra_registry.MustGetService[*UserService]("users")
    
    users, err := userService.GetAll()
    if err != nil {
        return ctx.Api.InternalError(err.Error())
    }
    
    return ctx.Api.Ok(users)
})
```

**Pros**:
- ✅ Clear error messages (panics with service name)
- ✅ No nil checks needed
- ✅ Fail-fast behavior

**Cons**:
- ❌ Slow (map lookup every request)
- ❌ Panics (not ideal for production APIs)

**Performance**: ~100-200ns overhead per call

---

### Method 3: service.LazyLoad() - Cached Access ⭐ RECOMMENDED

**Use case**: Production code, high-traffic endpoints, package-level access

```go
// Package-level: Cached after first access
var userService = service.LazyLoad[*UserService]("users")

r.GET("/users", func() (*response.ApiHelper, error) {
    api := response.NewApiHelper()
    
    // ✅ Cached! Only 1-5ns overhead
    users, err := userService.MustGet().GetAll()
    if err != nil {
        api.InternalError(err.Error())
        return api, nil
    }
    
    api.Ok(users)
    return api, nil
})
```

**Pros**:
- ✅ **20-100x faster** (cached after first access!)
- ✅ Clear errors with `.MustGet()`
- ✅ Clean code (no nil checks)
- ✅ Production-ready

**Cons**:
- ⚠️ Must be package-level or struct field (not function-local!)

**Performance**: 
- First access: ~100-200ns (one-time)
- Subsequent: ~1-5ns (cached)

---

## 🎨 LazyLoad: MustGet() vs Get()

### ✅ Recommended: MustGet()

**Clear error messages when service not found**:

```go
var userService = service.LazyLoad[*UserService]("users")

func handler(ctx *request.Context) error {
    users, err := userService.MustGet().GetAll()
    // If service not found:
    // Panic: "service 'users' not found or not initialized"
    // ✅ CLEAR! You know exactly what's wrong
}
```

---

### ⚠️ Not Recommended: Get()

**Confusing nil pointer errors**:

```go
var userService = service.LazyLoad[*UserService]("users")

func handler(ctx *request.Context) error {
    users, err := userService.Get().GetAll()
    // If service not found:
    // Panic: "runtime error: invalid memory address or nil pointer dereference"
    // ❌ CONFUSING! What caused nil? DB? Service? Something else?
}
```

**When to use Get()**:
Only when you want custom nil handling:

```go
svc := userService.Get()
if svc == nil {
    log.Warn("Service not available, using fallback")
    return fallbackResponse
}
users, err := svc.GetAll()
```

---

## 🔗 Service Dependencies

Services can depend on other services:

```go
type OrderService struct {
    userService    *UserService    // Dependency
    paymentService *PaymentService // Dependency
}

func NewOrderService() (*OrderService, error) {
    // Get dependencies from registry
    userSvc := lokstra_registry.MustGetService[*UserService]("users")
    paymentSvc := lokstra_registry.MustGetService[*PaymentService]("payments")
    
    return &OrderService{
        userService:    userSvc,
        paymentService: paymentSvc,
    }, nil
}

func (s *OrderService) CreateOrder(userID int, amount float64) (*Order, error) {
    // Use dependencies
    user, err := s.userService.GetByID(userID)
    if err != nil {
        return nil, err
    }
    
    payment, err := s.paymentService.Charge(user, amount)
    if err != nil {
        return nil, err
    }
    
    // Create order...
}
```

**Registration order**:
```go
func main() {
    // Register dependencies first
    lokstra_registry.RegisterServiceFactory("users", NewUserService)
    lokstra_registry.RegisterServiceFactory("payments", NewPaymentService)
    
    // Then register dependent services
    lokstra_registry.RegisterServiceFactory("orders", NewOrderService)
    
    app := lokstra.NewApp("myapp", ":8080", routers...)
    if err := app.Run(30 * time.Second); err != nil {
        fmt.Println("Error starting server:", err)
    }
}
```

---

## 🚀 Service as Router (UNIQUE FEATURE!)

**Automatically generate REST endpoints** from service methods!

### Traditional Approach (Manual)

```go
type UserService struct {
    users []User
}

func (s *UserService) GetAll() ([]User, error) { ... }
func (s *UserService) GetByID(id int) (*User, error) { ... }
func (s *UserService) Create(user *User) error { ... }

// Register service
lokstra_registry.RegisterServiceFactory("users", NewUserService)

// ❌ Manually create router and handlers
r := lokstra.NewRouter("api")
r.GET("/users", handleGetAll)
r.GET("/users/{id}", handleGetByID)
r.POST("/users", handleCreate)
// Tedious!
```

---

### Service as Router (Automatic!) ⭐

```go
type UserService struct {
    users []User
}

func (s *UserService) GetAll() ([]User, error) { ... }
func (s *UserService) GetByID(id int) (*User, error) { ... }
func (s *UserService) Create(user *User) error { ... }

// Register service WITH config
lokstra_registry.RegisterServiceFactory("users", NewUserService)
lokstra_registry.RegisterServiceConfig("users", map[string]any{
    "api.enabled": true,
    "api.prefix":  "/api/users",
})

// ✅ Auto-generate router!
userRouter := lokstra_registry.MustGetServiceAsRouter("users")

// Routes automatically created:
// GET  /api/users          → GetAll()
// GET  /api/users/{id}     → GetByID(id)
// POST /api/users          → Create(user)
```

**Benefits**:
- ✅ No boilerplate handler code
- ✅ Type-safe automatically
- ✅ Consistent API structure
- ✅ Faster development
- ✅ Less code to maintain

**See Example 04 for full details!**

---

## 🧪 Examples

All examples are runnable! Navigate to each folder and `go run main.go`

**Total learning time**: ~50 minutes

### [01 - Simple Service](examples/01-simple-service/) ⏱️ 10 min
**Learn**: Service registration, factory pattern, basic access

```go
lokstra_registry.RegisterServiceFactory("users", NewUserService)
var userService = service.LazyLoad[*UserService]("users")
users, err := userService.MustGet().GetAll()
```

**Key Concepts**: Factory pattern, registration, LazyLoad, MustGet()

---

### [02 - LazyLoad vs GetService](examples/02-lazyload-vs-getservice/) ⏱️ 12 min
**Learn**: Performance comparison, when to use each method

```go
// Slow: GetService (100-200ns per call)
userService := lokstra_registry.GetService[*UserService]("users")

// Fast: LazyLoad (1-5ns after first access)
var userService = service.LazyLoad[*UserService]("users")
users := userService.MustGet().GetAll()
```

**Key Concepts**: Performance, benchmarking, best practices

---

### [03 - Service Dependencies](examples/03-service-dependencies/) ⏱️ 15 min ⭐
**Learn**: Lazy registration, 3 factory modes, auto dependency resolution

```go
// Register in ANY order! Dependencies auto-resolved

// Mode 1: No params (simplest)
lokstra_registry.RegisterLazyService("user-repo", func() any {
    return repository.NewUserRepository()
}, nil)

// Mode 2: Config only
lokstra_registry.RegisterLazyService("db", func(cfg map[string]any) any {
    return db.NewConnection(cfg["dsn"].(string))
}, map[string]any{"dsn": "postgresql://localhost/main"})

// Mode 3: Full signature (deps + config)
lokstra_registry.RegisterLazyService("order-service", func(deps, cfg map[string]any) any {
    userSvc := lokstra_registry.MustGetService[*UserService]("user-service")
    return service.NewOrderService(userSvc)
}, nil)
```

**Key Concepts**: LazyService registration, 3 factory modes, no ordering required, per-instance config

---

### [04 - Service as Router](examples/04-service-as-router/) ⏱️ 20 min ⭐
**Learn**: Auto-generate endpoints from service methods (UNIQUE!)

```go
lokstra_registry.RegisterServiceConfig("users", map[string]any{
    "api.enabled": true,
    "api.prefix":  "/api/users",
})

// Auto-generates:
// GET  /api/users          → GetAll()
// GET  /api/users/{id}     → GetByID(id)
// POST /api/users          → Create(user)
// PUT  /api/users/{id}     → Update(id, user)
// DELETE /api/users/{id}   → Delete(id)
```

**Key Concepts**: Code generation, convention over configuration, rapid development

---

## 🎯 Common Patterns

### Pattern 1: Package-Level Service Access

```go
package handlers

import "github.com/primadi/lokstra/core/service"

// Package-level: Shared by all handlers
var (
    userService  = service.LazyLoad[*UserService]("users")
    orderService = service.LazyLoad[*OrderService]("orders")
)

func ListUsersHandler() (*response.ApiHelper, error) {
    api := response.NewApiHelper()
    users, err := userService.MustGet().GetAll()
    if err != nil {
        api.InternalError(err.Error())
        return api, nil
    }
    api.Ok(users)
    return api, nil
}
```

---

### Pattern 2: Struct Field Service Access

```go
type UserHandler struct {
    userService *service.Cached[*UserService]
}

func NewUserHandler() *UserHandler {
    return &UserHandler{
        userService: service.LazyLoad[*UserService]("users"),
    }
}

func (h *UserHandler) List(ctx *request.Context) error {
    users, err := h.userService.MustGet().GetAll()
    return ctx.Api.Ok(users)
}
```

---

### Pattern 3: Service with Repository Pattern

```go
type UserRepository interface {
    FindAll() ([]User, error)
    FindByID(id int) (*User, error)
    Create(user *User) error
}

type UserService struct {
    repo UserRepository
}

func NewUserService() (*UserService, error) {
    return &UserService{
        repo: NewPostgresUserRepository(),
    }, nil
}

func (s *UserService) GetAll() ([]User, error) {
    return s.repo.FindAll()
}
```

---

## 🚫 Common Mistakes

### ❌ Don't: Use LazyLoad in Function Scope

```go
func handler(ctx *request.Context) error {
    // ❌ Created every request! Cache useless!
    userService := service.LazyLoad[*UserService]("users")
    users, err := userService.MustGet().GetAll()
}
```

**✅ Do**: Use at package or struct level
```go
// ✅ Package-level: Created once, cached forever
var userService = service.LazyLoad[*UserService]("users")

func handler(ctx *request.Context) error {
    users, err := userService.MustGet().GetAll()
}
```

---

### ❌ Don't: Register Services After App Creation

```go
app := lokstra.NewApp("myapp", ":8080", routers...)

// ❌ TOO LATE! Services already initialized
lokstra_registry.RegisterServiceFactory("users", NewUserService)
```

**✅ Do**: Register before app creation
```go
// ✅ Register first
lokstra_registry.RegisterServiceFactory("users", NewUserService)

// Then create app
app := lokstra.NewApp("myapp", ":8080", routers...)
```

---

### ❌ Don't: Ignore Factory Errors

```go
func NewUserService() (*UserService, error) {
    db, err := ConnectDatabase()
    // ❌ Ignoring error!
    return &UserService{db: db}, nil
}
```

**✅ Do**: Propagate errors
```go
func NewUserService() (*UserService, error) {
    db, err := ConnectDatabase()
    if err != nil {
        return nil, fmt.Errorf("failed to connect database: %w", err)
    }
    return &UserService{db: db}, nil
}
```

---

## 🎓 Best Practices

### 1. **Always Use LazyLoad in Production**

```go
// ✅ Production: Fast, cached, clear errors
var userService = service.LazyLoad[*UserService]("users")

func handler(ctx *request.Context) error {
    users, err := userService.MustGet().GetAll()
}
```

---

### 2. **Use MustGet() for Clear Errors**

```go
// ✅ Clear error: "service 'users' not found"
users, err := userService.MustGet().GetAll()

// ❌ Confusing error: "nil pointer dereference"
users, err := userService.Get().GetAll()
```

---

### 3. **Keep Services Focused**

```go
// ✅ Good: Focused on users
type UserService struct {
    repo UserRepository
}

func (s *UserService) GetAll() ([]User, error) { ... }
func (s *UserService) GetByID(id int) (*User, error) { ... }
func (s *UserService) Create(user *User) error { ... }

// ❌ Bad: Too many responsibilities
type GodService struct {
    userRepo    UserRepository
    orderRepo   OrderRepository
    paymentRepo PaymentRepository
}
```

---

### 4. **Use Interfaces for Dependencies**

```go
// ✅ Good: Interface for testability
type UserService struct {
    repo UserRepository  // Interface
}

// Easy to mock in tests
func TestUserService(t *testing.T) {
    mockRepo := &MockUserRepository{}
    service := &UserService{repo: mockRepo}
    // ...
}

// ❌ Bad: Concrete dependency
type UserService struct {
    repo *PostgresUserRepository  // Hard to test
}
```

---

### 5. **Register Services in Order**

```go
// ✅ Good: Dependencies first
lokstra_registry.RegisterServiceFactory("database", NewDatabase)
lokstra_registry.RegisterServiceFactory("cache", NewCache)
lokstra_registry.RegisterServiceFactory("users", NewUserService)  // Depends on database
lokstra_registry.RegisterServiceFactory("orders", NewOrderService) // Depends on users

// ❌ Bad: Random order (may fail)
lokstra_registry.RegisterServiceFactory("orders", NewOrderService) // Error: users not found
lokstra_registry.RegisterServiceFactory("users", NewUserService)
```

---

## 📚 What's Next?

You now understand:
- ✅ Service registration and factory pattern
- ✅ 3 ways to access services (GetService, MustGetService, LazyLoad)
- ✅ LazyLoad for production (20-100x faster!)
- ✅ Service dependencies and injection
- ✅ Service as Router (auto-generate endpoints!)
- ✅ Best practices

### Next Steps:

**Continue Learning**:  
1. 👉 **[03 - Middleware](../03-middleware/README.md)** - Cross-cutting concerns
2. 👉 **[04 - Configuration](../04-configuration/README.md)** - Config-driven services
3. 👉 **[05 - App and Server](../05-app-and-server/README.md)** - Application lifecycle

**Deep Dive Topics**:
- [Service Lifecycle](../../02-deep-dive/service/lifecycle.md) (coming soon)
- [Service as Router Details](../../02-deep-dive/service/router-generation.md) (coming soon)
- [Testing Services](../../02-deep-dive/service/testing.md) (coming soon)

---

## 🔍 Quick Reference

### Registration
```go
lokstra_registry.RegisterServiceFactory("name", factory)
lokstra_registry.RegisterServiceConfig("name", config)
```

### Access Methods
```go
// Slow, returns nil
GetService[T](name)

// Slow, panics
MustGetService[T](name)

// Fast, cached ⭐
service.LazyLoad[T](name).MustGet()
service.LazyLoad[T](name).Get()
```

### Service as Router
```go
lokstra_registry.MustGetServiceAsRouter("name")
```

---

**Continue learning** → [03 - Middleware](../03-middleware/README.md)
