---
layout: docs
title: Schema
---

# Schema

> YAML schema definitions and validation for deployment configurations

## Overview

The `schema` package provides type definitions for YAML deployment configurations and embeds the JSON schema for validation. It defines the structure for all deployment-related configuration including services, routers, middleware, and deployment topology.

## Import Path

```go
import "github.com/primadi/lokstra/core/deploy/schema"
```

---

## Core Types

### DeployConfig
Root configuration structure for YAML files.

**Definition:**
```go
type DeployConfig struct {
    Configs                    map[string]any
    MiddlewareDefinitions      map[string]*MiddlewareDef
    ServiceDefinitions         map[string]*ServiceDef
    RouterDefinitions          map[string]*RouterDef           // Renamed from "Routers"
    Deployments                map[string]*DeploymentDefMap
    // RouterOverrides removed - overrides are now inline in RouterDef
}
```

**YAML Example:**
```yaml
configs:
  app.name: "MyApp"
  app.version: "1.0.0"

service-definitions:
  user-service:
    type: user-service-factory

router-definitions:
  user-router:
    convention: rest
    resource: user

deployments:
  production:
    servers:
      api-server:
        base-url: https://api.example.com
```

---

## Service Definitions

### ServiceDef
Defines a service instance.

**Definition:**
```go
type ServiceDef struct {
    Name      string
    Type      string         // Factory type
    DependsOn []string       // Dependencies
    Config    map[string]any // Optional config
}
```

**YAML Example:**
```yaml
service-definitions:
  user-service:
    type: user-service-factory
    depends-on:
      - db-service
      - cache-service
    config:
      max_connections: 100
      timeout: 30s
```

**Dependency Syntax:**
```yaml
# Simple dependency (uses service name as key)
depends-on:
  - db-service

# Explicit mapping (custom key)
depends-on:
  - userRepo:user-repository
  - paymentSvc:payment-service

---

## Router Definitions

### RouterDef
Defines a router auto-generated from a service, or middleware overrides for manual routers.

**Router Naming Convention:**
- Router name determines the service name using pattern `{service-name}-router`
- Example: `user-service-router` → service is `user-service`

**Manual routers** (registered via `RegisterRouter()`) can also use `router-definitions` for:
- Path prefix overrides
- Middleware injection
- Route-level customization

**Definition:**
```go
type RouterDef struct {
    // Service name is derived from router name pattern: "{service-name}-router"
    // Example: "user-service-router" → service is "user-service"
    
    Convention     string           // Convention type (rest, rpc, graphql) - for auto-generated only
    Resource       string           // Singular form (e.g., "user") - for auto-generated only
    ResourcePlural string           // Plural form (e.g., "users") - for auto-generated only
    
    // Inline overrides (works for both auto-generated AND manual routers)
    PathPrefix   string           // Path prefix override
    PathRewrites []PathRewriteDef // Regex-based path rewrite rules
    Middlewares  []string         // Router-level middleware names
    Hidden       []string         // Methods to hide (auto-generated only)
    Custom       []RouteDef       // Custom route definitions
}

type PathRewriteDef struct {
    Pattern     string // Regex pattern to match (e.g., "^/api/v1/(.*)$")
    Replacement string // Replacement string (e.g., "/api/v2/$1")
}
```

**YAML Example (Auto-Generated Router):**
```yaml
router-definitions:
  user-service-router:  # Service derived: "user-service"
    # Inline overrides
    path-prefix: /api/v1
    middlewares:
      - auth
      - logger
    custom:
      - name: Login
        method: POST
        path: /auth/login
        middlewares:
          - rate-limiter
```

**YAML Example (Manual Router Overrides):**
```yaml
# Code: Manual router registration
# r := router.New("")
# r.GET("/dashboard", handler.ShowDashboard, route.WithNameOption("showDashboard"))
# r.GET("/users", handler.ListUsers)  # Auto-name: "GET_/users"
# lokstra_registry.RegisterRouter("admin-router", r)

