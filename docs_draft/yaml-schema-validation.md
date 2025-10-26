# YAML Configuration JSON Schema Validation

## Overview

Lokstra framework sekarang dilengkapi dengan JSON Schema validation untuk YAML configuration files. Validasi ini memastikan bahwa konfigurasi YAML Anda memenuhi struktur yang benar sebelum digunakan oleh aplikasi.

## Fitur

- ✅ **Automatic Validation**: Semua file YAML divalidasi secara otomatis saat loading
- ✅ **Detailed Error Messages**: Pesan error yang jelas menunjukkan field mana yang bermasalah
- ✅ **Type Checking**: Memastikan tipe data yang benar untuk setiap field
- ✅ **Required Fields**: Memvalidasi bahwa semua field required ada
- ✅ **Format Validation**: Validasi format URL, dll
- ✅ **Min/Max Constraints**: Validasi panjang minimum string dan array

## Schema Structure

### Root Properties

```yaml
configs:      # Optional array of general configuration
services:     # Optional array of service configurations
middlewares:  # Optional array of middleware configurations
servers:      # Optional array of server configurations
```

### Configs (General Configuration)

```yaml
configs:
  - name: string          # Required, min length: 1
    value: any           # Required, can be any type
```

### Services

```yaml
services:
  - name: string         # Required, min length: 1
    type: string         # Required, min length: 1
    enable: boolean      # Optional, default: true
    config: object       # Optional, any properties allowed
```

### Middlewares

```yaml
middlewares:
  - name: string         # Required, min length: 1
    type: string         # Required, min length: 1
    enable: boolean      # Optional, default: true
    config: object       # Optional, any properties allowed
```

### Servers

```yaml
servers:
  - name: string              # Required, min length: 1
    baseUrl: string           # Optional, must be valid http/https URL
    deployment-id: string     # Optional
    apps: array               # Required, minimum 1 item
      - name: string          # Optional, auto-generated if not provided
        addr: string          # Required, min length: 1
        listener-type: string # Optional, default: "default"
        routers: array        # Optional
          - string            # Router name, min length: 1
```

## Usage

### Loading with Validation

Validasi dilakukan secara otomatis saat memanggil `LoadConfigFs` atau `LoadConfigFile`:

```go
import "github.com/primadi/lokstra/core/config"

// Load single file
cfg := config.New()
err := config.LoadConfigFile("config.yaml", cfg)
if err != nil {
    // Error will contain validation details if schema validation fails
    log.Fatal(err)
}

// Load from directory
cfg := config.New()
err := config.LoadConfigDir("configs/", cfg)
if err != nil {
    log.Fatal(err)
}

// Load from embedded filesystem
//go:embed configs
var configFS embed.FS

cfg := config.New()
err := config.LoadConfigFs(configFS, "configs/app.yaml", cfg)
if err != nil {
    log.Fatal(err)
}
```

### Manual Validation

Anda juga bisa melakukan validasi manual:

```go
import "github.com/primadi/lokstra/core/config"

// Validate YAML string
yamlContent := `
servers:
  - name: test-server
    apps:
      - addr: :8080
`
err := config.ValidateYAMLString(yamlContent)
if err != nil {
    log.Println("Validation failed:", err)
}

// Validate Config struct
cfg := &config.Config{
    Servers: []*config.Server{
        {
            Name: "test-server",
            Apps: []*config.App{
                {Addr: ":8080"},
            },
        },
    },
}
err = config.ValidateConfig(cfg)
if err != nil {
    log.Println("Validation failed:", err)
}
```

## Validation Error Examples

### Missing Required Field

```yaml
# ❌ ERROR: Missing required 'name' field
servers:
  - baseUrl: http://localhost
    apps:
      - addr: :8080
```

Error message:
```
YAML configuration validation failed:
  - servers.0.name: name is required
```

### Invalid Format

```yaml
# ❌ ERROR: Invalid baseUrl format
servers:
  - name: test-server
    baseUrl: not-a-valid-url
    apps:
      - addr: :8080
```

Error message:
```
YAML configuration validation failed:
  - servers.0.baseUrl: Does not match pattern '^https?://[^\s/$.?#].[^\s]*$'
```

### Empty Required Array

```yaml
# ❌ ERROR: Empty apps array
servers:
  - name: test-server
    apps: []
```

Error message:
```
YAML configuration validation failed:
  - servers.0.apps: Array must have at least 1 items
```

### Missing Required Property

```yaml
# ❌ ERROR: Missing required 'type' in service
services:
  - name: database
    config:
      host: localhost
```

Error message:
```
YAML configuration validation failed:
  - services.0.type: type is required
```

## Valid Configuration Example

```yaml
# Complete valid configuration
configs:
  - name: app-version
    value: "1.0.0"

services:
  - name: database
    type: postgres
    enable: true
    config:
      host: localhost
      port: 5432

middlewares:
  - name: cors
    type: cors
    config:
      allowOrigins: ["*"]

servers:
  - name: main-server
    baseUrl: http://localhost
    deployment-id: dev
    apps:
      - name: api
        addr: :8080
        listener-type: default
        routers:
          - product-api
          - order-api
```

## Schema File Location

The JSON schema is defined in:
```
core/config/schema.go
```

Validator functions are in:
```
core/config/validator.go
```

## Testing

Run the validation tests:

```bash
go test -v ./core/config -run TestValidate
```

## Benefits

1. **Early Error Detection**: Catch configuration errors before deployment
2. **Better DX**: Clear error messages help developers fix issues quickly
3. **Documentation**: Schema serves as documentation for configuration structure
4. **IDE Support**: JSON schema can be used by IDEs for autocompletion
5. **Type Safety**: Ensures correct data types are used throughout

## Best Practices

1. **Always validate early**: Load and validate configs at application startup
2. **Use descriptive names**: Make service/middleware names clear and meaningful
3. **Enable only what you need**: Set `enable: false` for unused services
4. **Group related configs**: Use deployment-id to group related servers
5. **Version your configs**: Include version in configs for tracking changes

## Troubleshooting

### Issue: "Additional property X is not allowed"
**Solution**: Check that you're using the correct property names as defined in the schema.

### Issue: "Does not match pattern"
**Solution**: For URLs, ensure they start with `http://` or `https://`

### Issue: "Array must have at least 1 items"
**Solution**: Ensure arrays like `apps` are not empty when required

### Issue: "X is required"
**Solution**: Add the missing required field to your configuration

## Future Enhancements

- [ ] Custom schema validators for specific service types
- [ ] Schema versioning support
- [ ] IDE integration with schema files
- [ ] Configuration migration tools
- [ ] Environment-specific schema variations
