# Service as Router - Lokstra's Unique Feature

⏱️ **Estimated time**: 20 minutes

## 🎯 What You'll Learn

This example demonstrates **Service as Router** - Lokstra's most unique feature that differentiates it from other Go frameworks. Instead of manually creating route handlers, Lokstra can **automatically generate REST endpoints** from your service methods!

**Key Concepts:**
- Auto-generating routers from service methods
- Metadata via `RegisterServiceType` options (no Remote struct needed!)
- Convention-based routing (REST, RPC, etc.)
- Comparison: Manual vs Auto-generated routing
- Clean file structure for complex examples

## 📁 File Structure

This example uses a **proper file structure** (not everything in one file):

```
04-service-as-router/
├── main.go                 # Bootstrap & router creation
├── model/
│   └── model.go           # Domain models (User, Product)
├── contract/
│   └── contract.go        # Request/Response params
├── service/
│   ├── user_service.go    # User business logic
│   └── product_service.go # Product business logic
├── README.md
└── test.http
```

**Why separate files?**
- ✅ Clear separation of concerns
- ✅ Easy to navigate
- ✅ Scalable for real projects
- ✅ Mirrors production code structure

**Note:** For simple examples (like Example 01), one file is fine. For complex features like Service as Router, proper structure helps understanding.

## 🌟 Why This Matters

**Traditional Go Frameworks** (Gin, Echo, Chi):
```go
// Manual routing - tedious and error-prone!
router.GET("/users", handleListUsers)
router.GET("/users/:id", handleGetUser)
router.POST("/users", handleCreateUser)
router.PUT("/users/:id", handleUpdateUser)
router.DELETE("/users/:id", handleDeleteUser)

// Then you need handler functions...
func handleListUsers(c *gin.Context) {
    users, err := userService.List(...)
    // ... response handling
}
// ... 4 more handlers
```

**Lokstra - Service as Router**:
```go
// 1. Define service methods (service/user_service.go)
type UserService struct {
    users []User
}

func (s *UserService) List(params *ListParams) ([]User, error) { ... }
func (s *UserService) GetByID(params *GetParams) (*User, error) { ... }
// Add Create, Update, Delete as needed

// 2. Register with metadata (main.go)
lokstra_registry.RegisterServiceType(
    "user-service",
    service.NewUserService,
    nil,  // No remote needed for simple examples
    deploy.WithResource("user", "users"),
    deploy.WithConvention("rest"),
)

// 3. Auto-generate router with ONE LINE!
userSvc := service.NewUserService()
router := lokstra_registry.NewRouterFromServiceType("user-service", userSvc)

// Done! ALL REST endpoints created automatically:
//   GET    /users       → List()
//   GET    /users/{id}  → GetByID()
//   POST   /users       → Create()  (if method exists)
//   PUT    /users/{id}  → Update()  (if method exists)
//   DELETE /users/{id}  → Delete()  (if method exists)
```

**Benefits:**
- ✅ **Write once, use everywhere**: Same service works locally AND remotely
- ✅ **Zero boilerplate**: No manual handler registration
- ✅ **Convention over configuration**: REST patterns auto-applied
- ✅ **Type-safe**: Go methods with strong typing
- ✅ **Microservice-ready**: Same code works in monolith or distributed

## 📋 Code Walkthrough

### Step 1: Define Your Service Methods

**File: `service/user_service.go`**

```go
package service

import "github.com/primadi/lokstra/docs/.../model"
import "github.com/primadi/lokstra/docs/.../contract"

type UserService struct {
    users []model.User
}

func NewUserService() *UserService {
    return &UserService{
        users: []model.User{
            {ID: 1, Name: "Alice", Email: "alice@example.com"},
            // ... more data
        },
    }
}

// Method signature determines the REST endpoint!
// List(params) → GET /users?role=...
func (s *UserService) List(p *contract.ListUsersParams) ([]model.User, error) {
    return s.users, nil
}

// GetByID(params) → GET /users/{id}
func (s *UserService) GetByID(p *contract.GetUserParams) (*model.User, error) {
    for _, user := range s.users {
        if user.ID == p.ID {
            return &user, nil
        }
    }
    return nil, fmt.Errorf("user not found")
}
```

**File: `contract/contract.go`** (Request Parameters)

```go
package contract

type ListUsersParams struct {
    Role string `query:"role"`  // From query string
}

type GetUserParams struct {
    ID int `path:"id"`  // From URL path
}
```

