package application

import (
	"fmt"

	"github.com/primadi/lokstra/common/logger"
)

// Example RouterService with Init() method

// @RouterService name="product-api", prefix="/api/products"
type ProductAPIService struct {
	// @Inject "product-repository"
	ProductRepo ProductRepository

	// @InjectCfgValue key="api.products.max-items", default=100
	MaxItems int

	// Internal cache (initialized in Init())
	categoryCache map[string][]string
}

// Init is called after dependency injection
func (s *ProductAPIService) Init() error {
	// Initialize internal cache
	s.categoryCache = make(map[string][]string)

	// Validate configuration
	if s.MaxItems <= 0 {
		return fmt.Errorf("max items must be positive, got %d", s.MaxItems)
	}

	// Pre-load data if needed
	logger.LogInfo("✅ ProductAPIService initialized: max_items=%d", s.MaxItems)
	return nil
}

// @Route "GET /"
func (s *ProductAPIService) ListProducts(req *ListProductsRequest) (*ProductListResponse, error) {
	// Use cache
	_ = s.categoryCache
	return &ProductListResponse{}, nil
}

// @Route "GET /{id}"
func (s *ProductAPIService) GetProduct(req *GetProductRequest) (*ProductResponse, error) {
	return &ProductResponse{}, nil
}

// DTOs
type ListProductsRequest struct {
	Category string `query:"category"`
}

type GetProductRequest struct {
	ID string `path:"id"`
}

type ProductResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ProductListResponse struct {
	Products []*ProductResponse `json:"products"`
}

// Domain interface
type ProductRepository interface {
	FindAll() ([]*ProductResponse, error)
}
