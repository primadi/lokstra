---
layout: docs
title: Router Registration
---

# Router Registration

> Router registration patterns, auto-router generation, and YAML-based router configuration

## Overview

Lokstra provides flexible router registration through factory functions and auto-router generation from service definitions. Both manually registered routers and auto-generated routers can be configured and overridden via YAML deployment configuration.

**Key Features:**
- ✅ Manual router registration via `RegisterRouter()`
- ✅ Auto-router generation from service definitions
- ✅ YAML-based configuration for both router types
- ✅ Router-level and route-level overrides
- ✅ Environment-specific middleware injection
- ✅ Path prefix and route customization

## Import Path

```go
import (
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/core/router"
    "github.com/primadi/lokstra/core/router/autogen"
    "github.com/primadi/lokstra/core/router/convention"
    "github.com/primadi/lokstra/core/route"
)
```

---

## Router Registration

### RegisterRouter
Registers a router instance in the runtime registry.

> **NEW:** Manual routers can now be configured via YAML `router-definitions` for middleware, path-prefix, and route-level overrides.

**Signature:**
```go
func RegisterRouter(name string, r router.Router)
```

**Example:**
```go
// Register manual router with named routes
adminRouter := router.New("")
adminRouter.GET("/dashboard", handlers.ShowDashboard, 
    route.WithNameOption("showDashboard"))
adminRouter.GET("/users", handlers.ListUsers)  // Auto-named: "GET_/users"
adminRouter.POST("/users", handlers.CreateUser)

lokstra_registry.RegisterRouter("admin-router", adminRouter)

// YAML can now override this router:
// router-definitions:
//   admin-router:
//     path-prefix: /api/v1/admin
//     middlewares: [admin-auth, audit-log]
//     custom:
//       - name: showDashboard
//         method: POST
//         path: /admin/main
```

**Use Cases:**
- Manual router registration with custom logic
- Admin panels and specialized endpoints
- Non-REST routing patterns
- Can be overridden from YAML per deployment

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
1. **YAML config (router-definitions)** - Highest priority (runtime override)
2. **RegisterServiceType options** - Medium priority (framework defaults)
3. **Auto-generate from service name** - Lowest priority (fallback)

**Example:**
```go
// Framework creates auto-router from service definition
router, err := lokstra_registry.BuildRouterFromDefinition("user-service-router")
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
Define routers in YAML configuration with inline overrides.

**Router Naming Convention:**
- Format: `{service-name}-router`
- Service name is derived by removing the `-router` suffix
- Examples:
  - `user-service-router` → service: `user-service`
  - `order-service-router` → service: `order-service`

**Auto-Generated Router Example:**
```yaml
router-definitions:
  user-service-router:  # Service name derived: "user-service"
    convention: rest
    resource: user
    resource-plural: users
    # Inline overrides
    path-prefix: /api/v1
    middlewares:
      - auth
      - logger
    hidden:
      - Delete
      - InternalHelper
    custom:
      - name: Login
        method: POST
        path: /auth/login
        middlewares:
          - rate-limiter
      - name: Logout
        method: POST
        path: /auth/logout
```

**Manual Router Override Example:**
```yaml
# Code: Manual router already registered
# r := router.New("")
# r.GET("/dashboard", handler.ShowDashboard, route.WithNameOption("showDashboard"))
# lokstra_registry.RegisterRouter("admin-router", r)

router-definitions:
  admin-router:  # Manual router (not auto-generated)
    # Override configuration from YAML
    path-prefix: /api/v1/admin
    middlewares:
      - admin-auth
      - audit-log
    # Route-level overrides
    custom:
      - name: showDashboard
        method: POST  # Change from GET to POST
        path: /admin/main  # Change path
        middlewares:
          - extra-logging
