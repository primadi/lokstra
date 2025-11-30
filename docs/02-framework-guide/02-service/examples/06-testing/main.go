package main

import (
	"fmt"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/response"
	"github.com/primadi/lokstra/lokstra_registry"
)

// Demonstrates mock services for testing

// Service Interface
type EmailService interface {
	Send(to, subject, body string) error
}

// Real Implementation
type SMTPEmailService struct{}

func (*SMTPEmailService) Send(to, subject, body string) error {
	fmt.Printf("📧 Sending email to %s: %s\n", to, subject)
	return nil
}

// Mock Implementation (for testing)
type MockEmailService struct {
	calls []string
}

func NewMockEmailService() *MockEmailService {
	return &MockEmailService{calls: []string{}}
}

func (m *MockEmailService) Send(to, subject, body string) error {
	m.calls = append(m.calls, fmt.Sprintf("%s|%s", to, subject))
	fmt.Printf("✓ Mock: Email logged for %s\n", to)
	return nil
}

func (m *MockEmailService) GetCalls() []string {
	return m.calls
}

// Service that uses EmailService
type NotificationService struct {
	email EmailService
}

func NewNotificationService(email EmailService) *NotificationService {
	return &NotificationService{email: email}
}

func (s *NotificationService) NotifyUser(email, message string) error {
	return s.email.Send(email, "Notification", message)
}

// Factories
func SMTPEmailServiceFactory(deps map[string]any, config map[string]any) any {
	return &SMTPEmailService{}
}

func MockEmailServiceFactory(deps map[string]any, config map[string]any) any {
	return NewMockEmailService()
}

func NotificationServiceFactory(deps map[string]any, config map[string]any) any {
	email := lokstra_registry.GetService[EmailService]("email-service")
	return NewNotificationService(email)
}

// HTTP Handlers
type NotifyParams struct {
	Email   string `json:"email" validate:"required,email"`
	Message string `json:"message" validate:"required"`
}

func Notify(params *NotifyParams) *response.ApiHelper {
	svc := lokstra_registry.GetService[*NotificationService]("notification-service")
	err := svc.NotifyUser(params.Email, params.Message)
	if err != nil {
		return response.NewApiInternalError(err.Error())
	}
	return response.NewApiOk(map[string]any{
		"status": "sent",
		"email":  params.Email,
	})
}

func GetMockCalls() *response.ApiHelper {
	mock := lokstra_registry.GetService[*MockEmailService]("email-service")
	return response.NewApiOk(map[string]any{
		"calls": mock.GetCalls(),
		"count": len(mock.GetCalls()),
	})
}

func Home() *response.Response {
	return response.NewHtmlResponse(`<!DOCTYPE html>
<html><head><title>Service Testing</title></head>
<body><h1>🧪 Service Testing Example</h1>
<p>Demonstrates mock services for testing.</p>
<ul>
<li>POST /notify - Send notification (uses mock)</li>
<li>GET /mock-calls - Get mock email calls</li>
</ul>
</body></html>`)
}

func main() {
	// Use mock for testing
	lokstra_registry.RegisterLazyService("email-service", MockEmailServiceFactory, nil)
	lokstra_registry.RegisterLazyService("notification-service", NotificationServiceFactory, nil)

	router := lokstra.NewRouter("service-testing")
	router.GET("/", Home)
	router.POST("/notify", Notify)
	router.GET("/mock-calls", GetMockCalls)

	app := lokstra.NewApp("service-testing", ":3000", router)
	fmt.Println("🚀 Service Testing Example (with mocks)")
	fmt.Println("📍 http://localhost:3000")

	if err := app.Run(0); err != nil {
		fmt.Println("Error:", err)
	}
}
