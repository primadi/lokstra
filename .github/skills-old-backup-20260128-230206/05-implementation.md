# SKILL 4-13: Implementation Skills

**Purpose:** Generate production-ready Lokstra Framework code from approved specifications.

**Prerequisites:** All documents approved (BRD → Requirements → API Spec → Schema).

---

## Implementation Order

```
1. Migrations → 2. Domain Models → 3. Repository → 4. Handler → 5. Config → 6. Tests
```

---

## SKILL 4: Create Module Structure

**Purpose:** Create folder structure and boilerplate files.

### Folder Structure

```
modules/
└── <module_name>/
    ├── handler/
    │   └── <entity>_handler.go
    ├── repository/
    │   └── <entity>_repository.go
    ├── domain/
    │   ├── <entity>.go          # Domain model
    │   └── <entity>_service.go  # Business logic (optional)
    └── migrations/
        └── <timestamp>_<description>.sql
```

### Command

```bash
mkdir -p modules/{handler,repository,domain,migrations}
```

**Example:**
```bash
mkdir -p modules/order/{handler,repository,domain}
```

---

## SKILL 5: Create Database Migrations

**Purpose:** Generate SQL migration files from schema document.

**Source:** `docs/modules/<module_name>/SCHEMA.md`

**Output:** `migrations/<module_name>/<timestamp>_<description>.sql`

### Migration Template

```sql
-- Migration: <description>
-- Module: <module_name>
-- Version: <schema_version>
-- Date: <date>

-- UP
<create_table_sql>
<create_indexes_sql>
<create_constraints_sql>

-- DOWN
DROP TABLE IF EXISTS <table_name> CASCADE;
```

### Example: Order Migration

**File:** `migrations/order/20260127_001_create_orders_table.sql`

```sql
-- Migration: Create orders table
-- Module: order
-- Version: 1.0.0
-- Date: 2026-01-27

-- UP
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    total_amount DECIMAL(10,2) NOT NULL CHECK (total_amount >= 0),
    shipping_address_id UUID NOT NULL REFERENCES addresses(id) ON DELETE RESTRICT,
    payment_method_id UUID NOT NULL REFERENCES payment_methods(id) ON DELETE RESTRICT,
    tracking_number VARCHAR(50) NULL,
    estimated_delivery DATE NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_orders_status_valid
      CHECK (status IN ('pending', 'processing', 'shipped', 'delivered', 'cancelled'))
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);
CREATE INDEX idx_orders_user_status ON orders(user_id, status);

-- DOWN
DROP TABLE IF EXISTS orders CASCADE;
```

**Run migration:**
```bash
# Using lokstra CLI (if available)
lokstra migrate up

# Or using psql
psql -U user -d database -f migrations/order/20260127_001_create_orders_table.sql
```

---

## SKILL 6: Generate Domain Models & DTOs

**Purpose:** Generate Go structs from schema and API spec.

**Source:** `docs/modules/<module_name>/SCHEMA.md` + `API_SPEC.md`

**Output:** `modules/<module_name>/domain/<entity>.go`

### Rules

1. **Domain models** match database schema (1:1)
2. **DTOs (Request/Response)** match API spec
3. Add `validate` tags from API spec validation rules
4. Use custom types for dates, decimals, etc.

### Example: Order Domain Models

**File:** `modules/order/domain/order.go`

