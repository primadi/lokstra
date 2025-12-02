package serviceapi

import "context"

// EmailMessage represents an email message to be sent
type EmailMessage struct {
	From        string   // Sender email address
	To          []string // Recipient email addresses
	Cc          []string // Carbon copy recipients (optional)
	Bcc         []string // Blind carbon copy recipients (optional)
	Subject     string   // Email subject
	Body        string   // Email body (plain text or HTML)
	HTMLBody    string   // HTML body (optional, if different from Body)
	Attachments []EmailAttachment
}

// EmailAttachment represents a file attachment
type EmailAttachment struct {
	Filename    string // Name of the file
	Content     []byte // File content
	ContentType string // MIME type (e.g., "application/pdf")
}

// EmailSender defines the interface for sending emails
type EmailSender interface {
	// Send sends an email message
	Send(ctx context.Context, message *EmailMessage) error

	// SendBatch sends multiple email messages
	SendBatch(ctx context.Context, messages []*EmailMessage) error
}
