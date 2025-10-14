# 05. Services - Complete Guide

Services are reusable components that provide specific functionality (database, caching, email, logging, etc.).

## What You'll Learn

1. **Service Factory** - Create services from config
2. **Service Container** - Proper caching pattern (CRITICAL!)
3. **Lazy Loading** - Services created only when needed
4. **Dependency Injection** - Services depending on other services
5. **Interface-based** - Easy testing with mocks
6. **Usage in Handlers** - How to use services in HTTP handlers

## Quick Start

```bash
go run .
```

Then test with `test.http` or curl commands shown in output.

## Key Concepts

### 1. Service Factory

Factory converts `map[string]any` (from YAML) to typed service:

```go
func EmailServiceFactory(params map[string]any) any {
    cfg := &EmailConfig{
        SMTPHost: utils.GetValueFromMap(params, "smtp_host", "localhost"),
        SMTPPort: utils.GetValueFromMap(params, "smtp_port", 587),
    }
    return NewEmailService(cfg)
}
```

**Register factory:**
```go
lokstra_registry.RegisterServiceFactory("email", EmailServiceFactory)
```

### 2. Service Container (CRITICAL PATTERN!)

**Problem:** Local variables don't cache services

```go
// ❌ WRONG: Local variable resets every call
func handler(c *lokstra.RequestContext) error {
    var email *EmailService  // Reset to nil on each request!
    email = lokstra_registry.GetService("email", email)
    // Always fetches from registry - no caching benefit
}
```

**Solution:** ServiceContainer with struct fields

```go
// ✅ CORRECT: Struct fields persist across calls
type ServiceContainer struct {
    emailCache *EmailService
    dbCache    *DBService
}

func (sc *ServiceContainer) GetEmail() *EmailService {
    sc.emailCache = lokstra_registry.GetService("email", sc.emailCache)
    return sc.emailCache  // Cached!
}

// Global container
var services = &ServiceContainer{}

// Handler uses container
func handler(c *lokstra.RequestContext) error {
    email := services.GetEmail()  // First call: loads, Second call: cached
}
```

**Why it works:**
- Struct field persists across method calls
- First call: `emailCache` is nil → fetches from registry
- Second call: `emailCache` is set → returns immediately
- Zero registry lookups after first access

### 3. Lazy Loading

Services created only when first accessed:

```go
// Registration: Just config, no instance created
lokstra_registry.RegisterLazyService("db", "db", config)

// First access: Creates instance
db := services.GetDB()  // Creates + caches

// Second access: Returns cached
db := services.GetDB()  // Instant return
```

**Benefits:**
- Memory efficient (unused services not created)
- Faster startup (no upfront cost)
- Flexible (services created in order of use)

### 4. Dependency Injection

Services can depend on other services:

```go
type UserService struct {
    db *DBService  // Dependency
}

func UserServiceFactory(params map[string]any) any {
    // Resolve dependency
    var db *DBService
    db = lokstra_registry.GetService("db", db)
    
    return NewUserService(db)
}
```

### 5. Interface-Based (Easy Testing)

```go
// Define interface
type Logger interface {
    Info(msg string)
    Error(msg string)
}

// Real implementation
type ConsoleLogger struct { ... }

// Mock for testing
type MockLogger struct {
    calls []string
}

func (m *MockLogger) Info(msg string) {
    m.calls = append(m.calls, msg)
}
```

## Service Creation Methods

### Method 1: NewService (Eager)

Creates service immediately:

```go
email := lokstra_registry.NewService[*EmailService](
    "email",           // service name
    "email",           // factory type
    map[string]any{    // configuration
        "smtp_host": "smtp.gmail.com",
    },
)
```

**Use when:** You know the service is always needed

### Method 2: RegisterLazyService + GetService (Lazy)

Registers config only, creates on first access:

```go
// Registration (startup)
lokstra_registry.RegisterLazyService("db", "db", config)

// Usage (runtime, lazy)
db := services.GetDB()  // Created on first call
```

**Use when:** Service might not be needed or is expensive to create

### Method 3: GetService (Retrieve Existing)

```go
var cache *CacheService
cache = lokstra_registry.GetService("cache", cache)
```

**Behavior:**
- If `cache != nil` → returns immediately
- If `cache == nil` → looks up registry or creates from lazy config
- If not found → **panics**

### Method 4: TryGetService (Safe Version)

```go
cache, ok := lokstra_registry.TryGetService[*CacheService]("cache", nil)
if !ok {
    // Service not found - handle gracefully
}
```

**Use when:** Service is optional

## Complete Example

See `main.go` for a complete runnable example showing:

1. ✅ Simple service (Email)
2. ✅ Service with dependencies (UserService needs DBService)
3. ✅ Interface-based service (Logger)
4. ✅ ServiceContainer with proper caching
5. ✅ Lazy loading demonstration
6. ✅ HTTP handlers using services

## Common Patterns

### Pattern 1: Service in Handler

