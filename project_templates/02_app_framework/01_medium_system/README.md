# Lokstra Medium System Template

A domain-driven modular architecture template for medium-sized applications (2-10 domain entities). This template demonstrates clean architecture principles with clear separation between domain, service, and repository layers.

## 🏗️ Architecture Overview

```
Medium System Architecture
┌─────────────────────────────────────────┐
│           HTTP Handlers                 │
│         (Auto-generated)                │
├─────────────────────────────────────────┤
│          Service Layer                  │
│    (Business Logic / Use Cases)         │
├─────────────────────────────────────────┤
│         Domain Layer                    │
│  (Entities, Interfaces, DTOs)           │
├─────────────────────────────────────────┤
│       Repository Layer                  │
│    (Data Access / Infrastructure)       │
└─────────────────────────────────────────┘
```

## 📁 Project Structure

```
01_medium_system/
├── domain/                  # Domain layer (entities, contracts, DTOs)
│   ├── user/
│   │   ├── entity.go       # User domain model
│   │   ├── contract.go     # Service & Repository interfaces
│   │   └── dto.go          # Request/Response DTOs
│   └── order/
│       ├── entity.go       # Order domain model
│       ├── contract.go     # Service & Repository interfaces
│       └── dto.go          # Request/Response DTOs
│
├── service/                 # Application/Service layer
│   ├── user_service.go     # User business logic (local)
│   ├── user_service_remote.go  # User service proxy (HTTP client)
│   ├── order_service.go    # Order business logic (local)
│   └── order_service_remote.go # Order service proxy (HTTP client)
│
├── repository/              # Infrastructure/Data access layer
│   ├── user_repository.go  # User data access
│   └── order_repository.go # Order data access
│
├── main.go                  # Application entry point
├── register.go              # Service registration
├── config.yaml              # Deployment configuration
├── test.http                # API testing file
└── README.md               # This file
```

## 🎯 Key Concepts

### 1. Domain Layer
Contains the core business entities, interfaces, and DTOs:
- **Entities**: Core domain models (`User`, `Order`)
- **Contracts**: Interfaces for services and repositories
- **DTOs**: Data Transfer Objects for API requests/responses

### 2. Service Layer
Application services implementing business logic:
- Orchestrates domain operations
- Implements domain service interfaces
- Handles use cases and business rules

**Local vs Remote Implementation:**
- **Local services** (`user_service.go`, `order_service.go`): Direct implementation for monolith deployment
- **Remote services** (`user_service_remote.go`, `order_service_remote.go`): HTTP proxy for microservices deployment
- Both implement the same domain interface, allowing transparent switching between monolith and microservices

### 3. Repository Layer
Infrastructure layer for data access:
- Implements repository interfaces
- Handles data persistence
- Database/storage abstractions

## 🚀 Running the Application

### Prerequisites
- Go 1.23 or higher
- Lokstra framework

### Start the Server

```bash
# From project root
go run ./project_templates/02_app_framework/01_medium_system

# Or from this directory
go run .
```

The server will start on `http://localhost:8080`

### Configuration
Edit `config.yaml` to configure:
- Server ports
- Middleware
- Service dependencies
- Repository connections

## 📡 API Endpoints

### Users
- `GET /api/users` - List all users
- `GET /api/users/{id}` - Get user by ID
- `POST /api/users` - Create new user
- `PUT /api/users/{id}` - Update user
- `DELETE /api/users/{id}` - Delete user

### Orders
- `GET /api/orders` - List all orders
- `GET /api/orders/{id}` - Get order by ID
- `GET /api/users/{user_id}/orders` - Get orders by user
- `POST /api/orders` - Create new order
- `PATCH /api/orders/{id}` - Update order status
- `DELETE /api/orders/{id}` - Delete order

## 🧪 Testing

Use the included `test.http` file with VS Code REST Client extension:

```http
### Create a new user
POST http://localhost:8080/api/users
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com"
}
```

## 🏛️ Design Principles

