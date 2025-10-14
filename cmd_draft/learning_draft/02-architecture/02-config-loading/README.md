# 02-Config Loading

Learn how to move service configuration from code to YAML files with environment variable support.

## Quick Start

```bash
# Run with default config
go run .

# Run with environment variables (PowerShell)
$env:APP_ENV="production"
$env:LOG_LEVEL="debug"
$env:COUNTER_SEED="5000"
go run .

# Run with environment variables (Bash)
APP_ENV=production LOG_LEVEL=debug COUNTER_SEED=5000 go run .
```

## What Changed from 01-Registry-Basics?

### Before (Code-based):
```go
// Hardcoded in main.go
lokstra_registry.RegisterLazyService("email-service", "email", map[string]any{
    "smtp_host": "smtp.gmail.com",
    "smtp_port": 587,
    "from":      "demo@lokstra.dev",
})
```

### After (Config-based):
```yaml
# config.yaml
services:
  - name: email-service
    type: email
    config:
      smtp_host: ${SMTP_HOST:smtp.gmail.com}
      smtp_port: ${SMTP_PORT:587}
      from: ${EMAIL_FROM:demo@lokstra.dev}
```

**Benefits:**
- ✅ Change config without recompiling
- ✅ Different configs per environment (dev, staging, prod)
- ✅ Secrets in environment variables (not in code)
- ✅ Version control for configs

## Configuration Syntax

### Environment Variables

```yaml
# Default ENV resolver
${VAR_NAME}                      # No default
${VAR_NAME:default_value}        # With default

# Custom resolvers (AWS, Vault, etc.)
${@RESOLVER:KEY}                 # No default
${@RESOLVER:KEY:default_value}   # With default
```

**Syntax:**
- `VAR_NAME` - Environment variable to read (ENV resolver is default)
- `default_value` - Used if variable is not set (can contain `:`)
- `@RESOLVER` - Custom resolver (ENV, AWS, VAULT, etc.) - optional

**Examples:**
```yaml
# ENV resolver (default - no @ prefix needed)
smtp_host: ${SMTP_HOST:localhost}
smtp_port: ${SMTP_PORT:587}
base_url: ${BASE_URL:http://localhost:8080}

# Default can contain colons
db_dsn: ${DATABASE_URL:postgresql://user:pass@localhost:5432/db}

# Custom resolvers (with @ prefix)
api_key: ${@AWS:api-key-secret}
db_password: ${@VAULT:database/password:fallback}
service_url: ${@CONSUL:service-url:http://default}

# Different values per environment
log_level: ${LOG_LEVEL:info}  # dev: info, prod: warn
```

### Service Definition

```yaml
services:
  - name: service-name           # Unique service name
    type: service-type           # Must match registered factory
    config:                      # Passed to factory
      key1: value1
      key2: ${ENV_VAR:default}
```

**How it works:**
1. Framework reads `config.yaml`
2. Replaces `${ENV_VAR:default}` with environment variables
3. Calls `RegisterLazyService(name, type, config)` automatically
4. Service created on first access via `GetService(name, cache)`

## Config Loading Flow

```
┌─────────────────────┐
│  1. Register        │  RegisterServiceFactory("email", EmailServiceFactory)
│     Factories       │  RegisterRouter("email-api", createEmailRouter())
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  2. Load YAML       │  config.LoadConfigFile("config.yaml", cfg)
│                     │  - Reads file
└──────────┬──────────┘  - Replaces ${ENV_VAR:default}
           │             - Parses YAML structure
           ▼
┌─────────────────────┐
│  3. Register Config │  lokstra_registry.RegisterConfig(cfg)
│                     │  - Automatically calls RegisterLazyService() for each service
└──────────┬──────────┘  - Stores server/app configurations
           │
           ▼
┌─────────────────────┐
│  4. Start Server    │  lokstra_registry.StartServer()
│                     │  - Creates apps based on config
│                     │  - Mounts routers
└─────────────────────┘  - Services created lazily on first request
```

## Complete Example

### 1. Register Factories (Code)

```go
func setupRegistry() {
    // Factories tell the framework how to create services
    lokstra_registry.RegisterServiceFactory("email", EmailServiceFactory)
    lokstra_registry.RegisterServiceFactory("counter", CounterServiceFactory)
    lokstra_registry.RegisterServiceFactory("logger", LoggerServiceFactory)
    
    // Routers for auto-discovery
    lokstra_registry.RegisterRouter("email-api", createEmailRouter())
    lokstra_registry.RegisterRouter("counter-api", createCounterRouter())
    lokstra_registry.RegisterRouter("logger-api", createLoggerRouter())
}
```

### 2. Define Services (YAML)

```yaml
# config.yaml
services:
  - name: email-service
    type: email
    config:
      smtp_host: ${SMTP_HOST:smtp.gmail.com}
      smtp_port: ${SMTP_PORT:587}
      from: ${EMAIL_FROM:demo@lokstra.dev}
  
  - name: counter-service
    type: counter
    config:
      name: ${COUNTER_NAME:yaml-counter}
      seed: ${COUNTER_SEED:200}
  
  - name: logger-service
    type: logger
    config:
      level: ${LOG_LEVEL:info}
      format: ${LOG_FORMAT:json}
      output: ${LOG_OUTPUT:stdout}
```

### 3. Load Config (Main)

```go
func main() {
    setupRegistry()
    
    // Load config
    cfg := config.New()
    config.LoadConfigFile("config.yaml", cfg)
    
    // Register and start
    lokstra_registry.RegisterConfig(cfg)
    lokstra_registry.SetCurrentServerName("config-demo-server")
    lokstra_registry.StartServer()
}
```

### 4. Use Services (Handler)

