# Provider Registry - Custom Config Value Resolvers

## Overview

Lokstra supports custom providers for resolving configuration values from various sources beyond environment variables. Providers follow a simple interface and can be registered to handle specific prefixes like `${@aws-secret:key}`, `${@vault:path}`, `${@k8s:configmap/key}`, etc.

## How It Works

### 2-Step Resolution Process

Config values are resolved at **YAML byte level** BEFORE unmarshaling:

**STEP 1**: Resolve all `${...}` EXCEPT `${@cfg:...}`
- Resolves: `${ENV_VAR}`, `${@env:VAR}`, `${@aws-secret:key}`, `${@vault:path}`, etc.
- Uses provider registry to handle custom providers
- Default provider: `@env` (environment variables)

**STEP 2**: Resolve `${@cfg:...}` using configs from step 1
- Resolves: `${@cfg:DB_HOST}`, `${@cfg:API_KEY}`, etc.
- Uses the `configs:` section from YAML (already resolved in step 1)
- Special provider: `@cfg` (references to other config values)

### Benefits

✅ **Efficient**: Resolve once at YAML level, not per-field after unmarshal  
✅ **Automatic**: All fields (including nested) are resolved automatically  
✅ **Extensible**: Easy to add custom providers (AWS, Vault, K8s, etc.)  
✅ **Maintainable**: New YAML fields auto-resolved without code changes  

## Provider Interface

```go
// Provider interface for custom value providers
type Provider interface {
    // Name returns the provider name (e.g., "env", "aws-secret", "vault")
    Name() string

    // Resolve resolves a key to its value
    // Returns the resolved value and whether it was found
    Resolve(key string) (string, bool)
}
```

## Syntax and Quote Escaping

### Basic Syntax

```yaml
# Without provider prefix (uses @env by default)
${VAR_NAME}                 # Environment variable
${VAR_NAME:default}         # With default value

# With provider prefix
${@provider:key}            # Custom provider
${@provider:key:default}    # Custom provider with default
```

### Quote Escaping for Keys with Colons

**Problem:** Keys containing `:` characters are ambiguous without quotes.

```yaml
# AMBIGUOUS - is this key or key:default?
password: ${@vault:secret/data/db:password}
# Interpreted as: key="secret/data/db", default="password" ❌

# CLEAR - single quotes preserve the full key
password: ${@vault:'secret/data/db:password'}  
# Interpreted as: key="secret/data/db:password", default="" ✅
```

**Why Single Quotes (`'`)?**
- **Avoids YAML syntax conflict** - Double quotes (`"`) are YAML string delimiters
- **Simpler to write** - No escaping needed in YAML
- **Clearer intent** - Explicitly marks quoted keys

**Rules:**
1. **Without quotes**: FIRST `:` after provider name is the key/default separator
2. **With single quotes**: `:` inside quotes is part of the key, not a separator

### Examples

**Environment Variables (no colons in key):**
```yaml
database:
  host: ${DB_HOST}                    # key="DB_HOST"
  host: ${DB_HOST:localhost}          # key="DB_HOST", default="localhost"
  port: ${DB_PORT:5432}               # key="DB_PORT", default="5432"
```

**URLs as Default Values:**
```yaml
database:
  # Key has no colons, URL is default value
  url: ${DB_URL:postgresql://localhost:5432/db}
  # Interpreted as: key="DB_URL", default="postgresql://localhost:5432/db" ✅
```

**AWS ARN (contains colons - needs quotes):**
```yaml
secrets:
  # WITHOUT quotes - WRONG interpretation
  password: ${@aws-secret:arn:aws:secretsmanager:us-east-1:123:secret:db}
  # Interpreted as: key="arn", default="aws:secretsmanager:us-east-1:123:secret:db" ❌
  
  # WITH single quotes - CORRECT interpretation  
  password: ${@aws-secret:'arn:aws:secretsmanager:us-east-1:123:secret:db'}
  # Interpreted as: key="arn:aws:secretsmanager:us-east-1:123:secret:db", default="" ✅
  
  # WITH quotes AND default
  password: ${@aws-secret:'arn:aws:secretsmanager:us-east-1:123:secret:db':fallback}
  # Interpreted as: key="arn:aws:secretsmanager:us-east-1:123:secret:db", default="fallback" ✅
```

