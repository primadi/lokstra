# App & Server - Application Lifecycle

> **Learn application containers and server management**  
> **Time**: 20-25 minutes • **Level**: Beginner • **Concepts**: 3

---

## 🎯 What You'll Learn

- Combine multiple routers into an App
- Manage multiple apps with a Server
- Start servers with graceful shutdown
- Handle termination signals automatically

---

## 📖 Concepts

### 1. App - Application Container

An **App** is a container that:
- Combines multiple routers
- Listens on a specific address
- Manages HTTP server lifecycle

**Think of it as**: One web application running on one port

```go
import "github.com/primadi/lokstra"

// Create routers
apiRouter := lokstra.NewRouter("api")
adminRouter := lokstra.NewRouter("admin")

// Combine into one app
app := lokstra.NewApp("my-app", ":8080", apiRouter, adminRouter)
```

**Why Apps?**
- ✅ Group related routers together
- ✅ Single listening address
- ✅ Unified middleware
- ✅ Easier management

### 2. Server - Server Manager

A **Server** manages multiple apps:
- Can run apps on different ports
- Handles graceful shutdown for all apps
- Coordinates startup and shutdown

**Think of it as**: A process manager for your applications

```go
// Create apps
apiApp := lokstra.NewApp("api", ":8080", apiRouter)
adminApp := lokstra.NewApp("admin", ":9000", adminRouter)

// Server manages both
server := lokstra.NewServer("main", apiApp, adminApp)
```

**Why Servers?**
- ✅ Run multiple apps in one process
- ✅ Unified shutdown handling
- ✅ Better resource management
- ✅ Production-ready pattern

### 3. Graceful Shutdown

Both App and Server support **graceful shutdown**:

1. **Stop** accepting new requests
2. **Wait** for active requests to complete (with timeout)
3. **Shutdown** cleanly

**Methods:**

```go
// Manual start/shutdown
app.Start()         // Blocks until server stops
app.Shutdown(30 * time.Second)

// Automatic signal handling (recommended!)
app.Run(30 * time.Second)  // Handles SIGINT/SIGTERM automatically
```

**Signals Handled:**
- `SIGINT` - Ctrl+C in terminal
- `SIGTERM` - Kubernetes, Docker, systemd

---

## 💻 Example 1: Basic App

**Single app with multiple routers:**

```go
package main

import (
    "log"
    "net/http"
    "time"
    
    "github.com/primadi/lokstra"
)

func main() {
    // Create API router
    apiRouter := lokstra.NewRouter("api")
    apiRouter.Get("/users", GetUsersHandler)
    apiRouter.Get("/products", GetProductsHandler)
    
    // Create admin router
    adminRouter := lokstra.NewRouter("admin")
    adminRouter.Get("/stats", GetStatsHandler)
    adminRouter.Get("/logs", GetLogsHandler)
    
    // Combine into one app
    app := lokstra.NewApp("web-app", ":8080", apiRouter, adminRouter)
    
    log.Println("Server starting on :8080")
    log.Println("Press Ctrl+C to stop")
    
    // Run with 30s graceful shutdown timeout
    if err := app.Run(30 * time.Second); err != nil {
        log.Fatal(err)
    }
}

func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`{"users": ["Alice", "Bob"]}`))
}

func GetProductsHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`{"products": ["Book", "Pen"]}`))
}

func GetStatsHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`{"requests": 1234, "uptime": "2h"}`))
}

func GetLogsHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`{"logs": ["Log 1", "Log 2"]}`))
}
```

**Run:**

```bash
go run main.go

# Test
curl http://localhost:8080/users
curl http://localhost:8080/products
curl http://localhost:8080/stats
curl http://localhost:8080/logs

# Stop with Ctrl+C - graceful shutdown happens automatically!
```

**Output on shutdown:**

```
^C
Received shutdown signal: interrupt
[NETHTTP] Initiating graceful shutdown for app at :8080
App 'web-app' has been gracefully shutdown.
```

---

## 💻 Example 2: Server with Multiple Apps

**Run multiple apps on different ports:**

