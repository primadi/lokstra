# Architecture

> **Understanding Lokstra's design - how all the pieces fit together**

---

## 🎯 Overview

Lokstra is built on **6 core components** that work together to create a flexible, scalable REST API framework:

```
┌─────────────────────────────────────────────────┐
│                   SERVER                        │
│  (Container - Lifecycle Management)             │
│                                                 │
│  ┌───────────────────────────────────────────┐ │
│  │               APP                         │ │
│  │  (HTTP Listener - ServeMux/FastHTTP)      │ │
│  │                                           │ │
│  │  ┌─────────────────────────────────────┐ │ │
│  │  │           ROUTER                    │ │ │
│  │  │  (Route Management + Middleware)    │ │ │
│  │  │                                     │ │ │
│  │  │  Route 1 → [MW1, MW2] → Handler    │ │ │
│  │  │  Route 2 → [MW3] → Handler         │ │ │
│  │  │  Route 3 → Handler → Service       │ │ │
│  │  └─────────────────────────────────────┘ │ │
│  └───────────────────────────────────────────┘ │
└─────────────────────────────────────────────────┘

Supporting Components:
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│   SERVICE   │  │MIDDLEWARE   │  │CONFIGURATION│
│ (Business)  │  │ (Filters)   │  │  (Settings) │
└─────────────┘  └─────────────┘  └─────────────┘
```

Let's explore each component:

---

## 🏗️ Component 1: Server

**Purpose**: Container for one or more Apps, manages lifecycle

### Responsibilities
- ✅ Start/stop Apps
- ✅ Graceful shutdown
- ✅ Signal handling (SIGTERM, SIGINT)
- ✅ Configuration loading

### Key Point: Not in Request Flow!
```
❌ WRONG: Request → Server → App → Router
✅ RIGHT: Request → App → Router
                    ↑
                  Server manages lifecycle only
```

### Example Usage
```go
// server.go
type Server struct {
    Name         string
    DeploymentID string
    Apps         []*App
}

func main() {
    // Create server with multiple apps
    server := &Server{
        Name: "my-server",
        Apps: []*App{
            NewApp("api-v1", ":8080", apiV1Router),
            NewApp("api-v2", ":8081", apiV2Router),
            NewApp("admin", ":9000", adminRouter),
        },
    }
    
    // Run with graceful shutdown (30s timeout)
    server.Run(30 * time.Second)
}
```

**When server receives SIGTERM**:
1. Stop accepting new connections
2. Wait for active requests (max 30s)
3. Close all apps
4. Exit

📖 **Learn more**: [App & Server Guide](../01-essentials/05-app-and-server/README.md)

---

## 🌐 Component 2: App

**Purpose**: HTTP listener that serves a Router

### Responsibilities
- ✅ Listen on address (`:8080`)
- ✅ Accept HTTP connections
- ✅ Pass requests to Router
- ✅ Implement `http.Handler` or FastHTTP handler

### Two Engine Types

#### Engine 1: Go Standard (ServeMux)
```go
app := lokstra.NewApp("api", ":8080", router)
// Uses net/http standard library
```

#### Engine 2: FastHTTP (High Performance)
```go
app := lokstra.NewAppFastHTTP("api", ":8080", router)
// Uses valyala/fasthttp for speed
```

### Example
```go
// app.go
type App struct {
    Name   string
    Addr   string      // ":8080"
    Router *Router     // The router to serve
}

func (a *App) Run() error {
    // Standard Go HTTP server
    return http.ListenAndServe(a.Addr, a.Router)
}
```

**Request Flow Through App**:
```
TCP Connection → App.ServeHTTP() → Router.ServeHTTP()
                                      ↓
                                   Matching Route
                                      ↓
                                   Middleware Chain
                                      ↓
                                   Handler
```

📖 **Learn more**: [App & Server Guide](../01-essentials/05-app-and-server/README.md)

---

## 🚦 Component 3: Router

**Purpose**: Route registration, middleware management, request dispatch

