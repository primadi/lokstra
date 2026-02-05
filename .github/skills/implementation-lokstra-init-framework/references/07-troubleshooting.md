# Troubleshooting Lokstra Initialization

## Common Errors and Solutions

### 1. "no published-services found" Error

**Error Message:**
```
panic: no published-services found in deployment configuration
```

**Cause:** No @Handler annotations are being discovered.

**Solutions:**

#### Solution A: Verify @Handler Syntax
```go
// ✅ Correct
// @Handler name="user-handler", prefix="/api/users"

// ❌ Wrong - missing name
// @Handler prefix="/api/users"

// ❌ Wrong - invalid syntax
// @Handler: name=user-handler
```

#### Solution B: Check File Location
Ensure handler files are in correct module structure:
```
modules/
├── user/
│   └── application/
│       └── handler.go  # @Handler annotation here
├── auth/
│   └── application/
│       └── handler.go  # @Handler annotation here
```

#### Solution C: Force Regeneration
```bash
go run . --generate-only
```

#### Solution D: Verify Generated Imports
Check that `zz_lokstra_imports.go` was generated:
```bash
cat zz_lokstra_imports.go
# Should contain auto-generated imports for all modules
```

**Note:** You do NOT need to manually import modules! All module imports are auto-generated in `zz_lokstra_imports.go` during bootstrap.

---

### 2. Service Not Found Error

**Error Message:**
```
panic: service "user-repo" not found in registry
```

**Cause:** Service is injected but never registered.

**Solutions:**

#### Solution A: Check Service Definition in config.yaml
```yaml
service-definitions:
  user-repo:
    type: user_repository
    config:
      # ... config
```

#### Solution B: Verify @Service Annotation
```go
// In modules/user/infrastructure/repository.go
// @Service "user-repo"
type UserRepository struct {
    db serviceapi.DBPool
}
```

#### Solution C: Check Published Services
```yaml
deployments:
  development:
    servers:
      api:
        published-services:
          - user-handler  # Must match @Handler name
```

---

### 3. Config File Not Loaded

**Error Message:**
```
panic: config file not found: configs/config.yaml
```

**Cause:** Config path incorrect or file doesn't exist.

**Solutions:**

#### Solution A: Create configs/config.yaml
```yaml
service-definitions: {}

deployments:
  development:
    servers:
      api:
        addr: ":8080"
        published-services: []
```

#### Solution B: Use Custom Config Path
```go
lokstra_init.BootstrapAndRun(
    lokstra_init.WithYAMLConfigPath(true, "config", "local-config"),
)
```

#### Solution C: Use Environment Variable
```bash
export LOKSTRA_CONFIG_PATH=./configs
go run .
```

---

### 4. Middleware Order Issues

**Problem:** Requests not logging, auth not working, etc.

**Cause:** Middleware registered in wrong order.

**Solution:** Always follow this order:
```go
// 1. Recovery MUST be first
recovery.Register()

// 2. Logging should be early
request_logger.Register()

// 3. CORS before auth
cors.Register([]string{"*"})

// 4. Authentication
auth_middleware.Register()

// 5. Other middleware
custom_middleware.Register()
```

**Why?** Middleware is applied in reverse order of registration.

---

### 5. Port Already in Use

**Error Message:**
```
panic: listen tcp :8080: bind: address already in use
```

**Solutions:**

#### Solution A: Change Port in config.yaml
```yaml
deployments:
  development:
    servers:
      api:
        addr: ":8081"  # Use different port
```

#### Solution B: Kill Process Using Port (Windows)
```powershell
netstat -ano | findstr :8080
taskkill /PID <PID> /F
```

#### Solution C: Kill Process Using Port (Linux/Mac)
```bash
lsof -ti:8080 | xargs kill -9
```

---

### 6. Database Connection Failed

**Error Message:**
```
panic: failed to connect to database: connection refused
```

**Solutions:**

#### Solution A: Check DSN in config.yaml
```yaml
service-definitions:
  db_main:
    type: dbpool_pg
    config:
      dsn: "postgres://user:pass@localhost:5432/mydb?sslmode=disable"
```

