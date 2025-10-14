package main

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/config"
	"github.com/primadi/lokstra/lokstra_registry"
)

// =============================================================================
// COMPLETE CONFIG-DRIVEN APPLICATION
// =============================================================================
// This example shows a complete e-commerce application where:
// - Services are defined in config.yaml
// - Routers are registered in code (setup_routers.go)
// - Server configuration is in config.yaml
// - main.go is minimal - just loads config and starts server
//
// Architecture:
// 1. Register factories (code) - how to create services
// 2. Register routers (code) - HTTP endpoints
// 3. Define services (YAML) - service instances with config
// 4. Define server (YAML) - deployment topology
// 5. Start server - framework assembles everything
//
// IMPORTANT: This example uses CORRECT lazy loading pattern:
// ✅ Factories store service NAMES from config (not instances)
// ✅ Lazy getter methods call GetService when needed (not in factory)
// ✅ Cache variables in struct fields for performance
// ✅ No registration order dependency
// ✅ Explicit dependencies in config.yaml
//
// See: 03-service-dependencies for detailed explanation of this pattern

func main() {
	fmt.Println("📚 Lokstra Learning: 04-Config-Driven Deployment")
	fmt.Println("===================================================")
	fmt.Println()

	// Step 1: Setup (register factories and routers)
	fmt.Println("🔧 Setting up application...")
	setupFactories()
	setupRouters()
	fmt.Println()

	// Step 2: Load configuration
	fmt.Println("📄 Loading configuration...")
	cfg := config.New()

	if err := config.LoadConfigFile("config.yaml", cfg); err != nil {
		fmt.Printf("❌ Failed to load config.yaml: %v\n", err)
		return
	}
	fmt.Println("   ✓ Loaded config.yaml")
	fmt.Println()

	// Step 3: Register config with framework
	lokstra_registry.RegisterConfig(cfg, "")

	// Print application info
	fmt.Println("📋 Application Info:")
	fmt.Printf("   Name: E-Commerce API\n")
	fmt.Printf("   Version: %s\n", lokstra_registry.GetConfig("app-version", "unknown"))
	fmt.Printf("   Environment: %s\n", lokstra_registry.GetConfig("app-env", "unknown"))
	fmt.Printf("   Server: %s\n", lokstra_registry.GetCurrentServerName())
	fmt.Println()

	// Step 4: Start server (framework does the rest)
	fmt.Println("🚀 Starting server...")
	lokstra_registry.PrintServerStartInfo()
	if err := lokstra_registry.StartServer(); err != nil {
		fmt.Printf("❌ Server error: %v\n", err)
	}
}

// =============================================================================
// INFRASTRUCTURE SERVICES (Layer 1)
// =============================================================================

type DBService struct {
	Host     string
	Port     int
	Database string
	Username string
}

func (s *DBService) Query(sql string) []map[string]any {
	fmt.Printf("💾 [DB] Query: %s\n", sql)
	return []map[string]any{}
}

func (s *DBService) Execute(sql string, args ...any) int64 {
	fmt.Printf("💾 [DB] Execute: %s (args: %v)\n", sql, args)
	return 1
}

func NewDBService(cfg map[string]any) any {
	fmt.Println("   🏗️  Creating DBService")
	return &DBService{
		Host:     utils.GetValueFromMap(cfg, "host", "localhost"),
		Port:     utils.GetValueFromMap(cfg, "port", 5432),
		Database: utils.GetValueFromMap(cfg, "database", "myapp"),
		Username: utils.GetValueFromMap(cfg, "username", "postgres"),
	}
}

type CacheService struct {
	Host string
	Port int
}

func (s *CacheService) Get(key string) (any, bool) {
	fmt.Printf("🗄️  [Cache] GET %s\n", key)
	return nil, false
}

func (s *CacheService) Set(key string, value any, ttl time.Duration) {
	fmt.Printf("🗄️  [Cache] SET %s (TTL: %v)\n", key, ttl)
}

func NewCacheService(cfg map[string]any) any {
	fmt.Println("   🏗️  Creating CacheService")
	return &CacheService{
		Host: utils.GetValueFromMap(cfg, "host", "localhost"),
		Port: utils.GetValueFromMap(cfg, "port", 6379),
	}
}

type EmailService struct {
	SMTPHost string
	SMTPPort int
	From     string
}

func (s *EmailService) Send(to, subject, body string) error {
	fmt.Printf("📧 [Email] To: %s, Subject: %s\n", to, subject)
	return nil
}

