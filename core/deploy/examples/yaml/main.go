package main

import (
	"fmt"
	"log"

	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/deploy/loader"
	"github.com/primadi/lokstra/core/service"
)

// ===== DOMAIN MODELS =====

type User struct {
	ID   int
	Name string
}

// ===== SERVICES =====

type DBPool struct {
	Host     string
	Port     int
	Database string
	MaxConns int
}

func (db *DBPool) Connect() {
	fmt.Printf("📦 Connected to PostgreSQL: %s:%d/%s (max conns: %d)\n",
		db.Host, db.Port, db.Database, db.MaxConns)
}

type Logger struct {
	Level  string
	Format string
}

func (l *Logger) Info(msg string) {
	fmt.Printf("ℹ️  [%s|%s] %s\n", l.Level, l.Format, msg)
}

type Cache struct {
	Host string
	Port int
}

func (c *Cache) Connect() {
	fmt.Printf("🔄 Connected to Redis: %s:%d\n", c.Host, c.Port)
}

type UserService struct {
	DB     *service.Cached[*DBPool]
	Logger *service.Cached[*Logger]
	Cache  *service.Cached[*Cache]
}

func (us *UserService) GetUser(id int) *User {
	logger := us.Logger.Get()
	logger.Info(fmt.Sprintf("Getting user %d", id))

	cache := us.Cache.Get()
	fmt.Printf("   💾 Checking cache at %s:%d\n", cache.Host, cache.Port)

	return &User{ID: id, Name: "John Doe"}
}

type OrderService struct {
	DB          *service.Cached[*DBPool]
	UserService *service.Cached[*UserService]
	Logger      *service.Cached[*Logger]
}

func (os *OrderService) GetOrder(id int) {
	logger := os.Logger.Get()
	logger.Info(fmt.Sprintf("Getting order %d", id))

	userService := os.UserService.MustGet()
	user := userService.GetUser(1)
	fmt.Printf("   📝 Order for user: %s\n", user.Name)
}

// ===== FACTORIES =====

func dbPoolFactory(deps map[string]any, config map[string]any) any {
	pool := &DBPool{
		Host:     config["host"].(string),
		Port:     int(config["port"].(int)),
		Database: config["database"].(string),
		MaxConns: int(config["max-conns"].(int)),
	}
	pool.Connect()
	return pool
}

func loggerFactory(deps map[string]any, config map[string]any) any {
	logger := &Logger{
		Level:  config["level"].(string),
		Format: config["format"].(string),
	}
	fmt.Printf("📝 Logger initialized (level: %s, format: %s)\n", logger.Level, logger.Format)
	return logger
}

func cacheFactory(deps map[string]any, config map[string]any) any {
	cache := &Cache{
		Host: config["host"].(string),
		Port: int(config["port"].(int)),
	}
	cache.Connect()
	return cache
}

func userServiceFactory(deps map[string]any, config map[string]any) any {
	return &UserService{
		DB:     service.Cast[*DBPool](deps["db"]),
		Logger: service.Cast[*Logger](deps["logger"]),
		Cache:  service.Cast[*Cache](deps["cache"]),
	}
}

func orderServiceFactory(deps map[string]any, config map[string]any) any {
	return &OrderService{
		DB:          service.Cast[*DBPool](deps["dbOrder"]),
		UserService: service.Cast[*UserService](deps["userSvc"]),
		Logger:      service.Cast[*Logger](deps["logger"]),
	}
}

// ===== MAIN =====

func main() {
	fmt.Println("🚀 Lokstra YAML Configuration Example")
	fmt.Println("=" + string(make([]byte, 60)))
	fmt.Println()

	// Create global registry
	reg := deploy.Global()

	// Register service factories
	fmt.Println("🔧 Registering service factories...")
	reg.RegisterServiceType("postgres-pool", dbPoolFactory, nil)
	reg.RegisterServiceType("logger-service", loggerFactory, nil)
	reg.RegisterServiceType("redis-cache", cacheFactory, nil)
	reg.RegisterServiceType("user-service-factory", userServiceFactory, nil)
	reg.RegisterServiceType("order-service-factory", orderServiceFactory, nil)
	fmt.Println()

	// Load configuration from YAML files
	fmt.Println("📂 Loading configuration from YAML files...")
	fmt.Println("   - config/base.yaml")
	fmt.Println("   - config/services.yaml")
	fmt.Println("   - config/deployments.yaml")
	fmt.Println()

	dep, err := loader.LoadAndBuildFromDir("config", "production", reg)
	if err != nil {
		log.Fatal("❌ Failed to load deployment:", err)
	}

	fmt.Println("✅ Configuration loaded and validated!")
	fmt.Println()

	// Get server and app
	server, ok := dep.GetServer("api-server")
	if !ok {
		log.Fatal("api-server not found")
	}

	apps := server.Apps()
	if len(apps) == 0 {
		log.Fatal("no apps found")
	}

	app := apps[0]

	// Instantiate services
	fmt.Println("🔨 Instantiating services from YAML configuration...")
	fmt.Println()

	_, err = app.GetService("user-service")
	if err != nil {
		log.Fatal(err)
	}

	orderSvc, err := app.GetService("order-service")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println()
	fmt.Println("✅ All services instantiated!")
	fmt.Println()

	// Use services
	fmt.Println("🎯 Using services...")
	fmt.Println("=" + string(make([]byte, 60)))
	fmt.Println()

	os := orderSvc.(*OrderService)
	os.GetOrder(1)

	fmt.Println()

	// Show deployment structure
	fmt.Println("📊 Deployment Structure (from YAML):")
	fmt.Println("=" + string(make([]byte, 60)))
	fmt.Printf("Deployment: %s\n", dep.Name())
	fmt.Printf("  Server: %s (%s)\n", server.Name(), server.BaseURL())
	fmt.Printf("    App (addr %s):\n", app.Addr())
	fmt.Printf("      Services: %d (loaded from YAML)\n", len(app.Services()))
	for name := range app.Services() {
		fmt.Printf("        - %s\n", name)
	}

	fmt.Println()
	fmt.Println("✨ YAML configuration example completed successfully!")
}
