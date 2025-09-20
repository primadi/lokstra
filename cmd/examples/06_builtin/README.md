# Built-in Services - Lokstra Framework Services

This section demonstrates comprehensive usage of Lokstra's built-in services for common application needs like logging, database connectivity, caching, metrics, and more.

## Learning Path

### 01. Logger Service (`01_logger_service/`)
**Foundation**: Comprehensive logging with multiple configurations
- Multiple logger instances with different levels
- Text and JSON logging formats
- Structured logging patterns
- Logger service lifecycle
- Performance considerations

**Key Concepts:**
- Logger configuration options
- Log level hierarchy (debug, info, warn, error, fatal)
- Structured vs. simple logging
- Multiple logger instances
- Service integration patterns

**Learning Objectives:**
- Master logger service configuration
- Understand logging level usage
- Implement structured logging
- Optimize logging performance
- Use loggers across application layers

---

## Built-in Services Overview

Lokstra provides comprehensive built-in services for common application needs:

### 1. **Logger Service** (`services/logger`)
```go
// Register logger module
regCtx.RegisterModule(logger.GetModule)

// Create logger with string config
regCtx.CreateService("lokstra.logger", "app-logger", false, "info")

// Create logger with detailed config
loggerConfig := map[string]any{
    "level":  "debug",
    "format": "json",
}
regCtx.CreateService("lokstra.logger", "debug-logger", false, loggerConfig)
```

**Features:**
- Multiple log levels (debug, info, warn, error, fatal)
- Text and JSON output formats
- Multiple logger instances
- Thread-safe logging
- Structured logging support

### 2. **Database Pool Service** (`services/dbpool_pg`)
PostgreSQL connection pooling with transaction management:
```go
regCtx.RegisterModule(dbpool_pg.GetModule)

dbConfig := map[string]any{
    "host":     "localhost",
    "port":     5432,
    "database": "myapp",
    "user":     "postgres",
    "password": "password",
}
regCtx.CreateService("lokstra.dbpool_pg", "main-db", false, dbConfig)
```

### 3. **Redis Service** (`services/redis`)
Redis connectivity with connection pooling:
```go
regCtx.RegisterModule(redis.GetModule)

redisConfig := map[string]any{
    "addr":     "localhost:6379",
    "password": "",
    "db":       0,
}
regCtx.CreateService("lokstra.redis", "cache-redis", false, redisConfig)
```

### 4. **KV Store Services**
In-memory and Redis-backed key-value stores:
```go
// In-memory KV store
regCtx.RegisterModule(kvstore_mem.GetModule)
regCtx.CreateService("lokstra.kvstore_mem", "memory-store", false, nil)

// Redis-backed KV store
regCtx.RegisterModule(kvstore_redis.GetModule)
regCtx.CreateService("lokstra.kvstore_redis", "persistent-store", false, redisConfig)
```

### 5. **Metrics Service** (`services/metrics`)
Application metrics collection and reporting:
```go
regCtx.RegisterModule(metrics.GetModule)
regCtx.CreateService("lokstra.metrics", "app-metrics", false, nil)
```

### 6. **Health Check Service** (`services/health_check`)
Health monitoring and status reporting:
```go
regCtx.RegisterModule(health_check.GetModule)
regCtx.CreateService("lokstra.health_check", "app-health", false, nil)
```

---

## Service Configuration Patterns

### String Configuration
Simple services often accept string configuration:
```go
// Logger with level string
regCtx.CreateService("lokstra.logger", "simple-logger", false, "info")

// Database with connection string
regCtx.CreateService("lokstra.dbpool_pg", "db", false, "postgres://user:pass@localhost/db")
```

### Map Configuration
Complex services use map configuration:
```go
config := map[string]any{
    "host":            "localhost",
    "port":            5432,
    "database":        "myapp",
    "user":            "postgres",
    "password":        "secret",
    "max_connections": 20,
    "ssl_mode":        "prefer",
}
regCtx.CreateService("lokstra.dbpool_pg", "main-db", false, config)
```

### Environment-based Configuration
Use environment variables for configuration:
```go
dbConfig := map[string]any{
    "host":     os.Getenv("DB_HOST"),
    "port":     os.Getenv("DB_PORT"),
    "database": os.Getenv("DB_NAME"),
    "user":     os.Getenv("DB_USER"),
    "password": os.Getenv("DB_PASSWORD"),
}
```

---

## Service Integration Patterns

### Type-Safe Service Retrieval
```go
// Built-in service types
logger, err := lokstra.GetService[serviceapi.Logger](regCtx, "app-logger")
dbPool, err := lokstra.GetService[serviceapi.DbPool](regCtx, "main-db")
redis, err := lokstra.GetService[serviceapi.Redis](regCtx, "cache-redis")
kvStore, err := lokstra.GetService[serviceapi.KvStore](regCtx, "memory-store")
```

### Service Dependency Injection
Services can depend on other services:
```go
func NewCustomService(regCtx *lokstra.RegistrationContext) (service.Service, error) {
    // Get dependent services
    logger, err := lokstra.GetService[serviceapi.Logger](regCtx, "app-logger")
    if err != nil {
        return nil, err
    }
    
    dbPool, err := lokstra.GetService[serviceapi.DbPool](regCtx, "main-db")
    if err != nil {
        return nil, err
    }
    
    return &CustomService{
        logger: logger,
        db:     dbPool,
    }, nil
}
```

