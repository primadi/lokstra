package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/config"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/route"
	"github.com/primadi/lokstra/core/server"
	"github.com/primadi/lokstra/lokstra_registry"
)

// Domain Models
type Product struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type Order struct {
	ID       string    `json:"id"`
	UserID   string    `json:"user_id"`
	Products []Product `json:"products"`
	Total    float64   `json:"total"`
	Status   string    `json:"status"`
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type PaymentRequest struct {
	OrderID string  `json:"order_id"`
	UserID  string  `json:"user_id"`
	Amount  float64 `json:"amount"`
}

type PaymentResponse struct {
	Status        string `json:"status"`
	TransactionID string `json:"transaction_id"`
}

// Service Interfaces (Contract between services)
type ProductService interface {
	GetProducts() ([]Product, error)
	ValidateProducts(ids []string) ([]Product, error)
	GetProductPrices(ids []string) (map[string]float64, error)
}

type UserService interface {
	GetUsers() ([]User, error)
	ValidateUser(userID string) (*User, error)
}

type PaymentService interface {
	ProcessPayment(req PaymentRequest) (*PaymentResponse, error)
}

type AnalyticsService interface {
	TrackOrderCreated(order Order) error
	GetAnalytics() (map[string]interface{}, error)
}

// Business Logic Handlers
func getProducts(w http.ResponseWriter, r *http.Request) {
	// ✨ Smart Service Integration - Framework automatically chooses implementation
	var defaultProductSvc ProductService = &LocalProductService{}
	productSvc := lokstra_registry.SmartGetService("product-service", defaultProductSvc).(ProductService)

	products, err := productSvc.GetProducts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"products":    products,
		"source":      "product-service",
		"deployment":  lokstra_registry.GetConfigString("deployment-type", "monolith"),
		"integration": "smart-service", // ✨ Indicates framework-handled service integration
	})
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	// Parse request (simplified)
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = "u1" // default for demo
	}

	productIDs := []string{"p1", "p2", "p3"} // simplified - normally from request body

	// ✨ Smart Service Integration - Framework automatically chooses implementation
	var defaultProductSvc ProductService = &LocalProductService{}
	var defaultUserSvc UserService = &LocalUserService{}
	var defaultPaymentSvc PaymentService = &LocalPaymentService{}
	var defaultAnalyticsSvc AnalyticsService = &LocalAnalyticsService{}

	productSvc := lokstra_registry.SmartGetService("product-service", defaultProductSvc).(ProductService)
	userSvc := lokstra_registry.SmartGetService("user-service", defaultUserSvc).(UserService)
	paymentSvc := lokstra_registry.SmartGetService("payment-service", defaultPaymentSvc).(PaymentService)
	analyticsSvc := lokstra_registry.SmartGetService("analytics-service", defaultAnalyticsSvc).(AnalyticsService)

	// 1. Validate user exists
	user, err := userSvc.ValidateUser(userID)
	if err != nil {
		http.Error(w, "Invalid user: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 2. Validate products and get prices
	products, err := productSvc.ValidateProducts(productIDs)
	if err != nil {
		http.Error(w, "Invalid products: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 3. Calculate total
	var total float64
	for _, product := range products {
		total += product.Price
	}

	// 4. Process payment
	paymentReq := PaymentRequest{
		OrderID: "ord-123",
		UserID:  userID,
		Amount:  total,
	}
	paymentResp, err := paymentSvc.ProcessPayment(paymentReq)
	if err != nil {
		http.Error(w, "Payment failed: "+err.Error(), http.StatusPaymentRequired)
		return
	}

	// 5. Create order
	order := Order{
		ID:       "ord-123",
		UserID:   userID,
		Products: products,
		Total:    total,
		Status:   "completed",
	}

	// 6. Track analytics
	analyticsSvc.TrackOrderCreated(order)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"order":         order,
		"payment":       paymentResp,
		"user":          user,
		"source":        "order-service",
		"deployment":    lokstra_registry.GetConfigString("deployment-type", "monolith"),
		"inter_service": "✅ Called product, user, payment & analytics services",
	})
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	// Simulate user management
	users := []User{
		{ID: "u1", Name: "John Doe", Email: "john@example.com"},
		{ID: "u2", Name: "Jane Smith", Email: "jane@example.com"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users":  users,
		"source": "user-service",
	})
}

func processPayment(w http.ResponseWriter, r *http.Request) {
	// Simulate payment processing
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"amount":  1109.97,
		"source":  "payment-service",
		"gateway": lokstra_registry.GetConfigString("payment-gateway", "stripe"),
	})
}

