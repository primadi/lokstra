---
layout: docs
title: API Reference
---

# API Reference

> Complete API documentation for all Lokstra packages

This section provides comprehensive API documentation for every exported type, function, and method in the Lokstra framework.

---

## 📚 Quick Navigation

### Core Framework
Start here for building Lokstra applications:

- **[lokstra](01-core-packages/lokstra.md)** - Main package (NewRouter, NewApp, NewServer)
- **[Router](01-core-packages/router.md)** - HTTP routing and handler registration
- **[App](01-core-packages/app.md)** - Application listener and lifecycle
- **[Server](01-core-packages/server.md)** - Server management and graceful shutdown
- **[Request Context](01-core-packages/request.md)** - Request handling and context
- **[Response](01-core-packages/response.md)** - Response helpers and formatting
- **[Service](01-core-packages/service.md)** - Service utilities (LazyLoad, dependency injection)

### Registry & Configuration
Service registration, middleware, and configuration:

- **[lokstra_registry](02-registry/lokstra_registry.md)** - Main registry API
- **[Service Registration](02-registry/service-registration.md)** - RegisterServiceType, DefineService
- **[Middleware Registration](02-registry/middleware-registration.md)** - RegisterMiddlewareType, DefineMiddleware
- **[Router Registration](02-registry/router-registration.md)** - Router factories and auto-router
- **[Configuration](03-configuration/config.md)** - Config package (core/config)
- **[Deployment](03-configuration/deploy.md)** - Deployment loader (core/deploy)
- **[Schema](03-configuration/schema.md)** - YAML schema and validation

### HTTP Client
Remote service communication:

- **[API Client](04-client/api-client.md)** - HTTP client (api_client package)
- **[Client Router](04-client/client-router.md)** - ClientRouter for convention-based calls
- **[Remote Service](04-client/remote-service.md)** - Remote service patterns

### Built-in Middleware
Ready-to-use middleware:

- **[CORS](05-middleware/cors.md)** - Cross-Origin Resource Sharing
- **[Request Logger](05-middleware/request-logger.md)** - HTTP request logging
- **[JWT Auth](05-middleware/jwtauth.md)** - JWT authentication
- **[Recovery](05-middleware/recovery.md)** - Panic recovery
- **[Body Limit](05-middleware/body-limit.md)** - Request body size limiter
- **[Gzip Compression](05-middleware/gzipcompression.md)** - Response compression
- **[Access Control](05-middleware/accesscontrol.md)** - Permission-based access
- **[Slow Request Logger](05-middleware/slow-request-logger.md)** - Slow request detection

### Built-in Services
Standard services and service APIs:

- **[Database Pool (PostgreSQL)](06-services/dbpool-pg.md)** - PostgreSQL connection pool
- **[Redis](06-services/redis.md)** - Redis client service
- **[KV Store](06-services/kvstore.md)** - Key-value store interface
- **[Metrics (Prometheus)](06-services/metrics.md)** - Prometheus metrics
- **[Auth Service](06-services/auth-service.md)** - Authentication service
- **[Auth Flow (Password)](06-services/auth-flow-password.md)** - Password authentication
- **[Auth Flow (OTP)](06-services/auth-flow-otp.md)** - OTP authentication
- **[Auth Session (Redis)](06-services/auth-session-redis.md)** - Redis-based sessions
- **[Auth Token (JWT)](06-services/auth-token-jwt.md)** - JWT token management
- **[Auth User Repository (PostgreSQL)](06-services/auth-user-repo-pg.md)** - PostgreSQL user storage
- **[Auth Validator](06-services/auth-validator.md)** - Auth validation utilities

### Helper Packages
Utility functions and helpers:

- **[common/cast](07-helpers/common-cast.md)** - Type casting utilities
- **[common/json](07-helpers/common-json.md)** - JSON parsing and error handling
- **[common/validator](07-helpers/common-validator.md)** - Validation utilities
- **[common/utils](07-helpers/common-utils.md)** - General utilities
- **[common/customtype](07-helpers/common-customtype.md)** - Custom types (NullableTime, etc.)
- **[common/response_writer](07-helpers/common-response-writer.md)** - Response writer helpers

