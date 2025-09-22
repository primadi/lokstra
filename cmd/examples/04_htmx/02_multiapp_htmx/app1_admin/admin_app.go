package admin

import (
	"embed"
	"fmt"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/router"
)

//go:embed layouts/* pages/* static/*
var AppFiles embed.FS

// Setup configures the admin dashboard application with the given prefix
func Setup(app *lokstra.App, prefix string) {
	adminGroup := app.Group(prefix)

	// Setup HTMX manager with embedded files
	adminGroup.AddHtmxLayouts(AppFiles, "layouts").
		AddHtmxPages(AppFiles, "pages").
		AddHtmxStatics(AppFiles, "static")

	// adminGroup.SetHTMXLayoutScriptInjection(htmx_fsmanager.NewDefaultScriptInjection(false))

	setupRoutes(adminGroup)
}

func setupRoutes(group router.Router) {
	// Dashboard page (main)
	group.GET("/", func(ctx *lokstra.Context) error {
		data := map[string]interface{}{
			"stats": map[string]interface{}{
				"totalUsers":  1247,
				"totalOrders": 3489,
				"revenue":     "$52,389",
				"growth":      "+12%",
			},
			"recentActivity": []map[string]interface{}{
				{"action": "New user registration", "user": "john.doe@example.com", "time": "2 minutes ago"},
				{"action": "Order completed", "user": "sarah.smith@example.com", "time": "5 minutes ago"},
				{"action": "Payment received", "user": "mike.johnson@example.com", "time": "8 minutes ago"},
			},
		}
		return ctx.HTMXFSPage("dashboard", data, "Dashboard - Admin Panel", "Overview of admin dashboard statistics and recent activities.")
	})

	// Users management
	group.GET("/users", func(ctx *lokstra.Context) error {
		users := []map[string]interface{}{
			{"id": 1, "name": "John Doe", "email": "john.doe@example.com", "role": "Admin", "status": "Active", "lastLogin": "2 hours ago"},
			{"id": 2, "name": "Sarah Smith", "email": "sarah.smith@example.com", "role": "Editor", "status": "Active", "lastLogin": "1 day ago"},
			{"id": 3, "name": "Mike Johnson", "email": "mike.johnson@example.com", "role": "User", "status": "Inactive", "lastLogin": "1 week ago"},
			{"id": 4, "name": "Emily Davis", "email": "emily.davis@example.com", "role": "User", "status": "Active", "lastLogin": "3 hours ago"},
		}
		data := map[string]interface{}{"users": users}
		return ctx.HTMXFSPage("users", data, "Users - Admin Panel", "Manage application users and their roles.")
	})

	// Settings page
	group.GET("/settings", func(ctx *lokstra.Context) error {
		settings := map[string]interface{}{
			"siteName":            "Admin Dashboard",
			"adminEmail":          "admin@example.com",
			"enableNotifications": true,
			"maintenanceMode":     false,
			"theme":               "blue",
		}
		data := map[string]interface{}{"settings": settings}
		return ctx.HTMXFSPage("settings", data, "Settings - Admin Panel", "Configure application settings.")
	})

	// Analytics page
	group.GET("/analytics", func(ctx *lokstra.Context) error {
		data := map[string]interface{}{
			"pageViews": []map[string]interface{}{
				{"page": "/dashboard", "views": 1250, "change": "+5%"},
				{"page": "/users", "views": 890, "change": "+12%"},
				{"page": "/settings", "views": 456, "change": "-2%"},
				{"page": "/analytics", "views": 234, "change": "+8%"},
			},
			"trafficSources": map[string]interface{}{
				"direct":   45,
				"search":   35,
				"social":   15,
				"referral": 5,
			},
		}
		return ctx.HTMXFSPage("analytics", data, "Analytics - Admin Panel", "View application analytics and performance metrics.")
	})

	// API endpoints for HTMX interactions
	group.GET("/api/stats", func(ctx *lokstra.Context) error {
		stats := map[string]interface{}{
			"totalUsers":  1247 + time.Now().Second(), // Simulate changing data
			"totalOrders": 3489 + time.Now().Minute(),
			"revenue":     fmt.Sprintf("$%d", 52389+time.Now().Hour()*100),
			"growth":      "+12%",
		}
		return ctx.Ok(stats)
	})

	group.POST("/api/users", func(ctx *lokstra.Context) error {
		// Simulate adding a new user
		name := ctx.Request.FormValue("name")
		email := ctx.Request.FormValue("email")
		role := ctx.Request.FormValue("role")

		newUser := map[string]interface{}{
			"id":        time.Now().Unix(),
			"name":      name,
			"email":     email,
			"role":      role,
			"status":    "Active",
			"lastLogin": "Just now",
		}
		return ctx.Ok(newUser)
	})
}