router-definitions:
  admin-router:  # Manual router (already registered in code)
    # Supported overrides for manual routers
    path-prefix: /api/v1/admin  # Change path prefix
    middlewares:
      - admin-auth
      - audit-log
    
    # Route-level overrides (by route name)
    custom:
      - name: showDashboard  # Update named route
        method: POST  # Change method from GET to POST
        path: /admin/main  # Change path
        middlewares:
          - extra-logging
      
      - name: GET_/users  # Update auto-named route
        middlewares:
          - rate-limiter  # Just add middlewares
    
    # Note: convention, resource, hidden are ignored for manual routers
```

**Router Naming Convention:**
- Format: `{service-name}-router`
- Service name: Remove `-router` suffix
- Examples:
  - `user-service-router` → `user-service`
  - `order-service-router` → `order-service`
  - `payment-router` → `payment`
  - `admin-router` → manual router (no auto-generation)

**Use Cases for Manual Router Overrides:**
- Apply environment-specific middlewares (auth only in prod)
- Add monitoring/logging per deployment
- Inject rate limiting from YAML configuration
- **Change path prefix per environment (API versioning)**
- **Rewrite paths using regex patterns (flexible transformations)**
- **Update specific route methods or paths per environment**
- **Add route-specific middlewares without changing code**
- Keep router logic in code, configuration in YAML

---

### Path Rewrites

**Path rewrites** allow regex-based transformation of route paths at build time. This is more flexible than `path-prefix` for complex URL transformations.

**When to use:**
- **`path-prefix`**: Simple prefix addition (e.g., add `/api/v1` to all routes)
- **`path-rewrites`**: Complex transformations (e.g., change `/api/v1/` to `/api/v2/`, add tenant IDs, etc.)

**Example: API Versioning**
```yaml
router-definitions:
  user-router:
    path-rewrites:
      - pattern: "^/api/v1/(.*)$"
        replacement: "/api/v2/$1"
```

**Result:**
```
Code:         r.GET("/api/v1/users", ...)
After rewrite: GET /api/v2/users
Internal name: GET_/api/v1/users (unchanged)
```

**Example: Multi-Tenant Paths**
```yaml
router-definitions:
  tenant-router:
    path-rewrites:
      - pattern: "^/(.*)$"
        replacement: "/tenant/${TENANT_ID}/$1"
```

**Result:**
```
Code:         r.GET("/users", ...)
After rewrite: GET /tenant/acme-corp/users
```

**Example: Multiple Rules**
```yaml
router-definitions:
  api-router:
    path-rewrites:
      - pattern: "^/v1/(.*)$"
        replacement: "/api/v2/$1"
      - pattern: "^/legacy/(.*)$"
        replacement: "/api/v2/compat/$1"
```

**Rules:**
- Applied in order (first match wins)
- Uses Go's `regexp` package
- Supports capture groups: `$1`, `$2`, etc.
- Applied at router build time (zero runtime overhead)
- Route names unchanged (only HTTP paths affected)

**Combining with path-prefix:**
```yaml
router-definitions:
  api-router:
    path-prefix: /v2       # Applied first
    path-rewrites:         # Applied after prefix
      - pattern: "^/v2/old/(.*)$"
        replacement: "/v2/new/$1"
```

---

### RouteDef
Defines a single route override.

**Definition:**
```go
type RouteDef struct {
    Name        string   // Route name (manual or auto-generated)
    Method      string   // HTTP method override
    Path        string   // Path override
    Middlewares []string // Route-level middleware names
}
```

**YAML Example:**
```yaml
custom:
  - name: Login
    method: POST
    path: /auth/login
    middlewares:  # Route-level middlewares
      - rate-limiter-strict
  - name: Logout
    method: POST
    path: /auth/logout
  - name: AdminReset
    method: POST
    path: /admin/reset
    middlewares:
      - admin-auth
      - audit-log
