# Router Registration

> Router registration patterns, auto-router generation, and router factories

## Overview

Lokstra provides flexible router registration through factory functions and auto-router generation from service definitions. This guide covers router registration patterns, auto-router generation from services, and router factory patterns.

## Import Path

```go
import (
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/core/router"
    "github.com/primadi/lokstra/core/router/autogen"
    "github.com/primadi/lokstra/core/router/convention"
)
```

---

## Router Registration

### RegisterRouter
Registers a router instance in the runtime registry.

**Signature:**
```go
func RegisterRouter(name string, r router.Router)
```

**Example:**
```go
userRouter := lokstra.NewRouter()
userRouter.GET("/users", handlers.GetUsers)
userRouter.POST("/users", handlers.CreateUser)
userRouter.GET("/users/:id", handlers.GetUser)

lokstra_registry.RegisterRouter("user-router", userRouter)
```

**Use Cases:**
- Manual router registration
- Custom routing logic
- Pre-built router instances

---

### GetRouter
Retrieves a registered router instance.

**Signature:**
```go
func GetRouter(name string) router.Router
```

**Returns:**
- Router instance, or `nil` if not found

**Example:**
```go
userRouter := lokstra_registry.GetRouter("user-router")
if userRouter != nil {
    userRouter.GET("/users/export", handlers.ExportUsers)
}
```

---

### GetAllRouters
Returns all registered routers.

**Signature:**
```go
func GetAllRouters() map[string]router.Router
```

**Returns:**
- Map of router name to router instance

**Example:**
```go
routers := lokstra_registry.GetAllRouters()
for name, router := range routers {
    fmt.Printf("Router: %s\n", name)
    router.PrintRoutes()
}
```

---

## Auto-Router Generation

### Overview
Auto-routers are automatically generated from service types with resource metadata. This eliminates boilerplate routing code and ensures consistent RESTful APIs.

**Benefits:**
- ✅ No manual routing code
- ✅ Consistent API patterns
- ✅ Convention-based routing
- ✅ Automatic CRUD operations
- ✅ Customizable via metadata

---

### Service Registration with Auto-Router

**Pattern:**
```go
lokstra_registry.RegisterServiceType(serviceType, local, remote,
    deploy.WithResource(singular, plural),
    deploy.WithConvention(convention),
    deploy.WithPathPrefix(prefix),
    deploy.WithMiddleware(names...),
    deploy.WithRouteOverride(methodName, pathSpec),
    deploy.WithHiddenMethods(methods...),
)
```

**Example:**
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

func (s *UserService) Update(ctx *request.Context) error {
    id := ctx.Req.PathParam("id")
    var user User
    ctx.Req.BindJSON(&user)
    s.db.Get().Update(id, &user)
    return ctx.Api.Ok(user)
}

func (s *UserService) Delete(ctx *request.Context) error {
    id := ctx.Req.PathParam("id")
    s.db.Get().Delete(id)
    return ctx.Api.Ok(nil)
}

// Register with auto-router
lokstra_registry.RegisterServiceType("user-service",
    userFactory,
    nil,
    deploy.WithResource("user", "users"),
    deploy.WithConvention("rest"),
)
```

**Generated Routes:**
```
GET    /users           -> List()
POST   /users           -> Create()
GET    /users/:id       -> Get()
PUT    /users/:id       -> Update()
DELETE /users/:id       -> Delete()
```

---

### BuildRouterFromDefinition
Creates a router instance from a router definition (used internally by framework).

**Signature:**
```go
func BuildRouterFromDefinition(routerName string) (router.Router, error)
```

**Metadata Resolution Priority:**
1. **YAML config (router-overrides)** - Highest priority (runtime override)
2. **RegisterServiceType options** - Medium priority (framework defaults)
3. **Auto-generate from service name** - Lowest priority (fallback)

**Example:**
```go
// Framework creates auto-router from service definition
router, err := lokstra_registry.BuildRouterFromDefinition("user-router")
if err != nil {
    log.Fatal(err)
}
```

---

## Routing Conventions

### REST Convention (Default)
Standard RESTful routing pattern.

**Method Mapping:**
| Service Method | HTTP Method | Path           | Description          |
|----------------|-------------|----------------|----------------------|
| `List`         | GET         | `/resources`   | List all resources   |
| `Create`       | POST        | `/resources`   | Create new resource  |
| `Get`          | GET         | `/resources/:id` | Get single resource |
| `Update`       | PUT         | `/resources/:id` | Update resource     |
| `Delete`       | DELETE      | `/resources/:id` | Delete resource     |

**Example:**
```go
deploy.WithResource("user", "users")
deploy.WithConvention("rest")

