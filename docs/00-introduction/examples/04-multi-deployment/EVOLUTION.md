# Evolution: Manual → Automated Patterns

> **This document explains the evolution from manual patterns (shown in this example) to automated patterns (covered in later chapters).**

---

## 🎯 Why Manual First?

This example deliberately uses **manual patterns** to teach fundamentals:

### Educational Benefits:
1. ✅ **Understand what happens under the hood**
2. ✅ **See the mechanics of service-to-router conversion**
3. ✅ **Learn how proxy communication works**
4. ✅ **Appreciate the automation later**
5. ✅ **Ability to debug and customize**

### Production Reality:
- Manual patterns are verbose but transparent
- Automated patterns are concise but require understanding
- **Learn manual → Use automated → Customize when needed**

---

## 📊 Pattern Evolution Comparison

### 1. Handler Creation

#### Manual Approach (This Example):
```go
// Create handler manually for each service method
func listUsersHandler(ctx *request.Context) error {
    users, err := userService.MustGet().List(&appservice.ListUsersParams{})
    if err != nil {
        return ctx.Api.Error(500, "INTERNAL_ERROR", err.Error())
    }
    return ctx.Api.Ok(users)
}

func getUserHandler(ctx *request.Context) error {
    var params appservice.GetUserParams
    if err := ctx.Req.BindPath(&params); err != nil {
        return ctx.Api.BadRequest("INVALID_ID", "Invalid user ID")
    }
    
    user, err := userService.MustGet().GetByID(&params)
    if err != nil {
        return ctx.Api.Error(404, "NOT_FOUND", err.Error())
    }
    return ctx.Api.Ok(user)
}

// Register manually
r.GET("/users", listUsersHandler)
r.GET("/users/{id}", getUserHandler)
```

**Pros**: Full control, explicit, easy to understand  
**Cons**: Verbose, repetitive, error-prone

#### Automated Approach (Future):
```go
// Define convention and overrides once
restConvention := conventions.RESTful

// WithOverrides supports multiple use cases:
routeOverrides := router.WithOverrides(map[string]any{
    // 1. Custom route mapping
    "GetByID": "GET /users/{id}",
    "List":    "GET /users",
    
    // 2. Hide methods (don't expose as routes)
    "Delete": router.Hidden,  // or nil, or "-"
    
    // 3. Add middleware to specific methods
    "GetByID": router.Override{
        Route: "GET /users/{id}",
        Middleware: []string{"SlowLogger", "Cache"},
    },
    
    // 4. Complete custom configuration
    "Update": router.Override{
        Route: "PUT /users/{id}",
        Middleware: []string{"Auth", "RateLimiter"},
        Hidden: false,
    },
})

// Auto-generate router from service
userRouter := router.NewFromService("users", restConvention, routeOverrides)

// Generated routes:
// ✅ GetByID() → GET /users/{id} (with SlowLogger, Cache middleware)
// ✅ List() → GET /users
// ✅ Create() → POST /users
// ✅ Update() → PUT /users/{id} (with Auth, RateLimiter middleware)
// ❌ Delete() → Hidden, not exposed
```

**Pros**: Concise, DRY, consistent, less errors, fine-grained control  
**Cons**: Need to understand conventions and override options

---

### 2. Proxy Service Communication

#### Manual Approach (This Example):
```go
type UserServiceRemote struct {
    proxy *proxy.Router
}

func (u *UserServiceRemote) GetByID(p *GetUserParams) (*User, error) {
    var JsonWrapper struct {
        Status string `json:"status"`
        Data   *User  `json:"data"`
    }
    
    // Manual path construction
    // Manual JSON handling
    err := u.proxy.DoJSON("GET", fmt.Sprintf("/users/%d", p.ID), nil, nil, &JsonWrapper)
    if err != nil {
        return nil, proxy.ParseRouterError(err)
    }
    return JsonWrapper.Data, nil
}

func (u *UserServiceRemote) List(p *ListUsersParams) ([]*User, error) {
    var JsonWrapper struct {
        Status string  `json:"status"`
        Data   []*User `json:"data"`
    }
    
    // Manual path construction
    // Manual JSON handling
    err := u.proxy.DoJSON("GET", "/users", nil, nil, &JsonWrapper)
    if err != nil {
        return nil, proxy.ParseRouterError(err)
    }
    return JsonWrapper.Data, nil
}

func NewUserServiceRemote() *UserServiceRemote {
    return &UserServiceRemote{
        proxy: proxy.NewRemoteRouter("http://localhost:3004"),
    }
}
```

