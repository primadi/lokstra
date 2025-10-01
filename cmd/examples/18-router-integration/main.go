package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/config"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/route"
	"github.com/primadi/lokstra/lokstra_registry"
)

// Product represents a product in our e-commerce system
type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
}

// Order represents an order in our system
type Order struct {
	ID       string    `json:"id"`
	UserID   string    `json:"user_id"`
	Products []Product `json:"products"`
	Total    float64   `json:"total"`
	Status   string    `json:"status"`
}

// Fake data for demonstration
var products = []Product{
	{ID: "1", Name: "Laptop", Description: "Gaming Laptop", Price: 1500.00, Category: "Electronics"},
	{ID: "2", Name: "Mouse", Description: "Wireless Mouse", Price: 25.00, Category: "Electronics"},
	{ID: "3", Name: "Book", Description: "Programming Guide", Price: 45.00, Category: "Books"},
}

var orders = []Order{}

// =============================================================================
// PRODUCT API ROUTES (can be deployed as separate microservice)
// =============================================================================

func getProductsHandler(c *request.Context) error {
	return c.Api.Ok(map[string]interface{}{
		"status":   "success",
		"products": products,
	})
}

func getProductByIDHandler(c *request.Context) error {
	id := c.Req.PathParam("id", "")

	for _, p := range products {
		if p.ID == id {
			return c.Api.Ok(map[string]interface{}{
				"status":  "success",
				"product": p,
			})
		}
	}

	return c.Api.NotFound(fmt.Sprintf("Product %s not found", id))
} // =============================================================================
// ORDER API ROUTES (can be deployed as separate microservice)
// =============================================================================

func createOrderHandler(c *request.Context) error {
	var req struct {
		UserID     string   `json:"user_id"`
		ProductIDs []string `json:"product_ids"`
	}

	if err := c.Req.BindBody(&req); err != nil {
		return c.Api.BadRequest("Invalid request body", err.Error())
	}

	// 🔗 Call Product API to validate products and get details
	// This will automatically use local (httptest) or remote (HTTP) calls
	// depending on deployment configuration
	productClient := lokstra_registry.GetRouterClient("product-api")
	if productClient == nil {
		return c.Api.InternalError("Product service unavailable")
	}

	// Fetch product details
	var orderProducts []Product
	var total float64

	for _, productID := range req.ProductIDs {
		resp, err := productClient.GET("/products/" + productID)
		if err != nil {
			return c.Api.InternalError(fmt.Sprintf("Failed to fetch product %s: %v", productID, err))
		}

		var productResp struct {
			Status  string  `json:"status"`
			Product Product `json:"product"`
		}

		if err := lokstra_registry.ParseJSONResponse(resp, &productResp); err != nil {
			return c.Api.InternalError(fmt.Sprintf("Failed to parse product response %s: %v", productID, err))
		}

		if productResp.Status != "success" {
			return c.Api.BadRequest("PRODUCT_NOT_FOUND", fmt.Sprintf("Product %s not found", productID))
		}

		orderProducts = append(orderProducts, productResp.Product)
		total += productResp.Product.Price
	}

	// Create order
	order := Order{
		ID:       fmt.Sprintf("order-%d", len(orders)+1),
		UserID:   req.UserID,
		Products: orderProducts,
		Total:    total,
		Status:   "pending",
	}

	orders = append(orders, order)

	return c.Api.Created(order, "Order created successfully")
}

func getOrdersHandler(c *request.Context) error {
	return c.Api.Ok(map[string]interface{}{
		"status": "success",
		"orders": orders,
	})
}

// =============================================================================
// ROUTER SETUP - Define business logic routers
// =============================================================================

