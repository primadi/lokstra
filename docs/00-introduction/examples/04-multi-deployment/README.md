# Example 4: Multi-Deployment with Auto-Router & Proxy

**Demonstrates**: Convention-based auto-router generation, metadata-driven routing, and seamless local/remote service switching

---

## 📌 About This Example

This example showcases Lokstra's **production-ready patterns** for building flexible, deployment-agnostic applications:

### Key Features:
- ✅ **Auto-router generation** from service methods using conventions
- ✅ **Metadata-driven routing** with `RemoteServiceMeta` interface
- ✅ **Single source of truth** for routing configuration
- ✅ **Seamless proxy services** with `proxy.Service` and `proxy.CallWithData`
- ✅ **Single binary, multiple deployments** (monolith, microservices)
- ✅ **YAML-driven configuration** with optional code-based overrides
- ✅ **Clean Architecture** with separated layers (contract, model, service, repository)

### What's New vs Manual Approach:
- 🚀 **No manual handler creation** - auto-generated from service methods
- 🚀 **No manual route definitions** - convention-based mapping
- 🚀 **No manual proxy.Router calls** - `proxy.Service` handles it
- 🚀 **Metadata in service code** - no redundant YAML config
- 🚀 **Clean separation** - contracts, models, services, repositories in separate packages

---

## 🎯 Learning Objectives

1. **Convention-Based Routing**: How services auto-generate RESTful endpoints
2. **Clean Architecture**: Separation of concerns with contract, model, service, repository layers
3. **Metadata Architecture**: Single source of truth in `RemoteServiceMetaAdapter`
4. **3-Level Metadata System**: Service code → RegisterServiceType → YAML config
5. **Auto-Proxy Pattern**: `proxy.CallWithData` with convention mapping
6. **Deployment Flexibility**: Same code, different runtime behavior
7. **Interface-Based DI**: Depend on contracts, not implementations

---

## 🏗️ Architecture

### Deployment 1: Monolith (1 Server)
```
┌────────────────────────────────────────┐
│   API Server (Port 3003)               │
│                                        │
│  ┌──────────────────────────────────┐  │
│  │  UserServiceImpl (local)         │  │
│  │  Auto-Router:                    │  │
│  │  • GET /users                    │  │
│  │  • GET /users/{id}               │  │
│  └──────────────────────────────────┘  │
│                ↑                       │
│  ┌──────────────────────────────────┐  │
│  │  OrderServiceImpl (local)        │  │
│  │  Auto-Router:                    │  │
│  │  • GET /orders/{id}              │  │
│  │  • GET /users/{user_id}/orders   │  │
│  └──────────────────────────────────┘  │
│                                        │
│  Direct method calls (in-process)      │
└────────────────────────────────────────┘
```

### Deployment 2: Microservices (2 Servers)
```
┌───────────────────────┐         ┌─────────────────────────────┐
│  User Server          │         │  Order Server               │
│  (Port 3004)          │         │  (Port 3005)                │
│                       │         │                             │
│  ┌─────────────────┐  │  HTTP   │  ┌───────────────────────┐  │
│  │ UserServiceImpl │  │◄────────┤  │ UserServiceRemote     │  │
│  │ Auto-Router:    │  │         │  │ (proxy.Service)       │  │
│  │ • GET /users    │  │         │  │ • CallWithData        │  │
│  │ • GET /users/{id}  │         │  └───────────────────────┘  │
│  └─────────────────┘  │         │            ↑                │
│                       │         │  ┌───────────────────────┐  │
└───────────────────────┘         │  │ OrderServiceImpl      │  │
                                  │  │ Auto-Router:          │  │
                                  │  │ • GET /orders/{id}    │  │
                                  │  │ • GET /users/{uid}/   │  │
                                  │  │   orders              │  │
                                  │  └───────────────────────┘  │
                                  └─────────────────────────────┘

Key: UserServiceRemote uses metadata to auto-map methods to HTTP endpoints
```

---

## 📦 Project Structure

