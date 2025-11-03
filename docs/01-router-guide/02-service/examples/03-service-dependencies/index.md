# Service Dependencies - Lazy Service Registration

⏱️ **Estimated time**: 15 minutes

## 🎯 What You'll Learn

This example demonstrates **Lazy Service Registration** - a powerful pattern that eliminates dependency ordering problems. Services can be registered in any order, and dependencies are automatically resolved on first access!

**Key Concepts:**
- `RegisterLazyService` for on-demand instantiation
- Automatic dependency resolution
- No need to worry about service creation order
- Thread-safe singleton pattern

## 🌟 The Problem: Dependency Order

**Traditional Approach (Order Matters!):**
```go
// ❌ ERROR! userService depends on userRepo, must create repo first!
userService := service.NewUserService(userRepo)  // userRepo not defined yet!
lokstra_registry.RegisterService("user-service", userService)

userRepo := repository.NewUserRepository()
lokstra_registry.RegisterService("user-repo", userRepo)
```

**Fixed (But Tedious):**
```go
// ✅ Must create in specific order
userRepo := repository.NewUserRepository()
lokstra_registry.RegisterService("user-repo", userRepo)

userService := service.NewUserService(userRepo)  // OK now
lokstra_registry.RegisterService("user-service", userService)

orderService := service.NewOrderService(userService)  // Depends on userService
lokstra_registry.RegisterService("order-service", orderService)
```

**Problem:** Complex dependency graphs become hard to manage!

## 💡 The Solution: Lazy Service Registration

