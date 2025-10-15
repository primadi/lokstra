package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/lokstra_registry"
)

// Services in Lokstra
//
// Services are reusable components that provide specific functionality:
// - Database access
// - Caching
// - Email sending
// - Logging
// - Authentication
//
// Key Concepts:
// 1. Service Factory - Creates services from config (map[string]any)
// 2. Service Container - Caches services with proper pattern
// 3. Lazy Loading - Services created only when first accessed
// 4. Dependency Injection - Services can depend on other services
// 5. Interface-based - Easy testing with mocks
//
// Run: go run .

// === 1. SIMPLE SERVICE (Email) ===

type EmailConfig struct {
	SMTPHost string
	SMTPPort int
	From     string
}

type EmailService struct {
	cfg *EmailConfig
}

func (s *EmailService) Send(to, subject, body string) {
	fmt.Printf("📧 Email: %s -> %s: %s\n", s.cfg.From, to, subject)
}

func NewEmailService(cfg *EmailConfig) *EmailService {
	fmt.Println("   🏗️  Creating EmailService")
	return &EmailService{cfg: cfg}
}

// Factory: converts map[string]any (from YAML) to typed service
func EmailServiceFactory(params map[string]any) any {
	cfg := &EmailConfig{
		SMTPHost: utils.GetValueFromMap(params, "smtp_host", "localhost"),
		SMTPPort: utils.GetValueFromMap(params, "smtp_port", 587),
		From:     utils.GetValueFromMap(params, "from", "noreply@app.com"),
	}
	return NewEmailService(cfg)
}

// === 2. SERVICE WITH DEPENDENCIES (User Service needs DB) ===

type DBService struct {
	name string
}

func (s *DBService) Query(sql string) string {
	fmt.Printf("💾 DB Query: %s\n", sql)
	return "result from " + s.name
}

func NewDBService(name string) *DBService {
	fmt.Printf("   🏗️  Creating DBService: %s\n", name)
	return &DBService{name: name}
}

func DBServiceFactory(params map[string]any) any {
	name := utils.GetValueFromMap(params, "name", "postgres")
	return NewDBService(name)
}

type UserService struct {
	db *service.Cached[*DBService]
}

func (s *UserService) GetUser(id string) map[string]any {
	result := s.db.MustGet().Query("SELECT * FROM users WHERE id = " + id)
	return map[string]any{
		"id":     id,
		"name":   "User " + id,
		"source": result,
	}
}

func (s *UserService) CreateUser(name string) map[string]any {
	s.db.MustGet().Query("INSERT INTO users (name) VALUES ('" + name + "')")
	return map[string]any{"id": "new-id", "name": name}
}

func NewUserService() *UserService {
	return &UserService{db: service.LazyLoad[*DBService]("db")}
}

// === 3. INTERFACE-BASED SERVICE (Easy testing) ===

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type ConsoleLogger struct {
	prefix string
}

func (l *ConsoleLogger) Info(msg string) {
	fmt.Printf("[%s] INFO: %s\n", l.prefix, msg)
}

func (l *ConsoleLogger) Error(msg string) {
	fmt.Printf("[%s] ERROR: %s\n", l.prefix, msg)
}

func NewConsoleLogger(prefix string) *ConsoleLogger {
	fmt.Printf("   🏗️  Creating ConsoleLogger: %s\n", prefix)
	return &ConsoleLogger{prefix: prefix}
}

func ConsoleLoggerFactory(params map[string]any) any {
	prefix := utils.GetValueFromMap(params, "prefix", "APP")
	return NewConsoleLogger(prefix)
}

// === 4. SERVICE CONTAINER (Proper Caching Pattern) ===

type ServiceContainer struct {
	emailCache  *service.Cached[*EmailService]
	dbCache     *service.Cached[*DBService]
	userCache   *service.Cached[*UserService]
	loggerCache *service.Cached[*ConsoleLogger]
}

// Getters with caching (CRITICAL PATTERN!)
func (sc *ServiceContainer) GetEmail() *EmailService {
	return sc.emailCache.MustGet()
}

func (sc *ServiceContainer) GetDB() *DBService {
	return sc.dbCache.MustGet()
}

func (sc *ServiceContainer) GetUser() *UserService {
	return sc.userCache.MustGet()
}

func (sc *ServiceContainer) GetLogger() *ConsoleLogger {
	return sc.loggerCache.MustGet()
}

// Global container
var services = &ServiceContainer{}

