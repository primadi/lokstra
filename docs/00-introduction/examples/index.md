---
layout: docs
title: Examples
---

# Lokstra Examples

> 🎯 **Progressive learning path: Manual basics → Production patterns**

Learn Lokstra step by step, from basic routing to production-ready middleware and architecture.

---

## 📚 Learning Path

```
01-hello-world
    ↓ Learn: Router basics, simple handlers
02-handler-forms
    ↓ Learn: 29 handler variations, request/response patterns
03-crud-api
    ↓ Learn: Services, dependency injection, manual routing
04-multi-deployment-yaml
    ↓ Learn: YAML config, auto-router, microservices
05-multi-deployment-pure-code ⭐ NEW!
    ↓ Learn: Pure code config, no YAML, type safety
06-external-services
    ↓ Learn: External API integration, proxy.Service, route overrides
07-remote-router ⭐ NEW!
    ↓ Learn: Quick API access with proxy.Router
08-middleware
    ↓ Learn: Global/route middleware, auth, recovery, rate limiting
```

**Time investment**: ~10-14 hours to complete all examples  
**Outcome**: Ready to build production REST APIs with Lokstra

---

## 📂 Examples

### [01-hello-world](./01-hello-world/)

**Your first Lokstra API**

- Simple router with GET handlers
- Auto JSON responses
- Basic string and map returns

```bash
cd 01-hello-world && go run main.go
curl http://localhost:3000/
```

**Why manual?** Perfect for quick prototyping and learning basics!

---

### [02-handler-forms](./02-handler-forms/)

**Explore 29 handler variations**

- Request binding (JSON, path, query, header)
- Response forms (string, map, struct, error handling)
- Context access patterns

```bash
cd 02-handler-forms && go run main.go
```

**Why manual?** Understanding handler flexibility is fundamental!

---

### [03-crud-api](./03-crud-api/)

**Full CRUD with service pattern**

- Service-based architecture
- Dependency injection
- Manual router registration

```bash
cd 03-crud-api && go run main.go
curl http://localhost:3000/users
```

**Features:**
- ✅ Service factories
- ✅ Lazy dependency injection
- ✅ Clean separation of concerns
- ✅ Manual route registration (understand the foundation!)

---

### [04-multi-deployment-yaml](./04-multi-deployment-yaml/)

**One binary, multiple deployments (YAML config)**

- YAML-based configuration
- Monolith vs Microservices
- Service interface pattern (local vs remote)
- Cross-service communication

```bash
# Run as monolith
go run . -server=monolith.api-server

# Run as microservices
go run . -server=microservice.user-server    # Terminal 1
go run . -server=microservice.order-server   # Terminal 2
```

**Key Learning:**
- Auto-router generation from service metadata
- Interface abstraction (UserService local vs remote)
- Proxy pattern for remote calls
- YAML deployment configuration

---

### [05-multi-deployment-pure-code](./05-multi-deployment-pure-code/) ⭐ NEW!

**Pure code deployment (no YAML)**

Same as example 04, but 100% code-based configuration!

- ✅ `RegisterLazyService` for service definitions
- ✅ `RegisterDeployment` for deployment topology
- ✅ Type safety with IDE autocomplete
- ✅ Refactoring-friendly

```bash
# Run as monolith
go run . -server=monolith.api-server

# Run as microservices
go run . -server=microservice.user-server    # Terminal 1
go run . -server=microservice.order-server   # Terminal 2
```

**Key Difference:**
```go
// Instead of config.yaml
lokstra_registry.RegisterLazyService("user-service", "user-service-factory", 
    map[string]any{"depends-on": []string{"user-repository"}})

lokstra_registry.RegisterDeployment("monolith", &lokstra_registry.DeploymentConfig{
    Servers: map[string]*lokstra_registry.ServerConfig{
        "api-server": {
            BaseURL: "http://localhost",
            Addr: ":3003",
            PublishedServices: []string{"user-service", "order-service"},
        },
    },
})
```

**Benefits:**
- ✅ Type safety (compile-time errors)
- ✅ IDE autocomplete
- ✅ Safe refactoring
- ✅ Dynamic configuration (conditionals, loops)
- ✅ Single language (no YAML context switching)

**When to use:**
- YAML (04): Ops teams, runtime config, non-coders
- Pure Code (05): Dev teams, version control, compile-time safety

---

### [06-external-services](./06-external-services/) ⭐

