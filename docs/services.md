# Services

Lokstra's service system provides a powerful dependency injection container that manages the lifecycle of your application components. Services are registered by name and can be retrieved throughout your application using type-safe methods.

## Understanding Services

Services in Lokstra are components that implement the `service.Service` interface and provide specific functionality like database connections, logging, metrics, caching, and more. The service system supports:

- **Factory-based registration** - Services are created on-demand using factory functions
- **Type-safe retrieval** - Get services with compile-time type checking
- **Configuration support** - Pass configuration data to service factories
- **Lifecycle management** - Services are created once and reused
- **Module system** - Group related services into reusable modules

## Service Interface

All services must implement the basic service interface:

```go
type Service interface {
    GetSetting(key string) any
}
```

The `GetSetting` method allows services to expose configuration or internal state.

### Optional Interfaces

Services can implement optional interfaces for additional functionality:

#### Shutdownable Interface

Services that need to perform cleanup tasks when the application shuts down can implement the `service.Shutdownable` interface:

```go
type Shutdownable interface {
    Shutdown() error
}
```

Services implementing this interface will be gracefully shut down when the server stops, allowing them to release resources, close connections, or save state. The shutdown process runs in parallel for better performance.

## Service Registration

### Register Service Factories

Service factories create service instances from configuration:

```go
// Basic factory registration
regCtx.RegisterServiceFactory("my-cache", func(config any) (service.Service, error) {
    return &MemoryCache{}, nil
})

// Factory with configuration
regCtx.RegisterServiceFactory("database", func(config any) (service.Service, error) {
    cfg := config.(map[string]any)
    dsn := cfg["dsn"].(string)
    
    return NewDatabaseConnection(dsn)
})
```

### Register Service Modules

Modules group related services and factories:

```go
// Register all default services
defaults.RegisterAllServices(regCtx)

// Register individual modules
regCtx.RegisterModule(logger.GetModule)
regCtx.RegisterModule(dbpool_pg.GetModule) 
regCtx.RegisterModule(redis.GetModule)
```

### Direct Service Registration

Register service instances directly:

```go
logger := &MyLogger{}
regCtx.RegisterService("my-logger", logger, false) // allowReplace = false
```

### Create Service Instances

Create services from registered factories:

```go
// Create a new service instance (fails if already exists)
dbPool, err := regCtx.CreateService("lokstra.dbpool_pg", "db.main", false, config)

// Create or replace an existing service instance
logger, err := regCtx.CreateService("lokstra.logger", "app-logger", true, "info")

// Create multiple services with same factory
cache1, err := regCtx.CreateService("lokstra.redis", "cache.session", false, "redis://localhost:6379/0")
cache2, err := regCtx.CreateService("lokstra.redis", "cache.user", false, "redis://localhost:6379/1")
```

### Get or Create Service

Get an existing service or create it if it doesn't exist:

```go
// Get existing service or create with factory
dbPool, err := regCtx.GetOrCreateService("lokstra.dbpool_pg", "db.main", config)

// Multiple config parameters
service, err := regCtx.GetOrCreateService("my-factory", "my-service", param1, param2, param3)
```

## Service Retrieval

### Type-Safe Service Retrieval

Get services with compile-time type checking:

```go
// Using the serviceapi helper
dbPool, err := serviceapi.GetService[serviceapi.DbPool](regCtx, "db.main")
if err != nil {
    return err
}

// Using the lokstra helper 
logger, err := lokstra.GetService[serviceapi.Logger](regCtx, "logger")
if err != nil {
    return err
}
```

### Create or Get Service

Create a service if it doesn't exist, or return the existing one:

```go
// Using the registration context
dbPool, err := regCtx.GetOrCreateService("lokstra.dbpool_pg", "db.main", config)

// Using the lokstra helper with type safety
logger, err := lokstra.GetOrCreateService[serviceapi.Logger](
    regCtx, "logger", "lokstra.logger", "info")
```

### Raw Service Retrieval

Get services without type checking:

```go
service, err := regCtx.GetService("my-service")
if err != nil {
    return err
}

// Type assertion required
logger := service.(serviceapi.Logger)
```

## Built-in Services

Lokstra provides several built-in service implementations:

### Database Pool Service

PostgreSQL connection pool with schema support:

```go
// Register the module
regCtx.RegisterModule(dbpool_pg.GetModule)

// Create with DSN string
dbPool, err := regCtx.CreateService("lokstra.dbpool_pg", "db.main", false,
    "postgres://user:pass@localhost/mydb")

// Create with configuration map
config := map[string]any{
    "host":     "localhost",
    "port":     5432,
    "database": "mydb",
    "username": "user",
    "password": "pass",
    "max_connections": 20,
    "min_connections": 5,
}
dbPool, err := regCtx.CreateService("lokstra.dbpool_pg", "db.main", false, config)

// Use the service
db, err := serviceapi.GetService[serviceapi.DbPool](regCtx, "db.main")
conn, err := db.Acquire(ctx, "public")
defer conn.Release()
```

