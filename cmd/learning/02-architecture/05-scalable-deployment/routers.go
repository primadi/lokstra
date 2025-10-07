package main

import (
	"fmt"

	lokstra "github.com/primadi/lokstra"
	"github.com/primadi/lokstra/lokstra_registry"
)

func setupRouters() {
	fmt.Println("🔌 Registering Routers...")

	// Product API Router
	productRouter := lokstra.NewRouter("product-api")
	productRouter.GET("/products", getProducts)
	productRouter.GET("/products/{id}", getProduct)
	lokstra_registry.RegisterRouter("product-api", productRouter)
	fmt.Println("   ✅ product-api registered")

	// Order API Router
	orderRouter := lokstra.NewRouter("order-api")
	orderRouter.POST("/orders", createOrder)
	orderRouter.GET("/orders/{id}", getOrder)
	lokstra_registry.RegisterRouter("order-api", orderRouter)
	fmt.Println("   ✅ order-api registered")

	// Health API Router
	healthRouter := lokstra.NewRouter("health-api")
	healthRouter.GET("/health", healthCheck)
	lokstra_registry.RegisterRouter("health-api", healthRouter)
	fmt.Println("   ✅ health-api registered")

	fmt.Println("✅ All routers registered")
}
