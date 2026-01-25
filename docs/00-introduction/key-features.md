# Key Features

> **What makes Lokstra special - the features that set it apart**

---

## 🎯 Overview

Lokstra has several **killer features** that make building REST APIs faster, cleaner, and more flexible:

1. **29 Handler Forms** - Write handlers your way
2. **Service as Router** - Auto HTTP endpoints from services
3. **One Binary, Multiple Deployments** - Monolith ↔ Microservices
4. **Built-in Lazy DI** - No external framework needed
5. **Flexible Configuration** - Code + YAML patterns
6. **Annotation-Driven Development** - 83% less boilerplate

Let's dive into each feature:

---

## 🎨 Feature 1: 29 Handler Forms

### The Problem
Most frameworks lock you into one handler pattern:

```go
// Gin - must use this signature
func Handler(c *gin.Context) {
    // forced pattern
}

// Echo - must use this
func Handler(c echo.Context) error {
    // forced pattern
}
```

### The Lokstra Solution
**29 different handler forms** - use what makes sense:

#### Form 1: Simplest - No Params, Return Value
```go
r.GET("/ping", func() string {
    return "pong"
})

r.GET("/config", func() map[string]any {
    return map[string]any{
        "version": "1.0",
        "env": "production",
    }
})
```

**When**: Simple endpoints, no errors possible

---

#### Form 2: With Error Handling (Most Common!)
```go
r.GET("/users", func() ([]User, error) {
    users, err := db.GetAllUsers()
    if err != nil {
        return nil, err  // Auto 500 error
    }
    return users, nil  // Auto 200 OK
})
```

**When**: Operations that can fail (90% of cases)

---

#### Form 3: With Request Binding
```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

r.POST("/users", func(req *CreateUserRequest) (*User, error) {
    // req auto-bound from JSON body
    // auto-validated
    return db.CreateUser(req.Name, req.Email)
})
```

**When**: Need request data (POST/PUT)

---

#### Form 4: Full Control
```go
r.GET("/complex", func(ctx *request.Context) (*response.Response, error) {
    // Access everything
    token := ctx.R.Header.Get("Authorization")
    userAgent := ctx.R.Header.Get("User-Agent")
    
    // Custom response
    return response.Success(data).
        WithHeader("X-Custom", "value").
        WithStatus(201), nil
})
```

**When**: Need headers, cookies, custom status codes

---

### Why This Matters

**Developer Experience**:
```go
// Simple case? Simple code!
r.GET("/ping", func() string { return "pong" })

// Complex case? Full power!
r.GET("/api", func(ctx *request.Context, req *ComplexRequest) (*response.Response, error) {
    // Do complex stuff
})
```

**One size doesn't fit all** - Lokstra adapts to your needs.

📖 **See all 29 forms**: [Deep Dive: Handler Forms](../02-deep-dive/router/handler-forms)

---

## ⚡ Feature 2: Service as Router

### The Problem
Repetitive routing code for CRUD operations:

```go
// Traditional approach - lots of boilerplate
r.GET("/users", listUsers)
r.GET("/users/{id}", getUser)
r.POST("/users", createUser)
r.PUT("/users/{id}", updateUser)
r.DELETE("/users/{id}", deleteUser)

// And handlers calling services...
func listUsers() ([]User, error) {
    return userService.GetAll()
}
func getUser(req *GetUserReq) (*User, error) {
    return userService.GetByID(req.ID)
}
// ... more boilerplate
```

### The Lokstra Solution
**Service methods automatically become HTTP endpoints**:

```go
// 1. Define service with methods
type UserService struct {
    DB *Database
}

type GetAllParams struct {}
type GetByIDParams struct {
    ID int `path:"id"`
}
type CreateParams struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

func (s *UserService) GetAll(p *GetAllParams) ([]User, error) {
    return s.DB.Query("SELECT * FROM users")
}

func (s *UserService) GetByID(p *GetByIDParams) (*User, error) {
    return s.DB.QueryOne("SELECT * FROM users WHERE id = ?", p.ID)
}

func (s *UserService) Create(p *CreateParams) (*User, error) {
    return s.DB.Insert("INSERT INTO users ...", p.Name, p.Email)
}

// 2. Register service factory
func UserServiceFactory(deps map[string]any, config map[string]any) any {
    return &UserService{
        DB: deps["db"].(*Database),
    }
}

lokstra_registry.RegisterServiceType("users-factory", UserServiceFactory, nil)

// 3. Auto-generate router from service
userRouter := router.NewFromService(
    lokstra_registry.GetService[*UserService]("users"),
    "/users",  // Base path
)
```

