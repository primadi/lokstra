# Example 05: Auto-Router & Proxy Service# Example 05: Auto-Router & Proxy Service with Service Discovery# Example 4: Multi-Deployment



> **Demonstrates**: Automatic router generation from services and cross-service communication using proxy



---This example demonstrates advanced features of Lokstra's configuration-driven architecture:**Demonstrates**: One binary, three deployment modes - flexible server architecture



## 🎯 What This Example Shows



This example demonstrates two powerful Lokstra features working together:1. **Auto-Router Generation** - Services automatically generate REST routers---



1. **Auto-Router** - Automatically generate REST routes from a service2. **Proxy Service** - Remote services without hardcoded URLs

2. **Proxy Service** - Call remote services transparently using the same interface

3. **Service Discovery** - Automatic URL resolution via `published-router-services`## 📌 About This Example

### Key Concepts

4. **Microservices Deployment** - Multiple servers in one deployment configuration

- ✅ `autogen.NewFromService()` - Generate router from service methods

- ✅ `proxy.Service` - HTTP client with convention-based routing> **Note**: This example demonstrates the **manual approach** to service and router registration. It's designed to help you understand:

- ✅ `proxy.Call()` - Forward method calls to remote service

- ✅ Service interface pattern - Same code, local or remote implementation## Key Concepts> - How service-to-router conversion works under the hood



---> - How to manually create handlers from service methods



## 🏗️ Architecture### 1. Auto-Router Configuration> - How proxy services work for cross-service communication



```> - Manual service registration for different deployment modes

┌─────────────────────────────────┐         ┌──────────────────────────────────┐

│   USER SERVICE (Server)         │         │   ORDER SERVICE (Client)         │Services can automatically generate routers using the `auto-router` configuration:

│   Port: 3000                    │         │   Port: 3002                     │

│                                 │         │                                  │### What This Example Shows (Manual Approach):

│  ┌───────────────────────────┐  │         │  ┌────────────────────────────┐  │

│  │  UserServiceImpl          │  │         │  │  OrderService              │  │```yaml- ✅ Manual handler creation from service methods

│  │  - List()                 │  │         │  │  - GetOrder()              │  │

│  │  - Get(id)                │  │         │  │  - GetUserOrders()         │  │service-definitions:- ✅ Manual `proxy.Router` usage with `DoJSON()` 

│  │  - Create()               │  │         │  │    ↓ calls UserService    │  │

│  │  - Update(id)             │  │         │  └────────────────────────────┘  │  user-service:- ✅ Manual service registration (`UserServiceImpl` vs `UserServiceRemote`)

│  └───────────────────────────┘  │         │             ↓                    │

│             ↓                    │         │  ┌────────────────────────────┐  │    type: user-service-factory- ✅ Manual router configuration per server

│  ┌───────────────────────────┐  │  HTTP   │  │  UserServiceRemote         │  │

│  │  AUTO-ROUTER              │  │◄────────┤  │  - List() → proxy.Call()   │  │    depends-on:

│  │  (autogen.NewFromService) │  │         │  │  - Get() → proxy.Call()    │  │

│  │                           │  │         │  └────────────────────────────┘  │      - database### Advanced Patterns (Coming in Later Chapters):

│  │  REST Convention:         │  │         │             ↓                    │

│  │  GET  /api/v1/users       │  │         │  ┌────────────────────────────┐  │    For production applications, Lokstra provides automated patterns:

│  │  GET  /api/v1/users/{id}  │  │         │  │  proxy.Service             │  │

│  │  POST /api/v1/users       │  │         │  │  http://localhost:3000     │  │    auto-router:- 🔄 **Auto service-to-router conversion**: `router.NewFromService()` with conventions

│  │  PUT  /api/v1/users/{id}  │  │         │  │  + REST Convention         │  │

│  └───────────────────────────┘  │         │  └────────────────────────────┘  │      convention: rest           # REST, RPC, or GraphQL- 🔄 **Convention-based routing**: RESTful, RPC, and custom conventions

└─────────────────────────────────┘         └──────────────────────────────────┘

