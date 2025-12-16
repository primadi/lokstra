package main

// This is a standalone example showing how to use the email SMTP service
// Run with: go run example_standalone.go
// Make sure MailHog is running on localhost:1025 for testing

import (
	"fmt"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/core/deploy/loader"
	"github.com/primadi/lokstra/lokstra_init"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/services/email_smtp"
)

func main() {
	// lokstra.SetLogLevel(lokstra.LogLevelDebug)

	lokstra_init.Bootstrap()

	fmt.Println("\n===========================================")
	fmt.Println("Email SMTP Service Example")
	fmt.Println("===========================================")
	fmt.Println("\nAvailable endpoints:")
	fmt.Println("  POST /api/send-email           - Send simple email")
	fmt.Println("  POST /api/send-welcome         - Send welcome email (HTML)")
	fmt.Println("  POST /api/send-with-cc         - Send email with CC/BCC")
	fmt.Println("  POST /api/send-batch           - Send batch emails")
	fmt.Println("  POST /api/send-with-attachment - Send email with attachment")
	fmt.Println("  GET  /api/health               - Health check")
	fmt.Println("===========================================")

	// Load config and run server (auto-registers services from deployments)
	if _, err := loader.LoadConfig("configs"); err != nil {
		logger.LogPanic(err.Error())
	}

	// Register email_smtp service type
	email_smtp.Register()

	// Register lazy router factory (will be instantiated after runtime resolution)
	lokstra_registry.RegisterRouterFactory("email-router", func() lokstra.Router {
		emailService := lokstra_registry.GetService[*EmailService]("email-api-service")
		return emailService.GetRouter()
	})

	if err := lokstra_registry.RunConfiguredServer(); err != nil {
		logger.LogPanic(err.Error())
	}
}
