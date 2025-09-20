# Built-in Services

Lokstra provides a comprehensive set of built-in services located in the `/services` directory, with corresponding service interfaces defined in `/serviceapi`. These services provide essential functionality for database connections, caching, logging, metrics, and health monitoring.

## Service Architecture

### Service Interfaces (`/serviceapi`)

All Lokstra services implement interfaces defined in `/serviceapi` for consistency and testability:

- **`DbPool`**: Database connection pooling interface
- **`KvStore`**: Key-value storage interface  
- **`Logger`**: Structured logging interface
- **`Metrics`**: Application metrics interface
- **`Redis`**: Redis client interface
- **`HealthService`**: Health monitoring interface
- **`I18n`**: Internationalization interface

This interface-based design enables:
- **Dependency injection** through the service container
- **Easy testing** with mock implementations
- **Service swapping** without code changes
- **Type safety** with generics support

## Available Services

### 1. PostgreSQL Database Pool (`dbpool_pg`)

PostgreSQL connection pool service with advanced features.

**Location**: `/services/dbpool_pg/`  
**Interface**: `serviceapi.DbPool`  
**Service Type**: `lokstra.dbpool_pg`

#### Features
- Connection pooling with configurable limits
- SSL/TLS connection support
- Multi-tenant schema support
- Connection health monitoring
- Automatic connection lifecycle management
- Transaction support
- Prepared statement caching

#### Configuration

```yaml
services:
  - name: "main_db"
    type: "lokstra.dbpool_pg"
    config:
      # Option 1: DSN string
      dsn: "postgres://user:password@localhost:5432/mydb?sslmode=require"
      
      # Option 2: Individual parameters
      host: "localhost"
      port: 5432
      database: "myapp"
      username: "postgres"
      password: "secret"
      sslmode: "require"
      
      # Connection pool settings
      min_connections: 2
      max_connections: 20
      max_idle_time: "30m"
      max_lifetime: "1h"
      connect_timeout: "10s"
      query_timeout: "30s"
      
      # Multi-tenant support
      tenant_mode: true
      default_schema: "public"
```

#### Usage Example

```go
// Get database service
dbPool, err := lokstra.GetService[serviceapi.DbPool](regCtx, "main_db")
if err != nil {
    return err
}

// Acquire connection
conn, err := dbPool.Acquire(ctx, "public")
if err != nil {
    return err
}
defer conn.Release()

// Execute query
rows, err := conn.Query(ctx, "SELECT id, name FROM users WHERE active = $1", true)
if err != nil {
    return err
}
defer rows.Close()

// Process results
for rows.Next() {
    var id int
    var name string
    if err := rows.Scan(&id, &name); err != nil {
        return err
    }
    // Process row...
}
```

### 2. In-Memory Key-Value Store (`kvstore_mem`)

Thread-safe in-memory key-value storage with TTL support.

**Location**: `/services/kvstore_mem/`  
**Interface**: `serviceapi.KvStore`  
**Service Type**: `lokstra.kvstore_mem`

#### Features
- Thread-safe concurrent access
- TTL (Time-To-Live) support with automatic cleanup
- Pattern-based key matching
- JSON serialization/deserialization
- Memory-efficient storage
- Bulk operations support

#### Configuration

```yaml
services:
  - name: "cache_memory"
    type: "lokstra.kvstore_mem"
    config:
      cleanup_interval: "5m"     # Cleanup expired keys every 5 minutes
```

#### Usage Example

```go
// Get memory cache service
cache, err := lokstra.GetService[serviceapi.KvStore](regCtx, "cache_memory")
if err != nil {
    return err
}

// Store data with TTL
user := &User{ID: 1, Name: "John Doe"}
err = cache.Set(ctx, "user:1", user, time.Hour)
if err != nil {
    return err
}

// Retrieve data
var retrievedUser User
err = cache.Get(ctx, "user:1", &retrievedUser)
if err != nil {
    return err
}

// Pattern matching
keys, err := cache.Keys(ctx, "user:*")
if err != nil {
    return err
}

// Bulk delete
err = cache.DeleteKeys(ctx, keys...)
if err != nil {
    return err
}
```

### 3. Redis Key-Value Store (`kvstore_redis`)

Redis-based key-value storage service.

**Location**: `/services/kvstore_redis/`  
**Interface**: `serviceapi.KvStore`  
**Service Type**: `lokstra.kvstore_redis`

