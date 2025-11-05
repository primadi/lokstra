package mainapp

import (
	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/project_templates/01_router/03_multi_app/shared"
)

// CreateApp creates the main public-facing API application
func CreateApp() *lokstra.App {
	// API router with business logic
	apiRouter := SetupMainAPIRouter()

	// Health check router (shared concern)
	healthRouter := shared.SetupHealthRouter("main")

	// Create app on port 3000
	app := lokstra.NewApp("main-app", ":3000", healthRouter, apiRouter)

	return app
}
