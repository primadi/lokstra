# SKILL 3: Generate Database Schema

**When to use:** After API spec approval, before implementation.

**Purpose:** Define complete database schema including tables, indexes, relationships, migrations, and performance optimization.

---

## Workflow

```
API Spec + Module Requirements → Identify Entities → Design Schema → Generate Migrations
```

### Step 1: Extract Entities from API Spec

From API spec and requirements, identify all database entities:

**Example (Order Module):**

| Entity               | Description                  | Relationships              |
|----------------------|------------------------------|----------------------------|
| `orders`             | Main order table             | → order_items (1:N)        |
| `order_items`        | Line items in order          | → orders (N:1)             |
| `order_status_history` | Audit trail for status changes | → orders (N:1)          |

---

### Step 2: Generate Schema Document

Save to: `docs/modules/<module_name>/SCHEMA.md`

Use template: [docs/templates/SCHEMA_TEMPLATE.md](../../docs/templates/SCHEMA_TEMPLATE.md)

**Example: Order Schema Documentation**

```markdown
# Database Schema: Order Management

**Version:** 1.0.0  
**Status:** draft  
**Parent Document:** [API Spec v1.0.0](API_SPEC.md)  
**Database:** PostgreSQL 15+  
**Last Updated:** 2026-01-27  

---

## Overview

This schema supports the complete order lifecycle including order creation, item management, status tracking, and audit history.

**Naming Conventions:**
- Tables: plural lowercase with underscores (`orders`, `order_items`)
- Foreign keys: `{table}_id` (e.g., `user_id`, `order_id`)
- Indexes: `idx_{table}_{columns}` (e.g., `idx_orders_user_id`)
- Constraints: `chk_{table}_{purpose}` (e.g., `chk_orders_total_positive`)

---

## Schema Diagram

```
┌─────────────────┐
│     orders      │
├─────────────────┤
│ id (PK)         │
│ user_id (FK)    │──┐
│ status          │  │
│ total_amount    │  │
│ created_at      │  │
│ updated_at      │  │
└─────────────────┘  │
         │           │
         │ 1:N       │
         ▼           │
┌─────────────────┐  │
│  order_items    │  │
├─────────────────┤  │
│ id (PK)         │  │
│ order_id (FK)   │──┘
│ product_id (FK) │
│ quantity        │
│ price           │
└─────────────────┘
         │
         │ 1:N
         ▼
┌─────────────────────┐
│ order_status_history│
├─────────────────────┤
│ id (PK)             │
│ order_id (FK)       │──┘
│ status              │
│ changed_by (FK)     │
│ created_at          │
└─────────────────────┘
```

---

## Tables

### 1. orders

**Purpose:** Store main order information.

**Columns:**

| Column            | Type         | Constraints                | Description                |
|-------------------|--------------|----------------------------|----------------------------|
| id                | UUID         | PRIMARY KEY, DEFAULT gen_random_uuid() | Order ID        |
| user_id           | UUID         | NOT NULL, REFERENCES users(id) | Customer ID         |
| status            | VARCHAR(20)  | NOT NULL, DEFAULT 'pending' | Order status               |
| total_amount      | DECIMAL(10,2)| NOT NULL, CHECK (total_amount >= 0) | Total order value   |
| shipping_address_id | UUID       | NOT NULL, REFERENCES addresses(id) | Shipping address   |
| payment_method_id | UUID         | NOT NULL, REFERENCES payment_methods(id) | Payment method |
| tracking_number   | VARCHAR(50)  | NULL                       | Courier tracking number    |
| estimated_delivery| DATE         | NULL                       | Expected delivery date     |
| created_at        | TIMESTAMPTZ  | NOT NULL, DEFAULT NOW()    | Creation timestamp         |
| updated_at        | TIMESTAMPTZ  | NOT NULL, DEFAULT NOW()    | Last update timestamp      |

**Indexes:**
```sql
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);
CREATE INDEX idx_orders_user_status ON orders(user_id, status);
```

**Constraints:**
```sql
ALTER TABLE orders
  ADD CONSTRAINT chk_orders_total_positive
  CHECK (total_amount >= 0);

ALTER TABLE orders
  ADD CONSTRAINT chk_orders_status_valid
  CHECK (status IN ('pending', 'processing', 'shipped', 'delivered', 'cancelled'));
