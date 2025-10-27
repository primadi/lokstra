---
layout: docs
title: Service Registration
---

# Service Registration

> Detailed guide to service registration patterns and factory functions

## Overview

Lokstra provides multiple ways to register services with flexible factory signatures, automatic dependency injection, and lazy loading support. This guide covers all service registration patterns in detail.

## Import Path

```go
import (
    "github.com/primadi/lokstra/core/deploy"
    "github.com/primadi/lokstra/core/deploy/schema"
    "github.com/primadi/lokstra/lokstra_registry"
)
```

---

## Service Type Registration

### RegisterServiceType

The primary method for registering reusable service factories.

**Signature:**
```go
func RegisterServiceType(
    serviceType string,
    local, remote any,
    options ...deploy.RegisterServiceTypeOption,
)
```

**Parameters:**
- `serviceType` - Unique identifier for this service type
- `local` - Local factory (in-process implementation)
- `remote` - Remote factory (API client wrapper), or `nil` if not needed
- `options` - Optional metadata for auto-router, dependencies, etc.

---

## Factory Signatures

Lokstra auto-wraps three factory signatures to the canonical form:

### 1. No Parameters (Simplest)
```go
func() any
```

**Use When:**
- Service has no configuration
- Service has no dependencies
- Simple stateless services

**Example:**
```go
lokstra_registry.RegisterServiceType("health-checker",
    func() any {
        return &HealthChecker{}
    },
    nil,
)
```

---

### 2. Config Only
```go
func(cfg map[string]any) any
```

**Use When:**
- Service needs configuration
- No dependencies on other services
- Configuration-driven initialization

**Example:**
```go
lokstra_registry.RegisterServiceType("db-service",
    func(cfg map[string]any) any {
        dsn := cfg["dsn"].(string)
        maxConn := cfg["max_connections"].(int)
        return db.NewConnection(dsn, maxConn)
    },
    nil,
)
```

**Config in YAML:**
```yaml
service-definitions:
  db:
    type: db-service
    config:
      dsn: "postgresql://localhost/mydb"
      max_connections: 10
```

---

### 3. Dependencies + Config (Full Control)
```go
func(deps, cfg map[string]any) any
```

**Use When:**
- Service depends on other services
- Need both dependencies and configuration
- Complex initialization logic

**Example:**
```go
lokstra_registry.RegisterServiceType("order-service",
    func(deps, cfg map[string]any) any {
        // Extract dependencies (lazy-loaded wrappers)
        userSvc := deps["userService"].(*service.Cached[*UserService])
        paymentSvc := deps["paymentService"].(*service.Cached[*PaymentService])
        
        // Extract config
        maxOrders := cfg["max_orders"].(int)
        
        return &OrderService{
            userService:    userSvc,
            paymentService: paymentSvc,
            maxOrders:      maxOrders,
        }
    },
    nil,
    deploy.WithDependencies("userService", "paymentService"),
)
```

**Dependency Resolution:**
- Dependencies are provided as `*service.Cached[T]` (lazy wrappers)
- Call `.Get()` to resolve the actual service instance
- Thread-safe, single initialization
- Prevents circular dependency issues

---

## Registration Options

### WithResource
Specifies resource names for auto-router generation.

**Signature:**
```go
deploy.WithResource(singular, plural string)
```

**Example:**
```go
lokstra_registry.RegisterServiceType("user-service",
    userFactory,
    nil,
    deploy.WithResource("user", "users"),
)
```

**Generated Routes:**
```
GET    /users           -> List
POST   /users           -> Create
GET    /users/:id       -> Get
PUT    /users/:id       -> Update
DELETE /users/:id       -> Delete
```

---

### WithConvention
Specifies routing convention (default: "rest").

**Signature:**
```go
deploy.WithConvention(convention string)
```

**Supported Conventions:**
- `"rest"` - RESTful routing (default)
- `"rpc"` - RPC-style routing
- Custom conventions (register with `convention.Register`)

**Example:**
```go
lokstra_registry.RegisterServiceType("user-service",
    userFactory,
    nil,
    deploy.WithResource("user", "users"),
    deploy.WithConvention("rest"),
)
```

---

### WithDependencies
Declares service dependencies for automatic injection.

**Signature:**
```go
deploy.WithDependencies(deps ...string)
```

**Example:**
```go
lokstra_registry.RegisterServiceType("order-service",
    orderFactory,
    nil,
    deploy.WithDependencies("userService", "paymentService", "db"),
)
```

