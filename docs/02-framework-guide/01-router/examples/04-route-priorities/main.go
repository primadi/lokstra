package main

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/response"
)

func main() {
	r := lokstra.NewRouter("Route Priorities Example")

	// ============================================
	// 1. Exact vs Parameter Routes
	// ============================================
	r.GET("/users/me", GetCurrentUser)
	r.GET("/users/active", GetActiveUsers)
	r.GET("/users/inactive", GetInactiveUsers)
	r.GET("/users/:id", GetUserByID)

	// ============================================
	// 2. Nested Routes with Mixed Types
	// ============================================
	r.GET("/posts/latest", GetLatestPost)
	r.GET("/posts/popular", GetPopularPosts)
	r.GET("/posts/:id", GetPostByID)
	r.GET("/posts/:id/comments", GetPostComments)
	r.GET("/posts/:id/comments/latest", GetLatestComment)
	r.GET("/posts/:id/comments/:commentId", GetCommentByID)

	// ============================================
	// 3. Wildcard Routes
	// ============================================
	r.GET("/files/upload", HandleFileUpload)
	r.GET("/files/info", GetFileInfo)
	r.GET("/files/:id", GetFileByID)
	r.GET("/files/*path", ServeFiles)

	// ============================================
	// 4. API Versioning with Wildcards
	// ============================================
	r.GET("/api/v1/status", GetAPIv1Status)
	r.GET("/api/v1/*path", HandleAPIv1)
	r.GET("/api/v2/status", GetAPIv2Status)
	r.GET("/api/v2/*path", HandleAPIv2)

	// ============================================
	// 5. Resource with Multiple Actions
	// ============================================
	r.GET("/products/search", SearchProducts)
	r.GET("/products/categories", GetCategories)
	r.GET("/products/featured", GetFeaturedProducts)
	r.GET("/products/:id", GetProduct)
	r.GET("/products/:id/reviews", GetProductReviews)
	r.GET("/products/:id/related", GetRelatedProducts)

	// ============================================
	// Home and Test Routes
	// ============================================
	r.GET("/", Home)
	r.GET("/test-priority", TestPriority)

	fmt.Println("🚀 Route Priorities Example")
	fmt.Println("📍 http://localhost:3000")
	fmt.Println("\n📋 Test routes:")
	fmt.Println("   GET /users/me          (exact)")
	fmt.Println("   GET /users/active      (exact)")
	fmt.Println("   GET /users/123         (parameter)")
	fmt.Println("   GET /posts/latest      (exact)")
	fmt.Println("   GET /posts/123         (parameter)")
	fmt.Println("   GET /files/upload      (exact)")
	fmt.Println("   GET /files/docs/readme (wildcard)")
	fmt.Println("\n🧪 Open test.http for more examples")

	app := lokstra.NewApp("route-priorities-example", ":3000", r)
	if err := app.Run(30 * time.Second); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

// ============================================
// 1. User Routes - Exact vs Parameter
// ============================================

func GetCurrentUser() *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"route":   "/users/me",
		"type":    "exact",
		"message": "Current user endpoint (exact match)",
		"user": map[string]any{
			"id":       "current",
			"username": "john_doe",
			"email":    "john@example.com",
		},
	})
}

func GetActiveUsers() *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"route":   "/users/active",
		"type":    "exact",
		"message": "Active users list",
		"count":   42,
		"users":   []string{"user1", "user2", "user3"},
	})
}

func GetInactiveUsers() *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"route":   "/users/inactive",
		"type":    "exact",
		"message": "Inactive users list",
		"count":   8,
	})
}

type UserIDParam struct {
	ID string `path:"id"`
}

func GetUserByID(params UserIDParam) *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"route":   "/users/:id",
		"type":    "parameter",
		"message": "User by ID (parameter route)",
		"id":      params.ID,
		"user": map[string]any{
			"id":       params.ID,
			"username": fmt.Sprintf("user_%s", params.ID),
		},
	})
}

// ============================================
// 2. Post Routes - Nested Parameters
// ============================================

func GetLatestPost() *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"route":   "/posts/latest",
		"type":    "exact",
		"message": "Latest post (exact match wins)",
		"post": map[string]any{
			"id":    "999",
			"title": "Latest Post",
		},
	})
}

func GetPopularPosts() *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"route":   "/posts/popular",
		"type":    "exact",
		"message": "Popular posts",
		"count":   10,
	})
}

