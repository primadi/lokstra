# Configuration

Lokstra provides a powerful YAML-based configuration system that allows you to declaratively define your entire application structure, from server settings to routes, services, and middleware. This approach enables environment-specific deployments and reduces boilerplate code.

## Configuration Overview

Lokstra's configuration system supports:

- **Multiple files** - Split configuration across multiple YAML files
- **Environment variables** - Dynamic configuration with variable substitution
- **File merging** - Combine configurations from different sources
- **Schema validation** - IntelliSense and validation support
- **Group includes** - Reusable route and middleware definitions
- **Modular structure** - Organize complex applications

## Basic Configuration Structure

### Minimal Configuration

```yaml
# config.yaml
server:
  name: my-app

apps:
  - name: web-api
    address: ":8080"
    routes:
      - method: GET
        path: /health
        handler: health.check
```

### Complete Configuration

```yaml
# Complete configuration example
server:
  name: production-server
  global_setting:
    log_level: info
    debug: false
    timeout: 30

apps:
  - name: api-app
    address: ":8080"
    listener_type: default
    routing_engine_type: httprouter
    setting:
      cors_enabled: true
    middleware:
      - cors
      - recovery
      - name: auth
        config:
          secret: ${JWT_SECRET}
    routes:
      - method: GET
        path: /users
        handler: user.list
        middleware: [rate_limit]
      - method: POST
        path: /users
        handler: user.create
    groups:
      - prefix: /admin
        middleware: [admin_auth]
        routes:
          - method: GET
            path: /stats
            handler: admin.stats

services:
  - name: db.main
    type: lokstra.dbpool_pg
    config:
      host: ${DB_HOST:localhost}
      port: ${DB_PORT:5432}
      database: ${DB_NAME}
      username: ${DB_USER}
      password: ${DB_PASSWORD}
      max_connections: 20

modules:
  - name: user-module
    path: ./modules/user
    settings:
      enabled: true
```

## Configuration Loading

### Single File Loading

```go
// Load single configuration file
cfg, err := lokstra.LoadConfigFile("config.yaml")
if err != nil {
    panic(err)
}

server, err := lokstra.NewServerFromConfig(regCtx, cfg)
if err != nil {
    panic(err)
}
```

### Directory Loading

```go
// Load all YAML files from directory
cfg, err := lokstra.LoadConfigDir("configs/production")
if err != nil {
    panic(err)
}

server, err := lokstra.NewServerFromConfig(regCtx, cfg)
if err != nil {
    panic(err)
}
```

### Multiple Environment Setup

```go
func createServer(env string) *lokstra.Server {
    regCtx := lokstra.NewGlobalRegistrationContext()
    
    // Register components
    registerAllComponents(regCtx)
    
    // Load environment-specific config
    configDir := fmt.Sprintf("configs/%s", env)
    cfg, err := lokstra.LoadConfigDir(configDir)
    if err != nil {
        panic(fmt.Sprintf("Failed to load %s config: %v", env, err))
    }
    
    server, err := lokstra.NewServerFromConfig(regCtx, cfg)
    if err != nil {
        panic(err)
    }
    
    return server
}

// Usage
devServer := createServer("development")
prodServer := createServer("production")
```

## Server Configuration

### Basic Server Settings

```yaml
server:
  name: my-application
  global_setting:
    log_level: info
    debug: false
    max_request_size: "10MB"
    request_timeout: "30s"
    shutdown_timeout: "10s"
```

### Advanced Server Settings

```yaml
server:
  name: high-performance-api
  global_setting:
    # Performance settings
    max_concurrent_requests: 1000
    read_timeout: "5s"
    write_timeout: "10s"
    idle_timeout: "120s"
    
    # Security settings
    rate_limit_enabled: true
    rate_limit_requests_per_minute: 1000
    
    # Monitoring settings
    metrics_enabled: true
    health_check_enabled: true
    
    # Custom application settings
    feature_flags:
      new_user_flow: true
      advanced_analytics: false
```

## Application Configuration

### Basic App Definition

```yaml
apps:
  - name: web-api
    address: ":8080"
    listener_type: default
    routing_engine_type: httprouter
    routes:
      - method: GET
        path: /
        handler: home.index
```