```go
var services = &ServiceContainer{}

func userHandler(c *lokstra.RequestContext) error {
    db := services.GetDB()      // Lazy + cached
    logger := services.GetLogger()
    
    logger.Info("Processing request")
    user := db.GetUser(id)
    
    return c.Api.Ok(user)
}
```

### Pattern 2: Service in Middleware

```go
func authMiddleware(c *lokstra.RequestContext) error {
    authService := services.GetAuth()
    
    user := authService.Validate(token)
    c.Set("user", user)
    
    return c.Next()
}
```

### Pattern 3: Service with Config from YAML

```yaml
# config.yaml
services:
  - name: email
    type: email
    config:
      smtp_host: ${SMTP_HOST:localhost}
      smtp_port: ${SMTP_PORT:587}
```

```go
// Load config and create services
config.LoadConfigFile("config.yaml", cfg)
lokstra_registry.CreateServicesFromConfig(cfg)
```

## Best Practices

✅ **DO:**
- Use ServiceContainer with struct fields
- Register factories at startup
- Use lazy services for expensive resources
- Define interfaces for testability
- Handle dependencies in factories

❌ **DON'T:**
- Use local variables for caching (they reset!)
- Create services in loops
- Forget to register factories before using
- Mix service creation methods inconsistently

## Decision Guide

| Scenario | Method | Why |
|----------|--------|-----|
| Always needed | `NewService` | Create upfront |
| Might be needed | `RegisterLazyService` | Create on demand |
| Expensive (DB, cache) | Lazy | Defer cost |
| Optional service | `TryGetService` | Safe failure |
| In handler/middleware | ServiceContainer | Proper caching |
| Testing required | Interface | Easy mocking |

## What's Next?

- **[02-architecture](../../02-architecture/)** - Config-driven services from YAML
- **[03-best-practices](../../03-best-practices/)** - Production service patterns
- Complete apps combining all concepts

## Key Takeaway

**ServiceContainer is the critical pattern:**
- Bridges service registry and handlers
- Provides proper caching via struct fields
- Enables lazy loading
- Clean, testable code

Without it, you lose caching benefits and repeat service lookups unnecessarily.

### [01-factory](01-factory/) - Service Factory Basics
**What you'll learn:**
- Create service factories that accept `map[string]any` config
- Register factories with `RegisterServiceFactory`
- Create services with `NewService`
- **CRITICAL:** Proper caching with `GetService` using struct fields
- Why local variables don't cache services

**Key takeaway:** ServiceContainer pattern with persistent struct fields

```go
type ServiceContainer struct {
    emailCache *EmailService
}

func (sc *ServiceContainer) GetEmail() *EmailService {
    sc.emailCache = lokstra_registry.GetService("email", sc.emailCache)
    return sc.emailCache
}
```

---

### [02-lazy](02-lazy/) - Lazy Service Initialization
**What you'll learn:**
- Register services with `RegisterLazyService` (config only, no instance)
- Services created only when first accessed
- Memory efficient - services created on-demand
- Once created, services are cached and reused

**Key takeaway:** Lazy loading for expensive services (database, cache, etc.)

```go
// Registration: just config, no instance created
lokstra_registry.RegisterLazyService("db", DB_TYPE, config)

// First access: creates instance
db := services.GetDB() // Creates + caches

// Second access: returns cached
db := services.GetDB() // Returns cached instantly
```

---

### [03-registry](03-registry/) - Registry Patterns
**What you'll learn:**
- Difference between `NewService`, `GetService`, `TryGetService`
- When to use each pattern
- Error handling strategies
- Service lifecycle management

**Key differences:**
- `NewService` - Creates and registers (eager)
- `GetService` - Gets or creates, **panics** if not found
- `TryGetService` - Gets or creates, returns `(service, bool)`

**Key takeaway:** Choose the right method for your use case

---

### [04-contracts](04-contracts/) - Service Interfaces (SOLID)
**What you'll learn:**
- Define service contracts (Go interfaces)
- Multiple implementations of same contract
- Dependency injection pattern
- Easy testing with mock services
- Loose coupling between components

**Key takeaway:** Program to interfaces, not implementations

```go
// Contract
type Logger interface {
    Info(msg string)
    Error(msg string)
}

// Multiple implementations
type ConsoleLogger struct { ... }
type FileLogger struct { ... }
type MockLogger struct { ... } // For testing

// Dependency injection
func NewUserService(logger Logger) *UserService {
    return &UserService{logger: logger}
}
```

---

## 🎯 Key Concepts Summary

### 1. Service Factory Pattern
```go
// Factory accepts map[string]any (from YAML)
func EmailServiceFactory(params map[string]any) any {
    cfg := mapToConfig(params)  // Convert to typed struct
    return NewEmailService(cfg)  // Return typed service
}
```