type PostIDParam struct {
	ID string `path:"id"`
}

func GetPostByID(params PostIDParam) *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"route":   "/posts/:id",
		"type":    "parameter",
		"message": "Post by ID",
		"id":      params.ID,
	})
}

func GetPostComments(params PostIDParam) *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"route":   "/posts/:id/comments",
		"type":    "parameter",
		"message": "Comments for post",
		"post_id": params.ID,
		"count":   15,
	})
}

func GetLatestComment(params PostIDParam) *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"route":   "/posts/:id/comments/latest",
		"type":    "exact + parameter",
		"message": "Latest comment (exact 'latest' wins over :commentId)",
		"post_id": params.ID,
	})
}

type CommentParams struct {
	ID        string `path:"id"`
	CommentID string `path:"commentId"`
}

func GetCommentByID(params CommentParams) *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"route":      "/posts/:id/comments/:commentId",
		"type":       "parameter",
		"message":    "Comment by ID",
		"post_id":    params.ID,
		"comment_id": params.CommentID,
	})
}

// ============================================
// 3. File Routes - Wildcard
// ============================================

func HandleFileUpload() *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"route":   "/files/upload",
		"type":    "exact",
		"message": "File upload endpoint (exact match wins)",
	})
}

func GetFileInfo() *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"route":   "/files/info",
		"type":    "exact",
		"message": "File info endpoint",
	})
}

type FileIDParam struct {
	ID string `path:"id"`
}

func GetFileByID(params FileIDParam) *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"route":   "/files/:id",
		"type":    "parameter",
		"message": "File by ID (parameter wins over wildcard)",
		"id":      params.ID,
	})
}

type WildcardParam struct {
	Path string `path:"path"`
}

func ServeFiles(params WildcardParam) *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"route":   "/files/*path",
		"type":    "wildcard",
		"message": "Wildcard file serving (lowest priority)",
		"path":    params.Path,
		"example": "Matches: /files/docs/readme.txt",
	})
}

// ============================================
// 4. API Versioning
// ============================================

func GetAPIv1Status() *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"route":   "/api/v1/status",
		"type":    "exact",
		"version": "1.0",
		"status":  "healthy",
	})
}

func HandleAPIv1(params WildcardParam) *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"route":   "/api/v1/*path",
		"type":    "wildcard",
		"version": "1.0",
		"path":    params.Path,
		"message": "API v1 wildcard handler",
	})
}

func GetAPIv2Status() *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"route":   "/api/v2/status",
		"type":    "exact",
		"version": "2.0",
		"status":  "healthy",
	})
}

func HandleAPIv2(params WildcardParam) *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"route":   "/api/v2/*path",
		"type":    "wildcard",
		"version": "2.0",
		"path":    params.Path,
		"message": "API v2 wildcard handler",
	})
}

// ============================================
// 5. Product Routes
// ============================================

func SearchProducts() *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"route":   "/products/search",
		"type":    "exact",
		"message": "Product search endpoint",
	})
}

func GetCategories() *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"route":      "/products/categories",
		"type":       "exact",
		"categories": []string{"Electronics", "Books", "Clothing"},
	})
}

func GetFeaturedProducts() *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"route":   "/products/featured",
		"type":    "exact",
		"message": "Featured products",
		"count":   5,
	})
}

type ProductIDParam struct {
	ID string `path:"id"`
}

func GetProduct(params ProductIDParam) *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"route":   "/products/:id",
		"type":    "parameter",
		"id":      params.ID,
		"product": fmt.Sprintf("Product %s", params.ID),
	})
}

func GetProductReviews(params ProductIDParam) *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"route":      "/products/:id/reviews",
		"type":       "parameter",
		"product_id": params.ID,
		"count":      23,
	})
}

func GetRelatedProducts(params ProductIDParam) *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"route":      "/products/:id/related",
		"type":       "parameter",
		"product_id": params.ID,
		"related":    []string{"prod1", "prod2"},
	})
}

// ============================================
// Home & Test Routes
// ============================================

