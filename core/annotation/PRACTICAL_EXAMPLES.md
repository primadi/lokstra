# Practical Examples: Smart @Inject

## Example 1: Multi-Environment Configuration

### Problem
Your app needs different JWT secrets, database timeouts, and API keys per environment (dev, staging, prod), but you don't want to change code.

### Solution: Indirect Config Injection

**Code (unchanged across environments):**

```go
// @EndpointService name="auth-service", prefix="/api/auth"
type AuthService struct {
    // @Inject "user-repo"
    UserRepo UserRepository
    
    // @Inject "cfg:@jwt.secret-key"  // Resolves dynamically
    JWTSecret string
    
    // @Inject "cfg:@db.timeout-key"  // Resolves dynamically
    DBTimeout time.Duration
}
```

**config.dev.yaml:**
```yaml
configs:
  jwt:
    secret-key: "security.dev-jwt"
  
  db:
    timeout-key: "database.dev-timeout"
  
  security:
    dev-jwt: "dev-secret-123"
  
  database:
    dev-timeout: "5s"
```

**config.prod.yaml:**
```yaml
configs:
  jwt:
    secret-key: "security.prod-jwt"
  
  db:
    timeout-key: "database.prod-timeout"
  
  security:
    prod-jwt: "prod-secret-xyz-secure"
  
  database:
    prod-timeout: "30s"
```

**Result:** Just swap config file, no code changes!

---

## Example 2: Feature Flags System

### Problem
You want to enable/disable features dynamically based on environment or user tier.

### Solution: Indirect Config with Slices

```go
// @Service "feature-service"
type FeatureService struct {
    // @Inject "cfg:@features.enabled-list"
    EnabledFeatures []string
    
    // @Inject "cfg:@features.beta-list"
    BetaFeatures []string
}

func (s *FeatureService) IsEnabled(feature string) bool {
    for _, f := range s.EnabledFeatures {
        if f == feature {
            return true
        }
    }
    return false
}
```

**config.yaml:**
```yaml
configs:
  features:
    enabled-list: "features.production"  # Switch to "features.beta" for beta env
    beta-list: "features.experimental"
  
  features:
    production:
      - "login"
      - "dashboard"
      - "reports"
    
    beta:
      - "login"
      - "dashboard"
      - "reports"
      - "ai-assistant"
      - "realtime-sync"
    
    experimental:
      - "ai-assistant"
      - "realtime-sync"
      - "blockchain-integration"
```

---

## Example 3: Multi-Tenant Database Selection

### Problem
Different tenants use different database implementations (PostgreSQL, MySQL, MongoDB).

### Solution: Service + Config Indirection

```go
// @EndpointService name="tenant-service", prefix="/api/tenants"
type TenantService struct {
    // @Inject "@tenant.db-provider"  // Service name from config
    DB database.Provider
    
    // @Inject "cfg:@tenant.max-connections"
    MaxConnections int
}
```

**config.yaml:**
```yaml
configs:
  tenant:
    db-provider: "database.tenant-a.provider"
    max-connections: "database.tenant-a.max-conn"
  
  database:
    tenant-a:
      provider: "postgres-db"
      max-conn: 100
    
    tenant-b:
      provider: "mysql-db"
      max-conn: 50
    
    tenant-c:
      provider: "mongo-db"
      max-conn: 200
```

Switch tenant by changing `tenant.db-provider` to `database.tenant-b.provider`!

---

## Example 4: API Rate Limiting (Per Plan)

### Problem
Different subscription plans have different rate limits.

```go
// @EndpointService name="api-gateway", prefix="/api"
type APIGateway struct {
    // @Inject "cfg:@rate-limit.requests-per-minute"
    RequestsPerMinute int
    
    // @Inject "cfg:@rate-limit.burst-size"
    BurstSize int
    
    // @Inject "cfg:@rate-limit.timeout"
    Timeout time.Duration
}
```

