# YAML Configuration Schema Validation

## Quick Start

JSON Schema validation telah diintegrasikan ke dalam Lokstra framework untuk memvalidasi YAML configuration files secara otomatis.

## Cara Penggunaan

### 1. Load Config dengan Auto-Validation

```go
import "github.com/primadi/lokstra/core/config"

// Semua load functions sekarang include validasi otomatis
cfg := config.New()
err := config.LoadConfigFile("config.yaml", cfg)
if err != nil {
    // Error akan berisi detail validasi jika gagal
    log.Fatal(err)
}
```

### 2. Manual Validation

```go
// Validasi YAML string
yamlContent := `...`
err := config.ValidateYAMLString(yamlContent)

// Validasi Config struct
cfg := &config.Config{...}
err := config.ValidateConfig(cfg)
```

### 3. Test Validation Tool

```bash
cd core/config
go run ./test-validate/main.go example-valid.yaml
```

## Schema Rules

### Required Fields

**Server**:
- `name` (string, min length: 1)
- `apps` (array, min items: 1)

**App**:
- `addr` (string, min length: 1)

**Service**:
- `name` (string, min length: 1)
- `type` (string, min length: 1)

**Middleware**:
- `name` (string, min length: 1)
- `type` (string, min length: 1)

**Config**:
- `name` (string, min length: 1)
- `value` (any type)

### Format Validation

**baseUrl**: Must match pattern `^https?://[^\s/$.?#].[^\s]*$`

## Example Errors

### Missing Required Field
```
❌ Validation FAILED:
validation failed for config.yaml: YAML configuration validation failed:
  - servers.0.name: String length must be greater than or equal to 1
```

### Invalid URL Format
```
❌ Validation FAILED:
validation failed for config.yaml: YAML configuration validation failed:
  - servers.0.baseUrl: Does not match pattern '^https?://[^\s/$.?#].[^\s]*$'
```

### Empty Array
```
❌ Validation FAILED:
validation failed for config.yaml: YAML configuration validation failed:
  - servers.0.apps: Array must have at least 1 items
```

## Files Created

```
core/config/
├── schema.go              # JSON Schema definition
├── validator.go           # Validation functions
├── validator_test.go      # Validation tests
├── example-valid.yaml     # Valid config example
├── example-invalid.yaml   # Invalid config examples
└── test-validate/         # CLI tool untuk test validasi
    └── main.go
```

## Documentation

Full documentation: [docs/yaml-schema-validation.md](../../docs/yaml-schema-validation.md)

## Testing

```bash
# Run validation tests
go test -v ./core/config -run TestValidate

# Test with example files
cd core/config
go run ./test-validate/main.go example-valid.yaml
go run ./test-validate/main.go test-invalid-1.yaml
```

## Benefits

✅ Early error detection  
✅ Clear error messages  
✅ Type safety  
✅ Better developer experience  
✅ Self-documenting configuration structure  

## Integration

Schema validation sudah terintegrasi ke dalam:
- `LoadConfigFs()`
- `LoadConfigFile()`
- `LoadConfigDirFs()`
- `LoadConfigDir()`

Tidak ada perubahan kode yang diperlukan pada existing code yang menggunakan functions tersebut.
