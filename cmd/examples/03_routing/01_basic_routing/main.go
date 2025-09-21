package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/primadi/lokstra"
)

// This example demonstrates basic routing patterns in Lokstra.
// It shows path parameters, query parameters, HTTP methods, and route organization.
//
// Learning Objectives:
// - Understand URL path parameter extraction
// - Learn query parameter handling
// - Explore different HTTP methods (GET, POST, PUT, DELETE)
// - See route organization patterns
//
// Documentation: https://github.com/primadi/lokstra/blob/main/docs/routing.md#basic-routing

func main() {
	regCtx := lokstra.NewGlobalRegistrationContext()
	app := lokstra.NewApp(regCtx, "basic-routing-app", ":8080")

	// ===== Static Routes =====

	app.GET("/", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"message": "Welcome to Lokstra Basic Routing Example",
			"routes": map[string]any{
				"static":       "/about, /contact",
				"path_params":  "/users/:id, /products/:category/:id",
				"query_params": "/search?q=term&page=1",
				"http_methods": "GET, POST, PUT, DELETE examples",
			},
		})
	})

	app.GET("/about", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"page":        "About",
			"description": "This is the about page",
			"version":     "1.0.0",
		})
	})

	app.GET("/contact", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"page":    "Contact",
			"email":   "contact@example.com",
			"phone":   "+1-555-0123",
			"address": "123 Main St, City, State 12345",
		})
	})

	// ===== Single Path Parameters =====

	// Simple path parameter
	app.GET("/users/:id", func(ctx *lokstra.Context) error {
		userID := ctx.GetPathParam("id")

		return ctx.Ok(map[string]any{
			"message": "User profile",
			"user_id": userID,
			"route":   "/users/:id",
		})
	})

	// Path parameter with type conversion
	type UserParams struct {
		ID int `path:"id"`
	}

	app.GET("/users/:id/profile", func(ctx *lokstra.Context, params *UserParams) error {
		return ctx.Ok(map[string]any{
			"message": "User profile with type safety",
			"user_id": params.ID,
			"id_type": fmt.Sprintf("%T", params.ID),
			"route":   "/users/:id/profile",
		})
	})

	// ===== Multiple Path Parameters =====

	app.GET("/products/:category/:id", func(ctx *lokstra.Context) error {
		category := ctx.GetPathParam("category")
		productID := ctx.GetPathParam("id")

		return ctx.Ok(map[string]any{
			"message":    "Product details",
			"category":   category,
			"product_id": productID,
			"route":      "/products/:category/:id",
		})
	})

	// Multiple path parameters with type safety
	type ProductParams struct {
		Category string `path:"category"`
		ID       int    `path:"id"`
	}

	app.GET("/catalog/:category/:id", func(ctx *lokstra.Context, params *ProductParams) error {
		return ctx.Ok(map[string]any{
			"message":    "Catalog item",
			"category":   params.Category,
			"product_id": params.ID,
			"route":      "/catalog/:category/:id",
		})
	})

	// Complex path with multiple segments
	app.GET("/api/v1/users/:userId/posts/:postId/comments/:commentId", func(ctx *lokstra.Context) error {
		userID := ctx.GetPathParam("userId")
		postID := ctx.GetPathParam("postId")
		commentID := ctx.GetPathParam("commentId")

		return ctx.Ok(map[string]any{
			"message":    "Nested resource access",
			"user_id":    userID,
			"post_id":    postID,
			"comment_id": commentID,
			"route":      "/api/v1/users/:userId/posts/:postId/comments/:commentId",
		})
	})

	// ===== Query Parameters =====

	// Simple query parameters
	app.GET("/search", func(ctx *lokstra.Context) error {
		query := ctx.GetQueryParam("q")
		page := ctx.GetQueryParam("page")
		limit := ctx.GetQueryParam("limit")

		// Parse page and limit with defaults
		pageNum := 1
		if page != "" {
			if p, err := strconv.Atoi(page); err == nil {
				pageNum = p
			}
		}

		limitNum := 10
		if limit != "" {
			if l, err := strconv.Atoi(limit); err == nil {
				limitNum = l
			}
		}

		return ctx.Ok(map[string]any{
			"message": "Search results",
			"query":   query,
			"page":    pageNum,
			"limit":   limitNum,
			"route":   "/search?q=<term>&page=<num>&limit=<num>",
		})
	})

	// Query parameters with type safety and defaults
	type SearchParams struct {
		Query    string   `query:"q"`
		Page     int      `query:"page" default:"1"`
		Limit    int      `query:"limit" default:"10"`
		Sort     string   `query:"sort" default:"relevance"`
		Category []string `query:"category"` // Multiple values
		MinPrice float64  `query:"min_price"`
		MaxPrice float64  `query:"max_price"`
	}

	app.GET("/advanced-search", func(ctx *lokstra.Context, params *SearchParams) error {
		return ctx.Ok(map[string]any{
			"message":    "Advanced search with type safety",
			"query":      params.Query,
			"page":       params.Page,
			"limit":      params.Limit,
			"sort":       params.Sort,
			"categories": params.Category,
			"price_range": map[string]float64{
				"min": params.MinPrice,
				"max": params.MaxPrice,
			},
			"route": "/advanced-search?q=<term>&page=<num>&category=<cat1>&category=<cat2>",
		})
	})

	// ===== HTTP Methods =====

	// GET - Read operations
	app.GET("/posts", func(ctx *lokstra.Context) error {
		return ctx.OkList([]map[string]any{
			{"id": 1, "title": "First Post", "content": "Hello World"},
			{"id": 2, "title": "Second Post", "content": "Lokstra is awesome"},
		}, map[string]any{
			"total": 2,
			"page":  1,
		})
	})

	app.GET("/posts/:id", func(ctx *lokstra.Context, params *UserParams) error {
		return ctx.Ok(map[string]any{
			"id":      params.ID,
			"title":   "Sample Post",
			"content": "This is a sample post content",
			"method":  "GET",
		})
	})

	// POST - Create operations
	type CreatePostRequest struct {
		Title   string   `json:"title" validate:"required"`
		Content string   `json:"content" validate:"required"`
		Tags    []string `json:"tags"`
	}

	app.POST("/posts", func(ctx *lokstra.Context, req *CreatePostRequest) error {
		return ctx.OkCreated(map[string]any{
			"id":      123,
			"title":   req.Title,
			"content": req.Content,
			"tags":    req.Tags,
			"method":  "POST",
			"message": "Post created successfully",
		})
	})

	// PUT - Update operations
	type UpdatePostParams struct {
		ID   int                `path:"id"`
		Data *CreatePostRequest `json:",inline"`
	}

	app.PUT("/posts/:id", func(ctx *lokstra.Context, params *UpdatePostParams) error {
		return ctx.OkUpdated(map[string]any{
			"id":      params.ID,
			"title":   params.Data.Title,
			"content": params.Data.Content,
			"tags":    params.Data.Tags,
			"method":  "PUT",
			"message": "Post updated successfully",
		})
	})

	// PATCH - Partial update operations
	type PatchPostRequest struct {
		Title   *string   `json:"title,omitempty"`
		Content *string   `json:"content,omitempty"`
		Tags    *[]string `json:"tags,omitempty"`
	}

	app.PATCH("/posts/:id", func(ctx *lokstra.Context, params *UserParams, req *PatchPostRequest) error {
		updates := make(map[string]any)

		if req.Title != nil {
			updates["title"] = *req.Title
		}
		if req.Content != nil {
			updates["content"] = *req.Content
		}
		if req.Tags != nil {
			updates["tags"] = *req.Tags
		}

		return ctx.OkUpdated(map[string]any{
			"id":      params.ID,
			"updates": updates,
			"method":  "PATCH",
			"message": "Post partially updated",
		})
	})

	// DELETE - Delete operations
	app.DELETE("/posts/:id", func(ctx *lokstra.Context, params *UserParams) error {
		return ctx.Ok(map[string]any{
			"id":      params.ID,
			"method":  "DELETE",
			"message": "Post deleted successfully",
		})
	})

	// ===== Route Patterns =====

	// Wildcard routes (catch-all)
	app.GET("/files/*filepath", func(ctx *lokstra.Context) error {
		filepath := ctx.GetPathParam("filepath")

		return ctx.Ok(map[string]any{
			"message":  "File access",
			"filepath": filepath,
			"route":    "/files/*filepath",
			"note":     "Wildcard captures everything after /files/",
		})
	})

	// Optional parameters simulation
	app.GET("/optional", func(ctx *lokstra.Context) error {
		name := ctx.GetQueryParam("name")
		if name == "" {
			name = "Anonymous"
		}

		format := ctx.GetQueryParam("format")
		if format == "" {
			format = "json"
		}

		return ctx.Ok(map[string]any{
			"message": "Optional parameters example",
			"name":    name,
			"format":  format,
			"route":   "/optional?name=<optional>&format=<optional>",
		})
	})

	// Route with constraints simulation
	app.GET("/numeric/:id", func(ctx *lokstra.Context) error {
		idStr := ctx.GetPathParam("id")

		// Validate that ID is numeric
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return ctx.ErrorBadRequest("ID must be a valid number")
		}

		if id <= 0 {
			return ctx.ErrorBadRequest("ID must be a positive number")
		}

		return ctx.Ok(map[string]any{
			"message": "Numeric ID validated",
			"id":      id,
			"route":   "/numeric/:id (numeric constraint)",
		})
	})

	// Route with string constraints
	app.GET("/slug/:slug", func(ctx *lokstra.Context) error {
		slug := ctx.GetPathParam("slug")

		// Validate slug format (alphanumeric and hyphens only)
		if !isValidSlug(slug) {
			return ctx.ErrorBadRequest("Slug must contain only letters, numbers, and hyphens")
		}

		return ctx.Ok(map[string]any{
			"message": "Valid slug",
			"slug":    slug,
			"route":   "/slug/:slug (alphanumeric + hyphens only)",
		})
	})

	lokstra.Logger.Infof("🚀 Basic Routing Example started on :8080")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Basic Routing Examples:")
	lokstra.Logger.Infof("  Static Routes:")
	lokstra.Logger.Infof("    GET  /                    - Home page with route overview")
	lokstra.Logger.Infof("    GET  /about               - Static about page")
	lokstra.Logger.Infof("    GET  /contact             - Static contact page")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("  Path Parameters:")
	lokstra.Logger.Infof("    GET  /users/123           - Single path parameter")
	lokstra.Logger.Infof("    GET  /products/books/456  - Multiple path parameters")
	lokstra.Logger.Infof("    GET  /catalog/tech/789    - Type-safe path parameters")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("  Query Parameters:")
	lokstra.Logger.Infof("    GET  /search?q=golang&page=2")
	lokstra.Logger.Infof("    GET  /advanced-search?q=test&category=tech&category=web")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("  HTTP Methods:")
	lokstra.Logger.Infof("    GET    /posts             - List posts")
	lokstra.Logger.Infof("    POST   /posts             - Create post")
	lokstra.Logger.Infof("    PUT    /posts/123         - Update post")
	lokstra.Logger.Infof("    PATCH  /posts/123         - Partial update")
	lokstra.Logger.Infof("    DELETE /posts/123         - Delete post")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("  Advanced Patterns:")
	lokstra.Logger.Infof("    GET  /files/docs/readme.txt - Wildcard route")
	lokstra.Logger.Infof("    GET  /numeric/123         - Numeric constraint")
	lokstra.Logger.Infof("    GET  /slug/my-blog-post   - String constraint")

	app.Start(true)
}

