# SKILL 8-13: Advanced Implementation Skills

**Purpose:** Complete the implementation workflow with handlers, configuration, testing, and validation.

---

## SKILL 8: Generate Handler (@Handler + @Route)

**Purpose:** Generate HTTP handlers using Lokstra annotations.

**Source:** `docs/modules/<module_name>/API_SPEC.md`

**Output:** `modules/<module_name>/handler/<entity>_handler.go`

### Rules

1. Use `@Handler` annotation with name and prefix
2. Use `@Route` annotation for each endpoint
3. Inject repositories via `@Inject`
4. Use pointer parameters for request binding
5. Return appropriate response types

### Example: Order Handler

**File:** `modules/order/handler/order_handler.go`

```go
package handler

import (
    "myapp/modules/order/domain"
    "myapp/modules/order/repository"
    "github.com/primadi/lokstra/core/request"
)

// @Handler name="order-handler", prefix="/api/orders"
type OrderHandler struct {
    // @Inject "order-repository"
    OrderRepo repository.OrderRepository
    
    // @Inject "inventory-repository"
    InventoryRepo interface{
        CheckStock(ctx context.Context, productID string, quantity int) (bool, error)
    }
}

// @Route "POST /", middlewares=["auth"]
func (h *OrderHandler) Create(ctx *request.Context, req *domain.CreateOrderRequest) error {
    // 1. Validate stock availability
    for _, item := range req.Items {
        available, err := h.InventoryRepo.CheckStock(ctx.Request.Context(), item.ProductID, item.Quantity)
        if err != nil {
            return ctx.Api.InternalServerError("Failed to check stock")
        }
        if !available {
            return ctx.Api.BadRequest(map[string]interface{}{
                "error": "Insufficient stock",
                "product_id": item.ProductID,
            })
        }
    }
    
    // 2. Calculate total (simplified - should get prices from product service)
    var totalAmount float64
    for _, item := range req.Items {
        // TODO: Get actual price from product repository
        price := 74.99 // Placeholder
        totalAmount += price * float64(item.Quantity)
    }
    
    // 3. Create order
    order := &domain.Order{
        UserID:            ctx.UserID(), // From JWT
        Status:            "pending",
        TotalAmount:       customtype.NewDecimal(totalAmount),
        ShippingAddressID: req.ShippingAddressID,
        PaymentMethodID:   req.PaymentMethodID,
    }
    
    items := make([]domain.OrderItem, len(req.Items))
    for i, item := range req.Items {
        items[i] = domain.OrderItem{
            ProductID: item.ProductID,
            Quantity:  item.Quantity,
            Price:     customtype.NewDecimal(74.99), // TODO: Actual price
        }
    }
    
    createdOrder, err := h.OrderRepo.Create(ctx.Request.Context(), order, items)
    if err != nil {
        return ctx.Api.InternalServerError("Failed to create order")
    }
    
    return ctx.Api.Created(createdOrder)
}

// @Route "GET /{id}", middlewares=["auth"]
func (h *OrderHandler) GetByID(ctx *request.Context, id string) error {
    order, items, err := h.OrderRepo.GetByIDWithItems(ctx.Request.Context(), id)
    if err != nil {
        if err.Error() == "order not found" {
            return ctx.Api.NotFound("Order not found")
        }
        return ctx.Api.InternalServerError("Failed to retrieve order")
    }
    
    // Authorization check: user can only view their own orders
    if order.UserID != ctx.UserID() && !ctx.IsAdmin() {
        return ctx.Api.Forbidden("Access denied")
    }
    
    // Convert to response DTO
    response := domain.OrderResponse{
        ID:                order.ID,
        UserID:            order.UserID,
        Status:            order.Status,
        TotalAmount:       order.TotalAmount.Float64(),
        ShippingAddressID: order.ShippingAddressID,
        PaymentMethodID:   order.PaymentMethodID,
        TrackingNumber:    order.TrackingNumber,
        CreatedAt:         order.CreatedAt.Format(time.RFC3339),
        UpdatedAt:         order.UpdatedAt.Format(time.RFC3339),
        Items:             make([]domain.OrderItemResponse, len(items)),
    }
    
    for i, item := range items {
        response.Items[i] = domain.OrderItemResponse{
            ID:        item.ID,
            ProductID: item.ProductID,
            Quantity:  item.Quantity,
            Price:     item.Price.Float64(),
        }
    }
    
    return ctx.Api.Ok(response)
}

// @Route "PATCH /{id}/status", middlewares=["auth", "admin"]
func (h *OrderHandler) UpdateStatus(ctx *request.Context, id string, req *domain.UpdateOrderStatusRequest) error {
    // Validate status transition (business rule)
    currentOrder, err := h.OrderRepo.GetByID(ctx.Request.Context(), id)
    if err != nil {
        return ctx.Api.NotFound("Order not found")
    }
    
    if !isValidTransition(currentOrder.Status, req.Status) {
        return ctx.Api.BadRequest(map[string]interface{}{
            "error": "Invalid status transition",
            "current_status": currentOrder.Status,
            "requested_status": req.Status,
        })
    }
    
    // Require tracking number for shipped status
    if req.Status == "shipped" && req.TrackingNumber == nil {
        return ctx.Api.BadRequest(map[string]interface{}{
            "error": "tracking_number required when status is shipped",
        })
    }
    
    err = h.OrderRepo.UpdateStatus(ctx.Request.Context(), id, req.Status, req.TrackingNumber)
    if err != nil {
        return ctx.Api.InternalServerError("Failed to update order status")
    }
    
    return ctx.Api.Ok(map[string]interface{}{
        "id": id,
        "status": req.Status,
    })
}

// @Route "DELETE /{id}", middlewares=["auth"]
func (h *OrderHandler) Cancel(ctx *request.Context, id string) error {
    // Get order
    order, err := h.OrderRepo.GetByID(ctx.Request.Context(), id)
    if err != nil {
        return ctx.Api.NotFound("Order not found")
    }
    
    // Authorization check
    if order.UserID != ctx.UserID() && !ctx.IsAdmin() {
        return ctx.Api.Forbidden("Access denied")
    }
    
    // Business rule: can only cancel if pending or processing
    if order.Status != "pending" && order.Status != "processing" {
        return ctx.Api.BadRequest(map[string]interface{}{
            "error": "Cannot cancel order",
            "current_status": order.Status,
            "message": "Orders can only be cancelled if status is pending or processing",
        })
    }
    
    err = h.OrderRepo.Delete(ctx.Request.Context(), id)
    if err != nil {
        return ctx.Api.InternalServerError("Failed to cancel order")
    }
    
    // TODO: Publish event for inventory release and refund
    
    return ctx.Api.Ok(map[string]interface{}{
        "id": id,
        "status": "cancelled",
    })
}

// @Route "GET /", middlewares=["auth"]
func (h *OrderHandler) List(ctx *request.Context, query *domain.ListOrdersQuery) error {
    userID := ctx.UserID()
    
    // Admin can view all orders
    if ctx.IsAdmin() && query.UserID != "" {
        userID = query.UserID
    }
    
    orders, total, err := h.OrderRepo.ListByUserID(ctx.Request.Context(), userID, *query)
    if err != nil {
        return ctx.Api.InternalServerError("Failed to list orders")
    }
    
    return ctx.Api.Ok(map[string]interface{}{
        "data": orders,
        "pagination": map[string]interface{}{
            "page":        query.Page,
            "limit":       query.Limit,
            "total":       total,
            "total_pages": (total + query.Limit - 1) / query.Limit,
        },
    })
}

// Helper function
func isValidTransition(current, next string) bool {
    transitions := map[string][]string{
        "pending":    {"processing", "cancelled"},
        "processing": {"shipped", "cancelled"},
        "shipped":    {"delivered"},
        "delivered":  {},
        "cancelled":  {},
    }
    
    allowed, exists := transitions[current]
    if !exists {
        return false
    }
    
    for _, s := range allowed {
        if s == next {
            return true
        }
    }
    return false
}
```

