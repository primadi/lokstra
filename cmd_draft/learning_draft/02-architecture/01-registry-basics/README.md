# 01-Registry Basics

Understanding Lokstra's registry pattern - the foundation for config-driven architecture.

## Quick Start

```bash
# Run the server
go run .

# Test with test.http file
# OR use curl:
curl http://localhost:8080/health
```

## What is the Registry?

The registry is a **central storage** that holds:

1. **Service Factories** - Templates for creating services
2. **Lazy Services** - Service configurations (created on first access)
3. **Routers** - Named routers for auto-discovery

Think of it as a **dependency injection container** that connects:
- Code (factories) ← Config (YAML) → Services (runtime)

## Key Concepts

### 1. Service Factory

A **function** that converts config → service instance:

```go
func EmailServiceFactory(cfg map[string]any) any {
    return &EmailService{
        SMTPHost: utils.GetValueFromMap(cfg, "smtp_host", "localhost"),
        SMTPPort: utils.GetValueFromMap(cfg, "smtp_port", 587),
        From:     utils.GetValueFromMap(cfg, "from", "noreply@example.com"),
    }
}
```

**Register it:**
```go
old_registry.RegisterServiceFactory("email", EmailServiceFactory)
```

### 2. Lazy Service

**Configuration** stored in registry, service created **only when accessed**:

```go
old_registry.RegisterLazyService("email-service", "email", map[string]any{
    "smtp_host": "smtp.gmail.com",
    "smtp_port": 587,
    "from":      "demo@lokstra.dev",
})
```

- `"email-service"` - Service name (used to get service)
- `"email"` - Factory type (must be registered)
- `map[string]any{...}` - Configuration passed to factory

### 3. Service Container (Best Practice)

**Cache services in struct fields** for fast access:

```go
type ServiceContainer struct {
    emailCache   *EmailService
    counterCache *CounterService
}

func (sc *ServiceContainer) GetEmail() *EmailService {
    // GetService creates on first call, returns cache on subsequent calls
    sc.emailCache = old_registry.GetService("email-service", sc.emailCache)
    return sc.emailCache
}

var services = &ServiceContainer{}
```

**Usage in handler:**
```go
r.POST("/api/email/send", func(c *lokstra.RequestContext) error {
    email := services.GetEmail()  // Fast! Uses cache
    email.SendEmail(to, subject, body)
    return c.Api.Ok("sent")
})
```

### 4. Router Registration

Register routers with **unique names** for config-driven deployment:

```go
old_registry.RegisterRouter("email-api", createEmailRouter())
old_registry.RegisterRouter("counter-api", createCounterRouter())
```

Later in `config.yaml`, reference by name:

```yaml
apps:
  - addr: ":8080"
    routers: [email-api, counter-api, health-api]
```

## Service Lifecycle

```
┌─────────────────┐
│ Register Factory│  RegisterServiceFactory("email", EmailServiceFactory)
│  (Code Setup)   │  
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Register Lazy   │  RegisterLazyService("email-service", "email", config)
│   Service       │  (Stored, NOT created yet)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  First Access   │  services.GetEmail()  →  GetService("email-service", cache)
│                 │  ✓ Creates service using factory + config
│                 │  ✓ Caches in struct field
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Subsequent Use  │  services.GetEmail()  →  Returns cached instance
│                 │  ✓ No factory call
│                 │  ✓ Same instance (stateful)
└─────────────────┘
```

## Registry Methods

### Service Methods

| Method | Description | When to Use |
|--------|-------------|-------------|
| `RegisterServiceFactory(type, factory)` | Register factory for service type | Code setup phase |
| `RegisterLazyService(name, type, config)` | Register service with config | Code setup phase |
| `GetService(name, cache)` | Get service, panic if not found | When service must exist |
| `TryGetService(name, cache)` | Get service, returns (service, bool) | When service is optional |

### Router Methods

