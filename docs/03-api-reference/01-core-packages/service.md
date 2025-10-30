# Service

> Service utilities and dependency injection helpers

## Overview

The `service` package provides utilities for lazy-loading services with type-safe dependency injection. The core type is `Cached[T]`, which enables lazy initialization, automatic caching, and thread-safe access to services.

## Import Path

```go
import "github.com/primadi/lokstra/core/service"
```

---

## Core Type

### Cached[T]
Type-safe lazy-loading service container.

**Definition:**
```go
type Cached[T any] struct {
    // Unexported fields
}
```

**Features:**
- ✅ Type-safe - No casting needed
- ✅ Lazy initialization - Service loaded only on first access
- ✅ Automatic caching - Single instance across multiple Get() calls
- ✅ Thread-safe - Uses `sync.Once` internally
- ✅ Zero-cost when not used - No initialization until Get()

---

## Functions

### LazyLoad
Creates a lazy service loader for a named service.

**Signature:**
```go
func LazyLoad[T any](serviceName string) *Cached[T]
```

**Type Parameters:**
- `T` - Type of the service

**Parameters:**
- `serviceName` - Service name registered in the registry

**Returns:**
- `*Cached[T]` - Lazy service loader

**Example:**
```go
type UserService struct {
    db     *service.Cached[*DBPool]
    logger *service.Cached[*Logger]
}

func NewUserService() *UserService {
    return &UserService{
        db:     service.LazyLoad[*DBPool]("db-pool"),
        logger: service.LazyLoad[*Logger]("logger"),
    }
}

func (s *UserService) CreateUser(user *User) error {
    // Services loaded on first access
    db := s.db.Get()
    logger := s.logger.Get()
    
    logger.Info("Creating user")
    return db.Insert(user)
}
```

**Use Cases:**
- Service dependencies in struct fields
- Avoiding circular dependencies
- Deferred service initialization
- Optional dependencies

---

### LazyLoadWith
Creates a lazy loader with custom loader function.

**Signature:**
```go
func LazyLoadWith[T any](loader func() T) *Cached[T]
```

**Type Parameters:**
- `T` - Type of the service

**Parameters:**
- `loader` - Custom function to load the service

**Returns:**
- `*Cached[T]` - Lazy service loader

**Example:**
```go
// Custom loader
dbLoader := service.LazyLoadWith(func() *DBPool {
    return connectToDB(os.Getenv("DB_URL"))
})

// In factory function
func MyServiceFactory(deps map[string]any, config map[string]any) any {
    return &MyService{
        db: service.LazyLoadWith(func() *DBPool {
            return deps["db"].(*DBPool)
        }),
    }
}
```

---

### LazyLoadFrom
Creates a lazy loader from a ServiceGetter interface.

**Signature:**
```go
func LazyLoadFrom[T any](getter ServiceGetter, serviceName string) *Cached[T]
```

**Parameters:**
- `getter` - Object implementing `GetService(serviceName string) (any, error)`
- `serviceName` - Name of service to load

**Returns:**
- `*Cached[T]` - Lazy service loader

**Example:**
```go
// Load from deployment app
app := deploy.GetServerApp("api", "crud-api")
userService := service.LazyLoadFrom[*UserService](app, "user-service")

// Service loaded on first access
users := userService.MustGet().GetAll()
```

**Use Cases:**
- Loading services from deployment apps
- Cross-app service access
- Testing with mock service getters

---

### LazyLoadFromConfig
Creates a lazy loader from factory configuration.

**Signature:**
```go
func LazyLoadFromConfig[T any](cfg map[string]any, key string) *Cached[T]
```

**Parameters:**
- `cfg` - Configuration map (from factory)
- `key` - Config key containing service name

**Returns:**
- `*Cached[T]` - Lazy service loader, or `nil` if key not found

**Example:**
```go
// In factory function
func UserServiceFactory(deps map[string]any, config map[string]any) any {
    return &UserService{
        db:    service.LazyLoadFromConfig[*DBPool](config, "db"),
        cache: service.LazyLoadFromConfig[*Cache](config, "cache"),
    }
}

// Config YAML:
// service-definitions:
//   user-service:
//     type: user-service-factory
//     config:
//       db: db-pool
//       cache: redis-cache
```

---

### MustLazyLoadFromConfig
Like `LazyLoadFromConfig` but panics if key is missing.

**Signature:**
```go
func MustLazyLoadFromConfig[T any](cfg map[string]any, key string) *Cached[T]
```

**Example:**
```go
func UserServiceFactory(deps map[string]any, config map[string]any) any {
    return &UserService{
        // Panic if "db" not in config (required dependency)
        db: service.MustLazyLoadFromConfig[*DBPool](config, "db"),
    }
}
```

---

### Value
Creates a Cached instance with a pre-loaded value (no lazy loading).

**Signature:**
```go
func Value[T any](value T) *Cached[T]
```

**Parameters:**
- `value` - Pre-loaded value

