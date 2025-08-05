package main

import (
	"github.com/primadi/lokstra"
)

// This example demonstrates how to mount reverse proxies in Lokstra:
// 1. Basic reverse proxy mounting
// 2. Reverse proxy with middleware
// 3. Multiple proxy targets
// 4. Load balancing simulation
func main() {
	ctx := lokstra.NewGlobalRegistrationContext()
	app := lokstra.NewApp(ctx, "reverse-proxy-app", ":8080")

	// Register middleware for proxy logging
	ctx.RegisterMiddlewareFunc("proxy_logger", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		lokstra.Logger.Infof("Registering proxy_logger middleware")
		return func(ctx *lokstra.Context) error {
			lokstra.Logger.Infof("[PROXY] Forwarding %s %s", ctx.Request.Method, ctx.Request.URL.Path)

			err := next(ctx)
			if ctx.ShouldStopMiddlewareChain(err) {
				return err
			}

			lokstra.Logger.Infof("[PROXY] After Forwarding %s %s", ctx.Request.Method, ctx.Request.URL.Path)
			return nil
		}
	})

	// Register auth middleware for protected proxies
	ctx.RegisterMiddlewareFunc("proxy_auth", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		lokstra.Logger.Infof("Registering proxy_auth middleware")
		return func(ctx *lokstra.Context) error {
			apiKey := ctx.Request.Header.Get("X-API-Key")
			if apiKey != "secret-key" {
				lokstra.Logger.Infof("[PROXY-AUTH] Missing or invalid API key")
				return ctx.ErrorBadRequest("API key required")
			}
			lokstra.Logger.Infof("[PROXY-AUTH] Authenticated request to %s", ctx.Request.URL.Path)

			err := next(ctx)
			if ctx.ShouldStopMiddlewareChain(err) {
				return err
			}

			lokstra.Logger.Infof("[PROXY-AUTH] After authenticated request to %s", ctx.Request.URL.Path)
			return nil
		}
	})

	// Local API endpoints (not proxied)
	app.GET("/api/local/status", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"service": "Local API",
			"status":  "running",
			"proxies": []string{
				"/api/external -> https://jsonplaceholder.typicode.com",
				"/api/secure -> https://httpbin.org",
				"/api/github -> https://api.github.com",
			},
		})
	})

	// Health check
	app.GET("/health", func(ctx *lokstra.Context) error {
		return ctx.Ok("Reverse Proxy Server is healthy")
	})

	// Information endpoint
	app.GET("/", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"app":         "Reverse Proxy Example",
			"version":     "1.0.0",
			"description": "Demonstrates reverse proxy mounting in Lokstra",
			"endpoints": map[string]string{
				"/":                 "This information",
				"/health":           "Health check",
				"/api/local/status": "Local API status",
				"/api/external/*":   "Proxy to JSONPlaceholder API",
				"/api/secure/*":     "Proxy to HTTPBin API",
				"/api/github/*":     "Proxy to GitHub API",
			},
			"usage": map[string]string{
				"external_api": "curl http://localhost:8080/api/external/posts/1",
				"httpbin_api":  "curl http://localhost:8080/api/secure/get",
				"github_api":   "curl http://localhost:8080/api/github/users/octocat",
			},
		})
	})

	// Mount reverse proxies

	// 1. Basic reverse proxy - forwards to external API
	// Routes like /api/external/posts/1 -> https://jsonplaceholder.typicode.com/posts/1
	app.MountReverseProxy("/api/external", "https://jsonplaceholder.typicode.com", false)

	// 2. HTTPBin proxy for testing (without middleware for now)
	// Routes like /api/secure/get -> https://httpbin.org/get
	app.MountReverseProxy("/api/secure", "https://httpbin.org", false, "proxy_logger", "proxy_auth")

	// 3. GitHub API proxy
	// Routes like /api/github/users/octocat -> https://api.github.com/users/octocat
	app.MountReverseProxy("/api/github", "https://api.github.com", false)

	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("🔄 Reverse Proxy Configuration:")
	lokstra.Logger.Infof("  /api/external/* → https://jsonplaceholder.typicode.com/*")
	lokstra.Logger.Infof("  /api/secure/*   → https://httpbin.org/*")
	lokstra.Logger.Infof("  /api/github/*   → https://api.github.com/*")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("📋 Test Commands:")
	lokstra.Logger.Infof("  # JSONPlaceholder API proxy")
	lokstra.Logger.Infof("  curl http://localhost:8080/api/external/posts/1")
	lokstra.Logger.Infof("  curl http://localhost:8080/api/external/users")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("  # HTTPBin API proxy")
	lokstra.Logger.Infof("  curl http://localhost:8080/api/secure/get")
	lokstra.Logger.Infof("  curl http://localhost:8080/api/secure/json")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("  # GitHub API proxy")
	lokstra.Logger.Infof("  curl http://localhost:8080/api/github/users/octocat")
	lokstra.Logger.Infof("  curl http://localhost:8080/api/github/repos/microsoft/vscode")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Visit http://localhost:8080 for more information")

	// Start the app with proper error handling
	if err := app.Start(); err != nil {
		lokstra.Logger.Errorf("Failed to start server: %v", err)
	}
}
