# RemoteService

> Automatic method-to-HTTP mapping for remote service calls

## Overview

`RemoteService` provides automatic HTTP mapping for method calls, eliminating boilerplate code for remote service communication. It uses convention-based routing to automatically determine HTTP methods and paths from Go method names.

## Import Path

```go
import "github.com/primadi/lokstra/api_client"
```

---

## RemoteService Type

**Definition:**
```go
type RemoteService struct {
    client             *ClientRouter
    basePath           string               // Base path (e.g., "/auth", "/users")
    convention         string               // "rest", "rpc", "kebab-case"
    resourceName       string               // Resource name (e.g., "user")
    pluralResourceName string               // Plural form (e.g., "users")
    routeOverrides     map[string]string    // Method → custom path
    methodOverrides    map[string]string    // Method → HTTP method
    parser             *router.ConventionParser // Reuses server-side logic
}
```

**Fields:**
- `client` - ClientRouter for actual HTTP calls
- `basePath` - Base path prefix for all routes
- `convention` - Routing convention ("rest", "rpc", "kebab-case")
- `resourceName` - Singular resource name
- `pluralResourceName` - Plural resource name (for REST)
- `routeOverrides` - Custom paths for specific methods
- `methodOverrides` - Custom HTTP methods for specific methods
- `parser` - Convention parser for path generation

---

## Creating RemoteService

### NewRemoteService
Creates a new RemoteService instance.

**Signature:**
```go
func NewRemoteService(client *ClientRouter, basePath string) *RemoteService
```

**Example:**
```go
client := lokstra_registry.GetClientRouter("user-service")
service := api_client.NewRemoteService(client, "/api/v1/users")
```

---

### Builder Methods

**WithConvention** - Sets routing convention:
```go
func (rs *RemoteService) WithConvention(convention string) *RemoteService
```

**WithResourceName** - Sets resource name:
```go
func (rs *RemoteService) WithResourceName(name string) *RemoteService
```

**WithPluralResourceName** - Sets plural resource name:
```go
func (rs *RemoteService) WithPluralResourceName(name string) *RemoteService
```

**WithRouteOverride** - Overrides path for specific method:
```go
func (rs *RemoteService) WithRouteOverride(methodName, path string) *RemoteService
```

**WithMethodOverride** - Overrides HTTP method:
```go
func (rs *RemoteService) WithMethodOverride(methodName, httpMethod string) *RemoteService
```

**Example:**
```go
service := api_client.NewRemoteService(client, "/auth").
    WithConvention("rest").
    WithResourceName("user").
    WithPluralResourceName("users").
    WithRouteOverride("ValidateToken", "/auth/validate").
    WithMethodOverride("ValidateToken", "POST")
```

---

## CallRemoteService

Generic function for calling remote service methods.

**Signature:**
```go
func CallRemoteService[TResponse any](
    c *RemoteService,
    methodName string,
    ctx *request.Context,
    req any,
) (TResponse, error)
```

**Type Parameters:**
- `TResponse` - Expected response type

**Parameters:**
- `c` - RemoteService instance
- `methodName` - Go method name to call
- `ctx` - Request context (optional, can be nil)
- `req` - Request payload (optional, can be nil)

**Returns:**
- `TResponse` - Parsed response
- `error` - Error if call fails

**Example:**
```go
// Simple call
user, err := api_client.CallRemoteService[*User](
    service, "GetUser", ctx, &GetUserRequest{ID: 123})

// Without context
users, err := api_client.CallRemoteService[[]*User](
    service, "ListUsers", nil, nil)

// With request body
created, err := api_client.CallRemoteService[*User](
    service, "CreateUser", ctx, newUser)
```

---

## Method Name Mapping

RemoteService automatically maps Go method names to HTTP methods and paths.

### HTTP Method Mapping

**POST Methods:**
- `Create*` → POST
- `Add*` → POST
- `Process*` → POST
- `Generate*` → POST
- `Login*` → POST
- `Register*` → POST
- `Validate*` → POST (if has request body)

**PUT Methods:**
- `Update*` → PUT
- `Modify*` → PUT

**DELETE Methods:**
- `Delete*` → DELETE
- `Remove*` → DELETE

**GET Methods:**
- `Get*` → GET
- `Find*` → GET
- `List*` → GET
- `Fetch*` → GET

