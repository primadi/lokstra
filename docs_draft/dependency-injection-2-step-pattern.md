# 2-Step Dependency Injection Pattern

## 🎯 Problem Statement

The original `GetLazyService` API was confusing for new developers:

```go
// ❌ Confusing: What is cfg? Service config or dependency config?
userService: lokstra_registry.MustGetLazyService[UserService](cfg, "user-service")
```

**Why confusing?**
1. ❌ Parameter `cfg` looks like it's for configuring the service being created
2. ❌ Actually `cfg` only contains dependency references (from YAML `depends-on`)
3. ❌ The key is manually typed string (typo-prone)
4. ❌ Doesn't make the dependency injection pattern explicit

## ✅ Solution: 2-Step Pattern

Make the intent explicit with two clear steps:

```go
// Step 1: Create Dependencies helper (makes it clear cfg contains deps)
deps := lokstra_registry.NewDependencies(cfg)

// Step 2: Load lazy dependencies (much clearer!)
userService: lokstra_registry.MustLazyLoad[UserService](deps, "user-service")
```

## 📖 Migration Guide

### Before (Old API)

```go
func CreateAuthServiceLocal(cfg map[string]any) any {
    jwtSecret := utils.GetValueFromMap(cfg, "jwt_secret", "default-secret")
    tokenExpiry := utils.GetValueFromMap(cfg, "token_expiry", 3600)

    return &authServiceLocal{
        jwtSecret:   jwtSecret,
        tokenExpiry: tokenExpiry,
        userService: lokstra_registry.MustGetLazyService[UserService](cfg, "user-service"),
        tokens:      make(map[string]string),
    }
}
```

### After (New 2-Step API)

```go
func CreateAuthServiceLocal(cfg map[string]any) any {
    jwtSecret := utils.GetValueFromMap(cfg, "jwt_secret", "default-secret")
    tokenExpiry := utils.GetValueFromMap(cfg, "token_expiry", 3600)

    // ✨ Step 1: Make it explicit that cfg contains dependencies
    deps := lokstra_registry.NewDependencies(cfg)

    return &authServiceLocal{
        jwtSecret:   jwtSecret,
        tokenExpiry: tokenExpiry,
        // ✨ Step 2: Load lazy dependencies
        userService: lokstra_registry.MustLazyLoad[UserService](deps, "user-service"),
        tokens:      make(map[string]string),
    }
}
```

## 🎨 Benefits

### 1. **Clearer Intent**
```go
deps := lokstra_registry.NewDependencies(cfg)  
// ✅ Clear: deps is for dependencies, not service config
```

### 2. **Better Readability**
```go
// Old: Hard to understand at first glance
userSvc := lokstra_registry.MustGetLazyService[UserService](cfg, "user-service")

// New: Immediately clear this is dependency injection
deps := lokstra_registry.NewDependencies(cfg)
userSvc := lokstra_registry.MustLazyLoad[UserService](deps, "user-service")
```

### 3. **Self-Documenting**
The code itself explains what's happening:
- `NewDependencies(cfg)` → Extract dependencies from config
- `LazyLoad[T](deps, key)` → Create lazy loader for dependency

### 4. **Easier to Debug**
```go
deps := lokstra_registry.NewDependencies(cfg)

// You can inspect dependencies before loading
if deps.HasDependency("user-service") {
    fmt.Println("User service available:", deps.GetServiceName("user-service"))
}

userSvc := lokstra_registry.LazyLoad[UserService](deps, "user-service")
```

## 📚 API Reference

### `NewDependencies(cfg map[string]any) *Dependencies`

Creates a Dependencies helper from factory config map.

```go
deps := lokstra_registry.NewDependencies(cfg)
```

### `LazyLoad[T](deps *Dependencies, key string) *Lazy[T]`

Creates a lazy-loading service dependency. Returns `nil` if not found.

```go
deps := lokstra_registry.NewDependencies(cfg)
userSvc := lokstra_registry.LazyLoad[UserService](deps, "user-service")
if userSvc == nil {
    // Handle optional dependency
}
```

### `MustLazyLoad[T](deps *Dependencies, key string) *Lazy[T]`

Like `LazyLoad` but panics if dependency is not found. Use for required dependencies.

```go
deps := lokstra_registry.NewDependencies(cfg)
userSvc := lokstra_registry.MustLazyLoad[UserService](deps, "user-service")
// Panics if "user-service" is not in cfg
```

### `Dependencies.HasDependency(key string) bool`

Check if a dependency exists in config.

```go
deps := lokstra_registry.NewDependencies(cfg)
if deps.HasDependency("cache-service") {
    cache := lokstra_registry.LazyLoad[CacheService](deps, "cache-service")
}
```

