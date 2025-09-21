package main

import (
	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/serviceapi"
)

// This example demonstrates Lokstra's registration context and dependency injection system.
// It shows how to register services, resolve dependencies, and manage application lifecycle.
//
// Learning Objectives:
// - Understand registration context for dependency injection
// - Learn how to register and retrieve services
// - See type-safe service container usage
// - Explore service lifecycle management
//
// Documentation: https://github.com/primadi/lokstra/blob/main/docs/core-concepts.md#registration-context

// Custom service interface and implementation
type EmailService interface {
	SendEmail(to, subject, body string) error
	GetServiceInfo() map[string]any
}

type SMTPEmailService struct {
	host     string
	port     int
	username string
}

func (s *SMTPEmailService) SendEmail(to, subject, body string) error {
	lokstra.Logger.Infof("📧 Sending email to %s: %s", to, subject)
	// In real implementation, this would send actual email
	return nil
}

func (s *SMTPEmailService) GetServiceInfo() map[string]any {
	return map[string]any{
		"type": "SMTP",
		"host": s.host,
		"port": s.port,
	}
}

// Factory function for creating EmailService
func NewSMTPEmailService(config map[string]any) (EmailService, error) {
	host := "localhost"
	port := 587
	username := ""

	if h, ok := config["host"].(string); ok {
		host = h
	}
	if p, ok := config["port"].(int); ok {
		port = p
	}
	if u, ok := config["username"].(string); ok {
		username = u
	}

	lokstra.Logger.Infof("🔧 Creating SMTP Email Service: %s:%d", host, port)

	return &SMTPEmailService{
		host:     host,
		port:     port,
		username: username,
	}, nil
}

// Another service that depends on EmailService
type NotificationService struct {
	emailService EmailService
	logger       serviceapi.Logger
}

func (n *NotificationService) SendWelcomeEmail(userEmail, userName string) error {
	subject := "Welcome to Our Platform!"
	body := "Hello " + userName + ", welcome to our amazing platform!"
	return n.emailService.SendEmail(userEmail, subject, body)
}

func (n *NotificationService) SendAlertEmail(adminEmail, alertMessage string) error {
	subject := "System Alert"
	return n.emailService.SendEmail(adminEmail, subject, alertMessage)
}

func NewNotificationService(regCtx lokstra.RegistrationContext) (*NotificationService, error) {
	// Resolve dependencies from registration context
	emailService, err := lokstra.GetService[EmailService](regCtx, "email")
	if err != nil {
		return nil, err
	}

	logger, err := lokstra.GetService[serviceapi.Logger](regCtx, "logger")
	if err != nil {
		return nil, err
	}

	lokstra.Logger.Infof("🔧 Creating Notification Service with dependencies")

	return &NotificationService{
		emailService: emailService,
		logger:       logger,
	}, nil
}

