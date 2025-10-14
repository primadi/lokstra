# Dependency Injection Pattern

## Overview

Lokstra provides automatic dependency injection through the `depends-on` configuration field. This document explains how dependencies are injected and how to use them in your services.

## How It Works

### 1. Declare Dependencies in Config

```yaml
services:
  - name: order-service
    type: order_service
    depends-on: [user-service, payment-service]  # ← Dependencies declared here
    config:
      user_service: user-service      # ← Will be auto-injected
      payment_service: payment-service
      tax_rate: 0.10
```

### 2. System Auto-Injects Dependencies

When `RegisterConfig()` is called, the system:

1. **Builds dependency map** - creates lazy service placeholders
2. **Injects into config** - replaces service name strings with lazy references
3. **Validates** - ensures all dependencies exist and are used

```go
// Internal flow (you don't need to do this):
depMap := buildDependencyMap(services)
// depMap["user-service"] = genericLazyService{name: "user-service"}

injectedCfg := injectDependencies(svc, depMap)
// injectedCfg["user_service"] = genericLazyService (not string anymore!)
```

### 3. Extract in Factory with `Dep[T]`

Use the short, clean `Dep[T]` helper to extract dependencies:

```go
func CreateOrderService(cfg map[string]any) any {
    return &OrderService{
        userSvc:    lokstra_registry.Dep[UserService](cfg, "user_service"),
        paymentSvc: lokstra_registry.Dep[PaymentService](cfg, "payment_service"),
        taxRate:    utils.GetValueFromMap(cfg, "tax_rate", 0.10),
    }
}
```

### 4. Use Lazily in Service Methods

```go
type OrderService struct {
    userSvc    *lokstra_registry.Lazy[UserService]
    paymentSvc *lokstra_registry.Lazy[PaymentService]
    taxRate    float64
}

func (s *OrderService) CreateOrder(ctx *request.Context, req *CreateOrderRequest) (*Order, error) {
    // Services are loaded lazily on first .Get() call
    user := s.userSvc.Get()
    payment := s.paymentSvc.Get()
    
    // Use the services...
    userInfo, err := user.GetUser(ctx, &GetUserRequest{ID: req.UserID})
    // ...
}
```

## API Reference

### `Dep[T](cfg, key)` ⭐ Recommended

Short, clean way to get service dependencies.

```go
userSvc := lokstra_registry.Dep[UserService](cfg, "user_service")
```

**Returns:** `*Lazy[T]` or `nil` if not found

### `MustDep[T](cfg, key)` 

Like `Dep` but panics if dependency is missing. Use for required dependencies.

```go
userSvc := lokstra_registry.MustDep[UserService](cfg, "user_service")
// Panics if "user_service" not in config
```

### `GetLazyService[T](cfg, key)` (Deprecated)

Legacy name for `Dep[T]`. Still works for backward compatibility.

```go
userSvc := lokstra_registry.GetLazyService[UserService](cfg, "user_service")
```

### `MustGetLazyService[T](cfg, key)` (Deprecated)

Legacy name for `MustDep[T]`.

## Benefits

### ✅ Type Safety

```go
// Compile-time type checking
userSvc := lokstra_registry.Dep[UserService](cfg, "user_service")
// userSvc is *Lazy[UserService], not any
```

### ✅ Lazy Loading

Services are only instantiated when first used:

```go
func (s *OrderService) CreateOrder(...) {
    user := s.userSvc.Get()  // ← Service created here (first call only)
    // Subsequent calls use cached instance
}
```

### ✅ No Manual Caching

No need to write boilerplate caching logic:

```go
// ❌ OLD WAY (80+ lines of boilerplate):
type OrderService struct {
    userServiceName string
    userServiceCache *UserService
}

func (s *OrderService) getUserService() *UserService {
    if s.userServiceCache == nil {
        s.userServiceCache = lokstra_registry.GetService[UserService](s.userServiceName, nil)
    }
    return s.userServiceCache
}

// ✅ NEW WAY (1 line):
type OrderService struct {
    userSvc *lokstra_registry.Lazy[UserService]
}
// Caching handled automatically!
```

### ✅ Validation at Load Time

System validates:
- All services in `depends-on` exist
- All `depends-on` are used in config
- All config service references are in `depends-on`
- No circular dependencies

