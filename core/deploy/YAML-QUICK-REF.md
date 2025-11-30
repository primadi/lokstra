# YAML Configuration Quick Reference

## File Structure

```yaml
configs:                      # Global configs (uppercase keys)
  KEY_NAME: value

services:                     # Service definitions (lowercase-with-dashes)
  service-name:
    type: factory-type        # REQUIRED
    depends-on:               # Optional
      - dependency
      - paramName:service
    config:                   # Optional
      key: value

routers:                      # Router definitions
  router-name:
    service: service-name     # REQUIRED
    overrides:                # Optional
      MethodName:
        hide: bool
        middleware: [...]

deployments:                  # Deployment targets
  deployment-name:
    config-overrides:         # Optional
      KEY_NAME: new-value
    servers:
      server-name:
        base-url: https://... # REQUIRED
        apps:
          - port: 8080        # REQUIRED (1-65535)
            services: [...]   # Optional
            routers: [...]    # Optional
            remote-services: [...] # Optional
```

## Loading Patterns

```go
// Single file
config, err := loader.LoadConfig("config.yaml")

// Multiple files (merge)
config, err := loader.LoadConfig(
    "base.yaml",
    "services.yaml",
    "prod.yaml",
)

// Directory (all .yaml and .yml)
config, err := loader.LoadConfigFromDir("config")

// Load and build deployment
dep, err := loader.LoadAndBuild(
    []string{"config.yaml"},
    "production",
    registry,
)

// Load directory and build
dep, err := loader.LoadAndBuildFromDir(
    "config",
    "production",
    registry,
)
```

## Naming Rules

| Type | Pattern | Examples |
|------|---------|----------|
| Configs | `[A-Z][A-Z0-9_]*` | `DB_HOST`, `API_KEY`, `MAX_RETRY` |
| Services | `[a-z][a-z0-9-]*` | `db-pool`, `user-service`, `api` |
| Dependencies | `[a-zA-Z][a-zA-Z0-9]*:[a-z][a-z0-9-]*` | `db:db-pool`, `userSvc:user-service` |
| URLs | `https?://` | `http://localhost`, `https://api.com` |
| Ports | `1-65535` | `3000`, `8080`, `443` |

## Config References

```yaml
configs:
  DB_HOST: localhost
  DB_PORT: 5432
  
  # Reference other configs
  DB_DSN: "postgres://${@cfg:DB_HOST}:${@cfg:DB_PORT}/mydb"

services:
  db:
    config:
      # Use config values
      host: ${@cfg:DB_HOST}
      port: ${@cfg:DB_PORT}
```

## Service Dependencies

```yaml
services:
  # No dependencies
  logger:
    type: logger-service
    
  # Simple dependency (param name = service name)
  api:
    type: api-service
    depends-on:
      - logger
      
  # Aliased dependencies
  order-service:
    type: order-service
    depends-on:
      - db:db-pool              # maps to 'db' parameter
      - userSvc:user-service    # maps to 'userSvc' parameter
      - logger                  # maps to 'logger' parameter
```

## Multi-File Merging

```yaml
# base.yaml
configs:
  LOG_LEVEL: info
services:
  db: {...}
  cache: {...}

# services.yaml
services:
  api: {...}
  worker: {...}

# prod.yaml
configs:
  LOG_LEVEL: warn    # Overrides base
deployments:
  production: {...}

# Result: All 4 services, LOG_LEVEL=warn
```

## Deployment Patterns

### Pattern 1: Environment-Specific Files
```
config/
  ├── common.yaml
  ├── development.yaml
  ├── staging.yaml
  └── production.yaml
```

```go
// Load environment-specific
env := os.Getenv("ENV")
dep, _ := loader.LoadAndBuild(
    []string{"config/common.yaml", "config/" + env + ".yaml"},
    env,
    registry,
)
```

### Pattern 2: Feature-Based Files
```
config/
  ├── infrastructure.yaml  # DB, cache, etc.
  ├── auth.yaml           # Auth services
  ├── api.yaml            # API services
  └── deployments.yaml    # All deployments
```

```go
// Load all and select deployment
dep, _ := loader.LoadAndBuildFromDir("config", "production", registry)
```

### Pattern 3: Single File (Simple)
```yaml
# all.yaml - Everything in one file
configs: {...}
services: {...}
deployments: {...}
```

```go
dep, _ := loader.LoadAndBuild([]string{"all.yaml"}, "dev", registry)
```

## Validation

Automatic validation happens on load:

```go
config, err := loader.LoadConfig("config.yaml")
if err != nil {
    // Validation errors are returned here
    fmt.Println(err)
    // Output:
    // schema validation failed:
    //   - services.InvalidName: Does not match pattern
    //   - deployments.prod.servers.api.apps.0.port: Must be >= 1
}
```