**User Session Config (dynamic per user):**
```yaml
configs:
  # Loaded based on user's subscription plan
  rate-limit:
    requests-per-minute: "plans.enterprise.rpm"
    burst-size: "plans.enterprise.burst"
    timeout: "plans.enterprise.timeout"
  
  plans:
    free:
      rpm: 10
      burst: 2
      timeout: "1s"
    
    pro:
      rpm: 100
      burst: 20
      timeout: "5s"
    
    enterprise:
      rpm: 1000
      burst: 200
      timeout: "30s"
```

---

## Example 5: Email Provider Selection (Development vs Production)

```go
// @Service "email-service"
type EmailService struct {
    // @Inject "@email.provider"  // Service injection from config
    Provider email.Sender
    
    // @Inject "cfg:@email.from-address"
    FromAddress string
    
    // @Inject "cfg:@email.reply-to"
    ReplyTo string
}
```

**config.dev.yaml:**
```yaml
configs:
  email:
    provider: "providers.email.dev"
    from-address: "email.addresses.dev-from"
    reply-to: "email.addresses.dev-reply"
  
  providers:
    email:
      dev: "mailhog-service"  # Local email testing
  
  email:
    addresses:
      dev-from: "dev@localhost"
      dev-reply: "noreply@localhost"
```

**config.prod.yaml:**
```yaml
configs:
  email:
    provider: "providers.email.prod"
    from-address: "email.addresses.prod-from"
    reply-to: "email.addresses.prod-reply"
  
  providers:
    email:
      prod: "sendgrid-service"  # Real email service
  
  email:
    addresses:
      prod-from: "hello@company.com"
      prod-reply: "support@company.com"
```

---

## Example 6: Cache Strategy Selection

```go
// @EndpointService name="product-service", prefix="/api/products"
type ProductService struct {
    // @Inject "@cache.provider"
    Cache cache.Provider
    
    // @Inject "cfg:@cache.ttl"
    CacheTTL time.Duration
    
    // @Inject "cfg:@cache.max-size"
    MaxCacheSize int
}
```

**config.yaml:**
```yaml
configs:
  cache:
    provider: "cache-strategies.products"
    ttl: "cache-ttl.products"
    max-size: "cache-size.products"
  
  cache-strategies:
    products: "redis-cache"      # Or "memory-cache" for dev
    sessions: "memory-cache"
    
  cache-ttl:
    products: "1h"
    sessions: "15m"
  
  cache-size:
    products: 10000
    sessions: 5000
```

---

## Example 7: Logging Level (Per Module)

```go
// @Service "user-service"
type UserService struct {
    // @Inject "logger"
    Logger logger.Logger
    
    // @Inject "cfg:@logging.level"
    LogLevel string
    
    // @Inject "cfg:@logging.output"
    LogOutput string
}

func (s *UserService) Init() error {
    s.Logger.SetLevel(s.LogLevel)
    s.Logger.SetOutput(s.LogOutput)
    return nil
}
```

**config.yaml:**
```yaml
configs:
  logging:
    level: "log-levels.user-service"
    output: "log-outputs.user-service"
  
  log-levels:
    user-service: "debug"      # Per-module control
    auth-service: "info"
    payment-service: "warn"
  
  log-outputs:
    user-service: "stdout"
    auth-service: "file:/var/log/auth.log"
    payment-service: "file:/var/log/payment.log"
```

---

## Benefits Demonstrated

1. **Zero Code Changes**: Switch behavior by changing config only
2. **Type Safety**: All conversions handled by generated code
3. **Environment Agnostic**: Same code, different configs
4. **Multi-Tenant**: Easy per-tenant customization
5. **Feature Flags**: Dynamic feature enablement
6. **A/B Testing**: Easy config-based experimentation

## Migration Path

### Step 1: Identify Hard-Coded Config Keys
```go
// Before
// @Inject "cfg:app.jwt-secret"
JWTSecret string
```

### Step 2: Add Indirection Layer
```go
// After
// @Inject "cfg:@jwt.secret-key"
JWTSecret string
```

### Step 3: Update Config
```yaml
# Add indirection
configs:
  jwt:
    secret-key: "security.production-jwt"
  
  security:
    production-jwt: "actual-secret"
```

### Step 4: Profit!
Now you can switch `jwt.secret-key` to point to different secrets per environment!
