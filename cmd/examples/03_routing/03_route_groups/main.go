package main

import (
	"strconv"

	"github.com/primadi/lokstra"
)

// This example demonstrates route grouping and organization in Lokstra.
// It shows how to structure routes logically, apply group-specific middleware,
// and build maintainable API hierarchies.
//
// Learning Objectives:
// - Understand route grouping concepts
// - Learn to organize routes by functionality
// - Explore nested groups and middleware inheritance
// - See real-world API structure patterns
//
// Documentation: https://github.com/primadi/lokstra/blob/main/docs/routing.md#route-groups

func main() {
	regCtx := lokstra.NewGlobalRegistrationContext()
	app := lokstra.NewApp(regCtx, "route-groups-app", ":8080")

	// ===== Global Middleware =====
	app.Use(func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
		lokstra.Logger.Infof("🌐 Global middleware: %s %s", ctx.Request.Method, ctx.Request.URL.Path)
		return next(ctx)
	})

	// ===== Root Routes =====
	app.GET("/", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]interface{}{
			"message": "Route Groups Example API",
			"version": "1.0.0",
			"endpoints": map[string]string{
				"health": "/health",
				"api_v1": "/api/v1/*",
				"api_v2": "/api/v2/*",
				"admin":  "/admin/*",
				"public": "/public/*",
				"docs":   "/docs/*",
			},
		})
	})

	app.GET("/health", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]interface{}{
			"status":    "healthy",
			"timestamp": "2024-01-01T00:00:00Z",
		})
	})

	// ===== API Version 1 Group =====
	v1 := app.Group("/api/v1")

	// V1 middleware
	v1.Use(func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
		lokstra.Logger.Infof("📱 API v1 middleware")
		ctx.Response.WithHeader("API-Version", "1.0")
		return next(ctx)
	})

	// V1 Users subgroup
	v1Users := v1.Group("/users")

	type UserV1 struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Version string `json:"version"`
	}

	v1Users.GET("", func(ctx *lokstra.Context) error {
		users := []UserV1{
			{ID: 1, Name: "John Doe", Email: "john@example.com", Version: "v1"},
			{ID: 2, Name: "Jane Smith", Email: "jane@example.com", Version: "v1"},
		}
		return ctx.Ok(map[string]interface{}{
			"users":       users,
			"total":       len(users),
			"api_version": "1.0",
		})
	})

	v1Users.GET("/:id", func(ctx *lokstra.Context) error {
		idStr := ctx.GetPathParam("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return ctx.ErrorBadRequest("Invalid user ID")
		}

		user := UserV1{
			ID:      id,
			Name:    "User " + idStr,
			Email:   "user" + idStr + "@example.com",
			Version: "v1",
		}

		return ctx.Ok(map[string]interface{}{
			"user":        user,
			"api_version": "1.0",
		})
	})

	type CreateUserV1Request struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	v1Users.POST("", func(ctx *lokstra.Context, req *CreateUserV1Request) error {
		user := UserV1{
			ID:      99,
			Name:    req.Name,
			Email:   req.Email,
			Version: "v1",
		}

		return ctx.OkCreated(map[string]interface{}{
			"message":     "User created (v1)",
			"user":        user,
			"api_version": "1.0",
		})
	})

	// V1 Posts subgroup
	v1Posts := v1.Group("/posts")

	type PostV1 struct {
		ID       int    `json:"id"`
		Title    string `json:"title"`
		Content  string `json:"content"`
		AuthorID int    `json:"author_id"`
		Version  string `json:"version"`
	}

	v1Posts.GET("", func(ctx *lokstra.Context) error {
		posts := []PostV1{
			{ID: 1, Title: "First Post", Content: "Hello World", AuthorID: 1, Version: "v1"},
			{ID: 2, Title: "Second Post", Content: "Hello Again", AuthorID: 2, Version: "v1"},
		}
		return ctx.Ok(map[string]interface{}{
			"posts":       posts,
			"total":       len(posts),
			"api_version": "1.0",
		})
	})

	v1Posts.GET("/:id", func(ctx *lokstra.Context) error {
		idStr := ctx.GetPathParam("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return ctx.ErrorBadRequest("Invalid post ID")
		}

		post := PostV1{
			ID:       id,
			Title:    "Post " + idStr,
			Content:  "Content for post " + idStr,
			AuthorID: 1,
			Version:  "v1",
		}

		return ctx.Ok(map[string]interface{}{
			"post":        post,
			"api_version": "1.0",
		})
	})

	// ===== API Version 2 Group =====
	v2 := app.Group("/api/v2")

	// V2 middleware
	v2.Use(func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
		lokstra.Logger.Infof("📱 API v2 middleware")
		ctx.Response.WithHeader("API-Version", "2.0")
		ctx.Response.WithHeader("X-Feature-Set", "extended")
		return next(ctx)
	})

	// V2 Users subgroup (enhanced version)
	v2Users := v2.Group("/users")

	type UserV2 struct {
		ID        int      `json:"id"`
		Name      string   `json:"name"`
		Email     string   `json:"email"`
		Avatar    string   `json:"avatar"`
		Roles     []string `json:"roles"`
		CreatedAt string   `json:"created_at"`
		Version   string   `json:"version"`
	}

	v2Users.GET("", func(ctx *lokstra.Context) error {
		pageStr := ctx.GetQueryParam("page")
		page := 1
		if pageStr != "" {
			var err error
			page, err = strconv.Atoi(pageStr)
			if err != nil || page < 1 {
				page = 1
			}
		}

		limitStr := ctx.GetQueryParam("limit")
		limit := 10
		if limitStr != "" {
			var err error
			limit, err = strconv.Atoi(limitStr)
			if err != nil || limit < 1 || limit > 100 {
				limit = 10
			}
		}

		users := []UserV2{
			{
				ID:        1,
				Name:      "John Doe",
				Email:     "john@example.com",
				Avatar:    "https://api.dicebear.com/7.x/avataaars/svg?seed=John",
				Roles:     []string{"user", "moderator"},
				CreatedAt: "2024-01-01T00:00:00Z",
				Version:   "v2",
			},
			{
				ID:        2,
				Name:      "Jane Smith",
				Email:     "jane@example.com",
				Avatar:    "https://api.dicebear.com/7.x/avataaars/svg?seed=Jane",
				Roles:     []string{"user", "admin"},
				CreatedAt: "2024-01-02T00:00:00Z",
				Version:   "v2",
			},
		}

		return ctx.Ok(map[string]interface{}{
			"users": users,
			"pagination": map[string]interface{}{
				"page":  page,
				"limit": limit,
				"total": len(users),
			},
			"api_version": "2.0",
		})
	})

	v2Users.GET("/:id", func(ctx *lokstra.Context) error {
		idStr := ctx.GetPathParam("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return ctx.ErrorBadRequest("Invalid user ID")
		}

		user := UserV2{
			ID:        id,
			Name:      "User " + idStr,
			Email:     "user" + idStr + "@example.com",
			Avatar:    "https://api.dicebear.com/7.x/avataaars/svg?seed=User" + idStr,
			Roles:     []string{"user"},
			CreatedAt: "2024-01-01T00:00:00Z",
			Version:   "v2",
		}

		return ctx.Ok(map[string]interface{}{
			"user":        user,
			"api_version": "2.0",
		})
	})

	// V2 Posts subgroup (enhanced version)
	v2Posts := v2.Group("/posts")

	type PostV2 struct {
		ID        int      `json:"id"`
		Title     string   `json:"title"`
		Content   string   `json:"content"`
		Author    UserV2   `json:"author"`
		Tags      []string `json:"tags"`
		CreatedAt string   `json:"created_at"`
		UpdatedAt string   `json:"updated_at"`
		Version   string   `json:"version"`
	}

	v2Posts.GET("", func(ctx *lokstra.Context) error {
		author := UserV2{
			ID:      1,
			Name:    "John Doe",
			Email:   "john@example.com",
			Version: "v2",
		}

		posts := []PostV2{
			{
				ID:        1,
				Title:     "Advanced Post",
				Content:   "This is an advanced post with more features",
				Author:    author,
				Tags:      []string{"tech", "programming"},
				CreatedAt: "2024-01-01T00:00:00Z",
				UpdatedAt: "2024-01-01T00:00:00Z",
				Version:   "v2",
			},
		}

		return ctx.Ok(map[string]interface{}{
			"posts": posts,
			"meta": map[string]interface{}{
				"total_count": len(posts),
				"api_version": "2.0",
				"features":    []string{"pagination", "filtering", "sorting"},
			},
		})
	})

	// ===== Admin Group =====
	admin := app.Group("/admin")

	// Admin authentication middleware
	admin.Use(func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
		token := ctx.GetHeader("Admin-Token")
		if token != "admin-secret" {
			return ctx.ErrorBadRequest("Admin access required")
		}
		lokstra.Logger.Infof("🔐 Admin authenticated")
		return next(ctx)
	})

	admin.GET("/dashboard", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]interface{}{
			"message": "Admin Dashboard",
			"stats": map[string]interface{}{
				"total_users":     150,
				"total_posts":     75,
				"active_sessions": 23,
			},
		})
	})

	// Nested admin groups
	adminUsers := admin.Group("/users")
	adminUsers.GET("", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]interface{}{
			"message": "Admin: All users",
			"count":   150,
		})
	})

	adminUsers.DELETE("/:id", func(ctx *lokstra.Context) error {
		userID := ctx.GetPathParam("id")
		return ctx.Ok(map[string]interface{}{
			"message": "Admin: User deleted",
			"user_id": userID,
		})
	})

	adminSystem := admin.Group("/system")
	adminSystem.GET("/logs", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]interface{}{
			"message": "System logs",
			"logs": []string{
				"INFO: Server started",
				"INFO: Database connected",
				"WARN: High memory usage",
			},
		})
	})

	adminSystem.POST("/restart", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]interface{}{
			"message": "System restart initiated",
			"status":  "pending",
		})
	})

	// ===== Public Content Group =====
	public := app.Group("/public")

	public.Use(func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
		lokstra.Logger.Infof("🌍 Public content middleware")
		ctx.Response.WithHeader("Cache-Control", "public, max-age=3600")
		return next(ctx)
	})

	publicBlog := public.Group("/blog")
	publicBlog.GET("/latest", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]interface{}{
			"message": "Latest blog posts",
			"posts": []string{
				"How to build REST APIs",
				"Understanding middleware",
				"Route grouping best practices",
			},
		})
	})

	publicPages := public.Group("/pages")
	publicPages.GET("/about", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]interface{}{
			"page":    "about",
			"content": "About our amazing API",
		})
	})

	publicPages.GET("/contact", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]interface{}{
			"page":    "contact",
			"content": "Contact us at hello@example.com",
		})
	})

	// ===== Documentation Group =====
	docs := app.Group("/docs")

	docs.Use(func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
		ctx.Response.WithHeader("Content-Type", "text/plain")
		return next(ctx)
	})

	docs.GET("/api", func(ctx *lokstra.Context) error {
		documentation := `
API Documentation
================

Available Endpoints:

1. API v1 (/api/v1)
   - GET    /users       - List users (basic)
   - GET    /users/:id   - Get user by ID
   - POST   /users       - Create user
   - GET    /posts       - List posts (basic)
   - GET    /posts/:id   - Get post by ID

2. API v2 (/api/v2)
   - GET    /users       - List users (enhanced with pagination)
   - GET    /users/:id   - Get user by ID (with more fields)
   - GET    /posts       - List posts (enhanced with author details)

3. Admin (/admin) - Requires Admin-Token header
   - GET    /dashboard   - Admin dashboard
   - GET    /users       - Admin user management
   - DELETE /users/:id   - Delete user
   - GET    /system/logs - System logs
   - POST   /system/restart - Restart system

4. Public (/public)
   - GET    /blog/latest - Latest blog posts
   - GET    /pages/about - About page
   - GET    /pages/contact - Contact page

5. Documentation (/docs)
   - GET    /api         - API documentation
   - GET    /examples    - Usage examples
`
		return ctx.Response.WriteRaw("text/plain", 200, []byte(documentation))
	})

	docs.GET("/examples", func(ctx *lokstra.Context) error {
		examples := `
Usage Examples
==============

# List V1 users
curl http://localhost:8080/api/v1/users

# List V2 users with pagination
curl "http://localhost:8080/api/v2/users?page=1&limit=5"

# Get specific user (V2)
curl http://localhost:8080/api/v2/users/1

# Create user (V1)
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"New User","email":"user@example.com"}'

# Admin dashboard (requires token)
curl -H "Admin-Token: admin-secret" http://localhost:8080/admin/dashboard

# Public content
curl http://localhost:8080/public/blog/latest

# Health check
curl http://localhost:8080/health
`
		return ctx.Response.WriteRaw("text/plain", 200, []byte(examples))
	})

	lokstra.Logger.Infof("🚀 Route Groups Example started on :8080")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Route Structure:")
	lokstra.Logger.Infof("  /                         - API overview")
	lokstra.Logger.Infof("  /health                   - Health check")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("  /api/v1/                  - API Version 1")
	lokstra.Logger.Infof("    /users                  - V1 Users (basic)")
	lokstra.Logger.Infof("    /posts                  - V1 Posts (basic)")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("  /api/v2/                  - API Version 2")
	lokstra.Logger.Infof("    /users                  - V2 Users (enhanced)")
	lokstra.Logger.Infof("    /posts                  - V2 Posts (enhanced)")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("  /admin/                   - Admin Panel (token required)")
	lokstra.Logger.Infof("    /dashboard              - Admin dashboard")
	lokstra.Logger.Infof("    /users                  - User management")
	lokstra.Logger.Infof("    /system                 - System management")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("  /public/                  - Public Content")
	lokstra.Logger.Infof("    /blog                   - Blog posts")
	lokstra.Logger.Infof("    /pages                  - Static pages")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("  /docs/                    - Documentation")
	lokstra.Logger.Infof("    /api                    - API docs")
	lokstra.Logger.Infof("    /examples               - Usage examples")

	app.Start()
}

