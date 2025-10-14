# YAML Configuration in Lokstra

Lokstra provides a powerful YAML-based configuration system that allows you to define your entire application architecture declaratively. This approach promotes Infrastructure as Code (IaC) principles and makes it easy to manage complex microservice architectures.

## Overview

The YAML configuration system in Lokstra supports four main top-level sections:

- **`routers`** - Define HTTP routing and middleware configuration
- **`services`** - Configure dependency injection services (databases, caches, etc.)
- **`middlewares`** - Define reusable middleware components
- **`servers`** - Configure server instances and their applications

## Configuration Structure

### Top-Level Structure

```yaml
routers:
  - name: api-router
    # router configuration

services:
  - name: database
    # service configuration

middlewares:
  - name: cors-middleware
    # middleware configuration

servers:
  - name: web-server
    # server configuration
```

## Multi-File Configuration

Lokstra supports loading configuration from multiple YAML files, which are merged together. This allows you to:

- Split configuration by environment (dev.yaml, prod.yaml)
- Organize by component (routers.yaml, services.yaml, middlewares.yaml)
- Share common configurations across teams

### Loading Configuration

```go
// Load single file
var config config.Config
err := config.LoadConfigFile("app.yaml", &config)

// Load from directory (merges all .yaml/.yml files)
var config config.Config  
err := config.LoadConfigDir("./config", &config)
```

## Configuration Sections

### 1. Routers Configuration

Routers define HTTP routing rules and associate middleware chains.

```yaml
routers:
  - name: blog-api-router
    engine-type: default          # optional, default: "default"
    enable: true                  # optional, default: true
    use: [cors-mw, auth-mw]      # router-level middleware
    override-parent-mw: false     # optional, default: false
    routes:
      - name: get-posts
        path: /posts
        method: GET               # optional, default: "GET"
        enable: true             # optional, default: true
        override-parent-mw: false # optional, default: false
        use: [cache-mw]          # route-specific middleware
        handler: GetPostsHandler  # handler function name
      
      - name: create-post
        path: /posts
        method: POST
        use: [validate-mw]
        handler: CreatePostHandler
```

**Router Properties:**

- `name` - Unique router identifier (required)
- `engine-type` - Router engine type (default: "default")
- `enable` - Enable/disable router (default: true)
- `use` - Array of middleware names to apply to all routes
- `override-parent-mw` - Whether to override parent middleware chain
- `routes` - Array of route definitions

**Route Properties:**

- `name` - Unique route identifier within router (required)
- `path` - HTTP path pattern (required)
- `method` - HTTP method (default: "GET")
- `enable` - Enable/disable route (default: true)
- `override-parent-mw` - Override router middleware chain
- `use` - Array of route-specific middleware names
- `handler` - Handler function name (must be registered in registry)

### 2. Services Configuration

Services are dependency injection components like databases, caches, or external APIs.

```yaml
services:
  - name: main-db
    type: postgres              # service factory type
    enable: true               # optional, default: true
    config:
      dsn: "postgres://user:pass@localhost/mydb?sslmode=disable"
      max_connections: 25
      max_idle: 5
  
  - name: redis-cache
    type: redis
    config:
      addr: "localhost:6379"
      password: ""
      db: 0
      
  - name: email-service
    type: smtp
    config:
      host: "smtp.gmail.com"
      port: 587
      username: "user@gmail.com"
      password: "app-password"
```

**Service Properties:**

- `name` - Unique service identifier for dependency injection (required)
- `type` - Service factory type (must be registered) (required)
- `enable` - Enable/disable service (default: true)
- `config` - Service-specific configuration (varies by type)

### 3. Middlewares Configuration

Middlewares are reusable HTTP request/response processors.

```yaml
middlewares:
  - name: cors-api
    type: cors                 # middleware factory type
    enable: true              # optional, default: true
    config:
      allowed_origins: ["http://localhost:3000", "https://myapp.com"]
      allowed_methods: ["GET", "POST", "PUT", "DELETE"]
      allowed_headers: ["Content-Type", "Authorization"]
      expose_headers: ["X-Total-Count"]
      allow_credentials: true
      max_age: 3600
  
  - name: auth-jwt
    type: jwt
    config:
      secret: "my-jwt-secret-key"
      algorithm: "HS256"
      token_lookup: "header:Authorization"
      auth_scheme: "Bearer"
      
  - name: rate-limiter
    type: rate_limit
    config:
      requests_per_minute: 100
      burst: 10
      key_generator: "ip"
```

**Middleware Properties:**

- `name` - Unique middleware identifier (required)
- `type` - Middleware factory type (must be registered) (required)
- `enable` - Enable/disable middleware (default: true)
- `config` - Middleware-specific configuration (varies by type)

### 4. Servers Configuration

Servers define application instances and their deployment configuration.

