# Code Comparison: Simple vs Layered Services

## Factory Pattern Comparison

### 🔴 SIMPLE SERVICES - OrderService Factory (60+ lines)

```go
type OrderService struct {
    // String names for lazy loading
    repoServiceName     string
    productServiceName  string
    userServiceName     string
    emailServiceName    string
    
    // Cache variables for performance
    repoCache           *OrderRepository
    productServiceCache *ProductService
    userServiceCache    *UserService
    emailServiceCache   *EmailService
    
    // Config
    taxRate             float64
    minOrderAmount      float64
}

// Manual lazy loading with caching (15 lines per dependency!)
func (s *OrderService) getRepo() *OrderRepository {
    s.repoCache = lokstra_registry.GetService(s.repoServiceName, s.repoCache)
    return s.repoCache
}

func (s *OrderService) getProductService() *ProductService {
    s.productServiceCache = lokstra_registry.GetService(s.productServiceName, s.productServiceCache)
    return s.productServiceCache
}

func (s *OrderService) getUserService() *UserService {
    s.userServiceCache = lokstra_registry.GetService(s.userServiceName, s.userServiceCache)
    return s.userServiceCache
}

func (s *OrderService) getEmailService() *EmailService {
    s.emailServiceCache = lokstra_registry.GetService(s.emailServiceName, s.emailServiceCache)
    return s.emailServiceCache
}

func NewOrderService(cfg map[string]any) any {
    return &OrderService{
        repoServiceName:    utils.GetValueFromMap(cfg, "repository_service", "order-repository"),
        productServiceName: utils.GetValueFromMap(cfg, "product_service", "product-service"),
        userServiceName:    utils.GetValueFromMap(cfg, "user_service", "user-service"),
        emailServiceName:   utils.GetValueFromMap(cfg, "email_service", "email-service"),
        taxRate:            utils.GetValueFromMap(cfg, "tax_rate", 0.10),
        minOrderAmount:     utils.GetValueFromMap(cfg, "min_order_amount", 10.0),
    }
}

func (s *OrderService) CreateOrder(userID, productID string, quantity int) (map[string]any, error) {
    // Must call getter methods (verbose)
    product := s.getProductService().GetProduct(productID)
    order := s.getRepo().Create(userID, productID, quantity, total)
    user := s.getUserService().GetUser(userID)
    s.getEmailService().Send(...)
    return order, nil
}
```

**Problems:**
- ❌ 60+ lines of boilerplate
- ❌ 8 struct fields for 4 dependencies
- ❌ Manual lazy loading logic (repeated 4 times)
- ❌ Verbose getter methods
- ❌ No compile-time type checking

---

### 🟢 LAYERED SERVICES - OrderService Factory (15 lines)

```go
type OrderService struct {
    // Generic lazy containers - type safe!
    repo           *service.Cached[OrderRepository]
    product        *service.Cached[ProductService]
    user           *service.Cached[UserService]
    email          *service.Cached[EmailService]
    
    // Config
    taxRate        float64
    minOrderAmount float64
}

func NewOrderService(cfg map[string]any) any {
    return &OrderService{
        repo:           service.GetLazyService[OrderRepository](cfg, "repository_service"),
        product:        service.GetLazyService[ProductService](cfg, "product_service"),
        user:           service.GetLazyService[UserService](cfg, "user_service"),
        email:          service.GetLazyService[EmailService](cfg, "email_service"),
        taxRate:        utils.GetValueFromMap(cfg, "tax_rate", 0.10),
        minOrderAmount: utils.GetValueFromMap(cfg, "min_order_amount", 10.0),
    }
}

func (s *OrderService) CreateOrder(userID, productID string, quantity int) (map[string]any, error) {
    // Direct access - clean and type-safe!
    product := s.product.MustGet().GetProduct(productID)
    order := s.repo.MustGet().Create(userID, productID, quantity, total)
    user := s.user.MustGet().GetUser(userID)
    s.email.MustGet().Send(...)
    return order, nil
}
```

**Benefits:**
- ✅ 15 lines (75% reduction!)
- ✅ 4 struct fields for 4 dependencies (simple)
- ✅ No manual lazy loading (handled by Lazy[T])
- ✅ No getter boilerplate
- ✅ Type-safe with generics
- ✅ Auto-caching with sync.Once

---

## Configuration Comparison

### 🔴 SIMPLE SERVICES Config

```yaml
services:
  - name: order-service
    type: order
    config:
      repository_service: order-repository    # String refs
      product_service: product-service        # No validation
      user_service: user-service              # Hidden dependencies
      email_service: email-service            # Can't see architecture
```

**Problems:**
- ❌ Flat structure - no layer visibility
- ❌ Dependencies implicit (only in factory code)
- ❌ No validation - typos cause runtime errors
- ❌ Can't see service dependencies without reading code

---

### 🟢 LAYERED SERVICES Config

```yaml
services:
  domain:
    - name: order-service
      type: order
      depends-on:                           # ✅ Explicit dependencies
        - order-repository                  # ✅ Validated at load time
        - product-service                   # ✅ Must be in config
        - user-service                      # ✅ Architecture visible
        - email-service
      config:
        repository_service: order-repository # ✅ Injected as Lazy
        product_service: product-service
        user_service: user-service
        email_service: email-service
```

**Benefits:**
- ✅ Layered structure - clear architecture
- ✅ Explicit dependencies in config
- ✅ Validation at load time
- ✅ Dependencies visible without code
- ✅ Layer violations detected
- ✅ Unused dependencies detected

---

## Metrics Summary

| Metric | Simple | Layered | Improvement |
|--------|--------|---------|-------------|
| **Lines per service** | 60+ | 15 | **75% less** |
| **Struct fields (4 deps)** | 8 | 4 | **50% less** |
| **Manual getters** | 4 | 0 | **No getters** |
| **Type casts** | Required | None | **Type-safe** |
| **Boilerplate** | High | Minimal | **Clean** |
| **Architecture visibility** | Code only | Config | **Visible** |
| **Validation** | None | Automated | **Safe** |

---

## Real-World Example

### Service with 10 Dependencies

**Simple Pattern:**
- 20 struct fields (10 names + 10 caches)
- 10 getter methods (~150 lines)
- 10 manual GetService calls
- Total: **~180 lines of boilerplate**

**Layered Pattern:**
- 10 struct fields (just Lazy containers)
- 0 getter methods
- 10 GetLazyService calls
- Total: **~30 lines** (**83% reduction!**)

---

## Try Both!

```bash
# Run with simple services
go run . simple

# Run with layered services  
go run . layered

# Same API, same behavior - just different patterns!
curl http://localhost:8080/api/products
curl http://localhost:8080/api/orders -d '{"user_id":"123","product_id":"1","quantity":2}'
```