```

**Migration (Up):**
```sql
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
```

**Migration (Down):**
```sql
DROP TABLE IF EXISTS orders CASCADE;
```

---

### 2. order_items

**Purpose:** Store line items for each order.

**Columns:**

| Column      | Type         | Constraints                | Description                |
|-------------|--------------|----------------------------|----------------------------|
| id          | UUID         | PRIMARY KEY, DEFAULT gen_random_uuid() | Line item ID    |
| order_id    | UUID         | NOT NULL, REFERENCES orders(id) ON DELETE CASCADE | Order ID |
| product_id  | UUID         | NOT NULL, REFERENCES products(id) | Product ID         |
| quantity    | INT          | NOT NULL, CHECK (quantity > 0) | Item quantity         |
| price       | DECIMAL(10,2)| NOT NULL, CHECK (price >= 0) | Price at time of order |

**Indexes:**
```sql
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_product_id ON order_items(product_id);
```

**Constraints:**
```sql
ALTER TABLE order_items
  ADD CONSTRAINT chk_order_items_quantity_positive
  CHECK (quantity > 0);

ALTER TABLE order_items
  ADD CONSTRAINT chk_order_items_price_positive
  CHECK (price >= 0);
```

**Migration (Up):**
```sql
CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity INT NOT NULL CHECK (quantity > 0),
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0)
);

CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_product_id ON order_items(product_id);
```

**Migration (Down):**
```sql
DROP TABLE IF EXISTS order_items CASCADE;
```

---

### 3. order_status_history

**Purpose:** Audit trail for order status changes.

**Columns:**

| Column      | Type         | Constraints                | Description                |
|-------------|--------------|----------------------------|----------------------------|
| id          | UUID         | PRIMARY KEY, DEFAULT gen_random_uuid() | History entry ID |
| order_id    | UUID         | NOT NULL, REFERENCES orders(id) ON DELETE CASCADE | Order ID |
| status      | VARCHAR(20)  | NOT NULL                   | New status                 |
| changed_by  | UUID         | NOT NULL, REFERENCES users(id) | User who made change  |
| created_at  | TIMESTAMPTZ  | NOT NULL, DEFAULT NOW()    | Change timestamp           |

**Indexes:**
```sql
CREATE INDEX idx_order_status_history_order_id ON order_status_history(order_id);
CREATE INDEX idx_order_status_history_created_at ON order_status_history(created_at DESC);
```

**Migration (Up):**
```sql
CREATE TABLE order_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL,
    changed_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_order_status_history_order_id ON order_status_history(order_id);
CREATE INDEX idx_order_status_history_created_at ON order_status_history(created_at DESC);
```

**Migration (Down):**
```sql
DROP TABLE IF EXISTS order_status_history CASCADE;
```

---

## Relationships

| Parent Table     | Child Table           | Type | Cascade Behavior       |
|------------------|-----------------------|------|------------------------|
| users            | orders                | 1:N  | RESTRICT (prevent deletion if orders exist) |
| orders           | order_items           | 1:N  | CASCADE (delete items when order deleted) |
| orders           | order_status_history  | 1:N  | CASCADE (delete history when order deleted) |
| products         | order_items           | 1:N  | RESTRICT (prevent deletion if in orders) |

---

## Triggers

### auto_update_updated_at

**Purpose:** Automatically update `updated_at` column on row modification.

**SQL:**
```sql
CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();
```

---

## Views

### v_order_summary

**Purpose:** Simplified view for order listing with item counts.

**SQL:**
```sql
CREATE VIEW v_order_summary AS
SELECT
    o.id,
    o.user_id,
    o.status,
    o.total_amount,
    o.created_at,
    COUNT(oi.id) AS item_count
FROM orders o
LEFT JOIN order_items oi ON o.id = oi.order_id
GROUP BY o.id, o.user_id, o.status, o.total_amount, o.created_at;
```

---

## Stored Procedures

### sp_create_order

**Purpose:** Atomically create order with items and initial status history.

**SQL:**
```sql
CREATE OR REPLACE FUNCTION sp_create_order(
    p_user_id UUID,
    p_shipping_address_id UUID,
    p_payment_method_id UUID,
    p_items JSONB
) RETURNS UUID AS $$
DECLARE
    v_order_id UUID;
    v_item JSONB;
    v_total DECIMAL(10,2) := 0;
BEGIN
    -- Create order
    INSERT INTO orders (user_id, shipping_address_id, payment_method_id, total_amount)
    VALUES (p_user_id, p_shipping_address_id, p_payment_method_id, 0)
    RETURNING id INTO v_order_id;
    
    -- Insert items and calculate total
    FOR v_item IN SELECT * FROM jsonb_array_elements(p_items)
    LOOP
        INSERT INTO order_items (order_id, product_id, quantity, price)
        VALUES (
            v_order_id,
            (v_item->>'product_id')::UUID,
            (v_item->>'quantity')::INT,
            (v_item->>'price')::DECIMAL
        );
        
        v_total := v_total + ((v_item->>'price')::DECIMAL * (v_item->>'quantity')::INT);
    END LOOP;
    
    -- Update order total
    UPDATE orders SET total_amount = v_total WHERE id = v_order_id;
    
    -- Insert initial status history
    INSERT INTO order_status_history (order_id, status, changed_by)
    VALUES (v_order_id, 'pending', p_user_id);
    
    RETURN v_order_id;
