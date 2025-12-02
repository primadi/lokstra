package email_smtp

import (
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net/smtp"
	"strings"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/serviceapi"
)

const SERVICE_TYPE = "email_smtp"

// Config represents the configuration for SMTP email service
type Config struct {
	Host         string `json:"host" yaml:"host"`                     // SMTP server host
	Port         int    `json:"port" yaml:"port"`                     // SMTP server port
	Username     string `json:"username" yaml:"username"`             // SMTP username
	Password     string `json:"password" yaml:"password"`             // SMTP password
	FromEmail    string `json:"from_email" yaml:"from_email"`         // Default from email
	FromName     string `json:"from_name" yaml:"from_name"`           // Default from name
	UseTLS       bool   `json:"use_tls" yaml:"use_tls"`               // Use TLS encryption
	SkipVerify   bool   `json:"skip_verify" yaml:"skip_verify"`       // Skip TLS certificate verification
	UseSTARTTLS  bool   `json:"use_starttls" yaml:"use_starttls"`     // Use STARTTLS
	AuthMethod   string `json:"auth_method" yaml:"auth_method"`       // Auth method: "plain", "login", "crammd5"
	PoolSize     int    `json:"pool_size" yaml:"pool_size"`           // Connection pool size
	MaxBatchSize int    `json:"max_batch_size" yaml:"max_batch_size"` // Maximum batch size
}

type emailSMTP struct {
	cfg *Config
}

var _ serviceapi.EmailSender = (*emailSMTP)(nil)

func (e *emailSMTP) Send(ctx context.Context, message *serviceapi.EmailMessage) error {
	// Use default from if not specified
	from := message.From
	if from == "" {
		if e.cfg.FromName != "" {
			from = fmt.Sprintf("%s <%s>", e.cfg.FromName, e.cfg.FromEmail)
		} else {
			from = e.cfg.FromEmail
		}
	}

	// Prepare recipients
	recipients := append([]string{}, message.To...)
	recipients = append(recipients, message.Cc...)
	recipients = append(recipients, message.Bcc...)

	if len(recipients) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	// Build email
	emailBody, err := e.buildEmail(from, message)
	if err != nil {
		return fmt.Errorf("failed to build email: %w", err)
	}

	// Send email
	return e.sendSMTP(from, recipients, emailBody)
}

func (e *emailSMTP) SendBatch(ctx context.Context, messages []*serviceapi.EmailMessage) error {
	if len(messages) == 0 {
		return nil
	}

	maxBatchSize := e.cfg.MaxBatchSize
	if maxBatchSize <= 0 {
		maxBatchSize = 100
	}

	// Process in batches
	for i := 0; i < len(messages); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(messages) {
			end = len(messages)
		}

		batch := messages[i:end]
		for _, msg := range batch {
			if err := e.Send(ctx, msg); err != nil {
				return fmt.Errorf("failed to send email in batch (index %d): %w", i, err)
			}
		}
	}

	return nil
}

func (e *emailSMTP) buildEmail(from string, message *serviceapi.EmailMessage) ([]byte, error) {
	var builder strings.Builder

	// Headers
	builder.WriteString(fmt.Sprintf("From: %s\r\n", from))
	builder.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(message.To, ", ")))

	if len(message.Cc) > 0 {
		builder.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(message.Cc, ", ")))
	}

	builder.WriteString(fmt.Sprintf("Subject: %s\r\n", message.Subject))
	builder.WriteString("MIME-Version: 1.0\r\n")

	// Check if we have attachments
	if len(message.Attachments) > 0 {
		boundary := "boundary-lokstra-email"
		builder.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n", boundary))

		// Body part
		builder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		if message.HTMLBody != "" {
			builder.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")
			builder.WriteString(message.HTMLBody)
		} else {
			builder.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
			builder.WriteString(message.Body)
		}
		builder.WriteString("\r\n")

		// Attachments
		for _, att := range message.Attachments {
			builder.WriteString(fmt.Sprintf("--%s\r\n", boundary))

			contentType := att.ContentType
			if contentType == "" {
				contentType = "application/octet-stream"
			}

			filename := mime.QEncoding.Encode("UTF-8", att.Filename)
			builder.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"\r\n", contentType, filename))
			builder.WriteString("Content-Transfer-Encoding: base64\r\n")
			builder.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n\r\n", filename))

			// Base64 encode attachment
			encoded := encodeBase64(att.Content)
			builder.WriteString(encoded)
			builder.WriteString("\r\n")
		}

		builder.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else {
		// No attachments - simple email
		if message.HTMLBody != "" {
			builder.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")
			builder.WriteString(message.HTMLBody)
		} else {
			builder.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
			builder.WriteString(message.Body)
		}
	}

	return []byte(builder.String()), nil
}

