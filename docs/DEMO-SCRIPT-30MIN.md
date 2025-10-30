# Demo Script: 30-Minute Live Code Exploration

**Format**: Walk through existing examples  
**Audience**: Programmers  
**Goal**: Show real working code, not slides  

---

## 🎯 Preparation Checklist

### Before Demo (15 minutes setup)

```bash
# 1. Clone repo ke folder mudah diakses
cd ~/demos
git clone https://github.com/primadi/lokstra
cd lokstra/docs/00-introduction/examples

# 2. Test run examples
cd 01-hello-world && go run main.go  # Ctrl+C after test
cd ../03-crud-api && go run main.go  # Ctrl+C after test
cd ../04-multi-deployment-yaml && go run . -server "monolith.all-in-one"  # Test

# 3. Open VS Code
cd ~/demos/lokstra
code .

# 4. Prepare terminal windows
# Terminal 1: For running apps
# Terminal 2: For testing (curl/httpie)

# 5. Bookmark URLs
# - http://localhost:3000
# - http://localhost:3001
# - http://localhost:3002
```

### Screen Layout
```
┌────────────────────────────┬─────────────────────────┐
│  VS Code                   │  Terminal 1 (Run)       │
│  - Show code               │  $ go run main.go       │
│  - Navigate files          │                         │
│  - Highlight concepts      │  Terminal 2 (Test)      │
│                            │  $ curl localhost:3000  │
└────────────────────────────┴─────────────────────────┘
```

---

## 📝 Demo Script

### [00:00-01:00] Opening

**Say**:
> "Okay, slides done. Now let's see REAL code. Everything I'm showing you is in the repo. You can clone and run this yourself. Let me share my screen..."

**Do**:
```bash
# Show terminal
pwd  # ~/demos/lokstra/docs/00-introduction/examples
ls   # Show 8 example folders
```

**Say**:
> "8 examples here. We'll walk through 3-4 key ones that show Lokstra's power. Starting simple, then building up."

---

### [01:00-06:00] Example 01: Hello World

**Navigate**:
```bash
cd 01-hello-world
code .  # Or show in VS Code already open
```

**Show**: `README.md` (quick glance)
```markdown
# Example 01: Hello World
Minimal Lokstra setup in ~10 lines...
```

**Say**:
> "Simplest possible Lokstra app. Let me show you the code..."

**Show**: `main.go` (explain while scrolling)

```go
package main

import (
    "time"
    "github.com/primadi/lokstra"
)

func main() {
    // 1. Create router
    r := lokstra.NewRouter("api")
    
    // 2. Add routes - berbagai handler forms
    r.GET("/", func() string {
        return "Hello, Lokstra!"  // Form 1: Return string
    })
    
    r.GET("/ping", func() string {
        return "pong"
    })
    
    r.GET("/time", func() map[string]any {
        return map[string]any{  // Form 2: Return map
            "current_time": time.Now(),
            "message": "Auto JSON!",
        }
    })
    
    // 3. Create app & server
    app := lokstra.NewApp("hello", ":3000", r)
    server := lokstra.NewServer("my-server", app)
    
    // 4. Run with graceful shutdown
    server.Run(30 * time.Second)
}
```

**Highlight**:
- ✅ No boilerplate
- ✅ Multiple handler forms (return string, map)
- ✅ Auto JSON encoding
- ✅ Graceful shutdown built-in

**Run**:
```bash
# Terminal 1
go run main.go
# Show output: "Server started on :3000"

# Terminal 2 (test)
curl http://localhost:3000/
# Output: Hello, Lokstra!

curl http://localhost:3000/ping
# Output: pong

curl http://localhost:3000/time
# Output: {"current_time":"2025-10-30T10:30:00Z","message":"Auto JSON!"}
```

**Say**:
> "See that? map[string]any automatically becomes JSON. No manual encoding. This is just 1 of 29 handler forms Lokstra supports. Let's see more..."

**Stop app**: `Ctrl+C`

---

### [06:00-13:00] Example 03: CRUD API with Services

**Navigate**:
```bash
cd ../03-crud-api
ls  # Show structure
```

**Show**: File structure
```
03-crud-api/
├── main.go              # Entry point
├── service.go           # UserService
├── handlers.go          # HTTP handlers (optional alternative)
└── README.md
```

