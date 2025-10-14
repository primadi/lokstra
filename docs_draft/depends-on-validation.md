# Depends-On Validation

## Overview

Starting from this version, Lokstra **strictly validates** the `depends-on` field in service configuration. This ensures:

1. ✅ **All dependencies must exist** - No typos or missing services
2. ✅ **All dependencies must be used** - No unused entries in `depends-on`
3. ✅ **All service references must be declared** - Config values referencing services must be in `depends-on`

## Validation Rules

### Rule 1: Dependencies Must Exist

```yaml
services:
  - name: user-service
    type: user_service
    
  - name: order-service
    type: order_service
    depends-on: [user-service, payment-service]  # ❌ ERROR: payment-service doesn't exist
```

**Error:**
```
service 'order-service' depends on 'payment-service' which does not exist
```

### Rule 2: Dependencies Must Be Used in Config

```yaml
services:
  - name: user-service
    type: user_service
    
  - name: order-service
    type: order_service
    depends-on: [user-service]  # ✅ In depends-on
    config:
      storage: memory
      # ❌ ERROR: user-service not referenced in config
```

**Error:**
```
service 'order-service': dependency 'user-service' in depends-on but not used in config
```

**Fix:**
```yaml
services:
  - name: order-service
    type: order_service
    depends-on: [user-service]
    config:
      user_service: user-service  # ✅ FIXED: Added to config
      storage: memory
```

### Rule 3: Service References Must Be in Depends-On

```yaml
services:
  - name: user-service
    type: user_service
    
  - name: order-service
    type: order_service
    depends-on: []  # ❌ Empty depends-on
    config:
      user_service: user-service  # ❌ ERROR: Not in depends-on
```

**Error:**
```
service 'order-service': config key 'user_service' references service 'user-service' 
which is not in depends-on. Add it to depends-on: [user-service]
```

**Fix:**
```yaml
services:
  - name: order-service
    type: order_service
    depends-on: [user-service]  # ✅ FIXED: Added to depends-on
    config:
      user_service: user-service
```

## Naming Convention

**Important:** Use **underscores** in config keys to match YAML convention:

```yaml
# ✅ CORRECT:
services:
  - name: order-service
    depends-on: [user-service]  # Service names use dashes
    config:
      user_service: user-service  # Config keys use underscores
```

```go
// ✅ CORRECT: Factory matches YAML key convention
func CreateOrderService(cfg map[string]any) any {
    return &OrderService{
        userSvc: lokstra_registry.Dep[UserService](cfg, "user_service"), // underscore
    }
}
```

```go
// ❌ WRONG: Mismatch between YAML and code
// YAML has: user_service: user-service
// Code expects: "user-service" (with dash)
userSvc: lokstra_registry.Dep[UserService](cfg, "user-service") // ❌ Won't work!
```

## Validation Applies To Both Modes

### Simple Mode (Flat Array)

```yaml
services:
  - name: db-service
    type: db
    
  - name: user-service
    type: user
    depends-on: [db-service]
    config:
      db_service: db-service  # ✅ Must match
```

### Layered Mode

```yaml
services:
  infrastructure:
    - name: db-service
      type: db
  
  domain:
    - name: user-service
      type: user
      depends-on: [db-service]
      config:
        db_service: db-service  # ✅ Must match
```

## Migration Guide

### Fixing "dependency in depends-on but not used" Error

**Before:**
```yaml
services:
  - name: auth-service
    depends-on: [user-service]
    config:
      jwt_secret: secret
      # Missing user_service reference!
```

**After:**
```yaml
services:
  - name: auth-service
    depends-on: [user-service]
    config:
      user_service: user-service  # ✅ Added
      jwt_secret: secret
```

### Fixing Factory Code

**Before (will cause MustDep panic):**
```go
func CreateAuthService(cfg map[string]any) any {
    return &AuthService{
        // ❌ YAML has "user_service", code looks for "user-service"
        userSvc: lokstra_registry.MustDep[UserService](cfg, "user-service"),
    }
}
```