#### Features
- Redis connection with connection pooling
- TTL support using Redis expiration
- Pattern-based key operations
- Automatic JSON marshaling/unmarshaling
- Distributed caching capabilities
- Redis cluster support

#### Configuration

```yaml
services:
  - name: "cache_redis"
    type: "lokstra.kvstore_redis"
    config:
      redis: "redis_connection"   # Reference to Redis service
```

#### Usage Example

```go
// Same interface as memory cache
cache, err := lokstra.GetService[serviceapi.KvStore](regCtx, "cache_redis")
if err != nil {
    return err
}

// Usage identical to kvstore_mem
err = cache.Set(ctx, "session:abc123", sessionData, time.Hour*24)
```

### 4. Redis Service (`redis`)

Direct Redis client service for advanced Redis operations.

**Location**: `/services/redis/`  
**Interface**: `serviceapi.Redis`  
**Service Type**: `lokstra.redis`

#### Features
- Direct Redis client access
- Connection pooling and health monitoring
- Support for all Redis commands
- Pub/Sub functionality
- Pipeline operations
- Redis Sentinel support

#### Configuration

```yaml
services:
  - name: "redis_main"
    type: "lokstra.redis"
    config:
      addr: "localhost:6379"
      password: ""
      db: 0
      
      # Connection pool settings
      max_idle: 10
      max_active: 100
      idle_timeout: "5m"
      
      # Health monitoring
      ping_interval: "30s"
      max_retries: 3
```

#### Usage Example

```go
// Get Redis service
redis, err := lokstra.GetService[serviceapi.Redis](regCtx, "redis_main")
if err != nil {
    return err
}

// Direct Redis operations
client := redis.Client()

// Set key with expiration
err = client.Set(ctx, "counter", 1, time.Hour).Err()
if err != nil {
    return err
}

// Increment counter
val, err := client.Incr(ctx, "counter").Result()
if err != nil {
    return err
}

// Pub/Sub operations
pubsub := client.Subscribe(ctx, "notifications")
defer pubsub.Close()

for msg := range pubsub.Channel() {
    // Process message
    fmt.Printf("Received: %s\n", msg.Payload)
}
```

### 5. Logger Service (`logger`)

Structured logging service with multiple output formats and log rotation.

**Location**: `/services/logger/`  
**Interface**: `serviceapi.Logger`  
**Service Type**: `lokstra.logger`

#### Features
- Multiple log levels (debug, info, warn, error, fatal, panic)
- Structured logging with fields
- Multiple output formats (JSON, text, console)
- File rotation and archiving
- Context-aware logging
- Performance optimized

#### Configuration

```yaml
services:
  - name: "app_logger"
    type: "lokstra.logger"
    config:
      level: "info"               # debug, info, warn, error, fatal, panic
      format: "json"             # json, text, console
      output: "stdout"           # stdout, stderr, file
      
      # File output settings (when output: file)
      file_path: "./logs/app.log"
      max_size: 100              # MB
      max_backups: 5
      max_age: 30                # days
      compress: true
      
      # Advanced options
      caller: true               # Include caller information
      stacktrace: true           # Include stack trace for errors
```

#### Usage Example

```go
// Get logger service
logger, err := lokstra.GetService[serviceapi.Logger](regCtx, "app_logger")
if err != nil {
    return err
}

// Basic logging
logger.Info("Application started")
logger.Error("Database connection failed", "error", err)

// Structured logging with fields
logger.WithFields(serviceapi.LogFields{
    "user_id": 12345,
    "action":  "login",
    "ip":      "192.168.1.100",
}).Info("User logged in")

// Context-aware logging
logger.WithContext(ctx).Warn("Rate limit exceeded")

// Different log levels
logger.Debug("Debugging information")
logger.Warn("Warning message", "component", "auth")
logger.Error("Error occurred", "error", err, "details", details)
```

### 6. Metrics Service (`metrics`)

Application metrics collection and export service.

**Location**: `/services/metrics/`  
**Interface**: `serviceapi.Metrics`  
**Service Type**: `lokstra.metrics`

#### Features
- Prometheus-compatible metrics
- Counter, Gauge, Histogram, and Summary metrics
- Automatic HTTP endpoint exposure
- Custom metric definitions
- Label support for dimensions
- Built-in application metrics

#### Configuration

