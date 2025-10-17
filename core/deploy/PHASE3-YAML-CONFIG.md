# Phase 3: YAML Configuration & Validation

## Overview
Phase 3 implements multi-file YAML configuration with JSON Schema validation. Configuration can be split across multiple files and automatically merged and validated.

## Features

### ✅ Multi-File Support
- Load configuration from multiple YAML files
- Automatic merging with override strategy (later files override earlier ones)
- Load entire directories of YAML files
- Support for both relative and absolute paths

### ✅ JSON Schema Validation
- Embedded JSON Schema using `embed.FS`
- Automatic validation after loading
- Clear validation error messages
- Schema enforces naming conventions and structure

### ✅ Configuration Structure
```yaml
# Global configurations
configs:
  KEY_NAME: value

# Service definitions
services:
  service-name:
    type: factory-type
    depends-on:
      - dependency1
      - paramName:dependency2
    config:
      key: value

# Router definitions
routers:
  router-name:
    service: service-name
    overrides:
      MethodName:
        hide: true
        middleware: [mw1, mw2]

# Remote service definitions
remote-services:
  remote-name:
    url: https://api.example.com
    resource: item
    resource-plural: items

# Deployment configurations
deployments:
  deployment-name:
    config-overrides:
      KEY_NAME: override-value
    servers:
      server-name:
        base-url: https://example.com
        apps:
          - port: 8080
            services: [service1, service2]
            routers: [router1]
            remote-services: [remote1]
```

## API Usage

### Loading Single File
```go
config, err := loader.LoadConfig("config.yaml")
if err != nil {
    log.Fatal(err)
}
```

### Loading Multiple Files (Merge)
```go
config, err := loader.LoadConfig(
    "config/base.yaml",
    "config/services.yaml",
    "config/deployments.yaml",
)
```

### Loading Directory
```go
// Loads all .yaml and .yml files in directory
config, err := loader.LoadConfigFromDir("config")
```

### Building Deployment
```go
// Create registry and register factories
reg := deploy.Global()
reg.RegisterServiceType("my-service", myFactory, nil)

// Load and build deployment
dep, err := loader.LoadAndBuildFromDir("config", "production", reg)
if err != nil {
    log.Fatal(err)
}

// Use deployment
server, _ := dep.GetServer("api-server")
app := server.Apps()[0]
svc, _ := app.GetService("my-service")
```

## JSON Schema

Located at: `core/deploy/schema/lokstra.schema.json`

Embedded in binary using `//go:embed` directive for:
- Zero runtime dependencies
- Always available
- Version-locked with code

### Naming Conventions Enforced

**Configs**: `^[A-Z][A-Z0-9_]*$`
- Examples: `DB_HOST`, `API_KEY`, `MAX_RETRIES`

**Services**: `^[a-z][a-z0-9-]*$`
- Examples: `db-pool`, `user-service`, `api-client`

**Dependencies**: `^([a-zA-Z][a-zA-Z0-9]*:)?[a-z][a-z0-9-]*$`
- Examples: `db-pool`, `dbOrder:db-pool`, `userSvc:user-service`

**URLs**: `^https?://`
- Must start with `http://` or `https://`

**Ports**: `1-65535`
- Valid TCP port range

### Validation Errors

Clear error messages with field paths:
```
schema validation failed:
  - configs.invalid-name: Does not match pattern '^[A-Z][A-Z0-9_]*$'
  - services.MyService.type: Does not match pattern '^[a-z][a-z0-9-]*$'
  - deployments.prod.servers.api.apps.0.port: Must be greater than or equal to 1
```

## Multi-File Strategy

### File Organization Patterns

#### Pattern 1: By Concern
```
config/
  ├── base.yaml           # Configs and infrastructure
  ├── services.yaml       # Application services
  ├── routers.yaml        # API routers
  └── deployments.yaml    # Deployment targets
```

#### Pattern 2: By Environment
```
config/
  ├── common.yaml         # Shared configs
  ├── development.yaml    # Dev-specific
  ├── staging.yaml        # Staging-specific
  └── production.yaml     # Prod-specific
```

#### Pattern 3: By Feature
```
config/
  ├── infrastructure.yaml # DB, cache, etc.
  ├── auth.yaml          # Authentication services
  ├── payments.yaml      # Payment services
  └── deployments.yaml   # Deployment configs
```

### Merge Strategy

**Later files override earlier files:**
```go
LoadConfig("base.yaml", "override.yaml")
// Values in override.yaml replace values in base.yaml
```

**Map keys are merged:**
```yaml
# base.yaml
services:
  db: {...}
  cache: {...}

# services.yaml
services:
  api: {...}

# Result: All three services (db, cache, api)
```

**Same key = override:**
```yaml
# base.yaml
configs:
  LOG_LEVEL: info

# prod.yaml
configs:
  LOG_LEVEL: warn

# Result: LOG_LEVEL = warn
```

## Examples

### Example 1: Basic Configuration
```yaml
# config.yaml
configs:
  DB_HOST: localhost
  DB_PORT: 5432

services:
  db-pool:
    type: postgres-pool
    config:
      host: ${@cfg:DB_HOST}
      port: ${@cfg:DB_PORT}

deployments:
  dev:
    servers:
      main:
        base-url: http://localhost
        apps:
          - port: 3000
            services: [db-pool]
```

### Example 2: Multi-File with Overrides
```yaml
# base.yaml
configs:
  DB_HOST: localhost
  LOG_LEVEL: info

services:
  db-pool:
    type: postgres-pool
    config:
      host: ${@cfg:DB_HOST}
```

```yaml
# production.yaml
configs:
  DB_HOST: prod-db.example.com
  LOG_LEVEL: warn

deployments:
  production:
    config-overrides:
      LOG_LEVEL: error
    servers:
      api:
        base-url: https://api.example.com
        apps:
          - port: 8080
            services: [db-pool]
```

