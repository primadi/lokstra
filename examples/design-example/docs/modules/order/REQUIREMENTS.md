# Module Requirements: Order
## E-Commerce Order Management System

**Version:** 1.0.0  
**Status:** approved  
**BRD Reference:** BRD v1.0.0 (2026-01-28)  
**Last Updated:** 2026-02-05  
**Module Owner:** Bob Johnson (Tech Lead)  

---

## 1. Module Overview

**Purpose:** Handle complete order lifecycle from creation to delivery, including payment processing, status tracking, and order history.

**Bounded Context:** Order management, order items, order status transitions, payment coordination, shipping integration.

**Business Value:**
- Process 10,000+ daily orders with 99.9% uptime
- Reduce order processing time from 15 min to 4.5 min (70% improvement)
- Eliminate overselling with real-time stock reservation
- Provide customers with real-time order tracking
- Enable staff to manage orders efficiently

**Dependencies:**
- Auth Module (authentication, user data)
- Product Module (product details, stock management)
- Stripe API (payment processing)
- SendGrid (order notifications)
- Shipping API (tracking numbers)

**Dependent Modules:**
- None (top-level business process)

---

## 2. Functional Requirements

### FR-ORD-001: Create Order
**BRD Reference:** FR-003  
**Priority:** High  

**User Story:** As a customer, I want to create an order so that I can purchase products.

**Acceptance Criteria:**
- POST `/api/orders` endpoint
- Required fields: items (array of product_id + quantity)
- Validate: all products exist, sufficient stock for all items
- Reserve stock for all items (15-minute hold)
- Calculate order total: sum(item_price × quantity) + tax + shipping
- Process payment via Stripe
- Send order confirmation email
- Return created order with status "pending"
- Requires "customer" role (authenticated)

**Business Rules:**
- Minimum order value: $10
- Maximum items per order: 100
- Tax rate: 8.5% (California)
- Shipping: $5 flat rate for orders < $50, free for ≥ $50
- Stock reservation expires in 15 minutes (released if payment fails)
- Payment processed BEFORE order creation (no unpaid orders)

**Input Validation:**
```go
type CreateOrderRequest struct {
    Items []OrderItem `json:"items" validate:"required,min=1,max=100,dive"`
}

type OrderItem struct {
    ProductID string `json:"product_id" validate:"required"`
    Quantity  int    `json:"quantity" validate:"required,gte=1,lte=100"`
}
```

**Success Response (201 Created):**
```json
{
  "data": {
    "id": "ord_abc123",
    "user_id": "usr_xyz789",
    "items": [
      {
        "product_id": "prd_mouse",
        "product_name": "Wireless Mouse",
        "quantity": 2,
        "unit_price": 29.99,
        "subtotal": 59.98
      }
    ],
    "subtotal": 59.98,
    "tax": 5.10,
    "shipping": 0.00,
    "total": 65.08,
    "payment_id": "pi_stripe_123",
    "status": "pending",
    "created_at": "2026-02-05T10:30:00Z"
  },
  "error": null
}
```

**Error Response (400 Bad Request):**
```json
{
  "data": null,
  "error": {
    "code": "INSUFFICIENT_STOCK",
    "message": "Not enough stock for product prd_mouse",
    "details": {
      "product_id": "prd_mouse",
      "requested": 10,
      "available": 5
    }
  }
}
```

---

### FR-ORD-002: Get Order by ID
**BRD Reference:** FR-004  
**Priority:** High  

**User Story:** As a customer, I want to view my order details so that I can track its progress.

**Acceptance Criteria:**
- GET `/api/orders/{id}` endpoint
- Return full order details including items, status history, tracking info
- Customers can only view their own orders (check user_id = JWT user_id)
- Staff/Admin can view all orders
- Requires authentication

**Success Response (200 OK):**
```json
{
  "data": {
    "id": "ord_abc123",
    "user_id": "usr_xyz789",
    "items": [
      {
        "product_id": "prd_mouse",
        "product_name": "Wireless Mouse",
        "quantity": 2,
        "unit_price": 29.99,
        "subtotal": 59.98
      }
    ],
    "subtotal": 59.98,
    "tax": 5.10,
    "shipping": 0.00,
    "total": 65.08,
    "payment_id": "pi_stripe_123",
    "status": "shipped",
    "tracking_number": "1Z999AA10123456784",
    "estimated_delivery": "2026-02-08T18:00:00Z",
    "status_history": [
      {
        "status": "pending",
        "timestamp": "2026-02-05T10:30:00Z"
      },
      {
        "status": "processing",
        "timestamp": "2026-02-05T11:00:00Z"
      },
      {
        "status": "shipped",
        "timestamp": "2026-02-06T09:15:00Z"
      }
    ],
    "created_at": "2026-02-05T10:30:00Z",
    "updated_at": "2026-02-06T09:15:00Z"
  },
  "error": null
}
```