### Multiple Apps

```yaml
apps:
  # Main API application
  - name: api-app
    address: ":8080"
    middleware: [cors, auth]
    routes:
      - method: GET
        path: /api/users
        handler: user.list

  # Admin dashboard
  - name: admin-app
    address: ":8081"
    middleware: [cors, admin_auth]
    routes:
      - method: GET
        path: /dashboard
        handler: admin.dashboard

  # Metrics endpoint
  - name: metrics-app
    address: ":9090"
    routes:
      - method: GET
        path: /metrics
        handler: metrics.prometheus
```

### App Listener Types

```yaml
apps:
  # Standard HTTP listener
  - name: api
    address: ":8080"
    listener_type: default

  # FastHTTP listener for high performance
  - name: fast-api
    address: ":8081"
    listener_type: fasthttp

  # HTTP/3 with QUIC
  - name: http3-api
    address: ":8082"
    listener_type: http3
    setting:
      cert_file: "./certs/server.crt"
      key_file: "./certs/server.key"

  # Secure HTTPS listener
  - name: secure-api
    address: ":8443"
    listener_type: secure
    setting:
      cert_file: "./certs/server.crt"
      key_file: "./certs/server.key"
```

## Route Configuration

### Basic Routes

```yaml
apps:
  - name: api
    address: ":8080"
    routes:
      # Simple routes
      - method: GET
        path: /health
        handler: health.check
        
      - method: GET
        path: /users
        handler: user.list
        
      - method: POST
        path: /users
        handler: user.create
        
      # Path parameters
      - method: GET
        path: /users/{id}
        handler: user.get
        
      - method: PUT
        path: /users/{id}
        handler: user.update
        
      # Wildcard routes
      - method: GET
        path: /files/*filepath
        handler: file.serve
```

### Routes with Middleware

```yaml
apps:
  - name: api
    address: ":8080"
    routes:
      # Route with single middleware
      - method: GET
        path: /profile
        handler: user.profile
        middleware: [auth]
        
      # Route with multiple middleware
      - method: POST
        path: /admin/users
        handler: admin.createUser
        middleware: [auth, admin_only, audit]
        
      # Route with configured middleware
      - method: GET
        path: /api/data
        handler: data.get
        middleware:
          - rate_limit
          - name: cache
            config:
              ttl: 300
              key_prefix: "api_data"
```

## Route Groups

### Basic Groups

```yaml
apps:
  - name: api
    address: ":8080"
    groups:
      # API v1 group
      - prefix: /api/v1
        middleware: [auth]
        routes:
          - method: GET
            path: /users
            handler: v1.user.list
          - method: POST
            path: /users
            handler: v1.user.create
            
      # Admin group
      - prefix: /admin
        middleware: [auth, admin_only]
        routes:
          - method: GET
            path: /stats
            handler: admin.stats
          - method: GET
            path: /logs
            handler: admin.logs
```

### Nested Groups

```yaml
apps:
  - name: api
    address: ":8080"
    groups:
      - prefix: /api
        middleware: [cors]
        groups:
          # API v1
          - prefix: /v1
            middleware: [auth]
            routes:
              - method: GET
                path: /users
                handler: v1.user.list
                
          # API v2
          - prefix: /v2
            middleware: [auth, v2_features]
            routes:
              - method: GET
                path: /users
                handler: v2.user.list
                
          # Admin endpoints
          - prefix: /admin
            middleware: [admin_only]
            routes:
              - method: GET
                path: /health
                handler: admin.health
```

### Group Includes

Create reusable group configurations:

```yaml
# api-routes.yaml
routes:
  - method: GET
    path: /users
    handler: user.list
  - method: POST
    path: /users
    handler: user.create
  - method: GET
    path: /users/{id}
    handler: user.get

mount_static:
  - prefix: /uploads
    folder: ["./uploads"]
```

```yaml
# main config
apps:
  - name: api
    address: ":8080"
    groups:
      - prefix: /api/v1
        middleware: [auth]
        load_from:
          - api-routes.yaml
          - admin-routes.yaml
```

## Service Configuration

### Database Services