**Examples:**
```go
// POST
CreateUser      → POST /users
AddProduct      → POST /products
ProcessPayment  → POST /payments/process
LoginUser       → POST /auth/login

// PUT
UpdateUser      → PUT /users/{id}
ModifyProfile   → PUT /profiles/{id}

// DELETE
DeleteUser      → DELETE /users/{id}
RemoveItem      → DELETE /items/{id}

// GET
GetUser         → GET /users/{id}
ListUsers       → GET /users
FindByEmail     → GET /users/find-by-email
```

---

## Path Generation

### REST Convention

REST convention uses standard RESTful routing:

```go
service := api_client.NewRemoteService(client, "/api").
    WithConvention("rest").
    WithResourceName("user").
    WithPluralResourceName("users")

// Method mappings:
ListUsers       → GET    /users
GetUser         → GET    /users/{id}
CreateUser      → POST   /users
UpdateUser      → PUT    /users/{id}
DeleteUser      → DELETE /users/{id}
GetUserProfile  → GET    /users/{id}/profile
```

---

### RPC Convention

RPC convention uses method names as endpoints:

```go
service := api_client.NewRemoteService(client, "/rpc").
    WithConvention("rpc")

// Method mappings:
ValidateToken   → POST /rpc/ValidateToken
GetUserProfile  → POST /rpc/GetUserProfile
ProcessPayment  → POST /rpc/ProcessPayment
```

---

### Kebab-Case Convention

Kebab-case converts CamelCase to kebab-case:

```go
service := api_client.NewRemoteService(client, "/api").
    WithConvention("kebab-case")

// Method mappings:
ValidateToken   → POST /api/validate-token
GetUserProfile  → GET  /api/get-user-profile
ProcessPayment  → POST /api/process-payment
```

---

## Path Parameters

### Struct Tags

Use `path:"paramName"` tags to define path parameters:

```go
type GetUserRequest struct {
    ID int `path:"id"`
}

type UpdateUserRequest struct {
    ID   int    `path:"id"`
    Name string `json:"name"`
}

type GetDepartmentRequest struct {
    DeptID int `path:"dep"`
    UserID int `path:"id"`
}

// Calls:
api_client.CallRemoteService[*User](service, "GetUser", ctx, 
    &GetUserRequest{ID: 123})
// → GET /users/123

api_client.CallRemoteService[*User](service, "UpdateUser", ctx,
    &UpdateUserRequest{ID: 123, Name: "John"})
// → PUT /users/123 with body {"name": "John"}

api_client.CallRemoteService[*User](service, "GetDepartment", ctx,
    &GetDepartmentRequest{DeptID: 10, UserID: 123})
// → GET /departments/10/users/123
```

---

### Automatic Substitution

RemoteService automatically:
1. Extracts path parameters from struct tags
2. Substitutes placeholders in path template
3. Removes parameters from request body

```go
// Request
req := &UpdateUserRequest{
    ID:   123,    // path parameter
    Name: "John", // body field
}

// Becomes:
// Path: /users/123
// Body: {"name": "John"}  // ID excluded from body
```

---

## Method Overrides

### Route Override

Override the entire path for specific methods:

```go
service := api_client.NewRemoteService(client, "/auth").
    WithRouteOverride("ValidateToken", "/auth/validate").
    WithRouteOverride("RefreshToken", "/auth/refresh")

// Mappings:
ValidateToken → /auth/validate  (not /auth/validate-token)
RefreshToken  → /auth/refresh   (not /auth/refresh-token)
```

---

### Method Override

Override HTTP method for specific methods:

```go
service := api_client.NewRemoteService(client, "/api").
    WithMethodOverride("SearchUsers", "POST"). // Force POST
    WithMethodOverride("ValidateEmail", "GET") // Force GET

// Mappings:
SearchUsers   → POST /api/search-users  (not GET)
ValidateEmail → GET  /api/validate-email (not POST)
```

---

### Struct Tag Override

Use `method:"HTTP_METHOD"` tag to override HTTP method:

```go
type SearchRequest struct {
    Query  string `json:"query" method:"POST"`
    Limit  int    `json:"limit"`
}

// Despite Get* prefix, uses POST because of struct tag
api_client.CallRemoteService[*SearchResult](service, "GetSearchResults", ctx, req)
// → POST /search-results (not GET)
```

---

## Complete Examples

### User Service
```go
package service

import (
    "github.com/primadi/lokstra/api_client"
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/core/request"
)

// Request types
type GetUserRequest struct {
    ID int `path:"id"`
}

type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

type UpdateUserRequest struct {
    ID    int    `path:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

