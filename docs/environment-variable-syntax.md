# Environment Variable Syntax

## Overview

Lokstra supports environment variable expansion in YAML configuration files using the `${...}` syntax. This allows you to:
- Use environment variables for sensitive data (API keys, passwords)
- Override configuration per environment (dev, staging, production)
- Support default values when variables are not set
- Integrate with external secret management systems (AWS Secrets Manager, HashiCorp Vault, etc.)

## Syntax

### Basic Environment Variable (ENV Resolver)

```yaml
# Without default - uses empty string if not set
port: ${PORT}

# With default value
port: ${PORT:8080}

# Default can contain colons (URLs, DSNs, etc.)
baseUrl: ${BASE_URL:http://localhost:8080}
databaseUrl: ${DATABASE_URL:postgresql://user:pass@localhost:5432/db}
```

**Rules for ENV resolver:**
- Format: `${KEY}` or `${KEY:default}`
- First `:` separates key from default value
- Everything after `:` is the default (can contain more `:`)
- If environment variable is not set, uses default value
- If no default provided and env not set, empty string is used

### Custom Resolver (@ Prefix)

Use `@` prefix to explicitly specify a custom resolver (AWS, VAULT, etc.):

```yaml
# Custom resolver without default
apiKey: ${@AWS:api-key-secret}

# Custom resolver with default
apiKey: ${@AWS:api-key-secret:fallback-key}

# Custom resolver with colon in default
serviceUrl: ${@VAULT:service/url:http://localhost:8080}
```

**Rules for custom resolver:**
- Format: `${@RESOLVER:KEY}` or `${@RESOLVER:KEY:default}`
- `@` prefix indicates custom resolver (not ENV)
- First `:` separates resolver from key
- Second `:` separates key from default value
- Everything after second `:` is the default (can contain more `:`)

## Supported Resolvers

### ENV (Default)
Built-in resolver that reads from environment variables using `os.Getenv()`.

```yaml
# These are equivalent (both use ENV resolver)
port: ${PORT:8080}
port: ${@ENV:PORT:8080}
```

### Custom Resolvers
You can register custom resolvers for external secret management:

- **AWS**: AWS Secrets Manager or Parameter Store → [Implementation](../core/config/custom_resolver/aws_secret.go)
- **K8S**: Kubernetes Secrets → [Implementation](../core/config/custom_resolver/k8s_secret.go)
- **VAULT**: HashiCorp Vault
- **CONSUL**: Consul KV Store
- **ETCD**: etcd key-value store
- ... or create your own!

**See**: [Custom Resolver README](../core/config/custom_resolver/README.md) for detailed documentation.

## Examples

### Development Configuration
```yaml
servers:
  - name: api-server
    baseUrl: ${BASE_URL:http://localhost:8080}
    apps:
      - addr: ${ADDR::8080}
        
services:
  - name: postgres
    type: dbpool_pg
    config:
      dsn: ${DATABASE_URL:postgresql://dev:dev@localhost:5432/devdb}
      
  - name: redis
    config:
      url: ${REDIS_URL:redis://localhost:6379/0}
```

### Production with AWS Secrets Manager
```yaml
servers:
  - name: api-server
    baseUrl: ${@AWS:api-base-url}
    apps:
      - addr: ${ADDR::8080}
        
services:
  - name: postgres
    type: dbpool_pg
    config:
      dsn: ${@AWS:database-connection-string}
      
  - name: redis
    config:
      url: ${@AWS:redis-url}
      token: ${@AWS:redis-auth-token}
```

### Production with Kubernetes Secrets
```yaml
servers:
  - name: api-server
    baseUrl: ${@K8S:app-config/base-url}
    apps:
      - addr: ${ADDR::8080}

services:
  - name: postgres
    type: dbpool_pg
    config:
      host: ${@K8S:database-credentials/host}
      port: ${@K8S:database-credentials/port}
      user: ${@K8S:database-credentials/username}
      password: ${@K8S:database-credentials/password}
      database: ${@K8S:database-credentials/database}
      
  - name: redis
    config:
      url: ${@K8S:redis-auth/url}
      password: ${@K8S:redis-auth/password}
```

