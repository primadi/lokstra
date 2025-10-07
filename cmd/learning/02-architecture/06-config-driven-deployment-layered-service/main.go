package main

import (
	"fmt"
	"os"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/config"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/lokstra_registry"
)

// =============================================================================
// LAYERED SERVICES COMPARISON
// =============================================================================
// This example compares TWO patterns for service configuration:
//
// 1. SIMPLE SERVICES (config-simple.yaml)
//    - Flat array format
//    - Manual lazy loading in factories
//    - ~15 lines of boilerplate per dependency
//
// 2. LAYERED SERVICES (config-layered.yaml)
//    - Grouped by architectural layers
//    - Generic Lazy[T] pattern
//    - ~3 lines per dependency (80% less!)
//    - Explicit depends-on with validation
//
// Run with:
//   go run . simple   → Uses config-simple.yaml
//   go run . layered  → Uses config-layered.yaml

func main() {
	fmt.Println("📚 Lokstra Learning: 06-Layered Services Comparison")
	fmt.Println("=====================================================")
	fmt.Println()

	// Determine which config to use
	configFile := "config-layered.yaml" // default
	mode := "layered"

	if len(os.Args) > 1 {
		arg := os.Args[1]
		switch arg {
		case "simple":
			configFile = "config-simple.yaml"
			mode = "simple"
		case "layered":
			configFile = "config-layered.yaml"
			mode = "layered"
		default:
			// Assume it's a filename
			configFile = arg
			if configFile == "config-simple.yaml" {
				mode = "simple"
			}
		}
	}

	fmt.Printf("🎯 Mode: %s\n", mode)
	fmt.Printf("📄 Config: %s\n", configFile)
	fmt.Println()

	// Step 1: Setup (register factories and routers)
	fmt.Println("🔧 Setting up application...")
	setupFactories()
	setupRouters()
	fmt.Println()

	// Step 2: Load configuration
	fmt.Println("📄 Loading configuration...")
	cfg := config.New()

	if err := config.LoadConfigFile(configFile, cfg); err != nil {
		fmt.Printf("❌ Failed to load %s: %v\n", configFile, err)
		return
	}
	fmt.Printf("   ✓ Loaded %s\n", configFile)

	// Show config mode
	if cfg.Services.IsSimple() {
		fmt.Println("   📋 Service Mode: SIMPLE (flat array)")
		fmt.Printf("      Total services: %d\n", len(cfg.Services.Simple))
	} else if cfg.Services.IsLayered() {
		fmt.Println("   📋 Service Mode: LAYERED (grouped by layer)")
		fmt.Printf("      Total layers: %d\n", len(cfg.Services.Order))
		for _, layerName := range cfg.Services.Order {
			fmt.Printf("      - %s: %d services\n", layerName, len(cfg.Services.Layered[layerName]))
		}

		// Validate layered services
		if err := config.ValidateLayeredServices(&cfg.Services); err != nil {
			fmt.Printf("   ❌ Validation failed: %v\n", err)
			return
		}
		fmt.Println("   ✅ Validation passed!")
	}
	fmt.Println()

	// Step 3: Register config with framework
	lokstra_registry.RegisterConfig(cfg)

	// Step 4: Set current server
	serverName := lokstra_registry.GetConfig("server-name", "ecommerce-server")
	lokstra_registry.SetCurrentServerName(serverName)

	// Print application info
	fmt.Println("📋 Application Info:")
	fmt.Printf("   Name: E-Commerce API\n")
	fmt.Printf("   Version: %s\n", lokstra_registry.GetConfig("app-version", "unknown"))
	fmt.Printf("   Environment: %s\n", lokstra_registry.GetConfig("app-env", "unknown"))
	fmt.Printf("   Server: %s\n", serverName)
	fmt.Println()

	// Print pattern comparison
	printPatternComparison(mode)

	// Step 5: Start server
	fmt.Println("🚀 Starting server...")
	lokstra_registry.PrintServerStartInfo()
	lokstra_registry.StartServer()
}