### Responsibilities
- ✅ Register routes (`GET`, `POST`, etc.)
- ✅ Match incoming requests to routes
- ✅ Apply middleware chains
- ✅ Execute handlers
- ✅ Support route groups

### Key Features

#### 1. Route Registration
```go
r := lokstra.NewRouter("api")

// Simple routes
r.GET("/users", getUsers)
r.POST("/users", createUser)

// With path parameters
r.GET("/users/{id}", getUser)

// Route groups
api := r.Group("/api/v1")
api.GET("/products", getProducts)  // /api/v1/products
```

#### 2. Middleware Scopes
```go
r := lokstra.NewRouter("api")

// Global middleware (all routes)
r.Use(logging.Middleware())

// Group middleware
auth := r.Group("/admin")
auth.Use(authMiddleware)
auth.GET("/users", getUsers)      // Has logging + auth
auth.GET("/settings", getSettings) // Has logging + auth

// Route-specific middleware
r.GET("/public", publicHandler)  // Only has logging
```

#### 3. Implements http.Handler
```go
type Router struct {
    routes     []*Route
    middleware []lokstra_handler.MiddlewareFunc
}

// Standard interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    // 1. Match route
    route := r.Match(req.Method, req.URL.Path)
    
    // 2. Build middleware chain
    chain := r.middleware + route.middleware
    
    // 3. Execute
    ctx := request.NewContext(w, req)
    chain.Execute(ctx, route.handler)
}
```

### Routing Algorithm

```
Request: GET /api/users/123

Step 1: Match method
  ✅ GET routes only

Step 2: Match path pattern
  ❌ /api/products/{id}
  ✅ /api/users/{id}
  
Step 3: Extract params
  id = "123"
  
Step 4: Build context
  ctx.PathParams["id"] = "123"
  
Step 5: Execute middleware chain
  [logging] → [auth] → [handler]
```

📖 **Learn more**: [Router Guide](../01-essentials/01-router/README.md)

---

## 🔧 Component 4: Service

**Purpose**: Business logic layer with dependency injection

### Responsibilities
- ✅ Implement business logic
- ✅ Database operations
- ✅ External API calls
- ✅ Manage dependencies (lazy loading)

### Core Pattern: Lazy Loading

```go
type UserService struct {
    DB    *service.Lazy[*Database]
    Cache *service.Lazy[*CacheService]
    Email *service.Lazy[*EmailService]
}

// Registered in registry
lokstra_registry.RegisterServiceFactory("users", func() any {
    return &UserService{
        DB:    service.LazyLoad[*Database]("db"),
        Cache: service.LazyLoad[*CacheService]("cache"),
        Email: service.LazyLoad[*EmailService]("email"),
    }
})
```

**Lazy = Created only when first accessed**:
```go
func (s *UserService) CreateUser(p *CreateParams) (*User, error) {
    // DB created here (first call)
    db := s.DB.Get()
    
    user, err := db.Insert("INSERT INTO users ...")
    if err != nil {
        return nil, err
    }
    
    // Email created here (first call)
    s.Email.Get().SendWelcome(user.Email)
    
    return user, nil
}
```

### Why Lazy Loading?

**Problem - Circular Dependencies**:
```go
// Eager loading fails!
userService := &UserService{
    Orders: orderService,  // Not created yet!
}
orderService := &OrderService{
    Users: userService,  // Not created yet!
}
```

**Solution - Lazy Loading**:
```go
// Lazy loading works!
userService := &UserService{
    Orders: service.LazyLoad[*OrderService]("orders"),
}
orderService := &OrderService{
    Users: service.LazyLoad[*UserService]("users"),
}
// Both reference each other - resolved when .Get() is called
```

### Service Method Requirements

**MUST use struct parameters**:
```go
// ✅ RIGHT: Struct parameter
type GetByIDParams struct {
    ID int `path:"id"`
}

func (s *UserService) GetByID(p *GetByIDParams) (*User, error) {
    return s.DB.Get().QueryOne("SELECT * FROM users WHERE id = ?", p.ID)
}

// ❌ WRONG: Primitive parameter
func (s *UserService) GetByID(id int) (*User, error) {
    // Can't bind from path/query/body!
}
```

