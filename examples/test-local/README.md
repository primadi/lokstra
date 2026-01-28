# Lokstra Enterprise Modular Template

**Domain-Driven Design with Bounded Contexts for Large-Scale Applications**

This template demonstrates how to structure a **large enterprise application** (10+ entities) using **Domain-Driven Design (DDD)** principles with **bounded contexts** as independent modules.

---

## 📋 Table of Contents

- [When to Use This Template](#when-to-use-this-template)
- [Architecture Overview](#architecture-overview)
- [Project Structure](#project-structure)
- [Module Structure](#module-structure)
- [Middleware](#middleware)
- [Configuration Strategy](#configuration-strategy)
- [Getting Started](#getting-started)
- [Adding New Modules](#adding-new-modules)
- [Deployment Strategies](#deployment-strategies)
- [Team Scalability](#team-scalability)
- [Comparison with Medium System](#comparison-with-medium-system)

---

## 🎯 When to Use This Template

Use this template when you have:

- **10+ domain entities** that need clear separation
- **Multiple teams** working on different parts of the system
- **Complex business logic** that benefits from bounded contexts
- **Need for microservices** in the future (without code changes)
- **Large-scale system** requiring modular architecture

**Don't use this template if:**
- You have < 10 entities → Use `01_medium_system` instead
- Simple CRUD operations only → Use `01_medium_system` or router templates
- Single team, simple domain → Overhead of modules is unnecessary

---

## 🏗 Architecture Overview

### Domain-Driven Design (DDD)

This template follows **DDD tactical patterns**:

- **Bounded Context**: Each module represents a distinct business capability
- **Ubiquitous Language**: Domain terminology consistent within each module
- **Domain Layer**: Pure business logic, no infrastructure concerns
- **Application Layer**: Use case orchestration
- **Infrastructure Layer**: Data access, external services

### Three-Layer Architecture per Module

```
modules/{module-name}/
├── domain/           # Core business logic (entities, interfaces)
├── application/      # Use case implementations (services)
└── infrastructure/   # Technical implementations (repositories, clients)
```

Each layer has specific responsibilities:

| Layer            | Responsibility                        | Dependencies     |
|------------------|---------------------------------------|------------------|
| **Domain**       | Business entities and rules           | None (pure)      |
| **Application**  | Use case orchestration                | Domain only      |
| **Infrastructure**| Data access, external services       | Domain only      |

---

## 📁 Project Structure

```
enterprise_modular/
├── modules/                    # Bounded contexts (business capabilities)
│   ├── user/                   # User management context
│   │   ├── domain/
│   │   │   ├── entity.go       # User, UserProfile entities
│   │   │   ├── service.go      # UserService interface
│   │   │   └── dto.go          # Request/Response contracts
│   │   ├── application/
│   │   │   └── user_service.go # UserService implementation
│   │   └── infrastructure/
│   │       └── repository/
│   │           └── user_repository.go
│   │
│   ├── order/                  # Order management context
│   │   ├── domain/
│   │   │   ├── entity.go       # Order, OrderItem entities
│   │   │   ├── service.go      # OrderService interface
│   │   │   └── dto.go          # Request/Response contracts
│   │   ├── application/
│   │   │   └── order_service.go
│   │   └── infrastructure/
│   │       └── repository/
│   │           └── order_repository.go
│   │
│   └── shared/                 # Shared kernel (common value objects)
│       └── domain/
│
├── config/                     # Per-module deployment configs
│   ├── user.yaml              # User module endpoints
│   └── order.yaml             # Order module endpoints
│
├── main.go                     # Application entry point
├── register.go                 # Module registration
├── test.http                   # API testing file
└── README.md                   # This file
```

### Key Principles

1. **Module Independence**: Each module can be developed, tested, deployed independently
2. **Clear Boundaries**: Modules communicate through interfaces, not direct calls
3. **Configuration-Driven**: Each module has its own config file
4. **Portability**: Copy module folder + config = portable module

---

## 🧩 Module Structure

### Domain Layer (`domain/`)

**Purpose**: Define business entities and contracts

```go
// entity.go - Business entities
type User struct {
    ID     int    `json:"id"`
    Name   string `json:"name"`
    Email  string `json:"email"`
    Status string `json:"status"`
}

// service.go - Business operations interface
type UserService interface {
    GetByID(p *GetUserRequest) (*User, error)
    Create(p *CreateUserRequest) (*User, error)
}

// dto.go - Request/Response contracts
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}
```

**Rules**:
- ✅ Pure business logic, no framework dependencies
- ✅ Defines interfaces, not implementations
- ❌ No database, HTTP, or external service code
- ❌ No dependencies on other modules

### Application Layer (`application/`)

**Purpose**: Implement use cases (business workflows)

```go
type UserServiceImpl struct {
    UserRepo *service.Cached[domain.UserRepository]
}

func (s *UserServiceImpl) Create(p *domain.CreateUserRequest) (*domain.User, error) {
    u := &domain.User{
        Name:   p.Name,
        Email:  p.Email,
        Status: "active",
    }
    return s.UserRepo.MustGet().Create(u)
}
```

**Rules**:
- ✅ Orchestrates business workflows
- ✅ Depends on domain interfaces only
- ❌ No direct database access (uses repositories)
- ❌ No HTTP handling (Lokstra handles it)

### Infrastructure Layer (`infrastructure/`)

**Purpose**: Technical implementations (databases, external APIs)

```go
type UserRepositoryImpl struct {
    users map[int]*domain.User
}

func (r *UserRepositoryImpl) Create(user *domain.User) (*domain.User, error) {
    // Database/storage implementation
    user.ID = r.nextID
    r.users[user.ID] = user
    return user, nil
}
```

**Rules**:
- ✅ Implements domain interfaces
- ✅ Contains database, cache, external API code
- ❌ Does not contain business logic
- ❌ Only accessed through interfaces

---

## 🔒 Middleware

This template includes custom middleware for authentication and logging.

### Built-in Middlewares

1. **`recovery`** - Panic recovery (from lokstra/middleware)
2. **`request-logger`** - HTTP request/response logging
3. **`simple-auth`** - Simple Bearer token authentication
4. **`mw-test`** - Example parameterized middleware

### Simple Auth Middleware

The `simple-auth` middleware provides basic Bearer token authentication:

**How it works:**
- Checks for `Authorization: Bearer <token>` header
- Validates token format: must start with `demo-`
- Extracts user ID from token (e.g., `demo-user123` → user ID: `user123`)
- Repositorys user info in request context

**Usage in @Handler:**
```go
// @Handler name="user-service", prefix="/api", middlewares=["recovery", "request-logger", "simple-auth"]
type UserServiceImpl struct {
    UserRepo domain.UserRepository
}
```

**Testing with HTTP client:**
```http
### Authenticated request
GET http://localhost:3000/api/users
Authorization: Bearer demo-user123

### Unauthenticated request (will fail with 401)
GET http://localhost:3000/api/users
```

**Valid tokens:**
- `demo-user123` ✅
- `demo-admin` ✅
- `demo-john.doe` ✅
- `invalid-token` ❌ (doesn't start with `demo-`)
- `` ❌ (missing header)

**Implementation** (`register.go`):
```go
func simpleAuthFactory(config map[string]any) request.HandlerFunc {
    return func(ctx *request.Context) error {
        authHeader := ctx.R.Header.Get("Authorization")
        if authHeader == "" {
            return ctx.Api.Unauthorized("Missing Authorization header")
        }
        
        // Extract Bearer token
        token := strings.TrimPrefix(authHeader, "Bearer ")
        if !strings.HasPrefix(token, "demo-") {
            return ctx.Api.Unauthorized("Invalid token")
        }
        
        // Repository user info in context
        userID := token[5:]
        ctx.Set("user_id", userID)
        ctx.Set("authenticated", true)
        
        return ctx.Next()
    }
}
```

### Adding Custom Middleware

**Step 1:** Register middleware factory in `register.go`:
```go
func registerMiddlewareTypes() {
    lokstra_registry.RegisterMiddlewareFactory("my-middleware", func(config map[string]any) request.HandlerFunc {
        return func(ctx *request.Context) error {
            // Before request
            log.Println("Before:", ctx.R.URL.Path)
            
            // Process request
            err := ctx.Next()
            
            // After request
            log.Println("After:", ctx.Resp.RespStatusCode)
            
            return err
        }
    })
}
```

**Step 2:** Use in `@Handler` annotation:
```go
// @Handler name="my-service", middlewares=["recovery", "my-middleware"]
type MyService struct {}
```

**Step 3:** Or use per-route:
```go
// @Route "POST /sensitive", middlewares=["auth", "admin-only"]
func (s *MyService) SensitiveAction() error {
    return nil
}
```

---

## ⚙️ Configuration Strategy

### Per-Module Configuration

Each module has its own YAML file in `config/`:

**config/user.yaml**:
```yaml
deployments:
  - name: api-server      # Deployment name
    type: server
    port: 3000
    
    services:
      - name: user-service
        factory: user-service-factory
        endpoints:
          - path: /api/users
            method: GET
            handler: List
```

**config/order.yaml**:
```yaml
deployments:
  - name: api-server      # Same name = merges with user.yaml
    type: server
    port: 3000
    
    services:
      - name: order-service
        factory: order-service-factory
        endpoints:
          - path: /api/orders
            method: GET
            handler: List
```

**Lokstra automatically merges** all YAML files with the same deployment name!

### Benefits

1. **Modularity**: Each team edits their own YAML
2. **Portability**: Move module + config = portable
3. **Flexibility**: Easy to split into microservices later
4. **No Conflicts**: Different files, no merge conflicts

---

## 🚀 Getting Started

### Prerequisites

- Go 1.23 or higher
- VS Code with REST Client extension (for test.http)

### Run the Application

```bash
# From project root (where go.mod is)
go run ./project_templates/02_app_framework/02_enterprise_modular
```

Server will start on `http://localhost:3000`

### Test the APIs

Open `test.http` in VS Code and click **"Send Request"** above each API call.

**User APIs**:
- `GET /api/users` - List all users
- `GET /api/users/{id}` - Get user by ID
- `POST /api/users` - Create new user
- `PUT /api/users/{id}` - Update user
- `POST /api/users/{id}/suspend` - Suspend user
- `POST /api/users/{id}/activate` - Activate user
- `DELETE /api/users/{id}` - Delete user

**Order APIs**:
- `GET /api/orders` - List all orders
- `GET /api/orders?user_id=2` - Get orders by user
- `GET /api/orders/{id}` - Get order by ID
- `POST /api/orders` - Create new order
- `PUT /api/orders/{id}/status` - Update order status
- `POST /api/orders/{id}/cancel` - Cancel order
- `DELETE /api/orders/{id}` - Delete order

---

## ➕ Adding New Modules

### Step 1: Create Module Structure

```bash
mkdir -p modules/product/{domain,application,infrastructure/repository}
```

### Step 2: Define Domain Layer

**modules/product/domain/entity.go**:
```go
package domain

type Product struct {
    ID    int     `json:"id"`
    Name  string  `json:"name"`
    Price float64 `json:"price"`
}
```

**modules/product/domain/service.go**:
```go
package domain

type ProductService interface {
    GetByID(p *GetProductRequest) (*Product, error)
    List(p *ListProductsRequest) ([]*Product, error)
}

type ProductRepository interface {
    GetByID(id int) (*Product, error)
    List() ([]*Product, error)
}
```

**modules/product/domain/dto.go**:
```go
package domain

type GetProductRequest struct {
    ID int `path:"id" validate:"required"`
}

type ListProductsRequest struct {
    Category string `query:"category"`
}
```

### Step 3: Implement Application Layer

**modules/product/application/product_service.go**:
```go
package application

import (
    "github.com/primadi/lokstra/core/service"
    "github.com/primadi/lokstra/project_templates/.../modules/product/domain"
)

type ProductServiceImpl struct {
    ProductRepo *service.Cached[domain.ProductRepository]
}

func (s *ProductServiceImpl) GetByID(p *domain.GetProductRequest) (*domain.Product, error) {
    return s.ProductRepo.MustGet().GetByID(p.ID)
}

func ProductServiceFactory(deps map[string]any, config map[string]any) any {
    return &ProductServiceImpl{
        ProductRepo: service.Cast[domain.ProductRepository](deps["product-repository"]),
    }
}
```

### Step 4: Implement Infrastructure Layer

**modules/product/infrastructure/repository/product_repository.go**:
```go
package repository

import "github.com/primadi/lokstra/project_templates/.../modules/product/domain"

type ProductRepositoryImpl struct {
    products map[int]*domain.Product
}

func (r *ProductRepositoryImpl) GetByID(id int) (*domain.Product, error) {
    // Implementation
}

func ProductRepositoryFactory(deps map[string]any, config map[string]any) any {
    return &ProductRepositoryImpl{
        products: make(map[int]*domain.Product),
    }
}
```

### Step 5: Register Module

**register.go**:
```go
import (
    productApp "github.com/primadi/lokstra/.../modules/product/application"
    productRepo "github.com/primadi/lokstra/.../modules/product/infrastructure/repository"
)

func registerServiceTypes() {
    // ... existing registrations ...
    
    // ==================== PRODUCT MODULE ====================
    lokstra_registry.RegisterServiceType("product-repository-factory",
        productRepo.ProductRepositoryFactory, nil)
    
    lokstra_registry.RegisterServiceType("product-service-factory",
        productApp.ProductServiceFactory, nil,
        deploy.WithResource("product", "products"),
        deploy.WithConvention("rest"),
    )
}
```

### Step 6: Create Module Config

**config/product.yaml**:
```yaml
deployments:
  - name: api-server
    type: server
    port: 3000
    
    services:
      - name: product-repository
        factory: product-repository-factory
      
      - name: product-service
        factory: product-service-factory
        dependencies:
          product-repository: product-repository
        endpoints:
          - path: /api/products/{id}
            method: GET
            handler: GetByID
          - path: /api/products
            method: GET
            handler: List
```

### Step 7: Done!

Restart the server. Product module is now deployed alongside user and order modules.

---

## 🚢 Deployment Strategies

### 1. Monolith (Current Setup)

**All modules in one server**

```yaml
# All configs use same deployment name
deployments:
  - name: api-server
    port: 3000
```

**Pros**: Simple, low latency, easy development  
**Cons**: All modules must scale together

---

### 2. Microservices

**Each module as separate service**

**config/user.yaml** (runs on port 3001):
```yaml
deployments:
  - name: user-service
    type: server
    port: 3001
    services:
      - name: user-service
        factory: user-service-factory
```

**config/order.yaml** (runs on port 3002):
```yaml
deployments:
  - name: order-service
    type: server
    port: 3002
    services:
      - name: order-service
        factory: order-service-factory
```

Run each as separate process:
```bash
go run . config/user.yaml    # User service on :3001
go run . config/order.yaml   # Order service on :3002
```

**Pros**: Independent scaling, deployment, technology  
**Cons**: Network latency, distributed complexity

---

### 3. Hybrid

**Some modules together, some separate**

Group related modules:
- Auth + User modules → auth-server (port 3001)
- Order + Payment modules → order-server (port 3002)
- Product + Inventory modules → catalog-server (port 3003)

**Pros**: Balance between monolith and microservices  
**Cons**: Need to decide which modules to group

---

## 👥 Team Scalability

### Scenario: 3 Teams

**Team A**: User Authentication  
**Team B**: Order Management  
**Team C**: Product Catalog

### How They Work Independently

1. **Code Isolation**: Each team owns their module folder
   ```
   Team A → modules/user/
   Team B → modules/order/
   Team C → modules/product/
   ```

2. **Config Isolation**: Each team owns their YAML
   ```
   Team A → config/user.yaml
   Team B → config/order.yaml
   Team C → config/product.yaml
   ```

3. **Minimal Coordination**:
   - Teams commit to their own folders
   - No merge conflicts (different files)
   - Register in `register.go` only when ready

4. **Independent Testing**:
   ```bash
   # Team A tests their module
   go test ./modules/user/...
   
   # Team B tests their module
   go test ./modules/order/...
   ```

5. **Independent Deployment** (if microservices):
   ```bash
   # Team A deploys user service
   go run . config/user.yaml
   
   # Team B deploys order service
   go run . config/order.yaml
   ```

---

## 📊 Comparison with Medium System

| Aspect              | Medium System (Flat)       | Enterprise Modular (DDD)    |
|---------------------|---------------------------|-----------------------------|
| **Best For**        | 2-10 entities             | 10+ entities                |
| **Structure**       | domain/ + service/ + repo/| modules/{bounded-context}/  |
| **Team Size**       | Single team               | Multiple teams              |
| **Complexity**      | Low                       | High                        |
| **Portability**     | Moderate                  | High (copy module folder)   |
| **Microservices**   | Harder to split           | Easy to split               |
| **Configuration**   | Single config.yaml        | Per-module YAML in config/  |
| **Learning Curve**  | Easy                      | Requires DDD understanding  |

**When to Migrate**:
- Growing from 10 to 20+ entities
- Need to split teams
- Planning microservices architecture
- Domain complexity requires bounded contexts

---

## 🎓 Key Concepts

### Bounded Context

A **bounded context** is a logical boundary around a business capability where:
- Specific terms have specific meanings
- Models are consistent within the boundary
- Different contexts can have different models for the same concept

**Example**: "User" means different things in different contexts:
- **User Module**: Authentication, profile management
- **Order Module**: Just a reference (user_id) to who placed order
- **Analytics Module**: Aggregated statistics

### Domain-Driven Design Benefits

1. **Ubiquitous Language**: Business and developers speak same language
2. **Modularity**: Clear boundaries, independent evolution
3. **Scalability**: Easy to add new bounded contexts
4. **Flexibility**: Change one context without affecting others

---

## 📝 Best Practices

1. **Keep Domain Pure**: No framework dependencies in `domain/`
2. **Interface Segregation**: Small, focused interfaces
3. **Dependency Direction**: Always point toward domain
4. **Module Independence**: Avoid cross-module dependencies
5. **Config per Module**: One YAML file per bounded context
6. **Consistent Naming**: Use same terms in code, config, API

---

## 🔧 Troubleshooting

### Module not found error

Make sure you're running from the project root:
```bash
cd /path/to/lokstra-project-root
go run ./project_templates/02_app_framework/02_enterprise_modular
```

### Config not loading

Lokstra looks for YAML files in the path you specify:
```go
lokstra_registry.RunServerFromConfig("config")  // looks in ./config/
```

### Endpoints not registered

Check `register.go` - all factories must be registered:
```go
lokstra_registry.RegisterServiceType("user-service-factory", ...)
```

---

## 📚 Learn More

- [Lokstra Documentation](https://primadi.github.io/lokstra/)
- [Domain-Driven Design](https://martinfowler.com/bliki/DomainDrivenDesign.html)
- [Bounded Context Pattern](https://martinfowler.com/bliki/BoundedContext.html)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)

---

## 📄 License

This template is part of the Lokstra framework. See LICENSE file in project root.
