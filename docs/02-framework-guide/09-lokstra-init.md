---
title: Lokstra Initialization (lokstra_init)
layout: default
parent: Framework Guide
nav_order: 9
---

# Lokstra Initialization (lokstra_init)

Lokstra provides a convenient initialization package (`lokstra_init`) that handles the complex setup process in the correct order. Instead of manually calling multiple initialization functions, use `lokstra_init` to bootstrap your application with sensible defaults.

## Quick Start

The simplest way to initialize Lokstra:

```go
package main

import "github.com/primadi/lokstra/lokstra_init"

func main() {
    // Initialize and run with defaults
    // Default: WithPanicOnConfigError(true) - panics on error
    lokstra_init.BootstrapAndRun()
}
```

This single call handles:
1. ✅ Annotation scanning and code generation
2. ✅ Loading YAML configuration
3. ✅ Setting up database pool manager
4. ✅ Running migrations (if enabled)
5. ✅ Starting the server

## Why Use lokstra_init?

### Without lokstra_init (Manual Setup)

```go
func main() {
    // 1. Set log level
    logger.SetLogLevel(logger.LogLevelInfo)
    
    // 2. Bootstrap annotations
    lokstra.Bootstrap("./services")
    
    // 3. Load config
    if err := lokstra_registry.LoadConfig("config.yaml"); err != nil {
        log.Fatal(err)
    }
    
    // 4. Register sync-config (if needed)
    sync_config_pg.Register("db_main", 5*time.Minute, 5*time.Second)
    
    // 5. Setup dbpool-manager
    lokstra_init.UsePgxDbPoolManager(false)
    
    // 6. Load pools from config
    if err := loader.LoadDbPoolManagerFromConfig(); err != nil {
        log.Fatal(err)
    }
    
    // 7. Check migrations (if needed)
    if err := lokstra_init.CheckDbMigrationsAuto("migrations"); err != nil {
        log.Fatal(err)
    }
    
    // 8. Run server
    if err := lokstra_registry.RunConfiguredServer(); err != nil {
        log.Fatal(err)
    }
}
```

**Problems:**
- ❌ Need to understand correct initialization order
- ❌ Easy to miss a step or get order wrong
- ❌ Lots of boilerplate code
- ❌ Manual error handling for each step

### With lokstra_init (Recommended)

```go
func main() {
    // Everything handled in correct order
    lokstra_init.BootstrapAndRun()
}
```

**Benefits:**
- ✅ Correct initialization order guaranteed
- ✅ Minimal code
- ✅ Sensible defaults
- ✅ Easy to customize with options

## Basic Usage

### Default Configuration

```go
import "github.com/primadi/lokstra/lokstra_init"

func main() {
    // Uses sensible defaults:
    // - Log level: Info
    // - Annotations: Enabled
    // - Config loading: Enabled
    // - DbPoolManager: Enabled (local mode)
    // - Migrations: Disabled
    // - Auto-run server: Enabled
    // - Panic on error: true (default)
    lokstra_init.BootstrapAndRun()
}
```

### Custom Configuration with Options

Use the options pattern to customize initialization:

```go
import (
    "github.com/primadi/lokstra/lokstra_init"
    "github.com/primadi/lokstra/common/logger"
    "time"
)

func main() {
    // Default: WithPanicOnConfigError(true) - panics on error
    lokstra_init.BootstrapAndRun(
        // Set log level to debug
        lokstra_init.WithLogLevel(logger.LogLevelDebug),
        
        // Enable annotations with custom paths
        lokstra_init.WithAnnotations(true, "./services", "./modules"),
        
        // Load config from multiple files
        lokstra_init.WithYAMLConfigPath(true, "config.yaml", "config.local.yaml"),
        
        // Enable database pool manager with sync mode
        lokstra_init.WithDbPoolManager(true, true), // enable, useSync
        
        // Enable sync-config with custom pool name
        lokstra_init.WithPgSyncMap(true, "db_main"),
        
        // Set sync-config intervals
        lokstra_init.WithPgxSyncMapIntervals(10*time.Minute, 10*time.Second),
        
        // Enable migrations
        lokstra_init.WithDbMigrations(true, "migrations"),
        
        // Custom server initialization
        lokstra_init.WithServerInitFunc(func() error {
            // Your custom initialization code
            return nil
        }),
        
        // Don't auto-run server (initialize only)
        lokstra_init.WithAutoRunServer(false),
    )
}
```

