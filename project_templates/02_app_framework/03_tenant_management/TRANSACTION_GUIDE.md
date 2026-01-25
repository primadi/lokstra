# Lazy Transaction Management Implementation

## Overview

Implementasi transaction management dengan lazy creation pattern untuk Lokstra framework. Transaction hanya dibuat ketika pertama kali ada operasi database, dan otomatis di-reuse untuk operasi selanjutnya dalam context yang sama.

## Key Features

✅ **Lazy Creation** - Transaction hanya dibuat saat pertama kali dibutuhkan  
✅ **Auto-Join** - Operasi database berikutnya otomatis join ke transaction yang sama  
✅ **Pool Name Based** - Tracking berdasarkan pool name (e.g., `"db_auth"`), bukan instance pointer  
✅ **Nested Support** - Pseudo-nested transaction dengan counter  
✅ **Auto Commit/Rollback** - Handled otomatis via defer  
✅ **Clean Service Layer** - Tidak perlu inject DbPool, cukup pakai nama pool

## Architecture

### Transaction Context Flow

```
BeginTransaction("db_auth")
    ↓
Context with TxContext marker
    ↓
First DB Operation (repository.Create)
    ↓
getExecutor checks context → finds TxContext
    ↓
Lazy create transaction (BEGIN)
    ↓
Second DB Operation (repository.Update)
    ↓
getExecutor checks context → reuses existing tx
    ↓
defer finishTx → COMMIT or ROLLBACK
```

### Components

1. **serviceapi/dbpool.go**
   - `TxContext` - Holds transaction state
   - Context keys for transaction tracking

2. **serviceapi/transaction.go**
   - `BeginTransaction(ctx, poolName)` - Mark context as needing transaction
   - `WithoutTransaction(ctx)` - Explicitly ignore parent transaction
   - `GetTransaction(ctx, poolName)` - Retrieve transaction context

3. **services/dbpool_pg/dbpool_postgres.go**
   - `getExecutor()` - Checks context and lazy creates transaction
   - Pool now has `poolName` field for transaction lookup

4. **services/dbpool_manager/dbpool_manager.go**
   - Updated to pass `poolName` when creating pool instances

## Usage Examples

### Basic Transaction (Recommended Pattern)

```go
// @Handler name="tenant-service"
type TenantService struct {
    // @Inject "@repository.tenant-repository"
    TenantRepository repository.TenantRepository
    
    // @Inject "@repository.user-repository"
    UserRepository repository.UserRepository
    
    // ❌ NO NEED to inject DbPool anymore!
}

// @Route "POST /"
func (s *TenantService) CreateTenant(ctx *request.Context, 
    req *domain.CreateTenantRequest) (result *domain.Tenant, err error) {
    
    // Mark context as needing transaction (lazy)
    newCtx, finishTx := serviceapi.BeginTransaction(ctx, "db_auth")
    defer finishTx(&err)
    
    // First operation - transaction created here
    if err := s.TenantRepository.Create(newCtx, tenant); err != nil {
        return nil, err // Auto rollback
    }
    
    // Second operation - joins same transaction
    if err := s.UserRepository.Create(newCtx, user); err != nil {
        return nil, err // Auto rollback
    }
    
    return tenant, nil // Auto commit
}
```

### Nested Transactions (Pseudo-Nested)

```go
func (s *TenantService) CreateWithAudit(ctx *request.Context, ...) (err error) {
    // Outer transaction
    ctx, finish1 := serviceapi.BeginTransaction(ctx, "db_auth")
    defer finish1(&err)
    
    // Nested transaction (just increments counter)
    ctx, finish2 := serviceapi.BeginTransaction(ctx, "db_auth")
    defer finish2(&err)
    
    s.TenantRepository.Create(ctx, tenant)
    s.UserRepository.Create(ctx, user)
    
    // Both finish functions work together
    // Only commits when counter reaches 0
    return nil
}
```

### Isolated Operations (Opt-Out)