```yaml
services:
  # PostgreSQL with DSN
  - name: db.main
    type: lokstra.dbpool_pg
    config: "postgres://user:pass@localhost/mydb"
    
  # PostgreSQL with parameters
  - name: db.analytics
    type: lokstra.dbpool_pg
    config:
      host: ${DB_HOST:localhost}
      port: ${DB_PORT:5432}
      database: analytics
      username: ${DB_USER}
      password: ${DB_PASSWORD}
      max_connections: 20
      min_connections: 5
      max_idle_time: "30m"
      max_lifetime: "1h"
      sslmode: require
```

### Cache Services

```yaml
services:
  # In-memory cache
  - name: cache.session
    type: lokstra.kvstore_mem
    
  # Redis cache
  - name: cache.user
    type: lokstra.kvstore_redis
    config: "redis://localhost:6379/0"
    
  # Redis with configuration
  - name: cache.main
    type: lokstra.redis
    config:
      host: ${REDIS_HOST:localhost}
      port: ${REDIS_PORT:6379}
      database: 0
      password: ${REDIS_PASSWORD}
      max_idle: 10
      max_active: 100
```

### Logger Services

```yaml
services:
  # Simple logger
  - name: logger
    type: lokstra.logger
    config: info
    
  # Configured logger
  - name: app.logger
    type: lokstra.logger
    config:
      level: ${LOG_LEVEL:info}
      format: json
      output: stdout
      
  # File logger
  - name: audit.logger
    type: lokstra.logger
    config:
      level: info
      format: json
      output: ./logs/audit.log
```

### Custom Services

```yaml
services:
  # Email service
  - name: email
    type: email.smtp
    config:
      host: ${SMTP_HOST}
      port: ${SMTP_PORT:587}
      username: ${SMTP_USER}
      password: ${SMTP_PASSWORD}
      
  # File storage service
  - name: storage
    type: storage.s3
    config:
      bucket: ${S3_BUCKET}
      region: ${AWS_REGION}
      access_key: ${AWS_ACCESS_KEY}
      secret_key: ${AWS_SECRET_KEY}
```

## Middleware Configuration

### Global Middleware

```yaml
apps:
  - name: api
    address: ":8080"
    middleware:
      # Simple middleware
      - cors
      - recovery
      
      # Configured middleware
      - name: rate_limit
        config:
          requests_per_minute: 1000
          burst_size: 100
          
      - name: auth
        config:
          jwt_secret: ${JWT_SECRET}
          token_header: Authorization
          token_prefix: "Bearer "
```

### Conditional Middleware

```yaml
apps:
  - name: api
    address: ":8080"
    middleware:
      # Always enabled
      - cors
      
      # Conditionally enabled
      - name: debug_logging
        enabled: ${DEBUG_MODE:false}
        config:
          include_headers: true
          include_body: false
          
      - name: rate_limit
        enabled: ${RATE_LIMIT_ENABLED:true}
        config:
          requests_per_minute: ${RATE_LIMIT_RPM:1000}
```

## Environment Variables

### Variable Substitution

Lokstra supports environment variable substitution with default values:

```yaml
# Syntax: ${VARIABLE_NAME:default_value}
server:
  name: ${APP_NAME:my-app}
  global_setting:
    log_level: ${LOG_LEVEL:info}
    debug: ${DEBUG_MODE:false}
    port: ${PORT:8080}

services:
  - name: db.main
    type: lokstra.dbpool_pg
    config:
      host: ${DB_HOST:localhost}
      port: ${DB_PORT:5432}
      database: ${DB_NAME}
      username: ${DB_USER}
      password: ${DB_PASSWORD}
      max_connections: ${DB_MAX_CONN:20}
```

### Environment-Specific Files

Organize configurations by environment:

```
configs/
├── development/
│   ├── server.yaml
│   ├── database.yaml
│   └── services.yaml
├── staging/
│   ├── server.yaml
│   ├── database.yaml
│   └── services.yaml
└── production/
    ├── server.yaml
    ├── database.yaml
    └── services.yaml
```

