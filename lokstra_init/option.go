package lokstra_init

import (
	"time"

	"github.com/primadi/lokstra/common/logger"
)

// Options pattern untuk QuickStart
type InitializeOption func(*InitializeConfig)

// set PanicOnConfigError to true to panic on configuration error
// default is true
func WithPanicOnConfigError(panicOnError bool) InitializeOption {
	return func(c *InitializeConfig) {
		c.PanicOnConfigError = panicOnError
	}
}

// set log level for lokstra logger
// default is logger.InfoLevel
func WithLogLevel(level logger.LogLevel) InitializeOption {
	return func(c *InitializeConfig) { c.LogLevel = level }
}

// enable annotations with optional scan paths
// if no paths provided, use default paths
// example annotations : @RouterService, @Service, @Route
// default enable is true, path is empty (current folder)
func WithAnnotations(enable bool, paths ...string) InitializeOption {
	return func(c *InitializeConfig) {
		c.EnableAnnotation = enable
		c.AnnotationScanPaths = paths
	}
}

// enable loading configuration from YAML files at the given paths
// paths can be files or folders
func WithYAMLConfigPath(enable bool, paths ...string) InitializeOption {
	return func(c *InitializeConfig) {
		c.EnableLoadConfig = enable
		c.ConfigPath = paths
	}
}

// enable PgxSyncMap with the given db pool name
// default enable is false
func WithPgSyncMap(enable bool, dbPoolName string) InitializeOption {
	return func(c *InitializeConfig) {
		c.EnablePgxSyncMap = enable
		c.PgxSyncMapDbPoolName = dbPoolName
	}
}

// set PgxSyncMap heartbeat and reconnect intervals
// default heartbeat is 5 minutes, reconnect is 5 seconds
func WithPgxSyncMapIntervals(heartBeatInterval, reconnectInterval time.Duration) InitializeOption {
	return func(c *InitializeConfig) {
		c.PgxSyncHeartbeatInterval = heartBeatInterval
		c.PgxSyncReconnectInterval = reconnectInterval
	}
}

// enable or disable database pool auto synchronization
// default enable is false
func WithDbPoolAutoSync(enable bool) InitializeOption {
	return func(c *InitializeConfig) {
		c.IsDbPoolAutoSync = enable
	}
}

// enable or disable database migrations with the given migration folder
// default enable is false
func WithDbMigrations(enable bool, folder string) InitializeOption {
	return func(c *InitializeConfig) {
		c.EnableDbMigration = enable
		c.MigrationFolder = folder
	}
}

// set server initialization function
func WithServerInitFunc(initFunc func() error) InitializeOption {
	return func(c *InitializeConfig) {
		c.ServerInitFunc = initFunc
	}
}

// enable or disable automatic running of the server after initialization
// default is true
func WithAutoRunServer(enable bool) InitializeOption {
	return func(c *InitializeConfig) {
		c.IsRunServer = enable
	}
}
