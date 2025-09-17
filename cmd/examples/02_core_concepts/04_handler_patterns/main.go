package main

import (
	"context"
	"errors"
	"time"

	"github.com/primadi/lokstra"
)

// This example demonstrates various handler patterns in Lokstra.
// It shows different ways to structure handlers, parameter binding, middleware, and response patterns.
//
// Learning Objectives:
// - Understand different handler signature patterns
// - Learn parameter binding variations
// - Explore middleware integration with handlers
// - See advanced handler patterns and best practices
//
// Documentation: https://github.com/primadi/lokstra/blob/main/docs/core-concepts.md#handler-patterns

// ===== Basic Handler Patterns =====

// 1. Simple handler with no parameters
func simpleHandler(ctx *lokstra.Context) error {
	return ctx.Ok("Simple handler response")
}

// 2. Handler with path parameters
func userHandler(ctx *lokstra.Context) error {
	userID := ctx.GetPathParam("id")
	return ctx.Ok(map[string]interface{}{
		"message": "User handler",
		"user_id": userID,
	})
}

// ===== Smart Binding Handler Patterns =====

// Path parameter binding
type UserPathParams struct {
	ID int `path:"id"`
}

func userWithPathBinding(ctx *lokstra.Context, params *UserPathParams) error {
	return ctx.Ok(map[string]interface{}{
		"message": "User with path binding",
		"user_id": params.ID,
	})
}

// Query parameter binding
type SearchQueryParams struct {
	Query  string `query:"q"`
	Page   int    `query:"page" default:"1"`
	Limit  int    `query:"limit" default:"10"`
	SortBy string `query:"sort" default:"name"`
}

func searchWithQueryBinding(ctx *lokstra.Context, params *SearchQueryParams) error {
	return ctx.Ok(map[string]interface{}{
		"message": "Search with query binding",
		"query":   params.Query,
		"page":    params.Page,
		"limit":   params.Limit,
		"sort_by": params.SortBy,
	})
}

