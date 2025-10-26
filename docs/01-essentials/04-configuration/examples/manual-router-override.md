# Manual Router Configuration Example

This example demonstrates how to register manual routers in code and apply middleware overrides via YAML configuration.

## Scenario

You have an admin panel with custom routing logic that you want to configure differently per environment:
- **Development**: No auth, verbose logging
- **Staging**: Basic auth, request logging
- **Production**: Strong auth, audit logging, rate limiting

## Code: Router Registration

```go
package main

import (
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/lokstra_registry"
)

// setupAdminRouter creates the admin router with base configuration
func setupAdminRouter() router.Router {
	r := router.New("")  // Empty prefix - will be set via YAML config
	
	// Base middlewares (applied in all environments)
	r.Use("recovery")  // Always recover from panics
	
	// Admin routes (relative paths) with explicit names
	r.GET("/dashboard", handler.ShowDashboard, route.WithNameOption("showDashboard"))
	r.GET("/users", handler.ListUsers)  // Auto-generated name: "GET_/users"
	r.POST("/users", handler.CreateUser, route.WithNameOption("createUser"))
	r.PUT("/users/:id", handler.UpdateUser)  // Auto: "PUT_/users/:id"
	r.DELETE("/users/:id", handler.DeleteUser)  // Auto: "DELETE_/users/:id"
	
	r.GET("/settings", handler.GetSettings)
	r.PUT("/settings", handler.UpdateSettings)
	
	r.GET("/logs", handler.ViewLogs)
	r.GET("/metrics", handler.ViewMetrics)
	
	return r
}

// setupPublicRouter creates the public API router
func setupPublicRouter() router.Router {
	r := router.New("/api/v1")
	
	// Base middlewares
	r.Use("recovery", "cors")
	
	// Public routes
	r.GET("/health", handler.HealthCheck)
	r.GET("/version", handler.GetVersion)
	
	return r
}

func init() {
	// Register manual routers
	lokstra_registry.RegisterRouter("admin-router", setupAdminRouter())
	lokstra_registry.RegisterRouter("public-router", setupPublicRouter())
}
```

## Configuration: Development

```yaml
# config/development.yaml

middleware-definitions:
  logger-verbose:
    type: logger
    config:
      level: DEBUG
      colorize: true
      show_request_body: true
      show_response_body: true

router-definitions:
  admin-router:
    path-prefix: /admin  # Development path
    middlewares:
      - logger-verbose  # Verbose logging for debugging

  public-router:
    path-prefix: /api
    middlewares:
      - logger-verbose

deployments:
  development:
    servers:
      dev-server:
        base-url: http://localhost:8080
        addr: ":8080"
        routers:
          - admin-router
          - public-router
```

**Result (Development):**
- Admin routes: `http://localhost:8080/admin/dashboard`, `/admin/users`, etc.
- Public routes: `http://localhost:8080/api/health`, `/api/version`
- Admin middleware: `recovery` → `logger-verbose` → handler
- Public middleware: `recovery` → `cors` → `logger-verbose` → handler
- No authentication in development for easier testing

## Configuration: Staging

```yaml
# config/staging.yaml

middleware-definitions:
  basic-auth:
    type: basic-auth
    config:
      users:
        admin: "${ADMIN_PASSWORD}"
        developer: "${DEV_PASSWORD}"

  request-logger:
    type: logger
    config:
      level: INFO
      log_request_headers: true

  request-dumper:
    type: request-dumper
    config:
      output_file: "/var/log/staging-requests.log"

router-definitions:
  admin-router:
    path-prefix: /api/staging/admin  # Staging API path
    middlewares:
      - basic-auth        # Simple auth for staging
      - request-logger    # Log all requests
      - request-dumper    # Dump for debugging

  public-router:
    path-prefix: /api/staging
    middlewares:
      - request-logger

deployments:
  staging:
    servers:
      staging-server:
        base-url: https://staging.example.com
        addr: ":443"
        routers:
          - admin-router
          - public-router
```

**Result (Staging):**
- Admin routes: `https://staging.example.com/api/staging/admin/dashboard`, etc.
- Public routes: `https://staging.example.com/api/staging/health`, etc.
- Admin middleware: `recovery` → `basic-auth` → `request-logger` → `request-dumper` → handler
- Public middleware: `recovery` → `cors` → `request-logger` → handler

## Configuration: Production

