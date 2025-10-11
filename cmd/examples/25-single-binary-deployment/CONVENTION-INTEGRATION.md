# Service Convention Integration

This document explains how the service convention system is integrated into Example 25.

## Overview

The service convention system automatically converts services into routers and client routers using configurable conventions. This eliminates boilerplate and ensures consistency across your application.

## How It Works

### 1. Service Definition

```go
// services/user_service.go
type UserService interface {
    GetUser(ctx *request.Context, req GetUserRequest) (GetUserResponse, error)
    ListUsers(ctx *request.Context, req ListUsersRequest) (ListUsersResponse, error)
    CreateUser(ctx *request.Context, req CreateUserRequest) (CreateUserResponse, error)
    UpdateUser(ctx *request.Context, req UpdateUserRequest) (UpdateUserResponse, error)
    DeleteUser(ctx *request.Context, req DeleteUserRequest) (DeleteUserResponse, error)
}
```

### 2. Convention Application (Automatic)

When you register a service with `lokstra_registry.RegisterService()`, the REST convention automatically applies:

```go
// In main.go or service factory
lokstra_registry.RegisterService("user-service", NewUserService)
```

The REST convention generates these routes:

| Service Method | HTTP Method | Path | Description |
|---------------|-------------|------|-------------|
| `GetUser` | GET | `/api/v1/users/{id}` | Get single user |
| `ListUsers` | GET | `/api/v1/users` | List all users |
| `CreateUser` | POST | `/api/v1/users` | Create new user |
| `UpdateUser` | PUT | `/api/v1/users/{id}` | Update existing user |
| `DeleteUser` | DELETE | `/api/v1/users/{id}` | Delete user |

### 3. Remote Service Client (Automatic)

The same convention is used on the client side:

```go
// services/user_service.go - Remote implementation
type RemoteUserService struct {
    client  *api_client.Client
    options router.ServiceRouterOptions
}

func (s *RemoteUserService) GetUser(ctx *request.Context, req GetUserRequest) (GetUserResponse, error) {
    // Convention automatically maps to: GET /api/v1/users/{id}
    return CallTyped[GetUserResponse](s.client, "GetUser", req)
}
```

## Configuration

### Default Configuration (REST)

By default, all services use the REST convention:

```yaml
# config-monolith.yaml
services:
  user-service:
    enabled: true
    lazy: false  # Local service
  
  auth-service:
    enabled: true
    lazy: false  # Local service
```

### Custom Convention

You can specify a different convention per service:

```go
// In service factory
func NewUserService(cfg *config.Config, lazy bool) interface{} {
    options := router.DefaultServiceRouterOptions().
        WithConvention("rpc").  // Use RPC instead of REST
        WithPrefix("/rpc/v1")
    
    // ...
}
```

### Override Specific Routes

For edge cases, override individual routes:

```go
options := router.DefaultServiceRouterOptions().
    WithPrefix("/api/v1").
    WithRouteOverride("Login", router.RouteMeta{
        HTTPMethod:   "POST",
        Path:         "/auth/login",
        AuthRequired: false,
    })
```

## Benefits in Example 25

### Before Convention System

**Manual Router Registration:**
```go
// Had to manually register each route
router.GET("/api/v1/users/:id", getUserHandler)
router.GET("/api/v1/users", listUsersHandler)
router.POST("/api/v1/users", createUserHandler)
router.PUT("/api/v1/users/:id", updateUserHandler)
router.DELETE("/api/v1/users/:id", deleteUserHandler)

// Had to manually implement each client method
func (c *UserClient) GetUser(id string) (*User, error) {
    return c.client.Get("/api/v1/users/" + id)
}
```

**Problems:**
- ❌ Lots of boilerplate
- ❌ Easy to make mistakes (typos, inconsistent paths)
- ❌ Server and client routes can drift
- ❌ Hard to maintain consistency across services

### After Convention System

