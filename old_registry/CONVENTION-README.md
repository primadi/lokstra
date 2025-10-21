# Service Convention System

A pluggable, extensible system for automatically converting Go service interfaces into HTTP routers and client routers.

## Quick Start

### 1. Define Your Service

```go
type UserService interface {
    GetUser(ctx *request.Context, req GetUserRequest) (GetUserResponse, error)
    ListUsers(ctx *request.Context, req ListUsersRequest) (ListUsersResponse, error)
    CreateUser(ctx *request.Context, req CreateUserRequest) (CreateUserResponse, error)
}
```

### 2. Register Service (Auto-routes Generated)

```go
lokstra_registry.RegisterService("user-service", NewUserService)
```

### 3. Routes Automatically Created

```
GET    /api/v1/users/{id}    -> GetUser
GET    /api/v1/users         -> ListUsers
POST   /api/v1/users         -> CreateUser
```

That's it! No manual route registration needed.

## Features

- ✅ **Zero Boilerplate** - Routes generated automatically from service interface
- ✅ **Bidirectional** - Same convention for server (router) and client (client router)
- ✅ **Extensible** - Create custom conventions (REST, RPC, GraphQL, etc.)
- ✅ **Flexible** - Override specific routes for edge cases
- ✅ **Type-Safe** - Based on Go interfaces and reflection
- ✅ **Consistent** - All services follow the same pattern

## Built-in Conventions

### REST Convention (Default)

Maps service methods to RESTful endpoints:

| Method Pattern | HTTP | Path | Example |
|---------------|------|------|---------|
| `Get{Resource}` | GET | `/{resources}/{id}` | `GetUser` → `GET /users/{id}` |
| `List{Resource}s` | GET | `/{resources}` | `ListUsers` → `GET /users` |
| `Create{Resource}` | POST | `/{resources}` | `CreateUser` → `POST /users` |
| `Update{Resource}` | PUT | `/{resources}/{id}` | `UpdateUser` → `PUT /users/{id}` |
| `Delete{Resource}` | DELETE | `/{resources}/{id}` | `DeleteUser` → `DELETE /users/{id}` |

## Advanced Usage

### Use Different Convention

```go
options := router.DefaultServiceRouterOptions().
    WithConvention("rpc").  // Use RPC instead of REST
    WithPrefix("/rpc/v1")
```

### Override Specific Routes

```go
options := router.DefaultServiceRouterOptions().
    WithPrefix("/api/v1").
    WithRouteOverride("Login", router.RouteMeta{
        HTTPMethod: "POST",
        Path:       "/auth/login",
    })
```

### Custom Resource Names

```go
options := router.DefaultServiceRouterOptions().
    WithResourceName("user").
    WithPluralResourceName("people")  // Use "people" instead of "users"
```

## Create Custom Convention

```go
type MyConvention struct{}

func (c *MyConvention) Name() string {
    return "my-convention"
}

func (c *MyConvention) GenerateRoutes(serviceType reflect.Type, options router.ServiceRouterOptions) (map[string]router.RouteMeta, error) {
    // Your custom route generation logic
}

func (c *MyConvention) GenerateClientMethod(method router.ServiceMethodInfo, options router.ServiceRouterOptions) (ClientMethodMeta, error) {
    // Your custom client method generation logic
}

// Register during init
func init() {
    router.MustRegisterConvention(&MyConvention{})
}
```

## Documentation

- **[Service Conventions](docs/service-conventions.md)** - Complete guide to the convention system
- **[Convention Examples](docs/convention-examples.md)** - Code examples and use cases
- **[Convention Integration](cmd/examples/25-single-binary-deployment/CONVENTION-INTEGRATION.md)** - How it integrates with Example 25

## Registry API

```go
// List available conventions
conventions := router.ListConventions()

// Get specific convention
conv, err := router.GetConvention("rest")

// Get default convention
conv, err := router.GetDefaultConvention()

// Set default convention
err := router.SetDefaultConvention("rpc")

// Register new convention
err := router.RegisterConvention(&MyConvention{})
```

## Benefits

### Before

```go
// Manual route registration
router.GET("/api/v1/users/:id", getUserHandler)
router.GET("/api/v1/users", listUsersHandler)
router.POST("/api/v1/users", createUserHandler)
router.PUT("/api/v1/users/:id", updateUserHandler)
router.DELETE("/api/v1/users/:id", deleteUserHandler)

// Manual client implementation
func (c *UserClient) GetUser(id string) (*User, error) {
    return c.client.Get("/api/v1/users/" + id)
}
func (c *UserClient) ListUsers() ([]*User, error) {
    return c.client.Get("/api/v1/users")
}
// ... more boilerplate
```

### After

```go
// Just define the interface
type UserService interface {
    GetUser(ctx *request.Context, req GetUserRequest) (GetUserResponse, error)
    ListUsers(ctx *request.Context, req ListUsersRequest) (ListUsersResponse, error)
    CreateUser(ctx *request.Context, req CreateUserRequest) (CreateUserResponse, error)
}

// Remote client - one line per method
func (s *RemoteUserService) GetUser(ctx *request.Context, req GetUserRequest) (GetUserResponse, error) {
    return CallTyped[GetUserResponse](s.client, "GetUser", req)
}
```

## Architecture

```
Service Interface
       ↓
Convention (Registry)
    ┌──────┴──────┐
    ↓             ↓
  Router      ClientRouter
 (Server)      (Client)
```

The same convention is used to generate both server routes and client methods, ensuring they always stay in sync.

## Use Cases

- **Microservices** - Ensure all services follow consistent API patterns
- **Monoliths** - Reduce boilerplate in large applications
- **Multi-team** - Enforce organization-wide API standards
- **API Versioning** - Easily switch between different API versions
- **Prototyping** - Quickly scaffold services without writing route handlers

## Future Conventions

The system is designed to support various API styles:

- **RPC Convention** - Function-call style APIs
- **GraphQL Convention** - Map to GraphQL queries/mutations
- **gRPC Convention** - Map to gRPC service definitions
- **WebSocket Convention** - Real-time bidirectional communication
- **CLI Convention** - Generate command-line interfaces

## License

Part of the Lokstra framework.
