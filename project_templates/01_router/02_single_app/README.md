# Lokstra Single App Template

This template demonstrates using Lokstra's **App** wrapper for production-ready features like graceful shutdown, multiple routers, and structured application lifecycle.

## What's Included

This template demonstrates:

- **Lokstra App**: Production-ready application wrapper
- **Multiple Routers**: API router + dedicated health check router
- **Graceful Shutdown**: Handles SIGINT/SIGTERM with timeout
- **Health Checks**: `/health` and `/ready` endpoints
- **Startup Info**: Automatic route listing and app information
- **Auto Signal Handling**: Ctrl+C gracefully stops the app
- **RESTful API**: Complete CRUD operations for users and roles

## Key Differences from Router-Only

| Feature | 01_router_only | 02_single_app |
|---------|---------------|---------------|
| Server | `http.ListenAndServe` | `app.Run()` |
| Shutdown | Immediate | Graceful (30s timeout) |
| Signal Handling | Manual | Automatic |
| Multiple Routers | Manual chaining | Built-in support |
| Startup Info | Manual | `app.PrintStartInfo()` |
| Production Ready | Basic | ✅ Yes |

## Project Structure

```
.
├── main.go         # Application entry with App initialization
├── router.go       # Multiple router setup (API + Health)
├── handlers.go     # Request handlers + health check handlers
├── middleware.go   # Custom middleware
├── test.http       # HTTP test file
├── go.mod          # Go module dependencies
└── README.md       # This file
```

## Quick Start

### 1. Install Dependencies

```bash
go mod download
```

### 2. Run the Application

```bash
go run .
```

You'll see:
```
Starting application...
Press Ctrl+C to gracefully shutdown
============================================
Application: demo-app
Address: :3000
Routers: 2
============================================
Health Router Routes:
  GET   /health
  GET   /ready
API Router Routes:
  GET   /users
  GET   /users/:id
  POST  /users
  ...
============================================
```

### 3. Test Graceful Shutdown

1. Start the app: `go run .`
2. Press `Ctrl+C` in terminal
3. Observe: "Received shutdown signal: interrupt"
4. App waits for active requests (up to 30 seconds)
5. Clean shutdown: "Application stopped gracefully"

### 4. Test the Endpoints

Use the `test.http` file with VS Code REST Client extension:

**Health Checks:**
```http
GET http://localhost:3000/health
GET http://localhost:3000/ready
```

**API Endpoints:** (Same as router-only template)
```http
GET http://localhost:3000/users
POST http://localhost:3000/users
...
```

## Key Concepts

### Multiple Routers in One App

One of Lokstra's powerful features: **multiple routers in a single app**

```go
// Create separate routers for different concerns
apiRouter := setupAPIRouter()
healthRouter := setupHealthRouter()

// Add both to the same app
// Order matters: health router is checked first
app := lokstra.NewApp("demo-app", ":3000", healthRouter, apiRouter)
```

**Why Multiple Routers?**
- ✅ Separation of concerns (API vs health checks)
- ✅ Different middleware per router
- ✅ Independent route organization
- ✅ Easy to add/remove functionality

### Health Check Router

Dedicated router for health checks, no middleware overhead:

```go
func setupHealthRouter() lokstra.Router {
    r := lokstra.NewRouter("health_router")
    r.GET("/health", handleHealth)   // Liveness probe
    r.GET("/ready", handleReady)     // Readiness probe
    return r
}
```

**Use Cases:**
- Kubernetes liveness/readiness probes
- Load balancer health checks
- Monitoring systems

### Graceful Shutdown

The app handles shutdown signals automatically:

```go
app.Run(30 * time.Second)  // 30 second graceful shutdown timeout
```

**What happens on Ctrl+C or SIGTERM:**
1. Stop accepting new requests
2. Wait for active requests to complete (up to 30 seconds)
3. Close connections and resources
4. Exit cleanly

### App Benefits Over Plain Router

```go
// ❌ Plain router - manual everything
router := lokstra.NewRouter("demo")
http.ListenAndServe(":3000", router)  // No graceful shutdown

// ✅ App - production ready
app := lokstra.NewApp("demo", ":3000", router)
app.Run(30 * time.Second)  // Graceful shutdown built-in
```

## Handler Examples

### Health Check Handler

```go
type HealthStatus struct {
    Status    string    `json:"status"`
    Timestamp time.Time `json:"timestamp"`
    Version   string    `json:"version"`
}

func handleHealth() (*HealthStatus, error) {
    return &HealthStatus{
        Status:    "healthy",
        Timestamp: time.Now(),
        Version:   "1.0.0",
    }, nil
}
```

### API Handlers

Same pattern as router-only template with auto-binding and validation.

## Customization

### Adding More Routers

```go
metricsRouter := lokstra.NewRouter("metrics")
metricsRouter.GET("/metrics", handleMetrics)

app := lokstra.NewApp("demo", ":3000", 
    healthRouter, 
    metricsRouter,  // Add new router
    apiRouter,
)
```

### Changing Shutdown Timeout

```go
app.Run(60 * time.Second)  // 60 seconds instead of 30
```

### Custom Startup Actions

```go
app := lokstra.NewApp("demo", ":3000", router)

// Do something before starting
log.Println("Initializing database...")
// ... initialization code ...

app.Run(30 * time.Second)
```

## Production Deployment

### Docker

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server .

FROM alpine:latest
COPY --from=builder /app/server /server
CMD ["/server"]
```

### Kubernetes

Health check configuration:
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 3000
  initialDelaySeconds: 10
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /ready
    port: 3000
  initialDelaySeconds: 5
  periodSeconds: 5
```

### Environment Variables

Add environment-based configuration:
```go
port := os.Getenv("PORT")
if port == "" {
    port = ":3000"
}
app := lokstra.NewApp("demo", port, router)
```

## Next Steps

- Add database connections
- Implement authentication middleware
- Add structured logging
- Set up metrics collection
- Configure environment-based settings
- Add unit tests for handlers

## Learn More

- [Router-Only Template](../01_router_only) - Simpler, for library usage
- [Multi-App Template](../03_multi_app) - Multiple apps on different ports
- [Lokstra Documentation](https://github.com/primadi/lokstra)

## Notes

**When to use Single App:**
- ✅ Production deployments
- ✅ Need graceful shutdown
- ✅ Want automatic signal handling
- ✅ Multiple routers in one service
- ✅ Health check endpoints

**When to use Router-Only:**
- Integrating into existing apps
- Maximum control over server setup
- Learning router concepts
- Library usage