**Dependency Mapping:**
```yaml
service-definitions:
  order-svc:
    type: order-service
    depends-on:
      - user-svc
      - payment-svc
      - db-connection
```

Framework automatically maps:
- `"userService"` → `user-svc` instance
- `"paymentService"` → `payment-svc` instance
- `"db"` → `db-connection` instance

---

### WithPathPrefix
Sets path prefix for all routes.

**Signature:**
```go
deploy.WithPathPrefix(prefix string)
```

**Example:**
```go
lokstra_registry.RegisterServiceType("user-service",
    userFactory,
    nil,
    deploy.WithResource("user", "users"),
    deploy.WithPathPrefix("/api/v1"),
)
```

**Generated Routes:**
```
GET    /api/v1/users
POST   /api/v1/users
GET    /api/v1/users/:id
...
```

---

### WithMiddleware
Attaches middleware to all service routes.

**Signature:**
```go
deploy.WithMiddleware(names ...string)
```

**Example:**
```go
lokstra_registry.RegisterServiceType("user-service",
    userFactory,
    nil,
    deploy.WithResource("user", "users"),
    deploy.WithMiddleware("auth", "logger", "rate-limiter"),
)
```

---

### WithRouteOverride
Customizes path for specific methods.

**Signature:**
```go
deploy.WithRouteOverride(methodName, path string)
```

**Example:**
```go
lokstra_registry.RegisterServiceType("user-service",
    userFactory,
    nil,
    deploy.WithResource("user", "users"),
    deploy.WithRouteOverride("Login", "/auth/login"),
    deploy.WithRouteOverride("Logout", "/auth/logout"),
)
```

**Result:**
```
POST /auth/login  -> user-service.Login()
POST /auth/logout -> user-service.Logout()
GET  /users       -> user-service.List()
```

---

### WithHiddenMethods
Excludes methods from auto-router generation.

**Signature:**
```go
deploy.WithHiddenMethods(methods ...string)
```

**Example:**
```go
lokstra_registry.RegisterServiceType("user-service",
    userFactory,
    nil,
    deploy.WithResource("user", "users"),
    deploy.WithHiddenMethods("Delete", "InternalHelper"),
)
```

---

## Local vs Remote Factories

### Local Factory
In-process service implementation.

**Example:**
```go
func UserServiceLocalFactory(deps, cfg map[string]any) any {
    db := deps["db"].(*service.Cached[*DBService])
    return &UserService{
        db: db,
    }
}
```

---

### Remote Factory
API client wrapper for external services.

**Example:**
```go
func UserServiceRemoteFactory(deps, cfg map[string]any) any {
    proxyService := deps["remote"].(*proxy.Service)
    return &UserServiceRemote{
        proxy: proxyService,
    }
}
```

**Registration:**
```go
lokstra_registry.RegisterServiceType("user-service",
    UserServiceLocalFactory,  // Local implementation
    UserServiceRemoteFactory, // Remote implementation
    deploy.WithResource("user", "users"),
    deploy.WithDependencies("db"),
)
```

**Framework automatically:**
- Uses local factory for local services
- Uses remote factory when service URL is provided
- Injects `remote` dependency with `*proxy.Service`

---

## Service Definition

### DefineService (Code)
Code-based service instance definition.

**Signature:**
```go
func DefineService(def *schema.ServiceDef)
```

**Example:**
```go
lokstra_registry.DefineService(&schema.ServiceDef{
    Name: "user-service",
    Type: "user-service-factory",
    Config: map[string]any{
        "max_connections": 100,
    },
    DependsOn: []string{"db-service"},
})
```

---

### DefineService (YAML)
YAML-based service definition.

**Example:**
```yaml
service-definitions:
  user-service:
    type: user-service-factory
    config:
      max_connections: 100
    depends-on:
      - db-service
```

---

## Lazy Service Registration

### RegisterLazyService
Simple lazy service registration without explicit dependency declaration.

**Signature:**
```go
func RegisterLazyService(name string, factory any, config map[string]any)
```

**Factory Signatures:**
```go
func() any                    // No params
func(cfg map[string]any) any  // With config
```

**Benefits:**
- ✅ Register services in any order
- ✅ Dependencies resolved on first access
- ✅ Thread-safe singleton
- ✅ Services only created when needed

**Example:**
```go
// No order required!
lokstra_registry.RegisterLazyService("user-service", func() any {
    repo := lokstra_registry.MustGetService[*UserRepo]("user-repo")
    return &UserService{repo: repo}
}, nil)

lokstra_registry.RegisterLazyService("user-repo", func() any {
    db := lokstra_registry.MustGetService[*DB]("db")
    return &UserRepository{db: db}
}, nil)

lokstra_registry.RegisterLazyService("db", func(cfg map[string]any) any {
    return db.Connect(cfg["dsn"].(string))
}, map[string]any{"dsn": "postgresql://localhost/mydb"})
```

