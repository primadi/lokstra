package main

import (
	"fmt"

	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/lokstra_registry"
)

// Mock service implementations for demonstration

// Request/Response types for services
type IDParam struct {
	ID int `path:"id"`
}

// UserService - Simple user management service
type UserService struct {
	users map[int]string
}

func (s *UserService) GetByID(params IDParam) (any, error) {
	name, exists := s.users[params.ID]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return map[string]any{"id": params.ID, "name": name}, nil
}

func (s *UserService) List() (any, error) {
	result := make([]map[string]any, 0)
	for id, name := range s.users {
		result = append(result, map[string]any{"id": id, "name": name})
	}
	return result, nil
}

func UserServiceFactory(deps map[string]any, config map[string]any) any {
	// Initialize with some mock data
	return &UserService{
		users: map[int]string{
			1: "Alice",
			2: "Bob",
			3: "Charlie",
		},
	}
}

// ProductService - Simple product catalog service
type ProductService struct {
	products map[int]map[string]any
}

func (s *ProductService) GetByID(params IDParam) (any, error) {
	product, exists := s.products[params.ID]
	if !exists {
		return nil, fmt.Errorf("product not found")
	}
	return product, nil
}

func (s *ProductService) List() (any, error) {
	result := make([]map[string]any, 0)
	for _, product := range s.products {
		result = append(result, product)
	}
	return result, nil
}

func ProductServiceFactory(deps map[string]any, config map[string]any) any {
	return &ProductService{
		products: map[int]map[string]any{
			1: {"id": 1, "name": "Laptop", "price": 1200.00},
			2: {"id": 2, "name": "Mouse", "price": 25.00},
			3: {"id": 3, "name": "Keyboard", "price": 75.00},
		},
	}
}

// OrderService - Order management service
type OrderService struct {
	orders map[int]map[string]any
}

func (s *OrderService) GetByID(params IDParam) (any, error) {
	order, exists := s.orders[params.ID]
	if !exists {
		return nil, fmt.Errorf("order not found")
	}
	return order, nil
}

func (s *OrderService) List() (any, error) {
	result := make([]map[string]any, 0)
	for _, order := range s.orders {
		result = append(result, order)
	}
	return result, nil
}

func OrderServiceFactory(deps map[string]any, config map[string]any) any {
	return &OrderService{
		orders: map[int]map[string]any{
			1: {"id": 1, "user_id": 1, "total": 1200.00, "status": "completed"},
			2: {"id": 2, "user_id": 2, "total": 100.00, "status": "pending"},
		},
	}
}

// Mock Middleware - Must return request.HandlerFunc
func LoggingMiddleware(config map[string]any) request.HandlerFunc {
	prefix := "API"
	if p, ok := config["prefix"]; ok {
		prefix = p.(string)
	}

	return request.HandlerFunc(func(c *request.Context) error {
		fmt.Printf("[%s] %s %s\n", prefix, c.R.Method, c.R.URL.Path)
		return c.Next()
	})
}

func AuthMiddleware(config map[string]any) request.HandlerFunc {
	return request.HandlerFunc(func(c *request.Context) error {
		fmt.Println("[AUTH] Checking authentication...")
		return c.Next()
	})
}

func CorsMiddleware(config map[string]any) request.HandlerFunc {
	return request.HandlerFunc(func(c *request.Context) error {
		fmt.Println("[CORS] Setting CORS headers...")
		c.W.Header().Set("Access-Control-Allow-Origin", "*")
		return c.Next()
	})
}

// Register all service types and middleware
func registerServiceTypes() {
	// Register middleware factories
	lokstra_registry.RegisterMiddlewareFactory("logging", LoggingMiddleware)
	lokstra_registry.RegisterMiddlewareFactory("auth", AuthMiddleware)
	lokstra_registry.RegisterMiddlewareFactory("cors", CorsMiddleware)

	// Register service types with REST metadata
	lokstra_registry.RegisterServiceType("user-service-type",
		UserServiceFactory,
		nil,
		deploy.WithResource("user", "users"),
		deploy.WithConvention("rest"),
	)

	lokstra_registry.RegisterServiceType("product-service-type",
		ProductServiceFactory,
		nil,
		deploy.WithResource("product", "products"),
		deploy.WithConvention("rest"),
	)

	lokstra_registry.RegisterServiceType("order-service-type",
		OrderServiceFactory,
		nil,
		deploy.WithResource("order", "orders"),
		deploy.WithConvention("rest"),
	)
}