---

## SKILL 9: Generate Configuration (config.yaml)

**Purpose:** Configure services, dependencies, and deployments.

**Source:** Module requirements + implementation

**Output:** `config.yaml`

### Example: Order Module Configuration

**File:** `config.yaml`

```yaml
# Application Configuration
configs:
  # Database
  database:
    dsn: "postgres://user:password@localhost:5432/myapp?sslmode=disable"
    max-connections: 50
    max-idle: 10
  
  # Repository selection
  repository:
    implementation: "postgres-order-repository"
  
  # Application settings
  app:
    timeout: "30s"
    max-upload-size: "10MB"
  
  # JWT
  jwt:
    secret: "${JWT_SECRET}"
    expiry: "24h"

# Service Definitions
service-definitions:
  # Database pool
  db-pool:
    type: dbpool-pg
    config:
      dsn: "@database.dsn"
      max-connections: "@database.max-connections"
  
  # Order repository
  order-repository:
    type: postgres-order-repository
    depends-on: [db-pool]
  
  # Inventory repository (cross-module)
  inventory-repository:
    type: postgres-inventory-repository
    depends-on: [db-pool]
  
  # Order handler
  order-handler:
    type: order-handler  # Auto-registered via @Handler
    depends-on: [order-repository, inventory-repository]

# Deployments
deployments:
  development:
    servers:
      api:
        addr: ":8080"
        published-services:
          - order-handler
        middlewares:
          - recovery
          - request-logger
          - cors
        cors:
          allowed-origins: ["*"]
  
  production:
    servers:
      api:
        addr: ":8080"
        published-services:
          - order-handler
        middlewares:
          - recovery
          - request-logger
          - cors
        cors:
          allowed-origins: ["https://myapp.com"]
```