```      resource: "user"           # Singular resource name- 🔄 **Auto proxy services**: `proxy.Service` with same conventions as router



**Flow:**      resource-plural: "users"   # Plural for collection endpoints- 🔄 **Config-driven deployment**: YAML/code-based deployment configuration

1. User Service: Service methods → Auto-Router → REST endpoints

2. Order Service: Method call → UserServiceRemote → proxy.Call() → HTTP request → User Service      path-prefix: "/api/v1"     # Base path for all routes



---      middlewares: []            # Router-level middlewaresThese advanced patterns will be covered in **01-essentials** and **02-advanced** chapters.



## 📦 Project Structure      hidden: []                 # Methods to hide



``````**For now, focus on understanding the manual approach - it's the foundation!**

05-auto-router-proxy/

├── main.go           # Complete example (server + client)

├── test.http         # HTTP tests

└── README.md         # This fileThis generates routes automatically:📖 **Want to see the evolution path?** Read [EVOLUTION.md](EVOLUTION.md) for detailed comparison of manual vs automated patterns.

```

- `GET /api/v1/users` - List all users

---

- `GET /api/v1/users/{id}` - Get user by ID---

## 🚀 Running the Example

- `POST /api/v1/users` - Create user

### Step 1: Start User Service (Server with Auto-Router)

- `PUT /api/v1/users/{id}` - Update user## 🎯 Learning Objectives

**Terminal 1:**

```bash- `DELETE /api/v1/users/{id}` - Delete user

cd docs/00-introduction/examples/05-auto-router-proxy

go run . -mode=serverThis example shows Lokstra's powerful deployment flexibility:

```

### 2. Proxy Service (No Hardcoded URLs!)

Output:

```1. **Single Binary**: One compiled binary can run as 3 different server types

🚀 Starting USER SERVICE (Auto-Router Server)

🌐 Server starting on :3000Remote services reference the original service definition:2. **Service Interface Pattern**: Same interface, multiple implementations (local vs remote)

```

3. **Transparent Cross-Service Calls**: HTTP calls hidden behind service interface

**What happens:**

- Creates `UserServiceImpl` instance```yaml4. **Deployment-Specific Registration**: Each server registers only what it needs

- Defines REST convention: resource="user", plural="users"

- Calls `autogen.NewFromService()` to generate routerremote-service-definitions:5. **Shared Handlers & Services**: Code reuse across all deployment modes

- Routes automatically created:

  - `GET /api/v1/users` → `UserServiceImpl.List()`  user-service-remote:

  - `GET /api/v1/users/{id}` → `UserServiceImpl.Get()`

  - `POST /api/v1/users` → `UserServiceImpl.Create()`    proxy-service: user-service  # Reference to service definition## 📐 Key Concepts

  - `PUT /api/v1/users/{id}` → `UserServiceImpl.Update()`

    # No URL needed! It will be resolved from published-router-services

### Step 2: Start Order Service (Client with Proxy)

```### Deployment vs Server

**Terminal 2:**

```bash

cd docs/00-introduction/examples/05-auto-router-proxy

go run . -mode=client**Benefits:**- **Deployment** = Complete infrastructure setup

```

- ✅ DRY - Single source of truth for routing configuration  - Monolith deployment: 1 server running all services

Output:

```- ✅ Environment-agnostic - Same config works in dev/staging/prod  - Microservices deployment: 2+ servers (user-service + order-service)

🚀 Starting ORDER SERVICE (Proxy Client)

🌐 Server starting on :3002- ✅ Automatic convention inheritance - Same REST patterns

```

- **Server** = Individual process with specific responsibilities

**What happens:**

- Creates `proxy.Service` pointing to http://localhost:3000### 3. Service Discovery via Published Router Services  - **Monolith server**: All services, all endpoints (port 3003)

- Uses same REST convention as user service

- Creates `UserServiceRemote` that uses the proxy  - **User-service server**: Only user service, user endpoints (port 3004)

- OrderService calls UserServiceRemote.List() which becomes HTTP GET /api/v1/users

Services publish their availability for other services to discover:  - **Order-service server**: Only order service, order endpoints (port 3005)

### Step 3: Test the Services



**Test User Service directly:**

```bash```yaml### Single Binary Approach

curl http://localhost:3000/api/v1/users

```microservices:



**Test Order Service calling User Service via proxy:**  servers:**One binary, three modes**:

```bash

curl http://localhost:3002/users/1/orders    user-api:```bash

```

      base-url: "http://localhost"# Same binary file

Server logs show the proxy call being made!

      apps:./app -server monolith       # Mode 1

---

        - addr: ":3004"./app -server user-service   # Mode 2

## 💡 Key Code Sections

          required-services:./app -server order-service  # Mode 3

### 1. Auto-Router Generation (Server)

            - user-service```

