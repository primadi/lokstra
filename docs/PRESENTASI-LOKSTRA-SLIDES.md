# Lokstra Framework Presentation
## Slide Deck for New Programmers

---

<!-- Slide 1: Title -->
# 🚀 Lokstra Framework

## Modern Go Web Framework
### Build APIs the Smart Way

**Presentasi untuk Programmer Baru**

---

<!-- Slide 2: Whoami -->
# 👋 Tentang Project

**Lokstra** adalah Go web framework yang:
- Production-ready
- Open source (MIT License)
- Active development
- Growing community

**Created by**: Primadi  
**GitHub**: [github.com/primadi/lokstra](https://github.com/primadi/lokstra)

---

<!-- Slide 3: Elevator Pitch -->
# 🎯 Lokstra dalam 30 Detik

### 4 Killer Features:

1. **29 Handler Forms** - Tulis handler sesukamu
2. **Zero Boilerplate** - Service → REST API otomatis
3. **One Binary** - Monolith ↔ Microservices
4. **Built-in DI** - Lazy loading, type-safe

### Filosofi:
> "Buat simple tetap simple,  
> buat complex tetap manageable"

---

<!-- Slide 4: Problem Statement -->
# 😫 Problem: Framework Go Yang Ada

### 1. Standard Library - Terlalu Verbose
```go
func Handler(w http.ResponseWriter, r *http.Request) {
    // Manual JSON parse
    // Manual JSON encode
    // Manual error handling
    // Manual routing
}
```

### 2. Gin/Echo/Chi - Terlalu Rigid
```go
func Handler(c *gin.Context) {
    // Terkunci pada satu pattern
}
```

### 3. Enterprise Frameworks - Terlalu Complex
- Learning curve terlalu tinggi
- Overkill untuk most projects

---

<!-- Slide 5: Solution Overview -->
# 💡 Lokstra Solution

## Balanced Approach

| Aspect | Lokstra |
|--------|---------|
| **Flexibility** | 29 handler forms |
| **Productivity** | Auto-generated routers |
| **Scalability** | Multi-deployment support |
| **DX** | Minimal boilerplate |
| **Type Safety** | Generics + compile-time checks |
| **Performance** | Production-ready (10k+ req/s) |

---

<!-- Slide 6: Feature 1 - Handler Forms -->
# 🎨 Feature #1: Handler Flexibility

## 29 Forms, Choose What Fits!

```go
// Simple - no params, no errors
r.GET("/ping", func() string { return "pong" })

// With error handling (most common!)
r.GET("/users", func() ([]User, error) {
    return db.GetAllUsers()
})

// With request binding
r.POST("/users", func(req *CreateUserReq) (*User, error) {
    return db.CreateUser(req)
})

// Full control
r.GET("/api", func(ctx *Context) (*Response, error) {
    // Access headers, cookies, custom status
})
```

**21 dari 29 menggunakan fast path (< 2μs)**

---

<!-- Slide 7: Feature 2 - Service as Router -->
# ⚡ Feature #2: Zero Boilerplate

## Problem: Traditional Approach
```go
// Banyak boilerplate untuk CRUD
r.GET("/users", listUsers)
r.GET("/users/{id}", getUser)
r.POST("/users", createUser)
r.PUT("/users/{id}", updateUser)
r.DELETE("/users/{id}", deleteUser)

func listUsers() ([]User, error) {
    return userService.GetAll()  // Boilerplate!
}
```

---

<!-- Slide 8: Service as Router Solution -->
# ⚡ Feature #2: Service as Router

## Lokstra Solution
```go
// 1. Tulis business logic
type UserService struct {
    DB *service.Cached[*Database]
}

func (s *UserService) GetAll(p *GetAllParams) ([]User, error) {
    return s.DB.MustGet().Query("SELECT * FROM users")
}

func (s *UserService) GetByID(p *GetByIDParams) (*User, error) {
    return s.DB.MustGet().QueryOne("...", p.ID)
}

// 2. Register dengan metadata
lokstra_registry.RegisterServiceType(
    "user-service-factory",
    UserServiceFactory, UserServiceRemoteFactory,
    deploy.WithResource("user", "users"),
    deploy.WithConvention("rest"),
)

// 3. DONE! Routes auto-generated
```

---

<!-- Slide 9: Auto-Generated Routes -->
# ⚡ Auto-Generated Routes

```
FROM SERVICE METHODS:
├─ GetAll()    → GET    /users
├─ GetByID()   → GET    /users/{id}
├─ Create()    → POST   /users
├─ Update()    → PUT    /users/{id}
└─ Delete()    → DELETE /users/{id}
```

**🎯 Fokus ke business logic, routing otomatis!**

---

<!-- Slide 10: Feature 3 - Multi-Deployment -->
# 🏗️ Feature #3: One Binary, Many Deployments

```yaml
deployments:
  # Monolith: All services in one
  monolith:
    servers:
      api-server:
        addr: ":8080"
        published-services: [users, orders, payments]

  # Microservices: Separate per service
  microservices:
    servers:
      user-server:
        addr: ":8001"
        published-services: [users]
      order-server:
        addr: ":8002"
        published-services: [orders, payments]
```

---

<!-- Slide 11: Multi-Deployment Benefits -->
# 🏗️ One Binary, Many Deployments

## Run with Flag:
```bash
# Monolith
./app -server "monolith.api-server"

# Microservices
./app -server "microservices.user-server"  # Instance 1
./app -server "microservices.order-server" # Instance 2
```

## Benefits:
✅ Start simple (monolith)  
✅ Scale easily (microservices)  
✅ **No code change!**  
✅ Test locally as monolith  
✅ Deploy as microservices

---

<!-- Slide 12: Feature 4 - Built-in DI -->
# 🔧 Feature #4: Built-in Lazy DI

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
    map[string]any{"depends-on": []string{"db", "users", "cache"}})
