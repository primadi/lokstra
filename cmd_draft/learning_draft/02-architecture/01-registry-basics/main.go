package main

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/lokstra_registry"
)

// =============================================================================
// REGISTRY BASICS
// =============================================================================
// The registry pattern allows you to:
// 1. Register service factories (code-based configuration)
// 2. Register routers with unique names
// 3. Create services lazily or eagerly
// 4. Access services anywhere in your application
//
// Key Components:
// - Service Factory: func(map[string]any) any - converts config to service
// - RegisterServiceFactory: Register factory with a type name
// - RegisterLazyService: Register service with config (created on first access)
// - RegisterRouter: Register router with unique name for auto-discovery

// =============================================================================
// SERVICES
// =============================================================================

// EmailService - Simple service with configuration
type EmailService struct {
	SMTPHost string
	SMTPPort int
	From     string
}

func (s *EmailService) SendEmail(to, subject, body string) error {
	fmt.Printf("📧 Sending email:\n")
	fmt.Printf("   From: %s\n", s.From)
	fmt.Printf("   To: %s\n", to)
	fmt.Printf("   Subject: %s\n", subject)
	fmt.Printf("   SMTP: %s:%d\n", s.SMTPHost, s.SMTPPort)
	return nil
}

// EmailServiceFactory creates EmailService from config
func EmailServiceFactory(cfg map[string]any) any {
	return &EmailService{
		SMTPHost: utils.GetValueFromMap(cfg, "smtp_host", "localhost"),
		SMTPPort: utils.GetValueFromMap(cfg, "smtp_port", 587),
		From:     utils.GetValueFromMap(cfg, "from", "noreply@example.com"),
	}
}

// CounterService - Stateful service example
type CounterService struct {
	Name  string
	Count int
}

func (s *CounterService) Increment() int {
	s.Count++
	fmt.Printf("🔢 Counter [%s] incremented to: %d\n", s.Name, s.Count)
	return s.Count
}

// CounterServiceFactory creates CounterService from config
func CounterServiceFactory(cfg map[string]any) any {
	return &CounterService{
		Name:  utils.GetValueFromMap(cfg, "name", "default-counter"),
		Count: utils.GetValueFromMap(cfg, "seed", 0),
	}
}

// =============================================================================
// SERVICE CONTAINER (Lazy Loading Pattern)
// =============================================================================

type ServiceContainer struct {
	emailCache   *EmailService
	counterCache *CounterService
}

func (sc *ServiceContainer) GetEmail() *EmailService {
	// GetService will create the service on first call, then cache it
	sc.emailCache = lokstra_registry.GetServiceCached("email-service", sc.emailCache)
	return sc.emailCache
}

func (sc *ServiceContainer) GetCounter() *CounterService {
	sc.counterCache = lokstra_registry.GetServiceCached("counter-service", sc.counterCache)
	return sc.counterCache
}

var services = &ServiceContainer{}

// =============================================================================
// ROUTERS
// =============================================================================

func createEmailRouter() lokstra.Router {
	r := lokstra.NewRouter("email-router")

	r.POST("/api/email/send", func(c *lokstra.RequestContext) error {
		var req struct {
			To      string `json:"to" binding:"required,email"`
			Subject string `json:"subject" binding:"required"`
			Body    string `json:"body" binding:"required"`
		}

		if err := c.Req.BindBody(&req); err != nil {
			return c.Api.BadRequest("INVALID_INPUT", err.Error())
		}

		email := services.GetEmail()
		if err := email.SendEmail(req.To, req.Subject, req.Body); err != nil {
			return c.Api.Error(500, "EMAIL_ERROR", err.Error())
		}

		return c.Api.Ok(map[string]any{
			"status":  "sent",
			"to":      req.To,
			"subject": req.Subject,
		})
	})

	return r
}

func createCounterRouter() lokstra.Router {
	r := lokstra.NewRouter("counter-router")

	r.POST("/api/counter/increment", func(c *lokstra.RequestContext) error {
		counter := services.GetCounter()
		newCount := counter.Increment()

		return c.Api.Ok(map[string]any{
			"counter": counter.Name,
			"count":   newCount,
		})
	})

	r.GET("/api/counter/value", func(c *lokstra.RequestContext) error {
		counter := services.GetCounter()

		return c.Api.Ok(map[string]any{
			"counter": counter.Name,
			"count":   counter.Count,
		})
	})

	return r
}

func createHealthRouter() lokstra.Router {
	r := lokstra.NewRouter("health-router")

	r.GET("/health", func(c *lokstra.RequestContext) error {
		return c.Api.Ok(map[string]any{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	return r
}

// =============================================================================
// REGISTRY SETUP
// =============================================================================

func setupRegistry() {
	fmt.Println("🔧 Setting up registry...")

	// Step 1: Register Service Factories
	// These are templates that know how to create services from config
	lokstra_registry.RegisterServiceFactory("email", EmailServiceFactory)
	lokstra_registry.RegisterServiceFactory("counter", CounterServiceFactory)
	fmt.Println("   ✓ Registered service factories: email, counter")

	// Step 2: Register Lazy Services
	// These services will be created only when first accessed
	lokstra_registry.RegisterLazyService("email-service", "email", map[string]any{
		"smtp_host": "smtp.gmail.com",
		"smtp_port": 587,
		"from":      "demo@lokstra.dev",
	})

	lokstra_registry.RegisterLazyService("counter-service", "counter", map[string]any{
		"name": "demo-counter",
		"seed": 100,
	})
	fmt.Println("   ✓ Registered lazy services: email-service, counter-service")

	// Step 3: Register Routers
	// These routers can be referenced by name in config.yaml for auto-discovery
	lokstra_registry.RegisterRouter("email-api", createEmailRouter())
	lokstra_registry.RegisterRouter("counter-api", createCounterRouter())
	lokstra_registry.RegisterRouter("health-api", createHealthRouter())
	fmt.Println("   ✓ Registered routers: email-api, counter-api, health-api")

	fmt.Println()
}

// =============================================================================
// MAIN
// =============================================================================

func main() {
	fmt.Println("📚 Lokstra Learning: 01-Registry Basics")
	fmt.Println("=========================================")
	fmt.Println()

	// Setup registry (factories, services, routers)
	setupRegistry()

	// Create app manually (without config.yaml)
	fmt.Println("🚀 Starting server...")

	emailRouter := lokstra_registry.GetRouter("email-api")
	counterRouter := lokstra_registry.GetRouter("counter-api")
	healthRouter := lokstra_registry.GetRouter("health-api")

	app := lokstra.NewApp("demo-app", ":8080", emailRouter, counterRouter, healthRouter)
	server := lokstra.NewServer("demo-server", app)

	server.PrintStartInfo()
	server.Run(5 * time.Second)
}
