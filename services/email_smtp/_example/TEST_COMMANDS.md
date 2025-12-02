# Email SMTP Service - Test Commands

This file contains curl commands to test the email SMTP service endpoints.

**Prerequisites:**
1. Start MailHog: `docker run -p 1025:1025 -p 8025:8025 mailhog/mailhog`
2. Run the example: `go run example_standalone.go`
3. View emails at: http://localhost:8025

## 1. Health Check

```bash
curl http://localhost:8080/api/health
```

## 2. Send Simple Email

```bash
curl -X POST http://localhost:8080/api/send-email \
  -H "Content-Type: application/json" \
  -d '{
    "to": "user@example.com",
    "subject": "Test Email",
    "message": "This is a test email from Lokstra Email Service",
    "is_html": false
  }'
```

## 3. Send HTML Email

```bash
curl -X POST http://localhost:8080/api/send-email \
  -H "Content-Type: application/json" \
  -d '{
    "to": "user@example.com",
    "subject": "HTML Test",
    "message": "<h1>Hello!</h1><p>This is an <strong>HTML</strong> email.</p>",
    "is_html": true
  }'
```

## 4. Send Welcome Email

```bash
curl -X POST http://localhost:8080/api/send-welcome \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "name": "John Doe"
  }'
```

## 5. Send Email with CC and BCC

```bash
curl -X POST http://localhost:8080/api/send-with-cc \
  -H "Content-Type: application/json" \
  -d '{
    "to": ["primary@example.com"],
    "cc": ["cc1@example.com", "cc2@example.com"],
    "bcc": ["bcc@example.com"],
    "subject": "Team Update",
    "message": "This is a team-wide announcement."
  }'
```

## 6. Send Batch Emails

```bash
curl -X POST http://localhost:8080/api/send-batch \
  -H "Content-Type: application/json" \
  -d '{
    "emails": [
      {
        "to": "user1@example.com",
        "subject": "Newsletter - Edition 1",
        "message": "Welcome to our first newsletter!"
      },
      {
        "to": "user2@example.com",
        "subject": "Newsletter - Edition 1",
        "message": "Welcome to our first newsletter!"
      },
      {
        "to": "user3@example.com",
        "subject": "Newsletter - Edition 1",
        "message": "Welcome to our first newsletter!"
      }
    ]
  }'
```

## 7. Send Email with Attachment

```bash
curl -X POST http://localhost:8080/api/send-with-attachment \
  -H "Content-Type: application/json" \
  -d '{
    "to": "user@example.com",
    "subject": "Document Attached",
    "message": "Please find the attached document."
  }'
```

## PowerShell Commands

For Windows PowerShell users:

### Send Simple Email
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/send-email" `
  -Method Post `
  -ContentType "application/json" `
  -Body (@{
    to = "user@example.com"
    subject = "Test Email"
    message = "This is a test email"
    is_html = $false
  } | ConvertTo-Json)
```

### Send Welcome Email
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/send-welcome" `
  -Method Post `
  -ContentType "application/json" `
  -Body (@{
    email = "newuser@example.com"
    name = "John Doe"
  } | ConvertTo-Json)
```

### Send Batch Emails
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/send-batch" `
  -Method Post `
  -ContentType "application/json" `
  -Body (@{
    emails = @(
      @{
        to = "user1@example.com"
        subject = "Newsletter"
        message = "Welcome to newsletter!"
      },
      @{
        to = "user2@example.com"
        subject = "Newsletter"
        message = "Welcome to newsletter!"
      }
    )
  } | ConvertTo-Json -Depth 3)
```

## Expected Response

Success response:
```json
{
  "status": "success",
  "message": "Email sent successfully"
}
```

Error response:
```json
{
  "error": "error message here"
}
```

## Verification

After sending emails, check the MailHog web UI at http://localhost:8025 to see the received emails.

## Troubleshooting

### Connection Refused
- Make sure MailHog is running: `docker ps`
- Check if port 1025 is available: `netstat -an | grep 1025` (Linux/Mac) or `netstat -an | findstr 1025` (Windows)

### No Emails Received
- Check MailHog web UI at http://localhost:8025
- Check the response from the API for errors
- Verify the email addresses in the request

### Invalid Request
- Make sure Content-Type is set to `application/json`
- Validate JSON syntax
- Check that required fields are provided