type DeleteUserRequest struct {
    ID int `path:"id"`
}

// Service wrapper
type UserService struct {
    remote *api_client.RemoteService
}

func NewUserService() *UserService {
    client := lokstra_registry.GetClientRouter("user-service")
    
    remote := api_client.NewRemoteService(client, "/api/v1/users").
        WithConvention("rest").
        WithResourceName("user").
        WithPluralResourceName("users")
    
    return &UserService{remote: remote}
}

// Service methods
func (s *UserService) ListUsers(ctx *request.Context) error {
    users, err := api_client.CallRemoteService[[]*User](
        s.remote, "ListUsers", ctx, nil)
    if err != nil {
        return handleError(ctx, err)
    }
    return ctx.Api.Ok(users)
}

func (s *UserService) GetUser(ctx *request.Context) error {
    req := &GetUserRequest{
        ID: ctx.Params.GetInt("id"),
    }
    
    user, err := api_client.CallRemoteService[*User](
        s.remote, "GetUser", ctx, req)
    if err != nil {
        return handleError(ctx, err)
    }
    return ctx.Api.Ok(user)
}

func (s *UserService) CreateUser(ctx *request.Context) error {
    req := &CreateUserRequest{
        Name:  ctx.Body.GetString("name"),
        Email: ctx.Body.GetString("email"),
    }
    
    user, err := api_client.CallRemoteService[*User](
        s.remote, "CreateUser", ctx, req)
    if err != nil {
        return handleError(ctx, err)
    }
    return ctx.Api.Created(user)
}

func (s *UserService) UpdateUser(ctx *request.Context) error {
    req := &UpdateUserRequest{
        ID:    ctx.Params.GetInt("id"),
        Name:  ctx.Body.GetString("name"),
        Email: ctx.Body.GetString("email"),
    }
    
    user, err := api_client.CallRemoteService[*User](
        s.remote, "UpdateUser", ctx, req)
    if err != nil {
        return handleError(ctx, err)
    }
    return ctx.Api.Ok(user)
}

func (s *UserService) DeleteUser(ctx *request.Context) error {
    req := &DeleteUserRequest{
        ID: ctx.Params.GetInt("id"),
    }
    
    _, err := api_client.CallRemoteService[any](
        s.remote, "DeleteUser", ctx, req)
    if err != nil {
        return handleError(ctx, err)
    }
    return ctx.Api.NoContent()
}

func handleError(ctx *request.Context, err error) error {
    if apiErr, ok := err.(*api_client.ApiError); ok {
        return ctx.Api.Error(apiErr.StatusCode, apiErr.Code, apiErr.Message)
    }
    return ctx.Api.InternalError(err.Error())
}
```

---

### Authentication Service
```go
package service

import (
    "github.com/primadi/lokstra/api_client"
    "github.com/primadi/lokstra/lokstra_registry"
)

// Request types
type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type GetProductRequest struct {
    ID int `json:"id"`
}

type UpdateProductRequest struct {
    ID          int     `json:"id"`
    Name        string  `json:"name"`
    Price       float64 `json:"price"`
    Description string  `json:"description"`
}

// Response types
type ProductResponse struct {
    ID          int     `json:"id"`
    Name        string  `json:"name"`
    Price       float64 `json:"price"`
    Description string  `json:"description"`
    CreatedAt   string  `json:"created_at"`
}

type Product struct {
    ID          int     `json:"id"`
    Name        string  `json:"name"`
    Price       float64 `json:"price"`
}

// Service
type ProductService struct {
    remote *api_client.RemoteService
}

func NewProductService() *ProductService {
    client := lokstra_registry.GetClientRouter("product-service")
    
    remote := api_client.NewRemoteService(client, "/products").
        WithConvention("kebab-case").
        WithRouteOverride("CreateProduct", "/products").
        WithRouteOverride("GetProduct", "/products/:id").
        WithRouteOverride("ListProducts", "/products")
    
    return &ProductService{remote: remote}
}

func (s *ProductService) CreateProduct(ctx *request.Context) error {
    req := &CreateProductRequest{
        Name:        ctx.Body.GetString("name"),
        Price:       ctx.Body.GetFloat("price"),
        Description: ctx.Body.GetString("description"),
    }
    
    product, err := api_client.CallRemoteService[*ProductResponse](
        s.remote, "CreateProduct", ctx, req)
    if err != nil {
        return handleProductError(ctx, err)
    }
    
    return ctx.Api.Created(product)
}