### Dependency Rule
Dependencies flow inward:
```
Repository → Domain ← Service
```
- Domain layer has no external dependencies
- Service depends on Domain
- Repository depends on Domain
- Domain defines interfaces, outer layers implement them

### Interface Segregation
Each domain defines its own interfaces:
- `UserService` interface for business operations
- `UserRepository` interface for data access
- DTOs for API contracts

### Dependency Injection
Services receive dependencies through factory functions:
```go
func UserServiceFactory(deps map[string]any, config map[string]any) any {
    return &UserServiceImpl{
        UserRepo: service.Cast[user.UserRepository](deps["user-repository"]),
    }
}
```

## 📦 Adding New Domains

To add a new domain (e.g., `product`):

### 1. Create Domain Layer
```bash
mkdir -p domain/product
```

Create files:
- `domain/product/entity.go` - Product model
- `domain/product/contract.go` - Interfaces
- `domain/product/dto.go` - Request/Response DTOs

### 2. Create Repository
```go
// repository/product_repository.go
type ProductRepositoryMemory struct { ... }
func NewProductRepositoryMemory(config map[string]any) *ProductRepositoryMemory { ... }
```

### 3. Create Service
```go
// service/product_service.go
type ProductServiceImpl struct { ... }
func ProductServiceFactory(deps map[string]any, config map[string]any) any { ... }
```

### 4. Register in register.go
```go
lokstra_registry.RegisterServiceType("product-repository-factory",
    repository.NewProductRepositoryMemory, nil)

lokstra_registry.RegisterServiceType("product-service-factory",
    service.ProductServiceFactory,
    service.ProductServiceRemoteFactory,
    deploy.WithResource("product", "products"),
    deploy.WithConvention("rest"),
)
```

### 5. Update config.yaml
Add product services and routes to the configuration.

## 🎓 When to Use This Template

### ✅ Good For:
- Medium-sized applications (2-10 entities)
- Monolithic deployments
- Single team projects
- Clear domain boundaries
- RESTful APIs

### ❌ Consider Enterprise Template For:
- 10+ domain entities
- Multiple teams
- Complex bounded contexts
- Microservices architecture
- Need for independent deployment per domain

## 🔄 Migration Path

This template can easily evolve:

**To Microservices:**

This template includes remote service implementations that enable microservices deployment:

1. **Local Service** (`user_service.go`): Direct implementation
   ```go
   type UserServiceImpl struct {
       UserRepo *service.Cached[user.UserRepository]
   }
   ```

2. **Remote Service** (`user_service_remote.go`): HTTP proxy
   ```go
   type UserServiceRemote struct {
       proxyService *proxy.Service
   }
   ```

Both implement the same interface, so switching is transparent:
```go
// Register both local and remote factories
lokstra_registry.RegisterServiceType("user-service-factory",
    service.UserServiceFactory,      // Local
    service.UserServiceRemoteFactory, // Remote
    deploy.WithResource("user", "users"),
    deploy.WithConvention("rest"),
)
```

**Deploy as Monolith:** All services use local implementations (current setup)

**Deploy as Microservices:** Split into separate servers with remote proxies:
- User Service: Runs user service locally, orders via HTTP
- Order Service: Runs order service locally, users via HTTP

See `register.go` for factory registration patterns.

**To Enterprise Template:**
- Reorganize into `modules/` structure
- Add domain events and messaging
- Implement CQRS patterns

## 📚 Additional Resources

- [Lokstra Documentation](../../docs)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Domain-Driven Design](https://martinfowler.com/bliki/DomainDrivenDesign.html)
- [Dependency Injection in Go](https://github.com/google/wire)

## 🤝 Contributing

When extending this template:
1. Keep domain layer pure (no external dependencies)
2. Use interfaces for all cross-layer communication
3. Follow naming conventions (entity.go, contract.go, dto.go)
4. Update tests and documentation
5. Maintain the dependency rule

## 📝 License

This template is part of the Lokstra framework.
