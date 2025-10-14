package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/primadi/lokstra/core/request"
)

// ============================================================================
// Product Models (for demonstrating various handler patterns)
// ============================================================================

type Product struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// Struct with tags for binding
type GetProductRequest struct {
	ID string `path:"id"` // Bind from path parameter
}

type SearchProductRequest struct {
	Query    string  `query:"q"`         // Bind from query parameter
	MinPrice float64 `query:"min_price"` // Bind from query parameter
	MaxPrice float64 `query:"max_price"` // Bind from query parameter
}

type CreateProductRequest struct {
	Name  string  `json:"name" validate:"required"`       // Bind from JSON body
	Price float64 `json:"price" validate:"required,gt=0"` // Bind from JSON body
}

type UpdateProductRequest struct {
	ID    string  `path:"id"`                             // Bind from path parameter
	Name  string  `json:"name" validate:"required"`       // Bind from JSON body
	Price float64 `json:"price" validate:"required,gt=0"` // Bind from JSON body
}

// ============================================================================
// Product Service - Demonstrating 7 Handler Patterns
// ============================================================================

type ProductService struct {
	products map[string]*Product
}

func NewProductService() *ProductService {
	return &ProductService{
		products: map[string]*Product{
			"1": {ID: "1", Name: "Laptop", Price: 999.99},
			"2": {ID: "2", Name: "Mouse", Price: 29.99},
			"3": {ID: "3", Name: "Keyboard", Price: 79.99},
		},
	}
}

// Pattern 1: func(ctx *request.Context) error
// Manual control - handler writes response directly
func (s *ProductService) ListProducts(ctx *request.Context) error {
	products := make([]*Product, 0, len(s.products))
	for _, p := range s.products {
		products = append(products, p)
	}
	ctx.Resp.Json(map[string]any{
		"success": true,
		"data":    products,
	})
	return nil
}

// Pattern 2: func(ctx *request.Context) (data, error)
// Returns data - framework handles response formatting
// Using route override to avoid conflict with GetProduct
func (s *ProductService) DetailProduct(ctx *request.Context) (*Product, error) {
	id := ctx.Req.PathParam("id", "")
	product, exists := s.products[id]
	if !exists {
		return nil, fmt.Errorf("product not found")
	}
	return product, nil
}

// Pattern 3: func(ctx *request.Context, param any) error
// Auto-bind parameters from request + manual response
func (s *ProductService) GetProduct(ctx *request.Context, req *GetProductRequest) error {
	product, exists := s.products[req.ID]
	if !exists {
		ctx.Resp.WithStatus(404).Json(map[string]any{
			"success": false,
			"error":   "product not found",
		})
		return fmt.Errorf("product not found")
	}
	ctx.Resp.Json(map[string]any{
		"success": true,
		"data":    product,
	})
	return nil
}

// Pattern 4: func(ctx *request.Context, param any) (data, error)
// Auto-bind parameters + return data
func (s *ProductService) SearchProducts(ctx *request.Context, req *SearchProductRequest) ([]*Product, error) {
	results := make([]*Product, 0)
	for _, product := range s.products {
		matchesQuery := req.Query == "" || strings.Contains(
			strings.ToLower(product.Name),
			strings.ToLower(req.Query),
		)
		matchesMinPrice := req.MinPrice == 0 || product.Price >= req.MinPrice
		matchesMaxPrice := req.MaxPrice == 0 || product.Price <= req.MaxPrice
		if matchesQuery && matchesMinPrice && matchesMaxPrice {
			results = append(results, product)
		}
	}
	return results, nil
}

// Pattern 5: func(param any) error
// Pure business logic - no HTTP context
func (s *ProductService) CreateProduct(req *CreateProductRequest) error {
	if req.Price <= 0 {
		return fmt.Errorf("price must be greater than zero")
	}

	id := fmt.Sprintf("%d", len(s.products)+1)
	product := &Product{
		ID:    id,
		Name:  req.Name,
		Price: req.Price,
	}
	s.products[id] = product
	return nil
}

// Pattern 6: func(param any) (data, error)
// Pure business logic with return value
func (s *ProductService) UpdateProduct(req *UpdateProductRequest) (*Product, error) {
	product, exists := s.products[req.ID]
	if !exists {
		return nil, fmt.Errorf("product not found")
	}
	product.Name = req.Name
	product.Price = req.Price
	return product, nil
}

// Pattern 7: func(w http.ResponseWriter, r *http.Request)
// Raw HTTP handler - lowest level control
func (s *ProductService) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, exists := s.products[id]; !exists {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"success":false,"error":"product not found"}`))
		return
	}
	delete(s.products, id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true,"message":"product deleted"}`))
}
