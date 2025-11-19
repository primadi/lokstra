# Services

Lokstra provides a comprehensive collection of built-in services that handle common application needs like database connectivity, caching, and monitoring. All services follow a consistent registration and configuration pattern, making them easy to use and extend.

## Table of Contents

- [Overview](#overview)
- [Service Architecture](#service-architecture)
- [Available Services](#available-services)
- [Quick Start](#quick-start)
- [Service Pattern](#service-pattern)
- [Service Dependencies](#service-dependencies)
- [Configuration](#configuration)
- [Testing Services](#testing-services)
- [Creating Custom Services](#creating-custom-services)
- [Best Practices](#best-practices)

## Overview

Lokstra services are self-contained, reusable components that:
- Follow standardized interfaces defined in `serviceapi`
- Support lazy loading and dependency injection
- Can be configured via YAML or programmatically
- Handle their own lifecycle (initialization and shutdown)
- Are type-safe with Go generics

**Key Features:**

```
✓ Type-Safe Access       - Generic type parameters ensure compile-time safety
✓ Lazy Loading           - Services load only when first accessed
✓ Dependency Injection   - Services can depend on other services
✓ YAML Configuration     - Configure services declaratively
✓ Service Registry       - Centralized service management
✓ Modular Design         - Swap implementations easily
```

## Service Architecture

### Service Categories

```
services/
├── Infrastructure Services      (Redis, PostgreSQL, Metrics)
│   ├── redis                   - Redis client wrapper
│   ├── dbpool_pg               - PostgreSQL connection pooling
│   ├── dbpool_manager          - Centralized pool management
│   ├── kvstore_redis           - Key-value store with Redis
│   └── metrics_prometheus      - Prometheus metrics
│
└── Utilities
    └── register_all.go         - Bulk registration helper
```

### Service Lifecycle

```go
// 1. Registration Phase (startup)
kvstore_redis.Register()  // Registers factory function

// 2. Creation Phase (when referenced)
kvStore := lokstra_registry.NewService[serviceapi.KvStore](
    "my_cache", "kvstore_redis", config)

// 3. Usage Phase (lazy loading)
err := kvStore.Set(ctx, "key", "value", ttl)  // Service loads on first use

// 4. Shutdown Phase (application shutdown)
kvStore.Shutdown()  // Clean up resources
```

## Available Services

### Infrastructure Services

| Service | Type | Interface | Description |
|---------|------|-----------|-------------|
| **[Redis](redis)** | `redis` | (custom) | Redis client wrapper with connection pooling |
| **[DbPool](dbpool-pg)** | `dbpool_pg` | `serviceapi.DbPool` | PostgreSQL connection pool with pgx driver |
| **[DbPool Manager](dbpool-manager)** | `dbpool_manager` | `serviceapi.DbPoolManager` | Centralized pool management with multi-tenancy and named pools |
| **[KvStore](kvstore-redis)** | `kvstore_redis` | `serviceapi.KvStore` | Key-value store with Redis backend and prefix support |
| **[Metrics](metrics-prometheus)** | `metrics_prometheus` | `serviceapi.Metrics` | Prometheus metrics (counters, histograms, gauges) |

## Quick Start

### Using Built-in Services

**Register All Services:**

```go
package main

import (
    "github.com/primadi/lokstra/services"
)

func main() {
    // Option 1: Register all services
    services.RegisterAllServices()
    
    // Option 2: Register by category
    services.RegisterCoreServices()   // Redis, KvStore, Metrics, DbPool
}
```

**Create and Use a Service:**

```go
import (
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/serviceapi"
)

// Create a KvStore service
kvStore := lokstra_registry.NewService[serviceapi.KvStore](
    "cache",              // Service name
    "kvstore_redis",      // Service type
    map[string]any{
        "addr":   "localhost:6379",
        "prefix": "myapp",
    },
)

// Use the service (lazy loads on first access)
ctx := context.Background()
kvStore.Set(ctx, "user:123", userData, 5*time.Minute)

var user User
kvStore.Get(ctx, "user:123", &user)
```

### YAML Configuration

**Define services in your config file:**

```yaml
# config/dev.yaml
services:
  # Redis client
  my_redis:
    type: redis
    config:
      addr: localhost:6379
      db: 0
      pool_size: 10
  
  # KV Store using Redis
  my_cache:
    type: kvstore_redis
    config:
      addr: localhost:6379
      prefix: myapp
      
  # Database pool
  main_db:
    type: dbpool_pg
    config:
      host: localhost
      port: 5432
      database: mydb
      username: postgres
      password: ${DB_PASSWORD}
      max_connections: 20
      
  # Metrics
  metrics:
    type: metrics_prometheus
    config:
      namespace: myapp
      subsystem: api
```

**Services are automatically created when referenced in your application.**

## Service Pattern

All Lokstra services follow this standard pattern:

```go
package myservice

import (
    "github.com/primadi/lokstra/common/utils"
    "github.com/primadi/lokstra/lokstra_registry"
)

const SERVICE_TYPE = "myservice"

// 1. Configuration struct
type Config struct {
    Field1 string `json:"field1" yaml:"field1"`
    Field2 int    `json:"field2" yaml:"field2"`
}

// 2. Service implementation
type myService struct {
    cfg *Config
    // dependencies
}

// 3. Constructor function
func Service(cfg *Config) *myService {
    return &myService{cfg: cfg}
}

// 4. Factory function for registry
func ServiceFactory(params map[string]any) any {
    cfg := &Config{
        Field1: utils.GetValueFromMap(params, "field1", "default"),
        Field2: utils.GetValueFromMap(params, "field2", 42),
    }
    return Service(cfg)
}

// 5. Registration function
func Register() {
    lokstra_registry.RegisterServiceType(SERVICE_TYPE, ServiceFactory, nil)
}
```

**The Three Functions:**

1. **`Service()`** - Constructor that accepts typed configuration
2. **`ServiceFactory()`** - Factory that extracts config from `map[string]any`
3. **`Register()`** - Registers the factory with the service registry

## Service Dependencies

Some services depend on other services. Dependencies are resolved through lazy loading:

```
kvstore_redis
└── redis              (Redis client)

dbpool_manager
└── dbpool_pg          (Creates pools on-demand)
```

**Dependency Injection Example:**

```go
// Service with dependencies
type cacheService struct {
    cfg         *Config
    kvStore     *service.Cached[serviceapi.KvStore]  // Lazy-loaded dependency
    metrics     *service.Cached[serviceapi.Metrics]  // Lazy-loaded dependency
}

func ServiceFactory(params map[string]any) any {
    cfg := extractConfig(params)
    
    // Load dependencies lazily
    kvStore := service.LazyLoad[serviceapi.KvStore](cfg.KvStoreServiceName)
    metrics := service.LazyLoad[serviceapi.Metrics](cfg.MetricsServiceName)
    
    return Service(cfg, kvStore, metrics)
}
```

**Key Points:**

- Dependencies are wrapped in `service.Cached[T]`
- `service.LazyLoad[T]()` creates a lazy reference
- `.MustGet()` resolves and caches the service
- Dependencies load only when first accessed

## Configuration

### Programmatic Configuration

```go
import (
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/services/dbpool_pg"
)

// Register service type
dbpool_pg.Register()

// Create service programmatically
dbPool := lokstra_registry.NewService[any](
    "main_db",      // Service name
    "dbpool_pg",    // Service type
    map[string]any{
        "host":            "localhost",
        "port":            5432,
        "database":        "myapp",
        "username":        "postgres",
        "password":        "password",
        "max_connections": 20,
    },
)
```

### YAML Configuration

```yaml
services:
  main_db:
    type: dbpool_pg
    config:
      host: localhost
      port: 5432
      database: myapp
      username: postgres
      password: ${DB_PASSWORD}      # Environment variable
      max_connections: 20
      max_idle_time: 30m
      max_lifetime: 1h
      
  cache:
    type: kvstore_redis
    config:
      addr: ${REDIS_ADDR:localhost:6379}  # Default value
      prefix: ${APP_NAME}
      db: 0
```

### Environment Variables

Use Lokstra's variable expansion syntax:

- `${VAR_NAME}` - Required variable
- `${VAR_NAME:default}` - Variable with default value
- Works in all string configuration fields

### Configuration Extraction

Use `utils.GetValueFromMap` to extract typed values with defaults:

```go
import "github.com/primadi/lokstra/common/utils"

func ServiceFactory(params map[string]any) any {
    cfg := &Config{
        // String field
        Host: utils.GetValueFromMap(params, "host", "localhost"),
        
        // Integer field
        Port: utils.GetValueFromMap(params, "port", 5432),
        
        // Duration field
        Timeout: utils.GetValueFromMap(params, "timeout", 30*time.Second),
        
        // Nested map
        Options: utils.GetValueFromMap(params, "options", map[string]string{}),
    }
    return Service(cfg)
}
```

## Testing Services

### Unit Testing

```go
package myservice

import (
    "context"
    "testing"
)

func TestService(t *testing.T) {
    // Create service directly
    cfg := &Config{
        Field1: "test",
        Field2: 42,
    }
    svc := Service(cfg)
    
    // Test service methods
    result, err := svc.DoSomething(context.Background())
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    
    if result != expected {
        t.Errorf("expected %v, got %v", expected, result)
    }
}
```

### Integration Testing

```go
func TestServiceWithRegistry(t *testing.T) {
    // Register service
    myservice.Register()
    
    // Create via registry
    svc := lokstra_registry.NewService[MyInterface](
        "test_svc", "myservice",
        map[string]any{
            "field1": "test",
            "field2": 42,
        },
    )
    
    // Test service
    result, err := svc.DoSomething(context.Background())
    // assertions...
}
```

### Mocking Dependencies

```go
type mockDependency struct {
    // mock fields
}

func (m *mockDependency) DoSomething() error {
    return nil // mock implementation
}

func TestServiceWithMock(t *testing.T) {
    cfg := &Config{...}
    mockDep := &mockDependency{}
    
    // Inject mock
    svc := Service(cfg, service.NewCached(mockDep))
    
    // Test with mock
    result, err := svc.UsesDependency()
    // assertions...
}
```

## Creating Custom Services

### 1. Define Interface

```go
// serviceapi/custom/myservice.go
package custom

type MyService interface {
    DoSomething(ctx context.Context, input string) (string, error)
    Shutdown() error
}
```

### 2. Implement Service

```go
// services/myservice/module.go
package myservice

import (
    "context"
    "github.com/primadi/lokstra/common/utils"
    "github.com/primadi/lokstra/lokstra_registry"
    "myapp/serviceapi/custom"
)

const SERVICE_TYPE = "myservice"

type Config struct {
    Option1 string `json:"option1" yaml:"option1"`
    Option2 int    `json:"option2" yaml:"option2"`
}

type myService struct {
    cfg *Config
}

var _ custom.MyService = (*myService)(nil)

func (s *myService) DoSomething(ctx context.Context, input string) (string, error) {
    // Implementation
    return "result", nil
}

func (s *myService) Shutdown() error {
    // Cleanup
    return nil
}

func Service(cfg *Config) *myService {
    return &myService{cfg: cfg}
}

func ServiceFactory(params map[string]any) any {
    cfg := &Config{
        Option1: utils.GetValueFromMap(params, "option1", "default"),
        Option2: utils.GetValueFromMap(params, "option2", 100),
    }
    return Service(cfg)
}

func Register() {
    lokstra_registry.RegisterServiceType(SERVICE_TYPE, ServiceFactory, nil)
}
```

### 3. Register and Use

```go
package main

import (
    "myapp/services/myservice"
    "myapp/serviceapi/custom"
)

func main() {
    // Register
    myservice.Register()
    
    // Create
    svc := lokstra_registry.NewService[custom.MyService](
        "my_svc", "myservice",
        map[string]any{
            "option1": "value",
            "option2": 200,
        },
    )
    
    // Use
    result, err := svc.DoSomething(context.Background(), "input")
}
```

## Best Practices

### Configuration

```go
✓ DO: Provide sensible defaults
cfg := &Config{
    Host: utils.GetValueFromMap(params, "host", "localhost"),
    Port: utils.GetValueFromMap(params, "port", 5432),
}

✗ DON'T: Require all fields
cfg := &Config{
    Host: params["host"].(string),  // Panics if missing
    Port: params["port"].(int),     // Panics if wrong type
}
```

### Dependency Management

```go
✓ DO: Use lazy loading for dependencies
kvStore := service.LazyLoad[serviceapi.KvStore]("cache")
// Loads only when .MustGet() is called

✗ DON'T: Eagerly load dependencies
kvStore := lokstra_registry.GetService[serviceapi.KvStore]("cache")
// May fail if service not yet registered
```

### Error Handling

```go
✓ DO: Return descriptive errors
if user == nil {
    return nil, fmt.Errorf("user not found: %s", username)
}

✓ DO: Wrap errors with context
if err != nil {
    return nil, fmt.Errorf("failed to connect to database: %w", err)
}

✗ DON'T: Panic in service methods
if err != nil {
    panic(err)  // BAD: Let caller handle errors
}
```

### Resource Management

```go
✓ DO: Implement Shutdown method
func (s *myService) Shutdown() error {
    if s.client != nil {
        return s.client.Close()
    }
    return nil
}

✓ DO: Handle shutdown errors gracefully
func (s *myService) Shutdown() error {
    var errs []error
    
    if err := s.dependency1.Shutdown(); err != nil {
        errs = append(errs, err)
    }
    
    if err := s.dependency2.Shutdown(); err != nil {
        errs = append(errs, err)
    }
    
    if len(errs) > 0 {
        return fmt.Errorf("shutdown errors: %v", errs)
    }
    return nil
}
```

### Type Safety

```go
✓ DO: Use interface assertions
var _ serviceapi.KvStore = (*kvStoreService)(nil)  // Compile-time check

✓ DO: Use generics for type-safe access
svc := lokstra_registry.NewService[serviceapi.KvStore]("cache", "kvstore_redis", cfg)
// svc is typed as serviceapi.KvStore

✗ DON'T: Use untyped access
svc := lokstra_registry.NewService[any]("cache", "kvstore_redis", cfg)
// Loses type information
```

### Testing

```go
✓ DO: Test with real config
func TestService(t *testing.T) {
    cfg := &Config{Host: "localhost", Port: 5432}
    svc := Service(cfg)
    // Test with real configuration
}

✓ DO: Test factory function
func TestServiceFactory(t *testing.T) {
    params := map[string]any{
        "host": "localhost",
        "port": 5432,
    }
    svc := ServiceFactory(params)
    // Verify correct type and configuration
}

✓ DO: Mock external dependencies
func TestServiceWithMock(t *testing.T) {
    mockDB := &mockDatabase{}
    svc := Service(cfg, mockDB)
    // Test without real database
}
```

## Related Documentation

**Core Concepts:**
- [Service Registration](../../02-registry/service-registration) - How services are registered
- [Dependency Injection](../../02-registry/service-registration#dependency-injection) - Managing service dependencies
- [Configuration](../03-configuration/config) - YAML configuration system

**Service Documentation:**
- [Infrastructure Services](#infrastructure-services) - Redis, PostgreSQL, Metrics

**Advanced Topics:**
- [Creating Services](../../08-advanced/custom-services) - Building custom services
- [Service Testing](../../08-advanced/testing) - Testing strategies

---

**Next Steps:**
- Learn about [DbPool Service](dbpool-pg) for PostgreSQL connectivity
- Explore [DbPool Manager](dbpool-manager) for multi-tenant database management
- Review [Service Patterns](../../08-advanced/service-patterns) for advanced usage