func (s *ProductService) GetProduct(id int) (*Product, error) {
    req := &GetProductRequest{ID: id}
    
    return api_client.CallRemoteService[*Product](
        s.remote, "GetProduct", nil, req)
}

func (s *ProductService) UpdateProduct(ctx *request.Context, id int) error {
    req := &UpdateProductRequest{
        ID:          id,
        Name:        ctx.Body.GetString("name"),
        Price:       ctx.Body.GetFloat("price"),
        Description: ctx.Body.GetString("description"),
    }
    
    product, err := api_client.CallRemoteService[*ProductResponse](
        s.remote, "UpdateProduct", ctx, req)
    if err != nil {
        return handleProductError(ctx, err)
    }
    
    return ctx.Api.Ok(product)
}

func handleProductError(ctx *request.Context, err error) error {
    if apiErr, ok := err.(*api_client.ApiError); ok {
        switch {
        case apiErr.IsNotFound():
            return ctx.Api.NotFound("Product not found")
        case apiErr.IsBadRequest():
            return ctx.Api.BadRequest(apiErr.Message)
        default:
            return ctx.Api.Error(apiErr.StatusCode, apiErr.Code, apiErr.Message)
        }
    }
    return ctx.Api.InternalError("Product operation failed")
}
```

---

### Order Service with Nested Resources
```go
package service

import (
    "github.com/primadi/lokstra/api_client"
    "github.com/primadi/lokstra/lokstra_registry"
)

// Request types
type GetOrderRequest struct {
    ID int `path:"id"`
}

type GetOrderItemsRequest struct {
    OrderID int `path:"order_id"`
}

type AddOrderItemRequest struct {
    OrderID   int    `path:"order_id"`
    ProductID int    `json:"product_id"`
    Quantity  int    `json:"quantity"`
}

type UpdateOrderStatusRequest struct {
    OrderID int    `path:"order_id"`
    Status  string `json:"status"`
}

// Service
type OrderService struct {
    remote *api_client.RemoteService
}

func NewOrderService() *OrderService {
    client := lokstra_registry.GetClientRouter("order-service")
    
    remote := api_client.NewRemoteService(client, "/api/v1").
        WithConvention("rest").
        WithResourceName("order").
        WithPluralResourceName("orders").
        WithRouteOverride("GetOrderItems", "/orders/{order_id}/items").
        WithRouteOverride("AddOrderItem", "/orders/{order_id}/items").
        WithRouteOverride("UpdateOrderStatus", "/orders/{order_id}/status")
    
    return &OrderService{remote: remote}
}

func (s *OrderService) GetOrder(ctx *request.Context) error {
    req := &GetOrderRequest{
        ID: ctx.Params.GetInt("id"),
    }
    
    order, err := api_client.CallRemoteService[*Order](
        s.remote, "GetOrder", ctx, req)
    if err != nil {
        return handleError(ctx, err)
    }
    return ctx.Api.Ok(order)
}

func (s *OrderService) GetOrderItems(ctx *request.Context) error {
    req := &GetOrderItemsRequest{
        OrderID: ctx.Params.GetInt("order_id"),
    }
    
    items, err := api_client.CallRemoteService[[]*OrderItem](
        s.remote, "GetOrderItems", ctx, req)
    if err != nil {
        return handleError(ctx, err)
    }
    return ctx.Api.Ok(items)
}

func (s *OrderService) AddOrderItem(ctx *request.Context) error {
    req := &AddOrderItemRequest{
        OrderID:   ctx.Params.GetInt("order_id"),
        ProductID: ctx.Body.GetInt("product_id"),
        Quantity:  ctx.Body.GetInt("quantity"),
    }
    
    item, err := api_client.CallRemoteService[*OrderItem](
        s.remote, "AddOrderItem", ctx, req)
    if err != nil {
        return handleError(ctx, err)
    }
    return ctx.Api.Created(item)
}

func (s *OrderService) UpdateOrderStatus(ctx *request.Context) error {
    req := &UpdateOrderStatusRequest{
        OrderID: ctx.Params.GetInt("order_id"),
        Status:  ctx.Body.GetString("status"),
    }
    
    order, err := api_client.CallRemoteService[*Order](
        s.remote, "UpdateOrderStatus", ctx, req)
    if err != nil {
        return handleError(ctx, err)
    }
    return ctx.Api.Ok(order)
}
```

---

### RPC-Style Service
```go
package service

