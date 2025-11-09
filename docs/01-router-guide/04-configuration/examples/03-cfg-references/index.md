# CFG References Example

Demonstrates using CFG references for DRY (Don't Repeat Yourself) configuration.

## What This Demonstrates

- CFG references: `${@CFG:path.to.value}`
- Shared configuration in `configs` section
- Reusing values across configuration
- Dynamic path construction with references

## Files

- `config.yaml` - Configuration with CFG references
- `main.go` - Application code
- `test.http` - HTTP requests for testing (use with VS Code REST Client)

## Key Concepts

### CFG Reference Syntax
```yaml
${@CFG:path.to.value}
```

### Example Usage

**Define once in `configs` section:**
```yaml
configs:
  database:
    host: localhost
    port: 5432
  
  api:
    version: v1
    base_path: /api
```

**Reference anywhere:**
```yaml
services:
  - name: db
    config:
      host: ${@CFG:database.host}      # Resolves to "localhost"
      port: ${@CFG:database.port}      # Resolves to 5432

routers:
  - name: api
    routes:
      - path: ${@CFG:api.base_path}/${@CFG:api.version}/users
        # Resolves to "/api/v1/users"
```

## Benefits

✅ **DRY** - Define values once, use everywhere
✅ **Consistency** - Single source of truth
✅ **Easy updates** - Change in one place
✅ **Type-safe** - Values resolved during config load

## Run

```bash
go run main.go
```

## Test

```bash
# Health check
curl http://localhost:8080/api/health

# Get users (path built from CFG refs)
curl http://localhost:8080/api/v1/users

# Get config info
curl http://localhost:8080/api/v1/config
```

## Expected Output

```
🚀 Server starting on http://localhost:8080

📖 Try:
   curl http://localhost:8080/api/health
   curl http://localhost:8080/api/v1/users
   curl http://localhost:8080/api/v1/config

💡 This example demonstrates:
   - CFG references: ${@CFG:path.to.value}
   - Shared config values in 'configs' section
   - DRY principle in configuration
```

## Configuration Structure

The `config.yaml` shows:
1. **configs** section - Shared values (database, api, features)
2. **services** section - References to shared database config
3. **routers** section - Paths constructed with CFG references
4. **servers** section - Standard server configuration

All `${@CFG:...}` references are resolved when config is loaded, ensuring consistency across the application.
