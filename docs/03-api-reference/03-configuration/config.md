# Configuration

> Configuration management and YAML loading system

## Overview

The `config` package provides configuration management for Lokstra applications with support for YAML files, variable expansion, schema validation, and multi-file merging. It supports both simple flat configurations and layered service definitions.

## Import Path

```go
import "github.com/primadi/lokstra/core/config"
```

---

## Core Types

### Config
Top-level configuration structure.

**Definition:**
```go
type Config struct {
    Configs     []*GeneralConfig
    Services    ServicesConfig
    Middlewares []*Middleware
    Routers     []*Router
    Servers     []*Server
}
```

**Fields:**
- `Configs` - General key-value configurations
- `Services` - Service definitions (simple or layered)
- `Middlewares` - Middleware definitions
- `Routers` - Router configurations
- `Servers` - Server and app definitions

---

### ServicesConfig
Flexible service configuration supporting both flat arrays and layered maps.

**Definition:**
```go
type ServicesConfig struct {
    Simple  []*Service            // Flat array of services
    Layered map[string][]*Service // Services grouped by layer
    Order   []string              // Layer order (for layered mode)
}
```

**Methods:**
```go
func (sc *ServicesConfig) IsSimple() bool
func (sc *ServicesConfig) IsLayered() bool
func (sc *ServicesConfig) GetAllServices() []*Service
func (sc *ServicesConfig) Flatten() []*Service
```

**Simple Mode (Array):**
```yaml
services:
  - name: db-service
    type: postgres
  - name: user-service
    type: user-service-factory
```

**Layered Mode (Map):**
```yaml
services:
  infrastructure:
    - name: db-service
      type: postgres
    - name: cache-service
      type: redis
      
  business:
    - name: user-service
      type: user-service-factory
    - name: order-service
      type: order-service-factory
```

---

### GeneralConfig
Key-value configuration pairs.

**Definition:**
```go
type GeneralConfig struct {
    Name  string // Configuration key
    Value any    // Configuration value (string, number, bool, object, etc.)
}
```

**Example:**
```yaml
configs:
  - name: db.dsn
    value: "postgresql://localhost/mydb"
  - name: app.max_connections
    value: 100
  - name: app.features
    value:
      enable_logging: true
      enable_metrics: false
```

---

### Service
Service definition configuration.

**Definition:**
```go
type Service struct {
    Name       string
    Type       string
    Enable     *bool          // Default: true
    DependsOn  []string
    Config     map[string]any
    AutoRouter *AutoRouter
}
```

**Methods:**
```go
func (s *Service) IsEnabled() bool
func (s *Service) GetConvention(globalDefault string) string
func (s *Service) GetPathPrefix() string
func (s *Service) GetResourceName() string
func (s *Service) GetPluralResourceName() string
func (s *Service) GetRouteOverrides() []*RouteOverride
```

**Example:**
```yaml
services:
  - name: user-service
    type: user-service-factory
    enable: true
    depends-on:
      - db-service
      - cache-service
    config:
      max_items: 100
      timeout: 30s
    auto-router:
      convention: rest
      path-prefix: /api/v1
      resource-name: user
      plural-resource-name: users
```

---

### AutoRouter
Auto-router configuration for service-based routing.

**Definition:**
```go
type AutoRouter struct {
    Convention         string
    PathPrefix         string
    ResourceName       string
    PluralResourceName string
    Routes             []*RouteOverride
}
```

**Example:**
```yaml
auto-router:
  convention: rest
  path-prefix: /api/v1
  resource-name: user
  plural-resource-name: users
  routes:
    - name: Login
      method: POST
      path: /auth/login
    - name: Logout
      method: POST
      path: /auth/logout
```

---

### RouteOverride
Overrides for specific service routes.

**Definition:**
```go
type RouteOverride struct {
    Name   string // Function/method name
    Method string // HTTP method override
    Path   string // Path override
}
```

---

### Middleware
Middleware definition configuration.

**Definition:**
```go
type Middleware struct {
    Name   string
    Type   string
    Enable *bool          // Default: true
    Config map[string]any
}
```

**Methods:**
```go
func (m *Middleware) IsEnabled() bool
```

