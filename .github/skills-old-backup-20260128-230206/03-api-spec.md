# SKILL 2: Generate API Specification

**When to use:** After module requirements approval, before implementation.

**Purpose:** Define precise API contracts (OpenAPI 3.0 spec) for all endpoints, validation rules, and error responses.

---

## Workflow

```
Module Requirements → Extract Endpoints → Generate OpenAPI Spec → Validate with Requirements
```

### Step 1: Extract Endpoints from Requirements

From module requirements, identify all API operations:

**Example (Order Module):**

| Requirement      | HTTP Method | Endpoint           | Auth Required |
|------------------|-------------|--------------------|---------------|
| FR-ORDER-001     | POST        | /api/orders        | Yes           |
| FR-ORDER-002     | GET         | /api/orders/{id}   | Yes           |
| FR-ORDER-003     | PATCH       | /api/orders/{id}/status | Yes (Admin) |
| FR-ORDER-004     | DELETE      | /api/orders/{id}   | Yes           |
| FR-ORDER-005     | GET         | /api/orders        | Yes           |

---

### Step 2: Generate OpenAPI Specification

Save to: `docs/modules/<module_name>/API_SPEC.md`

Use template: [docs/templates/API_SPEC_TEMPLATE.md](../../docs/templates/API_SPEC_TEMPLATE.md)

**Example: Order API Specification**

```markdown
# API Specification: Order Management

**Version:** 1.0.0  
**Status:** draft  
**Parent Document:** [Module Requirements v1.0.0](REQUIREMENTS.md)  
**Base URL:** `/api/orders`  
**Last Updated:** 2026-01-27  

---

## Overview

This API manages the complete order lifecycle including creation, tracking, status updates, and cancellation.

**Authentication:** All endpoints require JWT Bearer token.

**Base Headers:**
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

---

## Endpoints

### 1. Create Order

**POST** `/api/orders`

**Purpose:** Create a new order with multiple items.

**Authorization:** User (authenticated)

**Request Body:**
```json
{
  "items": [
    {
      "product_id": "550e8400-e29b-41d4-a716-446655440000",
      "quantity": 2
    }
  ],
  "shipping_address_id": "660e8400-e29b-41d4-a716-446655440000",
  "payment_method_id": "770e8400-e29b-41d4-a716-446655440000"
}
```

**Validation Rules:**
```go
type CreateOrderRequest struct {
    Items []OrderItem `json:"items" validate:"required,min=1,max=100,dive"`
    ShippingAddressID string `json:"shipping_address_id" validate:"required,uuid"`
    PaymentMethodID string `json:"payment_method_id" validate:"required,uuid"`
}

type OrderItem struct {
    ProductID string `json:"product_id" validate:"required,uuid"`
    Quantity int `json:"quantity" validate:"required,min=1,max=100"`
}
```

**Success Response (201 Created):**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "user_id": "user-123",
  "status": "pending",
  "total_amount": 149.99,
  "shipping_address_id": "660e8400-e29b-41d4-a716-446655440000",
  "payment_method_id": "770e8400-e29b-41d4-a716-446655440000",
  "created_at": "2026-01-27T10:30:00Z",
  "updated_at": "2026-01-27T10:30:00Z",
  "items": [
    {
      "id": "item-1",
      "product_id": "550e8400-e29b-41d4-a716-446655440000",
      "quantity": 2,
      "price": 74.99
    }
  ]
}
```

**Error Responses:**

**400 Bad Request - Validation Error:**
```json
{
  "error": "Validation failed",
  "details": [
    {
      "field": "items",
      "message": "at least 1 item required"
    },
    {
      "field": "items[0].quantity",
      "message": "must be between 1 and 100"
    }
  ]
}
```

**400 Bad Request - Business Rule Violation:**
```json
{
  "error": "Insufficient stock",
  "details": {
    "product_id": "550e8400-e29b-41d4-a716-446655440000",
    "requested": 10,
    "available": 5
  }
}
```

**401 Unauthorized:**
```json
{
  "error": "Authentication required",
  "message": "Valid JWT token required"
}
```

**500 Internal Server Error:**
```json
{
  "error": "Internal server error",
  "message": "An unexpected error occurred"
}
```

---

### 2. Get Order by ID

**GET** `/api/orders/{id}`

**Purpose:** Retrieve order details including all items.

**Authorization:** User (own orders only) or Admin (all orders)

**Path Parameters:**
- `id` (string, required) - Order UUID

