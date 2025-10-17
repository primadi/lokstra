# Multi-Deployment Flow Diagrams

## Request Flow Comparison

### Monolith Deployment - GET /orders/1

```
┌─────────────────────────────────────────────────────────┐
│                  Monolith Server (:3003)                │
│                                                         │
│  HTTP Request                                           │
│  GET /orders/1                                          │
│       ↓                                                 │
│  ┌──────────────────────────────────────────────────┐   │
│  │  getOrderHandler (handlers.go)                   │   │
│  │  - Bind path params                              │   │
│  │  - Call orderService.MustGet().GetByID()         │   │
│  └──────────────┬───────────────────────────────────┘   │
│                 ↓                                       │
│  ┌──────────────────────────────────────────────────┐   │
│  │  OrderServiceImpl (appservice/order_service.go)  │   │
│  │  - Get order from DB                             │   │
│  │  - Call s.Users.MustGet().GetByID()              │   │
│  └──────────────┬───────────────────────────────────┘   │
│                 ↓                                       │
│  ┌──────────────────────────────────────────────────┐   │
│  │  UserServiceImpl (appservice/user_service.go)    │   │
│  │  - Get user from DB (direct method call)         │   │
│  │  - Return User                                   │   │
│  └──────────────┬───────────────────────────────────┘   │
│                 ↓                                       │
│  Return OrderWithUser (order + user)                    │
│       ↓                                                 │
│  JSON Response                                          │
│                                                         │
└─────────────────────────────────────────────────────────┘

Time: ~1ms (all in-process)
```

### Microservices Deployment - GET /orders/1

```
┌─────────────────────────────┐      ┌─────────────────────────────┐
│ Order-Service Server (:3005)│      │ User-Service Server (:3004) │
│                             │      │                             │
│  HTTP Request               │      │                             │
│  GET /orders/1              │      │                             │
│       ↓                     │      │                             │
│  ┌────────────────────────┐ │      │                             │
│  │ getOrderHandler        │ │      │                             │
│  │ - Bind path params     │ │      │                             │
│  │ - Call orderService... │ │      │                             │
│  └──────┬─────────────────┘ │      │                             │
│         ↓                   │      │                             │
│  ┌────────────────────────┐ │      │                             │
│  │ OrderServiceImpl       │ │      │                             │
│  │ - Get order from DB    │ │      │                             │
│  │ - Call s.Users.MustGet │ │      │                             │
│  └──────┬─────────────────┘ │      │                             │
│         ↓                   │      │                             │
│  ┌────────────────────────┐ │      │                             │
│  │ UserServiceRemote      │ │  HTTP GET /users/1                 │
│  │ - proxy.DoJSON(...)    │─┼──────┼────────────────────────────►│
│  └────────────────────────┘ │      │  ┌────────────────────────┐ │
│                             │      │  │ getUserHandler         │ │
│                             │      │  │ - Bind path params     │ │
│                             │      │  └──────┬─────────────────┘ │
│                             │      │         ↓                   │
│                             │      │  ┌────────────────────────┐ │
│                             │      │  │ UserServiceImpl        │ │
│                             │      │  │ - Get user from DB     │ │
│                             │      │  │ - Return User          │ │
│                             │      │  └──────┬─────────────────┘ │
│                             │      │         ↓                   │
│  ┌────────────────────────┐ │      │  JSON Response {user}       │
│  │ Parse JSON response    │◄┼──────┼─────────────────────────────┤
│  │ Return User            │ │      │                             │
│  └──────┬─────────────────┘ │      │                             │
│         ↓                   │      │                             │
│  Return OrderWithUser       │      │                             │
│         ↓                   │      │                             │
│  JSON Response              │      │                             │
│                             │      │                             │
└─────────────────────────────┘      └─────────────────────────────┘

Time: ~10ms (includes network latency)
```

## Service Resolution Flow

### Monolith - Service Registration

