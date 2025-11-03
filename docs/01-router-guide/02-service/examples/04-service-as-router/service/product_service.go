package service

import (
	"fmt"

	"github.com/primadi/lokstra/docs/01-essentials/02-service/examples/04-service-as-router/contract"
	"github.com/primadi/lokstra/docs/01-essentials/02-service/examples/04-service-as-router/model"
)

// ProductService handles product-related business logic
type ProductService struct {
	products []model.Product
}

// NewProductService creates a new ProductService instance
func NewProductService() *ProductService {
	return &ProductService{
		products: []model.Product{
			{ID: 1, Name: "Laptop", Price: 999.99},
			{ID: 2, Name: "Mouse", Price: 29.99},
			{ID: 3, Name: "Keyboard", Price: 79.99},
		},
	}
}

// List returns all products (optionally filtered by category)
func (s *ProductService) List(p *contract.ListProductsParams) ([]model.Product, error) {
	// In real app, would filter by p.Category
	return s.products, nil
}

// GetByID returns a product by ID
func (s *ProductService) GetByID(p *contract.GetProductParams) (*model.Product, error) {
	for _, product := range s.products {
		if product.ID == p.ID {
			return &product, nil
		}
	}
	return nil, fmt.Errorf("product with ID %d not found", p.ID)
}