// Route Groups Key Concepts:
//
// 1. Group Creation:
//    - app.Group("/prefix") creates a new route group
//    - Groups inherit global middleware automatically
//    - Groups can have their own middleware
//
// 2. Nested Groups:
//    - Groups can contain sub-groups
//    - Middleware inheritance flows down the hierarchy
//    - URL paths are automatically combined
//
// 3. Middleware Inheritance:
//    - Global → Group → Sub-group → Route
//    - Each level can add its own middleware
//    - Execution order follows the hierarchy
//
// 4. Organization Benefits:
//    - Logical grouping of related routes
//    - Shared middleware for route families
//    - Version management (v1, v2, etc.)
//    - Access control by group
//
// 5. Common Patterns:
//    - API versioning (/api/v1, /api/v2)
//    - Access levels (/public, /protected, /admin)
//    - Feature grouping (/users, /posts, /orders)
//    - Content types (/api, /docs, /static)

// Test Commands:
//
// # Root endpoints
// curl http://localhost:8080/
// curl http://localhost:8080/health
//
// # API v1 endpoints
// curl http://localhost:8080/api/v1/users
// curl http://localhost:8080/api/v1/users/1
// curl http://localhost:8080/api/v1/posts
//
// # API v2 endpoints (enhanced)
// curl http://localhost:8080/api/v2/users
// curl "http://localhost:8080/api/v2/users?page=1&limit=5"
// curl http://localhost:8080/api/v2/posts
//
// # Admin endpoints (require Admin-Token header)
// curl -H "Admin-Token: admin-secret" http://localhost:8080/admin/dashboard
// curl -H "Admin-Token: admin-secret" http://localhost:8080/admin/users
// curl -H "Admin-Token: admin-secret" http://localhost:8080/admin/system/logs
//
// # Public endpoints
// curl http://localhost:8080/public/blog/latest
// curl http://localhost:8080/public/pages/about
//
// # Documentation
// curl http://localhost:8080/docs/api
// curl http://localhost:8080/docs/examples
//
// # Create user (v1)
// curl -X POST http://localhost:8080/api/v1/users \
//   -H "Content-Type: application/json" \
//   -d '{"name":"Test User","email":"test@example.com"}'