**Automatic Generation:**
```go
// Server: Just define the service interface
type UserService interface {
    GetUser(ctx *request.Context, req GetUserRequest) (GetUserResponse, error)
    // ... other methods
}

// Client: One-line remote methods
func (s *RemoteUserService) GetUser(ctx *request.Context, req GetUserRequest) (GetUserResponse, error) {
    return CallTyped[GetUserResponse](s.client, "GetUser", req)
}
```

**Benefits:**
- ✅ Zero boilerplate
- ✅ Automatic consistency
- ✅ Server and client always in sync
- ✅ Easy to maintain
- ✅ Type-safe

## Deployment Modes

The convention system works seamlessly across all deployment modes:

### 1. Monolith Deployment

All services are local, convention generates routes:

```
┌─────────────────┐
│  Single Binary  │
│                 │
│  GET /users/:id │──┐
│  POST /users    │  │ All routes auto-generated
│  GET /orders/:id│  │ by REST convention
│  POST /orders   │──┘
└─────────────────┘
```

### 2. Multiport Deployment

Services on different ports, each with consistent routes:

```
Port 8080               Port 8081
┌─────────────┐        ┌─────────────┐
│ User Service│        │Order Service│
│             │        │             │
│ GET /users  │        │ GET /orders │
│ POST /users │        │ POST /orders│
└─────────────┘        └─────────────┘
        │                      │
        └──────────┬───────────┘
                   │
              Same REST
             Convention
```

### 3. Microservices Deployment

Each service is a separate binary, convention ensures consistency:

```
user-service:8080      auth-service:8081    order-service:8082
┌────────────┐         ┌────────────┐       ┌────────────┐
│GET /users  │         │POST /login │       │GET /orders │
│POST /users │         │POST /logout│       │POST /orders│
└────────────┘         └────────────┘       └────────────┘
     │                       │                     │
     └───────────────────────┴─────────────────────┘
              All use REST convention
             (can be overridden per service)
```

## Advanced Usage

### Creating a Custom Convention for Your Organization

```go
// conventions/company_convention.go
package conventions

type CompanyAPIConvention struct{}

func (c *CompanyAPIConvention) Name() string {
    return "company-api"
}

func (c *CompanyAPIConvention) GenerateRoutes(serviceType reflect.Type, options router.ServiceRouterOptions) (map[string]router.RouteMeta, error) {
    // Your company's specific API conventions
    // Example: All actions use POST, path includes version and action
    // POST /v1/user/get
    // POST /v1/user/list
    // POST /v1/user/create
}

func init() {
    lokstra_registry.MustRegisterConvention(&CompanyAPIConvention{})
}
```

### Using It in Your Services

```go
import _ "myapp/conventions"

options := router.DefaultServiceRouterOptions().
    WithConvention("company-api")

// All your services now follow your company's convention!
```

## Testing

Convention system makes testing easier:

```go
// Test that convention generates expected routes
func TestUserServiceConvention(t *testing.T) {
    convention, _ := lokstra_registry.GetConvention("rest")
    
    options := router.DefaultServiceRouterOptions().
        WithPrefix("/api/v1")
    
    routes, _ := convention.GenerateRoutes(
        reflect.TypeOf((*UserService)(nil)).Elem(),
        options,
    )
    
    // Assert expected routes
    assert.Equal(t, "GET", routes["GetUser"].HTTPMethod)
    assert.Equal(t, "/api/v1/users/{id}", routes["GetUser"].Path)
}
```

## Summary

The service convention system in Example 25:

1. **Eliminates boilerplate** - No manual route registration
2. **Ensures consistency** - All services follow the same pattern
3. **Supports all deployment modes** - Works with monolith, multiport, and microservices
4. **Is extensible** - Create custom conventions for your needs
5. **Is bidirectional** - Same convention for server and client
6. **Is type-safe** - Based on Go interfaces
7. **Supports overrides** - Handle edge cases when needed

This makes the framework more powerful while reducing the amount of code you need to write!
