package main

import (
	"github.com/primadi/lokstra"
)

// This example demonstrates Lokstra's smart request binding feature.
// It shows how to automatically bind request data to struct parameters.
//
// Learning Objectives:
// - Understand smart binding with struct tags
// - Learn query parameter binding
// - Explore path parameter binding
// - See JSON body binding in action
//
// Documentation: https://github.com/primadi/lokstra/blob/main/docs/core-concepts.md#smart-request-binding

// User represents user data with smart binding tags
type User struct {
	ID     int    `path:"id"`      // Bind from URL path
	Name   string `query:"name"`   // Bind from query parameter
	Email  string `json:"email"`   // Bind from JSON body
	Active bool   `query:"active"` // Bind boolean from query
}

// CreateUserRequest represents request body for creating a user
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"min=18"`
}

// SearchParams represents search query parameters
type SearchParams struct {
	Query  string `query:"q"`
	Limit  int    `query:"limit" default:"10"`
	Offset int    `query:"offset" default:"0"`
	Sort   string `query:"sort" default:"name"`
}

func main() {
	regCtx := lokstra.NewGlobalRegistrationContext()
	app := lokstra.NewApp(regCtx, "smart-binding-app", ":8080")

	// Example 1: Query parameter binding
	// Test: curl "http://localhost:8080/search?q=john&limit=5&sort=email"
	app.GET("/search", func(ctx *lokstra.Context, params *SearchParams) error {
		return ctx.Ok(map[string]interface{}{
			"message": "Search completed",
			"params":  params,
		})
	})

	// Example 2: Path parameter binding
	// Test: curl "http://localhost:8080/users/123?name=John&active=true"
	app.GET("/users/:id", func(ctx *lokstra.Context, user *User) error {
		return ctx.Ok(map[string]interface{}{
			"message": "User retrieved",
			"user":    user,
		})
	})

	// Example 3: JSON body binding with validation
	// Test: curl -X POST "http://localhost:8080/users" \
	//       -H "Content-Type: application/json" \
	//       -d '{"name":"John Doe","email":"john@example.com","age":25}'
	app.POST("/users", func(ctx *lokstra.Context, req *CreateUserRequest) error {
		// In a real app, you would save the user to database here
		return ctx.Ok(map[string]interface{}{
			"message": "User created successfully",
			"user":    req,
		})
	})

	// Example 4: Combined binding (path + query + body)
	// Test: curl -X PUT "http://localhost:8080/users/123?active=false" \
	//       -H "Content-Type: application/json" \
	//       -d '{"name":"John Smith","email":"johnsmith@example.com","age":30}'
	type UpdateUserParams struct {
		ID     int                `path:"id"`
		Active *bool              `query:"active"`
		Data   *CreateUserRequest `json:",inline"`
	}

	app.PUT("/users/:id", func(ctx *lokstra.Context, params *UpdateUserParams) error {
		return ctx.Ok(map[string]interface{}{
			"message": "User updated successfully",
			"params":  params,
		})
	})

	// Example 5: Error handling with validation
	// Test invalid data: curl -X POST "http://localhost:8080/users" \
	//                    -H "Content-Type: application/json" \
	//                    -d '{"name":"","email":"invalid-email","age":15}'
	app.POST("/users/validated", func(ctx *lokstra.Context, req *CreateUserRequest) error {
		// Lokstra automatically validates based on struct tags
		// If validation fails, it returns 400 Bad Request automatically
		return ctx.Ok(map[string]interface{}{
			"message": "User validation passed",
			"user":    req,
		})
	})

	lokstra.Logger.Infof("Smart Binding Example started on :8080")
	lokstra.Logger.Infof("Try these endpoints:")
	lokstra.Logger.Infof("  GET  /search?q=test&limit=5")
	lokstra.Logger.Infof("  GET  /users/123?name=John&active=true")
	lokstra.Logger.Infof("  POST /users (with JSON body)")
	lokstra.Logger.Infof("  PUT  /users/123?active=false (with JSON body)")

	app.Start()
}

// Smart Binding Features Demonstrated:
//
// 1. Query Parameters: `query:"name"`
//    - Automatically binds URL query parameters to struct fields
//    - Supports type conversion (string, int, bool, etc.)
//    - Supports default values
//
// 2. Path Parameters: `path:"id"`
//    - Binds URL path segments to struct fields
//    - Supports type conversion
//
// 3. JSON Body: `json:"email"`
//    - Automatically parses JSON request body
//    - Binds to struct fields
//
// 4. Validation: `validate:"required,email,min=18"`
//    - Automatic validation based on struct tags
//    - Returns 400 Bad Request for validation errors
//
// 5. Default Values: `default:"10"`
//    - Sets default values for missing parameters
//
// 6. Inline Binding: `json:",inline"`
//    - Flattens nested structs for complex binding scenarios
