# 🚀 Lokstra Framework
## Modern Go Web Framework dengan Declarative Service Management

**Presentasi untuk Programmer Baru**

---

## 📋 Agenda

1. [Apa itu Lokstra?](#apa-itu-lokstra)
2. [Mengapa Lokstra Berbeda?](#mengapa-lokstra-berbeda)
3. [Fitur Unggulan](#fitur-unggulan)
4. [Demo Live: Hello World ke Production](#demo-live)
5. [Performa & Benchmarks](#performa--benchmarks)
6. [Arsitektur & Design Patterns](#arsitektur)
7. [Roadmap & Kontribusi](#roadmap--kontribusi)
8. [Q&A](#qa)

---

## 🎯 Apa itu Lokstra?

### Elevator Pitch (30 detik)

**Lokstra** adalah Go web framework yang:
- ✅ **Fleksibel** - 29 bentuk handler (tulis sesukamu)
- ✅ **Produktif** - Service jadi REST API otomatis
- ✅ **Scalable** - 1 binary → monolith ATAU microservices
- ✅ **Type-Safe** - Built-in lazy DI dengan generics
- ✅ **Production-Ready** - Sudah dipakai di aplikasi production

### Filosofi Utama

```
┌─────────────────────────────────────────────────────────┐
│ "Buat simple tetap simple,                              │
│  buat complex tetap manageable"                         │
└─────────────────────────────────────────────────────────┘
```

**Convention over Configuration** - Smart defaults, configure ketika butuh

---

## 💡 Mengapa Lokstra Berbeda?

### Problem: Framework Go Yang Ada

#### 1️⃣ Gin/Echo/Chi - Terlalu Rigid
```go
// Terkunci pada satu pattern
func Handler(c *gin.Context) {
    // Harus pakai c.JSON(), c.Bind()
    // Tidak fleksibel
}
```

#### 2️⃣ Standard Library - Terlalu Verbose
```go
// Boilerplate everywhere
func Handler(w http.ResponseWriter, r *http.Request) {
    // Manual JSON parse
    // Manual JSON encode
    // Manual error handling
    // Manual routing
}
```

#### 3️⃣ Enterprise Frameworks - Terlalu Complex
- Terlalu banyak konsep
- Learning curve terlalu tinggi
- Overkill untuk kebanyakan project

---

### Solution: Lokstra's Balanced Approach

#### 🎨 1. Handler Flexibility (29 Forms!)

```go
// Simple endpoint? Simple code!
r.GET("/ping", func() string {
    return "pong"
})

// Need error handling?
r.GET("/users", func() ([]User, error) {
    return db.GetAllUsers()
})

// Need request data?
type CreateUserReq struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

r.POST("/users", func(req *CreateUserReq) (*User, error) {
    return db.CreateUser(req)
})

// Need full control?
r.GET("/complex", func(ctx *request.Context) (*response.Response, error) {
    // Access headers, cookies, custom status
    return response.Success(data).WithStatus(201), nil
})
```

**🔑 Key Point**: Kamu pilih pattern yang sesuai dengan use case!

---

#### ⚡ 2. Zero Boilerplate dengan Service as Router

**Masalah Tradisional**:
```go
// Banyak boilerplate untuk CRUD
r.GET("/users", listUsers)
r.GET("/users/{id}", getUser)
r.POST("/users", createUser)
r.PUT("/users/{id}", updateUser)
r.DELETE("/users/{id}", deleteUser)

// Plus handler yang cuma forward ke service
func listUsers() ([]User, error) {
    return userService.GetAll()  // Boilerplate!
}
```

**Lokstra Solution**:
```go
// 1. Tulis business logic di service
type UserService struct {
    DB *service.Cached[*Database]
}

func (s *UserService) GetAll(p *GetAllParams) ([]User, error) {
    return s.DB.MustGet().Query("SELECT * FROM users")
}

func (s *UserService) GetByID(p *GetByIDParams) (*User, error) {
    return s.DB.MustGet().QueryOne("SELECT * FROM users WHERE id = ?", p.ID)
}

// 2. Register dengan metadata
lokstra_registry.RegisterServiceType(
    "user-service-factory",
    UserServiceFactory, UserServiceRemoteFactory,
    deploy.WithResource("user", "users"),
    deploy.WithConvention("rest"),
)

// 3. DONE! Routes auto-generated:
// GET    /users       → UserService.GetAll()
// GET    /users/{id}  → UserService.GetByID()
// POST   /users       → UserService.Create()
// PUT    /users/{id}  → UserService.Update()
// DELETE /users/{id}  → UserService.Delete()
```

**🎯 Result**: Fokus ke business logic, routing otomatis!

---

#### 🏗️ 3. One Binary, Multiple Deployments

**Game Changer**: Deploy sebagai monolith atau microservices **tanpa ubah code**!

```yaml
# config.yaml
service-definitions:
  user-service:
    type: user-service-factory
  order-service:
    type: order-service-factory
    depends-on: [user-service]

deployments:
  # Monolith: Semua service di 1 server
  monolith:
    servers:
      api-server:
        addr: ":8080"
        published-services:
          - user-service
          - order-service

  # Microservices: Terpisah per service
  microservices:
    servers:
      user-server:
        addr: ":8001"
        published-services: [user-service]
      
      order-server:
        addr: ":8002"
        published-services: [order-service]
        # user-service jadi remote call otomatis!
```

**Jalankan dengan flag**:
```bash
# Monolith
./app -server "monolith.api-server"

# Microservices
./app -server "microservices.user-server"  # Instance 1
./app -server "microservices.order-server" # Instance 2
```

**🎯 Benefit**:
- Start simple (monolith)
- Scale easily (microservices)
- No code change!
- Test locally as monolith
- Deploy as microservices

---

#### 🔧 4. Built-in Lazy Dependency Injection

**No External DI Framework Needed!**

```go
// Define service dengan dependencies
type OrderService struct {
    DB    *service.Cached[*Database]
    Users *service.Cached[*UserService]
    Cache *service.Cached[*CacheService]
}

// Register dengan dependency chain
lokstra_registry.RegisterServiceFactory("orders-factory", 
    func(deps map[string]any, config map[string]any) any {
        return &OrderService{
            DB:    service.Cast[*Database](deps["db"]),
            Users: service.Cast[*UserService](deps["users"]),
            Cache: service.Cast[*CacheService](deps["cache"]),
        }
    })

lokstra_registry.RegisterLazyService("orders", "orders-factory", 
    map[string]any{
        "depends-on": []string{"db", "users", "cache"},
    })

// Use dengan lazy loading (optimal!)
var orderService = service.LazyLoad[*OrderService]("orders")

r.GET("/orders", func() ([]Order, error) {
    return orderService.MustGet().GetAll()
    // First call: Creates & caches
    // Next calls: Returns cached instance (zero overhead)
})
```

**Benefits**:
- ✅ Type-safe dengan generics
- ✅ Zero reflection overhead setelah first load
- ✅ Circular dependency OK
- ✅ Memory efficient
- ✅ Thread-safe

---

## 🌟 Fitur Unggulan

### 1. 29 Handler Forms - Unmatched Flexibility

| Use Case | Handler Form | Code |
|----------|--------------|------|
| Simple return | `func() T` | `func() string { return "ok" }` |
| With errors | `func() (T, error)` | `func() ([]User, error) { ... }` |
| Request binding | `func(*Req) (T, error)` | `func(req *CreateReq) (*User, error)` |
| Full control | `func(*Context) (*Response, error)` | Full access |
| HTTP compat | `http.HandlerFunc` | Standard interface |

**Dan 24 bentuk lainnya!**

📖 [Lihat semua 29 forms →](./02-deep-dive/router/handler-forms.md)

---

### 2. Service Layer Architecture

```
┌──────────────────────────────────────────────┐
│           PRESENTATION LAYER                 │
│  (Router, Middleware, Handlers)              │
├──────────────────────────────────────────────┤
│           BUSINESS LAYER                     │
│  (Services with Lazy DI)                     │
├──────────────────────────────────────────────┤
│           DATA LAYER                         │
│  (Database, Cache, External APIs)            │
└──────────────────────────────────────────────┘
```

**Service Categories**:

#### 🔹 Local-Only Services (Infrastructure)
Tidak pernah exposed via HTTP:
- `db-service`, `redis-service`, `logger-service`
- Always loaded locally
- Used by other services

#### 🔹 Remote-Only Services (External APIs)
Third-party services:
- `stripe-service`, `sendgrid-service`, `twilio-service`
- Always HTTP calls
- Wrapped dengan Lokstra interface

#### 🔹 Local + Remote Services (Business Logic)
Your business services:
- `user-service`, `order-service`, `product-service`
- Can be local OR remote
- **Transparent**: Code tidak tahu local atau remote!

---

### 3. Multi-Deployment Patterns

```yaml
# Pattern 1: Monolith (Simple startup)
deployments:
  startup:
    servers:
      all-in-one:
        addr: ":8080"
        published-services: [users, orders, payments, products]

# Pattern 2: Partial Split (Scaling specific services)
deployments:
  scaling:
    servers:
      main-api:
        addr: ":8080"
        published-services: [users, orders]
      
      heavy-processing:
        addr: ":8081"
        published-services: [payments, analytics]

# Pattern 3: Full Microservices (Enterprise)
deployments:
  enterprise:
    servers:
      user-service:
        addr: ":8001"
        published-services: [users]
      
      order-service:
        addr: ":8002"
        published-services: [orders]
      
      payment-service:
        addr: ":8003"
        published-services: [payments]
```

**Same code, different topology!**

---

### 4. Middleware System

```go
// Global middleware
r.Use(loggingMiddleware, corsMiddleware)

// Group middleware
admin := r.Group("/admin")
admin.Use(authMiddleware)

// Route-specific middleware
r.GET("/special", specialHandler, rateLimitMiddleware)

// By-name middleware (from registry)
lokstra_registry.RegisterMiddleware("auth", authMiddleware)
r.Use("auth", "logging")
```

**Execution Order**:
```
Request
  ↓
Global Middleware [logging, cors]
  ↓
Group Middleware [auth]
  ↓
Route Middleware [rateLimit]
  ↓
Handler
  ↓
Response (back through middleware)
```

---

## 🎬 Demo Live: Hello World ke Production

### Step 1: Hello World (30 detik)

```go
package main

import (
    "time"
    "github.com/primadi/lokstra"
)

func main() {
    r := lokstra.NewRouter("api")
    
    r.GET("/", func() string {
        return "Hello, Lokstra!"
    })
    
    r.GET("/ping", func() string {
        return "pong"
    })
    
    app := lokstra.NewApp("hello", ":3000", r)
    server := lokstra.NewServer("my-server", app)
    server.Run(30 * time.Second)
}
```

```bash
go run main.go
# http://localhost:3000 → "Hello, Lokstra!"
# http://localhost:3000/ping → "pong"
```

---

### Step 2: JSON API (1 menit)

```go
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    r := lokstra.NewRouter("api")
    
    // Auto JSON response!
    r.GET("/users", func() ([]User, error) {
        users := []User{
            {ID: 1, Name: "Alice", Email: "alice@example.com"},
            {ID: 2, Name: "Bob", Email: "bob@example.com"},
        }
        return users, nil
    })
    
    // Request binding + validation
    type CreateUserReq struct {
        Name  string `json:"name" validate:"required"`
        Email string `json:"email" validate:"required,email"`
    }
    
    r.POST("/users", func(req *CreateUserReq) (*User, error) {
        // req already validated!
        newUser := &User{
            ID:    3,
            Name:  req.Name,
            Email: req.Email,
        }
        return newUser, nil
    })
    
    app := lokstra.NewApp("api", ":3000", r)
    server := lokstra.NewServer("my-server", app)
    server.Run(30 * time.Second)
}
```

---

### Step 3: With Services (5 menit)

```go
// service.go
type UserService struct {
    DB *service.Cached[*Database]
}

func (s *UserService) GetAll(p *GetAllParams) ([]User, error) {
    return s.DB.MustGet().FindAll()
}

func (s *UserService) GetByID(p *GetByIDParams) (*User, error) {
    return s.DB.MustGet().FindByID(p.ID)
}

func (s *UserService) Create(p *CreateParams) (*User, error) {
    return s.DB.MustGet().Insert(p)
}

// main.go
func main() {
    // Register services
    lokstra_registry.RegisterServiceType("db-factory", NewDatabase, nil)
    lokstra_registry.RegisterLazyService("db", "db-factory", nil)
    
    lokstra_registry.RegisterServiceFactory("users-factory",
        func(deps map[string]any, config map[string]any) any {
            return &UserService{
                DB: service.Cast[*Database](deps["db"]),
            }
        })
    lokstra_registry.RegisterLazyService("users", "users-factory",
        map[string]any{"depends-on": []string{"db"}})
    
    // Use service with lazy loading (optimal)
    var userService = service.LazyLoad[*UserService]("users")
    
    r := lokstra.NewRouter("api")
    
    r.GET("/users", func() ([]User, error) {
        return userService.MustGet().GetAll(&GetAllParams{})
    })
    
    r.GET("/users/{id}", func(ctx *request.Context) (*User, error) {
        id := ctx.PathParamInt("id")
        return userService.MustGet().GetByID(&GetByIDParams{ID: id})
    })
    
    app := lokstra.NewApp("api", ":3000", r)
    server := lokstra.NewServer("my-server", app)
    server.Run(30 * time.Second)
}
```

---

### Step 4: Auto-Router from Service (2 menit)

```go
// Metadata di RegisterServiceType
lokstra_registry.RegisterServiceType(
    "user-service-factory",
    UserServiceFactory, UserServiceRemoteFactory,
    deploy.WithResource("user", "users"),
    deploy.WithConvention("rest"),
)

// Auto-generate router
userRouter := lokstra_registry.NewRouterFromServiceType("user-service")

// DONE! Routes auto-created:
// GET    /users       → UserService.GetAll()
// GET    /users/{id}  → UserService.GetByID()
// POST   /users       → UserService.Create()
// PUT    /users/{id}  → UserService.Update()
// DELETE /users/{id}  → UserService.Delete()
```

---

### Step 5: Multi-Deployment (10 menit)

```yaml
# config.yaml
service-definitions:
  user-service:
    type: user-service-factory
    depends-on: [db]
  
  order-service:
    type: order-service-factory
    depends-on: [db, user-service]
  
  db:
    type: database-factory

deployments:
  # Development: Monolith
  dev:
    servers:
      api-server:
        addr: ":3000"
        published-services: [user-service, order-service]
  
  # Production: Microservices
  prod:
    servers:
      user-server:
        base-url: "https://user-api.example.com"
        addr: ":8001"
        published-services: [user-service]
      
      order-server:
        base-url: "https://order-api.example.com"
        addr: ":8002"
        published-services: [order-service]
        # user-service jadi remote call otomatis!
```

```bash
# Development
./app -server "dev.api-server"

# Production
./app -server "prod.user-server"   # User microservice
./app -server "prod.order-server"  # Order microservice
```

**🎯 No code change between deployments!**

---

## ⚡ Performa & Benchmarks

### Router Performance

Lokstra menggunakan Go's standard `ServeMux` (default) yang sangat cepat:

```
┌─────────────────────────────────────────────────────────────┐
│ Static Routes (ns/op) - Lower is Better                    │
├─────────────────────────────────────────────────────────────┤
│ ServeMux (Lokstra) ████ 200.7ns                             │
│ Chi Router         ███████ 350.7ns (1.7x slower)            │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ Path Parameters (ns/op)                                     │
├─────────────────────────────────────────────────────────────┤
│ ServeMux (Lokstra) █████ 278.2ns                            │
│ Chi Router         █████████████ 689.7ns (2.4x slower)      │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ Parallel Requests (ns/op)                                   │
├─────────────────────────────────────────────────────────────┤
│ ServeMux (Lokstra) ██ 48.25ns                               │
│ Chi Router         ████████ 204.9ns (4.2x slower)           │
└─────────────────────────────────────────────────────────────┘
```

**🏆 Winner: Lokstra menggunakan router tercepat!**

---

### Handler Performance

Fast path optimization untuk handler patterns:

```
┌─────────────────────────────────────────────────────────────┐
│ Handler Pattern Performance (ns/op)                         │
├─────────────────────────────────────────────────────────────┤
│ http.HandlerFunc          ██ 434ns   (Fastest)              │
│ func(*Context) error      ████ 1,626ns (Fast path)          │
│ func() (any, error)       ████ 1,651ns (Fast path)          │
│ func(*Context) any        ████ 1,645ns (Fast path)          │
│ func(*Req) (any, error)   ██████ 2,600ns (Reflection)       │
└─────────────────────────────────────────────────────────────┘
```

**Key Findings**:
- ✅ **21 dari 29 patterns** menggunakan fast path (< 2μs)
- ✅ **Reflection overhead** minimal (hanya ~1μs extra)
- ✅ **Production-ready** performance

---

### Service Lazy Loading

```
┌──────────────────────────────────────────────────────────────┐
│ Service Resolution Method (per request)                      │
├──────────────────────────────────────────────────────────────┤
│ Registry lookup           ████ ~50ns (map lookup overhead)   │
│ LazyLoad (cached)         █ ~5ns (cached pointer access)     │
└──────────────────────────────────────────────────────────────┘
```

**🎯 Recommendation**: Always use `service.LazyLoad` for optimal performance!

```go
// ❌ Not optimal: Registry lookup on every request
r.GET("/users", func() ([]User, error) {
    users := lokstra_registry.GetService[*UserService]("users")
    return users.GetAll()  // Map lookup on EVERY request
})

// ✅ Optimal: Cached service resolution
var userService = service.LazyLoad[*UserService]("users")

r.GET("/users", func() ([]User, error) {
    return userService.MustGet().GetAll()  // Cached access (fast!)
})
```

---

### Memory & Allocations

```
Handler Pattern Allocations:
├─ Fast path handlers:     26-29 allocs/op
├─ Reflection handlers:    34 allocs/op
└─ HTTP compat handlers:   11 allocs/op (lowest)

Total Memory per Request:
├─ Simple handlers:        ~2KB
├─ JSON response:          ~2KB
└─ Complex handlers:       ~3KB
```

**Production Metrics** (real application):
- Average response time: **< 10ms**
- P99 latency: **< 50ms**
- Memory per request: **< 5KB**
- Throughput: **10,000+ req/s** (single instance)

---

## 🏛️ Arsitektur

### Component Overview

```
┌──────────────────────────────────────────────┐
│                   SERVER                     │
│  (Lifecycle Management)                      │
│                                              │
│  ┌─────────────────────────────────────────┐ │
│  │               APP                       │ │
│  │  (HTTP Listener)                        │ │
│  │                                         │ │
│  │  ┌────────────────────────────────────┐ │ │
│  │  │           ROUTER                   │ │ │
│  │  │  (Route Matching + Middleware)     │ │ │
│  │  │                                    │ │ │
│  │  │  Route → [MW Chain] → Handler     │ │ │
│  │  └────────────────────────────────────┘ │ │
│  └─────────────────────────────────────────┘ │
└──────────────────────────────────────────────┘

Supporting Components:
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│   SERVICE   │  │ MIDDLEWARE  │  │   CONFIG    │
│ (Business)  │  │ (Filters)   │  │ (Settings)  │
└─────────────┘  └─────────────┘  └─────────────┘
```

---

### Request Flow

```
1. TCP Connection → App
   ↓
2. App.ServeHTTP() → Router
   ↓
3. Router matches route
   - Method: GET ✓
   - Path: /users/{id} ✓
   - Extract params: {id: "123"}
   ↓
4. Build middleware chain
   - Global: [logging, cors]
   - Group: [auth]
   - Route: [rateLimit]
   ↓
5. Execute chain
   logging.before()
     cors.before()
       auth.before()
         rateLimit.before()
           HANDLER (business logic)
         rateLimit.after()
       auth.after()
     cors.after()
   logging.after()
   ↓
6. Write response
```

---

### Service Resolution

```yaml
# Deployment topology
deployments:
  microservices:
    servers:
      user-server:
        addr: ":8001"
        published-services: [user-service]
      
      order-server:
        addr: ":8002"
        published-services: [order-service]
```

**Auto-Discovery**:
```
1. order-service depends on user-service
   ↓
2. Framework scans topology
   ↓
3. Finds: user-service → http://localhost:8001
   ↓
4. Creates: user-service-remote (HTTP proxy)
   ↓
5. Injects: Remote client into OrderService
   ↓
6. TRANSPARENT: Code tidak tahu local atau remote!
```

---

## 🗺️ Roadmap & Kontribusi

### 🎯 Current Status (v2.x)

**Production Ready Features**:
- ✅ 29 handler forms
- ✅ Service as router
- ✅ Multi-deployment
- ✅ Lazy DI
- ✅ Middleware system
- ✅ Path rewrites
- ✅ Remote services

---

### 🚀 Next Release (v2.1 - Q4 2025)

#### 1. 🎨 HTMX Support
Build modern web apps tanpa complex JavaScript:

```go
// Coming soon!
r.GET("/users", func() templ.Component {
    users := userService.GetAll()
    return views.UserList(users)
})

r.POST("/users", func(req *CreateUserReq) templ.Component {
    user := userService.Create(req)
    return views.UserRow(user)  // Partial update
})
```

**Features**:
- Template rendering (templ, html/template)
- HTMX helpers & middleware
- Form handling patterns
- Server-sent events (SSE)

---

#### 2. 🛠️ CLI Tools
Speed up development:

```bash
# Project scaffolding
lokstra new my-api --template=rest-api

# Code generation
lokstra generate service user
lokstra generate router api
lokstra generate middleware auth

# Hot reload
lokstra dev --port 3000

# Migrations
lokstra migrate create add_users_table
lokstra migrate up
```

---

#### 3. 📦 Standard Library

**Production-Ready Middleware**:
```go
// Metrics & monitoring
r.Use(middleware.Prometheus())
r.Use(middleware.OpenTelemetry())

// Authentication
r.Use(middleware.JWT(jwtConfig))
r.Use(middleware.OAuth2(oauthConfig))

// Rate limiting
r.Use(middleware.RateLimit(100, time.Minute))

// Security
r.Use(middleware.CSRF())
r.Use(middleware.SecureHeaders())
```

**Standard Services**:
```go
// Health checks
health := lokstra_registry.GetService[*HealthService]("health")
health.AddCheck("database", dbHealthCheck)

// Metrics
metrics := lokstra_registry.GetService[*MetricsService]("metrics")

// Tracing
tracer := lokstra_registry.GetService[*TracingService]("tracing")
```

---

### 🔮 Future Vision (v2.2+)

- 🔌 **Plugin System** - Extend framework
- 📊 **Admin Dashboard** - Built-in API explorer
- 🌐 **GraphQL Support** - Alternative to REST
- 🔄 **WebSocket Support** - Real-time communication
- 📝 **OpenAPI/Swagger** - Auto-generate docs

---

### 🤝 How to Contribute

#### Ways to Get Involved

1. **📖 Documentation**
   - Improve existing docs
   - Add tutorials
   - Translate to other languages

2. **💻 Code Contributions**
   - Fix bugs
   - Add features
   - Improve performance

3. **🧪 Testing & Examples**
   - Write tests
   - Create example apps
   - Share production use cases

4. **💡 Ideas & Feedback**
   - Feature suggestions
   - Bug reports
   - Architecture discussions

#### Getting Started

```bash
# Clone repo
git clone https://github.com/primadi/lokstra
cd lokstra

# Install dependencies
go mod download

# Run tests
go test ./...

# Run benchmarks
go test -bench=. ./...

# Check examples
cd docs/00-introduction/examples
```

#### Contribution Process

1. **Fork** the repository
2. **Create** feature branch (`git checkout -b feature/amazing`)
3. **Make** your changes
4. **Test** thoroughly
5. **Commit** with clear messages
6. **Push** to your fork
7. **Create** Pull Request

---

### 🎯 Ideal Contributors

**We're looking for**:
- 🔰 **Beginners** - Documentation, examples, tutorials
- 🧑‍💻 **Experienced Go devs** - Core features, optimizations
- 🏗️ **Architects** - Design patterns, best practices
- 📝 **Technical writers** - Documentation, blog posts
- 🎨 **Designers** - Logo, website, branding
- 🌍 **Translators** - Multi-language support

**Everyone is welcome!** 🎉

---

### 📊 Community Goals

| Metric | Current | Target (2026) | Target (2027) |
|--------|---------|---------------|---------------|
| GitHub Stars | - | 1,000+ | 10,000+ |
| Contributors | - | 50+ | 200+ |
| Production Deployments | Active | 100+ | 500+ |
| Documentation Languages | EN | EN, ID | EN, ID, ZH, JP |
| Example Apps | 7 | 20+ | 50+ |

---

## 📚 Resources

### Documentation

- **🏠 Home**: [lokstra.dev](https://primadi.github.io/lokstra/)
- **📖 Full Docs**: [docs/](./docs/)
- **🚀 Quick Start**: [docs/00-introduction/quick-start.md](./docs/00-introduction/quick-start.md)
- **💡 Examples**: [docs/00-introduction/examples/](./docs/00-introduction/examples/)
- **🎓 Essentials**: [docs/01-essentials/](./docs/01-essentials/)
- **🔬 Deep Dive**: [docs/02-deep-dive/](./docs/02-deep-dive/)
- **📘 API Reference**: [docs/03-api-reference/](./docs/03-api-reference/)

### Learning Path

**For Beginners** (2-3 hours):
1. Read [Why Lokstra?](./docs/00-introduction/why-lokstra.md) (10 min)
2. Follow [Quick Start](./docs/00-introduction/quick-start.md) (15 min)
3. Work through [Examples](./docs/00-introduction/examples/) (2 hours)

**For Intermediate** (1 week):
1. Study [Architecture](./docs/00-introduction/architecture.md) (30 min)
2. Complete [Essentials Guide](./docs/01-essentials/) (4-6 hours)
3. Build your first app (varies)

**For Advanced** (2-3 weeks):
1. Read [Deep Dive](./docs/02-deep-dive/) (8-10 hours)
2. Explore [API Reference](./docs/03-api-reference/) (as needed)
3. Contribute to framework (ongoing)

---

### Example Applications

1. **Hello World** - Minimal setup (5 min)
2. **Handler Forms** - Explore 29 patterns (10 min)
3. **CRUD API** - Complete service example (30 min)
4. **Multi-Deployment** - Monolith vs microservices (1 hour)
5. **External Services** - Third-party API integration (30 min)
6. **Remote Router** - Direct HTTP calls (20 min)
7. **Middleware** - Custom middleware patterns (45 min)

**Total learning time**: 4-6 hours from zero to production-ready!

---

## 🎤 Q&A

### Frequently Asked Questions

#### Q: Apakah Lokstra stable untuk production?
**A**: Ya! Sudah digunakan di aplikasi production. Active development dan maintenance.

#### Q: Bagaimana performance dibanding Gin/Echo?
**A**: Comparable. Tidak secepat raw fasthttp, tapi cukup untuk 99% use cases. ServeMux engine yang digunakan adalah salah satu router tercepat.

#### Q: Bisa migrate dari Gin/Echo?
**A**: Ya! Gradual migration possible. Mulai dengan fitur baru di Lokstra, keep existing code.

#### Q: Apa advantage terbesar Lokstra?
**A**: 
1. **Flexibility** (29 handler forms)
2. **Productivity** (zero boilerplate dengan service-as-router)
3. **Scalability** (monolith ↔ microservices tanpa code change)

#### Q: Learning curve nya gimana?
**A**: 
- **Basic usage**: 2-3 hours
- **Production ready**: 1 week
- **Advanced patterns**: 2-3 weeks

Lebih mudah dari framework enterprise, lebih powerful dari Gin/Echo.

#### Q: Ada breaking changes?
**A**: Mengikuti semantic versioning. Breaking changes hanya di major versions. API sedang stabilizing.

#### Q: Support untuk GraphQL/gRPC?
**A**: 
- GraphQL: Planned untuk v2.3+
- gRPC: Bisa digunakan bersamaan, tapi Lokstra focus ke REST/HTTP

#### Q: Bisa pakai dengan database apa saja?
**A**: Semua! Lokstra tidak tied ke specific database. Use any Go database library.

#### Q: Butuh IDE khusus?
**A**: Tidak. Any Go-compatible IDE (VS Code, GoLand, dll).

---

## 🎯 Call to Action

### Try Lokstra Today!

```bash
# Install
go get github.com/primadi/lokstra

# Create your first app
# (Copy Hello World example from slides)

# Run
go run main.go

# Open browser
# http://localhost:3000
```

---

### Join the Community

- 🌟 **Star on GitHub**: [github.com/primadi/lokstra](https://github.com/primadi/lokstra)
- 🐛 **Report Issues**: [GitHub Issues](https://github.com/primadi/lokstra/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/primadi/lokstra/discussions)
- 📧 **Contact**: [primadi@example.com](mailto:primadi@example.com)

---

### Share Your Experience

Built something with Lokstra?
- 📝 Write a blog post
- 🐦 Tweet about it
- 📺 Create a video tutorial
- 🗣️ Present at meetups

**Tag us**: `@lokstra` `#LokstraFramework`

---

## 🙏 Thank You!

### Credits

**Created by**: Primadi  
**Contributors**: Growing community  
**Inspired by**: Go philosophy, modern web frameworks  
**License**: MIT

---

### Let's Build Better Go APIs Together! 🚀

```
┌─────────────────────────────────────────────────────────┐
│                                                         │
│  "Start simple, scale effortlessly"                     │
│                                                         │
│  - Lokstra Framework                                    │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

---

**Questions? Let's discuss!** 💬

📖 **Documentation**: [primadi.github.io/lokstra](https://primadi.github.io/lokstra/)  
💻 **GitHub**: [github.com/primadi/lokstra](https://github.com/primadi/lokstra)  
🌟 **Star if you like it!**

---

*Presentasi ini dibuat dengan ❤️ menggunakan Markdown*  
*Dapat diexport ke PDF, PowerPoint, atau Google Slides*

---

## 📎 Appendix: Code Examples

### Complete CRUD Service Example

```go
package main

import (
    "time"
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/core/service"
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/core/deploy"
)

// Models
type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

// Service
type UserService struct {
    DB *service.Cached[*Database]
}

type GetAllParams struct{}
type GetByIDParams struct {
    ID int `path:"id"`
}
type CreateParams struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}
type UpdateParams struct {
    ID    int    `path:"id"`
    Name  string `json:"name"`
    Email string `json:"email" validate:"email"`
}
type DeleteParams struct {
    ID int `path:"id"`
}

func (s *UserService) GetAll(p *GetAllParams) ([]*User, error) {
    return s.DB.MustGet().FindAll()
}

func (s *UserService) GetByID(p *GetByIDParams) (*User, error) {
    return s.DB.MustGet().FindByID(p.ID)
}

func (s *UserService) Create(p *CreateParams) (*User, error) {
    return s.DB.MustGet().Insert(p)
}

func (s *UserService) Update(p *UpdateParams) (*User, error) {
    return s.DB.MustGet().Update(p)
}

func (s *UserService) Delete(p *DeleteParams) error {
    return s.DB.MustGet().Delete(p.ID)
}

// Factories
func UserServiceFactory(deps map[string]any, config map[string]any) any {
    return &UserService{
        DB: service.Cast[*Database](deps["db"]),
    }
}

func main() {
    // Register services
    lokstra_registry.RegisterServiceType("db-factory", NewDatabase, nil)
    lokstra_registry.RegisterLazyService("db", "db-factory", nil)
    
    lokstra_registry.RegisterServiceType(
        "user-service-factory",
        UserServiceFactory, nil,
        deploy.WithResource("user", "users"),
        deploy.WithConvention("rest"),
    )
    lokstra_registry.RegisterLazyService("users", "user-service-factory",
        map[string]any{"depends-on": []string{"db"}})
    
    // Auto-generate router
    userRouter := lokstra_registry.NewRouterFromServiceType("user-service")
    
    // Create app
    app := lokstra.NewApp("api", ":3000", userRouter)
    server := lokstra.NewServer("my-server", app)
    
    // Run
    server.Run(30 * time.Second)
}
```

**Generated routes:**
```
GET    /users       → UserService.GetAll()
GET    /users/{id}  → UserService.GetByID()
POST   /users       → UserService.Create()
PUT    /users/{id}  → UserService.Update()
DELETE /users/{id}  → UserService.Delete()
```

---

### Multi-Deployment Complete Example

See: [docs/00-introduction/examples/04-multi-deployment-yaml/](./docs/00-introduction/examples/04-multi-deployment-yaml/)

**3 deployment modes dari 1 codebase:**
1. Monolith (1 server, all services local)
2. Multi-port (multiple servers, all services local)
3. Microservices (multiple servers, remote service calls)

**No code change, just config!**

---

*End of Presentation* 🎉