```

---

## Middleware Definitions

### MiddlewareDef
Defines a middleware instance.

**Definition:**
```go
type MiddlewareDef struct {
    Name   string
    Type   string         // Factory type
    Config map[string]any // Optional config
}
```

**YAML Example:**
```yaml
middleware-definitions:
  logger-debug:
    type: logger
    config:
      level: DEBUG
      colorize: true
      
  cors-dev:
    type: cors
    config:
      allow_origin: "*"
      allow_methods: "*"
      allow_headers: "*"
      
  rate-limiter-strict:
    type: rate-limiter
    config:
      requests_per_minute: 10
      burst: 5
```

---

## Configuration

### ConfigDef
Defines a configuration value.

**Definition:**
```go
type ConfigDef struct {
    Name  string
    Value any // Can be string or ${...} reference
}
```

**YAML Example:**
```yaml
configs:
  app.name: "MyApp"
  app.port: 8080
  app.env: "${APP_ENV:development}"
  db.dsn: "${DATABASE_URL}"
  log.level: "INFO"
```

---

## Deployment Topology

### DeploymentDefMap
Deployment using map structure.

**Definition:**
```go
type DeploymentDefMap struct {
    ConfigOverrides map[string]any
    Servers         map[string]*ServerDefMap
}
```

**YAML Example:**
```yaml
deployments:
  production:
    config-overrides:
      log.level: INFO
      db.pool_size: 100
    servers:
      api-server-1:
        # ...
      api-server-2:
        # ...
```

---

### ServerDefMap
Server using map structure.

**Definition:**
```go
type ServerDefMap struct {
    BaseURL string
    Apps    []*AppDefMap
    
    // Helper fields (shorthand for single app)
    HelperAddr              string
    HelperRouters           []string
    HelperPublishedServices []string
}
```

**YAML Example (Full):**
```yaml
servers:
  api-server:
    base-url: https://api.example.com
    apps:
      - addr: ":443"
        routers:
          - user-router
          - order-router
      - addr: ":8080"
        routers:
          - admin-router
```

**YAML Example (Shorthand):**
```yaml
servers:
  api-server:
    base-url: https://api.example.com
    # Shorthand for single app
    addr: ":443"
    routers:
      - user-router
      - order-router
```

---

### AppDefMap
App using map structure.

**Definition:**
```go
type AppDefMap struct {
    Addr              string
    Routers           []string
    PublishedServices []string // Auto-generate routers for these services
    ReverseProxies    []*ReverseProxyDef
    MountSpa          []*MountSpaDef
    MountStatic       []*MountStaticDef
}
```

**YAML Example:**
```yaml
apps:
  - addr: ":8080"
    routers:
      - user-router
      - order-router
      
  - addr: ":9090"
    # Auto-generate routers from services
    published-services:
      - user-service
      - order-service
      
  - addr: ":3000"
    # Handler configurations
    reverse-proxies:
      - prefix: /api/v1
        target: http://backend:8080
        strip-prefix: true
    mount-spa:
      - prefix: /admin
        dir: ./dist/admin
    mount-static:
      - prefix: /assets
        dir: ./public/assets
```

---

## Schema Validation

### GetSchemaBytes
Returns the embedded JSON schema for validation.

**Signature:**
```go
func GetSchemaBytes() []byte
```

**Example:**
```go
schemaBytes := schema.GetSchemaBytes()
// Use with JSON schema validator
```

---

## Complete Examples

### Minimal Configuration
```yaml
service-definitions:
  user-service:
    type: user-service-factory

router-definitions:
  user-service-router:  # Service derived: "user-service"
    convention: rest

deployments:
  production:
    servers:
      api-server:
        base-url: https://api.example.com
        addr: ":443"
        routers:
          - user-service-router
```

---

### Full-Featured Configuration
```yaml
# Configuration values
configs:
  app.name: "MyApp"
  app.version: "1.0.0"
  app.env: "${APP_ENV:production}"
  db.dsn: "${DATABASE_URL}"
  db.pool_size: 100

# Middleware definitions
middleware-definitions:
  logger:
    type: logger
    config:
      level: INFO
      
  auth:
    type: jwt-auth
    config:
      secret: "${JWT_SECRET}"
      
  rate-limiter:
    type: rate-limiter
    config:
      requests_per_minute: 60
      burst: 10

