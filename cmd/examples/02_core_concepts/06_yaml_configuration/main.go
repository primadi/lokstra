package main

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/middleware/cors"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/services/logger"
)

// This example demonstrates how to run a Lokstra application from YAML configuration.
// It shows comprehensive YAML-based setup including server configuration, middleware,
// services, and route definitions.
//
// Learning Objectives:
// - Understand YAML configuration structure
// - Learn service configuration via YAML
// - See middleware setup through configuration
// - Explore server and application settings
// - Master environment variable integration
//
// Documentation: https://github.com/primadi/lokstra/blob/main/docs/configuration.md

func main() {
	fmt.Println("🚀 YAML Configuration Example - Loading from config.yaml")
	fmt.Println("")

	// Load configuration from YAML file
	config, err := lokstra.LoadConfigDir(".")
	if err != nil {
		lokstra.Logger.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Printf("✅ Configuration loaded successfully\n")
	fmt.Printf("📋 Server name: %s\n", config.Server.Name)
	fmt.Printf("📋 Apps configured: %d\n", len(config.Apps))
	fmt.Printf("📋 Services configured: %d\n", len(config.Services))
	fmt.Printf("📋 Modules configured: %d\n", len(config.Modules))
	fmt.Println("")

	// Create registration context
	regCtx := lokstra.NewGlobalRegistrationContext()

	// Register required service modules
	regCtx.RegisterModule(logger.GetModule)
	regCtx.RegisterModule(recovery.GetModule)
	regCtx.RegisterModule(cors.GetModule)

	// Register custom middleware functions
	registerCustomMiddleware(regCtx)

	// Register custom handlers
	registerHandlers(regCtx)

	// Start server from configuration
	server, err := lokstra.NewServerFromConfig(regCtx, config)
	if err != nil {
		lokstra.Logger.Fatalf("Failed to create server from config: %v", err)
	}

	fmt.Println("🎯 Starting server from YAML configuration...")
	fmt.Println("📡 Check the endpoints defined in config.yaml")
	fmt.Println("")

	// Start the server
	if err := server.Start(); err != nil {
		lokstra.Logger.Fatalf("Failed to start server: %v", err)
	}
}

// registerCustomMiddleware registers middleware functions referenced in YAML
func registerCustomMiddleware(regCtx lokstra.RegistrationContext) {
	// Request timing middleware
	regCtx.RegisterMiddlewareFunc("timing", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			start := time.Now()

			// Execute next handler
			err := next(ctx)

			// Add timing header
			duration := time.Since(start)
			ctx.WithHeader("X-Response-Time", duration.String())

			return err
		}
	})

	// Custom authentication middleware
	regCtx.RegisterMiddlewareFunc("custom_auth", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			apiKey := ctx.GetHeader("X-API-Key")
			if apiKey == "" {
				ctx.SetStatusCode(401)
				return ctx.ErrorBadRequest("API key required")
			}

			if apiKey != "yaml-config-key-123" {
				ctx.SetStatusCode(401)
				return ctx.ErrorBadRequest("Invalid API key")
			}

			return next(ctx)
		}
	})

	// Request validation middleware
	regCtx.RegisterMiddlewareFunc("validate_content", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			contentType := ctx.GetHeader("Content-Type")
			if ctx.Request.Method == "POST" || ctx.Request.Method == "PUT" {
				if contentType == "" {
					return ctx.ErrorBadRequest("Content-Type header required for POST/PUT requests")
				}
			}

			return next(ctx)
		}
	})
}

