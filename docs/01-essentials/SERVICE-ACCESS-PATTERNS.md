# Service Access Patterns - Best Practices Guide

> **Comprehensive guide to accessing services in Lokstra**

This guide explains the different ways to access services and when to use each pattern.

---

## 📊 Quick Comparison

| Pattern | Performance | Error Handling | Use Case |
|---------|-------------|----------------|----------|
| `GetService[T]()` | ⚠️ **Slow** (map lookup every call) | Returns `nil` (confusing errors) | Quick prototypes, dynamic service access |
| `MustGetService[T]()` | ⚠️ **Slow** (map lookup every call) | **Panics** with clear error | Fail-fast prototypes |
| **`LazyLoad[T]() + MustGet()`** | ✅ **Fast** (cached after first access) | ✅ **Panics** with clear error | **Production (recommended)** |
| `LazyLoad[T]() + Get()` | ✅ **Fast** (cached after first access) | Returns `nil` (confusing errors) | When nil handling needed |

**Recommended**: Always use `.MustGet()` for clearer error messages!

---

## Pattern 1: GetService (Direct Registry Access)

### Signature
```go
func GetService[T any](serviceName string) T
```

### Example
```go
func listUsersHandler(ctx *request.Context) error {
    // ⚠️ Lookup in registry EVERY request
    userService := lokstra_registry.GetService[*UserService]("users")
    
    // ⚠️ Check for nil!
    if userService == nil {
        return ctx.Api.Error(500, "SERVICE_UNAVAILABLE", "User service not found")
    }
    
    users, err := userService.GetAll()
    // ...
}
```

### When to Use
✅ **Good for**:
- Quick prototypes
- Dynamic service name (decided at runtime)
- Services that may or may not exist

❌ **Avoid for**:
- Production code with high traffic
- Handlers called frequently
- When service name is known at compile time

### Performance Impact
```
Per Request:
1. Map lookup in registry
2. Type assertion
3. Return value

Overhead: ~100-200ns per call
```

---

## Pattern 2: MustGetService (Fail-Fast Registry Access)

### Signature
```go
func MustGetService[T any](serviceName string) T
```

### Example
```go
func listUsersHandler(ctx *request.Context) error {
    // ⚠️ Lookup in registry EVERY request
    // ⚠️ Panics if service not found!
    userService := lokstra_registry.MustGetService[*UserService]("users")
    
    users, err := userService.GetAll()
    // ...
}
```

### When to Use
✅ **Good for**:
- Development/testing
- Fail-fast prototypes
- Critical services (app should crash if missing)

