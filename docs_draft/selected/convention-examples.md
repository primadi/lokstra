# Service Convention System - Quick Start Examples

## Example 1: Basic Usage with Default REST Convention

```go
package main

import (
    "github.com/primadi/lokstra/core/router"
    "github.com/primadi/lokstra/lokstra_registry"
)

// 1. Define your service interface
type UserService interface {
    GetUser(ctx *request.Context, req GetUserRequest) (GetUserResponse, error)
    ListUsers(ctx *request.Context, req ListUsersRequest) (ListUsersResponse, error)
    CreateUser(ctx *request.Context, req CreateUserRequest) (CreateUserResponse, error)
    UpdateUser(ctx *request.Context, req UpdateUserRequest) (UpdateUserResponse, error)
    DeleteUser(ctx *request.Context, req DeleteUserRequest) (DeleteUserResponse, error)
}

// 2. Use convention to auto-generate routes
func main() {
    // Get the default REST convention
    convention, _ := router.GetDefaultConvention()
    
    // Configure options
    options := router.DefaultServiceRouterOptions().
        WithPrefix("/api/v1")
    
    // Generate routes from service
    routes, _ := convention.GenerateRoutes(
        reflect.TypeOf((*UserService)(nil)).Elem(),
        options,
    )
    
    // Routes automatically generated:
    // GET    /api/v1/users/{id}   -> GetUser
    // GET    /api/v1/users        -> ListUsers
    // POST   /api/v1/users        -> CreateUser
    // PUT    /api/v1/users/{id}   -> UpdateUser
    // DELETE /api/v1/users/{id}   -> DeleteUser
}
```

## Example 2: Using Different Conventions

```go
// Use RPC convention instead of REST
options := router.DefaultServiceRouterOptions().
    WithConvention("rpc").
    WithPrefix("/rpc/v1")

// Results in:
// POST /rpc/v1/user.GetUser
// POST /rpc/v1/user.ListUsers
// POST /rpc/v1/user.CreateUser
```

## Example 3: Override Specific Routes

```go
// Use convention but override special cases
options := router.DefaultServiceRouterOptions().
    WithPrefix("/api/v1").
    WithRouteOverride("Login", router.RouteMeta{
        HTTPMethod:   "POST",
        Path:         "/auth/login",
        AuthRequired: false,
    }).
    WithRouteOverride("RefreshToken", router.RouteMeta{
        HTTPMethod:   "POST",
        Path:         "/auth/refresh",
        AuthRequired: false,
    })

// Most routes follow convention:
// GET  /api/v1/users/{id}   -> GetUser
// POST /api/v1/users        -> CreateUser
//
// But special auth routes are overridden:
// POST /auth/login          -> Login
// POST /auth/refresh        -> RefreshToken
```

## Example 4: Custom Resource Names

```go
// Override auto-detected names
options := router.DefaultServiceRouterOptions().
    WithResourceName("user").
    WithPluralResourceName("people")  // Use "people" instead of "users"

// Results in:
// GET  /people/{id}    -> GetUser
// GET  /people         -> ListUsers
// POST /people         -> CreateUser
```

## Example 5: Multiple Services with Different Conventions

```go
// User service uses REST
userOptions := router.DefaultServiceRouterOptions().
    WithConvention("rest").
    WithPrefix("/api/v1")

// Analytics service uses RPC
analyticsOptions := router.DefaultServiceRouterOptions().
    WithConvention("rpc").
    WithPrefix("/rpc/v1")

// Each service can use its own convention!
```

## Example 6: Creating a Custom Convention

```go
package conventions

import (
    "reflect"
    "strings"
    "github.com/primadi/lokstra/core/router"
    "github.com/primadi/lokstra/lokstra_registry"
)

// Simple RPC-style convention
type SimpleRPCConvention struct{}

func (c *SimpleRPCConvention) Name() string {
    return "simple-rpc"
}

func (c *SimpleRPCConvention) GenerateRoutes(serviceType reflect.Type, options router.ServiceRouterOptions) (map[string]router.RouteMeta, error) {
    routes := make(map[string]router.RouteMeta)
    
    serviceName := strings.TrimSuffix(serviceType.Name(), "Service")
    serviceName = strings.ToLower(serviceName)
    
    // All methods use POST with path: /service/method
    for i := 0; i < serviceType.NumMethod(); i++ {
        method := serviceType.Method(i)
        methodName := method.Name
        
        routes[methodName] = router.RouteMeta{
            MethodName: methodName,
            HTTPMethod: "POST",
            Path:       "/" + serviceName + "/" + strings.ToLower(methodName),
        }
    }
    
    return routes, nil
}

func (c *SimpleRPCConvention) GenerateClientMethod(method router.ServiceMethodInfo, options router.ServiceRouterOptions) (router.ClientMethodMeta, error) {
    serviceName := options.ResourceName
    
    return router.ClientMethodMeta{
        MethodName: method.Name,
        HTTPMethod: "POST",
        Path:       "/" + serviceName + "/" + strings.ToLower(method.Name),
        HasBody:    true,
    }, nil
}

// Register during init
func init() {
    router.MustRegisterConvention(&SimpleRPCConvention{})
}
```

