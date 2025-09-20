package main

import (
	"time"

	"github.com/primadi/lokstra"
)

func createPageDataRoutes(app *lokstra.App) {
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

	// Stats page data
	pageData.GET("/stats", func(ctx *lokstra.Context) error {
		return ctx.HtmxPageData(
			"Live Statistics Dashboard",
			"Real-time metrics and performance data",
			map[string]any{
				"users_online":   42,
				"requests_today": 1523,
				"uptime":         "99.9%",
				"response_time":  "12ms",
				"last_updated":   time.Now().Format("January 2, 2006 at 3:04 PM"),
			},
		)
	})

	// Featured Content page data
	pageData.GET("/featured-content", func(ctx *lokstra.Context) error {
		return ctx.HtmxPageData(
			"Featured Content Hub",
			"Discover trending articles, videos, and tutorials",
			map[string]any{
				"featured_title":       "Getting Started with HTMX",
				"featured_description": "A comprehensive guide to building dynamic web applications with HTMX and Lokstra",
				"featured_id":          "htmx-guide",
				"last_updated":         time.Now().Format("January 2, 2006 at 3:04 PM"),
			},
		)
	})
}
