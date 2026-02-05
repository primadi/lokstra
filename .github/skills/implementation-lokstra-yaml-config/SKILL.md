---
name: implementation-lokstra-yaml-config
description: Create and manage multi-file YAML configuration in configs/ folder. Auto-merge multiple YAML files, environment variable substitution, service definitions, and deployment configurations. Use after design-lokstra-schema-design to set up application configuration.
phase: implementation
order: 2
license: MIT
compatibility:
  lokstra_version: ">=0.1.0"
  go_version: ">=1.18"
---

# Implementation: YAML Configuration Management

## When to Use

Use this skill when:
- Setting up configuration files for new project
- Adding service definitions (database, cache, etc.)
- Managing environment-specific configs
- Using interface-based injection (config-selected implementations)
- Configuring deployment topologies

Prerequisites:
- ✅ Project structure created
- ✅ main.go setup complete (see: implementation-lokstra-init-framework)
- ✅ Service definitions planned from design specs

## Quick Reference

### Minimal Config
```yaml
configs:
  app:
    name: "MyApp"

service-definitions:
  db-main:
    type: dbpool_pg
    config:
      dsn: "postgres://localhost:5432/mydb"

deployments:
  development:
    servers:
      api:
        base-url: "http://localhost"
        addr: ":8080"
        published-services: [my-handler]
```

### Key Concepts
- **configs**: Global key-value configuration
- **service-definitions**: Service instances including database pools (type: dbpool_pg)
- **middleware-definitions**: Reusable middleware configurations
- **router-definitions**: Custom router configurations
- **deployments**: Server topology and service distribution
- **config-overrides**: Deployment-specific config values

### Configuration Injection
```go
// @Inject "cfg:app.timeout", "30s"
Timeout time.Duration

// @Inject "@repositories.user-implementation"  
UserRepo UserRepository  // From config value
```

---

## Multi-File Configuration Pattern

### Project Structure

```
myapp/
├── config.yaml              # Base config (global defaults)
├── config/                  # Additional configs folder
│   ├── database.yaml        # Database-specific
│   ├── cache.yaml           # Cache configuration
│   ├── user-module.yaml     # User module config
│   └── auth-module.yaml     # Auth module config
└── main.go
```

**Configuration Loading:**
- By default, `config.yaml` in the root is loaded
- Additional YAML files can be loaded from `config/` folder
- All files are auto-merged in alphabetical order
- Later files override earlier ones for duplicate keys
- Enable with `lokstra_init.BootstrapAndRun(lokstra_init.WithConfigPath("config.yaml", "config"))`

### Base Configuration (config.yaml)

```yaml
# yaml-language-server: $schema=https://primadi.github.io/lokstra/schema/lokstra.schema.json

# Global configurations
configs:
  # Interface-based injection (select implementation via config)
  repositories:
    user-implementation: "postgres-user-repository"  # ← Switch here for different DB
    order-implementation: "postgres-order-repository"
  
  # Application-wide settings
  app:
    name: "MyApp"
    timeout: "30s"
    environment: "development"  # Can be overridden by LOKSTRA_DEPLOYMENT env var
    debug: true
    
  # JWT/Auth configuration
  jwt:
    secret: "dev-secret"  # Use config-overrides in deployment for prod
    issuer: "myapp.local"

# Service definitions (infrastructure)
service-definitions:
  # Database pool service (PostgreSQL)
  db-main:
    type: dbpool_pg
    config:
      dsn: "postgres://postgres:admin@localhost:5432/mydb"
      schema: "public"
      
  # Redis cache service
  cache-main:
    type: redis
    config:
      url: "redis://localhost:6379"
      
  # User repository with db dependency
  postgres-user-repository:
    type: postgres-user-repository
    depends-on: [db-main]  # Automatically injects DbPool

# Deployment configuration
deployments:
  development:
    servers:
      api:
        base-url: "http://localhost"
        addr: ":8080"
        published-services: [user-handler, auth-handler]
        
  production:
    # Override sensitive configs per deployment
    config-overrides:
      jwt:
        secret: "${JWT_SECRET}"  # From environment variable
      app:
        debug: false
        environment: "production"
    
    servers:
      api:
        base-url: "https://api.myapp.com"
        addr: "0.0.0.0:8080"
        published-services: [user-handler, auth-handler]
```

---

## Environment Variables

### Recommended Pattern: Config Overrides

Use `config-overrides` in deployment sections for environment-specific values:

