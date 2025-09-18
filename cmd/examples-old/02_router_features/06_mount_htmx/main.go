package main

import (
	"embed"
	"io/fs"
	"os"
	"time"

	"github.com/primadi/lokstra"
)

//go:embed admin_app
var embedAdminApp embed.FS

//go:embed web_app
var embedWebApp embed.FS

//go:embed static
var embedStaticFiles embed.FS

type assetFS struct {
	adminApp    fs.FS
	webApp      fs.FS
	staticFiles fs.FS
}

func initAssetFS(embed bool) *assetFS {
	if embed {
		fsAdminApp, _ := fs.Sub(embedAdminApp, "admin_app")
		fsWebApp, _ := fs.Sub(embedWebApp, "web_app")
		fsStatic, _ := fs.Sub(embedStaticFiles, "static")

		return &assetFS{
			adminApp:    fsAdminApp,
			webApp:      fsWebApp,
			staticFiles: fsStatic,
		}
	}

	return &assetFS{
		adminApp:    os.DirFS("./admin_app"),
		webApp:      os.DirFS("./web_app"),
		staticFiles: os.DirFS("./static"),
	}
}

func main() {
	// 1. Create global registration context
	regCtx := lokstra.NewGlobalRegistrationContext()

	// 2. Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 3. Create application
	app := lokstra.NewApp(regCtx, "htmx-demo", ":"+port)

	assetFs := initAssetFS(false)

	// 4. Mount Htmx App
	app.MountHtmx("/", nil, assetFs.webApp)

	// 5. Mount Admin App
	app.MountHtmx("/admin", nil, assetFs.adminApp)

	// 6. Mount static files (for assets like CSS, JS, images)
	app.MountStatic("/static", false, assetFs.staticFiles)

	// 6. Register page_data to serve page data for web app
	registerPageData(app)

	// 7. Register API endpoints
	registerApi(app)

	// 8. Start App
	if err := app.Start(); err != nil {
		lokstra.Logger.Fatalf("Failed to start app: %v", err)
	}
}

func registerPageData(app *lokstra.App) {
	// Page Data API endpoints - these provide dynamic data for HTMX pages
	// The HTMX handler will call these internally via /page-data/* routes

	pageDataGroup := app.Group("/page-data")
	pageDataGroup.GET("/", func(ctx *lokstra.Context) error {
		return ctx.HtmxPageData("Home Page", "", map[string]any{
			"message":   "Welcome to Lokstra HTMX Demo",
			"timestamp": time.Now(),
			"features": []string{
				"HTMX page serving with layouts",
				"Static asset fallback",
				"Partial rendering support",
				"Template-based rendering",
			},
		})
	})

	pageDataGroup.GET("/about", func(ctx *lokstra.Context) error {
		return ctx.HtmxPageData("About Us",
			"This is the about page with dynamic content", map[string]any{
				"team": []map[string]string{
					{"name": "Alice", "role": "Developer"},
					{"name": "Bob", "role": "Designer"},
					{"name": "Charlie", "role": "Product Manager"},
				},
			})
	})

	pageDataGroup.GET("/products", func(ctx *lokstra.Context) error {
		return ctx.HtmxPageData("Our Products", "", map[string]any{
			"products": []map[string]any{
				{"id": 1, "name": "Widget A", "price": 29.99},
				{"id": 2, "name": "Widget B", "price": 39.99},
				{"id": 3, "name": "Widget C", "price": 49.99},
			},
		})
	})

	pageDataGroup.GET("/contact", func(ctx *lokstra.Context) error {
		return ctx.HtmxPageData("Contact Us", "", map[string]any{
			"email":   "contact@example.com",
			"phone":   "+1-555-0123",
			"address": "123 Main St, City, State 12345",
		})
	})

	adminPageDataGroup := pageDataGroup.Group("/admin")
	adminPageDataGroup.GET("/", func(ctx *lokstra.Context) error {
		return ctx.HtmxPageData("Admin Dashboard", "", nil)
	})
}

func registerApi(app *lokstra.App) {
	// API endpoints for HTMX interactions
	apiGroup := app.Group("/api")
	apiGroup.POST("/contact", func(ctx *lokstra.Context) error {
		var form struct {
			Name    string `json:"name"`
			Email   string `json:"email"`
			Message string `json:"message"`
		}
		if err := ctx.BindBody(&form); err != nil {
			return ctx.ErrorBadRequest("Invalid form data")
		}

		// Process contact form (save to database, send email, etc.)
		lokstra.Logger.Infof("Contact form submitted: %+v", form)

		return ctx.Ok(map[string]any{
			"success": true,
			"message": "Thank you for your message! We'll get back to you soon.",
		})
	})

	apiGroup.GET("/products/{id}", func(ctx *lokstra.Context) error {
		id := ctx.GetPathParam("id")

		// Mock product data
		products := map[string]map[string]any{
			"1": {"id": 1, "name": "Widget A", "price": 29.99, "description": "A great widget"},
			"2": {"id": 2, "name": "Widget B", "price": 39.99, "description": "An even better widget"},
			"3": {"id": 3, "name": "Widget C", "price": 49.99, "description": "The best widget"},
		}

		product, exists := products[id]
		if !exists {
			return ctx.ErrorNotFound("Product not found")
		}

		return ctx.Ok(product)
	})

	// Health check and info endpoints
	apiGroup.GET("/info", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"app":     "HTMX Pages with Layout Example",
			"version": "1.0.0",
			"htmx_mounts": []map[string]any{
				{
					"path":        "/",
					"description": "Main HTMX application",
					"sources": []string{
						"./htmx_content (highest priority)",
						"./project/htmx",
						"htmx_app embed.FS (lowest priority)",
					},
				},
				{
					"path":        "/admin",
					"description": "Admin HTMX section",
					"sources": []string{
						"./admin_htmx (highest priority)",
						"htmx_app embed.FS (fallback)",
					},
				},
			},
		})
	})

	app.GET("/health", func(ctx *lokstra.Context) error {
		return ctx.Ok("OK")
	})
}
