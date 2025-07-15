# Getting Started with Lokstra

Lokstra is a modular, scalable Go web framework designed to grow with your needs - from simple single-file applications to complex microservices architectures.

## Quick Start

### 1. Simple Application (Single File)

Create a `main.go` file:

```go
package main

import "lokstra"

func main() {
    ctx := lokstra.NewGlobalContext()
    app := lokstra.NewApp(ctx, "my-app", ":8080")
    
    app.GET("/", func(ctx *lokstra.Context) error {
        return ctx.Ok("Hello, Lokstra!")
    })
    
    app.Start()
}
```

Run your application:
```bash
go run main.go
```

### 2. Adding Middleware

```go
package main

import (
    "lokstra"
    "lokstra/middleware/recovery"
    "lokstra/middleware/cors"
)

func main() {
    ctx := lokstra.NewGlobalContext()
    app := lokstra.NewApp(ctx, "my-app", ":8080")
    
    // Register middleware modules
    ctx.RegisterMiddlewareModule(recovery.GetModule())
    ctx.RegisterMiddlewareModule(cors.GetModule())
    
    // Use middleware
    app.Use("lokstra.recovery")
    app.Use("lokstra.cors")
    
    app.GET("/", func(ctx *lokstra.Context) error {
        return ctx.Ok("Hello with middleware!")
    })
    
    app.Start()
}
```

### 3. Adding Services

```go
package main

import (
    "lokstra"
    "lokstra/services/logger"
    "lokstra/services/redis"
)

func main() {
    ctx := lokstra.NewGlobalContext()
    
    // Register service modules
    ctx.RegisterServiceModule(logger.GetModule())
    ctx.RegisterServiceModule(redis.GetModule())
    
    app := lokstra.NewApp(ctx, "my-app", ":8080")
    
    app.GET("/cache/:key", func(ctx *lokstra.Context) error {
        key := ctx.Param("key")
        
        // Get Redis service
        service, err := ctx.GetService("redis")
        if err != nil {
            return ctx.ErrorInternal("Redis not available")
        }
        
        redisService := service.(*redis.RedisService)
        value, err := redisService.Get(ctx.Context(), key)
        if err != nil {
            return ctx.ErrorNotFound("Key not found")
        }
        
        return ctx.Ok(map[string]any{
            "key": key,
            "value": value,
        })
    })
    
    app.Start()
}
```

### 4. YAML Configuration

Create `config/server.yaml`:
```yaml
server:
  name: my-server
  global_setting:
    log_level: info
```

Create `config/apps.yaml`:
```yaml
apps:
  - name: api-app
    address: :8080
    routes:
      - method: GET
        path: /health
        handler: health.check
    middleware:
      - name: recovery
        enabled: true
      - name: cors
        enabled: true
```

Create `config/services.yaml`:
```yaml
services:
  - type: lokstra.logger
    name: default-logger
    config:
      level: info
  - type: lokstra.redis
    name: cache
    config:
      addr: localhost:6379
```

Use configuration in your application:
```go
package main

import (
    "lokstra"
    "fmt"
)

func main() {
    ctx := lokstra.NewGlobalContext()
    
    // Register components
    registerComponents(ctx)
    
    // Load configuration
    cfg, err := lokstra.LoadConfigDir("config")
    if err != nil {
        panic(err)
    }
    
    // Create server from config
    server, err := lokstra.NewServerFromConfig(ctx, cfg)
    if err != nil {
        panic(err)
    }
    
    server.Start()
}

func registerComponents(ctx *lokstra.GlobalContext) {
    // Register handlers
    ctx.RegisterHandler("health.check", func(ctx *lokstra.Context) error {
        return ctx.Ok(map[string]any{"status": "healthy"})
    })
    
    // Register middleware and services
    // ... (register your modules here)
}
```

## Next Steps

- Explore [examples](../cmd/examples/) for more advanced usage patterns
- Read the [Architecture Guide](architecture.md) to understand Lokstra's design
- Learn about [Service Development](service-development.md) to create custom services
- Check out [Middleware Development](middleware-development.md) for custom middleware
- See [Configuration Reference](configuration-reference.md) for complete YAML options

## Key Concepts

- **Apps**: HTTP applications that listen on specific ports
- **Server**: Container for multiple apps with shared services
- **Services**: Reusable components (database, cache, email, etc.)
- **Middleware**: Request/response processing pipeline
- **Modules**: Packaged services and middleware with metadata
- **Configuration**: YAML-based setup for production deployments
