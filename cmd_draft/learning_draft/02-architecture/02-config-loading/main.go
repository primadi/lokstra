package main

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/config"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/lokstra_registry"
)

// =============================================================================
// CONFIG-DRIVEN ARCHITECTURE
// =============================================================================
// In this example, we:
// 1. Register service factories (code)
// 2. Define services in config.yaml (with env vars)
// 3. Load config and start server
//
// Benefits:
// - Environment-specific config without code changes
// - Easy deployment to different environments
// - Config versioning in git
// - Secrets in environment variables

// =============================================================================
// SERVICES (Same as before)
// =============================================================================

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

func EmailServiceFactory(cfg map[string]any) any {
	return &EmailService{
		SMTPHost: utils.GetValueFromMap(cfg, "smtp_host", "localhost"),
		SMTPPort: utils.GetValueFromMap(cfg, "smtp_port", 587),
		From:     utils.GetValueFromMap(cfg, "from", "noreply@example.com"),
	}
}

type CounterService struct {
	Name  string
	Count int
}

func (s *CounterService) Increment() int {
	s.Count++
	fmt.Printf("🔢 Counter [%s] incremented to: %d\n", s.Name, s.Count)
	return s.Count
}

func CounterServiceFactory(cfg map[string]any) any {
	return &CounterService{
		Name:  utils.GetValueFromMap(cfg, "name", "default-counter"),
		Count: utils.GetValueFromMap(cfg, "seed", 0),
	}
}

type LoggerService struct {
	Level  string
	Format string
	Output string
}

func (s *LoggerService) Log(message string) {
	fmt.Printf("📝 [%s] %s (format=%s, output=%s)\n", s.Level, message, s.Format, s.Output)
}

func LoggerServiceFactory(cfg map[string]any) any {
	return &LoggerService{
		Level:  utils.GetValueFromMap(cfg, "level", "info"),
		Format: utils.GetValueFromMap(cfg, "format", "json"),
		Output: utils.GetValueFromMap(cfg, "output", "stdout"),
	}
}

// =============================================================================
// SERVICE CONTAINER
// =============================================================================

type ServiceContainer struct {
	emailCache   *service.Cached[*EmailService]
	counterCache *service.Cached[*CounterService]
	loggerCache  *service.Cached[*LoggerService]
}

func (sc *ServiceContainer) GetEmail() *EmailService {
	return sc.emailCache.MustGet()
}

func (sc *ServiceContainer) GetCounter() *CounterService {
	return sc.counterCache.MustGet()
}

func (sc *ServiceContainer) GetLogger() *LoggerService {
	return sc.loggerCache.MustGet()
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
			"config": map[string]any{
				"smtp_host": email.SMTPHost,
				"smtp_port": email.SMTPPort,
				"from":      email.From,
			},
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

func createLoggerRouter() lokstra.Router {
	r := lokstra.NewRouter("logger-router")

	r.POST("/api/log", func(c *lokstra.RequestContext) error {
		var req struct {
			Message string `json:"message" binding:"required"`
		}

		if err := c.Req.BindBody(&req); err != nil {
			return c.Api.BadRequest("INVALID_INPUT", err.Error())
		}

		logger := services.GetLogger()
		logger.Log(req.Message)

		return c.Api.Ok(map[string]any{
			"status": "logged",
			"config": map[string]any{
				"level":  logger.Level,
				"format": logger.Format,
				"output": logger.Output,
			},
		})
	})

	return r
}

func createHealthRouter() lokstra.Router {
	r := lokstra.NewRouter("health-router")

	r.GET("/health", func(c *lokstra.RequestContext) error {
		// Get app environment from config
		appEnv := lokstra_registry.GetConfig("app-env", "unknown")

		return c.Api.Ok(map[string]any{
			"status":      "healthy",
			"environment": appEnv,
			"time":        time.Now().Format(time.RFC3339),
		})
	})

	return r
}

// =============================================================================
// SETUP REGISTRY
// =============================================================================

func setupRegistry() {
	fmt.Println("🔧 Setting up registry...")

	// Register service factories
	lokstra_registry.RegisterServiceFactory("email", EmailServiceFactory)
	lokstra_registry.RegisterServiceFactory("counter", CounterServiceFactory)
	lokstra_registry.RegisterServiceFactory("logger", LoggerServiceFactory)
	fmt.Println("   ✓ Registered service factories: email, counter, logger")

	// Register routers
	lokstra_registry.RegisterRouter("email-api", createEmailRouter())
	lokstra_registry.RegisterRouter("counter-api", createCounterRouter())
	lokstra_registry.RegisterRouter("logger-api", createLoggerRouter())
	lokstra_registry.RegisterRouter("health-api", createHealthRouter())
	fmt.Println("   ✓ Registered routers: email-api, counter-api, logger-api, health-api")

	fmt.Println()
}

// =============================================================================
// MAIN
// =============================================================================

func main() {
	fmt.Println("📚 Lokstra Learning: 02-Config Loading")
	fmt.Println("=========================================")
	fmt.Println()

	// Step 1: Setup registry (factories and routers)
	setupRegistry()

	// Step 2: Load configuration from config.yaml
	fmt.Println("📄 Loading configuration...")
	cfg := config.New()

	if err := config.LoadConfigFile("config.yaml", cfg); err != nil {
		fmt.Printf("❌ Failed to load config.yaml: %v\n", err)
		return
	}
	fmt.Println("   ✓ Loaded config.yaml")
	fmt.Println()

	// Step 3: Register config and set server name
	lokstra_registry.RegisterConfig(cfg, "")

	fmt.Println("🔍 Configuration loaded:")
	fmt.Printf("   Server: %s\n", lokstra_registry.GetCurrentServerName())
	fmt.Printf("   Environment: %s\n", lokstra_registry.GetConfig("app-env", "unknown"))
	fmt.Println()

	// Step 4: Start server (config-driven)
	fmt.Println("🚀 Starting config-driven server...")
	lokstra_registry.PrintServerStartInfo()
	if err := lokstra_registry.StartServer(); err != nil {
		fmt.Printf("❌ Server error: %v\n", err)
	}
}
