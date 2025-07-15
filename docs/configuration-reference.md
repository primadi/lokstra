# Configuration Reference

This document provides a complete reference for Lokstra's YAML configuration system.

## Configuration Structure

Lokstra configuration is split into logical files that can be loaded from a directory:

```
config/
├── server.yaml          # Server-level settings
├── apps_api.yaml        # API application configuration
├── apps_web.yaml        # Web application configuration
├── services_db.yaml     # Database services
├── services_cache.yaml  # Cache services
└── modules_auth.yaml    # Authentication modules
```

## Server Configuration

### server.yaml

```yaml
server:
  name: my-server                    # Server name
  global_setting:                    # Global settings
    log_level: info                  # Log level: debug, info, warn, error
    environment: production          # Environment identifier
    shutdown_timeout: 30s            # Graceful shutdown timeout
```

## Application Configuration

### apps_*.yaml

```yaml
apps:
  - name: api-app                    # Application name
    address: :8080                   # Listen address
    listener_type: http              # Listener type: http, https, fasthttp, http3, unix
    router_engine_type: httprouter   # Router engine
    
    # Application settings
    setting:
      cors: true
      max_body_size: 10MB
    
    # Global middleware for this app
    middleware:
      - name: lokstra.recovery       # Middleware name
        enabled: true                # Enable/disable
        config:                      # Middleware configuration
          log_panics: true
      - name: lokstra.cors
        enabled: true
        config:
          allowed_origins: ["*"]
          allowed_methods: ["GET", "POST", "PUT", "DELETE"]
    
    # Routes
    routes:
      - method: GET                  # HTTP method
        path: /health                # Route path
        handler: health.check        # Handler name
        middleware:                  # Route-specific middleware
          - name: lokstra.requestid
            enabled: true
    
    # Route groups
    groups:
      - prefix: /api/v1             # Group prefix
        middleware:                 # Group middleware
          - name: lokstra.jwt_auth
            enabled: true
            config:
              service_name: jwt_auth
        override_middleware: false  # Override parent middleware
        
        routes:
          - method: GET
            path: /users
            handler: user.list
          - method: POST
            path: /users
            handler: user.create
        
        groups:                     # Nested groups
          - prefix: /admin
            middleware:
              - name: admin.auth
                enabled: true
            routes:
              - method: GET
                path: /stats
                handler: admin.stats
    
    # Static file serving
    mount_static:
      - prefix: /static             # URL prefix
        folder: ./public            # Local folder
    
    # SPA serving
    mount_spa:
      - prefix: /app                # URL prefix
        fallback_file: ./dist/index.html  # Fallback file
    
    # Reverse proxy
    mount_reverse_proxy:
      - prefix: /api                # URL prefix
        target: http://backend:8080 # Target URL
```

## Service Configuration

### services_*.yaml

```yaml
services:
  # Logger Service
  - type: lokstra.logger            # Service type
    name: default-logger            # Service instance name
    config:
      level: ${LOG_LEVEL:info}      # Environment variable with default
      format: json                  # Log format: json, text
      output: stdout                # Output: stdout, stderr, file
      file_path: /var/log/app.log   # File path (if output=file)
  
  # Database Service
  - type: lokstra.dbpool_pg
    name: main-db
    config:
      dsn: ${DATABASE_URL:postgres://user:pass@localhost/db}
      max_connections: 20
      min_connections: 5
      connection_timeout: 30s
  
  # Redis Service
  - type: lokstra.redis
    name: cache
    config:
      addr: ${REDIS_ADDR:localhost:6379}
      password: ${REDIS_PASSWORD:}
      db: ${REDIS_DB:0}
      pool_size: 10
      min_idle_conns: 5
  
  # Email Service
  - type: lokstra.email
    name: mailer
    config:
      smtp_host: ${SMTP_HOST:localhost}
      smtp_port: ${SMTP_PORT:587}
      username: ${SMTP_USER:}
      password: ${SMTP_PASS:}
      from: ${SMTP_FROM:noreply@example.com}
  
  # Metrics Service
  - type: lokstra.metrics
    name: prometheus
    config:
      namespace: myapp
      subsystem: api
  
  # Health Check Service
  - type: lokstra.health
    name: health
    config:
      checks:
        - name: database
          timeout: 5s
        - name: redis
          timeout: 3s
```

## Module Configuration

### modules_*.yaml

```yaml
modules:
  # JWT Authentication Module
  - name: jwt-auth
    path: lokstra/modules/jwt_auth_basic
    entry: GetModule
    settings:
      secret: ${JWT_SECRET:default-secret}
      expires_hours: 24
      algorithm: HS256
    permissions:
      required_role: user
  
  # Custom Business Module
  - name: user-management
    path: myapp/modules/user
    entry: GetUserModule
    settings:
      default_role: user
      password_min_length: 8
```