END;
$$ LANGUAGE plpgsql;
```

---

## Performance Considerations

### Query Patterns

**Get order by ID (with items):**
```sql
-- Index used: PRIMARY KEY on orders, idx_order_items_order_id
SELECT o.*, oi.*
FROM orders o
LEFT JOIN order_items oi ON o.id = oi.order_id
WHERE o.id = $1;
```

**List user orders (paginated):**
```sql
-- Index used: idx_orders_user_status, idx_orders_created_at
SELECT * FROM orders
WHERE user_id = $1 AND status = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;
```

### Expected Data Volume

| Table                | Estimated Rows | Growth Rate     |
|----------------------|----------------|-----------------|
| orders               | 10,000,000     | 10,000/day      |
| order_items          | 30,000,000     | 30,000/day      |
| order_status_history | 40,000,000     | 40,000/day      |

### Partitioning Strategy (Future)

**When table > 50M rows:**
```sql
-- Partition by created_at (monthly)
CREATE TABLE orders_2026_01 PARTITION OF orders
FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
```

---

## Backup & Retention

- **Full backup:** Daily at 2 AM UTC
- **Incremental backup:** Every 6 hours
- **Retention:** 30 days for orders, 90 days for audit logs
- **Archival:** Orders older than 2 years moved to cold storage

---

## Security

### Row-Level Security (RLS)

**Enable RLS on orders table:**
```sql
ALTER TABLE orders ENABLE ROW LEVEL SECURITY;

-- Users can only see their own orders
CREATE POLICY orders_select_policy ON orders
FOR SELECT
USING (user_id = current_setting('app.user_id')::UUID OR current_setting('app.role') = 'admin');
```

---

## Migration Files

### File Naming Convention
```
<timestamp>_<description>.sql

Examples:
20260127_001_create_orders_table.sql
20260127_002_create_order_items_table.sql
20260127_003_create_order_status_history_table.sql
```

### Migration Order

1. `001_create_orders_table.sql`
2. `002_create_order_items_table.sql`
3. `003_create_order_status_history_table.sql`
4. `004_create_triggers.sql`
5. `005_create_views.sql`
6. `006_create_stored_procedures.sql`

---

## Testing Requirements

### Data Fixtures
```sql
-- Test order
INSERT INTO orders (id, user_id, status, total_amount, shipping_address_id, payment_method_id)
VALUES ('123e4567-e89b-12d3-a456-426614174000', 'user-123', 'pending', 149.99, 'addr-123', 'pm-123');

-- Test order items
INSERT INTO order_items (order_id, product_id, quantity, price)
VALUES ('123e4567-e89b-12d3-a456-426614174000', 'prod-123', 2, 74.99);
```

### Performance Tests
- Create order: < 100ms
- Get order with items: < 50ms
- List paginated orders: < 200ms

---

## Changelog

| Version | Date       | Author      | Changes              |
|---------|------------|-------------|----------------------|
| 1.0.0   | 2026-01-27 | Bob Johnson | Initial schema       |
```

---

### Step 3: Generate Migration Files

Create migration files in: `migrations/<module_name>/`

**Example: 001_create_orders_table.sql**

```sql
-- Migration: Create orders table
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

---

### Step 4: Save Schema & Migrations

```bash
# Save schema doc
docs/modules/order/SCHEMA.md

# Create migrations directory
mkdir -p migrations/order

# Save migration files
migrations/order/001_create_orders_table.sql
migrations/order/002_create_order_items_table.sql
migrations/order/003_create_order_status_history_table.sql

# Version control
git add docs/modules/order/SCHEMA.md migrations/order/
git commit -m "docs: add order schema v1.0.0 with migrations"
```

---

## Validation Checklist

Before implementation:

- [ ] All entities from API spec have corresponding tables
- [ ] Foreign keys match relationships in requirements
- [ ] Indexes created for all query patterns
- [ ] Constraints enforce business rules
- [ ] Triggers and stored procedures implemented
- [ ] Migration files created in correct order
- [ ] Performance considerations documented
- [ ] Status = `approved`

---

## Next Step

Once schema approved, proceed to:
- **SKILL 4:** [05-implementation.md](05-implementation.md) - Implement module code

---

**Template:** [docs/templates/SCHEMA_TEMPLATE.md](../../docs/templates/SCHEMA_TEMPLATE.md)