**External API integration with best DX**

This is where you learn production-ready request handling!

- ✅ Global middleware (applied to all routes)
- ✅ Route-specific middleware (per-endpoint auth)
- ✅ Custom middleware creation
- ✅ Built-in middleware (CORS, Recovery, Logger)
- ✅ Middleware chaining and execution order

```bash
cd 05-middleware
go run main.go

# Test with different scenarios
curl http://localhost:3000/                           # Public
curl http://localhost:3000/protected -H "X-API-Key: secret-key-123"  # Auth required
curl http://localhost:3000/api/admin/dashboard -H "X-API-Key: admin-key-456"  # Admin only
curl http://localhost:3000/panic                      # Recovery middleware
```

**What you'll learn:**
- ✅ **Global middleware**: Recovery, CORS, Logger, Rate Limiting
- ✅ **Auth middleware**: API key validation
- ✅ **Role-based access**: Admin-only endpoints
- ✅ **Custom middleware**: LoggingMiddleware, RateLimitMiddleware
- ✅ **Middleware chain**: Multiple middleware per route
- ✅ **Override parent**: Route with `WithOverrideParentMwOption(true)`

**Production patterns covered:**
- Panic recovery (graceful error handling)
- Request logging with timing
- Rate limiting per IP
- Authentication & Authorization
- CORS for API access

**Code size**: ~180 lines  
**Endpoints**: 11 routes with various middleware combinations

**This is essential for production!** 🚀

---

### [06-external-services](./06-external-services/) ⭐ NEW!

**External API integration with best DX**

Learn how to integrate third-party APIs (payment gateways, email services, etc.) as Lokstra services.

- ✅ **ServiceMeta** interface for metadata (works for local & remote!)
- ✅ **Route overrides in code** (not config!)
- ✅ **Auto-wrapper creation** from `external-service-definitions`
- ✅ **Convention-based proxy** with `proxy.Service`
- ✅ **Custom routes** for non-standard APIs

```bash
# Terminal 1: Start mock payment gateway
cd mock-payment-gateway
go run main.go

# Terminal 2: Start main app
cd ..
go run main.go

# Test
curl -X POST http://localhost:3000/orders \
  -H "Content-Type: application/json" \
  -d '{\"user_id\":1,\"items\":[\"Laptop\"],\"total_amount\":1299.99}'
```

**What you'll learn:**
- ✅ **External service definition**: Auto-create wrappers with `type` field
- ✅ **Clean metadata pattern**: All metadata in `RegisterServiceType` options
- ✅ **Route overrides**: Custom routes via `deploy.WithRouteOverride()`
- ✅ **Smart method names**: Use `Create`, `Get`, `Refund` (match REST convention when possible)
- ✅ **DX improvements**: Single source of truth - no duplication!

**Key Pattern:**
```go
// Simple service wrapper - no embedded metadata!
type PaymentServiceRemote struct {
    proxyService *proxy.Service
}

func (s *PaymentServiceRemote) CreatePayment(p *CreatePaymentParams) (*Payment, error) {
    return proxy.CallWithData[*Payment](s.proxyService, "CreatePayment", p)
}

// Metadata in RegisterServiceType (single source of truth!)
lokstra_registry.RegisterServiceType(
    "payment-service-remote-factory",
    nil, service.PaymentServiceRemoteFactory,
    deploy.WithResource("payment", "payments"),
    deploy.WithConvention("rest"),
    deploy.WithRouteOverride("CreatePayment", "POST /payments"),
    deploy.WithRouteOverride("Refund", "POST /payments/{id}/refund"),
)
```

**Code size**: ~400 lines  
**Endpoints**: 3 order routes + 3 payment gateway routes

**Real-world ready!** Use this pattern for Stripe, SendGrid, Twilio, etc.

---

### [07-remote-router](./07-remote-router/) ⭐ NEW!

**Quick API access without service wrappers**

Learn when to use `proxy.Router` for simple, direct HTTP calls vs `proxy.Service`.

- ✅ **Simple URL config** (no router-definitions!)
- ✅ **No service wrapper needed**
- ✅ **Direct HTTP calls** with `DoJSON()`
- ✅ **Quick integration** for one-off API calls
- ✅ **Comparison**: proxy.Router vs proxy.Service

```bash
# Terminal 1: Start mock weather API
cd mock-weather-api
go run main.go

# Terminal 2: Start main app
cd ..
go run main.go

# Test
curl -X POST "http://localhost:3001/weather-reports?city=jakarta&forecast=true&days=5"
```

