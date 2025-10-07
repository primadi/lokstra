# 03-Service Dependencies

Learn the **CORRECT way** to handle service dependencies with lazy loading and proper caching.

## Quick Start

```bash
go run .
```

## ❌ Common Mistake (Don't Do This!)

```go
// WRONG! GetService called during registration phase
func UserServiceFactory(cfg map[string]any) any {
    var db *DBService
    var cache *CacheService
    
    // BAD! This runs when service is REGISTERED, not when it's USED
    db = lokstra_registry.GetService("db-service", db)
    cache = lokstra_registry.GetService("cache-service", cache)
    
    return NewUserService(db, cache)
}
```

**Problems:**
1. ❌ No lazy loading - all services created at startup
2. ❌ No caching benefit - GetService cache doesn't work
3. ❌ Registration order dependency - DB must exist before User
4. ❌ Tight coupling - hardcoded dependency names
5. ❌ Can't test with mocks

## ✅ Correct Pattern - Lazy Loading

### Step 1: Store Service Names (Not Instances)

### Step 1: Store Service Names (Not Instances)

```go
type UserService struct {
    // Store service NAMES from config
    dbServiceName    string
    cacheServiceName string
    
    // Cache variables (nil until first use)
    dbCache    *DBService
    cacheCache *CacheService
}
```

**Key Points:**
- Store **names** (strings), not instances
- Cache variables start as `nil`
- No GetService calls yet!

### Step 2: Create Lazy Getter Methods

```go
// Lazy getter - only calls GetService when needed
func (s *UserService) getDB() *DBService {
    // GetService pattern: cache variable is updated
    s.dbCache = lokstra_registry.GetService(s.dbServiceName, s.dbCache)
    return s.dbCache
}

func (s *UserService) getCache() *CacheService {
    s.cacheCache = lokstra_registry.GetService(s.cacheServiceName, s.cacheCache)
    return s.cacheCache
}
```

**What happens:**
1. First call: `dbCache` is nil → GetService creates service → stores in `dbCache`
2. Second call: `dbCache` is not nil → returns immediately (no creation)
3. **This is true lazy loading + caching!**

### Step 3: Use Lazy Getters in Business Methods

```go
func (s *UserService) GetUser(id string) map[string]any {
    // Lazy load only when needed
    cache := s.getCache()  // Created on first call
    db := s.getDB()        // Created on first call
    
    // Use cache and db...
    if cached, ok := cache.Get("user:" + id); ok {
        return parseUser(cached)
    }
    
    result := db.Query("SELECT * FROM users WHERE id = " + id)
    return parseUser(result)
}
```

**Benefits:**
- Services only created when handler is called
- If handler never called, services never created
- Each subsequent call reuses cached instances

### Step 4: Factory Just Stores Names from Config

```go
func UserServiceFactory(cfg map[string]any) any {
    // NO GetService calls! Just store service names.
    return &UserService{
        dbServiceName:    utils.GetValueFromMap(cfg, "db_service", "db-service"),
        cacheServiceName: utils.GetValueFromMap(cfg, "cache_service", "cache-service"),
        // dbCache and cacheCache are nil - filled on first use
    }
}
```

**What's different:**
- ✅ No GetService calls in factory
- ✅ Service names come from config (flexible!)
- ✅ Cache variables stay nil (lazy!)
- ✅ No registration order dependency

### Step 5: Config - Explicit Dependencies

```yaml
services:
  # Layer 1: Infrastructure
  - name: db-service
    type: db
    config:
      host: localhost
      port: 5432

  - name: cache-service
    type: cache
    config:
      host: localhost
      port: 6379

  # Layer 2: Domain - Dependencies VISIBLE in config
  - name: user-service
    type: user
    config:
      db_service: db-service       # Which DB to use
      cache_service: cache-service # Which cache to use
```

**Benefits:**
- ✅ Dependencies visible in YAML (self-documenting)
- ✅ Can use different service names
- ✅ Easy to override for testing
- ✅ Multiple instances with different configs

## Complete Flow

