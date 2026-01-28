# Module Requirements: Product
## E-Commerce Order Management System

**Version:** 1.0.0  
**Status:** approved  
**BRD Reference:** BRD v1.0.0 (2026-01-28)  
**Last Updated:** 2026-02-05  
**Module Owner:** Bob Johnson (Tech Lead)  

---

## 1. Module Overview

**Purpose:** Manage product catalog with CRUD operations, search, filtering, and inventory tracking.

**Bounded Context:** Product information management, category taxonomy, inventory levels, product images.

**Business Value:**
- Enable customers to discover and browse 10,000+ products
- Support real-time inventory synchronization (reduce overselling by 100%)
- Improve search performance (< 200ms response time for 100,000+ products)
- Provide staff with efficient product management tools

**Dependencies:**
- Auth Module (authentication/authorization)

**Dependent Modules:**
- Order Module (product details, stock validation)

---

## 2. Functional Requirements

### FR-PROD-001: Create Product
**BRD Reference:** FR-002  
**Priority:** High  

**User Story:** As a staff member, I want to create products so that customers can purchase them.

**Acceptance Criteria:**
- POST `/api/products` endpoint
- Required fields: name, description, price, sku, category_id, stock_quantity
- Optional fields: images (up to 5)
- Return created product with generated ID
- Requires "staff" or "admin" role

**Business Rules:**
- SKU must be unique (return 409 Conflict if duplicate)
- Price must be > 0
- Stock quantity must be ≥ 0
- Category must exist (foreign key constraint)
- Images: max 5, JPEG/PNG only, max 2MB each

**Input Validation:**
```go
type CreateProductRequest struct {
    Name          string   `json:"name" validate:"required,min=3,max=100"`
    Description   string   `json:"description" validate:"required,min=10,max=1000"`
    Price         float64  `json:"price" validate:"required,gt=0"`
    SKU           string   `json:"sku" validate:"required,min=3,max=50,alphanum"`
    CategoryID    string   `json:"category_id" validate:"required"`
    StockQuantity int      `json:"stock_quantity" validate:"required,gte=0"`
    Images        []string `json:"images" validate:"omitempty,max=5,dive,url"`
}
```

**Success Response (201 Created):**
```json
{
  "data": {
    "id": "prd_xyz789",
    "name": "Wireless Mouse",
    "description": "Ergonomic wireless mouse with 3-year battery life",
    "price": 29.99,
    "sku": "MOUSE-WL-001",
    "category_id": "cat_electronics",
    "stock_quantity": 150,
    "images": [
      "https://cdn.example.com/products/mouse-01.jpg"
    ],
    "status": "active",
    "created_at": "2026-02-05T10:30:00Z",
    "updated_at": "2026-02-05T10:30:00Z"
  },
  "error": null
}
```

---

### FR-PROD-002: Update Product
**BRD Reference:** FR-002  
**Priority:** High  

**User Story:** As a staff member, I want to update product details so that information is accurate.

**Acceptance Criteria:**
- PATCH `/api/products/{id}` endpoint
- Updatable fields: name, description, price, stock_quantity, images, status
- SKU NOT updatable (immutable after creation)
- Requires "staff" or "admin" role

**Business Rules:**
- Price must be > 0 (if provided)
- Stock quantity must be ≥ 0 (if provided)
- Status enum: active, inactive (archived products)

**Input Validation:**
```go
type UpdateProductRequest struct {
    Name          *string  `json:"name" validate:"omitempty,min=3,max=100"`
    Description   *string  `json:"description" validate:"omitempty,min=10,max=1000"`
    Price         *float64 `json:"price" validate:"omitempty,gt=0"`
    StockQuantity *int     `json:"stock_quantity" validate:"omitempty,gte=0"`
    Images        []string `json:"images" validate:"omitempty,max=5,dive,url"`
    Status        *string  `json:"status" validate:"omitempty,oneof=active inactive"`
}
```

**Success Response (200 OK):**
```json
{
  "data": {
    "id": "prd_xyz789",
    "name": "Wireless Mouse Pro",
    "price": 34.99,
    "updated_at": "2026-02-06T14:20:00Z"
  },
  "error": null
}
```

---

### FR-PROD-003: Get Product by ID
**BRD Reference:** FR-002  
**Priority:** High  

