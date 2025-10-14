# Why Lokstra?

> **Understanding the problems Lokstra solves and when to use it**

---

## 🤔 The Problem

Building REST APIs in Go, you typically face these challenges:

### 1. **Standard Library is Too Low-Level**
```go
// stdlib net/http - verbose boilerplate
func main() {
    mux := http.NewServeMux()
    
    // Manual routing
    mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "GET" {
            http.Error(w, "Method not allowed", 405)
            return
        }
        
        // Manual JSON parsing
        var users []User
        // ... database code ...
        
        // Manual JSON encoding
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(users)
    })
    
    http.ListenAndServe(":8080", mux)
}
```

**Problems**:
- Too much boilerplate
- Manual request/response handling
- No built-in middleware system
- No dependency injection

---

### 2. **Popular Frameworks Have Limitations**

#### Gin / Echo / Chi
Great frameworks, but:

```go
// Handler signature is fixed
func GetUsers(c *gin.Context) {
    // Must use c.JSON(), c.Bind(), etc
    // Locked into framework-specific types
}
```

**Problems**:
- ❌ Fixed handler patterns
- ❌ No built-in service layer
- ❌ No dependency injection
- ❌ Config-driven deployment limited
- ❌ Hard to migrate monolith → microservices

---

### 3. **Enterprise Frameworks Too Heavy**

Complex frameworks with:
- Too many concepts to learn
- Over-engineered for simple APIs
- Steep learning curve
- Overkill for most projects

---

## 💡 The Lokstra Solution

Lokstra addresses these problems with a balanced approach:

### 1. **Flexible Handler Signatures**

**Problem**: Most frameworks force one pattern

**Lokstra**: Write handlers that make sense for your use case

```go
// Simple - no params, no errors
r.GET("/ping", func() string {
    return "pong"
})

// With error handling
r.GET("/users", func() ([]User, error) {
    return db.GetAllUsers()
})

// With request binding
r.POST("/users", func(req *CreateUserRequest) (*User, error) {
    return db.CreateUser(req)
})

// Full control
r.GET("/complex", func(ctx *request.Context) (*response.Response, error) {
    // Access headers, cookies, etc.
    return response.Success(data), nil
})
```

**29 different handler forms supported!** Use what fits your needs.

---

### 2. **Service-First Architecture**

**Problem**: Business logic scattered in handlers

**Lokstra**: Services are first-class citizens

```go
// Define service
type UserService struct {
    DB *service.Lazy[*Database]
}

// Business logic in service
func (s *UserService) GetUsers() ([]User, error) {
    return s.DB.Get().Query("SELECT * FROM users")
}

// Use in handler
r.GET("/users", func() ([]User, error) {
    users := lokstra_registry.GetService[*UserService]("users")
    return users.GetUsers()
})

// OR: Service becomes router automatically!
userRouter := router.NewFromService(userService, "/users")
// Creates GET /users, POST /users, etc automatically
```

---

### 3. **Built-in Dependency Injection**

**Problem**: Manual dependency wiring or external DI containers

**Lokstra**: Simple, built-in DI with lazy loading

```go
// Register factories
lokstra_registry.RegisterServiceFactory("db", createDB)
lokstra_registry.RegisterServiceFactory("users", func() any {
    return &UserService{
        DB: service.LazyLoad[*Database]("db"),
    }
})

// Use anywhere - services auto-resolved
users := lokstra_registry.GetService[*UserService]("users")
```

**Benefits**:
- No external framework needed
- Type-safe with generics
- Lazy loading (efficient)
- Testable (easy mocking)

---

### 4. **Deploy Anywhere Without Code Changes**

**Problem**: Monolith → Microservices requires rewrite

**Lokstra**: Configure deployment, not code

```yaml
# Same code, different deployment!

# Monolith
servers:
  - name: monolith
    deployment-id: monolith
    apps:
      - addr: ":8080"
        services: [users, orders, payments]

# Microservices
servers:
  - name: user-service
    deployment-id: microservices
    base-url: http://user-service
    apps:
      - addr: ":8001"
        services: [users]
  
  - name: order-service
    deployment-id: microservices
    base-url: http://order-service
    apps:
      - addr: ":8002"
        services: [orders, payments]
```

**Change deployment with just a flag**:
```bash
./app --server=monolith        # Run as monolith
./app --server=user-service    # Run as microservice
```

**One binary, infinite architectures!**

---

## 📊 Comparison Matrix

| Feature | stdlib | Gin/Echo | Chi | Lokstra |
|---------|--------|----------|-----|---------|
| **Handler Flexibility** | ❌ | ⚠️ 1 form | ⚠️ 1 form | ✅ 29 forms |
| **Auto JSON Response** | ❌ | ✅ | ⚠️ | ✅ |
| **Service Layer** | ❌ | ❌ | ❌ | ✅ Built-in |
| **Dependency Injection** | ❌ | ❌ | ❌ | ✅ Built-in |
| **Service as Router** | ❌ | ❌ | ❌ | ✅ Unique |
| **Config-Driven Deploy** | ❌ | ⚠️ Limited | ❌ | ✅ Full |
| **Multi-Deployment** | ❌ | ❌ | ❌ | ✅ 1 binary |
| **Middleware System** | ⚠️ Basic | ✅ | ✅ | ✅ Enhanced |
| **Learning Curve** | Easy | Easy | Easy | Medium |
| **Boilerplate** | High | Low | Medium | Very Low |
| **Type Safety** | ✅ | ⚠️ | ✅ | ✅ |
| **Performance** | ⚡⚡⚡ | ⚡⚡⚡ | ⚡⚡⚡ | ⚡⚡⚡ |

