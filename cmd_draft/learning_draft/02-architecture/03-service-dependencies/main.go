package main

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/config"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/old_registry"
)

// =============================================================================
// SERVICE DEPENDENCIES - BEST PRACTICES
// =============================================================================
// This example demonstrates the CORRECT way to handle service dependencies.
//
// ❌ WRONG PATTERN (Don't do this!):
//    func ServiceFactory(cfg map[string]any) any {
//        var dep *DepService
//        dep = old_registry.GetService("dep", dep)  // BAD! Called at registration
//        return NewService(dep)
//    }
//
// Problems:
// 1. GetService called during REGISTRATION phase (not lazy!)
// 2. Forces dependencies to exist before this service
// 3. No lazy loading benefit
// 4. No caching benefit
// 5. Tight coupling in factory
//
// ✅ CORRECT PATTERN (Pattern 2 - Explicit Lazy Dependencies):
//
// 1. Store service NAMES in struct (not instances):
//    type MyService struct {
//        depServiceName string  // Store name, not instance
//        depCache *DepService   // Cache variable (nil until first use)
//    }
//
// 2. Create lazy getter methods:
//    func (s *MyService) getDep() *DepService {
//        s.depCache = old_registry.GetService(s.depServiceName, s.depCache)
//        return s.depCache
//    }
//
// 3. Use lazy getters in business methods:
//    func (s *MyService) DoWork() {
//        dep := s.getDep()  // Lazy load on first call, cached after
//        dep.Query()
//    }
//
// 4. Factory just stores names from config:
//    func ServiceFactory(cfg map[string]any) any {
//        return &MyService{
//            depServiceName: utils.GetValueFromMap(cfg, "dep_service", "dep-service"),
//        }
//    }
//
// Benefits:
// ✅ True lazy loading (only created when needed)
// ✅ Proper caching (GetService pattern works correctly)
// ✅ No registration order dependency
// ✅ Flexible service naming (from config)
// ✅ Easy testing (override service names in config)
// ✅ Clear dependencies (visible in YAML config)

// =============================================================================
// LAYER 1: INFRASTRUCTURE SERVICES (No dependencies)
// =============================================================================

// DBService - Database connection
type DBService struct {
	Host     string
	Port     int
	Database string
}

func (s *DBService) Query(sql string) string {
	fmt.Printf("💾 [DB:%s] Query: %s\n", s.Database, sql)
	return fmt.Sprintf("result from %s", s.Database)
}

func (s *DBService) Execute(sql string) bool {
	fmt.Printf("💾 [DB:%s] Execute: %s\n", s.Database, sql)
	return true
}

func NewDBService(host string, port int, database string) *DBService {
	fmt.Printf("   🏗️  Creating DBService: %s:%d/%s\n", host, port, database)
	return &DBService{Host: host, Port: port, Database: database}
}

func DBServiceFactory(cfg map[string]any) any {
	return NewDBService(
		utils.GetValueFromMap(cfg, "host", "localhost"),
		utils.GetValueFromMap(cfg, "port", 5432),
		utils.GetValueFromMap(cfg, "database", "myapp"),
	)
}

// CacheService - Redis cache
type CacheService struct {
	Host string
	Port int
}

func (s *CacheService) Get(key string) (string, bool) {
	fmt.Printf("🗄️  [Cache:%s:%d] GET %s\n", s.Host, s.Port, key)
	return "", false
}

func (s *CacheService) Set(key, value string) {
	fmt.Printf("🗄️  [Cache:%s:%d] SET %s = %s\n", s.Host, s.Port, key, value)
}

func NewCacheService(host string, port int) *CacheService {
	fmt.Printf("   🏗️  Creating CacheService: %s:%d\n", host, port)
	return &CacheService{Host: host, Port: port}
}

func CacheServiceFactory(cfg map[string]any) any {
	return NewCacheService(
		utils.GetValueFromMap(cfg, "host", "localhost"),
		utils.GetValueFromMap(cfg, "port", 6379),
	)
}

// =============================================================================
// LAYER 2: DOMAIN SERVICES (Depend on infrastructure)
// =============================================================================

// UserService - User management (depends on DB + Cache)
// BEST PRACTICE: Store service NAMES, not instances. Use lazy getters.
type UserService struct {
	// Service names from config
	dbServiceName    string
	cacheServiceName string

	// Cache variables (nil until first GetService call)
	dbCache    *service.Cached[*DBService]
	cacheCache *service.Cached[*CacheService]
}

