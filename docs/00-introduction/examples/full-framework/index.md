---
layout: docs
title: Full Framework Examples
---

# Track 2: Full Framework Examples

> **Use Lokstra as a complete application framework (like NestJS, Spring Boot)**  
> **Time**: 8-12 hours • **Level**: Intermediate to Advanced

---

## 📚 What You'll Learn

This track covers the **complete Lokstra framework** with dependency injection, services, and deployment patterns:

- ✅ Service layer and dependency injection
- ✅ Auto-generated REST routers from services
- ✅ Configuration-driven deployment (YAML or Code)
- ✅ Monolith → Microservices migration
- ✅ External service integration
- ✅ Production-ready patterns

**Full enterprise features** - DI, auto-router, multi-deployment!

---

## 🎯 Learning Path

### [01 - CRUD API with Services](./01-crud-api/) ⏱️ 1 hour

Service-based architecture with dependency injection.

```go
type UserService struct {
    DB *service.Cached[*Database]
}

func (s *UserService) GetAll() ([]User, error) {
    return s.DB.MustGet().Query("SELECT * FROM users")
}

// Register service
lokstra_registry.RegisterServiceFactory("users", NewUserService)

// Use in handlers
var userService = service.LazyLoad[*UserService]("users")

r.GET("/users", func() ([]User, error) {
    return userService.MustGet().GetAll()
})
```

**What you'll learn:**
- Service factory pattern
- Lazy dependency injection
- Service registration and access
- Manual routing with services

---

### [02 - Multi-Deployment (YAML Config)](./02-multi-deployment-yaml/) ⏱️ 2-3 hours ⭐

One binary, multiple deployment topologies with YAML configuration.

```yaml
# config.yaml
service-definitions:
  user-service:
    type: user-service-factory
    depends-on: [database]

deployments:
  monolith:
    servers:
      api-server:
        addr: ":8080"
        published-services: [user-service, order-service]
  
  microservices:
    servers:
      user-server:
        addr: ":8001"
        published-services: [user-service]
      order-server:
        addr: ":8002"
        published-services: [order-service]
```

```bash
# Same code, different deployment!
go run . -server=monolith.api-server
go run . -server=microservices.user-server
```

**What you'll learn:**
- YAML-based configuration
- Auto-router generation from services
- Service interfaces (local vs remote)
- Proxy pattern for remote calls
- Monolith vs microservices topology

---

### [03 - Multi-Deployment (Pure Code)](./03-multi-deployment-pure-code/) ⏱️ 30 min

Same as Example 02, but 100% code-based (no YAML).

```go
// Type-safe configuration in Go
lokstra_registry.RegisterLazyService("user-service", "user-service-factory",
    map[string]any{"depends-on": []string{"database"}})

lokstra_registry.RegisterDeployment("monolith", &lokstra_registry.DeploymentConfig{
    Servers: map[string]*lokstra_registry.ServerConfig{
        "api-server": {
            Addr: ":8080",
            PublishedServices: []string{"user-service", "order-service"},
        },
    },
})
```

**What you'll learn:**
- Code-based configuration (vs YAML)
- Type safety and IDE autocomplete
- Dynamic configuration (conditionals, loops)
- When to use code vs YAML

**Benefits:**
- ✅ Compile-time type checking
- ✅ IDE autocomplete
- ✅ Refactoring-friendly
- ✅ Single language (no YAML)

---

### [04 - External Services](./04-external-services/) ⏱️ 1-2 hours ⭐

Integrate third-party APIs (payment gateways, email, etc.) as Lokstra services.

```go
// Clean service wrapper
type PaymentServiceRemote struct {
    proxyService *proxy.Service
}

func (s *PaymentServiceRemote) CreatePayment(p *CreatePaymentParams) (*Payment, error) {
    return proxy.CallWithData[*Payment](s.proxyService, "CreatePayment", p)
}

// Metadata in registry (single source of truth)
lokstra_registry.RegisterServiceType(
    "payment-service-remote-factory",
    nil, PaymentServiceRemoteFactory,
    deploy.WithResource("payment", "payments"),
    deploy.WithRouteOverride("CreatePayment", "POST /payments"),
    deploy.WithRouteOverride("Refund", "POST /payments/{id}/refund"),
)
```

