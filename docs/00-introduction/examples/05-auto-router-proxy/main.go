package main

import (
	"flag"
	"log"

	lokstra "github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/proxy"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/router/autogen"
	"github.com/primadi/lokstra/core/router/convention"
)

func main() {
	// Parse flags
	mode := flag.String("mode", "client", "Run mode: 'server' or 'client'")
	flag.Parse()

	switch *mode {
	case "server":
		runUserServer()
	case "client":
		runOrderClient()
	default:
		log.Fatal("Invalid mode. Use -mode=server or -mode=client")
	}
}

// ========================================
// USER SERVICE (interface + impl + remote)
// ========================================

// UserService is the interface exposed to consumers
type UserService interface {
	List(ctx *request.Context) error
	Get(ctx *request.Context) error
	Create(ctx *request.Context) error
	Update(ctx *request.Context) error
}

// UserServiceImpl is the local implementation (server-side)
type UserServiceImpl struct{}

func (s *UserServiceImpl) List(ctx *request.Context) error {
	log.Println("📡 UserService.List() called (local)")
	return ctx.Api.Ok(map[string]any{
		"users": []map[string]any{
			{"id": "1", "name": "Alice", "email": "alice@example.com"},
			{"id": "2", "name": "Bob", "email": "bob@example.com"},
		},
	})
}

func (s *UserServiceImpl) Get(ctx *request.Context) error {
	id := ctx.Req.PathParam("id", "")
	log.Printf("📡 UserService.Get(%s) called (local)\n", id)
	return ctx.Api.Ok(map[string]any{
		"user": map[string]any{
			"id":    id,
			"name":  "User " + id,
			"email": "user" + id + "@example.com",
		},
	})
}

func (s *UserServiceImpl) Create(ctx *request.Context) error {
	log.Println("📡 UserService.Create() called (local)")
	return ctx.Api.Created(map[string]any{"message": "User created"}, "User created successfully")
}

func (s *UserServiceImpl) Update(ctx *request.Context) error {
	id := ctx.Req.PathParam("id", "")
	log.Printf("📡 UserService.Update(%s) called (local)\n", id)
	return ctx.Api.Ok(map[string]any{"message": "User " + id + " updated"})
}

// UserServiceRemote implements UserService but forwards calls to proxy.Service
type UserServiceRemote struct {
	userProxy *proxy.Service
}

func NewUserServiceRemote(p *proxy.Service) *UserServiceRemote {
	return &UserServiceRemote{userProxy: p}
}

func (s *UserServiceRemote) List(ctx *request.Context) error {
	return proxy.Call(s.userProxy, "List", ctx)
}

func (s *UserServiceRemote) Get(ctx *request.Context) error {
	return proxy.Call(s.userProxy, "Get", ctx)
}

func (s *UserServiceRemote) Create(ctx *request.Context) error {
	return proxy.Call(s.userProxy, "Create", ctx)
}

func (s *UserServiceRemote) Update(ctx *request.Context) error {
	return proxy.Call(s.userProxy, "Update", ctx)
}

// runUserServer starts a local user service using UserServiceImpl
func runUserServer() {
	log.Println("🚀 Starting USER SERVICE (Auto-Router Server)")

	userService := &UserServiceImpl{}

	conversionRule := autogen.ConversionRule{
		Convention:     convention.REST,
		Resource:       "user",
		ResourcePlural: "users",
	}

	routerOverride := autogen.RouteOverride{PathPrefix: "/api/v1", Hidden: []string{}}

	// Generate router from the concrete implementation
	router := autogen.NewFromService(userService, conversionRule, routerOverride)

	router.GET("/", func(ctx *request.Context) error {
		return ctx.Api.Ok(map[string]any{"service": "user-service", "message": "Auto-Router enabled REST API"})
	})

	app := lokstra.NewApp("user-service", ":3000", router)
	app.PrintStartInfo()

	if err := app.Run(0); err != nil {
		log.Fatal("❌ Failed to start user service:", err)
	}
}

// ========================================
// ORDER SERVICE (Client using Proxy)
// ========================================

type OrderService struct {
	userSvc UserService
}

func NewOrderService(userSvc UserService) *OrderService {
	return &OrderService{userSvc: userSvc}
}

func (s *OrderService) GetOrder(ctx *request.Context) error {
	orderID := ctx.Req.PathParam("id", "")
	log.Printf("📦 OrderService.GetOrder(%s) called\n", orderID)

	return ctx.Api.Ok(map[string]any{
		"order": map[string]any{
			"id":     orderID,
			"status": "shipped",
			"total":  12500,
		},
	})
}

func (s *OrderService) GetUserOrders(ctx *request.Context) error {
	userID := ctx.Req.PathParam("user_id", "")
	log.Printf("📦 OrderService.GetUserOrders(%s) - fetching user list via proxy...\n", userID)

	// Call List() method from UserService (remote or local)
	// This will forward to GET /api/v1/users
	return s.userSvc.List(ctx)
}

func runOrderClient() {
	log.Println("🚀 Starting ORDER SERVICE (Proxy Client)")

	// Create a UserServiceRemote backed by proxy.Service
	userProxy := proxy.NewService(
		"http://localhost:3000",
		autogen.ConversionRule{
			Convention:     convention.REST,
			Resource:       "user",
			ResourcePlural: "users",
		},
		autogen.RouteOverride{
			PathPrefix: "/api/v1",
		},
	)
	userRemote := NewUserServiceRemote(userProxy)

	// Create order service with user service (remote impl)
	orderService := NewOrderService(userRemote)

	// Create router
	router := lokstra.NewRouter("order-service")

	router.GET("/", func(ctx *request.Context) error {
		return ctx.Api.Ok(map[string]any{
			"service":      "order-service",
			"message":      "Using proxy.Service to access user-service",
			"user_service": "http://localhost:3000/api/v1/users",
		})
	})

	router.GET("/orders/{id}", orderService.GetOrder)
	router.GET("/users/{user_id}/orders", orderService.GetUserOrders)

	app := lokstra.NewApp("order-service", ":3002", router)
	app.PrintStartInfo()

	if err := app.Run(0); err != nil {
		log.Fatal("❌ Failed to start order service:", err)
	}
}