func (e *emailSMTP) sendSMTP(from string, recipients []string, body []byte) error {
	// Extract email from "Name <email>" format
	fromEmail := extractEmail(from)

	addr := fmt.Sprintf("%s:%d", e.cfg.Host, e.cfg.Port)

	var auth smtp.Auth
	if e.cfg.Username != "" {
		switch strings.ToLower(e.cfg.AuthMethod) {
		case "crammd5":
			auth = smtp.CRAMMD5Auth(e.cfg.Username, e.cfg.Password)
		case "login":
			auth = LoginAuth(e.cfg.Username, e.cfg.Password)
		default: // "plain"
			auth = smtp.PlainAuth("", e.cfg.Username, e.cfg.Password, e.cfg.Host)
		}
	}

	// Handle TLS
	if e.cfg.UseTLS {
		return e.sendWithTLS(addr, auth, fromEmail, recipients, body)
	} else if e.cfg.UseSTARTTLS {
		return e.sendWithSTARTTLS(addr, auth, fromEmail, recipients, body)
	}

	// Plain connection
	return smtp.SendMail(addr, auth, fromEmail, recipients, body)
}

func (e *emailSMTP) sendWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	tlsConfig := &tls.Config{
		ServerName:         e.cfg.Host,
		InsecureSkipVerify: e.cfg.SkipVerify,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to dial TLS: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, e.cfg.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", addr, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}

func (e *emailSMTP) sendWithSTARTTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	defer client.Close()

	if err = client.Hello("localhost"); err != nil {
		return fmt.Errorf("failed to say hello: %w", err)
	}

	tlsConfig := &tls.Config{
		ServerName:         e.cfg.Host,
		InsecureSkipVerify: e.cfg.SkipVerify,
	}

	if err = client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", addr, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}

// Helper functions

func extractEmail(address string) string {
	// Extract email from "Name <email@example.com>" format
	start := strings.Index(address, "<")
	end := strings.Index(address, ">")
	if start != -1 && end != -1 && end > start {
		return address[start+1 : end]
	}
	return address
}

func encodeBase64(data []byte) string {
	// Simple base64 encoding with line breaks
	const base64Table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var result strings.Builder

	for i := 0; i < len(data); i += 3 {
		b1 := data[i]
		var b2, b3 byte

		if i+1 < len(data) {
			b2 = data[i+1]
		}
		if i+2 < len(data) {
			b3 = data[i+2]
		}

		result.WriteByte(base64Table[b1>>2])
		result.WriteByte(base64Table[((b1&0x03)<<4)|(b2>>4)])

		if i+1 < len(data) {
			result.WriteByte(base64Table[((b2&0x0f)<<2)|(b3>>6)])
		} else {
			result.WriteByte('=')
		}

		if i+2 < len(data) {
			result.WriteByte(base64Table[b3&0x3f])
		} else {
			result.WriteByte('=')
		}

		// Line break every 76 characters
		if (i+3)%57 == 0 {
			result.WriteString("\r\n")
		}
	}

	return result.String()
}

// LoginAuth implements LOGIN authentication
type loginAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, fmt.Errorf("unknown fromServer")
		}
	}
	return nil, nil
}

// Service creates a new SMTP email sender service
func Service(cfg *Config) *emailSMTP {
	return &emailSMTP{cfg: cfg}
}

// ServiceFactory creates an email sender service from configuration map
func ServiceFactory(params map[string]any) any {
	cfg := &Config{
		Host:         utils.GetValueFromMap(params, "host", "localhost"),
		Port:         utils.GetValueFromMap(params, "port", 587),
		Username:     utils.GetValueFromMap(params, "username", ""),
		Password:     utils.GetValueFromMap(params, "password", ""),
		FromEmail:    utils.GetValueFromMap(params, "from_email", "noreply@example.com"),
		FromName:     utils.GetValueFromMap(params, "from_name", ""),
		UseTLS:       utils.GetValueFromMap(params, "use_tls", false),
		SkipVerify:   utils.GetValueFromMap(params, "skip_verify", false),
		UseSTARTTLS:  utils.GetValueFromMap(params, "use_starttls", true),
		AuthMethod:   utils.GetValueFromMap(params, "auth_method", "plain"),
		PoolSize:     utils.GetValueFromMap(params, "pool_size", 10),
		MaxBatchSize: utils.GetValueFromMap(params, "max_batch_size", 100),
	}
	return Service(cfg)
}

// Register registers the SMTP email service type
func Register() {
	lokstra_registry.RegisterServiceType(SERVICE_TYPE, ServiceFactory)
}
