# Architecture

> **Understanding Lokstra's design - how all the pieces fit together**

**📝 Note**: All code examples in this document are **runnable** - you can copy-paste and execute them directly. Examples are using the actual Lokstra API, not simplified pseudocode.

---

## 🎯 Overview

Lokstra is built on **6 core components** that work together to create a flexible, scalable REST API framework:

```
┌──────────────────────────────────────────────┐
│                   SERVER                     │
│  (Container - Lifecycle Management)          │
│                                              │
│  ┌─────────────────────────────────────────┐ │
│  │               APP                       │ │
│  │  (HTTP Listener - Standard net/http)    │ │
│  │                                         │ │
│  │  ┌────────────────────────────────────┐ │ │
│  │  │           ROUTER                   │ │ │
│  │  │  (Route Management + Middleware)   │ │ │
│  │  │                                    │ │ │
│  │  │  Route 1 → [MW1, MW2] → Handler    │ │ │
│  │  │  Route 2 → [MW3] → Handler         │ │ │
│  │  │  Route 3 → Handler → Service       │ │ │
│  │  └────────────────────────────────────┘ │ │
│  └─────────────────────────────────────────┘ │
└──────────────────────────────────────────────┘

Supporting Components:
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│   SERVICE   │  │ MIDDLEWARE  │  │CONFIGURATION│
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
import "github.com/primadi/lokstra"

func main() {
    // Create routers
    apiV1Router := lokstra.NewRouter("api-v1")
    apiV2Router := lokstra.NewRouter("api-v2")
    adminRouter := lokstra.NewRouter("admin")
    
    setupApiV1Router(apiV1Router)
    setupApiV2Router(apiV2Router)
    setupAdminRouter(adminRouter)
    
    // Create server with multiple apps
    server := lokstra.NewServer("my-server",
        lokstra.NewApp("api-v1", ":8080", apiV1Router),
        lokstra.NewApp("api-v2", ":8081", apiV2Router),
        lokstra.NewApp("admin", ":9000", adminRouter),
    )
    
    // Run with graceful shutdown (30s timeout)
    server.Run(30 * time.Second)
}
```

**When server receives SIGTERM/SIGINT**:
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

#### Engine 1: Go Standard (ServeMux) - Default
```go
// Default engine (ServeMux)
app := lokstra.NewApp("api", ":8080", router)
```

#### Engine 2: Custom Listener (Advanced)
```go
// Custom configuration
app := lokstra.NewAppWithConfig("api", ":8080", "fasthttp", 
    map[string]any{
        // custom config options
    }, 
    router,
)
```

**Note:** Lokstra currently uses Go's standard `net/http` with ServeMux. FastHTTP support can be added via custom listener configuration if needed.

### Example
```go
import "github.com/primadi/lokstra"

func main() {
    // Create router
    router := lokstra.NewRouter("api")
    router.GET("/ping", func() string { return "pong" })
    
    // Create app
    app := lokstra.NewApp("api", ":8080", router)
    
    // Create server and run
    server := lokstra.NewServer("my-server", app)
    server.Run(30 * time.Second)
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
r.Use(loggingMiddleware, corsMiddleware)

// Group middleware
auth := r.Group("/admin")
auth.Use(authMiddleware)
auth.GET("/users", getUsers)      // Has logging + cors + auth
auth.GET("/settings", getSettings) // Has logging + cors + auth

// Route-level middleware (single route only)
r.GET("/special", specialHandler, rateLimitMiddleware, cacheMiddleware)
// Has: logging + cors + rateLimit + cache

// No middleware (only global)
r.GET("/public", publicHandler)  // Only has logging + cors
```

**Middleware by name:**
```go
// Register middleware in registry
lokstra_registry.RegisterMiddleware("auth", authMiddleware)
lokstra_registry.RegisterMiddleware("cors", corsMiddleware)
lokstra_registry.RegisterMiddleware("logging", loggingMiddleware)

// Use by name (string reference)
r := lokstra.NewRouter("api")
r.Use("logging", "cors")

auth := r.Group("/admin")
auth.Use("auth")
auth.GET("/users", getUsers)

// Mix direct function and by-name
r.GET("/special", specialHandler, "rateLimit", cacheMiddleware)
```