// Generated:
// GET    /users
// POST   /users
// GET    /users/:id
// PUT    /users/:id
// DELETE /users/:id
```

---

### RPC Convention
Remote procedure call style routing.

**Method Mapping:**
| Service Method | HTTP Method | Path                | Description      |
|----------------|-------------|---------------------|------------------|
| `GetUsers`     | GET         | `/GetUsers`         | RPC-style call   |
| `CreateUser`   | POST        | `/CreateUser`       | RPC-style call   |
| Custom methods | Auto-detect | `/MethodName`       | Based on prefix  |

**Example:**
```go
deploy.WithResource("user", "users")
deploy.WithConvention("rpc")

// Generated:
// GET    /GetUsers
// POST   /CreateUser
// GET    /GetUser
// PUT    /UpdateUser
// DELETE /DeleteUser
```

---

### Custom Conventions
Register custom routing conventions.

**Example:**
```go
// Register custom convention
convention.Register("graphql", myGraphQLConvention)

// Use in service
deploy.WithConvention("graphql")
```

---

## Route Customization

### WithPathPrefix
Sets path prefix for all routes.

**Example:**
```go
lokstra_registry.RegisterServiceType("user-service",
    userFactory,
    nil,
    deploy.WithResource("user", "users"),
    deploy.WithPathPrefix("/api/v1"),
)

// Generated:
// GET    /api/v1/users
// POST   /api/v1/users
// GET    /api/v1/users/:id
```

---

### WithRouteOverride
Customizes path for specific methods.

**Signature:**
```go
deploy.WithRouteOverride(methodName, pathSpec string)
```

**Path Spec Formats:**
```go
"/custom/path"              // Auto-detect HTTP method from method name
"POST /custom/path"         // Explicit HTTP method
"GET /users/{id}/orders"    // With path parameters
```

**Example:**
```go
lokstra_registry.RegisterServiceType("user-service",
    userFactory,
    nil,
    deploy.WithResource("user", "users"),
    deploy.WithRouteOverride("Login", "POST /auth/login"),
    deploy.WithRouteOverride("Logout", "POST /auth/logout"),
    deploy.WithRouteOverride("ChangePassword", "/users/:id/password"),
)

// Generated:
// POST /auth/login               -> Login()
// POST /auth/logout              -> Logout()
// PUT  /users/:id/password       -> ChangePassword()
// GET  /users                    -> List()
// POST /users                    -> Create()
// GET  /users/:id                -> Get()
```

---

### WithHiddenMethods
Excludes methods from auto-router generation.

**Example:**
```go
type UserService struct {
    db *service.Cached[*DBService]
}

func (s *UserService) List(ctx *request.Context) error { /* ... */ }
func (s *UserService) Create(ctx *request.Context) error { /* ... */ }
func (s *UserService) Delete(ctx *request.Context) error { /* ... */ }
func (s *UserService) InternalHelper() { /* ... */ }
func (s *UserService) ValidateUser(user *User) error { /* ... */ }

lokstra_registry.RegisterServiceType("user-service",
    userFactory,
    nil,
    deploy.WithResource("user", "users"),
    deploy.WithHiddenMethods("Delete", "InternalHelper", "ValidateUser"),
)

// Generated (only List and Create):
// GET  /users  -> List()
// POST /users  -> Create()
// (Delete, InternalHelper, ValidateUser are hidden)
```

---

### WithMiddleware
Attaches middleware to all service routes.

**Example:**
```go
lokstra_registry.RegisterServiceType("user-service",
    userFactory,
    nil,
    deploy.WithResource("user", "users"),
    deploy.WithMiddleware("auth", "logger", "rate-limiter"),
)