**What you'll learn:**
- External service integration
- `proxy.Service` for structured APIs
- Route overrides for custom endpoints
- Service metadata patterns
- Real-world API integration (Stripe, SendGrid, etc.)

---

### [05 - Remote Router](./05-remote-router/) ⏱️ 30 min

Quick API integration without service wrappers.

```go
type WeatherService struct {
    weatherAPI *proxy.Router
}

func (s *WeatherService) GetWeather(city string) (*Weather, error) {
    var weather Weather
    err := s.weatherAPI.DoJSON("GET", fmt.Sprintf("/weather/%s", city), 
        nil, nil, &weather)
    return &weather, err
}

// Simple factory with URL
func WeatherServiceFactory(deps, cfg map[string]any) any {
    url := cfg["url"].(string)
    return &WeatherService{
        weatherAPI: proxy.NewRemoteRouter(url),
    }
}
```

**What you'll learn:**
- `proxy.Router` for simple HTTP calls
- When to use Router vs Service
- Quick API prototyping
- Direct HTTP calls vs structured services

**Use for:** Weather APIs, currency converters, one-off integrations

---

## 🚀 Running Examples

### Simple Examples (01):
```bash
cd 01-crud-api
go run main.go
curl http://localhost:3000/users
```

### Multi-Server Examples (02-03):
```bash
cd 02-multi-deployment-yaml  # or 03

# Option 1: Monolith
go run . -server=monolith.api-server

# Option 2: Microservices (2 terminals)
go run . -server=microservices.user-server   # Terminal 1
go run . -server=microservices.order-server  # Terminal 2
```

### External Service Examples (04):
```bash
cd 04-external-services

# Terminal 1: Mock gateway
cd mock-payment-gateway && go run main.go

# Terminal 2: Main app
cd .. && go run main.go
```

---

## 📊 Skills Progression

```
Example 01:  Service Architecture
    → Services, DI, factory pattern

Example 02:  YAML Configuration
    → Auto-router, multi-deployment, YAML config

Example 03:  Code Configuration
    → Pure code config, type safety

Example 04:  External Integration
    → proxy.Service, route overrides, real-world patterns

Example 05:  Quick Integration
    → proxy.Router, simple HTTP calls
```

---

## 🎯 After This Track

### Continue Learning:
- **[Framework Guide](../../../02-framework-guide/)** - Advanced DI patterns
- **[Configuration Reference](../../../03-api-reference/03-configuration/)** - Full YAML schema
- **[Production Deployment](../../../02-framework-guide/)** - Microservices patterns

### Build Real Projects:
- Apply multi-deployment patterns
- Integrate external services
- Build microservices architectures
- Use auto-router for rapid development

---

## 💡 When to Use Full Framework

**Perfect for:**
- ✅ Enterprise applications
- ✅ Microservices architectures
- ✅ Teams wanting DI and auto-router
- ✅ Configuration-driven deployment
- ✅ Scalable, maintainable codebases

**Framework advantages:**
- Type-safe dependency injection
- Auto-generated REST routers
- Zero-code deployment changes
- Service abstraction (local/remote)
- Production-ready patterns

---

## 🔄 Comparison with Other Frameworks

**Lokstra Framework is similar to:**
- **NestJS** (Node.js) - DI, decorators, modular architecture
- **Spring Boot** (Java) - Enterprise DI, auto-configuration
- **Uber Fx** (Go) - Dependency injection framework
- **Buffalo** (Go) - Full-stack web framework

**Lokstra advantages:**
- ✅ Type-safe generics (no `interface{}`)
- ✅ Auto-router from service methods
- ✅ Zero-code deployment topology changes
- ✅ Code or YAML configuration (your choice)

---

**Ready to start?** → [01 - CRUD API with Services](./01-crud-api/)

**Coming from Router Track?** This builds on routing basics with DI and services!
