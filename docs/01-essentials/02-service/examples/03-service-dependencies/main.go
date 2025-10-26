package main

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/docs/01-essentials/02-service/examples/03-service-dependencies/model"
	"github.com/primadi/lokstra/docs/01-essentials/02-service/examples/03-service-dependencies/repository"
	svc "github.com/primadi/lokstra/docs/01-essentials/02-service/examples/03-service-dependencies/service"
	"github.com/primadi/lokstra/lokstra_registry"
)

// Package-level service access (cached!)
var (
	userService  = service.LazyLoad[*svc.UserService]("user-service")
	orderService = service.LazyLoad[*svc.OrderService]("order-service")
)

func main() {
	fmt.Println("🚀 Service Dependencies Example - 3 Factory Modes")
	fmt.Println("=" + "===============================================")
	fmt.Println()

	// ============================================================
	// REGISTER SERVICES - ORDER DOESN'T MATTER!
	// Demonstrates 3 factory modes supported by RegisterLazyService
	// ============================================================

	// Mode 1: func() any - No params (simplest!)
	// Use when: No config needed, dependencies from registry
	lokstra_registry.RegisterLazyService("user-repo", repository.NewUserRepository, nil)
	lokstra_registry.RegisterLazyService("user-service", svc.NewUserService, nil)

	// Mode 2: func(cfg map[string]any) any - Config only
	// Use when: Need per-instance configuration
	lokstra_registry.RegisterLazyService("order-repo", repository.NewOrderRepository,
		map[string]any{
			"price_per_item": 15.99, // Custom price per item
		})

	// Mode 3: func(deps, cfg map[string]any) any - Full signature
	// Use when: Need to distinguish between service deps and config values
	lokstra_registry.RegisterLazyService("order-service", svc.NewOrderService,
		map[string]any{
			"max_items": 5, // Maximum items per order
		})

	fmt.Println("📝 All services registered (lazy - not created yet)")
	fmt.Println()

	// ============================================================
	// CREATE ROUTER AND HANDLERS
	// ============================================================

	r := lokstra.NewRouter("api")

	// User endpoints
	r.GET("/users", func() ([]model.User, error) {
		return userService.MustGet().GetAllUsers()
	})

	type GetUserRequest struct {
		ID int `path:"id"`
	}

	r.GET("/users/{id}", func(req GetUserRequest) (*model.User, error) {
		return userService.MustGet().GetUser(req.ID)
	})

	// Order endpoints
	r.GET("/orders", func() ([]model.Order, error) {
		return orderService.MustGet().GetAllOrders()
	})

	type GetOrderRequest struct {
		ID int `path:"id"`
	}

	r.GET("/orders/{id}", func(req GetOrderRequest) (*model.Order, error) {
		return orderService.MustGet().GetOrder(req.ID)
	})

	type CreateOrderRequest struct {
		UserID int      `json:"user_id"`
		Items  []string `json:"items"`
	}

	r.POST("/orders", func(req CreateOrderRequest) (*model.Order, error) {
		return orderService.MustGet().CreateOrder(req.UserID, req.Items)
	})

	// ============================================================
	// START APPLICATION
	// ============================================================

	app := lokstra.NewApp("service-deps", ":3003", r)

	fmt.Println()
	fmt.Println("✅ App created, services will initialize on first access")
	fmt.Println()
	fmt.Println("📡 Server starting on http://localhost:3003")
	fmt.Println()
	fmt.Println("Try these requests:")
	fmt.Println("  GET  http://localhost:3003/users")
	fmt.Println("  GET  http://localhost:3003/users/1")
	fmt.Println("  GET  http://localhost:3003/orders")
	fmt.Println("  POST http://localhost:3003/orders")
	fmt.Println("       { \"user_id\": 1, \"items\": [\"Laptop\", \"Mouse\"] }")
	fmt.Println()

	app.Run(30 * time.Second)
}