#### 3. Implements http.Handler

**Conceptual Overview** (simplified for understanding):

```go
type Router struct {
    routes     []*Route
    middleware []request.HandlerFunc
}

// Standard http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    // Conceptually:
    // 1. Match route based on method and path
    // 2. Build middleware chain (global + route-specific)
    // 3. Create request context
    // 4. Execute chain with handler
    
    // Actual implementation uses optimized routing and handler adaptation
}
```

**Note:** This is pseudo-code for conceptual understanding. The actual implementation in `core/router/router_impl.go` is optimized and handles:
- Path parameter extraction
- Handler form adaptation (29 variations)
- Middleware chain building
- Context creation and execution
- Error handling and response writing

See `core/router/router_impl.go` for real implementation details.

### Routing Algorithm

```
Request: GET /api/users/123
```

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


📖 **Learn more**: [Router Guide](../01-essentials/01-router/README.md)

---
## 🔧 Component 4: Service

**Purpose**: Business logic layer with dependency injection and service abstraction

### Responsibilities
- ✅ Implement business logic
- ✅ Database operations
- ✅ External API calls
- ✅ Manage dependencies (lazy loading)
- ✅ Support local AND remote execution

### Three Service Types

Lokstra recognizes three distinct service patterns based on deployment needs:

```
┌─────────────────────────────────────────────────────────────┐
│                    SERVICE TYPES                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1️.  LOCAL ONLY (Infrastructure)                            │
│      • Never exposed via HTTP                               │
│      • Always loaded locally                                │
│      • Examples: db, cache, logger, queue                   │
│                                                             │
│  2️.  REMOTE ONLY (External APIs)                            │
│      • Third-party services                                 │
│      • Always accessed via HTTP                             │
│      • Examples: stripe, sendgrid, twilio                   │
│                                                             │
│  3️.  LOCAL + REMOTE (Business Logic)                        │
│      • Your business services                               │
│      • Can be local OR remote                               │
│      • Auto-published via HTTP when needed                  │
│      • Examples: user-service, order-service                │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Service Categories

Lokstra supports three distinct service patterns based on deployment needs:

#### 1. Local-Only Services (Infrastructure)
Services that never need HTTP exposure:

**Examples:**
- Database connections (`db-service`, `postgres-service`)
- Cache systems (`redis-service`, `memcached-service`)
- Logging (`logger-service`)
- Message queues (`rabbitmq-service`, `kafka-service`)
- File storage (local filesystem)

**Characteristics:**
- ✅ Loaded locally in process
- ❌ Never published as router
- ❌ No remote variant exists
- Used by other services via dependency injection

**Example configuration:**
```yaml
service-definitions:
  db-service:
    type: database-factory
  
  redis-service:
    type: redis-factory
  
  logger-service:
    type: logger-factory
  
  user-repository:
    type: user-repository-factory
    depends-on: [db-service]
  
  user-service:
    type: user-service-factory
    depends-on: [user-repository, redis-service, logger-service]

deployments:
  app:
    servers:
      api-server:
        published-services:
          - user-service
        # Framework resolves dependency chain:
        # user-service → [user-repository, redis-service, logger-service]
        # user-repository → [db-service]
        # All services created lazily when first accessed
```

**How it works:**
- Define services in `service-definitions` with `depends-on`
- Framework automatically resolves entire dependency chain
- Services created lazily (only when first accessed)
- No need to worry about load order - dependencies resolved on-demand
- Clean zero-config deployment!

#### 2. Remote-Only Services (External)
Third-party APIs or external systems wrapped as Lokstra service:

**Examples:**
- Payment gateways (`stripe-service`, `paypal-service`)
- Email providers (`sendgrid-service`, `mailgun-service`)
- SMS services (`twilio-service`)
- Cloud storage (`s3-service`, `gcs-service`)
- External APIs (`weather-api-service`, `maps-service`)

**Characteristics:**
- ❌ No local implementation (always HTTP)
- ✅ Uses `proxy.Service` (if follows convention) OR `proxy.Router` (custom endpoints)
- ✅ Configured via `external-service-definitions`
- Can override routes for non-standard APIs

```yaml
service-definitions:
  order-repository:
    type: order-repository-factory
    depends-on: [db-service]
  
  order-service:
    type: order-service-factory
    depends-on: [order-repository, payment-gateway]

