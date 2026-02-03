---
name: implementation-lokstra-create-service
description: Create @Service annotated infrastructure services. Generate repository implementations, data access objects, and service factory functions with dependency injection. Use after database schema is designed to implement data persistence layer.
phase: implementation
order: 4
license: MIT
compatibility:
  lokstra_version: ">=0.1.0"
  go_version: ">=1.18"
---

# Implementation: @Service Creation

## When to Use

Use this skill when:
- Implementing repository layer (data persistence)
- Creating infrastructure services (database, cache, email, etc.)
- Setting up service factory with custom logic
- Configuring interface-based dependency injection
- Building swappable implementations for testing/multi-tenant

Prerequisites:
- ✅ Domain models and interfaces designed (see: design-lokstra-module-requirements)
- ✅ Database schema finalized (see: design-lokstra-schema-design)
- ✅ config.yaml with service definitions ready (see: implementation-lokstra-yaml-config)

## Quick Reference

### @Service Annotation

```go
// @Service "service-name"
type MyService struct {
    // @Inject "dependency-name"
    Dependency SomeInterface
}
```

### Injection Patterns

| Pattern | Syntax | Description |
|---------|--------|-------------|
| Direct service | `@Inject "service-name"` | Inject registered service |
| Config service ref | `@Inject "@config.key"` | Service name from config |
| Config value | `@Inject "cfg:key"` | Direct config value |
| Indirect config | `@Inject "cfg:@key"` | Config value reference |

---

## Service Types

### 1. Repository Service (Database Access)

The most common service type - implements data persistence using `serviceapi.DbPool`.

**File:** `modules/user/infrastructure/postgres_user_repository.go`

```go
package infrastructure

import (
	"context"
	"fmt"

	"myapp/modules/user/domain"
	"myapp/modules/user/repository"
	"github.com/primadi/lokstra/serviceapi"
)

// @Service "postgres-user-repository"
type PostgresUserRepository struct {
	// @Inject "db-main"
	dbPool serviceapi.DbPool
}

// Compile-time interface check
var _ repository.UserRepository = (*PostgresUserRepository)(nil)

// Create inserts a new user into the database
func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, name, email, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.dbPool.Exec(ctx, query,
		user.ID, user.Name, user.Email, user.Status,
		user.CreatedAt, user.UpdatedAt,
	)
	return err
}

// Get retrieves a user by ID
func (r *PostgresUserRepository) Get(ctx context.Context, userID string) (*domain.User, error) {
	query := `
		SELECT id, name, email, status, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	user := &domain.User{}
	err := r.dbPool.QueryRow(ctx, query, userID).Scan(
		&user.ID, &user.Name, &user.Email, &user.Status,
		&user.CreatedAt, &user.UpdatedAt,
	)
	
	if r.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("user not found: %s", userID)
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Update modifies an existing user
func (r *PostgresUserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET name = $1, email = $2, status = $3, updated_at = $4
		WHERE id = $5
	`
	result, err := r.dbPool.Exec(ctx, query,
		user.Name, user.Email, user.Status, user.UpdatedAt, user.ID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found: %s", user.ID)
	}
	return nil
}

// Delete removes a user by ID
func (r *PostgresUserRepository) Delete(ctx context.Context, userID string) error {
	result, err := r.dbPool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found: %s", userID)
	}
	return nil
}

// List retrieves all users
func (r *PostgresUserRepository) List(ctx context.Context) ([]*domain.User, error) {
	query := `
		SELECT id, name, email, status, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`
	rows, err := r.dbPool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(
			&user.ID, &user.Name, &user.Email, &user.Status,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

// GetByEmail retrieves a user by email address
func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, name, email, status, created_at, updated_at
		FROM users WHERE email = $1
	`
	user := &domain.User{}
	err := r.dbPool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Status,
		&user.CreatedAt, &user.UpdatedAt,
	)
	
	if r.dbPool.IsErrorNoRows(err) {
		return nil, nil // Not found is not an error for lookup
	}
	return user, err
}
```

**Interface Definition:** `modules/user/repository/user_repository.go`

```go
package repository

