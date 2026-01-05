package annotation

// This file contains example code demonstrating the @Inject annotation
// DO NOT DELETE - Used as reference for testing

/* Example 1: @Inject with cfg: prefix for config values

// @RouterService name="auth-service", prefix="/api/auth"
type AuthService struct {
	// @Inject "cfg:app.timeout"
	Timeout time.Duration

	// @Inject "cfg:app.max-attempts"
	MaxAttempts int

	// @Inject "user-repository"
	UserRepo UserRepository
}

config.yaml:
  configs:
    app:
      timeout: "30s"
      max-attempts: 5
*/

/* Example 2: @Inject with cfg:@ prefix (indirect config)

// @RouterService name="auth-service", prefix="/api/auth"
type AuthService struct {
	// @Inject "cfg:@jwt.key-path"
	JWTSecret string

	// @Inject "cfg:@db.timeout-key"
	DBTimeout time.Duration
}

config.yaml:
  configs:
    jwt:
      key-path: "app.production-jwt-secret"  # Points to actual config key

    db:
      timeout-key: "database.connection-timeout"

    app:
      production-jwt-secret: "super-secret-key-xyz"

    database:
      connection-timeout: "15s"
*/

/* Example 3: Config value injection with indirection

// @Service "config-service"
type ConfigService struct {
	// @Inject "cfg:@jwt.key-path"
	JWTSecret string

	// @Inject "cfg:app.name"
	AppName string
}

config.yaml:
  configs:
    jwt:
      key-path: "security.jwt-secret"

    security:
      jwt-secret: "my-secret-key"

    app:
      name: "MyApp"
*/

/* Example 4: Mixed injection patterns

// @RouterService name="user-service", prefix="/api/users"
type UserService struct {
	// Direct service injection
	// @Inject "user-repository"
	UserRepo UserRepository

	// Service from config
	// @Inject "@store.implementation"
	Store Store

	// Direct config value
	// @Inject "cfg:app.page-size"
	PageSize int

	// Indirect config value
	// @Inject "cfg:@cache.ttl-key"
	CacheTTL time.Duration

	// Config value
	// @Inject "cfg:app.name"
	AppName string

	// Indirect config
	// @Inject "cfg:@feature.flags-key"
	FeatureFlags []string
}

config.yaml:
  configs:
    store:
      implementation: "postgres-store"

    app:
      page-size: 20
      name: "UserService"

    cache:
      ttl-key: "cache.default-ttl"
      default-ttl: "5m"

    feature:
      flags-key: "features.enabled"

    features:
      enabled:
        - "feature-a"
        - "feature-b"
*/

/* Example 5: Environment-specific config resolution

// @Service "email-service"
type EmailService struct {
	// @Inject "cfg:@email.provider-key"
	Provider string  // Resolves to different providers per environment

	// @Inject "cfg:@email.api-key-path"
	APIKey string  // Resolves to environment-specific API key
}

config.yaml (development):
  configs:
    email:
      provider-key: "providers.email.dev"
      api-key-path: "secrets.email.dev-key"

    providers:
      email:
        dev: "mailhog"

    secrets:
      email:
        dev-key: "dev-api-key-123"

config.yaml (production):
  configs:
    email:
      provider-key: "providers.email.prod"
      api-key-path: "secrets.email.prod-key"

    providers:
      email:
        prod: "sendgrid"

    secrets:
      email:
        prod-key: "prod-api-key-xyz"
*/