---

### RegisterLazyServiceWithDeps
Lazy service with explicit dependency injection.

**Signature:**
```go
func RegisterLazyServiceWithDeps(
    name string,
    factory any,
    deps map[string]string,
    config map[string]any,
    opts ...deploy.LazyServiceOption,
)
```

**Factory Signature:**
```go
func(deps, cfg map[string]any) any
```

**Dependency Mapping:**
```go
deps := map[string]string{
    "userRepo":   "user-repo",     // key in factory -> service name
    "paymentSvc": "payment-service",
}
```

**Benefits:**
- ✅ Explicit dependency declaration
- ✅ Framework auto-injects dependencies
- ✅ No manual `MustGetService()` calls
- ✅ Clear dependency graph

**Example:**
```go
lokstra_registry.RegisterLazyServiceWithDeps("order-service",
    func(deps, cfg map[string]any) any {
        // deps already contains resolved services!
        userRepo := deps["userRepo"].(*UserRepository)
        paymentSvc := deps["paymentSvc"].(*PaymentService)
        maxOrders := cfg["max_orders"].(int)
        
        return &OrderService{
            userRepo:   userRepo,
            paymentSvc: paymentSvc,
            maxOrders:  maxOrders,
        }
    },
    map[string]string{
        "userRepo":   "user-repo",
        "paymentSvc": "payment-service",
    },
    map[string]any{"max_orders": 100},
)
```

**Registration Modes:**
```go
// Panic if already registered (default)
lokstra_registry.RegisterLazyServiceWithDeps(name, factory, deps, cfg)

// Skip if already registered (idempotent)
lokstra_registry.RegisterLazyServiceWithDeps(name, factory, deps, cfg,
    deploy.WithRegistrationMode(deploy.LazyServiceSkip))

// Override existing registration
lokstra_registry.RegisterLazyServiceWithDeps(name, factory, deps, cfg,
    deploy.WithRegistrationMode(deploy.LazyServiceOverride))
```

---

## Complete Examples

### Simple Service (No Dependencies)
```go
package main

import "github.com/primadi/lokstra/lokstra_registry"

type HealthChecker struct{}

func (h *HealthChecker) Check() string {
    return "OK"
}

func main() {
    // Register type
    lokstra_registry.RegisterServiceType("health-checker",
        func() any {
            return &HealthChecker{}
        },
        nil,
    )
    
    // Define instance
    lokstra_registry.DefineService(&schema.ServiceDef{
        Name: "health",
        Type: "health-checker",
    })
    
    // Access
    health := lokstra_registry.MustGetService[*HealthChecker]("health")
    status := health.Check()
}
```

---

### Service with Config
```go
type DBService struct {
    dsn     string
    maxConn int
}

func main() {
    lokstra_registry.RegisterServiceType("db-service",
        func(cfg map[string]any) any {
            return &DBService{
                dsn:     cfg["dsn"].(string),
                maxConn: cfg["max_connections"].(int),
            }
        },
        nil,
    )
    
    lokstra_registry.DefineService(&schema.ServiceDef{
        Name: "db",
        Type: "db-service",
        Config: map[string]any{
            "dsn":             "postgresql://localhost/mydb",
            "max_connections": 10,
        },
    })
}
```

---

### Service with Dependencies
```go
type UserService struct {
    db     *service.Cached[*DBService]
    logger *service.Cached[*Logger]
}

func (s *UserService) GetUser(id int) (*User, error) {
    db := s.db.Get()     // Lazy load
    logger := s.logger.Get()
    
    logger.Info("Getting user", id)
    return db.QueryUser(id)
}

func main() {
    lokstra_registry.RegisterServiceType("user-service",
        func(deps, cfg map[string]any) any {
            return &UserService{
                db:     deps["db"].(*service.Cached[*DBService]),
                logger: deps["logger"].(*service.Cached[*Logger]),
            }
        },
        nil,
        deploy.WithDependencies("db", "logger"),
    )
    
    lokstra_registry.DefineService(&schema.ServiceDef{
        Name:      "user-svc",
        Type:      "user-service",
        DependsOn: []string{"db", "logger-svc"},
    })
}
```

---