func getAnalytics(w http.ResponseWriter, r *http.Request) {
	// Simulate analytics data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_orders":  156,
		"total_revenue": 12543.67,
		"active_users":  89,
		"source":        "analytics-service",
		"cache_enabled": lokstra_registry.GetConfigBool("enable-cache", false),
	})
}

// Service Implementations
// =====================

// Local Product Service (for monolith and direct calls)
type LocalProductService struct{}

func (s *LocalProductService) GetProducts() ([]Product, error) {
	return []Product{
		{ID: "p1", Name: "Laptop", Price: 999.99},
		{ID: "p2", Name: "Mouse", Price: 29.99},
		{ID: "p3", Name: "Keyboard", Price: 79.99},
	}, nil
}

func (s *LocalProductService) ValidateProducts(ids []string) ([]Product, error) {
	products, _ := s.GetProducts()
	var result []Product
	for _, id := range ids {
		for _, p := range products {
			if p.ID == id {
				result = append(result, p)
				break
			}
		}
	}
	return result, nil
}

func (s *LocalProductService) GetProductPrices(ids []string) (map[string]float64, error) {
	products, _ := s.ValidateProducts(ids)
	prices := make(map[string]float64)
	for _, p := range products {
		prices[p.ID] = p.Price
	}
	return prices, nil
}

// Local User Service
type LocalUserService struct{}

func (s *LocalUserService) GetUsers() ([]User, error) {
	return []User{
		{ID: "u1", Name: "John Doe", Email: "john@example.com"},
		{ID: "u2", Name: "Jane Smith", Email: "jane@example.com"},
	}, nil
}

func (s *LocalUserService) ValidateUser(userID string) (*User, error) {
	users, _ := s.GetUsers()
	for _, u := range users {
		if u.ID == userID {
			return &u, nil
		}
	}
	return nil, fmt.Errorf("user %s not found", userID)
}

// Local Payment Service
type LocalPaymentService struct{}

func (s *LocalPaymentService) ProcessPayment(req PaymentRequest) (*PaymentResponse, error) {
	gateway := lokstra_registry.GetConfigString("payment-gateway", "stripe")
	return &PaymentResponse{
		Status:        "success",
		TransactionID: fmt.Sprintf("txn-%s-%s", gateway, req.OrderID),
	}, nil
}

// Local Analytics Service
type LocalAnalyticsService struct{}

func (s *LocalAnalyticsService) TrackOrderCreated(order Order) error {
	fmt.Printf("📊 Analytics: Order %s created for user %s, total $%.2f\n",
		order.ID, order.UserID, order.Total)
	return nil
}

func (s *LocalAnalyticsService) GetAnalytics() (map[string]interface{}, error) {
	return map[string]interface{}{
		"total_orders":  156,
		"total_revenue": 12543.67,
		"active_users":  89,
		"last_updated":  "2024-10-01T22:53:00Z",
	}, nil
}

// HTTP Client Services (for microservices deployment)
type HTTPProductService struct {
	client *lokstra_registry.HTTPServiceClient
}

func createHTTPProductService(baseURL string) ProductService {
	return &HTTPProductService{
		client: lokstra_registry.NewHTTPServiceClient(baseURL),
	}
}

func (s *HTTPProductService) GetProducts() ([]Product, error) {
	var response struct {
		Products []Product `json:"products"`
	}

	if err := s.client.GET("/products", &response); err != nil {
		// Fallback to local service on error
		fmt.Printf("⚠️ HTTP call failed, using local fallback: %v\n", err)
		local := &LocalProductService{}
		return local.GetProducts()
	}

	return response.Products, nil
}

