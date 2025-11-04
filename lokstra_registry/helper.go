package lokstra_registry

import (
	"log"
	"time"
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
