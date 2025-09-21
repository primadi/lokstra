package main

import (
	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/services/logger"
)

// This example demonstrates comprehensive usage of Lokstra's built-in logger service.
// It shows different logger configurations, logging levels, structured logging,
// and best practices for application logging.
//
// Learning Objectives:
// - Understand logger service configuration options
// - Learn different logging levels and their usage
// - Explore structured logging patterns
// - See logger integration in HTTP handlers
// - Master logger service lifecycle
//
// Documentation: https://github.com/primadi/lokstra/blob/main/docs/services.md

func main() {
	regCtx := lokstra.NewGlobalRegistrationContext()
	app := lokstra.NewApp(regCtx, "logger-service-app", ":8080")

	// ===== Logger Service Registration =====

	// Register the logger service module
	regCtx.RegisterModule(logger.GetModule)

	// Create different logger configurations

	// 1. Basic logger with string level
	factoryName := logger.GetModule().Name()
	basicLogger, err := registration.CreateService[serviceapi.Logger](regCtx, factoryName, "basic-logger", true, "info")
	if err != nil {
		lokstra.Logger.Fatalf("Failed to create basic logger: %v", err)
	}

	// 2. Debug logger with detailed configuration
	debugConfig := map[string]any{
		"level":  "debug",
		"format": "text", // or "json"
	}
	debugLogger, err := registration.CreateService[serviceapi.Logger](regCtx, factoryName, "debug-logger", true, debugConfig)
	if err != nil {
		lokstra.Logger.Fatalf("Failed to create debug logger: %v", err)
	}

	// 3. JSON logger for production-like logging
	jsonConfig := map[string]any{
		"level":  "warn",
		"format": "json",
	}
	jsonLogger, err := registration.CreateService[serviceapi.Logger](regCtx, factoryName, "json-logger", true, jsonConfig)
	if err != nil {
		lokstra.Logger.Fatalf("Failed to create JSON logger: %v", err)
	}

	// 4. Error-only logger
	errorConfig := map[string]any{
		"level":  "error",
		"format": "text",
	}
	errorLogger, err := registration.CreateService[serviceapi.Logger](regCtx, factoryName, "error-logger", true, errorConfig)
	if err != nil {
		lokstra.Logger.Fatalf("Failed to create error logger: %v", err)
	}

	// ===== HTTP Handlers with Logger Usage =====

	// Home endpoint - basic logging
	app.GET("/", func(ctx *lokstra.Context) error {
		logger, err := lokstra.GetService[serviceapi.Logger](regCtx, "basic-logger")
		if err != nil {
			return ctx.ErrorInternal("Logger service unavailable")
		}

		logger.Infof("Home endpoint accessed from %s", ctx.Request.RemoteAddr)

		return ctx.Ok(map[string]any{
			"message": "Logger Service Example",
			"loggers": []string{
				"basic-logger (" + basicLogger.GetLogLevel().String() + ")",
				"debug-logger (" + debugLogger.GetLogLevel().String() + ")",
				"json-logger (" + jsonLogger.GetLogLevel().String() + ")",
				"error-logger (" + errorLogger.GetLogLevel().String() + ")",
			},
			"client_ip": ctx.Request.RemoteAddr,
		})
	})

	// Logging levels demonstration
	app.GET("/log-levels", func(ctx *lokstra.Context) error {
		// Demonstrate all logging levels
		debugLogger.Debugf("Debug message - detailed information for debugging")
		debugLogger.Infof("Info message - general application information")
		debugLogger.Warnf("Warning message - potentially harmful situation")
		debugLogger.Errorf("Error message - error occurred but application continues")

		// Note: Fatalf would terminate the application, so we don't use it here

		return ctx.Ok(map[string]any{
			"message": "All logging levels demonstrated",
			"levels": map[string]string{
				"debug": "Detailed debugging information",
				"info":  "General application information",
				"warn":  "Potentially harmful situations",
				"error": "Error conditions",
				"fatal": "Critical errors that cause termination",
			},
			"note": "Check server logs to see the output",
		})
	})

	// Structured logging with context
	app.POST("/users/:id/update", func(ctx *lokstra.Context) error {
		userID := ctx.GetPathParam("id")

		// Simulate user update process with structured logging
		jsonLogger.Infof("User update initiated - ID: %s, IP: %s", userID, ctx.Request.RemoteAddr)

		// Simulate validation
		if userID == "invalid" {
			jsonLogger.Warnf("Invalid user ID provided: %s", userID)
			return ctx.ErrorBadRequest("Invalid user ID")
		}

		// Simulate database operation
		if userID == "error" {
			jsonLogger.Errorf("Database error during user update - ID: %s", userID)
			return ctx.ErrorInternal("Database operation failed")
		}

		// Success case
		jsonLogger.Infof("User updated successfully - ID: %s", userID)

		return ctx.Ok(map[string]any{
			"message": "User updated successfully",
			"user_id": userID,
			"logged":  "Check JSON logs for structured output",
		})
	})

	// Error logging demonstration
	app.GET("/error-demo", func(ctx *lokstra.Context) error {
		// Simulate various error scenarios
		errorLogger.Errorf("Simulated database connection error")
		errorLogger.Errorf("Simulated API rate limit exceeded")
		errorLogger.Errorf("Simulated authentication failure")

		return ctx.Ok(map[string]any{
			"message": "Error scenarios logged",
			"note":    "Error logger only shows warn and error level messages",
		})
	})

	// Logger configuration info
	app.GET("/logger-info", func(ctx *lokstra.Context) error {
		basicLogger.Infof("Logger configuration requested")

		return ctx.Ok(map[string]any{
			"loggers": map[string]any{
				"basic-logger": map[string]any{
					"level":       "info",
					"format":      "text",
					"description": "Standard application logging",
				},
				"debug-logger": map[string]any{
					"level":       "debug",
					"format":      "text",
					"description": "Detailed debugging information",
				},
				"json-logger": map[string]any{
					"level":       "warn",
					"format":      "json",
					"description": "Structured JSON logging for production",
				},
				"error-logger": map[string]any{
					"level":       "error",
					"format":      "text",
					"description": "Error-only logging",
				},
			},
			"level_hierarchy": []string{
				"debug (most verbose)",
				"info",
				"warn",
				"error",
				"fatal (least verbose)",
			},
		})
	})

	// Benchmark logging performance
	app.GET("/log-benchmark", func(ctx *lokstra.Context) error {
		// Log multiple messages to demonstrate performance
		count := 100
		for i := range count {
			debugLogger.Debugf("Benchmark log message %d", i)
		}

		debugLogger.Infof("Logged %d benchmark messages", count)

		return ctx.Ok(map[string]any{
			"message":     "Logging benchmark completed",
			"log_count":   count,
			"performance": "Check server logs for timing",
		})
	})

	// Smart binding with logging context
	app.POST("/log-request", func(ctx *lokstra.Context, req *LogRequest) error {
		// Log based on request level
		switch req.Level {
		case "debug":
			debugLogger.Debugf("Custom log: %s", req.Message)
		case "info":
			basicLogger.Infof("Custom log: %s", req.Message)
		case "warn":
			jsonLogger.Warnf("Custom log: %s", req.Message)
		case "error":
			errorLogger.Errorf("Custom log: %s", req.Message)
		default:
			basicLogger.Infof("Custom log (default): %s", req.Message)
		}

		return ctx.Ok(map[string]any{
			"message": "Custom log message recorded",
			"level":   req.Level,
			"logged":  req.Message,
		})
	})

	lokstra.Logger.Infof("🚀 Logger Service Example started on :8080")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Available Endpoints:")
	lokstra.Logger.Infof("  GET  /                    - Home with logger overview")
	lokstra.Logger.Infof("  GET  /log-levels          - Demonstrate all logging levels")
	lokstra.Logger.Infof("  POST /users/:id/update    - Structured logging example")
	lokstra.Logger.Infof("  GET  /error-demo          - Error logging demonstration")
	lokstra.Logger.Infof("  GET  /logger-info         - Logger configuration info")
	lokstra.Logger.Infof("  GET  /log-benchmark       - Logging performance test")
	lokstra.Logger.Infof("  POST /log-request         - Custom log with level selection")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Logger Configurations:")
	lokstra.Logger.Infof("  - basic-logger: info level, text format")
	lokstra.Logger.Infof("  - debug-logger: debug level, text format")
	lokstra.Logger.Infof("  - json-logger: warn level, JSON format")
	lokstra.Logger.Infof("  - error-logger: error level, text format")

	app.Start(true)
}