// Lazy getter for DB - only calls GetService when needed
func (s *UserService) getDB() *DBService {
	return s.dbCache.MustGet()
}

// Lazy getter for Cache - only calls GetService when needed
func (s *UserService) getCache() *CacheService {
	return s.cacheCache.MustGet()
}

func (s *UserService) GetUser(id string) map[string]any {
	// Lazy load dependencies when needed
	cache := s.getCache()
	db := s.getDB()

	// Try cache first
	cacheKey := "user:" + id
	if cached, ok := cache.Get(cacheKey); ok {
		fmt.Println("   ✓ Cache hit!")
		return map[string]any{"id": id, "name": cached, "source": "cache"}
	}

	// Cache miss, query database
	result := db.Query("SELECT * FROM users WHERE id = " + id)
	userName := "User " + id

	// Store in cache
	cache.Set(cacheKey, userName)

	return map[string]any{
		"id":     id,
		"name":   userName,
		"source": "database",
		"result": result,
	}
}

func (s *UserService) CreateUser(name string) map[string]any {
	db := s.getDB() // Lazy load
	db.Execute("INSERT INTO users (name) VALUES ('" + name + "')")
	return map[string]any{"id": "new-id", "name": name}
}

// Factory: NO GetService calls! Just store service names.
// This is the key to lazy loading and proper dependency management.
func UserServiceFactory(cfg map[string]any) any {
	fmt.Println("   🏗️  Creating UserService (lazy dependencies)")

	// Get service names from config (with defaults)
	dbServiceName := utils.GetValueFromMap(cfg, "db_service", "db-service")
	cacheServiceName := utils.GetValueFromMap(cfg, "cache_service", "cache-service")

	return &UserService{
		dbServiceName:    dbServiceName,
		cacheServiceName: cacheServiceName,
		// dbCache and cacheCache are nil - will be filled on first use
	}
}

// OrderService - Order management (depends on UserService + DB)
// BEST PRACTICE: Store service names, use lazy getters
type OrderService struct {
	// Service names from config
	dbServiceName   string
	userServiceName string

	// Cache variables (nil until first use)
	dbCache   *service.Cached[*DBService]
	userCache *service.Cached[*UserService]
}

// Lazy getters
func (s *OrderService) getDB() *DBService {
	return s.dbCache.MustGet()
}

func (s *OrderService) getUser() *UserService {
	return s.userCache.MustGet()
}

func (s *OrderService) CreateOrder(userID string, product string, amount float64) map[string]any {
	// Lazy load dependencies
	db := s.getDB()
	users := s.getUser()

	// Get user info
	user := users.GetUser(userID)

	// Create order
	orderID := "order-" + fmt.Sprint(time.Now().Unix())
	db.Execute(fmt.Sprintf("INSERT INTO orders (user_id, product, amount) VALUES ('%s', '%s', %.2f)", userID, product, amount))

	return map[string]any{
		"order_id": orderID,
		"user":     user["name"],
		"product":  product,
		"amount":   amount,
	}
}

func (s *OrderService) GetOrder(orderID string) map[string]any {
	db := s.getDB() // Lazy load
	result := db.Query("SELECT * FROM orders WHERE id = " + orderID)
	return map[string]any{
		"order_id": orderID,
		"result":   result,
	}
}

// Factory: NO GetService! Just store service names.
func OrderServiceFactory(cfg map[string]any) any {
	fmt.Println("   🏗️  Creating OrderService (lazy dependencies)")

	// Get service names from config (with defaults)
	dbServiceName := utils.GetValueFromMap(cfg, "db_service", "db-service")
	userServiceName := utils.GetValueFromMap(cfg, "user_service", "user-service")

	return &OrderService{
		dbServiceName:   dbServiceName,
		userServiceName: userServiceName,
		// Caches are nil - filled on first use
	}
}

// =============================================================================
// SERVICE CONTAINER
// =============================================================================

type ServiceContainer struct {
	dbCache    *service.Cached[*DBService]
	cacheCache *service.Cached[*CacheService]
	userCache  *service.Cached[*UserService]
	orderCache *service.Cached[*OrderService]
}

func (sc *ServiceContainer) GetDB() *DBService {
	return sc.dbCache.MustGet()
}

func (sc *ServiceContainer) GetCache() *CacheService {
	return sc.cacheCache.MustGet()
}

func (sc *ServiceContainer) GetUser() *UserService {
	return sc.userCache.MustGet()
}