### Handler Integration
Use services in HTTP handlers:
```go
app.GET("/users/:id", func(ctx *lokstra.Context) error {
    logger, err := lokstra.GetService[serviceapi.Logger](regCtx, "app-logger")
    if err != nil {
        return ctx.ErrorInternal("Logger unavailable")
    }
    
    dbPool, err := lokstra.GetService[serviceapi.DbPool](regCtx, "main-db")
    if err != nil {
        return ctx.ErrorInternal("Database unavailable")
    }
    
    userID := ctx.GetPathParam("id")
    logger.Infof("Fetching user: %s", userID)
    
    // Use database
    user, err := fetchUser(dbPool, userID)
    if err != nil {
        logger.Errorf("Error fetching user %s: %v", userID, err)
        return ctx.ErrorInternal("Database error")
    }
    
    return ctx.Ok(user)
})
```

---

## Configuration Best Practices

### 1. **Environment-based Configuration**
```go
// Use environment variables with defaults
config := map[string]any{
    "level":  getEnv("LOG_LEVEL", "info"),
    "format": getEnv("LOG_FORMAT", "text"),
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

### 2. **Configuration Validation**
```go
func validateLoggerConfig(config map[string]any) error {
    level, ok := config["level"].(string)
    if !ok {
        return errors.New("level must be string")
    }
    
    validLevels := []string{"debug", "info", "warn", "error", "fatal"}
    if !contains(validLevels, level) {
        return fmt.Errorf("invalid level: %s", level)
    }
    
    return nil
}
```

### 3. **Service Health Checks**
```go
func checkServiceHealth(regCtx *lokstra.RegistrationContext) error {
    // Check critical services
    if _, err := lokstra.GetService[serviceapi.Logger](regCtx, "app-logger"); err != nil {
        return fmt.Errorf("logger service unavailable: %w", err)
    }
    
    if _, err := lokstra.GetService[serviceapi.DbPool](regCtx, "main-db"); err != nil {
        return fmt.Errorf("database service unavailable: %w", err)
    }
    
    return nil
}
```

### 4. **Graceful Service Handling**
```go
app.GET("/status", func(ctx *lokstra.Context) error {
    status := map[string]string{
        "app": "healthy",
    }
    
    // Check optional services gracefully
    if logger, err := lokstra.GetService[serviceapi.Logger](regCtx, "app-logger"); err == nil {
        status["logger"] = "available"
        logger.Debugf("Status check performed")
    } else {
        status["logger"] = "unavailable"
    }
    
    return ctx.Ok(status)
})
```

---

## Service Lifecycle Management

### Registration Phase
```go
// Register all service modules
regCtx.RegisterModule(logger.GetModule)
regCtx.RegisterModule(dbpool_pg.GetModule)
regCtx.RegisterModule(redis.GetModule)
regCtx.RegisterModule(kvstore_mem.GetModule)
```

### Creation Phase
```go
// Create service instances
services := []struct {
    factory string
    name    string
    config  any
}{
    {"lokstra.logger", "app-logger", "info"},
    {"lokstra.dbpool_pg", "main-db", dbConfig},
    {"lokstra.redis", "cache-redis", redisConfig},
}

for _, svc := range services {
    if _, err := regCtx.CreateService(svc.factory, svc.name, false, svc.config); err != nil {
        log.Fatalf("Failed to create service %s: %v", svc.name, err)
    }
}
```

### Usage Phase
```go
// Services are available throughout application lifetime
// No manual lifecycle management needed
// Services are automatically available in handlers
```

---

## Common Integration Patterns

### Database with Logging
```go
app.POST("/users", func(ctx *lokstra.Context, req *CreateUserRequest) error {
    logger, _ := lokstra.GetService[serviceapi.Logger](regCtx, "app-logger")
    dbPool, _ := lokstra.GetService[serviceapi.DbPool](regCtx, "main-db")
    
    logger.Infof("Creating user: %s", req.Email)
    
    userID, err := createUser(dbPool, req)
    if err != nil {
        logger.Errorf("Failed to create user: %v", err)
        return ctx.ErrorInternal("User creation failed")
    }
    
    logger.Infof("User created successfully: %d", userID)
    return ctx.OkCreated(map[string]any{"user_id": userID})
})
```

### Caching with Redis
```go
app.GET("/users/:id", func(ctx *lokstra.Context) error {
    redis, _ := lokstra.GetService[serviceapi.Redis](regCtx, "cache-redis")
    logger, _ := lokstra.GetService[serviceapi.Logger](regCtx, "app-logger")
    
    userID := ctx.GetPathParam("id")
    cacheKey := fmt.Sprintf("user:%s", userID)
    
    // Try cache first
    if cached, err := redis.Get(cacheKey).Result(); err == nil {
        logger.Debugf("Cache hit for user: %s", userID)
        return ctx.Ok(cached)
    }
    
    // Cache miss - fetch from database
    user, err := fetchUser(userID)
    if err != nil {
        return ctx.ErrorNotFound("User not found")
    }
    
    // Cache the result
    redis.Set(cacheKey, user, time.Hour)
    logger.Debugf("Cached user: %s", userID)
    
    return ctx.Ok(user)
})
```

---

## Next Steps

After mastering built-in services, explore:

1. **Advanced Service Patterns** - Complex service architectures
2. **Service Monitoring** - Health checks and metrics
3. **Performance Optimization** - Connection pooling and caching
4. **Custom Service Development** - Building your own services
5. **Service Testing** - Mocking and integration testing

---

## Running Examples

Each example includes:
- **Service Configuration**: Multiple service instances
- **Integration Patterns**: Service usage in handlers
- **Error Handling**: Graceful service unavailability
- **Best Practices**: Configuration and lifecycle management

```bash
# Navigate to any example directory
cd 01_logger_service/

# Run the example
go run main.go

# Test the endpoints
curl http://localhost:8080/
curl http://localhost:8080/log-levels
curl http://localhost:8080/logger-info
```

Each example demonstrates real-world service integration patterns with comprehensive documentation and testing commands.
</content>