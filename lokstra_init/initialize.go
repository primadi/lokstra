package lokstra_init

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/core/deploy/loader"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/services/sync_config_pg"
)

type InitializeConfig struct {
	// If true, panic on configuration error
	PanicOnConfigError bool

	// 1. SetLogLevel
	LogLevel logger.LogLevel

	// 2.Bootstrap
	EnableAnnotation    bool
	AnnotationScanPaths []string

	// 3. LoadConfig
	EnableLoadConfig bool
	ConfigPath       []string

	// 4. EnablePgxSyncMap
	//    If true, You have to use DbPoolManager also
	EnablePgxSyncMap         bool
	PgxSyncMapDbPoolName     string
	PgxSyncHeartbeatInterval time.Duration
	PgxSyncReconnectInterval time.Duration

	// 5. EnableDbMigration
	EnableDbMigration   bool
	MigrationFolder     string
	SkipMigrationOnProd bool

	// 6. ServerInit Func
	ServerInitFunc func() error

	// 7. Init and Run Server
	IsRunServer bool
}

func (c *InitializeConfig) returnError(err error) error {
	if c.PanicOnConfigError && err != nil {
		panic(err)
	}

	return err
}

// Initialize lokstra framework with options
// defaults:
//   - PanicOnConfigError=true,
//   - LogLevel=Info,
//   - EnableLoadConfig=true,
//   - EnableAnnotation=true,
//   - IsRunServer=true,
//
// Db Migration Options:
//   - EnableDbMigration=false,
//   - SkipMigrationOnProd=true,
//   - MigrationFolder="migrations"
//
// PgxSyncMap Options:
//   - EnablePgxSyncMap=false,
//   - PgxSyncMapDbPoolName="db_main",
//   - PgxSyncHeartbeatInterval=5min,
//   - PgxSyncReconnectInterval=5s,
func BootstrapAndRun(opts ...InitializeOption) error {
	cfg := &InitializeConfig{
		PanicOnConfigError:       true,
		LogLevel:                 logger.LogLevelInfo,
		EnableLoadConfig:         true,
		EnableAnnotation:         true, // Auto-detect @EndpointService
		EnablePgxSyncMap:         false,
		SkipMigrationOnProd:      true,
		PgxSyncMapDbPoolName:     "db_main",
		PgxSyncHeartbeatInterval: 5 * time.Minute,
		PgxSyncReconnectInterval: 5 * time.Second,
		EnableDbMigration:        false,
		MigrationFolder:          "migrations",
		IsRunServer:              true,
	}

	// Apply options
	for _, opt := range opts {
		opt(cfg)
	}

	return BootstrapAndRunWithConfig(cfg)
}

// Initialize lokstra framework with given config
func BootstrapAndRunWithConfig(cfg *InitializeConfig) error {
	// Validate config
	if cfg.EnablePgxSyncMap {
		if len(cfg.PgxSyncMapDbPoolName) == 0 {
			return cfg.returnError(fmt.Errorf("PgxSyncMapDbPoolName must be set when UsePgxSyncMap is true"))
		}
	}

	// 1. Set log level
	logger.SetLogLevel(cfg.LogLevel)

	// 2. Bootstrap
	if cfg.EnableAnnotation {
		Bootstrap(cfg.AnnotationScanPaths...)
	}

	// 3. LoadConfig
	if cfg.EnableLoadConfig {
		if _, err := loader.LoadConfig(cfg.ConfigPath...); err != nil {
			return cfg.returnError(err)
		}
	}

	// 4. Use Pgx SyncMap (MUST be before DbPoolManager if IsDbPoolAutoSync=true)
	if cfg.EnablePgxSyncMap {
		if len(cfg.PgxSyncMapDbPoolName) == 0 {
			return cfg.returnError(fmt.Errorf(
				"PgxSyncMapDbPoolName must be set when UsePgxSyncMap is true"))
		}
		sync_config_pg.Register(cfg.PgxSyncMapDbPoolName,
			cfg.PgxSyncHeartbeatInterval, cfg.PgxSyncReconnectInterval)
	}

	// 5. Check DB Migrations
	if cfg.EnableDbMigration {
		if mode := GetRuntimeMode(); mode != "prod" || !cfg.SkipMigrationOnProd {
			if err := CheckDbMigrationsAuto(cfg.MigrationFolder); err != nil {
				return cfg.returnError(err)
			}
		}
	}

	// 6. Server Init Func
	if cfg.ServerInitFunc != nil {
		if err := cfg.ServerInitFunc(); err != nil {
			return cfg.returnError(err)
		}
	}

	// 7. Init and Run Server
	if cfg.IsRunServer {
		if err := lokstra_registry.RunConfiguredServer(); err != nil {
			return cfg.returnError(err)
		}
	}

	return nil
}
