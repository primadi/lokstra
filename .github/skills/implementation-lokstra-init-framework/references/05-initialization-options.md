# Lokstra Initialization Options Reference

## Available Options

### 1. WithLogLevel
Set the logging level for the application.

```go
lokstra_init.WithLogLevel(logger.LogLevelDebug)
```

**Values:**
- `logger.LogLevelDebug` - Verbose debugging information
- `logger.LogLevelInfo` - General information (default)
- `logger.LogLevelWarn` - Warning messages only
- `logger.LogLevelError` - Error messages only

---

### 2. WithAnnotations
Enable/disable annotation scanning and specify paths.

```go
lokstra_init.WithAnnotations(true, "./modules", "./services")
```

**Parameters:**
- `enable` (bool) - Enable annotation scanning (default: true)
- `paths` (variadic string) - Additional scan paths (default: current directory)

---

### 3. WithYAMLConfigPath
Configure YAML configuration loading.

```go
lokstra_init.WithYAMLConfigPath(true, "configs", "local-configs")
```

**Parameters:**
- `enable` (bool) - Enable YAML config loading (default: true)
- `paths` (variadic string) - Config file/folder paths (default: "configs")

**Note:** Multiple YAML files are auto-merged. See skill: implementation-lokstra-yaml-config

---

### 4. WithDbMigrations
Enable automatic database migrations on startup.

```go
lokstra_init.WithDbMigrations(true, "migrations")
```

**Parameters:**
- `enable` (bool) - Enable migrations (default: false)
- `folder` (string) - Migration folder path (default: "migrations")

**Behavior:**
- Runs migrations in alphabetical order
- Skips on production by default (override with SkipMigrationOnProd option)
- Uses db_main pool by default

---

### 5. WithPgSyncMap
Enable PostgreSQL-backed distributed configuration sync.

```go
lokstra_init.WithPgSyncMap(true, "db_main")
```

**Parameters:**
- `enable` (bool) - Enable PgSyncMap (default: false)
- `dbPoolName` (string) - Database pool name (default: "db_main")

**Use Case:** Multi-instance deployments where config changes need to sync across instances.

---

### 6. WithPgxSyncMapIntervals
Configure PgSyncMap heartbeat and reconnect intervals.

```go
lokstra_init.WithPgxSyncMapIntervals(
    5*time.Minute,  // heartbeat interval
    5*time.Second,  // reconnect interval
)
```

**Default Values:**
- Heartbeat: 5 minutes
- Reconnect: 5 seconds

---

### 7. WithServerInitFunc
Register a custom initialization function that runs before server starts.

```go
lokstra_init.WithServerInitFunc(func() error {
    // Custom initialization logic
    fmt.Println("Server initializing...")
    
    // Register routers, middleware, services
    registerCustomServices()
    
    return nil
})
```

**Use Cases:**
- Custom router registration
- Additional service setup
- Environment-specific configuration
- Pre-startup validation

---

### 8. WithAutoRunServer
Control whether the server automatically starts.

```go
lokstra_init.WithAutoRunServer(false)
```

**Parameters:**
- `enable` (bool) - Auto-start server (default: true)

**Use Case:** Testing environments where you want to initialize without starting the server.

---

### 9. WithPanicOnConfigError
Control error handling behavior.

```go
lokstra_init.WithPanicOnConfigError(false)
```

**Parameters:**
- `panicOnError` (bool) - Panic on config errors (default: true)

**Default:** true - Application panics on configuration errors for fail-fast behavior.

---

## Common Combinations

### Development Setup
```go
lokstra_init.BootstrapAndRun(
    lokstra_init.WithLogLevel(logger.LogLevelDebug),
    lokstra_init.WithDbMigrations(true, "migrations"),
)
```

### Production Setup
```go
lokstra_init.BootstrapAndRun(
    lokstra_init.WithLogLevel(logger.LogLevelInfo),
    lokstra_init.WithPgSyncMap(true, "db_main"),
)
```

### Testing Setup
```go
lokstra_init.BootstrapAndRun(
    lokstra_init.WithLogLevel(logger.LogLevelError),
    lokstra_init.WithAutoRunServer(false),
    lokstra_init.WithPanicOnConfigError(false),
)
```
