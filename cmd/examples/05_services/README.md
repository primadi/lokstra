# Services - Dependency Injection and Service Management

This section demonstrates Lokstra's powerful service system for dependency injection, service lifecycle management, and modular architecture.

## Learning Path

### 01. Basic Services (`01_basic_services/`)
**Foundation**: Service registration, creation, and retrieval
- Service factory registration
- Service instance creation
- Type-safe service retrieval
- Built-in logger service integration
- Custom service implementation

**Key Concepts:**
- `RegisterServiceFactory()` for factory registration
- `CreateService()` for instance creation
- `lokstra.GetService[T]()` for type-safe retrieval
- Service configuration and lifecycle
- Custom service interfaces

**Learning Objectives:**
- Understand service registration patterns
- Learn service factory implementation
- Master type-safe service retrieval
- Explore service configuration options
- Build custom services with proper interfaces

---

## Service System Overview

Lokstra's service system provides:

### 1. **Service Interface**
All services implement the `service.Service` interface:
```go
type Service = any  // Type alias, services can be any type
```

Services typically implement `GetSetting(key string) any` for configuration access.

### 2. **Service Registration**
- **Factory Registration**: `RegisterServiceFactory(name, factory)`
- **Module Registration**: `RegisterModule(moduleFunc)`
- **Direct Registration**: `RegisterService(name, instance)`

### 3. **Service Creation**
- **Create Instance**: `CreateService(factoryName, serviceName, allowReplace, config...)`
- **Get or Create**: `GetOrCreateService(factoryName, serviceName, config...)`
- **Configuration**: Passed to factory functions as `any`

### 4. **Service Retrieval**
- **Basic Retrieval**: `GetService(serviceName)`
- **Type-Safe Retrieval**: `lokstra.GetService[T](ctx, serviceName)`
- **Type Assertions**: For custom service types

### 5. **Service Lifecycle**
- Services are singletons within the application
- Created once and reused across requests
- Configuration handled during creation
- Available throughout application lifetime

---

## Built-in Services

Lokstra provides several built-in services:

### Logger Service
```go
// Register logger module
regCtx.RegisterModule(logger.GetModule)

// Create logger instance
regCtx.CreateService("lokstra.logger", "app-logger", false, "info")

// Use type-safe retrieval
logger, err := lokstra.GetService[serviceapi.Logger](regCtx, "app-logger")
```

### Other Built-in Services
- **Database Pool**: `dbpool_pg.GetModule`
- **Redis**: `redis.GetModule`
- **KV Store**: `kvstore_mem.GetModule`, `kvstore_redis.GetModule`
- **Metrics**: `metrics.GetModule`
- **Health Check**: `health_check.GetModule`

---

## Custom Service Implementation

### Basic Custom Service
```go
type CounterService struct {
    name  string
    count int
}

func (c *CounterService) GetSetting(key string) any {
    switch key {
    case "name":
        return c.name
    case "count":
        return c.count
    default:
        return nil
    }
}

// Register factory
regCtx.RegisterServiceFactory("counter", func(config any) (service.Service, error) {
    name := "default"
    if config != nil {
        if configMap, ok := config.(map[string]any); ok {
            if n, exists := configMap["name"]; exists {
                name = n.(string)
            }
        }
    }
    return &CounterService{name: name, count: 0}, nil
})
```

### Advanced Service Features
- **Configuration Validation**: Validate config in factory functions
- **Dependency Injection**: Services can depend on other services
- **Lifecycle Hooks**: Implement initialization and cleanup
- **Service Groups**: Organize related services in modules
- **Service Discovery**: Dynamic service registration and lookup

---

## Best Practices

### 1. **Service Design**
- Keep services focused and single-purpose
- Implement proper configuration validation
- Use dependency injection for service composition
- Design for testability with interface abstractions

### 2. **Registration Patterns**
- Group related services in modules
- Use descriptive factory and service names
- Validate configuration early in factory functions
- Handle errors gracefully during creation

### 3. **Retrieval Patterns**
- Use type-safe retrieval when possible
- Handle service unavailability gracefully
- Cache service references when appropriate
- Use proper error handling and fallbacks

### 4. **Configuration Management**
- Use structured configuration objects
- Validate configuration completeness
- Provide sensible defaults
- Document configuration options

### 5. **Testing Services**
- Create mock implementations for testing
- Test service factories independently
- Validate service lifecycle behavior
- Test error conditions and edge cases

---

## Common Patterns

### Service with Dependencies
```go
func NewDatabaseService(config any) (service.Service, error) {
    // Get dependent services
    logger, err := lokstra.GetService[serviceapi.Logger](regCtx, "app-logger")
    if err != nil {
        return nil, err
    }
    
    return &DatabaseService{logger: logger}, nil
}
```

### Configuration Validation
```go
func NewAPIService(config any) (service.Service, error) {
    cfg, ok := config.(APIConfig)
    if !ok {
        return nil, errors.New("invalid configuration type")
    }
    
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("configuration validation failed: %w", err)
    }
    
    return &APIService{config: cfg}, nil
}
```

### Service Module
```go
func GetAPIModule() lokstra.Module {
    return lokstra.Module{
        Name: "api-services",
        RegisterFunc: func(regCtx *lokstra.RegistrationContext) error {
            regCtx.RegisterServiceFactory("api-client", NewAPIClientService)
            regCtx.RegisterServiceFactory("api-cache", NewAPICacheService)
            regCtx.RegisterServiceFactory("api-auth", NewAPIAuthService)
            return nil
        },
    }
}
```

---

## Next Steps

After mastering basic services, explore:

1. **Dependency Injection Patterns** - Complex service composition
2. **Built-in Services Integration** - Database, Redis, metrics
3. **Custom Service Modules** - Organizing service groups
4. **Service Testing Strategies** - Mocking and testing patterns
5. **Advanced Service Features** - Lifecycle hooks, discovery, monitoring

---

## Running Examples

Each example includes:
- **Comprehensive Documentation**: Inline comments and concepts
- **Test Commands**: Curl commands for API testing
- **Learning Objectives**: Clear goals for each example
- **Key Concepts**: Essential patterns and practices

```bash
# Navigate to any example directory
cd 01_basic_services/

# Run the example
go run main.go

# Test the endpoints (from example documentation)
curl http://localhost:8080/
curl http://localhost:8080/services
curl -X POST http://localhost:8080/counters/request/increment
```

Each example builds upon previous concepts while introducing new patterns and advanced features.
</content>