**What you'll learn:**
- ✅ **When to use proxy.Router**: One-off calls, prototyping, simple APIs
- ✅ **Simple config**: Just URL, no special definitions
- ✅ **Direct HTTP**: `router.DoJSON(method, path, ...)`
- ✅ **vs proxy.Service**: When to upgrade to service wrapper

**Key Pattern:**
```go
type WeatherService struct {
    weatherAPI *proxy.Router
}

func (s *WeatherService) Create(p *GetWeatherReportParams) (*WeatherReport, error) {
    // Direct HTTP call - no wrapper!
    var current WeatherData
    err := s.weatherAPI.DoJSON("GET", fmt.Sprintf("/weather/%s", p.City), 
        nil, nil, &current)
    
    return &WeatherReport{Current: &current}, nil
}

// Factory creates router from URL
func WeatherServiceFactory(deps map[string]any, config map[string]any) any {
    url := config["weather-api-url"].(string)
    return &WeatherService{
        weatherAPI: proxy.NewRemoteRouter(url),
    }
}
```

**Code size**: ~200 lines  
**Endpoints**: 1 weather report route + 2 mock API routes

**Perfect for**: Weather APIs, currency converters, quick integrations!

---

### [08-middleware](./08-middleware/) ⭐

**Global and route-specific middleware**

This is where you learn production-ready request handling!

- ✅ Global middleware (applied to all routes)
- ✅ Route-specific middleware (per-endpoint auth)
- ✅ Custom middleware creation
- ✅ Built-in middleware (CORS, Recovery, Logger)
- ✅ Middleware chaining and execution order

```bash
cd 08-middleware
go run main.go

# Test with different scenarios
curl http://localhost:3000/                           # Public
curl http://localhost:3000/protected -H "X-API-Key: secret-key-123"  # Auth required
curl http://localhost:3000/api/admin/dashboard -H "X-API-Key: admin-key-456"  # Admin only
curl http://localhost:3000/panic                      # Recovery middleware
```

**What you'll learn:**
- ✅ **Global middleware**: Recovery, CORS, Logger, Rate Limiting
- ✅ **Auth middleware**: API key validation
- ✅ **Role-based access**: Admin-only endpoints
- ✅ **Custom middleware**: LoggingMiddleware, RateLimitMiddleware
- ✅ **Middleware chain**: Multiple middleware per route
- ✅ **Override parent**: Route with `WithOverrideParentMwOption(true)`

**Production patterns covered:**
- Panic recovery (graceful error handling)
- Request logging with timing
- Rate limiting per IP
- Authentication & Authorization
- CORS for API access

**Code size**: ~180 lines  
**Endpoints**: 11 routes with various middleware combinations

**This is essential for production!** 🚀

---

## 🎯 What You'll Learn

### 📊 Feature Coverage

| Example | What's Covered |
|---------|----------------|
| **01** | ✅ Basic Router, ✅ Simple Handlers |
| **02** | ✅ Routes, ✅ 29 Handler Forms |
| **03** | ✅ Manual Router, ✅ Services, ✅ Dependency Injection |
| **04** | ✅ YAML Config, ✅ Auto-Router, ✅ Microservices |
| **05** | ✅ Pure Code Config, ✅ Type Safety, ✅ No YAML |
| **06** | ✅ External APIs, ✅ proxy.Service, ✅ Route Overrides |
| **07** | ✅ proxy.Router, ✅ Quick Integration, ✅ Direct HTTP Calls |
| **08** | ✅ Global Middleware, ✅ Auth, ✅ Production Patterns |

### 🎓 Skills Progression

```
Example 01-02:  Basic Foundations
    → Router creation, handler patterns

Example 03:     Service Architecture  
    → DI, service layer, manual routing

Example 04-05:  Advanced Deployment
    → Auto-router, microservices, YAML vs Pure Code

Example 06-07:  External Integration
    → proxy.Service (structured), proxy.Router (simple)

Example 08:     Production Ready
    → Middleware chains, auth, recovery, CORS
```

---

## 🔄 Recommended Learning Strategy

### Week 1: Foundations (5-6 hours)
- **Day 1**: Example 01 (15min) + Example 02 (30min)
- **Day 2**: Example 03 (1 hour)
- **Day 3**: Example 04 - YAML Config (1-2 hours)
- **Day 4**: Example 05 - Pure Code (30min, compare with 04)
- **Day 5**: Review and build small API