external-service-definitions:
  payment-gateway:
    url: "https://payment-api.example.com"
    type: payment-service-remote-factory

deployments:
  app:
    servers:
      api-server:
        # Framework auto-detects external services from URL
        published-services:
          - order-service  # Depends on payment-gateway (external)
```

**How it works:**
- Define external service in `external-service-definitions` with URL
- Add to `depends-on` in `service-definitions`
- Framework automatically creates remote proxy
- Auto-resolves based on URL presence!

**Implementation with proxy.Service:**
```go
// Simple external service wrapper
type PaymentServiceRemote struct {
    proxyService *proxy.Service
}

func NewPaymentServiceRemote(proxyService *proxy.Service) *PaymentServiceRemote {
    return &PaymentServiceRemote{
        proxyService: proxyService,
    }
}

func (s *PaymentServiceRemote) CreatePayment(p *CreatePaymentParams) (*Payment, error) {
    return proxy.CallWithData[*Payment](s.proxyService, "CreatePayment", p)
}

func (s *PaymentServiceRemote) Refund(p *RefundParams) (*Refund, error) {
    return proxy.CallWithData[*Refund](s.proxyService, "Refund", p)
}

// Factory for remote service
func PaymentServiceRemoteFactory(deps map[string]any, config map[string]any) any {
    return NewPaymentServiceRemote(
        service.CastProxyService(config["remote"]),
    )
}

// Register service type with metadata
lokstra_registry.RegisterServiceType(
    "payment-service-remote-factory",
    nil, PaymentServiceRemoteFactory,
    deploy.WithResource("payment", "payments"),
    deploy.WithConvention("rest"),
    deploy.WithRouteOverride("Refund", "POST /payments/{id}/refund"),
)
```

📖 **See**: [Example 06 - External Services](./examples/06-external-services/) for complete demo

#### 3. Local + Remote Services (Business Logic)
Business services that can be deployed locally OR accessed remotely:

**Examples:**
- Business entities (`user-service`, `order-service`, `product-service`)
- Domain logic (`accounting-service`, `inventory-service`)
- Application services (`notification-service`, `report-service`)

**Characteristics:**
- ✅ Has local implementation (business logic + DB)
- ✅ Has remote implementation (proxy for microservices)
- ✅ Published as router when local (via `published-services`)
- ✅ Auto-generates HTTP endpoints from service methods
- ✅ Interface abstraction for deployment flexibility
- Follow REST/RPC convention

**How Auto-Publishing Works:**

When a business service is listed in `published-services`, the framework:

1. **Reads metadata** from service instance (if implements `ServiceMeta`)
2. **Auto-generates router** using convention (REST/RPC)
3. **Creates HTTP endpoints** for each public method
4. **Makes service accessible** remotely via HTTP

```yaml
service-definitions:
  user-service:
    type: user-service-factory
    depends-on: [user-repository]

deployments:
  microservice:
    servers:
      user-server:
        published-services:
          - user-service  # ← Framework auto-exposes UserService via HTTP
```

2. **Auto-generates router** using convention (REST/RPC)
3. **Creates HTTP endpoints** for each public method
4. **Makes service accessible** remotely via HTTP

**Example - User Service with REST convention:**
```go
// UserService methods:
type UserService struct {
    DB *service.Cached[*Database]
}

// Optional: Implement ServiceMeta for custom routing
func (s *UserService) GetResourceName() (string, string) {
    return "user", "users"
}

func (s *UserService) GetConventionName() string {
    return "rest"
}

func (s *UserService) GetRouteOverride() autogen.RouteOverride {
    return autogen.RouteOverride{
        Custom: map[string]autogen.Route{
            // Custom route for non-standard method
            "Activate": {Method: "POST", Path: "/users/{id}/activate"},
        },
    }
}

func (s *UserService) GetByID(p *GetByIDParams) (*User, error) {
    return s.DB.MustGet().QueryOne("SELECT * FROM users WHERE id = ?", p.ID)
}