### `Dependencies.GetServiceName(key string) string`

Get the actual service name for a dependency key. Returns empty string if not found.

```go
deps := lokstra_registry.NewDependencies(cfg)
serviceName := deps.GetServiceName("user-service")
fmt.Println("Will load service:", serviceName) // e.g., "user_service"
```

## 🔄 Complete Example

### Service Factory

```go
package services

import (
    "github.com/primadi/lokstra/common/utils"
    "github.com/primadi/lokstra/lokstra_registry"
)

type OrderService struct {
    orderRepo    OrderRepository
    userService  *lokstra_registry.Lazy[UserService]
    paymentSvc   *lokstra_registry.Lazy[PaymentService]
    emailService *lokstra_registry.Lazy[EmailService]  // Optional
}

func CreateOrderService(cfg map[string]any) any {
    // Extract service-specific config
    maxRetries := utils.GetValueFromMap(cfg, "max_retries", 3)
    timeout := utils.GetValueFromMap(cfg, "timeout", 30)

    // ✨ Step 1: Create dependencies helper
    deps := lokstra_registry.NewDependencies(cfg)

    // ✨ Step 2: Load dependencies
    return &OrderService{
        orderRepo: NewOrderRepository(),
        // Required dependencies
        userService:  lokstra_registry.MustLazyLoad[UserService](deps, "user-service"),
        paymentSvc:   lokstra_registry.MustLazyLoad[PaymentService](deps, "payment-service"),
        // Optional dependency
        emailService: lokstra_registry.LazyLoad[EmailService](deps, "email-service"),
    }
}

func (s *OrderService) CreateOrder(userID, productID string) (*Order, error) {
    // Lazy-load and use dependencies
    user := s.userService.MustGet().GetUser(userID)
    
    order := &Order{
        UserID:    userID,
        ProductID: productID,
    }
    
    // Optional dependency handling
    if s.emailService != nil {
        emailSvc := s.emailService.Get()
        emailSvc.SendOrderConfirmation(user.Email, order)
    }
    
    return order, nil
}
```

### YAML Configuration

```yaml
services:
  order_service:
    type: local
    config:
      max_retries: 5
      timeout: 60
      depends-on:
        user-service: "user_service"      # Auto-injected into cfg
        payment-service: "payment_service"
        email-service: "email_service"
```

## 🚀 Migration Checklist

- [ ] Find all usages of `GetLazyService` in your codebase
- [ ] Replace with 2-step pattern:
  1. Add `deps := lokstra_registry.NewDependencies(cfg)`
  2. Replace `GetLazyService[T](cfg, key)` with `LazyLoad[T](deps, key)`
- [ ] Find all usages of `MustGetLazyService`
- [ ] Replace with `MustLazyLoad[T](deps, key)`
- [ ] Test that dependencies still load correctly
- [ ] Update internal documentation and comments

## 🔍 Finding Old API Usage

```bash
# Find GetLazyService usage
grep -r "GetLazyService" --include="*.go"

# Find MustGetLazyService usage
grep -r "MustGetLazyService" --include="*.go"
```

## 📝 Notes

- The old API (`GetLazyService`, `MustGetLazyService`) is deprecated but still works
- Old API will be removed in v2.0
- Migration is straightforward and improves code clarity
- The 2-step pattern is now the recommended approach

## ❓ FAQ

### Q: Why can't we use generic methods on Dependencies struct?

A: Go currently doesn't support generic methods on structs, only generic functions. That's why we use standalone functions like `LazyLoad[T](deps, key)` instead of `deps.LazyLoad[T](key)`.

### Q: Is there any performance difference?

A: No performance difference. The new API is just a clearer wrapper around the same underlying mechanism.

### Q: Can I mix old and new APIs?

A: Yes, but it's recommended to migrate fully for consistency.

### Q: What if I have many dependencies?

A: The 2-step pattern actually makes it cleaner:

```go
deps := lokstra_registry.NewDependencies(cfg)  // Once

return &MyService{
    dep1: lokstra_registry.LazyLoad[Service1](deps, "svc1"),
    dep2: lokstra_registry.LazyLoad[Service2](deps, "svc2"),
    dep3: lokstra_registry.LazyLoad[Service3](deps, "svc3"),
    dep4: lokstra_registry.LazyLoad[Service4](deps, "svc4"),
}
```

Much clearer than:
```go
return &MyService{
    dep1: lokstra_registry.GetLazyService[Service1](cfg, "svc1"),
    dep2: lokstra_registry.GetLazyService[Service2](cfg, "svc2"),
    dep3: lokstra_registry.GetLazyService[Service3](cfg, "svc3"),
    dep4: lokstra_registry.GetLazyService[Service4](cfg, "svc4"),
}
```
