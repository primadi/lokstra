package main

import (
	"net/http"

	"github.com/primadi/lokstra/common/logger"
)

func main() {
	// Initialize the router with a name
	router := setupRouter()

	// Start the HTTP server
	logger.LogInfo("Starting server on :3000")

	router.PrintRoutes()
	if err := http.ListenAndServe(":3000", router); err != nil {
		logger.LogPanic("Server failed to start: %v", err)
	}
}