```go
package main

import (
    "log"
    "net/http"
    "time"
    
    "github.com/primadi/lokstra"
)

func main() {
    // API app on port 8080
    apiRouter := lokstra.NewRouter("api")
    apiRouter.Get("/health", HealthHandler)
    apiRouter.Get("/users", GetUsersHandler)
    apiApp := lokstra.NewApp("api-app", ":8080", apiRouter)
    
    // Admin app on port 9000
    adminRouter := lokstra.NewRouter("admin")
    adminRouter.Get("/dashboard", DashboardHandler)
    adminRouter.Get("/users", AdminUsersHandler)
    adminApp := lokstra.NewApp("admin-app", ":9000", adminRouter)
    
    // Metrics app on port 9090
    metricsRouter := lokstra.NewRouter("metrics")
    metricsRouter.Get("/metrics", MetricsHandler)
    metricsApp := lokstra.NewApp("metrics-app", ":9090", metricsRouter)
    
    // Server manages all apps
    server := lokstra.NewServer("main-server", apiApp, adminApp, metricsApp)
    
    // All apps start together, shutdown together
    log.Println("Server starting all apps...")
    if err := server.Run(30 * time.Second); err != nil {
        log.Fatal(err)
    }
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`{"status": "ok"}`))
}

func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`{"users": ["Alice", "Bob", "Charlie"]}`))
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`{"dashboard": "admin"}`))
}

func AdminUsersHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`{"users": ["Alice", "Bob"], "actions": ["edit", "delete"]}`))
}

func MetricsHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`{"cpu": "25%", "memory": "512MB", "requests": 1234}`))
}
```

**Run:**

```bash
go run main.go
```

**Output:**

```
Server 'main-server' starting with 3 app(s):
Starting [api-app] with 1 router(s) on address :8080
  GET /health
  GET /users
Starting [admin-app] with 1 router(s) on address :9000
  GET /dashboard
  GET /users
Starting [metrics-app] with 1 router(s) on address :9090
  GET /metrics
Press CTRL+C to stop the server...
```

**Test in different terminals:**

```bash
# API app
curl http://localhost:8080/health
curl http://localhost:8080/users

# Admin app
curl http://localhost:9000/dashboard
curl http://localhost:9000/users

# Metrics app
curl http://localhost:9090/metrics
```

**Stop with Ctrl+C:**

```
^C
Received shutdown signal: interrupt
[NETHTTP] Initiating graceful shutdown for app at :8080
[NETHTTP] Initiating graceful shutdown for app at :9000
[NETHTTP] Initiating graceful shutdown for app at :9090
App 'api-app' has been gracefully shutdown.
App 'admin-app' has been gracefully shutdown.
App 'metrics-app' has been gracefully shutdown.
Server 'main-server' has been gracefully shutdown.
```

---

## 💻 Example 3: App with Router Chaining

**Mount routers at different paths:**

```go
package main

import (
    "log"
    "net/http"
    "time"
    
    "github.com/primadi/lokstra"
)

func main() {
    // Create routers
    apiV1Router := lokstra.NewRouter("api-v1")
    apiV1Router.Get("/users", V1GetUsersHandler)
    apiV1Router.Post("/users", V1CreateUserHandler)
    
    apiV2Router := lokstra.NewRouter("api-v2")
    apiV2Router.Get("/users", V2GetUsersHandler)
    apiV2Router.Post("/users", V2CreateUserHandler)
    
    publicRouter := lokstra.NewRouter("public")
    publicRouter.Get("/about", AboutHandler)
    publicRouter.Get("/contact", ContactHandler)
    
    // Chain routers - they combine into one handler
    // All routes are merged together
    app := lokstra.NewApp("api", ":8080", 
        apiV1Router,  // Handles /users (v1)
        apiV2Router,  // Handles /users (v2)
        publicRouter, // Handles /about, /contact
    )
    
    log.Println("Server starting on :8080")
    app.Run(30 * time.Second)
}

func V1GetUsersHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`{"version": "v1", "users": ["Alice", "Bob"]}`))
}

func V1CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`{"version": "v1", "created": true}`))
}

func V2GetUsersHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`{"version": "v2", "users": [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]}`))
}

func V2CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`{"version": "v2", "created": true, "id": 3}`))
}

func AboutHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`{"page": "about", "content": "About us..."}`))
}

func ContactHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`{"page": "contact", "email": "contact@example.com"}`))
}
```