```yaml
configs:
  jwt:
    secret: "dev-default"  # Safe default for development

deployments:
  production:
    config-overrides:
      jwt:
        secret: "${JWT_SECRET}"  # From env var in production only
```

### Direct Environment Variable Substitution

You can also use `${VAR_NAME}` syntax directly (but less flexible):

```yaml
configs:
  database:
    url: "${DATABASE_URL}"  # Must be set or error
  redis:
    url: "${REDIS_URL:-redis://localhost:6379}"  # Shell-style default
```

### Runtime Usage

```bash
# Development (uses config defaults)
go run .

# Production (with deployment and env vars)
export JWT_SECRET="prod-secret-key"
export LOKSTRA_DEPLOYMENT="production"
go run .
```

---

## Config-Based Interface Injection

### Problem

Multiple implementations of same interface (PostgreSQL vs MySQL):

```go
type UserRepository interface {
    GetUser(id string) (*User, error)
    SaveUser(user *User) error
}
```

### Solution

Select implementation via config without code changes:

**Infrastructure services (infrastructure/postgres_user_repository.go):**

```go
package infrastructure

import (
    "context"
    "github.com/primadi/lokstra/serviceapi"
    "myapp/domain"
)

// @Service "postgres-user-repository"
type PostgresUserRepository struct {
    // @Inject "db-main"
    Pool serviceapi.DbPool  // Auto-injected from service-definitions
}

var _ domain.UserRepository = (*PostgresUserRepository)(nil)

func (r *PostgresUserRepository) GetUser(ctx context.Context, id string) (*domain.User, error) {
    query := `SELECT id, name, email FROM users WHERE id = $1`
    var user domain.User
    err := r.Pool.QueryRow(ctx, query, id).Scan(&user.ID, &user.Name, &user.Email)
    return &user, err
}
```

**Business handler (application/user_handler.go):**

```go
// @Handler name="user-handler", prefix="/api/users"
type UserHandler struct {
    // @Inject "@repositories.user-implementation"
    UserRepo domain.UserRepository  // Injected from config!
}

// @Route "GET /{id}"
func (h *UserHandler) GetByID(id string) (*User, error) {
    return h.UserRepo.GetByID(id)
}
```

**config.yaml:**

```yaml
configs:
  repositories:
    user-implementation: "postgres-user-repository"  # ← Change here to switch DB!

service-definitions:
  # Database pool
  db-main:
    type: dbpool_pg
    config:
      dsn: "postgres://localhost:5432/mydb"
      schema: "public"

  # Repository implementations
  postgres-user-repository:
    type: postgres-user-repository
    depends-on: [db-main]  # Auto-injects DbPool
  
  # Alternative MySQL implementation (switch by changing config above)
  mysql-user-repository:
    type: mysql-user-repository
    depends-on: [db-main]
```

**Benefit:** Change database implementation with 1 line - no code recompilation needed!

---

## Multi-Environment Setup

### Separate deployment configs

```yaml
deployments:
  development:
    servers:
      api:
        base-url: "http://localhost"
        addr: ":8080"
        published-services: [user-handler]
        
  staging:
    config-overrides:
      app:
        environment: "staging"
    servers:
      api:
        base-url: "https://staging.myapp.com"
        addr: ":8080"
        published-services: [user-handler]
        
  production:
    config-overrides:
      app:
        environment: "production"
        debug: false
      jwt:
        secret: "${JWT_SECRET}"  # From environment
    servers:
      api:
        base-url: "https://api.myapp.com"
        addr: "0.0.0.0:8080"
        published-services: [user-handler]
```

Run specific deployment:

```bash
# Set deployment via environment variable
export LOKSTRA_DEPLOYMENT=production
go run .

# Or use command line flag
go run . --deployment=production
```

---

## Module-Specific Configs

### File: config/user-module.yaml

```yaml
configs:
  user:
    max-password-length: 128
    min-password-length: 8
    require-email-verification: true
    password-reset-ttl: "24h"

service-definitions:
  user-validator:
    type: user-validator
```

### Access in Code

