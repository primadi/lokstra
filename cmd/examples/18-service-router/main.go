package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/router"
)

func main() {
	// Create services
	userService := NewUserService()
	productService := NewProductService()

	// 1. Manual Router (Traditional approach - for comparison)
	fmt.Println("=== Creating Manual Router ===")
	manualRouter := CreateManualRouter(userService)

	// 2. Service Router (Convention-based - basic CRUD)
	fmt.Println("\n=== Creating Service Router (Auto-generated) ===")
	serviceRouter := router.NewFromService(
		userService,
		router.DefaultServiceRouterOptions().
			WithPrefix("/api/v1/auto"),
	)
	fmt.Println("✓ Service Router created successfully")

	// 3. Pattern Router (Demonstrating 7 handler patterns)
	fmt.Println("\n=== Creating Pattern Router (7 Handler Patterns) ===")
	patternRouter := router.NewFromService(
		productService,
		router.DefaultServiceRouterOptions().
			WithPrefix("/api/v2/patterns").
			WithRouteOverride("DetailProduct", router.RouteMeta{
				HTTPMethod: "GET",
				Path:       "/products/{id}/detail",
			}),
	)
	fmt.Println("✓ Pattern Router created successfully")

	// Print all routes for comparison
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("REGISTERED ROUTES COMPARISON")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Println("\n--- 1. Manual Router (Traditional) ---")
	manualRouter.PrintRoutes()

	fmt.Println("\n--- 2. Service Router (Auto-generated CRUD) ---")
	serviceRouter.PrintRoutes()

	fmt.Println("\n--- 3. Pattern Router (7 Handler Patterns) ---")
	patternRouter.PrintRoutes()

	// Create application with all three routers
	application := app.New("service-router-comparison", ":3000", manualRouter, serviceRouter, patternRouter)

	// Print usage information
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("SERVER STARTED ON :3000")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Println("\n📌 MANUAL ROUTER ENDPOINTS (Traditional Approach)")
	fmt.Println("   Path: /api/v1/manual")
	fmt.Println("   GET    http://localhost:3000/api/v1/manual/users")
	fmt.Println("   GET    http://localhost:3000/api/v1/manual/users/1")
	fmt.Println("   POST   http://localhost:3000/api/v1/manual/users")
	fmt.Println("   PUT    http://localhost:3000/api/v1/manual/users/1")
	fmt.Println("   DELETE http://localhost:3000/api/v1/manual/users/1")
	fmt.Println("   GET    http://localhost:3000/api/v1/manual/users/search?q=john")

	fmt.Println("\n🚀 SERVICE ROUTER ENDPOINTS (Auto-generated from UserService)")
	fmt.Println("   Path: /api/v1/auto")
	fmt.Println("   GET    http://localhost:3000/api/v1/auto/users")
	fmt.Println("   GET    http://localhost:3000/api/v1/auto/users/1")
	fmt.Println("   POST   http://localhost:3000/api/v1/auto/users")
	fmt.Println("   PUT    http://localhost:3000/api/v1/auto/users/1")
	fmt.Println("   DELETE http://localhost:3000/api/v1/auto/users/1")
	fmt.Println("   GET    http://localhost:3000/api/v1/auto/users/search?q=john")

	fmt.Println("\n✨ PATTERN ROUTER ENDPOINTS (7 Handler Patterns Demo)")
	fmt.Println("   Path: /api/v2/patterns")
	fmt.Println("   GET    http://localhost:3000/api/v2/patterns/products              # Pattern 1")
	fmt.Println("   GET    http://localhost:3000/api/v2/patterns/products/1/detail     # Pattern 2")
	fmt.Println("   GET    http://localhost:3000/api/v2/patterns/products/1            # Pattern 3")
	fmt.Println("   GET    http://localhost:3000/api/v2/patterns/products/search?q=lap # Pattern 4")
	fmt.Println("   POST   http://localhost:3000/api/v2/patterns/products              # Pattern 5")
	fmt.Println("   PUT    http://localhost:3000/api/v2/patterns/products/1            # Pattern 6")
	fmt.Println("   DELETE http://localhost:3000/api/v2/patterns/products/1            # Pattern 7")

	fmt.Println("\n" + strings.Repeat("=", 80))

	if err := application.Run(0); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
