# lokstra Package

> Main entry point for Lokstra framework

## Overview

The `lokstra` package is the main entry point for building Lokstra applications. It provides convenience functions and type aliases for creating routers, apps, and servers.

This package re-exports the most commonly used types and functions from sub-packages, giving you a clean import path for basic application development.

## Import Path

```go
import "github.com/primadi/lokstra"
```

---

## Type Aliases

### Server
```go
type Server = server.Server
```

Server container that manages one or more Apps. See [Server documentation](server) for details.

**Example:**
```go
server := lokstra.NewServer("my-server", app1, app2)
```

---

### App
```go
type App = app.App
```

HTTP listener that serves routers. See [App documentation](app) for details.

**Example:**
```go
app := lokstra.NewApp("api", ":8080", router)
```

---

### Router
```go
type Router = router.Router
```

HTTP router interface for registering routes and middleware. See [Router documentation](router) for details.

**Example:**
```go
router := lokstra.NewRouter("api")
router.GET("/users", getUsersHandler)
```

---

### RequestContext
```go
type RequestContext = request.Context
```

Request context passed to handlers and middleware. See [Request Context documentation](request) for details.

**Example:**
```go
func handler(c *lokstra.RequestContext) error {
    userID := c.Req.Param("id")
    return c.Api.Success(user)
}
```

---

### HandlerFunc
```go
type HandlerFunc = request.HandlerFunc
```

Handler function signature for middleware and interceptors.

**Signature:**
```go
type HandlerFunc func(*Context) error
```

**Example:**
```go
func loggingMiddleware(c *lokstra.RequestContext) error {
    log.Printf("Request: %s %s", c.R.Method, c.R.URL.Path)
    return c.Next()
}
```

---

### Handler
```go
type Handler = request.Handler
```

Interface for HTTP handlers that can be registered with Router.

---

## Functions

### NewRouter
Creates a new Router instance with default engine.

**Signature:**
```go
func NewRouter(name string) Router
```

**Parameters:**
- `name` - Router identifier (used for debugging and registration)

**Returns:**
- `Router` - New router instance

**Example:**
```go
api := lokstra.NewRouter("api-v1")
api.GET("/health", healthCheckHandler)
api.POST("/users", createUserHandler)

app := lokstra.NewApp("api", ":8080", api)
```