```yaml
services:
  - name: "app_metrics"
    type: "lokstra.metrics"
    config:
      enabled: true
      endpoint: "/metrics"       # Prometheus scrape endpoint
      namespace: "lokstra"       # Metric namespace
      subsystem: "app"          # Metric subsystem
      
      # Built-in metrics
      collect_runtime: true      # Go runtime metrics
      collect_process: true      # Process metrics
      collect_http: true         # HTTP request metrics
      
      # Custom metrics configuration
      buckets: [0.1, 0.5, 1.0, 2.5, 5.0, 10.0]  # Histogram buckets
```

#### Usage Example

```go
// Get metrics service
metrics, err := lokstra.GetService[serviceapi.Metrics](regCtx, "app_metrics")
if err != nil {
    return err
}

// Counter metrics
metrics.Counter("api_requests_total", "Total API requests").
    WithLabels(map[string]string{
        "method": "GET",
        "endpoint": "/users",
    }).Inc()

// Gauge metrics
metrics.Gauge("active_connections", "Active database connections").
    Set(float64(activeConnections))

// Histogram metrics
timer := metrics.Histogram("request_duration_seconds", "Request duration").
    WithLabels(map[string]string{
        "method": ctx.Request.Method,
        "status": fmt.Sprintf("%d", statusCode),
    })

start := time.Now()
// ... process request ...
timer.Observe(time.Since(start).Seconds())

// Summary metrics
metrics.Summary("response_size_bytes", "Response size distribution").
    WithLabels(map[string]string{
        "content_type": "application/json",
    }).Observe(float64(responseSize))
```

### 7. Health Check Service (`health_check`)

Comprehensive health monitoring service for application and dependency health.

**Location**: `/services/health_check/`  
**Interface**: `serviceapi.HealthService`  
**Service Type**: `lokstra.health_check`

#### Features
- Multiple health check types (application, database, memory, disk, Redis)
- Concurrent health check execution
- Kubernetes-ready liveness/readiness probes
- Prometheus metrics export
- Configurable timeouts and thresholds
- Extensible with custom health checks

#### Configuration

```yaml
services:
  - name: "health_service"
    type: "lokstra.health_check"
    config:
      timeout: "10s"                    # Overall health check timeout
      
      # Built-in health checks
      checks:
        application:
          enabled: true
        memory:
          enabled: true
          threshold_mb: 1024            # Memory threshold
        disk:
          enabled: true
          path: "/tmp"
          threshold_percent: 80.0       # Disk usage threshold
        database:
          enabled: true
          service: "main_db"           # Database service to check
        redis:
          enabled: true
          service: "redis_main"        # Redis service to check
```

#### Built-in Health Checks

1. **Application Health**: Basic application status and uptime
2. **Memory Health**: Memory usage monitoring with thresholds
3. **Disk Health**: Disk space monitoring for specified paths
4. **Database Health**: Database connectivity and response time
5. **Redis Health**: Redis connectivity and ping response

#### Health Check Endpoints

The service automatically exposes several endpoints:

- `GET /health` - Basic health status
- `GET /health/liveness` - Kubernetes liveness probe
- `GET /health/readiness` - Kubernetes readiness probe
- `GET /health/detailed` - Detailed health information
- `GET /health/metrics` - Prometheus metrics

#### Usage Example

```go
// Get health service
health, err := lokstra.GetService[serviceapi.HealthService](regCtx, "health_service")
if err != nil {
    return err
}

// Check overall health
result := health.CheckHealth(ctx)
if result.Status == serviceapi.HealthStatusHealthy {
    fmt.Println("Application is healthy")
} else {
    fmt.Printf("Application health: %s\n", result.Status)
}

// Register custom health check
health.RegisterCheck("external_api", func(ctx context.Context) serviceapi.HealthCheck {
    // Custom health check logic
    return serviceapi.HealthCheck{
        Name:    "external_api",
        Status:  serviceapi.HealthStatusHealthy,
        Message: "External API is responsive",
    }
})
```

## Service Registration

All built-in services are automatically registered when using the defaults package:

```go
import "github.com/primadi/lokstra/defaults"

func main() {
    regCtx := lokstra.NewGlobalRegistrationContext()
    
    // Register all built-in services
    defaults.RegisterAllServices(regCtx)
    
    // Your app configuration...
}
```

## Service Dependencies

Services can depend on other services through the registration context:

