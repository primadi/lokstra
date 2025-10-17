package main

import (
	"fmt"
	"log"

	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/deploy/schema"
	"github.com/primadi/lokstra/core/service"
)

// ===== DOMAIN MODELS =====

type User struct {
	ID   int
	Name string
}

type Order struct {
	ID     int
	UserID int
	Total  float64
}

// ===== SERVICES =====

type DBPool struct {
	DSN      string
	MaxConns int
}

func (db *DBPool) Connect() {
	fmt.Printf("📦 Connected to database: %s (max conns: %d)\n", db.DSN, db.MaxConns)
}

type Logger struct {
	Level string
}

func (l *Logger) Info(msg string) {
	fmt.Printf("ℹ️  [%s] %s\n", l.Level, msg)
}

type UserService struct {
	DB     *service.Cached[*DBPool] // Lazy-loaded DBPool
	Logger *service.Cached[*Logger] // Lazy-loaded Logger
}

func (us *UserService) GetUser(id int) *User {
	// Resolve lazy-loaded logger (typed, no cast needed!)
	logger := us.Logger.Get()
	logger.Info(fmt.Sprintf("Getting user %d", id))
	return &User{ID: id, Name: "John Doe"}
}

type OrderService struct {
	DB          *service.Cached[*DBPool]      // Lazy-loaded DBPool
	UserService *service.Cached[*UserService] // Lazy-loaded UserService
	Logger      *service.Cached[*Logger]      // Lazy-loaded Logger
}

func (os *OrderService) GetOrder(id int) *Order {
	// Resolve lazy-loaded dependencies (typed, no cast needed!)
	logger := os.Logger.Get()
	logger.Info(fmt.Sprintf("Getting order %d", id))

	// Use user service (also lazy-loaded, MustGet for fail-fast)
	userService := os.UserService.MustGet()
	user := userService.GetUser(1)
	fmt.Printf("   📝 Order for user: %s\n", user.Name)

	return &Order{ID: id, UserID: 1, Total: 99.99}
}

// ===== FACTORIES =====

func dbPoolFactory(deps map[string]any, config map[string]any) any {
	pool := &DBPool{
		DSN:      config["dsn"].(string),
		MaxConns: config["max-conns"].(int),
	}
	pool.Connect()
	return pool
}

func loggerFactory(deps map[string]any, config map[string]any) any {
	logger := &Logger{
		Level: config["level"].(string),
	}
	fmt.Printf("📝 Logger initialized (level: %s)\n", logger.Level)
	return logger
}

func userServiceFactory(deps map[string]any, config map[string]any) any {
	return &UserService{
		DB:     service.Cast[*DBPool](deps["db"]),
		Logger: service.Cast[*Logger](deps["logger"]),
	}
}

func orderServiceFactory(deps map[string]any, config map[string]any) any {
	return &OrderService{
		DB:          service.Cast[*DBPool](deps["dbOrder"]),
		UserService: service.Cast[*UserService](deps["userSvc"]),
		Logger:      service.Cast[*Logger](deps["logger"]),
	}
}

// ===== SETUP =====