```

---

<!-- Slide 13: Lazy Loading Benefits -->
# 🔧 Lazy Loading Benefits

```go
// ❌ Not optimal: Map lookup per request
r.GET("/orders", func() ([]Order, error) {
    orders := lokstra_registry.GetService[*OrderService]("orders")
    return orders.GetAll()  // Lookup overhead
})

// ✅ Optimal: Cached resolution
var orderService = service.LazyLoad[*OrderService]("orders")

r.GET("/orders", func() ([]Order, error) {
    return orderService.MustGet().GetAll()  // Cached!
})
```

**Performance**:
- Registry lookup: ~50ns per request
- LazyLoad cached: ~5ns per request
- **10x faster!**

---

<!-- Slide 14: Demo Time -->
# 🎬 Live Demo

## From Hello World to Production

1. **Hello World** (30 seconds)
2. **JSON API** (1 minute)
3. **With Services** (5 minutes)
4. **Auto-Router** (2 minutes)
5. **Multi-Deployment** (10 minutes)

**Let's code! 💻**

---

<!-- Slide 15: Demo 1 - Hello World -->
# Demo #1: Hello World (30s)

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
```

---

<!-- Slide 16: Demo 2 - JSON API -->
# Demo #2: JSON API (1 min)

```go
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

r.GET("/users", func() ([]User, error) {
    return []User{
        {ID: 1, Name: "Alice", Email: "alice@example.com"},
        {ID: 2, Name: "Bob", Email: "bob@example.com"},
    }, nil
})

type CreateUserReq struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

r.POST("/users", func(req *CreateUserReq) (*User, error) {
    return &User{ID: 3, Name: req.Name, Email: req.Email}, nil
})
```

**Auto JSON response + validation!**

---

<!-- Slide 17: Demo 3 - Services -->
# Demo #3: With Services (5 min)

```go
type UserService struct {
    DB *service.Cached[*Database]
}

func (s *UserService) GetAll(p *GetAllParams) ([]User, error) {
    return s.DB.MustGet().FindAll()
}

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
    
    // Use with lazy loading
    var userService = service.LazyLoad[*UserService]("users")
    
    r.GET("/users", func() ([]User, error) {
        return userService.MustGet().GetAll(&GetAllParams{})
    })
}
```

---

<!-- Slide 18: Demo 4 - Auto Router -->
# Demo #4: Auto-Router (2 min)

```go
// Register dengan metadata
lokstra_registry.RegisterServiceType(
    "user-service-factory",
    UserServiceFactory, UserServiceRemoteFactory,
    deploy.WithResource("user", "users"),
    deploy.WithConvention("rest"),
)

// Auto-generate router
userRouter := lokstra_registry.NewRouterFromServiceType("user-service")

// DONE! Routes created:
// GET    /users       → UserService.GetAll()
// GET    /users/{id}  → UserService.GetByID()
// POST   /users       → UserService.Create()
// PUT    /users/{id}  → UserService.Update()
// DELETE /users/{id}  → UserService.Delete()

app := lokstra.NewApp("api", ":3000", userRouter)
```