import (
	"context"
	"myapp/modules/user/domain"
)

// UserRepository defines the contract for user data access
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	Get(ctx context.Context, userID string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, userID string) error
	List(ctx context.Context) ([]*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}
```

---

## DbPool API Reference (serviceapi.DbPool)

Lokstra provides `serviceapi.DbPool` interface for database operations. This abstracts the underlying database driver (pgx for PostgreSQL).

### Core Methods

```go
// DbPool interface - main database operations
type DbPool interface {
	// Acquire gets a connection from the pool
	Acquire(ctx context.Context) (DbConn, error)
	
	// Direct query methods (auto-acquire/release connection)
	Exec(ctx context.Context, query string, args ...any) (CommandResult, error)
	Query(ctx context.Context, query string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) Row
	
	// Helper methods
	SelectOne(ctx context.Context, query string, args []any, dest ...any) error
	SelectMustOne(ctx context.Context, query string, args []any, dest ...any) error
	SelectOneRowMap(ctx context.Context, query string, args ...any) (RowMap, error)
	SelectManyRowMap(ctx context.Context, query string, args ...any) ([]RowMap, error)
	
	// Error checking
	IsErrorNoRows(err error) bool
	IsExists(ctx context.Context, query string, args ...any) (bool, error)
	
	// Transaction support
	Transaction(ctx context.Context, fn func(tx DbExecutor) error) error
}
```

### Usage Examples

```go
// Simple query with auto-connection management
func (r *UserRepo) GetByID(ctx context.Context, id string) (*User, error) {
	user := &User{}
	err := r.dbPool.QueryRow(ctx, 
		"SELECT id, name FROM users WHERE id = $1", id,
	).Scan(&user.ID, &user.Name)
	
	if r.dbPool.IsErrorNoRows(err) {
		return nil, ErrUserNotFound
	}
	return user, err
}

// Using SelectMustOne (returns error if no rows)
func (r *UserRepo) GetByIDStrict(ctx context.Context, id string) (*User, error) {
	user := &User{}
	err := r.dbPool.SelectMustOne(ctx,
		"SELECT id, name FROM users WHERE id = $1",
		[]any{id}, &user.ID, &user.Name,
	)
	return user, err
}

// Using SelectOneRowMap (returns map[string]any)
func (r *UserRepo) GetUserMap(ctx context.Context, id string) (map[string]any, error) {
	return r.dbPool.SelectOneRowMap(ctx,
		"SELECT id, name, email FROM users WHERE id = $1", id,
	)
}

// Check existence
func (r *UserRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	return r.dbPool.IsExists(ctx,
		"SELECT 1 FROM users WHERE email = $1", email,
	)
}
```

---

## Transaction Support

### Using Transaction Method

```go
func (r *PostgresUserRepository) CreateWithAudit(ctx context.Context, user *domain.User) error {
	return r.dbPool.Transaction(ctx, func(tx serviceapi.DbExecutor) error {
		// Insert user
		_, err := tx.Exec(ctx, `
			INSERT INTO users (id, name, email) VALUES ($1, $2, $3)
		`, user.ID, user.Name, user.Email)
		if err != nil {
			return err // Automatic rollback
		}
		
		// Insert audit log
		_, err = tx.Exec(ctx, `
			INSERT INTO audit_logs (entity, action, entity_id, created_at)
			VALUES ('user', 'create', $1, NOW())
		`, user.ID)
		if err != nil {
			return err // Automatic rollback
		}
		
		return nil // Automatic commit
	})
}
```

### Manual Transaction Control

```go
func (r *PostgresUserRepository) TransferBalance(ctx context.Context, fromID, toID string, amount float64) error {
	// Acquire connection for explicit transaction control
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	
	// Start transaction
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	
	// Deduct from source
	_, err = tx.Exec(ctx, 
		"UPDATE accounts SET balance = balance - $1 WHERE id = $2",
		amount, fromID,
	)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}
	
	// Add to destination
	_, err = tx.Exec(ctx,
		"UPDATE accounts SET balance = balance + $1 WHERE id = $2",
		amount, toID,
	)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}
	
	return tx.Commit(ctx)
}
```

---

## Multiple Implementations (Swappable)

Create different implementations of the same interface for different databases or environments.

**PostgreSQL Implementation:**

```go
// @Service "postgres-user-repository"
type PostgresUserRepository struct {
	// @Inject "db-main"
	dbPool serviceapi.DbPool
}

