# 04-Multi-Deployment - Quick Summary

> **⚠️ Note**: This example demonstrates the **manual approach** for educational purposes.  
> Advanced patterns (auto service-to-router, conventions, config-driven) will be covered in later chapters.

## 🎯 Core Concept

**One binary, three deployment modes** - Same codebase runs as monolith or microservices through interface-based abstraction and deployment-specific service registration.

**Manual approach used here**:
- Manual handler creation from service methods
- Manual `proxy.Router` with `DoJSON()` calls
- Manual service registration per deployment
- Hardcoded configuration

**What you'll learn in later chapters**:
- Auto service-to-router with `router.NewFromService()`
- Convention-based routing (RESTful, RPC)
- Auto proxy with `proxy.Service`
- Config-driven deployment (YAML/code)

## 📁 File Structure

```
├── appservice/                    # Deployment-agnostic services
│   ├── database.go               # Models + in-memory DB
│   ├── user_service.go           # UserServiceImpl (local)
│   ├── user_service_remote.go    # UserServiceRemote (HTTP proxy)
│   └── order_service.go          # OrderServiceImpl (uses UserService interface)
│
├── handlers.go                   # Shared handlers (all deployments)
├── registration.go               # Deployment-specific service registration
├── main.go                       # 3 server entry points
└── test.http                     # HTTP tests
```

## 🔧 How It Works

### The Magic: Service Interface

```go
// Interface (in appservice/)
type UserService interface {
    GetByID(p *GetUserParams) (*User, error)
    List(p *ListUsersParams) ([]*User, error)
}

// Implementation 1: Local (direct DB)
type UserServiceImpl struct { ... }

// Implementation 2: Remote (HTTP proxy to port 3004)
type UserServiceRemote struct {
    proxy *proxy.Router
}
```

### OrderService Uses Interface

```go
type OrderServiceImpl struct {
    Users *service.Cached[UserService]  // ← Interface!
}

func (s *OrderServiceImpl) GetByID(...) (*OrderWithUser, error) {
    // Works with both UserServiceImpl and UserServiceRemote
    user, err := s.Users.MustGet().GetByID(...)
}
```

### Different Registration Per Deployment

**Monolith**:
```go
old_registry.RegisterServiceType("usersFactory", appservice.NewUserService)
// Returns UserServiceImpl (local)
```

**Order-Service**:
```go
old_registry.RegisterServiceTypeRemote("usersFactory", appservice.NewUserServiceRemote)
// Returns UserServiceRemote (HTTP proxy)
```

## 🚀 Run Modes

```bash
# Build once
go build .

# Run in 3 modes
./app -server monolith        # Port 3003, all services local
./app -server user-service    # Port 3004, only user service
./app -server order-service   # Port 3005, user via HTTP to 3004
```

## 📊 Service Registration Matrix

| Server | Database | UserService | OrderService |
|--------|----------|-------------|--------------|
| **Monolith** | Local | `UserServiceImpl` | `OrderServiceImpl` |
| **User-service** | Local | `UserServiceImpl` | ❌ Not registered |
| **Order-service** | Local | `UserServiceRemote` (→ :3004) | `OrderServiceImpl` |

## ✅ What's Shared (100% Reuse)

- ✅ Service interfaces
- ✅ Service implementations (both local & remote)
- ✅ Handlers
- ✅ Models

## ❌ What's Different

- ❌ Service registration (`registration.go`)
- ❌ Router configuration (`main.go`)
- ❌ Port numbers

## 🎓 Key Learnings

1. **Interface abstraction** enables transparent local/remote calls
2. **proxy.Router** provides clean HTTP client wrapper (manual approach)
3. **Single binary** eliminates version skew between services
4. **Deployment-specific registration** is the only varying part
5. **No unified config** in this example (hardcoded for clarity)
6. **Manual patterns teach fundamentals** before automation

## 🔗 Key Files to Study

1. **`appservice/user_service_remote.go`** - Manual HTTP proxy pattern with DoJSON
2. **`registration.go`** - Manual service registration differs per deployment
3. **`handlers.go`** - Manual handler creation from service methods
4. **`appservice/order_service.go`** - Cross-service dependency via interface

## 💡 Evolution Path

This example shows **manual approach**. To evolve to production patterns:

### Current (Manual):
```go
// Manual handler
func getUserHandler(ctx *request.Context) error {
    var params GetUserParams
    ctx.Req.BindPath(&params)
    user, err := userService.MustGet().GetByID(&params)
    return ctx.Api.Ok(user)
}
r.GET("/users/{id}", getUserHandler)

// Manual proxy
err := u.proxy.DoJSON("GET", fmt.Sprintf("/users/%d", p.ID), nil, nil, &JsonWrapper)

// Manual registration
old_registry.RegisterServiceType("usersFactory", NewUserService)
old_registry.RegisterServiceTypeRemote("usersFactory", NewUserServiceRemote)
```

### Future (Automated - covered in later chapters):
```go
// Auto router from service
restConvention := conventions.RESTful
overrides := router.WithOverrides(map[string]string{
    "GetByID": "GET /users/{id}",
})
userRouter := router.NewFromService("users", restConvention, overrides)

// Auto proxy with same conventions
userServiceRemote := proxy.NewService("users", "http://localhost:3004",
    restConvention, overrides)

// Config-driven registration
# deployment.yaml
servers:
  - name: monolith
    services: [users, orders]
  - name: user-service  
    services: [users]
```

## 💡 Production Next Steps

- [ ] Add unified config for ports & URLs
- [ ] Add circuit breaker to UserServiceRemote
- [ ] Add metrics & tracing
- [ ] Add health checks
- [ ] Separate databases per service
- [ ] Add service discovery (Consul/K8s)