**Success Response (200 OK):**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "user_id": "user-123",
  "status": "shipped",
  "total_amount": 149.99,
  "shipping_address_id": "660e8400-e29b-41d4-a716-446655440000",
  "payment_method_id": "770e8400-e29b-41d4-a716-446655440000",
  "tracking_number": "TRACK123456",
  "estimated_delivery": "2026-01-30",
  "created_at": "2026-01-27T10:30:00Z",
  "updated_at": "2026-01-28T14:20:00Z",
  "items": [
    {
      "id": "item-1",
      "product_id": "550e8400-e29b-41d4-a716-446655440000",
      "product_name": "Wireless Keyboard",
      "quantity": 2,
      "price": 74.99
    }
  ],
  "status_history": [
    {
      "status": "pending",
      "changed_at": "2026-01-27T10:30:00Z"
    },
    {
      "status": "processing",
      "changed_at": "2026-01-27T15:00:00Z"
    },
    {
      "status": "shipped",
      "changed_at": "2026-01-28T14:20:00Z",
      "changed_by": "admin-456"
    }
  ]
}
```

**Error Responses:**

**404 Not Found:**
```json
{
  "error": "Order not found",
  "order_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

**403 Forbidden:**
```json
{
  "error": "Access denied",
  "message": "You can only access your own orders"
}
```

---

### 3. Update Order Status

**PATCH** `/api/orders/{id}/status`

**Purpose:** Update order status (admin only).

**Authorization:** Admin

**Path Parameters:**
- `id` (string, required) - Order UUID

**Request Body:**
```json
{
  "status": "shipped",
  "tracking_number": "TRACK123456"
}
```

**Validation Rules:**
```go
type UpdateOrderStatusRequest struct {
    Status string `json:"status" validate:"required,oneof=pending processing shipped delivered cancelled"`
    TrackingNumber *string `json:"tracking_number" validate:"omitempty,min=5,max=50"`
}
```

**Business Rules:**
- `tracking_number` required when status = `shipped`
- Status transitions must follow allowed flow (see requirements)

**Success Response (200 OK):**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "status": "shipped",
  "tracking_number": "TRACK123456",
  "updated_at": "2026-01-28T14:20:00Z"
}
```

**Error Responses:**

**400 Bad Request - Invalid Transition:**
```json
{
  "error": "Invalid status transition",
  "current_status": "delivered",
  "requested_status": "shipped",
  "message": "Cannot change status from delivered"
}
```

**403 Forbidden:**
```json
{
  "error": "Admin access required",
  "message": "Only administrators can update order status"
}
```

---

### 4. Cancel Order

**DELETE** `/api/orders/{id}`

**Purpose:** Cancel order (user or admin).

**Authorization:** User (own orders) or Admin

**Path Parameters:**
- `id` (string, required) - Order UUID

**Success Response (200 OK):**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "status": "cancelled",
  "cancelled_at": "2026-01-27T16:00:00Z",
  "refund_initiated": true
}
```

**Error Responses:**

**400 Bad Request - Cannot Cancel:**
```json
{
  "error": "Cannot cancel order",
  "current_status": "shipped",
  "message": "Orders can only be cancelled if status is pending or processing"
}
```

---

### 5. List User Orders

**GET** `/api/orders`

**Purpose:** List user's orders with filtering and pagination.

**Authorization:** User (own orders) or Admin (all orders)

**Query Parameters:**
- `status` (string, optional) - Filter by status
- `from_date` (string, optional) - ISO 8601 date (e.g., "2026-01-01")
- `to_date` (string, optional) - ISO 8601 date
- `page` (int, optional, default=1) - Page number
- `limit` (int, optional, default=20) - Items per page (max=100)

**Example Request:**
```
GET /api/orders?status=shipped&page=1&limit=10
```

**Success Response (200 OK):**
```json
{
  "data": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "status": "shipped",
      "total_amount": 149.99,
      "created_at": "2026-01-27T10:30:00Z",
      "item_count": 2
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 45,
    "total_pages": 5
  }
}
```

---

## Error Response Schema

**Standard Error Format:**
```go
type ErrorResponse struct {
    Error   string                 `json:"error"`
    Message string                 `json:"message,omitempty"`
    Details map[string]interface{} `json:"details,omitempty"`
}
```

**HTTP Status Codes:**
- `200` - Success
- `201` - Created
- `400` - Bad Request (validation error, business rule violation)
- `401` - Unauthorized (missing/invalid JWT)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found
- `500` - Internal Server Error

---

## Data Models

### Order Response Model
```go
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
    StatusHistory     []StatusHistory     `json:"status_history,omitempty"`
}

type OrderItemResponse struct {
    ID          string  `json:"id"`
    ProductID   string  `json:"product_id"`
    ProductName string  `json:"product_name"`
    Quantity    int     `json:"quantity"`
    Price       float64 `json:"price"`
}
```

---

## Rate Limiting

- **Rate Limit:** 100 requests/minute per user
- **Header:** `X-RateLimit-Remaining`
- **Response (429 Too Many Requests):**
```json
{
  "error": "Rate limit exceeded",
  "retry_after": 60
}
```

---

## Changelog

| Version | Date       | Author      | Changes              |
|---------|------------|-------------|----------------------|
| 1.0.0   | 2026-01-27 | Bob Johnson | Initial draft        |
```

---

### Step 3: Validate Against Requirements

**Cross-Check:**

| Requirement ID  | API Endpoint               | Status |
|-----------------|----------------------------|--------|
| FR-ORDER-001    | POST /api/orders           | ✅     |
| FR-ORDER-002    | GET /api/orders/{id}       | ✅     |
| FR-ORDER-003    | PATCH /api/orders/{id}/status | ✅  |
| FR-ORDER-004    | DELETE /api/orders/{id}    | ✅     |
| FR-ORDER-005    | GET /api/orders            | ✅     |

---

### Step 4: Save API Spec

```bash
# Save to module docs
docs/modules/order/API_SPEC.md

# Version control
git add docs/modules/order/API_SPEC.md
git commit -m "docs: add order API spec v1.0.0"
```

---

## Validation Checklist

Before generating schema:

- [ ] All functional requirements have corresponding endpoints
- [ ] Request/response schemas defined
- [ ] Validation rules specified (`validate` tags)
- [ ] Error responses documented
- [ ] Authorization requirements specified
- [ ] Status = `approved`

---

## Next Step

Once API spec approved, proceed to:
- **SKILL 3:** [04-schema.md](04-schema.md) - Generate database schema

---

**Template:** [docs/templates/API_SPEC_TEMPLATE.md](../../docs/templates/API_SPEC_TEMPLATE.md)