# Service definitions
service-definitions:
  db-service:
    type: postgres-factory
    config:
      dsn: "${@cfg:db.dsn}"
      pool_size: "${@cfg:db.pool_size}"
      
  cache-service:
    type: redis-factory
    config:
      addr: "${REDIS_URL:localhost:6379}"
      
  user-service:
    type: user-service-factory
    depends-on:
      - db-service
      - cache-service
    config:
      max_users: 10000
      
  order-service:
    type: order-service-factory
    depends-on:
      - db-service
      - user-service
    config:
      max_orders: 50000

# Router definitions
router-definitions:
  user-service-router:  # Service derived: "user-service"
    # Inline overrides
    path-prefix: /api/v1
    middlewares:
      - auth
      - logger
    custom:
      - name: Login
        method: POST
        path: /auth/login
        middlewares:
          - rate-limiter-strict
      - name: Logout
        method: POST
        path: /auth/logout

# Deployments
deployments:
  production:
    config-overrides:
      log.level: INFO
      db.pool_size: 100
    servers:
      api-server-1:
        base-url: https://api-1.example.com
        apps:
          - addr: ":443"
            routers:
              - user-service-router
              - order-service-router
      
      api-server-2:
        base-url: https://api-2.example.com
        apps:
          - addr: ":443"
            routers:
              - user-service-router
              - order-service-router
```

---

### Multi-Environment Configuration
```yaml
# Base configuration (shared)
service-definitions:
  user-service:
    type: user-service-factory

router-definitions:
  user-service-router:
    convention: rest

# Development deployment
deployments:
  development:
    config-overrides:
      log.level: DEBUG
      db.pool_size: 10
    servers:
      dev-server:
        base-url: http://localhost:8080
        addr: ":8080"
        routers:
          - user-service-router

# Staging deployment
deployments:
  staging:
    config-overrides:
      log.level: INFO
      db.pool_size: 50
    servers:
      staging-server:
        base-url: https://staging.example.com
        addr: ":443"
        routers:
          - user-service-router

# Production deployment
deployments:
  production:
    config-overrides:
      log.level: WARN
      db.pool_size: 100
    servers:
      api-server-1:
        base-url: https://api-1.example.com
        addr: ":443"
        routers:
          - user-service-router
      api-server-2:
        base-url: https://api-2.example.com
        addr: ":443"
        routers:
          - user-service-router
```

---

### Microservices Architecture
```yaml
# User service deployment
service-definitions:
  user-service:
    type: user-service-factory

deployments:
  user-deployment:
    servers:
      user-server:
        base-url: https://user-api.example.com
        addr: ":443"
        published-services:
          - user-service

---

# Order service deployment (separate file)
service-definitions:
  order-service:
    type: order-service-factory

# External user service
external-service-definitions:
  user-service:
    url: https://user-api.example.com
    type: user-service-factory

deployments:
  order-deployment:
    servers:
      order-server:
        base-url: https://order-api.example.com
        addr: ":443"
        published-services:
          - order-service
```

---

### Handler Configurations
```yaml
# Full example with reverse proxies, SPAs, and static files
service-definitions:
  api-service:
    type: api-service-factory

deployments:
  production:
    servers:
      web-server:
        base-url: https://example.com
        apps:
          # Backend API
          - addr: ":8080"
            routers:
              - api-service-router
          
          # Frontend + Gateway
          - addr: ":3000"
            # Reverse proxy to backend API
            reverse-proxies:
              - prefix: /api
                target: http://localhost:8080
                strip-prefix: false
              
              # Proxy to legacy system
              - prefix: /legacy
                target: http://legacy-system:9000
                strip-prefix: true
                rewrite:
                  path-pattern: "^/legacy/(.*)$"
                  path-replacement: "/v1/$1"
            
            # SPA applications
            mount-spa:
              - prefix: /admin
                dir: ./dist/admin
              - prefix: /
                dir: ./dist/app
            
            # Static assets
            mount-static:
              - prefix: /assets
                dir: ./public/assets
              - prefix: /uploads
                dir: ./storage/uploads
