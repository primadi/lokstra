package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/lokstra_registry"
)

// ========================================
// Models
// ========================================

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Order struct {
	ID      int     `json:"id"`
	UserID  int     `json:"user_id"`
	Product string  `json:"product"`
	Amount  float64 `json:"amount"`
}

// ========================================
// Database (In-Memory)
// ========================================

type Database struct {
	users  map[int]*User
	orders map[int]*Order
	mu     sync.RWMutex
}

func NewDatabase() *Database {
	db := &Database{
		users:  make(map[int]*User),
		orders: make(map[int]*Order),
	}

	// Seed users
	db.users[1] = &User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	db.users[2] = &User{ID: 2, Name: "Bob", Email: "bob@example.com"}

	// Seed orders
	db.orders[1] = &Order{ID: 1, UserID: 1, Product: "Laptop", Amount: 1200.00}
	db.orders[2] = &Order{ID: 2, UserID: 1, Product: "Mouse", Amount: 25.00}
	db.orders[3] = &Order{ID: 3, UserID: 2, Product: "Keyboard", Amount: 75.00}

	return db
}

func (db *Database) GetUser(id int) (*User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	user, exists := db.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (db *Database) GetAllUsers() ([]*User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	users := make([]*User, 0, len(db.users))
	for _, user := range db.users {
		users = append(users, user)
	}
	return users, nil
}

func (db *Database) GetOrder(id int) (*Order, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	order, exists := db.orders[id]
	if !exists {
		return nil, fmt.Errorf("order not found")
	}
	return order, nil
}

func (db *Database) GetOrdersByUser(userID int) ([]*Order, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	orders := make([]*Order, 0)
	for _, order := range db.orders {
		if order.UserID == userID {
			orders = append(orders, order)
		}
	}
	return orders, nil
}

// ========================================
// Services
// ========================================

// User Service
type UserService struct {
	DB *service.Cached[*Database]
}

type GetUserParams struct {
	ID int `path:"id"`
}

type ListUsersParams struct{}

func (s *UserService) GetByID(p *GetUserParams) (*User, error) {
	return s.DB.Get().GetUser(p.ID)
}

func (s *UserService) List(p *ListUsersParams) ([]*User, error) {
	return s.DB.Get().GetAllUsers()
}

// Order Service
type OrderService struct {
	DB    *service.Cached[*Database]
	Users *service.Cached[*UserService] // Lazy cross-service dependency
}

type GetOrderParams struct {
	ID int `path:"id"`
}

type GetUserOrdersParams struct {
	UserID int `path:"user_id"`
}

type OrderWithUser struct {
	Order *Order `json:"order"`
	User  *User  `json:"user"`
}

func (s *OrderService) GetByID(p *GetOrderParams) (*OrderWithUser, error) {
	// Get order
	order, err := s.DB.Get().GetOrder(p.ID)
	if err != nil {
		return nil, err
	}

	// Get associated user (cross-service call)
	// In monolith: Direct method call
	// In microservices: HTTP call to user-service
	user, err := s.Users.Get().GetByID(&GetUserParams{ID: order.UserID})
	if err != nil {
		return nil, fmt.Errorf("order found but user not found: %v", err)
	}

	return &OrderWithUser{
		Order: order,
		User:  user,
	}, nil
}

func (s *OrderService) GetByUserID(p *GetUserOrdersParams) ([]*Order, error) {
	// Verify user exists (cross-service call)
	_, err := s.Users.Get().GetByID(&GetUserParams{ID: p.UserID})
	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	return s.DB.Get().GetOrdersByUser(p.UserID)
}

// ========================================
// Handlers
// ========================================

// Package-level cached services (optimal pattern)
var (
	userService  = service.LazyLoad[*UserService]("users")
	orderService = service.LazyLoad[*OrderService]("orders")
)

func listUsersHandler(ctx *request.Context) error {
	users, err := userService.MustGet().List(&ListUsersParams{})
	if err != nil {
		return ctx.Api.Error(500, "INTERNAL_ERROR", err.Error())
	}
	return ctx.Api.Ok(users)
}

func getUserHandler(ctx *request.Context) error {
	var params GetUserParams
	if err := ctx.Req.BindPath(&params); err != nil {
		return ctx.Api.BadRequest("INVALID_ID", "Invalid user ID")
	}

	user, err := userService.MustGet().GetByID(&params)
	if err != nil {
		return ctx.Api.Error(404, "NOT_FOUND", err.Error())
	}
	return ctx.Api.Ok(user)
}

func getOrderHandler(ctx *request.Context) error {
	var params GetOrderParams
	if err := ctx.Req.BindPath(&params); err != nil {
		return ctx.Api.BadRequest("INVALID_ID", "Invalid order ID")
	}

	orderWithUser, err := orderService.MustGet().GetByID(&params)
	if err != nil {
		return ctx.Api.Error(404, "NOT_FOUND", err.Error())
	}
	return ctx.Api.Ok(orderWithUser)
}

func getUserOrdersHandler(ctx *request.Context) error {
	var params GetUserOrdersParams
	if err := ctx.Req.BindPath(&params); err != nil {
		return ctx.Api.BadRequest("INVALID_ID", "Invalid user ID")
	}

	orders, err := orderService.MustGet().GetByUserID(&params)
	if err != nil {
		return ctx.Api.Error(404, "NOT_FOUND", err.Error())
	}
	return ctx.Api.Ok(orders)
}

// ========================================
// Deployment Configurations
// ========================================

func runMonolith() {
	log.Println("🚀 Starting MONOLITH deployment")
	log.Println("   All services in one process")

	// Register services
	registerServices()

	// Create combined router
	r := lokstra.NewRouter("monolith")

	r.GET("/", func() map[string]any {
		return map[string]any{
			"deployment": "monolith",
			"message":    "All services running in one process",
			"endpoints": map[string]any{
				"users": []string{
					"GET /users",
					"GET /users/{id}",
				},
				"orders": []string{
					"GET /orders/{id}",
					"GET /users/{user_id}/orders",
				},
			},
		}
	})

	// Register user routes
	r.GET("/users", listUsersHandler)
	r.GET("/users/{id}", getUserHandler)

	// Register order routes
	r.GET("/orders/{id}", getOrderHandler)
	r.GET("/users/{user_id}/orders", getUserOrdersHandler)

	// Run
	app := lokstra.NewApp("monolith", ":3003", r)
	app.PrintStartInfo()
	app.Run(30 * time.Second)
}

func runUserService() {
	log.Println("🚀 Starting USER-SERVICE (microservices mode)")
	log.Println("   Only user-related endpoints")

	// Register only user service dependencies
	lokstra_registry.RegisterServiceFactory("db", func() any {
		return NewDatabase()
	})

	lokstra_registry.RegisterServiceFactory("users", func() any {
		return &UserService{
			DB: service.LazyLoad[*Database]("db"),
		}
	})

	// Create user router
	r := lokstra.NewRouter("user-service")

	r.GET("/", func() map[string]any {
		return map[string]any{
			"service":    "user-service",
			"deployment": "microservices",
			"endpoints": []string{
				"GET /users",
				"GET /users/{id}",
			},
		}
	})

	// Register user routes
	r.GET("/users", listUsersHandler)
	r.GET("/users/{id}", getUserHandler)

	// Run
	app := lokstra.NewApp("user-service", ":3004", r)
	app.PrintStartInfo()
	app.Run(30 * time.Second)
}

func runOrderService() {
	log.Println("🚀 Starting ORDER-SERVICE (microservices mode)")
	log.Println("   Only order-related endpoints")

	// Register order service dependencies
	// Note: In real microservices, this would make HTTP calls to user-service
	// For this demo, we keep the DB shared for simplicity
	registerServices()

	// Create order router
	r := lokstra.NewRouter("order-service")

	r.GET("/", func() map[string]any {
		return map[string]any{
			"service":    "order-service",
			"deployment": "microservices",
			"endpoints": []string{
				"GET /orders/{id}",
				"GET /users/{user_id}/orders",
			},
			"dependencies": []string{
				"user-service (for user data)",
			},
		}
	})

	// Register order routes
	r.GET("/orders/{id}", getOrderHandler)
	r.GET("/users/{user_id}/orders", getUserOrdersHandler)

	// Run
	app := lokstra.NewApp("order-service", ":3005", r)
	app.PrintStartInfo()
	app.Run(30 * time.Second)
}

// ========================================
// Service Registration
// ========================================

func registerServices() {
	// Register database
	lokstra_registry.RegisterServiceFactory("db", func() any {
		return NewDatabase()
	})

	// Register user service
	lokstra_registry.RegisterServiceFactory("users", func() any {
		return &UserService{
			DB: service.LazyLoad[*Database]("db"),
		}
	})

	// Register order service with user dependency
	lokstra_registry.RegisterServiceFactory("orders", func() any {
		return &OrderService{
			DB:    service.LazyLoad[*Database]("db"),
			Users: service.LazyLoad[*UserService]("users"), // Cross-service dependency
		}
	})
}

// ========================================
// Main
// ========================================

func main() {
	deployment := flag.String("mode", "monolith", "Deployment mode: monolith, user-service, or order-service")
	flag.Parse()

	fmt.Printf("\n")
	fmt.Printf("╔═══════════════════════════════════════════════╗\n")
	fmt.Printf("║   LOKSTRA MULTI-DEPLOYMENT DEMO               ║\n")
	fmt.Printf("╚═══════════════════════════════════════════════╝\n")
	fmt.Printf("\n")

	switch *deployment {
	case "monolith":
		fmt.Println("📦 Mode: MONOLITH")
		fmt.Println("   • All services in one process")
		fmt.Println("   • Port: 3003")
		fmt.Println()
		runMonolith()

	case "user-service":
		fmt.Println("🔷 Mode: USER-SERVICE (Microservices)")
		fmt.Println("   • Only user endpoints")
		fmt.Println("   • Port: 3004")
		fmt.Println()
		runUserService()

	case "order-service":
		fmt.Println("🔶 Mode: ORDER-SERVICE (Microservices)")
		fmt.Println("   • Only order endpoints")
		fmt.Println("   • Port: 3005")
		fmt.Println()
		runOrderService()

	default:
		log.Fatalf("Unknown deployment mode: %s\nUse: monolith, user-service, or order-service", *deployment)
	}
}
