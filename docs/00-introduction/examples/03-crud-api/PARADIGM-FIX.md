# Paradigm Fix: Remove lokstra_registry Dependency

## 🎯 Problem Identified

Original implementation had **paradigm confusion**:

### ❌ WRONG Approach (Before)
```go
// runWithConfig() was registering to OLD lokstra_registry
lokstra_registry.RegisterService("database", dbRaw)
lokstra_registry.RegisterService("user-service", userServiceRaw)

// Handlers were looking up from lokstra_registry
var userService = service.LazyLoadWith(func() *UserService {
    return lokstra_registry.GetInstance("user-service").(*UserService)
})
```

**Issues:**
1. ❌ Using OLD paradigm (`lokstra_registry`) in NEW paradigm code
2. ❌ Double registration: both `deploy.Registry` AND `lokstra_registry`
3. ❌ Unnecessary complexity: lazy load from registry when we already have instance
4. ❌ Mixing two different systems

---

## ✅ CORRECT Approach (After)

### Key Principle
> **Handlers access services via package-level variables, NOT from registry**

### Implementation

#### 1️⃣ **Package-level variable (handlers.go)**
```go
// Package-level cached service
// - Set once in main() before starting server
// - Accessed directly by handlers (no registry lookup)
// - Thread-safe via service.Cached[T]
var userService *service.Cached[*UserService]
```

#### 2️⃣ **Mode 1: Manual Instantiation (runWithCode)**
```go
func runWithCode() {
    // 1. Create services manually
    db := NewDatabase()
    userSvc := &UserService{
        DB: service.Value(db),
    }

    // 2. Set package variable for handlers
    userService = service.Value(userSvc)

    // 3. Setup router and run
    setupRouterAndRun()
}
```

**No lokstra_registry involved!** ✅

#### 3️⃣ **Mode 2: YAML Configuration (runWithConfig)**
```go
func runWithConfig() {
    // 1. Get NEW deploy registry
    reg := deploy.Global()

    // 2. Register service factories to NEW registry
    reg.RegisterServiceType("database-factory", DatabaseFactory, nil)
    reg.RegisterServiceType("user-service-factory", UserServiceFactory, nil)

    // 3. Load from YAML (uses NEW deploy registry)
    dep, err := loader.LoadAndBuild(
        []string{"config.yaml"},
        "development",
        reg,
    )

    // 4. Get service instance
    server, _ := dep.GetServer("api")
    deployApp := server.Apps()[0]
    userServiceRaw, err := deployApp.GetService("user-service")

    // 5. Set package variable for handlers
    userService = service.Value(userServiceRaw.(*UserService))

    // 6. Setup router and run
    setupRouterAndRun()
}
```

**No lokstra_registry involved!** ✅

#### 4️⃣ **Handlers (handlers.go)**
```go
func listUsersHandler(ctx *request.Context) error {
    // Access service directly from package variable
    users, err := userService.MustGet().GetAll()
    if err != nil {
        return ctx.Api.Error(500, "INTERNAL_ERROR", err.Error())
    }
    return ctx.Api.Ok(users)
}
```

**Direct access, no registry lookup!** ✅

---

## 🔑 Key Differences: OLD vs NEW Paradigm

| Aspect | OLD Paradigm (lokstra_registry) | NEW Paradigm (deploy.Registry) |
|--------|--------------------------------|-------------------------------|
| **Registration** | `lokstra_registry.RegisterService()` | `deploy.Global().RegisterServiceType()` |
| **Config Loading** | Manual parsing + registration | `loader.LoadAndBuild()` with YAML |
| **Service Access** | Runtime lookup: `lokstra_registry.GetInstance()` | Direct access via package variable |
| **Lazy Loading** | Via registry lookup on every call | Via `service.Cached[T]` set once at startup |
| **Type Safety** | Type assertion on every access | Type-safe `Cached[T]` |
| **Performance** | Registry lookup + mutex lock every call | Zero overhead after first access |
| **Dependency Injection** | Manual wiring | Automatic via YAML config |
| **Thread Safety** | Manual locking required | Built-in via `sync.Once` |

---

## 📝 Migration Checklist

When updating examples to NEW paradigm:

- [ ] ❌ Remove `lokstra_registry` import
- [ ] ❌ Remove `lokstra_registry.RegisterService()` calls
- [ ] ❌ Remove `lokstra_registry.GetInstance()` calls
- [ ] ✅ Use `deploy.Global()` for registry
- [ ] ✅ Use `service.Value()` or `service.Cached[T]` for instances
- [ ] ✅ Set package variables in `main()` before starting server
- [ ] ✅ Access services directly in handlers (no registry lookup)

---

## 🎯 Benefits of NEW Paradigm

1. **Clear Separation**: Deploy registry for configuration, package variables for runtime
2. **Better Performance**: No registry lookup overhead in request handlers
3. **Type Safety**: `service.Cached[T]` provides compile-time type checking
4. **Simpler Code**: Direct access is more readable than registry lookups
5. **YAML Support**: Full configuration-driven deployment
6. **Lazy DI**: Automatic dependency resolution from YAML
7. **No Double Registration**: Single source of truth (deploy.Registry)

---

## 🚀 Testing

Both modes should work without errors:

```bash
# Test manual mode
go run . -mode code

# Test YAML mode
go run . -mode config
```

Both should:
- ✅ Start server successfully
- ✅ Respond to API requests
- ✅ Handle CRUD operations
- ✅ Use same handlers (shared code)
- ✅ No dependency on lokstra_registry

---

## 📌 Important Notes

### For Future Examples

**DO:**
- ✅ Use `deploy.Global()` for service registration
- ✅ Use `service.Cached[T]` for type-safe lazy loading
- ✅ Set package variables once at startup
- ✅ Access services directly in handlers

**DON'T:**
- ❌ Use `lokstra_registry` (old paradigm, will be removed)
- ❌ Do registry lookups in handlers (performance overhead)
- ❌ Mix old and new paradigms
- ❌ Register same service to multiple registries

### Deprecation Notice

`lokstra_registry` package is **LEGACY** and will be removed in future versions.

All new code should use:
- `core/deploy` for deployment configuration
- `core/service` for lazy loading
- Package variables for service access

---

## ✅ Result

**CLEAN SEPARATION between paradigms:**

```
OLD Paradigm (lokstra_registry):
- Manual registration
- Runtime lookups
- Will be deprecated

NEW Paradigm (core/deploy + core/service):
- YAML configuration
- Factory registration
- Package-level instances
- Type-safe lazy loading
- Production-ready ✅
```

This example now demonstrates the NEW paradigm correctly! 🎉