---

## SKILL 10: Generate Unit Tests

**Purpose:** Test business logic, validation, and repository methods.

**Source:** Module requirements + implementation

**Output:** `modules/<module_name>/*_test.go`

### Example: Order Repository Test

**File:** `modules/order/repository/order_repository_test.go`

```go
package repository

import (
    "context"
    "testing"
    "time"
    "myapp/modules/order/domain"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestOrderRepository_Create(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer db.Close()
    
    repo := &PostgresOrderRepository{DB: db}
    ctx := context.Background()
    
    // Test data
    order := &domain.Order{
        UserID:            "user-123",
        Status:            "pending",
        TotalAmount:       customtype.NewDecimal(149.99),
        ShippingAddressID: "addr-123",
        PaymentMethodID:   "pm-123",
    }
    
    items := []domain.OrderItem{
        {
            ProductID: "prod-123",
            Quantity:  2,
            Price:     customtype.NewDecimal(74.99),
        },
    }
    
    // Execute
    createdOrder, err := repo.Create(ctx, order, items)
    
    // Assert
    require.NoError(t, err)
    assert.NotEmpty(t, createdOrder.ID)
    assert.Equal(t, "user-123", createdOrder.UserID)
    assert.Equal(t, "pending", createdOrder.Status)
    assert.False(t, createdOrder.CreatedAt.IsZero())
}

func TestOrderRepository_GetByID(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := &PostgresOrderRepository{DB: db}
    ctx := context.Background()
    
    // Create test order
    order := createTestOrder(t, repo)
    
    // Execute
    retrieved, err := repo.GetByID(ctx, order.ID)
    
    // Assert
    require.NoError(t, err)
    assert.Equal(t, order.ID, retrieved.ID)
    assert.Equal(t, order.UserID, retrieved.UserID)
}

func TestOrderRepository_GetByID_NotFound(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := &PostgresOrderRepository{DB: db}
    ctx := context.Background()
    
    // Execute
    _, err := repo.GetByID(ctx, "non-existent-id")
    
    // Assert
    require.Error(t, err)
    assert.Contains(t, err.Error(), "order not found")
}

func TestOrderRepository_UpdateStatus(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := &PostgresOrderRepository{DB: db}
    ctx := context.Background()
    
    // Create test order
    order := createTestOrder(t, repo)
    
    // Execute
    trackingNumber := "TRACK123"
    err := repo.UpdateStatus(ctx, order.ID, "shipped", &trackingNumber)
    
    // Assert
    require.NoError(t, err)
    
    // Verify
    updated, _ := repo.GetByID(ctx, order.ID)
    assert.Equal(t, "shipped", updated.Status)
    assert.Equal(t, "TRACK123", *updated.TrackingNumber)
}

// Test helpers
func setupTestDB(t *testing.T) *sql.DB {
    dsn := os.Getenv("TEST_DATABASE_DSN")
    if dsn == "" {
        t.Skip("TEST_DATABASE_DSN not set")
    }
    
    db, err := sql.Open("postgres", dsn)
    require.NoError(t, err)
    
    // Run migrations
    runMigrations(t, db)
    
    return db
}

func createTestOrder(t *testing.T, repo *PostgresOrderRepository) *domain.Order {
    order := &domain.Order{
        UserID:            "user-123",
        Status:            "pending",
        TotalAmount:       customtype.NewDecimal(149.99),
        ShippingAddressID: "addr-123",
        PaymentMethodID:   "pm-123",
    }
    
    items := []domain.OrderItem{
        {ProductID: "prod-123", Quantity: 2, Price: customtype.NewDecimal(74.99)},
    }
    
    created, err := repo.Create(context.Background(), order, items)
    require.NoError(t, err)
    
    return created
}
```