## Common Patterns

### Database Service
```yaml
services:
  db-pool:
    type: postgres-pool
    config:
      host: ${@cfg:DB_HOST}
      port: ${@cfg:DB_PORT}
      database: ${@cfg:DB_NAME}
      max-conns: 20
      ssl-mode: disable
```

### Service with Multiple Dependencies
```yaml
services:
  order-service:
    type: order-service-factory
    depends-on:
      - dbOrder:db-pool
      - dbUser:user-db-pool
      - cache:redis-cache
      - logger:main-logger
      - userSvc:user-service
    config:
      timeout: 30
      max-retries: 3
```

### Multi-Environment Deployment
```yaml
deployments:
  development:
    config-overrides:
      LOG_LEVEL: debug
      DB_HOST: localhost
    servers:
      dev: {...}
      
  staging:
    config-overrides:
      LOG_LEVEL: info
      DB_HOST: staging-db.company.com
    servers:
      staging: {...}
      
  production:
    config-overrides:
      LOG_LEVEL: warn
      DB_HOST: prod-db.company.com
    servers:
      api-01: {...}
      api-02: {...}
```

## IDE Setup

### VS Code
`.vscode/settings.json`:
```json
{
  "yaml.schemas": {
    "./core/deploy/schema/lokstra.schema.json": "config/**/*.yaml"
  }
}
```

### GoLand / IntelliJ
Settings → Languages & Frameworks → JSON Schema Mappings:
- Add `lokstra.schema.json`
- Map to `**/config/*.yaml`

## Complete Example

```yaml
# Production deployment configuration

configs:
  # Database
  DB_HOST: prod-db.company.com
  DB_PORT: 5432
  DB_NAME: myapp_production
  DB_MAX_CONNS: 100
  
  # Cache
  REDIS_HOST: prod-redis.company.com
  REDIS_PORT: 6379
  
  # Application
  LOG_LEVEL: warn
  API_TIMEOUT: 30

services:
  db-pool:
    type: postgres-pool
    config:
      host: ${@cfg:DB_HOST}
      port: ${@cfg:DB_PORT}
      database: ${@cfg:DB_NAME}
      max-conns: ${@cfg:DB_MAX_CONNS}

  redis-cache:
    type: redis-cache
    config:
      host: ${@cfg:REDIS_HOST}
      port: ${@cfg:REDIS_PORT}

  logger:
    type: logger-service
    config:
      level: ${@cfg:LOG_LEVEL}
      format: json

  user-service:
    type: user-service-factory
    depends-on:
      - db:db-pool
      - cache:redis-cache
      - logger
    config:
      cache-ttl: 3600

  order-service:
    type: order-service-factory
    depends-on:
      - dbOrder:db-pool
      - userSvc:user-service
      - logger

remote-services:
  payment-gateway:
    url: https://payments.company.com
    resource: payment

deployments:
  production:
    servers:
      api-server-01:
        base-url: https://api1.company.com
        apps:
          - port: 8080
            services:
              - db-pool
              - redis-cache
              - logger
              - user-service
              - order-service
            remote-services:
              - payment-gateway

      api-server-02:
        base-url: https://api2.company.com
        apps:
          - port: 8080
            services:
              - db-pool
              - redis-cache
              - logger
              - user-service
              - order-service
            remote-services:
              - payment-gateway
```

## Tips & Best Practices

### ✅ DO:
- Use uppercase for config keys (`DB_HOST`)
- Use lowercase-with-dashes for services (`user-service`)
- Split large configs into multiple files
- Use config references (`${@cfg:KEY}`)
- Validate locally before deployment
- Keep related services in same file
- Document complex configurations with comments

### ❌ DON'T:
- Mix naming conventions
- Put secrets in YAML (use env vars)
- Create circular dependencies
- Override same key multiple times unintentionally
- Forget to register factories before loading YAML

## Error Handling

```go
config, err := loader.LoadConfig("config.yaml")
if err != nil {
    switch {
    case strings.Contains(err.Error(), "validation failed"):
        // Schema validation error
        log.Fatal("Invalid config structure:", err)
        
    case strings.Contains(err.Error(), "failed to read"):
        // File not found or permission error
        log.Fatal("Cannot read config file:", err)
        
    case strings.Contains(err.Error(), "failed to parse"):
        // YAML syntax error
        log.Fatal("Invalid YAML syntax:", err)
        
    default:
        log.Fatal("Unexpected error:", err)
    }
}
```

## See Also

- Full documentation: `PHASE3-YAML-CONFIG.md`
- JSON Schema: `schema/lokstra.schema.json`
- Working example: `examples/yaml/`
- Test suite: `loader/loader_test.go`
