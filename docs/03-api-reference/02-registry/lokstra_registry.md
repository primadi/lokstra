# lokstra_registry

> Simplified API for the Lokstra registry system

## Overview

The `lokstra_registry` package provides a clean, package-level API for registering and accessing services, middleware, routers, and configuration. It wraps `deploy.GlobalRegistry()` to provide:

- ✅ Shorter import path
- ✅ Package-level functions (no singleton access)
- ✅ Generic helper functions like `GetService[T]`
- ✅ Cleaner developer experience

## Import Path

```go
import "github.com/primadi/lokstra/lokstra_registry"
```

---

## Service Registration

### RegisterServiceType
Registers a service factory with optional metadata for auto-router generation.

**Signature:**
```go
func RegisterServiceType(
    serviceType string,
    local, remote any,
    options ...deploy.RegisterServiceTypeOption,
)
```

**Parameters:**
- `serviceType` - Unique identifier for the service type
- `local` - Local factory function (for in-process services)
- `remote` - Remote factory function (for external API wrappers), or `nil`
- `options` - Optional metadata (resource name, convention, dependencies, etc.)

**Factory Signatures (All Auto-Wrapped):**
```go
// Simplest - no dependencies, no config
func() any

// With config
func(cfg map[string]any) any

// Full control - with dependencies and config
func(deps, cfg map[string]any) any
```

**Example:**
```go
// Simple factory (no deps, no config)
lokstra_registry.RegisterServiceType("user-service",
    func() any {
        return &UserService{}
    },
    nil, // No remote implementation
    deploy.WithResource("user", "users"),
)

// With config
lokstra_registry.RegisterServiceType("db-service",
    func(cfg map[string]any) any {
        dsn := cfg["dsn"].(string)
        return db.NewConnection(dsn)
    },
    nil,
)

// Full signature with deps
lokstra_registry.RegisterServiceType("order-service",
    func(deps, cfg map[string]any) any {
        userSvc := deps["userService"].(*service.Cached[*UserService])
        maxItems := cfg["max_items"].(int)
        return &OrderService{
            userService: userSvc,
            maxItems:    maxItems,
        }
    },
    nil,
    deploy.WithDependencies("userService"),
)

// With remote implementation
lokstra_registry.RegisterServiceType("payment-service",
    // Local implementation
    func(deps, cfg map[string]any) any {
        return &PaymentService{db: deps["db"].(*DBService)}
    },
    // Remote implementation (API client)
    func(deps, cfg map[string]any) any {
        proxy := deps["remote"].(*proxy.Service)
        return &PaymentServiceRemote{proxy: proxy}
    },
    deploy.WithResource("payment", "payments"),
    deploy.WithDependencies("db"),
)
```

**Options:**
```go
deploy.WithResource(singular, plural string)
deploy.WithConvention(convention string)
deploy.WithDependencies(deps ...string)
deploy.WithPathPrefix(prefix string)
deploy.WithMiddleware(names ...string)
deploy.WithRouteOverride(methodName, path string)
deploy.WithHiddenMethods(methods ...string)
```

---

### DefineService
Defines a service instance in the global registry (code-based config).

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
        "max_connections": 10,
    },
    DependsOn: []string{"db-service"},
})
```

**Use Cases:**
- Code-based service configuration
- Dynamic service definitions
- Testing scenarios

---

### RegisterLazyService
Registers a lazy service that will be instantiated on first access.

**Signature:**
```go
func RegisterLazyService(name string, factory any, config map[string]any)
```

**Factory Signatures:**
```go
func() any                    // No params (simplest!)
func(cfg map[string]any) any  // With config
```

**Benefits:**
- ✅ No creation order requirements
- ✅ Dependencies resolved automatically
- ✅ Thread-safe singleton pattern
- ✅ Services only created when needed

**Example:**
```go
// With config
lokstra_registry.RegisterLazyService("db-main", func(cfg map[string]any) any {
    dsn := cfg["dsn"].(string)
    return db.NewConnection(dsn)
}, map[string]any{
    "dsn": "postgresql://localhost/main",
})