var _ repository.UserRepository = (*PostgresUserRepository)(nil)

func (r *PostgresUserRepository) Get(ctx context.Context, id string) (*domain.User, error) {
	// PostgreSQL-specific implementation
}
```

**In-Memory Implementation (for testing):**

```go
// @Service "inmemory-user-repository"
type InMemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]*domain.User
}

var _ repository.UserRepository = (*InMemoryUserRepository)(nil)

// Init initializes the in-memory store
func (r *InMemoryUserRepository) Init() error {
	r.users = make(map[string]*domain.User)
	return nil
}

func (r *InMemoryUserRepository) Get(ctx context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	user, exists := r.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", id)
	}
	return user, nil
}
```

**Select via config (no code change required):**

```yaml
# config.yaml
configs:
  repositories:
    user-implementation: "postgres-user-repository"  # or "inmemory-user-repository"

service-definitions:
  postgres-user-repository:
    type: postgres-user-repository
    depends-on: [db-main]
    
  inmemory-user-repository:
    type: inmemory-user-repository
```

**Usage in handler:**

```go
// @Handler name="user-handler", prefix="/api/users"
type UserHandler struct {
	// @Inject "@repositories.user-implementation"
	UserRepo repository.UserRepository  // Injected based on config!
}
```

---

## Service Initialization

Services can implement an `Init()` method that is called after dependency injection.

```go
// @Service "user-repository"
type UserRepositoryImpl struct {
	mu      sync.RWMutex
	users   map[int]*domain.User
	nextID  int
	byEmail map[string]*domain.User
}

// Init is automatically called after all dependencies are injected
func (r *UserRepositoryImpl) Init() error {
	r.users = make(map[int]*domain.User)
	r.byEmail = make(map[string]*domain.User)
	r.nextID = 1

	// Seed initial data
	seedUsers := []*domain.User{
		{Name: "Admin User", Email: "admin@example.com", Status: "active"},
		{Name: "John Doe", Email: "john@example.com", Status: "active"},
	}

	for _, u := range seedUsers {
		r.Create(u)
	}

	return nil
}
```

---

## Built-in Infrastructure Services

### Database Pool (PostgreSQL - dbpool_pg)

**config.yaml:**

```yaml
service-definitions:
  db-main:
    type: dbpool_pg
    config:
      dsn: "${DATABASE_URL:postgres://postgres:admin@localhost:5432/mydb}"
      schema: "public"
      min-cons: 2
      max-cons: 10
      max-idle-time: "30m"
      max-lifetime: "1h"
      sslmode: "disable"
```

**Register in main.go:**

```go
import "github.com/primadi/lokstra/services/dbpool_pg"

func main() {
	lokstra.Bootstrap()
	dbpool_pg.Register()  // Register dbpool_pg service type
	lokstra_init.BootstrapAndRun()
}
```

**Config Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `dsn` | string | - | Full PostgreSQL connection string |
| `host` | string | localhost | Database host (if no DSN) |
| `port` | int | 5432 | Database port |
| `database` | string | postgres | Database name |
| `username` | string | postgres | Username |
| `password` | string | - | Password |
| `schema` | string | public | Default schema |
| `min-cons` | int | 0 | Minimum pool connections |
| `max-cons` | int | 4 | Maximum pool connections |
| `max-idle-time` | duration | 30m | Max idle time before closing |
| `max-lifetime` | duration | 1h | Max connection lifetime |
| `sslmode` | string | disable | SSL mode |
| `rls-context` | map | {} | Row-Level Security context |

### Email Service (SMTP)

**config.yaml:**

```yaml
service-definitions:
  email-smtp:
    type: email_smtp
    config:
      host: "${SMTP_HOST:smtp.gmail.com}"
      port: ${SMTP_PORT:587}
      username: "${SMTP_USER}"
      password: "${SMTP_PASS}"
      from_email: "noreply@myapp.com"
      from_name: "MyApp"
      use_starttls: true
