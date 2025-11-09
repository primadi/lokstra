package adminapp

import (
	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/project_templates/01_router/03_multi_app/shared"
)

// CreateApp creates the admin API application
func CreateApp() *lokstra.App {
	// Admin router with admin-specific logic
	adminRouter := SetupAdminAPIRouter()

	// Health check router (shared concern)
	healthRouter := shared.SetupHealthRouter("admin")

	// Create app on port 3001
	app := lokstra.NewApp("admin-app", ":3001", healthRouter, adminRouter)

	return app
}