// All routes have: auth -> logger -> rate-limiter -> handler
```

---

## YAML-Based Router Configuration

### Router Definitions
Define routers in YAML configuration.

**Example:**
```yaml
# Auto-router from service
router-definitions:
  user-router:
    service: user-service
    convention: rest
    resource: user
    resource-plural: users
    overrides: user-router-overrides

# Router overrides
router-overrides:
  user-router-overrides:
    path-prefix: /api/v1
    hidden:
      - Delete
      - InternalHelper
    custom:
      - name: Login
        method: POST
        path: /auth/login
      - name: Logout
        method: POST
        path: /auth/logout
```

---

## Complete Examples

### Basic Auto-Router
```go
package main

import (
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/core/deploy"
    "github.com/primadi/lokstra/lokstra_registry"
)

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
    // Register service with auto-router
    lokstra_registry.RegisterServiceType("user-service",
        func(deps, cfg map[string]any) any {
            return &UserService{
                db: deps["db"].(*service.Cached[*DBService]),
            }
        },
        nil,
        deploy.WithResource("user", "users"),
        deploy.WithDependencies("db"),
    )
    
    // Define service instance
    lokstra_registry.DefineService(&schema.ServiceDef{
        Name:      "user-svc",
        Type:      "user-service",
        DependsOn: []string{"db"},
    })
    
    // Router auto-generated and registered
    // GET    /users
    // POST   /users
    // GET    /users/:id
}
```

---

### Custom Routes with Overrides
```go
type UserService struct {
    db *service.Cached[*DBService]
}

// Standard CRUD
func (s *UserService) List(ctx *request.Context) error { /* ... */ }
func (s *UserService) Get(ctx *request.Context) error { /* ... */ }
func (s *UserService) Create(ctx *request.Context) error { /* ... */ }

// Custom endpoints
func (s *UserService) Login(ctx *request.Context) error {
    var creds Credentials
    ctx.Req.BindJSON(&creds)
    token, err := authenticate(creds)
    if err != nil {
        return ctx.Api.Unauthorized("Invalid credentials")
    }
    return ctx.Api.Ok(map[string]string{"token": token})
}

func (s *UserService) Logout(ctx *request.Context) error {
    token := ctx.Req.HeaderParam("Authorization")
    invalidateToken(token)
    return ctx.Api.Ok(nil)
}

func (s *UserService) ChangePassword(ctx *request.Context) error {
    id := ctx.Req.PathParam("id")
    var data PasswordChange
    ctx.Req.BindJSON(&data)
    s.db.Get().UpdatePassword(id, data.NewPassword)
    return ctx.Api.Ok(nil)
}

// Internal methods (not exposed)
func (s *UserService) validateEmail(email string) bool { /* ... */ }
func (s *UserService) hashPassword(password string) string { /* ... */ }

func main() {
    lokstra_registry.RegisterServiceType("user-service",
        userFactory,
        nil,
        deploy.WithResource("user", "users"),
        deploy.WithPathPrefix("/api/v1"),
        deploy.WithRouteOverride("Login", "POST /auth/login"),
        deploy.WithRouteOverride("Logout", "POST /auth/logout"),
        deploy.WithRouteOverride("ChangePassword", "PUT /users/:id/password"),
        deploy.WithHiddenMethods("validateEmail", "hashPassword"),
        deploy.WithMiddleware("logger"),
    )
}

// Generated routes:
// GET    /api/v1/users              -> List()
// POST   /api/v1/users              -> Create()
// GET    /api/v1/users/:id          -> Get()
// POST   /api/v1/auth/login         -> Login()
// POST   /api/v1/auth/logout        -> Logout()
// PUT    /api/v1/users/:id/password -> ChangePassword()
```

---

### Multi-Service API
```go
func main() {
    // User service
    lokstra_registry.RegisterServiceType("user-service",
        userFactory,
        nil,
        deploy.WithResource("user", "users"),
        deploy.WithPathPrefix("/api/v1"),
        deploy.WithMiddleware("auth", "logger"),
    )
    
    // Order service
    lokstra_registry.RegisterServiceType("order-service",
        orderFactory,
        nil,
        deploy.WithResource("order", "orders"),
        deploy.WithPathPrefix("/api/v1"),
        deploy.WithMiddleware("auth", "logger"),
        deploy.WithDependencies("userService"),
    )
    
    // Product service
    lokstra_registry.RegisterServiceType("product-service",
        productFactory,
        nil,
        deploy.WithResource("product", "products"),
        deploy.WithPathPrefix("/api/v1"),
        deploy.WithMiddleware("logger"),
    )
    
    // App with all routers
    app := lokstra.NewApp("api", ":8080")
    app.AddRouter(lokstra_registry.GetRouter("user-router"))
    app.AddRouter(lokstra_registry.GetRouter("order-router"))
    app.AddRouter(lokstra_registry.GetRouter("product-router"))
    
    app.Run()
}

