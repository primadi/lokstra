package service

import (
	"fmt"

	"github.com/primadi/lokstra/docs/01-router-guide/02-service/examples/03-service-dependencies/model"
	"github.com/primadi/lokstra/docs/01-router-guide/02-service/examples/03-service-dependencies/repository"
	"github.com/primadi/lokstra/lokstra_registry"
)

type OrderService struct {
	userService *UserService
	orderRepo   *repository.OrderRepository
	maxItems    int
}

// NewOrderService - Mode 3: Full signature (deps + config)
// Shows clear separation: deps for services, cfg for config values
func NewOrderService(deps, cfg map[string]any) *OrderService {
	// Get service dependencies from registry
	userSvc := lokstra_registry.MustGetService[*UserService]("user-service")
	orderRepo := lokstra_registry.MustGetService[*repository.OrderRepository]("order-repo")

	// Get config values
	maxItems := 10 // default
	if max, ok := cfg["max_items"].(int); ok {
		maxItems = max
	}

	fmt.Printf("✅ OrderService created (Mode 3: deps + config, maxItems=%d)\n", maxItems)

	return &OrderService{
		userService: userSvc,
		orderRepo:   orderRepo,
		maxItems:    maxItems,
	}
}

func (s *OrderService) CreateOrder(userID int, items []string) (*model.Order, error) {
	// Validate user exists
	_, err := s.userService.GetUser(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user: %w", err)
	}

	// Validate max items
	if len(items) > s.maxItems {
		return nil, fmt.Errorf("too many items (max %d)", s.maxItems)
	}

	// Create order
	return s.orderRepo.Create(userID, items)
}

func (s *OrderService) GetOrder(id int) (*model.Order, error) {
	return s.orderRepo.FindByID(id)
}

func (s *OrderService) GetAllOrders() ([]model.Order, error) {
	return s.orderRepo.FindAll()
}