**See Also:**
- [Router documentation](router) for detailed router API
- [NewRouterWithEngine](#newrouterwithengine) for custom engine

---

### NewRouterWithEngine
Creates a new Router instance with specific engine type.

**Signature:**
```go
func NewRouterWithEngine(name string, engineType string) Router
```

**Parameters:**
- `name` - Router identifier
- `engineType` - Engine type ("default", "servemux", etc.)

**Returns:**
- `Router` - New router instance with specified engine

**Example:**
```go
// Use Go's standard ServeMux engine
router := lokstra.NewRouterWithEngine("api", "servemux")

// Use default Lokstra engine
router := lokstra.NewRouterWithEngine("api", "default")
```

**Engine Types:**
- `"default"` - Lokstra's default router engine (fast, flexible)
- `"servemux"` - Go's standard `http.ServeMux` (compatible with stdlib)

**Notes:**
- Default engine supports all Lokstra features (middleware, auto-binding, etc.)
- ServeMux engine has limited features but maximum stdlib compatibility

---

### NewApp
Creates a new App instance with given routers.

**Signature:**
```go
func NewApp(name string, addr string, routers ...Router) *App
```

**Parameters:**
- `name` - App identifier (used for logging and management)
- `addr` - Listen address (e.g., ":8080", "127.0.0.1:8080", "unix:/tmp/app.sock")
- `routers` - One or more routers to mount in the app

**Returns:**
- `*App` - New app instance

**Example:**
```go
// Single router
apiRouter := lokstra.NewRouter("api")
app := lokstra.NewApp("my-app", ":8080", apiRouter)

// Multiple routers (chained automatically)
apiV1 := lokstra.NewRouter("api-v1")
apiV2 := lokstra.NewRouter("api-v2")
app := lokstra.NewApp("my-app", ":8080", apiV1, apiV2)

// TCP listener
app := lokstra.NewApp("api", ":8080", router)

// Unix socket
app := lokstra.NewApp("api", "unix:/tmp/api.sock", router)

// Specific interface
app := lokstra.NewApp("api", "127.0.0.1:8080", router)
```

**Address Formats:**
- `":8080"` - Listen on all interfaces, port 8080
- `"127.0.0.1:8080"` - Listen on localhost only
- `"192.168.1.100:8080"` - Listen on specific IP
- `"unix:/tmp/api.sock"` - Unix domain socket

**See Also:**
- [App documentation](app) for detailed app API
- [NewAppWithConfig](#newappwithconfig) for custom configuration

---

### NewAppWithConfig
Creates a new App instance with custom listener configuration.

**Signature:**
```go
func NewAppWithConfig(name string, addr string, listenerType string, 
    config map[string]any, routers ...Router) *App
```

**Parameters:**
- `name` - App identifier
- `addr` - Listen address
- `listenerType` - Listener type ("default", "tls", "h2c", etc.)
- `config` - Listener configuration map
- `routers` - Routers to mount

**Returns:**
- `*App` - New app instance with custom configuration

**Example:**
```go
// HTTPS with TLS
tlsConfig := map[string]any{
    "cert-file": "/path/to/cert.pem",
    "key-file":  "/path/to/key.pem",
}
app := lokstra.NewAppWithConfig("api", ":443", "tls", tlsConfig, router)

// HTTP/2 Cleartext (h2c)
h2cConfig := map[string]any{
    "read-timeout":  "30s",
    "write-timeout": "30s",
}
app := lokstra.NewAppWithConfig("api", ":8080", "h2c", h2cConfig, router)

// Custom timeouts
config := map[string]any{
    "read-timeout":       "10s",
    "write-timeout":      "10s",
    "idle-timeout":       "60s",
    "read-header-timeout": "5s",
}
app := lokstra.NewAppWithConfig("api", ":8080", "default", config, router)
```

**Listener Types:**
- `"default"` - Standard HTTP listener
- `"tls"` - HTTPS listener (requires cert and key)
- `"h2c"` - HTTP/2 Cleartext
- Custom types can be registered via registry

**Configuration Keys:**
- `read-timeout` - Maximum duration for reading request (string duration)
- `write-timeout` - Maximum duration for writing response
- `idle-timeout` - Maximum idle time between requests
- `read-header-timeout` - Time to read request headers
- `cert-file` - TLS certificate file path (for "tls" type)
- `key-file` - TLS private key file path (for "tls" type)

---

### NewServer
Creates a new Server instance with given apps.

**Signature:**
```go
func NewServer(name string, apps ...*App) *Server
```

**Parameters:**
- `name` - Server identifier
- `apps` - One or more apps to manage

**Returns:**
- `*Server` - New server instance

**Example:**
```go
// Single app
app := lokstra.NewApp("api", ":8080", router)
server := lokstra.NewServer("my-server", app)

// Multiple apps on different ports
apiApp := lokstra.NewApp("api", ":8080", apiRouter)
adminApp := lokstra.NewApp("admin", ":9000", adminRouter)
server := lokstra.NewServer("my-server", apiApp, adminApp)

// Run server
if err := server.Run(30 * time.Second); err != nil {
    fmt.Println("Error starting server:", err)
}
```

**See Also:**
- [Server documentation](server) for lifecycle management

---

### FetchAndCast
Generic HTTP client function for making requests with automatic type casting.

**Signature:**
```go
func FetchAndCast[T any](client *api_client.ClientRouter, path string, 
    opts ...api_client.FetchOption) (T, error)
```

**Type Parameters:**
- `T` - Target type for response deserialization

**Parameters:**
- `client` - ClientRouter instance
- `path` - Request path (relative to client base URL)
- `opts` - Fetch options (method, headers, body, query, etc.)

**Returns:**
- `T` - Deserialized response data
- `error` - Error if request fails or deserialization fails

**Example:**
```go
import (
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/common/api_client"
)

client := api_client.NewClientRouter("https://api.example.com")

// GET request
user, err := lokstra.FetchAndCast[*User](client, "/users/123")
if err != nil {
    log.Fatal(err)
}

// With query parameters
users, err := lokstra.FetchAndCast[[]User](client, "/users",
    api_client.WithMethod("GET"),
    api_client.WithQuery(map[string]string{
        "status": "active",
        "limit":  "10",
    }),
)

// POST request
newUser := &User{Name: "John", Email: "john@example.com"}
created, err := lokstra.FetchAndCast[*User](client, "/users",
    api_client.WithMethod("POST"),
    api_client.WithBody(newUser),
)

// With headers
user, err := lokstra.FetchAndCast[*User](client, "/users/123",
    api_client.WithHeaders(map[string]string{
        "Authorization": "Bearer token123",
        "X-Request-ID":  "req-456",
    }),
)
```

**Supported Types:**
- Structs: `*User`, `*Order`
- Slices: `[]User`, `[]*Order`
- Maps: `map[string]any`
- Primitives: `string`, `int`, `bool`
- Any JSON-deserializable type

**See Also:**
- [API Client documentation](../04-client/api-client) for complete client API
- [FetchOption documentation](../04-client/api-client#fetchoption) for all options

---

## Complete Example

```go
package main

import (
    "time"
    "github.com/primadi/lokstra"
)

func main() {
    // Create router
    api := lokstra.NewRouter("api")
    
    // Register routes
    api.GET("/health", healthCheck)
    api.GET("/users/:id", getUser)
    api.POST("/users", createUser)
    api.PUT("/users/:id", updateUser)
    api.DELETE("/users/:id", deleteUser)
    
    // Add middleware
    api.Use(loggingMiddleware, authMiddleware)
    
    // Create app
    app := lokstra.NewApp("api", ":8080", api)
    
    // Create server
    server := lokstra.NewServer("my-server", app)
    
    // Run with 30s graceful shutdown
    if err := server.Run(30 * time.Second); err != nil {
        fmt.Println("Error starting server:", err)
    }    
}

func healthCheck(c *lokstra.RequestContext) error {
    return c.Api.Success(map[string]string{"status": "ok"})
}

func getUser(c *lokstra.RequestContext) error {
    id := c.Req.Param("id")
    // ... fetch user from database
    return c.Api.Success(user)
}

func createUser(c *lokstra.RequestContext, input *CreateUserInput) error {
    // Input automatically bound from request body
    // ... create user
    return c.Api.Created(user)
}

func updateUser(c *lokstra.RequestContext, input *UpdateUserInput) error {
    id := c.Req.Param("id")
    // ... update user
    return c.Api.Success(user)
}

func deleteUser(c *lokstra.RequestContext) error {
    id := c.Req.Param("id")
    // ... delete user
    return c.Api.NoContent()
}

func loggingMiddleware(c *lokstra.RequestContext) error {
    log.Printf("%s %s", c.R.Method, c.R.URL.Path)
    return c.Next()
}

func authMiddleware(c *lokstra.RequestContext) error {
    token := c.Req.Header("Authorization")
    if token == "" {
        return c.Api.Unauthorized("Missing authorization token")
    }
    // ... validate token
    return c.Next()
}
```

---

## See Also

- **[Router](router)** - Complete router API
- **[App](app)** - App lifecycle and configuration
- **[Server](server)** - Server management
- **[Request Context](request)** - Request handling
- **[Response](response)** - Response formatting
- **[API Client](../04-client/api-client)** - HTTP client

---

## Related Guides

- **[Quick Start](../../00-introduction/quick-start)** - Build your first app
- **[Router Essentials](../../01-router-guide/01-router/)** - Learn routing basics
- **[App & Server Guide](../../01-router-guide/05-app-and-server/)** - Lifecycle management