func (s *HTTPProductService) ValidateProducts(ids []string) ([]Product, error) {
	// For demo, use simple GET and filter locally
	// In real implementation, this would be POST /products/validate
	products, err := s.GetProducts()
	if err != nil {
		return nil, err
	}

	var result []Product
	for _, id := range ids {
		for _, p := range products {
			if p.ID == id {
				result = append(result, p)
				break
			}
		}
	}
	return result, nil
}

func (s *HTTPProductService) GetProductPrices(ids []string) (map[string]float64, error) {
	products, err := s.ValidateProducts(ids)
	if err != nil {
		return nil, err
	}

	prices := make(map[string]float64)
	for _, p := range products {
		prices[p.ID] = p.Price
	}
	return prices, nil
}

// HTTP User Service
type HTTPUserService struct {
	client *lokstra_registry.HTTPServiceClient
}

func createHTTPUserService(baseURL string) UserService {
	return &HTTPUserService{
		client: lokstra_registry.NewHTTPServiceClient(baseURL),
	}
}

func (s *HTTPUserService) GetUsers() ([]User, error) {
	var response struct {
		Users []User `json:"users"`
	}

	if err := s.client.GET("/users", &response); err != nil {
		fmt.Printf("⚠️ HTTP call failed, using local fallback: %v\n", err)
		local := &LocalUserService{}
		return local.GetUsers()
	}

	return response.Users, nil
}

func (s *HTTPUserService) ValidateUser(userID string) (*User, error) {
	users, err := s.GetUsers()
	if err != nil {
		return nil, err
	}

	for _, u := range users {
		if u.ID == userID {
			return &u, nil
		}
	}
	return nil, fmt.Errorf("user %s not found", userID)
}

// HTTP Payment Service
type HTTPPaymentService struct {
	client *lokstra_registry.HTTPServiceClient
}

func createHTTPPaymentService(baseURL string) PaymentService {
	return &HTTPPaymentService{
		client: lokstra_registry.NewHTTPServiceClient(baseURL),
	}
}

func (s *HTTPPaymentService) ProcessPayment(req PaymentRequest) (*PaymentResponse, error) {
	var response PaymentResponse

	if err := s.client.POST("/payments", req, &response); err != nil {
		fmt.Printf("⚠️ HTTP payment call failed, using local fallback: %v\n", err)
		local := &LocalPaymentService{}
		return local.ProcessPayment(req)
	}

	return &response, nil
}

// HTTP Analytics Service
type HTTPAnalyticsService struct {
	client *lokstra_registry.HTTPServiceClient
}

func createHTTPAnalyticsService(baseURL string) AnalyticsService {
	return &HTTPAnalyticsService{
		client: lokstra_registry.NewHTTPServiceClient(baseURL),
	}
}

func (s *HTTPAnalyticsService) TrackOrderCreated(order Order) error {
	// Fire-and-forget analytics call
	if err := s.client.POST("/analytics/orders", order, nil); err != nil {
		fmt.Printf("⚠️ Analytics tracking failed: %v\n", err)
		// Don't fail the main operation for analytics
	}
	return nil
}

func (s *HTTPAnalyticsService) GetAnalytics() (map[string]interface{}, error) {
	var response struct {
		Analytics map[string]interface{} `json:"analytics"`
	}

	if err := s.client.GET("/analytics", &response); err != nil {
		fmt.Printf("⚠️ HTTP analytics call failed, using local fallback: %v\n", err)
		local := &LocalAnalyticsService{}
		return local.GetAnalytics()
	}

	return response.Analytics, nil
}

// Helper Functions
func getServiceNameFromServer(serverName string) string {
	// Extract service name from server name
	switch serverName {
	case "product-service":
		return "product"
	case "order-service":
		return "order"
	case "user-service":
		return "user"
	case "payment-service":
		return "payment"
	case "analytics-service":
		return "analytics"
	default:
		return "unknown"
	}
}

