# App

> HTTP listener and application lifecycle management

## Overview

The App type represents an HTTP listener that serves one or more routers. It manages the lifecycle of HTTP servers, including graceful shutdown, router chaining, and reverse proxy configuration.

An App is a container for routers and handles the actual HTTP listening and serving. Multiple Apps can run on different ports within a single Server.

## Import Path

```go
import "github.com/primadi/lokstra/core/app"

// Or use the main package
import "github.com/primadi/lokstra"
app := lokstra.NewApp("api", ":8080", router)
```

---

## Type Definition

```go
type App struct {
    // Unexported fields
}
```

---

## Constructor Functions

### New
Creates a new App with default listener configuration.

**Signature:**
```go
func New(name string, addr string, routers ...router.Router) *App
```

**Parameters:**
- `name` - App identifier (used for logging)
- `addr` - Listen address
- `routers` - Zero or more routers to mount

**Returns:**
- `*App` - New app instance

**Example:**
```go
import "github.com/primadi/lokstra/core/app"

apiRouter := router.New("api")
app := app.New("my-app", ":8080", apiRouter)
```

**See Also:**
- [lokstra.NewApp](lokstra.md#newapp) - Convenience function

---

### NewWithConfig
Creates a new App with custom listener configuration.

**Signature:**
```go
func NewWithConfig(name string, addr string, listenerType string, 
    cfg map[string]any, routers ...router.Router) *App
```

**Parameters:**
- `name` - App identifier
- `addr` - Listen address
- `listenerType` - Listener type ("default", "tls", "h2c", etc.)
- `cfg` - Listener configuration map
- `routers` - Routers to mount

**Returns:**
- `*App` - New app instance

**Example:**
```go
// HTTPS app
tlsConfig := map[string]any{
    "cert-file": "/path/to/cert.pem",
    "key-file":  "/path/to/key.pem",
}
app := app.NewWithConfig("api", ":443", "tls", tlsConfig, router)

// HTTP/2 Cleartext
h2cConfig := map[string]any{
    "read-timeout":  "30s",
    "write-timeout": "30s",
}
app := app.NewWithConfig("api", ":8080", "h2c", h2cConfig, router)
```

**Listener Types:**
- `"default"` - Standard HTTP listener
- `"tls"` - HTTPS listener
- `"h2c"` - HTTP/2 Cleartext
- Custom types can be registered

**Configuration Keys:**
- `read-timeout` - Max duration for reading request (string duration, e.g., "10s")
- `write-timeout` - Max duration for writing response
- `idle-timeout` - Max idle time between requests
- `read-header-timeout` - Time to read request headers
- `cert-file` - TLS certificate file path (for "tls")
- `key-file` - TLS private key file path (for "tls")

**See Also:**
- [lokstra.NewAppWithConfig](lokstra.md#newappwithconfig) - Convenience function

---

## Methods

### GetName
Returns the app name.

**Signature:**
```go
func (a *App) GetName() string
```

**Returns:**
- `string` - App identifier

**Example:**
```go
app := lokstra.NewApp("api", ":8080", router)
fmt.Println(app.GetName()) // Output: api
```

---

### GetAddress
Returns the listening address.

**Signature:**
```go
func (a *App) GetAddress() string
```

**Returns:**
- `string` - Listen address (e.g., ":8080", "127.0.0.1:8080")

**Example:**
```go
app := lokstra.NewApp("api", ":8080", router)
fmt.Println(app.GetAddress()) // Output: :8080
```

---

### GetRouter
Returns the main router of the app.

**Signature:**
```go
func (a *App) GetRouter() router.Router
```

**Returns:**
- `router.Router` - Main router instance

**Example:**
```go
app := lokstra.NewApp("api", ":8080", router)
mainRouter := app.GetRouter()
mainRouter.PrintRoutes()
```

---

### AddRouter
Adds a router to the app. If there's already a router, it will be chained.

**Signature:**
```go
func (a *App) AddRouter(rt router.Router)
```

**Parameters:**
- `rt` - Router to add

**Example:**
```go
app := lokstra.NewApp("api", ":8080")

apiRouter := router.New("api")
adminRouter := router.New("admin")

app.AddRouter(apiRouter)   // First router
app.AddRouter(adminRouter)  // Chained to first
```

**Notes:**
- Routers are cloned to avoid side effects
- Later routers are chained to previous ones
- Request flows through router chain until matched

---

### AddRouterWithPrefix
Adds a router with a path prefix.

**Signature:**
```go
func (a *App) AddRouterWithPrefix(rt router.Router, appPrefix string)
```

**Parameters:**
- `rt` - Router to add
- `appPrefix` - Path prefix for the router

**Example:**
```go
app := lokstra.NewApp("api", ":8080")

apiRouter := router.New("api")
adminRouter := router.New("admin")

app.AddRouter(apiRouter)                        // All paths
app.AddRouterWithPrefix(adminRouter, "/admin")  // Only /admin/*
```

---

### AddReverseProxies
Adds reverse proxy configurations to the app.

**Signature:**
```go
func (a *App) AddReverseProxies(proxies []*ReverseProxyConfig)
```

**Parameters:**
- `proxies` - Slice of reverse proxy configurations

**Example:**
```go
proxies := []*app.ReverseProxyConfig{
    {
        Prefix:      "/api",
        StripPrefix: true,
        Target:      "http://api-server:8080",
    },
    {
        Prefix:      "/auth",
        StripPrefix: false,
        Target:      "http://auth-server:9000",
        Rewrite: &app.ReverseProxyRewrite{
            From: "^/auth/(.*)$",
            To:   "/v2/$1",
        },
    },
}

app.AddReverseProxies(proxies)
```

**Notes:**
- Typically called from config loader
- Reverse proxy router is prepended (processed first)
- Useful for microservice architectures

**See Also:**
- [ReverseProxyConfig](#reverseproxyconfig) type
- [ReverseProxyRewrite](#reverseproxyrewrite) type

---

### Start
Starts the app listener. Blocks until the app stops or an error occurs.

**Signature:**
```go
func (a *App) Start() error
```

**Returns:**
- `error` - Error if listener fails to start

**Example:**
```go
app := lokstra.NewApp("api", ":8080", router)

// Start in background
go func() {
    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}()
```

**Notes:**
- Blocks until server stops
- Should be called in goroutine if you need non-blocking start
- Use `Run()` for automatic signal handling

---

### Shutdown
Gracefully shuts down the app with a timeout.

**Signature:**
```go
func (a *App) Shutdown(timeout time.Duration) error
```

**Parameters:**
- `timeout` - Maximum time to wait for shutdown

**Returns:**
- `error` - Error if shutdown fails

**Example:**
```go
app := lokstra.NewApp("api", ":8080", router)

// Start in background
go app.Start()

// Later... shutdown
if err := app.Shutdown(30 * time.Second); err != nil {
    log.Printf("Shutdown error: %v", err)
}
```

**Graceful Shutdown Process:**
1. Stop accepting new connections
2. Wait for active requests to complete (up to timeout)
3. Force close remaining connections after timeout
4. Return control

---

### Run
Starts the app and blocks until a termination signal is received. Handles graceful shutdown automatically.

**Signature:**
```go
func (a *App) Run(timeout time.Duration) error
```

**Parameters:**
- `timeout` - Graceful shutdown timeout

**Returns:**
- `error` - Error if app fails to start or shutdown fails

**Example:**
```go
app := lokstra.NewApp("api", ":8080", router)

// Run with 30s graceful shutdown
if err := app.Run(30 * time.Second); err != nil {
    fmt.Println("Error starting server:", err)
}
```

**Signals Handled:**
- `SIGINT` (Ctrl+C)
- `SIGTERM` (kill command)

**Use Cases:**
- Simple applications with single app
- Development and testing
- Scripts and CLI tools

**Notes:**
- For multiple apps, use `Server.Run()` instead
- Automatically prints start info
- Blocks until signal received

---

### PrintStartInfo
Prints app startup information to stdout.

**Signature:**
```go
func (a *App) PrintStartInfo()
```

**Example:**
```go
app := lokstra.NewApp("api", ":8080", router)
app.PrintStartInfo()
// Output:
// Starting [api] with 1 router(s) on address :8080
// GET /users
// POST /users
// ...
```

---

## Types

### ReverseProxyConfig
Configuration for a single reverse proxy.

**Definition:**
```go
type ReverseProxyConfig struct {
    Prefix      string               // URL prefix to match (e.g., "/api")
    StripPrefix bool                 // Whether to strip the prefix before forwarding
    Target      string               // Target backend URL (e.g., "http://api-server:8080")
    Rewrite     *ReverseProxyRewrite // Path rewrite rules (optional)
}
```

**Fields:**
- `Prefix` - Request path prefix to match
- `StripPrefix` - If true, removes prefix before forwarding
- `Target` - Backend server URL
- `Rewrite` - Optional path rewriting rules

**Example:**
```go
config := &app.ReverseProxyConfig{
    Prefix:      "/api/v1",
    StripPrefix: true,
    Target:      "http://backend:8080",
}
// Request: /api/v1/users
// Forwarded: http://backend:8080/users (prefix stripped)

config2 := &app.ReverseProxyConfig{
    Prefix:      "/api/v1",
    StripPrefix: false,
    Target:      "http://backend:8080",
}
// Request: /api/v1/users
// Forwarded: http://backend:8080/api/v1/users (prefix kept)
```

---

### ReverseProxyRewrite
Path rewrite rules for reverse proxy.

**Definition:**
```go
type ReverseProxyRewrite struct {
    From string // Pattern to match in path (regex supported)
    To   string // Replacement pattern
}
```

**Fields:**
- `From` - Regex pattern to match
- `To` - Replacement string (supports capture groups: `$1`, `$2`)

**Example:**
```go
rewrite := &app.ReverseProxyRewrite{
    From: "^/old/(.*)$",
    To:   "/new/$1",
}
// /old/users -> /new/users
// /old/posts/123 -> /new/posts/123

rewrite2 := &app.ReverseProxyRewrite{
    From: "^/api/v1/(.+)$",
    To:   "/v2/api/$1",
}
// /api/v1/users -> /v2/api/users
```

---

## Complete Examples

### Basic App
```go
package main

import (
    "time"
    "github.com/primadi/lokstra"
)

func main() {
    // Create router
    router := lokstra.NewRouter("api")
    router.GET("/health", healthCheck)
    router.GET("/users", listUsers)
    
    // Create app
    app := lokstra.NewApp("api", ":8080", router)
    
    // Run with graceful shutdown
    if err := app.Run(30 * time.Second); err != nil {
        log.Fatal(err)
    }
}

func healthCheck(c *lokstra.RequestContext) error {
    return c.Api.Success(map[string]string{"status": "ok"})
}

func listUsers(c *lokstra.RequestContext) error {
    return c.Api.Success(users)
}
```

### Multiple Routers
```go
func main() {
    // API router
    apiRouter := lokstra.NewRouter("api")
    apiRouter.GET("/users", listUsers)
    apiRouter.POST("/users", createUser)
    
    // Admin router
    adminRouter := lokstra.NewRouter("admin")
    adminRouter.Use("auth", "admin-role")
    adminRouter.GET("/stats", getStats)
    
    // Create app with both routers
    app := lokstra.NewApp("api", ":8080", apiRouter, adminRouter)
    
    // Or add separately
    app := lokstra.NewApp("api", ":8080")
    app.AddRouter(apiRouter)
    app.AddRouterWithPrefix(adminRouter, "/admin")
    
    if err := app.Run(30 * time.Second); err != nil {
        fmt.Println("Error starting server:", err)
    }
}
```

### HTTPS App
```go
func main() {
    router := lokstra.NewRouter("api")
    router.GET("/users", listUsers)
    
    // HTTPS configuration
    tlsConfig := map[string]any{
        "cert-file": "/etc/ssl/certs/server.crt",
        "key-file":  "/etc/ssl/private/server.key",
    }
    
    app := lokstra.NewAppWithConfig("api", ":443", "tls", tlsConfig, router)
    if err := app.Run(30 * time.Second); err != nil {
        fmt.Println("Error starting server:", err)
    }
}
```

### Reverse Proxy
```go
func main() {
    router := lokstra.NewRouter("api")
    router.GET("/health", healthCheck)
    
    app := lokstra.NewApp("gateway", ":8080", router)
    
    // Add reverse proxies
    proxies := []*app.ReverseProxyConfig{
        {
            Prefix:      "/api/users",
            StripPrefix: true,
            Target:      "http://user-service:8081",
        },
        {
            Prefix:      "/api/orders",
            StripPrefix: true,
            Target:      "http://order-service:8082",
        },
    }
    
    app.AddReverseProxies(proxies)
    if err := app.Run(30 * time.Second); err != nil {
        fmt.Println("Error starting server:", err)
    }
}
```

### Manual Lifecycle Control
```go
func main() {
    app := lokstra.NewApp("api", ":8080", router)
    
    // Start in background
    errCh := make(chan error, 1)
    go func() {
        if err := app.Start(); err != nil {
            errCh <- err
        }
    }()
    
    // Custom signal handling
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
    
    select {
    case sig := <-stop:
        log.Printf("Received signal: %v", sig)
        // Custom cleanup
        cleanupResources()
        // Graceful shutdown
        if err := app.Shutdown(30 * time.Second); err != nil {
            log.Printf("Shutdown error: %v", err)
        }
    case err := <-errCh:
        log.Fatal("App error:", err)
    }
}
```

---

## See Also

- **[lokstra](lokstra.md)** - Convenience functions (NewApp, NewAppWithConfig)
- **[Server](server.md)** - Managing multiple apps
- **[Router](router.md)** - Router API
- **[Listener](../08-advanced/listener.md)** - Custom listeners

---

## Related Guides

- **[App & Server Guide](../../01-essentials/05-app-and-server/)** - Lifecycle management tutorial
- **[Configuration](../../01-essentials/04-configuration/)** - Configuring apps from YAML
- **[Deployment](../../02-deep-dive/app-and-server/)** - Production deployment patterns