func NewEmailService(cfg map[string]any) any {
	fmt.Println("   🏗️  Creating EmailService")
	return &EmailService{
		SMTPHost: utils.GetValueFromMap(cfg, "smtp_host", "localhost"),
		SMTPPort: utils.GetValueFromMap(cfg, "smtp_port", 587),
		From:     utils.GetValueFromMap(cfg, "from", "noreply@example.com"),
	}
}

// =============================================================================
// REPOSITORIES (Layer 2)
// =============================================================================

type UserRepository struct {
	dbServiceName string
	dbCache       *DBService
}

func (r *UserRepository) getDB() *DBService {
	r.dbCache = lokstra_registry.GetServiceCached(r.dbServiceName, r.dbCache)
	return r.dbCache
}

func (r *UserRepository) FindByID(id string) map[string]any {
	db := r.getDB()
	db.Query("SELECT * FROM users WHERE id = " + id)
	return map[string]any{"id": id, "name": "User " + id, "email": "user" + id + "@example.com"}
}

func (r *UserRepository) Create(name, email, password string) map[string]any {
	db := r.getDB()
	db.Execute("INSERT INTO users (name, email, password) VALUES (?, ?, ?)", name, email, password)
	return map[string]any{"id": "new-user-id", "name": name, "email": email}
}

func NewUserRepository(cfg map[string]any) any {
	fmt.Println("   🏗️  Creating UserRepository")
	return &UserRepository{
		dbServiceName: utils.GetValueFromMap(cfg, "db_service", "db-service"),
	}
}

type ProductRepository struct {
	dbServiceName    string
	cacheServiceName string
	dbCache          *DBService
	cacheCache       *CacheService
}

func (r *ProductRepository) getDB() *DBService {
	r.dbCache = lokstra_registry.GetServiceCached(r.dbServiceName, r.dbCache)
	return r.dbCache
}

func (r *ProductRepository) getCache() *CacheService {
	r.cacheCache = lokstra_registry.GetServiceCached(r.cacheServiceName, r.cacheCache)
	return r.cacheCache
}

var productList = []map[string]any{
	{"id": "1", "name": "Laptop", "price": 1299.99},
	{"id": "2", "name": "Mouse", "price": 5.00},
	{"id": "3", "name": "Keyboard", "price": 79.99},
}

func (r *ProductRepository) FindByID(id string) map[string]any {
	cache := r.getCache()
	cacheKey := "product:" + id
	if cached, ok := cache.Get(cacheKey); ok {
		return cached.(map[string]any)
	}

	db := r.getDB()
	db.Query("SELECT * FROM products WHERE id = " + id)
	for _, product := range productList {
		if product["id"] == id {
			cache.Set(cacheKey, product, 5*time.Minute)
			return product
		}
	}
	return nil
}

func (r *ProductRepository) List() []map[string]any {
	db := r.getDB()
	db.Query("SELECT * FROM products")
	return productList
}

func NewProductRepository(cfg map[string]any) any {
	fmt.Println("   🏗️  Creating ProductRepository")
	return &ProductRepository{
		dbServiceName:    utils.GetValueFromMap(cfg, "db_service", "db-service"),
		cacheServiceName: utils.GetValueFromMap(cfg, "cache_service", "cache-service"),
	}
}

type OrderRepository struct {
	dbServiceName string
	dbCache       *DBService
}

func (r *OrderRepository) getDB() *DBService {
	r.dbCache = lokstra_registry.GetServiceCached(r.dbServiceName, r.dbCache)
	return r.dbCache
}

func (r *OrderRepository) Create(userID, productID string, quantity int, amount float64) map[string]any {
	db := r.getDB()
	db.Execute("INSERT INTO orders (user_id, product_id, quantity, amount) VALUES (?, ?, ?, ?)",
		userID, productID, quantity, amount)
	return map[string]any{
		"id":         "order-" + fmt.Sprint(time.Now().Unix()),
		"user_id":    userID,
		"product_id": productID,
		"quantity":   quantity,
		"amount":     amount,
	}
}

func NewOrderRepository(cfg map[string]any) any {
	fmt.Println("   🏗️  Creating OrderRepository")
	return &OrderRepository{
		dbServiceName: utils.GetValueFromMap(cfg, "db_service", "db-service"),
	}
}

// =============================================================================
// DOMAIN SERVICES (Layer 3)
// =============================================================================

type UserService struct {
	repoServiceName   string
	repoCache         *UserRepository
	passwordMinLength int
}

func (s *UserService) getRepo() *UserRepository {
	s.repoCache = lokstra_registry.GetServiceCached(s.repoServiceName, s.repoCache)
	return s.repoCache
}

func (s *UserService) GetUser(id string) map[string]any {
	repo := s.getRepo()
	return repo.FindByID(id)
}

