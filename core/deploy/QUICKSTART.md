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

### Step 3: Test Config Resolution

```go
package main

import (
    "testing"
    "github.com/primadi/lokstra/core/deploy"
    "github.com/primadi/lokstra/core/deploy/schema"
)

func TestConfigResolution(t *testing.T) {
    // Create registry
    reg := deploy.NewGlobalRegistry()
    
    // Define config
    reg.DefineConfig(&schema.ConfigDef{
        Name: "DB_MAX_CONNS",
        Value: 20,
    })
    
    // Resolve
    if err := reg.ResolveConfigs(); err != nil {
        t.Fatal(err)
    }
    
    // Get resolved value
    value, ok := reg.GetResolvedConfig("DB_MAX_CONNS")
    if !ok {
        t.Fatal("config not found")
    }
    
    // Verify type preserved
    if value != 20 {
        t.Errorf("expected int 20, got %v (type %T)", value, value)
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

## Custom Resolver Example

```go
package main

import "github.com/primadi/lokstra/core/deploy/resolver"

// Consul resolver
type ConsulResolver struct{}

func (c *ConsulResolver) Name() string {
    return "consul"
}

func (c *ConsulResolver) Resolve(key string) (string, bool) {
    // Fetch from Consul
    value, err := consulClient.Get(key)
    if err != nil {
        return "", false
    }
    return value, true
}

func init() {
    // Register custom resolver
    deploy.Global().RegisterResolver(&ConsulResolver{})
}
```

**Use in YAML:**
```yaml
configs:
  - name: API_KEY
    value: ${@consul:config/api-key:fallback}
```

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

### Mock Resolver for Tests

```go
func TestWithMockResolver(t *testing.T) {
    reg := deploy.NewGlobalRegistry()
    
    // Register mock resolver
    reg.RegisterResolver(resolver.NewStaticResolver("mock", map[string]string{
        "api-key": "test-key-123",
    }))
    
    // Define config using mock resolver
    reg.DefineConfig(&schema.ConfigDef{
        Name: "API_KEY",
        Value: "${@mock:api-key}",
    })
    
    // Resolve and test
    if err := reg.ResolveConfigs(); err != nil {
        t.Fatal(err)
    }
    
    value, _ := reg.GetResolvedConfig("API_KEY")
    if value != "test-key-123" {
        t.Errorf("expected 'test-key-123', got %v", value)
    }
}
```

## Troubleshooting

### Config Not Found Error

```
Error: config key DB_MAX_CONNS not found (referenced in ${@cfg:DB_MAX_CONNS})
```

**Solution:** Make sure the config is defined before resolution:
```yaml
configs:
  - name: DB_MAX_CONNS
    value: 20
```

### Type Mismatch

```yaml
# Wrong: This becomes string "20"
config:
  max-conns: "20"

# Correct: This stays as int 20
configs:
  - name: DB_MAX_CONNS
    value: 20

config:
  max-conns: ${@cfg:DB_MAX_CONNS}
```

### Resolver Not Found

```
Error: resolver consul not found (in ${@consul:config/api-key})
```

**Solution:** Register the resolver first:
```go
deploy.Global().RegisterResolver(&ConsulResolver{})
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

**Q: Do I need to use @cfg for all config references?**  
A: Only when you want to reference another config value. Direct env vars don't need it: `${DATABASE_URL}`

**Q: Why ${@cfg:...} instead of ${cfg:...}?**  
A: The `@` prefix clearly distinguishes custom resolvers from env vars, making parsing easier.

**Q: Can I have multiple deployments in one YAML?**  
A: Yes! The `deployments:` section is an array. See example.yaml.

**Q: How do I switch between deployments?**  
A: Pass deployment name to `deploy.LoadYAML(file, deploymentName)` (Phase 2 feature, coming soon).

**Q: Is the global registry thread-safe?**  
A: Yes! All operations are protected with RWMutex.
