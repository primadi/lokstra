package main

import (
	"embed"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/common/static_files"
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

	// ===== HTMX Setup =====

	// Create static file handler with embedded filesystem
	sf := static_files.New().
		WithEmbedFS(htmxFS, "htmx_app")

	// Optional script injection for enhanced functionality
	scriptInjection := static_files.NewScriptInjection().
		AddNamedScriptInjection("default")

	// Mount HTMX pages at root
	app.MountHtmx("/", scriptInjection, sf.Sources...)

	// Mount static assets
	app.MountStatic("/static/", false, sf.Sources...)

	// ===== Page Data Handlers =====

	// Create page data routes
	pageData := app.Group("/page-data")

	// Home page data
	pageData.GET("/", func(ctx *lokstra.Context) error {
		return ctx.HtmxPageData(
			"Welcome to HTMX Example",             // Page title
			"Learn HTMX integration with Lokstra", // Description
			map[string]any{ // Page data
				"subtitle": "Build dynamic web applications with ease",
				"features": []string{
					"🚀 Fast server-side rendering",
					"⚡ Dynamic content without page reloads",
					"🎨 Clean, maintainable templates",
					"📱 Mobile-friendly responsive design",
				},
				"current_time": time.Now().Format("January 2, 2006 at 3:04 PM"),
			},
		)
	})

	// About page data
	pageData.GET("/about", func(ctx *lokstra.Context) error {
		return ctx.HtmxPageData(
			"About Us",
			"Learn about our company and mission",
			map[string]any{
				"company":  "Lokstra Framework",
				"founded":  2024,
				"location": "Global",
				"team": []map[string]any{
					{"name": "Alice Johnson", "role": "Framework Developer", "bio": "Go enthusiast with 10+ years experience"},
					{"name": "Bob Smith", "role": "Frontend Designer", "bio": "HTMX expert and UI/UX specialist"},
					{"name": "Carol White", "role": "DevOps Engineer", "bio": "Infrastructure and deployment expert"},
				},
				"mission": "To make web development simple, fast, and enjoyable for everyone.",
			},
		)
	})

	// Products page data
	pageData.GET("/products", func(ctx *lokstra.Context) error {
		products := []map[string]any{
			{
				"id":          1,
				"name":        "Lokstra Framework",
				"price":       "Free",
				"description": "A modern Go web framework with HTMX integration",
				"features":    []string{"Fast", "Simple", "Powerful"},
			},
			{
				"id":          2,
				"name":        "HTMX Components",
				"price":       "Free",
				"description": "Pre-built HTMX components for rapid development",
				"features":    []string{"Reusable", "Customizable", "Well-documented"},
			},
			{
				"id":          3,
				"name":        "Developer Tools",
				"price":       "Free",
				"description": "Debugging and development utilities",
				"features":    []string{"Debug mode", "Hot reload", "Error handling"},
			},
		}

		return ctx.HtmxPageData(
			"Our Products",
			"Discover the tools and frameworks we offer",
			map[string]any{
				"products": products,
				"total":    len(products),
			},
		)
	})

	// ===== Dynamic Content API =====

	// API routes for dynamic content loading
	api := app.Group("/api")

	// Dynamic content for home page
	api.GET("/featured-content", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"title": "Featured Content",
			"items": []map[string]any{
				{"title": "Getting Started Guide", "description": "Learn the basics of Lokstra", "link": "/docs/getting-started"},
				{"title": "HTMX Tutorial", "description": "Master HTMX with practical examples", "link": "/docs/htmx"},
				{"title": "Best Practices", "description": "Tips for building great apps", "link": "/docs/best-practices"},
			},
		})
	})

	// Dynamic stats
	api.GET("/stats", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"users_online":   42,
			"requests_today": 1523,
			"uptime":         "99.9%",
			"response_time":  "12ms",
			"last_updated":   time.Now().Format("15:04:05"),
		})
	})

	// Product details
	api.GET("/products/:id", func(ctx *lokstra.Context) error {
		productID := ctx.GetPathParam("id")

		// Simulate product lookup
		product := map[string]any{
			"id":          productID,
			"name":        "Product " + productID,
			"description": "Detailed information about product " + productID,
			"price":       "$99.99",
			"in_stock":    true,
			"rating":      4.5,
			"reviews": []map[string]any{
				{"user": "John D.", "rating": 5, "comment": "Excellent product!"},
				{"user": "Jane S.", "rating": 4, "comment": "Very good, would recommend."},
			},
		}

		return ctx.Ok(product)
	})

	// ===== Form Handling =====

	// Contact form submission
	api.POST("/contact", func(ctx *lokstra.Context) error {
		// Get form data
		name := ctx.GetQueryParam("name")
		email := ctx.GetQueryParam("email")
		message := ctx.GetQueryParam("message")

		// Validate required fields
		if name == "" || email == "" || message == "" {
			return ctx.ErrorBadRequest("All fields are required")
		}

		// Simulate processing
		time.Sleep(500 * time.Millisecond)

		// Return success message
		return ctx.Ok(map[string]any{
			"success": true,
			"message": "Thank you for your message! We'll get back to you soon.",
			"name":    name,
		})
	})

	// ===== Search Functionality =====

	api.GET("/search", func(ctx *lokstra.Context) error {
		query := ctx.GetQueryParam("q")

		if query == "" {
			return ctx.Ok(map[string]any{
				"results": []map[string]any{},
				"message": "Enter a search term to get started",
			})
		}

		// Simulate search results
		results := []map[string]any{
			{"title": "HTMX Basics", "description": "Learn the fundamentals of HTMX", "url": "/docs/htmx-basics"},
			{"title": "Lokstra Framework", "description": "Getting started with Lokstra", "url": "/docs/lokstra"},
			{"title": "Dynamic Content", "description": "Building dynamic web pages", "url": "/docs/dynamic"},
		}

		// Filter results based on query
		filteredResults := []map[string]any{}
		for _, result := range results {
			title := result["title"].(string)
			description := result["description"].(string)
			if containsIgnoreCase(title, query) || containsIgnoreCase(description, query) {
				filteredResults = append(filteredResults, result)
			}
		}

		return ctx.Ok(map[string]any{
			"results": filteredResults,
			"query":   query,
			"count":   len(filteredResults),
		})
	})

	lokstra.Logger.Infof("🚀 HTMX Basic Example started on :8080")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("HTMX Pages:")
	lokstra.Logger.Infof("  /                         - Home page with dynamic content")
	lokstra.Logger.Infof("  /about                    - About page with team info")
	lokstra.Logger.Infof("  /products                 - Products page with details")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Dynamic Content APIs:")
	lokstra.Logger.Infof("  GET  /api/featured-content - Load featured content")
	lokstra.Logger.Infof("  GET  /api/stats            - Live statistics")
	lokstra.Logger.Infof("  GET  /api/products/:id     - Product details")
	lokstra.Logger.Infof("  POST /api/contact          - Contact form submission")
	lokstra.Logger.Infof("  GET  /api/search?q=term    - Search functionality")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Page Data Endpoints:")
	lokstra.Logger.Infof("  /page-data/               - Home page data")
	lokstra.Logger.Infof("  /page-data/about          - About page data")
	lokstra.Logger.Infof("  /page-data/products       - Products page data")

	app.Start()
}

// Helper function for case-insensitive search
func containsIgnoreCase(s, substr string) bool {
	s = toLowerCase(s)
	substr = toLowerCase(substr)
	return contains(s, substr)
}

func toLowerCase(s string) string {
	result := ""
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			result += string(r + 32)
		} else {
			result += string(r)
		}
	}
	return result
}

func contains(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
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
