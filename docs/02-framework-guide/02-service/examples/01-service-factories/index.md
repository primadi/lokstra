# Service Factories

This example demonstrates various service factory patterns in Lokstra, including simple factories, configuration-based factories, dependency injection, and lifecycle management.

## Overview

Service factories are functions that create and initialize service instances. They provide flexibility in how services are instantiated and configured.

**Topics Covered:**
- Simple factory pattern
- Configuration-based initialization
- Dependency injection
- Lifecycle management (start/stop)
- Factory registration

---

## Factory Patterns

### 1. Simple Factory

Basic service instantiation with minimal configuration:

```go
type SimpleService struct {
    name string
}

func NewSimpleService(name string) *SimpleService {
    return &SimpleService{name: name}
}

func SimpleServiceFactory(deps map[string]any, config map[string]any) any {
    name := "simple-service"
    if nameVal, ok := config["name"].(string); ok {
        name = nameVal
    }
    return NewSimpleService(name)
}
```

**Registration:**
```go
lokstra_registry.RegisterServiceType("simple-service", SimpleServiceFactory, nil)
lokstra_registry.RegisterLazyService("simple-service", SimpleServiceFactory, map[string]any{
    "name": "My Simple Service",
})
```

**Usage:**
```go
svc := lokstra_registry.GetService[*SimpleService]("simple-service")
info := svc.GetInfo()
```

---

### 2. Configurable Factory

Services that read configuration with defaults:

```go
type ConfigurableService struct {
    apiKey     string
    maxRetries int
    timeout    time.Duration
}

func ConfigurableServiceFactory(deps map[string]any, config map[string]any) any {
    // Read with defaults
    apiKey := ""
    if key, ok := config["api_key"].(string); ok {
        apiKey = key
    }

    maxRetries := 3
    if retries, ok := config["max_retries"].(int); ok {
        maxRetries = retries
    }

    timeout := 30 * time.Second
    if timeoutVal, ok := config["timeout_seconds"].(int); ok {
        timeout = time.Duration(timeoutVal) * time.Second
    }

    return NewConfigurableService(apiKey, maxRetries, timeout)
}
```

**Configuration:**
```go
lokstra_registry.RegisterLazyService("configurable-service", ConfigurableServiceFactory, map[string]any{
    "api_key":         "sk-test-12345",
    "max_retries":     5,
    "timeout_seconds": 60,
})
```

**Benefits:**
- ✅ Type-safe configuration reading
- ✅ Default values for optional settings
- ✅ Validation at initialization
- ✅ Environment-specific configuration

---

### 3. Factory with Dependencies

Services that depend on other services:

```go
type DependentService struct {
    cache  *CacheService
    logger *LoggerService
}

func DependentServiceFactory(deps map[string]any, config map[string]any) any {
    cache := deps["cache-service"].(*service.Cached[*CacheService])
    logger := deps["logger-service"].(*service.Cached[*LoggerService])

    return NewDependentService(cache.Get(), logger.Get())
}
```

**Registration with Dependencies:**
```go
lokstra_registry.RegisterLazyServiceWithDeps("dependent-service",
    DependentServiceFactory,
    map[string]string{
        "cache-service":  "cache-service",
        "logger-service": "logger-service",
    },
    nil, nil,
)
```

**Dependency Resolution:**
- Framework resolves dependencies automatically
- Dependencies are instantiated lazily
- Circular dependencies are detected
- Type-safe dependency injection

---

### 4. Lifecycle Management

Services with start/stop lifecycle:

```go
type LifecycleService struct {
    name       string
    startTime  time.Time
    isRunning  bool
    background context.CancelFunc
}

func (s *LifecycleService) Start(ctx context.Context) error {
    if s.isRunning {
        return fmt.Errorf("service already running")
    }

    bgCtx, cancel := context.WithCancel(ctx)
    s.background = cancel
    s.isRunning = true

    // Start background task
    go func() {
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-bgCtx.Done():
                return
            case <-ticker.C:
                // Background work
            }
        }
    }()

    return nil
}

func (s *LifecycleService) Stop() error {
    if s.background != nil {
        s.background()
    }
    s.isRunning = false
    return nil
}
```

