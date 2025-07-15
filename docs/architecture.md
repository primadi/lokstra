# Lokstra Architecture

Lokstra follows a modular architecture designed for scalability and maintainability.

## Core Components

### 1. Global Context
The `GlobalContext` is the central registry for all components:
- Service factories and instances
- Middleware factories
- Named handlers
- Configuration

### 2. Server and Apps
- **Server**: Top-level container that manages multiple apps
- **App**: HTTP application that listens on a specific address
- Each app can have its own middleware stack and routes
- Apps can run on different ports for microservices architecture

### 3. Services
Services are reusable components that provide functionality:
- Database connections
- Cache systems
- Email sending
- Metrics collection
- Health checks

#### Service Interface
```go
type Service interface {
    InstanceName() string
    GetConfig(key string) any
}
```

#### Service Module Interface
```go
type ServiceModule interface {
    Name() string
    Factory(config any) (Service, error)
    Meta() *ServiceMeta
}
```

### 4. Middleware
Middleware processes requests and responses in a pipeline:
- Authentication
- Logging
- CORS
- Rate limiting
- Security headers

#### Middleware Interface
```go
type MiddlewareFunc = func(next HandlerFunc) HandlerFunc
type MiddlewareModule interface {
    Name() string
    Factory(config any) MiddlewareFunc
    Meta() *MiddlewareMeta
}
```

### 5. Request Context
The `Context` provides access to:
- HTTP request/response
- URL parameters
- Services
- Middleware data
- JSON binding/response helpers

## Module System

### Service Modules
Service modules encapsulate:
- Service implementation
- Configuration handling
- Metadata (description, tags)
- Factory function

Example:
```go
type RedisModule struct{}

func (r *RedisModule) Name() string {
    return "lokstra.redis"
}

func (r *RedisModule) Factory(config any) (iface.Service, error) {
    // Create and configure Redis service
}

func (r *RedisModule) Meta() *iface.ServiceMeta {
    return &iface.ServiceMeta{
        Description: "Redis connection pool service",
        Tags:        []string{"cache", "storage"},
    }
}
```

### Middleware Modules
Middleware modules provide:
- Middleware implementation
- Configuration handling
- Priority ordering
- Metadata

Example:
```go
type CORSModule struct{}

func (c *CORSModule) Name() string {
    return "lokstra.cors"
}

func (c *CORSModule) Factory(config any) MiddlewareFunc {
    return func(next HandlerFunc) HandlerFunc {
        return func(ctx *Context) error {
            // CORS logic
            return next(ctx)
        }
    }
}

func (c *CORSModule) Meta() *MiddlewareMeta {
    return &MiddlewareMeta{
        Priority:    50,
        Description: "CORS middleware for cross-origin requests",
        Tags:        []string{"cors", "security"},
    }
}
```

## Configuration System

### YAML Structure
Configuration is split into logical files:
- `server.yaml`: Server-level settings
- `apps_*.yaml`: Application definitions
- `services_*.yaml`: Service configurations
- `modules_*.yaml`: Module settings

### Configuration Loading
1. Load all YAML files from directory
2. Merge configurations
3. Resolve environment variables
4. Validate structure
5. Create server from configuration

### Environment Variable Override
Use `${VAR_NAME:default}` syntax in YAML:
```yaml
services:
  - type: lokstra.redis
    config:
      addr: ${REDIS_ADDR:localhost:6379}
      password: ${REDIS_PASSWORD:}
```

## Request Lifecycle

1. **Request Received**: HTTP server receives request
2. **Context Creation**: Create request context with services access
3. **Middleware Pipeline**: Execute middleware in priority order
4. **Route Matching**: Find matching route handler
5. **Handler Execution**: Execute route handler
6. **Response**: Send response back to client
7. **Cleanup**: Clean up resources

## Scalability Patterns

### Single Application
```go
app := lokstra.NewApp(ctx, "api", ":8080")
app.Start()
```

### Multiple Applications (Microservices)
```go
server := lokstra.NewServer(ctx, "my-server")
server.AddApp(userApp)    // :8080
server.AddApp(orderApp)   // :8081
server.AddApp(gatewayApp) // :80 (reverse proxy)
server.Start()
```

### Configuration-Driven Deployment
```yaml
apps:
  - name: user-service
    address: :8080
  - name: order-service
    address: :8081
  - name: gateway
    address: :80
    mount_reverse_proxy:
      - prefix: /users
        target: http://localhost:8080
      - prefix: /orders
        target: http://localhost:8081
```

## Design Principles

1. **Modularity**: Everything is a module that can be registered and configured
2. **Composability**: Mix and match services and middleware as needed
3. **Configuration**: Production deployments driven by YAML configuration
4. **Scalability**: Single binary can run as monolith or microservices
5. **Simplicity**: Start simple, add complexity as needed
6. **Extensibility**: Easy to add custom services and middleware