```
04-multi-deployment/
├── contract/                    # Application Layer - Interfaces & DTOs
│   ├── user_contract.go         # UserService interface + request/response types
│   └── order_contract.go        # OrderService interface + DTOs
│
├── model/                       # Domain Layer - Pure business entities
│   ├── user.go                  # User entity
│   └── order.go                 # Order entity
│
├── repository/                  # Infrastructure Layer - Data access
│   ├── user_repository.go       # UserRepository interface + in-memory impl
│   └── order_repository.go      # OrderRepository interface + in-memory impl
│
├── service/                     # Application Layer - Business logic
│   ├── user_service.go          # UserServiceImpl (local implementation)
│   ├── user_service_remote.go   # UserServiceRemote (HTTP proxy)
│   ├── order_service.go         # OrderServiceImpl (local implementation)
│   └── order_service_remote.go  # OrderServiceRemote (HTTP proxy)
│
├── config.yaml                  # Multi-deployment configuration
├── main.go                      # Entry point with service registration
└── test.http                    # API tests
```

### 🏛️ Clean Architecture Layers

```
┌─────────────────────────────────────────────────────────┐
│  Application Layer (service/, contract/)                │
│  - Business logic & use cases                           │
│  - Service interfaces & DTOs                            │
│  - Depends on: Domain Layer                             │
├─────────────────────────────────────────────────────────┤
│  Domain Layer (model/)                                  │
│  - Pure business entities                               │
│  - No external dependencies                             │
├─────────────────────────────────────────────────────────┤
│  Infrastructure Layer (repository/)                     │
│  - Data access implementations                          │
│  - External service adapters                            │
│  - Depends on: Domain interfaces                        │
└─────────────────────────────────────────────────────────┘

Dependency Rule: Outer layers depend on inner layers, never the reverse
```

---

## 🔑 Key Concepts

### 1. **Clean Architecture Pattern**

This example follows **Clean Architecture** principles with clear separation of concerns:

#### **Layer 1: Domain (model/)**
Pure business entities with no external dependencies:

```go
// model/user.go
package model

type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}
```

**Characteristics**:
- ✅ No framework dependencies
- ✅ No infrastructure dependencies  
- ✅ Pure Go structs
- ✅ Represents business concepts

#### **Layer 2: Application (contract/, service/)**

**contract/user_contract.go** - Service interfaces & DTOs:
```go
package contract

import "example/model"

// Service interface (application boundary)
type UserService interface {
    GetByID(p *GetUserParams) (*model.User, error)
    List(p *ListUsersParams) ([]*model.User, error)
}

// DTOs (Data Transfer Objects)
type GetUserParams struct {
    ID int `path:"id"`
}
```

**service/user_service.go** - Business logic implementation:
```go
package service

import "example/contract"
import "example/repository"

type UserServiceImpl struct {
    UserRepo *service.Cached[repository.UserRepository]  // Depend on interface!
}

var _ contract.UserService = (*UserServiceImpl)(nil)  // Ensure implementation

func (s *UserServiceImpl) GetByID(p *contract.GetUserParams) (*model.User, error) {
    return s.UserRepo.MustGet().GetByID(p.ID)
}
```

**Characteristics**:
- ✅ Depends on domain models
- ✅ Depends on repository interfaces (not implementations!)
- ✅ Contains business logic
- ✅ Framework-agnostic (testable!)

#### **Layer 3: Infrastructure (repository/)**

**repository/user_repository.go** - Data access:
```go
package repository

import "example/model"

// Repository interface (defined by application layer needs)
type UserRepository interface {
    GetByID(id int) (*model.User, error)
    List() ([]*model.User, error)
}

// Implementation (infrastructure detail)
type UserRepositoryMemory struct {
    users map[int]*model.User
}

var _ UserRepository = (*UserRepositoryMemory)(nil)

func (r *UserRepositoryMemory) GetByID(id int) (*model.User, error) {
    // Implementation details
}
```

**Characteristics**:
- ✅ Implements repository interfaces
- ✅ Can be swapped (memory → postgres → redis)
- ✅ Infrastructure details hidden behind interface

#### **Dependency Flow**

```
main.go → service → repository interface
                 ↓
              model

Dependencies point INWARD (toward domain)
```

