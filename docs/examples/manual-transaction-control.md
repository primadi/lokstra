# Manual Transaction Control

This guide covers advanced scenarios where you need explicit control over transaction commit/rollback behavior.

## Automatic Transaction Management (Default)

By default, transactions are automatically managed based on response status:

```go
// @Route "POST /"
func (s *UserService) CreateUser(ctx *request.Context, req *CreateUserRequest) error {
    ctx.BeginTransaction("main-db")
    
    // These operations join the transaction
    user, err := s.repo.Create(ctx, req)
    if err != nil {
        return err // ✅ Auto rollback
    }
    
    return ctx.Api.Ok(user) // ✅ Auto commit
}
```

**Auto-rollback triggers:**
- Handler returns `error`
- Response status code >= 400 (e.g., `ctx.Api.BadRequest()`)

**Auto-commit triggers:**
- Handler returns `nil` error
- Response status code < 400 (e.g., `ctx.Api.Ok()`)

## Manual Control

### 1. Dry-Run Mode (Rollback with 200 OK)

```go
// @Route "POST /dry-run"
func (s *UserService) DryRunCreate(ctx *request.Context, req *CreateUserRequest) error {
    ctx.BeginTransaction("main-db")
    
    // Execute all operations normally
    user, err := s.repo.Create(ctx, req)
    if err != nil {
        return err
    }
    
    s.auditRepo.Log(ctx, "user_created", user.ID)
    
    // ✅ Manual rollback - data not persisted
    ctx.RollbackTransaction("main-db")
    
    // Return 200 OK with simulation results
    return ctx.Api.Ok(map[string]any{
        "message": "Dry run successful",
        "user": user,
        "note": "No data was actually saved",
    })
}
```

### 2. Testing & Validation

```go
// @Route "POST /validate"
func (s *UserService) ValidateOperation(ctx *request.Context, req *ComplexRequest) error {
    ctx.BeginTransaction("main-db")
    
    // Test complex multi-step operation
    result1, err := s.repo.Step1(ctx, req)
    if err != nil {
        return err
    }
    
    result2, err := s.repo.Step2(ctx, result1)
    if err != nil {
        return err
    }
    
    // Validate business rules
    if !s.validateBusinessRules(result2) {
        // ✅ Manual rollback with custom response
        ctx.RollbackTransaction("main-db")
        return ctx.Api.Ok(map[string]any{
            "valid": false,
            "message": "Business rules not satisfied",
            "details": result2,
        })
    }
    
    // ✅ Manual commit
    ctx.CommitTransaction("main-db")
    
    return ctx.Api.Ok(map[string]any{
        "valid": true,
        "message": "All validations passed",
    })
}
```

### 3. Partial Success Handling

```go
// @Route "POST /batch"
func (s *UserService) BatchCreate(ctx *request.Context, req *BatchRequest) error {
    ctx.BeginTransaction("main-db")
    
    var succeeded []string
    var failed []string
    
    for _, item := range req.Items {
        if err := s.repo.Create(ctx, item); err != nil {
            failed = append(failed, item.ID)
        } else {
            succeeded = append(succeeded, item.ID)
        }
    }
    
    // Business logic: commit only if at least 80% succeeded
    successRate := float64(len(succeeded)) / float64(len(req.Items))
    
    if successRate < 0.8 {
        // ✅ Rollback all
        ctx.RollbackTransaction("main-db")
        return ctx.Api.Ok(map[string]any{
            "status": "rolled_back",
            "reason": "Success rate below threshold",
            "succeeded": succeeded,
            "failed": failed,
        })
    }
    
    // ✅ Commit partial success, auto commit because return 200
    // no need to do this:
    // ctx.CommitTransaction("main-db")
    
    return ctx.Api.Ok(map[string]any{
        "status": "committed",
        "succeeded": succeeded,
        "failed": failed,
    })
}
```

### 4. Multiple Pools with Selective Control

```go
// @Route "POST /cross-db"
func (s *UserService) CrossDatabaseOperation(ctx *request.Context, req *Request) error {
    // Start transactions on both pools
    ctx.BeginTransaction("db_auth")
    ctx.BeginTransaction("db_tenant")
    
    // Auth DB operations
    user, err := s.authRepo.Create(ctx, req.User)
    if err != nil {
        return err // Both auto-rollback
    }
    
    // Tenant DB operations
    tenant, err := s.tenantRepo.Create(ctx, req.Tenant)
    if err != nil {
        // ✅ Manually rollback auth DB first
        ctx.RollbackTransaction("db_auth")
        return err // Tenant DB also auto-rollback
    }
    
    // Business rule: Commit auth but rollback tenant
    if req.DryRunTenant {
        ctx.RollbackTransaction("db_tenant")  // ✅ Tenant rolled back
        // db_auth will auto-commit (200 OK)
    }
    
    return ctx.Api.Ok(map[string]any{
        "user": user,
        "tenant": tenant,
    })
}
```