**File: `model/model.go`** (Domain Models)

```go
package model

type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}
```

### Step 2: Register Service with Metadata

**File: `main.go`**

Instead of creating a separate Remote struct, provide metadata directly via options:

```go
import (
    "github.com/primadi/lokstra/core/deploy"
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/docs/.../service"
)

// Register service type with metadata
lokstra_registry.RegisterServiceType(
    "user-service",
    func(deps, cfg map[string]any) any {
        return service.NewUserService()
    },
    nil,  // No remote factory needed for this example
    deploy.WithResource("user", "users"),      // Metadata!
    deploy.WithConvention("rest"),             // REST convention
)

// Create and register instance
userSvc := service.NewUserService()
lokstra_registry.RegisterService("user-service", userSvc)
```

**Metadata Options Available:**
- `WithResource(singular, plural)` - Resource names
- `WithConvention(name)` - "rest", "rpc", "graphql"
- `WithPathPrefix(prefix)` - e.g., "/api/v1"
- `WithHiddenMethods(methods...)` - Don't expose certain methods
- `WithRouteOverride(method, path)` - Custom path for method
- `WithMiddlewares(names...)` - Apply middlewares

**Why This is Better:**
- ✅ No separate Remote struct needed for simple examples
- ✅ Metadata in one place (RegisterServiceType)
- ✅ Cleaner code, less boilerplate
- ✅ Still supports Remote struct for complex cases (see 04-multi-deployment)

### Step 3: Auto-Generate Router

Still in `main.go`:

```go
// AUTO-GENERATE router from service + metadata!
// Metadata comes from RegisterServiceType options above
autoUserRouter := lokstra_registry.NewRouterFromServiceType(
    "user-service",  // Service type name
    userSvc,         // Service instance
)

// Done! Router has ALL endpoints automatically:
// GET /users       → List()
// GET /users/{id}  → GetByID()
```

**That's it!** One function call generates the entire router.

Compare with traditional approach:
```go
// Traditional (manual) - tedious!
manualRouter := lokstra.NewRouter("manual-api")
manualRouter.GET("/manual/users", func() ([]User, error) {
    return userService.List(&ListUsersParams{})
})
manualRouter.GET("/manual/users/{id}", func(p *GetUserParams) (*User, error) {
    return userService.GetByID(p)
})
// ... repeat for every endpoint
```

### Step 4: Create App and Run

```go
app := lokstra.NewApp("service-as-router-demo", ":3000",
    autoUserRouter,
    autoProductRouter,
)

app.Run(30 * time.Second)
```

That's it! Your API is ready with auto-generated endpoints.

## 🔧 How It Works

### REST Convention Mapping

The REST convention automatically maps method names to HTTP methods and paths:

| Service Method | HTTP Method | Path | Description |
|---------------|-------------|------|-------------|
| `List()` | GET | `/users` | List all resources |
| `GetByID()` | GET | `/users/{id}` | Get single resource |
| `Create()` | POST | `/users` | Create new resource |
| `Update()` | PUT | `/users/{id}` | Update resource |
| `Delete()` | DELETE | `/users/{id}` | Delete resource |
| `Patch()` | PATCH | `/users/{id}` | Partial update |

### Parameter Binding

Lokstra automatically binds request data to your parameter structs:

```go
type GetUserParams struct {
    ID int `path:"id"`  // From URL: /users/123
}

type ListUsersParams struct {
    Role   string `query:"role"`      // From query: ?role=admin
    Page   int    `query:"page"`      // ?page=2
    Limit  int    `query:"limit"`     // ?limit=10
}

type CreateUserParams struct {
    Name  string `json:"name"`   // From JSON body
    Email string `json:"email"`  // From JSON body
}
```

### Service Reusability

The **SAME service** can be used in multiple ways:

1. **Local Service**: Direct method calls
   ```go
   users, err := userService.List(params)
   ```

2. **Auto-Generated Router**: REST API endpoints
   ```go
   router := lokstra_registry.NewRouterFromServiceType("user-service", userSvc)
   app := lokstra.NewApp("api", ":3000", router)
   ```

3. **Remote Service** (for microservices): HTTP proxy calls
   ```go
   // In another microservice - use Remote struct with ProxyService
   remoteUser := NewUserServiceRemote(proxyService)
   users, err := remoteUser.List(params)  // HTTP call via proxy!
   
   // See 04-multi-deployment example for full implementation
   ```

