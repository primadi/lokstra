# Quick Start Guide

## Installation

The `deploy` package is part of Lokstra core. No additional installation needed.

```go
import "github.com/primadi/lokstra/core/deploy"
```

## 5-Minute Tutorial

### Step 1: Create YAML Configuration

**deployment.yaml:**
```yaml
# Global configs
configs:
  - name: DB_MAX_CONNS
    value: 20
  - name: DB_URL
    value: ${DATABASE_URL:postgres://localhost/mydb}

# Services
services:
  - name: db-pool
    type: dbpool_pg
    config:
      dsn: ${DB_URL}
      max-conns: ${@cfg:DB_MAX_CONNS}
  
  - name: my-service
    type: my-service-factory
    depends-on: [db-pool]

# Service routers
service-routers:
  - name: my-service-api
    service: my-service
    convention: rest

# Deployment
deployments:
  - name: dev
    servers:
      - name: dev-server
        base-url: http://localhost
        apps:
          - port: 3000
            services: [db-pool, my-service]
            service-routers:
              - name: my-service-api
```

### Step 2: Register Factories (Code)

**main.go:**
```go
package main

import (
    "log"
    "github.com/primadi/lokstra/core/deploy"
    "github.com/primadi/lokstra/core/service"
)

// Service implementation
type MyService struct {
    db *DBPool
}

// Factory functions
func myServiceFactoryLocal(deps map[string]any, config map[string]any) any {
    dbPool := deps["db-pool"].(service.Cached[*DBPool])
    return &MyService{db: dbPool.Get()}
}

func dbPoolFactory(deps map[string]any, config map[string]any) any {
    dsn := config["dsn"].(string)
    maxConns := config["max-conns"].(int)  // Note: int, not string!
    return NewDBPool(dsn, maxConns)
}

func init() {
    // Register factories
    deploy.Global().RegisterServiceType("my-service-factory", 
        myServiceFactoryLocal, 
        nil) // no remote factory for this example
    
    deploy.Global().RegisterServiceType("dbpool_pg", 
        dbPoolFactory, 
        nil)
}

func main() {
    // TODO: Load YAML and run deployment
    // (Phase 2 implementation)
    log.Println("Factories registered successfully!")
}
```

### Step 3: Test Config Access

```go
package main

import (
    "testing"
    "github.com/primadi/lokstra/lokstra_registry"
)

func TestConfigAccess(t *testing.T) {
    // Set config value
    lokstra_registry.SetConfig("db.max-conns", 20)
    
    // Get config value with type-safe generic
    maxConns := lokstra_registry.GetConfig("db.max-conns", 10)
    
    // Verify value and type
    if maxConns != 20 {
        t.Errorf("expected int 20, got %v (type %T)", maxConns, maxConns)
    }
    
    // Test nested access
    lokstra_registry.SetConfig("database", map[string]any{
        "host": "localhost",
        "port": 5432,
    })
    
    // Access as nested map
    dbConfig := lokstra_registry.GetConfig[map[string]any]("database", nil)
    if dbConfig["host"] != "localhost" {
        t.Errorf("expected 'localhost', got %v", dbConfig["host"])
    }
}
```

## Common Patterns

### Pattern 1: Environment Variables with Defaults

```yaml
configs:
  - name: PORT
    value: ${PORT:3000}
  
  - name: LOG_LEVEL
    value: ${LOG_LEVEL:info}
```

### Pattern 2: Config References (Type Preservation)

```yaml
configs:
  - name: MAX_CONNECTIONS
    value: 100

services:
  - name: db
    config:
      max-conns: ${@cfg:MAX_CONNECTIONS}  # Stays as int 100
```

### Pattern 3: Service Dependencies

```yaml
services:
  - name: user-service
    type: user-factory
    depends-on: [db, logger]  # Simple

  - name: order-service
    type: order-factory
    depends-on: ["dbPool:db", "userSvc:user-service"]  # With aliases
```

### Pattern 4: Router Overrides

```yaml
router-overrides:
  - name: public-api
    path-prefix: /api/v1
    middlewares: [cors]
    hidden: [Delete, AdminReset]
    routes:
      - name: Create
        middlewares: [rate-limit]
```

### Pattern 5: Deployment Config Overrides

```yaml
deployments:
  - name: dev
    config-overrides:
      LOG_LEVEL: debug
      MAX_CONNECTIONS: 10
  
  - name: prod
    config-overrides:
      LOG_LEVEL: error
      MAX_CONNECTIONS: 1000
```