**Pros**: Full control over HTTP calls, explicit  
**Cons**: Repetitive, error-prone, manual path management

#### Automated Approach (Future):
```go
type UserServiceRemote struct {
    proxy *proxy.Service
}

// Auto-proxy with same convention as router
func NewUserServiceRemote() *UserServiceRemote {
    return &UserServiceRemote{
        proxy: proxy.NewService("users", "http://localhost:3004",
            restConvention,  // Same convention as router!
            routeOverrides,  // Same overrides as router!
        ),
    }
}

// Methods auto-mapped to endpoints!
func (u *UserServiceRemote) GetByID(p *GetUserParams) (*User, error) {
    return u.proxy.Call("GetByID", p)  // Auto-maps to GET /users/{id}
}

func (u *UserServiceRemote) List(p *ListUsersParams) ([]*User, error) {
    return u.proxy.Call("List", p)  // Auto-maps to GET /users
}
```

**Pros**: DRY, consistent with router, automatic path mapping  
**Cons**: Need to understand conventions

**Key Insight**: Convention defined ONCE, used in BOTH router AND proxy!

```go
// Define once
restConvention := conventions.RESTful
routeOverrides := router.WithOverrides(...)

// Use in router
userRouter := router.NewFromService("users", restConvention, routeOverrides)

// Use in proxy (same convention!)
userProxy := proxy.NewService("users", url, restConvention, routeOverrides)

// Communication guaranteed to work!
```

---

### 3. Service Registration

#### Manual Approach (This Example):
```go
// Monolith - register local implementation
func registerMonolithServices() {
    lokstra_registry.RegisterServiceType("dbFactory", appservice.NewDatabase)
    lokstra_registry.RegisterServiceType("usersFactory", appservice.NewUserService)
    lokstra_registry.RegisterServiceType("ordersFactory", appservice.NewOrderService)
    
    lokstra_registry.RegisterLazyService("db", "dbFactory", nil)
    lokstra_registry.RegisterLazyService("users", "usersFactory", nil)
    lokstra_registry.RegisterLazyService("orders", "ordersFactory", nil)
}

// User-service - register local implementation
func registerUserServices() {
    lokstra_registry.RegisterServiceType("dbFactory", appservice.NewDatabase)
    lokstra_registry.RegisterServiceType("usersFactory", appservice.NewUserService)
    
    lokstra_registry.RegisterLazyService("db", "dbFactory", nil)
    lokstra_registry.RegisterLazyService("users", "usersFactory", nil)
}

// Order-service - register remote user service
func registerOrderServices() {
    lokstra_registry.RegisterServiceType("dbFactory", appservice.NewDatabase)
    lokstra_registry.RegisterServiceType("ordersFactory", appservice.NewOrderService)
    lokstra_registry.RegisterServiceTypeRemote("usersFactory", appservice.NewUserServiceRemote)
    
    lokstra_registry.RegisterLazyService("db", "dbFactory", nil)
    lokstra_registry.RegisterLazyService("orders", "ordersFactory", nil)
    lokstra_registry.RegisterLazyService("users", "usersFactory", nil)
}
```

**Pros**: Explicit control over what's registered  
**Cons**: Repetitive, need separate function per deployment

#### Automated Approach (Future - YAML):
```yaml
# deployment.yaml

services:
  # Define all services with local AND remote factories
  db:
    type: local
    factory: appservice.NewDatabase
  
  users:
    type: local
    factory: appservice.NewUserService
    remote:
      factory: appservice.NewUserServiceRemote
      base-url: http://user-service:3004
  
  orders:
    type: local
    factory: appservice.NewOrderService

# Define deployments
deployments:
  monolith:
    servers:
      - name: monolith
        port: 3003
        services:
          - db: local
          - users: local
          - orders: local
  
  microservices:
    servers:
      - name: user-service
        port: 3004
        services:
          - db: local
          - users: local
      
      - name: order-service
        port: 3005
        services:
          - db: local
          - users: remote  # Auto-uses remote factory!
          - orders: local
```

**Pros**: Declarative, no code duplication, easy to modify  
**Cons**: Need to understand config schema