**Benefits**:
1. **Testability**: Easy to mock repositories and test services
2. **Flexibility**: Swap implementations without changing business logic
3. **Maintainability**: Clear boundaries between layers
4. **Scalability**: Add features without touching existing code
5. **Team collaboration**: Clear ownership per layer

---

### 2. **RemoteServiceMeta Interface**

Services provide routing metadata via embedded `RemoteServiceMetaAdapter`:

**service/user_service_remote.go**:
```go
type UserServiceRemote struct {
    service.RemoteServiceMetaAdapter  // Metadata + proxy.Service
}

func NewUserServiceRemote(proxyService *proxy.Service) *UserServiceRemote {
    return &UserServiceRemote{
        RemoteServiceMetaAdapter: service.RemoteServiceMetaAdapter{
            Resource:     "user",
            Plural:       "users",
            Convention:   "rest",
            ProxyService: proxyService,
        },
    }
}
```

**Benefits**:
- ✅ Single source of truth for routing
- ✅ Used by auto-router generation
- ✅ Used by proxy.Service for HTTP calls
- ✅ No separate field for proxy.Service

### 3. **Auto-Router Generation**

Framework scans service methods and generates routes using conventions:

```go
// Service method
func (s *UserServiceImpl) GetByID(p *GetUserParams) (*User, error)

// Auto-generates
GET /users/{id} -> UserService.GetByID
```

**Convention mapping**:
| Method Name | HTTP Method | Path |
|-------------|-------------|------|
| `List()` | GET | `/users` |
| `GetByID(params)` | GET | `/users/{id}` |
| `Create(params)` | POST | `/users` |
| `Update(params)` | PUT | `/users/{id}` |
| `Delete(params)` | DELETE | `/users/{id}` |
| Custom actions | POST | `/actions/{snake_case}` |

### 4. **Custom Route Overrides**

Services can override convention-based routes:

**service/order_service_remote.go**:
```go
RemoteServiceMetaAdapter: service.RemoteServiceMetaAdapter{
    Resource:     "order",
    Plural:       "orders",
    Convention:   "rest",
    ProxyService: proxyService,
    Override: autogen.RouteOverride{
        Custom: map[string]autogen.Route{
            "GetByUserID": {Method: "GET", Path: "/users/{user_id}/orders"},
        },
    },
}
```

Result:
- `GetByID()` → `GET /orders/{id}` (convention)
- `GetByUserID()` → `GET /users/{user_id}/orders` (custom override)

### 5. **Proxy.Service Pattern**

Remote services use `proxy.CallWithData` for type-safe HTTP calls:

```go
func (u *UserServiceRemote) GetByID(params *GetUserParams) (*User, error) {
    return proxy.CallWithData[*User](u.GetProxyService(), "GetByID", params)
}
```

**What happens**:
1. Framework resolves method name to HTTP route using metadata
2. Extracts path params from struct tags (`path:"id"`)
3. Makes HTTP request
4. Auto-extracts data from JSON wrapper (`{"data": {...}}`)
5. Returns typed result

**No manual URL construction, no manual JSON parsing!**

### 6. **3-Level Metadata System**

Metadata can be provided in 3 places with priority:

```
Priority 1 (HIGH):  YAML config (router-overrides)     ← Deployment-specific
Priority 2 (MED):   XXXRemote struct (code)            ← Service-level defaults
Priority 3 (LOW):   RegisterServiceType options        ← Framework defaults
```

**Recommended**: Put metadata in `XXXRemote` struct only. Use YAML only for deployment-specific overrides.

---

## 🚀 Running the Examples

### Option 1: Monolith Deployment

```powershell
go run . -server "monolith.api-server"
```

Output:
```
Starting [api-server] with 2 router(s) on address :3003
[user-auto] GET /users/{id} -> user-auto.GetByID
[user-auto] GET /users -> user-auto.List
[order-auto] GET /orders/{id} -> order-auto.GetByID
[order-auto] GET /users/{user_id}/orders -> order-auto.GetByUserID
```

All endpoints on **port 3003**.

### Option 2: Microservices Deployment

**Terminal 1** - User Server:
```powershell
go run . -server "microservice.user-server"
```