```yaml
servers:
  - name: monolith-server
    description: "Single server hosting all applications"
    services: [main-db, redis-cache, email-service]  # services to inject
    apps:
      - name: web-app
        addr: ":8080"
        listener-type: default    # optional, default: "default"
        routers: [blog-api-router, admin-router]
        reverse-proxies:          # optional
          - path: /api/external
            strip-prefix: ""      # optional, default: ""
            target: "http://external-service:8080"
  
  - name: microservice-blog
    description: "Blog microservice"
    services: [main-db]
    apps:
      - name: blog-service
        addr: ":8081"
        routers: [blog-api-router]
        
  - name: microservice-auth  
    description: "Authentication microservice"
    services: [main-db, redis-cache]
    apps:
      - name: auth-service
        addr: ":8082"
        routers: [auth-router]
```

**Server Properties:**

- `name` - Unique server identifier (required)
- `description` - Server description (optional)
- `services` - Array of service names to make available to apps
- `apps` - Array of application instances

**App Properties:**

- `name` - Unique app identifier within server (required)
- `addr` - Listen address (e.g., ":8080") (required)
- `listener-type` - Listener type (default: "default")
- `routers` - Array of router names to mount
- `reverse-proxies` - Array of reverse proxy configurations (optional)

**Reverse Proxy Properties:**

- `path` - Path prefix to proxy (required)
- `strip-prefix` - Prefix to strip before forwarding (default: "")
- `target` - Target URL to proxy to (required)

## Configuration Application

### Apply Specific Components

```go
var cfg config.Config
config.LoadConfigFile("app.yaml", &cfg)

// Apply specific routers
err := config.ApplyRoutersConfig(&cfg, "blog-api-router", "admin-router")

// Apply specific services
err := config.ApplyServicesConfig(&cfg, "main-db", "redis-cache")

// Apply specific middlewares  
err := config.ApplyMiddlewareConfig(&cfg, "cors-api", "auth-jwt")

// Apply specific server
server, err := config.ApplyServerConfig(&cfg, "monolith-server")
```

### Apply Complete Configuration

```go
var cfg config.Config
config.LoadConfigDir("./config", &cfg)

// Apply all components needed for a server
server, err := config.ApplyAllConfig(&cfg, "monolith-server")
if err != nil {
    log.Fatal(err)
}

// Start the server
server.Start()
```

## Best Practices

### 1. Configuration Organization

**Environment-based:**
```
config/
├── base.yaml          # Common configuration
├── development.yaml   # Development overrides
├── staging.yaml       # Staging overrides
└── production.yaml    # Production overrides
```

**Component-based:**
```
config/
├── routers.yaml       # All router definitions
├── services.yaml      # All service definitions
├── middlewares.yaml   # All middleware definitions
└── servers.yaml       # All server definitions
```

### 2. Configuration Validation

Always validate configuration after loading:

```go
var cfg config.Config
if err := config.LoadConfigDir("./config", &cfg); err != nil {
    log.Fatal("Failed to load config:", err)
}

if err := cfg.Validate(); err != nil {
    log.Fatal("Config validation failed:", err)
}
```

### 3. Environment Variables

Use environment variables for sensitive data:

```yaml
services:
  - name: main-db
    type: postgres
    config:
      dsn: "${DATABASE_URL}"  # Use environment variable
      max_connections: 25

middlewares:
  - name: auth-jwt
    type: jwt
    config:
      secret: "${JWT_SECRET}"  # Use environment variable
```

### 4. Default Values

Leverage default values to keep configuration minimal:

```yaml
# Minimal configuration using defaults
routers:
  - name: api-router      # engine-type: default, enable: true
    routes:
      - name: health      # method: GET, enable: true
        path: /health
        handler: HealthHandler
```

### 5. Middleware Chains

Design middleware chains thoughtfully:

```yaml
routers:
  - name: api-router
    use: [cors-api, request-logging]  # Applied to all routes
    routes:
      - name: public-endpoint
        path: /public
        handler: PublicHandler
        # Inherits: cors-api, request-logging
        
      - name: protected-endpoint  
        path: /protected
        use: [auth-jwt, rate-limiter]  # Additional middleware
        handler: ProtectedHandler
        # Final chain: cors-api, request-logging, auth-jwt, rate-limiter
```

## Error Handling

The configuration system provides detailed error messages:

```go
var cfg config.Config
if err := config.LoadConfigFile("app.yaml", &cfg); err != nil {
    // Possible errors:
    // - File not found
    // - YAML parsing errors
    // - Validation errors (duplicate names, missing references)
    log.Printf("Configuration error: %v", err)
}
```

## Integration with Registry

The configuration system integrates with Lokstra's registry system:

```go
// Register factories before applying config
lokstra_registry.RegisterServiceFactory("postgres", NewPostgresService)
lokstra_registry.RegisterMiddlewareFactory("cors", NewCorsMiddleware)
lokstra_registry.RegisterHandler("GetPostsHandler", GetPostsHandler)

// Apply configuration
var cfg config.Config
config.LoadConfigDir("./config", &cfg)
server, err := config.ApplyAllConfig(&cfg, "web-server")
```

This declarative approach to configuration makes Lokstra applications highly maintainable and allows for easy deployment across different environments while keeping the codebase clean and focused on business logic.