// ===== Request Types =====

type LogRequest struct {
	Level   string `json:"level" validate:"required,oneof=debug info warn error"`
	Message string `json:"message" validate:"required,min=1,max=500"`
}

// Logger Service Key Concepts:
//
// 1. Logger Configuration:
//    - Level: debug, info, warn, error, fatal
//    - Format: text or json
//    - String config: just the level
//    - Map config: level, format, and other options
//
// 2. Logging Levels:
//    - Debug: Detailed debugging information
//    - Info: General application information
//    - Warn: Potentially harmful situations
//    - Error: Error conditions
//    - Fatal: Critical errors (terminates app)
//
// 3. Logger Methods:
//    - Debugf, Infof, Warnf, Errorf, Fatalf
//    - Printf-style formatting
//    - Automatic level filtering
//    - Context-aware logging
//
// 4. Best Practices:
//    - Use appropriate log levels
//    - Include context in log messages
//    - Use structured logging for production
//    - Avoid logging sensitive information
//    - Consider performance implications
//
// 5. Service Integration:
//    - Type-safe retrieval with lokstra.GetService[serviceapi.Logger]
//    - Multiple logger instances with different configs
//    - Service lifecycle managed by framework
//    - Available across all handlers

// Test Commands:
//
// # Start the server
// go run main.go
//
// # Test basic endpoints
// curl http://localhost:8080/
// curl http://localhost:8080/log-levels
// curl http://localhost:8080/logger-info
//
// # Test structured logging
// curl -X POST http://localhost:8080/users/123/update
// curl -X POST http://localhost:8080/users/invalid/update
// curl -X POST http://localhost:8080/users/error/update
//
// # Test error logging
// curl http://localhost:8080/error-demo
//
// # Test logging benchmark
// curl http://localhost:8080/log-benchmark
//
// # Test custom logging
// curl -X POST http://localhost:8080/log-request \
//   -H "Content-Type: application/json" \
//   -d '{"level":"info","message":"Custom info message"}'
//
// curl -X POST http://localhost:8080/log-request \
//   -H "Content-Type: application/json" \
//   -d '{"level":"error","message":"Custom error message"}'
//
// curl -X POST http://localhost:8080/log-request \
//   -H "Content-Type: application/json" \
//   -d '{"level":"debug","message":"Custom debug message"}'
//
// # Check server console output to see different log formats and levels