## Initialization Steps

`lokstra_init.BootstrapAndRun()` executes the following steps in order:

### 1. Set Log Level
```go
logger.SetLogLevel(cfg.LogLevel) // Default: Info
```

### 2. Bootstrap Annotations
```go
if cfg.EnableAnnotation {
    lokstra.Bootstrap(cfg.AnnotationScanPaths...)
}
```
- Scans for `@Handler`, `@Service`, `@Route` annotations
- Generates router code if needed
- Auto-registers services

### 3. Load Configuration
```go
if cfg.EnableLoadConfig {
    loader.LoadConfig(cfg.ConfigPath...)
}
```
- Loads YAML configuration files
- Resolves environment variables
- Validates configuration structure

### 4. Setup Sync-Config (if enabled)
```go
if cfg.EnablePgxSyncMap {
    sync_config_pg.Register(
        cfg.PgxSyncMapDbPoolName,
        cfg.PgxSyncHeartbeatInterval,
        cfg.PgxSyncReconnectInterval,
    )
}
```
- **Important:** Must be before DbPoolManager if using sync mode
- Registers PostgreSQL-based distributed configuration

### 5. Setup DbPoolManager
```go
if cfg.EnableDbPoolManager {
    lokstra_init.UsePgxDbPoolManager(cfg.IsDbPoolAutoSync)
    loader.LoadDbPoolManagerFromConfig()
}
```
- Creates `dbpool-manager` service
- Loads named pools from `dbpool-definitions:` section in config
- Auto-registers pools as services

### 6. Check Migrations (if enabled)
```go
if cfg.EnableDbMigration {
    mode := lokstra_init.GetRuntimeMode()
    if mode != "prod" || !cfg.SkipMigrationOnProd {
        lokstra_init.CheckDbMigrationsAuto(cfg.MigrationFolder)
    }
}
```
- Validates database migrations
- Skips in production (unless configured otherwise)

### 7. Server Initialization
```go
if cfg.ServerInitFunc != nil {
    cfg.ServerInitFunc()
}
```
- Executes custom initialization code
- Useful for setup that requires all services to be ready

### 8. Run Server (if enabled)
```go
if cfg.IsRunServer {
    lokstra_registry.RunConfiguredServer()
}
```
- Starts the HTTP server
- Blocks until server stops

## Common Patterns

### Pattern 1: Basic Application

```go
func main() {
    // Simple initialization with defaults
    // Default: WithPanicOnConfigError(true) - panics on error
    lokstra_init.BootstrapAndRun()
}
```

**Use when:**
- Single instance deployment
- Standard configuration
- No migrations needed

### Pattern 2: Multi-Instance with Sync

```go
func main() {
    // Default: WithPanicOnConfigError(true) - panics on error
    lokstra_init.BootstrapAndRun(
        lokstra_init.WithDbPoolManager(true, true), // enable, useSync
        lokstra_init.WithPgSyncMap(true, "db_main"),
    )
}
```

**Use when:**
- Multiple application instances
- Need shared pool configurations
- Using distributed sync-config

**Config required:**
```yaml
configs:
  dbpool-definitions:
    use_sync: true

dbpool-definitions:
  db_main:
    dsn: "postgres://localhost/mydb"
```

### Pattern 3: Initialize Only (No Server)

```go
func main() {
    // Default: WithPanicOnConfigError(true) - panics on error
    lokstra_init.BootstrapAndRun(
        lokstra_init.WithAutoRunServer(false),
        lokstra_init.WithServerInitFunc(func() error {
            // Run migrations, seed data, etc.
            return performSetup()
        }),
    )
    
    // Manual server control
    // lokstra_registry.RunConfiguredServer()
}
```

**Use when:**
- Need custom server lifecycle
- Running setup scripts
- Testing initialization

### Pattern 4: Development with Migrations

```go
func main() {
    // Default: WithPanicOnConfigError(true) - panics on error
    lokstra_init.BootstrapAndRun(
        lokstra_init.WithLogLevel(logger.LogLevelDebug),
        lokstra_init.WithDbMigrations(true, "migrations"),
        lokstra_init.WithDbPoolManager(true, false), // local mode
    )
}
```