Development configuration:
```yaml
# configs/development/server.yaml
server:
  name: dev-server
  global_setting:
    log_level: debug
    debug: true

# configs/development/database.yaml
services:
  - name: db.main
    type: lokstra.dbpool_pg
    config:
      host: localhost
      database: myapp_dev
      max_connections: 5
```

Production configuration:
```yaml
# configs/production/server.yaml
server:
  name: prod-server
  global_setting:
    log_level: warn
    debug: false

# configs/production/database.yaml
services:
  - name: db.main
    type: lokstra.dbpool_pg
    config:
      host: ${DB_HOST}
      database: ${DB_NAME}
      max_connections: 50
      sslmode: require
```

## Module Configuration

### External Modules

```yaml
modules:
  - name: user-management
    path: ./modules/user
    settings:
      enabled: true
      features:
        - registration
        - authentication
        - profile_management
        
  - name: payment-processing
    path: ./modules/payment
    settings:
      enabled: ${PAYMENT_ENABLED:false}
      providers:
        - stripe
        - paypal
```

## File Organization Patterns

### Single File Structure

```yaml
# app.yaml - Everything in one file
server:
  name: simple-app

apps:
  - name: web
    address: ":8080"
    routes:
      - method: GET
        path: /
        handler: home.index

services:
  - name: logger
    type: lokstra.logger
    config: info
```

### Multi-File Structure

```
configs/
├── server.yaml      # Server configuration
├── apps.yaml        # Application definitions
├── services.yaml    # Service configurations
├── middleware.yaml  # Middleware definitions
└── modules.yaml     # Module configurations
```

```yaml
# server.yaml
server:
  name: production-api
  global_setting:
    log_level: info

# apps.yaml  
apps:
  - name: api
    address: ":8080"
    middleware:
      - cors
      - auth

# services.yaml
services:
  - name: db.main
    type: lokstra.dbpool_pg
    config:
      host: ${DB_HOST}
```

### Domain-Driven Structure

```
configs/
├── core/
│   ├── server.yaml
│   └── middleware.yaml
├── user/
│   ├── routes.yaml
│   └── services.yaml
├── payment/
│   ├── routes.yaml
│   └── services.yaml
└── admin/
    ├── routes.yaml
    └── services.yaml
```

## Schema Validation

### VS Code Integration

Add schema validation to your YAML files:

```yaml
# yaml-language-server: $schema=https://lokstra.dev/schema/lokstra.json

server:
  name: my-app  # IntelliSense will provide suggestions
```

### Available Schemas

```yaml
# Main configuration schema
# yaml-language-server: $schema=https://lokstra.dev/schema/lokstra.json

# Group include schema
# yaml-language-server: $schema=https://lokstra.dev/schema/lokstra-group.json
```

## Configuration Validation

### Runtime Validation

```go
// Validate configuration before using
func validateConfig(cfg *lokstra.LokstraConfig) error {
    if cfg.Server == nil {
        return errors.New("server configuration is required")
    }
    
    if len(cfg.Apps) == 0 {
        return errors.New("at least one app must be configured")
    }
    
    for _, app := range cfg.Apps {
        if app.Address == "" {
            return fmt.Errorf("address is required for app %s", app.Name)
        }
        
        if len(app.Routes) == 0 && len(app.Groups) == 0 {
            return fmt.Errorf("app %s must have routes or groups", app.Name)
        }
    }
    
    return nil
}

// Load and validate
cfg, err := lokstra.LoadConfigDir("configs/production")
if err != nil {
    panic(err)
}

if err := validateConfig(cfg); err != nil {
    panic(fmt.Sprintf("Invalid configuration: %v", err))
}
```

## Advanced Configuration Patterns

### Feature Flags

```yaml
server:
  global_setting:
    features:
      new_user_ui: ${FEATURE_NEW_USER_UI:false}
      advanced_search: ${FEATURE_ADVANCED_SEARCH:true}
      beta_features: ${FEATURE_BETA:false}

apps:
  - name: api
    routes:
      - method: GET
        path: /search
        handler: search.basic
        middleware:
          - name: feature_gate
            config:
              feature: advanced_search
              fallback_handler: search.simple
```

### Dynamic Service Configuration