```

**Register in main.go:**

```go
import "github.com/primadi/lokstra/services/email_smtp"

func main() {
	lokstra.Bootstrap()
	email_smtp.Register()
	lokstra_init.BootstrapAndRun()
}
```

**Usage in service:**

```go
// @Service "notification-service"
type NotificationService struct {
	// @Inject "email-smtp"
	EmailSender serviceapi.EmailSender
}

func (s *NotificationService) SendWelcome(ctx context.Context, email, name string) error {
	return s.EmailSender.Send(ctx, &serviceapi.EmailMessage{
		To:      []string{email},
		Subject: "Welcome to MyApp!",
		Body:    fmt.Sprintf("Hello %s, welcome!", name),
	})
}
```

### Key-Value Store (PostgreSQL)

**config.yaml:**

```yaml
service-definitions:
  kvstore:
    type: kvrepository_postgres
    depends-on: [db-main]
    config:
      prefix: "app"  # Optional key prefix
```

**Register in main.go:**

```go
import "github.com/primadi/lokstra/services/kvstore/kvstore_postgres"

func main() {
	lokstra.Bootstrap()
	kvstore_postgres.Register()
	lokstra_init.BootstrapAndRun()
}
```

**Usage:**

```go
// @Service "cache-service"
type CacheService struct {
	// @Inject "kvstore"
	KV serviceapi.KvRepository
}

func (s *CacheService) SetUserSession(ctx context.Context, userID string, data any) error {
	return s.KV.Set(ctx, "session:"+userID, data, 24*time.Hour)
}

func (s *CacheService) GetUserSession(ctx context.Context, userID string, dest any) error {
	return s.KV.Get(ctx, "session:"+userID, dest)
}
```

### Key-Value Store (Redis)

**config.yaml:**

```yaml
service-definitions:
  kvstore-redis:
    type: kvrepository_redis
    config:
      url: "${REDIS_URL:redis://localhost:6379}"
      database: 0
      prefix: "myapp"
```

**Register in main.go:**

```go
import "github.com/primadi/lokstra/services/kvstore/kvstore_redis"

func main() {
	lokstra.Bootstrap()
	kvstore_redis.Register()
	lokstra_init.BootstrapAndRun()
}
```

---

## Lazy Loading Services

For optional dependencies or circular dependency resolution, use `service.Cached`:

```go
import "github.com/primadi/lokstra/core/service"

// @Service "user-service"
type UserService struct {
	// Required dependency - injected immediately
	// @Inject "db-main"
	dbPool serviceapi.DbPool
	
	// Optional/lazy dependency - loaded on first use
	Logger *service.Cached[*LoggerService]
	Cache  *service.Cached[*CacheService]
}

func NewUserService() *UserService {
	return &UserService{
		Logger: service.LazyLoad[*LoggerService]("logger-service"),
		Cache:  service.LazyLoad[*CacheService]("cache-service"),
	}
}

func (s *UserService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	// Lazy load logger only when needed
	if logger := s.Logger.Get(); logger != nil {
		logger.Info("GetByID called", "id", id)
	}
	
	// Check cache first
	if cache := s.Cache.Get(); cache != nil {
		if user, found := cache.GetUser(id); found {
			return user, nil
		}
	}
	
	// Query database
	// ...
}
```

### Lazy Load Patterns

```go
// LazyLoad - load from global registry by name
cache := service.LazyLoad[*CacheService]("cache-service")