func (s *UserService) Register(name, email, password string) (map[string]any, error) {
	if len(password) < s.passwordMinLength {
		return nil, fmt.Errorf("password must be at least %d characters", s.passwordMinLength)
	}
	repo := s.getRepo()
	return repo.Create(name, email, password), nil
}

func NewUserService(cfg map[string]any) any {
	fmt.Println("   🏗️  Creating UserService")
	return &UserService{
		repoServiceName:   utils.GetValueFromMap(cfg, "repository_service", "user-repository"),
		passwordMinLength: utils.GetValueFromMap(cfg, "password_min_length", 8),
	}
}

type ProductService struct {
	repoServiceName string
	repoCache       *ProductRepository
	defaultCurrency string
}

func (s *ProductService) getRepo() *ProductRepository {
	s.repoCache = lokstra_registry.GetServiceCached(s.repoServiceName, s.repoCache)
	return s.repoCache
}

func (s *ProductService) GetProduct(id string) map[string]any {
	repo := s.getRepo()
	product := repo.FindByID(id)
	product["currency"] = s.defaultCurrency
	return product
}

func (s *ProductService) ListProducts() []map[string]any {
	repo := s.getRepo()
	products := repo.List()
	for _, p := range products {
		p["currency"] = s.defaultCurrency
	}
	return products
}

func NewProductService(cfg map[string]any) any {
	fmt.Println("   🏗️  Creating ProductService")
	return &ProductService{
		repoServiceName: utils.GetValueFromMap(cfg, "repository_service", "product-repository"),
		defaultCurrency: utils.GetValueFromMap(cfg, "default_currency", "USD"),
	}
}

type OrderService struct {
	repoServiceName     string
	productServiceName  string
	userServiceName     string
	emailServiceName    string
	repoCache           *OrderRepository
	productServiceCache *ProductService
	userServiceCache    *UserService
	emailServiceCache   *EmailService
	taxRate             float64
	minOrderAmount      float64
}

func (s *OrderService) getRepo() *OrderRepository {
	s.repoCache = lokstra_registry.GetServiceCached(s.repoServiceName, s.repoCache)
	return s.repoCache
}

func (s *OrderService) getProductService() *ProductService {
	s.productServiceCache = lokstra_registry.GetServiceCached(s.productServiceName, s.productServiceCache)
	return s.productServiceCache
}

func (s *OrderService) getUserService() *UserService {
	s.userServiceCache = lokstra_registry.GetServiceCached(s.userServiceName, s.userServiceCache)
	return s.userServiceCache
}

func (s *OrderService) getEmailService() *EmailService {
	s.emailServiceCache = lokstra_registry.GetServiceCached(s.emailServiceName, s.emailServiceCache)
	return s.emailServiceCache
}

func (s *OrderService) CreateOrder(userID, productID string, quantity int) (map[string]any, error) {
	// Get product (lazy loaded)
	productService := s.getProductService()
	product := productService.GetProduct(productID)
	price := product["price"].(float64)

	// Calculate total
	subtotal := price * float64(quantity)
	if subtotal < s.minOrderAmount {
		return nil, fmt.Errorf("minimum order amount is %.2f", s.minOrderAmount)
	}

	tax := subtotal * s.taxRate
	total := subtotal + tax

	// Create order (lazy loaded)
	repo := s.getRepo()
	order := repo.Create(userID, productID, quantity, total)
	order["subtotal"] = subtotal
	order["tax"] = tax
	order["total"] = total

	// Get user and send confirmation (lazy loaded)
	userService := s.getUserService()
	user := userService.GetUser(userID)

	emailService := s.getEmailService()
	emailService.Send(
		user["email"].(string),
		"Order Confirmation",
		fmt.Sprintf("Your order #%s has been placed. Total: $%.2f", order["id"], total),
	)

	return order, nil
}

func NewOrderService(cfg map[string]any) any {
	fmt.Println("   🏗️  Creating OrderService")
	return &OrderService{
		repoServiceName:    utils.GetValueFromMap(cfg, "repository_service", "order-repository"),
		productServiceName: utils.GetValueFromMap(cfg, "product_service", "product-service"),
		userServiceName:    utils.GetValueFromMap(cfg, "user_service", "user-service"),
		emailServiceName:   utils.GetValueFromMap(cfg, "email_service", "email-service"),
		taxRate:            utils.GetValueFromMap(cfg, "tax_rate", 0.10),
		minOrderAmount:     utils.GetValueFromMap(cfg, "min_order_amount", 10.0),
	}
}

// =============================================================================
// SERVICE CONTAINER
// =============================================================================

