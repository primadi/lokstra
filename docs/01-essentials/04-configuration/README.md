---
layout: docs
title: Configuration - YAML & Environment Setup
---

# Configuration - YAML & Environment Setup

> **Learn configuration management and environment strategies**  
> **Time**: 30-35 minutes • **Level**: Beginner • **Concepts**: 4

---

## 🎯 What You'll Learn

- Load configuration from YAML files
- Use environment variables in configuration
- Organize multi-environment configs
- Validate configuration automatically
- Apply Code + Config pattern (recommended!)

---

## 📖 Concepts

### 1. Configuration Basics

Lokstra supports **pure code**, **pure YAML**, or **code + YAML hybrid** approaches.

**Recommended: Code + Config Pattern**
- Write core logic in code
- Configure instances via YAML
- Best of both worlds!

```go
// Code: Define service factories
func NewPostgresService(params map[string]any) lokstra.Service {
    host := cast.GetValueFromMap(params, "host", "localhost")
    port := cast.GetValueFromMap(params, "port", 5432)
    // ... create and return service
}

func init() {
    lokstra_registry.RegisterServiceFactory("postgres", NewPostgresService)
}
```

```yaml
# YAML: Configure instances
services:
  - name: main-db
    type: postgres
    config:
      host: ${DB_HOST:localhost}
      port: ${DB_PORT:5432}
      database: myapp
```

**Why This Pattern?**
- ✅ Type-safe code
- ✅ Flexible configuration
- ✅ Easy environment management
- ✅ Best for production

### 2. Loading Configuration

#### Single File

```go
import "github.com/primadi/lokstra/core/config"

cfg := config.New()
err := config.LoadConfigFile("config.yaml", cfg)
if err != nil {
    log.Fatal(err)
}
```

#### Multiple Files (Merge)

```go
cfg := config.New()

// Load base config
config.LoadConfigFile("config/base.yaml", cfg)

// Load environment-specific (merges with base)
config.LoadConfigFile("config/production.yaml", cfg)
```

#### Directory (Auto-merge)

```go
cfg := config.New()

// Loads all .yaml and .yml files
// Merges them in alphabetical order
err := config.LoadConfigDir("config/", cfg)
```

**File Merge Example:**

```
config/
├── 01-base.yaml      # Loaded first
├── 02-services.yaml  # Merged second
└── 03-prod.yaml      # Merged last (overrides)
```

### 3. Environment Variables

Use `${VAR_NAME}` or `${VAR_NAME:default}` syntax:

```yaml
services:
  - name: database
    type: postgres
    config:
      # Required environment variable
      password: ${DB_PASSWORD}
      
      # With default value
      host: ${DB_HOST:localhost}
      port: ${DB_PORT:5432}
      
      # Multiple variables
      dsn: "postgres://${DB_USER:postgres}:${DB_PASSWORD}@${DB_HOST:localhost}:${DB_PORT:5432}/${DB_NAME:myapp}"
```

**Set environment variables:**

```bash
# Linux/Mac
export DB_HOST=prod-db.example.com
export DB_PASSWORD=secret123

# Windows
set DB_HOST=prod-db.example.com
set DB_PASSWORD=secret123
```

**In code:**

```go
os.Setenv("DB_HOST", "localhost")
os.Setenv("DB_PASSWORD", "secret")

cfg := config.New()
config.LoadConfigFile("config.yaml", cfg)
// Variables are expanded automatically
```

### 4. Configuration Validation

**Automatic Validation**

All config loading functions validate automatically:

```go
cfg := config.New()
err := config.LoadConfigFile("config.yaml", cfg)
if err != nil {
    // Error includes validation details
    log.Fatal(err)
    // Output:
    // validation failed: 
    //   - services.0.name: This field is required
    //   - servers.0.apps.0.addr: Does not match pattern
}
```

**Manual Validation**

```go
// Validate YAML string
yamlContent := `...`
err := config.ValidateYAMLString(yamlContent)

// Validate config struct
cfg := &config.Config{...}
err := config.ValidateConfig(cfg)
```

**What Gets Validated:**
- Required fields present
- Valid URL formats
- Name patterns (alphanumeric + underscore)
- Port ranges (1-65535)
- Array constraints

---

## 💻 Example 1: Basic YAML Configuration

**File: `config.yaml`**

```yaml
# Define services
services:
  - name: logger
    type: logger
    config:
      level: info
      format: json

# Define routers
routers:
  - name: api
    routes:
      - name: health
        path: /health
        handler: HealthCheckHandler
      
      - name: version
        path: /version
        handler: VersionHandler

# Define server
servers:
  - name: web-server
    baseUrl: http://localhost:8080
    apps:
      - name: api-app
        addr: /api
        routers:
          - api
```

**File: `main.go`**