Output:
```
Starting [user-server] with 1 router(s) on address :3004
[user-auto] GET /users/{id} -> user-auto.GetByID
[user-auto] GET /users -> user-auto.List
```

**Terminal 2** - Order Server:
```powershell
go run . -server "microservice.order-server"
```

Output:
```
Starting [order-server] with 1 router(s) on address :3005
[order-auto] GET /orders/{id} -> order-auto.GetByID
[order-auto] GET /users/{user_id}/orders -> order-auto.GetByUserID
```

---

## 📝 Configuration Walkthrough

### config.yaml

```yaml
service-definitions:
  # Infrastructure layer - Repositories
  user-repository:
    type: user-repository-factory

  order-repository:
    type: order-repository-factory

  # Application layer - Services
  user-service:
    type: user-service-factory
    depends-on: [user-repository]
  
  order-service:
    type: order-service-factory
    depends-on: [order-repository, user-service]  # Can be local OR remote

# Routers auto-generated from published-services using metadata
# Optional overrides commented out - metadata in XXXRemote is enough!

deployments:
  monolith:
    servers:
      api-server:
        base-url: "http://localhost"
        required-services: [user-repository, order-repository]
        addr: ":3003"
        published-services:
          - user-service
          - order-service

  microservice:
    servers:
      user-server:
        base-url: "http://localhost"
        required-services: [user-repository]
        addr: ":3004"
        published-services: [user-service]

      order-server:
        base-url: "http://localhost"
        required-services: [order-repository]
        required-remote-services: [user-service-remote]  # Auto-resolved!
        addr: ":3005"
        published-services: [order-service]
```

**Key Points**:
- **Layered architecture**: Repositories (infrastructure) → Services (application)
- `published-services`: Services that expose HTTP endpoints
- `required-remote-services`: Services accessed via HTTP proxy
- URLs auto-resolved from `published-services` in other servers
- No manual router definitions needed!

### main.go

```go
func main() {
    // Register repositories (infrastructure layer)
    lokstra_registry.RegisterServiceType("user-repository-factory",
        func(deps map[string]any, config map[string]any) any {
            return repository.NewUserRepositoryMemory()
        }, nil)

    lokstra_registry.RegisterServiceType("order-repository-factory",
        func(deps map[string]any, config map[string]any) any {
            return repository.NewOrderRepositoryMemory()
        }, nil)

    // Register services (application layer)
    // Metadata comes from XXXRemote structs!
    lokstra_registry.RegisterServiceType("user-service-factory",
        service.UserServiceFactory,
        service.UserServiceRemoteFactory,
        // Metadata in UserServiceRemote.RemoteServiceMetaAdapter
    )

    lokstra_registry.RegisterServiceType("order-service-factory",
        service.OrderServiceFactory,
        service.OrderServiceRemoteFactory,
        // Metadata + custom routes in OrderServiceRemote.Override!
    )

    // Load config - auto-builds ALL deployments
    lokstra_registry.LoadAndBuild([]string{"config.yaml"})

    // Run server based on flag
    lokstra_registry.RunServer(*server, 30*time.Second)
}
```

**What's simplified**:
- ❌ No manual router setup
- ❌ No manual handler registration
- ❌ No deployment-specific registration functions
- ✅ Just factory registration + LoadAndBuild!

---

## 🧪 Testing

Use `test.http` with VS Code REST Client extension:

```http
### Monolith - Get all users
GET http://localhost:3003/users

### Monolith - Get specific user
GET http://localhost:3003/users/1

### Monolith - Get order with user (cross-service)
GET http://localhost:3003/orders/1

### Monolith - Get user's orders
GET http://localhost:3003/users/1/orders

### Microservices - User server
GET http://localhost:3004/users/1

### Microservices - Order server (makes HTTP call to user server)
GET http://localhost:3005/orders/1
GET http://localhost:3005/users/1/orders
```

---

## 🔍 How It Works

### Auto-Router Generation Flow