**Note**: When multiple routers define the same path, the **last one wins**. For API versioning, use route groups instead (see Router documentation).

---

## 🎯 Best Practices

### 1. App Organization

**✅ DO: One app per service**

```go
// Good: Separate concerns
apiApp := lokstra.NewApp("api", ":8080", apiRouter)
adminApp := lokstra.NewApp("admin", ":9000", adminRouter)

server := lokstra.NewServer("main", apiApp, adminApp)
```

**✅ DO: Group related routers in one app**

```go
// Good: Related functionality together
app := lokstra.NewApp("web", ":8080",
    publicRouter,    // Public pages
    apiRouter,       // API endpoints
    websocketRouter, // WebSocket connections
)
```

**✗ DON'T: Mix unrelated concerns**

```go
// Bad: Admin and API in same app without clear separation
app := lokstra.NewApp("everything", ":8080", 
    publicRouter,
    adminRouter,  // Security risk: admin on public port
)
```

### 2. Server Usage

**✅ DO: Use Server for multiple apps**

```go
// Good: Multiple services managed together
server := lokstra.NewServer("main",
    lokstra.NewApp("api", ":8080", apiRouter),
    lokstra.NewApp("admin", ":9000", adminRouter),
    lokstra.NewApp("metrics", ":9090", metricsRouter),
)
```

**✅ DO: Use Run() for production**

```go
// Good: Automatic signal handling
server.Run(30 * time.Second)
```

**✗ DON'T: Use Start() in production**

```go
// Bad: No graceful shutdown
server.Start()  // Must handle signals manually
```

### 3. Graceful Shutdown

**✅ DO: Set appropriate timeout**

```go
// Good: Enough time for requests to complete
app.Run(30 * time.Second)  // 30s timeout

// For long-running requests
app.Run(60 * time.Second)  // 60s timeout
```

**✅ DO: Log shutdown events**

```go
log.Println("Starting server...")
if err := server.Run(30 * time.Second); err != nil {
    log.Printf("Server stopped with error: %v", err)
} else {
    log.Println("Server stopped gracefully")
}
```

**✗ DON'T: Use very short timeouts**

```go
// Bad: Requests may be killed mid-processing
app.Run(1 * time.Second)  // Too short!
```

### 4. Port Selection

**✅ DO: Use standard ports**

```go
// Good: Standard conventions
apiApp := lokstra.NewApp("api", ":8080", apiRouter)      // HTTP
adminApp := lokstra.NewApp("admin", ":9000", adminRouter) // Admin
metricsApp := lokstra.NewApp("metrics", ":9090", metricsRouter) // Metrics
```

**✅ DO: Use environment variables**

```go
// Good: Configurable
port := os.Getenv("PORT")
if port == "" {
    port = ":8080"
}
app := lokstra.NewApp("api", port, apiRouter)
```

**✗ DON'T: Hardcode non-standard ports**

```go
// Bad: Non-standard, hard to remember
app := lokstra.NewApp("api", ":37294", apiRouter)
```

---

## 🔍 Common Patterns

### Pattern 1: Microservices Architecture

```go
func main() {
    // User service
    userRouter := lokstra.NewRouter("user-api")
    setupUserRoutes(userRouter)
    userApp := lokstra.NewApp("user-service", ":8001", userRouter)
    
    // Product service
    productRouter := lokstra.NewRouter("product-api")
    setupProductRoutes(productRouter)
    productApp := lokstra.NewApp("product-service", ":8002", productRouter)
    
    // Order service
    orderRouter := lokstra.NewRouter("order-api")
    setupOrderRoutes(orderRouter)
    orderApp := lokstra.NewApp("order-service", ":8003", orderRouter)
    
    // Run all services in one process
    server := lokstra.NewServer("microservices", userApp, productApp, orderApp)
    server.Run(30 * time.Second)
}
```

