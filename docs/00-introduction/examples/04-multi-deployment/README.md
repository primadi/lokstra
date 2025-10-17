# Example 4: Multi-Deployment

**Demonstrates**: One binary, three deployment modes - flexible server architecture

---

## 📌 About This Example

> **Note**: This example demonstrates the **manual approach** to service and router registration. It's designed to help you understand:
> - How service-to-router conversion works under the hood
> - How to manually create handlers from service methods
> - How proxy services work for cross-service communication
> - Manual service registration for different deployment modes

### What This Example Shows (Manual Approach):
- ✅ Manual handler creation from service methods
- ✅ Manual `proxy.Router` usage with `DoJSON()` 
- ✅ Manual service registration (`UserServiceImpl` vs `UserServiceRemote`)
- ✅ Manual router configuration per server

### Advanced Patterns (Coming in Later Chapters):
For production applications, Lokstra provides automated patterns:
- 🔄 **Auto service-to-router conversion**: `router.NewFromService()` with conventions
- 🔄 **Convention-based routing**: RESTful, RPC, and custom conventions
- 🔄 **Auto proxy services**: `proxy.Service` with same conventions as router
- 🔄 **Config-driven deployment**: YAML/code-based deployment configuration

These advanced patterns will be covered in **01-essentials** and **02-advanced** chapters.

**For now, focus on understanding the manual approach - it's the foundation!**

📖 **Want to see the evolution path?** Read [EVOLUTION.md](EVOLUTION.md) for detailed comparison of manual vs automated patterns.

---

## 🎯 Learning Objectives

This example shows Lokstra's powerful deployment flexibility:

1. **Single Binary**: One compiled binary can run as 3 different server types
2. **Service Interface Pattern**: Same interface, multiple implementations (local vs remote)
3. **Transparent Cross-Service Calls**: HTTP calls hidden behind service interface
4. **Deployment-Specific Registration**: Each server registers only what it needs
5. **Shared Handlers & Services**: Code reuse across all deployment modes

## 📐 Key Concepts

### Deployment vs Server

- **Deployment** = Complete infrastructure setup
  - Monolith deployment: 1 server running all services
  - Microservices deployment: 2+ servers (user-service + order-service)

- **Server** = Individual process with specific responsibilities
  - **Monolith server**: All services, all endpoints (port 3003)
  - **User-service server**: Only user service, user endpoints (port 3004)
  - **Order-service server**: Only order service, order endpoints (port 3005)

### Single Binary Approach

**One binary, three modes**:
```bash
# Same binary file
./app -server monolith       # Mode 1
./app -server user-service   # Mode 2
./app -server order-service  # Mode 3
```

Each mode registers different services and exposes different endpoints.

## 🏗️ Architecture

### Deployment 1: Monolith (1 Server)
```
┌────────────────────────────────────────┐
│   Monolith Server (Port 3003)          │
│                                        │
│  ┌──────────────────────────────────┐  │
│  │  UserServiceImpl (local)         │  │
│  │  - GetByID()                     │  │
│  │  - List()                        │  │
│  └──────────────────────────────────┘  │
│                ↑                       │
│  ┌──────────────────────────────────┐  │
│  │  OrderServiceImpl                │  │
│  │  - GetByID() → calls UserService │  │
│  │  - GetByUserID()                 │  │
│  └──────────────────────────────────┘  │
│                                        │
│  Direct method calls (fast)            │
│  Shared database                       │
└────────────────────────────────────────┘
```