```yaml
services:
  # Production database
  - name: db.main
    type: lokstra.dbpool_pg
    config:
      host: ${DB_HOST}
      port: ${DB_PORT:5432}
      database: ${DB_NAME}
      
  # Read replica (if configured)
  - name: db.readonly
    type: lokstra.dbpool_pg
    enabled: ${DB_READONLY_ENABLED:false}
    config:
      host: ${DB_READONLY_HOST}
      port: ${DB_READONLY_PORT:5432}
      database: ${DB_NAME}
```

### Configuration Inheritance

Base configuration:
```yaml
# configs/base/server.yaml
server: &default-server
  global_setting:
    request_timeout: "30s"
    shutdown_timeout: "10s"

apps: &default-middleware
  middleware:
    - cors
    - recovery
```

Environment-specific:
```yaml
# configs/production/server.yaml
server:
  <<: *default-server
  name: production-api
  global_setting:
    <<: *default-server.global_setting
    log_level: warn
```

## Troubleshooting

### Common Configuration Issues

1. **YAML Syntax Errors**
   ```
   unmarshal yaml config.yaml: yaml: line 10: found character that cannot start any token
   ```
   - Check indentation (use spaces, not tabs)
   - Verify YAML syntax with a validator
   - Ensure proper quoting for special characters

2. **Environment Variable Not Found**
   ```
   expand group includes for app api: variable DB_HOST not found
   ```
   - Set missing environment variables
   - Provide default values: `${DB_HOST:localhost}`

3. **Handler Not Found**
   ```
   handler 'user.list' not found
   ```
   - Ensure handlers are registered before loading config
   - Check handler name spelling and case

4. **Service Type Not Found**
   ```
   service factory 'my.service' not found
   ```
   - Register service factory before loading config
   - Verify service type name matches registration

### Debugging Configuration

```go
func debugConfig(cfg *lokstra.LokstraConfig) {
    fmt.Printf("Server: %+v\n", cfg.Server)
    
    for i, app := range cfg.Apps {
        fmt.Printf("App %d: %s at %s\n", i, app.Name, app.Address)
        fmt.Printf("  Routes: %d\n", len(app.Routes))
        fmt.Printf("  Groups: %d\n", len(app.Groups))
        fmt.Printf("  Middleware: %d\n", len(app.Middleware))
    }
    
    for i, svc := range cfg.Services {
        fmt.Printf("Service %d: %s (type: %s)\n", i, svc.Name, svc.Type)
    }
}
```

## Best Practices

### 1. Environment Separation

Keep environment-specific configurations separate:

```
configs/
├── base/           # Common configuration
├── development/    # Dev overrides
├── staging/        # Staging overrides
└── production/     # Production overrides
```

### 2. Security

- Never commit secrets to version control
- Use environment variables for sensitive data
- Implement configuration validation
- Use secure defaults

```yaml
# Good: Use environment variables
database:
  password: ${DB_PASSWORD}

# Bad: Hardcoded secrets
database:
  password: supersecret123
```

### 3. Documentation

Document your configuration structure:

```yaml
# Server configuration
server:
  name: my-api                    # Application name for logging/metrics
  global_setting:
    log_level: info               # debug, info, warn, error
    request_timeout: "30s"        # Max request processing time
    max_request_size: "10MB"      # Maximum request body size
```

### 4. Validation

Implement configuration validation:

```go
type Config struct {
    RequiredField string `yaml:"required_field" validate:"required"`
    OptionalField string `yaml:"optional_field"`
}

func LoadAndValidate(path string) (*Config, error) {
    cfg, err := LoadConfig(path)
    if err != nil {
        return nil, err
    }
    
    if err := validator.Validate(cfg); err != nil {
        return nil, fmt.Errorf("invalid configuration: %w", err)
    }
    
    return cfg, nil
}
```

## Next Steps

- [Getting Started](./getting-started.md) - Build your first configured app
- [Services](./services.md) - Configure services and dependencies  
- [Middleware](./middleware.md) - Configure middleware pipelines
- [HTMX Integration](./htmx-integration.md) - Configure HTMX applications

---

*Configuration-driven development with Lokstra enables maintainable, environment-aware applications. Master these patterns to build robust, scalable systems.*