// Usage in same deployment
userService := lokstra_registry.GetService[*UserService]("user-service")
user, err := userService.GetByID(&GetByIDParams{ID: 123})
// ✅ Direct method call - fast!
```

**Note:** Implementing `ServiceMeta` is optional for local services. If not implemented, the framework uses metadata from service registration.

**Remote Implementation:**
```go
// Simple remote service wrapper
type UserServiceRemote struct {
    proxyService *proxy.Service
}

// Constructor receives proxy.Service from framework
func NewUserServiceRemote(proxyService *proxy.Service) *UserServiceRemote {
    return &UserServiceRemote{
        proxyService: proxyService,
    }
}

// Method uses proxy.CallWithData for HTTP calls
func (s *UserServiceRemote) GetByID(p *GetByIDParams) (*User, error) {
    return proxy.CallWithData[*User](s.proxyService, "GetByID", p)
}

// Factory for remote service (framework calls this)
func UserServiceRemoteFactory(deps map[string]any, config map[string]any) any {
    return NewUserServiceRemote(
        service.CastProxyService(config["remote"]),
    )
}

// Metadata in RegisterServiceType (not in struct!)
lokstra_registry.RegisterServiceType(
    "user-service-factory",
    UserServiceFactory, UserServiceRemoteFactory,
    deploy.WithResource("user", "users"),
    deploy.WithConvention("rest"),
)

// Usage in different deployment
userRemote := lokstra_registry.GetService[*UserServiceRemote]("user-service-remote")
user, err := userRemote.GetByID(&GetByIDParams{ID: 123})
// ✅ HTTP call - transparent!
```

📖 **See**: [Example 04 - Multi-Deployment](../examples/04-multi-deployment/) for complete demo

---

### Proxy Patterns

Lokstra provides two proxy patterns for different remote access scenarios:

#### proxy.Service - Convention-Based Remote Services

**Use when:**
- ✅ Service follows REST/RPC convention
- ✅ Need auto-routing from method names
- ✅ Internal microservices or external APIs with standard patterns
- ✅ Consistent API patterns across methods

**Features:**
- Auto-generates URLs from convention + metadata
- `CallWithData[T]()` for type-safe calls
- Route override support for custom endpoints
- Metadata-driven (resource, plural, convention)
- Framework auto-injects `proxy.Service` via `config["remote"]`

**Example:**
```go
// Simple remote wrapper
type UserServiceRemote struct {
    proxyService *proxy.Service
}

func NewUserServiceRemote(proxyService *proxy.Service) *UserServiceRemote {
    return &UserServiceRemote{
        proxyService: proxyService,
    }
}

// Metadata from RegisterServiceType
lokstra_registry.RegisterServiceType(
    "user-service-factory",
    UserServiceFactory, UserServiceRemoteFactory,
    deploy.WithResource("user", "users"),
    deploy.WithConvention("rest"),
)

// Auto-generates: GET /users/{id}
func (s *UserServiceRemote) GetByID(p *GetByIDParams) (*User, error) {
    return proxy.CallWithData[*User](s.proxyService, "GetByID", p)
}
```

📖 **See**: [Example 04](../examples/04-multi-deployment/) and [Example 06](../examples/06-external-services/)

---

#### proxy.Router - Direct HTTP Calls

**Use when:**
- ✅ Quick access to external API without creating service
- ✅ Non-standard or legacy API endpoints
- ✅ One-off calls or simple integrations
- ✅ Prototype/testing external APIs
- ❌ Don't need auto-routing or convention

**Features:**
- Direct HTTP calls to any endpoint
- No metadata or convention required
- Flexible for any REST API
- Type-safe response parsing
- Good for quick integrations

**Example:**
```go
// Create router proxy
router := proxy.NewRouter("https://api.external.com")

// Direct calls - no service wrapper needed
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

// GET /users/123
user, err := proxy.CallRouter[*User](
    router,
    "GET",
    "/users/{id}",
    map[string]any{"id": 123},
)