**Why?** Lokstra uses struct tags to bind request data:
- `path:"id"` - from URL path
- `query:"name"` - from query string
- `json:"email"` - from JSON body
- `header:"Authorization"` - from headers

📖 **Learn more**: [Service Guide](../01-essentials/02-service/README.md)

---

## 🔗 Component 5: Middleware

**Purpose**: Request/response filters, cross-cutting concerns

### Responsibilities
- ✅ Logging
- ✅ Authentication
- ✅ CORS
- ✅ Rate limiting
- ✅ Request validation
- ✅ Response transformation

### Middleware Pattern

```go
type MiddlewareFunc func(ctx *request.Context, next func() error) error

// Example: Logging middleware
func LoggingMiddleware() MiddlewareFunc {
    return func(ctx *request.Context, next func() error) error {
        start := time.Now()
        
        // Before handler
        log.Printf("→ %s %s", ctx.R.Method, ctx.R.URL.Path)
        
        // Execute next middleware/handler
        err := next()
        
        // After handler
        duration := time.Since(start)
        log.Printf("← %s %s (%v)", ctx.R.Method, ctx.R.URL.Path, duration)
        
        return err
    }
}
```

### Middleware Chain Execution

```
Request
  ↓
[Middleware 1] ─ before
  ↓
[Middleware 2] ─ before
  ↓
[Middleware 3] ─ before
  ↓
[Handler] ─ execute
  ↓
[Middleware 3] ─ after
  ↓
[Middleware 2] ─ after
  ↓
[Middleware 1] ─ after
  ↓
Response
```

### Example Flow

```go
// Setup
r := lokstra.NewRouter("api")
r.Use(loggingMiddleware, corsMiddleware)

auth := r.Group("/admin")
auth.Use(authMiddleware)
auth.GET("/users", getUsers)

// Request: GET /admin/users
// Execution:
logging.before()
  cors.before()
    auth.before()
      getUsers() // handler
    auth.after()
  cors.after()
logging.after()
```

### Two Usage Methods

#### Method 1: Direct Function
```go
r.Use(func(ctx *request.Context, next func() error) error {
    // middleware logic
    return next()
})
```

#### Method 2: By Name (Registry)
```go
// Register
lokstra_registry.RegisterMiddleware("auth", authMiddleware)

// Use by name
r.Use(middleware.ByName("auth"))
```

📖 **Learn more**: [Middleware Guide](../01-essentials/03-middleware/README.md)

---

## ⚙️ Component 6: Configuration

**Purpose**: Application settings and deployment management

### Responsibilities
- ✅ Load YAML config files
- ✅ Environment variable substitution
- ✅ Multi-deployment support
- ✅ Service/Router registration

### Configuration Structure

```yaml
# config.yaml
servers:
  - name: dev-server
    deployment-id: dev
    base-url: http://localhost:8080
    
    apps:
      - addr: ":8080"
        routers: [api]          # Which routers to include
        services: [db, users]   # Which services to load
        
    middleware:
      - name: logging
        type: builtin
      - name: cors
        type: builtin
        
    services:
      - name: db
        type: postgres
        config:
          host: ${DB_HOST:localhost}
          port: 5432
```

### Multi-Deployment Architecture

**Key concept**: Same code, different deployment configurations

```yaml
servers:
  # Monolith deployment
  - name: monolith
    deployment-id: dev
    apps:
      - addr: ":8080"
        services: [users, orders, products]
        
  # Microservices deployment
  - name: user-service
    deployment-id: prod
    base-url: http://user-service
    apps:
      - addr: ":8001"
        services: [users]
        
  - name: order-service
    deployment-id: prod
    base-url: http://order-service
    apps:
      - addr: ":8002"
        services: [orders]
```

**Critical fields**:
- `deployment-id`: Identifies deployment environment
- `base-url`: Base URL for service-to-service communication