### Logger Service

Structured logging with multiple output formats:

```go
// Register the module
regCtx.RegisterModule(logger.GetModule)

// Create with log level
logger, err := regCtx.CreateService("lokstra.logger", "app-logger", false, "info")

// Create with configuration
config := map[string]any{
    "level":  "debug",
    "format": "json",
    "output": "stdout",
}
logger, err := regCtx.CreateService("lokstra.logger", "app-logger", false, config)

// Use the service
log, err := serviceapi.GetService[serviceapi.Logger](regCtx, "app-logger")
log.Info("Application started")
log.Error("Something went wrong", "error", err)
```

### Redis Service

Redis client with connection pooling:

```go
// Register the module
regCtx.RegisterModule(redis.GetModule)

// Create with connection string
redis, err := regCtx.CreateService("lokstra.redis", "cache", false,
    "redis://localhost:6379/0")

// Use the service
cache, err := serviceapi.GetService[serviceapi.Redis](regCtx, "cache")
err = cache.Set(ctx, "key", "value", time.Hour)
value, err := cache.Get(ctx, "key")
```

### Key-Value Store Services

In-memory and Redis-backed key-value stores:

```go
// In-memory store
regCtx.RegisterModule(kvstore_mem.GetModule)
memStore, err := regCtx.CreateService("lokstra.kvstore_mem", "session-store", false, nil)

// Redis store
regCtx.RegisterModule(kvstore_redis.GetModule)
redisStore, err := regCtx.CreateService("lokstra.kvstore_redis", "user-cache", false,
    "redis://localhost:6379/1")

// Use the services
kv, err := serviceapi.GetService[serviceapi.KvStore](regCtx, "session-store")
err = kv.Set(ctx, "session:123", userData, time.Hour)
data, err := kv.Get(ctx, "session:123")
```

### Metrics Service

Application metrics collection:

```go
// Register the module
regCtx.RegisterModule(metrics.GetModule)

// Create metrics service
metrics, err := regCtx.CreateService("lokstra.metrics", "app-metrics", false, nil)

// Use the service
m, err := serviceapi.GetService[serviceapi.Metrics](regCtx, "app-metrics")
m.Counter("requests_total").Inc()
m.Histogram("request_duration").Observe(duration.Seconds())
```

### Health Check Service

Health monitoring for your application:

```go
// Register the module
regCtx.RegisterModule(health_check.GetModule)

// Create health check service
health, err := regCtx.CreateService("lokstra.health", "app-health", false, nil)

// Use the service
hc, err := serviceapi.GetService[serviceapi.HealthCheck](regCtx, "app-health")
hc.AddCheck("database", func(ctx context.Context) error {
    return dbPool.Ping(ctx)
})
```

## Service Configuration

### Configuration Types

Services accept various configuration types:

```go
// String configuration
service, err := regCtx.CreateService("my-service", "instance", false, "simple-config")

// Map configuration
config := map[string]any{
    "host": "localhost",
    "port": 8080,
    "ssl":  true,
}
service, err := regCtx.CreateService("my-service", "instance", false, config)

// Struct configuration
type MyConfig struct {
    Host string `json:"host"`
    Port int    `json:"port"`
}
config := MyConfig{Host: "localhost", Port: 8080}
service, err := regCtx.CreateService("my-service", "instance", false, config)
```

### Configuration Helpers

Use utility functions for common configuration patterns:

```go
// Get service from configuration
dbPool, err := registration.GetServiceFromConfig[serviceapi.DbPool](
    regCtx, config, "database_service")
```

## Custom Services

### Creating Custom Services

Implement the service interface:

```go
type EmailService struct {
    smtpHost string
    smtpPort int
    username string
    password string
}

func (e *EmailService) GetSetting(key string) any {
    switch key {
    case "smtp_host":
        return e.smtpHost
    case "smtp_port":
        return e.smtpPort
    default:
        return nil
    }
}

func (e *EmailService) SendEmail(to, subject, body string) error {
    // Implementation here
    return nil
}

// Implement the optional Shutdownable interface for graceful shutdown
func (e *EmailService) Shutdown() error {
    // Close SMTP connection pool
    // Wait for pending emails to finish
    // Clean up resources
    return nil
}
```

### Register Custom Service Factory

```go
func emailServiceFactory(config any) (service.Service, error) {
    cfg := config.(map[string]any)
    
    return &EmailService{
        smtpHost: cfg["host"].(string),
        smtpPort: cfg["port"].(int),
        username: cfg["username"].(string),
        password: cfg["password"].(string),
    }, nil
}

regCtx.RegisterServiceFactory("email", emailServiceFactory)
```

