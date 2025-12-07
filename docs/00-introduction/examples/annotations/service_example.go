package application

import (
	"time"
)

// Example: @Service annotation with @Inject and @InjectCfgValue

// @Service name="auth-service"
type AuthService struct {
	// @Inject "user-repository"
	UserRepo UserRepository

	// @Inject "cache-service"
	Cache CacheService

	// @InjectCfgValue "auth.jwt-secret"
	JwtSecret string

	// @InjectCfgValue key="auth.token-expiry", default="24h"
	TokenExpiry time.Duration

	// @InjectCfgValue key="auth.max-attempts", default=5
	MaxAttempts int

	// @InjectCfgValue key="auth.debug-mode", default="false"
	DebugMode bool
}

// Login authenticates user and returns JWT token
func (s *AuthService) Login(email, password string) (string, error) {
	// Check cache first if available
	if s.Cache != nil {
		if cachedUser, err := s.Cache.Get("user:" + email); err == nil {
			// Use cached user
			_ = cachedUser
		}
	}

	user, err := s.UserRepo.GetByEmail(email)
	if err != nil {
		return "", err
	}

	// Verify password (simplified)
	if user.Email != email {
		return "", nil
	}

	// Generate JWT token with configured expiry
	token := "jwt_token_for_" + user.ID + "_expires_in_" + s.TokenExpiry.String()

	// Cache user if cache is available
	if s.Cache != nil {
		_ = s.Cache.Set("user:"+email, user, s.TokenExpiry)
	}

	if s.DebugMode {
		println("Debug: Login successful for", email)
	}

	return token, nil
}

// @Service name="notification-service"
type NotificationService struct {
	// @InjectCfgValue "smtp.host"
	SMTPHost string

	// @InjectCfgValue key="smtp.port", default="587"
	SMTPPort int

	// @InjectCfgValue key="smtp.from-email", default="noreply@example.com"
	FromEmail string

	// @InjectCfgValue key="notification.enabled", default="true"
	Enabled bool
}

// SendEmail sends notification email
func (s *NotificationService) SendEmail(to, subject, body string) error {
	if !s.Enabled {
		return nil // Skip if disabled
	}

	// Send email via SMTP
	println("Sending email from", s.FromEmail, "to", to, "via", s.SMTPHost+":"+string(rune(s.SMTPPort)))
	return nil
}