**Say**:
> "This is more realistic. CRUD API with service layer, lazy DI, and proper architecture. Let me show the service first..."

**Show**: `service.go` (key parts)

```go
// Business logic in service
type UserService struct {
    // Lazy-loaded dependencies
    DB *service.Cached[*Database]
}

// Service method dengan struct param (important!)
func (s *UserService) GetAll(p *GetAllParams) ([]*User, error) {
    return s.DB.MustGet().FindAll()
}

func (s *UserService) GetByID(p *GetByIDParams) (*User, error) {
    return s.DB.MustGet().FindByID(p.ID)
}

func (s *UserService) Create(p *CreateParams) (*User, error) {
    // Validation happens automatically via struct tags
    return s.DB.MustGet().Insert(p)
}
```

**Highlight**:
- ✅ Clean service layer
- ✅ Lazy-loaded dependencies (DB)
- ✅ Struct params for auto-binding
- ✅ Error handling built-in

**Show**: `main.go` (registration)

```go
func main() {
    // Register services
    lokstra_registry.RegisterServiceType("db-factory", 
        NewDatabase, nil)
    lokstra_registry.RegisterLazyService("db", 
        "db-factory", nil)
    
    lokstra_registry.RegisterServiceFactory("users-factory",
        func(deps map[string]any, config map[string]any) any {
            return &UserService{
                DB: service.Cast[*Database](deps["db"]),
            }
        })
    lokstra_registry.RegisterLazyService("users", 
        "users-factory",
        map[string]any{"depends-on": []string{"db"}})
    
    // Use with lazy loading (OPTIMAL!)
    var userService = service.LazyLoad[*UserService]("users")
    
    r := lokstra.NewRouter("api")
    
    // Handler menggunakan service
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
    
    // ... more routes
}
```

**Highlight**:
- ✅ Service registration dengan dependencies
- ✅ `service.LazyLoad` - cached resolution (fast!)
- ✅ Handler super clean - just call service
- ✅ Auto JSON request/response

**Say**:
> "Notice the pattern: LazyLoad di variable level, bukan di dalam handler. Ini penting untuk performance. First call creates service, subsequent calls use cached instance. Zero overhead."

**Run**:
```bash
# Terminal 1
go run .
# Output: Server started...

# Terminal 2 (test)
# GET all users
curl http://localhost:3001/users

# GET single user
curl http://localhost:3001/users/1

# POST create user
curl -X POST http://localhost:3001/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com"}'

# PUT update user
curl -X PUT http://localhost:3001/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Jane Doe","email":"jane@example.com"}'
```

**Say**:
> "Full CRUD working. Auto JSON. Validation built-in. Service layer clean. This is how you structure real Lokstra apps. But wait, there's more..."

**Stop app**: `Ctrl+C`

---

### [13:00-21:00] Example 04: Multi-Deployment (KILLER DEMO ⭐)

**Say**:
> "This is the KILLER feature. Same code, multiple deployments. Watch this..."

**Navigate**:
```bash
cd ../04-multi-deployment-yaml
ls
```

**Show**: File structure
```
04-multi-deployment-yaml/
├── config.yaml          # ⭐ Deployment config
├── main.go              # Entry point
├── user_service.go      # User service
├── order_service.go     # Order service (depends on user!)
└── README.md
```

**Show**: `config.yaml` (explain while scrolling)

```yaml
# Service definitions (global)
service-definitions:
  user-service:
    type: user-service-factory
    depends-on: [db]
  
  order-service:
    type: order-service-factory
    depends-on: [db, user-service]  # Depends on user!
  
  db:
    type: database-factory

deployments:
  # Deployment 1: MONOLITH (all services local)
  monolith:
    servers:
      all-in-one:
        base-url: "http://localhost"
        addr: ":3003"
        published-services:
          - user-service
          - order-service
        # Both services in same process!
  
  # Deployment 2: MICROSERVICES (services separated)
  microservices:
    servers:
      user-server:
        base-url: "http://localhost"
        addr: ":3004"
        published-services:
          - user-service
      
      order-server:
        base-url: "http://localhost"
        addr: ":3005"
        published-services:
          - order-service
        # order-service calls user-service via HTTP automatically!
```

**Highlight**:
- ✅ Same service definitions
- ✅ Different deployment topology
- ✅ Framework auto-detects local vs remote

