package main

import (
	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/services/logger"
)

// This example demonstrates basic service registration and retrieval.
// It shows how to register service factories, create service instances,
// and use them in handlers with type-safe dependency injection.
//
// Learning Objectives:
// - Understand service factory registration
// - Learn service creation and configuration
// - Explore type-safe service retrieval
// - See service usage in HTTP handlers
//
// Documentation: https://github.com/primadi/lokstra/blob/main/docs/services.md

func main() {
	regCtx := lokstra.NewGlobalRegistrationContext()
	app := lokstra.NewApp(regCtx, "services-basic-app", ":8080")

	// ===== Service Registration =====

	// Register the logger service module
	regCtx.RegisterModule(logger.GetModule)

	// Create a logger service instance
	_, err := regCtx.CreateService("lokstra.logger", "app-logger", true, "info")
	if err != nil {
		lokstra.Logger.Fatalf("Failed to create logger service: %v", err)
	}

	// Create a logger with custom configuration
	loggerConfig := map[string]any{
		"level":  "debug",
		"format": "text", // or "json"
	}
	_, err = regCtx.CreateService("lokstra.logger", "debug-logger", true, loggerConfig)
	if err != nil {
		lokstra.Logger.Fatalf("Failed to create debug logger: %v", err)
	}

	// Register a custom service factory
	regCtx.RegisterServiceFactory("counter", func(config any) (service.Service, error) {
		name := "default"
		if config != nil {
			if configMap, ok := config.(map[string]any); ok {
				if n, exists := configMap["name"]; exists {
					name = n.(string)
				}
			}
		}
		return &CounterService{
			name:  name,
			count: 0,
		}, nil
	})

	// Create custom service instances
	_, err = regCtx.CreateService("counter", "request-counter", true, map[string]any{
		"name": "HTTP Requests",
	})
	if err != nil {
		lokstra.Logger.Fatalf("Failed to create counter service: %v", err)
	}

	_, err = regCtx.CreateService("counter", "user-counter", true, map[string]any{
		"name": "Active Users",
	})
	if err != nil {
		lokstra.Logger.Fatalf("Failed to create user counter: %v", err)
	}

	// ===== HTTP Handlers with Service Usage =====

	// Home endpoint - demonstrates basic service retrieval
	app.GET("/", func(ctx *lokstra.Context) error {
		// Get logger service with type safety
		logger, err := lokstra.GetService[serviceapi.Logger](regCtx, "app-logger")
		if err != nil {
			return ctx.ErrorInternal("Logger service unavailable")
		}

		// Get counter service
		counterService, err := regCtx.GetService("request-counter")
		if err != nil {
			return ctx.ErrorInternal("Counter service unavailable")
		}
		counter := counterService.(*CounterService)

		// Use the services
		logger.Infof("Home endpoint accessed")
		counter.Increment()

		return ctx.Ok(map[string]any{
			"message":       "Services Basic Example",
			"request_count": counter.GetCount(),
			"services": []string{
				"app-logger (Logger)",
				"debug-logger (Logger)",
				"request-counter (Counter)",
				"user-counter (Counter)",
			},
		})
	})

	// Service info endpoint
	app.GET("/services", func(ctx *lokstra.Context) error {
		logger, err := lokstra.GetService[serviceapi.Logger](regCtx, "debug-logger")
		if err != nil {
			return ctx.ErrorInternal("Debug logger unavailable")
		}

		// Get both counter services
		requestCounterService, err := regCtx.GetService("request-counter")
		if err != nil {
			return ctx.ErrorInternal("Request counter unavailable")
		}
		requestCounter := requestCounterService.(*CounterService)

		userCounterService, err := regCtx.GetService("user-counter")
		if err != nil {
			return ctx.ErrorInternal("User counter unavailable")
		}
		userCounter := userCounterService.(*CounterService)

		logger.Debugf("Services info requested")

		return ctx.Ok(map[string]any{
			"registered_services": map[string]any{
				"app-logger": map[string]any{
					"type":        "Logger",
					"level":       "info",
					"description": "Main application logger",
				},
				"debug-logger": map[string]any{
					"type":        "Logger",
					"level":       "debug",
					"description": "Debug level logger",
				},
				"request-counter": map[string]any{
					"type":        "Counter",
					"name":        requestCounter.GetSetting("name"),
					"count":       requestCounter.GetCount(),
					"description": "Tracks HTTP request count",
				},
				"user-counter": map[string]any{
					"type":        "Counter",
					"name":        userCounter.GetSetting("name"),
					"count":       userCounter.GetCount(),
					"description": "Tracks active user count",
				},
			},
		})
	})

	// Counter operations
	app.POST("/counters/:name/increment", func(ctx *lokstra.Context) error {
		counterName := ctx.GetPathParam("name")
		serviceName := counterName + "-counter"

		logger, err := lokstra.GetService[serviceapi.Logger](regCtx, "app-logger")
		if err != nil {
			return ctx.ErrorInternal("Logger unavailable")
		}

		counterService, err := regCtx.GetService(serviceName)
		if err != nil {
			logger.Errorf("Counter service '%s' not found", serviceName)
			return ctx.ErrorNotFound("Counter not found")
		}
		counter := counterService.(*CounterService)

		oldCount := counter.GetCount()
		counter.Increment()
		newCount := counter.GetCount()

		logger.Infof("Counter '%s' incremented from %d to %d", counterName, oldCount, newCount)

		return ctx.Ok(map[string]any{
			"message":   "Counter incremented",
			"counter":   counterName,
			"old_count": oldCount,
			"new_count": newCount,
		})
	})

	app.POST("/counters/:name/reset", func(ctx *lokstra.Context) error {
		counterName := ctx.GetPathParam("name")
		serviceName := counterName + "-counter"

		logger, err := lokstra.GetService[serviceapi.Logger](regCtx, "app-logger")
		if err != nil {
			return ctx.ErrorInternal("Logger unavailable")
		}

		counterService, err := regCtx.GetService(serviceName)
		if err != nil {
			logger.Errorf("Counter service '%s' not found", serviceName)
			return ctx.ErrorNotFound("Counter not found")
		}
		counter := counterService.(*CounterService)

		oldCount := counter.GetCount()
		counter.Reset()

		logger.Infof("Counter '%s' reset from %d to 0", counterName, oldCount)

		return ctx.Ok(map[string]any{
			"message":   "Counter reset",
			"counter":   counterName,
			"old_count": oldCount,
			"new_count": 0,
		})
	})

	// Service creation endpoint with smart binding
	app.POST("/services/counters", func(ctx *lokstra.Context, req *CreateCounterRequest) error {
		logger, err := lokstra.GetService[serviceapi.Logger](regCtx, "app-logger")
		if err != nil {
			return ctx.ErrorInternal("Logger unavailable")
		}

		// Check if service already exists
		_, err = regCtx.GetService(req.ServiceName)
		if err == nil {
			return ctx.ErrorDuplicate("Service already exists")
		}

		// Create new counter service
		_, err = regCtx.CreateService("counter", req.ServiceName, true, map[string]any{
			"name": req.Name,
		})
		if err != nil {
			logger.Errorf("Failed to create counter service: %v", err)
			return ctx.ErrorInternal("Failed to create service")
		}

		logger.Infof("Created new counter service: %s", req.ServiceName)

		return ctx.OkCreated(map[string]any{
			"message":      "Counter service created",
			"service_name": req.ServiceName,
			"counter_name": req.Name,
		})
	})

	// Service stats endpoint
	app.GET("/stats", func(ctx *lokstra.Context) error {
		logger, err := lokstra.GetService[serviceapi.Logger](regCtx, "debug-logger")
		if err != nil {
			return ctx.ErrorInternal("Logger unavailable")
		}

		// For simplicity, we'll manually count our known services
		// In a real application, you might maintain a registry
		serviceCount := 4 // app-logger, debug-logger, request-counter, user-counter
		counterServices := 2
		loggerServices := 2
		totalRequests := 0

		// Get request counter for total
		if requestCounterService, err := regCtx.GetService("request-counter"); err == nil {
			if counter, ok := requestCounterService.(*CounterService); ok {
				totalRequests = counter.GetCount()
			}
		}

		logger.Debugf("Stats requested - %d total services", serviceCount)

		return ctx.Ok(map[string]any{
			"service_statistics": map[string]any{
				"total_services":   serviceCount,
				"counter_services": counterServices,
				"logger_services":  loggerServices,
				"total_requests":   totalRequests,
			},
			"known_factories": []string{
				"lokstra.logger",
				"counter",
			},
		})
	})

	lokstra.Logger.Infof("🚀 Services Basic Example started on :8080")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Available Endpoints:")
	lokstra.Logger.Infof("  GET  /                           - Home with service demo")
	lokstra.Logger.Infof("  GET  /services                   - List all services")
	lokstra.Logger.Infof("  POST /counters/:name/increment   - Increment counter (request, user)")
	lokstra.Logger.Infof("  POST /counters/:name/reset       - Reset counter")
	lokstra.Logger.Infof("  POST /services/counters          - Create new counter service")
	lokstra.Logger.Infof("  GET  /stats                      - Service statistics")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Example Services:")
	lokstra.Logger.Infof("  - app-logger: Main application logger")
	lokstra.Logger.Infof("  - debug-logger: Debug level logger")
	lokstra.Logger.Infof("  - request-counter: HTTP request counter")
	lokstra.Logger.Infof("  - user-counter: Active user counter")

	app.Start(true)
}