func main() {
	fmt.Println("🎯 Services Demo - Complete Example")
	fmt.Println(strings.Repeat("=", 47))

	// === STEP 1: Register Service Factories ===
	fmt.Println("\n📝 Step 1: Registering Service Factories")

	lokstra_registry.RegisterServiceFactory("email", EmailServiceFactory)
	lokstra_registry.RegisterServiceFactory("db", DBServiceFactory)
	lokstra_registry.RegisterServiceFactory("user", NewUserService)
	lokstra_registry.RegisterServiceFactory("logger", ConsoleLoggerFactory)

	fmt.Println("   ✅ Registered: email, db, user, logger")

	// === STEP 2: Create Services (Two Ways) ===
	fmt.Println("\n🏗️  Step 2: Creating Services")

	// Way 1: NewService - Eager creation
	fmt.Println("\n   Method 1: NewService (eager)")
	emailSvc := lokstra_registry.NewService[*EmailService](
		"email",
		"email",
		map[string]any{
			"smtp_host": "smtp.gmail.com",
			"smtp_port": 587,
			"from":      "no-reply@myapp.com",
		},
	)

	// Way 2: RegisterLazyService - Lazy creation (only when accessed)
	fmt.Println("\n   Method 2: RegisterLazyService (lazy)")
	lokstra_registry.RegisterLazyService("db", "db", map[string]any{
		"name": "postgres-main",
	})
	lokstra_registry.RegisterLazyService("user", "user", map[string]any{})
	lokstra_registry.RegisterLazyService("logger", "logger", map[string]any{
		"prefix": "DEMO",
	})

	fmt.Println("   ✅ Lazy services registered (not created yet)")

	// === STEP 3: Use Services Directly ===
	fmt.Println("\n▶️  Step 3: Using Services Directly")
	emailSvc.Send("user@example.com", "Welcome", "Thanks for signing up!")

	// === STEP 4: Service Container Pattern ===
	fmt.Println("\n🔄 Step 4: Service Container Pattern (Recommended!)")

	fmt.Println("\n   First call - creates and caches:")
	logger1 := services.GetLogger()
	logger1.Info("Application started")

	fmt.Println("\n   Second call - returns cached:")
	logger2 := services.GetLogger()
	logger2.Info("Using cached logger")

	fmt.Printf("   Same instance? %v ✅\n", logger1 == logger2)

	// === STEP 5: Services with Dependencies ===
	fmt.Println("\n🔗 Step 5: Services with Dependencies")

	fmt.Println("\n   Getting UserService (depends on DBService):")
	userSvc := services.GetUser() // This triggers DBService creation

	user := userSvc.GetUser("123")
	fmt.Printf("   Retrieved: %v\n", user)

	// === STEP 6: HTTP Server with Services ===
	fmt.Println("\n🌐 Step 6: HTTP Server with Services")

	r := lokstra.NewRouter("services-demo")

	// Endpoint using services
	r.GET("/api/users/:id", func(c *lokstra.RequestContext) error {
		id := c.Req.PathParam("id", "")

		// Get services via container (lazy + cached)
		userService := services.GetUser()
		logger := services.GetLogger()

		logger.Info("Fetching user: " + id)
		user := userService.GetUser(id)

		return c.Api.Ok(user)
	})

	r.POST("/api/users", func(c *lokstra.RequestContext) error {
		type Input struct {
			Name string `json:"name" validate:"required"`
		}

		var input Input
		if err := c.Req.BindBody(&input); err != nil {
			return c.Api.BadRequest("INVALID_INPUT", err.Error())
		}

		userService := services.GetUser()
		logger := services.GetLogger()
		emailService := services.GetEmail()

		logger.Info("Creating user: " + input.Name)
		user := userService.CreateUser(input.Name)

		// Send welcome email
		emailService.Send(input.Name+"@example.com", "Welcome", "Account created")

		return c.Api.Created(user, "User created")
	})

	r.GET("/api/health", func(c *lokstra.RequestContext) error {
		logger := services.GetLogger()
		logger.Info("Health check")

		return c.Api.Ok(map[string]any{
			"status": "healthy",
			"time":   time.Now(),
		})
	})

	r.GET("/", func(c *lokstra.RequestContext) error {
		return c.Api.Ok(map[string]any{
			"service": "services-demo",
			"endpoints": []string{
				"GET  /api/users/:id",
				"POST /api/users",
				"GET  /api/health",
			},
		})
	})

	// === STEP 7: Demonstrate Patterns ===
	fmt.Println("\n💡 Key Patterns Demonstrated:")
	fmt.Println("   1. ✅ Service Factory (map[string]any -> typed service)")
	fmt.Println("   2. ✅ Service Container (proper caching pattern)")
	fmt.Println("   3. ✅ Lazy Loading (created only when needed)")
	fmt.Println("   4. ✅ Dependency Injection (UserService needs DBService)")
	fmt.Println("   5. ✅ Interface-based (Logger interface)")
	fmt.Println("   6. ✅ Usage in HTTP handlers")

	fmt.Println("\n⚠️  Critical: Service Container Pattern")
	fmt.Println("   ❌ DON'T: var cache *Service (local variable)")
	fmt.Println("   ✅ DO:    struct field + getter method")
	fmt.Println("   Why? Local variables reset on each call (no caching!)")

	fmt.Println("\n🚀 Server starting on http://localhost:8080")
	fmt.Println("\n💡 Test commands:")
	fmt.Println("   curl http://localhost:8080/api/users/123")
	fmt.Println("   curl -X POST http://localhost:8080/api/users -H 'Content-Type: application/json' -d '{\"name\":\"Alice\"}'")
	fmt.Println("   curl http://localhost:8080/api/health")
	fmt.Println()

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