**Result - Automatic routes**:
```
GET    /users       → UserService.GetAll()
GET    /users/{id}  → UserService.GetByID()
POST   /users       → UserService.Create()
PUT    /users/{id}  → UserService.Update()
DELETE /users/{id}  → UserService.Delete()
```

### Convention-Based Routing

Lokstra understands REST conventions:

| Method Name | HTTP Method | Path | Parameters |
|-------------|-------------|------|------------|
| `GetAll()` | GET | `/` | None |
| `GetByID()` | GET | `/{id}` | ID from path |
| `Create()` | POST | `/` | Body |
| `Update()` | PUT | `/{id}` | ID + Body |
| `Delete()` | DELETE | `/{id}` | ID from path |
| `List()` | GET | `/` | Query params |
| `Search()` | GET | `/search` | Query params |

### Why This Matters

**Zero Boilerplate**:
- ✅ No handler functions needed
- ✅ No routing registration
- ✅ No parameter extraction
- ✅ Focus on business logic only

**Business Logic in Services**:
```go
// All your logic in one place
type UserService struct {
    DB      *Database
    Email   *EmailService
    Cache   *CacheService
}

// Pure business logic
func (s *UserService) Create(p *CreateParams) (*User, error) {
    // Validate
    if err := validateUserData(p); err != nil {
        return nil, err
    }
    
    // Create - direct access, dependencies already injected
    user, err := s.DB.Insert(...)
    if err != nil {
        return nil, err
    }
    
    // Cache user data
    s.Cache.Set(user.ID, user)
    
    // Notify
    s.Email.SendWelcome(user.Email)
    
    return user, nil
}
```

📖 **Learn more**: [Framework Guide](../02-framework-guide/)

---

## 🏗️ Feature 3: One Binary, Multiple Deployments

### The Problem
Migrating from monolith to microservices requires:
- Code refactoring
- New repos
- Different builds
- Deployment complexity

### The Lokstra Solution
**Same code, different deployment configurations**:

```yaml
# config.yaml - One config file, multiple deployments

servers:
  # Deployment 1: Monolith (all services in one)
  - name: monolith
    deployment-id: monolith
    apps:
      - addr: ":8080"
        services: [users, orders, payments, products]
  
  # Deployment 2: Microservices (split services)
  - name: user-service
    deployment-id: microservices
    base-url: http://user-service
    apps:
      - addr: ":8001"
        services: [users]
  
  - name: order-service
    deployment-id: microservices
    base-url: http://order-service
    apps:
      - addr: ":8002"
        services: [orders, payments]
  
  - name: product-service
    deployment-id: microservices
    base-url: http://product-service
    apps:
      - addr: ":8003"
        services: [products]
```

**One binary, multiple modes**:
```bash
# Build once
go build -o myapp

# Deploy as monolith
./myapp --server=monolith

# Or deploy as microservices
./myapp --server=user-service    # Instance 1
./myapp --server=order-service   # Instance 2
./myapp --server=product-service # Instance 3
```

### How It Works

**Deployment Isolation**:
- Services in same `deployment-id` can communicate
- Services in different `deployment-id` are isolated
- Lokstra handles routing automatically

**Example**:
```go
// In OrderService
type OrderService struct {
    Users *UserService  // May be local or remote!
}

func (s *OrderService) CreateOrder(p *CreateOrderParams) (*Order, error) {
    // This works in BOTH deployments!
    user, err := s.Users.GetByID(p.UserID)
    // Monolith: Direct call
    // Microservices: HTTP call to user-service
    
    // ... create order
}
```

### Why This Matters

**Flexibility**:
- Start as monolith (simple deployment)
- Split to microservices when needed (no code change)
- Test locally as monolith
- Deploy as microservices

**Cost Efficiency**:
- Small projects: One server
- Growing: Split services gradually
- Enterprise: Full microservices

**Development Speed**:
- Local dev: Monolith (fast iteration)
- Staging: Multi-port (test separation)
- Production: Microservices (scale independently)

📖 **See it in action**: [Multi-Deployment Example](../05-examples/single-binary-deployment/)

---

## 🔧 Feature 4: Built-in Lazy Dependency Injection

### The Problem
Managing dependencies manually or using heavy DI frameworks:

```go
// Manual DI - order matters, error-prone
db := createDB()
cache := createCache()
userRepo := NewUserRepo(db, cache)  // Must create in order!
orderRepo := NewOrderRepo(db, cache)
userService := NewUserService(userRepo)
orderService := NewOrderService(orderRepo, userService)
```