```go
func (s *TenantService) CreateWithAuditLog(ctx *request.Context, ...) (err error) {
    ctx, finish := serviceapi.BeginTransaction(ctx, "db_auth")
    defer finish(&err)
    
    // This joins transaction
    s.TenantRepository.Create(ctx, tenant)
    
    // Audit log MUST commit even if transaction rollbacks
    isolatedCtx := serviceapi.WithoutTransaction(ctx)
    s.AuditRepository.Log(isolatedCtx, "tenant_created") // New connection
    
    // If error here, tenant rollbacks but audit committed
    return someError
}
```

### Multiple Pools

```go
func (s *Service) CrossDatabaseOperation(ctx *request.Context) (err error) {
    // Transaction for auth DB
    ctx1, finish1 := serviceapi.BeginTransaction(ctx, "db_auth")
    defer finish1(&err)
    
    // Transaction for tenant DB (independent)
    ctx2, finish2 := serviceapi.BeginTransaction(ctx1, "db_tenant")
    defer finish2(&err)
    
    s.AuthRepository.Create(ctx1, ...)   // Uses db_auth transaction
    s.TenantRepository.Create(ctx2, ...) // Uses db_tenant transaction
    
    return nil
}
```

## repository Implementation

Repositorys tidak perlu tahu tentang transaction! Tetap sederhana:

```go
// @Service "postgres-tenant-repository"
type PostgresTenantRepository struct {
    // @Inject "db_auth"
    dbPool serviceapi.DbPool
}

func (s *PostgresTenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
    query := `INSERT INTO tenants (...) VALUES (...)`
    
    // DbPool.Exec otomatis cek context dan join transaction jika ada
    _, err := s.dbPool.Exec(ctx, query, ...)
    return err
}
```

## Configuration

### config.yaml

```yaml
configs:
  repository:
    tenant-repository: postgres-tenant-repository
    user-repository: postgres-user-repository

dbpool-definitions:
  db_auth:
    dsn: ${GLOBAL_DB_DSN:postgres://postgres:adm1n@localhost:5432/lokstra_db}
    schema: ${GLOBAL_DB_SCHEMA:lokstra_auth}

servers:
  api-server:
    addr: ":3000"
    published-services: [tenant-service]
```

## Migration

SQL schema untuk tenant management:

```sql
-- migrations/001_init.sql
CREATE TABLE tenants (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    owner_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    ...
);

CREATE TABLE users (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    ...
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);
```

## Benefits

| Feature | Before | After |
|---------|--------|-------|
| **Service Code** | Inject DbPool + manual tx management | Clean, just use BeginTransaction |
| **Transaction Creation** | Eager (upfront cost) | Lazy (only when needed) |
| **Multi-repository Operations** | Manual coordination | Automatic join |
| **Nested Calls** | Complex savepoint logic | Simple counter |
| **Pool Isolation** | By instance pointer | By pool name (cleaner) |
| **repository Implementation** | Complex (check tx manually) | Simple (transparent) |

## Testing

Use the provided HTTP file:

```bash
# Start server
go run .

# Test create tenant with auto-owner creation
# See tenant-service.http for examples
```

## Implementation Notes

1. **Pool Name is Key** - Transaction tracking menggunakan pool name, bukan instance pointer
2. **Context Immutability** - Setiap BeginTransaction return new context dengan TxContext value
3. **Counter for Nesting** - Tidak support true nested transaction, hanya counter untuk koordinasi commit/rollback
4. **No Goroutine Sharing** - Transaction context TIDAK boleh di-share antar goroutine (pgx.Tx not thread-safe)
5. **Lazy is Efficient** - Transaction hanya dibuat jika ada operasi database, menghemat resource untuk read-only operations

## Future Enhancements

- [ ] Distributed transaction support (2-phase commit)
- [ ] Transaction timeout configuration
- [ ] Transaction metrics (duration, query count)
- [ ] Deadlock detection and retry
- [ ] Transaction isolation level control