---

## SKILL 11: Generate Integration Tests

**Purpose:** Test end-to-end API flows.

**Source:** API spec

**Output:** `modules/<module_name>/handler/*_test.go` or `tests/integration/*_test.go`

### Example: Order Handler Integration Test

**File:** `tests/integration/order_test.go`

```go
package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "myapp/modules/order/domain"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestOrderFlow_CreateAndGet(t *testing.T) {
    // Setup test server
    app := setupTestApp(t)
    server := httptest.NewServer(app.Router())
    defer server.Close()
    
    // Create order
    createReq := domain.CreateOrderRequest{
        Items: []domain.OrderItemRequest{
            {ProductID: "prod-123", Quantity: 2},
        },
        ShippingAddressID: "addr-123",
        PaymentMethodID:   "pm-123",
    }
    
    body, _ := json.Marshal(createReq)
    resp, err := http.Post(server.URL+"/api/orders", "application/json", bytes.NewReader(body))
    require.NoError(t, err)
    defer resp.Body.Close()
    
    // Assert create response
    assert.Equal(t, http.StatusCreated, resp.StatusCode)
    
    var createdOrder domain.OrderResponse
    json.NewDecoder(resp.Body).Decode(&createdOrder)
    assert.NotEmpty(t, createdOrder.ID)
    assert.Equal(t, "pending", createdOrder.Status)
    
    // Get order
    getResp, err := http.Get(server.URL + "/api/orders/" + createdOrder.ID)
    require.NoError(t, err)
    defer getResp.Body.Close()
    
    // Assert get response
    assert.Equal(t, http.StatusOK, getResp.StatusCode)
    
    var retrievedOrder domain.OrderResponse
    json.NewDecoder(getResp.Body).Decode(&retrievedOrder)
    assert.Equal(t, createdOrder.ID, retrievedOrder.ID)
}

func TestOrderFlow_UpdateStatus(t *testing.T) {
    app := setupTestApp(t)
    server := httptest.NewServer(app.Router())
    defer server.Close()
    
    // Create test order
    orderID := createTestOrder(t, server)
    
    // Update status
    updateReq := domain.UpdateOrderStatusRequest{
        Status:         "shipped",
        TrackingNumber: stringPtr("TRACK123"),
    }
    
    body, _ := json.Marshal(updateReq)
    req, _ := http.NewRequest("PATCH", server.URL+"/api/orders/"+orderID+"/status", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+getAdminToken(t))
    
    resp, err := http.DefaultClient.Do(req)
    require.NoError(t, err)
    defer resp.Body.Close()
    
    // Assert
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestOrderFlow_CancelOrder(t *testing.T) {
    app := setupTestApp(t)
    server := httptest.NewServer(app.Router())
    defer server.Close()
    
    // Create test order
    orderID := createTestOrder(t, server)
    
    // Cancel order
    req, _ := http.NewRequest("DELETE", server.URL+"/api/orders/"+orderID, nil)
    req.Header.Set("Authorization", "Bearer "+getUserToken(t))
    
    resp, err := http.DefaultClient.Do(req)
    require.NoError(t, err)
    defer resp.Body.Close()
    
    // Assert
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

---

## SKILL 12: Update Main (Bootstrap)

**Purpose:** Register module and run server.

**Output:** `main.go`

### Example: Main with Order Module

**File:** `main.go`

```go
package main