// POST /users with body
newUser, err := proxy.CallRouter[*User](
    router,
    "POST",
    "/users",
    map[string]any{
        "name":  "John",
        "email": "john@example.com",
    },
)
```

**When to use proxy.Router vs proxy.Service:**

| Aspect | proxy.Router | proxy.Service |
|--------|-------------|--------------|
| **Setup** | Minimal (just URL) | Service + metadata |
| **Auto-routing** | ❌ Manual paths | ✅ Convention-based |
| **Type safety** | ✅ Response only | ✅ Request + Response |
| **Use case** | Quick/simple calls | Structured services |
| **Maintenance** | Low effort | Better for large APIs |
| **Best for** | External APIs, prototyping | Microservices, internal APIs |

**Example - When proxy.Router is better:**

```go
// Scenario: Quick weather API integration

// With proxy.Router (simple!)
weatherRouter := proxy.NewRouter("https://api.weather.com")
weather, err := proxy.CallRouter[*WeatherData](
    weatherRouter,
    "GET",
    "/forecast/{city}",
    map[string]any{"city": "Jakarta"},
)

// With proxy.Service (overkill!)
// Need: WeatherServiceRemote struct, metadata, factory, etc.
// Too much boilerplate for one-off call!
```

📖 **See**: [Example 07 - Remote Router](../examples/07-remote-router/) for complete demo

---

### Key Concept: Interface Abstraction

**Same interface, different implementation:**

```go
// Define interface
type IUserService interface {
    GetByID(p *GetByIDParams) (*User, error)
    List(p *ListParams) ([]*User, error)
}

// Local implements interface
type UserService struct { ... }
func (s *UserService) GetByID(...) (*User, error) { /* DB call */ }

// Remote implements interface  
type UserServiceRemote struct { ... }
func (s *UserServiceRemote) GetByID(...) (*User, error) { /* HTTP call */ }

// OrderService doesn't know which one!
type OrderService struct {
    Users *service.Cached[IUserService]  // Could be local OR remote!
}

func (s *OrderService) CreateOrder(p *CreateParams) (*Order, error) {
    // This works for BOTH local and remote!
    user, err := s.Users.MustGet().GetByID(&GetByIDParams{ID: p.UserID})
    // ...
}
```

### Published Services

**Services that are exposed via HTTP endpoints:**

```yaml
# config.yaml
service-definitions:
  user-service:
    type: user-service-factory
    depends-on: [user-repository]

deployments:
  microservice:
    servers:
      user-server:
        base-url: "http://localhost"
        addr: ":3004"
        published-services:
          - user-service  # ← Makes UserService available via HTTP
```

**What happens:**
1. **Auto-router generated** from service metadata
2. **Routes created** for each service method
3. **HTTP endpoints** available for remote calls

**Example:**
```
published-services: [user-service]

Auto-generates:
  GET    /users           → UserService.List()
  GET    /users/{id}      → UserService.GetByID()
  POST   /users           → UserService.Create()
  PUT    /users/{id}      → UserService.Update()
  DELETE /users/{id}      → UserService.Delete()
```

### Service Resolution (Auto-Discovery)

**Lokstra automatically resolves service locations:**

```yaml
service-definitions:
  user-service:
    type: user-service-factory
    depends-on: [user-repository]
  
  order-service:
    type: order-service-factory
    depends-on: [order-repository, user-service]

deployments:
  microservice:
    servers:
      user-server:
        base-url: "http://localhost"
        addr: ":3004"
        published-services: [user-service]
      
      order-server:
        base-url: "http://localhost"
        addr: ":3005"
        published-services: [order-service]
        # No manual service listing needed!
```

**How it works:**
1. `user-service` published at `http://localhost:3004`
2. `order-service` depends on `IUserService` (from service-definitions)
3. Lokstra **auto-discovers**: `user-service` → `http://localhost:3004`
4. Creates `user-service-remote` client automatically
5. Injects remote client into OrderService

**No manual URL configuration needed!** ✅

### Service Types Comparison

| Aspect | Local Service | Remote Service | Published Service |
|--------|--------------|----------------|-------------------|
| **Execution** | In-process | HTTP call | Exposes via HTTP |
| **Performance** | Fast (ns) | Slower (ms) | Serves HTTP requests |
| **Usage** | Same deployment | Different deployment | Makes service accessible |
| **Code** | Business logic | HTTP proxy | Business logic + router |
| **Suffix** | `-service` | `-service-remote` | `-service` |

### Core Pattern: Lazy Loading

