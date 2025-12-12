package lokstra

import (
	"fmt"

	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/services/sync_config_pg"
)

type ConfigSourceType int

const (
	ConfigNone ConfigSourceType = iota
	ConfigFromFile
	ConfigFromFolder
)

type InitializeConfig struct {
	// 1. SetLogLevel
	LogLevel logger.LogLevel

	// 2.Bootstrap
	UseAnnotation       bool
	AnnotationScanPaths []string

	// 3. LoadConfig
	ConfigSource ConfigSourceType
	ConfigPath   []string

	// 4. UsePgxSyncMap
	// If true, You have to use DbPoolManager also
	UsePgxSyncMap        bool
	PgxSyncMapDbPoolName string

	// 5. DbPoolManagerEnable
	DbPoolManagerEnable bool

	// 6. CheckDbMigrations
	CheckDbMigration bool
	MigrationFolder  string

	// 7. ServerInit Func
	ServerInitFunc func() error

	// 8. Init and Run Server
	InitAndRunServer bool
}

func BootstrapAndRun(opts ...InitializeOption) error {
	cfg := &InitializeConfig{
		LogLevel:             logger.LogLevelInfo,
		UseAnnotation:        true, // Auto-detect @RouterService
		ConfigSource:         ConfigFromFolder,
		DbPoolManagerEnable:  true, // Auto if config has database.*
		UsePgxSyncMap:        true,
		PgxSyncMapDbPoolName: "db_main",
		CheckDbMigration:     true,
		MigrationFolder:      "migrations",
		InitAndRunServer:     true,
	}

	// Apply options
	for _, opt := range opts {
		opt(cfg)
	}

	return BootstrapAndRunWithConfig(cfg)
}

// Initialize lokstra framework with given config
func BootstrapAndRunWithConfig(cfg *InitializeConfig) error {
	if !cfg.DbPoolManagerEnable {
		if cfg.UsePgxSyncMap {
			return fmt.Errorf("PgxSyncMap requires DbPoolManager to be enabled")
		}
		if cfg.CheckDbMigration {
			return fmt.Errorf("DB Migration check requires DbPoolManager to be enabled")
		}
	}

	// 1. Set log level
	logger.SetLogLevel(cfg.LogLevel)

	// 2. Bootstrap
	if cfg.UseAnnotation {
		Bootstrap(cfg.AnnotationScanPaths...)
	}

	// 3. LoadConfig
	switch cfg.ConfigSource {
	case ConfigFromFile:
		if len(cfg.ConfigPath) == 0 {
			cfg.ConfigPath = []string{"config.yaml"}
		}
		if err := LoadConfig(cfg.ConfigPath...); err != nil {
			return err
		}
	case ConfigFromFolder:
		lenConfig := len(cfg.ConfigPath)
		if lenConfig == 0 {
			cfg.ConfigPath = []string{"config"}
		} else if lenConfig > 1 {
			return fmt.Errorf("ConfigPath for ConfigFromFolder should contain only one folder path")
		}

		if err := LoadConfigFromFolder(cfg.ConfigPath[0]); err != nil {
			return err
		}
	}

	// 4. Use Pgx SyncMap
	if cfg.UsePgxSyncMap {
		if len(cfg.PgxSyncMapDbPoolName) == 0 {
			return fmt.Errorf("PgxSyncMapDbPoolName must be set when UsePgxSyncMap is true")
		}
		sync_config_pg.Register(cfg.PgxSyncMapDbPoolName)
	}

	// 5. PgxDbPoolManager
	if cfg.DbPoolManagerEnable {
		UsePgxDbPoolManager(cfg.UsePgxSyncMap)

		if err := LoadNamedDbPoolsFromConfig(); err != nil {
			return err
		}
	}

	// 6. Check DB Migrations
	if cfg.CheckDbMigration {
		if err := CheckDbMigrationsAuto(cfg.MigrationFolder); err != nil {
			return err
		}
	}

	// 7. Server Init Func
	if cfg.ServerInitFunc != nil {
		if err := cfg.ServerInitFunc(); err != nil {
			return err
		}
	}

	// 8. Init and Run Server
	if cfg.InitAndRunServer {
		if err := RunConfiguredServer(); err != nil {
			return err
		}
	}

	return nil
}