**How it works**:
```go
// In OrderService
type OrderService struct {
    Users *service.Lazy[*UserService]
}

func (s *OrderService) CreateOrder(p *CreateParams) (*Order, error) {
    // Same deployment-id (dev):
    //   → Direct method call (in-memory)
    //
    // Different deployment-id (prod):
    //   → HTTP call to base-url
    user, err := s.Users.Get().GetByID(p.UserID)
}
```

📖 **Learn more**: [Configuration Guide](../01-essentials/04-configuration/README.md)

---

## 🔄 Complete Request Flow

Let's trace a request through all components:

### Example Setup
```go
// 1. Register services
lokstra_registry.RegisterServiceFactory("db", createDB)
lokstra_registry.RegisterServiceFactory("users", func() any {
    return &UserService{DB: service.LazyLoad[*Database]("db")}
})

// 2. Create router
r := lokstra.NewRouter("api")
r.Use(loggingMiddleware)

auth := r.Group("/admin")
auth.Use(authMiddleware)
auth.GET("/users/{id}", getUser)

// 3. Create app
app := lokstra.NewApp("api", ":8080", r)

// 4. Create server
server := &Server{Apps: []*App{app}}
server.Run(30 * time.Second)
```

### Request: `GET /admin/users/123`

```
Step 1: TCP Connection
  Client → App (port 8080)

Step 2: App receives request
  App.ServeHTTP(w, req)
    ↓
  Router.ServeHTTP(w, req)

Step 3: Router matches route
  Method: GET ✅
  Path: /admin/users/{id} ✅
  Extract params: {id: "123"}

Step 4: Build middleware chain
  Global: [loggingMiddleware]
  Group:  [authMiddleware]
  Route:  []
  Chain:  [logging, auth]

Step 5: Create context
  ctx := request.NewContext(w, req)
  ctx.PathParams["id"] = "123"

Step 6: Execute chain
  logging.before()
    → Log: "GET /admin/users/123"
    
  auth.before()
    → Check: Authorization header
    → Validate: JWT token
    
  handler.execute()
    → Call: getUser(ctx)
    → Extract: id from ctx.PathParams
    → Service: userService.GetByID(id)
    → DB: SELECT * FROM users WHERE id = 123
    → Response: user object
    
  auth.after()
    → (nothing)
    
  logging.after()
    → Log: "200 OK (45ms)"

Step 7: Write response
  HTTP/1.1 200 OK
  Content-Type: application/json
  
  {"id": 123, "name": "John", "email": "john@example.com"}
```

---

## 🏛️ Architecture Patterns

### Pattern 1: Layered Architecture

```
┌─────────────────────────────────────┐
│         Presentation Layer          │
│  (Router, Middleware, Handlers)     │
├─────────────────────────────────────┤
│          Business Layer             │
│         (Services)                  │
├─────────────────────────────────────┤
│          Data Layer                 │
│    (Database, Cache, APIs)          │
└─────────────────────────────────────┘
```

**Example**:
```go
// Presentation: Handler
func GetUserHandler(ctx *request.Context) (*User, error) {
    id := ctx.PathParam("id")
    return userService.GetByID(id)  // Call business layer
}

// Business: Service
func (s *UserService) GetByID(id string) (*User, error) {
    return s.DB.Get().QueryOne(...)  // Call data layer
}

// Data: Database
func (db *Database) QueryOne(query string) (*User, error) {
    // Execute SQL
}
```

### Pattern 2: Dependency Injection

```
Registry (Central)
   ↓
Services ←─── Lazy Load
   ↓
Handlers
```

**Example**:
```go
// Registry
lokstra_registry.RegisterServiceFactory("users", createUserService)

// Service with dependencies
type UserService struct {
    DB    *service.Lazy[*Database]
    Email *service.Lazy[*EmailService]
}

// Handler uses service
userService := lokstra_registry.GetService[*UserService]("users")
```

### Pattern 3: Convention over Configuration

**Example: Service as Router**

Instead of:
```go
// Configuration approach
r.GET("/users", listUsers)
r.GET("/users/{id}", getUser)
r.POST("/users", createUser)
r.PUT("/users/{id}", updateUser)
r.DELETE("/users/{id}", deleteUser)
```

