---
layout: default
title: "@InjectCfgValue Annotation"
parent: Framework Guide
nav_order: 8
---

# @InjectCfgValue Annotation

## Overview

The `@InjectCfgValue` annotation injects configuration values from `config.yaml` into service fields. It provides type-safe configuration injection with automatic type detection and optional default values.

## Basic Syntax

```go
// @InjectCfgValue "config.key"
FieldName FieldType

// or with default value
// @InjectCfgValue key="config.key", default="value"
FieldName FieldType
```

## Supported Formats

### 1. Positional Arguments

```go
// @InjectCfgValue "smtp.host"
SMTPHost string

// @InjectCfgValue "smtp.host", "localhost"
SMTPHost string
```

### 2. Named Arguments

```go
// @InjectCfgValue key="smtp.host"
SMTPHost string

// @InjectCfgValue key="smtp.host", default="localhost"
SMTPHost string
```

## Supported Types

The framework automatically detects the field type and uses the appropriate `GetConfig*` function:

| Go Type | Generated Function | Example Default |
|---------|-------------------|-----------------|
| `string` | `GetConfig` | `""` or custom |
| `int`, `int8`, `int16`, `int32`, `int64` | `GetConfigInt` | `0` or custom |
| `uint`, `uint8`, `uint16`, `uint32`, `uint64` | `GetConfigInt` | `0` or custom |
| `bool` | `GetConfigBool` | `false` or custom |
| `float32`, `float64` | `GetConfigFloat` | `0.0` or custom |
| `time.Duration` | `GetConfigDuration` | `0` or custom |

## Examples

### String Configuration

```go
// @Service name="email-service"
type EmailService struct {
    // No default - required in config
    // @InjectCfgValue "smtp.host"
    SMTPHost string
    
    // With default
    // @InjectCfgValue key="smtp.from", default="noreply@example.com"
    FromEmail string
}
```

**Generated:**
```go
SMTPHost:  lokstra_registry.GetConfig("smtp.host", ""),
FromEmail: lokstra_registry.GetConfig("smtp.from", "noreply@example.com"),
```

**Config (config.yaml):**
```yaml
configs:
  smtp:
    host: "smtp.gmail.com"
    # from: uses default "noreply@example.com"
```

### Integer Configuration

```go
// @Service name="rate-limiter"
type RateLimiter struct {
    // @InjectCfgValue key="rate.max-requests", default="100"
    MaxRequests int
    
    // @InjectCfgValue key="rate.window-seconds", default="60"
    WindowSeconds int64
}
```

**Generated:**
```go
MaxRequests:   lokstra_registry.GetConfigInt("rate.max-requests", 100),
WindowSeconds: lokstra_registry.GetConfigInt("rate.window-seconds", 60),
```

### Boolean Configuration

```go
// @Service name="feature-flags"
type FeatureFlags struct {
    // @InjectCfgValue key="features.new-ui", default="false"
    EnableNewUI bool
    
    // @InjectCfgValue key="features.debug-mode", default="false"
    DebugMode bool
}
```

**Generated:**
```go
EnableNewUI: lokstra_registry.GetConfigBool("features.new-ui", false),
DebugMode:   lokstra_registry.GetConfigBool("features.debug-mode", false),
```

### Duration Configuration

```go
// @Service name="cache-service"
type CacheService struct {
    // @InjectCfgValue key="cache.ttl", default="5m"
    TTL time.Duration
    
    // @InjectCfgValue key="cache.cleanup-interval", default="1h"
    CleanupInterval time.Duration
}
```

**Generated:**
```go
TTL:             lokstra_registry.GetConfigDuration("cache.ttl", 5*time.Minute),
CleanupInterval: lokstra_registry.GetConfigDuration("cache.cleanup-interval", 1*time.Hour),
```

### Float Configuration

```go
// @Service name="payment-service"
type PaymentService struct {
    // @InjectCfgValue key="payment.fee-percentage", default="2.5"
    FeePercentage float64
    
    // @InjectCfgValue key="payment.min-amount", default="10.0"
    MinAmount float32
}
```

**Generated:**
```go
FeePercentage: lokstra_registry.GetConfigFloat("payment.fee-percentage", 2.5),
MinAmount:     lokstra_registry.GetConfigFloat("payment.min-amount", 10.0),
```

## Complete Example

### Service with Mixed Config Types

```go
package application

import "time"

// @Service name="app-config"
type AppConfig struct {
    // String configs
    // @InjectCfgValue key="app.name", default="MyApp"
    AppName string
    
    // @InjectCfgValue "app.version"
    Version string  // Required, no default
    
    // Integer configs
    // @InjectCfgValue key="server.port", default="8080"
    ServerPort int
    
    // @InjectCfgValue key="server.max-connections", default="1000"
    MaxConnections int64
    
    // Boolean configs
    // @InjectCfgValue key="server.enable-gzip", default="true"
    EnableGzip bool
    
    // @InjectCfgValue key="server.debug", default="false"
    Debug bool
    
    // Duration configs
    // @InjectCfgValue key="server.read-timeout", default="30s"
    ReadTimeout time.Duration
    
    // @InjectCfgValue key="server.write-timeout", default="30s"
    WriteTimeout time.Duration
    
    // Float configs
    // @InjectCfgValue key="cache.eviction-ratio", default="0.1"
    CacheEvictionRatio float64
}

func (c *AppConfig) GetServerAddress() string {
    return fmt.Sprintf(":%d", c.ServerPort)
}
```