import (
    "github.com/primadi/lokstra/api_client"
    "github.com/primadi/lokstra/lokstra_registry"
)

// Request types
type ProcessPaymentRequest struct {
    OrderID      int     `json:"order_id"`
    Amount       float64 `json:"amount"`
    Currency     string  `json:"currency"`
    CardToken    string  `json:"card_token"`
}

type RefundPaymentRequest struct {
    PaymentID    int     `json:"payment_id"`
    Amount       float64 `json:"amount"`
    Reason       string  `json:"reason"`
}

type ValidateCardRequest struct {
    CardToken string `json:"card_token"`
}

// Response types
type PaymentResult struct {
    TransactionID string  `json:"transaction_id"`
    Status        string  `json:"status"`
    Amount        float64 `json:"amount"`
    ProcessedAt   string  `json:"processed_at"`
}

type RefundResult struct {
    RefundID    string  `json:"refund_id"`
    Amount      float64 `json:"amount"`
    Status      string  `json:"status"`
}

type CardValidation struct {
    Valid      bool   `json:"valid"`
    CardType   string `json:"card_type"`
    LastFour   string `json:"last_four"`
    ExpiryDate string `json:"expiry_date"`
}

// Service
type PaymentService struct {
    remote *api_client.RemoteService
}

func NewPaymentService() *PaymentService {
    client := lokstra_registry.GetClientRouter("payment-service")
    
    // RPC-style: all methods as POST to /rpc/MethodName
    remote := api_client.NewRemoteService(client, "/rpc").
        WithConvention("rpc")
    
    return &PaymentService{remote: remote}
}

func (s *PaymentService) ProcessPayment(ctx *request.Context) error {
    req := &ProcessPaymentRequest{
        OrderID:   ctx.Body.GetInt("order_id"),
        Amount:    ctx.Body.GetFloat64("amount"),
        Currency:  ctx.Body.GetString("currency"),
        CardToken: ctx.Body.GetString("card_token"),
    }
    
    result, err := api_client.CallRemoteService[*PaymentResult](
        s.remote, "ProcessPayment", ctx, req)
    if err != nil {
        return handlePaymentError(ctx, err)
    }
    
    return ctx.Api.Ok(result)
}

func (s *PaymentService) RefundPayment(ctx *request.Context) error {
    req := &RefundPaymentRequest{
        PaymentID: ctx.Body.GetInt("payment_id"),
        Amount:    ctx.Body.GetFloat64("amount"),
        Reason:    ctx.Body.GetString("reason"),
    }
    
    result, err := api_client.CallRemoteService[*RefundResult](
        s.remote, "RefundPayment", ctx, req)
    if err != nil {
        return handlePaymentError(ctx, err)
    }
    
    return ctx.Api.Ok(result)
}

func (s *PaymentService) ValidateCard(cardToken string) (*CardValidation, error) {
    req := &ValidateCardRequest{CardToken: cardToken}
    
    return api_client.CallRemoteService[*CardValidation](
        s.remote, "ValidateCard", nil, req)
}

func handlePaymentError(ctx *request.Context, err error) error {
    if apiErr, ok := err.(*api_client.ApiError); ok {
        switch apiErr.Code {
        case "INSUFFICIENT_FUNDS":
            return ctx.Api.PaymentRequired("Insufficient funds")
        case "INVALID_CARD":
            return ctx.Api.BadRequest("Invalid card")
        case "PAYMENT_DECLINED":
            return ctx.Api.BadRequest("Payment declined")
        default:
            return ctx.Api.Error(apiErr.StatusCode, apiErr.Code, apiErr.Message)
        }
    }
    return ctx.Api.InternalError("Payment processing failed")
}
```

---

### Multi-Service Aggregator
```go
package service

import (
    "sync"
    "github.com/primadi/lokstra/api_client"
    "github.com/primadi/lokstra/lokstra_registry"
)

type AggregatorService struct {
    userService    *api_client.RemoteService
    orderService   *api_client.RemoteService
    paymentService *api_client.RemoteService
}

func NewAggregatorService() *AggregatorService {
    userClient := lokstra_registry.GetClientRouter("user-service")
    orderClient := lokstra_registry.GetClientRouter("order-service")
    paymentClient := lokstra_registry.GetClientRouter("payment-service")
    
    return &AggregatorService{
        userService: api_client.NewRemoteService(userClient, "/api/v1/users").
            WithConvention("rest").
            WithResourceName("user").
            WithPluralResourceName("users"),
            
        orderService: api_client.NewRemoteService(orderClient, "/api/v1/orders").
            WithConvention("rest").
            WithResourceName("order").
            WithPluralResourceName("orders"),
            
        paymentService: api_client.NewRemoteService(paymentClient, "/rpc").
            WithConvention("rpc"),
    }
}

