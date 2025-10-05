package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/request"
)

// This example demonstrates three ways to handle request parameters in Lokstra:
// 1. Manual Parameter Access - c.Req.PathParam(), c.Req.QueryParam()
// 2. Manual Binding - c.Req.BindBody(), c.Req.BindQuery(), c.Req.BindPath()
// 3. Smart Binding - Additional parameters in handler signature (recommended!)
//
// Run: go run main.go
// Test: See test.http for all test cases

func main() {
	router := lokstra.NewRouter("handlers-demo")

	// ========================================
	// 1. MANUAL PARAMETER ACCESS
	// ========================================
	// Basic approach - manually extract each parameter
	// Good for: Simple cases, optional parameters, default values

	router.GET("/manual/users/:id", func(c *lokstra.RequestContext) error {
		// Extract path parameter
		id := c.Req.PathParam("id", "0")

		// Extract query parameters
		format := c.Req.QueryParam("format", "json")
		includeDetails := c.Req.QueryParam("details", "false")

		return c.Api.Ok(map[string]any{
			"method":          "manual-params",
			"user_id":         id,
			"format":          format,
			"include_details": includeDetails,
			"note":            "Parameters extracted manually",
		})
	})

	router.POST("/manual/users", func(c *lokstra.RequestContext) error {
		// Manual JSON decoding - old school
		rawBody, err := c.Req.RawRequestBody()
		if err != nil {
			return c.Api.BadRequest("INVALID_BODY", "Failed to read request body")
		}

		var body map[string]any
		if err := json.Unmarshal(rawBody, &body); err != nil {
			return c.Api.BadRequest("INVALID_JSON", "Invalid JSON format")
		}

		// Manually validate required fields
		name, ok := body["name"].(string)
		if !ok || name == "" {
			return c.Api.BadRequest("VALIDATION_ERROR", "Name is required")
		}

		email, ok := body["email"].(string)
		if !ok || email == "" {
			return c.Api.BadRequest("VALIDATION_ERROR", "Email is required")
		}

		return c.Api.Created(map[string]any{
			"method": "manual-decode",
			"name":   name,
			"email":  email,
			"note":   "Validated manually",
		}, "User created successfully")
	})

	// ========================================
	// 2. MANUAL BINDING
	// ========================================
	// Better approach - bind to struct, automatic validation
	// Good for: Structured data, validation tags

	type CreateUserRequest struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
		Age   int    `json:"age" validate:"min=1,max=120"`
		Role  string `json:"role" validate:"oneof=admin user guest"`
	}

	router.POST("/bind/users", func(c *lokstra.RequestContext) error {
		var req CreateUserRequest

		// BindBody automatically:
		// - Decodes JSON
		// - Validates using struct tags (NEW! automatic validation)
		// - Returns ValidationError if validation fails
		if err := c.Req.BindBody(&req); err != nil {
			// Check if it's a validation error with field details
			if valErr, ok := err.(*request.ValidationError); ok {
				// Return structured validation error with field-level details
				return c.Api.ValidationError("Validation failed", valErr.FieldErrors)
			}
			// Other errors (e.g., JSON parsing error)
			return c.Api.BadRequest("INVALID_REQUEST", err.Error())
		}

		return c.Api.Created(map[string]any{
			"method": "manual-bind",
			"user":   req,
			"note":   "Validated using struct tags - automatic validation!",
		}, "User created successfully")
	})

	type UserQueryParams struct {
		Page     int    `query:"page" validate:"min=1"`
		PageSize int    `query:"page_size" validate:"min=1,max=100"`
		Sort     string `query:"sort" validate:"oneof=name email created_at"`
		Order    string `query:"order" validate:"oneof=asc desc"`
	}

	router.GET("/bind/users", func(c *lokstra.RequestContext) error {
		var params UserQueryParams

		// BindQuery extracts and validates query parameters
		if err := c.Req.BindQuery(&params); err != nil {
			// Check if it's a validation error
			if valErr, ok := err.(*request.ValidationError); ok {
				return c.Api.ValidationError("Invalid query parameters", valErr.FieldErrors)
			}
			return c.Api.BadRequest("INVALID_REQUEST", err.Error())
		}

		// Simulate database query with pagination
		users := []map[string]any{
			{"id": 1, "name": "Alice", "email": "alice@example.com"},
			{"id": 2, "name": "Bob", "email": "bob@example.com"},
			{"id": 3, "name": "Charlie", "email": "charlie@example.com"},
		}

		return c.Api.Ok(map[string]any{
			"method":     "bind-query",
			"pagination": params,
			"users":      users,
			"note":       "Query parameters validated with struct tags",
		})
	})

	type UserPathParams struct {
		UserID string `path:"user_id" validate:"required"`
		PostID string `path:"post_id" validate:"required"`
	}

	router.GET("/bind/users/:user_id/posts/:post_id", func(c *lokstra.RequestContext) error {
		var params UserPathParams

		// BindPath extracts and validates path parameters
		if err := c.Req.BindPath(&params); err != nil {
			if valErr, ok := err.(*request.ValidationError); ok {
				return c.Api.ValidationError("Invalid path parameters", valErr.FieldErrors)
			}
			return c.Api.BadRequest("INVALID_REQUEST", err.Error())
		}

		return c.Api.Ok(map[string]any{
			"method":  "bind-path",
			"user_id": params.UserID,
			"post_id": params.PostID,
			"post": map[string]any{
				"title":   "Sample Post",
				"content": "This is a sample post",
			},
			"note": "Path parameters validated",
		})
	})

	// ========================================
	// 3. SMART BINDING (RECOMMENDED!)
	// ========================================
	// Best approach - declare parameters in handler signature
	// Lokstra automatically binds and validates before calling handler
	// Good for: Clean code, automatic validation, type safety
	//
	// RULE: Only ONE struct parameter allowed (besides context)!
	// But that struct can combine: path, query, header, json/body tags

	type CreateProductRequest struct {
		Name        string  `json:"name" validate:"required"`
		Description string  `json:"description"`
		Price       float64 `json:"price" validate:"required,gt=0"`
		Stock       int     `json:"stock" validate:"min=0"`
	}

	// Smart Bind: Just add ONE struct as a parameter!
	// Lokstra will automatically bind and validate before calling the handler
	// IMPORTANT: Only ONE struct parameter allowed (besides context)
	router.POST("/smart/products", func(c *lokstra.RequestContext, req *CreateProductRequest) error {
		// If we reach here, validation already passed!
		// No need for manual binding or validation

		return c.Api.Created(map[string]any{
			"method":  "smart-bind",
			"product": req,
			"note":    "Automatically validated before handler execution",
		}, "Product created successfully")
	})

	// Smart Bind: Combining multiple sources in ONE struct
	// You can mix path, query, header, and body tags in a single struct!
	type UpdateProductRequest struct {
		// Path parameter
		ID string `path:"id" validate:"required"`

		// Body parameters (JSON) - Optional fields use pointers
		Name        *string  `json:"name"`
		Description *string  `json:"description"`
		Price       *float64 `json:"price" validate:"omitempty,gt=0"`
		Stock       *int     `json:"stock" validate:"omitempty,min=0"`
	}

	// Smart Bind: Only ONE struct parameter allowed!
	// But that struct can contain fields from path, query, header, and body
	router.PUT("/smart/products/:id", func(c *lokstra.RequestContext, req *UpdateProductRequest) error {
		return c.Api.Ok(map[string]any{
			"method":     "smart-bind-combined",
			"product_id": req.ID, // From path
			"updates": map[string]any{
				"name":        req.Name,
				"description": req.Description,
				"price":       req.Price,
				"stock":       req.Stock,
			},
			"note": "Path + Body combined in ONE struct",
		})
	})

	// Smart Bind: All query parameters in one struct
	type SearchParams struct {
		Query    string `query:"q" validate:"required"`
		Category string `query:"category"`
		MinPrice int    `query:"min_price" validate:"omitempty,min=0"`
		MaxPrice int    `query:"max_price" validate:"omitempty,min=0"`
		Page     int    `query:"page" validate:"min=1"`
	}

	// Smart Bind: One struct parameter with query tags
	router.GET("/smart/products/search", func(c *lokstra.RequestContext, params *SearchParams) error {
		// Simulate search results
		products := []map[string]any{
			{
				"id":          1,
				"name":        "Laptop",
				"category":    params.Category,
				"price":       1000,
				"match_score": 0.95,
			},
			{
				"id":          2,
				"name":        "Laptop Stand",
				"category":    params.Category,
				"price":       50,
				"match_score": 0.75,
			},
		}

		return c.Api.Ok(map[string]any{
			"method":  "smart-bind-query",
			"query":   params.Query,
			"filters": params,
			"results": products,
			"note":    "Query parameters automatically validated",
		})
	})

	// Advanced Smart Bind: Combining ALL sources (path + query + header + body)
	type AdvancedRequest struct {
		// Path parameter
		ProductID string `path:"id" validate:"required"`

		// Query parameters
		Action string `query:"action" validate:"required,oneof=activate deactivate"`
		Notify bool   `query:"notify"`

		// Header
		APIKey string `header:"X-API-Key" validate:"required,min=10"`

		// Body (JSON)
		Reason  string `json:"reason" validate:"required"`
		Comment string `json:"comment"`
	}

	router.POST("/smart/products/:id/actions", func(c *lokstra.RequestContext, req *AdvancedRequest) error {
		return c.Api.Ok(map[string]any{
			"method":     "smart-bind-combined-all",
			"product_id": req.ProductID, // from path
			"action":     req.Action,    // from query
			"notify":     req.Notify,    // from query
			"api_key":    req.APIKey,    // from header
			"reason":     req.Reason,    // from body
			"comment":    req.Comment,   // from body
			"note":       "ONE struct combining path + query + header + body!",
		})
	})

	// ========================================
	// 4. CUSTOM VALIDATORS
	// ========================================
	// Demonstrates using custom validators registered in custom_validator.go
	// Custom validators: uuid, startswith, alphanum, url

	type CreateProductWithCustomValidatorsRequest struct {
		ID       string  `json:"id" validate:"required,uuid"`              // Custom UUID validator
		Code     string  `json:"code" validate:"required,startswith=PRD-"` // Custom startswith validator
		Name     string  `json:"name" validate:"required,min=3"`           // Built-in validators
		SKU      string  `json:"sku" validate:"required,alphanum"`         // Custom alphanum validator
		Price    float64 `json:"price" validate:"required,gt=0"`           // Built-in validators
		Quantity int     `json:"quantity" validate:"required,min=1"`       // Built-in validators
	}

	router.POST("/custom/products", func(c *lokstra.RequestContext, req *CreateProductWithCustomValidatorsRequest) error {
		return c.Api.Created(map[string]any{
			"method":  "custom-validators",
			"product": req,
			"note":    "Uses custom validators: uuid, startswith, alphanum",
		}, "Product created with custom validation")
	})

	type WebsiteRequest struct {
		Name    string `json:"name" validate:"required,min=3"`
		URL     string `json:"url" validate:"required,url"`        // Custom URL validator
		Owner   string `json:"owner" validate:"required,alphanum"` // Custom alphanum validator
		AdminID string `json:"admin_id" validate:"required,uuid"`  // Custom UUID validator
	}

	router.POST("/custom/websites", func(c *lokstra.RequestContext, req *WebsiteRequest) error {
		return c.Api.Created(map[string]any{
			"method":  "custom-validators-url",
			"website": req,
			"note":    "Uses custom validators: url, alphanum, uuid",
		}, "Website registered with custom validation")
	})

	// ========================================
	// COMPARISON ROUTE
	// ========================================
	router.GET("/", func(c *lokstra.RequestContext) error {
		return c.Api.Ok(map[string]any{
			"title": "Lokstra Handler Patterns Demo",
			"approaches": map[string]any{
				"manual-params": map[string]any{
					"description": "c.Req.PathParam(), c.Req.QueryParam()",
					"pros":        []string{"Simple", "Good for optional params", "Default values"},
					"cons":        []string{"Manual validation", "Verbose", "No type safety"},
					"example":     "GET /manual/users/:id?format=json&details=true",
				},
				"manual-bind": map[string]any{
					"description": "c.Req.BindBody(), c.Req.BindQuery(), c.Req.BindPath()",
					"pros":        []string{"Struct validation", "Type safety", "Reusable structs"},
					"cons":        []string{"Manual binding call", "Extra error handling"},
					"example":     "POST /bind/users with JSON body",
				},
				"smart-bind": map[string]any{
					"description": "One struct parameter in handler signature (can combine path, query, header, body tags)",
					"pros":        []string{"Cleanest code", "Automatic validation", "Type safe", "Can mix different sources"},
					"cons":        []string{"Only ONE struct parameter allowed"},
					"example":     "POST /smart/products with auto-binding",
				},
			},
			"recommendation": "Use Smart Binding (handler signature parameters) for best developer experience!",
		})
	})

	// Create and run the app
	app := lokstra.NewApp("handlers-demo", ":8080", router)

	fmt.Println("🎯 Handler Patterns Demo Server")
	fmt.Println("================================")
	fmt.Println("\n📋 Available Endpoints:")
	fmt.Println("\n1️⃣  Manual Parameter Access:")
	fmt.Println("  GET  http://localhost:8080/manual/users/:id?format=json&details=true")
	fmt.Println("  POST http://localhost:8080/manual/users")
	fmt.Println("\n2️⃣  Manual Binding:")
	fmt.Println("  POST http://localhost:8080/bind/users")
	fmt.Println("  GET  http://localhost:8080/bind/users?page=1&page_size=10&sort=name&order=asc")
	fmt.Println("  GET  http://localhost:8080/bind/users/:user_id/posts/:post_id")
	fmt.Println("\n3️⃣  Smart Binding (Recommended!):")
	fmt.Println("  POST http://localhost:8080/smart/products")
	fmt.Println("  PUT  http://localhost:8080/smart/products/:id")
	fmt.Println("  GET  http://localhost:8080/smart/products/search?q=laptop&category=electronics&page=1")
	fmt.Println("  POST http://localhost:8080/smart/products/:id/actions?action=activate&notify=true")
	fmt.Println("       (Advanced: Combines path + query + header + body in ONE struct!)")
	fmt.Println("\n4️⃣  Custom Validators:")
	fmt.Println("  POST http://localhost:8080/custom/products")
	fmt.Println("  POST http://localhost:8080/custom/websites")
	fmt.Println("       (Uses custom validators: uuid, startswith, alphanum, url)")
	fmt.Println("\n📖 Comparison:")
	fmt.Println("  GET  http://localhost:8080/")
	fmt.Println("\n🧪 See test.http for complete test suite!")
	fmt.Println("\n🚀 Server starting on http://localhost:8080")
	fmt.Println("Press Ctrl+C to stop")

	router.PrintRoutes()

	app.Run(30 * time.Second)
}