### The Lokstra Solution
**Built-in lazy DI with type safety**:

```go
import "github.com/primadi/lokstra/core/service"

// 1. Define services with dependencies
type OrderService struct {
    DB    *service.Cached[*Database]
    Users *service.Cached[*UserService]
    Cache *service.Cached[*CacheService]
}

// 2. Register factories (order doesn't matter with depends-on!)
lokstra_registry.RegisterServiceType("db", createDatabase)
lokstra_registry.RegisterServiceType("cache", createCache)

lokstra_registry.RegisterServiceFactory("users-factory", 
    func(deps map[string]any, config map[string]any) any {
        return &UserService{
            DB: service.Cast[*Database](deps["db"]),
        }
    })
lokstra_registry.RegisterLazyService("users", "users-factory", 
    map[string]any{"depends-on": []string{"db"}})

lokstra_registry.RegisterServiceFactory("orders-factory", 
    func(deps map[string]any, config map[string]any) any {
        return &OrderService{
            DB:    service.Cast[*Database](deps["db"]),
            Users: service.Cast[*UserService](deps["users"]),
            Cache: service.Cast[*CacheService](deps["cache"]),
        }
    })
lokstra_registry.RegisterLazyService("orders", "orders-factory", 
    map[string]any{"depends-on": []string{"db", "users", "cache"}})

// 3. Use anywhere - dependencies auto-injected
orders := lokstra_registry.GetService[*OrderService]("orders")
```

### Lazy Loading

Services created only when first accessed:

```go
type OrderService struct {
    DB *service.Cached[*Database]
}

func (s *OrderService) CreateOrder(p *CreateOrderParams) (*Order, error) {
    // DB created on first .Get() call
    db := s.DB.Get()  // Lazy load here!
    
    return db.Insert("INSERT INTO orders ...")
}
```

**Benefits**:
- ✅ Services created only if used
- ✅ Circular dependencies OK (lazy resolution)
- ✅ Faster startup (no upfront init)
- ✅ Memory efficient

### Type-Safe with Generics

```go
// Type-safe - compile-time checked
users := lokstra_registry.GetService[*UserService]("users")
// ✅ Type: *UserService

// Wrong type? Compile error!
users := lokstra_registry.GetService[*WrongType]("users")
// ❌ Compile error
```

### Why This Matters

**No External Framework**:
- No uber/dig
- No google/wire
- Built into Lokstra

**Simple API**:
```go
// Register
RegisterServiceType(name, factory)

// Use
GetService[T](name)

// That's it!
```

📖 **Learn more**: [Framework Guide](../02-framework-guide/)

---

## ⚙️ Feature 5: Flexible Configuration

### The Problem
Hardcoded configuration or limited config options:

```go
// Hardcoded - must recompile to change
const port = ":8080"
const dbHost = "localhost"
```

### The Lokstra Solution
**Code + YAML pattern** - best of both worlds:

#### Approach 1: Pure Code (Simple Apps)
```go
r := lokstra.NewRouter("api")
r.Use(logging.Middleware(), cors.Middleware())
r.GET("/users", getUsers)

app := lokstra.NewApp("demo", ":8080", r)
app.Run(30 * time.Second)
```

#### Approach 2: Code + Config (Recommended)
```go
// code: setup.go
func setupServices() {
    lokstra_registry.RegisterServiceType("db", createDB)
    lokstra_registry.RegisterServiceType("users", createUserService)
}

func setupRouters() {
    lokstra_registry.RegisterRouter("api", createAPIRouter())
}

// config: app.yaml
servers:
  - name: dev-server
    deployment-id: dev
    apps:
      - addr: ":3000"
        routers: [api]
        services: [db, users]
        
# main.go
func main() {
    setupServices()
    setupRouters()
    
    var cfg config.Config
    config.LoadConfigFile("app.yaml", &cfg)
    lokstra_registry.RegisterConfig(&cfg, "dev-server")
    
    lokstra_registry.RunServer(30 * time.Second)
}
```

### Environment Variables

```yaml
# config.yaml
services:
  - name: database
    type: postgres
    config:
      host: ${DB_HOST:localhost}           # env var with default
      port: ${DB_PORT:5432}
      password: ${DB_PASSWORD}             # required env var
      ssl_mode: ${DB_SSL_MODE:disable}
```

### Multi-File Configuration