// Helper function to validate slug format
func isValidSlug(slug string) bool {
	if slug == "" {
		return false
	}

	for _, char := range slug {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-') {
			return false
		}
	}

	// Don't allow leading or trailing hyphens
	return !strings.HasPrefix(slug, "-") && !strings.HasSuffix(slug, "-")
}

// Basic Routing Key Concepts:
//
// 1. Static Routes:
//    - Fixed URL paths that don't change
//    - Simple GET endpoints for pages, APIs, etc.
//    - Use app.GET("/path", handler)
//
// 2. Path Parameters:
//    - Dynamic segments in URL paths
//    - Extract with ctx.GetPathParam("name")
//    - Type-safe binding with struct tags
//    - Multiple parameters: /users/:userId/posts/:postId
//
// 3. Query Parameters:
//    - Optional parameters in URL query string
//    - Extract with ctx.GetQueryParam("name")
//    - Type-safe binding with defaults
//    - Multiple values supported
//
// 4. HTTP Methods:
//    - GET: Read operations
//    - POST: Create operations
//    - PUT: Full update operations
//    - PATCH: Partial update operations
//    - DELETE: Delete operations
//
// 5. Route Patterns:
//    - Wildcards: /*filepath for catch-all
//    - Constraints: Validate parameter format
//    - Optional parameters via query strings
//    - Complex nested resource paths