#### Automated Approach (Future - Code):
```go
// Define all services ONCE with both local and remote
func defineServices() *lokstra.ServiceDefinitions {
    defs := lokstra.NewServiceDefinitions()
    
    defs.Define("db",
        service.WithLocalFactory(appservice.NewDatabase))
    
    defs.Define("users",
        service.WithLocalFactory(appservice.NewUserService),
        service.WithRemoteFactory(appservice.NewUserServiceRemote))
    
    defs.Define("orders",
        service.WithLocalFactory(appservice.NewOrderService))
    
    return defs
}

// Define deployments
func defineDeployments() *lokstra.Deployments {
    deps := lokstra.NewDeployments()
    
    // Monolith deployment
    deps.Add("monolith",
        deployment.WithServer("monolith", ":3003",
            server.UseLocal("db", "users", "orders")))
    
    // Microservices deployment
    deps.Add("microservices",
        deployment.WithServer("user-service", ":3004",
            server.UseLocal("db", "users")),
        deployment.WithServer("order-service", ":3005",
            server.UseLocal("db", "orders"),
            server.UseRemote("users")))
    
    return deps
}

// Run deployment
func main() {
    app := lokstra.NewApp()
    app.RegisterServices(defineServices())
    app.RegisterDeployments(defineDeployments())
    app.RunDeployment(flag.Lookup("deployment").Value.String())
}
```

**Pros**: Type-safe, refactorable, IDE support  
**Cons**: More code than YAML

---

## 🎓 Learning Path

### Step 1: Master Manual (This Example) ✅
**What you learn**:
- Service interface abstraction
- Handler creation from service methods
- Proxy communication mechanics
- Deployment-specific registration

**Time**: 2-3 hours

### Step 2: Understand Conventions (01-Essentials)
**What you learn**:
- Convention system design
- RESTful convention
- RPC convention
- Custom conventions

**Time**: 3-4 hours

### Step 3: Auto Service-to-Router (01-Essentials)
**What you learn**:
- `router.NewFromService()`
- Convention rules
- Route overrides
- Handler generation

**Time**: 2-3 hours

### Step 4: Auto Proxy Service (01-Essentials)
**What you learn**:
- `proxy.NewService()`
- Method-to-endpoint mapping
- Convention reuse
- Error handling

**Time**: 2-3 hours

### Step 5: Config-Driven Deployment (02-Advanced)
**What you learn**:
- YAML deployment config
- Code-based deployment config
- Service definition system
- Multi-environment setup

**Time**: 4-5 hours

---

## � WithOverrides - Powerful Customization

The `WithOverrides` option provides fine-grained control over route generation:

### 1. Simple Route Mapping
```go
router.WithOverrides(map[string]any{
    "GetByID": "GET /users/{id}",
    "List":    "GET /users",
})
```

### 2. Hide Methods
Don't expose certain methods as routes:
```go
router.WithOverrides(map[string]any{
    "Delete":         router.Hidden,  // Option 1
    "SoftDelete":     nil,            // Option 2
    "InternalMethod": "-",            // Option 3
})
```

**Use cases**:
- Internal methods not for external API
- Admin-only methods (exposed separately)
- Deprecated methods

### 3. Add Method-Specific Middleware
```go
router.WithOverrides(map[string]any{
    "GetByID": router.Override{
        Route:      "GET /users/{id}",
        Middleware: []string{"Cache", "SlowLogger"},
    },
    
    "Update": router.Override{
        Route:      "PUT /users/{id}",
        Middleware: []string{"Auth", "RateLimiter", "AuditLog"},
    },
    
    "List": router.Override{
        Route:      "GET /users",
        Middleware: []string{"Cache"},
    },
})
```

**Use cases**:
- Add caching to read operations
- Add auth to write operations
- Add rate limiting to expensive operations
- Add audit logging to sensitive operations

### 4. Combined Configuration
```go
router.WithOverrides(map[string]any{
    // Custom route + middleware
    "GetByID": router.Override{
        Route:      "GET /users/{id}",
        Middleware: []string{"Cache", "SlowLogger"},
    },
    
    // Hide method
    "Delete": router.Hidden,
    
    // Simple route mapping
    "List": "GET /users",
    
    // Full control
    "Create": router.Override{
        Route:      "POST /users",
        Middleware: []string{"Auth", "Validator", "RateLimiter"},
        Hidden:     false,
    },
})
```

