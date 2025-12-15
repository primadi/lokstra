# Email SMTP Service - Standalone Example

This example demonstrates how to use the Email SMTP service in a **standalone** application (without full deployment framework).

## Key Differences: Standalone vs Framework Mode

### Framework Mode (Production)

```go
func main() {
    lokstra.Bootstrap()
    
    // Load config
    lokstra_registry.LoadConfig("config.yaml")
    
    // Services auto-registered when server runs
    lokstra_registry.InitAndRunServer()
}
```

**Config (config.yaml):**
```yaml
configs:
  server: development.api

service-definitions:
  email_sender:
    type: email-smtp
    config:
      host: smtp.example.com
      # ...

deployments:
  development:
    servers:
      api:
        addr: ":8080"
        published-services: [email_sender]  # Auto-registered!
```

### Framework Mode (This Example - Recommended)

```go
func main() {
    lokstra.Bootstrap()
    
    // Register service types
    email_smtp.Register()
    lokstra_registry.RegisterServiceType("email-api-service", EmailAPIServiceFactory)
    
    // Load config and run (auto-registers services from deployments)
    lokstra_registry.LoadConfigFromFolder("configs")
    lokstra_registry.InitAndRunServer()
}
```

**Config (configs/email_smtp.yaml):**
```yaml
configs:
  server: development.api

service-definitions:
  email-sender:                   # Service instance name
    type: email-smtp              # Service type (from email_smtp.Register())
    config:
      host: smtp.example.com
      # ...

  email-api-service:              # Router service
    type: email-api-service
    depends-on:
      - EmailSender:email-sender  # Inject email-sender as EmailSender field

deployments:
  development:
    servers:
      api:
        addr: ":8080"
        routers: [email-router]
        published-services: [email-sender, email-api-service]  # Auto-registered!
```

**How it works:**
1. ✅ `LoadConfig()` loads YAML and stores service definitions
2. ✅ `InitAndRunServer()` reads `deployments.development.api.published-services`
3. ✅ Auto-registers `email-sender` and `email-api-service` with dependency injection
4. ✅ Mounts `email-router` to server
5. ✅ Server starts on `:8080`

## Why Manual Registration?

**Important:** `LoadConfig()` **DOES NOT** automatically register services. It only:
1. ✅ Loads YAML files
2. ✅ Stores config values
3. ✅ Stores service **definitions** (metadata)
4. ❌ Does NOT instantiate services

**Service registration happens during `RunServer()`** when deployment topology is analyzed.

For standalone apps (no deployment), you must **manually register** services.

## Quick Start

### 1. Start MailHog (for testing)

Using Docker:
```bash
docker run -p 1025:1025 -p 8025:8025 mailhog/mailhog
```

Or install and run directly:
```bash
# Install
go install github.com/mailhog/MailHog@latest

# Run
MailHog
```

MailHog provides:
- SMTP server on port 1025
- Web UI on http://localhost:8025

### 2. Run the Example

```bash
cd services/email_smtp/_example
go run example_standalone.go
```

The server will start on http://localhost:8080

### 3. Test the Endpoints

See [TEST_COMMANDS.md](TEST_COMMANDS.md) for curl commands.

Quick test:
```bash
curl -X POST http://localhost:8080/api/send-email \
  -H "Content-Type: application/json" \
  -d '{
    "to": "user@example.com",
    "subject": "Test Email",
    "message": "Hello from Lokstra!",
    "is_html": false
  }'
```

### 4. View Sent Emails

Open http://localhost:8025 in your browser to see all emails sent by the example.

## Available Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health` | Health check |
| POST | `/api/send-email` | Send simple email |
| POST | `/api/send-welcome` | Send welcome email (HTML) |
| POST | `/api/send-with-cc` | Send email with CC/BCC |
| POST | `/api/send-batch` | Send batch emails |
| POST | `/api/send-with-attachment` | Send email with attachment |

## Files

- `example_standalone.go` - Complete standalone example
- `TEST_COMMANDS.md` - Curl commands for testing
- `README.md` - This file

## Configuration

The example uses these default settings:
```go
Host:      "localhost"
Port:      1025          // MailHog default
FromEmail: "noreply@example.com"
FromName:  "Email Service Demo"
```

To use a real SMTP server, modify the configuration in `example_standalone.go`:
```go
emailSender := email_smtp.Service(&email_smtp.Config{
    Host:        "smtp.gmail.com",
    Port:        587,
    Username:    "your-email@gmail.com",
    Password:    "your-app-password",
    FromEmail:   "noreply@yourdomain.com",
    FromName:    "Your App Name",
    UseSTARTTLS: true,
    AuthMethod:  "plain",
})
```

## Troubleshooting

**Port already in use:**
```bash
# Find process using port 8080
netstat -ano | findstr :8080  # Windows
lsof -i :8080                 # Linux/Mac

# Kill the process or change the port in example_standalone.go
```

**MailHog not running:**
```bash
# Check if MailHog is running
docker ps | grep mailhog

# Or check the process
ps aux | grep MailHog
```

**Cannot connect to SMTP:**
- Ensure MailHog is running on port 1025
- Check firewall settings
- Try telnet: `telnet localhost 1025`

## Learn More

- [Main README](../README.md) - Full documentation
- [EXAMPLES.md](../EXAMPLES.md) - More usage examples
- [Lokstra Docs](https://primadi.github.io/lokstra/)
