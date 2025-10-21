# 06-Config-Driven Deployment with Layered Services

Comparison between **Simple Services** (flat array) and **Layered Services** (explicit layers with `depends-on`) using the same e-commerce application.

## Quick Start

### Run with Simple Services (Backward Compatible)
```bash
go run . simple
# or
go run . config-simple.yaml
```

### Run with Layered Services (New Pattern)
```bash
go run . layered
# or
go run . config-layered.yaml
```

## What's Different?

This example demonstrates **TWO WAYS** to configure the same application:

| Feature | Simple Services | Layered Services |
|---------|----------------|------------------|
| **Config Format** | Flat array | Grouped by layer |
| **Dependencies** | Implicit (in code) | Explicit (`depends-on`) |
| **Validation** | None | Layer violations detected |
| **Architecture Visibility** | Hidden | Clear in config |
| **Factory Pattern** | Manual string refs | Generic `Lazy[T]` |
| **Boilerplate** | ~15 lines/dependency | ~3 lines/dependency |
| **Type Safety** | Manual casts | Generic type-safe |

## Simple Services (Old Pattern)

### Config Format
```yaml
services:
  - name: db-service
    type: db
  
  - name: user-repo
    type: user-repo
    config:
      db_service: db-service  # String reference
  
  - name: user-service
    type: user
    config:
      repository_service: user-repo  # String reference
```

### Factory Pattern (Manual)
```go
type UserRepository struct {
    dbServiceName string  // Store name
    dbCache       *DBService  // Cache instance
}

func (r *UserRepository) getDB() *DBService {
    // Manual lazy loading with caching
    r.dbCache = old_registry.GetService(r.dbServiceName, r.dbCache)
    return r.dbCache
}

func NewUserRepository(cfg map[string]any) any {
    return &UserRepository{
        dbServiceName: utils.GetValueFromMap(cfg, "db_service", "db-service"),
    }
}
```

**Problems:**
- ❌ 15+ lines of boilerplate per dependency
- ❌ Manual caching logic
- ❌ No compile-time type checking
- ❌ Dependencies hidden in factory code
- ❌ No validation that `db-service` exists

## Layered Services (New Pattern)

### Config Format
```yaml
services:
  infrastructure:
    - name: db-service
      type: db
  
  repository:
    - name: user-repo
      type: user-repo
      depends-on: [db-service]  # Explicit dependency
      config:
        db_service: db-service  # Injected as GenericLazyService
  
  domain:
    - name: user-service
      type: user
      depends-on: [user-repo]  # Explicit dependency
      config:
        repository_service: user-repo  # Injected as GenericLazyService
```

### Factory Pattern (Generic Lazy)
```go
type UserRepository struct {
    db *service.Cached[DBService]  // Generic lazy container
}

func (r *UserRepository) FindByID(id string) map[string]any {
    db := r.db.Get()  // Type-safe lazy load + auto-cache
    db.Query("SELECT * FROM users WHERE id = " + id)
    return map[string]any{}
}

func NewUserRepository(cfg map[string]any) any {
    return &UserRepository{
        db: service.GetLazyService[DBService](cfg, "db_service"),
    }
}
```

**Benefits:**
- ✅ 3 lines per dependency (vs 15+)
- ✅ Type-safe - no manual casts
- ✅ Automatic caching with `sync.Once`
- ✅ Dependencies visible in config
- ✅ Validation: `depends-on` must exist in config
- ✅ Layer violations detected at load time

## Architecture Layers

### Infrastructure Layer
Foundation services with no internal dependencies.

```yaml
infrastructure:
  - name: db-service        # PostgreSQL
  - name: cache-service     # Redis
  - name: email-service     # SMTP
```

### Repository Layer
Data access services - depend only on infrastructure.

```yaml
repository:
  - name: user-repo
    depends-on: [db-service]
  
  - name: product-repo
    depends-on: [db-service, cache-service]
  
  - name: order-repo
    depends-on: [db-service]
```

### Domain Layer
Business logic services - depend on repositories and other domain services.

```yaml
domain:
  - name: user-service
    depends-on: [user-repo]
  
  - name: product-service
    depends-on: [product-repo]
  
  - name: order-service
    depends-on: [order-repo, product-service, user-service, email-service]
```

## Validation Rules

Layered services are validated at config load time:

### ✅ Valid Dependencies
```yaml
domain:
  - name: order-service
    depends-on: [order-repo, product-service]  # OK - previous layers
```

### ❌ Invalid Dependencies
```yaml
repository:
  - name: user-repo
    depends-on: [user-service]  # ERROR - same or later layer
```

### ❌ Unused Dependencies
```yaml
repository:
  - name: user-repo
    depends-on: [db-service, cache-service]  # Lists cache-service
    config:
      db_service: db-service  # But only uses db-service
```
**Error:** `dependency 'cache-service' declared in depends-on but not used in config`

### ❌ Undeclared Dependencies
```yaml
repository:
  - name: product-repo
    depends-on: [db-service]  # Lists only db-service
    config:
      db_service: db-service
      cache_service: cache-service  # But uses cache-service
```
**Error:** `config references service 'cache-service' which is not in depends-on`

## Dependency Injection Comparison

### Simple Mode: Manual String References

