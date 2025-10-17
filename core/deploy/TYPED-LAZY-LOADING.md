# Typed Lazy Loading Pattern

## Overview
The deployment system uses **typed lazy loading** with `service.Cached[T]` for type-safe, lazy dependency injection. This eliminates runtime type assertions and provides compile-time type safety.

## Key Features

### ✅ Type Safety
No runtime type assertions needed - the compiler enforces types:

```go
// ❌ OLD: Runtime type assertion required
type UserService struct {
    DB     *service.Cached[any]  // Could be anything!
    Logger *service.Cached[any]
}

func (us *UserService) DoWork() {
    db := us.DB.Get().(*DBPool)      // Runtime panic if wrong type!
    logger := us.Logger.Get().(*Logger)  // Runtime panic if wrong type!
}

// ✅ NEW: Compile-time type safety
type UserService struct {
    DB     *service.Cached[*DBPool]  // Compiler knows it's *DBPool
    Logger *service.Cached[*Logger]  // Compiler knows it's *Logger
}

func (us *UserService) DoWork() {
    db := us.DB.Get()      // Returns *DBPool, no cast needed!
    logger := us.Logger.Get()  // Returns *Logger, no cast needed!
}
```

### ✅ Helper Functions

#### `service.Cast[T](value any) *Cached[T]`
Converts `map[string]any` dependency to typed `Cached[T]`:

```go
func userServiceFactory(deps map[string]any, config map[string]any) any {
    return &UserService{
        // ❌ OLD: Manual type assertion
        DB:     deps["db"].(*service.Cached[any]),
        Logger: deps["logger"].(*service.Cached[any]),
        
        // ✅ NEW: Helper function with type inference
        DB:     service.Cast[*DBPool](deps["db"]),
        Logger: service.Cast[*Logger](deps["logger"]),
    }
}
```

#### `MustGet() T`
Fail-fast variant of `Get()` - panics if service is nil:

```go
// Get() - returns value (could be nil)
db := us.DB.Get()  // Returns *DBPool (could be nil)

// MustGet() - panics if nil (fail-fast)
db := us.DB.MustGet()  // Returns *DBPool or panics
```

Use `MustGet()` when:
- Service is required (not optional)
- You want to fail fast if dependency is missing
- Service initialization should stop if dependency is unavailable

## Complete Example

### Service Definition
```go
type UserService struct {
    DB     *service.Cached[*DBPool]  // Typed lazy reference
    Logger *service.Cached[*Logger]  // Typed lazy reference
}

func (us *UserService) GetUser(id int) *User {
    // Type-safe access - no casting needed!
    logger := us.Logger.Get()
    logger.Info(fmt.Sprintf("Getting user %d", id))
    
    // Or use MustGet for fail-fast
    db := us.DB.MustGet()
    return db.QueryUser(id)
}
```

### Factory Function
```go
func userServiceFactory(deps map[string]any, config map[string]any) any {
    return &UserService{
        DB:     service.Cast[*DBPool](deps["db"]),      // Type-safe helper
        Logger: service.Cast[*Logger](deps["logger"]),  // Type-safe helper
    }
}
```

### Registry Definition
```go
reg.DefineService(&schema.ServiceDef{
    Name:      "user-service",
    Type:      "user-factory",
    DependsOn: []string{"db:db-pool", "logger"},  // Alias syntax
})
```

## Benefits

### 1. **Compile-Time Type Safety**
```go
// Compiler error if you try to use wrong type
logger := us.Logger.Get()
logger.Debug("test")  // ✅ Compiler knows Logger has Debug()

// vs old way
logger := us.Logger.Get().(*Logger)  // Could panic at runtime!
```

### 2. **Better IDE Support**
- Auto-completion works immediately after `.Get()`
- No need to remember concrete types
- Refactoring is safer

### 3. **Cleaner Code**
```go
// ❌ OLD: 2 operations (get + cast)
db := us.DB.Get().(*DBPool)

// ✅ NEW: 1 operation (get)
db := us.DB.Get()
```

### 4. **Fail-Fast with MustGet**
```go
// Critical dependencies should use MustGet
db := us.DB.MustGet()  // Panics immediately if missing

// Optional dependencies can use Get
cache := us.Cache.Get()  // Returns nil if missing, no panic
if cache != nil {
    cache.Set(key, value)
}
```

## Migration Guide

### Step 1: Update Service Struct
```go
// Before
type MyService struct {
    DB     *service.Cached[any]
    Logger *service.Cached[any]
}

// After
type MyService struct {
    DB     *service.Cached[*DBPool]
    Logger *service.Cached[*Logger]
}
```