### Service with Auto-Router
```go
type UserService struct {
    db *service.Cached[*DBService]
}

func (s *UserService) List(ctx *request.Context) error {
    users := s.db.Get().QueryAll()
    return ctx.Api.Ok(users)
}

func (s *UserService) Get(ctx *request.Context) error {
    id := ctx.Req.PathParam("id")
    user := s.db.Get().QueryUser(id)
    return ctx.Api.Ok(user)
}

func (s *UserService) Create(ctx *request.Context) error {
    var user User
    ctx.Req.BindJSON(&user)
    s.db.Get().Insert(&user)
    return ctx.Api.Created(user)
}

func main() {
    lokstra_registry.RegisterServiceType("user-service",
        userFactory,
        nil,
        deploy.WithResource("user", "users"),
        deploy.WithConvention("rest"),
        deploy.WithDependencies("db"),
        deploy.WithMiddleware("auth", "logger"),
    )
    
    // Auto-generates:
    // GET    /users     -> List
    // POST   /users     -> Create
    // GET    /users/:id -> Get
    // PUT    /users/:id -> Update
    // DELETE /users/:id -> Delete
}
```

---

### Remote Service Pattern
```go
// Local implementation
type UserServiceLocal struct {
    db *service.Cached[*DBService]
}

func (s *UserServiceLocal) GetUser(id int) (*User, error) {
    return s.db.Get().QueryUser(id)
}

// Remote implementation (API client)
type UserServiceRemote struct {
    proxy *proxy.Service
}

func (s *UserServiceRemote) GetUser(id int) (*User, error) {
    return proxy.CallWithData[*User](s.proxy, "GetUser", id)
}

// Factories
func UserServiceLocalFactory(deps, cfg map[string]any) any {
    return &UserServiceLocal{
        db: deps["db"].(*service.Cached[*DBService]),
    }
}

func UserServiceRemoteFactory(deps, cfg map[string]any) any {
    return &UserServiceRemote{
        proxy: deps["remote"].(*proxy.Service),
    }
}

func main() {
    lokstra_registry.RegisterServiceType("user-service",
        UserServiceLocalFactory,
        UserServiceRemoteFactory,
        deploy.WithResource("user", "users"),
        deploy.WithDependencies("db"),
    )
}
```

**YAML Configuration:**
```yaml
# Local service
servers:
  - name: api
    services:
      - user-service

# Remote service
servers:
  - name: web
    external-service-definitions:
      user-service:
        url: http://api.example.com
        type: user-service  # Uses remote factory
```

---

## Best Practices

### 1. Use Appropriate Factory Signature
```go
// ✅ Good: Minimal signature for simple services
lokstra_registry.RegisterServiceType("health", func() any {
    return &HealthChecker{}
}, nil)

// 🚫 Avoid: Unnecessary params
lokstra_registry.RegisterServiceType("health", func(deps, cfg map[string]any) any {
    return &HealthChecker{}
}, nil)
```

---

### 2. Declare Dependencies Explicitly
```go
// ✅ Good: Clear dependencies
lokstra_registry.RegisterServiceType("user-service", factory, nil,
    deploy.WithDependencies("db", "logger"))

// 🚫 Avoid: Hidden dependencies
lokstra_registry.RegisterServiceType("user-service", factory, nil)
```

---

### 3. Use Lazy Wrappers for Dependencies
```go
// ✅ Good: Lazy-loaded dependencies
type UserService struct {
    db *service.Cached[*DBService]
}

// 🚫 Avoid: Direct dependencies (breaks lazy loading)
type UserService struct {
    db *DBService
}
```

---

### 4. Provide Both Local and Remote When Needed
```go
// ✅ Good: Supports both local and remote deployment
lokstra_registry.RegisterServiceType("user-service",
    UserServiceLocalFactory,
    UserServiceRemoteFactory,
    deploy.WithResource("user", "users"),
)

// 🚫 Avoid: Only local (can't call remotely)
lokstra_registry.RegisterServiceType("user-service",
    UserServiceLocalFactory,
    nil,
    deploy.WithResource("user", "users"),
)
```

---

## See Also

- **[lokstra_registry](./lokstra_registry.md)** - Registry API
- **[Service](../01-core-packages/service.md)** - Lazy service loading
- **[Middleware Registration](./middleware-registration.md)** - Middleware patterns
- **[Deploy](../03-configuration/deploy.md)** - Deployment configuration

---

## Related Guides

- **[Service Essentials](../../01-essentials/02-service/)** - Service basics
- **[Dependency Injection](../../02-deep-dive/service/)** - Advanced DI patterns
- **[Auto-Router](../../02-deep-dive/auto-router/)** - Auto-router generation