**Example:**
```yaml
middlewares:
  - name: logger-debug
    type: logger
    enable: true
    config:
      level: DEBUG
      colorize: true
      
  - name: cors-dev
    type: cors
    config:
      allow_origin: "*"
      allow_methods: "*"
```

---

### Router
Router configuration.

**Definition:**
```go
type Router struct {
    Name        string
    PathPrefix  string
    Middlewares []string
}
```

**Example:**
```yaml
routers:
  - name: user-router
    path-prefix: /api/v1
    middlewares:
      - auth
      - logger
      
  - name: public-router
    path-prefix: /public
    middlewares:
      - cors
```

---

### Server
Server configuration with multiple apps.

**Definition:**
```go
type Server struct {
    Name         string
    BaseUrl      string
    DeploymentID string
    Apps         []*App
}
```

**Methods:**
```go
func (s *Server) GetBaseUrl() string      // Default: "http://localhost"
func (s *Server) GetDeploymentID() string
```

**Example:**
```yaml
servers:
  - name: api-server
    base-url: http://localhost:8080
    deployment-id: production
    apps:
      - name: rest-api
        addr: ":8080"
        services:
          - user-service
          - order-service
        routers:
          - user-router
          - order-router
```

---

### App
Application configuration within a server.

**Definition:**
```go
type App struct {
    Name           string
    Addr           string
    ListenerType   string // Default: "default"
    Services       []string
    Routers        []string
    ReverseProxies []*ReverseProxyConfig
}
```

**Methods:**
```go
func (a *App) GetListenerType() string      // Default: "default"
func (a *App) GetName(index int) string     // Auto-generates if empty
```

**Example:**
```yaml
apps:
  - name: api
    addr: ":8080"
    listener-type: default
    services:
      - user-service
      - order-service
    routers:
      - user-router
      - order-router
    reverse-proxies:
      - prefix: /external
        target: http://external-api:9000
        strip-prefix: true
```

---

### ReverseProxyConfig
Reverse proxy configuration for proxying requests.

**Definition:**
```go
type ReverseProxyConfig struct {
    Prefix      string
    StripPrefix bool
    Target      string
    Rewrite     *ReverseProxyRewrite
}
```

**Example:**
```yaml
reverse-proxies:
  - prefix: /api
    strip-prefix: true
    target: http://backend-api:8080
    rewrite:
      from: ^/api/v1/(.*)
      to: /v2/$1
```

---

### ReverseProxyRewrite
Path rewrite rules for reverse proxy.

**Definition:**
```go
type ReverseProxyRewrite struct {
    From string // Pattern to match (regex supported)
    To   string // Replacement pattern
}
```

---

## Loading Functions

### LoadConfigFile
Loads a single YAML configuration file from OS filesystem.

**Signature:**
```go
func LoadConfigFile(fileName string, config *Config) error
```

**Parameters:**
- `fileName` - Path to YAML file (absolute or relative)
- `config` - Target config structure to merge into

**Returns:**
- `error` - Error if file not found, invalid YAML, or validation fails

**Example:**
```go
cfg := config.New()
err := config.LoadConfigFile("config/app.yaml", cfg)
if err != nil {
    log.Fatal(err)
}
```

**Features:**
- ✅ Variable expansion (${ENV_VAR}, ${@cfg:key})
- ✅ JSON schema validation
- ✅ Merges with existing config
- ✅ Two-pass expansion for ${@cfg:...}

---

### LoadConfigFs
Loads a single YAML configuration file from any filesystem.

**Signature:**
```go
func LoadConfigFs(fsys fs.FS, fileName string, config *Config) error
```

**Parameters:**
- `fsys` - Filesystem interface (os.DirFS, embed.FS, etc.)
- `fileName` - Path to YAML file within filesystem
- `config` - Target config structure

**Example:**
```go
import "embed"

//go:embed config/*.yaml
var configFS embed.FS

cfg := config.New()
err := config.LoadConfigFs(configFS, "config/app.yaml", cfg)
if err != nil {
    log.Fatal(err)
}
```

**Use Cases:**
- Embedded configurations
- Testing with virtual filesystems
- Custom filesystem implementations

---

### LoadConfigDir
Loads and merges multiple YAML files from a directory.

**Signature:**
```go
func LoadConfigDir(dirName string, config *Config) error
```

