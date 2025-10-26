# Strict Route Configuration Examples

## Comparison: Minimal vs Full Configuration

### 🎯 Scenario 1: Minimal Config (Code-Only Middleware)

When you're happy with middleware defined in code, **no router configuration needed**.

#### Code Registration:
```go
// Setup router with middleware in code
apiRouter := router.New("api")
apiRouter.Use(loggingMiddleware, corsMiddleware)  // Router-level

apiRouter.GET("/health", healthHandler, 
    route.WithNameOption("health"))

apiRouter.GET("/users", userHandler, 
    authMiddleware,  // Route-level middleware in code
    route.WithNameOption("users"))

lokstra_registry.RegisterRouter("api", apiRouter)
```

#### YAML Configuration:
```yaml
# Minimal - only services and servers!
services:
  - name: database
    type: postgres
    config: { host: "localhost" }

servers:
  - name: web-server
    services: [database]
    apps:
      - name: web-app
        addr: ":8080"
        routers: [api]  # Uses middleware from code registration
```

**Result**: Router uses exactly the middleware defined in code.

---

### 🎯 Scenario 2: Full Config (Deployment-Specific Middleware)

When you need different middleware per environment.

#### Code Registration:
```go
// Setup router with minimal middleware in code
apiRouter := router.New("api")
// No middleware in code - will be added via config

apiRouter.GET("/health", healthHandler, 
    route.WithNameOption("health"))

apiRouter.GET("/users", userHandler,
    route.WithNameOption("users"))

lokstra_registry.RegisterRouter("api", apiRouter)
```

#### YAML Configuration:
```yaml
middlewares:
  - name: logger
    type: logger
    config: { level: "info" }
  - name: cors
    type: cors
  - name: auth
    type: auth
  - name: rate-limit
    type: rate-limit

services:
  - name: database
    type: postgres
    config: { host: "localhost" }

routers:
  - name: api
    use: [logger, cors]           # Router-level middleware
    routes:
      - name: health
        use: [rate-limit]         # health gets: logger, cors, rate-limit
        
      - name: users
        use: [auth, rate-limit]   # users gets: logger, cors, auth, rate-limit

servers:
  - name: web-server
    services: [database]
    apps:
      - name: web-app
        addr: ":8080" 
        routers: [api]
```

**Result**: Router gets middleware from YAML configuration.

---

### 🎯 Scenario 3: Hybrid (Code + Config Override)

Code provides base middleware, config adds environment-specific layers.

#### Code Registration:
```go
// Base middleware in code
apiRouter := router.New("api")
apiRouter.Use(loggingMiddleware)  // Always present

apiRouter.GET("/health", healthHandler,
    route.WithNameOption("health"))

apiRouter.GET("/users", userHandler, 
    authMiddleware,  // Auth required in code
    route.WithNameOption("users"))

lokstra_registry.RegisterRouter("api", apiRouter)
```

#### Development YAML:
```yaml
middlewares:
  - name: debug
    type: debug

routers:
  - name: api
    use: [debug]                 # Add debug to existing logging
    routes:
      - name: users
        use: []                  # users gets: debug only (no auth!)
        override-parent-mw: true
```

#### Production YAML:
```yaml
middlewares:
  - name: monitoring
    type: monitoring
  - name: rate-limit
    type: rate-limit

routers:
  - name: api
    use: [monitoring, rate-limit]  # Add production middleware
    # No route overrides - keep auth from code
```

---

## Configuration Strategies

### Strategy 1: **Code-Heavy** (Minimal Config)
**Best for**: Small teams, simple deployments, prototype/internal tools

✅ **Pros**:
- Simple configuration files
- Middleware behavior predictable from code
- Less configuration management overhead

❌ **Cons**:
- Harder to customize per environment
- Need code changes for middleware adjustments

### Strategy 2: **Config-Heavy** (Full Config)
**Best for**: Multi-environment deployments, SaaS products, enterprise

✅ **Pros**:
- Complete deployment flexibility
- Different middleware per environment without code changes
- Fine-grained control over routes

❌ **Cons**:
- More complex configuration
- Need to understand both code and config
- More files to maintain

### Strategy 3: **Hybrid** (Base + Override)
**Best for**: Medium/large applications with varying security needs

✅ **Pros**:
- Core functionality guaranteed in code
- Deployment customization available
- Can override security rules per environment

❌ **Cons**:
- Most complex approach
- Need careful coordination between code and config
- Override behavior needs good documentation

---

## Route Configuration Reference

### Route Configuration Fields:
```yaml
routes:
  - name: route-name               # Required: Must match code registration
    use: [middleware1, middleware2] # Optional: Add middleware to route
    enable: true                   # Optional: Default true, set false to disable
    override-parent-mw: false      # Optional: Default false, set true to ignore router middleware
```

### Middleware Inheritance:
```yaml
# Normal inheritance (default behavior)
routers:
  - name: api
    use: [logger, cors]      # Router middleware
    routes:
      - name: users
        use: [auth]          # Route middleware
        # Result: [logger, cors, auth]

# Override inheritance  
routers:
  - name: api
    use: [logger, cors]      # Router middleware
    routes:
      - name: public
        use: [cache]         # Route middleware only
        override-parent-mw: true
        # Result: [cache] only
```

### Environment-Specific Examples:

#### Development:
```yaml
routers:
  - name: api
    use: [debug, cors-permissive]
    routes:
      - name: admin
        use: []              # No auth in development
        override-parent-mw: true
```

#### Production:
```yaml
routers:
  - name: api
    use: [monitoring, rate-limit, cors-strict]
    routes:
      - name: admin
        use: [auth, audit]   # Strong security in production
```

This strict policy makes configuration **predictable**, **safe**, and **focused** on deployment concerns rather than application structure.