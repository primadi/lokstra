# CFG Resolver - Configuration Reference System

## Overview

The CFG resolver is a built-in variable resolver that allows you to reference values from the `configs` section of your YAML configuration file. This enables you to define configuration values once and reuse them throughout your configuration, eliminating duplication and making configurations easier to maintain.

## Why CFG Resolver?

Before the CFG resolver, there was a circular dependency problem:
1. Variable expansion happened before YAML parsing
2. CFG resolver needed access to parsed config values
3. But parsing couldn't happen until variables were expanded

The solution is a **two-pass expansion system** that happens entirely within `expandVariables()`:
1. **Phase 1**: Expand all resolvers EXCEPT CFG (ENV, AWS, K8S, etc.)
2. **Phase 2**: Parse YAML partially to extract `configs` section, store in temporary registry, then expand CFG placeholders

This self-contained approach eliminates the circular dependency while keeping the expansion logic centralized.

## Syntax

```yaml
${@CFG:config.key}                    # Reference a config value
${@CFG:config.key:default}            # With default value
${@CFG:nested.path.to.value}          # Nested configuration paths
```

## Configuration Structure

The `configs` section is an array of name-value pairs:

```yaml
configs:
  - name: database.host
    value: postgres.example.com
  - name: database.port
    value: 5432
  - name: database.name
    value: myapp
  - name: features.debug
    value: false
  - name: features.timeout
    value: 30
```

## Examples

### Basic Usage

Define configuration values once and reuse them:

```yaml
configs:
  - name: database.host
    value: postgres.example.com
  - name: database.port
    value: 5432

servers:
  - name: api-server
    baseUrl: "http://${@CFG:database.host}:${@CFG:database.port}"
    apps:
      - name: api
        addr: /api

services:
  - name: postgres
    type: database
    config:
      host: "${@CFG:database.host}"
      port: ${@CFG:database.port}
```

### Nested Configuration Paths

Use dot notation for nested configurations:

```yaml
configs:
  - name: database.primary.host
    value: db1.example.com
  - name: database.primary.port
    value: 5432
  - name: database.replica.host
    value: db2.example.com
  - name: database.replica.port
    value: 5433

services:
  - name: db-service
    type: database
    config:
      primary_dsn: "postgres://${@CFG:database.primary.host}:${@CFG:database.primary.port}/db"
      replica_dsn: "postgres://${@CFG:database.replica.host}:${@CFG:database.replica.port}/db"
```

### Default Values

Provide fallback values for missing configurations:

```yaml
configs:
  - name: database.host
    value: postgres.example.com
  # port is not defined

servers:
  - name: api-server
    baseUrl: "http://${@CFG:database.host}:${@CFG:database.port:5432}"
    # Uses 5432 as default since database.port is not in configs
```

### Mixed Resolvers

Combine CFG with other resolvers (ENV, AWS, K8S, etc.):

```yaml
configs:
  - name: features.debug
    value: false
  - name: features.timeout
    value: 30

servers:
  - name: "${APP_ENV}-server"  # From environment variable
    baseUrl: "http://localhost:8080"
    apps:
      - name: api
        addr: /

middlewares:
  - name: custom
    type: custom
    config:
      environment: "${APP_ENV}"              # ENV resolver
      debug: ${@CFG:features.debug}          # CFG resolver
      timeout: ${@CFG:features.timeout}      # CFG resolver
      aws_region: "${@AWS:region}"           # AWS resolver (if configured)
```

### Data Types

CFG resolver supports all YAML data types:

```yaml
configs:
  - name: app.name
    value: "MyApp"                    # String
  - name: app.port
    value: 8080                        # Integer
  - name: app.debug
    value: true                        # Boolean
  - name: app.timeout
    value: 30.5                        # Float

servers:
  - name: "${@CFG:app.name}"
    baseUrl: "http://localhost:${@CFG:app.port}"
    apps:
      - name: api
        addr: /

services:
  - name: app-service
    type: custom
    config:
      name: "${@CFG:app.name}"
      port: ${@CFG:app.port}
      debug: ${@CFG:app.debug}
      timeout: ${@CFG:app.timeout}
```