```go
package application

import (
    "github.com/primadi/lokstra/lokstra_registry"
    "time"
)

// @Handler name="user-handler", prefix="/api/users"
type UserHandler struct {
    // Inject config values directly
    // @Inject "cfg:user.max-password-length", "128"
    MaxPasswordLength int
    
    // @Inject "cfg:user.require-email-verification", "true"
    RequireEmailVerification bool
    
    // @Inject "cfg:user.password-reset-ttl", "24h"
    PasswordResetTTL time.Duration
}

// Or fetch at runtime
func (h *UserHandler) ValidatePassword(password string) error {
    maxLen := lokstra_registry.GetConfigInt("user.max-password-length", 128)
    if len(password) > maxLen {
        return fmt.Errorf("password too long")
    }
    return nil
}
```

---

## Configuration Validation

### Ensure all required configs are set:

```bash
# Validate and print merged configuration
go run . --generate-only

# Check deployment structure
LOKSTRA_DEPLOYMENT=production go run . --generate-only

# Debug configuration loading
go run . --debug-config
```

---

## Configuration Loading Order

1. **Base config**: `config.yaml` in root directory
2. **Additional configs**: Files in `config/` folder (alphabetical order)
3. **Merge**: Later files override earlier ones for duplicate keys
4. **Deployment selection**: Via `LOKSTRA_DEPLOYMENT` env var or `--deployment` flag
5. **Deployment overrides**: `config-overrides` section in selected deployment
6. **Environment variables**: `${VAR_NAME}` substitution happens during load

**Example:**
```yaml
# config.yaml
configs:
  app:
    name: "MyApp"
    debug: true

# config/production.yaml (loaded if exists)
configs:
  app:
    debug: false  # Overrides base config

# Selected deployment (e.g., LOKSTRA_DEPLOYMENT=production)
deployments:
  production:
    config-overrides:
      app:
        name: "MyApp-Prod"  # Final override
```

**Result**: `app.name = "MyApp-Prod"`, `app.debug = false`

---

## Advanced Configuration Patterns

### 1. Handler Configurations (SPA, Static, Proxy)

Mount additional handlers at the app level:

```yaml
deployments:
  production:
    servers:
      web:
        base-url: "https://myapp.com"
        apps:
          - addr: ":8080"
            published-services: [api-handler]
            
            # Mount Single Page Applications
            mount-spa:
              - prefix: "/admin"
                dir: "./dist/admin"
              - prefix: "/"
                dir: "./dist/landing"  # Catch-all (must be last)
            
            # Mount static file directories
            mount-static:
              - prefix: "/assets"
                dir: "./public/assets"
              - prefix: "/uploads"
                dir: "./storage/uploads"
            
            # Reverse proxy to other services
            reverse-proxies:
              - prefix: "/api/v2"
                target: "http://backend-v2:9000"
                strip-prefix: true
              - prefix: "/legacy"
                target: "http://old-system:8080"
                strip-prefix: false
```

### 2. Router Customization

Auto-generate routers with custom configuration:

```yaml
service-definitions:
  user-service:
    type: user-service-factory
    router:
      convention: rest         # rest, rpc, graphql
      resource: user           # Resource name (affects paths)
      path-prefix: /api/v1     # Prefix all routes
      middlewares: [auth, rate-limit]  # Apply to all routes
      hidden-routes: [Delete]  # Hide specific routes
```

### 3. Service Configuration and Dependencies

Pass configuration to service factories:

```yaml
service-definitions:
  email-service:
    type: email-service-factory
    depends-on:
      - smtp:smtp-client    # Named dependency
      - logger              # Unnamed (uses service name)
    config:
      from-email: "noreply@myapp.com"
      retry-count: 3
      timeout: "30s"
```

In your factory:

```go
func EmailServiceFactory(deps map[string]any, config map[string]any) any {
    return &EmailService{
        SMTP:       service.Cast[SMTPClient](deps["smtp"]),
        Logger:     service.Cast[Logger](deps["logger"]),
        FromEmail:  config["from-email"].(string),
        RetryCount: int(config["retry-count"].(float64)),
    }
}
```

### 4. External Service Definitions

Define external/remote services:

```yaml
external-service-definitions:
  payment-api:
    type: payment-client-factory
    config:
      base-url: "https://api.payment.com"
      api-key: "${PAYMENT_API_KEY}"
      timeout: "30s"
```

### 5. Multiple Database Pools

Define multiple named database pools using `type: dbpool_pg`:

```yaml
service-definitions:
  # Main application database
  db-main:
    type: dbpool_pg
    config:
      dsn: "postgres://localhost:5432/myapp"
      schema: "public"
  
  # Read replica
  db-read:
    type: dbpool_pg
    config:
      dsn: "postgres://replica:5432/myapp"
      schema: "public"
  
  # Analytics database
  db-analytics:
    type: dbpool_pg
    config:
      dsn: "postgres://localhost:5432/analytics"
      schema: "analytics"
    max-conns: 10
```