**Legend**:
- ✅ Full support
- ⚠️ Partial support
- ❌ Not supported
- ⚡ Performance rating

---

## ✅ When to Use Lokstra

### Perfect For:

#### 1. **REST APIs (Sweet Spot)**
```go
// Build REST APIs fast with less code
r := lokstra.NewRouter("api")
r.GET("/users", getUsers)
r.POST("/users", createUser)
r.PUT("/users/{id}", updateUser)
```

#### 2. **Microservices Architecture**
```yaml
# Same code, different servers
servers:
  - name: user-service
  - name: order-service
  - name: payment-service
```

#### 3. **Monolith with Migration Plan**
```yaml
# Start as monolith
servers:
  - name: monolith
    apps:
      - services: [users, orders, payments]

# Later: Split to microservices (no code change!)
```

#### 4. **Service-Heavy Applications**
```go
// Rich business logic in services
type OrderService struct {
    Users    *service.Lazy[*UserService]
    Payments *service.Lazy[*PaymentService]
    Inventory *service.Lazy[*InventoryService]
}
```

#### 5. **Multi-Environment Deployments**
```yaml
# dev.yaml
servers:
  - name: dev-server
    apps:
      - addr: ":3000"

# prod.yaml
servers:
  - name: prod-server
    apps:
      - addr: ":80"
```

---

## 🚫 When NOT to Use Lokstra

### Consider Alternatives For:

#### 1. **GraphQL-First APIs**
Lokstra is optimized for REST. For GraphQL:
- Use `gqlgen` or `graphql-go`
- Lokstra can host GraphQL, but not optimized for it

#### 2. **Pure gRPC Services**
Lokstra focuses on HTTP/REST. For gRPC:
- Use official `grpc-go`
- Or use Lokstra for REST + separate gRPC service

#### 3. **Static File Servers**
```go
// Just serving files? stdlib is enough
http.FileServer(http.Dir("./static"))
```

#### 4. **Learning Go**
If you're **new to Go**:
- Start with stdlib `net/http` first
- Learn Go fundamentals
- Then adopt Lokstra for productivity

#### 5. **Extreme Performance Requirements**
If you need **absolute fastest** (microseconds matter):
- Use `fasthttp` directly
- Or lightweight frameworks like `fiber`
- Lokstra is fast, but prioritizes features over raw speed

#### 6. **Simple CRUD with Database Only**
If it's just database CRUD with no business logic:
- Consider simpler frameworks (Gin, Echo)
- Or even stdlib + SQL library
- Lokstra's power shines with complex logic

---

## 🎯 Lokstra's Philosophy

### Core Principles:

#### 1. **Convention over Configuration**
Smart defaults, configure only when needed:
```go
// Minimal config - just works
r := lokstra.NewRouter("api")
r.GET("/users", getUsers)
```

#### 2. **Service-Oriented**
Business logic belongs in services:
```go
// Not in handlers
func handler() { /* business logic */ }  // ❌

// In services
type UserService struct {}
func (s *UserService) CreateUser() {}    // ✅
```

#### 3. **Flexible, Not Opinionated**
Multiple ways to solve problems:
```go
// Use what fits your needs
r.Use(middleware.Direct())     // Option 1
r.Use("middleware_name")       // Option 2
```

#### 4. **Production-Ready**
Built for real applications:
- Graceful shutdown
- Multi-environment support
- Observability hooks
- Error handling

#### 5. **Developer Experience**
Make developers happy:
- Clear error messages
- Good debugging tools (`PrintRoutes()`, `Walk()`)
- Comprehensive documentation
- Runnable examples

---

## 🚀 Getting Started

Convinced? Here's what to do next:

### 1. **Quick Evaluation (5 minutes)**
👉 [Quick Start Guide](quick-start.md) - Build your first API

### 2. **Deep Understanding (20 minutes)**
👉 [Architecture](architecture.md) - How Lokstra works internally

### 3. **Learn by Doing (2-3 hours)**
👉 [Essentials Tutorial](../01-essentials/README.md) - Hands-on learning

### 4. **See It in Action**
👉 [Complete Examples](../05-examples/README.md) - Real applications

---

## 💭 Still Deciding?

### Common Questions:

**Q: Is Lokstra mature enough for production?**  
A: Yes! Already used in production applications. Active development and maintenance.

**Q: How's the performance?**  
A: Comparable to Gin/Echo. Not as fast as raw fasthttp, but fast enough for 99% of use cases.

**Q: Can I migrate from Gin/Echo?**  
A: Yes! Gradual migration is possible. Start with new features in Lokstra, keep existing code.

**Q: What about community and support?**  
A: Growing community, active GitHub discussions, comprehensive docs, and responsive maintainers.

**Q: Is it stable? Breaking changes?**  
A: API is stabilizing. We follow semantic versioning. Breaking changes only in major versions.

---

## 📚 Learn More

- **Next**: [Architecture Overview](architecture.md) - Understand how it works
- **Or**: [Quick Start](quick-start.md) - Start coding now
- **Or**: [Key Features](key-features.md) - Deep dive into unique features

---

**Ready to build better APIs?** 🚀
