package lokstra_registry

import (
	"log"
	"path/filepath"
	"time"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/deploy/loader"
)

// ===== LEGACY API (For Backward Compatibility) =====

// RunServerFromConfig loads configuration from specified YAML file(s) and runs the server.
func RunServerFromConfig(config ...string) {
	// 1. Load config (loads ALL deployments into Global registry)
	if err := LoadConfig(config...); err != nil {
		log.Fatal("❌ Failed to load config:", err)
	}

	server := GetConfig("server", "")
	if server == "" {
		log.Fatal("❌ 'server' not specified in config, Please add this to your config.yaml:\n" +
			"configs:\n" +
			"  server: ${SERVER}                          # mandatory, default first_server_defined\n" +
			"  shutdown_timeout: ${SHUTDOWN_TIMEOUT:30s}  # optional, default 30s")
	}

	var timeout time.Duration

	timeoutStr := GetConfig("shutdown_timeout", "30s")
	if dur, err := time.ParseDuration(timeoutStr); err == nil {
		timeout = dur
	} else {
		timeout = 30 * time.Second
	}

	// 2. Run server
	if err := RunServer(server, timeout); err != nil {
		log.Fatal("❌ Failed to run server:", err)
	}
}

// RunServerFromConfigFolder loads all YAML files from the specified folder and runs the server.
func RunServerFromConfigFolder(configFolder string) {
	// Load all YAML files in the specified config folder
	basePath := utils.GetBasePath()
	configFolder = filepath.Join(basePath, configFolder)
	files, err := filepath.Glob(filepath.Join(configFolder, "*.yaml"))
	if err != nil {
		log.Fatalf("failed to read config folder: %v", err)
	}

	if len(files) == 0 {
		log.Printf("no YAML config found in folder: %s", configFolder)
		return
	}

	lenPrefix := len(basePath) + 1
	for i, f := range files {
		files[i] = f[lenPrefix:]
	}

	// Kirim semua file ke fungsi berikutnya
	RunServerFromConfig(files...)
}

// ===== NEW API (Recommended Flow) =====

// LoadConfig loads YAML configuration file(s) and registers lazy load services.
// This makes config available for service/middleware registration.
// Config paths are relative to the project base path.
//
// Example:
//
//	if err := lokstra_registry.LoadConfig("config.yaml"); err != nil {
//	    log.Fatal(err)
//	}
//
// After calling LoadConfig, you can:
//   - Access config via GetConfig()
//   - Register services/middlewares (they can read config)
//   - Call InitAndRunServer() to start the server
func LoadConfig(configPaths ...string) error {
	if len(configPaths) == 0 {
		configPaths = []string{"config.yaml"}
	}

	// Load config (loads ALL deployments into Global registry)
	if err := loader.LoadAndBuild(configPaths); err != nil {
		return err
	}

	deploy.LogDebug("✅ Config loaded successfully from: %v", configPaths)
	return nil
}

// LoadConfigFromFolder loads all YAML files from the specified folder.
// This is a convenience wrapper around LoadConfig.
//
// Example:
//
//	if err := lokstra_registry.LoadConfigFromFolder("config"); err != nil {
//	    log.Fatal(err)
//	}
func LoadConfigFromFolder(configFolder string) error {
	// Load all YAML files in the specified config folder
	basePath := utils.GetBasePath()
	configFolder = filepath.Join(basePath, configFolder)
	files, err := filepath.Glob(filepath.Join(configFolder, "*.yaml"))
	if err != nil {
		return err
	}

	if len(files) == 0 {
		log.Printf("⚠️  No YAML config found in folder: %s", configFolder)
		return nil
	}

	lenPrefix := len(basePath) + 1
	for i, f := range files {
		files[i] = f[lenPrefix:]
	}

	return LoadConfig(files...)
}

// SetupNamedDbPools sets up database pools from loaded config.
// Must be called AFTER LoadConfig() if you use named-db-pools in config.
// Call this explicitly only if you need DB pools.
//
// Example:
//
//	if err := lokstra_registry.LoadConfig("config.yaml"); err != nil {
//	    log.Fatal(err)
//	}
//	if err := lokstra_registry.SetupNamedDbPools(); err != nil {
//	    log.Fatal(err)
//	}
func SetupNamedDbPools() error {
	return loader.SetupNamedDbPools()
}

// InitAndRunServer initializes and runs the server based on loaded config.
// Must be called after LoadConfig() and service/middleware registration.
//
// This function:
//  1. Reads server selection from config (or uses first server if not specified)
//  2. Reads shutdown timeout from config (default: 30s)
//  3. Runs the selected server
//
// Config keys used:
//   - server: Server composite key "deployment.server" (optional, uses first if not specified)
//   - shutdown_timeout: Graceful shutdown timeout duration (optional, default: 30s)
//
// Example:
//
//	if err := lokstra_registry.InitAndRunServer(); err != nil {
//	    log.Fatal(err)
//	}
func InitAndRunServer() error {
	server := GetConfig("server", "")

	var timeout time.Duration
	timeoutStr := GetConfig("shutdown_timeout", "30s")
	if dur, err := time.ParseDuration(timeoutStr); err == nil {
		timeout = dur
	} else {
		timeout = 30 * time.Second
	}

	// Run server
	return RunServer(server, timeout)
}