### Deployment 2: Microservices (2 Servers)
```
┌───────────────────────┐         ┌─────────────────────────────┐
│  User-Service Server  │         │  Order-Service Server       │
│  (Port 3004)          │         │  (Port 3005)                │
│                       │         │                             │
│  ┌─────────────────┐  │  HTTP   │  ┌───────────────────────┐  │
│  │ UserServiceImpl │  │◄────────┤  │ UserServiceRemote     │  │
│  │ (local)         │  │         │  │ (proxy to :3004)      │  │
│  │ - GetByID()     │  │         │  └───────────────────────┘  │
│  │ - List()        │  │         │            ↑                │
│  └─────────────────┘  │         │  ┌───────────────────────┐  │
│                       │         │  │ OrderServiceImpl      │  │
│  Endpoints:           │         │  │ - GetByID()           │  │
│  • GET /users         │         │  │ - GetByUserID()       │  │
│  • GET /users/{id}    │         │  └───────────────────────┘  │
│                       │         │                             │
└───────────────────────┘         │  Endpoints:                 │
                                  │  • GET /orders/{id}         │
                                  │  • GET /users/{id}/orders   │
                                  └─────────────────────────────┘

Key: OrderService uses UserServiceRemote which makes HTTP calls to port 3004
```

## 📦 Project Structure

```
04-multi-deployment/
├── appservice/              # Service definitions (deployment-agnostic)
│   ├── database.go         # In-memory database with User & Order models
│   ├── user_service.go     # UserServiceImpl (local implementation)
│   ├── user_service_remote.go  # UserServiceRemote (HTTP proxy)
│   └── order_service.go    # OrderServiceImpl (uses UserService interface)
│
├── handlers.go             # HTTP handlers (shared across all deployments)
├── registration.go         # Service registration for each server mode
├── main.go                 # Server entry points (3 functions)
└── test.http               # Test requests for all deployment modes
```

### Key Insight: Separation of Concerns

- **`/appservice`**: Service logic (same for all deployments)
- **`handlers.go`**: HTTP layer (same for all deployments)
- **`registration.go`**: What differs between deployments
- **`main.go`**: Server configuration & routing

## 📚 Code Walkthrough

### 1. Service Interface Pattern

**appservice/user_service.go** - Interface + Local Implementation:
```go
// Interface (used by all)
type UserService interface {
    GetByID(p *GetUserParams) (*User, error)
    List(p *ListUsersParams) ([]*User, error)
}

// Local implementation (for monolith & user-service server)
type UserServiceImpl struct {
    DB *service.Cached[*Database]
}

func (s *UserServiceImpl) GetByID(p *GetUserParams) (*User, error) {
    return s.DB.MustGet().GetUser(p.ID)
}

func (s *UserServiceImpl) List(p *ListUsersParams) ([]*User, error) {
    return s.DB.MustGet().GetAllUsers()
}

func NewUserService() UserService {
    return &UserServiceImpl{
        DB: service.LazyLoad[*Database]("db"),
    }
}
```

**appservice/user_service_remote.go** - Remote Implementation (HTTP Proxy):
```go
// Remote implementation (for order-service server in microservices mode)
type UserServiceRemote struct {
    proxy *proxy.Router
}

func (u *UserServiceRemote) GetByID(p *GetUserParams) (*User, error) {
    var JsonWrapper struct {
        Status string `json:"status"`
        Data   *User  `json:"data"`
    }
    
    // Makes HTTP GET to user-service server
    err := u.proxy.DoJSON("GET", fmt.Sprintf("/users/%d", p.ID), nil, nil, &JsonWrapper)
    if err != nil {
        return nil, proxy.ParseRouterError(err)
    }
    return JsonWrapper.Data, nil
}

func NewUserServiceRemote() *UserServiceRemote {
    return &UserServiceRemote{
        proxy: proxy.NewRemoteRouter("http://localhost:3004"),
    }
}
```

**Key Benefit**: OrderService doesn't know if it's calling local or remote!

### 2. OrderService Uses Interface