### 5. Real-World Example
```go
// Public API - hide sensitive methods, add caching
publicRouter := router.NewFromService("users", 
    conventions.RESTful,
    router.WithOverrides(map[string]any{
        // Read operations - add cache
        "GetByID": router.Override{
            Route:      "GET /users/{id}",
            Middleware: []string{"Cache", "RateLimiter"},
        },
        "List": router.Override{
            Route:      "GET /users",
            Middleware: []string{"Cache"},
        },
        
        // Write operations - hide from public
        "Create":  router.Hidden,
        "Update":  router.Hidden,
        "Delete":  router.Hidden,
    }),
)

// Admin API - expose all methods with auth
adminRouter := router.NewFromService("users",
    conventions.RESTful,
    router.WithPrefix("/admin"),
    router.WithOverrides(map[string]any{
        // All methods require auth + audit
        "Create": router.Override{
            Route:      "POST /users",
            Middleware: []string{"AdminAuth", "AuditLog"},
        },
        "Update": router.Override{
            Route:      "PUT /users/{id}",
            Middleware: []string{"AdminAuth", "AuditLog"},
        },
        "Delete": router.Override{
            Route:      "DELETE /users/{id}",
            Middleware: []string{"AdminAuth", "AuditLog", "SoftDelete"},
        },
    }),
)
```

### 6. Override Structure
```go
type Override struct {
    Route      string   // Custom route pattern
    Middleware []string // Method-specific middleware
    Hidden     bool     // Hide from router
}

// Convenience values
const (
    Hidden = "-"  // Shorthand for hiding
)
```

### 7. Proxy Service with Same Overrides
**Critical**: Proxy must use the SAME overrides as router!

```go
// Define once
restConvention := conventions.RESTful
overrides := router.WithOverrides(map[string]any{
    "GetByID": "GET /users/{id}",
    "Delete":  router.Hidden,  // Hidden in router
})

// Router - creates routes
userRouter := router.NewFromService("users", restConvention, overrides)

// Proxy - must use SAME overrides
userProxy := proxy.NewService("users", "http://localhost:3004",
    restConvention,
    overrides,  // SAME overrides!
)

// Why? So proxy knows:
// - GetByID maps to GET /users/{id}
// - Delete is hidden (don't try to call it)
```

---

##  Key Insights

### 1. WithOverrides is Your Power Tool
The same overrides used in router MUST be used in proxy for fine-grained control:
```go
overrides := router.WithOverrides(
    "List":    {Path: "/", Method: "GET"},
    "GetByID": {Path: "/{id}", Method: "GET"},
    "Delete":  router.Hidden,  // Hide from API!
    "Update": {
        Path: "/{id}",
        Method: "PUT",
        Middleware: []router.Middleware{SlowLogger},  // Add per-method middleware!
    },
)

// Router uses overrides
router := router.NewFromService("users", conventions.RESTful, overrides)

// Proxy uses SAME overrides
proxy := proxy.NewService("users", url, conventions.RESTful, overrides)

// Both stay in sync: Delete is hidden, Update has SlowLogger
```

### 2. Convention is King
The same convention used in router MUST be used in proxy:
```go
convention := conventions.RESTful
overrides := router.WithOverrides(...)

// Router uses convention
router := router.NewFromService("users", convention, overrides)

// Proxy uses SAME convention
proxy := proxy.NewService("users", url, convention, overrides)

// Communication works automatically!
```

### 3. Define Once, Use Everywhere
In manual approach:
- Define routes multiple times
- Define proxy paths multiple times
- Easy to get out of sync

In automated approach:
- Define convention once
- Define overrides once (hide methods, add middleware)
- Router generates routes
- Proxy uses same convention + overrides
- Always in sync!

### 4. Trade-offs
**Manual**:
- ➕ Explicit, transparent, full control
- ➕ Easy to debug and customize
- ➖ Verbose, repetitive, error-prone

**Automated**:
- ➕ Concise, DRY, consistent
- ➕ Less errors, easier maintenance
- ➕ WithOverrides for fine-grained control (hide methods, add middleware)
- ➖ Need to understand conventions

### 5. Progressive Enhancement
You don't have to choose one or the other:
```go
// Auto-generate most routes
router := router.NewFromService("users", conventions.RESTful)

// Manually add special routes
router.POST("/users/bulk-import", bulkImportHandler)
router.GET("/users/export", exportHandler)
```

Best of both worlds!

---

## 🚀 Ready for Advanced Patterns?

Once you understand this manual example, proceed to:

### 01-Essentials
- Convention-based routing
- Auto service-to-router
- Auto proxy services
- Advanced DI patterns

### 02-Advanced
- Config-driven deployment
- Custom conventions
- Multi-environment setup
- Production patterns

### 03-Production
- Service discovery
- Observability
- Resilience patterns
- Scaling strategies

---

**Remember**: Manual patterns are not "wrong" - they're the foundation. Automated patterns build on this foundation!