**Factory Code:**
```go
type OrderService struct {
    repoServiceName     string
    productServiceName  string
    userServiceName     string
    emailServiceName    string
    repoCache           *OrderRepository
    productServiceCache *ProductService
    userServiceCache    *UserService
    emailServiceCache   *EmailService
}

func (s *OrderService) getRepo() *OrderRepository {
    s.repoCache = old_registry.GetService(s.repoServiceName, s.repoCache)
    return s.repoCache
}

func (s *OrderService) getProductService() *ProductService {
    s.productServiceCache = old_registry.GetService(s.productServiceName, s.productServiceCache)
    return s.productServiceCache
}

func (s *OrderService) getUserService() *UserService {
    s.userServiceCache = old_registry.GetService(s.userServiceName, s.userServiceCache)
    return s.userServiceCache
}

func (s *OrderService) getEmailService() *EmailService {
    s.emailServiceCache = old_registry.GetService(s.emailServiceName, s.emailServiceCache)
    return s.emailServiceCache
}

func NewOrderService(cfg map[string]any) any {
    return &OrderService{
        repoServiceName:    utils.GetValueFromMap(cfg, "repository_service", "order-repository"),
        productServiceName: utils.GetValueFromMap(cfg, "product_service", "product-service"),
        userServiceName:    utils.GetValueFromMap(cfg, "user_service", "user-service"),
        emailServiceName:   utils.GetValueFromMap(cfg, "email_service", "email-service"),
    }
}
```

**Lines of code:** ~60 lines for 4 dependencies

### Layered Mode: Generic Lazy[T]

**Factory Code:**
```go
type OrderService struct {
    repo    *service.Cached[OrderRepository]
    product *service.Cached[ProductService]
    user    *service.Cached[UserService]
    email   *service.Cached[EmailService]
}

func NewOrderService(cfg map[string]any) any {
    return &OrderService{
        repo:    service.GetLazyService[OrderRepository](cfg, "repository_service"),
        product: service.GetLazyService[ProductService](cfg, "product_service"),
        user:    service.GetLazyService[UserService](cfg, "user_service"),
        email:   service.GetLazyService[EmailService](cfg, "email_service"),
    }
}

func (s *OrderService) CreateOrder(userID, productID string, quantity int) (map[string]any, error) {
    // Direct access - lazy loaded, cached, type-safe
    product := s.product.MustGet().GetProduct(productID)
    order := s.repo.MustGet().Create(userID, productID, quantity, total)
    user := s.user.MustGet().GetUser(userID)
    s.email.MustGet().Send(user["email"].(string), "Order Confirmation", body)
    return order, nil
}
```

**Lines of code:** ~15 lines for 4 dependencies (75% reduction!)

## Testing Both Patterns

### Test Simple Services
```bash
curl http://localhost:8080/api/products
curl http://localhost:8080/api/products/1
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id":"123","product_id":"1","quantity":2}'
```

### Test Layered Services
Same API, same behavior - just different config!

```bash
# Use layered config
go run . layered

# Same endpoints work
curl http://localhost:8080/api/products
curl http://localhost:8080/api/orders -d '...'
```

## Migration Path

### Step 1: Add depends-on to existing services
```yaml
# Before
services:
  - name: user-repo
    config:
      db_service: db-service

# After (still simple mode)
services:
  - name: user-repo
    config:
      db_service: db-service
```

### Step 2: Group by layers
```yaml
# Layered mode
services:
  infrastructure:
    - name: db-service
  repository:
    - name: user-repo
      depends-on: [db-service]
      config:
        db_service: db-service
```

### Step 3: Update factories to use Lazy[T]
```go
// Before
type UserRepository struct {
    dbServiceName string
    dbCache       *DBService
}

// After
type UserRepository struct {
    db *service.Cached[DBService]
}
```

## Benefits Summary

| Metric | Simple | Layered | Improvement |
|--------|--------|---------|-------------|
| Lines per dependency | ~15 | ~3 | **80% less** |
| Type safety | Manual casts | Generic | **Type-safe** |
| Dependency visibility | In code only | In config | **Visible** |
| Validation | None | Automated | **Validated** |
| Architecture clarity | Hidden | Explicit layers | **Clear** |
| Boilerplate | High | Minimal | **Clean** |

## Key Takeaways

1. **Simple Services**
   - ✅ Backward compatible
   - ✅ Familiar pattern
   - ❌ Verbose factory code
   - ❌ Dependencies hidden
   - ❌ No validation

2. **Layered Services**
   - ✅ Type-safe with generics
   - ✅ 80% less boilerplate
   - ✅ Architecture visible in config
   - ✅ Automatic validation
   - ✅ Clear dependency graph
   - ⚠️ Requires Go 1.18+ (generics)

3. **Both Patterns**
   - Same application behavior
   - Same service registry
   - Same runtime performance
   - Can coexist in same codebase

## Next Steps

- **../04-config-driven-deployment** - Original simple services example
- **../05-scalable-deployment** - Deployment patterns with layered services
- **../../03-best-practices** - Best practices for both patterns

## Summary

This example demonstrates:
- ✅ Side-by-side comparison of simple vs layered services
- ✅ Generic `Lazy[T]` pattern reduces boilerplate by 80%
- ✅ Explicit `depends-on` makes architecture visible
- ✅ Automatic validation prevents configuration errors
- ✅ Type-safe dependency injection
- ✅ Backward compatible - both patterns work
