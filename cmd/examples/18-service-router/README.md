# Example 18: Service Router Comparison

This example demonstrates **three different approaches** to creating HTTP routes in Lokstra:

1. **Manual Router** - Traditional explicit route registration
2. **Service Router** - Convention-based automatic route generation
3. **Pattern Router** - Demonstrating 7 flexible handler patterns

## Project Structure

```
18-service-router/
├── main.go                           # Main application entry point
├── manual-router.go                  # Traditional manual route registration
├── service-router.go                 # User service with convention-based routing
├── service-router-with-patterns.go   # Product service demonstrating 7 handler patterns
├── test-all-routers.http             # Comprehensive HTTP tests for all routers
└── README.md                         # This file
```

## Quick Start

```bash
# Run the server
go run main.go

# Server starts on :3000 with three routers:
# - Manual Router:   /api/v1/manual/*
# - Service Router:  /api/v1/auto/*
# - Pattern Router:  /api/v2/patterns/*
```

## Router Comparison

### 1. Manual Router (`/api/v1/manual/*`)

**Traditional explicit route registration** - Complete control but verbose.

```go
r := router.New("manual-user-router")
r.SetPathPrefix("/api/v1/manual")

r.GET("/users", func(ctx *request.Context) error {
    users, err := service.ListUsers(ctx)
    // Manual response handling...
    return nil
})

r.POST("/users", func(ctx *request.Context) error {
    var req CreateUserRequest
    ctx.Req.BindBody(&req)
    // Manual request binding and response handling...
    return nil
})
// ... repeat for PUT, DELETE, etc.
```

**Pros:**
- ✅ Complete control over every route
- ✅ Explicit and easy to understand
- ✅ Fine-grained customization

**Cons:**
- ❌ Verbose and repetitive
- ❌ Lots of boilerplate code
- ❌ Manual error handling for each route

### 2. Service Router (`/api/v1/auto/*`)

**Convention-based automatic routing** - Minimal code, maximum productivity.

```go
// Just create the service
type UserService struct {
    users map[string]*User
}

func (s *UserService) ListUsers(ctx *request.Context) ([]*User, error) { /*...*/ }
func (s *UserService) GetUser(ctx *request.Context, id string) (*User, error) { /*...*/ }
func (s *UserService) CreateUser(ctx *request.Context, req *CreateUserRequest) (*User, error) { /*...*/ }
func (s *UserService) UpdateUser(ctx *request.Context, id string, req *UpdateUserRequest) (*User, error) { /*...*/ }
func (s *UserService) DeleteUser(ctx *request.Context, id string) error { /*...*/ }
func (s *UserService) SearchUsers(ctx *request.Context) ([]*User, error) { /*...*/ }

// Auto-generate router from service (1 line!)
router := router.NewFromService(
    userService,
    router.DefaultServiceRouterOptions().WithPrefix("/api/v1/auto"),
)
```

**Convention Rules:**

| Method Name | HTTP Method | Path | Example |
|-------------|-------------|------|---------|
| `Get{Resource}` | GET | `/{resources}/{id}` | `GetUser` → `GET /users/{id}` |
| `List{Resources}` | GET | `/{resources}` | `ListUsers` → `GET /users` |
| `Create{Resource}` | POST | `/{resources}` | `CreateUser` → `POST /users` |
| `Update{Resource}` | PUT | `/{resources}/{id}` | `UpdateUser` → `PUT /users/{id}` |
| `Delete{Resource}` | DELETE | `/{resources}/{id}` | `DeleteUser` → `DELETE /users/{id}` |
| `Search{Resources}` | GET | `/{resources}/search` | `SearchUsers` → `GET /users/search` |

**Pros:**
- ✅ Minimal boilerplate (95% less code)
- ✅ Automatic request/response handling
- ✅ Consistent API patterns
- ✅ Self-documenting code

**Cons:**
- ❌ Less control over individual routes
- ❌ Must follow naming conventions