```
1. Config loading (LoadAndBuild):
   ├─ Read config.yaml
   ├─ Find published-services: [user-service, order-service]
   ├─ Create router definitions: [user-service-router, order-service-router]
   └─ Register to global registry

2. Server startup (RunServer):
   ├─ Get router definitions for current server
   ├─ Call BuildRouterFromDefinition for each:
   │  ├─ Instantiate remote factory (UserServiceRemoteFactory)
   │  ├─ Read metadata from RemoteServiceMetaAdapter
   │  ├─ Call autogen.NewFromService(svc, rule, override)
   │  │  ├─ Scan service methods via reflection
   │  │  ├─ Map methods to routes using convention
   │  │  ├─ Apply custom overrides from metadata
   │  │  └─ Create auto-generated handlers
   │  └─ Return router
   └─ Mount all routers to app

3. Request handling:
   ├─ HTTP request arrives
   ├─ Router matches path to auto-generated handler
   ├─ Handler calls service method
   ├─ Service method:
   │  ├─ Monolith: Direct method call (UserServiceImpl)
   │  └─ Microservices: HTTP proxy call (UserServiceRemote)
   │     └─ proxy.CallWithData resolves method → HTTP using metadata
   └─ Return response
```

### Cross-Service Call Flow (Microservices)

```
Client → Order Server → User Server
                ↓
GET /orders/1   OrderServiceImpl.GetByID()
                ↓
                s.Users.MustGet().GetByID(...)
                ↓
                UserServiceRemote.GetByID()
                ↓
                proxy.CallWithData[*User](service, "GetByID", params)
                ↓
                [Metadata resolution]
                Resource: "user", Plural: "users", Convention: "rest"
                Method: "GetByID" → Convention: GET /users/{id}
                ↓
                HTTP GET http://localhost:3004/users/1
                ↓
                UserServiceImpl.GetByID() @ User Server
                ↓
                Return User
```

---

## 💡 Design Patterns

### 1. **Convention Over Configuration**

Instead of:
```yaml
# ❌ Manual route definitions
routers:
  user-router:
    routes:
      - path: /users
        method: GET
        handler: listUsers
      - path: /users/{id}
        method: GET
        handler: getUser
```

We have:
```go
// ✅ Convention-based auto-generation
type UserService interface {
    List()     // → GET /users
    GetByID()  // → GET /users/{id}
}
```

### 2. **Metadata-Driven Architecture**

Single source of truth in service code:
```go
RemoteServiceMetaAdapter{
    Resource:   "order",
    Plural:     "orders",
    Convention: "rest",
    Override: autogen.RouteOverride{
        Custom: map[string]autogen.Route{
            "GetByUserID": {Method: "GET", Path: "/users/{user_id}/orders"},
        },
    },
}
```

Used by:
- ✅ Auto-router generation (server-side)
- ✅ Proxy.Service (client-side)
- ✅ Documentation generation (future)
- ✅ API gateway configuration (future)

### 3. **Interface-Based Dependency Injection**

```go
type OrderServiceImpl struct {
    Users *service.Cached[UserService]  // Interface!
}
```

Runtime resolution:
- Monolith: `UserServiceImpl` (local)
- Microservices: `UserServiceRemote` (proxy)

Same code, different behavior!

### 4. **Zero-Boilerplate Remote Calls**

Before (manual):
```go
var wrapper struct {
    Data *User `json:"data"`
}
err := proxy.DoJSON("GET", fmt.Sprintf("/users/%d", id), nil, nil, &wrapper)
return wrapper.Data, err
```

After (auto):
```go
return proxy.CallWithData[*User](service, "GetByID", params)
```

Framework handles:
- ✅ URL construction
- ✅ Path parameter extraction
- ✅ JSON wrapper unwrapping
- ✅ Error handling

---

## 🎓 Advanced Topics

### Custom Conventions

Create your own routing conventions:

```go
convention.Register("api-v2", &convention.Definition{
    List:     "GET /{resource}",
    GetByID:  "GET /{resource}/{id}",
    Create:   "POST /{resource}",
    // ... custom patterns
})
```

### Deployment-Specific Overrides

Override metadata per environment:

```yaml
# production.yaml
routers:
  order-service-router:
    overrides: prod-overrides

router-overrides:
  prod-overrides:
    path-prefix: /api/v2  # All routes prefixed
    hidden: [InternalMethod]  # Hide from public
```