```go
package domain

import (
    "time"
    "github.com/primadi/lokstra/common/customtype"
)

// Order represents the main order entity
type Order struct {
    ID                string                  `json:"id" db:"id"`
    UserID            string                  `json:"user_id" db:"user_id"`
    Status            string                  `json:"status" db:"status"`
    TotalAmount       customtype.Decimal      `json:"total_amount" db:"total_amount"`
    ShippingAddressID string                  `json:"shipping_address_id" db:"shipping_address_id"`
    PaymentMethodID   string                  `json:"payment_method_id" db:"payment_method_id"`
    TrackingNumber    *string                 `json:"tracking_number,omitempty" db:"tracking_number"`
    EstimatedDelivery *customtype.Date        `json:"estimated_delivery,omitempty" db:"estimated_delivery"`
    CreatedAt         time.Time               `json:"created_at" db:"created_at"`
    UpdatedAt         time.Time               `json:"updated_at" db:"updated_at"`
}

// OrderItem represents a line item in an order
type OrderItem struct {
    ID        string             `json:"id" db:"id"`
    OrderID   string             `json:"order_id" db:"order_id"`
    ProductID string             `json:"product_id" db:"product_id"`
    Quantity  int                `json:"quantity" db:"quantity"`
    Price     customtype.Decimal `json:"price" db:"price"`
}

// OrderStatusHistory represents audit trail
type OrderStatusHistory struct {
    ID        string    `json:"id" db:"id"`
    OrderID   string    `json:"order_id" db:"order_id"`
    Status    string    `json:"status" db:"status"`
    ChangedBy string    `json:"changed_by" db:"changed_by"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// --- DTOs (Request/Response) ---

// CreateOrderRequest - from API Spec
type CreateOrderRequest struct {
    Items             []OrderItemRequest `json:"items" validate:"required,min=1,max=100,dive"`
    ShippingAddressID string             `json:"shipping_address_id" validate:"required,uuid"`
    PaymentMethodID   string             `json:"payment_method_id" validate:"required,uuid"`
}

type OrderItemRequest struct {
    ProductID string `json:"product_id" validate:"required,uuid"`
    Quantity  int    `json:"quantity" validate:"required,min=1,max=100"`
}

// UpdateOrderStatusRequest - from API Spec
type UpdateOrderStatusRequest struct {
    Status         string  `json:"status" validate:"required,oneof=pending processing shipped delivered cancelled"`
    TrackingNumber *string `json:"tracking_number" validate:"omitempty,min=5,max=50"`
}

// OrderResponse - from API Spec
type OrderResponse struct {
    ID                string              `json:"id"`
    UserID            string              `json:"user_id"`
    Status            string              `json:"status"`
    TotalAmount       float64             `json:"total_amount"`
    ShippingAddressID string              `json:"shipping_address_id"`
    PaymentMethodID   string              `json:"payment_method_id"`
    TrackingNumber    *string             `json:"tracking_number,omitempty"`
    EstimatedDelivery *string             `json:"estimated_delivery,omitempty"`
    CreatedAt         string              `json:"created_at"`
    UpdatedAt         string              `json:"updated_at"`
    Items             []OrderItemResponse `json:"items"`
}

type OrderItemResponse struct {
    ID          string  `json:"id"`
    ProductID   string  `json:"product_id"`
    ProductName string  `json:"product_name"`
    Quantity    int     `json:"quantity"`
    Price       float64 `json:"price"`
}

// ListOrdersQuery - from API Spec
type ListOrdersQuery struct {
    Status   string `query:"status" validate:"omitempty,oneof=pending processing shipped delivered cancelled"`
    FromDate string `query:"from_date" validate:"omitempty,datetime=2006-01-02"`
    ToDate   string `query:"to_date" validate:"omitempty,datetime=2006-01-02"`
    Page     int    `query:"page" validate:"omitempty,min=1"`
    Limit    int    `query:"limit" validate:"omitempty,min=1,max=100"`
}
```

---

## SKILL 7: Generate Repository (Data Access)

**Purpose:** Generate repository interface and PostgreSQL implementation.

**Source:** `docs/modules/<module_name>/REQUIREMENTS.md` (functional requirements)

**Output:** `modules/<module_name>/repository/<entity>_repository.go`

### Rules

1. **Interface** in domain layer or repository file
2. **Implementation** uses `@Service` annotation
3. Methods match functional requirements (CRUD + custom queries)
4. Use parameterized queries (prevent SQL injection)

### Example: Order Repository

**File:** `modules/order/repository/order_repository.go`

```go
package repository

import (
    "context"
    "database/sql"
    "fmt"
    "myapp/modules/order/domain"
)