**Show**: `order_service.go` (key part)

```go
type OrderService struct {
    DB    *service.Cached[*Database]
    Users *service.Cached[IUserService]  // Interface!
}

func (s *OrderService) Create(p *CreateParams) (*Order, error) {
    // Get user - framework handles local OR remote automatically!
    user, err := s.Users.MustGet().GetByID(&GetByIDParams{
        ID: p.UserID,
    })
    
    // In monolith: direct method call (fast)
    // In microservices: HTTP call (transparent)
    
    // ... create order
}
```

**Say**:
> "See that? OrderService doesn't know if UserService is local or remote. Framework handles it. Let me show you both deployments..."

**Demo 1: Run as MONOLITH**:
```bash
# Terminal 1
go run . -server "monolith.all-in-one"
# Output: 
# Starting server: monolith.all-in-one
# Loaded services: [db, user-service, order-service]
# Published services: [user-service, order-service]
# Server listening on :3003

# Terminal 2 (test)
# Create user
curl -X POST http://localhost:3003/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'
# Output: {"id":1,"name":"Alice",...}

# Create order (user-service call is LOCAL)
curl -X POST http://localhost:3003/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id":1,"product":"Laptop","amount":1000}'
# Output: {"id":1,"user_id":1,"product":"Laptop",...}
```

**Say**:
> "Working. Both services in same process. Now watch - same code, microservices deployment..."

**Stop app**: `Ctrl+C`

**Demo 2: Run as MICROSERVICES**:
```bash
# Terminal 1 - User Service
go run . -server "microservices.user-server"
# Output:
# Starting server: microservices.user-server
# Loaded services: [db, user-service]
# Published services: [user-service]
# Server listening on :3004

# Open NEW Terminal 3 - Order Service
go run . -server "microservices.order-server"
# Output:
# Starting server: microservices.order-server
# Loaded services: [db, order-service]
# Remote services: [user-service → http://localhost:3004]
# Published services: [order-service]
# Server listening on :3005

# Terminal 2 (test)
# Create user (ke user-service directly)
curl -X POST http://localhost:3004/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Bob","email":"bob@example.com"}'
# Output: {"id":2,"name":"Bob",...}

# Create order (ke order-service, calls user-service via HTTP!)
curl -X POST http://localhost:3005/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id":2,"product":"Mouse","amount":50}'
# Output: {"id":2,"user_id":2,"product":"Mouse",...}
```

**Say**:
> "BOOM! Same code, different deployment. Look at Terminal 3 - it says 'Remote services: user-service → http://localhost:3004'. Framework auto-detected it! OrderService thinks it's calling local UserService, but it's actually HTTP. Zero code change!"

**Show**: Both terminals side by side
```
Terminal 1 (User Service :3004)    Terminal 3 (Order Service :3005)
--------------------------------   ----------------------------------
Server listening on :3004          Server listening on :3005
[LOG] GET /users/2                 [LOG] Calling remote user-service
                                   [LOG] POST /orders
```

**Highlight**:
- ✅ NO code change between deployments
- ✅ Framework auto-discovers services
- ✅ Transparent local vs remote calls
- ✅ Config-driven architecture

**Say**:
> "This is the killer feature. Start as monolith for simplicity. Scale to microservices when needed. No refactoring required!"

**Stop apps**: `Ctrl+C` both terminals

---

### [21:00-26:00] Example 02: Handler Forms (BONUS)

**Say** (if time permits):
> "Let me quickly show you the 29 handler forms I mentioned..."

**Navigate**:
```bash
cd ../02-handler-forms
```

**Show**: `main.go` (quick scroll, highlight variety)

```go
// Form 1: Simple return
r.GET("/ping", func() string { return "pong" })

// Form 2: Return with error
r.GET("/users", func() ([]User, error) { ... })

// Form 3: Context + error
r.GET("/user", func(ctx *Context) error { ... })

// Form 4: Context + any + error
r.GET("/data", func(ctx *Context) (any, error) { ... })

// Form 5: Struct param (auto-binding!)
r.POST("/user", func(req *CreateReq) (*User, error) { ... })

// Form 6: HTTP compatible
r.GET("/std", http.HandlerFunc(func(w, r) { ... }))

// ... and 23 more variations!
```