## Use Cases

### 1. Environment-Specific Configuration

Define base configurations that can be overridden per environment:

```yaml
configs:
  - name: database.host
    value: "${DB_HOST:localhost}"      # From ENV with default
  - name: database.port
    value: 5432
  - name: redis.host
    value: "${REDIS_HOST:localhost}"
  - name: redis.port
    value: 6379

services:
  - name: postgres
    type: database
    config:
      host: "${@CFG:database.host}"
      port: ${@CFG:database.port}
  
  - name: redis
    type: redis
    config:
      host: "${@CFG:redis.host}"
      port: ${@CFG:redis.port}
```

### 2. Shared Configuration Across Services

Define common settings once and share across multiple services:

```yaml
configs:
  - name: common.timeout
    value: 30
  - name: common.retries
    value: 3
  - name: common.log_level
    value: info

services:
  - name: api-service
    type: api
    config:
      timeout: ${@CFG:common.timeout}
      retries: ${@CFG:common.retries}
      log_level: "${@CFG:common.log_level}"
  
  - name: worker-service
    type: worker
    config:
      timeout: ${@CFG:common.timeout}
      retries: ${@CFG:common.retries}
      log_level: "${@CFG:common.log_level}"
```

### 3. Complex Connection Strings

Build connection strings from individual configuration values:

```yaml
configs:
  - name: db.user
    value: "${DB_USER:postgres}"
  - name: db.password
    value: "${DB_PASSWORD:secret}"
  - name: db.host
    value: "${DB_HOST:localhost}"
  - name: db.port
    value: 5432
  - name: db.name
    value: "${DB_NAME:myapp}"

services:
  - name: postgres
    type: database
    config:
      dsn: "postgres://${@CFG:db.user}:${@CFG:db.password}@${@CFG:db.host}:${@CFG:db.port}/${@CFG:db.name}"
```

### 4. Feature Flags

Centralize feature flag management:

```yaml
configs:
  - name: features.new_api
    value: true
  - name: features.experimental_cache
    value: false
  - name: features.debug_mode
    value: "${DEBUG:false}"

middlewares:
  - name: feature-flags
    type: feature-flags
    config:
      new_api: ${@CFG:features.new_api}
      experimental_cache: ${@CFG:features.experimental_cache}
      debug_mode: ${@CFG:features.debug_mode}
```

## Implementation Details

### Two-Pass Expansion

The CFG resolver uses a two-pass expansion system:

1. **Phase 1 (expandExceptCFG)**:
   - Expand all resolvers EXCEPT CFG (ENV, AWS, K8S, custom resolvers)
   - This ensures that environment variables used in `configs` section are expanded first

2. **Phase 2 (expandCFGWithTempRegistry)**:
   - Parse the YAML to extract `configs` section
   - Store configs in temporary registry
   - Expand all CFG placeholders using temporary registry
   - Return fully expanded YAML

### Temporary CFG Registry

During expansion, a `TempCFGResolver` is created that:
- Parses the `configs` section from YAML
- Stores name-value pairs in a map
- Implements the `VariableResolver` interface
- Supports nested paths using dot notation
- Handles all YAML data types (string, int, bool, float)

### Placeholder Preservation

If a CFG key doesn't exist and no default is provided:
- The original placeholder is preserved: `${@CFG:missing.key}`
- This allows for debugging and identifying missing configurations
- With default value: `${@CFG:missing.key:default}` → `default`

## Best Practices

### 1. Use Descriptive Names

```yaml
# Good
configs:
  - name: database.primary.host
  - name: database.replica.host
  - name: redis.cache.host

# Avoid
configs:
  - name: db1
  - name: db2
  - name: cache
```

### 2. Group Related Configurations

