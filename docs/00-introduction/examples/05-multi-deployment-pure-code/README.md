# Pure Code Deployment Example

This example demonstrates **100% code-based deployment configuration** without any YAML files.

## What's Different from 04-multi-deployment?

**04-multi-deployment:**
- Uses `config.yaml` for service definitions and deployments
- Calls `lokstra_registry.LoadAndBuild([]string{"config.yaml"})`

**05-pure-code-new:**
- ❌ NO config.yaml
- ✅ Uses `RegisterLazyService` for service definitions
- ✅ Uses `RegisterDeployment` for deployment topology
- ✅ 100% pure Go code

## What's the Same?

Everything else is identical:
- Same service factories (with LOCAL and REMOTE factories)
- Same clean architecture structure (contract, model, service, repository)
- Same deployments: monolith (port 3003) and microservice (ports 3004, 3005)
- Same auto-router generation from metadata

## Running the Example

### Monolith (all services in one process)
```bash
go run . -server monolith.api-server
```

### Microservices

**User Server** (port 3004):
```bash
go run . -server microservice.user-server
```

**Order Server** (port 3005):
```bash
# In another terminal:
go run . -server microservice.order-server
```

## Testing

Use `test.http` with REST Client extension (same as 04-multi-deployment).

## Key Code Changes

**Before (04-multi-deployment):**
```go
lokstra_registry.LoadAndBuild([]string{"config.yaml"})
```

**After (05-pure-code-new):**
```go
// Service definitions
lokstra_registry.RegisterLazyService("user-service", "user-service-factory", 
    map[string]any{"depends-on": []string{"user-repository"}})

lokstra_registry.RegisterLazyService("order-service", "order-service-factory", 
    map[string]any{"depends-on": []string{"order-repository", "user-service"}})

// Deployments
lokstra_registry.RegisterDeployment("monolith", &lokstra_registry.DeploymentConfig{
    Servers: map[string]*lokstra_registry.ServerConfig{
        "api-server": {
            BaseURL: "http://localhost",
            Addr: ":3003",
            PublishedServices: []string{"user-service", "order-service"},
        },
    },
})
```

## Benefits of Pure Code Approach

1. **Type Safety** - IDE autocomplete and compile-time checking
2. **Refactoring** - Rename services safely with IDE refactoring tools
3. **No Schema Validation Needed** - Go compiler is the validator
4. **Dynamic Configuration** - Can use conditionals, loops, functions
5. **Single Language** - No context switching between Go and YAML

## When to Use Which?

| Approach | Best For |
|----------|----------|
| **YAML** (04) | Ops teams, non-coders, runtime config changes |
| **Pure Code** (05) | Dev teams, version control, compile-time safety |

Both approaches produce identical runtime behavior!

## Architecture Details

See `index.md` (copied from 04-multi-deployment) for complete architecture documentation, including:
- Clean Architecture patterns
- Auto-router generation
- Metadata-driven routing
- Proxy service patterns
- Deployment topologies