```go

func runUserServer() {          published-router-services:    # ← Publish for discovery

    userService := &UserServiceImpl{}

            - user-serviceEach mode registers different services and exposes different endpoints.

    // Define REST convention

    conversionRule := autogen.ConversionRule{    

        Convention:     convention.REST,

        Resource:       "user",    order-api:## 🏗️ Architecture

        ResourcePlural: "users",

    }      base-url: "http://localhost"



    routerOverride := autogen.RouteOverride{      apps:### Deployment 1: Monolith (1 Server)

        PathPrefix: "/api/v1",

        Hidden:     []string{},        - addr: ":3005"```

    }

          required-services:┌────────────────────────────────────────┐

    // ✨ MAGIC: Generate router automatically

    router := autogen.NewFromService(userService, conversionRule, routerOverride)            - order-service│   Monolith Server (Port 3003)          │

    

    app := lokstra.NewApp("user-service", ":3000", router)          required-remote-services:      # ← Discover and use│                                        │

    app.Run(0)

}            - user-service-remote│  ┌──────────────────────────────────┐  │

```

```│  │  UserServiceImpl (local)         │  │

### 2. Proxy Service (Client)

│  │  - GetByID()                     │  │

```go

func runOrderClient() {**How it works:**│  │  - List()                        │  │

    // Create proxy with same convention as server

    userProxy := proxy.NewService(1. `order-service` depends on `user-service`│  └──────────────────────────────────┘  │

        "http://localhost:3000",

        autogen.ConversionRule{2. Framework checks if `user-service` is in `required-services` (local) - **NOT FOUND**│                ↑                       │

            Convention:     convention.REST,

            Resource:       "user",3. Framework checks if `user-service-remote` is in `required-remote-services` - **FOUND**│  ┌──────────────────────────────────┐  │

            ResourcePlural: "users",

        },4. Framework looks up `user-service-remote` definition → `proxy-service: user-service`│  │  OrderServiceImpl                │  │

        autogen.RouteOverride{

            PathPrefix: "/api/v1",5. Framework searches deployment for server that publishes `user-service`│  │  - GetByID() → calls UserService │  │

        },

    )6. Framework finds `user-api` publishes `user-service`│  │  - GetByUserID()                 │  │



    userRemote := NewUserServiceRemote(userProxy)7. Framework resolves URL: `http://localhost:3004`│  └──────────────────────────────────┘  │

    // Now userRemote.List() will make HTTP GET http://localhost:3000/api/v1/users

}8. Framework creates HTTP proxy client for `order-service` to call `user-service`│                                        │

```

│  Direct method calls (fast)            │

### 3. UserServiceRemote Implementation

## Project Structure│  Shared database                       │

```go

type UserServiceRemote struct {└────────────────────────────────────────┘

    userProxy *proxy.Service

}``````



func (s *UserServiceRemote) List(ctx *request.Context) error {05-auto-router-proxy/

    // Forward to proxy - will make HTTP GET /api/v1/users

    return proxy.Call(s.userProxy, "List", ctx)├── config.yaml              # Configuration with auto-router & service discovery### Deployment 2: Microservices (2 Servers)

}

```├── main.go                  # Entry point (deployment selector)```



**How it works:**├── handlers.go              # HTTP handlers┌───────────────────────┐         ┌─────────────────────────────┐

- `proxy.Call()` uses the convention to determine the HTTP method and path

- Makes the HTTP request to the remote service├── factories.go             # Service factories│  User-Service Server  │         │  Order-Service Server       │

- Returns response through ctx

└── appservice/             # Business logic│  (Port 3004)          │         │  (Port 3005)                │

---

    ├── database.go         # Database service (stub)│                       │         │                             │

## 🔑 Key Features Demonstrated

    ├── user_service.go     # User business logic│  ┌─────────────────┐  │  HTTP   │  ┌───────────────────────┐  │

### 1. **Zero Boilerplate Routing**

    ├── user_service_remote.go   # Remote user service client│  │ UserServiceImpl │  │◄────────┤  │ UserServiceRemote     │  │

**Before (Manual):**

```go    └── order_service.go    # Order business logic│  │ (local)         │  │         │  │ (proxy to :3004)      │  │

r.GET("/api/v1/users", userService.List)

r.GET("/api/v1/users/{id}", userService.Get)```│  │ - GetByID()     │  │         │  └───────────────────────┘  │

r.POST("/api/v1/users", userService.Create)

r.PUT("/api/v1/users/{id}", userService.Update)│  │ - List()        │  │         │            ↑                │