**User Story:** As a customer, I want to view product details so that I can make informed purchase decisions.

**Acceptance Criteria:**
- GET `/api/products/{id}` endpoint
- Return full product details including category name
- Public endpoint (no authentication required)

**Success Response (200 OK):**
```json
{
  "data": {
    "id": "prd_xyz789",
    "name": "Wireless Mouse",
    "description": "Ergonomic wireless mouse...",
    "price": 29.99,
    "sku": "MOUSE-WL-001",
    "category": {
      "id": "cat_electronics",
      "name": "Electronics"
    },
    "stock_quantity": 150,
    "images": ["https://cdn.example.com/products/mouse-01.jpg"],
    "status": "active",
    "created_at": "2026-02-05T10:30:00Z"
  },
  "error": null
}
```

**Error Response (404 Not Found):**
```json
{
  "data": null,
  "error": {
    "code": "PRODUCT_NOT_FOUND",
    "message": "Product with ID prd_xyz789 not found"
  }
}
```

---

### FR-PROD-004: List Products (with Pagination)
**BRD Reference:** FR-002  
**Priority:** High  

**User Story:** As a customer, I want to browse products so that I can find items to purchase.

**Acceptance Criteria:**
- GET `/api/products` endpoint
- Query parameters: page, limit, category_id, in_stock, sort
- Default pagination: 20 products per page
- Public endpoint (no authentication required)

**Query Parameters:**
```
?page=1               # Page number (default: 1)
&limit=20             # Items per page (default: 20, max: 100)
&category_id=cat_123  # Filter by category
&in_stock=true        # Filter: stock_quantity > 0
&sort=price_asc       # Sort: price_asc, price_desc, name_asc, newest
```

**Success Response (200 OK):**
```json
{
  "data": [
    {
      "id": "prd_xyz789",
      "name": "Wireless Mouse",
      "price": 29.99,
      "sku": "MOUSE-WL-001",
      "category": {
        "id": "cat_electronics",
        "name": "Electronics"
      },
      "stock_quantity": 150,
      "images": ["https://cdn.example.com/products/mouse-01.jpg"],
      "status": "active"
    }
  ],
  "error": null,
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 523,
    "total_pages": 27
  }
}
```

---

### FR-PROD-005: Search Products
**BRD Reference:** FR-007  
**Priority:** Medium  

**User Story:** As a customer, I want to search for products so that I can quickly find what I need.

**Acceptance Criteria:**
- GET `/api/products/search` endpoint
- Query parameter: `q` (search query)
- Search in: product name, description
- Full-text search (PostgreSQL tsvector)
- Case-insensitive
- Minimum 2 characters
- Response time: < 200ms p95

**Query Parameters:**
```
?q=wireless mouse     # Search query (min 2 chars)
&category_id=cat_123  # Optional: filter by category
&in_stock=true        # Optional: in-stock only
&page=1&limit=20      # Pagination
```

**Success Response (200 OK):**
```json
{
  "data": [
    {
      "id": "prd_xyz789",
      "name": "Wireless Mouse",
      "description": "Ergonomic wireless mouse...",
      "price": 29.99,
      "stock_quantity": 150,
      "relevance_score": 0.95
    }
  ],
  "error": null,
  "meta": {
    "query": "wireless mouse",
    "total": 12,
    "page": 1,
    "limit": 20
  }
}
```

---

### FR-PROD-006: Delete Product (Soft Delete)
**BRD Reference:** FR-002  
**Priority:** Low  

**User Story:** As a staff member, I want to delete products so that discontinued items don't appear in listings.

**Acceptance Criteria:**
- DELETE `/api/products/{id}` endpoint
- Soft delete: Set `deleted_at` timestamp (do NOT hard delete)
- Product still accessible via direct ID (for historical orders)
- Product excluded from list/search results
- Requires "staff" or "admin" role

**Business Rules:**
- Cannot delete products with pending orders (return 409 Conflict)
- Deleted products can be restored by admin (set deleted_at = NULL)

**Success Response (200 OK):**
```json
{
  "data": {
    "message": "Product deleted successfully"
  },
  "error": null
}
```

---

### FR-PROD-007: Create Category
**BRD Reference:** FR-002  
**Priority:** Medium  

**User Story:** As a staff member, I want to create categories so that products are organized.

