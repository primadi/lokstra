# Flow Execution Guide

## 1. How to Run Flow

### Basic Execution
```go
// Create flow context
ctx := flow.NewContext(context.Background(), dbPool, "schema_name")

// Create handler with steps  
handler := flow.NewHandler(regCtx, "my-flow").
    BeginTx().
    ExecSql("INSERT INTO users (name) VALUES (?)", "John").
    AffectOne().
    Done().
    CommitOrRollback()

// Execute flow
err := flow.ExecuteFlow(ctx, handler)
if err != nil {
    log.Fatal(err)
}
```

### With Input/Output Management
```go
// Build flow with database configuration
handler := flow.NewHandler(regCtx, "my-flow").
    SetDbService("lokstra.dbpool.main").  // Configure database service
    SetDbSchema("main").                  // Configure database schema
    BeginTx().
    ExecSql("INSERT INTO users (name) VALUES (?)", "John").
    AffectOne().
    Done().
    CommitOrRollback()

// Prepare input (only variables and context)
input := flow.NewFlowInput().
    WithContext(context.Background()).
    WithVariable("user_name", "John Doe").
    WithVariable("user_email", "john@example.com")

// Execute with input/output
output, err := flow.ExecuteFlowWithInputOutput(handler, input)
if err != nil {
    log.Fatal(err)
}

// Read results
userID := output.GetInt64("new_user_id")
userData := output.GetRowMap("user_data")
```

## 2. Input Parameter Management

### Static Parameters
```go
handler := flow.NewHandler(regCtx, "static-params").
    ExecSql("INSERT INTO users (name, email) VALUES (?, ?)", "John", "john@example.com").
    Done()
```

### Dynamic Parameters (Runtime)
```go
// Using ArgsFn for dynamic parameters
handler := flow.NewHandler(regCtx, "dynamic-params").
    ExecSql("INSERT INTO users (name, email) VALUES (?, ?)").
    ArgsFn(func(ctx *flow.Context) []any {
        name, _ := ctx.Get("user_name")
        email, _ := ctx.Get("user_email")
        return []any{name, email}
    }).
    Done()

// Or using helper function
handler := flow.NewHandler(regCtx, "dynamic-params").
    ExecSql("UPDATE users SET status = ? WHERE id = ?").
    ArgsFn(createDynamicArgsFn("status", "user_id")).
    Done()
```

### Variable Sharing Between Steps
```go
handler := flow.NewHandler(regCtx, "variable-sharing").
    ExecSql("INSERT INTO users (name) VALUES (?)", "John").
    SaveAs("insert_result").  // Save result to context variable
    QueryRowSql("SELECT LAST_INSERT_ID() as user_id").
    SaveAs("new_user_id").    // Save query result to context variable
    ExecSql("UPDATE users SET status = 'active' WHERE id = ?").
    ArgsFn(func(ctx *flow.Context) []any {
        userID, _ := ctx.Get("new_user_id")
        return []any{userID}
    }).
    Done()
```

## 3. Output Management

### Accessing Results
```go
// After flow execution, get variables saved by SaveAs()
output, err := flow.ExecuteFlowWithInputOutput(handler, input)

// Get different types of results
userID := output.GetInt64("new_user_id")
userName := output.GetString("user_name")
userData := output.GetRowMap("user_data")          // Single row result
userList := output.GetRowMaps("user_list")         // Multiple rows result

// Check if variable exists
if value, exists := output.GetVariable("optional_data"); exists {
    // Process value
}
```