**Use when:**
- Development environment
- Need migration validation
- Local database only

## Available Options

### WithLogLevel(level logger.LogLevel)
Set log level for Lokstra logger.

```go
lokstra_init.WithLogLevel(logger.LogLevelDebug)
```

### WithAnnotations(enable bool, paths ...string)
Enable/disable annotation scanning with optional paths.

```go
lokstra_init.WithAnnotations(true, "./services", "./modules")
```

### WithYAMLConfigPath(enable bool, paths ...string)
Enable/disable config loading with file/folder paths.

```go
lokstra_init.WithYAMLConfigPath(true, "config.yaml", "config.local.yaml")
```

### WithDbPoolManager(enable bool, isDbPoolAutoSync bool)
Enable database pool manager.

- `enable`: Enable dbpool-manager service
- `isDbPoolAutoSync`: Use sync mode (requires sync-config)

```go
lokstra_init.WithDbPoolManager(true, true) // enable, useSync
```

### WithPgSyncMap(enable bool, dbPoolName string)
Enable PostgreSQL-based distributed sync-config.

```go
lokstra_init.WithPgSyncMap(true, "db_main")
```

### WithPgxSyncMapIntervals(heartbeat, reconnect time.Duration)
Set sync-config heartbeat and reconnect intervals.

```go
lokstra_init.WithPgxSyncMapIntervals(10*time.Minute, 10*time.Second)
```

### WithDbMigrations(enable bool, folder string)
Enable database migration validation.

```go
lokstra_init.WithDbMigrations(true, "migrations")
```

### WithServerInitFunc(initFunc func() error)
Set custom server initialization function.

```go
lokstra_init.WithServerInitFunc(func() error {
    // Your initialization code
    return nil
})
```

### WithAutoRunServer(enable bool)
Enable/disable automatic server startup.

```go
lokstra_init.WithAutoRunServer(false) // Initialize only, don't run server
```

### WithPanicOnConfigError(panicOnError bool)
Control error handling behavior.

**Default:** `true` (panics on error)

```go
// Default behavior (panic on error)
lokstra_init.BootstrapAndRun() // Will panic if config error occurs

// Return error instead of panic
err := lokstra_init.BootstrapAndRun(
    lokstra_init.WithPanicOnConfigError(false),
)
if err != nil {
    log.Fatal(err)
}
```

## Initialization Order

The initialization order is **critical** and `lokstra_init` ensures correct sequence:

```
1. Set Log Level
   ↓
2. Bootstrap Annotations
   ↓
3. Load Configuration
   ↓
4. Setup Sync-Config (if enabled)
   ↓
5. Setup DbPoolManager (requires sync-config if using sync mode)
   ↓
6. Check Migrations (requires DbPoolManager)
   ↓
7. Server Init Func
   ↓
8. Run Server (if enabled)
```

**Important Notes:**
- Sync-Config **must** be before DbPoolManager if using sync mode
- DbPoolManager **must** be before Migrations
- Configuration **must** be loaded before services that depend on it

## Error Handling

**Default behavior:** `BootstrapAndRun` panics on configuration errors (`WithPanicOnConfigError(true)` by default). This is the recommended approach for most applications.

If you need to handle errors manually (e.g., for testing or custom error handling):

```go
err := lokstra_init.BootstrapAndRun(
    lokstra_init.WithPanicOnConfigError(false), // Return error instead of panic
)
if err != nil {
    log.Fatal(err)
}
```

## Runtime Mode

Lokstra detects runtime mode from environment:

- `LOKSTRA_MODE=prod` → Production mode
- `LOKSTRA_MODE=dev` → Development mode  
- `LOKSTRA_MODE=debug` → Debug mode
- Default → Development mode

Use `lokstra_init.GetRuntimeMode()` to check current mode:

```go
mode := lokstra_init.GetRuntimeMode()
if mode == lokstra_init.RunModeProd {
    // Production-specific logic
}
```

## See Also

- [Database Pools](./08-database-pools.md) - Database pool configuration
- [Configuration](./04-config/index.md) - YAML configuration guide
- [Service Registration](./02-service/index.md) - Service setup
- [DbPool Manager API](../03-api-reference/06-services/dbpool-manager.md) - API details