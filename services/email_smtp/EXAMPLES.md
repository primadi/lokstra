# Email SMTP Service - Usage Examples

Complete examples of using the email SMTP service in Lokstra applications.

## Example 1: Simple Email Service with Router Mode

```go
package main

import (
    "context"
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
}

func main() {
    // Create email sender
    emailSender := email_smtp.Service(&email_smtp.Config{
        Host:      "localhost",
        Port:      1025,
        FromEmail: "noreply@example.com",
        FromName:  "My App",
    })

    // Create router
    r := lokstra.NewRouter("api")

    // Add email sending endpoint
    r.POST("/send-email", func(ctx *request.Context, req *SendEmailRequest) error {
        msg := &serviceapi.EmailMessage{
            To:      []string{req.To},
            Subject: req.Subject,
            Body:    req.Message,
        }

        err := emailSender.Send(context.Background(), msg)
        if err != nil {
            return ctx.Api.InternalServerError(err.Error())
        }

        return ctx.Api.Ok(map[string]string{
            "status": "Email sent successfully",
        })
    })

    // Run app
    app := lokstra.NewApp("email-app", ":8080", r)
    app.Run(30 * time.Second)
}
```

## Example 2: Framework Mode with Config

### config.yaml

```yaml
service-definitions:
  email-service:
    type: email_smtp
    params:
      host: smtp.gmail.com
      port: 587
      username: ${SMTP_USERNAME}
      password: ${SMTP_PASSWORD}
      from_email: noreply@myapp.com
      from_name: My Application
      use_starttls: true
      auth_method: plain

deployments:
  development:
    servers:
      api:
        addr: ":8080"
        published-services: [notification-service]
```

### main.go

```go
package main

import (
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/services"
    
    _ "myapp/application"
)

func main() {
    lokstra.Bootstrap()
    services.RegisterAllServices()
    _ = "myapp/application"
    lokstra_registry.RunServerFromConfig()
}
```

### application/notification_service.go

```go
package application

import (
    "context"
    "fmt"

    "github.com/primadi/lokstra/serviceapi"
)

type SendEmailParams struct {
    To      string `json:"to" validate:"required,email"`
    Subject string `json:"subject" validate:"required"`
    Message string `json:"message" validate:"required"`
}

// @Handler name="notification-service", prefix="/api/notifications"
type NotificationService struct {
    // @Inject "email-service"
    EmailSender serviceapi.EmailSender
}

// @Route "POST /send-email"
func (s *NotificationService) SendEmail(p *SendEmailParams) error {
    msg := &serviceapi.EmailMessage{
        To:      []string{p.To},
        Subject: p.Subject,
        Body:    p.Message,
    }

    return s.EmailSender.Send(context.Background(), msg)
}
```

## Example 3: Welcome Email Service

```go
package application

import (
    "context"
    "fmt"

    "github.com/primadi/lokstra/serviceapi"
)

// @Handler name="user-service", prefix="/api/users"
type UserService struct {
    // @Inject "email-service"
    EmailSender serviceapi.EmailSender
}

type CreateUserParams struct {
    Name     string `json:"name" validate:"required,min=3"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}

// @Route "POST /"
func (s *UserService) CreateUser(p *CreateUserParams) (map[string]any, error) {
    // Create user in database...
    // (omitted for brevity)

    // Send welcome email
    welcomeHTML := fmt.Sprintf(`
        <html>
            <body>
                <h1>Welcome, %s!</h1>
                <p>Thank you for joining our platform.</p>
                <p>Your account has been successfully created.</p>
            </body>
        </html>
    `, p.Name)

    msg := &serviceapi.EmailMessage{
        To:       []string{p.Email},
        Subject:  "Welcome to Our Platform",
        HTMLBody: welcomeHTML,
    }

    err := s.EmailSender.Send(context.Background(), msg)
    if err != nil {
        // Log error but don't fail user creation
        fmt.Printf("Failed to send welcome email: %v\n", err)
    }

    return map[string]any{
        "status": "User created successfully",
        "email":  p.Email,
    }, nil
}
```

## Example 4: Email with Attachments

```go
package application

import (
    "context"
    "fmt"
    "os"

    "github.com/primadi/lokstra/serviceapi"
)

// @Handler name="invoice-service", prefix="/api/invoices"
type InvoiceService struct {
    // @Inject "email-service"
    EmailSender serviceapi.EmailSender
}

type SendInvoiceParams struct {
    Email     string `json:"email" validate:"required,email"`
    InvoiceID string `json:"invoice_id" validate:"required"`
}

// @Route "POST /send"
func (s *InvoiceService) SendInvoice(p *SendInvoiceParams) error {
    // Generate PDF invoice
    pdfBytes, err := s.generateInvoicePDF(p.InvoiceID)
    if err != nil {
        return fmt.Errorf("failed to generate PDF: %w", err)
    }

    msg := &serviceapi.EmailMessage{
        To:      []string{p.Email},
        Subject: fmt.Sprintf("Invoice #%s", p.InvoiceID),
        HTMLBody: `
            <html>
                <body>
                    <h2>Your Invoice</h2>
                    <p>Please find your invoice attached to this email.</p>
                    <p>Thank you for your business!</p>
                </body>
            </html>
        `,
        Attachments: []serviceapi.EmailAttachment{
            {
                Filename:    fmt.Sprintf("invoice-%s.pdf", p.InvoiceID),
                Content:     pdfBytes,
                ContentType: "application/pdf",
            },
        },
    }

    return s.EmailSender.Send(context.Background(), msg)
}

func (s *InvoiceService) generateInvoicePDF(invoiceID string) ([]byte, error) {
    // Generate PDF logic here
    return []byte("PDF content"), nil
}
```

## Example 5: Bulk Email Sending

```go
package application

