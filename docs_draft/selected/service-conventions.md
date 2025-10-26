# Service Convention System

## Overview

The Service Convention System provides a pluggable and extensible way to automatically convert Go service interfaces into HTTP routers and client routers. This allows you to:

1. **Auto-generate HTTP routes** from service methods
2. **Auto-generate client router methods** using the same convention
3. **Create custom conventions** for different API styles (REST, RPC, GraphQL, etc.)
4. **Override individual routes** for edge cases while keeping convention benefits

## Package Location

The convention system is located in `core/router` package:
- `core/router/service_convention.go` - Convention interface and registry
- `core/router/convention_rest.go` - Default REST convention implementation

This placement avoids circular dependencies and makes it easy to use throughout the framework.

## Architecture

```
Service Interface
       ↓
Convention (Registry)
       ↓
    ┌──────┴──────┐
    ↓             ↓
Router         ClientRouter
(Server)       (Client)
```

## Built-in Conventions

### REST Convention (Default)

The REST convention maps service methods to RESTful HTTP endpoints:

| Method Pattern | HTTP Method | Path | Example |
|---------------|-------------|------|---------|
| `Get{Resource}` | GET | `/{resources}/{id}` | `GetUser` → `GET /users/{id}` |
| `List{Resource}s` | GET | `/{resources}` | `ListUsers` → `GET /users` |
| `Create{Resource}` | POST | `/{resources}` | `CreateUser` → `POST /users` |
| `Update{Resource}` | PUT | `/{resources}/{id}` | `UpdateUser` → `PUT /users/{id}` |
| `Delete{Resource}` | DELETE | `/{resources}/{id}` | `DeleteUser` → `DELETE /users/{id}` |
| `Patch{Resource}` | PATCH | `/{resources}/{id}` | `PatchUser` → `PATCH /users/{id}` |
| Other methods | POST | `/{resources}/{method-name}` | `ResetPassword` → `POST /users/reset-password` |

## Basic Usage

### 1. Define Your Service

```go
type UserService interface {
    GetUser(ctx *request.Context, req GetUserRequest) (GetUserResponse, error)
    ListUsers(ctx *request.Context, req ListUsersRequest) (ListUsersResponse, error)
    CreateUser(ctx *request.Context, req CreateUserRequest) (CreateUserResponse, error)
    UpdateUser(ctx *request.Context, req UpdateUserRequest) (UpdateUserResponse, error)
    DeleteUser(ctx *request.Context, req DeleteUserRequest) (DeleteUserResponse, error)
}
```

### 2. Auto-generate Router (Server Side)

```go
// Using default REST convention
options := router.DefaultServiceRouterOptions().
    WithPrefix("/api/v1")

// Convention automatically generates:
// GET    /api/v1/users/{id}      -> GetUser
// GET    /api/v1/users           -> ListUsers
// POST   /api/v1/users           -> CreateUser
// PUT    /api/v1/users/{id}      -> UpdateUser
// DELETE /api/v1/users/{id}      -> DeleteUser
```

### 3. Auto-generate Client Router (Client Side)

```go
// Client uses the same convention to make HTTP calls
client := api_client.NewClient("http://localhost:8080")

// Convention automatically maps to:
// client.Get("/api/v1/users/{id}") for GetUser
// client.Get("/api/v1/users") for ListUsers
// client.Post("/api/v1/users", body) for CreateUser
// etc.
```

## Advanced Usage

### Using Different Conventions

```go
// Use a different convention
options := router.DefaultServiceRouterOptions().
    WithConvention("rpc").  // Use RPC convention instead of REST
    WithPrefix("/rpc/v1")
```

### Override Specific Routes

```go
// Use convention for most routes, override for edge cases
options := router.DefaultServiceRouterOptions().
    WithPrefix("/api/v1").
    WithRouteOverride("ResetPassword", router.RouteMeta{
        HTTPMethod: "POST",
        Path:       "/auth/reset-password",  // Custom path
        AuthRequired: false,
    })
```

### Disable Conventions (Manual Routing)

```go
// Disable conventions and specify all routes manually
options := router.DefaultServiceRouterOptions().
    WithoutConventions().
    WithRouteOverride("GetUser", router.RouteMeta{...}).
    WithRouteOverride("CreateUser", router.RouteMeta{...})
```

### Custom Resource Names

```go
// Override auto-detected resource names
options := router.DefaultServiceRouterOptions().
    WithResourceName("user").           // Singular: "user"
    WithPluralResourceName("people")    // Plural: "people" instead of "users"

// Results in:
// GET /people/{id}      -> GetUser
// GET /people           -> ListUsers
// POST /people          -> CreateUser
```

## Creating Custom Conventions

You can create your own conventions for different API styles (RPC, GraphQL, etc.).

### Example: Simple RPC Convention

