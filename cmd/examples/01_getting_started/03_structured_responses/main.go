package main

import (
	"fmt"

	"github.com/primadi/lokstra"
)

// This example demonstrates Lokstra's structured response system.
// It shows different types of responses that provide consistent API structure.
//
// Learning Objectives:
// - Understand Lokstra's structured response format
// - Learn success response types (Ok, OkCreated, OkUpdated, OkList)
// - Explore error response types with proper HTTP status codes
// - See validation and field error responses
//
// Documentation: https://github.com/primadi/lokstra/blob/main/docs/core-concepts.md#structured-responses

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	InStock     bool    `json:"in_stock"`
}

type CreateProductRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,min=0"`
}

func main() {
	regCtx := lokstra.NewGlobalRegistrationContext()
	app := lokstra.NewApp(regCtx, "structured-response-app", ":8080")

	// In-memory storage for demonstration
	products := []Product{
		{ID: 1, Name: "Laptop", Description: "Gaming laptop", Price: 1299.99, InStock: true},
		{ID: 2, Name: "Mouse", Description: "Wireless mouse", Price: 29.99, InStock: true},
	}
	nextID := 3

	// ===== SUCCESS RESPONSES =====

	// Example 1: Ok() - Standard success response
	// Test: curl http://localhost:8080/health
	app.GET("/health", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"status":    "healthy",
			"timestamp": "2024-01-01T10:00:00Z",
			"version":   "1.0.0",
		})
	})

	// Example 2: OkList() - List response with metadata
	// Test: curl http://localhost:8080/products
	app.GET("/products", func(ctx *lokstra.Context) error {
		meta := map[string]any{
			"total":    len(products),
			"page":     1,
			"per_page": 10,
		}
		return ctx.OkList(products, meta)
	})

	// Example 3: OkCreated() - Resource creation response
	// Test: curl -X POST http://localhost:8080/products \
	//       -H "Content-Type: application/json" \
	//       -d '{"name":"Keyboard","description":"Mechanical keyboard","price":89.99}'
	app.POST("/products", func(ctx *lokstra.Context, req *CreateProductRequest) error {
		product := Product{
			ID:          nextID,
			Name:        req.Name,
			Description: req.Description,
			Price:       req.Price,
			InStock:     true,
		}
		products = append(products, product)
		nextID++

		return ctx.OkCreated(product)
	})

	// Example 4: OkUpdated() - Resource update response
	// Test: curl -X PUT http://localhost:8080/products/1 \
	//       -H "Content-Type: application/json" \
	//       -d '{"name":"Gaming Laptop","description":"High-end gaming laptop","price":1499.99}'
	type UpdateProductParams struct {
		ID   int                   `path:"id"`
		Data *CreateProductRequest `json:",inline"`
	}

	app.PUT("/products/:id", func(ctx *lokstra.Context, params *UpdateProductParams) error {
		// Find product
		for i, product := range products {
			if product.ID == params.ID {
				products[i].Name = params.Data.Name
				products[i].Description = params.Data.Description
				products[i].Price = params.Data.Price
				return ctx.OkUpdated(products[i])
			}
		}
		return ctx.ErrorNotFound("Product not found")
	})

	// ===== ERROR RESPONSES =====

	// Example 5: ErrorNotFound() - 404 response
	// Test: curl http://localhost:8080/products/999
	type GetProductParams struct {
		ID int `path:"id"`
	}

	app.GET("/products/:id", func(ctx *lokstra.Context, params *GetProductParams) error {
		for _, product := range products {
			if product.ID == params.ID {
				return ctx.Ok(product)
			}
		}
		return ctx.ErrorNotFound(fmt.Sprintf("Product with ID %d not found", params.ID))
	})

	// Example 6: ErrorBadRequest() - 400 response
	// Test: curl -X DELETE http://localhost:8080/products/abc
	type DeleteProductParams struct {
		ID int `path:"id"`
	}

	app.DELETE("/products/:id", func(ctx *lokstra.Context, params *DeleteProductParams) error {
		// Find and remove product
		for i, product := range products {
			if product.ID == params.ID {
				products = append(products[:i], products[i+1:]...)
				return ctx.Ok(map[string]any{
					"message": "Product deleted successfully",
					"id":      params.ID,
				})
			}
		}
		return ctx.ErrorNotFound("Product not found")
	})

	// Example 7: ErrorValidation() - Validation errors with field details
	// Test: curl -X POST http://localhost:8080/products/validate \
	//       -H "Content-Type: application/json" \
	//       -d '{"name":"","price":-10}'
	app.POST("/products/validate", func(ctx *lokstra.Context, req *CreateProductRequest) error {
		// Manual validation for demonstration
		fieldErrors := make(map[string]string)

		if req.Name == "" {
			fieldErrors["name"] = "Product name is required"
		}
		if req.Price < 0 {
			fieldErrors["price"] = "Price must be greater than or equal to 0"
		}
		if len(req.Name) > 100 {
			fieldErrors["name"] = "Product name must be less than 100 characters"
		}

		if len(fieldErrors) > 0 {
			return ctx.ErrorValidation("Validation failed", fieldErrors)
		}

		// Create product if validation passes
		product := Product{
			ID:          nextID,
			Name:        req.Name,
			Description: req.Description,
			Price:       req.Price,
			InStock:     true,
		}
		products = append(products, product)
		nextID++

		return ctx.OkCreated(product)
	})

	// Example 8: ErrorDuplicate() - 409 conflict response
	// Test: curl -X POST http://localhost:8080/products/unique \
	//       -H "Content-Type: application/json" \
	//       -d '{"name":"Laptop","price":999.99}'
	app.POST("/products/unique", func(ctx *lokstra.Context, req *CreateProductRequest) error {
		// Check for duplicate name
		for _, product := range products {
			if product.Name == req.Name {
				return ctx.ErrorDuplicate("A product with this name already exists")
			}
		}

		product := Product{
			ID:          nextID,
			Name:        req.Name,
			Description: req.Description,
			Price:       req.Price,
			InStock:     true,
		}
		products = append(products, product)
		nextID++

		return ctx.OkCreated(product)
	})

	// Example 9: ErrorInternal() - 500 server error
	// Test: curl http://localhost:8080/error-demo
	app.GET("/error-demo", func(ctx *lokstra.Context) error {
		// Simulate internal server error
		return ctx.ErrorInternal("Something went wrong on our end. Please try again later.")
	})

	lokstra.Logger.Infof("Structured Response Example started on :8080")
	lokstra.Logger.Infof("Try these endpoints:")
	lokstra.Logger.Infof("  Success responses:")
	lokstra.Logger.Infof("    GET  /health              - Ok() response")
	lokstra.Logger.Infof("    GET  /products            - OkList() with metadata")
	lokstra.Logger.Infof("    POST /products            - OkCreated() response")
	lokstra.Logger.Infof("    PUT  /products/1          - OkUpdated() response")
	lokstra.Logger.Infof("  Error responses:")
	lokstra.Logger.Infof("    GET  /products/999        - ErrorNotFound() 404")
	lokstra.Logger.Infof("    POST /products/validate   - ErrorValidation() 400")
	lokstra.Logger.Infof("    POST /products/unique     - ErrorDuplicate() 409")
	lokstra.Logger.Infof("    GET  /error-demo          - ErrorInternal() 500")

	app.Start()
}

// Structured Response Format:
//
// All Lokstra responses follow this consistent structure:
//
// Success Response:
// {
//   "success": true,
//   "message": "OK",
//   "data": <your_data>,
//   "meta": <optional_metadata>
// }
//
// Error Response:
// {
//   "success": false,
//   "message": "Error description",
//   "field_errors": {          // Optional for validation errors
//     "field_name": "error_msg"
//   }
// }
//
// Benefits:
// - Consistent API structure across all endpoints
// - Easy client-side error handling
// - Proper HTTP status codes
// - Structured validation error reporting
// - Metadata support for pagination and additional info
