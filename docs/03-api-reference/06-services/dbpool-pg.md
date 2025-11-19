# DbPool (PostgreSQL)

The `dbpool_pg` service provides connection pooling for PostgreSQL databases using the pgx driver. It supports multi-tenancy with schema isolation and Row-Level Security (RLS), along with convenient query methods.

## Table of Contents

- [Overview](#overview)
- [Configuration](#configuration)
- [Registration](#registration)
- [Connection Management](#connection-management)
- [Query Operations](#query-operations)
- [Transactions](#transactions)
- [Multi-Tenancy Support](#multi-tenancy-support)
- [Advanced Features](#advanced-features)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Overview

**Service Type:** `dbpool_pg`

**Interface:** `serviceapi.DbPool`

**Key Features:**

```
✓ Connection Pooling       - Efficient connection reuse
✓ Multi-Tenancy            - Schema and RLS support
✓ Transaction Support      - Manual and automatic transactions
✓ Type-Safe Queries        - Generic query helpers
✓ Row Mapping              - Automatic struct/map conversion
✓ Health Checks            - Ping and connection validation
```

## Configuration

### Config Struct

```go
type Config struct {
    // Connection using DSN string
    DSN string `json:"dsn" yaml:"dsn"`
    
    // Or individual connection parameters
    Host     string `json:"host" yaml:"host"`
    Port     int    `json:"port" yaml:"port"`
    Database string `json:"database" yaml:"database"`
    Username string `json:"username" yaml:"username"`
    Password string `json:"password" yaml:"password"`
    
    // Pool settings
    MinConnections int           `json:"min_connections" yaml:"min_connections"`
    MaxConnections int           `json:"max_connections" yaml:"max_connections"`
    MaxIdleTime    time.Duration `json:"max_idle_time" yaml:"max_idle_time"`
    MaxLifetime    time.Duration `json:"max_lifetime" yaml:"max_lifetime"`
    
    // SSL configuration
    SSLMode string `json:"sslmode" yaml:"sslmode"` // disable, require, verify-ca, verify-full
}
```

### YAML Configuration

**Using DSN:**

```yaml
services:
  main_db:
    type: dbpool_pg
    config:
      dsn: postgres://user:pass@localhost:5432/mydb?sslmode=disable&pool_max_conns=20
```

**Using Individual Parameters:**

```yaml
services:
  main_db:
    type: dbpool_pg
    config:
      host: localhost
      port: 5432
      database: myapp
      username: postgres
      password: ${DB_PASSWORD}
      
      # Pool configuration
      min_connections: 2
      max_connections: 20
      max_idle_time: 30m
      max_lifetime: 1h
      
      # SSL
      sslmode: disable
```

**Production Configuration:**

```yaml
services:
  prod_db:
    type: dbpool_pg
    config:
      host: ${DB_HOST}
      port: ${DB_PORT:5432}
      database: ${DB_NAME}
      username: ${DB_USER}
      password: ${DB_PASSWORD}
      
      # Production pool settings
      min_connections: 5
      max_connections: 50
      max_idle_time: 10m
      max_lifetime: 30m
      
      sslmode: verify-full
```

### Programmatic Configuration

```go
import (
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/services/dbpool_pg"
)

// Register service
dbpool_pg.Register()

// Create with DSN
dbPool := lokstra_registry.NewService[serviceapi.DbPool](
    "main_db", "dbpool_pg",
    map[string]any{
        "dsn": "postgres://user:pass@localhost:5432/mydb?sslmode=disable",
    },
)

// Or with individual parameters
dbPool := lokstra_registry.NewService[serviceapi.DbPool](
    "main_db", "dbpool_pg",
    map[string]any{
        "host":            "localhost",
        "port":            5432,
        "database":        "myapp",
        "username":        "postgres",
        "password":        "secret",
        "max_connections": 20,
    },
)
```

## Registration

### Basic Registration

```go
import "github.com/primadi/lokstra/services/dbpool_pg"

func init() {
    dbpool_pg.Register()
}
```

### Bulk Registration

```go
import "github.com/primadi/lokstra/services"

func main() {
    // Registers all services including dbpool_pg
    services.RegisterAllServices()
    
    // Or register only core services
    services.RegisterCoreServices()
}
```

## Connection Management

### Acquiring Connections

**Basic Connection:**

```go
import (
    "context"
    "github.com/primadi/lokstra/serviceapi"
)

ctx := context.Background()

// Acquire connection with schema
conn, err := dbPool.Acquire(ctx, "public")
if err != nil {
    log.Fatal(err)
}
defer conn.Release()

// Use connection
rows, err := conn.Query(ctx, "SELECT * FROM users")
```

**Multi-Tenant Connection:**

```go
// Acquire connection with schema AND tenant context
conn, err := dbPool.AcquireMultiTenant(ctx, "public", "tenant-123")
if err != nil {
    log.Fatal(err)
}
defer conn.Release()

// All queries will have RLS context set
// Query automatically filtered by tenant
rows, err := conn.Query(ctx, "SELECT * FROM users")
```

### Connection Interface

```go
type DbConn interface {
    // Query operations
    Exec(ctx context.Context, query string, args ...any) (CommandResult, error)
    Query(ctx context.Context, query string, args ...any) (Rows, error)
    QueryRow(ctx context.Context, query string, args ...any) Row
    
    // Convenience methods
    SelectOne(ctx context.Context, query string, args []any, dest ...any) error
    SelectMustOne(ctx context.Context, query string, args []any, dest ...any) error
    SelectOneRowMap(ctx context.Context, query string, args ...any) (RowMap, error)
    SelectManyRowMap(ctx context.Context, query string, args ...any) ([]RowMap, error)
    SelectManyWithMapper(ctx context.Context, fnScan func(Row) (any, error), 
        query string, args ...any) (any, error)
    
    // Transactions
    Begin(ctx context.Context) (DbTx, error)
    Transaction(ctx context.Context, fn func(tx DbExecutor) error) error
    
    // Utilities
    IsExists(ctx context.Context, query string, args ...any) (bool, error)
    IsErrorNoRows(err error) bool
    Ping(ctx context.Context) error
    Release() error
}
```

## Query Operations

### Execute Commands (INSERT, UPDATE, DELETE)

```go
// Insert
result, err := conn.Exec(ctx, 
    "INSERT INTO users (name, email) VALUES ($1, $2)",
    "John Doe", "john@example.com",
)
if err != nil {
    log.Fatal(err)
}
rowsAffected := result.RowsAffected()

// Update
result, err = conn.Exec(ctx,
    "UPDATE users SET email = $1 WHERE id = $2",
    "newemail@example.com", userID,
)

// Delete
result, err = conn.Exec(ctx,
    "DELETE FROM users WHERE id = $1",
    userID,
)
```

### Query Rows

**Manual Scanning:**

```go
rows, err := conn.Query(ctx, "SELECT id, name, email FROM users")
if err != nil {
    log.Fatal(err)
}
defer rows.Close()

for rows.Next() {
    var id int
    var name, email string
    if err := rows.Scan(&id, &name, &email); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("User: %d, %s, %s\n", id, name, email)
}

if err := rows.Err(); err != nil {
    log.Fatal(err)
}
```

**Query Single Row:**

```go
var user User
err := conn.QueryRow(ctx, 
    "SELECT id, name, email FROM users WHERE id = $1", 
    userID,
).Scan(&user.ID, &user.Name, &user.Email)

if err != nil {
    if conn.IsErrorNoRows(err) {
        return nil, ErrUserNotFound
    }
    return nil, err
}
```

### Convenience Methods

**SelectOne - Single Row:**

```go
var id int
var name, email string

err := conn.SelectOne(ctx,
    "SELECT id, name, email FROM users WHERE id = $1",
    []any{userID},
    &id, &name, &email,
)

if err != nil {
    if conn.IsErrorNoRows(err) {
        return nil, ErrUserNotFound
    }
    return nil, err
}
```

**SelectMustOne - Exactly One Row:**

```go
// Fails if zero or more than one row returned
err := conn.SelectMustOne(ctx,
    "SELECT id, name FROM users WHERE email = $1",
    []any{email},
    &id, &name,
)

if err != nil {
    // Returns error if no rows or multiple rows
    return nil, err
}
```

**SelectOneRowMap - Map Result:**

```go
rowMap, err := conn.SelectOneRowMap(ctx,
    "SELECT * FROM users WHERE id = $1",
    userID,
)

if err != nil {
    return nil, err
}

// Access as map
id := rowMap["id"].(int)
name := rowMap["name"].(string)
```

**SelectManyRowMap - Multiple Rows as Maps:**

```go
rows, err := conn.SelectManyRowMap(ctx,
    "SELECT id, name, email FROM users WHERE active = $1",
    true,
)

if err != nil {
    return nil, err
}

for _, row := range rows {
    fmt.Printf("User: %v, %v\n", row["id"], row["name"])
}
```

**SelectManyWithMapper - Custom Mapper:**

```go
type User struct {
    ID    int
    Name  string
    Email string
}

// Define mapper function
mapper := func(row serviceapi.Row) (any, error) {
    var user User
    err := row.Scan(&user.ID, &user.Name, &user.Email)
    return user, err
}

// Query with mapper
result, err := conn.SelectManyWithMapper(ctx, mapper,
    "SELECT id, name, email FROM users WHERE active = $1",
    true,
)

if err != nil {
    return nil, err
}

// Type assert to slice
users := result.([]User)
```

### Check Existence

```go
exists, err := conn.IsExists(ctx,
    "SELECT 1 FROM users WHERE email = $1",
    email,
)

if err != nil {
    return err
}

if exists {
    return ErrEmailAlreadyExists
}
```

## Transactions

### Manual Transaction Management

```go
// Begin transaction
tx, err := conn.Begin(ctx)
if err != nil {
    return err
}

// Execute queries
_, err = tx.Exec(ctx, 
    "INSERT INTO orders (user_id, amount) VALUES ($1, $2)",
    userID, amount,
)
if err != nil {
    tx.Rollback(ctx)
    return err
}

_, err = tx.Exec(ctx,
    "UPDATE users SET balance = balance - $1 WHERE id = $2",
    amount, userID,
)
if err != nil {
    tx.Rollback(ctx)
    return err
}

// Commit transaction
if err := tx.Commit(ctx); err != nil {
    return err
}
```

### Automatic Transaction Management

```go
// Transaction function handles commit/rollback automatically
err := conn.Transaction(ctx, func(tx serviceapi.DbExecutor) error {
    // All operations in this function are transactional
    
    _, err := tx.Exec(ctx,
        "INSERT INTO orders (user_id, amount) VALUES ($1, $2)",
        userID, amount,
    )
    if err != nil {
        return err // Triggers rollback
    }
    
    _, err = tx.Exec(ctx,
        "UPDATE users SET balance = balance - $1 WHERE id = $2",
        amount, userID,
    )
    if err != nil {
        return err // Triggers rollback
    }
    
    return nil // Triggers commit
})

if err != nil {
    log.Printf("Transaction failed: %v", err)
}
```

**Transaction Best Practices:**

```go
✓ DO: Use automatic transactions for simple cases
err := conn.Transaction(ctx, func(tx serviceapi.DbExecutor) error {
    // Transactional operations
    return nil
})

✓ DO: Return errors to trigger rollback
return fmt.Errorf("validation failed: %w", err)

✗ DON'T: Panic inside transactions
if err != nil {
    panic(err) // BAD: Use return instead
}

✗ DON'T: Commit manually in automatic transactions
return tx.Commit(ctx) // BAD: Already handled automatically
```

## Multi-Tenancy Support

### Schema Isolation

```go
// Each tenant gets its own schema
conn, err := dbPool.Acquire(ctx, "tenant_123")
if err != nil {
    return err
}
defer conn.Release()

// All queries use the tenant's schema
rows, err := conn.Query(ctx, "SELECT * FROM users")
// Queries tenant_123.users table
```

### Row-Level Security (RLS)

**Database Setup:**

```sql
-- Enable RLS on table
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

-- Create RLS policy
CREATE POLICY tenant_isolation ON users
    USING (tenant_id = current_setting('app.current_tenant')::text);
```

**Application Code:**

```go
// Acquire connection with tenant context
conn, err := dbPool.AcquireMultiTenant(ctx, "public", "tenant-123")
if err != nil {
    return err
}
defer conn.Release()

// All queries automatically filtered by tenant_id
// This query only returns users for tenant-123
users, err := conn.SelectManyRowMap(ctx, "SELECT * FROM users")
```

### Multi-Tenant Example

```go
func GetUsers(ctx context.Context, tenantID string) ([]User, error) {
    // Get database connection from registry
    dbPool := lokstra_registry.GetService[serviceapi.DbPool]("main_db")
    
    // Acquire connection with tenant context
    conn, err := dbPool.AcquireMultiTenant(ctx, "public", tenantID)
    if err != nil {
        return nil, fmt.Errorf("failed to acquire connection: %w", err)
    }
    defer conn.Release()
    
    // Define mapper
    mapper := func(row serviceapi.Row) (any, error) {
        var user User
        err := row.Scan(&user.ID, &user.TenantID, &user.Name, &user.Email)
        return user, err
    }
    
    // Query - RLS automatically filters by tenantID
    result, err := conn.SelectManyWithMapper(ctx, mapper,
        "SELECT id, tenant_id, name, email FROM users WHERE active = true",
    )
    
    if err != nil {
        return nil, fmt.Errorf("failed to query users: %w", err)
    }
    
    return result.([]User), nil
}
```

## Advanced Features

### Connection Health Checks

```go
// Ping connection
if err := conn.Ping(ctx); err != nil {
    log.Printf("Connection unhealthy: %v", err)
    return err
}

// Check connection at startup
func init() {
    dbPool := lokstra_registry.GetService[serviceapi.DbPool]("main_db")
    conn, err := dbPool.Acquire(context.Background(), "public")
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer conn.Release()
    
    if err := conn.Ping(context.Background()); err != nil {
        log.Fatal("Database ping failed:", err)
    }
    
    log.Println("Database connection established")
}
```

### Dynamic DSN Building

```go
// Config automatically builds DSN from individual parameters
cfg := &dbpool_pg.Config{
    Host:           "localhost",
    Port:           5432,
    Database:       "myapp",
    Username:       "postgres",
    Password:       "secret",
    MinConnections: 2,
    MaxConnections: 20,
    MaxIdleTime:    30 * time.Minute,
    MaxLifetime:    time.Hour,
    SSLMode:        "disable",
}

// GetFinalDSN() builds the DSN string
dsn := cfg.GetFinalDSN()
// Result: postgres://postgres:secret@localhost:5432/myapp?sslmode=disable&pool_min_conns=2&pool_max_conns=20&...
```

### Custom Settings

```go
// Get DSN setting
dsn := dbPool.GetSetting("dsn").(string)
log.Printf("Connected to: %s", dsn)
```

## Best Practices

### Connection Management

```go
✓ DO: Always release connections
conn, err := dbPool.Acquire(ctx, schema)
if err != nil {
    return err
}
defer conn.Release()  // Always use defer

✓ DO: Use context for cancellation
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
conn, err := dbPool.Acquire(ctx, schema)

✗ DON'T: Hold connections unnecessarily
conn, _ := dbPool.Acquire(ctx, schema)
time.Sleep(time.Hour)  // BAD: Holds connection too long
conn.Release()

✗ DON'T: Share connections across goroutines
// BAD: Connection is not safe for concurrent use
go func() { conn.Query(ctx, "...") }()
go func() { conn.Exec(ctx, "...") }()
```

### Query Construction

```go
✓ DO: Use parameterized queries
conn.Query(ctx, "SELECT * FROM users WHERE id = $1", userID)

✗ DON'T: Concatenate user input
query := "SELECT * FROM users WHERE name = '" + userName + "'"  // SQL injection!
conn.Query(ctx, query)

✓ DO: Check for no rows error
if err != nil {
    if conn.IsErrorNoRows(err) {
        return nil, ErrNotFound
    }
    return nil, err
}

✓ DO: Close rows when done
rows, err := conn.Query(ctx, "SELECT * FROM users")
if err != nil {
    return err
}
defer rows.Close()  // Important!
```

### Transaction Management

```go
✓ DO: Keep transactions short
err := conn.Transaction(ctx, func(tx serviceapi.DbExecutor) error {
    // Fast operations only
    _, err := tx.Exec(ctx, "UPDATE ...")
    return err
})

✗ DON'T: Do slow operations in transactions
err := conn.Transaction(ctx, func(tx serviceapi.DbExecutor) error {
    _, err := tx.Exec(ctx, "UPDATE ...")
    time.Sleep(10 * time.Second)  // BAD: Holds locks
    return err
})

✓ DO: Handle errors properly in transactions
err := conn.Transaction(ctx, func(tx serviceapi.DbExecutor) error {
    if _, err := tx.Exec(ctx, query1); err != nil {
        return fmt.Errorf("failed step 1: %w", err)
    }
    if _, err := tx.Exec(ctx, query2); err != nil {
        return fmt.Errorf("failed step 2: %w", err)
    }
    return nil
})
```

### Pool Configuration

```go
✓ DO: Configure appropriate pool sizes
config:
  min_connections: 2      # Small minimum
  max_connections: 20     # Based on workload
  max_idle_time: 30m      # Close idle connections
  max_lifetime: 1h        # Recycle connections

✗ DON'T: Set pool too large
max_connections: 1000     # BAD: Too many connections

✗ DON'T: Set pool too small
max_connections: 1        # BAD: Bottleneck under load
```

## Examples

### Complete CRUD Operations

```go
package repository

import (
    "context"
    "fmt"
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/serviceapi"
)

type UserRepository struct {
    dbPool serviceapi.DbPool
}

func NewUserRepository() *UserRepository {
    return &UserRepository{
        dbPool: lokstra_registry.GetService[serviceapi.DbPool]("main_db"),
    }
}

// Create user
func (r *UserRepository) Create(ctx context.Context, user *User) error {
    conn, err := r.dbPool.Acquire(ctx, "public")
    if err != nil {
        return err
    }
    defer conn.Release()
    
    err = conn.QueryRow(ctx,
        "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
        user.Name, user.Email,
    ).Scan(&user.ID)
    
    return err
}

// Get user by ID
func (r *UserRepository) GetByID(ctx context.Context, id int) (*User, error) {
    conn, err := r.dbPool.Acquire(ctx, "public")
    if err != nil {
        return nil, err
    }
    defer conn.Release()
    
    var user User
    err = conn.SelectOne(ctx,
        "SELECT id, name, email, created_at FROM users WHERE id = $1",
        []any{id},
        &user.ID, &user.Name, &user.Email, &user.CreatedAt,
    )
    
    if err != nil {
        if conn.IsErrorNoRows(err) {
            return nil, ErrUserNotFound
        }
        return nil, err
    }
    
    return &user, nil
}

// List users
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]User, error) {
    conn, err := r.dbPool.Acquire(ctx, "public")
    if err != nil {
        return nil, err
    }
    defer conn.Release()
    
    mapper := func(row serviceapi.Row) (any, error) {
        var user User
        err := row.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
        return user, err
    }
    
    result, err := conn.SelectManyWithMapper(ctx, mapper,
        "SELECT id, name, email, created_at FROM users ORDER BY id LIMIT $1 OFFSET $2",
        limit, offset,
    )
    
    if err != nil {
        return nil, err
    }
    
    return result.([]User), nil
}

// Update user
func (r *UserRepository) Update(ctx context.Context, user *User) error {
    conn, err := r.dbPool.Acquire(ctx, "public")
    if err != nil {
        return err
    }
    defer conn.Release()
    
    result, err := conn.Exec(ctx,
        "UPDATE users SET name = $1, email = $2 WHERE id = $3",
        user.Name, user.Email, user.ID,
    )
    
    if err != nil {
        return err
    }
    
    if result.RowsAffected() == 0 {
        return ErrUserNotFound
    }
    
    return nil
}

// Delete user
func (r *UserRepository) Delete(ctx context.Context, id int) error {
    conn, err := r.dbPool.Acquire(ctx, "public")
    if err != nil {
        return err
    }
    defer conn.Release()
    
    result, err := conn.Exec(ctx, "DELETE FROM users WHERE id = $1", id)
    if err != nil {
        return err
    }
    
    if result.RowsAffected() == 0 {
        return ErrUserNotFound
    }
    
    return nil
}
```

### Multi-Tenant Repository

```go
type TenantUserRepository struct {
    dbPool serviceapi.DbPool
}

func (r *TenantUserRepository) GetUsers(ctx context.Context, tenantID string) ([]User, error) {
    // Acquire multi-tenant connection
    conn, err := r.dbPool.AcquireMultiTenant(ctx, "public", tenantID)
    if err != nil {
        return nil, err
    }
    defer conn.Release()
    
    // Query automatically filtered by RLS
    rows, err := conn.SelectManyRowMap(ctx,
        "SELECT id, name, email FROM users WHERE active = true",
    )
    
    if err != nil {
        return nil, err
    }
    
    users := make([]User, len(rows))
    for i, row := range rows {
        users[i] = User{
            ID:    row["id"].(int),
            Name:  row["name"].(string),
            Email: row["email"].(string),
        }
    }
    
    return users, nil
}
```

## Related Documentation

- [Services Overview](index) - Service architecture and patterns
- [DbPool Manager](dbpool-manager) - Centralized pool management with multi-tenancy
- [KvStore Service](kvstore-redis) - Key-value caching
- [Configuration](../03-configuration/config) - YAML configuration system

---

**Next:** [KvStore Service](kvstore-redis) - Redis-based key-value store