### Response Flow Integration
```go
// In HTTP handler context
func MyHandler(ctx *request.Context) error {
    // Build flow
    handler := flow.NewHandler(regCtx, "user-api").
        BeginTx().
        ExecSql("INSERT INTO users (name, email) VALUES (?, ?)", "John", "john@example.com").
        SaveAs("user_creation").
        QueryRowSql("SELECT * FROM users WHERE id = LAST_INSERT_ID()").
        SaveAs("new_user").
        CommitOrRollback()
    
    // Prepare flow input
    input := flow.NewFlowInput().
        WithContext(ctx.StdContext()).
        WithDbPool(getDbPool()).
        WithSchema("main")
    
    // Execute flow
    output, err := flow.ExecuteFlowWithInputOutput(handler, input)
    if err != nil {
        return ctx.ErrorInternalServer(err.Error())
    }
    
    // Return response
    userData := output.GetRowMap("new_user")
    return ctx.WithData(userData).JSON()
}
```

## 4. Common Patterns

### Custom Logic with Do()
```go
handler := flow.NewHandler(regCtx, "custom-logic").
    QueryRowSql("SELECT balance FROM accounts WHERE id = ?", accountID).
    SaveAs("account_data").
    Do(func(ctx *flow.Context) error {
        // Custom business logic
        accountData, _ := ctx.Get("account_data")
        balance := accountData.(map[string]any)["balance"].(float64)
        
        if balance < 100.0 {
            ctx.Set("low_balance", true)
            ctx.Set("warning_message", "Low balance warning")
        }
        
        return nil
    }).
    ExecSql("INSERT INTO notifications (message) VALUES (?)", "{{.warning_message}}").
    Done()

// Integration with external services
handler := flow.NewHandler(regCtx, "external-service").
    QueryRowSql("SELECT email, amount FROM orders WHERE id = ?", orderID).
    SaveAs("order_data").
    Do(func(ctx *flow.Context) error {
        // Call external payment service
        orderData, _ := ctx.Get("order_data")
        order := orderData.(map[string]any)
        
        paymentResult, err := paymentService.Process(order["email"].(string), order["amount"].(float64))
        if err != nil {
            return fmt.Errorf("payment failed: %w", err)
        }
        
        ctx.Set("payment_id", paymentResult.ID)
        return nil
    }).
    ExecSql("UPDATE orders SET payment_id = ? WHERE id = ?", "{{.payment_id}}", orderID).
    Done()
```

### Transaction Management
```go
handler := flow.NewHandler(regCtx, "transaction-pattern").
    BeginTx().                    // Start transaction
    // ... SQL operations ...
    CommitOrRollback()           // Commit on success, rollback on error

// Or forced rollback (for testing)
handler := flow.NewHandler(regCtx, "test-pattern").
    BeginTx().
    // ... SQL operations ...
    Rollback()                   // Always rollback
```

### Error Handling with Guards
```go
handler := flow.NewHandler(regCtx, "guarded-operations").
    ExecSql("UPDATE users SET status = ? WHERE id = ?", "active", 123).
    AffectOne().                 // Must affect exactly 1 row
    Done().
    QueryRowSql("SELECT * FROM users WHERE id = ?", 123).
    EnsureExists(errors.New("user not found")).  // Must return a row
    SaveAs("user_data")
```

### Conditional Logic
```go
handler := flow.NewHandler(regCtx, "conditional-flow").
    QueryRowSql("SELECT balance FROM accounts WHERE id = ?", accountID).
    ScanTo(func(row serviceapi.Row) error {
        var balance float64
        if err := row.Scan(&balance); err != nil {
            return err
        }
        if balance < withdrawAmount {
            return errors.New("insufficient balance")
        }
        return nil
    }).
    ExecSql("UPDATE accounts SET balance = balance - ? WHERE id = ?", withdrawAmount, accountID).
    AffectOne().
    Done()
```

## 5. Best Practices

### 1. Always Use Transactions for Multi-Step Operations
```go
// ✅ Good
handler := flow.NewHandler(regCtx, "transfer").
    BeginTx().
    ExecSql("UPDATE accounts SET balance = balance - ? WHERE id = ?", amount, fromAccount).
    ExecSql("UPDATE accounts SET balance = balance + ? WHERE id = ?", amount, toAccount).
    CommitOrRollback()

// ❌ Bad - No transaction
handler := flow.NewHandler(regCtx, "transfer").
    ExecSql("UPDATE accounts SET balance = balance - ? WHERE id = ?", amount, fromAccount).
    ExecSql("UPDATE accounts SET balance = balance + ? WHERE id = ?", amount, toAccount)
```