### Using the Custom Convention

```go
import _ "myapp/conventions"  // Register custom convention

func main() {
    options := router.DefaultServiceRouterOptions().
        WithConvention("simple-rpc")
    
    // Results in:
    // POST /user/getuser
    // POST /user/listusers
    // POST /user/createuser
}
```

## Example 7: Disable Conventions (Manual Mode)

```go
// Completely manual routing
options := router.DefaultServiceRouterOptions().
    WithoutConventions().  // Disable auto-generation
    WithRouteOverride("GetUser", router.RouteMeta{
        HTTPMethod: "GET",
        Path:       "/v1/fetch-user/{id}",
    }).
    WithRouteOverride("CreateUser", router.RouteMeta{
        HTTPMethod: "POST",
        Path:       "/v1/add-user",
    })

// Only explicitly defined routes are created
```

## Example 8: Convention Registry Operations

```go
package main

import (
    "fmt"
    "github.com/primadi/lokstra/lokstra_registry"
)

func main() {
    // List all available conventions
    conventions := router.ListConventions()
    fmt.Println("Available:", conventions)
    // Output: Available: [rest simple-rpc]
    
    // Get specific convention
    restConvention, err := router.GetConvention("rest")
    if err != nil {
        panic(err)
    }
    fmt.Println("Got convention:", restConvention.Name())
    
    // Change default convention
    err = router.SetDefaultConvention("simple-rpc")
    if err != nil {
        panic(err)
    }
    
    // Get the new default
    defaultConv, _ := router.GetDefaultConvention()
    fmt.Println("Default is now:", defaultConv.Name())
}
```

## Example 9: Integration with lokstra_registry Service Factory

```go
// In your service factory
func NewUserService(cfg *config.Config, lazy bool) interface{} {
    if lazy {
        // Remote service - use client with convention
        client := api_client.NewClient(cfg.GetString("user-service.url"))
        
        options := router.DefaultServiceRouterOptions().
            WithPrefix("/api/v1")
        
        return &RemoteUserService{
            client:  client,
            options: options,  // Convention applied here
        }
    }
    
    // Local service
    return &LocalUserService{
        db: cfg.MustGet("db").(*sql.DB),
    }
}
```

## Example 10: Full Service Setup (Monolith)

```go
package main

import (
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/core/router"
)

func main() {
    // Register services
    lokstra_registry.RegisterService("user-service", NewUserService)
    lokstra_registry.RegisterService("order-service", NewOrderService)
    lokstra_registry.RegisterService("payment-service", NewPaymentService)
    
    // Each service auto-generates its routes using REST convention
    userOpts := router.DefaultServiceRouterOptions().
        WithPrefix("/api/v1").
        WithResourceName("user")
    
    orderOpts := router.DefaultServiceRouterOptions().
        WithPrefix("/api/v1").
        WithResourceName("order")
    
    paymentOpts := router.DefaultServiceRouterOptions().
        WithPrefix("/api/v1").
        WithResourceName("payment")
    
    // All routes auto-generated:
    // GET    /api/v1/users/{id}
    // GET    /api/v1/orders/{id}
    // GET    /api/v1/payments/{id}
    // etc...
}
```

## Cheat Sheet

| Task | Code |
|------|------|
| Use default REST convention | `options := router.DefaultServiceRouterOptions()` |
| Use specific convention | `options.WithConvention("rpc")` |
| Set prefix | `options.WithPrefix("/api/v1")` |
| Override route | `options.WithRouteOverride("MethodName", meta)` |
| Custom resource name | `options.WithResourceName("user")` |
| Disable conventions | `options.WithoutConventions()` |
| List conventions | `router.ListConventions()` |
| Get convention | `router.GetConvention("rest")` |
| Register convention | `router.RegisterConvention(conv)` |
| Set default | `router.SetDefaultConvention("rest")` |

## Benefits Summary

✅ **No boilerplate** - Routes generated automatically
✅ **Consistent** - All services follow same pattern
✅ **Flexible** - Override when needed
✅ **Extensible** - Create custom conventions
✅ **Type-safe** - Based on Go interfaces
✅ **Bidirectional** - Same convention for server & client

Start with the default REST convention, override edge cases, and create custom conventions only when needed!
