# Lokstra Learning Path

Progressive examples to master Lokstra framework - from basics to production deployment.

## 📚 Learning Structure

### 01-Basics - Foundation Concepts ✅ COMPLETE

Core building blocks you need before anything else.

| Example | Status | Description |
|---------|--------|-------------|
| [01-router](01-basics/01-router/) | ✅ Complete | Router basics, route registration, path parameters |
| [02-request-context](01-basics/02-request-context/) | ✅ Complete | Context structure, Req/Resp/Api helpers |
| [03-handlers](01-basics/03-handlers/) | ✅ Complete | Request binding, validation, custom validators |
| [04-middleware](01-basics/04-middleware/) | ✅ Complete | Logging, auth, recovery, rate-limiting, CORS |
| [05-services](01-basics/05-services/) | ✅ Complete | Service factory, container, lazy loading, DI |

### 02-Architecture - Config-Driven Patterns ✅ COMPLETE

How to build config-driven applications with the registry pattern.

| Example | Status | Description |
|---------|--------|-------------|
| [01-registry-basics](02-architecture/01-registry-basics/) | ✅ Complete | Service factories, lazy services, router registration |
| [02-config-loading](02-architecture/02-config-loading/) | ✅ Complete | YAML configuration with environment variables |
| [03-service-dependencies](02-architecture/03-service-dependencies/) | ✅ Complete | Services that depend on other services (layered architecture) |
| [04-config-driven-deployment](02-architecture/04-config-driven-deployment/) | ✅ Complete | Complete e-commerce app from YAML config |

### 03-Best-Practices - Production Deployment ✅ COMPLETE

Real-world deployment strategies with same code, different configs.

| Example | Status | Description |
|---------|--------|-------------|
| [01-monolith-single](03-best-practices/01-monolith-single/) | ✅ Complete | Single process, single port (simplest deployment) |
| [02-monolith-multi](03-best-practices/02-monolith-multi/) | ✅ Complete | Single process, multiple ports (logical separation) |
| [03-microservices](03-best-practices/03-microservices/) | ✅ Complete | Multiple services, independent deployment |
| [04-deployment-patterns](03-best-practices/04-deployment-patterns/) | ✅ Complete | Complete comparison and decision guide |

## 🎯 Recommended Learning Path

### Phase 1: Core Concepts (Start Here!)

1. **[01-basics/01-router](01-basics/01-router/)** - HTTP routing fundamentals
   - Route registration and groups
   - Path parameters and query strings
   - Run `go run .` and test with `test.http`

2. **[01-basics/02-request-context](01-basics/02-request-context/)** - Understanding Context (CRITICAL)
   - Req, Resp, Api helpers
   - Request binding patterns
   - This is the foundation for everything

3. **[01-basics/03-handlers](01-basics/03-handlers/)** - Request handling patterns
   - BindAll, BindBody, BindQuery, BindPath
   - Validation with struct tags
   - Custom validators

4. **[01-basics/04-middleware](01-basics/04-middleware/)** - Middleware patterns
   - Logging with timing
   - Authentication and authorization
   - Recovery, rate limiting, CORS
   - Middleware chaining

5. **[01-basics/05-services](01-basics/05-services/)** - Services and DI
   - Service factory pattern
   - ServiceContainer for caching
   - Lazy loading
   - Dependency injection

### Phase 2: Config-Driven Architecture

1. **[02-architecture/01-registry-basics](02-architecture/01-registry-basics/)** - Registry foundation
   - RegisterServiceFactory
   - RegisterLazyService
   - RegisterRouter
   - ServiceContainer pattern

2. **[02-architecture/02-config-loading](02-architecture/02-config-loading/)** - YAML configuration
   - Define services in YAML
   - Environment variables with defaults
   - Config-driven service creation

3. **[02-architecture/03-service-dependencies](02-architecture/03-service-dependencies/)** - Layered architecture
   - Infrastructure layer (DB, Cache)
   - Repository layer (Data access)
   - Domain layer (Business logic)
   - Automatic dependency resolution

4. **[02-architecture/04-config-driven-deployment](02-architecture/04-config-driven-deployment/)** - Complete application
   - 9 services across 3 layers
   - Complete e-commerce API
   - Everything from YAML

### Phase 3: Production Deployment Strategies

1. **[03-best-practices/01-monolith-single](03-best-practices/01-monolith-single/)** - Start simple
   - Single process, single port
   - Zero network overhead
   - Lowest cost

2. **[03-best-practices/02-monolith-multi](03-best-practices/02-monolith-multi/)** - Logical separation
   - Public vs Internal APIs
   - Different middleware per app
   - Still simple deployment

3. **[03-best-practices/03-microservices](03-best-practices/03-microservices/)** - Full independence
   - Separate services
   - Independent deployment
   - Independent scaling

4. **[03-best-practices/04-deployment-patterns](03-best-practices/04-deployment-patterns/)** - Make decisions
   - Complete comparison
   - Decision matrix
   - Migration path

## 🔑 Key Concepts

### request.Context - The Foundation

Every handler receives `*request.Context`:

```go
func MyHandler(c *request.Context) error {
    // c.Req  - Request data (path, query, body, headers)
    // c.Resp - Response building (Json, Html, Text)
    // c.Api  - API responses (Ok, NotFound, BadRequest)
    
    return c.Api.Ok(data)
}
```

**Three Helper Layers:**

1. **Req** - Unopinionated request access
   ```go
   id := c.Req.PathParam("id", "")
   c.Req.BindAll(&input)  // Recommended!
   ```

2. **Resp** - Flexible response building
   ```go
   c.Resp.Json(data)                  // Custom structure
   c.Resp.Html("<h1>Admin</h1>")     // HTML pages
   ```

3. **Api** - Opinionated API responses
   ```go
   c.Api.Ok(user)                     // Success
   c.Api.NotFound("Not found")        // Error
   c.Api.BadRequest("CODE", "msg")    // Detailed error
   ```

### Services

```go
// Factory pattern
dbService := lokstra_registry.NewService[*db.Service](app, "db")

// Registry pattern (shared)
dbService := lokstra_registry.GetService[*db.Service](app, "db")
```

### Middleware

```go
func LoggingMiddleware(c *request.Context) error {
    log.Printf("→ %s %s", c.R.Method, c.R.URL.Path)
    
    err := c.Next()  // Continue chain
    
    log.Printf("← Done")
    return err
}
```

### Config-Driven Deployment

```yaml
servers:
  - name: api
    deployment-id: monolith-single-port  # User-defined string
    port: ${PORT:":8080"}
    routes:
      - path: /api/v1
        router: api_v1
```

## 📖 Documentation

- **[learning.md](learning.md)** - Original learning notes
- **[../docs/](../docs/)** - Comprehensive framework documentation

## 🚀 Quick Start

```bash
# Start with Context (most important!)
cd 01-basics/02-request-context
go run .

# Then explore other examples
cd ../01-router
go run .
```

## 💡 Tips

1. **Start with 02-request-context** - It's the foundation for everything else
2. **Run examples** - See output, modify code, experiment
3. **Read READMEs** - Each example has detailed explanations
4. **Check docs/** - Deeper dives into specific topics
5. **Use BindAll()** - Best way to handle request data

## 🔄 Migration Note

Old structure is being reorganized:
- `02-services/*` → `01-basics/05-services/*`
- `03-middleware/*` → `01-basics/04-middleware/*`

Old examples remain functional during transition.

## Status Legend

- ✅ Complete - Fully implemented with README
- 🔄 Next - Currently being developed
- 📝 Planned - Scheduled for future development