// registerHandlers registers handler functions referenced in YAML routes
func registerHandlers(regCtx lokstra.RegistrationContext) {
	// Home handler
	regCtx.RegisterHandler("home", func(ctx *lokstra.Context) error {
		logger, _ := serviceapi.GetService[serviceapi.Logger](regCtx, "app-logger")
		logger.Infof("Home endpoint accessed from YAML configuration")

		return ctx.Ok(map[string]interface{}{
			"message": "YAML Configuration Example",
			"config_info": map[string]interface{}{
				"loaded_from":           "config.yaml",
				"server_configured":     true,
				"services_configured":   true,
				"middleware_configured": true,
				"routes_configured":     true,
			},
			"endpoints": []string{
				"GET /",
				"GET /health",
				"GET /config-info",
				"POST /api/data",
				"GET /api/protected/profile",
			},
		})
	})

	// Health check handler
	regCtx.RegisterHandler("health_check", func(ctx *lokstra.Context) error {
		logger, _ := serviceapi.GetService[serviceapi.Logger](regCtx, "app-logger")
		logger.Debugf("Health check performed")

		return ctx.Ok(map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now(),
			"source":    "YAML configuration",
		})
	})

	// Configuration info handler
	regCtx.RegisterHandler("config_info", func(ctx *lokstra.Context) error {
		logger, _ := serviceapi.GetService[serviceapi.Logger](regCtx, "debug-logger")
		logger.Debugf("Configuration info requested")

		return ctx.Ok(map[string]interface{}{
			"configuration": map[string]interface{}{
				"format": "YAML",
				"file":   "config.yaml",
				"features": []string{
					"Server configuration",
					"Service registration",
					"Middleware setup",
					"Route definitions",
					"Environment variables",
				},
			},
			"services": map[string]interface{}{
				"app-logger":   "Application logger service",
				"debug-logger": "Debug logger service",
			},
			"middleware": []string{
				"timing",
				"custom_auth",
				"validate_content",
			},
		})
	})

	// Data handler with smart binding
	regCtx.RegisterHandler("data_handler", func(ctx *lokstra.Context, req *DataRequest) error {
		logger, _ := serviceapi.GetService[serviceapi.Logger](regCtx, "app-logger")
		logger.Infof("Data received: type=%s, from YAML config route", req.Type)

		return ctx.Ok(map[string]interface{}{
			"message": "Data processed successfully",
			"data": map[string]interface{}{
				"type":         req.Type,
				"content":      req.Content,
				"processed_at": time.Now(),
				"via":          "YAML configuration",
			},
		})
	})

	// Protected profile handler
	regCtx.RegisterHandler("protected_profile", func(ctx *lokstra.Context) error {
		logger, _ := serviceapi.GetService[serviceapi.Logger](regCtx, "app-logger")
		logger.Infof("Protected profile accessed with YAML auth")

		return ctx.Ok(map[string]interface{}{
			"profile": map[string]interface{}{
				"user":              "YAML Config User",
				"email":             "yaml.user@example.com",
				"role":              "demo",
				"authenticated_via": "YAML middleware",
			},
			"access_info": map[string]interface{}{
				"method":        "API Key Authentication",
				"configured_in": "config.yaml",
				"middleware":    "custom_auth",
			},
		})
	})
}

// Request types for smart binding
type DataRequest struct {
	Type    string `json:"type" validate:"required,oneof=user content analytics"`
	Content string `json:"content" validate:"required,min=1"`
}

// YAML Configuration Key Concepts:
//
// 1. Server Configuration:
//    - Server name and global settings
//    - Application definitions with addresses
//    - Listener types and configurations
//    - Environment variable integration
//
// 2. Service Configuration:
//    - Service definitions with factories
//    - Service-specific configurations
//    - Dependency declarations
//    - Lifecycle management
//
// 3. Middleware Configuration:
//    - Global middleware chains
//    - Application-specific middleware
//    - Route-specific middleware
//    - Middleware parameters and settings
//
// 4. Route Configuration:
//    - HTTP method and path definitions
//    - Handler function mappings
//    - Route-specific middleware
//    - Parameter validation
//
// 5. Environment Integration:
//    - Environment variable substitution
//    - Default value specifications
//    - Conditional configurations
//    - Deployment environment handling

// Configuration Benefits:
//
// 1. Declarative Setup:
//    - Infrastructure as code
//    - Version-controlled configuration
//    - Environment-specific configs
//    - No code changes for deployment
//
// 2. Centralized Management:
//    - Single configuration file
//    - Clear service dependencies
//    - Middleware orchestration
//    - Route organization
//
// 3. Environment Flexibility:
//    - Development vs production configs
//    - Environment variable integration
//    - Conditional service enabling
//    - Dynamic configuration loading
//
// 4. Maintenance Benefits:
//    - Reduced code complexity
//    - Configuration validation
//    - Documentation through config
//    - Easier troubleshooting

// Test Commands:
//
// # Start the application
// go run main.go
//
// # Test public endpoints
// curl http://localhost:8080/
// curl http://localhost:8080/health
// curl http://localhost:8080/config-info
//
// # Test protected endpoint (requires API key)
// curl http://localhost:8080/api/protected/profile
// # Should fail with 401
//
// curl -H "X-API-Key: yaml-config-key-123" \
//   http://localhost:8080/api/protected/profile
// # Should succeed
//
// # Test data endpoint with JSON
// curl -X POST http://localhost:8080/api/data \
//   -H "Content-Type: application/json" \
//   -d '{"type":"user","content":"Test data from YAML config"}'
//
// # Test data endpoint without Content-Type (should fail)
// curl -X POST http://localhost:8080/api/data \
//   -d '{"type":"user","content":"Test"}'
//
// # Test with different environment variables
// LOG_LEVEL=debug go run main.go
// API_PORT=9090 go run main.go
