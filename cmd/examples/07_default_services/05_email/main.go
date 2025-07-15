package main

import (
	"lokstra"
	"lokstra/services/email"
)

func main() {
	ctx := lokstra.NewGlobalContext()

	ctx.RegisterServiceModule(email.GetModule())

	app := lokstra.NewApp(ctx, "email-app", ":8080")

	app.POST("/send-email", func(ctx *lokstra.Context) error {
		var body map[string]any
		if err := ctx.BindJSON(&body); err != nil {
			return ctx.ErrorBadRequest("Invalid JSON body")
		}

		to, ok := body["to"].(string)
		if !ok {
			return ctx.ErrorBadRequest("To field is required")
		}

		subject, ok := body["subject"].(string)
		if !ok {
			return ctx.ErrorBadRequest("Subject field is required")
		}

		message, ok := body["message"].(string)
		if !ok {
			return ctx.ErrorBadRequest("Message field is required")
		}

		service, err := ctx.GetService("email")
		if err != nil {
			return ctx.ErrorInternal("Email service not available")
		}

		emailService := service.(*email.EmailService)
		
		err = emailService.SendPlainEmail([]string{to}, subject, message)
		if err != nil {
			return ctx.ErrorInternal("Failed to send email: " + err.Error())
		}

		return ctx.Ok(map[string]any{
			"message": "Email sent successfully",
			"to":      to,
			"subject": subject,
		})
	})

	app.POST("/send-html-email", func(ctx *lokstra.Context) error {
		var body map[string]any
		if err := ctx.BindJSON(&body); err != nil {
			return ctx.ErrorBadRequest("Invalid JSON body")
		}

		to, ok := body["to"].(string)
		if !ok {
			return ctx.ErrorBadRequest("To field is required")
		}

		subject, ok := body["subject"].(string)
		if !ok {
			return ctx.ErrorBadRequest("Subject field is required")
		}

		htmlBody, ok := body["html_body"].(string)
		if !ok {
			return ctx.ErrorBadRequest("HTML body field is required")
		}

		service, err := ctx.GetService("email")
		if err != nil {
			return ctx.ErrorInternal("Email service not available")
		}

		emailService := service.(*email.EmailService)
		
		err = emailService.SendHTMLEmail([]string{to}, subject, htmlBody)
		if err != nil {
			return ctx.ErrorInternal("Failed to send HTML email: " + err.Error())
		}

		return ctx.Ok(map[string]any{
			"message": "HTML email sent successfully",
			"to":      to,
			"subject": subject,
		})
	})

	lokstra.Logger.Infof("Email service example started on :8080")
	lokstra.Logger.Infof("Configure email service in YAML with SMTP settings")
	app.Start()
}