// OrderRepository interface
type OrderRepository interface {
    Create(ctx context.Context, order *domain.Order, items []domain.OrderItem) (*domain.Order, error)
    GetByID(ctx context.Context, id string) (*domain.Order, error)
    GetByIDWithItems(ctx context.Context, id string) (*domain.Order, []domain.OrderItem, error)
    UpdateStatus(ctx context.Context, id, status string, trackingNumber *string) error
    Delete(ctx context.Context, id string) error
    ListByUserID(ctx context.Context, userID string, query domain.ListOrdersQuery) ([]domain.Order, int, error)
}

// @Service "order-repository"
type PostgresOrderRepository struct {
    // @Inject "db-pool"
    DB *sql.DB
}

// Ensure interface compliance
var _ OrderRepository = (*PostgresOrderRepository)(nil)

// Create creates a new order with items
func (r *PostgresOrderRepository) Create(ctx context.Context, order *domain.Order, items []domain.OrderItem) (*domain.Order, error) {
    tx, err := r.DB.BeginTx(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback()

    // Insert order
    query := `
        INSERT INTO orders (user_id, status, total_amount, shipping_address_id, payment_method_id)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, created_at, updated_at
    `
    err = tx.QueryRowContext(ctx, query,
        order.UserID, order.Status, order.TotalAmount,
        order.ShippingAddressID, order.PaymentMethodID,
    ).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
    if err != nil {
        return nil, fmt.Errorf("insert order: %w", err)
    }

    // Insert items
    itemQuery := `
        INSERT INTO order_items (order_id, product_id, quantity, price)
        VALUES ($1, $2, $3, $4)
    `
    for _, item := range items {
        _, err = tx.ExecContext(ctx, itemQuery, order.ID, item.ProductID, item.Quantity, item.Price)
        if err != nil {
            return nil, fmt.Errorf("insert order item: %w", err)
        }
    }

    // Insert status history
    historyQuery := `
        INSERT INTO order_status_history (order_id, status, changed_by)
        VALUES ($1, $2, $3)
    `
    _, err = tx.ExecContext(ctx, historyQuery, order.ID, order.Status, order.UserID)
    if err != nil {
        return nil, fmt.Errorf("insert status history: %w", err)
    }

    if err = tx.Commit(); err != nil {
        return nil, fmt.Errorf("commit transaction: %w", err)
    }

    return order, nil
}

// GetByID retrieves order by ID
func (r *PostgresOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
    query := `
        SELECT id, user_id, status, total_amount, shipping_address_id, payment_method_id,
               tracking_number, estimated_delivery, created_at, updated_at
        FROM orders
        WHERE id = $1
    `
    
    order := &domain.Order{}
    err := r.DB.QueryRowContext(ctx, query, id).Scan(
        &order.ID, &order.UserID, &order.Status, &order.TotalAmount,
        &order.ShippingAddressID, &order.PaymentMethodID,
        &order.TrackingNumber, &order.EstimatedDelivery,
        &order.CreatedAt, &order.UpdatedAt,
    )
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("order not found: %s", id)
    }
    if err != nil {
        return nil, fmt.Errorf("query order: %w", err)
    }

    return order, nil
}

// GetByIDWithItems retrieves order with all items
func (r *PostgresOrderRepository) GetByIDWithItems(ctx context.Context, id string) (*domain.Order, []domain.OrderItem, error) {
    order, err := r.GetByID(ctx, id)
    if err != nil {
        return nil, nil, err
    }

    query := `
        SELECT id, order_id, product_id, quantity, price
        FROM order_items
        WHERE order_id = $1
    `
    
    rows, err := r.DB.QueryContext(ctx, query, id)
    if err != nil {
        return nil, nil, fmt.Errorf("query items: %w", err)
    }
    defer rows.Close()

    items := []domain.OrderItem{}
    for rows.Next() {
        var item domain.OrderItem
        err = rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price)
        if err != nil {
            return nil, nil, fmt.Errorf("scan item: %w", err)
        }
        items = append(items, item)
    }

    return order, items, nil
}

