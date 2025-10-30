# Lokstra Framework Cheatsheet

Quick reference untuk programmer baru. Print atau bagikan sebagai handout!

---

## 🚀 Quick Start

```bash
# Install
go get github.com/primadi/lokstra

# Hello World
package main

import (
    "time"
    "github.com/primadi/lokstra"
)

func main() {
    r := lokstra.NewRouter("api")
    r.GET("/", func() string { return "Hello!" })
    
    app := lokstra.NewApp("api", ":3000", r)
    server := lokstra.NewServer("my-server", app)
    if err := server.Run(30 * time.Second); err != nil {
        log.Fatal("❌ Failed to start server:", err)
    }
}
```

---

## 🎨 Handler Forms (Most Common)

```go
// 1. Simple return
r.GET("/ping", func() string {
    return "pong"
})

// 2. With error (MOST COMMON!)
r.GET("/users", func() ([]User, error) {
    return db.GetAllUsers()
})

// 3. With request binding
type CreateReq struct {
    Name string `json:"name" validate:"required"`
}

r.POST("/users", func(req *CreateReq) (*User, error) {
    return db.CreateUser(req)
})

// 4. Full control
r.GET("/api", func(ctx *request.Context) error {
    data := "Hello Lokstra"
    return ctx.Resp.Json(data)
})

// 5. HTTP compatible
r.GET("/std", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("ok"))
}))
```

**Total: 29+ forms available!**

---

## 🔧 Service Pattern

```go
// Define service
type UserService struct {
    DB *service.Cached[*Database]
}

func (s *UserService) GetAll(p *GetAllParams) ([]User, error) {
    return s.DB.MustGet().Query("SELECT * FROM users")
}

// Register service
lokstra_registry.RegisterServiceFactory("users-factory",
    func(deps map[string]any, config map[string]any) any {
        return &UserService{
            DB: service.Cast[*Database](deps["db"]),
        }
    })

lokstra_registry.RegisterLazyService("users", "users-factory",
    map[string]any{"depends-on": []string{"db"}})

// Use with lazy loading (OPTIMAL!)
var userService = service.LazyLoad[*UserService]("users")

r.GET("/users", func() ([]User, error) {
    return userService.MustGet().GetAll(&GetAllParams{})
})
```

---

## ⚡ Service as Router

```go
// Service methods
func (s *UserService) GetAll(p *GetAllParams) ([]User, error) { ... }
func (s *UserService) GetByID(p *GetByIDParams) (*User, error) { ... }
func (s *UserService) Create(p *CreateParams) (*User, error) { ... }
func (s *UserService) Update(p *UpdateParams) (*User, error) { ... }
func (s *UserService) Delete(p *DeleteParams) error { ... }

// Register with metadata
lokstra_registry.RegisterServiceType(
    "user-service-factory",
    UserServiceFactory, UserServiceRemoteFactory,
    deploy.WithResource("user", "users"),
    deploy.WithConvention("rest"),
)

// Auto-generate router
userRouter := lokstra_registry.NewRouterFromServiceType("user-service")

// Routes created automatically:
// GET    /users       → GetAll()
// GET    /users/{id}  → GetByID()
// POST   /users       → Create()
// PUT    /users/{id}  → Update()
// DELETE /users/{id}  → Delete()
```

---

## 🏗️ Multi-Deployment

```yaml
# config.yaml
service-definitions:
  user-service:
    type: user-service-factory
  order-service:
    type: order-service-factory
    depends-on: [user-service]

deployments:
  # Monolith
  dev:
    servers:
      api-server:
        addr: ":3000"
        published-services: [user-service, order-service]
  
  # Microservices
  prod:
    servers:
      user-server:
        addr: ":8001"
        published-services: [user-service]
      
      order-server:
        addr: ":8002"
        published-services: [order-service]
```

```bash
# Run
./app -server "dev.api-server"       # Monolith
./app -server "prod.user-server"     # Microservice 1
./app -server "prod.order-server"    # Microservice 2
```

---

## 🔗 Middleware

```go
// Global middleware
r.Use(loggingMiddleware, corsMiddleware)

// Group middleware
admin := r.Group("/admin")
admin.Use(authMiddleware)
admin.GET("/users", getUsers)

// Route-specific middleware
r.GET("/special", specialHandler, rateLimitMiddleware)

// By-name (from registry)
lokstra_registry.RegisterMiddleware("auth", authMiddleware)
r.Use("auth", "logging")

// Middleware function signature
func MyMiddleware() request.MiddlewareFunc {
    return func(ctx *request.Context, next func() error) error {
        // Before handler
        log.Println("Before")
        
        // Execute handler
        err := next()
        
        // After handler
        log.Println("After")
        
        return err
    }
}
```

---

## 📦 Request/Response Helpers