```
main() → runMonolithServer()
   ↓
registerMonolithServices()
   ↓
   ├─ RegisterServiceType("dbFactory", NewDatabase)
   ├─ RegisterServiceType("usersFactory", NewUserService)
   │     ↓ Returns UserServiceImpl
   └─ RegisterServiceType("ordersFactory", NewOrderService)
   ↓
   ├─ RegisterLazyService("db", "dbFactory", nil)
   ├─ RegisterLazyService("users", "usersFactory", nil)
   │     ↓ Lazy resolver: "users" → UserServiceImpl
   └─ RegisterLazyService("orders", "ordersFactory", nil)
```

### Order-Service - Service Registration

```
main() → runOrderServiceServer()
   ↓
registerOrderServices()
   ↓
   ├─ RegisterServiceType("dbFactory", NewDatabase)
   ├─ RegisterServiceTypeRemote("usersFactory", NewUserServiceRemote)
   │     ↓ Returns UserServiceRemote (HTTP proxy)
   └─ RegisterServiceType("ordersFactory", NewOrderService)
   ↓
   ├─ RegisterLazyService("db", "dbFactory", nil)
   ├─ RegisterLazyService("users", "usersFactory", nil)
   │     ↓ Lazy resolver: "users" → UserServiceRemote
   └─ RegisterLazyService("orders", "ordersFactory", nil)
```

## Handler Call Flow

```
HTTP Request
    ↓
Router.ServeHTTP()
    ↓
Match route → getOrderHandler()
    ↓
ctx.Req.BindPath(&params)
    ↓
orderService.MustGet()
    ↓
    ├─ Monolith: Returns OrderServiceImpl instance
    └─ Order-Service: Returns OrderServiceImpl instance
    ↓
.GetByID(&params)
    ↓
OrderServiceImpl.GetByID()
    ↓
s.Users.MustGet()
    ↓
    ├─ Monolith: Returns UserServiceImpl instance
    └─ Order-Service: Returns UserServiceRemote instance
    ↓
.GetByID(&GetUserParams{ID: 1})
    ↓
    ├─ Monolith: UserServiceImpl.GetByID() → DB.GetUser()
    └─ Order-Service: UserServiceRemote.GetByID() → HTTP GET :3004/users/1
    ↓
Return User
    ↓
Combine Order + User → OrderWithUser
    ↓
ctx.Api.Ok(orderWithUser)
    ↓
JSON Response
```

## Key Insight: Same Handler, Different Behavior

```go
// handlers.go (SAME for all deployments)
func getOrderHandler(ctx *request.Context) error {
    orderWithUser, err := orderService.MustGet().GetByID(&params)
    return ctx.Api.Ok(orderWithUser)
}
```

**Monolith**:
- `orderService` → `OrderServiceImpl`
- `s.Users` → `UserServiceImpl` (local call)

**Order-Service**:
- `orderService` → `OrderServiceImpl`  
- `s.Users` → `UserServiceRemote` (HTTP call to :3004)

**Zero code changes!** Only registration differs.

## Registration Comparison

```
┌─────────────────────────────────────────────────────────────────┐
│                 Service: "users"                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Monolith Registration:                                         │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │ RegisterServiceType("usersFactory", NewUserService)       │  │
│  │   → Returns: UserServiceImpl                              │  │
│  │   → Calls: DB directly                                    │  │
│  └───────────────────────────────────────────────────────────┘  │
│                                                                 │
│  Order-Service Registration:                                    │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │ RegisterServiceTypeRemote("usersFactory",                 │  │
│  │                          NewUserServiceRemote)            │  │
│  │   → Returns: UserServiceRemote                            │  │
│  │   → Calls: HTTP GET http://localhost:3004/users/{id}      │  │
│  └───────────────────────────────────────────────────────────┘  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Type Relationships

```
UserService (interface)
    ↑
    ├── UserServiceImpl (struct)
    │   └── DB *service.Cached[*Database]
    │       └── Methods: GetByID(), List()
    │           └── Direct DB calls
    │
    └── UserServiceRemote (struct)
        └── proxy *proxy.Router
            └── Methods: GetByID(), List()
                └── HTTP calls to :3004

OrderServiceImpl (struct)
    └── Users *service.Cached[UserService]  ← Interface!
        └── Runtime resolution:
            ├── Monolith: UserServiceImpl
            └── Order-Service: UserServiceRemote
```