// UpdateStatus updates order status
func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, id, status string, trackingNumber *string) error {
    tx, err := r.DB.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback()

    // Update order
    query := `
        UPDATE orders
        SET status = $1, tracking_number = $2, updated_at = NOW()
        WHERE id = $3
    `
    result, err := tx.ExecContext(ctx, query, status, trackingNumber, id)
    if err != nil {
        return fmt.Errorf("update order: %w", err)
    }
    
    rows, _ := result.RowsAffected()
    if rows == 0 {
        return fmt.Errorf("order not found: %s", id)
    }

    // Insert status history
    historyQuery := `
        INSERT INTO order_status_history (order_id, status, changed_by)
        VALUES ($1, $2, $3)
    `
    _, err = tx.ExecContext(ctx, historyQuery, id, status, "system") // TODO: Use actual user ID
    if err != nil {
        return fmt.Errorf("insert status history: %w", err)
    }

    return tx.Commit()
}

// Delete marks order as cancelled
func (r *PostgresOrderRepository) Delete(ctx context.Context, id string) error {
    return r.UpdateStatus(ctx, id, "cancelled", nil)
}

// ListByUserID lists orders for a user with filtering and pagination
func (r *PostgresOrderRepository) ListByUserID(ctx context.Context, userID string, query domain.ListOrdersQuery) ([]domain.Order, int, error) {
    // Set defaults
    if query.Page == 0 {
        query.Page = 1
    }
    if query.Limit == 0 {
        query.Limit = 20
    }

    // Build WHERE clause
    where := "user_id = $1"
    args := []interface{}{userID}
    argIndex := 2

    if query.Status != "" {
        where += fmt.Sprintf(" AND status = $%d", argIndex)
        args = append(args, query.Status)
        argIndex++
    }

    if query.FromDate != "" {
        where += fmt.Sprintf(" AND created_at >= $%d", argIndex)
        args = append(args, query.FromDate)
        argIndex++
    }

    if query.ToDate != "" {
        where += fmt.Sprintf(" AND created_at <= $%d", argIndex)
        args = append(args, query.ToDate)
        argIndex++
    }

    // Count total
    countQuery := fmt.Sprintf("SELECT COUNT(*) FROM orders WHERE %s", where)
    var total int
    err := r.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total)
    if err != nil {
        return nil, 0, fmt.Errorf("count orders: %w", err)
    }

    // Get paginated results
    offset := (query.Page - 1) * query.Limit
    selectQuery := fmt.Sprintf(`
        SELECT id, user_id, status, total_amount, shipping_address_id, payment_method_id,
               tracking_number, estimated_delivery, created_at, updated_at
        FROM orders
        WHERE %s
        ORDER BY created_at DESC
        LIMIT $%d OFFSET $%d
    `, where, argIndex, argIndex+1)
    
    args = append(args, query.Limit, offset)
    
    rows, err := r.DB.QueryContext(ctx, selectQuery, args...)
    if err != nil {
        return nil, 0, fmt.Errorf("query orders: %w", err)
    }
    defer rows.Close()

    orders := []domain.Order{}
    for rows.Next() {
        var order domain.Order
        err = rows.Scan(
            &order.ID, &order.UserID, &order.Status, &order.TotalAmount,
            &order.ShippingAddressID, &order.PaymentMethodID,
            &order.TrackingNumber, &order.EstimatedDelivery,
            &order.CreatedAt, &order.UpdatedAt,
        )
        if err != nil {
            return nil, 0, fmt.Errorf("scan order: %w", err)
        }
        orders = append(orders, order)
    }

    return orders, total, nil
}
```

---

**(Continue in next message due to length...)**

**Next sections:**
- SKILL 8: Generate Handler (@Handler with @Route)
- SKILL 9: Generate Config (config.yaml)
- SKILL 10: Generate Unit Tests
- SKILL 11: Generate Integration Tests
- SKILL 12: Update Main (Bootstrap)
- SKILL 13: Consistency Check

Would you like me to continue with the remaining skills?