// Middleware Factories
func createAuthMiddleware(config map[string]any) request.HandlerFunc {
	return func(ctx *request.Context) error {
		// Simple auth check
		token := ctx.R.Header.Get("Authorization")
		if token == "" {
			return ctx.Api.Error(http.StatusUnauthorized, "AUTH_REQUIRED", "Authorization required")
		}
		return ctx.Next()
	}
}

func createRateLimitMiddleware(config map[string]any) request.HandlerFunc {
	return func(ctx *request.Context) error {
		// Rate limiting logic here
		limit := lokstra_registry.GetConfigInt("rate-limit-rps", 100)
		fmt.Printf("🚦 Rate limit: %d requests per second\n", limit)
		return ctx.Next()
	}
}

func createCacheMiddleware(config map[string]any) request.HandlerFunc {
	return func(ctx *request.Context) error {
		ttl := lokstra_registry.GetConfigInt("cache-ttl-seconds", 300)
		fmt.Printf("💾 Cache TTL: %d seconds\n", ttl)
		return ctx.Next()
	}
}

func createLoggingMiddleware(config map[string]any) request.HandlerFunc {
	return func(ctx *request.Context) error {
		service := lokstra_registry.GetConfigString("service-name", "unknown")
		fmt.Printf("📝 [%s] %s %s\n", service, ctx.R.Method, ctx.R.URL.Path)
		return ctx.Next()
	}
}

func createCORSMiddleware(config map[string]any) request.HandlerFunc {
	return func(ctx *request.Context) error {
		origins := lokstra_registry.GetConfigString("cors-origins", "*")
		ctx.W.Header().Set("Access-Control-Allow-Origin", origins)
		ctx.W.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.W.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if ctx.R.Method == "OPTIONS" {
			ctx.W.WriteHeader(http.StatusOK)
			return nil
		}
		return ctx.Next()
	}
}

// Infrastructure Service Factories
func createDatabaseService(config map[string]any) any {
	dbURL := lokstra_registry.GetConfigString("database-url", "postgres://localhost:5432/ecommerce")
	poolSize := lokstra_registry.GetConfigInt("database-pool-size", 10)

	fmt.Printf("🗄️  Database: %s (pool: %d)\n", dbURL, poolSize)

	// Simulate database connection
	return map[string]any{
		"url":       dbURL,
		"pool_size": poolSize,
		"connected": true,
	}
}

func createRedisService(config map[string]any) any {
	redisURL := lokstra_registry.GetConfigString("redis-url", "redis://localhost:6379")

	fmt.Printf("🔴 Redis: %s\n", redisURL)

	// Simulate Redis connection
	return map[string]any{
		"url":       redisURL,
		"connected": true,
	}
}

func createEmailService(config map[string]any) any {
	provider := lokstra_registry.GetConfigString("email-provider", "sendgrid")

	fmt.Printf("📧 Email Provider: %s\n", provider)

	return map[string]any{
		"provider": provider,
		"ready":    true,
	}
}

// Business Service Factories (Key Innovation!)
// These factories return different implementations based on deployment type

func createProductService(config map[string]any) any {
	deploymentType := lokstra_registry.GetConfigString("deployment-type", "monolith")

	switch deploymentType {
	case "monolith":
		fmt.Printf("🏢 Product Service: Local implementation (monolith)\n")
		return &LocalProductService{}
	case "microservices":
		baseURL := lokstra_registry.GetConfigString("product-service-url", "http://localhost:8081")
		fmt.Printf("🔄 Product Service: HTTP client to %s (microservices)\n", baseURL)
		return createHTTPProductService(baseURL)
	default:
		return &LocalProductService{}
	}
}

func createUserService(config map[string]any) any {
	deploymentType := lokstra_registry.GetConfigString("deployment-type", "monolith")

	switch deploymentType {
	case "monolith":
		fmt.Printf("🏢 User Service: Local implementation (monolith)\n")
		return &LocalUserService{}
	case "microservices":
		baseURL := lokstra_registry.GetConfigString("user-service-url", "http://localhost:8083")
		fmt.Printf("🔄 User Service: HTTP client to %s (microservices)\n", baseURL)
		return createHTTPUserService(baseURL)
	default:
		return &LocalUserService{}
	}
}