**appservice/order_service.go**:
```go
type OrderService interface {
    GetByID(p *GetOrderParams) (*OrderWithUser, error)
    GetByUserID(p *GetUserOrdersParams) ([]*Order, error)
}

type OrderServiceImpl struct {
    DB    *service.Cached[*Database]
    Users *service.Cached[UserService]  // ← Interface, not concrete type!
}

func (s *OrderServiceImpl) GetByID(p *GetOrderParams) (*OrderWithUser, error) {
    order, err := s.DB.MustGet().GetOrder(p.ID)
    if err != nil {
        return nil, err
    }

    // Cross-service call - local or HTTP, doesn't matter!
    user, err := s.Users.MustGet().GetByID(&GetUserParams{ID: order.UserID})
    if err != nil {
        return nil, fmt.Errorf("order found but user not found: %v", err)
    }

    return &OrderWithUser{Order: order, User: user}, nil
}
```

**Magic**: `s.Users.MustGet()` returns `UserService` interface.
- In monolith: It's `UserServiceImpl` (local calls)
- In order-service server: It's `UserServiceRemote` (HTTP calls)

### 3. Shared Handlers

**handlers.go** - Same code for all deployments:
```go
var (
    userService  = service.LazyLoad[appservice.UserService]("users")
    orderService = service.LazyLoad[appservice.OrderService]("orders")
)

func listUsersHandler(ctx *request.Context) error {
    users, err := userService.MustGet().List(&appservice.ListUsersParams{})
    if err != nil {
        return ctx.Api.Error(500, "INTERNAL_ERROR", err.Error())
    }
    return ctx.Api.Ok(users)
}

func getOrderHandler(ctx *request.Context) error {
    var params appservice.GetOrderParams
    if err := ctx.Req.BindPath(&params); err != nil {
        return ctx.Api.BadRequest("INVALID_ID", "Invalid order ID")
    }

    orderWithUser, err := orderService.MustGet().GetByID(&params)
    if err != nil {
        return ctx.Api.Error(404, "NOT_FOUND", err.Error())
    }
    return ctx.Api.Ok(orderWithUser)
}
```

Handlers don't care about deployment mode - they just call services!

### 4. Deployment-Specific Registration

**registration.go** - This is where the magic happens:

**Monolith Server**:
```go
func registerMonolithServices() {
    // Register service factories
    lokstra_registry.RegisterServiceType("dbFactory", appservice.NewDatabase)
    lokstra_registry.RegisterServiceType("usersFactory", appservice.NewUserService)
    lokstra_registry.RegisterServiceType("ordersFactory", appservice.NewOrderService)

    // Register lazy services
    lokstra_registry.RegisterLazyService("db", "dbFactory", nil)
    lokstra_registry.RegisterLazyService("users", "usersFactory", nil)  // ← UserServiceImpl
    lokstra_registry.RegisterLazyService("orders", "ordersFactory", nil)
}
```

**User-Service Server**:
```go
func registerUserServices() {
    // Only user-related services
    lokstra_registry.RegisterServiceType("dbFactory", appservice.NewDatabase)
    lokstra_registry.RegisterServiceType("usersFactory", appservice.NewUserService)

    lokstra_registry.RegisterLazyService("db", "dbFactory", nil)
    lokstra_registry.RegisterLazyService("users", "usersFactory", nil)  // ← UserServiceImpl
    // No orders service!
}
```

**Order-Service Server**:
```go
func registerOrderServices() {
    lokstra_registry.RegisterServiceType("dbFactory", appservice.NewDatabase)
    lokstra_registry.RegisterServiceType("ordersFactory", appservice.NewOrderService)
    
    // Remote user service - makes HTTP calls!
    lokstra_registry.RegisterServiceTypeRemote("usersFactory",
        appservice.NewUserServiceRemote)  // ← UserServiceRemote!

    lokstra_registry.RegisterLazyService("db", "dbFactory", nil)
    lokstra_registry.RegisterLazyService("orders", "ordersFactory", nil)
    lokstra_registry.RegisterLazyService("users", "usersFactory", nil)  // ← UserServiceRemote!
}
```

**Critical Difference**: `users` service:
- Monolith & user-service: `UserServiceImpl` (local)
- Order-service: `UserServiceRemote` (HTTP proxy)

### 5. Server Entry Points