## Best Practices

### ⚠️ CRITICAL: Avoid Manual Control with Nested Calls

**DO NOT** use manual commit/rollback if your handler calls other handlers:

```go
// ❌ DANGEROUS: Manual commit with nested handler call
func (s *ServiceA) Create(ctx *request.Context) error {
    ctx.BeginTransaction("db")
    
    s.repo.Create(ctx, data)
    
    // ❌ BAD: Manual commit here
    ctx.CommitTransaction("db")
    
    // ❌ DANGER: Calling another handler creates nested transaction
    return s.serviceB.Process(ctx)  // Transaction state corrupted!
}
```

**Why dangerous?**
- ServiceA commits transaction
- ServiceB calls `BeginTransaction()` on same context
- Transaction context still exists but already committed
- Can cause data corruption or inconsistent state

**Solution: Let auto-finalization handle it**

```go
// ✅ SAFE: Auto-finalization
func (s *ServiceA) Create(ctx *request.Context) error {
    ctx.BeginTransaction("db")
    
    s.repo.Create(ctx, data)
    
    // ✅ GOOD: No manual commit, let it auto-finalize
    return s.serviceB.Process(ctx)  // Safe!
}
```

### ✅ DO

```go
// Use manual control for explicit edge cases
func (s *Service) DryRun(ctx *request.Context) error {
    ctx.BeginTransaction("db")
    // ... operations
    ctx.RollbackTransaction("db") // Clear intent
    return ctx.Api.Ok(result)
}
```

### ❌ DON'T

```go
// Don't use manual control for normal error handling
func (s *Service) Create(ctx *request.Context) error {
    ctx.BeginTransaction("db")
    
    user, err := s.repo.Create(ctx, data)
    if err != nil {
        ctx.RollbackTransaction("db") // ❌ Unnecessary
        return err // Auto-rollback already happens
    }
    
    ctx.CommitTransaction("db") // ❌ Unnecessary
    return ctx.Api.Ok(user) // Auto-commit already happens
}
```

**Rule of thumb:** Only use manual control when automatic behavior doesn't match your use case **AND** you're in a single, isolated handler with no nested calls.

### Safe vs Unsafe Manual Control

| Scenario | Safe? | Reason |
|----------|-------|--------|
| **Single handler, dry-run** | ✅ Safe | No nested calls, transaction fully controlled |
| **Single handler, validation** | ✅ Safe | No nested calls, clear lifecycle |
| **Handler → calls another handler** | ❌ **UNSAFE** | Nested context, transaction state corrupted |
| **Service layer shared by handlers** | ❌ **UNSAFE** | Called from multiple contexts |
| **Middleware with manual control** | ❌ **UNSAFE** | Applied to many handlers |
| **Batch processing (no nested calls)** | ✅ Safe | Single handler, isolated logic |

### When in Doubt: Use Auto-Finalization

If you're unsure whether manual control is safe:
- **DON'T** use manual commit/rollback
- Let auto-finalization handle it
- Manual control is for **rare edge cases only**

## Transaction Lifecycle

```
BeginTransaction(poolName)
    ↓
Repository finalizer in map[poolName]
    ↓
Handler executes
    ↓
┌─────────────────────────────────┐
│ Manual Control (Optional)       │
│ - RollbackTransaction(poolName) │
│ - CommitTransaction(poolName)   │
└─────────────────────────────────┘
    ↓
FinalizeResponse()
    ↓
Auto-finalize remaining transactions
(skip already manually finalized)
```

## Common Patterns

| Use Case | Pattern | Example |
|----------|---------|---------|
| **Normal CRUD** | Auto (default) | Return error → rollback, return success → commit |
| **Dry-Run** | Manual rollback + 200 OK | `ctx.RollbackTransaction()` then `ctx.Api.Ok()` |
| **Validation Test** | Manual based on validation | Rollback if invalid, commit if valid |
| **Batch Processing** | Conditional manual control | Commit if threshold met, rollback otherwise |
| **Audit Logs** | Split transaction | Use `serviceapi.WithoutTransaction()` for audit |

## See Also

- [Transaction Guide](../02-framework-guide/08-database-pools.md#transaction-management)
- [Database Pools](../03-api-reference/06-services/dbpool-manager.md)
- [Request Context API](../03-api-reference/02-core/request-context.md)