```go
package myapp

import (
    "reflect"
    "strings"
    "github.com/primadi/lokstra/core/router"
    "github.com/primadi/lokstra/lokstra_registry"
)

type RPCConvention struct{}

func (c *RPCConvention) Name() string {
    return "rpc"
}

func (c *RPCConvention) GenerateRoutes(serviceType reflect.Type, options router.ServiceRouterOptions) (map[string]router.RouteMeta, error) {
    routes := make(map[string]router.RouteMeta)
    
    serviceName := extractServiceName(serviceType.Name())
    
    for i := 0; i < serviceType.NumMethod(); i++ {
        method := serviceType.Method(i)
        methodName := method.Name
        
        // RPC style: POST /rpc/{ServiceName}.{MethodName}
        routes[methodName] = router.RouteMeta{
            MethodName: methodName,
            HTTPMethod: "POST",
            Path:       strings.ToLower(serviceName) + "." + methodName,
        }
    }
    
    return routes, nil
}

func (c *RPCConvention) GenerateClientMethod(method router.ServiceMethodInfo, options router.ServiceRouterOptions) (router.ClientMethodMeta, error) {
    serviceName := options.ResourceName
    
    return router.ClientMethodMeta{
        MethodName: method.Name,
        HTTPMethod: "POST",
        Path:       serviceName + "." + method.Name,
        HasBody:    true,
    }, nil
}

func extractServiceName(typeName string) string {
    name := strings.TrimSuffix(typeName, "Service")
    return strings.ToLower(name)
}

// Register during init
func init() {
    router.MustRegisterConvention(&RPCConvention{})
}
```

### Using Your Custom Convention

```go
import _ "myapp/conventions"  // Register RPC convention

func main() {
    // Use your custom RPC convention
    options := router.DefaultServiceRouterOptions().
        WithConvention("rpc").
        WithPrefix("/rpc/v1")
    
    // Results in:
    // POST /rpc/v1/user.GetUser
    // POST /rpc/v1/user.ListUsers
    // POST /rpc/v1/user.CreateUser
}
```

## Convention Registry API

### Register a Convention

```go
// Register during init (panics on error)
func init() {
    router.MustRegisterConvention(&MyConvention{})
}

// Or register at runtime (returns error)
err := router.RegisterConvention(&MyConvention{})
if err != nil {
    log.Fatal(err)
}
```

### Get a Convention

```go
// Get specific convention
convention, err := router.GetConvention("rest")
if err != nil {
    log.Fatal(err)
}

// Get default convention
convention, err := router.GetDefaultConvention()
```

### Set Default Convention

```go
// Change the default convention
err := router.SetDefaultConvention("rpc")
if err != nil {
    log.Fatal(err)
}
```

### List Available Conventions

```go
conventions := router.ListConventions()
fmt.Println("Available conventions:", conventions)
// Output: Available conventions: [rest rpc graphql]
```

## Convention Interface

```go
type ServiceConvention interface {
    // Name returns the convention name (e.g., "rest", "rpc")
    Name() string

    // GenerateRoutes generates route metadata from service methods
    // Used for server-side router generation
    GenerateRoutes(serviceType reflect.Type, options router.ServiceRouterOptions) (map[string]router.RouteMeta, error)

    // GenerateClientMethod generates client method metadata
    // Used for client-side router generation
    GenerateClientMethod(method router.ServiceMethodInfo, options router.ServiceRouterOptions) (ClientMethodMeta, error)
}
```

## Best Practices

### 1. Use Conventions for Consistency
```go
// ✅ Good: Convention ensures all services follow same pattern
options := router.DefaultServiceRouterOptions().
    WithPrefix("/api/v1")
```

### 2. Override Only When Necessary
```go
// ✅ Good: Use convention by default, override edge cases
options := router.DefaultServiceRouterOptions().
    WithPrefix("/api/v1").
    WithRouteOverride("Login", router.RouteMeta{
        Path: "/auth/login",  // Special auth endpoint
    })
```

### 3. Name Service Methods Consistently
```go
// ✅ Good: Convention-friendly names
type UserService interface {
    GetUser(...)      // GET /users/{id}
    ListUsers(...)    // GET /users
    CreateUser(...)   // POST /users
}

// ❌ Bad: Inconsistent names break conventions
type UserService interface {
    FetchUserById(...)     // Doesn't match convention
    AllUsers(...)          // Doesn't match convention
    AddNewUser(...)        // Doesn't match convention
}
```

### 4. Use Resource Names for Multi-Service Apps
```go
// ✅ Good: Explicit resource names avoid conflicts
userOptions := router.DefaultServiceRouterOptions().
    WithPrefix("/api/v1").
    WithResourceName("user")

orderOptions := router.DefaultServiceRouterOptions().
    WithPrefix("/api/v1").
    WithResourceName("order")
```

## Migration Guide

### From Manual Routes to Conventions

**Before:**
```go
// Manual route registration
router.GET("/api/v1/users/:id", getUserHandler)
router.GET("/api/v1/users", listUsersHandler)
router.POST("/api/v1/users", createUserHandler)
router.PUT("/api/v1/users/:id", updateUserHandler)
router.DELETE("/api/v1/users/:id", deleteUserHandler)
```

**After:**
```go
// Auto-generated from service + convention
options := router.DefaultServiceRouterOptions().
    WithPrefix("/api/v1")
// All routes generated automatically!
```

## Future Conventions

The convention system is designed to be extensible. Future built-in conventions may include:

- **GraphQL Convention**: Map services to GraphQL queries/mutations
- **gRPC Convention**: Map services to gRPC methods
- **WebSocket Convention**: Map services to WebSocket message handlers
- **Event Convention**: Map services to event handlers
- **CLI Convention**: Map services to CLI commands

## Summary

The Service Convention System provides:

✅ **Automatic route generation** - No manual route registration
✅ **Consistency** - All services follow the same pattern
✅ **Flexibility** - Override for edge cases
✅ **Extensibility** - Create custom conventions
✅ **Bidirectional** - Same convention for server and client
✅ **Type-safe** - Based on Go interfaces

This eliminates boilerplate while maintaining flexibility for advanced use cases.