```

**Use Cases:**
- **Auto-generated routers:** Configure convention, resource, and customize routes
- **Manual routers:** Apply environment-specific middlewares and path prefixes
- **Both types:** Support route-level method/path/middleware overrides

---

## Manual Router Overrides

### Overview
Manual routers registered via `RegisterRouter()` can now be configured from YAML deployment files. This enables environment-specific configuration without code changes.

**Supported Overrides:**
- ✅ `path-prefix` - Change router base path
- ✅ `middlewares` - Add router-level middlewares
- ✅ `custom` routes - Update individual route method, path, or middlewares

**Not Supported:**
- ❌ `convention` - Manual routers don't use conventions
- ❌ `resource`/`resource-plural` - Manual routers don't use resource names
- ❌ `hidden` - Manual routers control visibility in code

---

### Route Naming for Overrides

To override specific routes, they must have names. Routes are named either:

**1. Manual Names (Recommended):**
```go
r.GET("/dashboard", handler, route.WithNameOption("showDashboard"))
r.POST("/export", handler, route.WithNameOption("exportData"))
```

**2. Auto-Generated Names:**
```go
r.GET("/users", handler)     // Name: "GET_/users"
r.POST("/orders", handler)   // Name: "POST_/orders"
r.PUT("/items/:id", handler) // Name: "PUT_/items/:id"
```

**Best Practice:** Use manual names for routes you plan to override from YAML.

---

### Router-Level Overrides

Apply configuration to the entire router.

**Code:**
```go
adminRouter := router.New("")
adminRouter.GET("/dashboard", handlers.Dashboard)
adminRouter.GET("/users", handlers.Users)
adminRouter.POST("/settings", handlers.Settings)

lokstra_registry.RegisterRouter("admin-router", adminRouter)
```

**YAML (Development):**
```yaml
router-definitions:
  admin-router:
    path-prefix: /admin
    middlewares:
      - logger
```

**YAML (Production):**
```yaml
router-definitions:
  admin-router:
    path-prefix: /api/v1/admin
    middlewares:
      - admin-auth
      - audit-log
      - logger
```

**Result:**
- Dev: Routes at `/admin/*` with logger only
- Prod: Routes at `/api/v1/admin/*` with auth, audit, logger

---

### Route-Level Overrides

Override specific routes within a manual router.

**Code:**
```go
apiRouter := router.New("")
apiRouter.GET("/status", handlers.Status, route.WithNameOption("status"))
apiRouter.POST("/webhook", handlers.Webhook, route.WithNameOption("webhook"))
apiRouter.GET("/metrics", handlers.Metrics, route.WithNameOption("metrics"))

lokstra_registry.RegisterRouter("api-router", apiRouter)
```

**YAML:**
```yaml
router-definitions:
  api-router:
    path-prefix: /api/v1
    middlewares: [logger]
    
    custom:
      # Disable webhook in staging (change to invalid path)
      - name: webhook
        path: /disabled
      
      # Add rate limiting to metrics
      - name: metrics
        middlewares:
          - rate-limiter
      
      # Change status to POST and add auth
      - name: status
        method: POST
        path: /health-check
        middlewares:
          - admin-auth
```

**Result:**
```
GET    /api/v1/status → POST /api/v1/health-check (with admin-auth + logger)
POST   /api/v1/webhook → POST /api/v1/disabled (effectively disabled)
GET    /api/v1/metrics (with rate-limiter + logger)
```

---

### Use Cases

**1. Environment-Specific Middleware:**
```yaml
# Production: Full security
router-definitions:
  admin-router:
    middlewares: [admin-auth, audit-log, rate-limiter]

# Development: No auth for easier testing
router-definitions:
  admin-router:
    middlewares: [logger]
```

**2. API Versioning:**
```yaml
# v1 deployment
router-definitions:
  api-router:
    path-prefix: /api/v1

# v2 deployment (same code, different path)
router-definitions:
  api-router:
    path-prefix: /api/v2
```

**3. Feature Flags via Path:**
```yaml
# Enable feature
router-definitions:
  feature-router:
    custom:
      - name: newFeature
        path: /features/new

# Disable feature (point to 404)
router-definitions:
  feature-router:
    custom:
      - name: newFeature
        path: /disabled/new
```

**4. Route-Specific Security:**
```yaml
router-definitions:
  api-router:
    custom:
      # Public route - no auth
      - name: publicEndpoint
        middlewares: [rate-limiter]
      
      # Protected route - full auth
      - name: adminEndpoint
        middlewares: [admin-auth, audit-log]
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
    
    if err := app.Run(30 * time.Second); err != nil {
        fmt.Println("Error starting server:", err)
    }
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
    
    if err := app.Run(30 * time.Second); err != nil {
        fmt.Println("Error starting server:", err)
    }
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

- **[lokstra_registry](./lokstra_registry)** - Registry API
- **[Service Registration](./service-registration)** - Service patterns
- **[Router](../01-core-packages/router)** - Router interface
- **[Auto-Router](../08-advanced/auto-router)** - Auto-router internals

---

## Related Guides

- **[Router Essentials](../../01-essentials/01-router/)** - Router basics
- **[Auto-Router Guide](../../02-deep-dive/auto-router/)** - Auto-router patterns
- **[API Design](../../04-guides/api-design/)** - API best practices