**Factory with Initialization:**
```go
func LifecycleServiceFactory(deps map[string]any, config map[string]any) any {
    svc := NewLifecycleService(name)
    
    // Start service immediately
    if err := svc.Start(context.Background()); err != nil {
        log.Printf("Failed to start service: %v", err)
    }
    
    return svc
}
```

**Lifecycle Patterns:**
- Immediate startup in factory
- Background tasks with context
- Graceful shutdown support
- Status monitoring

---

## Factory Signature

All service factories must follow this signature:

```go
func ServiceFactory(deps map[string]any, config map[string]any) any
```

**Parameters:**
- `deps` - Resolved dependencies (map of dependency name → service instance)
- `config` - Configuration map for this service

**Returns:**
- `any` - Service instance (will be type-asserted by consumers)

---

## Best Practices

### 1. Use Constructor Functions

Create separate constructor functions for clarity:

```go
func NewMyService(db *Database, logger *Logger) *MyService {
    return &MyService{
        db:     db,
        logger: logger,
    }
}

func MyServiceFactory(deps map[string]any, config map[string]any) any {
    return NewMyService(
        deps["db"].(*Database),
        deps["logger"].(*Logger),
    )
}
```

### 2. Validate Configuration Early

Fail fast if configuration is invalid:

```go
func ServiceFactory(deps map[string]any, config map[string]any) any {
    apiKey, ok := config["api_key"].(string)
    if !ok || apiKey == "" {
        panic("api_key is required")
    }
    
    return NewService(apiKey)
}
```

### 3. Use Type-Safe Dependencies

Cast dependencies early for type safety:

```go
func ServiceFactory(deps map[string]any, config map[string]any) any {
    // Type-safe casting
    cache, ok := deps["cache"].(*service.Cached[*CacheService])
    if !ok {
        panic("cache dependency not found")
    }
    
    return NewService(cache.Get())
}
```

### 4. Log Initialization

Log service creation for debugging:

```go
func ServiceFactory(deps map[string]any, config map[string]any) any {
    fmt.Printf("✓ Creating MyService with config: %+v\n", config)
    return NewMyService(config)
}
```

---

## Registration Methods

### RegisterServiceType

Register a service type for use in YAML configuration:

```go
lokstra_registry.RegisterServiceType(
    "my-service",           // Type name
    LocalFactory,           // Local factory
    RemoteFactory,          // Remote factory (optional)
)
```

### RegisterLazyService

Register a service instance (no dependencies):

```go
lokstra_registry.RegisterLazyService(
    "service-name",         // Service name
    ServiceFactory,         // Factory function
    map[string]any{         // Configuration
        "key": "value",
    },
)
```

### RegisterLazyServiceWithDeps

Register a service with dependencies:

```go
lokstra_registry.RegisterLazyServiceWithDeps(
    "service-name",         // Service name
    ServiceFactory,         // Factory function
    map[string]string{      // Dependencies (key → service name)
        "db":     "database",
        "cache":  "redis-cache",
    },
    map[string]any{         // Configuration
        "timeout": 30,
    },
    nil,                    // Options (optional)
)
```

---

## Running the Example

```bash
cd docs/02-deep-dive/02-service/examples/01-service-factories
go run main.go
```

**Test Endpoints:**
```bash
# Simple factory
curl http://localhost:3000/simple

# Configurable factory
curl http://localhost:3000/configurable

# Dependent factory (with cache & logger)
curl -X POST http://localhost:3000/process \
  -H "Content-Type: application/json" \
  -d '{"key": "test", "value": "data"}'

# Lifecycle management
curl http://localhost:3000/lifecycle
```

Or use the `test.http` file for interactive testing.

---

## Key Takeaways

1. **Factory Pattern** - Clean separation between construction and usage
2. **Configuration** - Type-safe config reading with defaults
3. **Dependencies** - Framework handles dependency resolution
4. **Lifecycle** - Services can manage their own start/stop logic
5. **Type Safety** - Generic `GetService[T]` for type-safe retrieval

---

## Related Examples

- **[02-remote-services](../02-remote-services/)** - HTTP-based service communication
- **[04-service-composition](../04-service-composition/)** - Layered service patterns
- **[06-testing](../06-testing/)** - Mock service factories for testing

---

**Status**: ✅ Complete
