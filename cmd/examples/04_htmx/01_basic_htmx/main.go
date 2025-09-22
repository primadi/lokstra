package main

import (
	"github.com/primadi/lokstra"
)

// This example demonstrates basic HTMX integration with Lokstra.
// It shows how to set up HTMX pages, layouts, and page data handlers
// for building dynamic web applications without complex JavaScript.
//
// Learning Objectives:
// - Understand HTMX integration basics
// - Learn layout and page template system
// - Explore page data handlers
// - See HTMX navigation patterns
//
// Documentation: https://github.com/primadi/lokstra/blob/main/docs/htmx-integration.md

func main() {
	regCtx := lokstra.NewGlobalRegistrationContext()
	app := lokstra.NewApp(regCtx, "htmx-basic-app", ":8080")

	createHtmxRoutes(app)
	createApiRoutes(app)

	app.Start(true)
}