// JSON body binding
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"min=18,max=120"`
}

func createUserWithBodyBinding(ctx *lokstra.Context, req *CreateUserRequest) error {
	return ctx.OkCreated(map[string]interface{}{
		"message": "User created with body binding",
		"user":    req,
	})
}

// Combined parameter binding
type UpdateUserCombined struct {
	ID   int                `path:"id"`
	Data *CreateUserRequest `json:",inline"`
}

func updateUserCombined(ctx *lokstra.Context, params *UpdateUserCombined) error {
	return ctx.OkUpdated(map[string]interface{}{
		"message": "User updated with combined binding",
		"id":      params.ID,
		"data":    params.Data,
	})
}

// ===== Advanced Handler Patterns =====

// Handler with service dependency injection
func handlerWithServices(ctx *lokstra.Context, regCtx lokstra.RegistrationContext) error {
	// Get logger service
	logger, err := lokstra.GetService[interface{}](regCtx, "logger")
	if err != nil {
		return ctx.ErrorInternal("Logger service not available")
	}

	return ctx.Ok(map[string]interface{}{
		"message": "Handler with service injection",
		"logger":  logger != nil,
	})
}

// Handler with custom validation
type CustomValidationRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r *CustomValidationRequest) Validate() error {
	if len(r.Username) < 3 {
		return errors.New("Username must be at least 3 characters")
	}
	if len(r.Password) < 8 {
		return errors.New("Password must be at least 8 characters")
	}
	return nil
}

func handlerWithCustomValidation(ctx *lokstra.Context, req *CustomValidationRequest) error {
	// Custom validation would be called automatically by Lokstra
	// This is just for demonstration
	if err := req.Validate(); err != nil {
		return ctx.ErrorBadRequest(err.Error())
	}

	return ctx.Ok(map[string]interface{}{
		"message":  "Custom validation passed",
		"username": req.Username,
	})
}

// ===== Middleware-Aware Handler Patterns =====

// Handler that uses middleware-set context values
func handlerWithMiddlewareData(ctx *lokstra.Context) error {
	// Get data set by middleware (if context keys were properly set)
	userID := "unknown"
	if val := ctx.Value("user_id"); val != nil {
		userID = val.(string)
	}

	startTime := "unknown"
	if val := ctx.Value("request_start"); val != nil {
		if t, ok := val.(time.Time); ok {
			startTime = t.Format(time.RFC3339)
		}
	}

	return ctx.Ok(map[string]interface{}{
		"message":         "Handler using middleware data",
		"user_id":         userID,
		"request_start":   startTime,
		"processing_time": "calculated by middleware",
	})
}

// ===== Error Handling Patterns =====

func handlerWithErrorHandling(ctx *lokstra.Context) error {
	// Demonstrate different error response patterns
	errorType := ctx.GetQueryParam("error")

	switch errorType {
	case "validation":
		return ctx.ErrorValidation("Validation failed", map[string]string{
			"field1": "Field1 is required",
			"field2": "Field2 must be valid email",
		})
	case "notfound":
		return ctx.ErrorNotFound("Resource not found")
	case "badrequest":
		return ctx.ErrorBadRequest("Invalid request parameters")
	case "duplicate":
		return ctx.ErrorDuplicate("Resource already exists")
	case "internal":
		return ctx.ErrorInternal("Internal server error occurred")
	default:
		return ctx.Ok(map[string]interface{}{
			"message": "No error - success response",
			"available_errors": []string{
				"validation", "notfound", "badrequest", "duplicate", "internal",
			},
		})
	}
}

// ===== Async/Background Handler Patterns =====

func asyncHandler(ctx *lokstra.Context) error {
	// Start background processing
	go func() {
		time.Sleep(2 * time.Second)
		lokstra.Logger.Infof("Background task completed")
	}()

	return ctx.Ok(map[string]interface{}{
		"message": "Background task started",
		"status":  "processing",
	})
}

// ===== Response Pattern Variations =====

func responsePatternHandler(ctx *lokstra.Context) error {
	format := ctx.GetQueryParam("format")

	switch format {
	case "list":
		data := []string{"item1", "item2", "item3"}
		meta := map[string]interface{}{"total": 3, "page": 1}
		return ctx.OkList(data, meta)

	case "created":
		newResource := map[string]interface{}{"id": 123, "name": "New Resource"}
		return ctx.OkCreated(newResource)

	case "updated":
		updatedResource := map[string]interface{}{"id": 456, "name": "Updated Resource"}
		return ctx.OkUpdated(updatedResource)

	case "html":
		return ctx.HTML("<h1>HTML Response</h1><p>This is HTML content</p>")

	case "raw":
		return ctx.WriteRaw("text/plain", 200, []byte("Raw text response"))

	default:
		return ctx.Ok(map[string]interface{}{
			"message": "Standard response",
			"available_formats": []string{
				"list", "created", "updated", "html", "raw",
			},
		})
	}
}

func main() {
	regCtx := lokstra.NewGlobalRegistrationContext()
	app := lokstra.NewApp(regCtx, "handler-patterns-app", ":8080")

	// Middleware to set context values for demonstration
	app.Use(func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
		// Simulate setting user ID from authentication
		ctx.Context = context.WithValue(ctx.Context, "user_id", "user123")

		// Set request start time
		ctx.Context = context.WithValue(ctx.Context, "request_start", time.Now())

		return next(ctx)
	})

	// ===== Register Handler Patterns =====

	// Basic patterns
	app.GET("/simple", simpleHandler)
	app.GET("/user-basic/:id", userHandler)

	// Smart binding patterns
	app.GET("/user/:id", userWithPathBinding)
	app.GET("/search", searchWithQueryBinding)
	app.POST("/users", createUserWithBodyBinding)
	app.PUT("/users/:id", updateUserCombined)

	// Advanced patterns
	app.GET("/with-services", func(ctx *lokstra.Context) error {
		return handlerWithServices(ctx, regCtx)
	})
	app.POST("/custom-validation", handlerWithCustomValidation)

	// Middleware-aware patterns
	app.GET("/middleware-data", handlerWithMiddlewareData)

	// Error handling patterns
	app.GET("/error-demo", handlerWithErrorHandling)

	// Async patterns
	app.GET("/async", asyncHandler)

	// Response patterns
	app.GET("/response-patterns", responsePatternHandler)

	// ===== Handler with inline definition =====
	app.GET("/inline", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]interface{}{
			"message": "Inline handler definition",
			"pattern": "anonymous function",
		})
	})

	// ===== Handler with closure =====
	message := "Closure message"
	app.GET("/closure", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]interface{}{
			"message": message,
			"pattern": "closure handler",
		})
	})

	lokstra.Logger.Infof("🚀 Handler Patterns Example started on :8080")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Handler Pattern Examples:")
	lokstra.Logger.Infof("  Basic Patterns:")
	lokstra.Logger.Infof("    GET  /simple              - Simple handler")
	lokstra.Logger.Infof("    GET  /user-basic/123      - Basic path parameter")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("  Smart Binding Patterns:")
	lokstra.Logger.Infof("    GET  /user/123            - Path parameter binding")
	lokstra.Logger.Infof("    GET  /search?q=test&page=2 - Query parameter binding")
	lokstra.Logger.Infof("    POST /users               - JSON body binding")
	lokstra.Logger.Infof("    PUT  /users/123           - Combined binding")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("  Advanced Patterns:")
	lokstra.Logger.Infof("    GET  /with-services       - Service injection")
	lokstra.Logger.Infof("    POST /custom-validation   - Custom validation")
	lokstra.Logger.Infof("    GET  /middleware-data     - Middleware context")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("  Error & Response Patterns:")
	lokstra.Logger.Infof("    GET  /error-demo?error=validation - Error handling")
	lokstra.Logger.Infof("    GET  /response-patterns?format=list - Response formats")
	lokstra.Logger.Infof("    GET  /async                - Async processing")

	app.Start()
}

// Handler Pattern Best Practices:
//
// 1. Parameter Binding:
//    - Use struct binding for type safety
//    - Apply validation tags for automatic validation
//    - Use default values for optional parameters
//    - Combine different parameter sources when needed
//
// 2. Service Integration:
//    - Inject services through registration context
//    - Keep handlers focused on HTTP concerns
//    - Move business logic to service layer
//    - Handle service errors gracefully
//
// 3. Error Handling:
//    - Use appropriate error response methods
//    - Provide meaningful error messages
//    - Include field-level errors for validation
//    - Log internal errors without exposing details
//
// 4. Middleware Integration:
//    - Use context values for cross-cutting concerns
//    - Keep middleware stateless when possible
//    - Handle middleware errors properly
//    - Document context value contracts
//
// 5. Response Patterns:
//    - Use structured responses for consistency
//    - Choose appropriate HTTP status codes
//    - Include metadata for lists and pagination
//    - Support multiple response formats when needed

// Test Commands:
//
// # Basic patterns
// curl http://localhost:8080/simple
// curl http://localhost:8080/user-basic/123
//
// # Smart binding
// curl http://localhost:8080/user/456
// curl "http://localhost:8080/search?q=golang&page=2&limit=5"
// curl -X POST http://localhost:8080/users -H "Content-Type: application/json" -d '{"name":"John","email":"john@example.com","age":25}'
// curl -X PUT http://localhost:8080/users/123 -H "Content-Type: application/json" -d '{"name":"Updated","email":"updated@example.com","age":30}'
//
// # Advanced patterns
// curl http://localhost:8080/with-services
// curl -X POST http://localhost:8080/custom-validation -H "Content-Type: application/json" -d '{"username":"jo","password":"short"}'
// curl http://localhost:8080/middleware-data
//
// # Error handling
// curl "http://localhost:8080/error-demo?error=validation"
// curl "http://localhost:8080/error-demo?error=notfound"
//
// # Response patterns
// curl "http://localhost:8080/response-patterns?format=list"
// curl "http://localhost:8080/response-patterns?format=html"
// curl http://localhost:8080/async