**Parameters:**
- `dirName` - Directory path containing YAML files
- `config` - Target config structure

**Returns:**
- `error` - Error if directory not found or any file fails validation

**Example:**
```go
cfg := config.New()
err := config.LoadConfigDir("config/", cfg)
if err != nil {
    log.Fatal(err)
}
```

**Behavior:**
- Loads all `.yaml` and `.yml` files in directory
- Files loaded in alphabetical order
- Configurations merged sequentially
- Variable expansion applied per file

**Directory Structure:**
```
config/
  ├── 01-base.yaml       # Loaded first
  ├── 02-services.yaml   # Merged second
  ├── 03-servers.yaml    # Merged third
  └── 04-overrides.yaml  # Merged last
```

---

### LoadConfigDirFs
Loads and merges multiple YAML files from a filesystem directory.

**Signature:**
```go
func LoadConfigDirFs(fsys fs.FS, dirName string, config *Config) error
```

**Example:**
```go
cfg := config.New()
err := config.LoadConfigDirFs(os.DirFS("."), "config/", cfg)
if err != nil {
    log.Fatal(err)
}
```

---

## Variable Expansion System

Lokstra provides an **extensible variable expansion system** powered by **Variable Resolvers**. This allows you to:
- Pull configuration from multiple sources (ENV, AWS Secrets, Vault, Kubernetes ConfigMaps, etc.)
- Create custom resolvers for your own config sources
- Reference other config values to avoid duplication
- Use two-pass expansion for complex scenarios

### Architecture

**Variable Resolver Interface:**
```go
type VariableResolver interface {
    Resolve(source string, key string, defaultValue string) (string, bool)
}
```

**Built-in Resolvers:**
- `ENV` - Environment variables (default)
- `CFG` - Config registry references (two-pass expansion)

**Custom Resolvers:**
You can add your own resolvers for AWS Secrets Manager, HashiCorp Vault, Kubernetes ConfigMaps, etc.

---

## Built-in Resolvers

### ENV Resolver (Environment Variables)
Reference environment variables in YAML files.

**Syntax:**
```yaml
${ENV_VAR_NAME}
${ENV_VAR_NAME:default_value}
${@ENV:VAR_NAME}
${@ENV:VAR_NAME:default_value}
```

**Example:**
```yaml
configs:
  - name: db.dsn
    value: "${DATABASE_URL:postgresql://localhost/dev}"
  - name: app.port
    value: "${PORT:8080}"
  - name: app.env
    value: "${APP_ENV:development}"
```

**Shell:**
```bash
export DATABASE_URL="postgresql://prod-server/prod_db"
export PORT="3000"
# APP_ENV not set, uses default "development"
```

**Result:**
```yaml
configs:
  - name: db.dsn
    value: "postgresql://prod-server/prod_db"
  - name: app.port
    value: "3000"
  - name: app.env
    value: "development"
```

---

### CFG Resolver (Config References)
Reference other config values using two-pass expansion.

**Syntax:**
```yaml
${@CFG:config.key}
${@CFG:config.key:default_value}
```

**Note:** CFG resolver uses uppercase `CFG`, not lowercase `cfg`.

**Example:**
```yaml
configs:
  - name: db.host
    value: "localhost"
  - name: db.port
    value: "5432"
  - name: db.name
    value: "myapp"
  - name: db.dsn
    value: "postgresql://${@CFG:db.host}:${@CFG:db.port}/${@CFG:db.name}"
```

**Result:**
```yaml
configs:
  - name: db.dsn
    value: "postgresql://localhost:5432/myapp"
```

**How it works:**
1. **Pass 1:** All non-CFG resolvers (ENV, AWS, etc.) are expanded
2. **Pass 2:** `configs` section is parsed and stored in temporary registry
3. **Pass 3:** CFG resolver expands using temporary config values

This allows `${@CFG:...}` to work even before the full config registry is built.

---

## Custom Resolvers

### Adding Custom Resolvers

You can create custom resolvers for any configuration source:

**1. Implement the `VariableResolver` interface:**