### Create Custom Module

Group related services into modules:

```go
type EmailModule struct{}

func (m *EmailModule) Name() string {
    return "email-services"
}

func (m *EmailModule) Description() string {
    return "Email sending services"
}

func (m *EmailModule) Register(regCtx registration.Context) error {
    // Register email service
    regCtx.RegisterServiceFactory("email.smtp", emailServiceFactory)
    
    // Register additional email-related services
    regCtx.RegisterServiceFactory("email.template", templateServiceFactory)
    
    return nil
}

func GetEmailModule() registration.Module {
    return &EmailModule{}
}

// Register the module
regCtx.RegisterModule(GetEmailModule)
```

## Service Usage in Handlers

### Dependency Injection in Handlers

Access services in your HTTP handlers:

```go
func createUserHandler(ctx *lokstra.Context) error {
    // Get database service
    dbPool, err := serviceapi.GetService[serviceapi.DbPool](ctx.RegistrationContext, "db.main")
    if err != nil {
        return ctx.ErrorInternal("Database unavailable")
    }
    
    // Get logger service
    logger, err := serviceapi.GetService[serviceapi.Logger](ctx.RegistrationContext, "logger")
    if err != nil {
        return ctx.ErrorInternal("Logger unavailable")
    }
    
    // Business logic
    conn, err := dbPool.Acquire(ctx.Context, "public")
    if err != nil {
        logger.Error("Failed to acquire database connection", "error", err)
        return ctx.ErrorInternal("Database connection failed")
    }
    defer conn.Release()
    
    // Create user...
    logger.Info("User created successfully", "user_id", userID)
    
    return ctx.OkCreated(user)
}
```

### Service Caching Pattern

Cache frequently used services:

```go
type UserHandler struct {
    dbPool serviceapi.DbPool
    logger serviceapi.Logger
    cache  serviceapi.KvStore
}

func NewUserHandler(regCtx registration.Context) (*UserHandler, error) {
    dbPool, err := serviceapi.GetService[serviceapi.DbPool](regCtx, "db.main")
    if err != nil {
        return nil, err
    }
    
    logger, err := serviceapi.GetService[serviceapi.Logger](regCtx, "logger")
    if err != nil {
        return nil, err
    }
    
    cache, err := serviceapi.GetService[serviceapi.KvStore](regCtx, "user-cache")
    if err != nil {
        return nil, err
    }
    
    return &UserHandler{
        dbPool: dbPool,
        logger: logger,
        cache:  cache,
    }, nil
}

func (h *UserHandler) CreateUser(ctx *lokstra.Context) error {
    // Use cached services
    conn, err := h.dbPool.Acquire(ctx.Context, "public")
    if err != nil {
        h.logger.Error("Database connection failed", "error", err)
        return ctx.ErrorInternal("Database unavailable")
    }
    defer conn.Release()
    
    // Business logic...
    
    return ctx.OkCreated(user)
}
```

## Multi-Tenant Services

### Tenant-Specific Database Pools

Lokstra supports tenant-specific services:

```go
import "github.com/primadi/lokstra/common/tenant_dbpool"

// Register tenant DSNs
tenant_dbpool.RegisterTenantDSN("tenant1", "postgres://user:pass@db1/tenant1")
tenant_dbpool.RegisterTenantDSN("tenant2", "postgres://user:pass@db2/tenant2")

// Get tenant-specific database pool
func getUsersHandler(ctx *lokstra.Context) error {
    tenantID := ctx.GetHeader("X-Tenant-ID")
    
    dbPool, exists := tenant_dbpool.GetTenantDbPool(tenantID)
    if !exists {
        return ctx.ErrorBadRequest("Invalid tenant")
    }
    
    conn, err := dbPool.Acquire(ctx.Context, "public")
    if err != nil {
        return ctx.ErrorInternal("Database unavailable")
    }
    defer conn.Release()
    
    // Query tenant-specific data...
    
    return ctx.Ok(users)
}
```

## Default Services

### Setting Default Services

Set global default services for convenience:

```go
import "github.com/primadi/lokstra/core/flow"

// Set default services
flow.SetDefaultDbPool(regCtx, "db.main")
flow.SetDefaultLogger(regCtx, "logger")
flow.SetDefaultMetrics(regCtx, "metrics")

// Or set directly
dbPool, _ := serviceapi.GetService[serviceapi.DbPool](regCtx, "db.main")
flow.SetDefaultDbPoolService(dbPool)
```

### Using Default Services

Access default services without explicit dependency injection:

