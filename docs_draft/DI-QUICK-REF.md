# Dependency Injection - Quick Reference

## 🆕 Recommended: 2-Step Pattern

```go
// Step 1: Create Dependencies helper
deps := lokstra_registry.NewDependencies(cfg)

// Step 2: Load lazy dependencies
userSvc := lokstra_registry.MustLazyLoad[UserService](deps, "user-service")
```

### Why Better?

✅ **Explicit Intent**: Clear that `cfg` contains dependencies, not service config  
✅ **Self-Documenting**: Code explains what it does  
✅ **Less Confusing**: Especially for new developers  
✅ **Easier to Debug**: Can inspect dependencies before loading  

## 📚 Quick API

```go
// Create helper
deps := lokstra_registry.NewDependencies(cfg)

// Load optional dependency
svc := lokstra_registry.LazyLoad[Service](deps, "key")

// Load required dependency (panics if missing)
svc := lokstra_registry.MustLazyLoad[Service](deps, "key")

// Check if dependency exists
if deps.HasDependency("key") { ... }

// Get actual service name
name := deps.GetServiceName("key")
```

## 🔄 Migration from Old API

```go
// OLD (Deprecated)
userSvc := lokstra_registry.MustGetLazyService[UserService](cfg, "user-service")

// NEW (Recommended)
deps := lokstra_registry.NewDependencies(cfg)
userSvc := lokstra_registry.MustLazyLoad[UserService](deps, "user-service")
```

## 📖 Full Documentation

See: [dependency-injection-2-step-pattern.md](./dependency-injection-2-step-pattern.md)