**After:**
```go
func CreateAuthService(cfg map[string]any) any {
    return &AuthService{
        // ✅ Matches YAML key: user_service
        userSvc: lokstra_registry.Dep[UserService](cfg, "user_service"),
    }
}
```

## Common Patterns

### Optional Dependencies

Use `Dep[T]` for optional dependencies:

```go
func CreateOrderService(cfg map[string]any) any {
    emailSvc := lokstra_registry.Dep[EmailService](cfg, "email_service")
    // emailSvc can be nil - handle gracefully
    
    return &OrderService{
        emailSvc: emailSvc,
    }
}
```

```yaml
services:
  - name: order-service
    # email_service is optional - can omit from depends-on
    config:
      storage: memory
```

### Required Dependencies

Use `MustDep[T]` for required dependencies (fails fast):

```go
func CreateOrderService(cfg map[string]any) any {
    return &OrderService{
        userSvc: lokstra_registry.MustDep[UserService](cfg, "user_service"),
        // Panics at startup if missing
    }
}
```

```yaml
services:
  - name: order-service
    depends-on: [user-service]  # ✅ Required
    config:
      user_service: user-service  # ✅ Must be present
```

## Best Practices

### 1. Always Declare Dependencies

```yaml
# ✅ GOOD: Clear dependency graph
services:
  - name: order-service
    depends-on: [user-service, payment-service]
    config:
      user_service: user-service
      payment_service: payment-service
```

```yaml
# ❌ BAD: Hidden dependencies
services:
  - name: order-service
    config:
      user_service: user-service  # Dependency not visible
```

### 2. Keep Dependencies Minimal

```yaml
# ✅ GOOD: Only what's needed
services:
  - name: order-service
    depends-on: [user-service]
    config:
      user_service: user-service
```

```yaml
# ❌ BAD: Unnecessary dependencies
services:
  - name: order-service
    depends-on: [user-service, db-service, cache-service, logger]
    # If order-service only needs user-service, don't list others
```

### 3. Use Consistent Naming

**Service names:** `kebab-case` (user-service, payment-service)  
**Config keys:** `snake_case` (user_service, payment_service)

```yaml
services:
  - name: payment-service  # kebab-case
    depends-on: [user-service]  # kebab-case
    config:
      user_service: user-service  # snake_case key, kebab-case value
```

## Troubleshooting

### Error: "dependency in depends-on but not used"

**Cause:** Service listed in `depends-on` but not referenced in `config`.

**Solution:** Add the service reference to config:

```yaml
depends-on: [user-service]
config:
  user_service: user-service  # Add this line
```

### Error: "depends on 'X' which does not exist"

**Cause:** Service name typo or service not defined.

**Solution:** 
1. Check service name spelling
2. Ensure service is defined before it's used as a dependency

### Error: "MustDep: missing required dependency"

**Cause:** Factory code looks for a config key that doesn't exist.

**Solution:** Check key name matches between YAML and factory code:

```yaml
# YAML
config:
  user_service: user-service  # underscore
```

```go
// Factory - must match!
lokstra_registry.Dep[UserService](cfg, "user_service")  // underscore
```

### Validation Not Running

**Cause:** Using old config loading code that doesn't call `ValidateServices()`.

**Solution:** Use `lokstra_registry.RegisterConfig()` which includes validation:

```go
// ✅ GOOD: Includes validation
lokstra_registry.RegisterConfig(cfg, serverName)

// ❌ BAD: Manual registration without validation
for _, svc := range cfg.Services {
    lokstra_registry.RegisterLazyService(svc.Name, svc.Type, svc.Config)
}
```

## Implementation Details

Validation is performed in `processServices()` before service registration:

```go
func processServices(services *config.ServicesConfig) error {
    // Validate first - fail fast if config is invalid
    if err := config.ValidateServices(services); err != nil {
        return fmt.Errorf("service validation failed: %w", err)
    }
    
    // Then process services...
}
```

This ensures **fail-fast** behavior - configuration errors are caught at startup, not at runtime.

## See Also

- [Dependency Injection Guide](./dependency-injection.md)
- [Configuration Guide](./configuration.md)
- [Service Configuration](./services.md)