### 3. Pattern Router (`/api/v2/patterns/*`)

**Flexible handler patterns** - Demonstrating 7 different ways to write handlers.

```go
router := router.NewFromService(
    productService,
    router.DefaultServiceRouterOptions().
        WithPrefix("/api/v2/patterns").
        WithRouteOverride("DetailProduct", router.RouteMeta{
            HTTPMethod: "GET",
            Path:       "/products/{id}/detail",
        }),
)
```

**Pros:**
- ✅ Multiple handler styles for different needs
- ✅ From high-level (pure business logic) to low-level (raw HTTP)
- ✅ Choose the right abstraction for each endpoint

## 7 Handler Patterns

The Pattern Router demonstrates **7 different handler signatures**, each with different levels of abstraction:

### Pattern 1: Manual Control
```go
func (s *ProductService) ListProducts(ctx *request.Context) error
```
- Full manual control
- Handler writes response directly
- Use when: Need complete control over response format

### Pattern 2: Return Data
```go
func (s *ProductService) DetailProduct(ctx *request.Context) (*Product, error)
```
- Framework handles response formatting
- Just return your data
- Use when: Standard JSON responses

### Pattern 3: Auto-Bind + Manual Response
```go
func (s *ProductService) GetProduct(ctx *request.Context, req *GetProductRequest) error
```
- Automatic parameter binding from path/query/body
- Manual response control
- Use when: Complex request binding with custom responses

**Request binding with struct tags:**
```go
type GetProductRequest struct {
    ID string `path:"id"`  // Binds from URL path parameter
}
```

### Pattern 4: Auto-Bind + Return Data
```go
func (s *ProductService) SearchProducts(ctx *request.Context, req *SearchProductRequest) ([]*Product, error)
```
- Automatic parameter binding
- Automatic response formatting
- Use when: Standard CRUD with query parameters

**Multi-source binding:**
```go
type SearchProductRequest struct {
    Query    string  `query:"q"`           // From query string
    MinPrice float64 `query:"min_price"`   // From query string
    MaxPrice float64 `query:"max_price"`   // From query string
}
```

### Pattern 5: Pure Business Logic
```go
func (s *ProductService) CreateProduct(req *CreateProductRequest) error
```
- No HTTP context - pure business logic
- Framework handles everything (binding + response)
- Use when: Unit testable business logic

**Body binding:**
```go
type CreateProductRequest struct {
    Name  string  `json:"name" validate:"required"`
    Price float64 `json:"price" validate:"required,gt=0"`
}
```

### Pattern 6: Pure Logic + Return
```go
func (s *ProductService) UpdateProduct(req *UpdateProductRequest) (*Product, error)
```
- Pure business logic with return value
- Easiest to test
- Use when: Business logic that returns data

**Combined path + body binding:**
```go
type UpdateProductRequest struct {
    ID    string  `path:"id"`       // From URL path
    Name  string  `json:"name"`     // From JSON body
    Price float64 `json:"price"`    // From JSON body
}
```

### Pattern 7: Raw HTTP
```go
func (s *ProductService) DeleteProduct(w http.ResponseWriter, r *http.Request)
```
- Raw HTTP handler (stdlib compatible)
- Lowest level control
- Use when: Need direct access to HTTP primitives

## Struct Tag Reference

The router supports automatic binding from multiple sources:

```go
type ComplexRequest struct {
    ID          string  `path:"id"`                // URL path parameter
    Query       string  `query:"q"`                // Query string
    Page        int     `query:"page"`             // Query string (converted to int)
    AuthToken   string  `header:"Authorization"`   // HTTP header
    Name        string  `json:"name"`              // JSON body field
    Description string  `body:"description"`       // Alternative body binding
}
```

**Supported tags:**
- `path:"name"` - Bind from URL path parameter
- `query:"name"` - Bind from query string
- `header:"name"` - Bind from HTTP header
- `json:"name"` - Bind from JSON body
- `body:"name"` - Alternative body binding