func (s *AggregatorService) GetUserDashboard(ctx *request.Context) error {
    userID := ctx.Params.GetInt("user_id")
    
    type Result struct {
        User     *User
        Orders   []*Order
        Payments []*Payment
        Error    error
    }
    
    var wg sync.WaitGroup
    result := &Result{}
    
    // Parallel fetch
    wg.Add(3)
    
    // Fetch user
    go func() {
        defer wg.Done()
        user, err := api_client.CallRemoteService[*User](
            s.userService, "GetUser", ctx, 
            &GetUserRequest{ID: userID})
        result.User = user
        if err != nil {
            result.Error = err
        }
    }()
    
    // Fetch orders
    go func() {
        defer wg.Done()
        orders, err := api_client.CallRemoteService[[]*Order](
            s.orderService, "ListUserOrders", ctx,
            &ListUserOrdersRequest{UserID: userID})
        result.Orders = orders
        if err != nil && result.Error == nil {
            result.Error = err
        }
    }()
    
    // Fetch payments
    go func() {
        defer wg.Done()
        payments, err := api_client.CallRemoteService[[]*Payment](
            s.paymentService, "ListUserPayments", ctx,
            &ListUserPaymentsRequest{UserID: userID})
        result.Payments = payments
        if err != nil && result.Error == nil {
            result.Error = err
        }
    }()
    
    wg.Wait()
    
    if result.Error != nil {
        return handleError(ctx, result.Error)
    }
    
    dashboard := map[string]any{
        "user":     result.User,
        "orders":   result.Orders,
        "payments": result.Payments,
    }
    
    return ctx.Api.Ok(dashboard)
}
```

---

## Best Practices

### 1. Use Appropriate Convention
```go
// ✅ Good: REST for CRUD operations
userService := api_client.NewRemoteService(client, "/users").
    WithConvention("rest")

// ✅ Good: RPC for action-oriented operations
paymentService := api_client.NewRemoteService(client, "/rpc").
    WithConvention("rpc")

// 🚫 Avoid: Wrong convention for use case
crudService := api_client.NewRemoteService(client, "/api").
    WithConvention("rpc") // Should use REST
```

---

### 2. Define Request Types with Tags
```go
// ✅ Good: Clear path parameters
type UpdateUserRequest struct {
    ID   int    `path:"id"`
    Name string `json:"name"`
}

// 🚫 Avoid: Missing tags
type UpdateUserRequest struct {
    ID   int
    Name string
}
```

---

### 3. Use Route Overrides for Special Cases
```go
// ✅ Good: Override non-standard paths
service.WithRouteOverride("ValidateToken", "/auth/validate")

// 🚫 Avoid: Hardcoding paths in method calls
// (defeats the purpose of RemoteService)
```

---

### 4. Handle Errors Appropriately
```go
// ✅ Good: Check ApiError type
if apiErr, ok := err.(*api_client.ApiError); ok {
    return ctx.Api.Error(apiErr.StatusCode, apiErr.Code, apiErr.Message)
}

// 🚫 Avoid: Generic error handling
return ctx.Api.InternalError(err.Error())
```

---

### 5. Reuse RemoteService Instances
```go
// ✅ Good: Create once, reuse
type Service struct {
    remote *api_client.RemoteService
}

func NewService() *Service {
    return &Service{
        remote: api_client.NewRemoteService(client, "/api"),
    }
}

// 🚫 Avoid: Creating on every call
func (s *Service) GetUser(id int) (*User, error) {
    remote := api_client.NewRemoteService(client, "/api")
    return api_client.CallRemoteService[*User](remote, "GetUser", nil, &GetUserRequest{ID: id})
}

---

## See Also

- **[API Client](./api-client)** - FetchAndCast and options
- **[ClientRouter](./client-router)** - HTTP client routing
- **[Service](../01-core-packages/service)** - Service patterns

---

## Related Guides

- **[Remote Services](../../04-guides/remote-services/)** - Remote service patterns
- **[API Design](../../04-guides/api-design/)** - API design principles
- **[Conventions](../../04-guides/conventions/)** - Routing conventions