```go
package main

import (
    "log"
    "net/http"
    
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/core/config"
    lokstra_registry "github.com/primadi/lokstra/lokstra_registry"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`{"status": "ok"}`))
}

func VersionHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`{"version": "1.0.0"}`))
}

func main() {
    // Register handlers
    lokstra_registry.RegisterHandler("HealthCheckHandler", HealthCheckHandler)
    lokstra_registry.RegisterHandler("VersionHandler", VersionHandler)
    
    // Load configuration
    cfg := config.New()
    if err := config.LoadConfigFile("config.yaml", cfg); err != nil {
        log.Fatal(err)
    }
    
    // Apply configuration and get server
    server, err := config.ApplyAllConfig(cfg, "web-server")
    if err != nil {
        log.Fatal(err)
    }
    
    // Start server
    log.Println("Server starting on :8080")
    server.Start()
}
```

**Run:**

```bash
go run main.go

# Test
curl http://localhost:8080/api/health
curl http://localhost:8080/api/version
```

---

## 💻 Example 2: Environment-Based Configuration

**Structure:**

```
myapp/
├── config/
│   ├── base.yaml       # Shared config
│   ├── dev.yaml        # Development
│   ├── staging.yaml    # Staging
│   └── prod.yaml       # Production
└── main.go
```

**File: `config/base.yaml`**

```yaml
# Shared configuration
routers:
  - name: api
    routes:
      - name: users
        path: /users
        handler: GetUsersHandler
        method: GET

services:
  - name: database
    type: postgres
    # Config will be overridden by environment files
```

**File: `config/dev.yaml`**

```yaml
# Development overrides
services:
  - name: database
    config:
      host: localhost
      port: 5432
      database: myapp_dev
      user: devuser
      password: devpass
      max_connections: 5

servers:
  - name: api-server
    baseUrl: http://localhost:3000
    apps:
      - name: api
        addr: /api/v1
        routers: [api]
```

**File: `config/prod.yaml`**

```yaml
# Production overrides
services:
  - name: database
    config:
      host: ${DB_HOST}
      port: ${DB_PORT:5432}
      database: ${DB_NAME}
      user: ${DB_USER}
      password: ${DB_PASSWORD}
      max_connections: 25
      ssl_mode: require

servers:
  - name: api-server
    baseUrl: ${API_BASE_URL}
    apps:
      - name: api
        addr: /api/v1
        routers: [api]
```

**File: `main.go`**

```go
package main

import (
    "log"
    "os"
    
    "github.com/primadi/lokstra/core/config"
    lokstra_registry "github.com/primadi/lokstra/lokstra_registry"
)

func main() {
    // Register factories and handlers
    lokstra_registry.RegisterServiceFactory("postgres", NewPostgresService)
    lokstra_registry.RegisterHandler("GetUsersHandler", GetUsersHandler)
    
    // Determine environment
    env := os.Getenv("APP_ENV")
    if env == "" {
        env = "dev"
    }
    
    // Load configuration
    cfg := config.New()
    
    // Load base
    if err := config.LoadConfigFile("config/base.yaml", cfg); err != nil {
        log.Fatal(err)
    }
    
    // Load environment-specific
    envFile := "config/" + env + ".yaml"
    if err := config.LoadConfigFile(envFile, cfg); err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Loaded configuration for environment: %s", env)
    
    // Apply and start
    server, err := config.ApplyAllConfig(cfg, "api-server")
    if err != nil {
        log.Fatal(err)
    }
    
    server.Start()
}
```

**Run:**

```bash
# Development
APP_ENV=dev go run main.go

# Production (with environment variables)
export APP_ENV=prod
export DB_HOST=prod-db.example.com
export DB_PORT=5432
export DB_NAME=myapp_prod
export DB_USER=produser
export DB_PASSWORD=secretpassword
export API_BASE_URL=https://api.example.com

go run main.go
```

---

## 💻 Example 3: Config References (CFG Resolver)

Use `${@CFG:path.to.config}` to reference other config values:

**File: `config.yaml`**

```yaml
# Define configuration values
configs:
  - name: features.debug
    value: true
  
  - name: features.timeout
    value: 30
  
  - name: database.max_connections
    value: 25
  
  - name: app.name
    value: MyApp

# Use config references
services:
  - name: logger
    type: logger
    config:
      debug: ${@CFG:features.debug}
      app_name: ${@CFG:app.name}
  
  - name: database
    type: postgres
    config:
      max_connections: ${@CFG:database.max_connections}
      connect_timeout: ${@CFG:features.timeout}

servers:
  - name: api-server
    baseUrl: http://localhost:8080
    apps:
      - name: api
        addr: /
```

**Benefits:**
- ✅ DRY - Define once, use many times
- ✅ Centralized configuration
- ✅ Easy to override per environment

**With Environment Overrides:**

```yaml
# base.yaml
configs:
  - name: features.debug
    value: true

# prod.yaml
configs:
  - name: features.debug
    value: false  # Override for production
```

---

## 🎯 Best Practices

### 1. Configuration Organization

**✅ DO: Use environment-based structure**

```
config/
├── base.yaml         # Shared configuration
├── dev.yaml          # Development
├── staging.yaml      # Staging
└── prod.yaml         # Production
```

**✅ DO: Use numbered prefixes for load order**

