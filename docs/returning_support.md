# RETURNING Clause Support in Lokstra Flow

The Lokstra Flow system now supports PostgreSQL-style `RETURNING` clauses in SQL EXEC operations, allowing you to capture data from INSERT, UPDATE, and DELETE operations.

## Overview

Traditional EXEC operations only return the number of affected rows. With RETURNING support, you can capture:
- Auto-generated IDs from INSERT operations
- Updated values and timestamps
- Data from deleted records for audit trails
- Computed values and aggregations

## Basic Usage

### Method 1: ExecReturning()
The most convenient way to execute a query with RETURNING:

```go
handler := NewHandler(regCtx, "example").
    ExecReturning("INSERT INTO users (name, email) VALUES (?, ?) RETURNING id, created_at", 
        "John Doe", "john@example.com").
    SaveAs("new_user").  // Saves: {"id": 123, "created_at": "2024-..."}
    Done()
```

### Method 2: WithReturning()
Mark an existing ExecSql step to handle RETURNING:

```go
handler := NewHandler(regCtx, "example").
    ExecSql("UPDATE users SET name = ? WHERE id = ? RETURNING id, name, updated_at", 
        "Jane Smith", 123).
    WithReturning().          // Mark as RETURNING query
    AffectOne().             // Guards work with RETURNING
    SaveAs("updated_user").  // Saves returned data
    Done()
```

## How It Works

When a step has RETURNING clause:
1. Instead of using `Exec()` (which returns only affected rows)
2. The system uses `QueryRowMap()` to capture returned data
3. Data is saved as `map[string]any` with column names as keys
4. Guards like `AffectOne()` still validate affected rows count

## Common Patterns

### 1. INSERT with Auto-Generated ID

```go
NewHandler(regCtx, "create-order").
    BeginTx().
    ExecReturning(`INSERT INTO orders (customer_id, total_amount, status, created_at) 
        VALUES (?, ?, 'pending', NOW()) 
        RETURNING id, created_at`, customerId, amount).
    SaveAs("order").
    
    // Use the generated ID in next step
    ExecSql("INSERT INTO order_items (order_id, product_id, quantity) VALUES (?, ?, ?)", 
        "{{.order.id}}", productId, quantity).
    Done().
    CommitOrRollback()
```

### 2. UPDATE with Audit Trail

```go
NewHandler(regCtx, "update-with-audit").
    BeginTx().
    ExecReturning(`UPDATE products 
        SET price = ?, updated_at = NOW() 
        WHERE id = ? 
        RETURNING id, name, price, updated_at`, newPrice, productId).
    WithName("product.update").
    AffectOne().
    SaveAs("updated_product").
    
    // Create audit log
    ExecSql(`INSERT INTO audit_logs (table_name, record_id, action, new_data) 
        VALUES ('products', ?, 'UPDATE', ?)`, 
        "{{.updated_product.id}}", "{{.updated_product}}").
    Done().
    CommitOrRollback()
```

### 3. DELETE with Recovery Data

```go
NewHandler(regCtx, "soft-delete").
    ExecReturning(`UPDATE users 
        SET deleted_at = NOW() 
        WHERE id = ? AND deleted_at IS NULL 
        RETURNING id, name, email, deleted_at`, userId).
    WithName("user.soft_delete").
    AffectOne().
    SaveAs("deleted_user").  // Save for potential recovery
    Done()
```

### 4. UPSERT Operations

```go
NewHandler(regCtx, "upsert-setting").
    ExecReturning(`INSERT INTO user_settings (user_id, key, value) 
        VALUES (?, ?, ?) 
        ON CONFLICT (user_id, key) 
        DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
        RETURNING id, 
            CASE WHEN xmax = 0 THEN 'inserted' ELSE 'updated' END as action`, 
        userId, settingKey, settingValue).
    SaveAs("upsert_result").
    
    DoNamed("log.upsert_action", func(ctx *Context) error {
        result, _ := ctx.Get("upsert_result")
        action := result.(map[string]any)["action"].(string)
        // Log whether record was inserted or updated
        return nil
    })
```

## Advanced Features

### Batch Operations with Summary Data