---

<!-- Slide 19: Demo 5 - Multi Deployment -->
# Demo #5: Multi-Deployment (10 min)

```yaml
# config.yaml
service-definitions:
  user-service:
    type: user-service-factory
  order-service:
    type: order-service-factory
    depends-on: [user-service]

deployments:
  dev:
    servers:
      api-server:
        addr: ":3000"
        published-services: [user-service, order-service]
  
  prod:
    servers:
      user-server:
        addr: ":8001"
        published-services: [user-service]
      order-server:
        addr: ":8002"
        published-services: [order-service]
```

**No code change between dev and prod!**

---

<!-- Slide 20: Performance -->
# ⚡ Performance & Benchmarks

## Router Performance (ns/op)
```
Static Routes:
├─ ServeMux (Lokstra):  200.7ns  ⭐
└─ Chi Router:          350.7ns  (1.7x slower)

Path Parameters:
├─ ServeMux (Lokstra):  278.2ns  ⭐
└─ Chi Router:          689.7ns  (2.4x slower)

Parallel Requests:
├─ ServeMux (Lokstra):  48.25ns  ⭐
└─ Chi Router:          204.9ns  (4.2x slower)
```

**🏆 Lokstra menggunakan router tercepat!**

---

<!-- Slide 21: Handler Performance -->
# ⚡ Handler Performance

```
Handler Pattern (ns/op):
├─ http.HandlerFunc:          434ns   (Fastest)
├─ func(*Context) error:      1,626ns (Fast path)
├─ func() (any, error):       1,651ns (Fast path)
├─ func(*Context) any:        1,645ns (Fast path)
└─ func(*Req) (any, error):   2,600ns (Reflection)
```

**Key Findings**:
- ✅ 21/29 patterns use fast path (< 2μs)
- ✅ Reflection overhead minimal (~1μs)
- ✅ Production-ready performance

**Real Application**:
- Average: < 10ms
- P99: < 50ms
- Throughput: 10,000+ req/s

---

<!-- Slide 22: Architecture Overview -->
# 🏛️ Architecture

```
┌──────────────────────────────────────┐
│           SERVER                     │
│  (Lifecycle Management)              │
│  ┌────────────────────────────────┐  │
│  │          APP                   │  │
│  │  (HTTP Listener)               │  │
│  │  ┌──────────────────────────┐  │  │
│  │  │        ROUTER            │  │  │
│  │  │  (Route + Middleware)    │  │  │
│  │  │                          │  │  │
│  │  │  Route → [MW] → Handler  │  │  │
│  │  └──────────────────────────┘  │  │
│  └────────────────────────────────┘  │
└──────────────────────────────────────┘

Supporting:
├─ SERVICE (Business Logic)
├─ MIDDLEWARE (Filters)
└─ CONFIG (Settings)
```

---

<!-- Slide 23: Request Flow -->
# 🔄 Request Flow

```
1. TCP → App
   ↓
2. App → Router
   ↓
3. Match route & extract params
   ↓
4. Build middleware chain
   - Global: [logging, cors]
   - Group: [auth]
   - Route: [rateLimit]
   ↓
5. Execute chain
   logging → cors → auth → rateLimit → HANDLER
   ↓
6. Response (back through middleware)
```

**Clean & predictable!**

---

<!-- Slide 24: Service Categories -->
# 🔧 Service Categories

## 1. Local-Only (Infrastructure)
Never exposed via HTTP:
- `db`, `redis`, `logger`, `queue`
- Always loaded locally

## 2. Remote-Only (External APIs)
Third-party services:
- `stripe`, `sendgrid`, `twilio`
- Always HTTP calls

## 3. Local + Remote (Business Logic)
Your services:
- `user-service`, `order-service`
- **Can be local OR remote**
- Code doesn't know which!

---

<!-- Slide 25: Service Resolution -->
# 🔧 Service Auto-Discovery

