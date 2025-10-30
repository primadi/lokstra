# KvStore (Redis)

The `kvstore_redis` service provides a Redis-backed key-value store with automatic JSON serialization, key prefixing for namespacing, and TTL support.

## Table of Contents

- [Overview](#overview)
- [Configuration](#configuration)
- [Registration](#registration)
- [Basic Operations](#basic-operations)
- [Key Management](#key-management)
- [Advanced Features](#advanced-features)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Overview

**Service Type:** `kvstore_redis`

**Interface:** `serviceapi.KvStore`

**Key Features:**

```
✓ Auto JSON Serialization  - Automatic encoding/decoding
✓ Key Prefixing            - Namespace isolation
✓ TTL Support              - Automatic expiration
✓ Batch Operations         - Delete multiple keys
✓ Pattern Matching         - Find keys by pattern
```

## Configuration

### Config Struct

```go
type Config struct {
    Addr     string `json:"addr" yaml:"addr"`          // Redis host:port
    Password string `json:"password" yaml:"password"`  // Redis password
    DB       int    `json:"db" yaml:"db"`              // Database number (0-15)
    PoolSize int    `json:"pool_size" yaml:"pool_size"` // Connection pool size
    Prefix   string `json:"prefix" yaml:"prefix"`      // Key prefix for namespacing
}
```

### YAML Configuration

**Basic Configuration:**

```yaml
services:
  cache:
    type: kvstore_redis
    config:
      addr: localhost:6379
      prefix: myapp
```

**Full Configuration:**

```yaml
services:
  cache:
    type: kvstore_redis
    config:
      addr: ${REDIS_ADDR:localhost:6379}
      password: ${REDIS_PASSWORD}
      db: 0
      pool_size: 20
      prefix: ${APP_NAME:myapp}
```

**Multiple KvStore Instances:**

```yaml
services:
  # User cache
  user_cache:
    type: kvstore_redis
    config:
      addr: localhost:6379
      db: 0
      prefix: users
      
  # Session cache
  session_cache:
    type: kvstore_redis
    config:
      addr: localhost:6379
      db: 1
      prefix: sessions
      
  # Rate limit cache
  ratelimit_cache:
    type: kvstore_redis
    config:
      addr: localhost:6379
      db: 2
      prefix: ratelimit
```

### Programmatic Configuration

```go
import (
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/serviceapi"
    "github.com/primadi/lokstra/services/kvstore_redis"
)

// Register service
kvstore_redis.Register()

// Create kvstore
kvStore := lokstra_registry.NewService[serviceapi.KvStore](
    "cache", "kvstore_redis",
    map[string]any{
        "addr":      "localhost:6379",
        "password":  "",
        "db":        0,
        "pool_size": 10,
        "prefix":    "myapp",
    },
)
```

## Registration

### Basic Registration

```go
import "github.com/primadi/lokstra/services/kvstore_redis"

func init() {
    kvstore_redis.Register()
}
```

### Bulk Registration

```go
import "github.com/primadi/lokstra/services"

func main() {
    // Registers all services including kvstore_redis
    services.RegisterAllServices()
    
    // Or register only core services
    services.RegisterCoreServices()
}
```

## Basic Operations

### Interface Definition

```go
type KvStore interface {
    Set(ctx context.Context, key string, value any, ttl time.Duration) error
    Get(ctx context.Context, key string, dest any) error
    Delete(ctx context.Context, key string) error
    DeleteKeys(ctx context.Context, keys ...string) error
    Keys(ctx context.Context, pattern string) ([]string, error)
}
```

### Set Values

**Basic Set:**

```go
ctx := context.Background()

// Set without expiration (permanent)
err := kvStore.Set(ctx, "user:123", userData, 0)

// Set with TTL (expires after 5 minutes)
err = kvStore.Set(ctx, "session:abc", sessionData, 5*time.Minute)
```

**Set Different Types:**

```go
// Struct
type User struct {
    ID    int
    Name  string
    Email string
}

user := User{ID: 123, Name: "John", Email: "john@example.com"}
err := kvStore.Set(ctx, "user:123", user, time.Hour)

// Map
data := map[string]any{
    "id":    123,
    "name":  "John",
    "email": "john@example.com",
}
err = kvStore.Set(ctx, "user:123", data, time.Hour)

// Slice
tags := []string{"go", "redis", "cache"}
err = kvStore.Set(ctx, "user:123:tags", tags, time.Hour)

// Primitive types
err = kvStore.Set(ctx, "counter", 42, 0)
err = kvStore.Set(ctx, "message", "Hello, World!", 10*time.Minute)
```

### Get Values

**Basic Get:**

```go
var user User
err := kvStore.Get(ctx, "user:123", &user)
if err != nil {
    if errors.Is(err, redis.Nil) {
        // Key doesn't exist
        return nil, ErrNotFound
    }
    return nil, err
}
```

**Get Different Types:**

```go
// Struct
var user User
err := kvStore.Get(ctx, "user:123", &user)

// Map
var data map[string]any
err = kvStore.Get(ctx, "user:123", &data)

// Slice
var tags []string
err = kvStore.Get(ctx, "user:123:tags", &tags)

// Primitive types
var count int
err = kvStore.Get(ctx, "counter", &count)

var message string
err = kvStore.Get(ctx, "message", &message)
```

**Handle Missing Keys:**

```go
var user User
err := kvStore.Get(ctx, "user:123", &user)
if err != nil {
    if errors.Is(err, redis.Nil) {
        // Key doesn't exist - return default value or error
        return nil, ErrUserNotFound
    }
    // Other error
    return nil, fmt.Errorf("failed to get user: %w", err)
}
```

### Delete Values

**Delete Single Key:**

```go
err := kvStore.Delete(ctx, "user:123")
if err != nil {
    return err
}
```

**Delete Multiple Keys:**

```go
err := kvStore.DeleteKeys(ctx, "user:1", "user:2", "user:3")
if err != nil {
    return err
}

// Delete with slice
userIDs := []string{"user:1", "user:2", "user:3"}
err = kvStore.DeleteKeys(ctx, userIDs...)
```

## Key Management

### Key Prefixing

All keys are automatically prefixed with the configured prefix:

```go
// Config
config:
  prefix: myapp

// Your code
kvStore.Set(ctx, "user:123", userData, 0)

// Actual Redis key
// myapp:user:123
```

**Benefits:**
- Namespace isolation (multiple apps can share Redis)
- Easy identification of keys
- Bulk operations by namespace

### Find Keys by Pattern

**Simple Patterns:**

```go
// Find all user keys
keys, err := kvStore.Keys(ctx, "user:*")
// Returns: ["user:1", "user:2", "user:3", ...]

// Find specific pattern
keys, err = kvStore.Keys(ctx, "session:abc*")
// Returns: ["session:abc123", "session:abc456", ...]
```

**Complex Patterns:**

```go
// All cache keys
keys, err := kvStore.Keys(ctx, "cache:*")

// Keys matching pattern
keys, err = kvStore.Keys(ctx, "user:*:profile")

// All keys (use with caution!)
keys, err = kvStore.Keys(ctx, "*")
```

**Returned Keys Have Prefix Removed:**

```go
// Config prefix: "myapp"
// Redis has key: "myapp:user:123"

keys, err := kvStore.Keys(ctx, "user:*")
// Returns: ["user:123"]  (prefix removed)
```

### Bulk Delete by Pattern

```go
// Find and delete all session keys
sessionKeys, err := kvStore.Keys(ctx, "session:*")
if err != nil {
    return err
}

if len(sessionKeys) > 0 {
    err = kvStore.DeleteKeys(ctx, sessionKeys...)
    if err != nil {
        return err
    }
}
```

## Advanced Features

### TTL Management

**Set with Expiration:**

```go
// 5 minutes
kvStore.Set(ctx, "otp:user123", otpCode, 5*time.Minute)

// 1 hour
kvStore.Set(ctx, "session:abc", sessionData, time.Hour)

// 24 hours
kvStore.Set(ctx, "cache:report", report, 24*time.Hour)

// No expiration (permanent)
kvStore.Set(ctx, "config:app", appConfig, 0)
```

**Common TTL Patterns:**

```go
const (
    OtpTTL      = 5 * time.Minute      // Short-lived OTP codes
    SessionTTL  = 24 * time.Hour       // User sessions
    CacheTTL    = time.Hour            // Cached data
    TempDataTTL = 15 * time.Minute     // Temporary data
)

kvStore.Set(ctx, "otp:"+userID, code, OtpTTL)
kvStore.Set(ctx, "session:"+token, session, SessionTTL)
kvStore.Set(ctx, "cache:users", users, CacheTTL)
```

### JSON Serialization

Values are automatically serialized to/from JSON:

```go
// Complex struct with nested fields
type UserProfile struct {
    ID       int
    Name     string
    Settings map[string]any
    Tags     []string
    Metadata struct {
        CreatedAt time.Time
        UpdatedAt time.Time
    }
}

profile := UserProfile{
    ID:   123,
    Name: "John Doe",
    Settings: map[string]any{
        "theme": "dark",
        "lang":  "en",
    },
    Tags: []string{"premium", "verified"},
}

// Automatically serialized to JSON
kvStore.Set(ctx, "profile:123", profile, time.Hour)

// Automatically deserialized from JSON
var retrieved UserProfile
kvStore.Get(ctx, "profile:123", &retrieved)
```

### Cache Patterns

**Cache-Aside Pattern:**

```go
func GetUser(ctx context.Context, userID int) (*User, error) {
    cacheKey := fmt.Sprintf("user:%d", userID)
    
    // Try cache first
    var user User
    err := kvStore.Get(ctx, cacheKey, &user)
    if err == nil {
        return &user, nil // Cache hit
    }
    
    // Cache miss - fetch from database
    user, err = userRepo.GetByID(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    _ = kvStore.Set(ctx, cacheKey, user, time.Hour)
    
    return &user, nil
}
```

**Write-Through Pattern:**

```go
func UpdateUser(ctx context.Context, user *User) error {
    // Update database
    err := userRepo.Update(ctx, user)
    if err != nil {
        return err
    }
    
    // Update cache
    cacheKey := fmt.Sprintf("user:%d", user.ID)
    _ = kvStore.Set(ctx, cacheKey, user, time.Hour)
    
    return nil
}
```

**Cache Invalidation:**

```go
func DeleteUser(ctx context.Context, userID int) error {
    // Delete from database
    err := userRepo.Delete(ctx, userID)
    if err != nil {
        return err
    }
    
    // Invalidate cache
    cacheKey := fmt.Sprintf("user:%d", userID)
    _ = kvStore.Delete(ctx, cacheKey)
    
    return nil
}
```

## Best Practices

### Key Naming

```go
✓ DO: Use hierarchical keys
"user:123"
"user:123:profile"
"user:123:settings"
"session:abc:data"

✗ DON'T: Use flat keys
"user123"
"userprofile123"
"usersettings123"

✓ DO: Use consistent separator
"user:123:profile"    // Use colon

✗ DON'T: Mix separators
"user-123:profile"    // Inconsistent
"user.123.profile"    // Inconsistent
```

### TTL Management

```go
✓ DO: Set appropriate TTLs
kvStore.Set(ctx, "otp:"+id, code, 5*time.Minute)    // Short for OTP
kvStore.Set(ctx, "session:"+id, data, time.Hour)    // Medium for session
kvStore.Set(ctx, "config:app", cfg, 0)              // Permanent for config

✗ DON'T: Use same TTL for everything
kvStore.Set(ctx, key, value, time.Hour)  // Same TTL for all

✓ DO: Use constants for TTLs
const (
    ShortTTL  = 5 * time.Minute
    MediumTTL = time.Hour
    LongTTL   = 24 * time.Hour
)

✗ DON'T: Use magic numbers
kvStore.Set(ctx, key, value, 3600*time.Second)  // What does 3600 mean?
```

### Error Handling

```go
✓ DO: Check for key not found
err := kvStore.Get(ctx, key, &value)
if err != nil {
    if errors.Is(err, redis.Nil) {
        return nil, ErrNotFound
    }
    return nil, err
}

✓ DO: Handle serialization errors gracefully
err := kvStore.Set(ctx, key, value, ttl)
if err != nil {
    log.Printf("failed to cache: %v", err)
    // Continue without cache
}

✗ DON'T: Panic on cache errors
err := kvStore.Get(ctx, key, &value)
if err != nil {
    panic(err)  // BAD: Cache shouldn't crash app
}
```

### Performance

```go
✓ DO: Use batch operations
keys := []string{"user:1", "user:2", "user:3"}
kvStore.DeleteKeys(ctx, keys...)  // Single operation

✗ DON'T: Loop for multiple operations
for _, key := range keys {
    kvStore.Delete(ctx, key)  // BAD: Multiple round-trips
}

✓ DO: Use appropriate prefixes
config:
  prefix: myapp  // Good namespace

✗ DON'T: Use overly long prefixes
config:
  prefix: my_super_long_application_name_v1_prod  // Too long

✓ DO: Be careful with Keys() on large datasets
keys, err := kvStore.Keys(ctx, "user:*")  // OK for small sets

✗ DON'T: Use Keys("*") in production
keys, err := kvStore.Keys(ctx, "*")  // BAD: Blocks Redis
```

## Examples

### User Cache Repository

```go
package repository

import (
    "context"
    "fmt"
    "time"
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/serviceapi"
)

type CachedUserRepository struct {
    cache  serviceapi.KvStore
    db     *UserRepository  // Database repository
}

func NewCachedUserRepository(db *UserRepository) *CachedUserRepository {
    return &CachedUserRepository{
        cache: lokstra_registry.GetService[serviceapi.KvStore]("cache"),
        db:    db,
    }
}

// Get user with caching
func (r *CachedUserRepository) GetByID(ctx context.Context, id int) (*User, error) {
    cacheKey := fmt.Sprintf("user:%d", id)
    
    // Try cache
    var user User
    err := r.cache.Get(ctx, cacheKey, &user)
    if err == nil {
        return &user, nil
    }
    
    // Cache miss - get from database
    user, err = r.db.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Cache the result
    _ = r.cache.Set(ctx, cacheKey, user, time.Hour)
    
    return &user, nil
}

// Update user and invalidate cache
func (r *CachedUserRepository) Update(ctx context.Context, user *User) error {
    // Update database
    if err := r.db.Update(ctx, user); err != nil {
        return err
    }
    
    // Invalidate cache
    cacheKey := fmt.Sprintf("user:%d", user.ID)
    _ = r.cache.Delete(ctx, cacheKey)
    
    return nil
}

// Delete user and invalidate cache
func (r *CachedUserRepository) Delete(ctx context.Context, id int) error {
    // Delete from database
    if err := r.db.Delete(ctx, id); err != nil {
        return err
    }
    
    // Invalidate cache
    cacheKey := fmt.Sprintf("user:%d", id)
    _ = r.cache.Delete(ctx, cacheKey)
    
    return nil
}
```

### OTP Service

```go
package service

import (
    "context"
    "crypto/rand"
    "fmt"
    "math/big"
    "time"
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/serviceapi"
)

type OTPService struct {
    kvStore serviceapi.KvStore
}

const (
    OtpLength = 6
    OtpTTL    = 5 * time.Minute
)

func NewOTPService() *OTPService {
    return &OTPService{
        kvStore: lokstra_registry.GetService[serviceapi.KvStore]("cache"),
    }
}

// Generate and store OTP
func (s *OTPService) Generate(ctx context.Context, userID string) (string, error) {
    // Generate 6-digit OTP
    otp, err := generateOTP(OtpLength)
    if err != nil {
        return "", err
    }
    
    // Store with TTL
    key := fmt.Sprintf("otp:%s", userID)
    err = s.kvStore.Set(ctx, key, otp, OtpTTL)
    if err != nil {
        return "", err
    }
    
    return otp, nil
}

// Verify OTP
func (s *OTPService) Verify(ctx context.Context, userID, otp string) (bool, error) {
    key := fmt.Sprintf("otp:%s", userID)
    
    // Get stored OTP
    var storedOTP string
    err := s.kvStore.Get(ctx, key, &storedOTP)
    if err != nil {
        return false, nil  // OTP not found or expired
    }
    
    // Compare
    if storedOTP != otp {
        return false, nil
    }
    
    // Valid - delete OTP (single use)
    _ = s.kvStore.Delete(ctx, key)
    
    return true, nil
}

func generateOTP(length int) (string, error) {
    digits := "0123456789"
    otp := make([]byte, length)
    for i := range otp {
        num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
        if err != nil {
            return "", err
        }
        otp[i] = digits[num.Int64()]
    }
    return string(otp), nil
}
```

### Rate Limiter

```go
package middleware

import (
    "context"
    "fmt"
    "time"
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/serviceapi"
)

type RateLimiter struct {
    kvStore serviceapi.KvStore
    limit   int
    window  time.Duration
}

type RateLimitInfo struct {
    Count int
    Limit int
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
    return &RateLimiter{
        kvStore: lokstra_registry.GetService[serviceapi.KvStore]("ratelimit_cache"),
        limit:   limit,
        window:  window,
    }
}

// Check if rate limit exceeded
func (r *RateLimiter) Check(ctx context.Context, identifier string) (bool, *RateLimitInfo, error) {
    key := fmt.Sprintf("ratelimit:%s", identifier)
    
    // Get current count
    var info RateLimitInfo
    err := r.kvStore.Get(ctx, key, &info)
    if err != nil {
        // First request - initialize
        info = RateLimitInfo{
            Count: 1,
            Limit: r.limit,
        }
        _ = r.kvStore.Set(ctx, key, info, r.window)
        return false, &info, nil  // Not exceeded
    }
    
    // Increment count
    info.Count++
    
    // Check limit
    if info.Count > r.limit {
        return true, &info, nil  // Exceeded
    }
    
    // Update count
    _ = r.kvStore.Set(ctx, key, info, r.window)
    
    return false, &info, nil  // Not exceeded
}
```

### Session Manager

```go
package session

import (
    "context"
    "fmt"
    "time"
    "github.com/google/uuid"
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/serviceapi"
)

type SessionManager struct {
    kvStore serviceapi.KvStore
    ttl     time.Duration
}

type SessionData struct {
    UserID   string
    TenantID string
    Metadata map[string]any
}

func NewSessionManager(ttl time.Duration) *SessionManager {
    return &SessionManager{
        kvStore: lokstra_registry.GetService[serviceapi.KvStore]("session_cache"),
        ttl:     ttl,
    }
}

// Create session
func (m *SessionManager) Create(ctx context.Context, data *SessionData) (string, error) {
    // Generate session ID
    sessionID := uuid.New().String()
    key := fmt.Sprintf("session:%s", sessionID)
    
    // Store session
    err := m.kvStore.Set(ctx, key, data, m.ttl)
    if err != nil {
        return "", err
    }
    
    return sessionID, nil
}

// Get session
func (m *SessionManager) Get(ctx context.Context, sessionID string) (*SessionData, error) {
    key := fmt.Sprintf("session:%s", sessionID)
    
    var data SessionData
    err := m.kvStore.Get(ctx, key, &data)
    if err != nil {
        return nil, err
    }
    
    return &data, nil
}

// Refresh session TTL
func (m *SessionManager) Refresh(ctx context.Context, sessionID string) error {
    // Get existing data
    data, err := m.Get(ctx, sessionID)
    if err != nil {
        return err
    }
    
    // Re-set with new TTL
    key := fmt.Sprintf("session:%s", sessionID)
    return m.kvStore.Set(ctx, key, data, m.ttl)
}

// Delete session
func (m *SessionManager) Delete(ctx context.Context, sessionID string) error {
    key := fmt.Sprintf("session:%s", sessionID)
    return m.kvStore.Delete(ctx, key)
}

// Delete all user sessions
func (m *SessionManager) DeleteUserSessions(ctx context.Context, userID string) error {
    // Find all sessions for user
    pattern := fmt.Sprintf("session:*")
    keys, err := m.kvStore.Keys(ctx, pattern)
    if err != nil {
        return err
    }
    
    // Filter by user ID
    var userSessions []string
    for _, key := range keys {
        var data SessionData
        err := m.kvStore.Get(ctx, key, &data)
        if err == nil && data.UserID == userID {
            userSessions = append(userSessions, key)
        }
    }
    
    // Delete user sessions
    if len(userSessions) > 0 {
        return m.kvStore.DeleteKeys(ctx, userSessions...)
    }
    
    return nil
}
```

## Related Documentation

- [Services Overview](index) - Service architecture and patterns
- [DbPool Service](dbpool-pg) - PostgreSQL connection pooling
- [Redis Service](redis) - Direct Redis client access
- [Auth Session Service](auth-session-redis) - Session management

---

**Next:** [Metrics Service](metrics-prometheus) - Prometheus metrics collection
