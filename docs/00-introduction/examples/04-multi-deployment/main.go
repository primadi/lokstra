package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/docs/00-introduction/examples/04-multi-deployment/appservice"
	"github.com/primadi/lokstra/lokstra_registry"
)

func main() {
	// Parse command line flags
	server := flag.String("server", "microservice.order-server", "Server to run (monolith.api-server or microservice.user-server, microservice.user-server, or microservice.order-server)")
	flag.Parse()

	fmt.Printf("\n")
	fmt.Printf("╔═════════════════════════════════════════════╗\n")
	fmt.Printf("║   LOKSTRA MULTI-DEPLOYMENT DEMO             ║\n")
	fmt.Printf("╚═════════════════════════════════════════════╝\n")
	fmt.Printf("\n")

	// 1. Register service factories
	lokstra_registry.RegisterServiceType("database-factory", DatabaseFactory, nil)
	lokstra_registry.RegisterServiceType("user-service-factory", UserServiceFactory, UserServiceRemoteFactory)
	lokstra_registry.RegisterServiceType("order-service-factory", OrderServiceFactory, OrderServiceRemoteFactory)

	// 2. Load config (loads ALL deployments into Global registry)
	if err := lokstra_registry.LoadAndBuild([]string{"config.yaml"}); err != nil {
		log.Fatal("❌ Failed to load config:", err)
	}

	// 3. Setup ALL routers (deployment-agnostic)
	setupMonolithRouter()
	setupUserServiceRouter()
	setupOrderServiceRouter()

	// 4. Run server
	if err := lokstra_registry.RunServer(*server, 30*time.Second); err != nil {
		log.Fatal("❌ Failed to run server:", err)
	}
}

// ================================================================================
// ROUTER SETUP FUNCTIONS
// These are registered once at startup, framework chooses which to use
// ================================================================================

func setupMonolithRouter() {
	// Get deployment from global registry
	dep, ok := deploy.Global().GetDeployment("monolith")
	if !ok {
		return // Deployment not loaded, skip
	}

	// Only setup if server exists in this deployment
	server, ok := dep.GetServer("api-server")
	if !ok {
		return // Server not in this deployment, skip
	}

	app := server.Apps()[0]

	// Publish services - they'll be available via DI
	userService := service.LazyLoadFrom[appservice.UserService](app, "user-service")
	orderService := service.LazyLoadFrom[appservice.OrderService](app, "order-service")

	// Create router with just the root endpoint
	router := lokstra.NewRouter("monolith-api")

	router.GET("/", func() map[string]any {
		return map[string]any{
			"server":  "monolith.api-server",
			"message": "All services running in one process",
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

	// Add service endpoints
	addUserRoutes(router, userService)
	addOrderRoutes(router, orderService)

	lokstra_registry.RegisterRouter("monolith-api", router)
}

func setupUserServiceRouter() {
	// Get deployment from global registry
	dep, ok := deploy.Global().GetDeployment("microservice")
	if !ok {
		return // Deployment not loaded, skip
	}

	// Only setup if server exists in this deployment
	server, ok := dep.GetServer("user-server")
	if !ok {
		return // Server not in this deployment, skip
	}

	app := server.Apps()[0]

	// Publish user service
	userService := service.LazyLoadFrom[appservice.UserService](app, "user-service")

	// Create router with just the root endpoint
	router := lokstra.NewRouter("user-api")

	router.GET("/", func() map[string]any {
		return map[string]any{
			"server": "user-server",
			"endpoints": []string{
				"GET /users",
				"GET /users/{id}",
			},
		}
	})

	// Add service endpoints
	addUserRoutes(router, userService)

	lokstra_registry.RegisterRouter("user-api", router)
}

func setupOrderServiceRouter() {
	// Get deployment from global registry
	dep, ok := deploy.Global().GetDeployment("microservice")
	if !ok {
		return // Deployment not loaded, skip
	}

	// Only setup if server exists in this deployment
	server, ok := dep.GetServer("order-server")
	if !ok {
		return // Server not in this deployment, skip
	}

	app := server.Apps()[0]

	// Publish order service
	orderService := service.LazyLoadFrom[appservice.OrderService](app, "order-service")

	// Create router with just the root endpoint
	router := lokstra.NewRouter("order-api")

	router.GET("/", func() map[string]any {
		return map[string]any{
			"server": "order-server",
			"endpoints": []string{
				"GET /orders/{id}",
				"GET /users/{user_id}/orders",
			},
			"dependencies": []string{
				"user-service (remote at http://localhost:3004)",
			},
		}
	})

	// Add service endpoints
	addOrderRoutes(router, orderService)

	lokstra_registry.RegisterRouter("order-api", router)
}

// ================================================================================
// SHARED ROUTE BUILDERS
// These functions add service endpoints to routers
// ================================================================================

func addUserRoutes(router lokstra.Router, userService *service.Cached[appservice.UserService]) {
	router.GET("/users", func(ctx *request.Context) error {
		users, err := userService.MustGet().List(&appservice.ListUsersParams{})
		if err != nil {
			return ctx.Api.Error(500, "INTERNAL_ERROR", err.Error())
		}
		return ctx.Api.Ok(users)
	})

	router.GET("/users/{id}", func(ctx *request.Context) error {
		var params appservice.GetUserParams
		if err := ctx.Req.BindAll(&params); err != nil {
			return ctx.Api.BadRequest("INVALID_ID", "Invalid user ID")
		}

		user, err := userService.MustGet().GetByID(&params)
		if err != nil {
			return ctx.Api.Error(404, "NOT_FOUND", err.Error())
		}
		return ctx.Api.Ok(user)
	})
}

func addOrderRoutes(router lokstra.Router, orderService *service.Cached[appservice.OrderService]) {
	router.GET("/orders/{id}", func(ctx *request.Context) error {
		var params appservice.GetOrderParams
		if err := ctx.Req.BindAll(&params); err != nil {
			return ctx.Api.BadRequest("INVALID_ID", "Invalid order ID")
		}

		orderWithUser, err := orderService.MustGet().GetByID(&params)
		if err != nil {
			return ctx.Api.Error(404, "NOT_FOUND", err.Error())
		}
		return ctx.Api.Ok(orderWithUser)
	})

	router.GET("/users/{user_id}/orders", func(ctx *request.Context) error {
		var params appservice.GetUserOrdersParams
		if err := ctx.Req.BindAll(&params); err != nil {
			return ctx.Api.BadRequest("INVALID_ID", "Invalid user ID")
		}

		orders, err := orderService.MustGet().GetByUserID(&params)
		if err != nil {
			return ctx.Api.Error(404, "NOT_FOUND", err.Error())
		}
		return ctx.Api.Ok(orders)
	})
}