```
config/
├── 01-base.yaml
├── 02-services.yaml
├── 03-middlewares.yaml
└── 04-production.yaml
```

**✗ DON'T: Mix concerns in single file**

```yaml
# BAD: Everything in one file
services: [...]
middlewares: [...]
routers: [...]
servers: [...]
# Hard to maintain!
```

### 2. Environment Variables

**✅ DO: Use env vars for sensitive data**

```yaml
services:
  - name: database
    config:
      password: ${DB_PASSWORD}
      api_key: ${API_SECRET}
```

**✅ DO: Provide defaults for non-sensitive values**

```yaml
services:
  - name: database
    config:
      host: ${DB_HOST:localhost}
      port: ${DB_PORT:5432}
```

**✗ DON'T: Hardcode secrets**

```yaml
# BAD: Credentials in file
services:
  - name: database
    config:
      password: "hardcoded_password"  # NEVER DO THIS!
```

### 3. Configuration Validation

**✅ DO: Check errors on load**

```go
cfg := config.New()
if err := config.LoadConfigFile("config.yaml", cfg); err != nil {
    log.Fatal("Config error:", err)
}
```

**✅ DO: Validate before deployment**

```bash
# Test config validation
go run ./cmd/validate-config config.yaml
```

**✗ DON'T: Ignore validation errors**

```go
config.LoadConfigFile("config.yaml", cfg)  // BAD: No error check
```

### 4. Code + Config Pattern

**✅ DO: Define factories in code**

```go
// Code - Type-safe and testable
func NewPostgresService(params map[string]any) lokstra.Service {
    cfg := ParsePostgresConfig(params)
    return &PostgresService{config: cfg}
}

func init() {
    lokstra_registry.RegisterServiceFactory("postgres", NewPostgresService)
}
```

**✅ DO: Configure instances in YAML**

```yaml
# YAML - Easy to change per environment
services:
  - name: main-db
    type: postgres  # References factory
    config:
      host: ${DB_HOST}
      port: ${DB_PORT}
```

**✗ DON'T: Put logic in YAML**

```yaml
# BAD: YAML can't contain logic
services:
  - name: database
    config:
      # This won't work - no conditionals in YAML!
      timeout: if debug then 60 else 30
```

---

## 🔍 Common Patterns

### Pattern 1: Feature Flags

```yaml
configs:
  - name: features.new_ui
    value: false
  
  - name: features.beta_api
    value: true

middlewares:
  - name: feature-flags
    type: feature-flags
    config:
      new_ui: ${@CFG:features.new_ui}
      beta_api: ${@CFG:features.beta_api}
```

### Pattern 2: Multi-Region Config

```yaml
# config/base.yaml
configs:
  - name: region
    value: ${REGION:us-east-1}

services:
  - name: database
    type: postgres
    config:
      host: db-${@CFG:region}.example.com
```

```bash
# Deploy to different regions
REGION=us-east-1 go run main.go
REGION=eu-west-1 go run main.go
REGION=ap-south-1 go run main.go
```

### Pattern 3: Service Composition

```yaml
services:
  # Base services
  - name: postgres
    type: postgres
    config:
      host: ${DB_HOST:localhost}
  
  - name: redis
    type: redis
    config:
      host: ${REDIS_HOST:localhost}
  
  # Composite service using others
  - name: user-service
    type: user-service
    depends_on:
      - postgres
      - redis
    config:
      cache_enabled: true
```

---

## 📚 Configuration Reference

### Complete YAML Structure

```yaml
# Configuration values
configs:
  - name: string           # Config key (use dotted notation)
    value: any             # Any value (string, number, bool, etc)

# Service definitions
services:
  - name: string           # Service name
    type: string           # Factory type
    depends_on: [string]   # Optional dependencies
    config: map            # Service-specific configuration

# Middleware definitions
middlewares:
  - name: string           # Middleware name
    type: string           # Factory type
    config: map            # Middleware-specific configuration

# Router definitions
routers:
  - name: string           # Router name
    engine_type: string    # Optional: default, gin, etc
    middleware: [string]   # Router-level middleware
    routes:
      - name: string       # Route name
        path: string       # URL path
        method: string     # HTTP method (GET, POST, etc)
        handler: string    # Handler name
        middleware: [string]  # Route-level middleware

# Server definitions
servers:
  - name: string           # Server name
    baseUrl: string        # Base URL
    apps:
      - name: string       # App name
        addr: string       # Mount path
        routers: [string]  # Router names
```

---

## ✅ Quick Checklist

After completing this section, you should be able to:

- [ ] Load configuration from YAML files
- [ ] Use environment variables in YAML
- [ ] Organize multi-environment configurations
- [ ] Validate configuration automatically
- [ ] Use CFG references for DRY config
- [ ] Apply Code + Config pattern

---

## 🚀 Next Steps

**Ready for more?** Continue to:

👉 [App & Server](../05-app-and-server/README.md) - Application lifecycle and server management

**Or explore:**
- [Complete Example](../06-putting-it-together/README.md) - Full application
- [API Reference - Configuration](../../03-api-reference/03-configuration/README.md) - Detailed docs