```yaml
# base.yaml - shared config
services:
  - name: database
    type: postgres
    
# dev.yaml - dev overrides
services:
  - name: database
    config:
      host: localhost
      
# prod.yaml - production overrides
services:
  - name: database
    config:
      host: ${DB_HOST}
      ssl_mode: require
```

Load and merge:
```go
var cfg config.Config
config.LoadConfigFile("base.yaml", &cfg)
config.LoadConfigFile("dev.yaml", &cfg)  // Merges with base
```

### Why This Matters

**Flexibility**:
- Simple apps: Pure code
- Complex apps: Code + config
- Enterprise: Full config-driven

**Environment Management**:
- One codebase
- Multiple environments
- Easy deployment

📖 **Learn more**: [Framework Guide](../02-framework-guide/)

---

## � Feature 6: Annotation-Driven Development

### The Problem
Setting up services with DI, routing, and wiring requires lots of boilerplate:

```go
// 70+ lines of manual setup for one service!

// 1. Define service
type UserService struct {
    DB *service.Cached[*Database]
}

// 2. Create factory
func createUserServiceFactory() any {
    return func(deps map[string]any, config map[string]any) any {
        return &UserService{
            DB: service.Cast[*Database](deps["db"]),
        }
    }
}

// 3. Register factory
lokstra_registry.RegisterServiceFactory("user-service-factory", 
    createUserServiceFactory())

// 4. Register lazy service
lokstra_registry.RegisterLazyService("user-service", 
    "user-service-factory",
    map[string]any{"depends-on": []string{"db"}})

// 5. Create router from service
func setupUserRouter() *lokstra.Router {
    userService := lokstra_registry.GetService[*UserService]("user-service")
    return router.NewFromService(userService, "/api")
}

// 6. Register router
lokstra_registry.RegisterRouter("user-router", setupUserRouter())

// 7. Mount in server config
// ... more YAML config
```

### The Lokstra Solution
**Annotations replace 70+ lines with 12 lines** - like NestJS decorators:

```go
// @Handler name="user-service", prefix="/api"
type UserServiceImpl struct {
    // @Inject "database"
    DB *service.Cached[*Database]
}

// @Route "GET /users"
func (s *UserServiceImpl) GetAll(p *GetAllRequest) ([]User, error) {
    return s.DB.MustGet().GetAllUsers()
}

// @Route "GET /users/{id}"
func (s *UserServiceImpl) GetByID(p *GetByIDRequest) (*User, error) {
    return s.DB.MustGet().GetUserByID(p.ID)
}

// Auto-generates: factory, DI wiring, routes, remote proxy!
```

### How It Works

**Step 1: Add Annotations**
```go
// @Handler name="user-service", prefix="/api", mount="/api"
type UserServiceImpl struct {
    // @Inject "database"
    DB *service.Cached[*Database]
}

// @Route "POST /users"
func (s *UserServiceImpl) Create(p *CreateUserRequest) (*User, error) {
    return s.DB.MustGet().CreateUser(p)
}
```

**Step 2: Run to Auto-Generate**
```bash
# Code generation happens automatically
go run .
```

**Step 3: Lokstra Generates Everything**

Generates `zz_generated.lokstra.go`:
```go
// ✅ Service factory
func init() {
    lokstra_registry.RegisterServiceFactory("user-service-factory", ...)
    lokstra_registry.RegisterLazyService("user-service", ...)
}

// ✅ Router with routes
func init() {
    r := lokstra.NewRouter("user-service")
    r.POST("/users", ...) // Auto-wired to Create()
    lokstra_registry.RegisterRouter("user-service", r)
}

// ✅ Remote service proxy (for microservices)
type UserServiceRemote struct { ... }
```

### Three Powerful Annotations

#### 1. @Handler - Define Service
```go
// @Handler name="user-service", prefix="/api", mount="/api"
type UserServiceImpl struct {}
```

**Generates**:
- Service factory
- DI registration
- Router with prefix
- Remote proxy for microservices

---

#### 2. @Inject - Dependency Injection
```go
type UserServiceImpl struct {
    // @Inject "database"
    DB *service.Cached[*Database]
    
    // @Inject "email-service"
    Email *service.Cached[*EmailService]
}
```

**Generates**:
- Dependency wiring in factory
- `depends-on` configuration
- Type-safe lazy loading

---