import (
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/lokstra_registry"
    
    // Import handlers (triggers code generation)
    _ "myapp/modules/order/handler"
    
    // Import services if manually registered
    // _ "myapp/services"
)

func main() {
    // Auto-generate code when @Handler changes
    lokstra.Bootstrap()
    
    // Run server from config.yaml
    lokstra_registry.RunServerFromConfig()
}
```

---

## SKILL 13: Consistency Check

**Purpose:** Validate implementation against specifications.

### Checklist

**1. Requirements → API Spec:**
- [ ] All functional requirements have corresponding endpoints
- [ ] Request/response schemas match requirements

**2. API Spec → Implementation:**
- [ ] All endpoints implemented as @Route handlers
- [ ] Validation rules match (validate tags)
- [ ] Error responses match spec

**3. Schema → Repository:**
- [ ] All tables have CRUD methods in repository
- [ ] Indexes match query patterns
- [ ] Foreign key relationships enforced

**4. Cross-Module Dependencies:**
- [ ] Repository injection configured in config.yaml
- [ ] Shared domain models in modules/shared/domain/

**5. Tests:**
- [ ] Unit test coverage > 80%
- [ ] Integration tests for critical flows
- [ ] All tests passing

### Automated Check Command

```bash
# Generate consistency report
lokstra check-consistency

# Output:
# ✅ Requirements: 5/5 implemented
# ✅ API Endpoints: 5/5 implemented
# ✅ Database Tables: 3/3 have repositories
# ⚠️  Test Coverage: 75% (target: 80%)
```

---

## Complete Implementation Workflow Summary

```
1. Create Module Structure (SKILL 4)
   ↓
2. Create Database Migrations (SKILL 5)
   ↓
3. Generate Domain Models (SKILL 6)
   ↓
4. Generate Repository (SKILL 7)
   ↓
5. Generate Handler (SKILL 8)
   ↓
6. Generate Config (SKILL 9)
   ↓
7. Update Main (SKILL 12)
   ↓
8. Run Code Generation (lokstra autogen .)
   ↓
9. Generate Unit Tests (SKILL 10)
   ↓
10. Generate Integration Tests (SKILL 11)
    ↓
11. Consistency Check (SKILL 13)
    ↓
12. ✅ Implementation Complete
```

---

**Previous:** [04-schema.md](04-schema.md) - Database schema generation  
**Next:** [06-consistency-check.md](06-consistency-check.md) - Validation and quality checks
