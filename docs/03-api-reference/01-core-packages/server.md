# Server

> Server container for managing multiple apps

## Overview

The Server type is a container that manages the lifecycle of one or more Apps. It handles concurrent startup, graceful shutdown, signal handling, and automatic app merging (when multiple apps listen on the same address).

**Key Point**: Server is NOT in the request flow. It only manages app lifecycle.

```
Request → App → Router → Handler
            ↑
          Server manages lifecycle only
```

## Import Path

```go
import "github.com/primadi/lokstra/core/server"

// Or use the main package
import "github.com/primadi/lokstra"
server := lokstra.NewServer("my-server", app1, app2)
```

---

## Type Definition

```go
type Server struct {
    Name         string     // Server identifier
    BaseUrl      string     // Base URL of the server (optional)
    DeploymentID string     // Deployment ID for grouping servers (optional)
    Apps         []*app.App // Apps to manage
}
```

**Fields:**
- `Name` - Server identifier (for logging and management)
- `BaseUrl` - External base URL (e.g., "https://api.example.com")
- `DeploymentID` - Deployment identifier (used in multi-deployment setups)
- `Apps` - List of apps managed by this server

---

## Constructor

### New
Creates a new Server instance with given apps.

**Signature:**
```go
func New(name string, apps ...*app.App) *Server
```

**Parameters:**
- `name` - Server identifier
- `apps` - Zero or more apps to manage

**Returns:**
- `*Server` - New server instance

**Example:**
```go
import "github.com/primadi/lokstra/core/server"

apiApp := app.New("api", ":8080", apiRouter)
adminApp := app.New("admin", ":9000", adminRouter)

server := server.New("my-server", apiApp, adminApp)
```