func setupRegistry() *deploy.GlobalRegistry {
	reg := deploy.Global()

	fmt.Println("🔧 Registering service factories...")

	// Register factories
	reg.RegisterServiceType("dbpool_pg", dbPoolFactory, nil)
	reg.RegisterServiceType("logger-factory", loggerFactory, nil)
	reg.RegisterServiceType("user-factory", userServiceFactory, nil)
	reg.RegisterServiceType("order-factory", orderServiceFactory, nil)

	// Define configs
	fmt.Println("⚙️  Defining configurations...")

	reg.DefineConfig(&schema.ConfigDef{
		Name:  "DB_USER_DSN",
		Value: "postgres://localhost/users",
	})

	reg.DefineConfig(&schema.ConfigDef{
		Name:  "DB_ORDER_DSN",
		Value: "postgres://localhost/orders",
	})

	reg.DefineConfig(&schema.ConfigDef{
		Name:  "DB_MAX_CONNS",
		Value: 20,
	})

	reg.DefineConfig(&schema.ConfigDef{
		Name:  "LOG_LEVEL",
		Value: "info",
	})

	// Resolve configs
	if err := reg.ResolveConfigs(); err != nil {
		log.Fatal(err)
	}

	// Define services
	fmt.Println("📋 Defining services...")

	reg.DefineService(&schema.ServiceDef{
		Name: "db-user",
		Type: "dbpool_pg",
		Config: map[string]any{
			"dsn":       "${@cfg:DB_USER_DSN}",
			"max-conns": "${@cfg:DB_MAX_CONNS}",
		},
	})

	reg.DefineService(&schema.ServiceDef{
		Name: "db-order",
		Type: "dbpool_pg",
		Config: map[string]any{
			"dsn":       "${@cfg:DB_ORDER_DSN}",
			"max-conns": "${@cfg:DB_MAX_CONNS}",
		},
	})

	reg.DefineService(&schema.ServiceDef{
		Name: "logger",
		Type: "logger-factory",
		Config: map[string]any{
			"level": "${@cfg:LOG_LEVEL}",
		},
	})

	reg.DefineService(&schema.ServiceDef{
		Name:      "user-service",
		Type:      "user-factory",
		DependsOn: []string{"db:db-user", "logger"},
	})

	reg.DefineService(&schema.ServiceDef{
		Name:      "order-service",
		Type:      "order-factory",
		DependsOn: []string{"dbOrder:db-order", "userSvc:user-service", "logger"},
	})

	return reg
}

// ===== MAIN =====

func main() {
	fmt.Println("🚀 Lokstra Deploy API Example")
	fmt.Println("=" + string(make([]byte, 50)) + "\n")

	// Setup global registry
	setupRegistry()

	fmt.Println("\n✨ Creating deployment...")

	// Create deployment
	dep := deploy.New("monolith")

	// Override config for this deployment
	dep.SetConfigOverride("LOG_LEVEL", "debug")

	// Create server
	fmt.Println("🖥️  Creating server...")
	server := dep.NewServer("main-server", "http://localhost")

	// Create app
	fmt.Println("📱 Creating app on port 3000...")
	app := server.NewApp(3000)

	// Add services
	fmt.Println("➕ Adding services to app...")
	app.AddServices("db-user", "db-order", "logger", "user-service", "order-service")

	fmt.Println("\n🏗️  Building deployment...")
	fmt.Println("=" + string(make([]byte, 50)) + "\n")

	// Get services (this will trigger instantiation with DI)
	fmt.Println("🔨 Instantiating services...")
	fmt.Println()

	// Get user service (will instantiate db-user, logger, and user-service)
	userSvc, err := app.GetService("user-service")
	if err != nil {
		log.Fatal(err)
	}

	// Get order service (will reuse existing instances and instantiate order-service)
	orderSvc, err := app.GetService("order-service")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println()
	fmt.Println("✅ All services instantiated!")
	fmt.Println()

	// Use services
	fmt.Println("🎯 Using services...")
	fmt.Println("=" + string(make([]byte, 50)) + "\n")

	us := userSvc.(*UserService)
	user := us.GetUser(1)
	fmt.Printf("👤 Got user: %+v\n\n", user)

	os := orderSvc.(*OrderService)
	order := os.GetOrder(1)
	fmt.Printf("📦 Got order: %+v\n\n", order)

	// Show deployment structure
	fmt.Println("📊 Deployment Structure:")
	fmt.Println("=" + string(make([]byte, 50)))
	fmt.Printf("Deployment: %s\n", dep.Name())
	fmt.Printf("  Server: %s (%s)\n", server.Name(), server.BaseURL())
	fmt.Printf("    App (port %d):\n", app.Port())
	fmt.Printf("      Services: %d\n", len(app.Services()))
	for name := range app.Services() {
		fmt.Printf("        - %s\n", name)
	}

	fmt.Println("\n✨ Example completed successfully!")
}