```
┌─────────────────────────────┐
│ 1. Registration Phase       │
│ - RegisterServiceFactory()  │
│ - RegisterLazyService()     │
│ - Factory creates UserService│
│   with service NAMES only   │
│ - NO GetService calls yet   │
└────────────┬────────────────┘
             │
             ▼
┌─────────────────────────────┐
│ 2. First Request            │
│ - Handler calls             │
│   services.GetUser()        │
│ - ServiceContainer calls    │
│   GetService("user-service")│
│ - Returns UserService       │
│   (already created at reg)  │
└────────────┬────────────────┘
             │
             ▼
┌─────────────────────────────┐
│ 3. First Business Call      │
│ - user.GetUser(id)          │
│ - Calls getDB()             │
│   → GetService("db-service")│
│   → Creates DBService       │
│   → Stores in dbCache       │
│ - Calls getCache()          │
│   → GetService("cache")     │
│   → Creates CacheService    │
│   → Stores in cacheCache    │
└────────────┬────────────────┘
             │
             ▼
┌─────────────────────────────┐
│ 4. Subsequent Calls         │
│ - user.GetUser(id)          │
│ - getDB() returns dbCache   │
│   (no GetService call!)     │
│ - getCache() returns cache  │
│   (no GetService call!)     │
│ - FAST! No creation overhead│
└─────────────────────────────┘
```

## Why This Pattern is Critical

### Performance

```go
// ❌ WRONG - Creates on registration (startup)
// Even if never used!
func Factory(cfg) {
    dep = GetService("dep", dep)  // Created at startup
    return NewService(dep)
}

// ✅ CORRECT - Creates on first use
// If never used, never created!
func Factory(cfg) {
    return &Service{
        depName: "dep",  // Just store name
    }
}

func (s *Service) getDep() {
    s.cache = GetService(s.depName, s.cache)  // Created when needed
    return s.cache
}
```

### Flexibility

```yaml
# Development - use local services
- name: user-service
  type: user
  config:
    db_service: db-local
    cache_service: cache-local

# Production - use production services
- name: user-service
  type: user
  config:
    db_service: db-production-primary
    cache_service: cache-production-cluster

# Testing - use mocks
- name: user-service
  type: user
  config:
    db_service: mock-db
    cache_service: mock-cache
```

### Multiple Instances

```yaml
# Write service - uses primary DB
- name: user-service-write
  type: user
  config:
    db_service: db-primary
    cache_service: cache-write

# Read service - uses replica DB
- name: user-service-read
  type: user
  config:
    db_service: db-replica
    cache_service: cache-read
```

## Dependency Tree

```
Layer 1: Infrastructure (no dependencies)
├─ db-service
└─ cache-service

Layer 2: Domain (depend on Layer 1)
├─ user-service
│  ├─→ db-service      (lazy loaded on first use)
│  └─→ cache-service   (lazy loaded on first use)
└─ order-service
   ├─→ db-service      (lazy loaded)
   └─→ user-service    (lazy loaded, which then loads its deps)
```

## Best Practices Summary

### ✅ DO

1. **Store service names in struct fields**
   ```go
   type Service struct {
       dbServiceName string  // Service name from config
       dbCache *DBService    // Cache variable
   }
   ```

2. **Create lazy getter methods with cache variables**
   ```go
   func (s *Service) getDB() *DBService {
       s.dbCache = lokstra_registry.GetService(s.dbServiceName, s.dbCache)
       return s.dbCache
   }
   ```

3. **Call GetService in getters (during use)**
   ```go
   func (s *Service) DoWork() {
       db := s.getDB()  // Lazy load when needed
       db.Query(...)
   }
   ```

4. **Keep factories simple - just store names**
   ```go
   func Factory(cfg map[string]any) any {
       return &Service{
           dbServiceName: utils.GetValueFromMap(cfg, "db_service", "db"),
       }
   }
   ```

5. **Declare dependencies explicitly in YAML**
   ```yaml
   config:
     db_service: my-db
     cache_service: my-cache
   ```

### ❌ DON'T

1. **Never call GetService in factory**
   ```go
   // ❌ WRONG - Called at registration time
   func Factory(cfg) {
       dep = GetService("dep", dep)  // NO!
       return NewService(dep)
   }
   ```

2. **Never create constructors that receive instances**
   ```go
   // ❌ WRONG - Forces eager loading
   func NewService(db *DBService) *Service {
       return &Service{db: db}
   }
   ```