func (sc *ServiceContainer) GetOrder() *OrderService {
	return sc.orderCache.MustGet()
}

var services = &ServiceContainer{}

// =============================================================================
// ROUTERS
// =============================================================================

func createUserRouter() lokstra.Router {
	r := lokstra.NewRouter("user-router")

	r.GET("/api/users/{id}", func(c *lokstra.RequestContext) error {
		id := c.Req.PathParam("id", "")
		user := services.GetUser()
		result := user.GetUser(id)

		return c.Api.Ok(result)
	})

	r.POST("/api/users", func(c *lokstra.RequestContext) error {
		var req struct {
			Name string `json:"name" binding:"required"`
		}

		if err := c.Req.BindBody(&req); err != nil {
			return c.Api.BadRequest("INVALID_INPUT", err.Error())
		}

		user := services.GetUser()
		result := user.CreateUser(req.Name)

		return c.Api.Created(result, "User created")
	})

	return r
}

func createOrderRouter() lokstra.Router {
	r := lokstra.NewRouter("order-router")

	r.POST("/api/orders", func(c *lokstra.RequestContext) error {
		var req struct {
			UserID  string  `json:"user_id" binding:"required"`
			Product string  `json:"product" binding:"required"`
			Amount  float64 `json:"amount" binding:"required,gt=0"`
		}

		if err := c.Req.BindBody(&req); err != nil {
			return c.Api.BadRequest("INVALID_INPUT", err.Error())
		}

		order := services.GetOrder()
		result := order.CreateOrder(req.UserID, req.Product, req.Amount)

		return c.Api.Created(result, "Order created")
	})

	r.GET("/api/orders/{id}", func(c *lokstra.RequestContext) error {
		id := c.Req.PathParam("id", "")
		order := services.GetOrder()
		result := order.GetOrder(id)

		return c.Api.Ok(result)
	})

	return r
}

func createHealthRouter() lokstra.Router {
	r := lokstra.NewRouter("health-router")

	r.GET("/health", func(c *lokstra.RequestContext) error {
		// Test all services
		db := services.GetDB()
		cache := services.GetCache()

		return c.Api.Ok(map[string]any{
			"status": "healthy",
			"services": map[string]any{
				"database": fmt.Sprintf("%s:%d/%s", db.Host, db.Port, db.Database),
				"cache":    fmt.Sprintf("%s:%d", cache.Host, cache.Port),
			},
			"time": time.Now().Format(time.RFC3339),
		})
	})

	return r
}

// =============================================================================
// SETUP
// =============================================================================

func setupRegistry() {
	fmt.Println("🔧 Setting up registry...")

	// Register factories
	old_registry.RegisterServiceType("db", DBServiceFactory)
	old_registry.RegisterServiceType("cache", CacheServiceFactory)
	old_registry.RegisterServiceType("user", UserServiceFactory)
	old_registry.RegisterServiceType("order", OrderServiceFactory)
	fmt.Println("   ✓ Registered factories: db, cache, user, order")

	// Register routers
	old_registry.RegisterRouter("user-api", createUserRouter())
	old_registry.RegisterRouter("order-api", createOrderRouter())
	old_registry.RegisterRouter("health-api", createHealthRouter())
	fmt.Println("   ✓ Registered routers: user-api, order-api, health-api")

	fmt.Println()
}

// =============================================================================
// MAIN
// =============================================================================

func main() {
	fmt.Println("📚 Lokstra Learning: 03-Service Dependencies")
	fmt.Println("===============================================")
	fmt.Println()

	setupRegistry()

	fmt.Println("📄 Loading configuration...")
	cfg := config.New()

	if err := config.LoadConfigFile("config.yaml", cfg); err != nil {
		fmt.Printf("❌ Failed to load config.yaml: %v\n", err)
		return
	}
	fmt.Println("   ✓ Loaded config.yaml")
	fmt.Println()

	old_registry.RegisterConfig(cfg, "")

	fmt.Println("🔍 Service dependency tree:")
	fmt.Println("   db-service (no dependencies)")
	fmt.Println("   cache-service (no dependencies)")
	fmt.Println("   user-service → db-service, cache-service")
	fmt.Println("   order-service → db-service, user-service")
	fmt.Println()

	fmt.Println("🚀 Starting server...")
	old_registry.PrintServerStartInfo()
	if err := old_registry.StartServer(); err != nil {
		fmt.Printf("❌ Server error: %v\n", err)
	}
}