**Error Response (403 Forbidden):**
```json
{
  "data": null,
  "error": {
    "code": "FORBIDDEN",
    "message": "You can only view your own orders"
  }
}
```

---

### FR-ORD-003: List User Orders
**BRD Reference:** FR-004  
**Priority:** High  

**User Story:** As a customer, I want to view all my orders so that I can track my purchase history.

**Acceptance Criteria:**
- GET `/api/orders` endpoint
- Return orders for authenticated user (from JWT user_id)
- Query parameters: status, page, limit, sort
- Default pagination: 20 orders per page
- Sort options: newest (default), oldest, total_asc, total_desc
- Requires authentication

**Query Parameters:**
```
?status=shipped       # Filter by status (optional)
&page=1               # Page number (default: 1)
&limit=20             # Items per page (default: 20, max: 100)
&sort=newest          # Sort: newest, oldest, total_asc, total_desc
```

**Success Response (200 OK):**
```json
{
  "data": [
    {
      "id": "ord_abc123",
      "items_count": 2,
      "total": 65.08,
      "status": "shipped",
      "tracking_number": "1Z999AA10123456784",
      "created_at": "2026-02-05T10:30:00Z"
    }
  ],
  "error": null,
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 47,
    "total_pages": 3
  }
}
```

---

### FR-ORD-004: Update Order Status
**BRD Reference:** FR-004  
**Priority:** High  

**User Story:** As a staff member, I want to update order status so that customers are informed of progress.

**Acceptance Criteria:**
- PATCH `/api/orders/{id}/status` endpoint
- Required field: status
- Optional fields: tracking_number (if status = shipped)
- Status flow validation: pending → processing → shipped → delivered
- Log status change in order_status_history table
- Send email notification on each status change
- Requires "staff" or "admin" role

**Business Rules:**
- Status flow: pending → processing → shipped → delivered
- Cannot skip statuses (must follow order)
- Cannot revert to previous status
- When status = "shipped": tracking_number required, deduct stock from inventory
- When status = "delivered": mark order as complete

**Input Validation:**
```go
type UpdateOrderStatusRequest struct {
    Status         string  `json:"status" validate:"required,oneof=processing shipped delivered"`
    TrackingNumber *string `json:"tracking_number" validate:"required_if=Status shipped"`
}
```

**Success Response (200 OK):**
```json
{
  "data": {
    "id": "ord_abc123",
    "status": "shipped",
    "tracking_number": "1Z999AA10123456784",
    "updated_at": "2026-02-06T09:15:00Z"
  },
  "error": null
}
```

**Error Response (400 Bad Request):**
```json
{
  "data": null,
  "error": {
    "code": "INVALID_STATUS_TRANSITION",
    "message": "Cannot transition from pending to shipped (must go through processing)"
  }
}
```

---

### FR-ORD-005: Cancel Order
**BRD Reference:** FR-005  
**Priority:** Medium  

**User Story:** As a customer, I want to cancel my order so that I can get a refund if I change my mind.

**Acceptance Criteria:**
- POST `/api/orders/{id}/cancel` endpoint
- Validate: status = "pending" or "processing" (cannot cancel if shipped/delivered)
- Process refund via Stripe
- Release reserved stock (call Product module)
- Set order status to "cancelled"
- Send cancellation email
- Customers can cancel their own orders, Staff/Admin can cancel any order

**Business Rules:**
- Can only cancel if status = "pending" or "processing"
- Cannot cancel if status = "shipped" or "delivered"
- Refund processed within 5-7 business days (Stripe policy)
- Cancelled orders still visible in order history
- Stock released immediately (available for other customers)

**Success Response (200 OK):**
```json
{
  "data": {
    "id": "ord_abc123",
    "status": "cancelled",
    "refund_id": "re_stripe_456",
    "refund_status": "pending",
    "cancelled_at": "2026-02-05T15:30:00Z"
  },
  "error": null
}
```

**Error Response (400 Bad Request):**
```json
{
  "data": null,
  "error": {
    "code": "CANNOT_CANCEL_ORDER",
    "message": "Cannot cancel order with status 'shipped'"
  }
}
```

---

### FR-ORD-006: List All Orders (Staff)
**BRD Reference:** FR-004  
**Priority:** Medium  

**User Story:** As a staff member, I want to view all orders so that I can manage fulfillment.

