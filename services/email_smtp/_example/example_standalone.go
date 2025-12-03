package main

// This is a standalone example showing how to use the email SMTP service
// Run with: go run example_standalone.go
// Make sure MailHog is running on localhost:1025 for testing

import (
	"context"
	"fmt"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/services/email_smtp"
)

type SendEmailRequest struct {
	To      string `json:"to" validate:"required,email"`
	Subject string `json:"subject" validate:"required"`
	Message string `json:"message" validate:"required"`
	IsHTML  bool   `json:"is_html"`
}

type SendBatchRequest struct {
	Emails []struct {
		To      string `json:"to" validate:"required,email"`
		Subject string `json:"subject" validate:"required"`
		Message string `json:"message" validate:"required"`
	} `json:"emails" validate:"required,min=1"`
}

func main() {
	fmt.Println("Starting Email SMTP Service Example...")

	// Create email sender service
	// For testing, use MailHog: docker run -p 1025:1025 -p 8025:8025 mailhog/mailhog
	emailSender := email_smtp.Service(&email_smtp.Config{
		Host:      "localhost",
		Port:      1025, // MailHog SMTP port
		FromEmail: "noreply@example.com",
		FromName:  "Email Service Demo",
	})

	// Create router
	r := lokstra.NewRouter("api")

	// Endpoint: Send single email
	r.POST("/send-email", func(ctx *request.Context, req *SendEmailRequest) error {
		msg := &serviceapi.EmailMessage{
			To:      []string{req.To},
			Subject: req.Subject,
		}

		if req.IsHTML {
			msg.HTMLBody = req.Message
		} else {
			msg.Body = req.Message
		}

		err := emailSender.Send(context.Background(), msg)
		if err != nil {
			return ctx.Api.InternalServerError(err.Error())
		}

		return ctx.Api.Ok(map[string]string{
			"status":  "success",
			"message": "Email sent successfully",
		})
	})

	// Endpoint: Send welcome email (with HTML)
	r.POST("/send-welcome", func(ctx *request.Context, req struct {
		Email string `json:"email" validate:"required,email"`
		Name  string `json:"name" validate:"required"`
	}) error {
		htmlBody := fmt.Sprintf(`
			<html>
				<head>
					<style>
						body { font-family: Arial, sans-serif; }
						.container { max-width: 600px; margin: 0 auto; padding: 20px; }
						.header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; }
						.content { padding: 20px; background-color: #f9f9f9; }
						.footer { text-align: center; padding: 10px; color: #666; font-size: 12px; }
					</style>
				</head>
				<body>
					<div class="container">
						<div class="header">
							<h1>Welcome to Our Platform!</h1>
						</div>
						<div class="content">
							<h2>Hello %s,</h2>
							<p>Thank you for joining our platform. We're excited to have you on board!</p>
							<p>Here are some things you can do:</p>
							<ul>
								<li>Complete your profile</li>
								<li>Explore our features</li>
								<li>Connect with other users</li>
							</ul>
							<p>If you have any questions, feel free to reach out to our support team.</p>
						</div>
						<div class="footer">
							<p>© 2024 Email Service Demo. All rights reserved.</p>
						</div>
					</div>
				</body>
			</html>
		`, req.Name)

		msg := &serviceapi.EmailMessage{
			To:       []string{req.Email},
			Subject:  "Welcome to Our Platform!",
			HTMLBody: htmlBody,
		}

		err := emailSender.Send(context.Background(), msg)
		if err != nil {
			return ctx.Api.InternalServerError(err.Error())
		}

		return ctx.Api.Ok(map[string]any{
			"status":  "success",
			"message": "Welcome email sent successfully",
			"to":      req.Email,
		})
	})

	// Endpoint: Send email with CC and BCC
	r.POST("/send-with-cc", func(ctx *request.Context, req struct {
		To      []string `json:"to" validate:"required,min=1"`
		Cc      []string `json:"cc"`
		Bcc     []string `json:"bcc"`
		Subject string   `json:"subject" validate:"required"`
		Message string   `json:"message" validate:"required"`
	}) error {
		msg := &serviceapi.EmailMessage{
			To:      req.To,
			Cc:      req.Cc,
			Bcc:     req.Bcc,
			Subject: req.Subject,
			Body:    req.Message,
		}

		err := emailSender.Send(context.Background(), msg)
		if err != nil {
			return ctx.Api.InternalServerError(err.Error())
		}

		return ctx.Api.Ok(map[string]any{
			"status":    "success",
			"message":   "Email sent with CC/BCC",
			"total_to":  len(req.To),
			"total_cc":  len(req.Cc),
			"total_bcc": len(req.Bcc),
		})
	})

	// Endpoint: Send batch emails
	r.POST("/send-batch", func(ctx *request.Context, req *SendBatchRequest) error {
		messages := make([]*serviceapi.EmailMessage, len(req.Emails))
		for i, email := range req.Emails {
			messages[i] = &serviceapi.EmailMessage{
				To:      []string{email.To},
				Subject: email.Subject,
				Body:    email.Message,
			}
		}

		err := emailSender.SendBatch(context.Background(), messages)
		if err != nil {
			return ctx.Api.InternalServerError(err.Error())
		}

		return ctx.Api.Ok(map[string]any{
			"status":     "success",
			"message":    "Batch emails sent successfully",
			"total_sent": len(messages),
		})
	})

	// Endpoint: Send email with attachment (simulated)
	r.POST("/send-with-attachment", func(ctx *request.Context, req struct {
		To      string `json:"to" validate:"required,email"`
		Subject string `json:"subject" validate:"required"`
		Message string `json:"message" validate:"required"`
	}) error {
		// Simulate a text file attachment
		attachmentContent := []byte("This is a sample attachment file.\nCreated by Email Service Demo.\n")

		msg := &serviceapi.EmailMessage{
			To:      []string{req.To},
			Subject: req.Subject,
			Body:    req.Message,
			Attachments: []serviceapi.EmailAttachment{
				{
					Filename:    "sample.txt",
					Content:     attachmentContent,
					ContentType: "text/plain",
				},
			},
		}

		err := emailSender.Send(context.Background(), msg)
		if err != nil {
			return ctx.Api.InternalServerError(err.Error())
		}

		return ctx.Api.Ok(map[string]any{
			"status":          "success",
			"message":         "Email sent with attachment",
			"attachment_name": "sample.txt",
		})
	})

	// Health check endpoint
	r.GET("/health", func() map[string]string {
		return map[string]string{
			"status":  "healthy",
			"service": "email-smtp-demo",
		}
	})

	// Create and run app
	app := lokstra.NewApp("email-demo", ":8080", r)

	fmt.Println("\n===========================================")
	fmt.Println("Email SMTP Service is running on :8080")
	fmt.Println("===========================================")
	fmt.Println("\nAvailable endpoints:")
	fmt.Println("  POST /api/send-email          - Send simple email")
	fmt.Println("  POST /api/send-welcome        - Send welcome email (HTML)")
	fmt.Println("  POST /api/send-with-cc        - Send email with CC/BCC")
	fmt.Println("  POST /api/send-batch          - Send batch emails")
	fmt.Println("  POST /api/send-with-attachment - Send email with attachment")
	fmt.Println("  GET  /api/health              - Health check")
	fmt.Println("\nNote: Make sure MailHog is running on localhost:1025")
	fmt.Println("View emails at: http://localhost:8025")
	fmt.Println("===========================================\n")

	app.Run(30 * time.Second)
}