**main.go**:
```go
func main() {
    server := flag.String("server", "monolith", "Server to run")
    
    switch *server {
    case "monolith":
        runMonolithServer()
    case "user-service":
        runUserServiceServer()
    case "order-service":
        runOrderServiceServer()
    }
}
```

Each function:
1. Calls appropriate registration function
2. Creates router with specific endpoints
3. Runs server on designated port

## 🚀 Running the Examples

### Option 1: Monolith Deployment (1 Server)

Run everything in one server process:
```powershell
go run . -server monolith
```

Access all endpoints on **port 3003**:
```http
GET http://localhost:3003/users
GET http://localhost:3003/users/1
GET http://localhost:3003/orders/1
GET http://localhost:3003/users/1/orders
```

**What's registered**:
- ✅ Database
- ✅ UserServiceImpl (local)
- ✅ OrderServiceImpl (local)

### Option 2: Microservices Deployment (2 Servers)

**Terminal 1** - Start User Service Server:
```powershell
go run . -server user-service
```

**Terminal 2** - Start Order Service Server:
```powershell
go run . -server order-service
```

Access services on different ports:
```http
# User Service Server (port 3004)
GET http://localhost:3004/users
GET http://localhost:3004/users/1

# Order Service Server (port 3005)
GET http://localhost:3005/orders/1
GET http://localhost:3005/users/1/orders
```

**What's registered in user-service**:
- ✅ Database
- ✅ UserServiceImpl (local)

**What's registered in order-service**:
- ✅ Database
- ✅ OrderServiceImpl (local)
- ✅ UserServiceRemote (HTTP proxy to localhost:3004)

## 🧪 Testing with test.http

The included `test.http` file has comprehensive tests for all deployment options. Open it in VS Code with REST Client extension.

## 🔍 Key Features Demonstrated

### 1. **Single Binary, Multiple Deployment Modes**

One compiled binary can run as 3 different servers:
```bash
# Build once
go build .

# Run in 3 different modes
./04-multi-deployment -server monolith
./04-multi-deployment -server user-service
./04-multi-deployment -server order-service
```

### 2. **Interface-Based Service Abstraction**

```go
type UserService interface {
    GetByID(p *GetUserParams) (*User, error)
    List(p *ListUsersParams) ([]*User, error)
}

// Implementation 1: Local (direct DB calls)
type UserServiceImpl struct { ... }

// Implementation 2: Remote (HTTP proxy)
type UserServiceRemote struct { ... }
```

Consumer code (OrderService, handlers) uses the interface - doesn't know which!

### 3. **Transparent Cross-Service Communication**

OrderService code:
```go
user, err := s.Users.MustGet().GetByID(&GetUserParams{ID: order.UserID})
```

Behavior:
- **Monolith**: Direct method call to `UserServiceImpl.GetByID()`
- **Microservices**: HTTP GET to `http://localhost:3004/users/{id}` via `UserServiceRemote`

Same code, different runtime behavior!

### 4. **Deployment-Specific Service Registration**

The **only** difference between deployments is what gets registered:

| Server | Database | UserService | OrderService |
|--------|----------|-------------|--------------|
| Monolith | Local | `UserServiceImpl` (local) | `OrderServiceImpl` (local) |
| User-service | Local | `UserServiceImpl` (local) | ❌ Not registered |
| Order-service | Local | `UserServiceRemote` (HTTP) | `OrderServiceImpl` (local) |

### 5. **Shared Code Across Deployments**

**What's shared** (100% reuse):
- ✅ All service interfaces
- ✅ All service implementations
- ✅ All handlers
- ✅ All models

**What's different**:
- ❌ Service registration
- ❌ Router configuration
- ❌ Port numbers

## 📊 Response Examples

### Monolith Server Info
```bash
GET http://localhost:3003/
```

Response:
```json
{
  "code": 200,
  "status": "success",
  "message": "OK",
  "data": {
    "server": "monolith",
    "message": "All services running in one process",
    "endpoints": {
      "users": ["GET /users", "GET /users/{id}"],
      "orders": ["GET /orders/{id}", "GET /users/{user_id}/orders"]
    }
  }
}
```