**Acceptance Criteria:**
- GET `/api/orders/all` endpoint
- Query parameters: status, user_id, date_from, date_to, page, limit
- Return orders for all users (not limited to JWT user_id)
- Requires "staff" or "admin" role

**Query Parameters:**
```
?status=processing    # Filter by status
&user_id=usr_xyz789   # Filter by user
&date_from=2026-02-01 # Orders created after (inclusive)
&date_to=2026-02-07   # Orders created before (inclusive)
&page=1&limit=50      # Pagination
```

**Success Response (200 OK):**
```json
{
  "data": [
    {
      "id": "ord_abc123",
      "user_id": "usr_xyz789",
      "user_email": "john@example.com",
      "items_count": 2,
      "total": 65.08,
      "status": "processing",
      "created_at": "2026-02-05T10:30:00Z"
    }
  ],
  "error": null,
  "meta": {
    "page": 1,
    "limit": 50,
    "total": 523,
    "total_pages": 11
  }
}
```

---

### FR-ORD-007: Get Order Statistics (Staff)
**BRD Reference:** Success Metrics  
**Priority:** Low  

**User Story:** As a staff member, I want to view order statistics so that I can monitor business performance.

**Acceptance Criteria:**
- GET `/api/orders/stats` endpoint
- Query parameters: date_from, date_to (default: last 30 days)
- Return: total orders, total revenue, average order value, orders by status
- Requires "staff" or "admin" role

**Success Response (200 OK):**
```json
{
  "data": {
    "period": {
      "from": "2026-01-05T00:00:00Z",
      "to": "2026-02-05T23:59:59Z"
    },
    "total_orders": 15234,
    "total_revenue": 1247893.45,
    "average_order_value": 81.92,
    "orders_by_status": {
      "pending": 234,
      "processing": 567,
      "shipped": 1234,
      "delivered": 13089,
      "cancelled": 110
    }
  },
  "error": null
}
```

---

## 3. Data Models

### Order Entity
```go
type Order struct {
    ID              string       `json:"id"`              // Primary key: ord_{ulid}
    UserID          string       `json:"user_id"`         // Foreign key to users
    Items           []OrderItem  `json:"items"`           // Embedded items
    Subtotal        float64      `json:"subtotal"`        // Sum of item subtotals
    Tax             float64      `json:"tax"`             // 8.5% of subtotal
    Shipping        float64      `json:"shipping"`        // $5 or $0
    Total           float64      `json:"total"`           // Subtotal + Tax + Shipping
    PaymentID       string       `json:"payment_id"`      // Stripe payment intent ID
    Status          string       `json:"status"`          // Enum: pending, processing, shipped, delivered, cancelled
    TrackingNumber  *string      `json:"tracking_number"` // FedEx/UPS tracking
    EstimatedDelivery *time.Time `json:"estimated_delivery"`
    CreatedAt       time.Time    `json:"created_at"`
    UpdatedAt       time.Time    `json:"updated_at"`
    CancelledAt     *time.Time   `json:"cancelled_at"`
}
```

### Order Item Entity
```go
type OrderItem struct {
    ID          string    `json:"id"`              // Primary key: oit_{ulid}
    OrderID     string    `json:"order_id"`        // Foreign key to orders
    ProductID   string    `json:"product_id"`      // Foreign key to products
    ProductName string    `json:"product_name"`    // Snapshot (product may change later)
    Quantity    int       `json:"quantity"`
    UnitPrice   float64   `json:"unit_price"`      // Snapshot (price may change later)
    Subtotal    float64   `json:"subtotal"`        // Quantity × UnitPrice
    CreatedAt   time.Time `json:"created_at"`
}
```

### Order Status History Entity
```go
type OrderStatusHistory struct {
    ID        string    `json:"id"`
    OrderID   string    `json:"order_id"`    // Foreign key to orders
    Status    string    `json:"status"`      // Enum: pending, processing, shipped, delivered, cancelled
    Notes     *string   `json:"notes"`       // Optional notes (e.g., "Out for delivery")
    ChangedBy string    `json:"changed_by"`  // User ID (staff member)
    ChangedAt time.Time `json:"changed_at"`
}
```

---

## 4. Business Rules

### BR-ORD-001: Order Validation
- Minimum order value: $10 (before tax/shipping)
- Maximum items per order: 100
- All products must exist and be active (status = "active")
- All products must have sufficient stock (quantity ≤ stock_quantity)

### BR-ORD-002: Pricing Calculation
```
Subtotal = sum(item_quantity × item_unit_price)
Tax = Subtotal × 0.085  (8.5% California)
Shipping = $5 if Subtotal < $50, else $0
Total = Subtotal + Tax + Shipping
```