```go
import "github.com/primadi/lokstra/core/config"

type AWSSecretsResolver struct {
    client *secretsmanager.Client
}

func (r *AWSSecretsResolver) Resolve(source string, key string, defaultValue string) (string, bool) {
    if source != "AWS" {
        return "", false
    }
    
    // Fetch from AWS Secrets Manager
    result, err := r.client.GetSecretValue(context.Background(), &secretsmanager.GetSecretValueInput{
        SecretId: aws.String(key),
    })
    
    if err != nil {
        return defaultValue, false
    }
    
    return *result.SecretString, true
}
```

**2. Register the resolver:**

```go
func init() {
    // Create AWS Secrets Manager client
    cfg, err := awsconfig.LoadDefaultConfig(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    
    client := secretsmanager.NewFromConfig(cfg)
    
    // Register resolver
    config.AddVariableResolver("AWS", &AWSSecretsResolver{
        client: client,
    })
}
```

**3. Use in YAML:**

```yaml
configs:
  - name: db.password
    value: "${@AWS:prod/db/password}"
  
  - name: api.key
    value: "${@AWS:prod/api/key:fallback-key}"
  
  - name: jwt.secret
    value: "${@AWS:prod/jwt/secret}"
```

---

### Example: HashiCorp Vault Resolver

```go
import (
    "github.com/hashicorp/vault/api"
    "github.com/primadi/lokstra/core/config"
)

type VaultResolver struct {
    client *api.Client
}

func NewVaultResolver(addr, token string) (*VaultResolver, error) {
    cfg := api.DefaultConfig()
    cfg.Address = addr
    
    client, err := api.NewClient(cfg)
    if err != nil {
        return nil, err
    }
    
    client.SetToken(token)
    
    return &VaultResolver{client: client}, nil
}

func (r *VaultResolver) Resolve(source string, key string, defaultValue string) (string, bool) {
    if source != "VAULT" {
        return "", false
    }
    
    // Read from Vault
    secret, err := r.client.Logical().Read(key)
    if err != nil || secret == nil {
        return defaultValue, false
    }
    
    // Get "value" field from secret data
    if value, ok := secret.Data["value"].(string); ok {
        return value, true
    }
    
    return defaultValue, false
}

// Register in init()
func init() {
    vaultAddr := os.Getenv("VAULT_ADDR")
    vaultToken := os.Getenv("VAULT_TOKEN")
    
    if vaultAddr != "" && vaultToken != "" {
        resolver, err := NewVaultResolver(vaultAddr, vaultToken)
        if err != nil {
            log.Printf("Failed to create Vault resolver: %v", err)
            return
        }
        
        config.AddVariableResolver("VAULT", resolver)
    }
}
```

**Usage:**
```yaml
configs:
  - name: db.password
    value: "${@VAULT:secret/data/db/password}"
  
  - name: api.key
    value: "${@VAULT:secret/data/api/key:default-key}"
```

---

### Example: Kubernetes ConfigMap Resolver

```go
import (
    "context"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
)

type K8sConfigMapResolver struct {
    client    *kubernetes.Clientset
    namespace string
}

func NewK8sConfigMapResolver() (*K8sConfigMapResolver, error) {
    // Create in-cluster config
    config, err := rest.InClusterConfig()
    if err != nil {
        return nil, err
    }
    
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, err
    }
    
    namespace := os.Getenv("POD_NAMESPACE")
    if namespace == "" {
        namespace = "default"
    }
    
    return &K8sConfigMapResolver{
        client:    clientset,
        namespace: namespace,
    }, nil
}

func (r *K8sConfigMapResolver) Resolve(source string, key string, defaultValue string) (string, bool) {
    if source != "K8S" {
        return "", false
    }
    
    // Format: configmap-name/key-name
    parts := strings.SplitN(key, "/", 2)
    if len(parts) != 2 {
        return defaultValue, false
    }
    
    configMapName := parts[0]
    keyName := parts[1]
    
    // Get ConfigMap
    cm, err := r.client.CoreV1().ConfigMaps(r.namespace).Get(
        context.Background(),
        configMapName,
        metav1.GetOptions{},
    )
    
    if err != nil {
        return defaultValue, false
    }
    
    // Get key from ConfigMap
    if value, ok := cm.Data[keyName]; ok {
        return value, true
    }
    
    return defaultValue, false
}

// Register
func init() {
    if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
        resolver, err := NewK8sConfigMapResolver()
        if err != nil {
            log.Printf("Failed to create K8s resolver: %v", err)
            return
        }
        
        config.AddVariableResolver("K8S", resolver)
    }
}
```

