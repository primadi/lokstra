package main

import (
	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/route"
	"github.com/primadi/lokstra/lokstra_registry"
)

// =============================================================================
// ROUTER SETUP - Define business logic routers
//
// Setup all routers and register them
// Routers can be deployed as separate microservices
// or combined in a monolithic deployment
//
// Each router handles its own routes and middleware
//
//	Router 1: Product API (product-api)
//	Router 2: Order API (order-api)
//	Router 3: Health Check API (health-api)
//
// =============================================================================
func setupRouters() {
	setupProductRouter()
	setupOrderRouter()
	setupHealthRouter()
}

func setupProductRouter() {
	productRouter := lokstra.NewRouter("product-api")
	productRouter.GET("/products", getProductsHandler, route.WithNameOption("products-list"))
	productRouter.GET("/products/{id}", getProductByIDHandler, route.WithNameOption("product-detail"))
	lokstra_registry.RegisterRouter("product-api", productRouter)
}

func setupOrderRouter() {
	orderRouter := lokstra.NewRouter("order-api")
	orderRouter.POST("/orders", createOrderHandler, route.WithNameOption("order-create"))
	orderRouter.GET("/orders", getOrdersHandler, route.WithNameOption("orders-list"))
	lokstra_registry.RegisterRouter("order-api", orderRouter)
}

func setupHealthRouter() {
	healthRouter := lokstra.NewRouter("health-api")
	healthRouter.GET("/health", func(c *lokstra.RequestContext) error {
		return c.Api.Ok("ok")
	}, route.WithNameOption("health-check"))
	lokstra_registry.RegisterRouter("health-api", healthRouter)
}
