package email

import (
	"fmt"
	"lokstra/common/iface"
	"net/smtp"
	"strings"
)

type EmailService struct {
	instanceName string
	config       map[string]any
	smtpHost     string
	smtpPort     string
	username     string
	password     string
	from         string
}

type EmailMessage struct {
	To      []string
	Subject string
	Body    string
	IsHTML  bool
}

func (e *EmailService) InstanceName() string {
	return e.instanceName
}

func (e *EmailService) GetConfig(key string) any {
	return e.config[key]
}

func (e *EmailService) SendEmail(msg *EmailMessage) error {
	auth := smtp.PlainAuth("", e.username, e.password, e.smtpHost)
	
	to := strings.Join(msg.To, ",")
	contentType := "text/plain"
	if msg.IsHTML {
		contentType = "text/html"
	}

	body := fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: %s; charset=UTF-8\r\n\r\n%s",
		to, msg.Subject, contentType, msg.Body)

	addr := fmt.Sprintf("%s:%s", e.smtpHost, e.smtpPort)
	return smtp.SendMail(addr, auth, e.from, msg.To, []byte(body))
}

func (e *EmailService) SendPlainEmail(to []string, subject, body string) error {
	return e.SendEmail(&EmailMessage{
		To:      to,
		Subject: subject,
		Body:    body,
		IsHTML:  false,
	})
}

func (e *EmailService) SendHTMLEmail(to []string, subject, body string) error {
	return e.SendEmail(&EmailMessage{
		To:      to,
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	})
}

func newEmailService(instanceName string, config map[string]any) (*EmailService, error) {
	smtpHost, ok := config["smtp_host"].(string)
	if !ok {
		return nil, fmt.Errorf("email service requires 'smtp_host' in config")
	}

	smtpPort, ok := config["smtp_port"].(string)
	if !ok {
		smtpPort = "587"
	}

	username, ok := config["username"].(string)
	if !ok {
		return nil, fmt.Errorf("email service requires 'username' in config")
	}

	password, ok := config["password"].(string)
	if !ok {
		return nil, fmt.Errorf("email service requires 'password' in config")
	}

	from, ok := config["from"].(string)
	if !ok {
		from = username
	}

	return &EmailService{
		instanceName: instanceName,
		config:       config,
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		username:     username,
		password:     password,
		from:         from,
	}, nil
}