```go
r.POST("/api/email/send", func(c *lokstra.RequestContext) error {
    var req EmailRequest
    c.Req.BindBody(&req)
    
    // Service config loaded from YAML
    email := services.GetEmail()
    email.SendEmail(req.To, req.Subject, req.Body)
    
    return c.Api.Ok(map[string]any{"status": "sent"})
})
```

## Environment-Specific Configuration

### Development (defaults in YAML)

```yaml
services:
  - name: logger-service
    type: logger
    config:
      level: ${LOG_LEVEL:debug}    # Dev: debug
      format: ${LOG_FORMAT:pretty}  # Dev: pretty (readable)
```

### Production (environment variables)

```bash
# .env or Dockerfile
export LOG_LEVEL=warn
export LOG_FORMAT=json
export SMTP_HOST=smtp.production.com
export COUNTER_SEED=0
```

### Testing (override in CI/CD)

```bash
# GitHub Actions / GitLab CI
export LOG_LEVEL=error
export SMTP_HOST=smtp.test.com
export COUNTER_SEED=1000
```

## Configuration Methods

### 1. Default in YAML

```yaml
smtp_host: ${SMTP_HOST:smtp.gmail.com}  # Uses smtp.gmail.com if SMTP_HOST not set
```

**Use when:** Good default for development/testing

### 2. Required Environment Variable

```yaml
smtp_host: ${SMTP_HOST}  # No default - must be set!
```

**Use when:** Production secrets (passwords, API keys)

### 3. Multiple Config Files

```go
// Override configs (later files win)
config.LoadConfigFile("config.yaml", cfg)       // Base config
config.LoadConfigFile("config.prod.yaml", cfg)  // Production overrides
```

### 4. Named Configs

```yaml
configs:
  - name: server-name
    value: ${SERVER_NAME:demo-server}
  
  - name: app-env
    value: ${APP_ENV:development}
```

**Access in code:**
```go
serverName := lokstra_registry.GetConfig("server-name", "default")
appEnv := lokstra_registry.GetConfig("app-env", "development")
```

## Best Practices

### ✅ DO

1. **Use environment variables for secrets**
   ```yaml
   password: ${DB_PASSWORD}  # Not hardcoded!
   ```

2. **Provide good defaults for development**
   ```yaml
   smtp_host: ${SMTP_HOST:localhost}  # Dev-friendly
   ```

3. **Use descriptive variable names**
   ```yaml
   # Good
   database_url: ${DATABASE_URL:postgresql://localhost}
   
   # Bad
   db: ${DB:postgres}
   ```

4. **Document required environment variables**
   ```yaml
   # .env.example
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   EMAIL_FROM=demo@lokstra.dev
   ```

5. **Version control config.yaml, ignore .env**
   ```gitignore
   .env
   config.prod.yaml  # Contains sensitive defaults
   ```

### ❌ DON'T

1. **Don't hardcode secrets in YAML**
   ```yaml
   # BAD!
   password: my-secret-password
   
   # GOOD!
   password: ${DB_PASSWORD}
   ```

2. **Don't use same config for all environments**
   ```yaml
   # BAD - forces production to use debug
   log_level: debug
   
   # GOOD - allows per-environment override
   log_level: ${LOG_LEVEL:info}
   ```

3. **Don't forget to validate required vars**
   ```go
   // Check critical environment variables
   if os.Getenv("DB_PASSWORD") == "" {
       log.Fatal("DB_PASSWORD is required")
   }
   ```

## Testing Configuration

### 1. Test Default Values

```bash
# No environment variables set
go run .

# Check if defaults work
curl http://localhost:8080/health
```

### 2. Test Environment Overrides

```bash
# PowerShell
$env:LOG_LEVEL="debug"
$env:COUNTER_SEED="9999"
go run .

# Bash
LOG_LEVEL=debug COUNTER_SEED=9999 go run .
```

### 3. Test Multiple Environments

```bash
# Development
APP_ENV=development LOG_LEVEL=debug go run .

# Staging
APP_ENV=staging LOG_LEVEL=info go run .

# Production
APP_ENV=production LOG_LEVEL=warn go run .
```

## Common Patterns

### Database Configuration

```yaml
services:
  - name: dbpool
    type: dbpool_pg
    config:
      host: ${DB_HOST:localhost}
      port: ${DB_PORT:5432}
      database: ${DB_NAME:myapp}
      username: ${DB_USER:postgres}
      password: ${DB_PASSWORD:postgres}
      min_conns: ${DB_MIN_CONNS:2}
      max_conns: ${DB_MAX_CONNS:10}
```

### Redis Configuration

```yaml
services:
  - name: redis
    type: redis
    config:
      addr: ${REDIS_ADDR:localhost:6379}
      password: ${REDIS_PASSWORD:}
      db: ${REDIS_DB:0}
      pool_size: ${REDIS_POOL_SIZE:10}
```

### External API Configuration

```yaml
services:
  - name: payment-api
    type: http_client
    config:
      base_url: ${PAYMENT_API_URL:https://api.stripe.com}
      api_key: ${PAYMENT_API_KEY}
      timeout: ${PAYMENT_TIMEOUT:30s}
```

## Next Steps

- **03-service-dependencies** - Services that depend on other services
- **04-config-driven-deployment** - Complete app entirely from config.yaml

## Comparison with 01-Registry-Basics

| Feature | 01-Registry-Basics | 02-Config-Loading |
|---------|-------------------|-------------------|
| Service Definition | Code (RegisterLazyService) | YAML (services:) |
| Configuration | Hardcoded in code | Environment variables |
| Flexibility | Requires recompile | Change without recompile |
| Secrets | In code (bad!) | In env vars (good!) |
| Environments | One config only | Multiple configs |
| Best for | Learning, prototyping | Production deployment |