### Step 2: Update Factory
```go
// Before
func myServiceFactory(deps map[string]any, config map[string]any) any {
    return &MyService{
        DB:     deps["db"].(*service.Cached[any]),
        Logger: deps["logger"].(*service.Cached[any]),
    }
}

// After
func myServiceFactory(deps map[string]any, config map[string]any) any {
    return &MyService{
        DB:     service.Cast[*DBPool](deps["db"]),
        Logger: service.Cast[*Logger](deps["logger"]),
    }
}
```

### Step 3: Update Service Methods
```go
// Before
func (s *MyService) DoWork() {
    db := s.DB.Get().(*DBPool)      // Manual cast
    logger := s.Logger.Get().(*Logger)  // Manual cast
}

// After
func (s *MyService) DoWork() {
    db := s.DB.Get()      // No cast needed!
    logger := s.Logger.Get()  // No cast needed!
    
    // Or use MustGet for required dependencies
    db := s.DB.MustGet()  // Fail-fast
}
```

## Advanced Patterns

### Optional Dependencies
```go
type MyService struct {
    DB    *service.Cached[*DBPool]  // Required
    Cache *service.Cached[*Cache]   // Optional
}

func (s *MyService) DoWork() {
    db := s.DB.MustGet()  // Panics if missing - required!
    
    cache := s.Cache.Get()  // Returns nil if missing - optional
    if cache != nil {
        cache.Set("key", "value")
    }
}
```

### Multiple Instances of Same Type
```go
type OrderService struct {
    DBOrder *service.Cached[*DBPool]  // Order database
    DBUser  *service.Cached[*DBPool]  // User database
}

func orderServiceFactory(deps map[string]any, config map[string]any) any {
    return &OrderService{
        DBOrder: service.Cast[*DBPool](deps["dbOrder"]),  // Alias: dbOrder:db-order
        DBUser:  service.Cast[*DBPool](deps["dbUser"]),   // Alias: dbUser:db-user
    }
}
```

## Implementation Details

### `service.Cast[T]` Implementation
```go
func Cast[T any](value any) *Cached[T] {
    if cached, ok := value.(*Cached[any]); ok {
        // Wrap the any Cached with a typed loader
        return LazyLoadWith(func() T {
            return cached.Get().(T)
        })
    }
    // If it's already the right type, return as-is
    if cached, ok := value.(*Cached[T]); ok {
        return cached
    }
    panic("Cast: value is not a *Cached type")
}
```

The cast creates a new typed `Cached[T]` that wraps the original `Cached[any]`. When `.Get()` is called:
1. Outer `Cached[T]` calls its loader
2. Loader calls inner `Cached[any].Get()`
3. Result is type-asserted to `T` once
4. Cached for subsequent calls

### Thread Safety
Both `Get()` and `MustGet()` use `sync.Once` internally:
```go
func (l *Cached[T]) Get() T {
    l.once.Do(func() {
        if l.loader != nil {
            l.cache = l.loader()  // Called once
        } else {
            l.cache = lokstra_registry.GetService[T](l.serviceName)
        }
    })
    return l.cache  // Subsequent calls return cached value
}
```

## Testing

All 31 tests pass with typed pattern:

```bash
$ go test -v
=== RUN   TestApp_GetService_WithDependencies
--- PASS: TestApp_GetService_WithDependencies (0.00s)
=== RUN   TestApp_GetService_WithAliases
--- PASS: TestApp_GetService_WithAliases (0.00s)
...
PASS
ok      github.com/primadi/lokstra/core/deploy  1.069s
```

## Best Practices

### ✅ DO:
- Use typed `Cached[T]` for all dependencies
- Use `service.Cast[T]()` in factories
- Use `MustGet()` for required dependencies
- Use `Get()` for optional dependencies
- Keep dependency types explicit in struct definitions

### ❌ DON'T:
- Don't use `Cached[any]` unless absolutely necessary
- Don't mix `Get()` and `MustGet()` for same dependency
- Don't call `.Get()` during factory - store the lazy reference
- Don't forget to handle nil from `Get()` if dependency is optional

## Summary

The typed lazy loading pattern provides:
- ✅ **Type Safety**: Compile-time type checking
- ✅ **Clean Code**: No manual type assertions
- ✅ **Better DX**: IDE auto-completion works perfectly
- ✅ **Fail-Fast**: `MustGet()` for required dependencies
- ✅ **Performance**: Zero overhead vs manual casting
- ✅ **Maintainability**: Refactoring is safer with types

This is now the recommended pattern for all service dependencies in Lokstra!
