# Email SMTP Service

SMTP-based email sender service implementation for Lokstra framework.

## Features

- Send plain text and HTML emails
- Support for attachments
- Batch email sending
- Multiple authentication methods (PLAIN, LOGIN, CRAMMD5)
- TLS/STARTTLS support
- Configurable connection pooling
- CC and BCC support

## Configuration

### YAML Configuration

```yaml
service-definitions:
  email-service:
    type: email_smtp
    params:
      host: smtp.gmail.com
      port: 587
      username: your-email@gmail.com
      password: your-app-password
      from_email: noreply@example.com
      from_name: My Application
      use_starttls: true
      auth_method: plain
      pool_size: 10
      max_batch_size: 100
```

### Configuration Options

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `host` | string | `localhost` | SMTP server hostname |
| `port` | int | `587` | SMTP server port |
| `username` | string | - | SMTP username (if authentication required) |
| `password` | string | - | SMTP password |
| `from_email` | string | `noreply@example.com` | Default sender email address |
| `from_name` | string | - | Default sender name |
| `use_tls` | bool | `false` | Use direct TLS connection |
| `skip_verify` | bool | `false` | Skip TLS certificate verification |
| `use_starttls` | bool | `true` | Use STARTTLS for encryption |
| `auth_method` | string | `plain` | Authentication method: `plain`, `login`, `crammd5` |
| `pool_size` | int | `10` | Connection pool size |
| `max_batch_size` | int | `100` | Maximum batch size for SendBatch |

## Usage

### Basic Usage

```go
import (
    "context"
    "github.com/primadi/lokstra/serviceapi"
    "github.com/primadi/lokstra/core/service"
)

type MyService struct {
    EmailSender serviceapi.EmailSender
}

func (s *MyService) SendWelcomeEmail(userEmail, userName string) error {
    msg := &serviceapi.EmailMessage{
        To:      []string{userEmail},
        Subject: "Welcome to Our Platform",
        Body:    fmt.Sprintf("Hello %s, welcome aboard!", userName),
    }
    
    return s.EmailSender.Send(context.Background(), msg)
}
```

### With Annotation (@EndpointService)

```go
// @EndpointService name="notification-service"
type NotificationService struct {
    // @Inject "email-service"
    EmailSender serviceapi.EmailSender
}

// @Route "POST /send-notification"
func (s *NotificationService) SendNotification(p *NotificationParams) error {
    msg := &serviceapi.EmailMessage{
        To:      []string{p.Email},
        Subject: p.Subject,
        Body:    p.Message,
    }
    
    return s.EmailSender.Send(context.Background(), msg)
}
```

### HTML Email

```go
msg := &serviceapi.EmailMessage{
    To:      []string{"user@example.com"},
    Subject: "Welcome!",
    HTMLBody: `
        <html>
            <body>
                <h1>Welcome to Our Platform</h1>
                <p>Thank you for signing up!</p>
            </body>
        </html>
    `,
}

err := emailSender.Send(ctx, msg)
```

### Email with Attachments

```go
msg := &serviceapi.EmailMessage{
    To:      []string{"user@example.com"},
    Subject: "Your Invoice",
    Body:    "Please find your invoice attached.",
    Attachments: []serviceapi.EmailAttachment{
        {
            Filename:    "invoice.pdf",
            Content:     pdfBytes,
            ContentType: "application/pdf",
        },
    },
}

err := emailSender.Send(ctx, msg)
```

### Batch Sending

```go
messages := []*serviceapi.EmailMessage{
    {
        To:      []string{"user1@example.com"},
        Subject: "Newsletter",
        Body:    "Latest updates...",
    },
    {
        To:      []string{"user2@example.com"},
        Subject: "Newsletter",
        Body:    "Latest updates...",
    },
}

err := emailSender.SendBatch(ctx, messages)
```

### CC and BCC

```go
msg := &serviceapi.EmailMessage{
    To:      []string{"primary@example.com"},
    Cc:      []string{"cc@example.com"},
    Bcc:     []string{"bcc@example.com"},
    Subject: "Team Update",
    Body:    "Important team announcement...",
}

err := emailSender.Send(ctx, msg)
```

## Common SMTP Configurations

### Gmail

```yaml
params:
  host: smtp.gmail.com
  port: 587
  username: your-email@gmail.com
  password: your-app-password  # Use app password, not regular password
  use_starttls: true
  auth_method: plain
```

### Outlook/Office 365

```yaml
params:
  host: smtp.office365.com
  port: 587
  username: your-email@outlook.com
  password: your-password
  use_starttls: true
  auth_method: plain
```

### SendGrid

```yaml
params:
  host: smtp.sendgrid.net
  port: 587
  username: apikey
  password: your-sendgrid-api-key
  use_starttls: true
  auth_method: plain
```

### Mailgun

```yaml
params:
  host: smtp.mailgun.org
  port: 587
  username: postmaster@your-domain.mailgun.org
  password: your-mailgun-password
  use_starttls: true
  auth_method: plain
```

### Local SMTP Server (Development)

```yaml
params:
  host: localhost
  port: 1025
  use_starttls: false
  auth_method: plain
```

## Testing with MailHog

For local development, use [MailHog](https://github.com/mailhog/MailHog):

```bash
# Install MailHog
go install github.com/mailhog/MailHog@latest

# Run MailHog
MailHog
```

Configuration:

```yaml
params:
  host: localhost
  port: 1025
  use_starttls: false
```

Access web UI at http://localhost:8025

## Error Handling

```go
err := emailSender.Send(ctx, msg)
if err != nil {
    // Handle specific errors
    if strings.Contains(err.Error(), "authentication failed") {
        // Invalid credentials
    } else if strings.Contains(err.Error(), "connection refused") {
        // SMTP server unreachable
    } else if strings.Contains(err.Error(), "no recipients") {
        // Missing recipients
    }
    
    log.Printf("Failed to send email: %v", err)
    return err
}
```

## Best Practices

1. **Use App Passwords**: For Gmail and other providers, use app-specific passwords instead of account passwords
2. **Connection Pooling**: Configure appropriate `pool_size` based on your sending volume
3. **Batch Wisely**: Use `SendBatch` for multiple emails, but don't exceed provider limits
4. **HTML Sanitization**: Always sanitize user-generated HTML content
5. **Attachment Size**: Be mindful of attachment sizes (most SMTP servers limit to 10-25MB)
6. **Rate Limiting**: Implement rate limiting to avoid hitting provider limits
7. **Error Logging**: Always log email sending errors for debugging
8. **Testing**: Use MailHog or similar tools for local testing

## Troubleshooting

### Authentication Failed

- Check username and password
- For Gmail, enable "Less secure app access" or use app password
- Verify `auth_method` is correct for your provider

### Connection Refused

- Check `host` and `port` configuration
- Verify firewall allows outbound connections on SMTP port
- Ensure SMTP server is running

### TLS Errors

- Try toggling between `use_tls` and `use_starttls`
- For local/development servers, set `skip_verify: true`
- Ensure server supports the TLS version

### Emails Not Received

- Check spam/junk folders
- Verify sender email is not blacklisted
- Check SMTP server logs
- Ensure `from_email` domain has proper SPF/DKIM records