**Returns:**
- `*Cached[T]` - Cached instance (already loaded)

**Example:**
```go
// For testing
func TestUserService(t *testing.T) {
    mockDB := &MockDBPool{}
    mockLogger := &MockLogger{}
    
    svc := &UserService{
        db:     service.Value(mockDB),
        logger: service.Value(mockLogger),
    }
    
    // Services already loaded, no registry needed
    svc.CreateUser(&User{Name: "Test"})
}
```

**Use Cases:**
- Unit testing with mocks
- Pre-initialized services
- Static configuration values

---

### Cast
Converts a dependency value from `map[string]any` to typed `Cached[T]`.

**Signature:**
```go
func Cast[T any](value any) *Cached[T]
```

**Parameters:**
- `value` - Dependency value from `deps` map

**Returns:**
- `*Cached[T]` - Typed cached instance

**Example:**
```go
func UserServiceFactory(deps map[string]any, config map[string]any) any {
    return &UserService{
        DB:     service.Cast[*DBPool](deps["db"]),
        Cache:  service.Cast[*Cache](deps["cache"]),
        Logger: service.Cast[*Logger](deps["logger"]),
    }
}
```

**Use Cases:**
- Factory functions receiving `deps` map
- Type-safe dependency extraction
- Framework-managed dependencies

---

### CastProxyService
Casts a dependency value to `*proxy.Service`.

**Signature:**
```go
func CastProxyService(value any) *proxy.Service
```

**Parameters:**
- `value` - Dependency value (should be `*proxy.Service`)

**Returns:**
- `*proxy.Service` - Proxy service instance

**Example:**
```go
// Remote service factory
func UserServiceRemoteFactory(deps map[string]any, config map[string]any) any {
    return &UserServiceRemote{
        proxyService: service.CastProxyService(deps["remote"]),
    }
}
```

**Use Cases:**
- Remote service implementations
- External API wrappers
- Proxy-based services

---

## Cached[T] Methods

### Get
Retrieves the service instance (loads on first call, cached thereafter).

**Signature:**
```go
func (c *Cached[T]) Get() T
```

**Returns:**
- `T` - Service instance

**Example:**
```go
type UserService struct {
    db *service.Cached[*DBPool]
}

func (s *UserService) GetUsers() []User {
    // Load DB on first call, use cached instance on subsequent calls
    db := s.db.Get()
    return db.QueryAll("SELECT * FROM users")
}
```

**Thread-Safety:**
- ✅ Safe to call from multiple goroutines
- ✅ Guaranteed single initialization
- ✅ No race conditions

---

### MustGet
Retrieves the service instance or panics if not found.

**Signature:**
```go
func (c *Cached[T]) MustGet() T
```

**Returns:**
- `T` - Service instance

**Panics:**
- If service is not registered or initialization returns zero value

**Example:**
```go
func (s *UserService) GetUsers() []User {
    // Panic if DB service not found (fail-fast)
    db := s.db.MustGet()
    return db.QueryAll("SELECT * FROM users")
}
```

**When to Use:**
- Required dependencies
- Fail-fast initialization
- Clear error reporting

---

### ServiceName
Returns the service name being loaded.

**Signature:**
```go
func (c *Cached[T]) ServiceName() string
```

**Returns:**
- `string` - Service name

**Example:**
```go
db := service.LazyLoad[*DBPool]("db-pool")
fmt.Println(db.ServiceName()) // Output: db-pool
```

---

### IsLoaded
Checks if the service has been loaded.

**Signature:**
```go
func (c *Cached[T]) IsLoaded() bool
```

**Returns:**
- `bool` - `true` if Get() was called at least once

**Example:**
```go
if !s.db.IsLoaded() {
    log.Println("DB not yet accessed")
}
```

**Use Cases:**
- Debugging
- Performance monitoring
- Conditional initialization

---

## Complete Examples

### Service with Dependencies
```go
package service

import "github.com/primadi/lokstra/core/service"

type UserService struct {
    db         *service.Cached[*DBPool]
    cache      *service.Cached[*Cache]
    logger     *service.Cached[*Logger]
    mailSender *service.Cached[*MailSender]
}

func NewUserService() *UserService {
    return &UserService{
        db:         service.LazyLoad[*DBPool]("db-pool"),
        cache:      service.LazyLoad[*Cache]("redis-cache"),
        logger:     service.LazyLoad[*Logger]("logger"),
        mailSender: service.LazyLoad[*MailSender]("mail-sender"),
    }
}

func (s *UserService) CreateUser(user *User) error {
    // Services loaded lazily on first access
    db := s.db.Get()
    cache := s.cache.Get()
    logger := s.logger.Get()
    
    logger.Info("Creating user", user.Email)
    
    if err := db.Insert("users", user); err != nil {
        logger.Error("Failed to create user", err)
        return err
    }
    
    // Invalidate cache
    cache.Delete("users:list")
    
    // Send welcome email (async)
    go s.sendWelcomeEmail(user)
    
    return nil
}

func (s *UserService) sendWelcomeEmail(user *User) {
    mailer := s.mailSender.Get()
    mailer.Send(user.Email, "Welcome!", "Welcome to our platform!")
}
```

