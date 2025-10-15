# Example 4: Multi-Deployment

**Demonstrates**: Same code, different deployment modes (monolith vs microservices)

## 🎯 Learning Objectives

This example shows Lokstra's deployment flexibility:

1. **Same Codebase**: One set of services works in multiple deployment modes
2. **Cross-Service Calls**: Service dependencies work transparently (local or HTTP)
3. **Flag-Based Deployment**: Simple runtime flags control deployment mode
4. **Service Isolation**: Each microservice runs independently with only needed dependencies

## 🏗️ Architecture

### Monolith Mode
```
┌─────────────────────────────────┐
│      Monolith (Port 3003)       │
│                                 │
│  ┌──────────────────────────┐  │
│  │    User Service          │  │
│  │  - List users            │  │
│  │  - Get user              │  │
│  └──────────────────────────┘  │
│                                 │
│  ┌──────────────────────────┐  │
│  │    Order Service         │  │
│  │  - Get order + user      │  │
│  │  - Get user's orders     │  │
│  └──────────────────────────┘  │
│                                 │
│  All services in one process    │
│  Direct method calls            │
└─────────────────────────────────┘
```

### Microservices Mode
```
┌─────────────────────┐      ┌──────────────────────┐
│  User Service       │      │  Order Service       │
│  (Port 3004)        │      │  (Port 3005)         │
│                     │      │                      │
│  • GET /users       │◄─────│  Calls user-service  │
│  • GET /users/{id}  │      │  to verify users     │
│                     │      │                      │
│  Standalone service │      │  • GET /orders/{id}  │
│  No dependencies    │      │  • GET /users/{id}/  │
│                     │      │    orders            │
└─────────────────────┘      └──────────────────────┘
```

## 📦 What's Inside

### Models

```go
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

type Order struct {
    ID      int     `json:"id"`
    UserID  int     `json:"user_id"`
    Product string  `json:"product"`
    Amount  float64 `json:"amount"`
}
```

### Services

**UserService**: Manages user data
```go
type UserService struct {
    DB *service.Lazy[*Database]
}

func (s *UserService) GetByID(p *GetUserParams) (*User, error) {
    return s.DB.Get().GetUser(p.ID)
}

func (s *UserService) List(p *ListUsersParams) ([]*User, error) {
    return s.DB.Get().GetAllUsers()
}
```

**OrderService**: Manages orders with user dependency
```go
type OrderService struct {
    DB    *service.Lazy[*Database]
    Users *service.Lazy[*UserService] // Cross-service dependency!
}

func (s *OrderService) GetByID(p *GetOrderParams) (*OrderWithUser, error) {
    // Get order
    order, err := s.DB.Get().GetOrder(p.ID)
    if err != nil {
        return nil, err
    }

    // Cross-service call to get user
    // In monolith: Direct method call
    // In microservices: Would be HTTP call
    user, err := s.Users.Get().GetByID(&GetUserParams{ID: order.UserID})
    if err != nil {
        return nil, err
    }

    return &OrderWithUser{Order: order, User: user}, nil
}
```

### Key Concept: Cross-Service Dependencies

OrderService has a dependency on UserService via `service.Lazy`:

```go
Users *service.Lazy[*UserService]
```

When you call:
```go
user, err := s.Users.Get().GetByID(&GetUserParams{ID: order.UserID})
```

**In monolith mode**:
- Direct method call to UserService in same process
- Fast, no network overhead
- Shared database instance

**In microservices mode**:
- Could be configured to make HTTP call to user-service
- Services run independently
- Each has own process/resources

## 🚀 Running the Examples

### 1. Monolith Mode (All-in-One)

Run everything in one process:
```powershell
go run main.go -mode monolith
```

Access all endpoints on **port 3003**:
```
GET http://localhost:3003/users
GET http://localhost:3003/users/1
GET http://localhost:3003/orders/1
GET http://localhost:3003/users/1/orders
```

### 2. Microservices Mode (Separate Services)

**Terminal 1** - Start User Service:
```powershell
go run main.go -mode user-service
```

**Terminal 2** - Start Order Service:
```powershell
go run main.go -mode order-service
```

Access services on different ports:
```
# User Service (port 3004)
GET http://localhost:3004/users
GET http://localhost:3004/users/1

# Order Service (port 3005)
GET http://localhost:3005/orders/1
GET http://localhost:3005/users/1/orders
```

## 🧪 Testing with test.http

The included `test.http` file has tests for both modes. Open it in VS Code with REST Client extension.

## 🔍 Key Features Demonstrated

### 1. **Deployment Flexibility**

