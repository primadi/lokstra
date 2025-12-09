package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/serviceapi"
)

// EmailService provides email API endpoints
// @Service "email-api-service"
type EmailService struct {
	// @Inject "email_smtp"
	EmailSender serviceapi.EmailSender
}

type SendEmailRequest struct {
	To      string `json:"to" validate:"required,email"`
	Subject string `json:"subject" validate:"required"`
	Message string `json:"message" validate:"required"`
	IsHTML  bool   `json:"is_html"`
}

type BatchEmailItem struct {
	To      string `json:"to" validate:"required,email"`
	Subject string `json:"subject" validate:"required"`
	Message string `json:"message" validate:"required"`
	IsHTML  bool   `json:"is_html"`
}

type SendBatchRequest struct {
	Emails []BatchEmailItem `json:"emails" validate:"required,min=1,dive"`
}

// GetRouter returns router with all email endpoints
func (s *EmailService) GetRouter() lokstra.Router {
	r := lokstra.NewRouter("email-api")

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

		err := s.EmailSender.Send(context.Background(), msg)
		if err != nil {
			return ctx.Api.InternalError(err.Error())
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

		err := s.EmailSender.Send(context.Background(), msg)
		if err != nil {
			return ctx.Api.InternalError(err.Error())
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

		err := s.EmailSender.Send(context.Background(), msg)
		if err != nil {
			return ctx.Api.InternalError(err.Error())
		}

		return ctx.Api.Ok(map[string]any{
			"status":    "success",
			"message":   "Email sent with CC/BCC",
			"total_to":  len(req.To),
			"total_cc":  len(req.Cc),
			"total_bcc": len(req.Bcc),
		})
	})

	// Endpoint: Send batch emails (with goroutine for better performance)
	r.POST("/send-batch", func(ctx *request.Context, req *SendBatchRequest) error {
		total := len(req.Emails)

		// Send batch emails in goroutine for faster response
		go func() {
			messages := make([]*serviceapi.EmailMessage, total)
			for i, email := range req.Emails {
				msg := &serviceapi.EmailMessage{
					To:      []string{email.To},
					Subject: email.Subject,
				}

				if email.IsHTML {
					msg.HTMLBody = email.Message
				} else {
					msg.Body = email.Message
				}

				messages[i] = msg
			}

			if err := s.EmailSender.SendBatch(context.Background(), messages); err != nil {
				fmt.Printf("❌ Batch email failed: %v\n", err)
			} else {
				fmt.Printf("✅ Successfully sent %d batch emails\n", total)
			}
		}()

		// Return 202 Accepted for async operation
		return ctx.Resp.WithStatus(http.StatusAccepted).Json(map[string]any{
			"status":  "queued",
			"message": fmt.Sprintf("%d emails queued for sending", total),
			"total":   total,
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

		err := s.EmailSender.Send(context.Background(), msg)
		if err != nil {
			return ctx.Api.InternalError(err.Error())
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

	r.GET("/info", func() map[string]string {
		return map[string]string{
			"host":    lokstra_registry.GetConfig("email_smtp.host", ""),
			"version": "1.0.0",
		}
	})

	return r
}
