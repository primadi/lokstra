# Lazy Loading Implementation Summary

## Overview
Successfully refactored `core/deploy` package to use **lazy dependency injection** via `service.Cached[T]` pattern. This eliminates initialization order problems and circular dependency risks.

## Key Changes

### 1. Enhanced `service.Cached[T]` (`core/service/lazy_load.go`)
Added support for custom loaders:
```go
type Cached[T any] struct {
    serviceName string
    loader      func() T  // NEW: Custom loader function
    once        sync.Once
    cache       T
}

// Create lazy reference with custom loader
func LazyLoadWith[T any](loader func() T) *Cached[T]

// Create pre-loaded value
func Value[T any](value T) *Cached[T]
```

When `.Get()` is called:
1. Check if custom `loader` exists → use it
2. Otherwise fall back to `lokstra_registry.Get(serviceName)`

### 2. Updated Factory Signature (`core/deploy/registry.go`)
**Before:**
```go
type ServiceFactory func(deps map[string]any, config map[string]any) any
// Services eagerly resolved all dependencies
```

**After:**
```go
type ServiceFactory func(deps map[string]any, config map[string]any) any
// deps now contain *service.Cached[any] instead of resolved instances
// Comment updated to indicate lazy loading
```

### 3. Updated `instantiateService` (`core/deploy/deployment.go`)
**Before - Eager Loading:**
```go
deps := make(map[string]any)
for _, depStr := range serviceDef.DependsOn {
    paramName, serviceName := parseDependency(depStr)
    
    // PROBLEM: Recursive GetService() call causes initialization order issues
    depInstance, err := a.GetService(serviceName)
    if err != nil {
        return nil, err
    }
    
    deps[paramName] = depInstance  // Eager - already resolved
}
```

**After - Lazy Loading:**
```go
deps := make(map[string]any)
for _, depStr := range serviceDef.DependsOn {
    paramName, serviceName := parseDependency(depStr)
    
    // Create lazy loader - no immediate resolution
    lazyDep := service.LazyLoadWith(func() any {
        depInstance, err := a.GetService(serviceName)
        if err != nil {
            panic(fmt.Sprintf("failed to resolve lazy dependency %s: %v",
                serviceName, err))
        }
        return depInstance
    })
    
    deps[paramName] = lazyDep  // Lazy - resolved on first .Get()
}
```

## Factory Pattern

### Service Struct Pattern
Services store lazy references:
```go
type UserService struct {
    DB     *service.Cached[any]  // Lazy-loaded
    Logger *service.Cached[any]  // Lazy-loaded
}
```

