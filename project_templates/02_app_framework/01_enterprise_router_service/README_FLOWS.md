# Bootstrap Flow Examples

This directory contains examples demonstrating both bootstrap flows.

## Files

- `main.go` - Old flow (current, still supported)
- `main_new_flow.go` - New flow (recommended)

## Quick Comparison

### Old Flow
```go
registerServiceTypes()
registerMiddlewareTypes()
lokstra_registry.RunServerFromConfigFolder("config")
```

**Problem:** Config not available during registration

### New Flow
```go
lokstra_registry.LoadConfigFromFolder("config")
registerServiceTypes()    // Config available here!
registerMiddlewareTypes() // Config available here!
lokstra.InitAndRunServer()
```

**Benefits:** Early config loading, services can access config

## Running Examples

```bash
# Old flow
go run main.go register.go

# New flow (recommended)
go run main_new_flow.go register.go
```

## See Also

- [BOOTSTRAP-FLOWS.md](../../../docs/BOOTSTRAP-FLOWS.md) - Complete documentation
- [AI-AGENT-GUIDE.md](../../../docs/AI-AGENT-GUIDE.md) - AI agent guide