### Get Order with User (Cross-Service Call)

**Monolith** - Direct method call:
```bash
GET http://localhost:3003/orders/1
```

**Order-Service (Microservices)** - HTTP call to user-service:
```bash
GET http://localhost:3005/orders/1
```

Both return identical response:
```json
{
  "code": 200,
  "status": "success",
  "message": "OK",
  "data": {
    "order": {
      "id": 1,
      "user_id": 1,
      "product": "Laptop",
      "amount": 1200
    },
    "user": {
      "id": 1,
      "name": "Alice",
      "email": "alice@example.com"
    }
  }
}
```

**Behind the scenes**:
- **Monolith**: `OrderServiceImpl` → `UserServiceImpl` (direct call)
- **Microservices**: `OrderServiceImpl` → `UserServiceRemote` → HTTP GET `/users/1` → `UserServiceImpl`

## 🎓 Advanced Patterns

### 1. **proxy.Router for HTTP Communication**

`UserServiceRemote` uses `proxy.Router` to make HTTP calls:

```go
type UserServiceRemote struct {
    proxy *proxy.Router
}

func NewUserServiceRemote() *UserServiceRemote {
    return &UserServiceRemote{
        proxy: proxy.NewRemoteRouter("http://localhost:3004"),
    }
}

func (u *UserServiceRemote) GetByID(p *GetUserParams) (*User, error) {
    var JsonWrapper struct {
        Status string `json:"status"`
        Data   *User  `json:"data"`
    }
    
    // Makes HTTP GET to http://localhost:3004/users/{id}
    err := u.proxy.DoJSON("GET", fmt.Sprintf("/users/%d", p.ID), nil, nil, &JsonWrapper)
    if err != nil {
        return nil, proxy.ParseRouterError(err)
    }
    return JsonWrapper.Data, nil
}
```

**Benefits**:
- Automatic JSON marshaling/unmarshaling
- Error handling with `proxy.ParseRouterError()`
- Consistent interface with local implementation

### 2. **Service Interface Contract**

Both implementations satisfy the same interface:

```go
var _ UserService = (*UserServiceImpl)(nil)   // Compile-time check
var _ UserService = (*UserServiceRemote)(nil) // Compile-time check
```

This ensures:
- Both have identical methods
- Can be swapped at runtime
- Type safety guaranteed

### 3. **Lazy Service Resolution**

Handlers use `service.LazyLoad`:

```go
var userService = service.LazyLoad[appservice.UserService]("users")

func getUserHandler(ctx *request.Context) error {
    // Resolved at first call based on registration
    user, err := userService.MustGet().GetByID(&params)
    ...
}
```

**Benefits**:
- Resolved based on what was registered
- No code changes in handlers
- Type-safe service access

### 4. **RegisterServiceTypeRemote**

Special registration for remote services:

```go
lokstra_registry.RegisterServiceTypeRemote("usersFactory",
    appservice.NewUserServiceRemote)
```

This tells Lokstra:
- Create instance using `NewUserServiceRemote()`
- Service will make HTTP calls
- Different from local factory registration

## 💡 Design Principles

### 1. **Separation of Concerns**

| Layer | Responsibility | Deployment Dependency |
|-------|----------------|----------------------|
| **Models** (`User`, `Order`) | Data structures | ❌ None |
| **Service Interfaces** | Contracts | ❌ None |
| **Service Implementations** | Business logic | ❌ None (both local & remote) |
| **Handlers** | HTTP layer | ❌ None |
| **Registration** | Wiring | ✅ **YES** - Only this changes! |

### 2. **Dependency Inversion**

```
OrderService depends on UserService interface (abstraction)
         ↓
Not on UserServiceImpl or UserServiceRemote (concrete)
```

This allows swapping implementations at runtime.

### 3. **Interface Segregation**

Each service interface is minimal:
- `UserService`: Only user operations
- `OrderService`: Only order operations