**Acceptance Criteria:**
- POST `/api/categories` endpoint
- Required fields: name, slug
- Optional field: parent_id (for subcategories)
- Requires "staff" or "admin" role

**Business Rules:**
- Slug must be unique, URL-friendly (lowercase, hyphens)
- Maximum category depth: 3 levels (parent → child → grandchild)

**Input Validation:**
```go
type CreateCategoryRequest struct {
    Name     string  `json:"name" validate:"required,min=2,max=50"`
    Slug     string  `json:"slug" validate:"required,min=2,max=50,lowercase,slug"`
    ParentID *string `json:"parent_id" validate:"omitempty"`
}
```

**Success Response (201 Created):**
```json
{
  "data": {
    "id": "cat_electronics",
    "name": "Electronics",
    "slug": "electronics",
    "parent_id": null,
    "created_at": "2026-02-05T10:30:00Z"
  },
  "error": null
}
```

---

### FR-PROD-008: List Categories
**BRD Reference:** FR-002  
**Priority:** Medium  

**User Story:** As a customer, I want to view categories so that I can browse products by type.

**Acceptance Criteria:**
- GET `/api/categories` endpoint
- Return hierarchical category tree
- Public endpoint (no authentication)

**Success Response (200 OK):**
```json
{
  "data": [
    {
      "id": "cat_electronics",
      "name": "Electronics",
      "slug": "electronics",
      "children": [
        {
          "id": "cat_computers",
          "name": "Computers",
          "slug": "computers",
          "children": []
        }
      ]
    }
  ],
  "error": null
}
```

---

### FR-PROD-009: Update Stock Quantity
**BRD Reference:** FR-006  
**Priority:** High  

**Purpose:** Internal method for Order module to reserve/deduct/release stock.

**Methods:**
```go
// Reserve stock when order created (pending payment)
ReserveStock(productID string, quantity int) error

// Deduct stock when order shipped (finalize reservation)
DeductStock(productID string, quantity int) error

// Release stock when order cancelled (undo reservation)
ReleaseStock(productID string, quantity int) error
```

**Business Rules:**
- Stock cannot go negative (return error if insufficient)
- Operations are atomic (use database transactions)
- Log all stock changes with reason (audit trail)

---

## 3. Data Models

### Product Entity
```go
type Product struct {
    ID            string    `json:"id"`              // Primary key: prd_{ulid}
    Name          string    `json:"name"`
    Description   string    `json:"description"`
    Price         float64   `json:"price"`           // USD, 2 decimal places
    SKU           string    `json:"sku"`             // Unique, indexed
    CategoryID    string    `json:"category_id"`     // Foreign key to categories
    StockQuantity int       `json:"stock_quantity"`
    Images        []string  `json:"images"`          // Array of CDN URLs
    Status        string    `json:"status"`          // Enum: active, inactive
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
    DeletedAt     *time.Time `json:"-"`              // Soft delete
}
```

### Category Entity
```go
type Category struct {
    ID        string     `json:"id"`         // Primary key: cat_{ulid}
    Name      string     `json:"name"`
    Slug      string     `json:"slug"`       // Unique, URL-friendly
    ParentID  *string    `json:"parent_id"`  // Foreign key (self-reference)
    CreatedAt time.Time  `json:"created_at"`
}
```

### Stock Change Log Entity
```go
type StockChangeLog struct {
    ID         string    `json:"id"`
    ProductID  string    `json:"product_id"`
    OldStock   int       `json:"old_stock"`
    NewStock   int       `json:"new_stock"`
    Reason     string    `json:"reason"`       // Enum: order_reserve, order_ship, order_cancel, manual_adjustment
    OrderID    *string   `json:"order_id"`     // Foreign key (if related to order)
    ChangedBy  string    `json:"changed_by"`   // User ID
    ChangedAt  time.Time `json:"changed_at"`
}
```

---

## 4. Business Rules

### BR-PROD-001: SKU Policy
- SKU must be unique across all products (active + deleted)
- Format: alphanumeric, hyphens allowed (e.g., MOUSE-WL-001)
- Length: 3-50 characters
- Case-insensitive uniqueness (stored as uppercase)

### BR-PROD-002: Pricing Policy
- Price must be > 0 (no free or negative prices)
- Price stored as `DECIMAL(10,2)` (max $99,999,999.99)
- Currency: USD only (v1.0)
- Price changes logged (audit trail)