**Usage:**
```yaml
configs:
  - name: app.config
    value: "${@K8S:app-config/database-url}"
  
  - name: feature.flag
    value: "${@K8S:feature-flags/new-ui:false}"
```

---

## Resolver Resolution Order

When a variable is expanded, resolvers are applied in this order:

1. **Pass 1: All resolvers EXCEPT CFG**
   - ENV resolver
   - AWS resolver
   - VAULT resolver
   - K8S resolver
   - ... (any custom resolvers)

2. **Pass 2: Parse configs and build temporary registry**

3. **Pass 3: CFG resolver**
   - Expands `${@CFG:...}` using temporary config values

This ensures that CFG references work even when referencing values that contain other resolver placeholders.

---

## Resolver Syntax Reference

| Syntax | Resolver | Example |
|--------|----------|---------|
| `${KEY}` | ENV (default) | `${DATABASE_URL}` |
| `${KEY:default}` | ENV with default | `${PORT:8080}` |
| `${@ENV:KEY}` | Explicit ENV | `${@ENV:API_KEY}` |
| `${@ENV:KEY:default}` | ENV with default | `${@ENV:PORT:8080}` |
| `${@CFG:key}` | Config reference | `${@CFG:db.host}` |
| `${@CFG:key:default}` | CFG with default | `${@CFG:db.port:5432}` |
| `${@AWS:secret}` | AWS Secrets | `${@AWS:prod/db/pass}` |
| `${@VAULT:path}` | Vault secret | `${@VAULT:secret/data/key}` |
| `${@K8S:cm/key}` | K8s ConfigMap | `${@K8S:app-config/url}` |

---

## Advanced Resolver Patterns

### Combined Expansion
Combine environment variables and config references.

**Example:**
```yaml
configs:
  - name: app.env
    value: "${APP_ENV:development}"
  - name: db.host
    value: "${DB_HOST:localhost}"
  - name: log.level
    value: "${LOG_LEVEL:INFO}"
  - name: app.name
    value: "MyApp (${@CFG:app.env})"
  - name: db.connection
    value: "postgresql://${@CFG:db.host}:5432/myapp_${@CFG:app.env}"
```

**Shell:**
```bash
export APP_ENV="production"
export DB_HOST="db.example.com"
```

**Result:**
```yaml
configs:
  - name: app.name
    value: "MyApp (production)"
  - name: db.connection
    value: "postgresql://db.example.com:5432/myapp_production"
```

---

### Multi-Source Resolution

Combine multiple resolvers in one config:

```yaml
configs:
  # From environment
  - name: app.env
    value: "${APP_ENV:development}"
  
  # From AWS Secrets Manager
  - name: db.password
    value: "${@AWS:${@CFG:app.env}/db/password}"
  
  # From Vault
  - name: jwt.secret
    value: "${@VAULT:secret/data/${@CFG:app.env}/jwt}"
  
  # From K8s ConfigMap
  - name: api.endpoint
    value: "${@K8S:app-config/api-endpoint:http://localhost}"
  
  # Composed from multiple sources
  - name: db.dsn
    value: "postgresql://${DB_USER}:${@AWS:prod/db/password}@${@CFG:db.host}:5432/${DB_NAME}"
```

**Expansion flow:**
1. **Pass 1:** `${APP_ENV}` → `production`, `${DB_USER}` → `myuser`, `${DB_NAME}` → `mydb`
2. **Pass 2:** Build config registry with `app.env=production`
3. **Pass 3:** `${@CFG:app.env}` → `production`, `${@AWS:production/db/password}` → `secret123`
4. **Result:** `postgresql://myuser:secret123@localhost:5432/mydb`

---

### Conditional Resolution by Environment

```yaml
configs:
  - name: app.env
    value: "${APP_ENV:development}"
  
  # Development: Use local values
  - name: db.password
    value: "${@CFG:app.env}" # Will check if dev/prod
  
services:
  - name: db-service
    type: postgres
    config:
      # Production: Use AWS Secrets
      # Development: Use environment variable
      password: "${@AWS:${@CFG:app.env}/db/password:${DB_PASSWORD:devpass}}"
```

