# SKILL 1: Generate Module Requirements

**When to use:** After BRD approval, before writing any code.

**Purpose:** Break down BRD into module-specific functional requirements, data models, and business rules.

---

## Workflow

```
Approved BRD → Identify Modules (Bounded Contexts) → Generate Module Requirements
```

### Step 1: Identify Modules (Bounded Context Analysis)

**Question:** "Can this be a separate microservice?"  
- If **yes** → Separate module
- If **no** → Part of existing module

**Example Breakdown:**

```
E-Commerce System (from BRD)
│
├── auth/              ← Authentication, sessions, JWT
├── user_profile/      ← User data, preferences, KYC
├── product/           ← Product catalog, categories, search
├── inventory/         ← Stock tracking, warehouse management
├── order/             ← Order creation, tracking, fulfillment
├── payment/           ← Payment processing, refunds, invoices
├── notification/      ← Email, SMS, push notifications
└── shared/
    └── domain/        ← Common models (Address, Money, etc.)
```

**Wrong Approach:** ❌
```
modules/
└── user/  ← Handles auth + profile + notifications (GOD MODULE!)
```

**Right Approach:** ✅
```
modules/
├── auth/
├── user_profile/
└── notification/
```

---

### Step 2: Generate Module Requirements

For **each module**, create: `docs/modules/<module_name>/REQUIREMENTS.md`

Use template: [docs/templates/MODULE_REQUIREMENTS_TEMPLATE.md](../../docs/templates/MODULE_REQUIREMENTS_TEMPLATE.md)

**Example: Order Module Requirements**

```markdown
# Module Requirements: Order Management

**Version:** 1.0.0  
**Status:** draft  
**Parent Document:** [BRD v1.0.0](../../BRD.md)  
**Last Updated:** 2026-01-27  

---

## 1. Module Overview

**Purpose:** Handle order lifecycle from creation to fulfillment.

**Boundaries:**
- ✅ In Scope: Order CRUD, status tracking, order history
- ❌ Out of Scope: Payment processing (payment module), inventory updates (inventory module)

**Dependencies:**
- `payment` module - Payment verification
- `inventory` module - Stock availability check
- `notification` module - Order status notifications
- `user_profile` module - User details, shipping address

---

## 2. Functional Requirements

### FR-ORDER-001: Create Order

**Priority:** High  
**User Story:** As a customer, I want to create an order with multiple items so that I can purchase products.

**Business Rules:**
- Minimum order value: $10
- Maximum items per order: 100
- Stock must be available for all items
- User must be authenticated
- Shipping address required

**Acceptance Criteria:**
- [ ] User can add 1-100 items
- [ ] System validates inventory availability (call inventory module)
- [ ] Order total = items + tax + shipping
- [ ] Order saved with status: `pending`
- [ ] Notification sent (async event)

**Validation Rules:**
```go
type CreateOrderRequest struct {
    Items []OrderItem `json:"items" validate:"required,min=1,max=100"`
    ShippingAddressID string `json:"shipping_address_id" validate:"required,uuid"`
    PaymentMethodID string `json:"payment_method_id" validate:"required,uuid"`
}

type OrderItem struct {
    ProductID string `json:"product_id" validate:"required,uuid"`
    Quantity int `json:"quantity" validate:"required,min=1,max=100"`
}
```

---

### FR-ORDER-002: Get Order by ID

**Priority:** High  
**User Story:** As a customer, I want to view my order details so that I can track my purchase.

**Business Rules:**
- User can only view their own orders (authorization check)
- Admin can view all orders

**Acceptance Criteria:**
- [ ] Return order with all items
- [ ] Include current status
- [ ] Show estimated delivery date
- [ ] Return 404 if order not found
- [ ] Return 403 if not authorized

---

### FR-ORDER-003: Update Order Status

**Priority:** High  
**User Story:** As an admin, I want to update order status so that customers can track fulfillment.

**Status Flow:**
```
pending → processing → shipped → delivered
                    ↓
                 cancelled