Same services work in different modes:
```go
func main() {
    deployment := flag.String("mode", "monolith", "Deployment mode")
    
    switch *deployment {
    case "monolith":
        runMonolith()
    case "user-service":
        runUserService()
    case "order-service":
        runOrderService()
    }
}
```

### 2. **Service Isolation**

Each microservice registers only what it needs:

**User Service** (minimal dependencies):
```go
func runUserService() {
    // Only register user-related services
    lokstra_registry.RegisterServiceFactory("db", func() any {
        return NewDatabase()
    })
    
    lokstra_registry.RegisterServiceFactory("users", func() any {
        return &UserService{DB: service.LazyLoad[*Database]("db")}
    })
}
```

**Order Service** (with user dependency):
```go
func runOrderService() {
    // Register all dependencies
    registerServices() // db + users + orders
}
```

### 3. **Cross-Service Communication**

OrderService calls UserService transparently:
```go
// This works in both monolith and microservices mode
user, err := s.Users.Get().GetByID(&GetUserParams{ID: order.UserID})
```

**Current behavior** (demo mode):
- Both modes use direct method calls
- Shared database for simplicity

**Production setup** (with remote services):
- Would use `client.NewRemoteService()` for HTTP calls
- Each service has own database
- Services communicate via REST API

### 4. **Independent Scaling**

In microservices mode, each service can:
- Scale independently (more order-service instances if needed)
- Deploy independently (update user-service without touching orders)
- Use different resources (different DB, cache, etc.)

## 📊 Response Examples

### Monolith Info
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
    "deployment": "monolith",
    "message": "All services running in one process",
    "endpoints": {
      "users": ["GET /users", "GET /users/{id}"],
      "orders": ["GET /orders/{id}", "GET /users/{user_id}/orders"]
    }
  }
}
```

### Get Order with User (Cross-Service Call)
```bash
GET http://localhost:3003/orders/1
```

Response:
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

Notice how the order includes user data - this required a cross-service call from OrderService to UserService!

## 🎓 Production Considerations

### Remote Service Configuration

For true microservices with HTTP communication, configure remote services:

```go
// In order-service, instead of direct dependency:
lokstra_registry.RegisterServiceFactory("users", func() any {
    return client.NewRemoteService(
        "users",                        // service name
        "http://user-service:3004/api", // base URL
    )
})
```

Then calls to `s.Users.Get().GetByID()` would make HTTP requests.

### Service Discovery

In production:
- Use environment variables for service URLs
- Integrate with service discovery (Consul, Kubernetes)
- Add health checks and retries
- Implement circuit breakers

### Database Strategy

**Monolith**: Shared database instance
```go
// All services share one DB
db := NewDatabase()
```

**Microservices**: Separate databases
```go
// User service has its own DB
userDB := NewUserDatabase()

// Order service has its own DB
orderDB := NewOrderDatabase()
```

### Configuration Example

```yaml
# config-monolith.yaml
deployment:
  mode: monolith
  port: 3003

# config-user-service.yaml
deployment:
  mode: microservices
  service: user-service
  port: 3004

# config-order-service.yaml
deployment:
  mode: microservices
  service: order-service
  port: 3005
  dependencies:
    user-service: http://user-service:3004/api
```

## 💡 When to Use Each Mode

### Monolith Mode
✅ **Good for**:
- Development and testing
- Small to medium applications
- Simple deployment requirements
- When latency is critical
- Cost-sensitive projects

❌ **Avoid when**:
- Need independent scaling
- Large teams working on different services
- Services have different resource requirements

### Microservices Mode
✅ **Good for**:
- Large applications
- Independent team ownership
- Different scaling requirements per service
- Polyglot architecture (mix languages)
- Fault isolation important

❌ **Avoid when**:
- Small team or application
- High inter-service communication
- Limited ops/infrastructure experience
- Cost/complexity not justified

## 🔗 Related Examples

- **Example 3 (CRUD API)**: Shows service layer pattern used here
- **Essentials Section**: Deep dive into services and remote services
- **Configuration Guide**: How to configure different deployment modes

## 📚 What You Learned

1. ✅ Same code runs in multiple deployment modes
2. ✅ Cross-service dependencies work transparently
3. ✅ Service isolation for independent deployment
4. ✅ Flag-based runtime configuration
5. ✅ Foundation for production microservices
6. ✅ When to choose monolith vs microservices

## 🎯 Next Steps

- Explore **Configuration Guide** for advanced deployment configs
- Read **Remote Services** guide for HTTP-based service communication
- Check **Service Guide** for dependency injection patterns
- See **Production Deployment** for scaling strategies