import (
    "context"

    "github.com/primadi/lokstra/serviceapi"
)

// @Handler name="newsletter-service", prefix="/api/newsletter"
type NewsletterService struct {
    // @Inject "email-service"
    EmailSender serviceapi.EmailSender
}

type SendNewsletterParams struct {
    Subject string   `json:"subject" validate:"required"`
    Content string   `json:"content" validate:"required"`
    ToEmails []string `json:"to_emails" validate:"required,min=1"`
}

// @Route "POST /send"
func (s *NewsletterService) SendNewsletter(p *SendNewsletterParams) (map[string]any, error) {
    // Build messages
    messages := make([]*serviceapi.EmailMessage, len(p.ToEmails))
    for i, email := range p.ToEmails {
        messages[i] = &serviceapi.EmailMessage{
            To:       []string{email},
            Subject:  p.Subject,
            HTMLBody: p.Content,
        }
    }

    // Send batch
    err := s.EmailSender.SendBatch(context.Background(), messages)
    if err != nil {
        return nil, err
    }

    return map[string]any{
        "status":      "Newsletter sent",
        "total_sent":  len(messages),
    }, nil
}
```

## Example 6: Password Reset Email

```go
package application

import (
    "context"
    "fmt"
    "time"

    "github.com/primadi/lokstra/serviceapi"
)

// @Handler name="auth-service", prefix="/api/auth"
type AuthService struct {
    // @Inject "email-service"
    EmailSender serviceapi.EmailSender
}

type ForgotPasswordParams struct {
    Email string `json:"email" validate:"required,email"`
}

// @Route "POST /forgot-password"
func (s *AuthService) ForgotPassword(p *ForgotPasswordParams) (map[string]any, error) {
    // Generate reset token
    resetToken := s.generateResetToken(p.Email)
    
    // Build reset URL
    resetURL := fmt.Sprintf("https://myapp.com/reset-password?token=%s", resetToken)
    
    emailBody := fmt.Sprintf(`
        <html>
            <body>
                <h2>Password Reset Request</h2>
                <p>You requested to reset your password.</p>
                <p>Click the link below to reset your password:</p>
                <p><a href="%s">Reset Password</a></p>
                <p>This link will expire in 1 hour.</p>
                <p>If you didn't request this, please ignore this email.</p>
            </body>
        </html>
    `, resetURL)

    msg := &serviceapi.EmailMessage{
        To:       []string{p.Email},
        Subject:  "Password Reset Request",
        HTMLBody: emailBody,
    }

    err := s.EmailSender.Send(context.Background(), msg)
    if err != nil {
        return nil, fmt.Errorf("failed to send reset email: %w", err)
    }

    return map[string]any{
        "status": "Password reset email sent",
    }, nil
}

func (s *AuthService) generateResetToken(email string) string {
    // Generate secure token logic
    return "sample-token-" + fmt.Sprint(time.Now().Unix())
}
```

## Example 7: Email Template System

```go
package application

import (
    "bytes"
    "context"
    "html/template"

    "github.com/primadi/lokstra/serviceapi"
)

// @Handler name="template-email-service", prefix="/api/emails"
type TemplateEmailService struct {
    // @Inject "email-service"
    EmailSender serviceapi.EmailSender
    
    templates map[string]*template.Template
}

type SendTemplateEmailParams struct {
    Template string         `json:"template" validate:"required"`
    To       string         `json:"to" validate:"required,email"`
    Subject  string         `json:"subject" validate:"required"`
    Data     map[string]any `json:"data"`
}

func (s *TemplateEmailService) Init() {
    // Initialize templates
    s.templates = map[string]*template.Template{
        "welcome": template.Must(template.New("welcome").Parse(`
            <html>
                <body>
                    <h1>Welcome, {{.Name}}!</h1>
                    <p>Thank you for joining {{.AppName}}.</p>
                </body>
            </html>
        `)),
        "order_confirmation": template.Must(template.New("order").Parse(`
            <html>
                <body>
                    <h2>Order Confirmation</h2>
                    <p>Order #{{.OrderID}}</p>
                    <p>Total: ${{.Total}}</p>
                </body>
            </html>
        `)),
    }
}

// @Route "POST /send-template"
func (s *TemplateEmailService) SendTemplateEmail(p *SendTemplateEmailParams) error {
    tmpl, exists := s.templates[p.Template]
    if !exists {
        return fmt.Errorf("template not found: %s", p.Template)
    }

    var body bytes.Buffer
    if err := tmpl.Execute(&body, p.Data); err != nil {
        return fmt.Errorf("failed to execute template: %w", err)
    }

    msg := &serviceapi.EmailMessage{
        To:       []string{p.To},
        Subject:  p.Subject,
        HTMLBody: body.String(),
    }

    return s.EmailSender.Send(context.Background(), msg)
}
```

## Running the Examples

### With MailHog (Local Testing)

1. Install and run MailHog:
   ```bash
   go install github.com/mailhog/MailHog@latest
   MailHog
   ```

2. Access web UI at http://localhost:8025

3. Use configuration:
   ```yaml
   params:
     host: localhost
     port: 1025
   ```

### With Real SMTP Server

Update `config.yaml` with your SMTP credentials:

```yaml
service-definitions:
  email-service:
    type: email_smtp
    params:
      host: ${SMTP_HOST}
      port: 587
      username: ${SMTP_USERNAME}
      password: ${SMTP_PASSWORD}
      from_email: noreply@yourdomain.com
      from_name: Your App Name
      use_starttls: true
```

Set environment variables:
```bash
export SMTP_HOST="smtp.gmail.com"
export SMTP_USERNAME="your-email@gmail.com"
export SMTP_PASSWORD="your-app-password"
```
