package main

import (
	"context"
	"time"

	"github.com/primadi/lokstra"
)

// This example demonstrates Lokstra's request context and its capabilities.
// It shows request/response handling, parameter extraction, headers, and context values.
//
// Learning Objectives:
// - Understand the request context structure
// - Learn parameter extraction (path, query, headers)
// - Explore request body handling
// - See context value storage and retrieval
//
// Documentation: https://github.com/primadi/lokstra/blob/main/docs/core-concepts.md#request-context

// Context keys for type safety
type contextKey string

const (
	requestStartKey   contextKey = "request_start"
	requestIDKey      contextKey = "request_id"
	processingTimeKey contextKey = "processing_time"
)

func main() {
	regCtx := lokstra.NewGlobalRegistrationContext()
	app := lokstra.NewApp(regCtx, "request-context-app", ":8080")

	// Middleware to demonstrate context value storage
	app.Use(func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
		// Store request start time in context
		start := time.Now()
		newCtx := context.WithValue(ctx.Context, requestStartKey, start)

		// Store request ID for tracking
		requestID := "req-" + time.Now().Format("20060102-150405")
		newCtx = context.WithValue(newCtx, requestIDKey, requestID)

		// Update the context
		ctx.Context = newCtx

		lokstra.Logger.Infof("🔄 [%s] Starting request: %s %s", requestID, ctx.Request.Method, ctx.Request.URL.Path)

		// Process request
		err := next(ctx)

		// Calculate processing time
		duration := time.Since(start)
		ctx.Context = context.WithValue(ctx.Context, processingTimeKey, duration)

		lokstra.Logger.Infof("✅ [%s] Completed in %v", requestID, duration)
		return err
	})

	// ===== Path Parameters =====
	// Test: curl http://localhost:8080/users/123/posts/456
	app.GET("/users/:userId/posts/:postId", func(ctx *lokstra.Context) error {
		// Extract path parameters using GetPathParam
		userID := ctx.GetPathParam("userId")
		postID := ctx.GetPathParam("postId")

		// Alternative: Use struct binding for type safety
		type PathParams struct {
			UserID int `path:"userId"`
			PostID int `path:"postId"`
		}
		var params PathParams
		// In Lokstra, this would be done automatically via handler parameter
		// For demonstration, we'll show manual values
		params.UserID = 123
		params.PostID = 456

		return ctx.Ok(map[string]any{
			"message":      "Retrieved user post",
			"user_id":      userID,
			"post_id":      postID,
			"typed_params": params,
			"request_id":   ctx.Value(requestIDKey),
		})
	})

	// ===== Query Parameters =====
	// Test: curl "http://localhost:8080/search?q=golang&page=2&limit=10&sort=date"
	app.GET("/search", func(ctx *lokstra.Context) error {
		// Extract individual query parameters
		query := ctx.GetQueryParam("q")
		page := ctx.GetQueryParam("page")
		limit := ctx.GetQueryParam("limit")
		sort := ctx.GetQueryParam("sort")

		// Show all query parameters
		allParams := ctx.Request.URL.Query()

		return ctx.Ok(map[string]any{
			"message":    "Search completed",
			"query":      query,
			"page":       page,
			"limit":      limit,
			"sort":       sort,
			"all_params": allParams,
			"request_id": ctx.Value(requestIDKey),
		})
	})

	// ===== Headers =====
	// Test: curl -H "Authorization: Bearer token123" -H "User-Agent: MyApp/1.0" http://localhost:8080/headers
	app.GET("/headers", func(ctx *lokstra.Context) error {
		// Extract specific headers
		auth := ctx.GetHeader("Authorization")
		userAgent := ctx.GetHeader("User-Agent")
		contentType := ctx.GetHeader("Content-Type")

		// Check if header contains specific value
		hasJsonContent := ctx.IsHeaderContainValue("Accept", "application/json")

		// Get all headers
		allHeaders := make(map[string][]string)
		for name, values := range ctx.Request.Header {
			allHeaders[name] = values
		}

		return ctx.Ok(map[string]any{
			"message":       "Headers extracted",
			"authorization": auth,
			"user_agent":    userAgent,
			"content_type":  contentType,
			"accepts_json":  hasJsonContent,
			"all_headers":   allHeaders,
			"request_id":    ctx.Value(requestIDKey),
		})
	})

	// ===== Request Body =====
	// Test: curl -X POST -H "Content-Type: application/json" -d '{"name":"John","age":30}' http://localhost:8080/body
	app.POST("/body", func(ctx *lokstra.Context) error {
		// Get raw request body
		rawBody, err := ctx.GetRawRequestBody()
		if err != nil {
			return ctx.ErrorBadRequest("Failed to read request body")
		}

		// Smart binding would normally handle this automatically
		// For demonstration, we'll show raw body and simulate parsed data
		type UserData struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		simulatedUser := UserData{
			Name: "John",
			Age:  30,
		}

		return ctx.Ok(map[string]any{
			"message":        "Body processed",
			"raw_body":       string(rawBody),
			"parsed_data":    simulatedUser,
			"content_length": len(rawBody),
			"request_id":     ctx.Value(requestIDKey),
		})
	})

	// ===== Context Values and Metadata =====
	// Test: curl http://localhost:8080/context-info
	app.GET("/context-info", func(ctx *lokstra.Context) error {
		// Get values stored in context by middleware
		requestID := ctx.Value(requestIDKey)
		startTime := ctx.Value(requestStartKey)

		// Request metadata
		method := ctx.Request.Method
		url := ctx.Request.URL.String()
		proto := ctx.Request.Proto
		remoteAddr := ctx.Request.RemoteAddr

		// Calculate current processing time
		var processingTime any = "in-progress"
		if start, ok := startTime.(time.Time); ok {
			processingTime = time.Since(start).String()
		}

		return ctx.Ok(map[string]any{
			"message":         "Context information",
			"request_id":      requestID,
			"start_time":      startTime,
			"processing_time": processingTime,
			"method":          method,
			"url":             url,
			"protocol":        proto,
			"remote_addr":     remoteAddr,
			"timestamp":       time.Now(),
		})
	})

	// ===== Combined Example =====
	// Test: curl -X PUT -H "Authorization: Bearer abc123" "http://localhost:8080/users/42?include=profile" -d '{"name":"Updated Name"}'
	type UpdateUserParams struct {
		UserID  int    `path:"userId"`
		Include string `query:"include"`
	}

	type UpdateUserRequest struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	app.PUT("/users/:userId", func(ctx *lokstra.Context, params *UpdateUserParams, req *UpdateUserRequest) error {
		// Get authorization header
		auth := ctx.GetHeader("Authorization")

		// Get context values
		requestID := ctx.Value(requestIDKey)

		// Simulate user update
		updatedUser := map[string]any{
			"id":      params.UserID,
			"name":    req.Name,
			"email":   req.Email,
			"include": params.Include,
		}

		return ctx.OkUpdated(map[string]any{
			"user":          updatedUser,
			"authorization": auth,
			"request_id":    requestID,
		})
	})

	// ===== Response Context =====
	// Test: curl http://localhost:8080/response-headers
	app.GET("/response-headers", func(ctx *lokstra.Context) error {
		// Set custom response headers
		ctx.Response.WithHeader("X-Custom-Header", "CustomValue")
		ctx.Response.WithHeader("X-Processing-Time", "123ms")

		if requestID := ctx.Value(requestIDKey); requestID != nil {
			ctx.Response.WithHeader("X-Request-ID", requestID.(string))
		}

		return ctx.Ok(map[string]any{
			"message":    "Custom headers set",
			"headers":    "Check response headers",
			"request_id": ctx.Value(requestIDKey),
		})
	})

	lokstra.Logger.Infof("🚀 Request Context Example started on :8080")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Try these endpoints:")
	lokstra.Logger.Infof("  Path params:")
	lokstra.Logger.Infof("    GET  /users/123/posts/456")
	lokstra.Logger.Infof("  Query params:")
	lokstra.Logger.Infof("    GET  /search?q=golang&page=2&limit=10")
	lokstra.Logger.Infof("  Headers:")
	lokstra.Logger.Infof("    GET  /headers (send Authorization header)")
	lokstra.Logger.Infof("  Request body:")
	lokstra.Logger.Infof("    POST /body (send JSON data)")
	lokstra.Logger.Infof("  Context info:")
	lokstra.Logger.Infof("    GET  /context-info")
	lokstra.Logger.Infof("  Combined:")
	lokstra.Logger.Infof("    PUT  /users/42?include=profile (send JSON + headers)")
	lokstra.Logger.Infof("  Response:")
	lokstra.Logger.Infof("    GET  /response-headers")

	app.Start(true)
}