### Factory Function with Dependencies
```go
func UserServiceFactory(deps map[string]any, config map[string]any) any {
    return &UserService{
        db:         service.Cast[*DBPool](deps["db"]),
        cache:      service.Cast[*Cache](deps["cache"]),
        logger:     service.Cast[*Logger](deps["logger"]),
        mailSender: service.Cast[*MailSender](deps["mail"]),
    }
}

// Register factory
lokstra_registry.RegisterServiceType(
    "user-service-factory",
    UserServiceFactory,
    nil,
    deploy.WithDependencies("db", "cache", "logger", "mail"),
)
```

### Optional Dependencies
```go
type UserService struct {
    db     *service.Cached[*DBPool]
    cache  *service.Cached[*Cache]  // Optional
    logger *service.Cached[*Logger] // Optional
}

func (s *UserService) GetUser(id int) (*User, error) {
    db := s.db.MustGet() // Required
    
    // Check optional cache
    if s.cache != nil && s.cache.IsLoaded() {
        if user := s.cache.Get().Get("user:" + id); user != nil {
            return user.(*User), nil
        }
    }
    
    user, err := db.QueryOne("SELECT * FROM users WHERE id = ?", id)
    if err != nil {
        // Log if logger available
        if s.logger != nil {
            s.logger.Get().Error("DB query failed", err)
        }
        return nil, err
    }
    
    // Cache if available
    if s.cache != nil {
        s.cache.Get().Set("user:"+id, user, 5*time.Minute)
    }
    
    return user, nil
}
```

### Testing with Mocks
```go
func TestUserService_CreateUser(t *testing.T) {
    // Mock dependencies
    mockDB := &MockDBPool{}
    mockCache := &MockCache{}
    mockLogger := &MockLogger{}
    
    // Create service with pre-loaded mocks
    svc := &UserService{
        db:     service.Value(mockDB),
        cache:  service.Value(mockCache),
        logger: service.Value(mockLogger),
    }
    
    // Test
    user := &User{Name: "John", Email: "john@example.com"}
    err := svc.CreateUser(user)
    
    assert.NoError(t, err)
    assert.True(t, mockDB.InsertCalled)
    assert.True(t, mockCache.DeleteCalled)
    assert.True(t, mockLogger.InfoCalled)
}
```

### Remote Service Implementation
```go
type UserServiceRemote struct {
    proxyService *proxy.Service
}

func NewUserServiceRemote(proxyService *proxy.Service) *UserServiceRemote {
    return &UserServiceRemote{
        proxyService: proxyService,
    }
}

func (s *UserServiceRemote) GetUser(id int) (*User, error) {
    return proxy.CallWithData[*User](s.proxyService, "GetUser", id)
}

func (s *UserServiceRemote) CreateUser(user *User) (*User, error) {
    return proxy.CallWithData[*User](s.proxyService, "CreateUser", user)
}

// Factory
func UserServiceRemoteFactory(deps map[string]any, config map[string]any) any {
    return NewUserServiceRemote(
        service.CastProxyService(deps["remote"]),
    )
}
```

---

## Best Practices

### 1. Use LazyLoad for Service Dependencies
```go
// ✅ Good: Lazy loading
type UserService struct {
    db *service.Cached[*DBPool]
}

// 🚫 Avoid: Direct references (breaks lazy initialization)
type UserService struct {
    db *DBPool
}
```

### 2. Use MustGet for Required Dependencies
```go
// ✅ Good: Fail-fast if missing
func (s *UserService) CreateUser(user *User) error {
    db := s.db.MustGet()
    return db.Insert(user)
}

// 🚫 Avoid: Silent failures
func (s *UserService) CreateUser(user *User) error {
    db := s.db.Get()
    if db == nil {
        return errors.New("db not available")
    }
    return db.Insert(user)
}
```

### 3. Use Value for Testing
```go
// ✅ Good: Test with mocks
func TestService(t *testing.T) {
    mockDB := &MockDB{}
    svc := &UserService{
        db: service.Value(mockDB),
    }
    // ...
}

// 🚫 Avoid: Registry in tests
func TestService(t *testing.T) {
    lokstra_registry.RegisterService("db", mockDB)
    svc := NewUserService()
    // ...cleanup registry...
}
```

---

## See Also

- **[lokstra_registry](../02-registry/lokstra_registry)** - Service registration
- **[Service Registration](../02-registry/service-registration)** - RegisterServiceType
- **[Proxy](../08-advanced/proxy)** - Remote services

---

## Related Guides

- **[Service Essentials](../../01-essentials/02-service/)** - Service basics
- **[Dependency Injection](../../02-deep-dive/service/)** - Advanced DI patterns
- **[Testing](../../04-guides/testing/)** - Testing with services