```go
type UserService struct {
    DB    *service.Cached[*Database]
    Cache *service.Cached[*CacheService]
    Email *service.Cached[*EmailService]
}

// Registered in registry
lokstra_registry.RegisterServiceFactory("user-service", func() any {
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
    s.Email.MustGet().SendWelcome(user.Email)
    
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
    Orders: service.LazyLoad[*OrderService]("order-service"),
}
orderService := &OrderService{
    Users: service.LazyLoad[*UserService]("user-service"),
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
    return s.DB.MustGet().QueryOne("SELECT * FROM users WHERE id = ?", p.ID)
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

### Deployment Flexibility

**Same code, different topology:**

```yaml
service-definitions:
  user-service:
    type: user-service-factory
    depends-on: [user-repository]
  
  order-service:
    type: order-service-factory
    depends-on: [order-repository, user-service]

# Monolith: UserService is local
deployments:
  monolith:
    servers:
      api-server:
        base-url: "http://localhost"
        addr: ":3003"
        published-services:
          - user-service
          - order-service
        # Framework auto-loads all dependencies:
        # user-repository, order-repository, etc.
  
  # Microservices: UserService is remote  
  microservice:
    servers:
      user-server:
        base-url: "http://localhost"
        addr: ":3004"
        published-services: [user-service]
      
      order-server:
        base-url: "http://localhost"
        addr: ":3005"
        published-services: [order-service]
        # OrderService.Users → UserServiceRemote (HTTP)
        # Framework auto-detects from topology
```

**OrderService code (unchanged):**

```go
type OrderService struct {
    Users *service.Cached[IUserService]
}

func (s *OrderService) CreateOrder(p *CreateParams) (*Order, error) {
    // In monolith: direct method call
    // In microservice: HTTP call
    user, err := s.Users.MustGet().GetByID(&GetByIDParams{ID: p.UserID})
}
```

**Key benefit**: Deploy as monolith OR microservices **without code changes!**

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
lokstra_registry.RegisterMiddleware("logging", loggingMiddleware)

// Use by name (string)
r.Use("auth", "logging")

// Mix direct and by-name
r.Use(corsMiddleware, "auth")
```

📖 **Learn more**: [Middleware Guide](../01-essentials/03-middleware/README.md)

---

## ⚙️ Component 6: Configuration

**Purpose**: Application settings and deployment topology management

### Responsibilities
- ✅ Load YAML config files
- ✅ Service/Router/Middleware registration
- ✅ Multi-deployment topology
- ✅ Service auto-discovery and resolution

### Configuration Structure

```yaml
# config.yaml

# ========================================
# Service Definitions (Global)
# ========================================
service-definitions:
  user-repository:
    type: user-repository-factory
    depends-on: [db-service]
  
  user-service:
    type: user-service-factory
    depends-on: [user-repository]
  
  order-repository:
    type: order-repository-factory
    depends-on: [db-service]
  
  order-service:
    type: order-service-factory
    depends-on: [order-repository, user-service]
  
  db-service:
    type: database-factory

# ========================================
# Deployments (Topology)
# ========================================
deployments:
  # Monolith: All services in one process
  monolith:
    servers:
      api-server:
        base-url: "http://localhost"
        addr: ":3003"
        published-services:
          - user-service
          - order-service
        # Framework resolves dependency chains:
        # user-service → [user-repository] → [db-service]
        # order-service → [order-repository, user-service] → [db-service]
        # All created lazily when first accessed
  
  # Microservices: Each service in separate process
  microservice:
    servers:
      user-server:
        base-url: "http://localhost"
        addr: ":3004"
        published-services: [user-service]
        # Resolves: user-service → [user-repository] → [db-service]
      
      order-server:
        base-url: "http://localhost"
        addr: ":3005"
        published-services: [order-service]
        # Resolves: order-service → [order-repository, user-service-remote]
        # order-repository → [db-service]
        # Auto-detects: user-service from user-server
```

**Key improvements:**
- ✅ No `required-services` - framework resolves dependency chains automatically
- ✅ No `required-remote-services` - framework auto-detects remote services
- ✅ Lazy loading - services created only when first accessed
- ✅ Zero-config - just list `published-services`!

### Key Configuration Concepts

#### 1. Service Definitions (Global)
Define services once, use in multiple deployments:

