package main

import (
	"embed"

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

//go:embed htmx_app
var htmxFS embed.FS

func main() {
	regCtx := lokstra.NewGlobalRegistrationContext()
	app := lokstra.NewApp(regCtx, "htmx-basic-app", ":8080")

	createHtmxAndStaticRoutes(app)
	createPageDataRoutes(app)
	createApiRoutes(app)

	app.Start()
}

// HTMX Integration Key Concepts:
//
// 1. File Structure:
//    - htmx_app/layouts/    - Layout templates
//    - htmx_app/pages/      - Page templates
//    - htmx_app/static/     - Static assets (CSS, JS, images)
//
// 2. Mounting HTMX:
//    - app.MountHtmx() mounts HTMX pages and routing
//    - app.MountStatic() mounts static assets
//    - Script injection adds automatic functionality
//
// 3. Page Data:
//    - ctx.HtmxPageData() provides structured page data
//    - Includes title, description, and custom data
//    - Available in templates via template variables
//
// 4. Dynamic Content:
//    - Regular API endpoints for HTMX requests
//    - JSON responses for partial page updates
//    - Form handling with HTMX
//
// 5. Navigation:
//    - HTMX handles navigation between pages
//    - Partial page updates for better performance
//    - URL state management with history API

// Test Commands:
//
// # Start the server
// go run main.go
//
// # Access the web application
// Open browser: http://localhost:8080
//
// # Test page navigation (HTMX will handle this smoothly)
// Click on navigation links in the web interface
//
// # Test dynamic content loading
// Click "Load Featured Content" button on home page
// Click "Load Stats" button on home page
//
// # Test search functionality
// Type in the search box and see live results
//
// # Test contact form
// Fill out and submit the contact form
//
// # API endpoints can also be tested directly:
// curl http://localhost:8080/api/featured-content
// curl http://localhost:8080/api/stats
// curl http://localhost:8080/api/search?q=htmx
// curl -X POST http://localhost:8080/api/contact -d "name=Test&email=test@example.com&message=Hello"