```go
// In flow steps, default services are automatically available
flow := lokstra.NewFlow[MyData](regCtx).
    Step(func(ctx *flow.Context[MyData]) error {
        // Default database pool is available
        conn, err := ctx.DbPool.Acquire(ctx.Context, "public")
        if err != nil {
            return err
        }
        defer conn.Release()
        
        // Business logic...
        return nil
    })
```

## Service Testing

### Mock Services for Testing

Create mock services for unit testing:

```go
type MockDbPool struct {
    connections map[string]*MockConn
}

func (m *MockDbPool) GetSetting(key string) any {
    return nil
}

func (m *MockDbPool) Acquire(ctx context.Context, schema string) (serviceapi.DbConn, error) {
    return m.connections[schema], nil
}

// Test setup
func TestUserHandler(t *testing.T) {
    regCtx := lokstra.NewRegistrationContext()
    
    // Register mock services
    mockDb := &MockDbPool{connections: make(map[string]*MockConn)}
    regCtx.RegisterService("db.main", mockDb)
    
    mockLogger := &MockLogger{}
    regCtx.RegisterService("logger", mockLogger)
    
    // Test handler
    handler := NewUserHandler(regCtx)
    // ... test logic
}
```

## Best Practices

### 1. Service Naming

Use consistent naming conventions:

```go
// Good: descriptive, hierarchical names
"db.main"           // Main database
"db.analytics"      // Analytics database  
"cache.session"     // Session cache
"cache.user"        // User cache
"logger.app"        // Application logger
"logger.audit"      // Audit logger
```

### 2. Configuration Management

Structure configuration clearly:

```go
type DatabaseConfig struct {
    Host            string `json:"host"`
    Port            int    `json:"port"`
    Database        string `json:"database"`
    Username        string `json:"username"`
    Password        string `json:"password"`
    MaxConnections  int    `json:"max_connections"`
    MinConnections  int    `json:"min_connections"`
    ConnectTimeout  string `json:"connect_timeout"`
}

func dbFactory(config any) (service.Service, error) {
    var cfg DatabaseConfig
    
    switch c := config.(type) {
    case string:
        // Parse DSN
        cfg = parseDSN(c)
    case map[string]any:
        // Map to struct
        cfg = mapToStruct(c)
    case DatabaseConfig:
        cfg = c
    default:
        return nil, errors.New("invalid configuration type")
    }
    
    return NewDatabase(cfg)
}
```

### 3. Error Handling

Handle service errors gracefully:

```go
func getDBPool(regCtx registration.Context) (serviceapi.DbPool, error) {
    dbPool, err := serviceapi.GetService[serviceapi.DbPool](regCtx, "db.main")
    if err != nil {
        // Fallback to default or return error
        return nil, fmt.Errorf("database service unavailable: %w", err)
    }
    return dbPool, nil
}
```

### 4. Service Lifecycle

Consider service initialization order:

```go
func setupServices(regCtx registration.Context) error {
    // Register modules first
    defaults.RegisterAllServices(regCtx)
    
    // Create core services
    _, err := regCtx.CreateService("lokstra.logger", "logger", false, "info")
    if err != nil {
        return err
    }
    
    // Create dependent services
    _, err = regCtx.CreateService("lokstra.dbpool_pg", "db.main", false, dbConfig)
    if err != nil {
        return err
    }
    
    // Set defaults
    flow.SetDefaultLogger(regCtx, "logger")
    flow.SetDefaultDbPool(regCtx, "db.main")
    
    return nil
}
```

## Troubleshooting

### Common Issues

1. **Service not found**
   ```
   service 'db.main' not found
   ```
   - Ensure the service factory is registered
   - Verify the service name is correct
   - Check if the service was created

2. **Type mismatch**
   ```
   service 'logger' is not of type serviceapi.Logger
   ```
   - Verify the service implements the expected interface
   - Check type assertions in service factories

3. **Configuration errors**
   ```
   invalid configuration type
   ```
   - Ensure configuration matches factory expectations
   - Validate configuration structure and types

### Debugging Services

```go
// List all registered services
services := regCtx.GetAllServices()
for name, service := range services {
    fmt.Printf("Service: %s, Type: %T\n", name, service)
}

// List all service factories
factories := regCtx.GetAllServiceFactories()
for name := range factories {
    fmt.Printf("Factory: %s\n", name)
}
```

## Next Steps

- [Core Concepts](./core-concepts.md) - Understand request/response handling
- [Configuration](./configuration.md) - Learn about YAML configuration
- [Middleware](./middleware.md) - Integrate services with middleware
- [Advanced Features](./advanced-features.md) - Service monitoring and optimization

---

*The service system is the foundation of dependency management in Lokstra. Master these patterns to build maintainable, testable applications with proper separation of concerns.*