## Config Resolution with Providers

Lokstra uses a **2-step resolution** at YAML byte level (before unmarshaling):

**STEP 1:** Resolve non-@cfg variables
```yaml
configs:
  database:
    host: ${DB_HOST:localhost}           # ENV variable
    secret: ${@aws-secret:db/password}   # AWS Secrets Manager
    token: ${@vault:secret/api-key}      # HashiCorp Vault
```

**STEP 2:** Resolve @cfg references (using step 1 results)
```yaml
configs:
  max-connections: 100

services:
  - name: db
    config:
      max-conns: ${@cfg:max-connections}  # References config value
```

**Benefits:**
- ✅ All fields auto-resolved (no manual iteration)
- ✅ Extensible via provider registry
- ✅ Type-safe after resolution
- ✅ Single quote escaping: `${@vault:'path:with:colons'}`

See `core/deploy/loader/provider_registry.go` for available providers.

## Testing Best Practices

### Isolated Registry for Tests

```go
func TestMyService(t *testing.T) {
    // Create isolated registry (not global)
    reg := deploy.NewGlobalRegistry()
    
    // Register test factory
    reg.RegisterServiceType("test-service", testFactory, nil)
    
    // Define test config
    reg.DefineConfig(&schema.ConfigDef{
        Name: "TEST_VALUE",
        Value: "test",
    })
    
    // Test...
}
```

### Config with Different Types

```go
func TestConfigTypes(t *testing.T) {
    // String config
    lokstra_registry.SetConfig("api.key", "test-key-123")
    apiKey := lokstra_registry.GetConfig("api.key", "")
    if apiKey != "test-key-123" {
        t.Errorf("expected 'test-key-123', got %v", apiKey)
    }
    
    // Integer config
    lokstra_registry.SetConfig("server.port", 8080)
    port := lokstra_registry.GetConfig("server.port", 3000)
    if port != 8080 {
        t.Errorf("expected 8080, got %v", port)
    }
    
    // Map config with auto-flattening
    lokstra_registry.SetConfig("database", map[string]any{
        "host": "localhost",
        "port": 5432,
    })
    
    // Access flattened keys
    dbHost := lokstra_registry.GetConfig("database.host", "")
    if dbHost != "localhost" {
        t.Errorf("expected 'localhost', got %v", dbHost)
    }
}
```

## Troubleshooting

### Config Not Found Error

```
Error: config 'db.max-conns' not found
```

**Solution:** Make sure the config is set before access:
```go
// In YAML (resolved at load time)
configs:
  database:
    max-conns: 20

// OR in code (runtime)
lokstra_registry.SetConfig("database.max-conns", 20)
```

### Type Mismatch

```yaml
# Wrong: String instead of int
configs:
  database:
    max-conns: "20"  # String

# Correct: Native YAML type
configs:
  database:
    max-conns: 20    # Integer

# Also correct: Reference preserves type
configs:
  max-connections: 100

service-definitions:
  db:
    config:
      max-conns: ${@cfg:max-connections}  # Stays as int 100
```

### Factory Not Registered

```
Panic: service type user-factory not registered in deployment dev
```

**Solution:** Register factory before loading YAML:
```go
deploy.Global().RegisterServiceType("user-factory", factory, nil)
```

## Next Steps

- Read [README.md](README.md) for complete documentation
- See [example.yaml](example.yaml) for full configuration example
- Check [IMPLEMENTATION-SUMMARY.md](IMPLEMENTATION-SUMMARY.md) for implementation details

## FAQ

**Q: Can I use code instead of YAML?**  
A: Yes! You can define everything programmatically using `registry.DefineConfig()`, `registry.DefineService()`, etc.

**Q: How does config resolution work?**  
A: 2-step resolution at YAML byte level: (1) ENV/provider variables → (2) @cfg references. All automatic before unmarshaling.

**Q: Can I add custom providers (like Vault, Consul)?**  
A: Yes! See `core/deploy/loader/provider_registry.go`. Implement the provider interface and register it.

**Q: Can I have multiple deployments in one YAML?**  
A: Yes! The `deployments:` section is an array. See example.yaml.

**Q: How do I switch between deployments?**  
A: Pass deployment name to `deploy.LoadYAML(file, deploymentName)` (Phase 2 feature, coming soon).

**Q: Is the global registry thread-safe?**  
A: Yes! All operations are protected with RWMutex.