No bloated interfaces with unused methods.

### 4. **Single Responsibility**

Each file has one job:
- `database.go`: Data storage
- `user_service.go`: Local user operations
- `user_service_remote.go`: Remote user operations  
- `order_service.go`: Order operations + user coordination
- `handlers.go`: HTTP request/response
- `registration.go`: Service wiring
- `main.go`: Server configuration

## 🚀 Production Considerations

### 1. **Configuration Management**

Currently uses hardcoded values. In production:

```go
// Use environment variables
func NewUserServiceRemote() *UserServiceRemote {
    baseURL := os.Getenv("USER_SERVICE_URL")
    if baseURL == "" {
        baseURL = "http://localhost:3004"
    }
    return &UserServiceRemote{
        proxy: proxy.NewRemoteRouter(baseURL),
    }
}
```

Or use Lokstra's unified config:
```yaml
services:
  user-service:
    url: http://user-service:3004
    timeout: 5s
    retry: 3
```

### 2. **Service Discovery**

Integrate with:
- **Kubernetes**: Service DNS (e.g., `http://user-service.default.svc.cluster.local`)
- **Consul**: Dynamic service discovery
- **Eureka**: Netflix service registry

### 3. **Resilience Patterns**

Add to `UserServiceRemote`:
- **Circuit breaker**: Stop calling failing services
- **Retries**: Retry failed requests with backoff
- **Timeouts**: Don't wait forever
- **Fallbacks**: Return cached data or defaults

```go
func (u *UserServiceRemote) GetByID(p *GetUserParams) (*User, error) {
    // Add circuit breaker, retry, timeout logic
    return u.circuitBreaker.Execute(func() (*User, error) {
        return u.doGetByID(p)
    })
}
```

### 4. **Monitoring & Observability**

Add:
- Request tracing (OpenTelemetry)
- Metrics (Prometheus)
- Logging (structured logs)

```go
func (u *UserServiceRemote) GetByID(p *GetUserParams) (*User, error) {
    span := trace.Start("UserService.GetByID")
    defer span.End()
    
    log.Info("Fetching user", "id", p.ID)
    // ... existing code
}
```

### 5. **Database Strategy**

**Development** (current):
- Shared in-memory database
- Simple, fast

**Production**:
- Each server has own database
- Data consistency via events or distributed transactions
- User-service: PostgreSQL
- Order-service: PostgreSQL + cache

### 6. **API Versioning**

When services evolve independently:

```go
proxy: proxy.NewRemoteRouter("http://user-service:3004/v1")
```

Allows:
- User-service to release v2 without breaking order-service
- Gradual migration
- Backward compatibility

## 💡 When to Use Each Deployment

### Monolith Deployment
✅ **Good for**:
- Development & testing
- Small to medium apps
- Simple operations
- Low latency requirements
- Cost-sensitive projects
- Single team

❌ **Avoid when**:
- Need independent scaling
- Multiple teams on different services
- Services have different resource needs

### Microservices Deployment
✅ **Good for**:
- Large applications
- Independent team ownership
- Different scaling per service
- Polyglot requirements
- Fault isolation
- Independent deployment cycles

❌ **Avoid when**:
- Small team/app
- High inter-service chattiness
- Limited ops experience
- Complexity not justified

## 🔗 Related Topics

- **Example 3 (CRUD API)**: Service layer patterns
- **Essentials / Services**: Deep dive into service registration & lazy loading
- **Essentials / Proxy Router**: HTTP client for inter-service communication
- **Configuration Guide**: Unified config for deployment settings
- **Production Guide**: Scaling, monitoring, and deployment strategies

## 📚 What You Learned

1. ✅ **Single binary, multiple deployment modes** - One build, three run options
2. ✅ **Interface-based abstraction** - UserService interface with local & remote implementations
3. ✅ **Transparent cross-service calls** - Same code works locally or via HTTP
4. ✅ **Deployment-specific registration** - Only registration changes, not business logic
5. ✅ **Code reuse** - Handlers, services, models shared across all deployments
6. ✅ **proxy.Router pattern** - Clean HTTP communication wrapper
7. ✅ **Design principles** - Separation of concerns, dependency inversion, interface segregation
8. ✅ **Production considerations** - Config, service discovery, resilience, monitoring