### Example 3: Service Dependencies
```yaml
services:
  db-pool:
    type: postgres-pool
    config:
      host: localhost

  logger:
    type: logger-service
    config:
      level: info

  user-service:
    type: user-service-factory
    depends-on:
      - db:db-pool          # Alias: paramName:serviceName
      - logger              # Direct: uses service name as param
    config:
      enable-cache: true

  order-service:
    type: order-service-factory
    depends-on:
      - dbOrder:db-pool     # Reuse db-pool with different param name
      - userSvc:user-service
      - logger
```

## Testing

### Loader Tests
```bash
$ cd core/deploy/loader
$ go test -v

=== RUN   TestLoadSingleFile
--- PASS: TestLoadSingleFile (0.00s)
=== RUN   TestLoadMultipleFiles
--- PASS: TestLoadMultipleFiles (0.00s)
=== RUN   TestLoadFromDirectory
--- PASS: TestLoadFromDirectory (0.00s)
=== RUN   TestMergeStrategy
--- PASS: TestMergeStrategy (0.00s)
=== RUN   TestValidation_ValidConfig
--- PASS: TestValidation_ValidConfig (0.00s)
...
PASS
ok      github.com/primadi/lokstra/core/deploy/loader   1.063s
```

### Running Examples
```bash
$ cd core/deploy/examples/yaml
$ go run main.go

🚀 Lokstra YAML Configuration Example
📂 Loading configuration from YAML files...
✅ Configuration loaded and validated!
✨ Example completed successfully!
```

## IDE Support

### VS Code
Add to `.vscode/settings.json`:
```json
{
  "yaml.schemas": {
    "./core/deploy/schema/lokstra.schema.json": [
      "**/deploy/*.yaml",
      "**/deploy/*.yml",
      "**/config/*.yaml",
      "**/config/*.yml"
    ]
  }
}
```

### IntelliJ / GoLand
1. Go to Settings → Languages & Frameworks → Schemas and DTDs → JSON Schema Mappings
2. Add new mapping:
   - Schema file: `core/deploy/schema/lokstra.schema.json`
   - Schema version: JSON Schema version 7
   - File path pattern: `**/config/*.yaml`

### Benefits of IDE Integration
- ✅ Auto-completion for keys
- ✅ Inline validation errors
- ✅ Documentation on hover
- ✅ Validation as you type
- ✅ Structural suggestions

## Implementation Details

### Embedded Schema
```go
//go:embed lokstra.schema.json
var schemaFS embed.FS

func ValidateConfig(config *schema.Config) error {
    schemaData, _ := schemaFS.ReadFile("lokstra.schema.json")
    schemaLoader := gojsonschema.NewBytesLoader(schemaData)
    // ... validation
}
```

**Benefits:**
- No external files needed at runtime
- Schema versioned with code
- Binary is self-contained
- Works in any environment

### YAML Parsing
Uses `gopkg.in/yaml.v3` for:
- Full YAML 1.2 support
- Preserves types (int, string, bool)
- Handles anchors and aliases
- Detailed error messages

### Validation Flow
```
Load YAML → Parse to struct → Convert to map → Validate against schema → Return
```

**Why convert to map?**
- JSON Schema expects JSON-like structure
- Direct struct validation is more complex
- Conversion allows flexible schema evolution

## Advanced Features

### Config References
```yaml
configs:
  DB_HOST: localhost
  DB_PORT: 5432
  DB_DSN: "postgres://${@cfg:DB_HOST}:${@cfg:DB_PORT}/mydb"
```

### Environment Variables
```yaml
configs:
  DB_PASSWORD: ${DB_PASSWORD}
  API_KEY: ${API_KEY:default-key}
```

### Service Aliasing
```yaml
services:
  order-service:
    depends-on:
      - dbOrder:db-pool     # Same service, different name
      - dbUser:db-pool      # Multiple uses with different aliases
```

## Migration from Programmatic Config

### Before (Programmatic)
```go
reg := deploy.Global()
reg.DefineConfig(&schema.ConfigDef{Name: "DB_HOST", Value: "localhost"})
reg.DefineService(&schema.ServiceDef{
    Name: "db-pool",
    Type: "postgres-pool",
    Config: map[string]any{"host": "${@cfg:DB_HOST}"},
})
dep := deploy.New("prod")
server := dep.NewServer("api", "https://api.example.com")
app := server.NewApp(8080)
app.AddService("db-pool")
```

### After (YAML)
```yaml
# config.yaml
configs:
  DB_HOST: localhost

services:
  db-pool:
    type: postgres-pool
    config:
      host: ${@cfg:DB_HOST}

deployments:
  prod:
    servers:
      api:
        base-url: https://api.example.com
        apps:
          - port: 8080
            services: [db-pool]
```

```go
// main.go
reg := deploy.Global()
reg.RegisterServiceType("postgres-pool", dbFactory, nil)
dep, _ := loader.LoadAndBuild([]string{"config.yaml"}, "prod", reg)
```

**Benefits:**
- ✅ Configuration in YAML (not code)
- ✅ Can be modified without recompile
- ✅ Environment-specific configs easy
- ✅ Validation catches errors early

## Summary

**Phase 3 Achievements:**
- ✅ Multi-file YAML configuration
- ✅ Automatic file merging
- ✅ JSON Schema validation (embedded)
- ✅ IDE support via schema
- ✅ Directory loading
- ✅ Clear error messages
- ✅ 10 comprehensive tests
- ✅ Working example
- ✅ Complete documentation

**Next Steps:**
- Router configuration from YAML
- Remote service configuration
- Middleware configuration
- YAML hot-reload (development mode)
- Config templates and includes