---

### Fallback Chain

Create a fallback chain across multiple resolvers:

```yaml
configs:
  # Try AWS → Vault → K8s → ENV → default
  - name: api.key
    value: "${@AWS:prod/api/key:${@VAULT:secret/api/key:${@K8S:secrets/api-key:${API_KEY:default-key}}}}"
```

**Resolution order:**
1. Try AWS Secrets Manager: `prod/api/key`
2. If not found, try Vault: `secret/api/key`
3. If not found, try K8s ConfigMap: `secrets/api-key`
4. If not found, try ENV: `API_KEY`
5. If not found, use default: `default-key`

---

## Schema Validation

All loaded configurations are automatically validated against the JSON schema.

**Validation Checks:**
- ✅ Required fields present
- ✅ Correct data types
- ✅ Valid enum values
- ✅ Service dependencies exist
- ✅ No duplicate names
- ✅ Valid middleware/service references

**Validation Error Example:**
```go
err := config.LoadConfigFile("invalid.yaml", cfg)
if err != nil {
    // Error: validation failed for invalid.yaml:
    // - services[0]: missing required field "type"
    // - middlewares[1].name: duplicate middleware name "logger"
    log.Fatal(err)
}
```

---

## Complete Examples

### Basic Configuration
```yaml
# config/app.yaml
configs:
  - name: app.name
    value: "My Application"
  - name: app.port
    value: 8080

middlewares:
  - name: logger
    type: logger
    config:
      level: INFO

services:
  - name: user-service
    type: user-service-factory

servers:
  - name: main
    base-url: http://localhost:8080
    apps:
      - addr: ":8080"
        services:
          - user-service
```

**Code:**
```go
cfg := config.New()
err := config.LoadConfigFile("config/app.yaml", cfg)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("App: %s\n", cfg.Configs[0].Value)
fmt.Printf("Services: %d\n", len(cfg.Services.Simple))
```

---

### Multi-File Configuration
```yaml
# config/01-base.yaml
configs:
  - name: app.env
    value: "${APP_ENV:development}"

# config/02-database.yaml
services:
  - name: db-service
    type: postgres
    config:
      dsn: "${DATABASE_URL}"

# config/03-api.yaml
services:
  - name: user-service
    type: user-service-factory
    depends-on:
      - db-service

servers:
  - name: api
    apps:
      - addr: ":8080"
        services:
          - user-service
```

**Code:**
```go
cfg := config.New()
err := config.LoadConfigDir("config/", cfg)
if err != nil {
    log.Fatal(err)
}
```

---

### Layered Services
```yaml
services:
  # Infrastructure layer
  infrastructure:
    - name: db-service
      type: postgres
    - name: cache-service
      type: redis
    - name: queue-service
      type: rabbitmq
  
  # Business logic layer
  business:
    - name: user-service
      type: user-service-factory
      depends-on:
        - db-service
        - cache-service
    - name: order-service
      type: order-service-factory
      depends-on:
        - db-service
        - queue-service
  
  # API layer
  api:
    - name: rest-api-service
      type: rest-api-factory
      depends-on:
        - user-service
        - order-service
```

**Code:**
```go
cfg := config.New()
config.LoadConfigFile("layered.yaml", cfg)

// Get all services in layer order
allServices := cfg.Services.Flatten()
for _, svc := range allServices {
    fmt.Printf("Service: %s (Type: %s)\n", svc.Name, svc.Type)
}
```

---

### Environment-Specific Config
```yaml
# config/base.yaml
configs:
  - name: app.name
    value: "MyApp"
  - name: app.env
    value: "${APP_ENV:development}"

services:
  - name: user-service
    type: user-service-factory

# config/development.yaml
configs:
  - name: db.dsn
    value: "postgresql://localhost/dev_db"
  - name: log.level
    value: "DEBUG"

# config/production.yaml
configs:
  - name: db.dsn
    value: "${DATABASE_URL}"
  - name: log.level
    value: "INFO"
```

**Code:**
```go
cfg := config.New()

// Load base config
config.LoadConfigFile("config/base.yaml", cfg)

// Load environment-specific config
env := os.Getenv("APP_ENV")
if env == "" {
    env = "development"
}
config.LoadConfigFile(fmt.Sprintf("config/%s.yaml", env), cfg)
```

