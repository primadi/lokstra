# Service Router - Supported Handler Patterns

The Service Router supports **7 different handler patterns** to give you maximum flexibility in how you define your service methods.

## Overview of Supported Patterns

| Pattern | Signature | Use Case |
|---------|-----------|----------|
| **1** | `func(ctx *request.Context) error` | Manual control over request/response |
| **2** | `func(ctx *request.Context) (data, error)` | Return data directly, auto-formatted |
| **3** | `func(ctx *request.Context, param any) error` | Auto-bind parameters with tags |
| **4** | `func(ctx *request.Context, param any) (data, error)` | Auto-bind + return data |
| **5** | `func(param any) error` | Pure business logic, no HTTP concerns |
| **6** | `func(param any) (data, error)` | Pure business logic with return value |
| **7** | `func(w http.ResponseWriter, r *http.Request)` | Raw HTTP handler for advanced use |

## Pattern Details

### Pattern 1: Context Only

```go
func (s *Service) ListProducts(ctx *request.Context) error {
    products := s.getAllProducts()
    
    ctx.Resp.Json(map[string]any{
        "success": true,
        "data":    products,
    })
    return nil
}
```

**Features:**
- Full control over request and response
- Manual parameter extraction
- Flexible response formatting

**Convention Mapping:**
- `ListProducts` → `GET /products`

---

### Pattern 2: Context with Return Value

```go
func (s *Service) GetProduct(ctx *request.Context) (*Product, error) {
    id := ctx.Req.PathParam("id", "")
    
    product, exists := s.products[id]
    if !exists {
        return nil, fmt.Errorf("product not found")
    }
    
    return product, nil  // Auto-wrapped in {"success":true,"data":...}
}
```

**Features:**
- Return value automatically formatted as JSON
- Error automatically formatted as error response
- Less boilerplate for simple cases

**Response Format:**
```json
// Success
{"success": true, "data": {...}}

// Error
{"success": false, "error": "error message"}
```

**Convention Mapping:**
- `GetProduct` → `GET /products/{id}`

---

### Pattern 3: Context with Struct Parameter

```go
type GetProductRequest struct {
    ID string `path:"id"`  // Auto-bind from path parameter
}

func (s *Service) GetProduct(ctx *request.Context, req *GetProductRequest) error {
    product, exists := s.products[req.ID]
    if !exists {
        return fmt.Errorf("product not found")
    }
    
    ctx.Resp.Json(map[string]any{
        "success": true,
        "data":    product,
    })
    return nil
}
```

**Features:**
- Automatic parameter binding using struct tags
- Supports: `path:"..."`, `query:"..."`, `header:"..."`, `json:"..."`, `body:"..."`
- Built-in validation support

**Supported Tags:**
- `path:"id"` - Bind from URL path parameter
- `query:"name"` - Bind from query parameter
- `header:"Authorization"` - Bind from HTTP header
- `json:"email"` or `body:"email"` - Bind from JSON body

**Convention Mapping:**
- `GetProduct` → `GET /products/{id}`

---

### Pattern 4: Context + Struct Parameter + Return Value

```go
type SearchProductRequest struct {
    Query    string  `query:"q"`
    MinPrice float64 `query:"min_price"`
    MaxPrice float64 `query:"max_price"`
}

func (s *Service) SearchProducts(ctx *request.Context, req *SearchProductRequest) ([]*Product, error) {
    results := s.search(req.Query, req.MinPrice, req.MaxPrice)
    return results, nil  // Auto-formatted
}
```

**Features:**
- Combines auto-binding with automatic response formatting
- Most convenient for typical REST APIs
- Clean separation of concerns

**Convention Mapping:**
- `SearchProducts` → `GET /products/search`

---

### Pattern 5: Struct Parameter Only (No Context)

```go
type CreateProductRequest struct {
    Name  string  `json:"name" validate:"required"`
    Price float64 `json:"price" validate:"required,gt=0"`
}

func (s *Service) CreateProduct(req *CreateProductRequest) error {
    product := &Product{
        ID:    generateID(),
        Name:  req.Name,
        Price: req.Price,
    }
    
    s.products[product.ID] = product
    return nil  // Success response auto-generated
}
```

**Features:**
- **Pure business logic** - no HTTP concerns
- Most testable pattern
- Framework handles all HTTP details
- Cannot customize response format

**Benefits:**
- Easy to test (no need to mock context)
- Reusable in non-HTTP contexts
- Clean architecture

**Convention Mapping:**
- `CreateProduct` → `POST /products`

---

### Pattern 6: Struct Parameter Only + Return Value

```go
type UpdateProductRequest struct {
    ID    string  `path:"id"`
    Name  string  `json:"name" validate:"required"`
    Price float64 `json:"price" validate:"required,gt=0"`
}

func (s *Service) UpdateProduct(req *UpdateProductRequest) (*Product, error) {
    product, exists := s.products[req.ID]
    if !exists {
        return nil, fmt.Errorf("product not found")
    }
    
    product.Name = req.Name
    product.Price = req.Price
    
    return product, nil
}
```

