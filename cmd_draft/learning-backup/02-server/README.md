# 03-server: Managing Multiple Apps with a Server

## What You'll Learn
- Create a Server to manage multiple Apps
- Run Apps on different ports
- Understand automatic App merging
- Centralized lifecycle management

## Key Concepts

### Server
A **Server** is the top-level orchestrator that:
- Manages multiple Apps
- Handles multiple network listeners
- Provides centralized graceful shutdown
- **Automatically merges Apps with the same address**

### App Merging
When multiple Apps listen on the same address:
- They are automatically merged into one listener
- All routers from both Apps are combined
- This saves resources and simplifies deployment
- The first App's listener config is used

## Architecture Hierarchy

```
Server                    (Top level - manages everything)
  ├── App1 (:8080)        (Container for routers)
  │   ├── Router A
  │   └── Router B
  ├── App2 (:8080)        (MERGED with App1 - same address!)
  │   └── Router C
  └── App3 (:8081)        (Separate listener)
      └── Router D
```

**Result**: Only 2 actual network listeners (8080 and 8081)

## Running the Example

```bash
cd cmd/learning/01-basics/03-server
go run main.go
```

## Testing

```bash
# Public API (port 8080)
curl http://localhost:8080/api/users
curl http://localhost:8080/api/products
curl http://localhost:8080/health
curl http://localhost:8080/internal/config  # From merged internal-api

# Admin API (port 8081)  
curl http://localhost:8081/admin/stats
curl http://localhost:8081/admin/users
curl http://localhost:8081/health
```

## Key Concepts Illustrated

### 1. Multiple Ports
Different Apps can listen on different ports:
- Public API on `:8080`
- Admin API on `:8081`
- This enables separation of concerns and security

### 2. Automatic Merging
Apps with the same address merge:
```go
publicApp := NewApp("public-api", ":8080", ...)
internalApp := NewApp("internal-api", ":8080", ...)
server.AddApp(publicApp)
server.AddApp(internalApp)
// Result: One listener on :8080 with routes from both apps
```

### 3. Shared Routers
The same router can be used in multiple Apps:
```go
healthRouter := createHealthRouter()
app1 := NewApp("app1", ":8080", healthRouter)
app2 := NewApp("app2", ":8081", healthRouter)
// Both apps have health endpoints
```

## Progression Summary

| Level | Concept | Purpose |
|-------|---------|---------|
| **Router** | Routes HTTP requests | Single domain/feature |
| **App** | Combines Routers | Group related features |
| **Server** | Manages Apps | Full application with multiple services |

## Real-World Use Cases

### Development
```go
// All on one port for simplicity
publicApp := NewApp("api", ":8080", userRouter, productRouter, adminRouter)
server.AddApp(publicApp)
```

### Production - Microservices
```go
// Separate services on different ports
userService := NewApp("users", ":8001", userRouter)
productService := NewApp("products", ":8002", productRouter)
server.AddApp(userService)
server.AddApp(productService)
```

### Production - Monolith
```go
// Public API on :80, Admin on :8080 (internal only)
publicApp := NewApp("public", ":80", userRouter, productRouter)
adminApp := NewApp("admin", ":8080", adminRouter)
server.AddApp(publicApp)
server.AddApp(adminApp)
```

## What's Next?
- **04-handlers**: Learn different ways to handle requests (manual, bind, smart bind)
- **05-config**: Learn how to configure Servers/Apps/Routers via YAML (no code changes!)
