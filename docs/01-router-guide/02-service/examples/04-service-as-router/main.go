package main

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/deploy"
	coreService "github.com/primadi/lokstra/core/service"

	"github.com/primadi/lokstra/docs/01-router-guide/02-service/examples/04-service-as-router/contract"
	"github.com/primadi/lokstra/docs/01-router-guide/02-service/examples/04-service-as-router/model"
	"github.com/primadi/lokstra/docs/01-router-guide/02-service/examples/04-service-as-router/service"
	"github.com/primadi/lokstra/lokstra_registry"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("Service as Router Example")
	fmt.Println("========================================")
	fmt.Println()

	// ========================================================================
	// STEP 1: Register Service Types with Metadata
	// ========================================================================
	// No need for separate Remote service structs!
	// Metadata is provided via options in RegisterServiceType

	fmt.Println("📝 Registering services with metadata...")

	lokstra_registry.RegisterServiceType(
		"user-service",
		service.NewUserService,
		nil, // No remote factory needed for this example
		deploy.WithResource("user", "users"),
		deploy.WithConvention("rest"),
	)

	lokstra_registry.RegisterServiceType(
		"product-service",
		service.NewProductService,
		nil,
		deploy.WithResource("product", "products"),
		deploy.WithConvention("rest"),
	)

	fmt.Println("✅ Services registered with metadata")
	fmt.Println()

	// ========================================================================
	// STEP 2: Register Lazy Service Instances
	// ========================================================================

	lokstra_registry.RegisterLazyService("user-service", service.NewUserService, nil)
	lokstra_registry.RegisterLazyService("product-service", service.NewProductService, nil)

	// ========================================================================
	// STEP 3: Create Routers
	// ========================================================================

	fmt.Println("🚀 Generating routers...")
	fmt.Println()

	// Method 1: Manual router (traditional way - tedious!)
	manualRouter := lokstra.NewRouter("manual-api")
	var manualUserSvc = coreService.LazyLoad[*service.UserService]("user-service")

	manualRouter.GET("/manual/users", func() ([]model.User, error) {
		return manualUserSvc.MustGet().List(&contract.ListUsersParams{})
	})
	manualRouter.GET("/manual/users/{id}", func(p *contract.GetUserParams) (*model.User, error) {
		return manualUserSvc.MustGet().GetByID(p)
	})

	fmt.Println("✅ Manual router created (2 endpoints manually defined):")
	fmt.Println("   GET /manual/users")
	fmt.Println("   GET /manual/users/{id}")
	fmt.Println()

	// Method 2: Auto-generated router (Service as Router - MAGIC!)
	// Metadata comes from RegisterServiceType options - no Remote struct needed!
	// Service instance is automatically created from factory!
	autoUserRouter := lokstra_registry.NewRouterFromServiceType("user-service-factory")

	fmt.Println("✅ Auto-generated router for user-service (ZERO manual routing!):")
	fmt.Println("   GET /users       → List() method")
	fmt.Println("   GET /users/{id}  → GetByID() method")
	fmt.Println()

	autoProductRouter := lokstra_registry.NewRouterFromServiceType("product-service-factory")

	fmt.Println("✅ Auto-generated router for product-service (ZERO manual routing!):")
	fmt.Println("   GET /products       → List() method")
	fmt.Println("   GET /products/{id}  → GetByID() method")
	fmt.Println()

	// ========================================================================
	// STEP 4: Create App with All Routers
	// ========================================================================

	app := lokstra.NewApp("service-as-router-demo", ":3000",
		manualRouter,
		autoUserRouter,
		autoProductRouter,
	)

	fmt.Println("========================================")
	fmt.Println("Server: http://localhost:3000")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("Manual Endpoints:")
	fmt.Println("  GET /manual/users       - List users (manual)")
	fmt.Println("  GET /manual/users/{id}  - Get user (manual)")
	fmt.Println()
	fmt.Println("Auto-Generated Endpoints:")
	fmt.Println("  GET /users              - List users (auto)")
	fmt.Println("  GET /users/{id}         - Get user (auto)")
	fmt.Println("  GET /products           - List products (auto)")
	fmt.Println("  GET /products/{id}      - Get product (auto)")
	fmt.Println()
	fmt.Println("🎯 Compare manual vs auto-generated!")
	fmt.Println("========================================")

	if err := app.Run(30 * time.Second); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
