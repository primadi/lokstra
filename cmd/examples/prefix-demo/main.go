package main

import (
	"net/http"

	"github.com/primadi/lokstra"
)

// Example demonstrating the difference between routerPrefix and pathPrefix
func main() {
	regCtx := lokstra.NewGlobalRegistrationContext()
	app := lokstra.NewApp(regCtx, "demo-app", ":8080")

	// Example 1: Router with base prefix
	apiRouter := app.WithPrefix("/api/v1") // routerPrefix = "/api/v1"

	// Exact match routes (routerPrefix + path)
	apiRouter.GET("/users", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"message": "GET /api/v1/users - User list",
			"type":    "exact match"})
	})

	apiRouter.POST("/users", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"message": "POST /api/v1/users - Create user",
			"type":    "exact match"})
	})

	// Catch-all routes (routerPrefix + pathPrefix + /*)
	apiRouter.GETPrefix("/files", func(ctx *lokstra.Context) error {
		path := ctx.Request.URL.Path
		return ctx.Ok(map[string]any{
			"message": "GET " + path + " - File handler (catch-all)",
			"type":    "prefix match"})
	})

	apiRouter.POSTPrefix("/upload", func(ctx *lokstra.Context) error {
		path := ctx.Request.URL.Path
		return ctx.Ok(map[string]any{
			"message": "POST " + path + " - Upload handler (catch-all)",
			"type":    "prefix match"})
	})

	// Example 2: Nested groups
	adminGroup := apiRouter.Group("/admin") // routerPrefix = "/api/v1/admin"

	adminGroup.GET("", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"message": "GET /api/v1/admin - Admin dashboard",
			"type":    "exact match"})
	})

	adminGroup.GETPrefix("/logs", func(ctx *lokstra.Context) error {
		path := ctx.Request.URL.Path
		return ctx.Ok(map[string]any{
			"message": "GET " + path + " - Admin logs (catch-all)",
			"type":    "prefix match"})
	})

	// Example 3: Root routes (no prefix)
	app.GET("", func(ctx *lokstra.Context) error {
		return ctx.HTML(http.StatusOK, `
		<h1>Router Prefix vs Path Prefix Demo</h1>
		<h2>Test these URLs:</h2>
		<ul>
			<li><strong>Exact routes:</strong>
				<ul>
					<li><a href="/api/v1/users">GET /api/v1/users</a> (exact match)</li>
					<li><a href="/api/v1/admin">GET /api/v1/admin</a> (exact match)</li>
				</ul>
			</li>
			<li><strong>Catch-all routes:</strong>
				<ul>
					<li><a href="/api/v1/files/document.pdf">GET /api/v1/files/document.pdf</a></li>
					<li><a href="/api/v1/files/images/photo.jpg">GET /api/v1/files/images/photo.jpg</a></li>
					<li><a href="/api/v1/upload/profile-pic">POST /api/v1/upload/profile-pic</a></li>
					<li><a href="/api/v1/admin/logs/system.log">GET /api/v1/admin/logs/system.log</a></li>
				</ul>
			</li>
			<li><strong>404 examples:</strong>
				<ul>
					<li><a href="/random">GET /random</a> (should be 404)</li>
					<li><a href="/api/v1/nonexistent">GET /api/v1/nonexistent</a> (should be 404)</li>
				</ul>
			</li>
		</ul>
		`)
	})

	server := lokstra.NewServer(regCtx, "demo-server").AddApps(app)
	server.Start(true)
}