## Testing

Use the provided HTTP test files:

```bash
# Test all three routers
test-all-routers.http

# Individual router tests
test-api-v1.http         # Manual and Service routers
test-api-v2.http         # Pattern router
test-patterns.http       # Legacy pattern tests
```

**VS Code**: Install "REST Client" extension and click "Send Request" above each request.

## Code Size Comparison

For the same functionality (6 CRUD endpoints):

| Approach | Lines of Code | Reduction |
|----------|---------------|-----------|
| Manual Router | ~150 lines | Baseline |
| Service Router | ~8 lines | **95% less** |
| Pattern Router | ~10 lines | **93% less** |

## When to Use Each Approach

### Use Manual Router When:
- ✅ Migrating legacy APIs with non-standard routes
- ✅ Need complete control over every detail
- ✅ Routes don't follow REST conventions
- ✅ One-off custom endpoints

### Use Service Router When:
- ✅ Building new RESTful APIs
- ✅ Standard CRUD operations
- ✅ Want to reduce boilerplate
- ✅ Team prefers conventions

### Use Pattern Router When:
- ✅ Need flexibility in handler complexity
- ✅ Mix of simple and complex endpoints
- ✅ Want testable business logic
- ✅ Gradual migration from patterns 1-7

## Example Output

When you run the server, it prints all registered routes:

```
=== Creating Manual Router ===
✓ Manual Router created successfully

=== Creating Service Router (Auto-generated) ===
✓ Service Router created successfully

=== Creating Pattern Router (7 Handler Patterns) ===
✓ Pattern Router created successfully

================================================================================
REGISTERED ROUTES COMPARISON
================================================================================

--- 1. Manual Router (Traditional) ---
[manual-user-router] GET /api/v1/manual/users
[manual-user-router] POST /api/v1/manual/users
[manual-user-router] GET /api/v1/manual/users/{id}
[manual-user-router] PUT /api/v1/manual/users/{id}
[manual-user-router] DELETE /api/v1/manual/users/{id}
[manual-user-router] GET /api/v1/manual/users/search

--- 2. Service Router (Auto-generated CRUD) ---
[UserService] GET /api/v1/auto/users -> UserService.ListUsers
[UserService] POST /api/v1/auto/users -> UserService.CreateUser
[UserService] GET /api/v1/auto/users/{id} -> UserService.GetUser
[UserService] PUT /api/v1/auto/users/{id} -> UserService.UpdateUser
[UserService] DELETE /api/v1/auto/users/{id} -> UserService.DeleteUser
[UserService] GET /api/v1/auto/users/search -> UserService.SearchUsers

--- 3. Pattern Router (7 Handler Patterns) ---
[ProductService] GET /api/v2/patterns/products -> ProductService.ListProducts
[ProductService] GET /api/v2/patterns/products/{id}/detail -> ProductService.DetailProduct
[ProductService] GET /api/v2/patterns/products/{id} -> ProductService.GetProduct
[ProductService] GET /api/v2/patterns/products/search -> ProductService.SearchProducts
[ProductService] POST /api/v2/patterns/products -> ProductService.CreateProduct
[ProductService] PUT /api/v2/patterns/products/{id} -> ProductService.UpdateProduct
[ProductService] DELETE /api/v2/patterns/products/{id} -> ProductService.DeleteProduct
```

## Key Takeaways

1. **Manual Router**: Traditional, explicit, verbose but fully controlled
2. **Service Router**: Convention-based, minimal code, highly productive
3. **Pattern Router**: Flexible patterns from low-level HTTP to pure business logic

All three approaches work together in the same application!

## Next Steps

- Explore other examples in `cmd/examples/`
- Read the main documentation in `docs/`
- Try customizing the conventions with route overrides
- Experiment with different handler patterns for your use cases