Access in code:

```go
// @Service "user-repository"
type UserRepository struct {
    // @Inject "db-main"
    MainPool serviceapi.DbPool
    
    // @Inject "db-read"
    ReadPool serviceapi.DbPool
}
```

### 6. Middleware Definitions

Define reusable middleware:

```yaml
middleware-definitions:
  cors:
    type: cors
    config:
      allowed-origins: ["https://myapp.com"]
      allowed-methods: ["GET", "POST", "PUT", "DELETE"]
      
  rate-limit:
    type: rate-limit
    config:
      requests-per-minute: 100
      
  auth:
    type: jwt-auth
    config:
      secret: "${JWT_SECRET}"
```

Apply to routes:

```yaml
service-definitions:
  api-handler:
    router:
      middlewares: [cors, rate-limit, auth]
```

---

## Best Practices

### 1. Separate Concerns

Organize configs by domain:

```
config.yaml           # Global app settings
config/
  database.yaml       # All database pools
  cache.yaml          # Redis, memcached configs
  auth.yaml           # JWT, OAuth settings
  user-module.yaml    # User module settings
  order-module.yaml   # Order module settings
```

### 2. Use Config Overrides for Environments

Keep base config with safe defaults:

```yaml
# config.yaml - Safe defaults
configs:
  app:
    debug: true
  jwt:
    secret: "dev-secret"

# Override per deployment
deployments:
  production:
    config-overrides:
      app:
        debug: false
      jwt:
        secret: "${JWT_SECRET}"
```

### 3. Document Configuration Schema

Add comments to explain each setting:

```yaml
configs:
  rate-limit:
    # Maximum requests per minute per IP
    requests-per-minute: 100
    
    # Block duration when limit exceeded (in minutes)
    block-duration: 15
```

### 4. Use Interface Injection for Flexibility

Define implementation selection in config:

```yaml
configs:
  storage:
    implementation: "s3-storage"  # Can switch to "local-storage"
  
  cache:
    implementation: "redis-cache"  # Can switch to "memory-cache"
```

```go
// @Handler name="upload-handler"
type UploadHandler struct {
    // @Inject "@storage.implementation"
    Storage StorageInterface  // Injected based on config
}
```

### 5. Validate Required Configs Early

Use `@Inject "cfg:..."` without defaults for required configs:

```go
// @Service "payment-service"
type PaymentService struct {
    // @Inject "cfg:payment.api-key"  // No default = required
    APIKey string
    
    // @Inject "cfg:payment.timeout", "30s"  // Has default = optional
    Timeout time.Duration
}
```

### 6. Keep Secrets Out of Config Files

Never commit secrets to version control:

```yaml
# ❌ BAD - Secret in config file
configs:
  jwt:
    secret: "my-super-secret-key"

# ✅ GOOD - Use environment variable
configs:
  jwt:
    secret: "${JWT_SECRET}"  # From environment
```

Set secrets via environment or secret management system:

```bash
export JWT_SECRET="actual-production-secret"
```

---

## Common Patterns

### Pattern 1: Multi-Tenant Configuration

```yaml
configs:
  tenants:
    tenant-a:
      db-schema: "tenant_a"
      max-users: 100
    tenant-b:
      db-schema: "tenant_b"
      max-users: 500
```

### Pattern 2: Feature Flags

```yaml
configs:
  features:
    new-ui: true
    beta-api: false
    experimental-search: true
```

```go
func (h *Handler) GetUsers() error {
    if lokstra_registry.GetConfigBool("features.new-ui", false) {
        return h.renderNewUI()
    }
    return h.renderOldUI()
}
```

### Pattern 3: Multi-Region Deployment

```yaml
deployments:
  us-east:
    config-overrides:
      region: "us-east-1"
    servers:
      api:
        base-url: "https://api-us.myapp.com"
        addr: ":8080"
        published-services: [api-handler]
  
  eu-west:
    config-overrides:
      region: "eu-west-1"
    servers:
      api:
        base-url: "https://api-eu.myapp.com"
        addr: ":8080"
        published-services: [api-handler]
```

---

## Troubleshooting

### Issue: Config value not found

