# Schema and Configuration Validation

Lokstra provides comprehensive JSON Schema validation for YAML configuration files, published at [https://lokstra.dev/schema](https://lokstra.dev/schema) for use with YAML language servers and IDE integration.

## Schema Overview

The schema system provides:
- **Auto-completion** for all configuration keys
- **Real-time validation** with error highlighting
- **Documentation** via hover tooltips
- **Type checking** for values
- **Enum validation** for restricted values

## Schema Files

### Main Schema (`lokstra.json`)

**URL**: [https://lokstra.dev/schema/lokstra.json](https://lokstra.dev/schema/lokstra.json)  
**Local Path**: `/schema/lokstra.json`

The main schema file defines the complete structure for Lokstra YAML configuration files including:

- **Server Configuration**: Global settings and server name
- **Apps Configuration**: HTTP listeners, routers, and routing
- **Services Configuration**: All built-in and custom services
- **Modules Configuration**: External module loading
- **Middleware Configuration**: Built-in and custom middleware

### Group Include Schema (`group-include.json`)

**URL**: [https://lokstra.dev/schema/group-include.json](https://lokstra.dev/schema/group-include.json)  
**Local Path**: `/schema/group-include.json`

Schema for files loaded via the `load_from` feature in group configurations:

- **Routes**: Route definitions to be included
- **Groups**: Nested group configurations
- **Mount Points**: Static, HTMX, and proxy mount configurations

## IDE Integration

### VS Code Setup

Add to your VS Code `settings.json`:

```json
{
  "yaml.schemas": {
    "https://lokstra.dev/schema/lokstra.json": [
      "**/configs/**/*.yaml",
      "**/config/**/*.yaml",
      "lokstra.yaml",
      "server.yaml",
      "**/lokstra/**/*.yaml"
    ],
    "https://lokstra.dev/schema/group-include.json": [
      "**/*group*.yaml",
      "**/*include*.yaml",
      "**/group-*.yaml",
      "**/include-*.yaml"
    ]
  },
  "yaml.format.enable": true,
  "yaml.validate": true,
  "yaml.completion": true,
  "yaml.hover": true
}
```

### Direct Schema Reference

Add to the top of your YAML files:

```yaml
# yaml-language-server: $schema=https://lokstra.dev/schema/lokstra.json

server:
  name: my-server
  global_setting:
    log_level: info

apps:
  - name: api-app
    # Auto-completion and validation available here
```

### Other Editors

Most modern editors with YAML language server support can use these schemas:

- **JetBrains IDEs**: Add schema mapping in Settings → JSON Schema Mappings
- **Vim/Neovim**: Use with coc-yaml or vim-lsp
- **Emacs**: Configure with lsp-mode
- **Sublime Text**: Use with LSP package

## Service Configuration Schemas

The schema provides conditional validation based on service type, with specific configuration properties for each service:

### Database Pool (`lokstra.dbpool_pg`)

```yaml
services:
  - name: "main_db"
    type: "lokstra.dbpool_pg"
    config:
      # Connection options (auto-completed)
      host: "localhost"          # string
      port: 5432                 # integer (1-65535)
      database: "myapp"          # string (required)
      username: "postgres"       # string (required)
      password: "secret"         # string
      
      # SSL configuration (enum validation)
      sslmode: "require"         # disable|allow|prefer|require|verify-ca|verify-full
      
      # Pool settings (with validation)
      min_connections: 2         # integer >= 0
      max_connections: 20        # integer >= 1
      max_idle_time: "30m"      # duration pattern
      max_lifetime: "1h"        # duration pattern
      
      # Advanced options
      tenant_mode: true          # boolean
      default_schema: "public"   # string
```

### Redis Service (`lokstra.redis`)

```yaml
services:
  - name: "redis_main"
    type: "lokstra.redis"
    config:
      # Connection options
      addr: "localhost:6379"    # string (required)
      password: ""              # string
      db: 0                     # integer (0-15)
      
      # Pool settings
      max_idle: 10              # integer >= 0
      max_active: 100           # integer >= 1
      idle_timeout: "5m"        # duration pattern
      
      # Advanced options
      ping_interval: "30s"      # duration pattern
      max_retries: 3            # integer >= 0
```

### Logger Service (`lokstra.logger`)

```yaml
services:
  - name: "app_logger"
    type: "lokstra.logger"
    config:
      # Basic settings (enum validation)
      level: "info"             # debug|info|warn|error|fatal|panic
      format: "json"            # json|text|console
      output: "stdout"          # stdout|stderr|file
      
      # File output (conditional validation)
      file_path: "./logs/app.log"  # string (required if output: file)
      max_size: 100             # integer > 0 (MB)
      max_backups: 5            # integer >= 0
      max_age: 30               # integer >= 0 (days)
      compress: true            # boolean
      
      # Advanced options
      caller: true              # boolean
      stacktrace: true          # boolean
```

### Metrics Service (`lokstra.metrics`)

```yaml
services:
  - name: "app_metrics"
    type: "lokstra.metrics"
    config:
      enabled: true             # boolean
      endpoint: "/metrics"      # string (path pattern)
      namespace: "lokstra"      # string
      subsystem: "app"          # string
      
      # Collection options
      collect_runtime: true     # boolean
      collect_process: true     # boolean
      collect_http: true        # boolean
      
      # Histogram configuration
      buckets:                  # array of numbers
        - 0.1
        - 0.5
        - 1.0
        - 2.5
        - 5.0
        - 10.0
```

### Health Check Service (`lokstra.health_check`)

```yaml
services:
  - name: "health_service"
    type: "lokstra.health_check"
    config:
      timeout: "10s"            # duration pattern
      
      # Built-in checks configuration
      checks:
        application:
          enabled: true         # boolean
        memory:
          enabled: true         # boolean
          threshold_mb: 1024    # integer > 0
        disk:
          enabled: true         # boolean
          path: "/tmp"          # string (path)
          threshold_percent: 80.0  # number (0-100)
        database:
          enabled: true         # boolean
          service: "main_db"    # string (service reference)
        redis:
          enabled: true         # boolean
          service: "redis_main" # string (service reference)
```

## Middleware Configuration Schemas

Built-in middleware have specific configuration schemas:

### Request Logger Middleware

```yaml
middleware:
  - name: "request_logger"
    enabled: true               # boolean
    config:
      include_request_body: false   # boolean
      include_response_body: false  # boolean
```

### Body Limit Middleware

```yaml
middleware:
  - name: "body_limit"
    enabled: true               # boolean
    config:
      max_size: 10485760        # integer > 0 (bytes)
      skip_large_payloads: false # boolean
      message: "Request body too large"  # string
      status_code: 413          # integer (400-599)
      skip_on_path:             # array of strings (path patterns)
        - "/uploads/*"
        - "/api/bulk/*"
```

### Recovery Middleware

```yaml
middleware:
  - name: "recovery"
    enabled: true               # boolean
    config:
      enable_stack_trace: true  # boolean
      print_stack: false        # boolean
      log_stack: true           # boolean
```

### CORS Middleware

```yaml
middleware:
  - name: "cors"
    enabled: true               # boolean
    config:
      allowed_origins:          # array of strings
        - "*"
        - "https://app.example.com"
      allowed_methods:          # array of HTTP methods
        - "GET"
        - "POST" 
        - "PUT"
        - "DELETE"
        - "OPTIONS"
      allowed_headers:          # array of strings
        - "*"
        - "Content-Type"
        - "Authorization"
      exposed_headers:          # array of strings
        - "Content-Length"
      allow_credentials: false  # boolean
      max_age: 86400           # integer >= 0 (seconds)
```

## Routing Configuration Schema

### Route Definitions

```yaml
routes:
  - method: "GET"               # enum: GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS|TRACE|CONNECT
    path: "/users/:id"          # string (route pattern)
    handler: "user.profile"     # string (handler name)
    override_middleware: false  # boolean
    middleware:                 # middleware list
      - name: "auth"
        enabled: true
```

### Group Configurations

```yaml
groups:
  - prefix: "/api/v1"           # string (path pattern starting with /)
    override_middleware: false  # boolean
    middleware:                 # middleware list
      - "request_logger"
      - name: "body_limit"
        config:
          max_size: 1048576
    
    routes:                     # array of routes
      - method: "GET"
        path: "/users"
        handler: "user.list"
    
    groups:                     # nested groups
      - prefix: "/admin"
        routes:
          - method: "GET"
            path: "/dashboard"
            handler: "admin.dashboard"
    
    load_from:                  # array of file paths
      - "./configs/api-routes.yaml"
      - "./configs/admin-routes.yaml"
```

### Mount Point Configurations

#### Static File Serving

```yaml
mount_static:
  - prefix: "/static"           # string (URL prefix starting with /)
    folder:                     # array of folder paths
      - "./public"
      - "./assets"
    spa: false                  # boolean (SPA mode)
```

#### HTMX Application Mounting

```yaml
mount_htmx:
  - prefix: "/"                 # string (URL prefix starting with /)
    sources:                    # array of source paths (required)
      - "./htmx_content"
      - "./htmx_app"
```

#### Reverse Proxy

```yaml
mount_reverse_proxy:
  - prefix: "/api"              # string (URL prefix starting with /)
    target: "http://localhost:3000"  # string (URI format)
```

## Schema Validation Rules

### Duration Format

Duration fields use Go duration format with pattern validation:

**Pattern**: `^[0-9]+(ns|us|µs|ms|s|m|h)$`

**Valid Examples**:
- `"5s"` (5 seconds)
- `"30m"` (30 minutes)  
- `"1h"` (1 hour)
- `"2h30m"` (2 hours 30 minutes)
- `"100ms"` (100 milliseconds)

**Invalid Examples**:
- `"5"` (missing unit)
- `"5 seconds"` (spaces not allowed)
- `"5sec"` (invalid unit)

### Port Number Validation

Port numbers are validated with range constraints:

```yaml
port: 8080                      # integer (1-65535)
```

### Path Pattern Validation

URL paths must start with `/`:

```yaml
prefix: "/api/v1"               # valid
prefix: "api/v1"                # invalid (missing leading /)
```

### HTTP Method Validation

HTTP methods are validated against allowed values:

```yaml
method: "GET"                   # valid
method: "get"                   # invalid (case sensitive)
method: "CUSTOM"                # invalid (not in enum)
```

### SSL Mode Validation

PostgreSQL SSL modes are validated against PostgreSQL standards:

```yaml
sslmode: "require"              # valid
sslmode: "required"             # invalid (not in enum)
```

**Valid SSL Modes**:
- `disable`
- `allow` 
- `prefer`
- `require`
- `verify-ca`
- `verify-full`

## Error Messages and Validation

The schema provides descriptive error messages for validation failures:

### Required Field Errors

```yaml
# Missing required field
services:
  - name: "db"
    type: "lokstra.dbpool_pg"
    config:
      host: "localhost"
      # Error: Missing required property 'database'
```

### Type Validation Errors

```yaml
# Incorrect type
services:
  - name: "db"
    type: "lokstra.dbpool_pg"
    config:
      port: "8080"              # Error: Expected integer, got string
```

### Enum Validation Errors

```yaml
# Invalid enum value
routes:
  - method: "PATCH"
    path: "/users"
    handler: "user.update"      # Valid
  - method: "INVALID"           # Error: Value not in enum [GET, POST, PUT, DELETE, ...]
```

### Pattern Validation Errors

```yaml
# Invalid duration format
services:
  - name: "db"
    type: "lokstra.dbpool_pg"
    config:
      max_idle_time: "30 minutes"  # Error: Does not match duration pattern
```

## Custom Validation

For custom services and middleware, extend the schema:

### Adding Custom Service Type

```json
{
  "if": {
    "properties": {
      "type": { "const": "my_custom.service" }
    }
  },
  "then": {
    "properties": {
      "config": {
        "type": "object",
        "properties": {
          "custom_setting": {
            "type": "string",
            "description": "Custom service setting"
          }
        },
        "required": ["custom_setting"],
        "additionalProperties": false
      }
    }
  }
}
```

### Adding Custom Middleware

```json
{
  "if": {
    "properties": {
      "name": { "const": "my_custom_middleware" }
    }
  },
  "then": {
    "properties": {
      "config": {
        "type": "object",
        "properties": {
          "custom_option": {
            "type": "boolean",
            "description": "Custom middleware option"
          }
        },
        "additionalProperties": false
      }
    }
  }
}
```

## Schema Versioning

The schema follows semantic versioning:

- **Major Version**: Breaking changes to existing configurations
- **Minor Version**: New service types or optional properties  
- **Patch Version**: Bug fixes, improved descriptions, examples

Current schema version is available at:
- [https://lokstra.dev/schema/version.json](https://lokstra.dev/schema/version.json)

## Online Schema Validation

Use the online schema for validation without local files:

```yaml
# yaml-language-server: $schema=https://lokstra.dev/schema/lokstra.json

# Your configuration with real-time validation
server:
  name: my-app
  
services:
  - name: "main_db"
    type: "lokstra.dbpool_pg"
    # Auto-completion and validation available
```

## Manual Validation

For programmatic validation, use JSON Schema validation libraries:

### Go Example

```go
package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    
    "github.com/xeipuuv/gojsonschema"
    "gopkg.in/yaml.v3"
)

func validateConfig(yamlFile string) error {
    // Load schema from URL
    schemaLoader := gojsonschema.NewReferenceLoader("https://lokstra.dev/schema/lokstra.json")
    
    // Read YAML file
    yamlData, err := ioutil.ReadFile(yamlFile)
    if err != nil {
        return err
    }
    
    // Convert YAML to JSON
    var yamlContent any
    err = yaml.Unmarshal(yamlData, &yamlContent)
    if err != nil {
        return err
    }
    
    jsonData, err := json.Marshal(yamlContent)
    if err != nil {
        return err
    }
    
    // Validate against schema
    documentLoader := gojsonschema.NewBytesLoader(jsonData)
    result, err := gojsonschema.Validate(schemaLoader, documentLoader)
    if err != nil {
        return err
    }
    
    if !result.Valid() {
        for _, err := range result.Errors() {
            fmt.Printf("Validation error: %s\n", err)
        }
        return fmt.Errorf("configuration validation failed")
    }
    
    return nil
}
```

### Node.js Example

```javascript
const Ajv = require('ajv');
const yaml = require('js-yaml');
const fs = require('fs');
const fetch = require('node-fetch');

async function validateConfig(yamlFile) {
    // Load schema
    const schemaResponse = await fetch('https://lokstra.dev/schema/lokstra.json');
    const schema = await schemaResponse.json();
    
    // Load and parse YAML
    const yamlContent = fs.readFileSync(yamlFile, 'utf8');
    const config = yaml.load(yamlContent);
    
    // Validate
    const ajv = new Ajv();
    const validate = ajv.compile(schema);
    const valid = validate(config);
    
    if (!valid) {
        console.log('Validation errors:', validate.errors);
        return false;
    }
    
    return true;
}
```

## Best Practices

### 1. Always Use Schema Validation

Enable schema validation in your IDE and CI/CD pipelines:

```yaml
# .github/workflows/validate.yml
name: Validate Configuration
on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Validate YAML
        run: |
          npm install -g ajv-cli
          ajv validate -s https://lokstra.dev/schema/lokstra.json -d "configs/**/*.yaml"
```

### 2. Use Environment-specific Configuration

Structure configuration files by environment:

```
configs/
├── development/
│   ├── server.yaml
│   └── services.yaml
├── staging/
│   ├── server.yaml
│   └── services.yaml
└── production/
    ├── server.yaml
    └── services.yaml
```

### 3. Validate Before Deployment

Always validate configuration before deploying:

```bash
# Validate all configuration files
ajv validate -s ./schema/lokstra.json -d "configs/**/*.yaml"

# Validate specific environment
ajv validate -s ./schema/lokstra.json -d "configs/production/*.yaml"
```

### 4. Keep Schema Updated

Regularly update to the latest schema version for new features and improvements.

### 5. Document Custom Extensions

When extending the schema for custom services, document the extensions:

```yaml
# yaml-language-server: $schema=./schema/lokstra-extended.json

# Custom service with documented configuration
services:
  - name: "my_service"
    type: "company.custom_auth"
    config:
      # Custom authentication provider
      provider_url: "https://auth.company.com"
      client_id: "app-client-id"
      # ... other custom options
```

## Next Steps

- [Built-in Services](./built-in-services.md) - Learn about available services
- [Built-in Middleware](./built-in-middleware.md) - Explore middleware options
- [Configuration](./configuration.md) - Advanced configuration patterns
- [Getting Started](./getting-started.md) - Quick start guide

---

*Schema validation in Lokstra ensures configuration correctness, provides excellent developer experience with IDE integration, and reduces configuration errors in production deployments.*