**Features:**
- Pure business logic with return value
- Auto-binding from path + body
- Most elegant for CRUD operations

**Convention Mapping:**
- `UpdateProduct` → `PUT /products/{id}`

---

### Pattern 7: Raw HTTP Handler

```go
func (s *Service) DeleteProduct(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")  // Go 1.22+
    
    if _, exists := s.products[id]; !exists {
        w.WriteHeader(http.StatusNotFound)
        w.Write([]byte(`{"error":"not found"}`))
        return
    }
    
    delete(s.products, id)
    w.WriteHeader(http.StatusOK)
}
```

**Features:**
- Maximum control over HTTP details
- Direct access to ResponseWriter and Request
- Useful for streaming, SSE, WebSocket upgrades, etc.
- No automatic response handling

**Use Cases:**
- File uploads/downloads
- Streaming responses
- WebSocket connections
- Custom protocols

**Convention Mapping:**
- `DeleteProduct` → `DELETE /products/{id}`

---

## Parameter Binding

### Struct Tags

```go
type ComplexRequest struct {
    ID          string   `path:"id"`              // From URL path
    Query       string   `query:"q"`              // From query string
    Page        int      `query:"page"`           // With type conversion
    Token       string   `header:"Authorization"` // From HTTP header
    Name        string   `json:"name"`            // From JSON body
    Email       string   `body:"email"`           // Alternative to json
    Tags        []string `query:"tags"`           // Array support
}
```

### Binding Order

1. **Path parameters** - Highest priority
2. **Query parameters** - Second priority
3. **Headers** - Third priority
4. **Body (JSON)** - Lowest priority

Later bindings override earlier ones if field names conflict.

## Validation

All patterns support built-in validation using `validate` tags:

```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=3,max=50"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"required,min=18,max=120"`
}
```

If validation fails, automatic 400 error response with field details.

## Response Handling

### Automatic Response Format

For patterns that return values:

```go
// Success
{
    "success": true,
    "data": <return-value>
}

// Error
{
    "success": false,
    "error": "error message"
}
```

### Custom Response Format

Use Pattern 1 or 3 with context for custom responses:

```go
func (s *Service) CustomResponse(ctx *request.Context) error {
    ctx.Resp.
        WithStatus(201).
        WithHeader("X-Custom", "value").
        Json(map[string]any{
            "custom": "format",
            "data":   "...",
        })
    return nil
}
```

## Best Practices

### When to Use Each Pattern

| Pattern | Best For |
|---------|----------|
| **1-2** | Custom response formatting, complex logic |
| **3-4** | Standard REST APIs with validation |
| **5-6** | Pure business logic, maximum testability |
| **7** | Streaming, WebSockets, file handling |

### Recommended Approach

**For typical CRUD APIs:** Use Pattern 4 or 6
```go
func (s *Service) GetUser(ctx *request.Context, req *GetUserRequest) (*User, error)
```

**For pure business logic:** Use Pattern 5 or 6
```go
func (s *Service) CreateUser(req *CreateUserRequest) (*User, error)
```

**For custom responses:** Use Pattern 1 or 3
```go
func (s *Service) CustomEndpoint(ctx *request.Context) error
```

## Examples

See `patterns.go` for a complete working example demonstrating all 7 patterns.

```bash
cd cmd/examples/18-service-router
go run patterns.go
```

## Testing

### Testing Pure Business Logic (Pattern 5-6)

```go
func TestCreateProduct(t *testing.T) {
    service := NewProductService()
    
    req := &CreateProductRequest{
        Name:  "Test Product",
        Price: 99.99,
    }
    
    err := service.CreateProduct(req)
    assert.NoError(t, err)
    
    // No need to mock HTTP context!
}
```

### Testing with Context (Pattern 1-4)

```go
func TestListProducts(t *testing.T) {
    service := NewProductService()
    ctx := createMockContext()
    
    err := service.ListProducts(ctx)
    assert.NoError(t, err)
}
```

## Migration Guide

### From Manual Router

**Before:**
```go
r.GET("/products/{id}", func(ctx *request.Context) error {
    id := ctx.Req.PathParam("id", "")
    product, err := service.GetProduct(id)
    if err != nil {
        ctx.Resp.WithStatus(400).Json(map[string]any{
            "success": false,
            "error":   err.Error(),
        })
        return err
    }
    ctx.Resp.Json(map[string]any{
        "success": true,
        "data":    product,
    })
    return nil
})
```

**After (Pattern 4):**
```go
type GetProductRequest struct {
    ID string `path:"id"`
}

func (s *Service) GetProduct(ctx *request.Context, req *GetProductRequest) (*Product, error) {
    return s.products[req.ID], nil
}

// Auto-register
router := router.NewFromService(service, router.DefaultServiceRouterOptions())
```

**Result:** 15 lines → 5 lines (67% reduction)