### Factory Pattern
Factories receive and store lazy references (don't call `.Get()` during construction):
```go
func userServiceFactory(deps map[string]any, config map[string]any) any {
    return &UserService{
        DB:     deps["db"].(*service.Cached[any]),      // Store lazy ref
        Logger: deps["logger"].(*service.Cached[any]),  // Store lazy ref
    }
}
```

### Usage Pattern
Services call `.Get()` only when actually needed:
```go
func (us *UserService) GetUser(id int) *User {
    // Resolve lazy-loaded logger only when method is called
    logger := us.Logger.Get().(*Logger)
    logger.Info(fmt.Sprintf("Getting user %d", id))
    
    return &User{ID: id, Name: "John Doe"}
}
```

## Benefits

### 1. **No Initialization Order Issues**
- Dependencies are not resolved during service construction
- Services can reference each other without worrying about initialization order
- Circular dependencies become possible (though still not recommended)

### 2. **On-Demand Resolution**
- Dependencies are only instantiated when actually used
- Unused dependencies never get instantiated
- Better resource management

### 3. **Cached After First Access**
- `sync.Once` ensures each dependency is instantiated only once
- Subsequent `.Get()` calls return cached instance
- Thread-safe

### 4. **Clear Dependency Flow**
```
1. Factory called with lazy deps
   ↓
2. Service stores lazy refs (no resolution yet)
   ↓
3. Service method called
   ↓
4. .Get() resolves dependency (first time only)
   ↓
5. Cached for future calls
```

## Test Results

All tests passing ✅:
```
=== Deployment Tests (core/deploy) ===
✅ TestDeployment_Creation
✅ TestDeployment_ConfigOverrides
✅ TestDeployment_ServerCreation
✅ TestDeployment_AppCreation
✅ TestApp_AddService
✅ TestApp_GetService_Simple
✅ TestApp_GetService_WithDependencies
✅ TestApp_GetService_WithAliases
✅ TestApp_FluentAPI
✅ TestParseDependency
✅ TestApp_ServiceNotFound
✅ TestApp_MissingDependency (updated for lazy behavior)
✅ TestGlobalRegistry_ConfigResolution
✅ TestGlobalRegistry_ConfigReference
✅ TestGlobalRegistry_ServiceDefinition
✅ TestGlobalRegistry_RouterOverride
✅ TestGlobalRegistry_FactoryRegistration
✅ TestGlobalRegistry_MiddlewareFactory
✅ TestGlobalSingleton

=== Resolver Tests (core/deploy/resolver) ===
✅ All 12 resolver tests passing

TOTAL: 31 tests passing
```

## Example Output

The `examples/basic/main.go` demonstrates lazy loading:
```
🚀 Lokstra Deploy API Example
🔧 Registering service factories...
⚙️  Defining configurations...
📋 Defining services...
✨ Creating deployment...
🖥️  Creating server...
📱 Creating app on port 3000...
➕ Adding services to app...
🏗️  Building deployment...
🔨 Instantiating services...

✅ All services instantiated!

🎯 Using services...
📝 Logger initialized (level: info)      ← Logger lazy-loaded here
ℹ️  [info] Getting user 1
👤 Got user: &{ID:1 Name:John Doe}
```

## Migration Guide

### For Existing Service Factories

**Before (Eager):**
```go
type MyService struct {
    DB     *DBPool
    Logger *Logger
}

func myServiceFactory(deps map[string]any, config map[string]any) any {
    return &MyService{
        DB:     deps["db"].(*DBPool),      // Eager resolution
        Logger: deps["logger"].(*Logger),  // Eager resolution
    }
}

func (s *MyService) DoWork() {
    s.Logger.Info("Working...")
}
```

**After (Lazy):**
```go
type MyService struct {
    DB     *service.Cached[any]  // Lazy reference
    Logger *service.Cached[any]  // Lazy reference
}

func myServiceFactory(deps map[string]any, config map[string]any) any {
    return &MyService{
        DB:     deps["db"].(*service.Cached[any]),      // Store lazy ref
        Logger: deps["logger"].(*service.Cached[any]),  // Store lazy ref
    }
}

func (s *MyService) DoWork() {
    logger := s.Logger.Get().(*Logger)  // Resolve when needed
    logger.Info("Working...")
}
```

## Important Notes

### ⚠️ Panic on Lazy Resolution Errors
When a lazy dependency fails to resolve (e.g., dependency not found), it **panics** instead of returning an error. This is by design because:
1. Lazy resolution happens inside service methods, not during instantiation
2. Methods typically don't have error returns for dependency resolution
3. Missing dependencies indicate configuration errors that should fail fast

### ✅ Error Detection During Instantiation
The service instantiation itself succeeds even if dependencies don't exist yet. Errors only occur when `.Get()` is called. This is the expected lazy loading behavior.

### 🔧 For Services Without Dependencies
Services that don't have dependencies work exactly as before:
```go
func dbPoolFactory(deps map[string]any, config map[string]any) any {
    pool := &DBPool{
        DSN:      config["dsn"].(string),
        MaxConns: config["max-conns"].(int),
    }
    pool.Connect()  // Can still do immediate initialization
    return pool
}
```

## Next Steps
- ✅ Phase 1: Config resolver + Registry (19 tests)
- ✅ Phase 2: Deployment API with lazy DI (12 tests)
- ⏳ Phase 3: YAML parser (upcoming)
- ⏳ Integration with existing Lokstra app initialization