// Generated API:
// User endpoints:
//   GET/POST/GET/PUT/DELETE /api/v1/users[/:id]
//
// Order endpoints:
//   GET/POST/GET/PUT/DELETE /api/v1/orders[/:id]
//
// Product endpoints:
//   GET/POST/GET/PUT/DELETE /api/v1/products[/:id]
```

---

### Versioned API with Router Groups
```go
func main() {
    // V1 services
    lokstra_registry.RegisterServiceType("user-service-v1",
        userFactoryV1,
        nil,
        deploy.WithResource("user", "users"),
        deploy.WithPathPrefix("/api/v1"),
    )
    
    // V2 services (breaking changes)
    lokstra_registry.RegisterServiceType("user-service-v2",
        userFactoryV2,
        nil,
        deploy.WithResource("user", "users"),
        deploy.WithPathPrefix("/api/v2"),
    )
    
    app := lokstra.NewApp("api", ":8080")
    app.AddRouter(lokstra_registry.GetRouter("user-router-v1"))
    app.AddRouter(lokstra_registry.GetRouter("user-router-v2"))
    
    app.Run()
}

// Generated API:
// V1: GET/POST/GET/PUT/DELETE /api/v1/users[/:id]
// V2: GET/POST/GET/PUT/DELETE /api/v2/users[/:id]
```

---

## Best Practices

### 1. Use Auto-Router for Standard CRUD
```go
// ✅ Good: Auto-router for standard operations
lokstra_registry.RegisterServiceType("user-service",
    userFactory,
    nil,
    deploy.WithResource("user", "users"),
)

// 🚫 Avoid: Manual routing for standard CRUD
router := lokstra.NewRouter()
router.GET("/users", handlers.List)
router.POST("/users", handlers.Create)
router.GET("/users/:id", handlers.Get)
// ...repetitive code
```

---

### 2. Override Only When Necessary
```go
// ✅ Good: Override only non-standard routes
deploy.WithRouteOverride("Login", "POST /auth/login")

// 🚫 Avoid: Overriding standard CRUD routes
deploy.WithRouteOverride("List", "GET /users")  // Unnecessary
deploy.WithRouteOverride("Create", "POST /users") // Unnecessary
```

---

### 3. Use Path Prefix for API Versioning
```go
// ✅ Good: Clear versioning
deploy.WithPathPrefix("/api/v1")
deploy.WithPathPrefix("/api/v2")

// 🚫 Avoid: Version in resource name
deploy.WithResource("user-v1", "users-v1")
```

---

### 4. Hide Internal Methods
```go
// ✅ Good: Hide non-endpoint methods
deploy.WithHiddenMethods("validateEmail", "hashPassword", "sendEmail")

// 🚫 Avoid: Exposing internal helpers
// (No hidden methods = all public methods become endpoints)
```

---

### 5. Apply Middleware at Service Level
```go
// ✅ Good: Service-level middleware
deploy.WithMiddleware("auth", "logger", "rate-limiter")

// 🚫 Avoid: Middleware on every route manually
router.GET("/users", auth, logger, rateLimiter, handler)
router.POST("/users", auth, logger, rateLimiter, handler)
// ...repetitive
```

---

## See Also

- **[lokstra_registry](./lokstra_registry.md)** - Registry API
- **[Service Registration](./service-registration.md)** - Service patterns
- **[Router](../01-core-packages/router.md)** - Router interface
- **[Auto-Router](../08-advanced/auto-router.md)** - Auto-router internals

---

## Related Guides

- **[Router Essentials](../../01-essentials/01-router/)** - Router basics
- **[Auto-Router Guide](../../02-deep-dive/auto-router/)** - Auto-router patterns
- **[API Design](../../04-guides/api-design/)** - API best practices