```yaml
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

**Auto-Discovery Process**:
```
1. order-service depends on user-service
2. Framework scans topology
3. Finds: user-service → http://localhost:8001
4. Creates: user-service-remote (HTTP proxy)
5. Injects: Remote client into OrderService
6. TRANSPARENT: Code tidak tahu!
```

---

<!-- Slide 26: Roadmap -->
# 🗺️ Roadmap

## Current (v2.x) ✅
- 29 handler forms
- Service as router
- Multi-deployment
- Lazy DI
- Middleware system

## Next (v2.1 - Q4 2025) 🚀
- 🎨 HTMX Support
- 🛠️ CLI Tools
- 📦 Standard Middleware Library
- 📦 Standard Service Library

## Future (v2.2+) 🔮
- Plugin System
- Admin Dashboard
- GraphQL Support
- WebSocket Support
- OpenAPI/Swagger

---

<!-- Slide 27: Community Goals -->
# 📊 Community Goals

| Metric | Target 2026 | Target 2027 |
|--------|-------------|-------------|
| GitHub Stars | 1,000+ | 10,000+ |
| Contributors | 50+ | 200+ |
| Production Apps | 100+ | 500+ |
| Languages | EN, ID | EN, ID, ZH, JP |
| Example Apps | 20+ | 50+ |

**Join us in building the future!** 🚀

---

<!-- Slide 28: How to Contribute -->
# 🤝 How to Contribute

## Ways to Get Involved:

1. **📖 Documentation**
   - Improve docs, add tutorials, translations

2. **💻 Code**
   - Fix bugs, add features, optimize

3. **🧪 Testing**
   - Write tests, create examples, share use cases

4. **💡 Ideas**
   - Feature suggestions, bug reports, discussions

**Everyone is welcome!** 🎉

---

<!-- Slide 29: Ideal Contributors -->
# 🎯 We Need You!

**Looking for**:
- 🔰 **Beginners** - Docs, examples, tutorials
- 🧑‍💻 **Go devs** - Features, optimizations
- 🏗️ **Architects** - Design patterns
- 📝 **Writers** - Documentation, blogs
- 🎨 **Designers** - Logo, branding
- 🌍 **Translators** - Multi-language

**Your contribution matters!**

---

<!-- Slide 30: Getting Started -->
# 🚀 Getting Started

```bash
# Install
go get github.com/primadi/lokstra

# Clone repo
git clone https://github.com/primadi/lokstra
cd lokstra

# Run examples
cd docs/00-introduction/examples/01-hello-world
go run main.go

# Run tests
go test ./...