**See Also:**
- [lokstra.NewServer](lokstra#newserver) - Convenience function

---

## Methods

### GetName
Returns the server name.

**Signature:**
```go
func (s *Server) GetName() string
```

**Returns:**
- `string` - Server identifier

**Example:**
```go
server := lokstra.NewServer("my-server", app)
fmt.Println(server.GetName()) // Output: my-server
```

---

### AddApp
Adds an app to the server.

**Signature:**
```go
func (s *Server) AddApp(a *app.App)
```

**Parameters:**
- `a` - App to add

**Example:**
```go
server := lokstra.NewServer("my-server")

apiApp := lokstra.NewApp("api", ":8080", apiRouter)
adminApp := lokstra.NewApp("admin", ":9000", adminRouter)

server.AddApp(apiApp)
server.AddApp(adminApp)
```

**Panics:**
- If called after server is built (after Start() or Run())

---

### Start
Starts all apps concurrently. Blocks until all apps stop or an error occurs.

**Signature:**
```go
func (s *Server) Start() error
```

**Returns:**
- `error` - Error if any app fails to start (may contain multiple errors)

**Example:**
```go
server := lokstra.NewServer("my-server", app1, app2)

// Start in background
go func() {
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}()
```

**Behavior:**
- Starts all apps concurrently in separate goroutines
- Automatically merges apps on same address
- Blocks until all apps complete
- Returns joined errors if any app fails

**Auto-Merging:**
```go
app1 := lokstra.NewApp("api", ":8080", apiRouter)
app2 := lokstra.NewApp("admin", ":8080", adminRouter) // Same port!

server := lokstra.NewServer("my-server", app1, app2)
// Server automatically merges app1 and app2 into single app
// Both routers served on :8080
```

---

### Shutdown
Gracefully shuts down all apps with a timeout.

**Signature:**
```go
func (s *Server) Shutdown(timeout any) error
```

**Parameters:**
- `timeout` - Shutdown timeout (multiple formats supported)

**Timeout Formats:**
- `time.Duration` - e.g., `30 * time.Second`
- `int` - Seconds, e.g., `30` (converted to duration)
- `string` - Duration string, e.g., `"30s"`, `"1m"`

**Returns:**
- `error` - Error if any app fails to shutdown gracefully

**Example:**
```go
// Duration
server.Shutdown(30 * time.Second)

// Integer (seconds)
server.Shutdown(30)

// String
server.Shutdown("30s")
server.Shutdown("1m")
```

**Graceful Shutdown Process:**
1. Calls `Shutdown()` on all apps concurrently
2. Each app waits for active requests (up to timeout)
3. Shuts down all registered services
4. Returns joined errors if any app fails

---

### Run
Starts the server and blocks until a termination signal is received. Handles graceful shutdown automatically.

**Signature:**
```go
func (s *Server) Run(timeout time.Duration) error
```

**Parameters:**
- `timeout` - Graceful shutdown timeout

**Returns:**
- `error` - Error if server fails to start or shutdown fails

**Example:**
```go
server := lokstra.NewServer("my-server", app1, app2)

// Run with 30s graceful shutdown
if err := server.Run(30 * time.Second); err != nil {
    log.Fatal(err)
}
```

**Signals Handled:**
- `SIGINT` (Ctrl+C)
- `SIGTERM` (kill command)

**Use Cases:**
- Production servers with multiple apps
- Microservices
- Complex deployments

**Notes:**
- Preferred method for running servers
- Automatically calls `PrintStartInfo()`
- Blocks until signal received
- Handles graceful shutdown automatically

---

### PrintStartInfo
Prints server startup information to stdout.

**Signature:**
```go
func (s *Server) PrintStartInfo()
```

**Example:**
```go
server := lokstra.NewServer("my-server", app1, app2)
server.PrintStartInfo()
// Output:
// Server 'my-server' starting with 2 app(s):
// Starting [api] with 1 router(s) on address :8080
// GET /users
// POST /users
// Starting [admin] with 1 router(s) on address :9000
// GET /stats
// Press CTRL+C to stop the server...
```

**Notes:**
- Called automatically by `Run()`
- Useful for debugging
- Shows all apps and their routes

---

## Complete Examples

### Single Server, Multiple Apps
```go
package main

import (
    "time"
    "github.com/primadi/lokstra"
)

func main() {
    // API app on :8080
    apiRouter := lokstra.NewRouter("api")
    apiRouter.GET("/users", listUsers)
    apiRouter.POST("/users", createUser)
    apiApp := lokstra.NewApp("api", ":8080", apiRouter)
    
    // Admin app on :9000
    adminRouter := lokstra.NewRouter("admin")
    adminRouter.Use("auth", "admin-role")
    adminRouter.GET("/stats", getStats)
    adminRouter.GET("/users", adminListUsers)
    adminApp := lokstra.NewApp("admin", ":9000", adminRouter)
    
    // Metrics app on :9090
    metricsRouter := lokstra.NewRouter("metrics")
    metricsRouter.GET("/metrics", prometheusHandler)
    metricsApp := lokstra.NewApp("metrics", ":9090", metricsRouter)
    
    // Create server with all apps
    server := lokstra.NewServer("my-server", apiApp, adminApp, metricsApp)
    
    // Run with 30s graceful shutdown
    if err := server.Run(30 * time.Second); err != nil {
        log.Fatal(err)
    }
}
```

### Production Server with HTTPS
```go
func main() {
    // HTTP app (redirect to HTTPS)
    httpRouter := lokstra.NewRouter("http")
    httpRouter.GET("/*", redirectToHTTPS)
    httpApp := lokstra.NewApp("http", ":80", httpRouter)
    
    // HTTPS app
    tlsConfig := map[string]any{
        "cert-file": "/etc/ssl/certs/server.crt",
        "key-file":  "/etc/ssl/private/server.key",
    }
    
    apiRouter := lokstra.NewRouter("api")
    apiRouter.GET("/users", listUsers)
    httpsApp := lokstra.NewAppWithConfig("https", ":443", "tls", tlsConfig, apiRouter)
    
    // Server with both apps
    server := lokstra.NewServer("production-server", httpApp, httpsApp)
    if err := server.Run(30 * time.Second); err != nil {
        fmt.Println("Error starting server:", err)
    }
}

func redirectToHTTPS(c *lokstra.RequestContext) error {
    target := "https://" + c.R.Host + c.R.URL.Path
    if c.R.URL.RawQuery != "" {
        target += "?" + c.R.URL.RawQuery
    }
    http.Redirect(c.W, c.R, target, http.StatusMovedPermanently)
    return nil
}
```

### Manual Lifecycle Control
```go
func main() {
    server := lokstra.NewServer("my-server", app1, app2, app3)
    
    // Start in background
    errCh := make(chan error, 1)
    go func() {
        if err := server.Start(); err != nil {
            errCh <- err
        }
    }()
    
    // Custom signal handling
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
    
    for {
        select {
        case sig := <-stop:
            switch sig {
            case syscall.SIGHUP:
                // Reload configuration
                log.Println("Reloading configuration...")
                reloadConfig()
            case syscall.SIGINT, syscall.SIGTERM:
                // Graceful shutdown
                log.Println("Shutting down...")
                if err := server.Shutdown(30 * time.Second); err != nil {
                    log.Printf("Shutdown error: %v", err)
                }
                return
            }
        case err := <-errCh:
            log.Fatal("Server error:", err)
        }
    }
}
```

### App Merging (Same Address)
```go
func main() {
    // Multiple apps on same port
    apiV1Router := lokstra.NewRouter("api-v1")
    apiV1Router.GET("/v1/users", v1ListUsers)
    apiV1App := lokstra.NewApp("api-v1", ":8080", apiV1Router)
    
    apiV2Router := lokstra.NewRouter("api-v2")
    apiV2Router.GET("/v2/users", v2ListUsers)
    apiV2App := lokstra.NewApp("api-v2", ":8080", apiV2Router) // Same port!
    
    adminRouter := lokstra.NewRouter("admin")
    adminRouter.GET("/admin/stats", getStats)
    adminApp := lokstra.NewApp("admin", ":8080", adminRouter) // Same port!
    
    // Server automatically merges all three apps
    server := lokstra.NewServer("my-server", apiV1App, apiV2App, adminApp)
    
    // Result: Single app on :8080 with all routers chained
    // Routes available:
    // - /v1/users
    // - /v2/users
    // - /admin/stats
    
    if err := server.Run(30 * time.Second); err != nil {
        fmt.Println("Error starting server:", err)
    }
}
```

### Health Check with Admin Port
```go
func main() {
    // Main API
    apiRouter := lokstra.NewRouter("api")
    apiRouter.GET("/users", listUsers)
    apiApp := lokstra.NewApp("api", ":8080", apiRouter)
    
    // Health check / admin interface (internal port)
    healthRouter := lokstra.NewRouter("health")
    healthRouter.GET("/health", healthCheck)
    healthRouter.GET("/ready", readinessCheck)
    healthRouter.GET("/metrics", prometheusMetrics)
    healthApp := lokstra.NewApp("health", "127.0.0.1:9090", healthRouter)
    
    server := lokstra.NewServer("my-server", apiApp, healthApp)
    if err := server.Run(30 * time.Second); err != nil {
        fmt.Println("Error starting server:", err)
    }
}
```

---

## Best Practices

### 1. Use Run() for Production
```go
// ✅ Recommended
if err := server.Run(30 * time.Second); err != nil {
    fmt.Println("Error starting server:", err)
}

// 🚫 Avoid (unless you need custom control)
go server.Start()
// ... manual signal handling
```

### 2. Separate Admin/Metrics Ports
```go
// ✅ Good: Admin on separate port
apiApp := lokstra.NewApp("api", ":8080", apiRouter)
adminApp := lokstra.NewApp("admin", "127.0.0.1:9000", adminRouter)

// 🚫 Avoid: Admin on same port (security risk)
apiApp.AddRouter(adminRouter)
```

### 3. Graceful Shutdown Timeout
```go
// ✅ Production: 30-60 seconds
if err := server.Run(30 * time.Second); err != nil {
    fmt.Println("Error starting server:", err)
}

// ✅ Development: 5-10 seconds
if err := server.Run(5 * time.Second); err != nil {
    fmt.Println("Error starting server:", err)
}

// 🚫 Too short: May terminate active requests
if err := server.Run(1 * time.Second); err != nil {
    fmt.Println("Error starting server:", err)
}
```

### 4. Error Handling
```go
// ✅ Check errors
if err := server.Run(30 * time.Second); err != nil {
    log.Fatal(err)
}

// 🚫 Ignore errors
if err := server.Run(30 * time.Second); err != nil {
    fmt.Println("Error starting server:", err)
}
```

---

## See Also

- **[lokstra](lokstra)** - Convenience function (NewServer)
- **[App](app)** - App lifecycle and configuration
- **[Router](router)** - Router API

---

## Related Guides

- **[App & Server Guide](../../01-router-guide/05-app-and-server/)** - Lifecycle management tutorial
- **[Multi-Deployment](../../00-introduction/examples/04-multi-deployment/)** - Example with multiple apps
- **[Production Deployment](../../02-deep-dive/app-and-server/)** - Best practices for production
