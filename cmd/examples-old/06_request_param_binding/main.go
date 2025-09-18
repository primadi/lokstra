// This example demonstrates various request parameter binding approaches in Lokstra.
// It shows manual binding, smart binding, and binding to map[string]any.
//
// Framework: Lokstra (https://github.com/primadi/lokstra)
// Documentation: /docs/core-concepts.md for binding details
package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/primadi/lokstra"
)

// UserRequest demonstrates struct with all binding tags
type UserRequest struct {
	// Path parameters
	ID string `path:"id"`

	// Query parameters
	Page   int      `query:"page"`
	Limit  int      `query:"limit"`
	Tags   []string `query:"tags"`
	Active bool     `query:"active"`

	// Headers
	Authorization string `header:"Authorization"`
	UserAgent     string `header:"User-Agent"`

	// Body (JSON)
	Name        string         `body:"name"`
	Email       string         `body:"email"`
	Age         int            `body:"age"`
	Preferences map[string]any `body:"preferences"`
}

// CreateUserRequest for body-only operations
type CreateUserRequest struct {
	Name  string `body:"name"`
	Email string `body:"email"`
	Age   int    `body:"age"`
}

// PathOnlyRequest for path-only operations
type PathOnlyRequest struct {
	ID string `path:"id"`
}

func main() {
	regCtx := lokstra.NewGlobalRegistrationContext()
	// Use default App factory (NewApp). The default router engine has been
	// configured to `servemux` in defaults so NewApp will pick it.
	app := lokstra.NewApp(regCtx, "binding-example", ":8080")

	// Register all binding example handlers
	setupRoutes(app)

	fmt.Println("🚀 Lokstra Request Binding Examples")
	fmt.Println("📖 See README.md for usage examples")
	fmt.Println("🌐 Server running on http://localhost:8080")
	fmt.Println()
	fmt.Println("Available endpoints:")
	fmt.Println("  GET /users/:id - Manual binding example")
	fmt.Println("  POST /users/:id/smart - Smart binding example")
	fmt.Println("  POST /users/create-map - BindBodySmart to map")
	fmt.Println("  POST /users/:id/all-map - BindAllSmart to map")
	fmt.Println("  GET /health - Health check")

	if err := app.StartAndWaitForShutdown(30 * time.Second); err != nil {
		log.Fatal(err)
	}
}

func setupRoutes(app *lokstra.App) {
	// Health check endpoint
	app.GET("/health", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"status":    "healthy",
			"timestamp": time.Now(),
			"service":   "request-binding-examples",
		})
	})

	// Register concrete/static paths and specific child routes before wildcard
	// wildcard routes to avoid httprouter conflicts.
	app.POST("/users/create-map", bindBodySmartToMapHandler)

	// Specific child routes under /users/:id (register before the wildcard)
	app.POST("/users/:id/smart", smartBindingHandler)
	app.POST("/users/:id/all-map", bindAllSmartToMapHandler)
	app.POST("/users/:id/hybrid", hybridBindingHandler)

	// 1. Manual binding - wildcard route for user id (register last)
	app.GET("/users/:id", manualBindingHandler)

	// 6. Complex query parameters
	app.GET("/search", complexQueryBindingHandler)
}

// manualBindingHandler demonstrates manual step-by-step binding
func manualBindingHandler(ctx *lokstra.Context) error {
	var req UserRequest

	// Manual binding step by step
	if err := ctx.BindPath(&req); err != nil {
		return ctx.ErrorBadRequest("Path binding failed: " + err.Error())
	}

	if err := ctx.BindQuery(&req); err != nil {
		return ctx.ErrorBadRequest("Query binding failed: " + err.Error())
	}

	if err := ctx.BindHeader(&req); err != nil {
		return ctx.ErrorBadRequest("Header binding failed: " + err.Error())
	}

	// Note: No body for GET request, but could bind if it was POST/PUT
	// if err := ctx.BindBody(&req); err != nil {
	//     return ctx.ErrorBadRequest("Body binding failed: " + err.Error())
	// }

	return ctx.Ok(map[string]any{
		"method": "manual_binding",
		"data": map[string]any{
			"id":            req.ID,
			"page":          req.Page,
			"limit":         req.Limit,
			"tags":          req.Tags,
			"active":        req.Active,
			"authorization": req.Authorization,
			"user_agent":    req.UserAgent,
		},
		"message": "Successfully bound using manual step-by-step approach",
	})
}

// smartBindingHandler demonstrates automatic smart binding
func smartBindingHandler(ctx *lokstra.Context, req *UserRequest) error {
	// Request automatically bound by Lokstra - no manual binding needed!

	return ctx.Ok(map[string]any{
		"method": "smart_binding",
		"data": map[string]any{
			"id":            req.ID,
			"page":          req.Page,
			"limit":         req.Limit,
			"tags":          req.Tags,
			"active":        req.Active,
			"authorization": req.Authorization,
			"user_agent":    req.UserAgent,
			"name":          req.Name,
			"email":         req.Email,
			"age":           req.Age,
			"preferences":   req.Preferences,
		},
		"message": "Successfully bound using smart binding (automatic)",
	})
}

