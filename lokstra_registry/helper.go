package lokstra_registry

import (
	"log"
	"path/filepath"
	"time"

	"github.com/primadi/lokstra/common/utils"
)

func RunServerFromConfig(config ...string) {

	if len(config) == 0 {
		config = []string{"config.yaml"}
	}

	// 1. Load config (loads ALL deployments into Global registry)
	if err := LoadAndBuild(config); err != nil {
		log.Fatal("❌ Failed to load config:", err)
	}

	server := GetConfig("server", "")
	// if server == "" {
	// 	log.Fatal("❌ 'server' not specified in config, Please add this to your config.yaml:\n" +
	// 		"configs:\n" +
	// 		"  server: ${SERVER}                          # mandatory, default first_server_defined\n" +
	// 		"  shutdown_timeout: ${SHUTDOWN_TIMEOUT:30s}  # optional, default 30s")
	// }

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