**With RegisterLazyService (Order Doesn't Matter!):**

```go
// Supports THREE factory signatures (auto-wrapped by framework):
// 1. func(deps, cfg map[string]any) any - full control with dependencies
// 2. func(cfg map[string]any) any       - only config (no dependencies)  
// 3. func() any                          - no params (simplest!)

// Multiple DB instances with different DSN (config only - mode 2)
lokstra_registry.RegisterLazyService("db-main", func(cfg map[string]any) any {
    return db.NewConnection(cfg["dsn"].(string))
}, map[string]any{
    "dsn": "postgresql://localhost/main",
})

lokstra_registry.RegisterLazyService("db-analytics", func(cfg map[string]any) any {
    return db.NewConnection(cfg["dsn"].(string))
}, map[string]any{
    "dsn": "postgresql://localhost/analytics",
})

// Services without params - simplest! (mode 3)
lokstra_registry.RegisterLazyService("user-repo", func() any {
    db := lokstra_registry.MustGetService[*DB]("db-main")
    return NewUserRepository(db)
}, nil)

lokstra_registry.RegisterLazyService("user-service", func() any {
    userRepo := lokstra_registry.MustGetService[*UserRepository]("user-repo")
    return NewUserService(userRepo)
}, nil)

lokstra_registry.RegisterLazyService("order-service", func() any {
    userSvc := lokstra_registry.MustGetService[*UserService]("user-service")
    return NewOrderService(userSvc)
}, nil)

// Advanced: Full signature with deps parameter (mode 1)
// Useful when you need to distinguish between service deps and config values
lokstra_registry.RegisterLazyService("advanced-service", func(deps, cfg map[string]any) any {
    // deps = service dependencies (if injected by framework)
    // cfg = configuration values
    timeout := cfg["timeout"].(int)
    userSvc := lokstra_registry.MustGetService[*UserService]("user-service")
    return NewAdvancedService(userSvc, timeout)
}, map[string]any{"timeout": 30})

// Services created ONLY when first accessed
// Dependencies resolved automatically
// Thread-safe singleton
```

**Benefits:**
- ✅ Register in any order
- ✅ Dependencies auto-resolved
- ✅ Lazy instantiation (performance!)
- ✅ Thread-safe
- ✅ Clean code
- ✅ **Config per instance** (e.g., multiple DB with different DSN)
- ✅ **Three factory modes** - use simplest one that fits your needs
- ✅ **Clear separation** - deps for services, cfg for config values

## 📋 Example Structure

```
03-service-dependencies/
├── main.go
├── repository/
│   ├── user_repository.go
│   └── order_repository.go
├── service/
│   ├── user_service.go
│   └── order_service.go
└── model/
    ├── user.go
    └── order.go
```

## 🔧 Code Walkthrough

### 1. Repository Layer (No Dependencies)

**repository/user_repository.go:**
```go
type UserRepository struct {
    users []model.User
}

func NewUserRepository() *UserRepository {
    return &UserRepository{
        users: []model.User{
            {ID: 1, Name: "Alice"},
            {ID: 2, Name: "Bob"},
        },
    }
}

func (r *UserRepository) FindByID(id int) (*model.User, error) {
    // ... implementation
}
```

### 2. Service Layer (Depends on Repository)

**service/user_service.go:**
```go
type UserService struct {
    repo *repository.UserRepository
}

// Constructor accepts dependency
func NewUserService(repo *repository.UserRepository) *UserService {
    return &UserService{repo: repo}
}

func (s *UserService) GetUser(id int) (*model.User, error) {
    return s.repo.FindByID(id)
}
```

### 3. Order Service (Depends on User Service)

**service/order_service.go:**
```go
type OrderService struct {
    userService *UserService
    orderRepo   *repository.OrderRepository
}

// Depends on BOTH UserService AND OrderRepository
func NewOrderService(
    userService *UserService,
    orderRepo *repository.OrderRepository,
) *OrderService {
    return &OrderService{
        userService: userService,
        orderRepo:   orderRepo,
    }
}

func (s *OrderService) CreateOrder(userID int, items []string) (*model.Order, error) {
    // Validate user exists
    user, err := s.userService.GetUser(userID)
    if err != nil {
        return nil, err
    }
    
    // Create order
    return s.orderRepo.Create(user.ID, items)
}
```

### 4. Main - Register with LazyService

**main.go:**
```go
func main() {
    // ============================================
    // Register ALL services - ORDER DOESN'T MATTER!
    // ============================================
    
    // Can register in any order!
    
    // Multiple DB instances with different config
    lokstra_registry.RegisterLazyService("db-main", func(cfg map[string]any) any {
        return db.NewConnection(cfg["dsn"].(string))
    }, map[string]any{
        "dsn": "postgresql://localhost/main",
    })
    
    lokstra_registry.RegisterLazyService("db-analytics", func(cfg map[string]any) any {
        return db.NewConnection(cfg["dsn"].(string))
    }, map[string]any{
        "dsn": "postgresql://localhost/analytics",
    })
    
    // Services with dependencies - no config
    lokstra_registry.RegisterLazyService("order-service", func(cfg map[string]any) any {
        // Dependencies resolved automatically
        userSvc := lokstra_registry.MustGetService[*service.UserService]("user-service")
        orderRepo := lokstra_registry.MustGetService[*repository.OrderRepository]("order-repo")
        return service.NewOrderService(userSvc, orderRepo)
    }, nil)
    
    lokstra_registry.RegisterLazyService("user-service", func(cfg map[string]any) any {
        repo := lokstra_registry.MustGetService[*repository.UserRepository]("user-repo")
        return service.NewUserService(repo)
    }, nil)
    
    lokstra_registry.RegisterLazyService("order-repo", func(cfg map[string]any) any {
        db := lokstra_registry.MustGetService[*DB]("db-main")
        return repository.NewOrderRepository(db)
    }, nil)
    
    lokstra_registry.RegisterLazyService("user-repo", func(cfg map[string]any) any {
        db := lokstra_registry.MustGetService[*DB]("db-main")
        return repository.NewUserRepository(db)
    }, nil)
    
    // ============================================
    // Access services - created on demand
    // ============================================
    
    // First access triggers creation chain:
    // 1. order-service factory called
    // 2. Needs user-service -> user-service factory called
    // 3. Needs user-repo -> user-repo factory called
    // 4. Needs db-main -> db-main factory called
    // 5. All cached for future use
    orderSvc := lokstra_registry.MustGetService[*service.OrderService]("order-service")
    
    order, err := orderSvc.CreateOrder(1, []string{"item1", "item2"})
    // ...
}
```

## 🎨 Pattern Comparison

### Pattern 1: Eager Registration (Original)
```go
// Must create in order
userRepo := repository.NewUserRepository()
lokstra_registry.RegisterService("user-repo", userRepo)

userSvc := service.NewUserService(userRepo)
lokstra_registry.RegisterService("user-service", userSvc)
```
- ❌ Order matters
- ❌ All created at startup
- ✅ Simple for small apps

### Pattern 2: Lazy Registration (New!)
```go
// Register in any order
lokstra_registry.RegisterLazyService("user-service", func() any {
    repo := lokstra_registry.MustGetService[*repository.UserRepository]("user-repo")
    return service.NewUserService(repo)
})

lokstra_registry.RegisterLazyService("user-repo", func() any {
    return repository.NewUserRepository()
})
```
- ✅ Order doesn't matter!
- ✅ Created on demand
- ✅ Auto dependency resolution
- ✅ Scales to complex apps

### Pattern 3: Hybrid (Best of Both)
```go
// Eager for simple services
lokstra_registry.RegisterService("config", config)

// Lazy for services with dependencies
lokstra_registry.RegisterLazyService("user-service", func() any {
    cfg := lokstra_registry.MustGetService[*Config]("config")
    return service.NewUserService(cfg)
})
```

## 💡 Key Takeaways

1. **Lazy Registration Solves Dependency Order Problems**
   - Register in any order
   - Dependencies resolved automatically
   - No more "undefined variable" errors

2. **Performance Benefits**
   - Services only created when needed
   - Singleton pattern (created once)
   - Thread-safe

3. **Clean Dependency Injection**
   - Constructor injection
   - Explicit dependencies
   - Easy to test

4. **When to Use Each Pattern:**
   - **Eager:** Simple services, no dependencies
   - **Lazy:** Complex dependencies, conditional usage
   - **Hybrid:** Mix both for optimal balance

## 🚀 Running the Example

```bash
cd docs/01-essentials/02-service/examples/03-service-dependencies
go run .
```

Test endpoints:
```http
POST /orders
{
    "user_id": 1,
    "items": ["Laptop", "Mouse"]
}

GET /orders/1
GET /users/1
```

## 🔗 Related

- Example 01: Simple Service (eager registration)
- Example 02: LazyLoad vs GetService (access patterns)
- Example 04: Service as Router (auto-generated endpoints)

---

**Remember:** `RegisterLazyService` is your friend for complex dependency graphs! 🎯