// LazyLoadWith - custom loader function
db := service.LazyLoadWith(func() *DBPool {
	return app.GetService("db-pool").(*DBPool)
})

// Value - pre-loaded value (for testing)
mockCache := service.Value[*CacheService](mockCacheInstance)

// MustGet - panics if service not found
user := s.Cache.MustGet().GetUser(id)
```

---

## Dependency Injection Patterns

### Direct Service Injection

```go
// @Service "order-service"
type OrderService struct {
	// Direct dependency on another @Service
	// @Inject "postgres-user-repository"
	UserRepo repository.UserRepository
	
	// Built-in service
	// @Inject "db-main"
	DB serviceapi.DbPool
}
```

### Config Value Injection

```go
// @Service "payment-service"
type PaymentService struct {
	// Config value injection
	// @Inject "cfg:app.timeout"
	Timeout time.Duration
	
	// @Inject "cfg:payment.api_key"
	APIKey string
	
	// @Inject "cfg:payment.sandbox_mode"
	SandboxMode bool
}
```

### Service from Config (Interface Selection)

```go
// @Service "notification-service"
type NotificationService struct {
	// Select implementation via config
	// @Inject "@providers.email"
	EmailProvider serviceapi.EmailSender
	
	// @Inject "@repositories.user"
	UserRepo repository.UserRepository
}
```

**config.yaml:**

```yaml
configs:
  providers:
    email: "email-smtp"  # Switch to "email-sendgrid" easily
  repositories:
    user: "postgres-user-repository"

service-definitions:
  email-smtp:
    type: email_smtp
    config:
      host: "${SMTP_HOST}"
      
  email-sendgrid:
    type: email_sendgrid
    config:
      api_key: "${SENDGRID_API_KEY}"
```

---

## Error Handling Patterns

### Using Custom Error Types

```go
// domain/errors.go
package domain

import "errors"

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrDuplicateEmail   = errors.New("email already exists")
	ErrInvalidInput     = errors.New("invalid input")
)

// repository implementation
func (r *PostgresUserRepository) Get(ctx context.Context, id string) (*domain.User, error) {
	user := &domain.User{}
	err := r.dbPool.QueryRow(ctx, "SELECT * FROM users WHERE id = $1", id).Scan(...)
	
	if r.dbPool.IsErrorNoRows(err) {
		return nil, domain.ErrUserNotFound
	}
	return user, err
}