Use:
```go
// Convention approach
router := router.NewFromService(userService, "/users")
// Auto-generates routes based on method names
```

---

## 🎯 Design Principles

### 1. Separation of Concerns
- **Router**: Routing only
- **Middleware**: Cross-cutting concerns
- **Handler**: Request/response
- **Service**: Business logic
- **Configuration**: Settings

### 2. Dependency Inversion
- High-level (handlers) depend on abstractions (services)
- Low-level (databases) implement abstractions
- Lazy loading for flexible resolution

### 3. Convention over Configuration
- Standard method names → Routes
- Struct tags → Parameter binding
- Sensible defaults

### 4. Flexibility
- 29 handler forms
- Multiple deployment modes
- Code or config-driven

### 5. Type Safety
- Generics for services
- Compile-time checks
- No reflection in hot path

---

## 📊 Component Interaction Diagram

```
┌──────────┐
│  Client  │
└────┬─────┘
     │ HTTP Request
     ↓
┌─────────────────────────────────────────────┐
│               SERVER                        │
│  (Lifecycle, Graceful Shutdown)             │
│                                             │
│  ┌───────────────────────────────────────┐ │
│  │              APP                      │ │
│  │  (HTTP Listener)                      │ │
│  │                                       │ │
│  │  ┌─────────────────────────────────┐ │ │
│  │  │         ROUTER                  │ │ │
│  │  │                                 │ │ │
│  │  │  Match Route                    │ │ │
│  │  │      ↓                          │ │ │
│  │  │  ┌────────────────────────┐    │ │ │
│  │  │  │   MIDDLEWARE CHAIN     │    │ │ │
│  │  │  │  [MW1] → [MW2] → [MW3] │    │ │ │
│  │  │  └──────────┬─────────────┘    │ │ │
│  │  │             ↓                   │ │ │
│  │  │  ┌────────────────────────┐    │ │ │
│  │  │  │      HANDLER           │    │ │ │
│  │  │  │  (Extract params)      │    │ │ │
│  │  │  └──────────┬─────────────┘    │ │ │
│  │  └─────────────┼──────────────────┘ │ │
│  └────────────────┼────────────────────┘ │
└───────────────────┼──────────────────────┘
                    ↓
        ┌────────────────────────┐
        │      SERVICE           │
        │  (Business Logic)      │
        │                        │
        │  ┌──────────────────┐  │
        │  │  Dependencies    │  │
        │  │  (Lazy Load)     │  │
        │  │                  │  │
        │  │  DB, Cache, etc  │  │
        │  └──────────────────┘  │
        └────────┬───────────────┘
                 ↓
     ┌────────────────────────┐
     │   External Resources   │
     │  (Database, APIs, etc) │
     └────────────────────────┘
```

---

## 💡 Key Takeaways

1. **Server**: Container, manages lifecycle, NOT in request flow
2. **App**: HTTP listener, serves router
3. **Router**: Route matching, middleware orchestration
4. **Middleware**: Request/response filters, cross-cutting concerns
5. **Service**: Business logic, lazy-loaded dependencies
6. **Configuration**: Settings, multi-deployment support

**Request Flow**:
```
App → Router → Middleware Chain → Handler → Service → Response
```

**Dependency Flow**:
```
Registry → Lazy Services → Handlers/Services → External Resources
```

---

## 📚 Learn More

**Next Steps**:
- [Essentials Guide](../01-essentials/README.md) - Hands-on tutorials
- [Deep Dive](../02-deep-dive/README.md) - Advanced patterns
- [API Reference](../03-api-reference/README.md) - Complete API docs

**Specific Components**:
- [Router](../01-essentials/01-router/README.md)
- [Service](../01-essentials/02-service/README.md)
- [Middleware](../01-essentials/03-middleware/README.md)
- [Configuration](../01-essentials/04-configuration/README.md)
- [App & Server](../01-essentials/05-app-and-server/README.md)

---

**Ready to start building?** 👉 [Quick Start](quick-start.md)