### Advanced Topics
Internal mechanisms and advanced patterns:

- **[Proxy](08-advanced/proxy.md)** - Remote service proxy (core/proxy)
- **[Route](08-advanced/route.md)** - Route definition internals (core/route)
- **[Auto-Router](08-advanced/auto-router.md)** - Auto-router generation
- **[Handler Utils](08-advanced/lokstra-handler.md)** - Built-in handlers (SPA, static, reverse proxy)

---

## 📦 Package Index

### Import Paths

```go
// Core framework
import "github.com/primadi/lokstra"
import "github.com/primadi/lokstra/core/router"
import "github.com/primadi/lokstra/core/app"
import "github.com/primadi/lokstra/core/server"
import "github.com/primadi/lokstra/core/request"
import "github.com/primadi/lokstra/core/response"
import "github.com/primadi/lokstra/core/service"

// Registry & config
import "github.com/primadi/lokstra/lokstra_registry"
import "github.com/primadi/lokstra/core/config"
import "github.com/primadi/lokstra/core/deploy"
import "github.com/primadi/lokstra/core/deploy/schema"

// Client
import "github.com/primadi/lokstra/api_client"

// Middleware (auto-registered when imported)
import _ "github.com/primadi/lokstra/middleware/cors"
import _ "github.com/primadi/lokstra/middleware/request_logger"
import _ "github.com/primadi/lokstra/middleware/jwtauth"
// ... see 05-middleware for complete list

// Services (auto-registered when imported)
import _ "github.com/primadi/lokstra/services/dbpool_pg"
import _ "github.com/primadi/lokstra/services/redis"
import _ "github.com/primadi/lokstra/services/auth_service"
// ... see 06-services for complete list

// Service APIs
import "github.com/primadi/lokstra/serviceapi"
import "github.com/primadi/lokstra/serviceapi/auth"

// Helpers
import "github.com/primadi/lokstra/common/cast"
import "github.com/primadi/lokstra/common/json"
import "github.com/primadi/lokstra/common/validator"
import "github.com/primadi/lokstra/common/utils"
import "github.com/primadi/lokstra/common/customtype"
import "github.com/primadi/lokstra/common/response_writer"

// Advanced
import "github.com/primadi/lokstra/core/proxy"
import "github.com/primadi/lokstra/core/route"
import "github.com/primadi/lokstra/lokstra_handler"
```

---

## 🎯 Common Use Cases

### Building a Basic API
```go
import "github.com/primadi/lokstra"

router := lokstra.NewRouter("api")
router.GET("/users", getUsersHandler)
router.POST("/users", createUserHandler)

app := lokstra.NewApp("api", ":8080", router)
server := lokstra.NewServer("my-server", app)
if err := server.Run(30 * time.Second); err != nil {
  fmt.Println("Error starting server:", err)
}
```

📖 See: [lokstra](01-core-packages/lokstra.md), [Router](01-core-packages/router.md)

### Using Services with Dependency Injection
```go
import "github.com/primadi/lokstra/lokstra_registry"
import "github.com/primadi/lokstra/core/service"

// Register service
lokstra_registry.RegisterServiceType("user-service",
    userServiceFactory, nil,
    deploy.WithResource("user", "users"),
)

// Use in another service
type OrderService struct {
    userService *service.Cached[*UserService]
}

func NewOrderService() *OrderService {
    return &OrderService{
        userService: service.LazyLoad[*UserService]("user-service"),
    }
}

func (s *OrderService) CreateOrder(userID int) {
    user := s.userService.Get() // Auto-loads on first access
    // ...
}
```

📖 See: [Service](01-core-packages/service.md), [lokstra_registry](02-registry/lokstra_registry.md)