#### Solution B: Use Environment Variables
```yaml
service-definitions:
  db_main:
    type: dbpool_pg
    config:
      dsn: ${DB_DSN:postgres://localhost:5432/mydb}
```

```bash
export DB_DSN="postgres://user:pass@localhost:5432/mydb"
go run .
```

#### Solution C: Verify Database is Running
```bash
# PostgreSQL
pg_isready -h localhost -p 5432

# Docker
docker ps | grep postgres
```

---

### 7. Circular Dependency Error

**Error Message:**
```
panic: circular dependency detected: A -> B -> C -> A
```

**Cause:** Services depend on each other in a loop.

**Solution:** Refactor dependencies:

```go
// ❌ Bad - Circular dependency
// @Service "service-a" depends on "service-b"
// @Service "service-b" depends on "service-a"

// ✅ Good - Extract common interface
// @Service "service-a" depends on "common-interface"
// @Service "service-b" depends on "common-interface"
// @Service "common-impl" implements "common-interface"
```

---

### 8. Migration Failed

**Error Message:**
```
panic: migration failed: syntax error at line 10
```

**Solutions:**

#### Solution A: Check SQL Syntax
```sql
-- migrations/001_users.up.sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL
);
-- Make sure syntax is valid PostgreSQL
```

#### Solution B: Skip Migrations on Development
```go
lokstra_init.BootstrapAndRun(
    lokstra_init.WithDbMigrations(false, "migrations"),
)
```

#### Solution C: Reset Database
```bash
# WARNING: This deletes all data!
dropdb mydb
createdb mydb
go run .
```

---

### 9. Hot Reload Not Working

**Problem:** Changes not reflected without manual restart.

**Solutions:**

#### Solution A: Verify Development Mode
```bash
# Should show: Environment detected: DEV
go run .
```

#### Solution B: Check File Permissions
```bash
# Linux/Mac
chmod -R 755 .
```

#### Solution C: Use --generate-only
```bash
# Manual regeneration
go run . --generate-only
go run .
```

---

### 10. Generated Code Not Updating

**Problem:** Changes to @Handler not reflected.

**Solutions:**

#### Solution A: Force Rebuild
```bash
go run . --generate-only
```

#### Solution B: Delete Cache Files
```bash
# Find and delete all generated files
find . -name "zz_generated.lokstra.go" -delete
find . -name "zz_lokstra_imports.go" -delete
go run .
```

#### Solution C: Check File Watcher
```bash
# Ensure no errors in console about file watching
```

---

## Debugging Tips

### 1. Enable Debug Logging
```go
lokstra_init.BootstrapAndRun(
    lokstra_init.WithLogLevel(logger.LogLevelDebug),
)
```

### 2. Print Registered Routes
```go
lokstra_init.WithServerInitFunc(func() error {
    // Routes are auto-printed on startup
    return nil
})
```

### 3. Verify Service Registry
```go
import "github.com/primadi/lokstra/lokstra_registry"

// In ServerInitFunc
fmt.Println("Registered services:", lokstra_registry.GetAllServiceNames())
```

### 4. Check Configuration
```go
import "github.com/primadi/lokstra/lokstra_registry"

// In ServerInitFunc
cfg := lokstra_registry.GetConfig("service-definitions.db_main")
fmt.Printf("DB Config: %+v\n", cfg)
```

---

## Getting Help

If you're still stuck:

1. Check the [examples](../../../examples/) folder
2. Review [project templates](../../../project_templates/)
3. Read the [complete documentation](../../../docs/)
4. Search for similar issues in the repository

## Related Skills

- [implementation-lokstra-yaml-config](../implementation-lokstra-yaml-config/SKILL.md) - Config troubleshooting
- [implementation-lokstra-create-handler](../implementation-lokstra-create-handler/SKILL.md) - Handler issues
- [implementation-lokstra-create-service](../implementation-lokstra-create-service/SKILL.md) - Service issues