```yaml
services:
  # Redis service
  - name: "redis_main"
    type: "lokstra.redis"
    config:
      addr: "localhost:6379"
  
  # Redis-based cache (depends on Redis service)
  - name: "cache_redis"
    type: "lokstra.kvstore_redis"
    config:
      redis: "redis_main"        # Reference to Redis service
  
  # Health check (depends on database and Redis)
  - name: "health_service"
    type: "lokstra.health_check"
    config:
      checks:
        database:
          service: "main_db"
        redis:
          service: "redis_main"
```

## Service Usage Patterns

### 1. Dependency Injection

```go
func setupHandlers(regCtx lokstra.RegistrationContext) {
    // Inject services into handlers
    regCtx.RegisterHandler("user.create", func(ctx *lokstra.Context) error {
        // Get services
        db, _ := lokstra.GetService[serviceapi.DbPool](ctx.RegistrationContext, "main_db")
        cache, _ := lokstra.GetService[serviceapi.KvStore](ctx.RegistrationContext, "cache_redis")
        logger, _ := lokstra.GetService[serviceapi.Logger](ctx.RegistrationContext, "app_logger")
        
        // Use services...
        return nil
    })
}
```

### 2. Service Composition

```go
type UserService struct {
    db     serviceapi.DbPool
    cache  serviceapi.KvStore
    logger serviceapi.Logger
}

func NewUserService(regCtx lokstra.RegistrationContext) (*UserService, error) {
    db, err := lokstra.GetService[serviceapi.DbPool](regCtx, "main_db")
    if err != nil {
        return nil, err
    }
    
    cache, err := lokstra.GetService[serviceapi.KvStore](regCtx, "cache_redis")
    if err != nil {
        return nil, err
    }
    
    logger, err := lokstra.GetService[serviceapi.Logger](regCtx, "app_logger")
    if err != nil {
        return nil, err
    }
    
    return &UserService{
        db:     db,
        cache:  cache,
        logger: logger,
    }, nil
}

func (s *UserService) GetUser(ctx context.Context, id int) (*User, error) {
    // Try cache first
    var user User
    cacheKey := fmt.Sprintf("user:%d", id)
    if err := s.cache.Get(ctx, cacheKey, &user); err == nil {
        s.logger.Debug("User found in cache", "user_id", id)
        return &user, nil
    }
    
    // Get from database
    conn, err := s.db.Acquire(ctx, "public")
    if err != nil {
        return nil, err
    }
    defer conn.Release()
    
    err = conn.QueryRow(ctx, "SELECT id, name, email FROM users WHERE id = $1", id).
        Scan(&user.ID, &user.Name, &user.Email)
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    s.cache.Set(ctx, cacheKey, &user, time.Hour)
    s.logger.Info("User loaded from database", "user_id", id)
    
    return &user, nil
}
```

### 3. Service Testing

```go
func TestUserService(t *testing.T) {
    // Create mock services
    mockDB := &MockDbPool{}
    mockCache := &MockKvStore{}
    mockLogger := &MockLogger{}
    
    service := &UserService{
        db:     mockDB,
        cache:  mockCache,
        logger: mockLogger,
    }
    
    // Test service methods
    user, err := service.GetUser(context.Background(), 1)
    assert.NoError(t, err)
    assert.Equal(t, "John Doe", user.Name)
}
```

## Best Practices

### 1. Service Configuration

- Use environment variables for sensitive configuration
- Separate configuration by environment (dev/staging/prod)
- Validate configuration on startup
- Use reasonable defaults with override capabilities

### 2. Error Handling

- Always check service availability before use
- Implement fallback mechanisms for critical services
- Log service errors with context
- Use graceful degradation when possible

### 3. Performance

- Configure appropriate connection pools
- Use caching effectively with TTL
- Monitor service performance with metrics
- Implement timeouts for external dependencies

### 4. Security

- Use SSL/TLS for database connections
- Implement proper authentication for Redis
- Don't log sensitive data
- Validate all service configurations

### 5. Monitoring

- Enable health checks for all critical services
- Export metrics for service monitoring
- Set up alerting for service failures
- Monitor resource usage (connections, memory)

## Next Steps

- [Middleware](./built-in-middleware.md) - Learn about built-in middleware
- [Configuration](./configuration.md) - Advanced configuration patterns
- [Schema Reference](./schema.md) - YAML schema documentation
- [Advanced Features](./advanced-features.md) - Production optimization

---

*Built-in services in Lokstra provide enterprise-grade functionality with comprehensive interfaces, robust error handling, and seamless integration through dependency injection.*