### BR-ORD-003: Status Transitions
```
Valid Flows:
pending → processing → shipped → delivered
pending → cancelled
processing → cancelled

Invalid Flows:
shipped → cancelled  (cannot cancel after shipped)
processing → delivered  (must go through shipped)
delivered → any  (terminal status)
```

### BR-ORD-004: Stock Management
- **Order Created (pending):** Reserve stock (15-minute hold)
- **Order Processing:** Stock remains reserved
- **Order Shipped:** Deduct stock from inventory (finalize reservation)
- **Order Cancelled:** Release reserved stock
- **Reservation Expires:** Release stock after 15 minutes (if payment fails)

### BR-ORD-005: Payment Processing
- Payment processed BEFORE order creation (via Stripe)
- Payment must be "succeeded" status before creating order
- Refunds processed when order cancelled
- Payment ID stored for reference (auditing, refunds)

---

## 5. Integration Points

| Integration      | Purpose                  | Protocol  | Auth Method        |
|------------------|--------------------------|-----------|---------------------|
| Stripe           | Payment processing       | REST API  | Secret Key          |
| SendGrid         | Order notifications      | REST API  | API Key             |
| Shipping API     | Tracking numbers         | REST API  | API Key             |
| Product Module   | Stock reservation/deduct | Internal  | Direct function call|

---

## 6. Error Codes

| Code                        | HTTP Status | Description                          |
|-----------------------------|-------------|--------------------------------------|
| ORDER_NOT_FOUND             | 404         | Order ID does not exist              |
| INSUFFICIENT_STOCK          | 400         | Product stock insufficient           |
| MINIMUM_ORDER_NOT_MET       | 400         | Order total < $10                    |
| INVALID_STATUS_TRANSITION   | 400         | Status transition not allowed        |
| CANNOT_CANCEL_ORDER         | 400         | Order already shipped/delivered      |
| PAYMENT_FAILED              | 402         | Stripe payment declined              |
| PRODUCT_NOT_FOUND           | 400         | Product in order does not exist      |
| ORDER_LIMIT_EXCEEDED        | 400         | More than 100 items in order         |

---

## 7. Performance Requirements

- **Order Creation:** < 300ms p95 (excluding payment processing)
- **Order List:** < 100ms p95 (paginated, 20 orders)
- **Order Details:** < 50ms p95 (single ID lookup)
- **Status Update:** < 50ms p95 (database update + email queue)
- **Concurrent Orders:** Support 1,000 orders/minute (peak traffic)

---

## 8. Security Requirements

- **Authorization:**
  - Order creation: "customer" role (authenticated)
  - Order list: Customers see own orders, Staff/Admin see all
  - Order update: "staff" or "admin" roles only
  - Order cancel: Owner or Staff/Admin
- **Data Privacy:**
  - Customers cannot view other users' orders (403 Forbidden)
  - Payment details NOT stored (only Stripe payment_id)
- **Audit Logging:**
  - Log all order status changes (who, when, why)
  - Log all cancellations (user_id, reason, refund_id)

---

## 9. Dependencies

**External:**
- Stripe API (payment processing)
- SendGrid (email notifications)
- Shipping API (tracking updates)

**Internal:**
- Auth Module (authentication, user data)
- Product Module (product details, stock management)

---

## 10. Testing Requirements

### Unit Tests
- Order creation logic (validation, pricing calculation)
- Status transition validation
- Stock reservation/release logic

### Integration Tests
- Complete order flow (create → process → ship → deliver)
- Order cancellation with refund
- Stock synchronization with Product module
- Payment processing with Stripe (test mode)

### Load Tests
- 1,000 concurrent order creations
- 10,000 order list requests/minute
- Race conditions (multiple orders for same product with low stock)

---

## 11. Acceptance Criteria

- [ ] All functional requirements implemented
- [ ] 80%+ code coverage (unit + integration tests)
- [ ] Performance benchmarks met (p95 < 300ms for order creation)
- [ ] Payment integration functional (Stripe test mode)
- [ ] Email notifications working (SendGrid)
- [ ] Stock synchronization tested (concurrent orders)
- [ ] API documentation complete (OpenAPI spec)

---

## Document History

| Version | Date       | Author      | Changes                         |
|---------|------------|-------------|---------------------------------|
| 0.1     | 2026-02-03 | Bob Johnson | Initial draft                   |
| 0.2     | 2026-02-04 | Alice Chen  | Added status transition rules   |
| 1.0.0   | 2026-02-05 | Bob Johnson | Approved after BRD alignment    |

---

**Next Steps:**
1. Generate API specification for Order module (SKILL 2)
2. Generate database schema for Order module (SKILL 3)
3. Begin implementation (SKILL 4+)