## 🎯 Key Takeaways

### Manual Approach for Learning

This example uses the **manual approach** intentionally to teach fundamentals:

**What you learned (Manual)**:
- ✅ How to create handlers from service methods manually
- ✅ How `proxy.Router` works with `DoJSON()` calls
- ✅ How to register different service implementations per deployment
- ✅ How interface abstraction enables transparent local/remote calls

**What's coming (Automated)**:
- 🔄 Auto service-to-router with `router.NewFromService()`
- 🔄 Convention-based routing (RESTful, RPC, custom)
- 🔄 Auto proxy with `proxy.Service` using same conventions
- 🔄 Config-driven deployment (YAML/code)

### Why Learn Manual First?

Understanding the manual approach helps you:
1. **Debug issues**: Know what's happening under the hood
2. **Customize behavior**: Override automated behavior when needed
3. **Appreciate automation**: Understand what the framework does for you
4. **Make informed decisions**: Choose manual vs automated wisely

### This Example Does NOT Use Unified Config

This example demonstrates deployment flexibility **without** Lokstra's unified config system. Everything is hardcoded for clarity:
- Port numbers in `main.go`
- Service URLs in `NewUserServiceRemote()`
- Flag-based server selection

**Next Level**: Later chapters will show:
- Unified config system
- Convention-based routing
- Automated service registration
- Config-driven deployment

### The Power of Interfaces

The magic is in this line:
```go
Users *service.Cached[UserService]  // Interface, not concrete type!
```

This single design choice enables:
- ✅ Swapping implementations at runtime
- ✅ Testing with mocks
- ✅ Deployment flexibility
- ✅ Zero code changes in consumers

### One Binary = Deployment Flexibility

Traditional approach:
```bash
user-service/      # Separate project
order-service/     # Separate project
shared-lib/        # Shared code (versioning nightmare)
```

Lokstra approach:
```bash
app/               # One project
  -server monolith       # Run option 1
  -server user-service   # Run option 2
  -server order-service  # Run option 3
```

Benefits:
- No version skew between services
- Shared code without libraries
- Easy refactoring across services
- Type-safe cross-service calls

## 🎯 Next Steps

### Within This Example (Manual Approach):
1. **Add More Services**: Create `PaymentService`, `ShippingService` manually
2. **Implement Caching**: Add Redis to UserServiceRemote
3. **Add Tests**: Unit test with mock UserService
4. **Add Metrics**: Track HTTP calls in UserServiceRemote
5. **Add Circuit Breaker**: Resilience patterns in remote calls

### Evolution to Advanced Patterns:
Continue your learning journey with these chapters:

**01-Essentials** (Recommended Next):
- 📚 **Convention-Based Routing**: Auto service-to-router conversion
- 📚 **Proxy Services**: `proxy.Service` with automatic method mapping
- 📚 **Service Registry Patterns**: Advanced registration strategies

**02-Advanced**:
- 📚 **Config-Driven Deployment**: YAML/code-based deployment configuration
- 📚 **Custom Conventions**: Create your own routing conventions
- 📚 **Multi-Environment Setup**: Dev, staging, production configs

**03-Production**:
- 📚 **Service Discovery**: Integration with Consul, Kubernetes
- 📚 **Observability**: Metrics, tracing, logging
- 📚 **Resilience Patterns**: Circuit breakers, retries, timeouts
3. **Add Config**: Use unified config for ports & URLs
4. **Add Metrics**: Track HTTP calls in UserServiceRemote
5. **Add Tests**: Unit test with mock UserService
6. **Add Circuit Breaker**: Resilience patterns in remote calls
7. **Try Kubernetes**: Deploy all three modes to K8s

Happy coding! 🚀