```yaml
configs:
  # Database configurations
  - name: database.host
    value: postgres.example.com
  - name: database.port
    value: 5432
  - name: database.name
    value: myapp
  
  # Redis configurations
  - name: redis.host
    value: redis.example.com
  - name: redis.port
    value: 6379
  
  # Feature flags
  - name: features.debug
    value: false
  - name: features.experimental
    value: true
```

### 3. Combine with Environment Variables

Use ENV resolver in `configs` for environment-specific overrides:

```yaml
configs:
  - name: app.env
    value: "${APP_ENV:development}"
  - name: database.host
    value: "${DB_HOST:localhost}"
  - name: database.port
    value: "${DB_PORT:5432}"
```

### 4. Always Provide Defaults

```yaml
# Good - has default
baseUrl: "http://${@CFG:app.host:localhost}:${@CFG:app.port:8080}"

# Risky - no default, might be empty if key missing
baseUrl: "http://${@CFG:app.host}:${@CFG:app.port}"
```

### 5. Use Type-Appropriate Placeholders

```yaml
# Strings - use quotes
name: "${@CFG:app.name}"

# Numbers - no quotes
port: ${@CFG:app.port}
timeout: ${@CFG:app.timeout}

# Booleans - no quotes
debug: ${@CFG:app.debug}
enabled: ${@CFG:feature.enabled}
```

## Limitations

### 1. No Array/Object Expansion

CFG resolver works on raw text before YAML parsing, so you cannot expand arrays or objects directly:

```yaml
# This does NOT work
configs:
  - name: cors.origins
    value:
      - https://app.example.com
      - https://admin.example.com

middlewares:
  - name: cors
    config:
      origins: ${@CFG:cors.origins}  # ❌ Cannot expand array
```

**Workaround**: Define individual values:

```yaml
configs:
  - name: cors.origin1
    value: "https://app.example.com"
  - name: cors.origin2
    value: "https://admin.example.com"

middlewares:
  - name: cors
    config:
      origins:
        - "${@CFG:cors.origin1}"
        - "${@CFG:cors.origin2}"
```

### 2. No Circular References

CFG values cannot reference other CFG values:

```yaml
# This does NOT work
configs:
  - name: base.url
    value: "https://example.com"
  - name: api.url
    value: "${@CFG:base.url}/api"  # ❌ Circular reference
```

**Workaround**: Use environment variables or define the complete value:

```yaml
configs:
  - name: base.url
    value: "${BASE_URL:https://example.com}"
  - name: api.url
    value: "${BASE_URL:https://example.com}/api"
```

## Testing

Comprehensive tests are available in:
- `core/config/var_resolver_cfg_test.go` - Unit tests for CFG expansion
- `core/config/cfg_integration_test.go` - Integration tests with full config loading

Run tests:
```bash
cd core/config
go test -v -run TestExpandVariablesWithCFG  # Unit tests
go test -v -run TestLoadConfigWith          # Integration tests
go test -v                                   # All tests
```

## Comparison with Other Resolvers

| Resolver | Source | Use Case |
|----------|--------|----------|
| ENV | Environment variables | OS-level configuration |
| AWS | AWS SSM Parameter Store | AWS-specific secrets |
| K8S | Kubernetes Secrets | Kubernetes cluster secrets |
| **CFG** | Config file itself | **Config reuse within same file** |

## Performance

- **Two-pass expansion overhead**: Minimal (< 1ms for typical configs)
- **Temporary registry**: In-memory, no I/O operations
- **YAML parsing**: Only `configs` section parsed during expansion
- **Caching**: None needed (expansion happens once during config load)

## Future Enhancements

Potential improvements (not yet implemented):
1. Support for array/object expansion
2. CFG-to-CFG references
3. Computed values (expressions)
4. Validation of CFG key existence at parse time
5. IDE support for autocomplete of CFG keys

## Related Documentation

- [Variable Resolver System](./environment-variable-syntax.md)
- [Custom Resolvers](../core/config/custom_resolver/README.md)
- [K8s Secret Resolver](../core/config/custom_resolver/README.md#kubernetes-secret-resolver)
- [YAML Configuration](./yaml-configuration-system.md)