### 2. Service Container Pattern (Recommended!)
```go
type ServiceContainer struct {
    dbCache    *DBService
    cacheCache *CacheService
}

func (sc *ServiceContainer) GetDB() *DBService {
    sc.dbCache = lokstra_registry.GetService("db", sc.dbCache)
    return sc.dbCache
}

// Global container
var services = &ServiceContainer{}

// Usage in handler
func handler(c *lokstra.RequestContext) error {
    db := services.GetDB()  // Lazy + cached
    // ...
}
```

### 3. Service Lifecycle

```
1. Registration Phase (app startup)
   ├─ RegisterServiceFactory("type", factory)
   ├─ RegisterLazyService("name", "type", config)
   └─ NewService("name", "type", config) // Optional eager

2. Runtime Phase (handling requests)
   ├─ GetService("name", cache) // First call: creates
   ├─ GetService("name", cache) // Subsequent: returns cache
   └─ TryGetService("name", cache) // Safe version
```

### 4. Best Practices

✅ **DO:**
- Use ServiceContainer with struct fields for caching
- Register factories at app startup
- Use lazy services for expensive resources
- Define interfaces for testability
- Use GetService with persistent cache variable

❌ **DON'T:**
- Use local variables for caching (they reset!)
- Create services inside loops
- Mix service types without interfaces
- Forget to handle TryGetService's bool return

---

## 🚀 Quick Start

Each example is self-contained and runnable:

```bash
# Run examples
cd 01-factory && go run .
cd 02-lazy && go run .
cd 03-registry && go run .
cd 04-contracts && go run .
```

All examples include:
- ✅ Runnable `main.go`
- ✅ Comprehensive `README.md`
- ✅ Clear console output explaining concepts

---

## 📖 Common Patterns

### Pattern 1: Simple Service
```go
// Define
type EmailService struct { cfg *Config }
func NewEmailService(cfg *Config) *EmailService { ... }

// Factory
func EmailServiceFactory(params map[string]any) any {
    return NewEmailService(mapToConfig(params))
}

// Register
lokstra_registry.RegisterServiceFactory("email", EmailServiceFactory)

// Create
email := lokstra_registry.NewService[*EmailService]("email", "email", config)
```

### Pattern 2: Service with Dependencies
```go
// Service B depends on Service A
type ServiceB struct {
    serviceA *ServiceA
}

// Factory resolves dependency
func ServiceBFactory(params map[string]any) any {
    var serviceA *ServiceA
    serviceA = lokstra_registry.GetService("serviceA", serviceA)
    return NewServiceB(serviceA)
}
```

### Pattern 3: Interface-based Service
```go
// Define interface
type Storage interface {
    Save(key, value string) error
    Load(key string) (string, error)
}

// Multiple implementations
type MemoryStorage struct { ... }
type RedisStorage struct { ... }

// Register by name
lokstra_registry.RegisterServiceFactory("memory_storage", MemoryStorageFactory)
lokstra_registry.RegisterServiceFactory("redis_storage", RedisStorageFactory)

// Choose implementation via config
storage := lokstra_registry.NewService[Storage]("storage", "memory_storage", nil)
```

---

## 🔗 Integration with Other Components

### With Handlers
```go
var services = &ServiceContainer{}

func userHandler(c *lokstra.RequestContext) error {
    db := services.GetDB()
    cache := services.GetCache()
    
    // Use services...
}
```

### With Middleware
```go
func authMiddleware(c *lokstra.RequestContext) error {
    authService := services.GetAuth()
    
    user := authService.Validate(token)
    c.Set("user", user)
    
    return c.Next()
}
```

### With Config-Driven Architecture
```yaml
# config.yaml
services:
  - name: db
    type: postgres
    config:
      host: ${DB_HOST:localhost}
      port: ${DB_PORT:5432}
  
  - name: cache
    type: redis
    config:
      host: ${REDIS_HOST:localhost}
```

---

## 📊 Decision Matrix

| Scenario | Use This | Why |
|----------|----------|-----|
| Need service in multiple places | ServiceContainer | Proper caching |
| Expensive initialization (DB) | RegisterLazyService | Create only when needed |
| Known dependencies | NewService | Eager initialization |
| Optional service | TryGetService | Returns bool, safe |
| Testing required | Interface pattern | Easy mocking |
| YAML configuration | Factory pattern | Config-driven |

---

## 🎓 What's Next?

After mastering services, you're ready for:

- **[02-architecture](../../02-architecture/)** - Config-driven application architecture
- **[03-best-practices](../../03-best-practices/)** - Production deployment patterns
- Complete app examples combining routers, middleware, and services

---

## 💡 Key Insight

**The ServiceContainer pattern is the bridge between:**
- Service Registry (global service storage)
- Handlers/Middleware (needs services)
- Proper caching (struct fields persist across calls)

Without ServiceContainer, you'd either:
- ❌ Create services repeatedly (inefficient)
- ❌ Use local variables (doesn't cache!)
- ❌ Use global variables (hard to test)

With ServiceContainer:
- ✅ Services cached properly
- ✅ Lazy loaded on first use
- ✅ Easy to test (inject mock container)
- ✅ Clean, explicit dependencies