### Calling Remote Services
```go
import "github.com/primadi/lokstra/api_client"

client := api_client.NewClientRouter("https://api.example.com")

// Convention-based call
user, err := api_client.FetchAndCast[*User](client, "/users/123")

// With options
users, err := api_client.FetchAndCast[[]User](client, "/users",
    api_client.WithMethod("GET"),
    api_client.WithQuery(map[string]string{"status": "active"}),
)
```

📖 See: [API Client](04-client/api-client.md), [Client Router](04-client/client-router.md)

### Using Middleware
```go
import "github.com/primadi/lokstra"
import _ "github.com/primadi/lokstra/middleware/cors"

router := lokstra.NewRouter("api")

// Apply middleware globally
router.Use("cors") // Middleware name from config/registry

// Apply to specific route
router.GET("/users", getUsersHandler, "auth", "logger")

// Apply to route group
admin := router.Group("/admin")
admin.Use("auth", "admin-only")
admin.GET("/users", listUsers)
```

📖 See: [Router](01-core-packages/router.md), [Middleware](05-middleware/)

### Configuration-Driven Deployment
```yaml
# config.yaml
service-definitions:
  user-service:
    type: user-service-factory
    depends-on: [db-service]

external-service-definitions:
  payment-gateway:
    url: "https://payment.example.com"
    type: payment-service-remote-factory

deployments:
  production:
    servers:
      api:
        base-url: "https://api.example.com"
        addr: ":8080"
        published-services:
          - user-service
```

```go
import "github.com/primadi/lokstra/core/deploy"

// Load config
server, err := deploy.LoadFromYamlFile("config.yaml", "production", "api")
server.Run(30 * time.Second)
```

📖 See: [Configuration](03-configuration/), [Schema](03-configuration/schema.md)

---

## 📖 Documentation Conventions

### Signature Format
```go
func FunctionName(param1 Type1, param2 Type2) ReturnType
```

### Type Definitions
```go
type TypeName struct {
    Field1 Type1 // Description
    Field2 Type2 // Description
}
```

### Generic Functions
```go
func GenericFunction[T any](param T) T
```

### Variadic Parameters
```go
func VariadicFunction(items ...Item)
```

### Functional Options
```go
func NewThing(opts ...Option) *Thing
```

---

## 🔍 Finding What You Need

### By Category
- **Routing & HTTP**: [01-core-packages](01-core-packages/)
- **Dependency Injection**: [02-registry](02-registry/), [Service](01-core-packages/service.md)
- **Configuration**: [03-configuration](03-configuration/)
- **Remote Calls**: [04-client](04-client/)
- **Pre-built Components**: [05-middleware](05-middleware/), [06-services](06-services/)
- **Utilities**: [07-helpers](07-helpers/)
- **Internals**: [08-advanced](08-advanced/)

### By Use Case
- **Building APIs**: Start with [lokstra](01-core-packages/lokstra.md) and [Router](01-core-packages/router.md)
- **Managing Services**: See [Service](01-core-packages/service.md) and [lokstra_registry](02-registry/lokstra_registry.md)
- **Microservices**: Check [Remote Service](04-client/remote-service.md) and [Proxy](08-advanced/proxy.md)
- **Auth & Security**: Browse [05-middleware](05-middleware/) and [06-services](06-services/)
- **Deployment**: Read [Configuration](03-configuration/)

---

## 💡 Tips

### IDE Support
Enable VS Code YAML validation:
```yaml
# yaml-language-server: $schema=https://primadi.github.io/lokstra/schema/lokstra.schema.json
```

### Godoc
Browse complete package documentation:
```bash
go doc github.com/primadi/lokstra
go doc github.com/primadi/lokstra/core/router
```

### Examples
Every section links to working examples in the main documentation.

---

## 🔗 Related Documentation

- **[Introduction](../00-introduction/)** - Getting started and overview
- **[Essentials](../01-essentials/)** - Core concepts and tutorials
- **[Examples](../00-introduction/examples/)** - Progressive learning examples
- **[Architecture](../00-introduction/architecture.md)** - System design and patterns

---

**Last Updated**: {% raw %}{{ date }}{% endraw %}  
**Lokstra Version**: {% raw %}{{ version }}{% endraw %}
