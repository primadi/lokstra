package main

import (
	"embed"
	"os"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/common/static_files"
)

//go:embed htmx_app/*
var htmxAppAssets embed.FS

//go:embed project/htmx/*
var projectHtmx embed.FS

// This example demonstrates how to serve HTMX pages with layout support in Lokstra:
// 1. Serving HTMX pages with layouts and partials
// 2. Static assets serving for CSS, JS, images
// 3. Page data API endpoints for dynamic content
// 4. Fallback between multiple sources (project overrides → embedded defaults)
func main() {
	ctx := lokstra.NewGlobalRegistrationContext()
	app := lokstra.NewApp(ctx, "htmx-pages-app", ":8080")

	// Mount HTMX pages with layout support:
	// Sources should contain:
	//   - "/layouts" for HTML layout templates
	//   - "/pages" for HTML page templates

	sf := static_files.New().
		WithSourceDir("./htmx_content").          // Project overrides (highest priority)
		WithEmbedFS(projectHtmx, "project/htmx"). // Project static files (middle priority)
		WithEmbedFS(htmxAppAssets, "htmx_app")    // Embedded static files (lowest priority)
	app.MountHtmx("/", nil, sf.Sources...)

	sfAdmin := static_files.New().
		WithSourceDir("./admin_htmx").         // Admin-specific overrides
		WithEmbedFS(htmxAppAssets, "htmx_app") // Fallback to main HTMX app
	app.MountHtmx("/admin", nil, sfAdmin.Sources...)

	app.MountStatic("/static/", false, os.DirFS("./static")) // Serve static files from ./static directory

	// Page Data API endpoints - these provide dynamic data for HTMX pages

	pageData := app.Group("/page-data")
	// The HTMX handler will call these internally via /page-data/* routes
	pageData.GET("/", func(ctx *lokstra.Context) error {
		return ctx.HtmxPageData("Home Page", "", map[string]any{
			"message":   "Welcome to Lokstra HTMX Demo",
			"timestamp": time.Now(),
			"features": []string{
				"HTMX page serving with layouts",
				"Static asset fallback",
				"Partial rendering support",
				"Template-based rendering",
			}})
	})

	pageData.GET("/about", func(ctx *lokstra.Context) error {
		return ctx.HtmxPageData("About Us",
			"This is the about page with dynamic content",
			map[string]any{
				"team": []map[string]string{
					{"name": "Alice", "role": "Developer"},
					{"name": "Bob", "role": "Designer"},
					{"name": "Charlie", "role": "Product Manager"},
				}})
	})

	pageData.GET("/products", func(ctx *lokstra.Context) error {
		return ctx.HtmxPageData("Our Products", "",
			map[string]any{
				"products": []map[string]any{
					{"id": 1, "name": "Widget A", "price": 29.99},
					{"id": 2, "name": "Widget B", "price": 39.99},
					{"id": 3, "name": "Widget C", "price": 49.99},
				}})
	})

	pageData.GET("/contact", func(ctx *lokstra.Context) error {
		return ctx.HtmxPageData("Contact Us", "",
			map[string]any{
				"email":   "contact@example.com",
				"phone":   "+1-555-0123",
				"address": "123 Main St, City, State 12345",
			})
	})

	adminPageData := pageData.Group("/admin")
	// Admin-specific page data
	adminPageData.GET("/", func(ctx *lokstra.Context) error {
		return ctx.HtmxPageData("Admin Dashboard",
			"Welcome to the admin section", nil,
		)
	})

	// API endpoints for HTMX interactions
	app.POST("/api/contact", func(ctx *lokstra.Context) error {
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

	app.GET("/api/products/{id}", func(ctx *lokstra.Context) error {
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
	app.GET("/api/info", func(ctx *lokstra.Context) error {
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

	lokstra.Logger.Infof("HTMX pages server started on :8080")
	lokstra.Logger.Infof("HTMX page endpoints:")
	lokstra.Logger.Infof("  /              - Home page (index.html)")
	lokstra.Logger.Infof("  /about         - About page")
	lokstra.Logger.Infof("  /products      - Products page")
	lokstra.Logger.Infof("  /contact       - Contact page")
	lokstra.Logger.Infof("  /admin/        - Admin section")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Static assets:")
	lokstra.Logger.Infof("  /static/*      - CSS, JS, images")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("API endpoints:")
	lokstra.Logger.Infof("  /api/info      - Server information")
	lokstra.Logger.Infof("  /api/contact   - Contact form submission")
	lokstra.Logger.Infof("  /api/products/{id} - Product details")
	lokstra.Logger.Infof("  /health        - Health check")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("HTMX Features:")
	lokstra.Logger.Infof("  • Layout-based rendering with <!-- layout: layout_name.html --> directive")
	lokstra.Logger.Infof("  • Partial rendering for HTMX requests (HX-Request header)")
	lokstra.Logger.Infof("  • Page data injection from /page-data/* endpoints")
	lokstra.Logger.Infof("  • Static asset serving with fallback chain")
	lokstra.Logger.Infof("  • Multiple source fallback: project → embedded")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Try accessing:")
	lokstra.Logger.Infof("  http://localhost:8080/")
	lokstra.Logger.Infof("  http://localhost:8080/about")
	lokstra.Logger.Infof("  http://localhost:8080/products")
	lokstra.Logger.Infof("  http://localhost:8080/contact")
	lokstra.Logger.Infof("  http://localhost:8080/static/style.css")

	app.StartAndWaitForShutdown(30 * time.Second)
}