**Goal**: Understand basics, service patterns, deployment configurations

### Week 2: Production Patterns (5-6 hours)
- **Day 1**: Example 08 - Middleware (1-2 hours)
- **Day 2**: Example 06 - External Services (1 hour)
- **Day 3**: Example 07 - Remote Router (30min)
- **Day 4-5**: Build your project using examples as template

**Goal**: Master production-ready architecture with middleware and external integrations

---

## 💡 Key Progression

### Example 01 → Router Basics
```go
r := lokstra.NewRouter("api")
r.GET("/ping", func() string { return "pong" })
```

### Example 02 → Handler Flexibility
```go
r.GET("/users/{id}", func(p *GetUserParams) (*User, error) {
    return db.GetUser(p.ID)
})
```

### Example 03 → Service Pattern
```go
type UserService struct {
    DB *service.Cached[*Database]
}

r.GET("/users", func() ([]*User, error) {
    return userService.List()
})
```

### Example 04 → Auto-Router + YAML Config
```go
# config.yaml
deployments:
  monolith:
    servers:
      api-server:
        addr: ":3003"
        published-services:
          - user-service
          - order-service

# Just define the service interface and implementation
# Routes auto-generated from metadata!
# GetByID() → GET /users/{id}
# List()    → GET /users
```

### Example 05 → Auto-Router + Pure Code Config
```go
// No YAML! 100% type-safe Go code
lokstra_registry.RegisterLazyService("user-service", "user-service-factory",
    map[string]any{"depends-on": []string{"user-repository"}})

lokstra_registry.RegisterDeployment("monolith", &lokstra_registry.DeploymentConfig{
    Servers: map[string]*lokstra_registry.ServerConfig{
        "api-server": {
            Addr: ":3003",
            PublishedServices: []string{"user-service", "order-service"},
        },
    },
})
```

### Example 06 → External Services Integration
```go
// Global middleware
r.Use(RecoveryMiddleware)
r.Use(CORSMiddleware)
r.Use(LoggerMiddleware)

// Route-specific auth
r.GET("/protected", ProtectedHandler, AuthMiddleware)
r.GET("/admin", AdminHandler, AuthMiddleware, AdminOnlyMiddleware)
```

### Example 06 → External Services Integration
```go
// Clean service wrapper - no embedded metadata
type PaymentServiceRemote struct {
    proxyService *proxy.Service
}

// Metadata in RegisterServiceType (single source of truth!)
lokstra_registry.RegisterServiceType(
    "payment-service-remote-factory",
    nil, service.PaymentServiceRemoteFactory,
    deploy.WithResource("payment", "payments"),
    deploy.WithRouteOverride("CreatePayment", "POST /payments"),
)
```

### Example 08 → Production Middleware
```go
// Global middleware
r.Use(RecoveryMiddleware)
r.Use(CORSMiddleware)
r.Use(LoggerMiddleware)

// Route-specific auth
r.GET("/protected", ProtectedHandler, AuthMiddleware)
r.GET("/admin", AdminHandler, AuthMiddleware, AdminOnlyMiddleware)
```

---

## 🚀 Running Examples

```bash
# Navigate to any example
cd 01-hello-world  # or 02, 03, 04, 05, 06

# Run it
go run main.go

# Test it (use test.http or curl from README)
curl http://localhost:3000/
```

**For multi-server examples:**

Example 04-05 (same commands for both):
```bash
cd 04-multi-deployment-yaml  # or 05-multi-deployment-pure-code

# Option 1: Monolith
go run . -server=monolith.api-server

# Option 2: Microservices (2 terminals)
go run . -server=microservice.user-server     # Terminal 1
go run . -server=microservice.order-server    # Terminal 2
```

Example 06:
```bash
cd 06-external-services

# Terminal 1: Mock gateway
cd mock-payment-gateway && go run main.go

# Terminal 2: Main app
cd .. && go run main.go
```

---

## 📚 Next Steps

After completing these examples:

- **Deep Dive**: [01-essentials](../../01-essentials)
- **API Reference**: [03-api-reference](../../03-api-reference) (coming soon)
- **Advanced Topics**: [02-deep-dive](../../02-deep-dive) (coming soon)

---

**Start here**: → [01-hello-world](./01-hello-world/) 🚀