```

**Business Rules:**
- Only admin can update status
- Cannot update if status = `delivered` or `cancelled`
- Must provide tracking number when status = `shipped`

**Acceptance Criteria:**
- [ ] Status updated in database
- [ ] Status history logged (audit trail)
- [ ] Notification sent on status change
- [ ] Return 400 if invalid status transition

---

### FR-ORDER-004: Cancel Order

**Priority:** Medium  
**User Story:** As a customer, I want to cancel my order before it ships.

**Business Rules:**
- Can only cancel if status = `pending` or `processing`
- Refund initiated automatically (payment module)
- Inventory returned (inventory module)

**Acceptance Criteria:**
- [ ] Order status set to `cancelled`
- [ ] Refund event published
- [ ] Inventory event published
- [ ] Cancellation email sent
- [ ] Return 400 if cannot cancel

---

### FR-ORDER-005: List User Orders

**Priority:** High  
**User Story:** As a customer, I want to see all my orders with filtering and pagination.

**Query Parameters:**
- `status` - Filter by status (optional)
- `from_date`, `to_date` - Date range (optional)
- `page`, `limit` - Pagination (default: page=1, limit=20)

**Acceptance Criteria:**
- [ ] Return paginated results
- [ ] Support filtering by status, date range
- [ ] Sort by created_at DESC
- [ ] Return empty array if no orders

---

## 3. Data Models

### Order Entity

```go
type Order struct {
    ID                string    `json:"id" db:"id"`
    UserID            string    `json:"user_id" db:"user_id"`
    Status            string    `json:"status" db:"status"`
    TotalAmount       float64   `json:"total_amount" db:"total_amount"`
    ShippingAddressID string    `json:"shipping_address_id" db:"shipping_address_id"`
    PaymentMethodID   string    `json:"payment_method_id" db:"payment_method_id"`
    TrackingNumber    *string   `json:"tracking_number,omitempty" db:"tracking_number"`
    CreatedAt         time.Time `json:"created_at" db:"created_at"`
    UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

type OrderItem struct {
    ID        string  `json:"id" db:"id"`
    OrderID   string  `json:"order_id" db:"order_id"`
    ProductID string  `json:"product_id" db:"product_id"`
    Quantity  int     `json:"quantity" db:"quantity"`
    Price     float64 `json:"price" db:"price"` // Price at time of order
}

type OrderStatusHistory struct {
    ID        string    `json:"id" db:"id"`
    OrderID   string    `json:"order_id" db:"order_id"`
    Status    string    `json:"status" db:"status"`
    ChangedBy string    `json:"changed_by" db:"changed_by"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}
```

---

## 4. Business Rules

### BR-ORDER-001: Order Total Calculation
```
total = sum(item.price * item.quantity) + tax + shipping
```

### BR-ORDER-002: Status Transition Rules
```go
allowedTransitions := map[string][]string{
    "pending":    {"processing", "cancelled"},
    "processing": {"shipped", "cancelled"},
    "shipped":    {"delivered"},
    "delivered":  {}, // Terminal state
    "cancelled":  {}, // Terminal state
}
```

### BR-ORDER-003: Stock Reservation
- Reserve stock when order created
- Release stock if order cancelled
- Deduct stock when order shipped

---

## 5. Cross-Module Communication

### Dependencies

| Module         | Interaction Type | Purpose                    |
|----------------|------------------|----------------------------|
| `inventory`    | Repository call  | Check stock, reserve items |
| `payment`      | Repository call  | Verify payment method      |
| `user_profile` | Repository call  | Get shipping address       |
| `notification` | Event (async)    | Send order notifications   |

**Repository Injection Pattern:**
```go
// @Handler name="order-handler", prefix="/api/orders"
type OrderHandler struct {
    // @Inject "order-repository"
    OrderRepo OrderRepository
    
    // @Inject "inventory-repository"   // Cross-module
    InventoryRepo shared.InventoryRepository
    
    // @Inject "payment-repository"     // Cross-module
    PaymentRepo shared.PaymentRepository
}
```

---

## 6. Non-Functional Requirements

### Performance
- Create order: < 300ms
- Get order: < 100ms
- List orders: < 200ms (paginated)

### Scalability
- Support 10,000 orders/day
- Database indexes on: user_id, status, created_at

### Security
- RBAC: Customers can only access their own orders
- Admin role required for status updates

---

## 7. Testing Requirements

### Unit Tests
- Business rule validation (status transitions, total calculation)
- DTO validation (CreateOrderRequest)

### Integration Tests
- End-to-end order creation flow
- Cross-module repository calls
- Database transactions

### Test Coverage
- Minimum: 80%

---

## 8. Acceptance Criteria (Module-Level)

- [ ] All functional requirements implemented
- [ ] Unit test coverage > 80%
- [ ] Integration tests pass
- [ ] API documentation complete
- [ ] Database migrations created
- [ ] Code reviewed and approved

---

## Document History

| Version | Date       | Author      | Changes              |
|---------|------------|-------------|----------------------|
| 1.0.0   | 2026-01-27 | Bob Johnson | Initial draft        |
```

---

### Step 3: Save Module Requirements

```bash
# Create module docs directory
mkdir -p docs/modules/order

# Save requirements
docs/modules/order/REQUIREMENTS.md

# Version control
git add docs/modules/order/REQUIREMENTS.md
git commit -m "docs: add order module requirements v1.0.0"
```

---

## Module Dependency Rules

### Allowed Dependencies

✅ **Business Module → Infrastructure Module:**
```
modules/order/handler → modules/order/repository
```

✅ **Module → Shared Domain:**
```
modules/order/handler → modules/shared/domain
```

✅ **Module → Other Module Repository (via injection):**
```
modules/order/handler → modules/inventory/repository
```

### Forbidden Dependencies

❌ **Handler → Handler (direct call):**
```
modules/order/handler → modules/payment/handler  // WRONG!
```
*Use repository injection instead.*

❌ **Circular Dependencies:**
```
modules/order → modules/payment → modules/order  // WRONG!
```

---

## Validation Checklist

Before generating API spec:

- [ ] All functional requirements have acceptance criteria
- [ ] Data models defined with validation tags
- [ ] Business rules documented
- [ ] Cross-module dependencies identified
- [ ] Repository injection pattern specified
- [ ] Status = `approved`

---

## Next Step

Once module requirements approved, proceed to:
- **SKILL 2:** [03-api-spec.md](03-api-spec.md) - Generate OpenAPI specification

---

**Template:** [docs/templates/MODULE_REQUIREMENTS_TEMPLATE.md](../../docs/templates/MODULE_REQUIREMENTS_TEMPLATE.md)