```

## Deployments│  └─────────────────┘  │         │  ┌───────────────────────┐  │

**After (Auto-Router):**

```go│                       │         │  │ OrderServiceImpl      │  │

router := autogen.NewFromService(userService, conversionRule, routerOverride)

```### 1. Monolith (All services in one process)│  Endpoints:           │         │  │ - GetByID()           │  │



All routes generated automatically!│  • GET /users         │         │  │ - GetByUserID()       │  │



### 2. **Convention-Based Remote Calls**```bash│  • GET /users/{id}    │         │  └───────────────────────┘  │



**Before (Manual HTTP):**go run . -deployment=monolith│                       │         │                             │

```go

url := "http://localhost:3000/api/v1/users"```└───────────────────────┘         │  Endpoints:                 │

req, _ := http.NewRequest("GET", url, nil)

resp, _ := http.DefaultClient.Do(req)                                  │  • GET /orders/{id}         │

// ... parse JSON, handle errors, etc.

```**Port:** 3003                                    │  • GET /users/{id}/orders   │



**After (Proxy):****Services:** database, user-service, order-service (all local)                                  └─────────────────────────────┘

```go

return proxy.Call(s.userProxy, "List", ctx)

```

Test:Key: OrderService uses UserServiceRemote which makes HTTP calls to port 3004

Convention handles URL construction + HTTP call + parsing!

```bash```

### 3. **Same Convention, Server & Client**

curl http://localhost:3003/

Both use identical configuration ensures compatibility!

curl http://localhost:3003/users## 📦 Project Structure

---

curl http://localhost:3003/users/1

## 🎓 What You Learned

curl http://localhost:3003/orders/1```

1. ✅ **Auto-Router Generation** - `autogen.NewFromService()` creates routes from service methods

2. ✅ **Proxy Service** - `proxy.Service` makes HTTP calls using conventionscurl http://localhost:3003/users/1/orders04-multi-deployment/

3. ✅ **Service Interface Pattern** - Same interface for local and remote implementations

4. ✅ **Convention Power** - Define once, use everywhere (server auto-generates routes, client auto-constructs URLs)```├── appservice/              # Service definitions (deployment-agnostic)



---│   ├── database.go         # In-memory database with User & Order models



## 🚀 Next Steps### 2. Microservices (Multiple servers with service discovery)│   ├── user_service.go     # UserServiceImpl (local implementation)



### Experiment│   ├── user_service_remote.go  # UserServiceRemote (HTTP proxy)



1. **Add more methods** to UserService (Delete, etc.)#### Start User Service:│   └── order_service.go    # OrderServiceImpl (uses UserService interface)

2. **Change convention** to RPC or custom

3. **Add middleware** to auto-generated router```bash│

4. **Add error handling** in remote calls

go run . -deployment=microservices -server=user-api├── handlers.go             # HTTP handlers (shared across all deployments)

### Related Examples

```├── registration.go         # Service registration for each server mode

- **Example 04**: Multi-deployment patterns

- **Example 03**: CRUD API - Service pattern basics├── main.go                 # Server entry points (3 functions)



---**Port:** 3004  └── test.http               # Test requests for all deployment modes



**Key Takeaway**: Auto-Router + Proxy = Zero boilerplate microservices! 🚀**Services:** database, user-service  ```


**Publishes:** user-service (for discovery)

### Key Insight: Separation of Concerns

Test:

```bash- **`/appservice`**: Service logic (same for all deployments)

curl http://localhost:3004/api/v1/users- **`handlers.go`**: HTTP layer (same for all deployments)

curl http://localhost:3004/api/v1/users/1- **`registration.go`**: What differs between deployments

```- **`main.go`**: Server configuration & routing



#### Start Order Service:## 📚 Code Walkthrough

```bash

go run . -deployment=microservices -server=order-api### 1. Service Interface Pattern

```

**appservice/user_service.go** - Interface + Local Implementation:

**Port:** 3005  ```go

**Services:** database, order-service  // Interface (used by all)

**Remote Services:** user-service-remote (auto-resolved to http://localhost:3004)type UserService interface {

    GetByID(p *GetUserParams) (*User, error)

Test:    List(p *ListUsersParams) ([]*User, error)

```bash}

curl http://localhost:3005/api/v1/orders/1

curl http://localhost:3005/api/v1/users/1/orders  # Uses remote user-service!// Local implementation (for monolith & user-service server)