func Home() *response.Response {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Route Priorities Example</title>
    <style>
        body { font-family: Arial; max-width: 1200px; margin: 40px auto; padding: 20px; }
        h1 { color: #333; }
        .section { margin: 30px 0; padding: 20px; background: #f5f5f5; border-radius: 8px; }
        .route { margin: 10px 0; padding: 10px; background: white; border-left: 4px solid #007bff; }
        .exact { border-left-color: #28a745; }
        .param { border-left-color: #ffc107; }
        .wildcard { border-left-color: #dc3545; }
        code { background: #e9ecef; padding: 2px 6px; border-radius: 3px; }
        .badge { display: inline-block; padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; }
        .badge-exact { background: #28a745; color: white; }
        .badge-param { background: #ffc107; color: black; }
        .badge-wildcard { background: #dc3545; color: white; }
    </style>
</head>
<body>
    <h1>🎯 Route Priorities Example</h1>
    <p>Understanding route matching order in Lokstra</p>

    <div class="section">
        <h2>Priority Rules</h2>
        <ol>
            <li><span class="badge badge-exact">EXACT</span> Exact matches (no parameters) - Highest priority</li>
            <li><span class="badge badge-param">PARAM</span> Parameterized routes (:id) - Medium priority</li>
            <li><span class="badge badge-wildcard">WILDCARD</span> Wildcard routes (*path) - Lowest priority</li>
        </ol>
    </div>

    <div class="section">
        <h2>1. User Routes</h2>
        <div class="route exact">
            <span class="badge badge-exact">EXACT</span>
            <a href="/users/me">/users/me</a> - Current user
        </div>
        <div class="route exact">
            <span class="badge badge-exact">EXACT</span>
            <a href="/users/active">/users/active</a> - Active users
        </div>
        <div class="route param">
            <span class="badge badge-param">PARAM</span>
            <a href="/users/123">/users/:id</a> - User by ID
        </div>
    </div>

    <div class="section">
        <h2>2. Post Routes (Nested)</h2>
        <div class="route exact">
            <span class="badge badge-exact">EXACT</span>
            <a href="/posts/latest">/posts/latest</a>
        </div>
        <div class="route param">
            <span class="badge badge-param">PARAM</span>
            <a href="/posts/123">/posts/:id</a>
        </div>
        <div class="route exact">
            <span class="badge badge-exact">EXACT</span>
            <a href="/posts/123/comments/latest">/posts/:id/comments/latest</a>
        </div>
        <div class="route param">
            <span class="badge badge-param">PARAM</span>
            <a href="/posts/123/comments/456">/posts/:id/comments/:commentId</a>
        </div>
    </div>

    <div class="section">
        <h2>3. File Routes (Wildcard)</h2>
        <div class="route exact">
            <span class="badge badge-exact">EXACT</span>
            <a href="/files/upload">/files/upload</a>
        </div>
        <div class="route param">
            <span class="badge badge-param">PARAM</span>
            <a href="/files/12345">/files/:id</a>
        </div>
        <div class="route wildcard">
            <span class="badge badge-wildcard">WILDCARD</span>
            <a href="/files/docs/readme.txt">/files/*path</a>
        </div>
    </div>

    <div class="section">
        <h2>4. API Versioning</h2>
        <div class="route exact">
            <span class="badge badge-exact">EXACT</span>
            <a href="/api/v1/status">/api/v1/status</a>
        </div>
        <div class="route wildcard">
            <span class="badge badge-wildcard">WILDCARD</span>
            <a href="/api/v1/users/list">/api/v1/*path</a>
        </div>
        <div class="route exact">
            <span class="badge badge-exact">EXACT</span>
            <a href="/api/v2/status">/api/v2/status</a>
        </div>
        <div class="route wildcard">
            <span class="badge badge-wildcard">WILDCARD</span>
            <a href="/api/v2/posts/trending">/api/v2/*path</a>
        </div>
    </div>

    <div class="section">
        <h2>📖 Documentation</h2>
        <p>See <code>index</code> for detailed explanation</p>
        <p>Use <code>test.http</code> for API testing</p>
    </div>
</body>
</html>`

	return response.NewHtmlResponse(html)
}

func TestPriority() *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"message": "Route priority test endpoint",
		"rules": []string{
			"1. Exact matches have highest priority",
			"2. Parameter routes have medium priority",
			"3. Wildcard routes have lowest priority",
			"4. Registration order doesn't matter",
		},
		"examples": map[string]string{
			"exact":    "/users/me (always matches first)",
			"param":    "/users/:id (matches when no exact route)",
			"wildcard": "/files/*path (matches everything else)",
		},
	})
}