```go
// Path parameters
id := ctx.PathParam("id")
idInt := ctx.PathParamInt("id")

// Query parameters
name := ctx.QueryParam("name")
page := ctx.QueryParamInt("page", 1) // with default

// Headers
auth := ctx.R.Header.Get("Authorization")

// JSON body (auto-parsed with struct param)
type CreateReq struct {
    Name string `json:"name"`
}

r.POST("/users", func(req *CreateReq) (*User, error) {
    // req already parsed from JSON body
})

// Response
return response.Success(data)
return response.SuccessWithMessage(data, "Created successfully")
return response.Error(err)
return response.ErrorWithMessage(err, "Failed to create")

// Custom status
return response.Success(data).WithStatus(201), nil
```

---

## 🎯 Common Patterns

### Pattern 1: CRUD Handler
```go
var userService = service.LazyLoad[*UserService]("users")

r.GET("/users", func() ([]*User, error) {
    return userService.MustGet().GetAll(&GetAllParams{})
})

r.GET("/users/{id}", func(ctx *request.Context) (*User, error) {
    id := ctx.PathParamInt("id")
    return userService.MustGet().GetByID(&GetByIDParams{ID: id})
})

r.POST("/users", func(req *CreateUserReq) (*User, error) {
    return userService.MustGet().Create(&CreateParams{
        Name:  req.Name,
        Email: req.Email,
    })
})
```

### Pattern 2: Service with Dependencies
```go
type OrderService struct {
    DB    *service.Cached[*Database]
    Users *service.Cached[*UserService]
    Email *service.Cached[*EmailService]
}

func (s *OrderService) Create(p *CreateParams) (*Order, error) {
    // Use dependencies
    user, _ := s.Users.MustGet().GetByID(&GetByIDParams{ID: p.UserID})
    order, _ := s.DB.MustGet().Insert(p)
    s.Email.MustGet().SendConfirmation(user.Email, order)
    return order, nil
}
```

### Pattern 3: External Service
```yaml
# config.yaml
external-service-definitions:
  payment-gateway:
    url: "https://payment-api.example.com"
    type: payment-service-remote-factory
```

```go
// Wrapper service
type PaymentServiceRemote struct {
    proxyService *proxy.Service
}

func (s *PaymentServiceRemote) CreatePayment(p *CreatePaymentParams) (*Payment, error) {
    return proxy.CallWithData[*Payment](s.proxyService, "CreatePayment", p)
}
```

---

## 🎨 Router Patterns

```go
// Static route
r.GET("/ping", handler)

// Path parameter
r.GET("/users/{id}", handler)
r.GET("/posts/{id}/comments/{commentId}", handler)

// Route group
api := r.Group("/api/v1")
api.GET("/users", getUsers)     // /api/v1/users
api.POST("/users", createUser)  // /api/v1/users

// Method chaining
r.GET("/users", getUsers).
  POST("/users", createUser).
  PUT("/users/{id}", updateUser)

// Print routes (debugging)
r.Build()
r.PrintRoutes()
```

---

## ⚡ Performance Tips

### ✅ DO: Use LazyLoad
```go
// Optimal - cached resolution
var userService = service.LazyLoad[*UserService]("users")

r.GET("/users", func() ([]User, error) {
    return userService.MustGet().GetAll(&GetAllParams{})
})
```

### ❌ DON'T: Registry lookup per request
```go
// Not optimal - map lookup overhead
r.GET("/users", func() ([]User, error) {
    users := lokstra_registry.GetService[*UserService]("users")
    return users.GetAll(&GetAllParams{})
})
```

### ✅ DO: Use fast path handlers
```go
// Fast path (< 2μs)
r.GET("/users", func() ([]User, error) { ... })
r.GET("/user", func(ctx *Context) (*Response, error) { ... })
```

### ❌ AVOID: Reflection when possible
```go
// Reflection (slower, but still fast ~2.6μs)
type GetUsersReq struct { Page int }
r.GET("/users", func(req *GetUsersReq) ([]User, error) { ... })

// Consider: Use query param instead if simple
r.GET("/users", func(ctx *Context) ([]User, error) {
    page := ctx.QueryParamInt("page", 1)
    ...
})
```

---

## 🐛 Debugging

```go
// Print routes
r.Build()
r.PrintRoutes()

// Walk routes
r.Walk(func(method, path string, handler any) {
    fmt.Printf("%s %s\n", method, path)
})

// Check errors
if err := r.Build(); err != nil {
    log.Fatal(err)
}

// Middleware logging
func LoggingMiddleware() request.MiddlewareFunc {
    return func(ctx *request.Context, next func() error) error {
        log.Printf("→ %s %s", ctx.R.Method, ctx.R.URL.Path)
        err := next()
        log.Printf("← %s %s", ctx.R.Method, ctx.R.URL.Path)
        return err
    }
}
```

---

## 📊 Common Struct Tags

```go
// Request binding
type CreateUserReq struct {
    // JSON body
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"min=18,max=100"`
    
    // Path parameter
    ID int `path:"id"`
    
    // Query parameter
    Page int `query:"page"`
    
    // Header
    Auth string `header:"Authorization"`
}