```type UserServiceImpl struct {

    DB *service.Cached[*Database]

## What's Different from Example 04?}



| Feature | Example 04 | Example 05 |func (s *UserServiceImpl) GetByID(p *GetUserParams) (*User, error) {

|---------|-----------|------------|    return s.DB.MustGet().GetUser(p.ID)

| **Router Creation** | Manual in code | Auto-generated from config |}

| **Remote Service URL** | Hardcoded in config | Auto-resolved from published services |

| **Deployment Structure** | 3 separate deployments | 1 microservices deployment with 2 servers |func (s *UserServiceImpl) List(p *ListUsersParams) ([]*User, error) {

| **Service Discovery** | ❌ Not available | ✅ Via `published-router-services` |    return s.DB.MustGet().GetAllUsers()

| **Convention Support** | ❌ Manual routes | ✅ REST/RPC/GraphQL conventions |}

| **Path Prefix** | Hardcoded in handlers | Configured in `auto-router` |

func NewUserService() UserService {

## Benefits    return &UserServiceImpl{

        DB: service.LazyLoad[*Database]("db"),

### 1. **Zero Boilerplate**    }

No need to manually create routers and routes - framework generates them from service definitions.}

```

### 2. **Environment Portability**

Same configuration works across environments:**appservice/user_service_remote.go** - Remote Implementation (HTTP Proxy):

```yaml```go

# Dev// Remote implementation (for order-service server in microservices mode)

user-api:type UserServiceRemote struct {

  base-url: "http://localhost"    proxy *proxy.Router

}

# Production

user-api:func (u *UserServiceRemote) GetByID(p *GetUserParams) (*User, error) {

  base-url: "https://user-api.example.com"    var JsonWrapper struct {

```        Status string `json:"status"`

Remote services automatically resolve to the correct URL!        Data   *User  `json:"data"`

    }

### 3. **Service Discovery**    

Framework handles service location - no need for service registry like Consul or Eureka for simple cases.    // Makes HTTP GET to user-service server

    err := u.proxy.DoJSON("GET", fmt.Sprintf("/users/%d", p.ID), nil, nil, &JsonWrapper)

### 4. **Consistent Conventions**    if err != nil {

All services follow the same REST conventions defined in `auto-router`.        return nil, proxy.ParseRouterError(err)

    }

### 5. **Easy Microservices Migration**    return JsonWrapper.Data, nil

Start with monolith, split to microservices by just changing deployment config:}

```yaml

# From monolith:func NewUserServiceRemote() *UserServiceRemote {

monolith:    return &UserServiceRemote{

  servers:        proxy: proxy.NewRemoteRouter("http://localhost:3004"),

    api:    }

      apps:}

        - required-services: [user-service, order-service]```



# To microservices:**Key Benefit**: OrderService doesn't know if it's calling local or remote!

microservices:

  servers:### 2. OrderService Uses Interface

    user-api:

      apps:**appservice/order_service.go**:

        - required-services: [user-service]```go

          published-router-services: [user-service]type OrderService interface {

    order-api:    GetByID(p *GetOrderParams) (*OrderWithUser, error)

      apps:    GetByUserID(p *GetUserOrdersParams) ([]*Order, error)

        - required-services: [order-service]}

          required-remote-services: [user-service-remote]

```type OrderServiceImpl struct {

    DB    *service.Cached[*Database]

## Next Steps    Users *service.Cached[UserService]  // ← Interface, not concrete type!

}

- [ ] Add middleware configuration (auth, rate-limiting, etc.)

- [ ] Implement custom route overridesfunc (s *OrderServiceImpl) GetByID(p *GetOrderParams) (*OrderWithUser, error) {

- [ ] Add health check endpoints    order, err := s.DB.MustGet().GetOrder(p.ID)

- [ ] Implement circuit breaker for remote calls    if err != nil {

- [ ] Add distributed tracing        return nil, err

- [ ] Support multiple environments (dev/staging/prod)    }



## Related Examples    // Cross-service call - local or HTTP, doesn't matter!

    user, err := s.Users.MustGet().GetByID(&GetUserParams{ID: order.UserID})

- **Example 04**: Multi-Deployment (manual routers, hardcoded URLs)    if err != nil {

- **Example 03**: CRUD API (basic service patterns)        return nil, fmt.Errorf("order found but user not found: %v", err)

- **Example 02**: Handler Forms (HTTP handler patterns)    }



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
remote-service-definitions:
  user-service-remote:
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