```bash
# Check what's loaded
go run . --generate-only

# Verify deployment selection
LOKSTRA_DEPLOYMENT=production go run . --generate-only
```

### Issue: Service dependency not resolved

- Check `depends-on` array in service definition
- Ensure dependent service is defined in `service-definitions`
- Verify service name spelling (case-insensitive)

### Issue: Environment variable not substituted

- Ensure format is `"${VAR_NAME}"` (with quotes)
- Check env var is set: `echo $VAR_NAME`
- Use config-overrides for deployment-specific substitution

### Issue: Wrong deployment running

```bash
# Explicitly set deployment
export LOKSTRA_DEPLOYMENT=production
go run .

# Or use flag
go run . --deployment=production
```

---

## Complete Working Example

### Directory Structure

```
myapp/
├── main.go
├── config.yaml
├── config/
│   ├── database.yaml
│   └── cache.yaml
├── domain/
│   └── user.go
├── infrastructure/
│   └── user_repository.go
└── application/
    └── user_handler.go
```

### config.yaml

```yaml
# yaml-language-server: $schema=https://primadi.github.io/lokstra/schema/lokstra.schema.json

configs:
  app:
    name: "User Management API"
    version: "1.0.0"
  
  repositories:
    user-implementation: "postgres-user-repository"

deployments:
  development:
    servers:
      api:
        base-url: "http://localhost"
        addr: ":8080"
        published-services: [user-handler]
  
  production:
    config-overrides:
      app:
        debug: false
    servers:
      api:
        base-url: "https://api.myapp.com"
        addr: "0.0.0.0:8080"
        published-services: [user-handler]
```

### config/database.yaml

```yaml
service-definitions:
  db-main:
    type: dbpool_pg
    config:
      dsn: "postgres://postgres:admin@localhost:5432/myapp"
      schema: "public"

  postgres-user-repository:
    type: postgres-user-repository
    depends-on: [db-main]
```

### config/cache.yaml

```yaml
service-definitions:
  cache-main:
    type: redis
    config:
      url: "redis://localhost:6379"
```

### domain/user.go

```go
package domain

type User struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

type UserRepository interface {
    GetUser(id string) (*User, error)
    CreateUser(user *User) error
}
```

### infrastructure/user_repository.go

```go
package infrastructure

import (
    "context"
    "myapp/domain"
    "github.com/primadi/lokstra/serviceapi"
)

// @Service "postgres-user-repository"
type PostgresUserRepository struct {
    // @Inject "db-main"
    Pool serviceapi.DbPool
}

func (r *PostgresUserRepository) GetUser(id string) (*domain.User, error) {
    ctx := context.Background()
    query := `SELECT id, name, email FROM users WHERE id = $1`
    
    var user domain.User
    err := r.Pool.QueryRow(ctx, query, id).Scan(&user.ID, &user.Name, &user.Email)
    return &user, err
}

func (r *PostgresUserRepository) CreateUser(user *domain.User) error {
    ctx := context.Background()
    query := `INSERT INTO users (id, name, email) VALUES ($1, $2, $3)`
    
    _, err := r.Pool.Exec(ctx, query, user.ID, user.Name, user.Email)
    return err
}
```

### application/user_handler.go

```go
package application

import (
    "myapp/domain"
    "github.com/primadi/lokstra/core/request"
)

// @Handler name="user-handler", prefix="/api/users"
type UserHandler struct {
    // @Inject "@repositories.user-implementation"
    UserRepo domain.UserRepository
}

// @Route "GET /{id}"
func (h *UserHandler) GetByID(id string) (*domain.User, error) {
    return h.UserRepo.GetUser(id)
}

// @Route "POST /"
func (h *UserHandler) Create(ctx *request.Context, user *domain.User) error {
    return h.UserRepo.CreateUser(user)
}
```

### main.go

```go
package main

import (
    "log"
    "myapp/application"
    "myapp/infrastructure"
    
    "github.com/primadi/lokstra/lokstra_init"
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/services"
)

func main() {
    // Register all service factories
    lokstra_registry.RegisterServiceType("postgres-user-repository", 
        infrastructure.NewPostgresUserRepositoryFactory())
    
    // Register built-in services (dbpool, redis, etc.)
    services.RegisterAll()
    
    // Bootstrap and run
    if err := lokstra_init.BootstrapAndRun(
        lokstra_init.WithConfigPath("config.yaml", "config"),
    ); err != nil {
        log.Fatal(err)
    }
}
```