❌ **Avoid for**:
- Production APIs (don't let requests panic!)
- Optional services
- Graceful error handling scenarios

### Behavior
- **Service exists**: Returns service instance
- **Service missing**: **PANICS** immediately
- **Service nil**: **PANICS** immediately

---

## Pattern 3: service.LazyLoad (Cached Access) ⭐ RECOMMENDED

### Signature
```go
func LazyLoad[T any](serviceName string) *Cached[T]

// Methods:
func (c *Cached[T]) Get() T       // Returns nil if not found (confusing errors!)
func (c *Cached[T]) MustGet() T   // Panics if not found (clear errors!) ✅ RECOMMENDED
```

### Example (Recommended: MustGet)
```go
// Package-level: Loaded once, cached forever
var userService = service.LazyLoad[*UserService]("users")

func listUsersHandler(ctx *request.Context) error {
    // ✅ MustGet: Panics with clear error message if service not found
    users, err := userService.MustGet().GetAll()
    if err != nil {
        return ctx.Api.Error(500, "INTERNAL_ERROR", err.Error())
    }
    
    return ctx.Api.Ok(users)
}
```

### MustGet() vs Get()

**✅ Recommended: MustGet() - Clear Error Messages**
```go
users, err := userService.MustGet().GetAll()

// If service not found:
// Panic: "service 'users' not found or not initialized"
// ✅ Clear! You know exactly what's wrong
```

**⚠️ Not Recommended: Get() - Confusing Errors**
```go
users, err := userService.MustGet().GetAll()

// If service not found:
// Panic: "runtime error: invalid memory address or nil pointer dereference"
// ❌ Confusing! What caused the nil? DB? Service? Something else?
```

**When to use Get()**:
Only when you explicitly want to handle nil case:
```go
svc := userService.Get()
if svc == nil {
    // Custom handling for missing service
    log.Warn("User service not available, using fallback")
    return fallbackResponse
}
users, err := svc.GetAll()
```

### When to Use
✅ **Best for**:
- **Production code** (highly recommended!)
- High-traffic handlers
- Package-level service access
- Struct field dependencies

❌ **Don't use for**:
- Function-local variables (cache useless!)
- Dynamic service names

### Performance Impact
```
First Access:
1. Map lookup in registry
2. Type assertion
3. Cache in sync.Once
Overhead: ~100-200ns (one time)

Subsequent Accesses:
1. Return cached value
Overhead: ~1-5ns (atomic check)

Performance Gain: 20-100x faster than GetService
```

### Advanced: Fail-Fast with MustGet()
```go
var userService = service.LazyLoad[*UserService]("users")

func handler(ctx *request.Context) error {
    // Panics if service not found (fail-fast on startup)
    users, err := userService.MustGet().GetAll()
    // ...
}
```

---

## 🎯 Scope Guidelines

### ✅ Package-Level (Optimal)
```go
package handlers

import "github.com/primadi/lokstra/core/service"

// Loaded once per application lifecycle
var (
    userService  = service.LazyLoad[*UserService]("users")
    orderService = service.LazyLoad[*OrderService]("orders")
)

func listUsersHandler(ctx *request.Context) error {
    // ✅ Use MustGet() for clear errors
    users, err := userService.MustGet().GetAll()
    // ...
}
```

**Why optimal?**
- Cached across all requests
- Loaded on first handler call
- Zero lookup cost after initialization
- Clear error messages with MustGet()

### ✅ Struct Field (Good for Services)
```go
type OrderService struct {
    Users *service.Cached[*UserService] // Cross-service dependency
    DB    *service.Cached[*Database]
}

func (s *OrderService) GetOrderWithUser(orderID int) (*OrderWithUser, error) {
    order := s.DB.MustGet().GetOrder(orderID)
    user := s.Users.MustGet().GetByID(order.UserID) // Cached + clear errors!
    // ...
}
```

**Why good?**
- Service-to-service dependencies
- Cached per service instance
- Clear dependency graph
- Fail-fast with MustGet()

### ❌ Function-Local (DON'T DO THIS!)
```go
func handler(ctx *request.Context) error {
    // ❌ BAD! Cache created every request, then thrown away!
    userService := service.LazyLoad[*UserService]("users")
    users, err := userService.MustGet().GetAll()
    // ...
}
```

**Why bad?**
- New `Cached` instance per request
- Cache never reused
- Wastes memory
- **No performance benefit!**

**Use GetService instead:**
```go
func handler(ctx *request.Context) error {
    // ✅ Better if you need function-local access
    userService := lokstra_registry.GetService[*UserService]("users")
    users, err := userService.GetAll()
    // ...
}
```

---

## 🔄 Migration Guide

### From GetService to LazyLoad

**Before (Slow)**:
```go
func listUsersHandler(ctx *request.Context) error {
    userService := lokstra_registry.GetService[*UserService]("users")
    users, err := userService.GetAll()
    return ctx.Api.Ok(users)
}

func getUserHandler(ctx *request.Context) error {
    userService := lokstra_registry.GetService[*UserService]("users")
    user, err := userService.GetByID(id)
    return ctx.Api.Ok(user)
}
```

**After (Fast + Clear Errors)**:
```go
// Add package-level cached service
var userService = service.LazyLoad[*UserService]("users")

func listUsersHandler(ctx *request.Context) error {
    // ✅ MustGet() for clear error messages
    users, err := userService.MustGet().GetAll()
    return ctx.Api.Ok(users)
}

func getUserHandler(ctx *request.Context) error {
    user, err := userService.MustGet().GetByID(id)
    return ctx.Api.Ok(user)
}
```

**Performance Gain**: 20-100x faster + clear error messages!

---

## 📚 Additional Resources

- **Service Registration Guide**: [02-service](./02-service)
- **CRUD API Example**: [../00-introduction/examples/03-crud-api/](../00-introduction/examples/03-crud-api/)

---

## 💡 Key Takeaways

1. ✅ **Use `service.LazyLoad[T]()` with `.MustGet()`** for production code (package-level or struct fields)
2. ✅ **Always prefer `.MustGet()` over `.Get()`** for clear error messages
3. ⚠️ **Avoid `GetService[T]()`** in high-traffic handlers (slow + confusing errors)
4. ⚠️ **Never use `LazyLoad` in function-local variables** (cache wasted)
5. ✅ **Package-level vars** provide optimal caching

**Remember**: 
- The goal is to minimize registry lookups. Cache once, use everywhere!
- Use `.MustGet()` for fail-fast behavior with clear error messages
- Avoid nil pointer panics - let MustGet tell you exactly what's missing!
