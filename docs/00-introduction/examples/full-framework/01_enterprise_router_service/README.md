# Lokstra Enterprise Router Service Template

**Annotation-Driven Router Services with Auto-Generated Code**

This template demonstrates how to use **Lokstra Annotations** to automatically generate router services, eliminating boilerplate code and streamlining development. This is the **recommended approach** for building enterprise applications with Lokstra.

---

## 📋 Table of Contents

- [What are Lokstra Annotations?](#what-are-lokstra-annotations)
- [When to Use This Template](#when-to-use-this-template)
- [Key Features](#key-features)
- [Architecture Overview](#architecture-overview)
- [Project Structure](#project-structure)
- [Lokstra Annotations Reference](#lokstra-annotations-reference)
- [How It Works](#how-it-works)
- [Getting Started](#getting-started)
- [Adding New Modules](#adding-new-modules)
- [Code Generation](#code-generation)
- [Deployment Strategies](#deployment-strategies)
- [Comparison with Manual Registration](#comparison-with-manual-registration)

---

## 🚀 What are Lokstra Annotations?

**Lokstra Annotations** are special comments in your Go code that automatically generate:
- ✅ Service factories
- ✅ Remote service proxies  
- ✅ Router configurations
- ✅ Dependency injection wiring
- ✅ Service registrations

**Write this:**
```go
// @EndpointService name="user-service", prefix="/api"
type UserServiceImpl struct {
    // @Inject "user-repository"
    UserRepo *service.Cached[domain.UserRepository]
}

// @Route "GET /users/{id}"
func (s *UserServiceImpl) GetByID(p *domain.GetUserRequest) (*domain.User, error) {
    return s.UserRepo.MustGet().GetByID(p.ID)
}
```

**Get this auto-generated:**
- Service factory function
- Remote HTTP proxy implementation
- Router registration with endpoints
- Dependency injection wiring
- Service type registration

**No more manual boilerplate!** 🎉

---

## 🎯 When to Use This Template

Use this template when you want:

- **✅ Rapid development** - Annotations eliminate boilerplate
- **✅ Type-safe DI** - Auto-detected dependencies  
- **✅ Auto-generated routers** - Routes from method signatures
- **✅ Microservice-ready** - Automatic remote proxies
- **✅ Clean code** - Business logic without framework noise
- **✅ Modular architecture** - Domain-Driven Design structure

**Don't use this template if:**
- You need full control over service registration
- Your project is extremely simple (< 3 services)
- You prefer explicit configuration over code generation

---

## ✨ Key Features

### 1. **Zero Boilerplate**
- No manual factory functions
- No router registration code
- No dependency wiring code
- Everything auto-generated from annotations

### 2. **Type-Safe Development**  
- Compile-time type checking
- IDE autocomplete for annotations
- Auto-detected dependency types
- Refactoring-friendly

### 3. **Hot Reload in Dev Mode**
- Automatic code regeneration on file changes
- No manual rebuild needed
- Instant feedback loop
- Works with debugger

### 4. **Production Ready**
- Generated code is readable and debuggable
- No runtime reflection overhead
- Cached and optimized
- Same performance as hand-written code

---

## 🏗 Architecture Overview

### Annotation-Driven Architecture

This template uses **three core annotations**:

| Annotation | Purpose | Example |
|------------|---------|---------|
| `@EndpointService` | Marks a service to be published as HTTP router | `@EndpointService name="user-service"` |
| `@Inject` | Auto-wires dependencies | `@Inject "user-repository"` |
| `@Route` | Maps methods to HTTP endpoints | `@Route "GET /users/{id}"` |

### Three-Layer Architecture per Module

```
modules/{module-name}/
├── domain/              # Business entities and interfaces
├── application/         # Service implementation with annotations
│   ├── user_service.go  # @EndpointService, @Inject, @Route
│   └── zz_generated.lokstra.go  # Auto-generated
└── infrastructure/      # Data access implementations
```

Each layer has specific responsibilities:

| Layer            | Responsibility                        | Annotations      |
|------------------|---------------------------------------|------------------|
| **Domain**       | Business entities and rules           | None             |
| **Application**  | Service with annotations              | `@EndpointService`, `@Inject`, `@Route` |
| **Infrastructure**| Repository implementations           | None             |

---

## 📁 Project Structure

```
01_enterprise_router_service/
├── modules/                       # Business modules
│   ├── user/                      # User management module
│   │   ├── domain/
│   │   │   ├── entity.go          # User entities
│   │   │   ├── service.go         # UserService interface
│   │   │   └── dto.go             # Request/Response types
│   │   ├── application/
│   │   │   ├── user_service.go    # ✨ With annotations
│   │   │   └── zz_generated.lokstra.go  # 🤖 Auto-generated
│   │   └── infrastructure/
│   │       └── repository/
│   │           └── user_repository.go
│   │
│   ├── order/                     # Order management module
│   │   └── ... (same structure)
│   │
│   └── shared/                    # Shared kernel
│
├── main.go                        # Entry point with lokstra.Bootstrap()
├── register.go                    # Module registration (minimal)
└── README.md
```

### Key Files

**main.go** - Application entry point:
```go
func main() {
    lokstra.Bootstrap()  // ✨ Magic happens here!
    
    registerServiceTypes()
    registerMiddlewareTypes()
    
    lokstra_registry.RunServerFromConfigFolder("config")
}
```

**user_service.go** - Annotated service:
```go
// @EndpointService name="user-service", prefix="/api"
type UserServiceImpl struct {
    // @Inject "user-repository"
    UserRepo *service.Cached[domain.UserRepository]
}

// @Route "GET /users/{id}"
func (s *UserServiceImpl) GetByID(p *domain.GetUserRequest) (*domain.User, error) {
    return s.UserRepo.MustGet().GetByID(p.ID)
}
```

**zz_generated.lokstra.go** - Auto-generated code:
```go
// AUTO-GENERATED - DO NOT EDIT
func init() {
    RegisterUserServiceImpl()  // Auto-registers on import
}

func UserServiceImplFactory(deps map[string]any, config map[string]any) any {
    return &UserServiceImpl{
        UserRepo: service.Cast[domain.UserRepository](deps["user-repository"]),
    }
}

func RegisterUserServiceImpl() {
    lokstra_registry.RegisterServiceType("user-service-factory",
        UserServiceImplFactory,
        UserServiceImplRemoteFactory,
        deploy.WithRouter(&deploy.ServiceTypeRouter{
            PathPrefix: "/api",
            CustomRoutes: map[string]string{
                "GetByID": "GET /users/{id}",
                // ... all routes auto-detected
            },
        }),
    )
}
```

**register.go** - Simple module loading:
```go
func registerServiceTypes() {
    user.Register()    // Triggers package init() -> auto-registration
    order.Register()   // Triggers package init() -> auto-registration
}
```

---

## 📝 Lokstra Annotations Reference

### @EndpointService

**Marks a struct as a router service** - generates factory, remote proxy, and router registration.

**Syntax:**
```go
// @EndpointService name="service-name", prefix="/api", middlewares=["recovery", "logger"]
type MyService struct { ... }
```

**Parameters:**
- `name` (required) - Service name for registration
- `prefix` (optional) - URL prefix for all routes (default: "")
- `middlewares` (optional) - Array of middleware names (default: [])

**Example:**
```go
// @EndpointService name="user-service", prefix="/api/v1", middlewares=["auth", "logging"]
type UserServiceImpl struct { ... }
```

**Generates:**
- `UserServiceImplFactory()` - Creates service instances
- `UserServiceImplRemote` - HTTP proxy implementation
- `RegisterUserServiceImpl()` - Auto-registration function
- Router configuration with all routes

---

### @Inject

**Auto-wires service dependencies** - generates dependency injection code.

**Syntax:**
```go
type MyService struct {
    // @Inject "dependency-service-name"
    DepField *service.Cached[InterfaceType]
}
```

**Parameters:**
- First string parameter - Service name to inject

**Example:**
```go
type UserServiceImpl struct {
    // @Inject "user-repository"
    UserRepo *service.Cached[domain.UserRepository]
    
    // @Inject "email-service"
    EmailSvc *service.Cached[domain.EmailService]
}
```

**Generates:**
```go
func UserServiceImplFactory(deps map[string]any, config map[string]any) any {
    return &UserServiceImpl{
        UserRepo: service.Cast[domain.UserRepository](deps["user-repository"]),
        EmailSvc: service.Cast[domain.EmailService](deps["email-service"]),
    }
}

// Auto-detected dependencies in registration
lokstra_registry.RegisterLazyService("user-service", "user-service-factory",
    map[string]any{"depends-on": []string{"user-repository", "email-service"}})
```

---

### @Route

**Maps a method to an HTTP endpoint** - generates route registration.

**Syntax:**
```go
// @Route "METHOD /path/{param}"
func (s *MyService) MethodName(p *RequestType) (*ResponseType, error) { ... }
```

**Supported HTTP Methods:**
- `GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `OPTIONS`, `HEAD`

**Path Parameters:**
- Use `{paramName}` for path variables
- Maps to fields in request struct

**Example:**
```go
// @Route "GET /users/{id}"
func (s *UserServiceImpl) GetByID(p *domain.GetUserRequest) (*domain.User, error) {
    return s.UserRepo.MustGet().GetByID(p.ID)
}

// @Route "POST /users"
func (s *UserServiceImpl) Create(p *domain.CreateUserRequest) (*domain.User, error) {
    return s.UserRepo.MustGet().Create(p)
}

// @Route "DELETE /users/{id}"
func (s *UserServiceImpl) Delete(p *domain.DeleteUserRequest) error {
    return s.UserRepo.MustGet().Delete(p.ID)
}
```

**Request Struct Binding:**
```go
type GetUserRequest struct {
    ID int `path:"id" validate:"required"`  // From URL path
}

type CreateUserRequest struct {
    Name  string `json:"name" validate:"required"`   // From JSON body
    Email string `json:"email" validate:"email"`
}

type ListUsersRequest struct {
    Page int `query:"page"`    // From query string
    Size int `query:"size"`
}
```

**Generated Route Map:**
```go
deploy.WithRouter(&deploy.ServiceTypeRouter{
    PathPrefix: "/api",
    CustomRoutes: map[string]string{
        "GetByID": "GET /users/{id}",
        "Create":  "POST /users",
        "Delete":  "DELETE /users/{id}",
    },
})
```

---

## 🔧 How It Works

### 1. Bootstrap Phase

When you call `lokstra.Bootstrap()` in `main()`:

```go
func main() {
    lokstra.Bootstrap()  // 🚀 Magic starts here
    // ... rest of your code
}
```

**Bootstrap does:**
1. **Detects run mode** - Production, Development, or Debug
2. **Checks for code changes** - Scans `.go` files for modifications
3. **Auto-generates code** - Creates `zz_generated.lokstra.go` files
4. **Relaunches if needed** - Restarts app to load new code

### 2. Annotation Processing

The annotation processor scans your code for:

```go
// @EndpointService name="user-service", prefix="/api"
type UserServiceImpl struct {
    // @Inject "user-repository"
    UserRepo *service.Cached[domain.UserRepository]
}

// @Route "GET /users/{id}"
func (s *UserServiceImpl) GetByID(...) { ... }
```

**Processor extracts:**
- Service metadata (name, prefix, middlewares)
- Dependencies from `@Inject` annotations
- Method signatures and return types
- Route definitions from `@Route` annotations

### 3. Code Generation

Generates `zz_generated.lokstra.go`:

```go
// AUTO-GENERATED CODE - DO NOT EDIT

package application

func init() {
    RegisterUserServiceImpl()  // Auto-registers on import
}

// Factory function
func UserServiceImplFactory(deps map[string]any, config map[string]any) any {
    return &UserServiceImpl{
        UserRepo: service.Cast[domain.UserRepository](deps["user-repository"]),
    }
}

// Remote HTTP proxy
type UserServiceImplRemote struct {
    proxyService *proxy.Service
}

func (s *UserServiceImplRemote) GetByID(p *domain.GetUserRequest) (*domain.User, error) {
    return proxy.CallWithData[*domain.User](s.proxyService, "GetByID", p)
}

// Registration function
func RegisterUserServiceImpl() {
    lokstra_registry.RegisterServiceType("user-service-factory",
        UserServiceImplFactory,
        UserServiceImplRemoteFactory,
        deploy.WithRouter(&deploy.ServiceTypeRouter{
            PathPrefix:  "/api",
            Middlewares: []string{"recovery", "request-logger"},
            CustomRoutes: map[string]string{
                "GetByID": "GET /users/{id}",
                // ... all routes
            },
        }),
    )
    
    lokstra_registry.RegisterLazyService("user-service",
        "user-service-factory",
        map[string]any{
            "depends-on": []string{"user-repository"},
        })
}
```

### 4. Auto-Registration via init()

When you import a module package:

```go
import (
    "github.com/.../modules/user"
)

func registerServiceTypes() {
    user.Register()  // Calls application.Register() -> triggers init()
}
```

The `init()` function in `zz_generated.lokstra.go` **automatically runs** and registers all services!

### 5. Cache System

**Lokstra caches annotation processing results** in `zz_cache.lokstra.json`:

```json
{
  "user_service.go": {
    "hash": "abc123...",
    "lastModified": "2025-11-11T10:30:00Z",
    "annotations": [...]
  }
}
```

**Benefits:**
- ✅ Only regenerates changed files
- ✅ Fast incremental builds
- ✅ Preserves code for unchanged files
- ✅ Minimal overhead in dev mode

### 6. Run Modes

| Mode | Detection | Behavior |
|------|-----------|----------|
| **Production** | Compiled binary | Skip autogen, use existing generated code |
| **Development** | `go run` | Auto-generate + relaunch with `go run` |
| **Debug** | Delve/VSCode debugger | Auto-generate + notify to restart debugger |

**Development Mode:**
```bash
$ go run .
[Lokstra] Environment detected: DEV
[Lokstra] Processing annotations...
[Lokstra] Code changed - relaunching...
# App restarts automatically with new code
```

**Debug Mode (VSCode):**
```
[Lokstra] Environment detected: DEBUG
[Lokstra] Processing annotations...

╔════════════════════════════════════════════════╗
║  AUTOGEN COMPLETED - DEBUGGER RESTART REQUIRED ║
╚════════════════════════════════════════════════╝

⚠️  Code generation detected changes.
⚠️  Please STOP and RESTART your debugger (F5)
```

---

## 🚀 Getting Started

### Prerequisites

- Go 1.23 or higher
- VS Code with REST Client extension (for test.http)

### Run the Application

```bash
# From project root
cd docs/00-introduction/examples/full-framework/01_enterprise_router_service

# Run in development mode (auto-reload)
go run .
```

**On first run, you'll see:**
```
[Lokstra] Environment detected: DEV
[Lokstra] Processing annotations...
Processing folder: .../modules/user/application
  - Updated: 1 files
  - Generated: zz_generated.lokstra.go
Processing folder: .../modules/order/application
  - Updated: 1 files
  - Generated: zz_generated.lokstra.go
[Lokstra] Relaunching with go run...

╔═══════════════════════════════════════════════╗
║   LOKSTRA ENTERPRISE ROUTER SERVICE           ║
║   Annotation-Driven Auto-Generated Routers    ║
╚═══════════════════════════════════════════════╝

Server starting on :3000
```

Server will start on `http://localhost:3000`

### Test the APIs

Open `test.http` in VS Code and click **"Send Request"** above each API call.

**User APIs:**
```http
### Get all users
GET http://localhost:3000/api/users

### Get user by ID
GET http://localhost:3000/api/users/1

### Create new user
POST http://localhost:3000/api/users
Content-Type: application/json

{
  "name": "Alice Johnson",
  "email": "alice@example.com",
  "role_id": 2
}

### Update user
PUT http://localhost:3000/api/users/1
Content-Type: application/json

{
  "name": "Alice Updated",
  "email": "alice.updated@example.com",
  "role_id": 2
}

### Suspend user
POST http://localhost:3000/api/users/1/suspend

### Delete user
DELETE http://localhost:3000/api/users/1
```

---

## ➕ Adding New Modules

### Step 1: Create Module Structure

```bash
mkdir -p modules/product/{domain,application,infrastructure/repository}
```

### Step 2: Define Domain Layer

**modules/product/domain/entity.go:**
```go
package domain

type Product struct {
    ID    int     `json:"id"`
    Name  string  `json:"name"`
    Price float64 `json:"price"`
}
```

**modules/product/domain/service.go:**
```go
package domain

type ProductService interface {
    GetByID(p *GetProductRequest) (*Product, error)
    List(p *ListProductsRequest) ([]*Product, error)
    Create(p *CreateProductRequest) (*Product, error)
}

type ProductRepository interface {
    GetByID(id int) (*Product, error)
    List() ([]*Product, error)
    Create(p *Product) (*Product, error)
}
```

**modules/product/domain/dto.go:**
```go
package domain

type GetProductRequest struct {
    ID int `path:"id" validate:"required"`
}

type ListProductsRequest struct {
    Category string `query:"category"`
}

type CreateProductRequest struct {
    Name  string  `json:"name" validate:"required"`
    Price float64 `json:"price" validate:"required,gt=0"`
}
```

### Step 3: Implement Application Layer with Annotations

**modules/product/application/product_service.go:**
```go
package application

import (
    "github.com/primadi/lokstra/core/service"
    "github.com/primadi/lokstra/.../modules/product/domain"
)

// @EndpointService name="product-service", prefix="/api", middlewares=["recovery", "request-logger"]
type ProductServiceImpl struct {
    // @Inject "product-repository"
    ProductRepo *service.Cached[domain.ProductRepository]
}

// Ensure implementation
var _ domain.ProductService = (*ProductServiceImpl)(nil)

// @Route "GET /products/{id}"
func (s *ProductServiceImpl) GetByID(p *domain.GetProductRequest) (*domain.Product, error) {
    return s.ProductRepo.MustGet().GetByID(p.ID)
}

// @Route "GET /products"
func (s *ProductServiceImpl) List(p *domain.ListProductsRequest) ([]*domain.Product, error) {
    return s.ProductRepo.MustGet().List()
}

// @Route "POST /products"
func (s *ProductServiceImpl) Create(p *domain.CreateProductRequest) (*domain.Product, error) {
    product := &domain.Product{
        Name:  p.Name,
        Price: p.Price,
    }
    return s.ProductRepo.MustGet().Create(product)
}

func Register() {
    // Empty function to trigger package load
}
```

### Step 4: Implement Infrastructure Layer

**modules/product/infrastructure/repository/product_repository.go:**
```go
package repository

import "github.com/primadi/lokstra/.../modules/product/domain"

type ProductRepositoryImpl struct {
    products map[int]*domain.Product
    nextID   int
}

func (r *ProductRepositoryImpl) GetByID(id int) (*domain.Product, error) {
    if p, exists := r.products[id]; exists {
        return p, nil
    }
    return nil, fmt.Errorf("product not found")
}

func (r *ProductRepositoryImpl) List() ([]*domain.Product, error) {
    result := make([]*domain.Product, 0, len(r.products))
    for _, p := range r.products {
        result = append(result, p)
    }
    return result, nil
}

func (r *ProductRepositoryImpl) Create(p *domain.Product) (*domain.Product, error) {
    r.nextID++
    p.ID = r.nextID
    r.products[p.ID] = p
    return p, nil
}

func ProductRepositoryFactory(deps map[string]any, config map[string]any) any {
    return &ProductRepositoryImpl{
        products: make(map[int]*domain.Product),
        nextID:   0,
    }
}
```

### Step 5: Create Module Registration

**modules/product/register.go:**
```go
package product

import (
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/.../modules/product/application"
    "github.com/primadi/lokstra/.../modules/product/infrastructure/repository"
)

func Register() {
    // Register repository
    lokstra_registry.RegisterServiceType("product-repository-factory",
        repository.ProductRepositoryFactory, nil)
    
    lokstra_registry.RegisterLazyService("product-repository",
        "product-repository-factory", nil)
    
    // Trigger auto-registration from annotations
    application.Register()
}
```

### Step 6: Register Module in Main

**register.go:**
```go
import (
    "github.com/primadi/lokstra/.../modules/user"
    "github.com/primadi/lokstra/.../modules/order"
    "github.com/primadi/lokstra/.../modules/product"  // Add this
)

func registerServiceTypes() {
    user.Register()
    order.Register()
    product.Register()  // Add this
}
```

### Step 7: Create Config (Optional)

**config/product.yaml:**
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
```

### Step 8: Run and Test!

```bash
go run .
```

**Bootstrap will automatically:**
1. ✅ Detect new `product_service.go` with annotations
2. ✅ Generate `zz_generated.lokstra.go` in product/application/
3. ✅ Register all routes, dependencies, and factories
4. ✅ Relaunch the app with new code

**Your product API is now live!**
```http
GET  /api/products
POST /api/products
GET  /api/products/{id}
```

**No manual registration needed!** 🎉

---

## 🔄 Code Generation

### Generated Files

Every folder with `@EndpointService` annotations gets:

**zz_generated.lokstra.go** - Generated Go code
```go
// AUTO-GENERATED CODE - DO NOT EDIT
package application

func init() { RegisterUserServiceImpl() }
func UserServiceImplFactory(...) { ... }
type UserServiceImplRemote struct { ... }
func RegisterUserServiceImpl() { ... }
```

**zz_cache.lokstra.json** - Cache metadata (gitignore this!)
```json
{
  "user_service.go": {
    "hash": "abc123...",
    "lastModified": "2025-11-11T10:30:00Z"
  }
}
```

### When Code is Regenerated

Code regenerates automatically when:

| Trigger | Action |
|---------|--------|
| **File modified** | Detected by hash + timestamp comparison |
| **Annotation changed** | Service name, routes, dependencies updated |
| **Method added/removed** | New routes auto-registered |
| **Dependency added** | Auto-wired in factory |

### Cache Behavior

**First run:**
```
Processing folder: modules/user/application
  - Updated: 1 files (user_service.go)
  - Generated: zz_generated.lokstra.go
```

**No changes:**
```
Processing folder: modules/user/application
  - Skipped: 1 files (no changes)
```

**File modified:**
```
Processing folder: modules/user/application
  - Updated: 1 files (user_service.go)
  - Regenerated: zz_generated.lokstra.go
```

### Manual Regeneration

Force regeneration (delete cache):
```bash
find . -name "zz_cache.lokstra.json" -delete
go run .
```

---

## 🚢 Deployment Strategies

### 1. Monolith (Current Setup)

**All modules in one server**

```yaml
# config/monolith.yaml
deployments:
  - name: api-server
    type: server
    port: 3000
    
    services:
      - name: user-service
        factory: user-service-factory
      - name: order-service
        factory: order-service-factory
```

```bash
go run . -config=config/monolith.yaml
```

**Pros:** Simple, low latency, easy development  
**Cons:** All modules must scale together

---

### 2. Microservices

**Each module as separate service**

**config/user-service.yaml:**
```yaml
deployments:
  - name: user-service
    type: server
    port: 3001
    
    services:
      - name: user-service
        factory: user-service-factory
      - name: user-repository
        factory: user-repository-factory
```

**config/order-service.yaml:**
```yaml
deployments:
  - name: order-service
    type: server
    port: 3002
    
    services:
      - name: order-service
        factory: order-service-factory
      - name: order-repository
        factory: order-repository-factory
```

Run each as separate process:
```bash
# Terminal 1
go run . -config=config/user-service.yaml

# Terminal 2
go run . -config=config/order-service.yaml
```

**Pros:** Independent scaling, deployment  
**Cons:** Network latency, distributed complexity

**Same code, different deployment!** ✨

---

## 📊 Comparison with Manual Registration

### Without Annotations (Manual)

**70+ lines of boilerplate per service:**

```go
// user_service.go
type UserServiceImpl struct {
    UserRepo *service.Cached[domain.UserRepository]
}

func (s *UserServiceImpl) GetByID(p *domain.GetUserRequest) (*domain.User, error) {
    return s.UserRepo.MustGet().GetByID(p.ID)
}

// ... all methods ...

// Manual factory
func UserServiceFactory(deps map[string]any, config map[string]any) any {
    return &UserServiceImpl{
        UserRepo: service.Cast[domain.UserRepository](deps["user-repository"]),
    }
}

// Manual remote proxy
type UserServiceRemote struct {
    proxyService *proxy.Service
}

func (s *UserServiceRemote) GetByID(p *domain.GetUserRequest) (*domain.User, error) {
    return proxy.CallWithData[*domain.User](s.proxyService, "GetByID", p)
}

// ... all proxy methods ...

func UserServiceRemoteFactory(deps, config map[string]any) any {
    proxyService := config["remote"].(*proxy.Service)
    return &UserServiceRemote{proxyService: proxyService}
}

// Manual registration
func init() {
    lokstra_registry.RegisterServiceType("user-service-factory",
        UserServiceFactory,
        UserServiceRemoteFactory,
        deploy.WithRouter(&deploy.ServiceTypeRouter{
            PathPrefix: "/api",
            CustomRoutes: map[string]string{
                "GetByID": "GET /users/{id}",
                "List":    "GET /users",
                "Create":  "POST /users",
                // ... manual route mapping
            },
        }),
    )
    
    lokstra_registry.RegisterLazyService("user-service",
        "user-service-factory",
        map[string]any{
            "depends-on": []string{"user-repository"},
        })
}
```

**Problems:**
- ❌ 70+ lines of boilerplate
- ❌ Error-prone manual route mapping
- ❌ Easy to forget dependency registration
- ❌ Remote proxy must match interface exactly
- ❌ Every method change requires 3 updates

---

### With Annotations (Auto-Generated)

**12 lines of business logic:**

```go
// @EndpointService name="user-service", prefix="/api"
type UserServiceImpl struct {
    // @Inject "user-repository"
    UserRepo *service.Cached[domain.UserRepository]
}

// @Route "GET /users/{id}"
func (s *UserServiceImpl) GetByID(p *domain.GetUserRequest) (*domain.User, error) {
    return s.UserRepo.MustGet().GetByID(p.ID)
}

// @Route "GET /users"
func (s *UserServiceImpl) List(p *domain.ListUsersRequest) ([]*domain.User, error) {
    return s.UserRepo.MustGet().List()
}

// @Route "POST /users"
func (s *UserServiceImpl) Create(p *domain.CreateUserRequest) (*domain.User, error) {
    // ... business logic
}

func Register() {}  // Trigger package load
```

**Benefits:**
- ✅ 12 lines vs 70+ lines
- ✅ No manual factory/proxy code
- ✅ Routes auto-detected from signatures
- ✅ Dependencies auto-wired
- ✅ Type-safe code generation
- ✅ Add method → auto-registered

### Comparison Table

| Aspect | Manual | With Annotations |
|--------|--------|------------------|
| **Lines of Code** | 70+ per service | 12 per service |
| **Factory Function** | Manual | Auto-generated |
| **Remote Proxy** | Manual implementation | Auto-generated |
| **Route Registration** | Manual mapping | Auto-detected |
| **Dependency Wiring** | Manual | Auto-detected |
| **Type Safety** | Manual sync required | Compiler-enforced |
| **Refactoring** | Update 3+ places | Update 1 place |
| **Error Prone** | High | Low |
| **Development Speed** | Slow | Fast |

**83% less code with annotations!** 🚀

---

## 🎓 Key Concepts

### Annotation-Driven Development

**Annotations are declarative metadata** that describe what you want, not how to implement it.

```go
// DECLARATIVE: What you want
// @EndpointService name="user-service", prefix="/api"
// @Route "GET /users/{id}"

// vs.

// IMPERATIVE: How to implement (manual)
router.GET("/api/users/:id", handler)
lokstra_registry.RegisterService(...)
```

### Code Generation vs. Reflection

**Lokstra uses code generation, not runtime reflection:**

| Approach | When | Performance | Type Safety | Debuggability |
|----------|------|-------------|-------------|---------------|
| **Reflection** | Runtime | Slower | Weak | Hard |
| **Code Generation** | Build time | Fast | Strong | Easy |

**Benefits:**
- ✅ Zero runtime overhead
- ✅ Full type checking at compile time
- ✅ Generated code is readable and debuggable
- ✅ No "magic" - you can see exactly what's generated

### Convention over Configuration

Lokstra follows **smart defaults** with **explicit overrides**:

```go
// Default convention (REST):
// Method name → Route
// GetByID   → GET /{resource}/{id}
// List      → GET /{resource}
// Create    → POST /{resource}

// Override when needed:
// @Route "POST /users/{id}/special-action"
func (s *UserService) SpecialAction(...) { ... }
```

---

## 📝 Best Practices

### 1. Keep Annotations Simple

```go
// ✅ Good - clear and concise
// @EndpointService name="user-service", prefix="/api"
type UserServiceImpl struct { ... }

// ❌ Bad - too complex
// @EndpointService name="user-service", prefix="/api/v1/internal/services", middlewares=["auth", "rbac", "logging", "metrics", "tracing"]
```

### 2. Use Descriptive Service Names

```go
// ✅ Good - follows naming convention
// @EndpointService name="user-service"

// ❌ Bad - inconsistent naming
// @EndpointService name="usrSvc"
```

### 3. Group Related Routes

```go
// ✅ Good - consistent prefix
// @EndpointService name="user-service", prefix="/api/v1/users"

// @Route "GET /{id}"
func GetByID(...) { ... }

// @Route "POST /"
func Create(...) { ... }
```

### 4. Document Business Logic

```go
// ✅ Good - document why, not what
// @Route "POST /users/{id}/suspend"
// Suspends user account - prevents login but preserves data
func (s *UserServiceImpl) Suspend(...) { ... }

// ❌ Bad - annotations already show what
// @Route "GET /users/{id}"
// Gets a user by ID
func (s *UserServiceImpl) GetByID(...) { ... }
```

### 5. Use Interface Assertion

```go
// ✅ Good - ensures implementation matches interface
var _ domain.UserService = (*UserServiceImpl)(nil)

// Compiler error if interface doesn't match!
```

---

## ⚠️ Important: Code Generation Before Build

### The Problem

**Auto-generation only happens during `go run`, not during `go build`!**

This means:
- ❌ Running `go build` directly **will fail** if generated files don't exist
- ❌ Running `go build` after code changes **will use old generated code**
- ✅ Must run `go run .` at least once before building
- ✅ Must run `go run .` after every annotation change

### Why This Happens

| Command | Run Mode | Code Generation |
|---------|----------|-----------------|
| `go run .` | Development | ✅ Auto-generates |
| `go build` | Production | ❌ Skips generation |
| `./compiled-binary` | Production | ❌ Skips generation |

**Bootstrap logic:**
```go
func Bootstrap() {
    Mode = detectRunMode()
    
    if Mode == RunModeProd {
        return  // Skip autogen in production!
    }
    
    // Only dev/debug modes auto-generate
    runAutoGen()
}
```

### Solutions to Avoid Forgetting

#### Solution 1: Use Build Scripts ⭐ **RECOMMENDED**

**Provided build scripts handle everything automatically!**

**Windows PowerShell:**
```powershell
.\build.ps1           # Generates + Builds for Windows
.\build.ps1 linux     # Generates + Builds for Linux
.\build.ps1 darwin    # Generates + Builds for macOS
```

**Windows CMD:**
```cmd
build.bat             # Generates + Builds for Windows
build.bat linux       # Generates + Builds for Linux
build.bat darwin      # Generates + Builds for macOS
```

**Linux/Mac:**
```bash
./build.sh            # Generates + Builds for current platform
./build.sh linux      # Generates + Builds for Linux
./build.sh windows    # Generates + Builds for Windows
./build.sh darwin     # Generates + Builds for macOS
```

**What the scripts do:**
1. ✅ Force code generation (`go run . --generate-only`)
2. ✅ Tidy dependencies (`go mod tidy`)
3. ✅ Build binary with platform-specific name
4. ✅ Cross-platform support (build for any OS from any OS)

**You never have to remember!** Just run the script. 🎉

---

#### Solution 2: Always Use `go run` in Development

**Recommended workflow:**
```bash
# During development - ALWAYS use go run
go run .

# Code autogenerates, then runs
# Edit code → Ctrl+C → go run . again
```

**Never use `go build` during development!**

---

#### Solution 3: Manual Generation Flag

**If you must build manually:**
```bash
# Step 1: Force generate (this is the key!)
go run . --generate-only

# Step 2: Build normally
go build -o app.exe .
```

The `--generate-only` flag:
- ✅ Forces rebuild of all generated code
- ✅ Deletes cache files automatically
- ✅ Exits after generation (doesn't run the app)
- ✅ Perfect for build scripts

---

#### Solution 4: CI/CD Pipeline

**GitHub Actions (.github/workflows/build.yml):**
```yaml
name: Build

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      
      # Use --generate-only flag
      - name: Generate Code
        run: go run . --generate-only
      
      - name: Tidy Dependencies
        run: go mod tidy
      
      - name: Build
        run: go build -v ./...
      
      - name: Test
        run: go test -v ./...
      
      # Build for multiple platforms
      - name: Build Releases
        run: |
          GOOS=linux GOARCH=amd64 go build -o dist/app-linux .
          GOOS=windows GOARCH=amd64 go build -o dist/app-windows.exe .
          GOOS=darwin GOARCH=amd64 go build -o dist/app-darwin .
```

---

### Recommended Workflow

**For Development:**
```bash
# 1. Edit your service files (add/change @Route, @Inject, etc.)
# 2. Run with auto-generation
go run .

# 3. Test your changes
curl http://localhost:3000/api/users

# 4. Repeat steps 1-3
```

**For Production Build:**

**Option A: Using Build Scripts (Recommended) ⭐**
```bash
# Windows (PowerShell)
.\build.ps1           # Build for Windows
.\build.ps1 linux     # Cross-compile for Linux
.\build.ps1 darwin    # Cross-compile for macOS

# Windows (CMD)
build.bat             # Build for Windows
build.bat linux       # Cross-compile for Linux
build.bat darwin      # Cross-compile for macOS

# Linux/Mac
./build.sh            # Build for current platform
./build.sh linux      # Build for Linux
./build.sh windows    # Cross-compile for Windows
./build.sh darwin     # Build for macOS
```

**What build scripts do:**
1. Run `go run . --generate-only` (force code generation)
2. Run `go mod tidy` (ensure dependencies)
3. Run `go build` (compile binary)
4. Create platform-specific binary:
   - Windows: `app-windows.exe`
   - Linux: `app-linux`
   - macOS: `app-darwin`

**Option B: Manual (Not Recommended)**
```bash
# Step 1: Force generate code
go run . --generate-only

# Step 2: Tidy dependencies (optional but recommended)
go mod tidy

# Step 3: Build
go build -o app.exe .

# For cross-platform:
GOOS=linux GOARCH=amd64 go build -o app-linux .
GOOS=windows GOARCH=amd64 go build -o app-windows.exe .
GOOS=darwin GOARCH=amd64 go build -o app-darwin .
```

**For CI/CD:**
```yaml
# GitHub Actions example
- name: Generate Code
  run: go run . --generate-only

- name: Tidy Dependencies
  run: go mod tidy

- name: Build
  run: go build -v ./...

- name: Test
  run: go test -v ./...
```

---

### Visual Workflow Reminder

```
┌─────────────────────────────────────────────┐
│  Development Mode                           │
│                                             │
│  Edit Code → go run . → Test → Repeat      │
│              ↑                              │
│              └─ Auto-generates here         │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│  Production Build (Recommended)             │
│                                             │
│  ./build.sh → Deploy binary                 │
│  ↑                                          │
│  └─ Generates + Builds in one step!        │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│  Production Build (Manual)                  │
│                                             │
│  go run . --generate-only → go build        │
│  ↑                                          │
│  └─ Force generation flag                  │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│  ❌ WRONG - Will Fail or Use Old Code!     │
│                                             │
│  Edit Code → go build → ERROR/STALE         │
│              ↑                              │
│              └─ No generation = problems!   │
└─────────────────────────────────────────────┘
```

### Available Build Scripts

The project includes cross-platform build scripts:

| Script | Platform | Description |
|--------|----------|-------------|
| `build.ps1` | Windows (PowerShell) | Recommended for Windows users |
| `build.bat` | Windows (CMD) | For legacy Windows environments |
| `build.sh` | Linux/macOS | Unix-based systems |

**All scripts support:**
- Current platform build
- Cross-compilation (Linux, Windows, macOS)
- Automatic code generation
- Dependency tidying

**Example usage:**
```bash
# Windows PowerShell
.\build.ps1           # Build for Windows
.\build.ps1 linux     # Cross-compile for Linux
.\build.ps1 darwin    # Cross-compile for macOS

# Windows CMD
build.bat windows     # Build for Windows
build.bat linux       # Cross-compile for Linux

# Linux/Mac
./build.sh            # Build for current platform
./build.sh windows    # Cross-compile for Windows
```

### The `--generate-only` Flag

**New in Lokstra!** Force code generation without running the app.

**Usage:**
```bash
go run . --generate-only
```

**What it does:**
1. ✅ Deletes all `zz_cache.lokstra.json` files
2. ✅ Forces rebuild of ALL generated code
3. ✅ Exits after generation (doesn't run the app)
4. ✅ Perfect for build scripts and CI/CD

**When to use:**
- Before building production binaries
- In CI/CD pipelines
- After major code refactoring
- When cache seems corrupted

**Example in build script:**
```bash
# Force generation before build
go run . --generate-only

# Now safe to build
go build -o app .
```

---

## 🔧 Troubleshooting

### Code Not Regenerating

**Problem:** Changed annotations but code not updating

**Solutions:**
```bash
# 1. Delete cache and rerun
find . -name "zz_cache.lokstra.json" -delete
go run .

# 2. Check run mode
[Lokstra] Environment detected: PROD  # Won't autogen in prod!

# 3. Force dev mode
export LOKSTRA_MODE=dev
go run .
```

### Debugger Not Stopping at Breakpoints

**Problem:** After code generation, debugger doesn't work

**Solution:**
```
[Lokstra] Environment detected: DEBUG
[Lokstra] Code changed - please RESTART debugger

→ Press Ctrl+C
→ Press F5 to restart debugger
```

### Import Cycle Errors

**Problem:** Circular imports between modules

**Solution:**
```
// ❌ Bad - circular dependency
modules/user → modules/order
modules/order → modules/user

// ✅ Good - shared domain
modules/user → modules/shared
modules/order → modules/shared
```

### Annotation Not Recognized

**Problem:** Annotation not being processed

**Checklist:**
- ✅ Correct syntax: `// @EndpointService` (with space after `//`)
- ✅ Correct parameter format: `name="value"`
- ✅ File saved before running
- ✅ Not in test file (`_test.go`)

---

## 📚 Learn More

- [Lokstra Documentation](https://primadi.github.io/lokstra/)
- [Annotation Processing Deep Dive](../../../03-api-reference/04-annotations/)
- [Service Layer Guide](../../../02-framework-guide/02-services/)
- [Dependency Injection Patterns](../../../02-framework-guide/03-dependency-injection/)

---

## 📄 License

This template is part of the Lokstra framework. See LICENSE file in project root.