### Run the Application

```bash
# Development
go run .

# Production
export LOKSTRA_DEPLOYMENT=production
go run .

# Or
go run . --deployment=production
```

---

## Next Steps

After setting up your configuration:

1. **Initialize Framework** - Set up main.go with service registrations
   - See: [implementation-lokstra-init-framework](../implementation-lokstra-init-framework/SKILL.md)

2. **Create Handlers** - Build @Handler annotated HTTP endpoints
   - See: [implementation-lokstra-create-handler](../implementation-lokstra-create-handler/SKILL.md)

3. **Create Services** - Build @Service annotated infrastructure services
   - See: [implementation-lokstra-create-service](../implementation-lokstra-create-service/SKILL.md)

4. **Database Setup** - Create migrations for your schema
   - See: [implementation-lokstra-create-migrations](../implementation-lokstra-create-migrations/SKILL.md)

5. **Add Middleware** - Create custom middleware for cross-cutting concerns
   - See: [advanced-lokstra-middleware](../advanced-lokstra-middleware/SKILL.md)

6. **Testing** - Add comprehensive tests for your application
   - See: [advanced-lokstra-tests](../advanced-lokstra-tests/SKILL.md)

---

## Configuration Reference

### Top-Level Sections

| Section | Purpose | Required |
|---------|---------|----------|
| `configs` | Global configuration values | No |
| `middleware-definitions` | Reusable middleware configurations | No |
| `service-definitions` | Local service definitions (including database pools) | No |
| `router-definitions` | Manual router configurations | No |
| `deployments` | Deployment topologies | Yes |
| `servers` | Shorthand for single deployment (auto-creates 'default') | No |

### Database Pool Service (type: dbpool_pg)

Database pools are defined in `service-definitions` with `type: dbpool_pg`:

```yaml
service-definitions:
  db-main:
    type: dbpool_pg
    config:
      dsn: "postgres://user:pass@localhost:5432/mydb"
      schema: "public"
```

| Config Field | Type | Description | Required |
|-------|------|-------------|----------|
| `dsn` | string | PostgreSQL connection string | Yes* |
| `host` | string | Database host | No* |
| `port` | int | Database port | No* |
| `database` | string | Database name | No* |
| `username` | string | Database user | No* |
| `password` | string | Database password | No* |
| `schema` | string | PostgreSQL schema | No |
| `min-conns` | int | Minimum pool connections | No (default: 2) |
| `max-conns` | int | Maximum pool connections | No (default: 10) |
| `max-idle-time` | duration | Max connection idle time | No (default: 30m) |
| `max-lifetime` | duration | Max connection lifetime | No (default: 1h) |
| `sslmode` | string | SSL mode (disable/require) | No (default: disable) |

*Either `dsn` OR component fields (host, port, etc.) required

### Service Definition Fields

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `type` | string | Service type/factory name | Yes |
| `depends-on` | array | Dependency service names | No |
| `config` | map | Service-specific configuration | No |
| `router` | object | Router auto-generation config | No |

### Router Config Fields

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `convention` | string | rest, rpc, graphql | rest |
| `resource` | string | Resource name | Service name |
| `path-prefix` | string | Prefix all routes | / |
| `middlewares` | array | Middleware names | [] |
| `hidden-routes` | array | Routes to hide | [] |

### Deployment Fields

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `config-overrides` | map | Override global configs | No |
| `servers` | map | Server definitions | Yes |

### Server Fields

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `base-url` | string | Base URL for this server | No |
| `addr` | string | Listen address (shorthand) | No* |
| `apps` | array | App configurations | No* |
| `published-services` | array | Services to publish (shorthand) | No |

*Either `addr` OR `apps` required

### App Fields

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `addr` | string | Listen address | Yes |
| `published-services` | array | Services to publish | No |
| `routers` | array | Router names to mount | No |
| `mount-spa` | array | SPA mount configurations | No |
| `mount-static` | array | Static file configurations | No |
| `reverse-proxies` | array | Reverse proxy configurations | No |

---

## Related Skills

- [implementation-lokstra-init-framework](../implementation-lokstra-init-framework/SKILL.md) - Framework bootstrap
- [implementation-lokstra-create-handler](../implementation-lokstra-create-handler/SKILL.md) - @Handler creation
- [implementation-lokstra-create-service](../implementation-lokstra-create-service/SKILL.md) - @Service creation