### Configuration File

```yaml
# config.yaml
configs:
  app:
    name: "ProductionApp"
    version: "1.2.3"  # Required
  
  server:
    port: 9000
    max-connections: 5000
    enable-gzip: true
    debug: false
    read-timeout: "60s"
    write-timeout: "60s"
  
  cache:
    eviction-ratio: 0.25
```

### Generated Code

```go
// zz_generated.lokstra.go
func RegisterAppConfig() {
    lokstra_registry.RegisterLazyService("app-config", func(deps map[string]any, cfg map[string]any) any {
        return &AppConfig{
            AppName:            lokstra_registry.GetConfig("app.name", "MyApp"),
            Version:            lokstra_registry.GetConfig("app.version", ""),
            ServerPort:         lokstra_registry.GetConfigInt("server.port", 8080),
            MaxConnections:     lokstra_registry.GetConfigInt("server.max-connections", 1000),
            EnableGzip:         lokstra_registry.GetConfigBool("server.enable-gzip", true),
            Debug:              lokstra_registry.GetConfigBool("server.debug", false),
            ReadTimeout:        lokstra_registry.GetConfigDuration("server.read-timeout", 30*time.Second),
            WriteTimeout:       lokstra_registry.GetConfigDuration("server.write-timeout", 30*time.Second),
            CacheEvictionRatio: lokstra_registry.GetConfigFloat("cache.eviction-ratio", 0.1),
        }
    }, nil)
}
```

## Default Value Behavior

### With Default Value

If config key is missing in `config.yaml`, uses the default:

```go
// @InjectCfgValue key="smtp.host", default="localhost"
SMTPHost string  // "localhost" if not in config
```

### Without Default Value

If config key is missing, uses type's zero value:

```go
// @InjectCfgValue "smtp.host"
SMTPHost string  // "" if not in config

// @InjectCfgValue "server.port"
Port int  // 0 if not in config

// @InjectCfgValue "debug"
Debug bool  // false if not in config
```

## Best Practices

### 1. Use Meaningful Config Keys

✅ **Good:**
```go
// @InjectCfgValue key="database.connection-timeout", default="30s"
DBTimeout time.Duration

// @InjectCfgValue key="auth.jwt-secret"
JWTSecret string
```

❌ **Bad:**
```go
// @InjectCfgValue "timeout"  // Too vague
Timeout time.Duration

// @InjectCfgValue "secret"  // Not descriptive
Secret string
```

### 2. Provide Sensible Defaults

✅ **Good:**
```go
// @InjectCfgValue key="server.port", default="8080"
Port int

// @InjectCfgValue key="cache.ttl", default="5m"
CacheTTL time.Duration
```

### 3. Group Related Configs

**config.yaml:**
```yaml
configs:
  database:
    host: "localhost"
    port: 5432
    timeout: "30s"
  
  smtp:
    host: "smtp.gmail.com"
    port: 587
    from: "noreply@example.com"
  
  features:
    enable-caching: true
    enable-logging: true
    debug-mode: false
```

**Service:**
```go
// @Service name="db-service"
type DBService struct {
    // @InjectCfgValue key="database.host", default="localhost"
    Host string
    
    // @InjectCfgValue key="database.port", default="5432"
    Port int
    
    // @InjectCfgValue key="database.timeout", default="30s"
    Timeout time.Duration
}
```

### 4. Mark Required Configs

```go
// @Service name="payment-service"
type PaymentService struct {
    // REQUIRED - no default
    // @InjectCfgValue "payment.api-key"
    APIKey string
    
    // Optional - has default
    // @InjectCfgValue key="payment.timeout", default="60s"
    Timeout time.Duration
}
```

## Combining @Inject and @InjectCfgValue

```go
// @Service name="notification-service"
type NotificationService struct {
    // Service dependencies
    // @Inject "user-repository"
    UserRepo UserRepository
    
    // @Inject service="email-service", optional=true
    EmailSvc EmailService
    
    // Configuration
    // @InjectCfgValue key="notifications.enabled", default="true"
    Enabled bool
    
    // @InjectCfgValue key="notifications.batch-size", default="100"
    BatchSize int
    
    // @InjectCfgValue key="notifications.retry-attempts", default="3"
    RetryAttempts int
}

func (s *NotificationService) SendNotification(userID, message string) error {
    if !s.Enabled {
        return nil  // Notifications disabled
    }
    
    user, err := s.UserRepo.GetByID(userID)
    if err != nil {
        return err
    }
    
    if s.EmailSvc != nil {
        return s.EmailSvc.Send(user.Email, message)
    }
    
    return nil
}
```

## Environment-Specific Configuration

Use different config files per environment:

**config.development.yaml:**
```yaml
configs:
  smtp:
    host: "localhost"
    port: 1025  # MailHog
  features:
    debug-mode: true
```

**config.production.yaml:**
```yaml
configs:
  smtp:
    host: "smtp.sendgrid.net"
    port: 587
  features:
    debug-mode: false
```

## See Also

- [@Service](06-service-annotation.md) - Service registration
- [@Inject](07-inject-annotation.md) - Dependency injection
- [Configuration Management](10-configuration.md) - Config file structure
- [Service Registry](09-service-registry.md) - Manual configuration