// ===== Request Types =====

type CreateCounterRequest struct {
	Name        string `json:"name" validate:"required"`
	ServiceName string `json:"service_name" validate:"required"`
}

// ===== Custom Service Implementation =====

// CounterService implements a simple counter service
type CounterService struct {
	name  string
	count int
}

func (c *CounterService) GetSetting(key string) any {
	switch key {
	case "name":
		return c.name
	case "count":
		return c.count
	default:
		return nil
	}
}

func (c *CounterService) GetCount() int {
	return c.count
}

func (c *CounterService) Increment() {
	c.count++
}

func (c *CounterService) Reset() {
	c.count = 0
}

func (c *CounterService) SetCount(count int) {
	c.count = count
}

// Service System Key Concepts:
//
// 1. Service Interface:
//    - All services implement service.Service interface
//    - GetSetting(key string) any method for configuration access
//    - Services can expose additional methods
//
// 2. Service Registration:
//    - RegisterServiceFactory() registers factory functions
//    - RegisterModule() registers groups of services
//    - RegisterService() registers direct instances
//
// 3. Service Creation:
//    - CreateService() creates instances from factories
//    - GetOrCreateService() creates or returns existing
//    - Configuration passed to factory functions
//
// 4. Service Retrieval:
//    - GetService() returns service instances
//    - lokstra.GetService[T]() provides type safety
//    - Type assertions for custom services
//
// 5. Service Lifecycle:
//    - Services created once and reused
//    - Factory functions handle configuration
//    - Services available throughout application

// Test Commands:
//
// # Start the server
// go run main.go
//
// # Test basic endpoints
// curl http://localhost:8080/
// curl http://localhost:8080/services
// curl http://localhost:8080/stats
//
// # Test counter operations
// curl -X POST http://localhost:8080/counters/request/increment
// curl -X POST http://localhost:8080/counters/user/increment
// curl -X POST http://localhost:8080/counters/request/reset
//
// # Create new counter service
// curl -X POST http://localhost:8080/services/counters \
//   -H "Content-Type: application/json" \
//   -d '{"name":"API Calls","service_name":"api-counter"}'
//
// # Use the new counter
// curl -X POST http://localhost:8080/counters/api/increment
//
// # Check updated stats
// curl http://localhost:8080/stats