#### 3. @Route - HTTP Endpoints
```go
// @Route "GET /users"
func (s *UserServiceImpl) GetAll(p *GetAllRequest) ([]User, error) {}

// @Route "POST /users"
func (s *UserServiceImpl) Create(p *CreateUserRequest) (*User, error) {}

// @Route "PUT /users/{id}"
func (s *UserServiceImpl) Update(p *UpdateUserRequest) (*User, error) {}
```

**Generates**:
- Route registration
- Parameter binding (path, query, body)
- HTTP method mapping
- Response serialization

### Why This Matters

**Massive Code Reduction**:
```
Traditional Approach: 70+ lines of boilerplate
With Annotations:    12 lines of business logic
Reduction:           83% less code!
```

**Developer Experience**:
- ✅ Like NestJS decorators (familiar pattern)
- ✅ Like Spring annotations (proven approach)
- ✅ But in Go (compile-time safe)
- ✅ No reflection at runtime (fast!)

**Productivity Boost**:
```go
// Before: Write 6 separate files
service.go          // Service implementation
factory.go          // Factory function
register.go         // Registration code
router.go           // Router setup
routes.go           // Route definitions
config.yaml         // YAML configuration

// After: Write 1 annotated file
user_service.go     // Everything in one place!
```

**Type Safety**:
- Annotations validated at build time
- Generated code is type-safe
- No runtime reflection overhead
- Full IDE support

### Comparison with Other Frameworks

| Framework | Pattern | Runtime Cost |
|-----------|---------|--------------|
| **NestJS** | Decorators | High (reflection) |
| **Spring** | Annotations | High (reflection) |
| **Lokstra** | Annotations | **Zero** (code generation) |

**Lokstra advantage**: All the DX benefits of annotations, none of the runtime cost!

### Build Workflow Integration

**Problem**: Code generation only happens during `go run`, not `go build`

**Solution**: Use `--generate-only` flag:

```bash
# Force code generation without running
go run . --generate-only

# Then build normally
go build -o myapp
```

**Or use provided build scripts**:
```bash
# Linux/Mac
./build.sh

# Windows (PowerShell)
.\build.ps1

# Windows (CMD)
.\build.bat
```

Scripts automatically:
1. Run `go run . --generate-only`
2. Run `go mod tidy`
3. Build for multiple platforms

### Real-World Example

See **Example 07: Enterprise Router Service**:
- Complete working example
- Full documentation
- Before/after comparison
- Build scripts included

📖 **Full guide**: [Example 07](../01-router-guide/07_enterprise_router_service/)

---

##  Feature Comparison

| Feature | Standard Libs | Gin/Echo | Lokstra |
|---------|--------------|----------|---------|
| **Handler Forms** | 1 | 1 | 29 |
| **Service Layer** | Manual | Manual | Built-in |
| **DI System** | None | None | Built-in (Lazy) |
| **Service→Router** | No | No | Yes (Auto) |
| **Multi-Deploy** | No | No | Yes (1 binary) |
| **Config-Driven** | Limited | Limited | Full (YAML) |
| **Lazy Loading** | Manual | Manual | Built-in |
| **Convention/Config** | Config | Config | Convention + Config |
| **Annotations** | No | No | Yes (83% less code) |

---

## 🚀 See Features in Action

### Quick Demos:

**Handler Forms**:
👉 [Handler Forms Example](../01-router-guide/01-router/examples/04-handler-forms/)

**Service as Router**:
👉 [Service Router Example](../01-router-guide/02-service/examples/)

**Multi-Deployment**:
👉 [Single Binary Example](../05-examples/single-binary-deployment/)

**Full Stack**:
👉 [Complete Todo API](../01-router-guide/06-putting-it-together/)

---

## 💡 What Makes Lokstra Unique?

Most frameworks focus on **one thing**:
- Fast routing (Chi)
- Simple API (Gin)
- Full features (Beego)

**Lokstra focuses on developer experience**:
- ✅ Flexible (29 handler forms)
- ✅ Productive (service as router)
- ✅ Scalable (multi-deployment)
- ✅ Clean (built-in DI)
- ✅ Simple (convention over config)
- ✅ Modern (annotation-driven like NestJS)

**Result**: Build faster, scale easier, maintain better.

---

## 📚 Learn More

**Deep Dives**:
- [Architecture](architecture) - How it all works together
- [Router Guide](../01-router-guide/) - Hands-on tutorials
- [Deep Dive](../02-deep-dive) - Advanced patterns

**Try It**:
- [Quick Start](quick-start) - Build your first API in 5 minutes
- [Examples](../05-examples) - Real applications

---

**Ready to experience these features?** 👉 [Get Started](quick-start)