func printPatternComparison(mode string) {
	if mode == "simple" {
		fmt.Println("📊 Simple Services Pattern:")
		fmt.Println("   ✅ Backward compatible")
		fmt.Println("   ✅ Familiar pattern")
		fmt.Println("   ❌ ~15 lines boilerplate per dependency")
		fmt.Println("   ❌ Manual lazy loading + caching")
		fmt.Println("   ❌ Dependencies hidden in code")
		fmt.Println("   ❌ No validation")
	} else {
		fmt.Println("📊 Layered Services Pattern:")
		fmt.Println("   ✅ Type-safe with Lazy[T]")
		fmt.Println("   ✅ ~3 lines per dependency (80% less!)")
		fmt.Println("   ✅ Explicit depends-on")
		fmt.Println("   ✅ Automatic validation")
		fmt.Println("   ✅ Architecture visible in config")
		fmt.Println("   ✅ Auto-caching with sync.Once")
	}
	fmt.Println()
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
// REPOSITORIES (Layer 2) - Using Generic Lazy[T]
// =============================================================================

type UserRepository struct {
	db *service.Lazy[DBService] // ✨ Generic lazy - type safe, auto-caching
}

func (r *UserRepository) FindByID(id string) map[string]any {
	db := r.db.Get() // ✨ One line - lazy load + cache
	db.Query("SELECT * FROM users WHERE id = " + id)
	return map[string]any{"id": id, "name": "User " + id, "email": "user" + id + "@example.com"}
}

func (r *UserRepository) Create(name, email, password string) map[string]any {
	db := r.db.Get()
	db.Execute("INSERT INTO users (name, email, password) VALUES (?, ?, ?)", name, email, password)
	return map[string]any{"id": "new-user-id", "name": name, "email": email}
}

func NewUserRepository(cfg map[string]any) any {
	fmt.Println("   🏗️  Creating UserRepository (Lazy[T] pattern)")
	return &UserRepository{
		db: service.GetLazyService[DBService](cfg, "db_service"), // ✨ 1 line vs 15!
	}
}

type ProductRepository struct {
	db    *service.Lazy[DBService]
	cache *service.Lazy[CacheService]
}

var productList = []map[string]any{
	{"id": "1", "name": "Laptop", "price": 1299.99},
	{"id": "2", "name": "Mouse", "price": 5.00},
	{"id": "3", "name": "Keyboard", "price": 79.99},
}

func (r *ProductRepository) FindByID(id string) map[string]any {
	cache := r.cache.Get()
	cacheKey := "product:" + id
	if cached, ok := cache.Get(cacheKey); ok {
		return cached.(map[string]any)
	}

	db := r.db.Get()
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
	db := r.db.Get()
	db.Query("SELECT * FROM products")
	return productList
}

func NewProductRepository(cfg map[string]any) any {
	fmt.Println("   🏗️  Creating ProductRepository (Lazy[T] pattern)")
	return &ProductRepository{
		db:    service.GetLazyService[DBService](cfg, "db_service"),       // ✨ Type-safe
		cache: service.GetLazyService[CacheService](cfg, "cache_service"), // ✨ No casts
	}
}

type OrderRepository struct {
	db *service.Lazy[DBService]
}

func (r *OrderRepository) Create(userID, productID string, quantity int, amount float64) map[string]any {
	db := r.db.Get()
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
	fmt.Println("   🏗️  Creating OrderRepository (Lazy[T] pattern)")
	return &OrderRepository{
		db: service.GetLazyService[DBService](cfg, "db_service"),
	}
}

// =============================================================================
// DOMAIN SERVICES (Layer 3) - Using Generic Lazy[T]
// =============================================================================

type UserService struct {
	repo              *service.Lazy[UserRepository] // ✨ Generic lazy container
	passwordMinLength int
}

func (s *UserService) GetUser(id string) map[string]any {
	repo := s.repo.Get() // ✨ Type-safe, no casts needed
	return repo.FindByID(id)
}

func (s *UserService) Register(name, email, password string) (map[string]any, error) {
	if len(password) < s.passwordMinLength {
		return nil, fmt.Errorf("password must be at least %d characters", s.passwordMinLength)
	}
	repo := s.repo.Get()
	return repo.Create(name, email, password), nil
}

func NewUserService(cfg map[string]any) any {
	fmt.Println("   🏗️  Creating UserService (Lazy[T] pattern)")
	return &UserService{
		repo:              service.GetLazyService[UserRepository](cfg, "repository_service"), // ✨ Clean!
		passwordMinLength: utils.GetValueFromMap(cfg, "password_min_length", 8),
	}
}

type ProductService struct {
	repo            *service.Lazy[ProductRepository]
	defaultCurrency string
}

func (s *ProductService) GetProduct(id string) map[string]any {
	repo := s.repo.Get()
	product := repo.FindByID(id)
	product["currency"] = s.defaultCurrency
	return product
}

func (s *ProductService) ListProducts() []map[string]any {
	repo := s.repo.Get()
	products := repo.List()
	for _, p := range products {
		p["currency"] = s.defaultCurrency
	}
	return products
}

func NewProductService(cfg map[string]any) any {
	fmt.Println("   🏗️  Creating ProductService (Lazy[T] pattern)")
	return &ProductService{
		repo:            service.GetLazyService[ProductRepository](cfg, "repository_service"),
		defaultCurrency: utils.GetValueFromMap(cfg, "default_currency", "USD"),
	}
}

type OrderService struct {
	repo           *service.Lazy[OrderRepository] // ✨ 4 dependencies
	product        *service.Lazy[ProductService]  // ✨ Only 4 lines of code!
	user           *service.Lazy[UserService]     // ✨ vs ~60 lines in old pattern
	email          *service.Lazy[EmailService]    // ✨ 93% reduction in boilerplate
	taxRate        float64
	minOrderAmount float64
}

func (s *OrderService) CreateOrder(userID, productID string, quantity int) (map[string]any, error) {
	// ✨ Direct access - type-safe, lazy loaded, auto-cached
	product := s.product.Get().GetProduct(productID)
	price := product["price"].(float64)

	// Calculate total
	subtotal := price * float64(quantity)
	if subtotal < s.minOrderAmount {
		return nil, fmt.Errorf("minimum order amount is %.2f", s.minOrderAmount)
	}

	tax := subtotal * s.taxRate
	total := subtotal + tax

	// Create order
	order := s.repo.Get().Create(userID, productID, quantity, total)
	order["subtotal"] = subtotal
	order["tax"] = tax
	order["total"] = total

	// Send confirmation email
	user := s.user.Get().GetUser(userID)
	s.email.Get().Send(
		user["email"].(string),
		"Order Confirmation",
		fmt.Sprintf("Your order #%s has been placed. Total: $%.2f", order["id"], total),
	)

	return order, nil
}

func NewOrderService(cfg map[string]any) any {
	fmt.Println("   🏗️  Creating OrderService (Lazy[T] pattern)")
	// ✨ Compare this to 60+ lines in simple pattern!
	return &OrderService{
		repo:           service.GetLazyService[OrderRepository](cfg, "repository_service"),
		product:        service.GetLazyService[ProductService](cfg, "product_service"),
		user:           service.GetLazyService[UserService](cfg, "user_service"),
		email:          service.GetLazyService[EmailService](cfg, "email_service"),
		taxRate:        utils.GetValueFromMap(cfg, "tax_rate", 0.10),
		minOrderAmount: utils.GetValueFromMap(cfg, "min_order_amount", 10.0),
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
	sc.userCache = lokstra_registry.GetService("user-service", sc.userCache)
	return sc.userCache
}

func (sc *ServiceContainer) GetProduct() *ProductService {
	sc.productCache = lokstra_registry.GetService("product-service", sc.productCache)
	return sc.productCache
}

func (sc *ServiceContainer) GetOrder() *OrderService {
	sc.orderCache = lokstra_registry.GetService("order-service", sc.orderCache)
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