```yaml
service-definitions:
  user-service:
    type: user-service-factory
    depends-on: [user-repository]
```

**What it does:**
- Registers service factory in global registry
- Declares dependencies
- Available to all deployments

#### 2. Deployments (Topology)
Define how services are distributed across servers:

```yaml
deployments:
  monolith:    # All-in-one
  microservice:  # Distributed
```

**Each deployment is independent topology**

#### 3. Zero-Config Service Resolution

**No manual service listing needed!**

The framework automatically:
- ✅ Loads all dependencies from `service-definitions`
- ✅ Detects remote services from published services in other servers
- ✅ Creates local or remote instances based on availability

```yaml
service-definitions:
  order-repository:
    type: order-repository-factory
    depends-on: [db-service]
  
  order-service:
    type: order-service-factory
    depends-on: [order-repository, user-service]
  
  user-service:
    type: user-service-factory
    depends-on: [user-repository]

deployments:
  microservice:
    servers:
      user-server:
        base-url: "http://localhost"
        addr: ":3004"
        published-services: [user-service]
      
      order-server:
        base-url: "http://localhost"
        addr: ":3005"
        published-services: [order-service]
        # Framework automatically:
        # - Loads order-repository (local)
        # - Detects user-service published on user-server
        # - Creates user-service-remote proxy
```

#### 4. Published Services
Services exposed via HTTP endpoints:

```yaml
servers:
  user-server:
    published-services:
      - user-service  # Creates HTTP endpoints automatically
```

**What happens:**
2. HTTP routes created for each method
3. Service becomes accessible remotely

#### 5. Service Auto-Discovery

**Framework automatically discovers service locations from topology!**

```yaml
service-definitions:
  user-service:
    type: user-service-factory
    depends-on: [user-repository]
  
  order-service:
    type: order-service-factory
    depends-on: [order-repository, user-service]

deployments:
  microservice:
    servers:
      user-server:
        base-url: "http://localhost"
        addr: ":3004"
        published-services: [user-service]
      
      order-server:
        base-url: "http://localhost"
        addr: ":3005"
        published-services: [order-service]
        # No manual service listing needed!
        # Framework auto-detects user-service from topology
```

**Lokstra automatically:**
1. Reads `order-service` dependencies from its `service-definitions`
2. Finds `user-service` published at `http://localhost:3004`
3. Auto-creates `user-service-remote` → `http://localhost:3004`
4. Injects remote client into OrderService

### Multi-Deployment Architecture

**Key concept**: Same code, different deployment configurations

**Example - User microservice:**

```bash
# Monolith deployment
go run . -server=monolith
# Loads: user-service (local) + order-service (local)

# User microservice deployment  
go run . -server=user-service
# Loads: user-service (local only)

# Order microservice deployment
go run . -server=order-service  
# Loads: order-service (local) + user-service-remote (HTTP)
```

**How it works:**
```go
// OrderService code (unchanged)
type OrderService struct {
    Users *service.Cached[IUserService]  // Interface!
}

func (s *OrderService) CreateOrder(p *CreateParams) (*Order, error) {
    // In monolith: direct method call
    // In microservice: HTTP call
    user, err := s.Users.MustGet().GetByID(&GetByIDParams{ID: p.UserID})
}
```

**Deployment determines implementation:**
- **Monolith**: `Users` → `UserService` (local)
- **Microservice**: `Users` → `UserServiceRemote` (HTTP)

### Configuration Loading

```go
// Load config
config, err := loader.LoadConfig("config.yaml")

// Build deployment topology
err = loader.LoadAndBuild([]string{"config.yaml"})

// Get topology for specific deployment
registry := deploy.Global()
topology := registry.GetDeploymentTopology("microservice")

// Build server from topology
server, err := registry.BuildServer("microservice", "order-server")
```

### External Service Definitions

For services outside your config (external APIs):

```yaml
external-service-definitions:
  payment-gateway-remote:
    url: "https://payment-api.example.com"
  
  email-service-remote:
    url: "https://email.example.com"
```

**Use case:** Third-party services not in your topology

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
    return s.DB.MustGet().QueryOne(...)  // Call data layer
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
    DB    *service.Cached[*Database]
    Email *service.Cached[*EmailService]
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