// Without params (resolve deps manually)
lokstra_registry.RegisterLazyService("user-repo", func() any {
    db := lokstra_registry.MustGetService[*DB]("db-main")
    return repository.NewUserRepository(db)
}, nil)
```

---

### RegisterLazyServiceWithDeps
Registers a lazy service with explicit dependency injection.

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

**Parameters:**
- `name` - Service name
- `factory` - Factory function `func(deps, cfg map[string]any) any`
- `deps` - Dependency mapping: `map[factoryKey]serviceName`
- `config` - Service configuration
- `opts` - Optional registration mode

**Example:**
```go
lokstra_registry.RegisterLazyServiceWithDeps("order-service",
    func(deps, cfg map[string]any) any {
        // deps already contains resolved services!
        userSvc := deps["userService"].(*UserService)
        orderRepo := deps["orderRepo"].(*OrderRepository)
        maxItems := cfg["max_items"].(int)
        return &OrderService{
            userService: userSvc,
            orderRepo:   orderRepo,
            maxItems:    maxItems,
        }
    },
    map[string]string{
        "userService": "user-service",
        "orderRepo":   "order-repo",
    },
    map[string]any{"max_items": 5},
)

// Skip if already registered
lokstra_registry.RegisterLazyServiceWithDeps(name, factory, deps, cfg,
    deploy.WithRegistrationMode(deploy.LazyServiceSkip))

// Override existing registration
lokstra_registry.RegisterLazyServiceWithDeps(name, factory, deps, cfg,
    deploy.WithRegistrationMode(deploy.LazyServiceOverride))
```

---

## Service Access

### GetService
Retrieves a service with type assertion (generic).

**Signature:**
```go
func GetService[T any](name string) T
```

**Returns:**
- Service instance of type `T`, or zero value if not found

**Example:**
```go
userSvc := lokstra_registry.GetService[*UserService]("user-service")
if userSvc != nil {
    users := userSvc.GetAll()
}
```

---

### MustGetService
Retrieves a service with type assertion (panics if not found).

**Signature:**
```go
func MustGetService[T any](name string) T
```

**Returns:**
- Service instance of type `T`

**Panics:**
- If service not found
- If type mismatch

**Example:**
```go
userSvc := lokstra_registry.MustGetService[*UserService]("user-service")
users := userSvc.GetAll() // Safe to use directly
```

**When to Use:**
- Required dependencies
- Fail-fast initialization
- Clear error reporting

---

### TryGetService
Retrieves a service with type assertion (safe version).

**Signature:**
```go
func TryGetService[T any](name string) (T, bool)
```

**Returns:**
- `(value, true)` if found and type matches
- `(zero, false)` otherwise

**Example:**
```go
if userSvc, ok := lokstra_registry.TryGetService[*UserService]("user-service"); ok {
    users := userSvc.GetAll()
} else {
    log.Println("User service not available")
}
```

---

### GetServiceAny
Retrieves a service without type assertion (non-generic).

**Signature:**
```go
func GetServiceAny(name string) (any, bool)
```

**Returns:**
- `(instance, true)` if found
- `(nil, false)` if not found

**Example:**
```go
instance, ok := lokstra_registry.GetServiceAny("user-service")
if ok {
    userSvc := instance.(*UserService)
    users := userSvc.GetAll()
}
```

---

### GetLazyService
Creates a lazy-loading service wrapper.

**Signature:**
```go
func GetLazyService[T any](serviceName string) *service.Cached[T]
```

**Returns:**
- `*service.Cached[T]` - Lazy service loader

**Example:**
```go
// In handler setup
type UserHandler struct {
    userService *service.Cached[*UserService]
}

func NewUserHandler() *UserHandler {
    return &UserHandler{
        userService: lokstra_registry.GetLazyService[*UserService]("user-service"),
    }
}

// In handler - service loaded only when first accessed
func (h *UserHandler) GetUsers(ctx *request.Context) error {
    users := h.userService.Get().GetAll() // Lazy loaded here!
    return ctx.Api.Ok(users)
}
```

---

### RegisterService
Registers a pre-instantiated service instance.

**Signature:**
```go
func RegisterService(name string, instance any)
```

**Example:**
```go
userSvc := &UserService{}
lokstra_registry.RegisterService("user-service", userSvc)
```

**Use Cases:**
- Manual service registration
- Testing with mocks
- Pre-initialized services

---

### GetServiceFactory
Returns the service factory for a service type.

**Signature:**
```go
func GetServiceFactory(serviceType string, isLocal bool) deploy.ServiceFactory
```

**Parameters:**
- `serviceType` - Service type name
- `isLocal` - `true` for local factory, `false` for remote

**Returns:**
- Factory function

---

## Middleware Registration

### RegisterMiddlewareFactory
Registers a middleware factory function.

**Signature:**
```go
func RegisterMiddlewareFactory(
    mwType string,
    factory any,
    opts ...RegisterOption,
)
```

**Factory Signatures:**
```go
func(config map[string]any) any
func(config map[string]any) request.HandlerFunc // Old pattern
```

**Example:**
```go
lokstra_registry.RegisterMiddlewareFactory("logger", func(cfg map[string]any) any {
    level := cfg["level"].(string)
    return func(ctx *request.Context) error {
        log.Printf("[%s] %s %s", level, ctx.R.Method, ctx.R.URL.Path)
        return ctx.Next()
    }
})