### BR-PROD-003: Stock Management
- Stock quantity must be ≥ 0 (cannot go negative)
- Stock reservations expire after 15 minutes (if order not completed)
- Low stock alert: email sent when stock < 10 units
- Out-of-stock products still visible but marked "Out of Stock"

### BR-PROD-004: Image Policy
- Maximum 5 images per product
- Allowed formats: JPEG, PNG
- Maximum file size: 2MB per image
- Images stored on CDN (Cloudflare)
- First image is primary (used in listings)

### BR-PROD-005: Search Ranking
- Relevance score based on:
  - Exact name match (highest priority)
  - Partial name match
  - Description match
- Active products ranked higher than inactive
- In-stock products ranked higher than out-of-stock

---

## 5. Integration Points

| Integration      | Purpose                  | Protocol  | Auth Method        |
|------------------|--------------------------|-----------|---------------------|
| Warehouse API    | Stock synchronization    | gRPC      | API Key             |
| Cloudflare CDN   | Image storage/delivery   | REST API  | API Token           |

---

## 6. Error Codes

| Code                    | HTTP Status | Description                          |
|-------------------------|-------------|--------------------------------------|
| PRODUCT_NOT_FOUND       | 404         | Product ID does not exist            |
| SKU_ALREADY_EXISTS      | 409         | SKU is already in use                |
| CATEGORY_NOT_FOUND      | 404         | Category ID does not exist           |
| INSUFFICIENT_STOCK      | 400         | Not enough stock for reservation     |
| INVALID_PRICE           | 400         | Price must be > 0                    |
| INVALID_IMAGE_FORMAT    | 400         | Image must be JPEG or PNG            |
| IMAGE_LIMIT_EXCEEDED    | 400         | Maximum 5 images allowed             |
| SEARCH_QUERY_TOO_SHORT  | 400         | Search query must be ≥ 2 characters  |

---

## 7. Performance Requirements

- **Product List:** < 100ms p95 (with pagination, 20 items)
- **Product Search:** < 200ms p95 (full-text search)
- **Product Details:** < 50ms p95 (single ID lookup)
- **Stock Update:** < 10ms p95 (atomic transaction)
- **Concurrent Requests:** Support 10,000 req/sec (product browsing)

---

## 8. Security Requirements

- **Authorization:**
  - Product CRUD: "staff" or "admin" roles only
  - Product list/search/details: Public (no auth required)
- **Input Validation:**
  - Sanitize product name/description (prevent XSS)
  - Validate image URLs (prevent SSRF attacks)
- **Rate Limiting:**
  - Product creation: 100 requests/hour per user
  - Product search: 1000 requests/hour per IP

---

## 9. Dependencies

**External:**
- Warehouse API (stock synchronization)
- Cloudflare CDN (image storage)

**Internal:**
- Auth Module (authentication, authorization)

---

## 10. Testing Requirements

### Unit Tests
- Product CRUD operations
- Stock reservation/deduction/release logic
- Search relevance scoring
- Category hierarchy (parent-child relationships)

### Integration Tests
- Complete product lifecycle (create → update → soft delete)
- Stock synchronization with Warehouse API
- Search with various queries
- Concurrent stock updates (race conditions)

### Load Tests
- 10,000 concurrent product list requests
- 5,000 concurrent search requests
- 1,000 concurrent stock updates

---

## 11. Acceptance Criteria

- [ ] All functional requirements implemented
- [ ] 80%+ code coverage (unit + integration tests)
- [ ] Performance benchmarks met (p95 < 200ms for search)
- [ ] Full-text search functional (PostgreSQL tsvector)
- [ ] Stock synchronization tested with Warehouse API
- [ ] API documentation complete (OpenAPI spec)

---

## Document History

| Version | Date       | Author      | Changes                         |
|---------|------------|-------------|---------------------------------|
| 0.1     | 2026-02-03 | Bob Johnson | Initial draft                   |
| 0.2     | 2026-02-04 | Alice Chen  | Added stock management rules    |
| 1.0.0   | 2026-02-05 | Bob Johnson | Approved after BRD alignment    |

---

**Next Steps:**
1. Generate API specification for Product module (SKILL 2)
2. Generate database schema for Product module (SKILL 3)
3. Begin implementation (SKILL 4+)