**Vault Paths:**
```yaml
secrets:
  # Simple path (no colons)
  api-key: ${@vault:secret/data/myapp/api-key}
  # Interpreted as: key="secret/data/myapp/api-key" ✅
  
  # Path with colons - WITHOUT quotes (ambiguous)
  password: ${@vault:secret/data/db:password}
  # Interpreted as: key="secret/data/db", default="password" ❌
  
  # Path with colons - WITH quotes (clear)
  password: ${@vault:'secret/data/db:password'}
  # Interpreted as: key="secret/data/db:password", default="" ✅
  
  # Path with colons AND default value
  password: ${@vault:'secret/data/db:password':fallback}
  # Interpreted as: key="secret/data/db:password", default="fallback" ✅
```

**Config References:**
```yaml
configs:
  database:
    host: "prod-db.example.com"
  db:url: "postgresql://localhost:5432/mydb"  # Key with colon

named-db-pools:
  main:
    # Simple config reference (no colons in key)
    host: ${@cfg:database.host}
    # Interpreted as: key="database.host" ✅
    
    # Config key contains colon - WITHOUT quotes (ambiguous)
    url: ${@cfg:db:url}
    # Interpreted as: key="db", default="url" ❌
    
    # Config key contains colon - WITH quotes (clear)
    url: ${@cfg:'db:url'}
    # Interpreted as: key="db:url" ✅
    
    # With default value
    url: ${@cfg:'db:url':postgresql://localhost:5432/fallback}
    # Interpreted as: key="db:url", default="postgresql://localhost:5432/fallback" ✅
```

### Best Practices

1. **Use single quotes when keys contain `:` characters**
   ```yaml
   # Good - single quotes for ARNs
   arn: ${@aws-secret:'arn:aws:secretsmanager:region:account:secret:name'}
   
   # Bad (will be parsed incorrectly)
   arn: ${@aws-secret:arn:aws:secretsmanager:region:account:secret:name}
   ```

2. **URLs as default values don't need quotes**
   ```yaml
   # Good - key has no colons
   db-url: ${DB_URL:postgresql://localhost:5432/db}
   ```

3. **Document your config key naming convention**
   ```yaml
   # Avoid colons in config keys if possible
   configs:
     database-url: "..."      # Good - use hyphens
     database_url: "..."      # Good - use underscores
     "database:url": "..."    # Works but requires quotes in references
   ```

4. **Always provide defaults for optional configs**
   ```yaml
   port: ${DB_PORT:5432}
   host: ${@cfg:db.host:localhost}
   ```

5. **No need to wrap YAML values in double quotes**
   ```yaml
   # Simple and clean
   password: ${@vault:'secret/data/db:password'}
   
   # Also works but unnecessary
   password: "${@vault:'secret/data/db:password'}"
   ```

### `@env` Provider (Default)

Resolves from environment variables and command-line flags.

**Syntax:**
```yaml
# Without prefix (uses @env by default)
database:
  host: ${DB_HOST}                    # → @env:DB_HOST
  port: ${DB_PORT:5432}               # → @env:DB_PORT with default

# Explicit @env prefix
database:
  host: ${@env:DB_HOST}
  port: ${@env:DB_PORT:5432}
```

**Resolution Order:**
1. Environment variable: `os.Getenv("DB_HOST")`
2. Command-line flag: `-DB_HOST=localhost` or `--DB_HOST=localhost`
3. Default value: `:5432` (if provided)

### `@cfg` Provider (Special)

References other values from `configs:` section. Resolved in STEP 2.

**Syntax:**
```yaml
configs:
  db:
    host: "prod-db.example.com"
    port: 5432
  
named-db-pools:
  main:
    host: ${@cfg:db.host}             # → "prod-db.example.com"
    port: ${@cfg:db.port}             # → 5432
```

**Features:**
- Case-insensitive lookup
- Supports nested config keys with dot notation
- Type preservation (integers stay integers)

## Custom Provider Examples

### Example 1: AWS Secrets Manager

```go
package awsprovider

import (
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/secretsmanager"
    "github.com/primadi/lokstra/core/deploy/loader"
)

type AWSSecretProvider struct {
    client *secretsmanager.SecretsManager
}

func NewAWSSecretProvider() *AWSSecretProvider {
    sess := session.Must(session.NewSession())
    return &AWSSecretProvider{
        client: secretsmanager.New(sess),
    }
}

func (p *AWSSecretProvider) Name() string {
    return "aws-secret"
}

func (p *AWSSecretProvider) Resolve(key string) (string, bool) {
    input := &secretsmanager.GetSecretValueInput{
        SecretId: &key,
    }
    
    result, err := p.client.GetSecretValue(input)
    if err != nil {
        return "", false
    }
    
    if result.SecretString != nil {
        return *result.SecretString, true
    }
    
    return "", false
}

// Register in main.go
func init() {
    loader.RegisterProvider(NewAWSSecretProvider())
}
```