| Method | Description | When to Use |
|--------|-------------|-------------|
| `RegisterRouter(name, router)` | Register router with unique name | Code setup phase |
| `TryGetRouter(name)` | Get router by name | Manual app creation |

## Example Walkthrough

### 1. Register Everything

```go
func setupRegistry() {
    // 1. Register factories (templates)
    old_registry.RegisterServiceFactory("email", EmailServiceFactory)
    old_registry.RegisterServiceFactory("counter", CounterServiceFactory)
    
    // 2. Register lazy services (configs)
    old_registry.RegisterLazyService("email-service", "email", map[string]any{
        "smtp_host": "smtp.gmail.com",
        "smtp_port": 587,
        "from":      "demo@lokstra.dev",
    })
    
    old_registry.RegisterLazyService("counter-service", "counter", map[string]any{
        "name": "demo-counter",
        "seed": 100,
    })
    
    // 3. Register routers
    old_registry.RegisterRouter("email-api", createEmailRouter())
    old_registry.RegisterRouter("counter-api", createCounterRouter())
}
```

### 2. Create Service Container

```go
type ServiceContainer struct {
    emailCache   *EmailService
    counterCache *CounterService
}

func (sc *ServiceContainer) GetEmail() *EmailService {
    sc.emailCache = old_registry.GetService("email-service", sc.emailCache)
    return sc.emailCache
}

func (sc *ServiceContainer) GetCounter() *CounterService {
    sc.counterCache = old_registry.GetService("counter-service", sc.counterCache)
    return sc.counterCache
}

var services = &ServiceContainer{}
```

### 3. Use in Handlers

```go
r.POST("/api/email/send", func(c *lokstra.RequestContext) error {
    var req EmailRequest
    if err := c.Req.BindJSON(&req); err != nil {
        return c.Api.BadRequest(err.Error())
    }
    
    // Get service from container (lazy loaded + cached)
    email := services.GetEmail()
    email.SendEmail(req.To, req.Subject, req.Body)
    
    return c.Api.Ok(map[string]any{"status": "sent"})
})
```

## Best Practices

### ✅ DO

1. **Always use ServiceContainer** - Struct fields cache services
2. **Register factories first** - Before registering services
3. **Use descriptive names** - `email-service` not `svc1`
4. **Register routers with unique names** - For config-driven deployment
5. **Use GetService with cache** - Pattern: `cache = GetService(name, cache)`

### ❌ DON'T

1. **Don't call GetService directly in handlers** - Use ServiceContainer
2. **Don't register duplicate names** - Each name must be unique
3. **Don't forget to register factory** - Before RegisterLazyService
4. **Don't use global variables** - Use ServiceContainer pattern

## When to Use Registry?

| Scenario | Use Registry? | Why |
|----------|---------------|-----|
| Simple app (1-2 services) | Optional | Direct instantiation is fine |
| Config-driven deployment | **YES** | Need YAML configuration |
| Multiple environments | **YES** | Different configs per env |
| Microservices architecture | **YES** | Auto-discovery needed |
| Testing with mocks | **YES** | Easy to swap services |

## Next Steps

- **02-config-loading** - Load services from YAML files
- **03-service-dependencies** - Services that depend on other services
- **04-config-driven-deployment** - Complete app from config.yaml

## Common Questions

**Q: Why not just use global variables?**
A: Registry provides:
- Lazy loading (created only when needed)
- Config-driven initialization
- Easy testing (swap implementations)
- Lifecycle management

**Q: When is the service created?**
A: On **first call** to `GetService()`. The factory function runs once, then the result is cached.

**Q: Can I register multiple services of the same type?**
A: YES! Register the factory once, then create multiple services:

```go
old_registry.RegisterServiceFactory("counter", CounterServiceFactory)

old_registry.RegisterLazyService("counter-1", "counter", map[string]any{"seed": 0})
old_registry.RegisterLazyService("counter-2", "counter", map[string]any{"seed": 100})
```

**Q: What if I call GetService before registering the service?**
A: **PANIC!** Always register factories and services during setup phase.
