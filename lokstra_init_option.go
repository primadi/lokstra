package lokstra

import (
	"github.com/primadi/lokstra/common/logger"
)

// Options pattern untuk QuickStart
type InitializeOption func(*InitializeConfig)

func WithLogLevel(level logger.LogLevel) InitializeOption {
	return func(c *InitializeConfig) { c.LogLevel = level }
}

func WithAnnotations(enable bool, paths ...string) InitializeOption {
	return func(c *InitializeConfig) {
		c.UseAnnotation = enable
		c.AnnotationScanPaths = paths
	}
}

func WithoutYAMLConfig() InitializeOption {
	return func(c *InitializeConfig) {
		c.ConfigSource = ConfigNone
		c.ConfigPath = nil
	}
}

func WithYAMLConfigPath(paths ...string) InitializeOption {
	return func(c *InitializeConfig) {
		c.ConfigSource = ConfigFromFile
		c.ConfigPath = paths
	}
}

func WithYAMLConfigFolder(folder string) InitializeOption {
	return func(c *InitializeConfig) {
		c.ConfigSource = ConfigFromFolder
		c.ConfigPath = []string{folder}
	}
}

func WithoutPgSyncMap() InitializeOption {
	return func(c *InitializeConfig) {
		c.UsePgxSyncMap = false
	}
}

func WithPgSyncMap(dbPoolName ...string) InitializeOption {
	return func(c *InitializeConfig) {
		c.UsePgxSyncMap = true
		if len(dbPoolName) == 0 {
			c.PgxSyncMapDbPoolName = "db_main"
		} else {
			if len(dbPoolName) > 1 {
				panic("WithPgSyncMap only accepts zero or one argument for dbPoolName")
			}
			c.PgxSyncMapDbPoolName = dbPoolName[0]
		}
	}
}

func WithoutPgDatabase() InitializeOption {
	return func(c *InitializeConfig) {
		c.DbPoolManagerEnable = false
	}
}

func WithPgDatabase() InitializeOption {
	return func(c *InitializeConfig) {
		c.DbPoolManagerEnable = true
	}
}

func WithoutDbMigrations() InitializeOption {
	return func(c *InitializeConfig) {
		c.CheckDbMigration = false
	}
}

func WithDbMigrations(folder string) InitializeOption {
	return func(c *InitializeConfig) {
		c.CheckDbMigration = true
		c.MigrationFolder = folder
	}
}

func WithServerInitFunc(initFunc func() error) InitializeOption {
	return func(c *InitializeConfig) {
		c.ServerInitFunc = initFunc
	}
}
func WithoutAutoRunServer() InitializeOption {
	return func(c *InitializeConfig) {
		c.InitAndRunServer = false
	}
}