```go
ExecReturning(`UPDATE inventory 
    SET quantity = quantity - ? 
    WHERE product_id IN (SELECT product_id FROM order_items WHERE order_id = ?) 
    RETURNING 
        COUNT(*) as updated_count,
        SUM(quantity) as total_remaining,
        ARRAY_AGG(product_id) as affected_products`, 
    reduceAmount, orderId).
SaveAs("inventory_summary")
```

### Complex Computed Values

```go
ExecReturning(`INSERT INTO transactions (account_id, amount, type) 
    VALUES (?, ?, ?) 
    RETURNING 
        id,
        amount,
        (SELECT balance FROM accounts WHERE id = ?) as previous_balance,
        (SELECT balance FROM accounts WHERE id = ?) + amount as new_balance`, 
    accountId, amount, transactionType, accountId, accountId).
SaveAs("transaction_with_balances")
```

## Data Access Patterns

### Using Template Syntax

Access returned data in subsequent steps using template syntax:

```go
SaveAs("user_data").
ExecSql("INSERT INTO logs (message) VALUES (?)", 
    "User {{.user_data.name}} created with ID {{.user_data.id}}")
```

### Using Custom Functions

Process returned data in custom functions:

```go
SaveAs("order_data").
DoNamed("process.order", func(ctx *Context) error {
    orderData, _ := ctx.Get("order_data")
    order := orderData.(map[string]any)
    
    orderId := order["id"].(int64)
    total := order["total_amount"].(float64)
    
    // Business logic here
    ctx.Set("processed_order_id", orderId)
    return nil
})
```

## Error Handling

### Panic Protection

`WithReturning()` will panic if used on non-EXEC queries:

```go
// This will panic - Query operations already return data
QuerySql("SELECT * FROM users").
    WithReturning()  // ❌ PANIC: WithReturning can only be used with EXEC queries
```

### Guard Integration

All existing guards work with RETURNING queries:

```go
ExecReturning("UPDATE users SET name = ? WHERE id = ? RETURNING id, name", name, id).
    AffectOne().     // ✅ Validates exactly one row was affected
    SaveAs("user").  // ✅ Data saved if guard passes
    Done()
```

## Performance Considerations

1. **Query Execution**: RETURNING queries use `QueryRowMap()` instead of `Exec()`, which may have slightly different performance characteristics
2. **Memory Usage**: Returned data is stored in context memory as `map[string]any`
3. **Network Traffic**: Only request columns you actually need in the RETURNING clause

## Database Compatibility

- **PostgreSQL**: Full support for all RETURNING features
- **SQLite**: Limited support (INSERT only)
- **MySQL**: No native RETURNING support
- **SQL Server**: Use OUTPUT clause instead

## Best Practices

1. **Meaningful Names**: Always use `WithName()` for better debugging
2. **Selective Columns**: Only return columns you need
3. **Guard Usage**: Use appropriate guards for data integrity
4. **Transaction Context**: Use within transactions for consistency
5. **Error Handling**: Check for data existence before type assertions

## Testing

Test RETURNING functionality with the provided test utilities:

```go
func TestMyReturningFlow(t *testing.T) {
    handler := flow.NewHandler(mockRegCtx, "test-returning").
        ExecReturning("INSERT INTO test_table (name) VALUES (?) RETURNING id, name", "test").
        SaveAs("result")
    
    // Test execution and data capture
    // Implementation depends on your test setup
}
```

## Migration Guide

### From Traditional EXEC
```go
// Before: Only get affected rows count
ExecSql("INSERT INTO users (name) VALUES (?)", name).
AffectOne().
Done()

// After: Get generated ID and other data
ExecReturning("INSERT INTO users (name) VALUES (?) RETURNING id, created_at", name).
AffectOne().
SaveAs("new_user").
Done()
```

### From Separate Query
```go
// Before: Two separate operations
ExecSql("INSERT INTO users (name) VALUES (?)", name).
AffectOne().
Done().
QuerySql("SELECT id, created_at FROM users WHERE name = ? ORDER BY id DESC LIMIT 1", name).
SaveAs("new_user")

// After: Single atomic operation
ExecReturning("INSERT INTO users (name) VALUES (?) RETURNING id, created_at", name).
AffectOne().
SaveAs("new_user").
Done()
```

This RETURNING support makes Lokstra Flow operations more efficient and provides better data flow capabilities for complex business operations.