type ServiceContainer struct {
	userCache    *UserService
	productCache *ProductService
	orderCache   *OrderService
}

func (sc *ServiceContainer) GetUser() *UserService {
	sc.userCache = lokstra_registry.GetServiceCached("user-service", sc.userCache)
	return sc.userCache
}

func (sc *ServiceContainer) GetProduct() *ProductService {
	sc.productCache = lokstra_registry.GetServiceCached("product-service", sc.productCache)
	return sc.productCache
}

func (sc *ServiceContainer) GetOrder() *OrderService {
	sc.orderCache = lokstra_registry.GetServiceCached("order-service", sc.orderCache)
	return sc.orderCache
}

var services = &ServiceContainer{}

// =============================================================================
// SETUP FACTORIES
// =============================================================================

func setupFactories() {
	// Infrastructure
	lokstra_registry.RegisterServiceFactory("db", NewDBService)
	lokstra_registry.RegisterServiceFactory("cache", NewCacheService)
	lokstra_registry.RegisterServiceFactory("email", NewEmailService)

	// Repositories
	lokstra_registry.RegisterServiceFactory("user-repo", NewUserRepository)
	lokstra_registry.RegisterServiceFactory("product-repo", NewProductRepository)
	lokstra_registry.RegisterServiceFactory("order-repo", NewOrderRepository)

	// Domain Services
	lokstra_registry.RegisterServiceFactory("user", NewUserService)
	lokstra_registry.RegisterServiceFactory("product", NewProductService)
	lokstra_registry.RegisterServiceFactory("order", NewOrderService)

	fmt.Println("   ✓ Registered 9 service factories")
}

// =============================================================================
// SETUP ROUTERS
// =============================================================================

func setupRouters() {
	// Product API
	productRouter := lokstra.NewRouter("product-router")
	productRouter.SetPathPrefix("/api/products")
	productRouter.GET("/", func(c *lokstra.RequestContext) error {
		products := services.GetProduct().ListProducts()
		return c.Api.Ok(products)
	})

	productRouter.GET("/{id}", func(c *lokstra.RequestContext) error {
		id := c.Req.PathParam("id", "")
		product := services.GetProduct().GetProduct(id)
		return c.Api.Ok(product)
	})

	// Order API
	orderRouter := lokstra.NewRouter("order-router")

	type createOrderRequest struct {
		UserID    string `json:"user_id" binding:"required"`
		ProductID string `json:"product_id" binding:"required"`
		Quantity  int    `json:"quantity" binding:"required,gt=0"`
	}

	orderRouter.POST("/api/orders", func(c *lokstra.RequestContext,
		req *createOrderRequest) error {
		order, err := services.GetOrder().CreateOrder(req.UserID, req.ProductID, req.Quantity)
		if err != nil {
			return c.Api.BadRequest("ORDER_ERROR", err.Error())
		}

		return c.Api.Created(order, "Order created successfully")
	})

	// User API
	userRouter := lokstra.NewRouter("user-router")

	userRouter.GET("/api/users/{id}", func(c *lokstra.RequestContext) error {
		id := c.Req.PathParam("id", "")
		user := services.GetUser().GetUser(id)
		return c.Api.Ok(user)
	})

	userRouter.POST("/api/users/register", func(c *lokstra.RequestContext) error {
		var req struct {
			Name     string `json:"name" binding:"required"`
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.Req.BindBody(&req); err != nil {
			return c.Api.BadRequest("INVALID_INPUT", err.Error())
		}

		user, err := services.GetUser().Register(req.Name, req.Email, req.Password)
		if err != nil {
			return c.Api.BadRequest("REGISTRATION_ERROR", err.Error())
		}

		return c.Api.Created(user, "User registered successfully")
	})

	// Health API
	healthRouter := lokstra.NewRouter("health-router")

	healthRouter.GET("/health", func(c *lokstra.RequestContext) error {
		appEnv := lokstra_registry.GetConfig("app-env", "unknown")
		appVersion := lokstra_registry.GetConfig("app-version", "unknown")

		return c.Api.Ok(map[string]any{
			"status":      "healthy",
			"environment": appEnv,
			"version":     appVersion,
			"time":        time.Now().Format(time.RFC3339),
		})
	})

	// Register routers
	lokstra_registry.RegisterRouter("product-api", productRouter)
	lokstra_registry.RegisterRouter("order-api", orderRouter)
	lokstra_registry.RegisterRouter("user-api", userRouter)
	lokstra_registry.RegisterRouter("health-api", healthRouter)

	fmt.Println("   ✓ Registered 4 routers")
}