// bindBodySmartToMapHandler demonstrates BindBodySmart to map[string]any
func bindBodySmartToMapHandler(ctx *lokstra.Context) error {
	var bodyData map[string]any

	// If content-type is form-urlencoded, parse form into map manually to avoid
	// reflection panics when binding into a map via BindBodySmart.
	contentType := ctx.Request.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		if err := ctx.Request.ParseForm(); err != nil {
			return ctx.ErrorBadRequest("Body binding failed: " + err.Error())
		}
		bodyData = make(map[string]any)
		for k, vs := range ctx.Request.PostForm {
			if len(vs) > 1 {
				arr := make([]any, len(vs))
				for i, v := range vs {
					arr[i] = v
				}
				bodyData[k] = arr
			} else {
				bodyData[k] = vs[0]
			}
		}
	} else {
		// BindBodySmart can handle JSON and other types into map
		if err := ctx.BindBodySmart(&bodyData); err != nil {
			return ctx.ErrorBadRequest("Body binding failed: " + err.Error())
		}
	}

	// Process the dynamic data
	result := map[string]any{
		"method":        "bind_body_smart_to_map",
		"received_data": bodyData,
		"data_type":     fmt.Sprintf("%T", bodyData),
		"message":       "Successfully bound body to map[string]any using BindBodySmart",
	}

	// Add some analysis of the received data
	if bodyData != nil {
		result["field_count"] = len(bodyData)

		fields := make([]string, 0, len(bodyData))
		for key := range bodyData {
			fields = append(fields, key)
		}
		result["fields"] = fields
	}

	return ctx.Ok(result)
}

// bindAllSmartToMapHandler demonstrates BindAllSmart limitations with maps
func bindAllSmartToMapHandler(ctx *lokstra.Context) error {
	// Note: This will demonstrate the limitation - BindAllSmart expects struct
	// We'll show the error and explain why hybrid approach is better

	var allData map[string]any

	// BindAllSmart expects a pointer to a struct with binding tags. It will
	// return an error when used with non-struct targets (we validate earlier
	// in the library). Use that error to demonstrate the limitation and
	// recommend the hybrid approach instead of recovering from a panic.
	if err := ctx.BindAllSmart(&allData); err != nil {
		return ctx.Ok(map[string]any{
			"method":         "bind_all_smart_to_map",
			"error":          err.Error(),
			"message":        "BindAllSmart to map[string]any failed as expected. Use hybrid approach instead.",
			"recommendation": "Use struct for path/query/header + BindBodySmart for dynamic body",
			"see_endpoint":   "/users/:id/hybrid",
		})
	}

	// If it somehow worked, show the result
	return ctx.Ok(map[string]any{
		"method":    "bind_all_smart_to_map",
		"data":      allData,
		"message":   "Unexpectedly succeeded with BindAllSmart to map",
		"surprised": true,
	})
}

// hybridBindingHandler demonstrates the recommended approach:
// struct for path/query/header + map for dynamic body
func hybridBindingHandler(ctx *lokstra.Context) error {
	// Struct for structured parameters
	var pathQuery struct {
		ID    string   `path:"id"`
		Page  int      `query:"page"`
		Limit int      `query:"limit"`
		Tags  []string `query:"tags"`
		Auth  string   `header:"Authorization"`
	}

	// Map for dynamic body content
	var bodyData map[string]any

	// Bind structured data
	if err := ctx.BindPath(&pathQuery); err != nil {
		return ctx.ErrorBadRequest("Path binding failed: " + err.Error())
	}
	if err := ctx.BindQuery(&pathQuery); err != nil {
		return ctx.ErrorBadRequest("Query binding failed: " + err.Error())
	}
	if err := ctx.BindHeader(&pathQuery); err != nil {
		return ctx.ErrorBadRequest("Header binding failed: " + err.Error())
	}

	// Bind dynamic body
	if err := ctx.BindBodySmart(&bodyData); err != nil {
		return ctx.ErrorBadRequest("Body binding failed: " + err.Error())
	}

	return ctx.Ok(map[string]any{
		"method": "hybrid_binding",
		"structured_data": map[string]any{
			"id":            pathQuery.ID,
			"page":          pathQuery.Page,
			"limit":         pathQuery.Limit,
			"tags":          pathQuery.Tags,
			"authorization": pathQuery.Auth,
		},
		"dynamic_body":   bodyData,
		"message":        "Successfully used hybrid approach: struct for params + map for body",
		"recommendation": "This is the recommended pattern for flexible APIs",
	})
}

// complexQueryBindingHandler demonstrates complex query parameter binding
func complexQueryBindingHandler(ctx *lokstra.Context) error {
	var searchReq struct {
		Query     string            `query:"q"`
		Filters   []string          `query:"filter"`
		Sort      string            `query:"sort"`
		Page      int               `query:"page"`
		Limit     int               `query:"limit"`
		Options   map[string]string `query:"opt"`
		DateRange []string          `query:"date"`
	}

	if err := ctx.BindQuery(&searchReq); err != nil {
		return ctx.ErrorBadRequest("Query binding failed: " + err.Error())
	}

	return ctx.Ok(map[string]any{
		"method": "complex_query_binding",
		"search_parameters": map[string]any{
			"query":      searchReq.Query,
			"filters":    searchReq.Filters,
			"sort":       searchReq.Sort,
			"page":       searchReq.Page,
			"limit":      searchReq.Limit,
			"options":    searchReq.Options,
			"date_range": searchReq.DateRange,
		},
		"message":     "Successfully bound complex query parameters",
		"example_url": "/search?q=lokstra&filter=type:web&filter=lang:go&sort=name&page=1&limit=10&opt[format]=json&opt[include]=docs&date=2023-01-01&date=2023-12-31",
	})
}