## Environment Variables

### Variable Syntax

Use `${VARIABLE_NAME:default_value}` syntax in YAML:

```yaml
config:
  database_url: ${DATABASE_URL:postgres://localhost/myapp}
  redis_addr: ${REDIS_ADDR:localhost:6379}
  log_level: ${LOG_LEVEL:info}
  debug: ${DEBUG:false}
  port: ${PORT:8080}
```

### Common Environment Variables

```bash
# Server
LOG_LEVEL=info
ENVIRONMENT=production
SHUTDOWN_TIMEOUT=30s

# Database
DATABASE_URL=postgres://user:pass@localhost/db
DB_MAX_CONNECTIONS=20
DB_CONNECTION_TIMEOUT=30s

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=secret
REDIS_DB=0

# Email
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=user@gmail.com
SMTP_PASS=password
SMTP_FROM=noreply@myapp.com

# JWT
JWT_SECRET=my-super-secret-key
JWT_EXPIRES_HOURS=24

# Application
PORT=8080
CORS_ORIGINS=http://localhost:3000,https://myapp.com
```

## Middleware Configuration

### Built-in Middleware

```yaml
middleware:
  # Recovery Middleware
  - name: lokstra.recovery
    enabled: true
    config:
      log_panics: true
      stack_trace: true
  
  # CORS Middleware
  - name: lokstra.cors
    enabled: true
    config:
      allowed_origins: ["*"]
      allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
      allowed_headers: ["*"]
      exposed_headers: ["X-Request-ID"]
      allow_credentials: true
      max_age: 86400
  
  # Request Logger
  - name: lokstra.request_logger
    enabled: true
    config:
      log_body: false
      log_headers: false
  
  # Rate Limiting
  - name: lokstra.ratelimit
    enabled: true
    config:
      limit: 100                    # Requests per window
      window_minutes: 1             # Time window
      key: ip                       # Rate limit key: ip, user
  
  # Request ID
  - name: lokstra.requestid
    enabled: true
    config:
      header_name: X-Request-ID
  
  # Timeout
  - name: lokstra.timeout
    enabled: true
    config:
      timeout_seconds: 30
  
  # Security Headers
  - name: lokstra.security
    enabled: true
    config:
      enable_hsts: true
      enable_csp: true
      csp_policy: "default-src 'self'"
      enable_x_frame_options: true
  
  # Body Size Limit
  - name: lokstra.bodysizelimit
    enabled: true
    config:
      max_size_mb: 10
  
  # JWT Authentication
  - name: lokstra.jwt_auth
    enabled: true
    config:
      service_name: jwt_auth
      skip_paths: ["/login", "/health"]
```

## Configuration Loading

### Directory Structure

```
config/
├── server.yaml
├── apps_api.yaml
├── apps_web.yaml
├── services_db.yaml
├── services_cache.yaml
└── modules_auth.yaml
```

### Loading in Code

```go
// Load from directory
cfg, err := lokstra.LoadConfigDir("config")
if err != nil {
    panic(err)
}

// Load from specific files
cfg, err := lokstra.LoadConfigFiles(
    "config/server.yaml",
    "config/apps.yaml",
    "config/services.yaml",
)
if err != nil {
    panic(err)
}

// Create server from config
server, err := lokstra.NewServerFromConfig(ctx, cfg)
if err != nil {
    panic(err)
}
```

### Configuration Validation

Lokstra validates configuration at startup:

- Required fields are present
- Service types are registered
- Middleware modules are available
- Port conflicts are detected
- Environment variables are resolved

## Best Practices

### 1. File Organization

- Split configuration by concern (apps, services, modules)
- Use descriptive file names
- Keep related configuration together

### 2. Environment Variables

- Use environment variables for secrets
- Provide sensible defaults
- Document required variables

### 3. Security

- Never commit secrets to version control
- Use environment variables for sensitive data
- Validate configuration at startup

### 4. Deployment

- Use different config directories for environments
- Override with environment-specific files
- Validate configuration in CI/CD

### 5. Documentation

- Document all configuration options
- Provide example configurations
- Keep documentation up to date

## Examples

See the [examples directory](../cmd/examples/) for complete configuration examples:

- [Basic YAML Config](../cmd/examples/01_basic_overview/03_with_yaml_config/)
- [Split Config Files](../cmd/examples/03_best_practices/03_split_config_files/)
- [Service Examples](../cmd/examples/07_default_services/)
- [Middleware Examples](../cmd/examples/08_default_middleware/)