// handler usage
func (h *UserHandler) GetByID(ctx *request.Context, id string) error {
	user, err := h.UserRepo.Get(ctx.Request.Context(), id)
	if errors.Is(err, domain.ErrUserNotFound) {
		return ctx.Api.NotFound("user not found")
	}
	if err != nil {
		return ctx.Api.InternalServerError("database error")
	}
	return ctx.Api.Ok(user)
}
```

### Wrapping Errors with Context

```go
func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	_, err := r.dbPool.Exec(ctx, `INSERT INTO users ...`, ...)
	if err != nil {
		// Check for duplicate key violation
		if strings.Contains(err.Error(), "unique_violation") {
			return fmt.Errorf("%w: %s", domain.ErrDuplicateEmail, user.Email)
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}
```

---

## File Organization

### Recommended Structure

```
modules/user/
├── domain/
│   ├── user.go              # Domain model
│   ├── user_dto.go          # Request/Response DTOs
│   └── errors.go            # Domain-specific errors
├── repository/
│   └── user_repository.go   # Interface definition
├── infrastructure/
│   ├── postgres_user_repository.go    # @Service PostgreSQL impl
│   ├── inmemory_user_repository.go    # @Service in-memory impl (optional)
│   └── zz_generated.lokstra.go        # Auto-generated
├── application/
│   ├── user_handler.go      # @Handler
│   └── zz_generated.lokstra.go
```

### Generated Code

The `zz_generated.lokstra.go` file is automatically generated and contains:

```go
// AUTO-GENERATED CODE - DO NOT EDIT
package infrastructure

import "github.com/primadi/lokstra/lokstra_registry"

func init() {
	RegisterPostgresUserRepository()
}

func RegisterPostgresUserRepository() {
	lokstra_registry.RegisterLazyService("postgres-user-repository", 
		func(deps map[string]any, cfg map[string]any) any {
			svc := &PostgresUserRepository{
				dbPool: deps["db-main"].(serviceapi.DbPool),
			}
			return svc
		}, map[string]any{
			"depends-on": []string{"db-main"},
		})
}
```

---

## Complete Example: Multi-Tenant Repository

```go
package infrastructure

import (
	"context"
	"fmt"
	"time"

	"myapp/modules/tenant/domain"
	"myapp/modules/tenant/repository"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

// @Service "postgres-tenant-repository"
type PostgresTenantRepository struct {
	// @Inject "db-main"
	dbPool serviceapi.DbPool
}

var _ repository.TenantRepository = (*PostgresTenantRepository)(nil)

func (r *PostgresTenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	metadata, _ := json.Marshal(tenant.Metadata)
	settings, _ := json.Marshal(tenant.Settings)

	query := `
		INSERT INTO tenants (
			id, name, domain, db_dsn, db_schema, status,
			settings, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.dbPool.Exec(ctx, query,
		tenant.ID, tenant.Name, tenant.Domain, tenant.DBDsn, tenant.DBSchema,
		tenant.Status, settings, metadata, time.Now(), time.Now(),
	)
	return err
}

func (r *PostgresTenantRepository) Get(ctx context.Context, tenantID string) (*domain.Tenant, error) {
	query := `
		SELECT id, name, domain, db_dsn, db_schema, status,
		       settings, metadata, created_at, updated_at
		FROM tenants WHERE id = $1
	`
	tenant := &domain.Tenant{}
	var metadata, settings []byte

	err := r.dbPool.QueryRow(ctx, query, tenantID).Scan(
		&tenant.ID, &tenant.Name, &tenant.Domain, &tenant.DBDsn, &tenant.DBSchema,
		&tenant.Status, &settings, &metadata, &tenant.CreatedAt, &tenant.UpdatedAt,
	)

	if r.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("tenant not found: %s", tenantID)
	}
	if err != nil {
		return nil, err
	}

	// Unmarshal JSON fields
	if len(metadata) > 0 {
		json.Unmarshal(metadata, &tenant.Metadata)
	}
	if len(settings) > 0 {
		json.Unmarshal(settings, &tenant.Settings)
	}

	return tenant, nil
}

func (r *PostgresTenantRepository) List(ctx context.Context) ([]*domain.Tenant, error) {
	query := `
		SELECT id, name, domain, status, created_at, updated_at
		FROM tenants
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`
	rows, err := r.dbPool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []*domain.Tenant
	for rows.Next() {
		t := &domain.Tenant{}
		err := rows.Scan(
			&t.ID, &t.Name, &t.Domain, &t.Status,
			&t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, t)
	}
	return tenants, rows.Err()
}

func (r *PostgresTenantRepository) SoftDelete(ctx context.Context, tenantID string) error {
	query := `UPDATE tenants SET deleted_at = NOW() WHERE id = $1`
	result, err := r.dbPool.Exec(ctx, query, tenantID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}
	return nil
}
```

---

## Next Steps

1. Create database migrations (see: implementation-lokstra-create-migrations)
2. Create @Handler endpoints (see: implementation-lokstra-create-handler)
3. Configure service definitions (see: implementation-lokstra-yaml-config)
4. Write integration tests (see: advanced-lokstra-tests)

---

## Related Skills

- [implementation-lokstra-create-handler](../implementation-lokstra-create-handler/SKILL.md) - Handler creation
- [implementation-lokstra-create-migrations](../implementation-lokstra-create-migrations/SKILL.md) - Database schema
- [implementation-lokstra-yaml-config](../implementation-lokstra-yaml-config/SKILL.md) - Service definitions
- [advanced-lokstra-tests](../advanced-lokstra-tests/SKILL.md) - Testing services
