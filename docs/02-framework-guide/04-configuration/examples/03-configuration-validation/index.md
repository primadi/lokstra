# Configuration Validation

Validate configuration at startup to catch errors early.

## Running

```bash
# With valid configuration
go run main.go valid

# With invalid configuration (will fail)
go run main.go invalid
```

## Validation Rules

Required configurations:
- `app_name` - Application name
- `app_version` - Version number
- `db_host` - Database host

## Benefits

- Fail fast on missing config
- Clear error messages
- Prevent runtime errors
- Better debugging experience

## Pattern

```go
func ValidateConfig(cfg *config.Config) error {
    required := []string{"app_name", "app_version", "db_host"}
    
    for _, req := range required {
        if !configMap[req] {
            return fmt.Errorf("missing required config: %s", req)
        }
    }
    
    return nil
}
```
