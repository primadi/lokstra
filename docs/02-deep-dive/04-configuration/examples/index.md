# Configuration Deep Dive - Examples

This folder contains advanced configuration patterns and multi-deployment strategies.

## Examples

### ✅ 01 - Monolith to Microservices
Same codebase, different deployment configurations.

**Topics**: Multi-deployment, monolith, microservices, hybrid

[View Example](./01-monolith-to-microservices/) | [main.go](./01-monolith-to-microservices/main.go) | [test.http](./01-monolith-to-microservices/test.http)

### ✅ 02 - Environment Management
Handle dev, staging, production environments.

**Topics**: Environment configs, overrides, feature flags

[View Example](./02-environment-management/) | [main.go](./02-environment-management/main.go) | [test.http](./02-environment-management/test.http)

### ✅ 03 - Configuration Validation
Validate configuration at startup and runtime.

**Topics**: Schema validation, required fields, fail-fast

[View Example](./03-configuration-validation/) | [main.go](./03-configuration-validation/main.go) | [test.http](./03-configuration-validation/test.http)

### ✅ 04 - Dynamic Configuration
Hot reload and feature toggle patterns.

**Topics**: Runtime updates, feature flags, thread-safety

[View Example](./04-dynamic-configuration/) | [main.go](./04-dynamic-configuration/main.go) | [test.http](./04-dynamic-configuration/test.http)

### ✅ 05 - Secrets Management
Secure handling of sensitive configuration.

**Topics**: Environment variables, secret stores, best practices

[View Example](./05-secrets-management/) | [main.go](./05-secrets-management/main.go) | [test.http](./05-secrets-management/test.http)

### ✅ 06 - Production Patterns
Real-world production configuration examples.

**Topics**: Health checks, metrics, graceful shutdown

[View Example](./06-production-patterns/) | [main.go](./06-production-patterns/main.go) | [test.http](./06-production-patterns/test.http)

---

## Running Examples

Each example follows this structure:
```
01-monolith-to-microservices/
├── main.go              # Application code
├── index             # Documentation
├── test.http            # HTTP test requests
├── config-*.yaml        # Configuration files
```

To run an example:
```bash
cd 01-monolith-to-microservices

# Run with specific config
go run main.go monolith
go run main.go microservices

# Test endpoints
# Use test.http file or curl
```

---

## Configuration Patterns

### Environment-Based Loading

```go
env := os.Getenv("APP_ENV")
if env == "" {
    env = "development"
}

cfg := config.New()
config.LoadConfigFile("config-base.yaml", cfg)
config.LoadConfigFile(fmt.Sprintf("config-%s.yaml", env), cfg)
```

### Validation

```go
func ValidateConfig(cfg *config.Config) error {
    required := []string{"app_name", "db_host"}
    for _, req := range required {
        if !hasConfig(cfg, req) {
            return fmt.Errorf("missing: %s", req)
        }
    }
    return nil
}
```

### Secrets

```go
// ✅ DO: Use environment variables
dbPassword := os.Getenv("DB_PASSWORD")

// ❌ DON'T: Hardcode in config files
config:
  db_password: "hardcoded"  # NEVER!
```

---

**Status**: ✅ All 6 configuration examples complete and ready to use!
