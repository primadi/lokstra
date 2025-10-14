package main

import (
	"fmt"

	lokstra "github.com/primadi/lokstra"
	"github.com/primadi/lokstra/api_client"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/lokstra_registry"
)

// =============================================================================
// PRODUCT HANDLERS
// =============================================================================

func getProducts(c *request.Context) error {
	fmt.Println("\n🎯 Handler: GET /products")

	products := services.GetProduct().GetProducts()

	return c.Api.Ok(map[string]any{
		"products": products,
		"count":    len(products),
	})
}

func getProduct(c *request.Context) error {
	id := c.Req.PathParam("id", "")
	fmt.Printf("\n🎯 Handler: GET /products/%s\n", id)

	if id == "" {
		return c.Api.BadRequest("INVALID_ID", "Product ID is required")
	}

	product := services.GetProduct().GetProduct(id)

	return c.Api.Ok(product)
}

// =============================================================================
// ORDER HANDLERS (with Cross-Router Communication)
// =============================================================================

// ClientRouterContainer for caching router clients
type ClientRouterContainer struct {
	productAPICache *api_client.ClientRouter
}

func (crc *ClientRouterContainer) GetProductAPI() *api_client.ClientRouter {
	crc.productAPICache = lokstra_registry.GetClientRouterCached("product-api", crc.productAPICache)
	return crc.productAPICache
}

var routerClients = &ClientRouterContainer{}

type createOrderRequest struct {
	UserID     string   `json:"user_id"`
	ProductIDs []string `json:"product_ids"`
}

func createOrder(c *request.Context, req *createOrderRequest) error {
	fmt.Println("\n🎯 Handler: POST /orders")

	if req.UserID == "" || len(req.ProductIDs) == 0 {
		return c.Api.BadRequest("INVALID_REQUEST", "user_id and product_ids are required")
	}

	// Validate products exist by calling product-api (cross-router communication!)
	fmt.Println("\n🔄 Cross-Router Call: order-api → product-api")
	productClient := routerClients.GetProductAPI()
	if productClient == nil {
		return c.Api.InternalError("Product API not available")
	}

	var totalAmount float64
	validProducts := []map[string]any{}

	for _, productID := range req.ProductIDs {
		fmt.Printf("   Fetching product %s from product-api...\n", productID)

		// Use FetchAndCast helper for type-safe cross-router calls
		prd, err := lokstra.FetchAndCast[map[string]any](productClient, "/products/"+productID, nil)
		if err != nil {
			fmt.Printf("   ❌ Failed to get product %s: %v\n", productID, err)
			continue
		}
		if price, ok := prd["price"].(float64); ok {
			totalAmount += price
			validProducts = append(validProducts, prd)
			fmt.Printf("   ✅ Product %s: $%.2f\n", productID, price)
		}
	}

	if len(validProducts) == 0 {
		return c.Api.BadRequest("NO_VALID_PRODUCTS", "No valid products found")
	}

	// Create order
	order := services.GetOrder().CreateOrder(req.UserID, req.ProductIDs, totalAmount)
	order["products"] = validProducts
	order["total"] = totalAmount

	fmt.Printf("\n✅ Order created: total $%.2f\n", totalAmount)

	return c.Api.Created(order, "/orders/order-123")
}

func getOrder(c *request.Context) error {
	id := c.Req.PathParam("id", "")
	fmt.Printf("\n🎯 Handler: GET /orders/%s\n", id)

	if id == "" {
		return c.Api.BadRequest("INVALID_ID", "Order ID is required")
	}

	order := services.GetOrder().GetOrder(id)

	return c.Api.Ok(order)
}

// =============================================================================
// HEALTH CHECK
// =============================================================================

func healthCheck(c *request.Context) error {
	serverName := lokstra_registry.GetCurrentServerName()

	return c.Api.Ok(map[string]any{
		"status": "healthy",
		"server": serverName,
		"deployment": map[string]any{
			"server_name": serverName,
		},
	})
}