// Request Context Key Features:
//
// 1. Parameter Extraction:
//    - Path parameters: ctx.GetPathParam("name")
//    - Query parameters: ctx.GetQueryParam("name")
//    - Smart binding: automatic struct binding via handler parameters
//
// 2. Header Handling:
//    - Single header: ctx.GetHeader("Authorization")
//    - Header checking: ctx.IsHeaderContainValue("Accept", "json")
//    - All headers: ctx.Request.Header
//
// 3. Request Body:
//    - Raw body: ctx.GetRawRequestBody()
//    - Smart binding: automatic JSON/form parsing
//    - Validation: built-in validation with struct tags
//
// 4. Context Values:
//    - Store values: ctx.SetValue(key, value)
//    - Retrieve values: ctx.Value(key)
//    - Middleware communication: share data between middleware and handlers
//
// 5. Request Metadata:
//    - HTTP method: ctx.Request.Method
//    - URL: ctx.Request.URL
//    - Protocol: ctx.Request.Proto
//    - Remote address: ctx.Request.RemoteAddr
//
// 6. Response Control:
//    - Custom headers: ctx.Response.WithHeader(name, value)
//    - Status codes: handled by response methods (Ok, ErrorBadRequest, etc.)
//    - Raw responses: ctx.WriteRaw() for custom content

// Test Commands:
//
// # Path parameters
// curl http://localhost:8080/users/123/posts/456
//
// # Query parameters
// curl "http://localhost:8080/search?q=golang&page=2&limit=10&sort=date"
//
// # Headers
// curl -H "Authorization: Bearer token123" -H "User-Agent: MyApp/1.0" http://localhost:8080/headers
//
// # Request body
// curl -X POST -H "Content-Type: application/json" -d '{"name":"John","age":30}' http://localhost:8080/body
//
// # Context information
// curl http://localhost:8080/context-info
//
// # Combined example
// curl -X PUT -H "Authorization: Bearer abc123" "http://localhost:8080/users/42?include=profile" -H "Content-Type: application/json" -d '{"name":"Updated Name","email":"updated@example.com"}'
//
// # Response headers
// curl -v http://localhost:8080/response-headers