# Run benchmarks
go test -bench=. ./...
```

---

<!-- Slide 31: Learning Path -->
# 📚 Learning Path

**For Beginners** (2-3 hours):
1. Read [Why Lokstra?](./docs/00-introduction/why-lokstra.md)
2. Follow [Quick Start](./docs/00-introduction/quick-start.md)
3. Work through [Examples](./docs/00-introduction/examples/)

**For Intermediate** (1 week):
1. Study [Architecture](./docs/00-introduction/architecture.md)
2. Complete [Essentials](./docs/01-essentials/)
3. Build your first app

**For Advanced** (2-3 weeks):
1. Read [Deep Dive](./docs/02-deep-dive/)
2. Explore [API Reference](./docs/03-api-reference/)
3. Contribute to framework

---

<!-- Slide 32: Resources -->
# 📖 Resources

**Documentation**:
- 🏠 Home: [primadi.github.io/lokstra](https://primadi.github.io/lokstra/)
- 📖 Full Docs: [docs/](./docs/)
- 🚀 Quick Start: [Quick Start Guide](./docs/00-introduction/quick-start.md)
- 💡 Examples: [7 Progressive Examples](./docs/00-introduction/examples/)

**Community**:
- 💻 GitHub: [github.com/primadi/lokstra](https://github.com/primadi/lokstra)
- 🐛 Issues: [GitHub Issues](https://github.com/primadi/lokstra/issues)
- 💬 Discussions: [GitHub Discussions](https://github.com/primadi/lokstra/discussions)

---

<!-- Slide 33: Q&A -->
# 🎤 Q&A

## Common Questions

**Q: Stable untuk production?**  
A: Ya! Sudah digunakan di production apps.

**Q: Performance vs Gin/Echo?**  
A: Comparable. ServeMux salah satu router tercepat.

**Q: Bisa migrate dari Gin/Echo?**  
A: Ya! Gradual migration possible.

**Q: Learning curve?**  
A: Basic: 2-3 jam, Production: 1 minggu.

**Q: Breaking changes?**  
A: Semantic versioning. Breaking hanya di major versions.

---

<!-- Slide 34: Why Choose Lokstra -->
# 🌟 Why Choose Lokstra?

## Unique Value Proposition

| Feature | Other Frameworks | Lokstra |
|---------|-----------------|---------|
| Handler Forms | 1 | **29** ✨ |
| Service Layer | Manual | **Built-in** ✨ |
| Auto-Router | No | **Yes** ✨ |
| Multi-Deploy | No | **Yes** ✨ |
| DI System | External | **Built-in** ✨ |
| Lazy Loading | Manual | **Built-in** ✨ |

**Productivity + Flexibility + Scalability**

---

<!-- Slide 35: Success Stories -->
# 📈 Success Stories

## Real-World Usage

**Production Applications**:
- REST APIs with 10k+ req/s
- Microservices architectures
- Monolithic applications
- API gateways

**Performance Metrics**:
- Average response: < 10ms
- P99 latency: < 50ms
- Uptime: 99.9%+
- Memory efficient: < 5KB per request

**Developer Experience**:
- Reduced boilerplate: 60-80%
- Faster development: 2-3x
- Easier maintenance

---

<!-- Slide 36: Comparison Matrix -->
# 📊 Framework Comparison

| Feature | stdlib | Gin | Echo | Lokstra |
|---------|--------|-----|------|---------|
| Handler Flex | ❌ | ⚠️ | ⚠️ | ✅ (29) |
| Auto JSON | ❌ | ✅ | ✅ | ✅ |
| Service Layer | ❌ | ❌ | ❌ | ✅ |
| DI System | ❌ | ❌ | ❌ | ✅ |
| Multi-Deploy | ❌ | ❌ | ❌ | ✅ |
| Auto-Router | ❌ | ❌ | ❌ | ✅ |
| Performance | ⚡⚡⚡ | ⚡⚡⚡ | ⚡⚡⚡ | ⚡⚡⚡ |
| Learning | Easy | Easy | Easy | Medium |

**Lokstra = Best of All Worlds**

---

<!-- Slide 37: When to Use -->
# ✅ When to Use Lokstra

**Perfect For**:
- ✅ REST APIs (sweet spot!)
- ✅ Microservices architecture
- ✅ Monolith with migration plan
- ✅ Service-heavy applications
- ✅ Multi-environment deployments

**Consider Alternatives For**:
- ❌ GraphQL-first APIs (use gqlgen)
- ❌ Pure gRPC services (use grpc-go)
- ❌ Static file servers (stdlib enough)
- ❌ Learning Go (start with stdlib)
- ❌ Extreme performance needs (use fasthttp)

---

<!-- Slide 38: Call to Action -->
# 🎯 Call to Action

## Try Lokstra Today!

```bash
# 1. Install
go get github.com/primadi/lokstra

# 2. Create first app (5 minutes)
# (Copy Hello World example)

# 3. Explore examples (2 hours)
cd docs/00-introduction/examples

# 4. Build your app!
```

**Join the revolution! 🚀**

---

<!-- Slide 39: Join Community -->
# 🌍 Join the Community

**Star on GitHub**:
[github.com/primadi/lokstra](https://github.com/primadi/lokstra)

**Get Involved**:
- 🐛 Report bugs
- 💡 Suggest features
- 📝 Write tutorials
- 💻 Contribute code
- 🗣️ Spread the word

**Tag us**: `@lokstra` `#LokstraFramework`

---

<!-- Slide 40: Thank You -->
# 🙏 Thank You!

## Let's Build Better Go APIs Together!

```
┌─────────────────────────────────────┐
│                                     │
│  "Start simple, scale effortlessly" │
│                                     │
│  - Lokstra Framework                │
│                                     │
└─────────────────────────────────────┘
```

**Questions? Let's discuss!** 💬

---

📖 **Documentation**: [primadi.github.io/lokstra](https://primadi.github.io/lokstra/)  
💻 **GitHub**: [github.com/primadi/lokstra](https://github.com/primadi/lokstra)  
📧 **Contact**: primadi@example.com

🌟 **Star if you like it!**

---

*End of Presentation*

**#LokstraFramework #GoLang #WebFramework #Microservices**