```

---

### ReverseProxyDef
HTTP reverse proxy configuration.

**Definition:**
```go
type ReverseProxyDef struct {
    Prefix      string
    StripPrefix bool
    Target      string
    Rewrite     *ReverseProxyRewriteDef
}

type ReverseProxyRewriteDef struct {
    PathPattern      string
    PathReplacement  string
    HostPattern      string
    HostReplacement  string
}
```

**YAML Example:**
```yaml
reverse-proxies:
  # Simple proxy
  - prefix: /api/v1
    target: http://backend:8080
    strip-prefix: true
    
  # Advanced rewrite
  - prefix: /api/v2
    target: http://new-backend:8080
    rewrite:
      path-pattern: "^/api/v2/(.*)$"
      path-replacement: "/v3/$1"
      host-pattern: "^api\\.(.*)$"
      host-replacement: "new-api.$1"
```

**Use Cases:**
- Proxy to backend services (microservices architecture)
- API gateway patterns
- Legacy system integration
- Dynamic path/host rewriting

---

### MountSpaDef
Single Page Application mounting configuration.

**Definition:**
```go
type MountSpaDef struct {
    Prefix string
    Dir    string
}
```

**YAML Example:**
```yaml
mount-spa:
  # Admin dashboard
  - prefix: /admin
    dir: ./dist/admin
    
  # Main application
  - prefix: /
    dir: ./dist/app
```

**Use Cases:**
- Serve React/Vue/Angular applications
- Multiple SPA deployments on one server
- Admin panels and dashboards
- Fallback to index.html for client-side routing

---

### MountStaticDef
Static file serving configuration.

**Definition:**
```go
type MountStaticDef struct {
    Prefix string
    Dir    string
}
```

**YAML Example:**
```yaml
mount-static:
  # Static assets
  - prefix: /assets
    dir: ./public/assets
    
  # User uploads
  - prefix: /uploads
    dir: ./storage/uploads
    
  # Documentation
  - prefix: /docs
    dir: ./public/docs
```

**Use Cases:**
- Serve CSS, JavaScript, images
- User-generated content (uploads)
- Documentation files
- Static resources

---

## Schema Validation Rules

### Required Fields

**ServiceDef:**
- ✅ `type` - Service factory type must be specified
- ❌ `config` - Optional
- ❌ `depends-on` - Optional

**RouterDef:**
- ❌ `service` - REMOVED (derived from router name pattern)
- ❌ `convention` - Optional (defaults to "rest")
- ❌ `resource` - Optional (auto-detected from service name)

**DeploymentDefMap:**
- ✅ `servers` - At least one server required
- ❌ `config-overrides` - Optional

**ServerDefMap:**
- ✅ `base-url` - Base URL required
- ✅ `apps` or helper fields (`addr` + `routers`/`published-services`)

---

### Field Types

**String Fields:**
- `type`, `name`, `service`, `convention`, `resource`, `url`, `addr`, `path-prefix`

**String Array Fields:**
- `depends-on`, `middlewares`, `routers`, `hidden`, `published-services`

**Map Fields:**
- `config`, `configs`, `config-overrides`

**Object Fields:**
- `service-definitions`, `routers`, `router-overrides`, `external-service-definitions`, `deployments`

---

## Best Practices

### 1. Use Meaningful Names
```yaml
# ✅ Good: Descriptive names
service-definitions:
  user-management-service:
    type: user-service-factory
  
  product-catalog-service:
    type: catalog-service-factory

# 🚫 Avoid: Cryptic names
service-definitions:
  svc1:
    type: user-service-factory
  svc2:
    type: catalog-service-factory
```

---

### 2. Group Related Configurations
```yaml
# ✅ Good: Logical grouping
configs:
  # Database settings
  db.host: "localhost"
  db.port: 5432
  db.name: "myapp"
  
  # Cache settings
  cache.host: "localhost"
  cache.port: 6379
  
  # App settings
  app.name: "MyApp"
  app.version: "1.0.0"

# 🚫 Avoid: Random order
configs:
  app.name: "MyApp"
  db.host: "localhost"
  cache.port: 6379
  db.port: 5432