**Say**:
> "29 forms total. Use what makes sense for your use case. Simple endpoint? Simple form. Need full control? Use Context. Need auto-binding? Use struct param. Framework adapts to you."

**Run** (quick test if time):
```bash
go run .
# Test 2-3 endpoints to show variety
curl http://localhost:3002/ping
curl http://localhost:3002/users
```

**Stop app**: `Ctrl+C`

---

### [26:00-30:00] Wrap-up & Documentation Tour

**Say**:
> "Okay, that's the live demo. You saw:
> - Example 01: Hello World in 10 lines
> - Example 03: Full CRUD with services
> - Example 04: Monolith OR microservices, same code
> - Example 02: 29 handler forms

> Everything is in the repo. Let me show you where to go next..."

**Navigate**: Show documentation structure
```bash
cd ../../..  # Back to root
ls docs/
```

**Show**: VS Code Explorer (docs folder)
```
docs/
├── 00-introduction/
│   ├── examples/          # ← You saw 3 of 8 examples
│   ├── quick-start.md     # ← Start here after presentation
│   ├── why-lokstra.md
│   ├── architecture.md
│   └── key-features.md
├── 01-essentials/         # ← Deep dive tutorials
├── 02-deep-dive/          # ← Advanced patterns
├── 03-api-reference/      # ← Complete API docs
└── ROADMAP.md             # ← What's coming
```

**Say**:
> "After this presentation:
> 1. Clone the repo
> 2. Run all 8 examples (they're all documented)
> 3. Read quick-start.md (10 minutes)
> 4. Build something!
> 
> All examples have README with explanations. Play with them, modify them, break them. That's how you learn.
>
> Okay, that's 30 minutes of code. Questions?"

---

## 🎯 Key Talking Points

### Throughout Demo, Emphasize:

1. **Real Code, Not Pseudocode**
   > "This is actual working code you can run right now."

2. **No Magic**
   > "It's just Go. No code generation, no complex build steps."

3. **Clone and Explore**
   > "Everything I showed is in the repo. Clone it!"

4. **Progressive Complexity**
   > "Start simple (Example 01), build up to production patterns (Example 04)."

5. **Unique Value**
   > "Show me another framework that can do monolith → microservices with zero code change."

---

## 🚨 Common Questions During Demo

### Q: "How's the performance?"
**A**: 
> "ServeMux router - one of the fastest. 200-700ns per request. 10k+ req/s on single instance. Fast path handlers < 2μs. Production-ready."

### Q: "Can I use my existing code?"
**A**: 
> "Yes! Lokstra supports standard http.HandlerFunc. You can mix old and new code. Gradual migration."

### Q: "What about database?"
**A**: 
> "Any Go database library works. Examples use simple in-memory, but in production use sqlx, gorm, pgx, whatever you want."

### Q: "Testing?"
**A**: 
> "Services are just structs. Easy to mock. Examples include test files. Standard Go testing."

### Q: "Production ready?"
**A**: 
> "Yes! Already used in production. Active maintenance. Following semantic versioning."

---

## 🎬 Recovery Plans

### If Demo Breaks:

**Port already in use**:
```bash
# Kill process
lsof -ti:3000 | xargs kill -9
# Or use different port
go run . -addr :3010
```

**Code doesn't compile**:
```bash
# Have backup: pre-compiled binaries
go build -o demo01 ./01-hello-world
go build -o demo03 ./03-crud-api
# Just run binary
./demo01
```

**Network issues**:
> "Let me show you the code instead. You can see it still works, here's the output..." (show pre-recorded terminal session or screenshots)

**Time running short**:
- Skip Example 02 (handler forms)
- Skip Example 06 (bonus)
- Focus on Example 01 + 04 (killer demo)

---

## 📝 Post-Demo Actions

### Immediate (during Q&A):
```bash
# Share screen with:
1. GitHub repo URL
2. Documentation homepage
3. Quick Start guide
4. Examples folder

# Paste in chat:
Repo: https://github.com/primadi/lokstra
Docs: https://primadi.github.io/lokstra/
Examples: https://github.com/primadi/lokstra/tree/dev2/docs/00-introduction/examples
```

### After Session:
- Share recording (if available)
- Share slide deck
- Share demo commands (this script)
- Follow up with interested people

---

**Good luck with the demo! 🚀**

Remember: **Show, don't tell. Let the code speak.**