---

### Auto-Router Configuration
```yaml
services:
  - name: user-service
    type: user-service-factory
    auto-router:
      convention: rest
      path-prefix: /api/v1
      resource-name: user
      plural-resource-name: users
      routes:
        - name: Login
          method: POST
          path: /auth/login
        - name: Logout
          method: POST
          path: /auth/logout
        - name: ChangePassword
          method: PUT
          path: /users/{id}/password

servers:
  - name: api
    apps:
      - addr: ":8080"
        services:
          - user-service
```

---

## Best Practices

### 1. Use Layered Services for Large Applications
```yaml
# ✅ Good: Clear separation of concerns
services:
  infrastructure:
    - name: db
    - name: cache
  business:
    - name: users
    - name: orders
  api:
    - name: rest-api

# 🚫 Avoid: Flat list in large apps
services:
  - name: db
  - name: cache
  - name: users
  - name: orders
  - name: rest-api
```

---

### 2. Use Config References for DRY
```yaml
# ✅ Good: Single source of truth
configs:
  - name: api.version
    value: "v1"
  - name: api.prefix
    value: "/api/${@cfg:api.version}"

# 🚫 Avoid: Duplication
configs:
  - name: user.path
    value: "/api/v1/users"
  - name: order.path
    value: "/api/v1/orders"
```

---

### 3. Use Environment Variables for Secrets
```yaml
# ✅ Good: Secrets from environment
configs:
  - name: db.password
    value: "${DB_PASSWORD}"
  - name: api.key
    value: "${API_KEY}"

# 🚫 Avoid: Hardcoded secrets
configs:
  - name: db.password
    value: "hardcoded_password"
```

---

### 4. Split Config into Multiple Files
```yaml
# ✅ Good: Organized by concern
config/
  ├── 01-base.yaml       # App-level config
  ├── 02-database.yaml   # Database services
  ├── 03-business.yaml   # Business services
  └── 04-servers.yaml    # Server topology

# 🚫 Avoid: Everything in one file
config/
  └── everything.yaml    # 1000+ lines
```

---

### 5. Use Custom Resolvers for External Config Sources
```yaml
# ✅ Good: Use appropriate resolver for each source
configs:
  - name: db.password
    value: "${@AWS:prod/db/password}"      # Secrets from AWS
  - name: feature.flags
    value: "${@K8S:app-config/features}"   # Config from K8s
  - name: app.env
    value: "${APP_ENV:development}"        # Simple env var

# 🚫 Avoid: Hardcoding external configs
configs:
  - name: db.password
    value: "hardcoded"
```

---

### 6. Design Resolver Fallback Chains
```yaml
# ✅ Good: Graceful fallback
configs:
  - name: api.key
    value: "${@AWS:prod/api/key:${API_KEY:default-key}}"

# 🚫 Avoid: No fallback (fails in dev)
configs:
  - name: api.key
    value: "${@AWS:prod/api/key}"  # Fails if AWS not configured
```

---

### 7. Use Explicit Resolver Names for Clarity
```yaml
# ✅ Good: Explicit and clear
configs:
  - name: db.host
    value: "${@ENV:DB_HOST:localhost}"
  - name: api.prefix
    value: "${@CFG:api.version}"

# 🚫 Avoid: Ambiguous (is it ENV or something else?)
configs:
  - name: db.host
    value: "${DB_HOST:localhost}"  # Works, but less clear
```

---

## See Also

- **[Deploy](./deploy)** - Deployment topology management
- **[Schema](./schema)** - YAML schema definitions
- **[lokstra_registry](../02-registry/lokstra_registry)** - Registry API
- **[Variable Resolvers](#custom-resolvers)** - Custom resolver implementation

---

## Related Guides

- **[Configuration Essentials](../../02-framework-guide/04-configuration/)** - Configuration basics
- **[Environment Variables](../../02-framework-guide/04-configuration/#environment)** - Environment management
- **[Deployment Patterns](../../04-guides/deployment/)** - Deployment strategies
- **[Secrets Management](../../04-guides/secrets-management/)** - Handling sensitive data