```yaml
services:
  - name: order-service
    depends-on: [user-service, payment-service]
    config:
      user_service: user-service
      # ❌ ERROR: payment-service in depends-on but not used in config
```

### ✅ Architecture Visibility

Dependencies are explicit and visible in config:

```yaml
depends-on: [user-service, payment-service]  # 👈 Clear dependency graph
```

## Best Practices

### 1. Always Declare Dependencies

```yaml
# ✅ GOOD: Explicit dependencies
services:
  - name: order-service
    depends-on: [user-service, payment-service]
    config:
      user_service: user-service
      payment_service: payment-service

# ❌ BAD: Hidden dependencies
services:
  - name: order-service
    config:
      user_service: user-service  # No depends-on declaration
```

### 2. Use `Dep[T]` for Optional Dependencies

```go
func CreateOrderService(cfg map[string]any) any {
    emailSvc := lokstra_registry.Dep[EmailService](cfg, "email_service")
    // emailSvc can be nil - handle gracefully
    
    return &OrderService{
        emailSvc: emailSvc,
    }
}
```

### 3. Use `MustDep[T]` for Required Dependencies

```go
func CreateOrderService(cfg map[string]any) any {
    return &OrderService{
        userSvc: lokstra_registry.MustDep[UserService](cfg, "user_service"),
        // Panics at startup if missing - fail fast!
    }
}
```

### 4. Keep Service Structs Simple

```go
// ✅ GOOD: Clean, minimal
type OrderService struct {
    userSvc    *lokstra_registry.Lazy[UserService]
    paymentSvc *lokstra_registry.Lazy[PaymentService]
    taxRate    float64
}

// ❌ BAD: Boilerplate caching logic
type OrderService struct {
    userServiceName string
    userServiceCache *UserService
    userServiceMutex sync.RWMutex
    // ... 50 more lines
}
```

## Layered Services

For complex applications, use layered services for automatic ordering:

```yaml
services:
  infrastructure:  # Layer 1
    - name: db-service
    - name: cache-service
  
  repository:     # Layer 2 - depends on layer 1
    - name: user-repo
      depends-on: [db-service]
      config:
        db_service: db-service
  
  domain:         # Layer 3 - depends on layer 2
    - name: user-service
      depends-on: [user-repo]
      config:
        repository_service: user-repo
```

## Troubleshooting

### "service not found" Error

```
panic: service user-service not found or type mismatch
```

**Solutions:**
1. Check service name spelling in `depends-on` and config
2. Ensure service is declared in `services` section
3. Verify service factory is registered

### Type Mismatch Error

```go
userSvc := lokstra_registry.Dep[WrongType](cfg, "user_service")
// Runtime error when .Get() is called
```

**Solution:** Ensure the generic type matches the actual service type.

### Circular Dependency

```yaml
services:
  - name: service-a
    depends-on: [service-b]
  - name: service-b
    depends-on: [service-a]  # ❌ Circular!
```

**Solution:** Refactor to break the cycle. Consider:
- Extracting shared logic to a third service
- Using events/callbacks instead of direct dependencies
- Reviewing your architecture

## Migration Guide

### From Manual String References

```go
// ❌ OLD:
type OrderService struct {
    userServiceName string
    userServiceCache *UserService
}

func (s *OrderService) getUser() *UserService {
    if s.userServiceCache == nil {
        s.userServiceCache = lokstra_registry.GetService[UserService](s.userServiceName, nil)
    }
    return s.userServiceCache
}

// ✅ NEW:
type OrderService struct {
    userSvc *lokstra_registry.Lazy[UserService]
}

func (s *OrderService) CreateOrder(...) {
    user := s.userSvc.Get()  // One line!
}
```

### From `GetLazyService` to `Dep`

```go
// ❌ OLD (verbose):
userSvc := lokstra_registry.GetLazyService[UserService](cfg, "user_service")

// ✅ NEW (concise):
userSvc := lokstra_registry.Dep[UserService](cfg, "user_service")
```

Both work, but `Dep` is shorter and cleaner.

## See Also

- [Configuration Guide](./configuration.md)
- [Service Guide](./services.md)
- [Layered Services Example](../cmd/examples/25-single-binary-deployment/)