func main() {
	// Create registration context
	regCtx := lokstra.NewGlobalRegistrationContext()

	// ===== Service Registration =====

	// 1. Register Email Service factory
	regCtx.RegisterServiceFactory("smtp-email", func(config any) (any, error) {
		configMap := make(map[string]any)
		if cfg, ok := config.(map[string]any); ok {
			configMap = cfg
		}
		return NewSMTPEmailService(configMap)
	})

	// 2. Create and register Email Service instance
	emailConfig := map[string]any{
		"host":     "smtp.example.com",
		"port":     587,
		"username": "noreply@example.com",
	}

	_, err := regCtx.CreateService("smtp-email", "email", true, emailConfig)
	if err != nil {
		lokstra.Logger.Errorf("Failed to create email service: %v", err)
		return
	}
	lokstra.Logger.Infof("✅ Email service registered successfully")

	// 3. Register Notification Service factory
	regCtx.RegisterServiceFactory("notification", func(config any) (any, error) {
		return NewNotificationService(regCtx)
	})

	// 4. Create Notification Service instance
	_, err = regCtx.CreateService("notification", "notification", true)
	if err != nil {
		lokstra.Logger.Errorf("Failed to create notification service: %v", err)
		return
	}
	lokstra.Logger.Infof("✅ Notification service registered successfully")

	// ===== Service Usage in Application =====

	app := lokstra.NewApp(regCtx, "registration-context-app", ":8080")

	// Endpoint to test email service directly
	app.POST("/send-email", func(ctx *lokstra.Context) error {
		type EmailRequest struct {
			To      string `json:"to" validate:"required,email"`
			Subject string `json:"subject" validate:"required"`
			Body    string `json:"body" validate:"required"`
		}

		var req EmailRequest
		// In Lokstra, binding happens automatically via handler parameter
		// For this example, we'll simulate manual binding
		req.To = "user@example.com"
		req.Subject = "Test Email"
		req.Body = "This is a test email from Lokstra!"

		// Get email service from context
		emailSvc, err := lokstra.GetService[EmailService](regCtx, "email")
		if err != nil {
			return ctx.ErrorInternal("Email service not available")
		}

		if err := emailSvc.SendEmail(req.To, req.Subject, req.Body); err != nil {
			return ctx.ErrorInternal("Failed to send email")
		}

		return ctx.Ok(map[string]any{
			"message": "Email sent successfully",
			"to":      req.To,
		})
	})

	// Endpoint to test notification service (with dependency injection)
	app.POST("/send-welcome", func(ctx *lokstra.Context) error {
		type WelcomeRequest struct {
			Email string `json:"email" validate:"required,email"`
			Name  string `json:"name" validate:"required"`
		}

		var req WelcomeRequest
		req.Email = "newuser@example.com"
		req.Name = "John Doe"

		// Get notification service
		notifSvc, err := lokstra.GetService[*NotificationService](regCtx, "notification")
		if err != nil {
			return ctx.ErrorInternal("Notification service not available")
		}

		if err := notifSvc.SendWelcomeEmail(req.Email, req.Name); err != nil {
			return ctx.ErrorInternal("Failed to send welcome email")
		}

		return ctx.Ok(map[string]any{
			"message": "Welcome email sent successfully",
			"email":   req.Email,
			"name":    req.Name,
		})
	})

	// Endpoint to get service information
	app.GET("/services", func(ctx *lokstra.Context) error {
		services := make(map[string]any)

		// Get email service info
		if emailSvc, err := lokstra.GetService[EmailService](regCtx, "email"); err == nil {
			services["email"] = emailSvc.GetServiceInfo()
		}

		// Get logger info
		if _, err := lokstra.GetService[serviceapi.Logger](regCtx, "logger"); err == nil {
			services["logger"] = map[string]any{
				"type":   "default",
				"format": "json", // or whatever format is configured
			}
		}

		return ctx.Ok(map[string]any{
			"message":  "Available services",
			"services": services,
		})
	})

	// Health check that verifies all services
	app.GET("/health", func(ctx *lokstra.Context) error {
		health := map[string]any{
			"status": "healthy",
		}

		// Check email service
		if _, err := lokstra.GetService[EmailService](regCtx, "email"); err != nil {
			health["email_service"] = "unavailable"
			health["status"] = "degraded"
		} else {
			health["email_service"] = "available"
		}

		// Check notification service
		if _, err := lokstra.GetService[*NotificationService](regCtx, "notification"); err != nil {
			health["notification_service"] = "unavailable"
			health["status"] = "degraded"
		} else {
			health["notification_service"] = "available"
		}

		return ctx.Ok(health)
	})

	lokstra.Logger.Infof("🚀 Registration Context Example started on :8080")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Try these endpoints:")
	lokstra.Logger.Infof("  GET  /health         - Check service availability")
	lokstra.Logger.Infof("  GET  /services       - Get service information")
	lokstra.Logger.Infof("  POST /send-email     - Test email service directly")
	lokstra.Logger.Infof("  POST /send-welcome   - Test notification service (with DI)")

	app.Start(true)
}

// Registration Context Key Concepts:
//
// 1. Service Factory Registration:
//    - Register factory functions that create service instances
//    - Factories can accept configuration parameters
//    - Use regCtx.RegisterServiceFactory(name, factory)
//
// 2. Service Instance Creation:
//    - Create instances using registered factories
//    - Use regCtx.CreateService(factoryName, serviceName, config...)
//    - Service instances are singletons by default
//
// 3. Dependency Injection:
//    - Services can depend on other services
//    - Resolve dependencies using lokstra.GetService[T](regCtx, name)
//    - Type-safe service resolution with generics
//
// 4. Service Lifecycle:
//    - Services are created once and reused
//    - Global registration context manages all services
//    - Services can be retrieved anywhere in the application
//
// 5. Type Safety:
//    - Use interface types for service contracts
//    - Generic functions provide compile-time type checking
//    - Clear separation between interface and implementation

// Test Commands:
//
// curl http://localhost:8080/health
// curl http://localhost:8080/services
// curl -X POST http://localhost:8080/send-email
// curl -X POST http://localhost:8080/send-welcome