func createPaymentService(config map[string]any) any {
	deploymentType := lokstra_registry.GetConfigString("deployment-type", "monolith")

	switch deploymentType {
	case "monolith":
		fmt.Printf("🏢 Payment Service: Local implementation (monolith)\n")
		return &LocalPaymentService{}
	case "microservices":
		baseURL := lokstra_registry.GetConfigString("payment-service-url", "http://localhost:8084")
		fmt.Printf("🔄 Payment Service: HTTP client to %s (microservices)\n", baseURL)
		return createHTTPPaymentService(baseURL)
	default:
		return &LocalPaymentService{}
	}
}

func createAnalyticsService(config map[string]any) any {
	deploymentType := lokstra_registry.GetConfigString("deployment-type", "monolith")

	switch deploymentType {
	case "monolith":
		fmt.Printf("🏢 Analytics Service: Local implementation (monolith)\n")
		return &LocalAnalyticsService{}
	case "microservices":
		baseURL := lokstra_registry.GetConfigString("analytics-service-url", "http://localhost:8085")
		fmt.Printf("🔄 Analytics Service: HTTP client to %s (microservices)\n", baseURL)
		return createHTTPAnalyticsService(baseURL)
	default:
		return &LocalAnalyticsService{}
	}
}

func main() {
	fmt.Println("🏪 Starting E-Commerce Application")

	// 1. Register all middleware types
	lokstra_registry.RegisterMiddlewareFactory("auth", createAuthMiddleware)
	lokstra_registry.RegisterMiddlewareFactory("rate-limit", createRateLimitMiddleware)
	lokstra_registry.RegisterMiddlewareFactory("cache", createCacheMiddleware)
	lokstra_registry.RegisterMiddlewareFactory("logging", createLoggingMiddleware)
	lokstra_registry.RegisterMiddlewareFactory("cors", createCORSMiddleware)

	// 2. Register all service types
	// Infrastructure services
	lokstra_registry.RegisterServiceFactory("database", createDatabaseService)
	lokstra_registry.RegisterServiceFactory("redis", createRedisService)
	lokstra_registry.RegisterServiceFactory("email", createEmailService)

	// ✨ NEW: Smart Service Integration (Framework-handled)
	// Define services with local implementations and HTTP client factories
	// Note: Using interface{} return type to handle different service interfaces
	lokstra_registry.DefineService("product-service", &LocalProductService{}, func(url string) interface{} {
		return createHTTPProductService(url)
	})
	lokstra_registry.DefineService("user-service", &LocalUserService{}, func(url string) interface{} {
		return createHTTPUserService(url)
	})
	lokstra_registry.DefineService("payment-service", &LocalPaymentService{}, func(url string) interface{} {
		return createHTTPPaymentService(url)
	})
	lokstra_registry.DefineService("analytics-service", &LocalAnalyticsService{}, func(url string) interface{} {
		return createHTTPAnalyticsService(url)
	})

	// Initialize Smart Service Integration (Auto-registers services based on deployment mode)
	lokstra_registry.ServiceIntegrationMiddleware()

	// OLD: Manual service factories (kept for comparison)
	lokstra_registry.RegisterServiceFactory("product-service-manual", createProductService)
	lokstra_registry.RegisterServiceFactory("user-service-manual", createUserService)
	lokstra_registry.RegisterServiceFactory("payment-service-manual", createPaymentService)
	lokstra_registry.RegisterServiceFactory("analytics-service-manual", createAnalyticsService)

	// 3. Create business logic routers (same in all deployments)
	productRouter := lokstra.NewRouter("product-api")
	productRouter.GET("/products", getProducts, route.WithNameOption("products-list"))
	lokstra_registry.RegisterRouter("product-api", productRouter)

	orderRouter := lokstra.NewRouter("order-api")
	orderRouter.POST("/orders", createOrder, route.WithNameOption("create-order"))
	lokstra_registry.RegisterRouter("order-api", orderRouter)

	userRouter := lokstra.NewRouter("user-api")
	userRouter.GET("/users", getUsers, route.WithNameOption("users-list"))
	lokstra_registry.RegisterRouter("user-api", userRouter)

	paymentRouter := lokstra.NewRouter("payment-api")
	paymentRouter.POST("/payments", processPayment, route.WithNameOption("process-payment"))
	lokstra_registry.RegisterRouter("payment-api", paymentRouter)

	analyticsRouter := lokstra.NewRouter("analytics-api")
	analyticsRouter.GET("/analytics", getAnalytics, route.WithNameOption("get-analytics"))
	lokstra_registry.RegisterRouter("analytics-api", analyticsRouter)

	// 4. Register all possible servers (will be determined by config which one to start)

	// Monolith server - single server with all services
	monolithApp := app.New("ecommerce-app", "/api")
	monolithServer := server.New("monolith-server", monolithApp)
	lokstra_registry.RegisterServer("monolith-server", monolithServer)

	// Microservices servers - individual servers per service
	productApp := app.New("product-app", "/")
	productServer := server.New("product-service", productApp)
	lokstra_registry.RegisterServer("product-service", productServer)

	orderApp := app.New("order-app", "/")
	orderServer := server.New("order-service", orderApp)
	lokstra_registry.RegisterServer("order-service", orderServer)

	userApp := app.New("user-app", "/")
	userServer := server.New("user-service", userApp)
	lokstra_registry.RegisterServer("user-service", userServer)

	paymentApp := app.New("payment-app", "/")
	paymentServer := server.New("payment-service", paymentApp)
	lokstra_registry.RegisterServer("payment-service", paymentServer)

	analyticsApp := app.New("analytics-app", "/")
	analyticsServer := server.New("analytics-service", analyticsApp)
	lokstra_registry.RegisterServer("analytics-service", analyticsServer)

	// 5. Load deployment configuration
	deploymentType := os.Getenv("DEPLOYMENT_TYPE")
	if deploymentType == "" {
		deploymentType = "monolith"
	}

	// 6. Determine which server to start and config to load
	serverName := os.Getenv("SERVER_NAME")
	var configPath string

	if deploymentType == "microservices" {
		if serverName == "" {
			log.Fatal("❌ SERVER_NAME environment variable required for microservices deployment")
		}
		configPath = fmt.Sprintf("config-microservices-%s.yaml", getServiceNameFromServer(serverName))
	} else {
		// Monolith deployment
		if serverName == "" {
			serverName = "monolith-server"
		}
		configPath = fmt.Sprintf("config-%s.yaml", deploymentType)
	}

	var cfg config.Config
	if err := config.LoadConfigFile(configPath, &cfg); err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	fmt.Printf("🚀 Deployment Type: %s\n", deploymentType)

	fmt.Printf("🎯 Starting Server: %s\n", serverName)

	// 7. Apply configuration and start server
	if err := config.ApplyAllConfig(&cfg, serverName); err != nil {
		log.Fatalf("❌ Failed to apply config: %v", err)
	}

	// Start the specified server
	server := lokstra_registry.GetServer(serverName)
	if server == nil {
		log.Fatalf("❌ Server %s not found", serverName)
	}

	fmt.Printf("✅ Server %s started successfully!\n", serverName)
	fmt.Println("🌍 Available endpoints:")

	// Show available endpoints based on deployment
	if deploymentType == "monolith" {
		fmt.Println("   GET  /api/products")
		fmt.Println("   POST /api/orders")
		fmt.Println("   GET  /api/users")
		fmt.Println("   POST /api/payments")
		fmt.Println("   GET  /api/analytics")
	} else {
		// Show service-specific endpoints
		switch serverName {
		case "product-service":
			fmt.Println("   GET  /products")
		case "order-service":
			fmt.Println("   POST /orders")
		case "user-service":
			fmt.Println("   GET  /users")
		case "payment-service":
			fmt.Println("   POST /payments")
		case "analytics-service":
			fmt.Println("   GET  /analytics")
		}
	}

	// Simulate server running (in real app, would call server.Start())
	fmt.Println("\n🎉 Application ready! (Press Ctrl+C to stop)")
	fmt.Println("Note: In production, call server.Start() to actually start HTTP server")
}
