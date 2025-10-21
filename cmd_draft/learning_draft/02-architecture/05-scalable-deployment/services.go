package main

import (
	"fmt"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/old_registry"
)

// =============================================================================
// INFRASTRUCTURE SERVICES (Layer 1)
// =============================================================================

type DBService struct {
	Host     string
	Port     int
	Database string
}

func (db *DBService) Query(query string) string {
	fmt.Printf("   [DB] Query: %s\n", query)
	return "db-result"
}

func (db *DBService) Execute(query string, args ...any) {
	fmt.Printf("   [DB] Execute: %s with %v\n", query, args)
}

func DBServiceFactory(cfg map[string]any) any {
	return &DBService{
		Host:     utils.GetValueFromMap(cfg, "host", "localhost"),
		Port:     utils.GetValueFromMap(cfg, "port", 5432),
		Database: utils.GetValueFromMap(cfg, "database", "ecommerce"),
	}
}

type CacheService struct {
	Host string
	Port int
}

func (c *CacheService) Get(key string) (any, bool) {
	fmt.Printf("   [Cache] GET %s -> miss\n", key)
	return nil, false
}

func (c *CacheService) Set(key string, value any) {
	fmt.Printf("   [Cache] SET %s\n", key)
}

func CacheServiceFactory(cfg map[string]any) any {
	return &CacheService{
		Host: utils.GetValueFromMap(cfg, "host", "localhost"),
		Port: utils.GetValueFromMap(cfg, "port", 6379),
	}
}

// =============================================================================
// REPOSITORY SERVICES (Layer 2)
// =============================================================================

type ProductRepository struct {
	dbServiceName    string
	cacheServiceName string
	dbCache          *service.Cached[*DBService]
	cacheCache       *service.Cached[*CacheService]
}

func (r *ProductRepository) getDB() *DBService {
	return r.dbCache.MustGet()
}

func (r *ProductRepository) getCache() *CacheService {
	return r.cacheCache.MustGet()
}

func (r *ProductRepository) FindByID(id string) map[string]any {
	cacheKey := "product:" + id
	if cached, ok := r.getCache().Get(cacheKey); ok {
		return cached.(map[string]any)
	}

	r.getDB().Query("SELECT * FROM products WHERE id = " + id)
	product := map[string]any{
		"id":    id,
		"name":  "Product " + id,
		"price": 99.99,
	}

	r.getCache().Set(cacheKey, product)
	return product
}

func (r *ProductRepository) FindAll() []map[string]any {
	r.getDB().Query("SELECT * FROM products")
	return []map[string]any{
		{"id": "1", "name": "Laptop", "price": 1299.99},
		{"id": "2", "name": "Mouse", "price": 29.99},
		{"id": "3", "name": "Keyboard", "price": 79.99},
	}
}

func ProductRepositoryFactory(cfg map[string]any) any {
	return &ProductRepository{
		dbServiceName:    utils.GetValueFromMap(cfg, "db_service", "db-service"),
		cacheServiceName: utils.GetValueFromMap(cfg, "cache_service", "cache-service"),
	}
}

type OrderRepository struct {
	dbServiceName string
	dbCache       *service.Cached[*DBService]
}

func (r *OrderRepository) getDB() *DBService {
	return r.dbCache.MustGet()
}

func (r *OrderRepository) Create(userID string, items []map[string]any, total float64) map[string]any {
	r.getDB().Execute("INSERT INTO orders (user_id, total) VALUES (?, ?)", userID, total)
	return map[string]any{
		"id":      "order-123",
		"user_id": userID,
		"items":   items,
		"total":   total,
		"status":  "pending",
	}
}

func (r *OrderRepository) FindByID(id string) map[string]any {
	r.getDB().Query("SELECT * FROM orders WHERE id = " + id)
	return map[string]any{
		"id":      id,
		"user_id": "user123",
		"total":   299.98,
		"status":  "pending",
	}
}

func OrderRepositoryFactory(cfg map[string]any) any {
	return &OrderRepository{
		dbServiceName: utils.GetValueFromMap(cfg, "db_service", "db-service"),
	}
}

// =============================================================================
// BUSINESS SERVICES (Layer 3)
// =============================================================================

type ProductService struct {
	repoServiceName string
	repoCache       *service.Cached[*ProductRepository]
}

func (s *ProductService) getRepo() *ProductRepository {
	return s.repoCache.MustGet()
}

func (s *ProductService) GetProduct(id string) map[string]any {
	fmt.Printf("📦 ProductService.GetProduct(%s)\n", id)
	return s.getRepo().FindByID(id)
}

func (s *ProductService) GetProducts() []map[string]any {
	fmt.Println("📦 ProductService.GetProducts()")
	return s.getRepo().FindAll()
}

func ProductServiceFactory(cfg map[string]any) any {
	return &ProductService{
		repoServiceName: utils.GetValueFromMap(cfg, "repo_service", "product-repository"),
	}
}

type OrderService struct {
	repoServiceName string
	repoCache       *service.Cached[*OrderRepository]
}

func (s *OrderService) getRepo() *OrderRepository {
	return s.repoCache.MustGet()
}

func (s *OrderService) CreateOrder(userID string, productIDs []string, total float64) map[string]any {
	fmt.Printf("🛒 OrderService.CreateOrder(user: %s, products: %v)\n", userID, productIDs)

	// Prepare order items
	items := make([]map[string]any, len(productIDs))
	for i, pid := range productIDs {
		items[i] = map[string]any{"product_id": pid}
	}

	return s.getRepo().Create(userID, items, total)
}

func (s *OrderService) GetOrder(id string) map[string]any {
	fmt.Printf("🛒 OrderService.GetOrder(%s)\n", id)
	return s.getRepo().FindByID(id)
}

func OrderServiceFactory(cfg map[string]any) any {
	return &OrderService{
		repoServiceName: utils.GetValueFromMap(cfg, "repo_service", "order-repository"),
	}
}

// =============================================================================
// SERVICE CONTAINER
// =============================================================================

type ServiceContainer struct {
	productCache *service.Cached[*ProductService]
	orderCache   *service.Cached[*OrderService]
}

func (sc *ServiceContainer) GetProduct() *ProductService {
	return sc.productCache.MustGet()
}

func (sc *ServiceContainer) GetOrder() *OrderService {
	return sc.orderCache.MustGet()
}

var services = &ServiceContainer{}

// =============================================================================
// FACTORY REGISTRATION
// =============================================================================

func setupFactories() {
	fmt.Println("🔧 Registering Service Factories...")

	// Layer 1: Infrastructure
	old_registry.RegisterServiceType("db", DBServiceFactory)
	old_registry.RegisterServiceType("cache", CacheServiceFactory)

	// Layer 2: Repositories
	old_registry.RegisterServiceType("product-repository", ProductRepositoryFactory)
	old_registry.RegisterServiceType("order-repository", OrderRepositoryFactory)

	// Layer 3: Business Services
	old_registry.RegisterServiceType("product", ProductServiceFactory)
	old_registry.RegisterServiceType("order", OrderServiceFactory)

	fmt.Println("✅ Service factories registered")
}