// Allow override
lokstra_registry.RegisterMiddlewareFactory("logger", newLoggerFactory,
    lokstra_registry.AllowOverride(true))
```

---

### RegisterMiddlewareName
Registers a named middleware instance with config.

**Signature:**
```go
func RegisterMiddlewareName(
    mwName string,
    mwType string,
    config map[string]any,
    opts ...RegisterOption,
)
```

**Example:**
```go
// Register factory
lokstra_registry.RegisterMiddlewareFactory("logger", loggerFactory)

// Register named instances with different configs
lokstra_registry.RegisterMiddlewareName("logger-debug", "logger",
    map[string]any{"level": "debug"})
lokstra_registry.RegisterMiddlewareName("logger-info", "logger",
    map[string]any{"level": "info"})
```

---

### RegisterMiddleware
Registers a pre-instantiated middleware.

**Signature:**
```go
func RegisterMiddleware(name string, handler request.HandlerFunc)
```

**Example:**
```go
logger := func(ctx *request.Context) error {
    log.Printf("%s %s", ctx.R.Method, ctx.R.URL.Path)
    return ctx.Next()
}
lokstra_registry.RegisterMiddleware("logger", logger)
```

---

### GetMiddleware
Retrieves a middleware instance.

**Signature:**
```go
func GetMiddleware(name string) (request.HandlerFunc, bool)
```

**Returns:**
- `(handler, true)` if found
- `(nil, false)` if not found

---

### CreateMiddleware
Creates a middleware from its definition and caches it.

**Signature:**
```go
func CreateMiddleware(name string) request.HandlerFunc
```

**Example:**
```go
logger := lokstra_registry.CreateMiddleware("logger-debug")
router.Use(logger)
```

---

## Router Registration

### RegisterRouter
Registers a router instance.

**Signature:**
```go
func RegisterRouter(name string, r router.Router)
```

**Example:**
```go
userRouter := lokstra.NewRouter()
userRouter.GET("/users", handlers.GetUsers)
lokstra_registry.RegisterRouter("user-router", userRouter)
```

---

### GetRouter
Retrieves a router instance.

**Signature:**
```go
func GetRouter(name string) router.Router
```

**Returns:**
- Router instance, or `nil` if not found

---

### GetAllRouters
Returns all registered routers.

**Signature:**
```go
func GetAllRouters() map[string]router.Router
```

**Returns:**
- Map of router name to router instance

---

## Configuration

### DefineConfig
Defines a configuration value.

**Signature:**
```go
func DefineConfig(name string, value any)
```

**Example:**
```go
lokstra_registry.DefineConfig("db.dsn", "postgresql://localhost/mydb")
lokstra_registry.DefineConfig("app.max_connections", 100)
```

---

### GetResolvedConfig
Gets a resolved configuration value.

**Signature:**
```go
func GetResolvedConfig(key string) (any, bool)
```

**Returns:**
- `(value, true)` if found
- `(nil, false)` if not found

---

### GetConfig
Retrieves a configuration with type assertion and default value.

**Signature:**
```go
func GetConfig[T any](name string, defaultValue T) T
```

**Returns:**
- Config value of type `T`, or `defaultValue` if not found

**Example:**
```go
dsn := lokstra_registry.GetConfig("db.dsn", "postgresql://localhost/default")
maxConn := lokstra_registry.GetConfig("app.max_connections", 10)
```

---

## Shutdown Management

### ShutdownServices
Gracefully shuts down all services implementing `Shutdownable`.

**Signature:**
```go
func ShutdownServices()
```

**Shutdownable Interface:**
```go
type Shutdownable interface {
    Shutdown() error
}
```

**Example Service:**
```go
type DatabaseService struct {
    conn *sql.DB
}