**Usage in YAML:**
```yaml
database:
  # Simple path (no colons in key)
  password: ${@aws-secret:prod/db/password}
  
  # ARN with colons - needs single quotes
  password: ${@aws-secret:'arn:aws:secretsmanager:us-east-1:123456789012:secret:prod/db/password'}
  
configs:
  api:
    key: ${@aws-secret:prod/api/key}
```

### Example 2: HashiCorp Vault

```go
package vaultprovider

import (
    "github.com/hashicorp/vault/api"
    "github.com/primadi/lokstra/core/deploy/loader"
)

type VaultProvider struct {
    client *api.Client
}

func NewVaultProvider(addr, token string) *VaultProvider {
    config := api.DefaultConfig()
    config.Address = addr
    
    client, _ := api.NewClient(config)
    client.SetToken(token)
    
    return &VaultProvider{client: client}
}

func (p *VaultProvider) Name() string {
    return "vault"
}

func (p *VaultProvider) Resolve(key string) (string, bool) {
    secret, err := p.client.Logical().Read(key)
    if err != nil || secret == nil {
        return "", false
    }
    
    // Assume key format: "secret/data/myapp"
    // Returns secret.Data["data"]["value"]
    if data, ok := secret.Data["data"].(map[string]interface{}); ok {
        if value, ok := data["value"].(string); ok {
            return value, true
        }
    }
    
    return "", false
}

// Register in main.go
func init() {
    vaultAddr := os.Getenv("VAULT_ADDR")
    vaultToken := os.Getenv("VAULT_TOKEN")
    loader.RegisterProvider(NewVaultProvider(vaultAddr, vaultToken))
}
```

**Usage in YAML:**
```yaml
database:
  # Simple Vault path (no colons in path)
  password: ${@vault:secret/data/db/password}
  
configs:
  jwt:
    secret: ${@vault:secret/data/jwt/signing-key}
    
  # If Vault path contains colons - use single quotes
  legacy:
    key: ${@vault:'secret/data/legacy:service:key'}
```

### Example 3: Kubernetes ConfigMap/Secret

```go
package k8sprovider

import (
    "context"
    "strings"
    
    "github.com/primadi/lokstra/core/deploy/loader"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type K8sProvider struct {
    clientset *kubernetes.Clientset
    namespace string
}

func NewK8sProvider(namespace string) *K8sProvider {
    config, _ := rest.InClusterConfig()
    clientset, _ := kubernetes.NewForConfig(config)
    
    return &K8sProvider{
        clientset: clientset,
        namespace: namespace,
    }
}

func (p *K8sProvider) Name() string {
    return "k8s"
}

func (p *K8sProvider) Resolve(key string) (string, bool) {
    // Key format: "configmap/myconfig/key" or "secret/mysecret/key"
    parts := strings.SplitN(key, "/", 3)
    if len(parts) != 3 {
        return "", false
    }
    
    resourceType := parts[0] // "configmap" or "secret"
    name := parts[1]
    dataKey := parts[2]
    
    ctx := context.Background()
    
    switch resourceType {
    case "configmap":
        cm, err := p.clientset.CoreV1().ConfigMaps(p.namespace).Get(ctx, name, metav1.GetOptions{})
        if err != nil {
            return "", false
        }
        if value, ok := cm.Data[dataKey]; ok {
            return value, true
        }
        
    case "secret":
        secret, err := p.clientset.CoreV1().Secrets(p.namespace).Get(ctx, name, metav1.GetOptions{})
        if err != nil {
            return "", false
        }
        if value, ok := secret.Data[dataKey]; ok {
            return string(value), true
        }
    }
    
    return "", false
}

// Register in main.go
func init() {
    loader.RegisterProvider(NewK8sProvider("default"))
}
```

**Usage in YAML:**
```yaml
database:
  host: ${@k8s:configmap/app-config/db-host}
  password: ${@k8s:secret/db-credentials/password}
  
  # If ConfigMap/Secret key contains colon - use single quotes
  url: ${@k8s:'configmap/app-config/db:url'}
```

