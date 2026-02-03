# Lokstra Framework Startup Flow

## Complete Bootstrap Sequence

```
┌─────────────────────────────────────────────────────────────┐
│ 1. main() starts                                            │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 2. Middleware Registration                                  │
│    - recovery.Register()                                    │
│    - request_logger.Register()                              │
│    - cors.Register()                                        │
│    - Custom middleware.Register()                           │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 3. Service Registration                                     │
│    - dbpool_pg.Register()                                   │
│    - eventbus.Register()                                    │
│    - kvstore.Register()                                     │
│    - Custom services.Register()                             │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 4. lokstra_init.BootstrapAndRun()                          │
└─────────────────────────────────────────────────────────────┘
                            ↓
        ┌───────────────────────────────────────────┐
        │ 4.1 Environment Detection                 │
        │     - Detect: prod / dev / debug mode     │
        │     - Check: --generate-only flag         │
        └───────────────────────────────────────────┘
                            ↓
        ┌───────────────────────────────────────────┐
        │ 4.2 Code Generation (dev/debug only)      │
        │     - Scan @Handler annotations           │
        │     - Scan @Service annotations           │
        │     - Generate zz_generated.lokstra.go    │
        │     - Generate route registration code    │
        │     - Auto-restart if code changed        │
        └───────────────────────────────────────────┘
                            ↓
        ┌───────────────────────────────────────────┐
        │ 4.3 Configuration Loading                 │
        │     - Load configs/*.yaml                 │
        │     - Merge multiple YAML files           │
        │     - Substitute ${ENV_VAR} values        │
        │     - Load service-definitions            │
        │     - Load deployments                    │
        └───────────────────────────────────────────┘
                            ↓
        ┌───────────────────────────────────────────┐
        │ 4.4 Database Migrations (optional)        │
        │     - Run .up.sql files in order          │
        │     - Track migrations in DB              │
        │     - Skip on production (configurable)   │
        └───────────────────────────────────────────┘
                            ↓
        ┌───────────────────────────────────────────┐
        │ 4.5 PgSyncMap Initialization (optional)   │
        │     - Connect to sync DB                  │
        │     - Setup heartbeat mechanism           │
        │     - Enable distributed config sync      │
        └───────────────────────────────────────────┘
                            ↓
        ┌───────────────────────────────────────────┐
        │ 4.6 Service Creation                      │
        │     - Resolve dependency graph            │
        │     - Create services in order            │
        │     - Inject dependencies                 │
        │     - Call service factory functions      │
        └───────────────────────────────────────────┘
                            ↓
        ┌───────────────────────────────────────────┐
        │ 4.7 Handler Creation                      │
        │     - Create @Handler instances           │
        │     - Inject @Service dependencies        │
        │     - Inject config values                │
        └───────────────────────────────────────────┘
                            ↓
        ┌───────────────────────────────────────────┐
        │ 4.8 ServerInitFunc (optional)             │
        │     - Run custom initialization hook      │
        │     - Register routers                    │
        │     - Additional service setup            │
        └───────────────────────────────────────────┘
                            ↓
        ┌───────────────────────────────────────────┐
        │ 4.9 Route Mounting                        │
        │     - Mount @Route endpoints              │
        │     - Apply middleware chains             │
        │     - Configure route handlers            │
        └───────────────────────────────────────────┘
                            ↓
        ┌───────────────────────────────────────────┐
        │ 4.10 Server Start                         │
        │      - Bind to configured address         │
        │      - Start HTTP listener                │
        │      - Print startup banner               │
        │      - Log registered routes              │
        └───────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 5. Server Running - Ready to Accept Requests               │
└─────────────────────────────────────────────────────────────┘
```

## Environment Modes

### Production Mode (`prod`)
- No code generation
- No auto-restart
- Optimized for performance
- Uses pre-generated code

### Development Mode (`dev`)
- Auto code generation on changes
- Auto-restart with `go run`
- File watching enabled
- Hot reload support

### Debug Mode (`debug`)
- Same as dev mode
- Auto-restart with Delve debugger
- Breakpoint support
- Step-through debugging

## Special Flags

### --generate-only
Force code generation and exit without starting server.

```bash
go run . --generate-only
```

**Use Cases:**
- CI/CD pipelines
- Pre-commit hooks
- Manual code generation
- Debugging annotation issues

## Code Generation Details

When code changes are detected:

1. **Scan Phase**
   - Find all files with @Handler annotations
   - Find all files with @Service annotations
   - Parse annotation parameters

2. **Generation Phase**
   - Generate `zz_generated.lokstra.go` per package
   - Generate `zz_lokstra_imports.go` at project root (auto-import all modules)
   - Generate route registration functions
   - Generate service factory functions
   - Generate dependency injection code

3. **Import Phase**
   - Load `zz_lokstra_imports.go` (contains all module imports)
   - Ensure all handlers are registered
   - No manual imports needed!

4. **Restart Phase** (if code changed)
   - Kill current process
   - Restart with same arguments
   - Preserve environment variables

## Configuration Loading

YAML files are loaded in this order:

1. **config.yaml** (root)
2. **configs/*.yaml** (all files, merged)
3. **Environment-specific overrides**

### Merge Strategy

```yaml
# configs/base.yaml
database:
  host: localhost
  port: 5432

# configs/production.yaml
database:
  host: prod-db.example.com
  # port: 5432 is inherited

# Result: host is overridden, port remains
```

## Service Creation Order

Services are created based on dependency graph:

```
db_main (no dependencies)
  ↓
user_repository (depends on: db_main)
  ↓
user_handler (depends on: user_repository)
  ↓
Server starts
```

**Circular dependencies** are detected and cause panic at startup.
