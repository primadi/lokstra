package main

import (
	"net/http"

	"github.com/primadi/lokstra"
	admin "github.com/primadi/lokstra/cmd/examples/04_htmx/02_multiapp_htmx/app1_admin"
	blog "github.com/primadi/lokstra/cmd/examples/04_htmx/02_multiapp_htmx/app2_blog"
)

// This example demonstrates multi-app HTMX integration with Lokstra.
// It shows how two different apps can run together with their own
// static files, layouts, and routing without conflicts.
//
// Learning Objectives:
// - Understand multi-app deployment patterns
// - Learn automatic static path injection with prefixes
// - Explore app isolation and conflict prevention
// - See how different apps can have different themes/styles
//
// Architecture:
// - App1: Admin Dashboard (/app1) - Professional blue theme
// - App2: Blog Platform (/app2) - Creative green theme
// - Each app has its own static files, layouts, and pages
// - Automatic static route registration prevents conflicts
//
// Documentation: https://github.com/primadi/lokstra/blob/main/docs/htmx-integration.md

func main() {
	regCtx := lokstra.NewGlobalRegistrationContext()
	app1 := lokstra.NewApp(regCtx, "admin-dashboard", ":8080")

	// Setup App1: Admin Dashboard
	admin.Setup(app1, "/app1")

	app2 := lokstra.NewApp(regCtx, "blog-platform", ":8080")
	// Setup App2: Blog Platform
	blog.Setup(app2, "/app2")

	// Optional: Root landing page
	appLanding := lokstra.NewApp(regCtx, "landing-page", ":8080")
	createLandingPage(appLanding)

	server := lokstra.NewServer(regCtx, "multiapp-htmx-server").
		AddApps(app1, app2, appLanding)

	server.Start(true)
}

// createLandingPage creates a simple landing page that showcases both apps
func createLandingPage(app *lokstra.App) {
	app.GET("/", func(ctx *lokstra.Context) error {
		html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Multi-App HTMX Demo</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 0;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .container {
            text-align: center;
            background: white;
            padding: 3rem;
            border-radius: 1rem;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            max-width: 600px;
        }
        h1 {
            color: #2d3748;
            margin-bottom: 1rem;
            font-size: 2.5rem;
        }
        p {
            color: #4a5568;
            margin-bottom: 2rem;
            font-size: 1.125rem;
        }
        .apps {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 2rem;
            margin-top: 2rem;
        }
        .app-card {
            padding: 2rem;
            border: 2px solid #e2e8f0;
            border-radius: 0.75rem;
            text-decoration: none;
            color: inherit;
            transition: all 0.3s ease;
        }
        .app-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 10px 25px rgba(0,0,0,0.1);
        }
        .admin-card {
            border-color: #3182ce;
            background: linear-gradient(135deg, #ebf8ff 0%, #bee3f8 100%);
        }
        .blog-card {
            border-color: #38a169;
            background: linear-gradient(135deg, #f0fff4 0%, #c6f6d5 100%);
        }
        .app-title {
            font-size: 1.5rem;
            font-weight: bold;
            margin-bottom: 0.5rem;
        }
        .app-desc {
            font-size: 0.875rem;
            color: #718096;
        }
        .feature-list {
            text-align: left;
            margin-top: 2rem;
            background: #f8fafc;
            padding: 1.5rem;
            border-radius: 0.5rem;
        }
        .feature-list h3 {
            margin-top: 0;
            color: #2d3748;
        }
        .feature-list ul {
            margin: 0;
            padding-left: 1.5rem;
        }
        .feature-list li {
            margin-bottom: 0.5rem;
            color: #4a5568;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Multi-App HTMX Demo</h1>
        <p>Explore two different applications running together with Lokstra framework</p>
        
        <div class="apps">
            <a href="/app1" class="app-card admin-card">
                <div class="app-title">🏢 Admin Dashboard</div>
                <div class="app-desc">Professional admin panel with analytics, user management, and system settings</div>
            </a>
            
            <a href="/app2" class="app-card blog-card">
                <div class="app-title">📝 Tech Blog</div>
                <div class="app-desc">Modern blog platform with articles, categories, and interactive features</div>
            </a>
        </div>
        
        <div class="feature-list">
            <h3>🎯 Demonstration Features</h3>
            <ul>
                <li>Multiple apps with different URL prefixes (/app1, /app2)</li>
                <li>Automatic static file routing without conflicts</li>
                <li>Separate themes and layouts for each app</li>
                <li>HTMX integration with dynamic content loading</li>
                <li>Embedded file systems for production deployment</li>
                <li>Responsive design and modern UI components</li>
            </ul>
        </div>
    </div>
</body>
</html>`

		return ctx.HTML(http.StatusOK, html)
	})
}
