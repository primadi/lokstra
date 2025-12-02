package email_smtp

import (
	"context"
	"testing"

	"github.com/primadi/lokstra/serviceapi"
)

// Unit tests run by default with: go test -v
//
// Integration tests (using mail.smtpbucket.com - free SMTP test server):
// Run with: go test -v -run TestSendEmail_Integration
// Run all integration tests: go test -v -run Integration
//
// Alternative: Use MailHog for local testing
// 1. Start MailHog: docker run -p 1025:1025 -p 8025:8025 mailhog/mailhog
// 2. Change Host to "localhost" and Port to 1025 in test configs
// 3. View emails at http://localhost:8025
//
// To debug integration tests in VS Code:
// - Set breakpoint in the test
// - Run "Debug Test" from the test function

func TestEmailSMTP_BuildEmail(t *testing.T) {
	cfg := &Config{
		Host:      "localhost",
		Port:      1025,
		FromEmail: "test@example.com",
		FromName:  "Test Sender",
	}

	service := Service(cfg)

	tests := []struct {
		name    string
		from    string
		message *serviceapi.EmailMessage
		wantErr bool
	}{
		{
			name: "simple text email",
			from: "Test <test@example.com>",
			message: &serviceapi.EmailMessage{
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
				Body:    "Test Body",
			},
			wantErr: false,
		},
		{
			name: "html email",
			from: "test@example.com",
			message: &serviceapi.EmailMessage{
				To:       []string{"recipient@example.com"},
				Subject:  "Test HTML",
				HTMLBody: "<html><body><h1>Test</h1></body></html>",
			},
			wantErr: false,
		},
		{
			name: "email with cc and bcc",
			from: "test@example.com",
			message: &serviceapi.EmailMessage{
				To:      []string{"to@example.com"},
				Cc:      []string{"cc@example.com"},
				Bcc:     []string{"bcc@example.com"},
				Subject: "Test CC/BCC",
				Body:    "Test Body",
			},
			wantErr: false,
		},
		{
			name: "email with attachment",
			from: "test@example.com",
			message: &serviceapi.EmailMessage{
				To:      []string{"recipient@example.com"},
				Subject: "Test Attachment",
				Body:    "See attachment",
				Attachments: []serviceapi.EmailAttachment{
					{
						Filename:    "test.txt",
						Content:     []byte("test content"),
						ContentType: "text/plain",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := service.buildEmail(tt.from, tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if body == nil {
				t.Error("buildEmail() returned nil body")
			}
		})
	}
}

func TestExtractEmail(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    string
	}{
		{
			name:    "email with name",
			address: "John Doe <john@example.com>",
			want:    "john@example.com",
		},
		{
			name:    "plain email",
			address: "john@example.com",
			want:    "john@example.com",
		},
		{
			name:    "email with unicode name",
			address: "José García <jose@example.com>",
			want:    "jose@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractEmail(tt.address); got != tt.want {
				t.Errorf("extractEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServiceFactory(t *testing.T) {
	params := map[string]any{
		"host":       "smtp.example.com",
		"port":       587,
		"username":   "user",
		"password":   "pass",
		"from_email": "noreply@example.com",
		"from_name":  "No Reply",
	}

	service := ServiceFactory(params)
	if service == nil {
		t.Error("ServiceFactory() returned nil")
	}

	emailSvc, ok := service.(*emailSMTP)
	if !ok {
		t.Error("ServiceFactory() did not return *emailSMTP")
	}

	if emailSvc.cfg.Host != "smtp.example.com" {
		t.Errorf("Host = %v, want %v", emailSvc.cfg.Host, "smtp.example.com")
	}

	if emailSvc.cfg.Port != 587 {
		t.Errorf("Port = %v, want %v", emailSvc.cfg.Port, 587)
	}
}

// TestSendEmail_Integration - integration test, requires actual SMTP server
// Skip by default, run with: go test -v -run TestSendEmail_Integration
func TestSendEmail_Integration(t *testing.T) {
	// t.Skip("Integration test - requires SMTP server (MailHog on localhost:1025)")

	cfg := &Config{
		Host:      "mail.smtpbucket.com",
		Port:      8025,
		FromEmail: "test@example.com",
		FromName:  "Test Sender",
	}

	sender := Service(cfg)

	msg := &serviceapi.EmailMessage{
		To:      []string{"recipient@example.com"},
		Subject: "Test Email",
		Body:    "This is a test email from integration test",
	}

	err := sender.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("Failed to send email: %v", err)
	}

	t.Log("Email sent successfully to mail.smtpbucket.com")
}

// TestSendHTMLEmail_Integration - test HTML email sending
func TestSendHTMLEmail_Integration(t *testing.T) {
	// t.Skip("Integration test - requires SMTP server (MailHog on localhost:1025)")

	cfg := &Config{
		Host:      "mail.smtpbucket.com",
		Port:      8025,
		FromEmail: "test@example.com",
		FromName:  "Test Sender",
	}

	sender := Service(cfg)

	msg := &serviceapi.EmailMessage{
		To:       []string{"recipient@example.com"},
		Subject:  "HTML Test Email",
		HTMLBody: "<html><body><h1>Hello!</h1><p>This is an <strong>HTML</strong> email.</p></body></html>",
	}

	err := sender.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("Failed to send HTML email: %v", err)
	}

	t.Log("HTML email sent successfully to mail.smtpbucket.com")
}

// TestSendEmailWithAttachment_Integration - test email with attachment
func TestSendEmailWithAttachment_Integration(t *testing.T) {
	// t.Skip("Integration test - requires SMTP server (MailHog on localhost:1025)")

	cfg := &Config{
		Host:      "mail.smtpbucket.com",
		Port:      8025,
		FromEmail: "test@example.com",
		FromName:  "Test Sender",
	}

	sender := Service(cfg)

	attachmentContent := []byte("This is a test attachment file.\nLine 2\nLine 3")

	msg := &serviceapi.EmailMessage{
		To:      []string{"recipient@example.com"},
		Subject: "Email with Attachment",
		Body:    "Please see the attached file.",
		Attachments: []serviceapi.EmailAttachment{
			{
				Filename:    "test-file.txt",
				Content:     attachmentContent,
				ContentType: "text/plain",
			},
		},
	}

	err := sender.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("Failed to send email with attachment: %v", err)
	}

	t.Log("Email with attachment sent successfully to mail.smtpbucket.com")
}

// TestSendBatchEmail_Integration - test batch email sending
func TestSendBatchEmail_Integration(t *testing.T) {
	// t.Skip("Integration test - requires SMTP server (MailHog on localhost:1025)")

	cfg := &Config{
		Host:      "mail.smtpbucket.com",
		Port:      8025,
		FromEmail: "test@example.com",
		FromName:  "Test Sender",
	}

	sender := Service(cfg)

	messages := []*serviceapi.EmailMessage{
		{
			To:      []string{"user1@example.com"},
			Subject: "Batch Email 1",
			Body:    "This is batch email 1",
		},
		{
			To:      []string{"user2@example.com"},
			Subject: "Batch Email 2",
			Body:    "This is batch email 2",
		},
		{
			To:      []string{"user3@example.com"},
			Subject: "Batch Email 3",
			Body:    "This is batch email 3",
		},
	}

	err := sender.SendBatch(context.Background(), messages)
	if err != nil {
		t.Fatalf("Failed to send batch emails: %v", err)
	}

	t.Logf("Sent %d emails successfully to mail.smtpbucket.com", len(messages))
}