## 🎨 Advanced: Custom Metadata Options

You can customize auto-generated routes using metadata options:

```go
lokstra_registry.RegisterServiceType(
    "user-service",
    service.NewUserService,
    nil,
    deploy.WithResource("user", "users"),
    deploy.WithConvention("rest"),
    deploy.WithPathPrefix("/api/v1"),           // All routes under /api/v1
    deploy.WithHiddenMethods("InternalMethod"), // Don't expose this
    deploy.WithRouteOverride("Search", "/users/search"), // Custom path
    deploy.WithMiddlewares("auth", "logging"),  // Apply middlewares
)
```

**Available Options:**
- `WithResource(singular, plural)` - Resource names (required)
- `WithConvention(name)` - "rest" (default), "rpc", "graphql"
- `WithPathPrefix(prefix)` - Prefix all routes
- `WithHiddenMethods(methods...)` - Hide methods from router
- `WithRouteOverride(method, path)` - Custom path for method
- `WithMiddlewares(names...)` - Apply middleware to all routes

## 🚀 Running the Example

1. **Start the server:**
   ```bash
   go run main.go
   ```

2. **Test with curl or REST Client:**
   ```bash
   # Manual endpoints
   curl http://localhost:3000/manual/users
   curl http://localhost:3000/manual/users/1

   # Auto-generated endpoints
   curl http://localhost:3000/users
   curl http://localhost:3000/users/1
   curl http://localhost:3000/products
   curl http://localhost:3000/products/1
   ```

3. **Use test.http** for all endpoints

## 📊 Output Analysis

When you run the example, you'll see:

```
========================================
Service as Router Example
========================================

🚀 Generating routers from services...

✅ Manual router created (2 endpoints manually defined):
   GET /manual/users
   GET /manual/users/{id}

✅ Auto-generated router for user-service (ZERO manual routing!):
   GET /users       → List() method
   GET /users/{id}  → GetByID() method

✅ Auto-generated router for product-service (ZERO manual routing!):
   GET /products       → List() method
   GET /products/{id}  → GetByID() method

========================================
Server: http://localhost:3000
========================================

Manual Endpoints:
  GET /manual/users       - List users (manual)
  GET /manual/users/{id}  - Get user (manual)

Auto-Generated Endpoints:
  GET /users              - List users (auto)
  GET /users/{id}         - Get user (auto)
  GET /products           - List products (auto)
  GET /products/{id}      - Get product (auto)

🎯 Compare manual vs auto-generated!
```

## 💡 Key Takeaways

1. **Service as Router** is Lokstra's killer feature
   - Write business logic once
   - Auto-generate REST endpoints
   - Works locally AND remotely (for microservices)

2. **Two Ways to Provide Metadata:**
   - **Simple (this example):** Via `RegisterServiceType` options
   - **Advanced (microservices):** Via Remote struct with `RemoteServiceMetaAdapter`

3. **Clean File Structure**
   - Separate files for model, contract, service
   - Not everything in one file!
   - Mirrors real production code

4. **Convention over Configuration**
   - Method names → HTTP methods
   - Parameters → Request binding
   - Automatic path generation

5. **Microservice-Ready**
   - Same service, multiple deployments
   - Monolith → Microservices without code changes
   - Remote services use HTTP proxies (see 04-multi-deployment)

## 🎓 What's Next?

**Other Service Topics:**
- Example 02: LazyLoad vs GetService (performance)
- Example 03: Service Dependencies (DI pattern)

**Related Topics:**
- 03-middleware: Adding auth, logging to auto-generated routes
- 04-configuration: Multi-deployment with YAML
- 00-introduction/examples/04-multi-deployment: Full example

## 🔗 Related Documentation

- `lokstra_registry/auto_router_helper.go` - Helper functions
- `core/router/autogen/autogen.go` - Auto-router implementation
- `core/router/convention/` - Convention registry
- `core/deploy/service_options.go` - Metadata options
- `docs/00-introduction/examples/04-multi-deployment/` - Complete microservices example with Remote structs

---

**Remember:** 
- **Simple examples:** Use `RegisterServiceType` with metadata options
- **Microservices:** Use Remote struct with `RemoteServiceMetaAdapter`
- **Both approaches** auto-generate routes - choose what fits your needs! 🎉