### 2. Use Guards for Data Validation
```go
handler := flow.NewHandler(regCtx, "safe-update").
    ExecSql("UPDATE users SET email = ? WHERE id = ?", newEmail, userID).
    AffectOne().  // Ensure exactly one row was updated
    Done()
```

### 3. Meaningful Variable Names for SaveAs
```go
// ✅ Good
handler := flow.NewHandler(regCtx, "user-creation").
    ExecSql("INSERT INTO users (name) VALUES (?)", userName).
    SaveAs("new_user_id").
    QueryRowSql("SELECT * FROM users WHERE id = ?", "{{.new_user_id}}").
    SaveAs("created_user_data")

// ❌ Bad
handler := flow.NewHandler(regCtx, "user-creation").
    ExecSql("INSERT INTO users (name) VALUES (?)", userName).
    SaveAs("result1").
    QueryRowSql("SELECT * FROM users WHERE id = ?", "{{.result1}}").
    SaveAs("result2")
```

### 4. Use Helper Functions for Complex Parameter Logic
```go
// Helper function
func createUserArgsFn(nameKey, emailKey, statusKey string) func(*flow.Context) []any {
    return func(ctx *flow.Context) []any {
        name, _ := ctx.Get(nameKey)
        email, _ := ctx.Get(emailKey)
        status, exists := ctx.Get(statusKey)
        if !exists {
            status = "active"  // default value
        }
        return []any{name, email, status}
    }
}

// Usage
handler := flow.NewHandler(regCtx, "create-user").
    ExecSql("INSERT INTO users (name, email, status) VALUES (?, ?, ?)").
    ArgsFn(createUserArgsFn("user_name", "user_email", "user_status")).
    Done()
```

### 5. Step Naming for Debugging and Telemetry

#### Default Step Names
```go
// Steps get automatic names based on their type
handler := flow.NewHandler(regCtx, "auto-naming").
    BeginTx().                                    // Name: "tx.begin"
    ExecSql("INSERT INTO users VALUES (?)").      // Name: "sql.exec"
    Done().
    QueryRowSql("SELECT * FROM users").           // Name: "sql.query_row"
    SaveAs("user_data").
    Do(func(ctx *Context) error { return nil }). // Name: "custom.function"
    CommitOrRollback()                           // Name: "tx.end"
```

#### Custom Step Names  
```go
// Provide meaningful names for better debugging
handler := flow.NewHandler(regCtx, "meaningful-naming").
    BeginTx().
    ExecSql("INSERT INTO users (name) VALUES (?)", "John").
    WithName("user.create").     // Custom name: "user.create"
    Done().
    QueryRowSql("SELECT id FROM users WHERE name = ?", "John").
    WithName("user.find_by_name"). // Custom name: "user.find_by_name"
    SaveAs("user_id").
    DoNamed("user.send_welcome_email", func(ctx *Context) error {
        // Custom step name: "user.send_welcome_email"
        return nil
    }).
    CommitOrRollback()
```

#### Naming Conventions
```go
// Domain-based naming
"user.create", "user.update", "user.find_by_email"
"order.validate", "order.fulfill", "order.cancel"
"payment.process", "payment.verify", "payment.refund"

// Integration naming
"stripe.charge_card", "sendgrid.send_email"
"slack.post_message", "redis.cache_set"

// Business process naming
"checkout.validate_cart", "inventory.reserve_stock"
"shipping.calculate_rate", "audit.log_action"
```

#### Benefits of Step Naming
- **Error Messages**: "step user.validate failed: invalid email format"
- **Performance Monitoring**: Track execution time per step type
- **Debugging**: Clear step identification in logs and traces
- **Telemetry**: Detailed metrics and observability