### Service Discovery Integration

Auto-resolve service URLs:

```yaml
deployments:
  kubernetes:
    servers:
      order-server:
        required-remote-services:
          - user-service-remote:
              url: "http://user-service.default.svc.cluster.local"
```

---

## 📊 Comparison: Manual vs Auto

| Aspect | Manual Approach | Auto (This Example) |
|--------|----------------|---------------------|
| Router Creation | Manual `r.GET()` | Auto-generated |
| Handler Code | Manual functions | Auto-generated |
| Proxy Calls | Manual `DoJSON()` | `CallWithData()` |
| Route Metadata | Hardcoded strings | Convention + metadata |
| Custom Routes | Manual registration | Override in metadata |
| Lines of Code | ~200 lines | ~40 lines |
| Refactoring | Manual updates | Auto-updates |

**Code reduction**: **80%** less boilerplate!

---

## 🚀 Production Considerations

### 1. **Service URLs**

Development (current):
```go
ProxyService: proxyService  // Framework-injected
```

Production:
```yaml
external-service-definitions:
  user-service-remote:
    url: "http://user-service.prod.internal:3004"
    timeout: 5s
```

### 2. **Error Handling**

Add circuit breakers, retries:
```go
func (s *OrderServiceImpl) GetByID(p *GetOrderParams) (*OrderWithUser, error) {
    user, err := s.Users.MustGet().GetByID(&GetUserParams{ID: order.UserID})
    if err != nil {
        // Handle remote call failure
        if apiErr, ok := err.(*api_client.ApiError); ok {
            return nil, fmt.Errorf("user service error: %s", apiErr.Message)
        }
        return nil, err
    }
    // ...
}
```

### 3. **Monitoring**

Framework logs auto-router generation:
```
✨ Auto-generated router 'user-service-router' from service 'user-service'
✨ Auto-generated router 'order-service-router' from service 'order-service'
```

Add custom metrics:
```go
proxy.CallWithData[*User](service, "GetByID", params)
// Framework can track: latency, errors, retries
```

---

## 🎯 Key Takeaways

### Why This Approach Is Better

1. **Less Code**: 80% reduction vs manual approach
2. **Type Safety**: Compile-time checks for method mapping
3. **Single Source of Truth**: Metadata in service code
4. **Refactoring-Friendly**: Rename method → route auto-updates
5. **Convention-Based**: Follow standards (REST, JSON:API, etc.)
6. **Flexible**: Override when needed via metadata or YAML

### When to Use Auto-Router

✅ **Good for**:
- RESTful APIs
- CRUD operations
- Microservices communication
- Rapid development
- Standard patterns

❌ **Consider manual when**:
- Highly custom routing
- Non-standard HTTP patterns
- Need fine-grained control
- Legacy API compatibility

### Production Checklist

Before deploying:
- [ ] Configure service URLs (env vars or config)
- [ ] Add health check endpoints
- [ ] Set up monitoring/metrics
- [ ] Configure timeouts and retries
- [ ] Test failure scenarios
- [ ] Document API endpoints
- [ ] Set up CI/CD pipelines

---

## 📚 Related Topics

- **01-essentials/auto-router**: Deep dive into convention system
- **01-essentials/proxy-service**: Advanced proxy patterns
- **01-essentials/metadata**: Metadata architecture
- **02-advanced/custom-conventions**: Build your own conventions
- **03-production/service-discovery**: Kubernetes, Consul integration

---

## 💡 What You Learned

1. ✅ **Clean Architecture** with contract, model, service, repository layers
2. ✅ **Auto-router generation** from service methods
3. ✅ **RemoteServiceMeta** interface for metadata
4. ✅ **3-level metadata system** (code → options → YAML)
5. ✅ **proxy.CallWithData** for type-safe HTTP calls
6. ✅ **Convention-based routing** (REST, custom)
7. ✅ **Single binary, multiple deployments**
8. ✅ **Zero-boilerplate** remote service calls
9. ✅ **Metadata-driven** architecture
10. ✅ **Interface-based dependency injection** for testability

**Next**: Explore custom conventions and advanced routing patterns! 🚀
