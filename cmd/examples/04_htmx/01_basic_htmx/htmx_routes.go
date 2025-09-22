package main

import (
	"embed"
	"time"

	"github.com/primadi/lokstra"
)

//go:embed htmx_app
var htmxFS embed.FS

func createHtmxRoutes(app *lokstra.App) {
	// Setup HTMX file manager for template processing AND static file serving
	app.AddHtmxLayouts(htmxFS, "htmx_app/layouts").
		AddHtmxPages(htmxFS, "htmx_app/pages").
		AddHtmxStatics(htmxFS, "htmx_app/static")
	// ✨ NEW: Static routes are now automatically registered during buildRouter!
	// The framework automatically creates "/static/*" route when HtmxFsManager has static files.
	// This eliminates manual route configuration and prevents conflicts between Apps.

	// HTMX page routes using HTMXFSPage context with inline data
	createHtmxPageRoutes(app)
}

func createHtmxPageRoutes(app *lokstra.App) {
	// Home page
	app.GET("/", func(ctx *lokstra.Context) error {
		return ctx.HTMXFSPage("index", map[string]any{
			"welcome_message": "Welcome to Lokstra HTMX Demo!",
			"features": []string{
				"🚀 Fast page loading with HTMX",
				"🎨 Beautiful layouts and components",
				"📱 Responsive design",
				"⚡ Real-time updates",
			},
			"stats": map[string]any{
				"users":    1250,
				"projects": 45,
				"uptime":   "99.9%",
			},
			"current_time": time.Now().Format("2006-01-02 15:04:05"),
		}, "Home - Lokstra HTMX Demo", "Welcome to the Lokstra HTMX demo application.")
	})

	// About page
	app.GET("/about", func(ctx *lokstra.Context) error {
		return ctx.HTMXFSPage("about", map[string]any{
			"company": map[string]any{
				"name":     "Lokstra Framework",
				"founded":  2025,
				"location": "Surabaya, Indonesia",
				"mission":  "Building modern web applications with Go and HTMX",
			},
			"team": []map[string]any{
				{
					"name":   "Alice Johnson",
					"role":   "Lead Developer",
					"skills": []string{"Go", "HTMX", "Architecture"},
					"avatar": "👩‍💻",
				},
				{
					"name":   "Bob Smith",
					"role":   "Frontend Specialist",
					"skills": []string{"HTMX", "CSS", "JavaScript"},
					"avatar": "👨‍💻",
				},
				{
					"name":   "Charlie Brown",
					"role":   "DevOps Engineer",
					"skills": []string{"Docker", "Kubernetes", "CI/CD"},
					"avatar": "👨‍🔧",
				},
			},
		}, "About Us - Lokstra HTMX Demo", "Learn more about the Lokstra Framework and our team.")
	})

	// Products page
	app.GET("/products", func(ctx *lokstra.Context) error {
		return ctx.HTMXFSPage("products", map[string]any{
			"products": []map[string]any{
				{
					"id":          1,
					"name":        "Lokstra Framework",
					"description": "Modern Go web framework with HTMX integration",
					"price":       "Free",
					"category":    "Framework",
					"features":    []string{"HTMX Support", "Router", "Middleware", "Services"},
					"badge":       "🚀",
				},
				{
					"id":          2,
					"name":        "Lokstra CLI",
					"description": "Command-line tool for Lokstra development",
					"price":       "Free",
					"category":    "Tool",
					"features":    []string{"Project Generator", "Dev Server", "Build Tools"},
					"badge":       "⚡",
				},
				{
					"id":          3,
					"name":        "Lokstra Templates",
					"description": "Ready-to-use templates for common applications",
					"price":       "Free",
					"category":    "Templates",
					"features":    []string{"Admin Dashboard", "Blog", "E-commerce", "API"},
					"badge":       "🎨",
				},
			},
			"categories": []string{"Framework", "Tool", "Templates"},
		}, "Products - Lokstra HTMX Demo", "Explore the products offered by Lokstra Framework.")
	})

	// Stats page
	app.GET("/stats", func(ctx *lokstra.Context) error {
		return ctx.HTMXFSPage("stats", map[string]any{
			"metrics": map[string]any{
				"total_requests": 125450,
				"active_users":   1250,
				"response_time":  "45ms",
				"uptime":         "99.9%",
				"error_rate":     "0.1%",
			},
			"charts": []map[string]any{
				{
					"title":  "Daily Active Users",
					"type":   "line",
					"data":   []int{1100, 1150, 1200, 1250, 1300, 1250, 1200},
					"labels": []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
				},
				{
					"title":  "Response Times",
					"type":   "bar",
					"data":   []int{42, 45, 38, 50, 43, 45, 40},
					"labels": []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
				},
			},
		}, "Stats - Lokstra HTMX Demo", "View application statistics and performance metrics.")
	})

	// Featured content page
	app.GET("/featured-content", func(ctx *lokstra.Context) error {
		return ctx.HTMXFSPage("featured-content", map[string]any{
			"featured_items": []map[string]any{
				{
					"title":       "Getting Started with HTMX",
					"description": "Learn the basics of HTMX with practical examples",
					"author":      "Alice Johnson",
					"date":        "2024-01-15",
					"category":    "Tutorial",
					"duration":    "15 min read",
					"difficulty":  "Beginner",
					"tags":        []string{"HTMX", "Web Development", "Tutorial"},
				},
				{
					"title":       "Advanced Lokstra Patterns",
					"description": "Master advanced patterns for building scalable applications",
					"author":      "Bob Smith",
					"date":        "2024-01-10",
					"category":    "Guide",
					"duration":    "25 min read",
					"difficulty":  "Advanced",
					"tags":        []string{"Lokstra", "Architecture", "Best Practices"},
				},
				{
					"title":       "Real-time Features with HTMX",
					"description": "Implement live updates and real-time functionality",
					"author":      "Charlie Brown",
					"date":        "2024-01-05",
					"category":    "Tutorial",
					"duration":    "20 min read",
					"difficulty":  "Intermediate",
					"tags":        []string{"HTMX", "Real-time", "WebSockets"},
				},
			},
		}, "Featured Content - Lokstra HTMX Demo", "Explore featured articles and tutorials on HTMX and Lokstra.")
	})

	// ✨ IMPORTANT: No manual static routes needed anymore!
	//
	// Previously, you had to manually add:
	//   app.GET("/static/*", func(ctx *lokstra.Context) error {
	//       return ctx.HTMXFSPage("", nil)
	//   })
	//
	// Now the framework automatically handles this during buildRouter when:
	// 1. HtmxFsManager has static files registered (via AddHtmxStatics)
	// 2. Router is the main app router (not API groups with prefixes)
	// 3. Static path injection and serving work seamlessly across all HTMX pages
	//
	// This eliminates configuration burden and prevents route conflicts between Apps!
}