func setupRouters() {
	// Product API Router (can be deployed separately)
	productRouter := lokstra.NewRouter("product-api")
	productRouter.GET("/products", getProductsHandler, route.WithNameOption("products-list"))
	productRouter.GET("/products/{id}", getProductByIDHandler, route.WithNameOption("product-detail"))
	lokstra_registry.RegisterRouter("product-api", productRouter)

	// Order API Router (can be deployed separately)
	orderRouter := lokstra.NewRouter("order-api")
	orderRouter.POST("/orders", createOrderHandler, route.WithNameOption("order-create"))
	orderRouter.GET("/orders", getOrdersHandler, route.WithNameOption("orders-list"))
	lokstra_registry.RegisterRouter("order-api", orderRouter)
}

// =============================================================================
// SERVER SETUP - Create HTTP servers based on deployment
// =============================================================================

func main() {
	fmt.Println("🚀 Starting Lokstra E-commerce Demo with Router Integration")

	// 1. Load configuration
	var cfg config.Config
	configFiles := []string{"cmd/examples/18-router-integration/config.yaml"}

	for _, file := range configFiles {
		if err := config.LoadConfigFile(file, &cfg); err == nil {
			fmt.Printf("📄 Loaded config: %s\n", file)
		} else {
			fmt.Printf("❌ Failed to load config: %s - %v\n", file, err)
		}
	}

	// 2. Setup all routers (business logic)
	setupRouters()

	// 3. Auto-configure router integration from config
	lokstra_registry.AutoConfigureRouterIntegration()

	// 4. Create and configure server
	port := lokstra_registry.GetConfigInt("server-port", 8080)
	serverAddr := ":" + strconv.Itoa(port)

	// Create app with routers
	mainApp := app.New("router-integration-demo", serverAddr)

	// Add routers to app based on configuration
	deploymentType := lokstra_registry.GetConfigString("deployment-type", "monolith")
	currentServerName := lokstra_registry.GetConfigString("server-name", "all")

	switch currentServerName {
	case "product-service":
		fmt.Println("🏪 Running Product Service")
		productRouter := lokstra_registry.GetRouter("product-api")
		if productRouter == nil {
			fmt.Println("❌ Product router is nil!")
		} else {
			fmt.Println("✅ Product router found, adding to app")
			mainApp.AddRouter(productRouter)
		}

	case "order-service":
		fmt.Println("📦 Running Order Service")
		orderRouter := lokstra_registry.GetRouter("order-api")
		if orderRouter == nil {
			fmt.Println("❌ Order router is nil!")
		} else {
			fmt.Println("✅ Order router found, adding to app")
			mainApp.AddRouter(orderRouter)
		}

	// case "all":
	default:
		fmt.Println("🏢 Running All Services (Monolith or Gateway)")

		productRouter := lokstra_registry.GetRouter("product-api")
		if productRouter == nil {
			fmt.Println("❌ Product router is nil!")
		} else {
			fmt.Println("✅ Product router found, adding to app")
			mainApp.AddRouter(productRouter)
		}

		orderRouter := lokstra_registry.GetRouter("order-api")
		if orderRouter == nil {
			fmt.Println("❌ Order router is nil!")
		} else {
			fmt.Println("✅ Order router found, adding to app")
			mainApp.AddRouter(orderRouter)
		}

		// Add a simple health check router
		healthRouter := lokstra.NewRouter("health")
		healthRouter.GET("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := map[string]interface{}{
				"status":     "healthy",
				"deployment": deploymentType,
				"server":     currentServerName,
			}
			json.NewEncoder(w).Encode(response)
		})
		fmt.Println("✅ Health router created, adding to app")
		mainApp.AddRouter(healthRouter)
	}

	// Create server with the app
	server := lokstra.NewServer(serverAddr, mainApp)

	fmt.Printf("🌟 Server starting on port %d\n", port)
	fmt.Printf("🔧 Router Integration Mode: %s\n", lokstra_registry.GetRouterIntegrationMode())

	// Print server info and run (blocks until CTRL+C)
	server.PrintStartInfo()
	if err := server.Run(30 * time.Second); err != nil {
		panic(err)
	}
}