func (s *DatabaseService) Shutdown() error {
    log.Println("Closing database connection")
    return s.conn.Close()
}
```

**Usage in main.go:**
```go
func main() {
    // Register services
    lokstra_registry.RegisterService("db", dbService)
    
    // Setup graceful shutdown
    defer lokstra_registry.ShutdownServices()
    
    // Or with signal handling
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        log.Println("Shutting down...")
        lokstra_registry.ShutdownServices()
        os.Exit(0)
    }()
    
    // Start server
    if err := server.Run(30 * time.Second); err != nil {
        fmt.Println("Error starting server:", err)
    }
}
```

---

## Advanced Functions

### Global
Returns the underlying global registry instance.

**Signature:**
```go
func Global() *deploy.GlobalRegistry
```

**Returns:**
- Global registry instance

**Example:**
```go
registry := lokstra_registry.Global()
// Access low-level registry APIs
```

---

## Complete Examples

### Service Registration Pattern
```go
package main

import (
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/core/deploy"
    "github.com/primadi/lokstra/lokstra_registry"
)

func main() {
    // Register service types
    lokstra_registry.RegisterServiceType("user-service",
        func(deps, cfg map[string]any) any {
            db := deps["db"].(*service.Cached[*DBService])
            return &UserService{db: db}
        },
        nil,
        deploy.WithResource("user", "users"),
        deploy.WithDependencies("db"),
    )
    
    lokstra_registry.RegisterServiceType("db-service",
        func(cfg map[string]any) any {
            dsn := cfg["dsn"].(string)
            return db.Connect(dsn)
        },
        nil,
    )
    
    // Define service instances
    lokstra_registry.DefineService(&schema.ServiceDef{
        Name:      "db",
        Type:      "db-service",
        Config:    map[string]any{"dsn": "postgresql://localhost/mydb"},
    })
    
    lokstra_registry.DefineService(&schema.ServiceDef{
        Name:      "user-svc",
        Type:      "user-service",
        DependsOn: []string{"db"},
    })
    
    // Access services
    userSvc := lokstra_registry.MustGetService[*UserService]("user-svc")
    users := userSvc.GetAll()
}
```

### Middleware Registration Pattern
```go
func main() {
    // Register middleware factory
    lokstra_registry.RegisterMiddlewareFactory("logger", func(cfg map[string]any) any {
        level := cfg["level"].(string)
        return func(ctx *request.Context) error {
            log.Printf("[%s] %s %s", level, ctx.R.Method, ctx.R.URL.Path)
            return ctx.Next()
        }
    })
    
    // Register named instances
    lokstra_registry.RegisterMiddlewareName("logger-debug", "logger",
        map[string]any{"level": "DEBUG"})
    lokstra_registry.RegisterMiddlewareName("logger-info", "logger",
        map[string]any{"level": "INFO"})
    
    // Use in router
    router := lokstra.NewRouter()
    router.Use(lokstra_registry.CreateMiddleware("logger-info"))
    router.GET("/users", handlers.GetUsers)
}
```

### Lazy Service Pattern
```go
func main() {
    // Register lazy services (no order required!)
    lokstra_registry.RegisterLazyService("db", func(cfg map[string]any) any {
        return db.Connect(cfg["dsn"].(string))
    }, map[string]any{"dsn": "postgresql://localhost/mydb"})
    
    lokstra_registry.RegisterLazyService("user-repo", func() any {
        db := lokstra_registry.MustGetService[*DB]("db")
        return repository.NewUserRepository(db)
    }, nil)
    
    lokstra_registry.RegisterLazyService("user-service", func() any {
        repo := lokstra_registry.MustGetService[*UserRepo]("user-repo")
        return service.NewUserService(repo)
    }, nil)
    
    // Services created on first access
    userSvc := lokstra_registry.MustGetService[*UserService]("user-service")
}
```

---

## See Also

- **[Service](../01-core-packages/service.md)** - Lazy service loading
- **[Service Registration](./service-registration.md)** - Detailed registration patterns
- **[Middleware Registration](./middleware-registration.md)** - Middleware patterns
- **[Router Registration](./router-registration.md)** - Router factories
- **[Deploy](../03-configuration/deploy.md)** - Deployment configuration

---

## Related Guides

- **[Service Essentials](../../01-essentials/02-service/)** - Service basics
- **[Dependency Injection](../../02-deep-dive/service/)** - Advanced DI patterns
- **[Testing](../../04-guides/testing/)** - Testing strategies