### Hybrid Configuration
```yaml
servers:
  - name: api-server
    # Use env var for local, AWS for production
    baseUrl: ${BASE_URL:http://localhost:8080}
    apps:
      - addr: ${ADDR::8080}
        
services:
  - name: postgres
    type: dbpool_pg
    config:
      # Sensitive data from AWS, connection details from env
      host: ${DB_HOST:localhost}
      port: ${DB_PORT:5432}
      database: ${DB_NAME:mydb}
      user: ${DB_USER:postgres}
      password: ${@AWS:db-password:dev-password}
```

## Implementation Details

### How It Works

1. **Parse YAML**: Configuration file is loaded as text
2. **Expand Variables**: All `${...}` patterns are replaced with actual values
3. **Validate**: Expanded YAML is validated against JSON Schema
4. **Load**: Final configuration is parsed into Go structs

### Parsing Logic

```go
// Format: ${@RESOLVER:KEY:DEFAULT} or ${KEY:DEFAULT}
//
// With @ prefix (custom resolver):
//   ${@AWS:secret-key:fallback} → source="AWS", key="secret-key", default="fallback"
//
// Without @ prefix (ENV default):
//   ${PORT:8080} → source="ENV", key="PORT", default="8080"
//   ${URL:http://localhost:8080} → source="ENV", key="URL", default="http://localhost:8080"
```

### Resolver Interface

Custom resolvers must implement:

```go
type VariableResolver interface {
    Resolve(source string, key string, defaultValue string) (string, bool)
}
```

Register custom resolver:

```go
config.AddVariableResolver("AWS", &AWSSecretsResolver{
    region: "us-east-1",
})
```

## Best Practices

### 1. Always Provide Defaults for Development
```yaml
# ✅ Good - works in dev without env vars
baseUrl: ${BASE_URL:http://localhost:8080}

# ❌ Bad - requires env var to be set
baseUrl: ${BASE_URL}
```

### 2. Use Custom Resolvers for Production Secrets
```yaml
# ✅ Good - explicit and secure
apiKey: ${@AWS:api-secret}

# ⚠️ Less secure - secrets in env vars
apiKey: ${API_KEY}
```

### 3. Complex Defaults with Colons Work Fine
```yaml
# ✅ All valid - colons in defaults are supported
databaseUrl: ${DSN:postgresql://user:pass@localhost:5432/db?sslmode=disable}
redisUrl: ${REDIS:redis://:password@localhost:6379/0}
timeFormat: ${TIME_FORMAT:15:04:05}
```

### 4. Mix and Match Resolvers
```yaml
services:
  - name: api
    config:
      # Public config from env (with defaults)
      baseUrl: ${API_URL:http://localhost:8080}
      timeout: ${TIMEOUT:30s}
      
      # Secrets from AWS
      apiKey: ${@AWS:api-key}
      privateKey: ${@AWS:rsa-private-key}
      
      # Vault for database credentials
      dbPassword: ${@VAULT:database/password}
```

## Why @ Prefix?

The `@` prefix solves ambiguity problems:

### Before (Ambiguous)
```yaml
# 🤔 Is this "RESOLVER:KEY" or "KEY:DEFAULT"?
value: ${ENV:API_KEY:default}
```

### After (Clear)
```yaml
# ✅ This is ENV resolver with key "API_KEY" and default "default"
value: ${@ENV:API_KEY:default}

# ✅ This is key "API_KEY" with default "default" (ENV is default resolver)
value: ${API_KEY:default}

# ✅ This is AWS resolver with key "api-key" and default "fallback"
value: ${@AWS:api-key:fallback}
```

## Migration Guide

If you have existing configs, **no changes needed**! The format is backward compatible:

```yaml
# Old format still works (ENV resolver is default)
port: ${PORT:8080}
baseUrl: ${BASE_URL:http://localhost:8080}

# Explicit ENV resolver (equivalent to above)
port: ${@ENV:PORT:8080}
baseUrl: ${@ENV:BASE_URL:http://localhost:8080}
```

Only use `@` prefix when you need custom resolvers like AWS or Vault.

## Testing

Environment variable expansion is thoroughly tested:

```bash
cd core/config
go test -v -run TestExpandVariables
```

Test coverage includes:
- ✅ Simple variables with/without defaults
- ✅ Defaults containing colons (URLs, DSNs)
- ✅ Multiple variables in strings
- ✅ Custom resolvers
- ✅ Real-world examples (databases, Redis, APIs)

## See Also

- [Configuration Loading](../cmd/learning/02-architecture/02-config-loading/README.md)
- [YAML Configuration System](./yaml-configuration-system.md)
- [JSON Schema Validation](./json-schema-implementation.md)