// Test Commands:
//
// # Static routes
// curl http://localhost:8080/
// curl http://localhost:8080/about
// curl http://localhost:8080/contact
//
// # Path parameters
// curl http://localhost:8080/users/123
// curl http://localhost:8080/products/books/456
// curl http://localhost:8080/catalog/tech/789
//
// # Query parameters
// curl "http://localhost:8080/search?q=golang&page=2&limit=5"
// curl "http://localhost:8080/advanced-search?q=test&category=tech&category=web&min_price=10&max_price=100"
//
// # HTTP methods
// curl http://localhost:8080/posts
// curl -X POST http://localhost:8080/posts -H "Content-Type: application/json" -d '{"title":"New Post","content":"Content here","tags":["go","web"]}'
// curl -X PUT http://localhost:8080/posts/123 -H "Content-Type: application/json" -d '{"title":"Updated Post","content":"Updated content"}'
// curl -X PATCH http://localhost:8080/posts/123 -H "Content-Type: application/json" -d '{"title":"Partially Updated"}'
// curl -X DELETE http://localhost:8080/posts/123
//
// # Advanced patterns
// curl http://localhost:8080/files/docs/readme.txt
// curl http://localhost:8080/numeric/123
// curl http://localhost:8080/slug/my-blog-post
// curl http://localhost:8080/slug/invalid_slug  # Should return error