3. **Never store instances directly in struct (without name)**
   ```go
   // ❌ WRONG - No lazy loading, no flexibility
   type Service struct {
       db *DBService  // Instance, not name
   }
   ```

4. **Never hardcode service names**
   ```go
   // ❌ WRONG - Not flexible
   func (s *Service) getDB() *DBService {
       s.dbCache = GetService("db-service", s.dbCache)  // Hardcoded!
   }
   ```

## Testing Benefits

### Easy Mocking

```yaml
# production.yaml
services:
  - name: user-service
    type: user
    config:
      db_service: postgres-production
      cache_service: redis-cluster

# test.yaml
services:
  - name: user-service
    type: user
    config:
      db_service: mock-db      # Just change names!
      cache_service: mock-cache
```

### No Need for Separate Test Factories

```go
// Same factory works for production and testing!
func UserServiceFactory(cfg map[string]any) any {
    return &UserService{
        dbServiceName: utils.GetValueFromMap(cfg, "db_service", "db-service"),
    }
}

// Just register different service names in test:
lokstra_registry.RegisterServiceFactory("mock-db", func(cfg) any {
    return &MockDB{}
})
```

## Run Example

```bash
cd cmd/learning/02-architecture/03-service-dependencies
go run main.go
```

### Test Endpoints

```bash
# Get user (triggers lazy loading of DB and Cache)
curl http://localhost:8080/users/123

# Get order (triggers lazy loading of DB and UserService)
curl http://localhost:8080/orders/456
```

### Expected Console Output

```
Starting service dependencies example...

=== Service Registration (Factories Only) ===
✓ RegisterServiceFactory: db-service
✓ RegisterServiceFactory: cache-service
✓ RegisterServiceFactory: user-service (stores names: "db-service", "cache-service")
✓ RegisterServiceFactory: order-service (stores names: "db-service", "user-service")

=== First Request: GET /users/123 ===
→ GetService("user-service") - already created at registration
→ user.GetUser() called
  → getDB() - first call, creates DBService
    ✓ Created DBService
  → getCache() - first call, creates CacheService
    ✓ Created CacheService
✓ Response: user data

=== Second Request: GET /users/456 ===
→ GetService("user-service") - returns cached instance
→ user.GetUser() called
  → getDB() - returns cached dbCache (no creation!)
  → getCache() - returns cached cacheCache (no creation!)
✓ Response: user data (FAST!)

=== Third Request: GET /orders/789 ===
→ GetService("order-service") - already created at registration
→ order.GetOrder() called
  → getDB() - first call, creates DBService
    ✓ Created DBService
  → getUser() - first call, gets UserService
    → UserService already created from previous requests
    → UserService's DB and Cache already cached!
✓ Response: order data with user info
```

## Key Takeaways

1. **GetService Phase Matters**
   - ❌ In factory = Registration phase (too early!)
   - ✅ In getter = Usage phase (correct!)

2. **Lazy Loading Requirements**
   - Store service **names** (not instances)
   - Create **getter methods**
   - Use **cache variables** (struct fields)
   - Call GetService **in getters** (not factory)

3. **Cache Pattern**
   - Cache variables start as `nil`
   - First call: nil → creates service → stores in cache
   - Subsequent calls: returns cached value immediately
   - **Must be struct field** for GetService pattern to work!

4. **Benefits**
   - True lazy loading (create only when used)
   - Proper caching (reuse instances)
   - No registration order dependency
   - Flexible service naming (config-driven)
   - Easy testing (swap names in config)
   - Multiple instances possible (different names)

5. **Critical for Framework Users**
   - This is fundamental to Lokstra architecture
   - Wrong pattern defeats lazy service benefits
   - Correct pattern enables all advanced features
   - Must be emphasized in documentation and examples!

## Next Steps

- **04-config-driven-deployment** - Complete app entirely from config.yaml

## Comparison

| Feature | Wrong Pattern | Correct Pattern (This Example) |
|---------|---------------|--------------------------------|
| GetService called | In factory (registration) | In getter (usage) |
| Loading | Eager (at startup) | Lazy (when needed) |
| Caching | Doesn't work properly | Works via struct fields |
| Registration order | Required, strict | Independent, flexible |
| Testing | Hard to mock | Easy to mock (change config) |
| Multiple instances | Difficult | Easy (different service names) |
| Performance | All services created at startup | Only used services created |