```yaml
# config/production.yaml

middleware-definitions:
  admin-auth:
    type: jwt-auth
    config:
      secret: "${JWT_SECRET}"
      required_role: "admin"
      token_header: "Authorization"

  audit-log:
    type: audit-logger
    config:
      destination: "database"
      log_table: "admin_audit_logs"
      log_user: true
      log_ip: true
      log_request: true
      log_response: true

  rate-limiter-admin:
    type: rate-limiter
    config:
      requests_per_minute: 60
      burst: 10
      by_user: true

  rate-limiter-public:
    type: rate-limiter
    config:
      requests_per_minute: 300
      burst: 50
      by_ip: true

  logger-prod:
    type: logger
    config:
      level: WARN
      output: "json"

router-definitions:
  admin-router:
    path-prefix: /api/v2/admin  # Production API v2 path
    middlewares:
      - admin-auth           # Strong JWT auth
      - audit-log            # Log all admin actions
      - rate-limiter-admin   # Prevent abuse
      - logger-prod          # Production logging
    
    # Route-level overrides
    custom:
      - name: showDashboard
        middlewares:
          - dashboard-metrics  # Extra middleware for dashboard
      
      - name: createUser
        method: POST  # Ensure POST (redundant but explicit)
        path: /users/new  # Change path for this route only
        middlewares:
          - user-validation
          - rate-limiter-strict

  public-router:
    path-prefix: /api/v2
    middlewares:
      - rate-limiter-public  # Rate limit by IP
      - logger-prod

deployments:
  production:
    config-overrides:
      db.pool_size: 100

    servers:
      api-server-1:
        base-url: https://api-1.example.com
        addr: ":443"
        routers:
          - admin-router
          - public-router

      api-server-2:
        base-url: https://api-2.example.com
        addr: ":443"
        routers:
          - admin-router
          - public-router
```

**Result (Production):**
- Admin dashboard: `https://api-1.example.com/api/v2/admin/dashboard`
  - Middlewares: `recovery` → `admin-auth` → `audit-log` → `rate-limiter-admin` → `logger-prod` → `dashboard-metrics` → handler
- Admin create user: `https://api-1.example.com/api/v2/admin/users/new` (path changed!)
  - Middlewares: `recovery` → `admin-auth` → `audit-log` → `rate-limiter-admin` → `logger-prod` → `user-validation` → `rate-limiter-strict` → handler
- Other admin routes: `https://api-1.example.com/api/v2/admin/{route}`
  - Middlewares: `recovery` → `admin-auth` → `audit-log` → `rate-limiter-admin` → `logger-prod` → handler
- Public routes: `https://api-1.example.com/api/v2/health`, etc.
  - Middlewares: `recovery` → `cors` → `rate-limiter-public` → `logger-prod` → handler
- Strong security, comprehensive logging, rate limiting

---

## API Versioning Example

The same code can serve multiple API versions by just changing the path prefix!

```yaml
# Production API v1 (legacy)
router-definitions:
  admin-router:
    path-prefix: /api/v1/admin
  public-router:
    path-prefix: /api/v1

# Production API v2 (current)
router-definitions:
  admin-router:
    path-prefix: /api/v2/admin
  public-router:
    path-prefix: /api/v2
```

Both versions run the same code, just with different URL paths!

## Middleware Execution Order

### Admin Router (Production)
```
Request
  ↓
recovery (from code)
  ↓
admin-auth (from YAML - production)
  ↓
audit-log (from YAML - production)
  ↓
rate-limiter-admin (from YAML - production)
  ↓
logger-prod (from YAML - production)
  ↓
handler.ListUsers
  ↓
Response
```

### Key Benefits

1. **Single Router Definition**: Define routes once in code
2. **Environment-Specific Config**: Different middlewares AND paths per environment via YAML
3. **No Code Changes**: Switch environments without code changes
4. **Type Safety**: Routes defined in strongly-typed Go code
5. **Configuration Flexibility**: Middlewares and paths configured in YAML
6. **Clear Separation**: Routing logic (code) vs deployment config (YAML)
7. **API Versioning**: Easy to serve multiple API versions from same code

## Running Different Environments

```bash
# Development
export LOKSTRA_CONFIG="config/development.yaml"
go run main.go

# Staging
export LOKSTRA_CONFIG="config/staging.yaml"
go run main.go

# Production
export LOKSTRA_CONFIG="config/production.yaml"
go run main.go
```

## Comparison: Auto-Generated vs Manual Routers

### Auto-Generated Router (from Service)
```yaml
router-definitions:
  user-service-router:
    convention: rest
    resource: user
    resource-plural: users
    path-prefix: /api/v1
    middlewares: [auth, logger]
    custom:
      - name: Login
        path: /auth/login
```

**Use when:**
- Standard CRUD operations
- Convention-based routing (REST, RPC)
- Service-oriented architecture
- Auto-generated from Go structs

### Manual Router (Custom Logic)
```yaml
router-definitions:
  admin-router:
    path-prefix: /api/v2/admin  # Prefix override
    middlewares: [admin-auth, audit-log]
```

**Use when:**
- Custom routing logic
- Non-standard endpoints
- Complex route hierarchies
- Fine-grained control needed
- Static file serving
- GraphQL endpoints
- WebSocket handlers
- API versioning with same code

## See Also

- **[Router API](../03-api-reference/01-core-packages/router.md)**
- **[Middleware Configuration](../03-api-reference/05-middleware/)**
- **[Schema Documentation](../03-api-reference/03-configuration/schema.md)**
- **[Migration Guide](../MIGRATION-SCHEMA-V2.md)**