## Registration

Register custom providers in your `main.go` BEFORE calling `lokstra.Bootstrap()`:

```go
package main

import (
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/core/deploy/loader"
    "myapp/providers/awsprovider"
    "myapp/providers/vaultprovider"
)

func main() {
    // Register custom providers BEFORE Bootstrap
    loader.RegisterProvider(awsprovider.NewAWSSecretProvider())
    loader.RegisterProvider(vaultprovider.NewVaultProvider(
        os.Getenv("VAULT_ADDR"),
        os.Getenv("VAULT_TOKEN"),
    ))
    
    // Now bootstrap (loads config with custom providers)
    lokstra.Bootstrap()
    
    // ... rest of main
}
```

## Complete Example

**config.yaml:**
```yaml
configs:
  db:
    host: ${DB_HOST:localhost}        # @env provider
    port: 5432
  aws:
    region: ${AWS_REGION:us-east-1}   # @env provider

named-db-pools:
  main:
    # Mix of providers
    host: ${@cfg:db.host}                           # @cfg provider
    port: ${@cfg:db.port}                             # @cfg provider
    database: ${DB_NAME}                             # @env provider (default)
    username: ${@aws-secret:prod/db/username}        # @aws-secret provider
    password: ${@aws-secret:prod/db/password}        # @aws-secret provider
    min-conns: 2
    max-conns: 10

service-definitions:
  api-service:
    type: api-service-factory
    config:
      jwt-secret: ${@vault:secret/data/jwt/key}     # @vault provider
      aws-region: ${@cfg:aws.region}                 # @cfg provider
```

**Environment Variables:**
```bash
export DB_HOST=prod-db.example.com
export DB_NAME=myapp
export AWS_REGION=ap-southeast-1
export VAULT_ADDR=https://vault.example.com
export VAULT_TOKEN=s.xxxxxxxxx
```

**Resolution Flow:**

1. **STEP 1** - Resolve non-@cfg:
   - `${DB_HOST:localhost}` → `"prod-db.example.com"` (@env)
   - `${AWS_REGION:us-east-1}` → `"ap-southeast-1"` (@env)
   - `${DB_NAME}` → `"myapp"` (@env)
   - `${@aws-secret:prod/db/username}` → `"admin"` (@aws-secret)
   - `${@aws-secret:prod/db/password}` → `"secret123"` (@aws-secret)
   - `${@vault:secret/data/jwt/key}` → `"jwt-signing-key-xyz"` (@vault)

2. **STEP 2** - Resolve @cfg:
   - `${@cfg:db.host}` → `"prod-db.example.com"` (from configs)
   - `${@cfg:db.port}` → `5432` (from configs)
   - `${@cfg:aws.region}` → `"ap-southeast-1"` (from configs)

## Best Practices

1. **Register providers early**: Register in `main.go` before `lokstra.Bootstrap()`
2. **Handle errors gracefully**: Return `("", false)` if resolution fails
3. **Cache expensive lookups**: Implement caching in your provider if needed
4. **Use default values**: Always provide defaults for optional configs: `${KEY:default}`
5. **Document key formats**: Clearly document expected key format for your provider

## Provider Priority

When a key is not prefixed with `@provider:`, the default `@env` provider is used:

```yaml
# These are equivalent:
host: ${DB_HOST}
host: ${@env:DB_HOST}
```

## Migration from Old ResolveConfigs

The old `ResolveConfigs()` method is **now deprecated** because:

- ❌ Resolved per-field after unmarshal (inefficient)
- ❌ Required manual iteration for each field type
- ❌ New fields required code updates

The new YAML-level resolution:

- ✅ Resolves at byte level before unmarshal (efficient)
- ✅ Automatic for all fields (including nested)
- ✅ New fields auto-resolved without code changes
- ✅ Extensible provider registry

## Troubleshooting

**Provider not found:**
```yaml
database:
  password: ${@unknown:key}
```
→ Returns `"${@unknown:key}"` unchanged (useful for debugging)

**Config key not found:**
```yaml
service:
  url: "${@cfg:nonexistent}"
```
→ Returns `"${@cfg:nonexistent}"` unchanged (useful for debugging)

**Use defaults to prevent errors:**
```yaml
database:
  password: "${@aws-secret:prod/db/password:fallback-password}"
```
→ Uses `"fallback-password"` if AWS Secrets Manager fails
