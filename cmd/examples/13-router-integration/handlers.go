package main

import (
	"fmt"
	"net/http"

	"github.com/primadi/lokstra/common/cast"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response/api_formatter"
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
	ID       string     `json:"id"`
	UserID   string     `json:"user_id"`
	Products []*Product `json:"products"`
	Total    float64    `json:"total"`
	Status   string     `json:"status"`
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
	return c.Api.Ok(products)
}

func getProductByIDHandler(c *request.Context) error {
	id := c.Req.PathParam("id", "")

	for _, p := range products {
		if p.ID == id {
			return c.Api.Ok(p)
		}
	}

	return c.Api.NotFound(fmt.Sprintf("Product %s not found", id))
}

// =============================================================================
// ORDER API ROUTES (can be deployed as separate microservice)
// =============================================================================

type createOrderParam struct {
	UserID     string   `json:"user_id"`
	ProductIDs []string `json:"product_ids"`
}

var productClient *lokstra_registry.ClientRouter

func getProduct(c *request.Context, productID string) (*Product, error) {
	resp, err := productClient.GET("/products/" + productID)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, c.Api.InternalError(fmt.Sprintf("Failed to fetch product %s: %v", productID, err))
	}

	formatter := api_formatter.GetGlobalFormatter()

	clientResp := &api_formatter.ClientResponse{}
	if err := formatter.ParseClientResponse(resp, clientResp); err != nil {
		return nil, c.Api.InternalError(fmt.Sprintf("Failed to parse product response %s: %v", productID, err))
	}

	if clientResp.Status != "success" {
		return nil, c.Api.BadRequest("PRODUCT_NOT_FOUND", fmt.Sprintf("Product %s not found", productID))
	}

	product := &Product{}
	if err := cast.ToStruct(clientResp.Data, product, true); err != nil {
		return nil, c.Api.InternalError(fmt.Sprintf("Failed to cast product data %s: %v", productID, err))
	}

	return product, nil
}

func createOrderHandler(c *request.Context, param *createOrderParam) error {
	if param.UserID == "" || len(param.ProductIDs) == 0 {
		return c.Api.BadRequest("INVALID_REQUEST", "UserID and ProductIDs are required")
	}

	// Automatically resolves product-api location based on deployment configuration
	// No need for manual RegisterRouterURL - auto-discovery handles this
	productClient = lokstra_registry.GetClientRouter("product-api", productClient)
	if productClient == nil {
		return c.Api.InternalError("Product service unavailable")
	}

	// Fetch product details
	var orderProducts []*Product
	var total float64

	for _, productID := range param.ProductIDs {
		resp, err := productClient.GET("/products/" + productID)
		if err != nil || resp.StatusCode != http.StatusOK {
			return c.Api.InternalError(fmt.Sprintf("Failed to fetch product %s: %v", productID, err))
		}

		product, err := getProduct(c, productID)
		if err != nil {
			return err
		}

		orderProducts = append(orderProducts, product)
		total += product.Price
	}

	// Create order
	order := Order{
		ID:       fmt.Sprintf("order-%d", len(orders)+1),
		UserID:   param.UserID,
		Products: orderProducts,
		Total:    total,
		Status:   "pending",
	}

	orders = append(orders, order)

	return c.Api.Created(order, "Order created successfully")
}

func getOrdersHandler(c *request.Context) error {
	return c.Api.Ok(orders)
}