### Pattern 2: API Gateway + Services

```go
func main() {
    // API Gateway (public-facing)
    gatewayRouter := lokstra.NewRouter("gateway")
    gatewayRouter.Get("/api/*", ProxyHandler) // Proxy to services
    gatewayApp := lokstra.NewApp("gateway", ":80", gatewayRouter)
    
    // Internal services
    userApp := lokstra.NewApp("users", ":8001", userRouter)
    productApp := lokstra.NewApp("products", ":8002", productRouter)
    
    server := lokstra.NewServer("main", gatewayApp, userApp, productApp)
    server.Run(30 * time.Second)
}
```

### Pattern 3: Separate Admin Interface

```go
func main() {
    // Public API
    apiRouter := lokstra.NewRouter("api")
    apiRouter.Use(PublicMiddleware)
    apiApp := lokstra.NewApp("api", ":8080", apiRouter)
    
    // Admin interface (different port, different middleware)
    adminRouter := lokstra.NewRouter("admin")
    adminRouter.Use(AdminAuthMiddleware)
    adminRouter.Use(AdminLoggingMiddleware)
    adminApp := lokstra.NewApp("admin", ":9000", adminRouter)
    
    server := lokstra.NewServer("main", apiApp, adminApp)
    server.Run(30 * time.Second)
}
```

### Pattern 4: Health Checks + Metrics

```go
func main() {
    // Main application
    apiApp := lokstra.NewApp("api", ":8080", apiRouter)
    
    // Health check endpoint (separate port for load balancers)
    healthRouter := lokstra.NewRouter("health")
    healthRouter.Get("/health", HealthHandler)
    healthRouter.Get("/ready", ReadyHandler)
    healthApp := lokstra.NewApp("health", ":8081", healthRouter)
    
    // Metrics (Prometheus)
    metricsRouter := lokstra.NewRouter("metrics")
    metricsRouter.Get("/metrics", PrometheusHandler)
    metricsApp := lokstra.NewApp("metrics", ":9090", metricsRouter)
    
    server := lokstra.NewServer("main", apiApp, healthApp, metricsApp)
    server.Run(30 * time.Second)
}
```

---

## 📊 App vs Server Comparison

| Aspect | App | Server |
|--------|-----|--------|
| **Purpose** | Run HTTP server on one port | Manage multiple apps |
| **Routers** | Multiple routers combined | N/A (apps manage routers) |
| **Ports** | Single port | Multiple ports (one per app) |
| **Use Case** | Single web application | Multiple services |
| **Shutdown** | Shuts down one app | Shuts down all apps |
| **Typical** | Small projects, single service | Production, microservices |

---

## 🚀 Start/Run Comparison

| Method | Behavior | When to Use |
|--------|----------|-------------|
| **Start()** | Starts server, blocks until error | Manual signal handling needed |
| **Run(timeout)** | Starts + auto signal handling + graceful shutdown | **Production (recommended!)** |
| **Shutdown(timeout)** | Manually trigger shutdown | Testing, custom logic |

**Example: Manual control**

```go
// Start server
go func() {
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}()

// Wait for custom condition
<-customShutdownSignal

// Manual shutdown
server.Shutdown(30 * time.Second)
```

**Example: Automatic (recommended)**

```go
// Everything handled automatically
server.Run(30 * time.Second)
```

---

## ✅ Quick Checklist

After completing this section, you should be able to:

- [ ] Create apps with multiple routers
- [ ] Create servers with multiple apps
- [ ] Use Run() for graceful shutdown
- [ ] Handle multiple ports in one process
- [ ] Understand shutdown flow

---

## 🚀 Next Steps

**Ready for the grand finale?** Continue to:

👉 [Putting It Together](../06-putting-it-together/README.md) - Build a complete REST API!

**Or review:**
- [Router](../01-router/README.md) - Router fundamentals
- [Service](../02-service/README.md) - Service patterns
- [Middleware](../03-middleware/README.md) - Request/response processing
- [Configuration](../04-configuration/README.md) - YAML configuration