// Response
type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}
```

---

## 🔑 Key Concepts

### 1. Lazy Loading
- Services created on first access
- Cached for subsequent calls
- Zero overhead after first load
- Thread-safe with `sync.Once`

### 2. Service Categories
- **Local-only**: `db`, `redis`, `logger` (never HTTP)
- **Remote-only**: `stripe`, `sendgrid` (always HTTP)
- **Local+Remote**: Your business services (can be either)

### 3. Multi-Deployment
- Same code, different config
- Monolith ↔ Microservices
- No code change needed
- Transparent service resolution

### 4. Convention over Configuration
- Smart defaults
- Auto-generated routes
- Minimal config needed
- Configure when necessary

---

## 📚 Resources

**Documentation**:
- Home: [primadi.github.io/lokstra](https://primadi.github.io/lokstra/)
- Quick Start: [docs/00-introduction/quick-start.md](./00-introduction/quick-start.md)
- Examples: [docs/00-introduction/examples/](./00-introduction/examples/)
- API Reference: [docs/03-api-reference/](./03-api-reference/)

**Community**:
- GitHub: [github.com/primadi/lokstra](https://github.com/primadi/lokstra)
- Issues: [GitHub Issues](https://github.com/primadi/lokstra/issues)
- Discussions: [GitHub Discussions](https://github.com/primadi/lokstra/discussions)

**Learning Path**:
1. Read Why Lokstra (10 min)
2. Hello World (5 min)
3. Work through 7 examples (2 hours)
4. Build your first app (varies)

---

## 🎯 Quick Commands

```bash
# Install
go get github.com/primadi/lokstra

# Run example
cd docs/00-introduction/examples/01-hello-world
go run main.go

# Test
go test ./...

# Benchmark
go test -bench=. ./...

# Build
go build -o myapp main.go

# Run with config
./myapp -server "dev.api-server"
```

---

## 💡 Pro Tips

1. **Always use `service.LazyLoad`** for services in handlers
2. **Use struct params** for service methods (auto-binding)
3. **Let router auto-generate** from services when possible
4. **Start with monolith**, split to microservices later
5. **Print routes** during development for debugging
6. **Use fast path handlers** (without struct params) when simple
7. **Implement metadata** in RegisterServiceType, not in structs
8. **Test locally as monolith**, deploy as microservices

---

## 🚨 Common Mistakes

❌ **Registry lookup in handler**
```go
r.GET("/users", func() ([]User, error) {
    users := lokstra_registry.GetService[*UserService]("users")
    return users.GetAll() // SLOW
})
```

✅ **Use LazyLoad**
```go
var userService = service.LazyLoad[*UserService]("users")
r.GET("/users", func() ([]User, error) {
    return userService.MustGet().GetAll() // FAST
})
```

❌ **Metadata in service struct**
```go
func (s *UserService) GetResourceName() (string, string) {
    return "user", "users" // Optional, not required
}
```

✅ **Metadata in RegisterServiceType**
```go
lokstra_registry.RegisterServiceType(
    "user-service-factory",
    UserServiceFactory, UserServiceRemoteFactory,
    deploy.WithResource("user", "users"), // Correct place
    deploy.WithConvention("rest"),
)
```

❌ **Forgetting Build()**
```go
r.GET("/users", handler)
// Missing r.Build()!
```

✅ **Call Build() before ServeHTTP**
```go
r.GET("/users", handler)
if err := r.Build(); err != nil {
    log.Fatal(err)
}
```

---

## 📋 Checklist untuk Project Baru

- [ ] Install Lokstra: `go get github.com/primadi/lokstra`
- [ ] Setup project structure (main.go, services/, handlers/)
- [ ] Register services dengan metadata
- [ ] Create config.yaml untuk deployments
- [ ] Implement service methods dengan struct params
- [ ] Use LazyLoad untuk service dependencies
- [ ] Auto-generate routers dari services
- [ ] Add middleware (logging, CORS, auth)
- [ ] Test with `r.PrintRoutes()`
- [ ] Test locally as monolith
- [ ] Configure multi-deployment
- [ ] Deploy!

---

## 🎓 Next Steps

**After this cheatsheet**:
1. ✅ Run Hello World example (5 min)
2. ✅ Work through CRUD example (30 min)
3. ✅ Try multi-deployment (1 hour)
4. ✅ Read Essentials guide (4-6 hours)
5. ✅ Build your first app
6. ✅ Join community & contribute!

---

**Happy coding with Lokstra! 🚀**

📖 **Full docs**: [primadi.github.io/lokstra](https://primadi.github.io/lokstra/)  
💻 **GitHub**: [github.com/primadi/lokstra](https://github.com/primadi/lokstra)  
🌟 **Star if you like it!**

---

*Quick Reference | v2.x | Last updated: October 30, 2025*