```

---

### 3. Use Config References
```yaml
# ✅ Good: DRY with config references
configs:
  api.version: "v1"
  api.prefix: "/api/${@cfg:api.version}"

router-definitions:
  user-service-router:
    path-prefix: "${@cfg:api.prefix}"

# 🚫 Avoid: Duplication
router-definitions:
  user-service-router:
    path-prefix: "/api/v1"
  order-service-router:
    path-prefix: "/api/v1"
```

---

### 4. Document Dependencies Clearly
```yaml
# ✅ Good: Clear dependencies
service-definitions:
  order-service:
    type: order-service-factory
    depends-on:
      - db-service       # Database access
      - user-service     # User validation
      - payment-service  # Payment processing
      - inventory-service # Stock management

# 🚫 Avoid: Undocumented dependencies
service-definitions:
  order-service:
    type: order-service-factory
    # Hidden dependencies!
```

---

### 5. Use Inline Router Overrides When Needed
```yaml
# ✅ Good: Custom routes documented
router-definitions:
  user-service-router:
    custom:
      - name: Login
        method: POST
        path: /auth/login
      - name: Logout
        method: POST
        path: /auth/logout
      - name: RefreshToken
        method: POST
        path: /auth/refresh

# 🚫 Avoid: Overriding standard CRUD
router-definitions:
  user-service-router:
    custom:
      - name: List
        method: GET
        path: /users  # Unnecessary override
```

---

### 6. Organize Handler Configurations by Purpose
```yaml
# ✅ Good: Clear separation of concerns
apps:
  # API backend
  - addr: ":8080"
    routers:
      - api-router
  
  # Frontend gateway
  - addr: ":3000"
    reverse-proxies:
      - prefix: /api
        target: http://localhost:8080
    mount-spa:
      - prefix: /
        dir: ./dist/app
    mount-static:
      - prefix: /assets
        dir: ./public/assets

# 🚫 Avoid: Mixing everything in one app
apps:
  - addr: ":8080"
    routers:
      - api-router
    reverse-proxies:
      - prefix: /other-api
        target: http://other:9000
    mount-spa:
      - prefix: /admin
        dir: ./dist/admin
    # Too many responsibilities!
```

---

### 7. Use Reverse Proxy for Backend Integration
```yaml
# ✅ Good: Proxy pattern for microservices
reverse-proxies:
  - prefix: /api/users
    target: http://user-service:8080
    strip-prefix: false
  - prefix: /api/orders
    target: http://order-service:8080
    strip-prefix: false

# ✅ Good: Strip prefix for clean routing
reverse-proxies:
  - prefix: /api/v1
    target: http://backend:8080
    strip-prefix: true  # /api/v1/users -> /users

# 🚫 Avoid: No strip-prefix when needed
reverse-proxies:
  - prefix: /api/v1
    target: http://backend:8080
    # Backend receives /api/v1/users instead of /users
```

---

### 8. Validate Directory Paths for Static Handlers
```yaml
# ✅ Good: Relative paths from project root
mount-spa:
  - prefix: /admin
    dir: ./dist/admin    # Clear relative path
  
mount-static:
  - prefix: /assets
    dir: ./public/assets # Clear relative path

# ⚠️ Caution: Absolute paths (environment-specific)
mount-spa:
  - prefix: /admin
    dir: /var/www/admin  # Only works in production

# 🚫 Avoid: Non-existent paths
mount-static:
  - prefix: /assets
    dir: ./assets  # Typo: should be ./public/assets
```

---

## See Also

- **[Config](./config)** - Configuration management
- **[Deploy](./deploy)** - Deployment topology
- **[Service Registration](../02-registry/service-registration)** - Service patterns
- **[Router Registration](../02-registry/router-registration)** - Router patterns

---

## Related Guides

- **[YAML Configuration](../../02-framework-guide/04-configuration/)** - Configuration basics
- **[Deployment Strategies](../../04-guides/deployment/)** - Deployment patterns
- **[Schema Validation](../../04-guides/validation/)** - Validation techniques
