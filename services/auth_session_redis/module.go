package auth_session_redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/serviceapi/auth"
	"github.com/redis/go-redis/v9"
)

const SERVICE_TYPE = "auth_session_redis"

// Config represents the configuration for Redis-based Session service.
type Config struct {
	Addr     string `json:"addr" yaml:"addr"`         // host:port address
	Password string `json:"password" yaml:"password"` // password
	DB       int    `json:"db" yaml:"db"`             // database number
	PoolSize int    `json:"pool_size" yaml:"pool_size"`
	Prefix   string `json:"prefix" yaml:"prefix"` // key prefix for namespacing
}

type sessionRedis struct {
	client *redis.Client
	prefix string
}

var _ auth.Session = (*sessionRedis)(nil)

func (s *sessionRedis) sessionKey(sessionID string) string {
	return fmt.Sprintf("%s:session:%s", s.prefix, sessionID)
}

func (s *sessionRedis) userSessionsKey(tenantID, userID string) string {
	return fmt.Sprintf("%s:user_sessions:%s:%s", s.prefix, tenantID, userID)
}

func (s *sessionRedis) Set(ctx context.Context, sessionID string, data *auth.SessionData, ttl time.Duration) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	pipe := s.client.Pipeline()

	// Store session data
	pipe.Set(ctx, s.sessionKey(sessionID), jsonData, ttl)

	// Add to user's session list
	pipe.SAdd(ctx, s.userSessionsKey(data.TenantID, data.UserID), sessionID)
	pipe.Expire(ctx, s.userSessionsKey(data.TenantID, data.UserID), ttl)

	_, err = pipe.Exec(ctx)
	return err
}

func (s *sessionRedis) Get(ctx context.Context, sessionID string) (*auth.SessionData, error) {
	data, err := s.client.Get(ctx, s.sessionKey(sessionID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found")
		}
		return nil, err
	}

	var sessionData auth.SessionData
	if err := json.Unmarshal(data, &sessionData); err != nil {
		return nil, err
	}

	return &sessionData, nil
}

func (s *sessionRedis) Delete(ctx context.Context, sessionID string) error {
	// Get session data first to remove from user's session list
	sessionData, err := s.Get(ctx, sessionID)
	if err != nil {
		// If session doesn't exist, just return
		return nil
	}

	pipe := s.client.Pipeline()
	pipe.Del(ctx, s.sessionKey(sessionID))
	pipe.SRem(ctx, s.userSessionsKey(sessionData.TenantID, sessionData.UserID), sessionID)

	_, err = pipe.Exec(ctx)
	return err
}

func (s *sessionRedis) DeleteByUser(ctx context.Context, tenantID, userID string) error {
	// Get all session IDs for the user
	sessionIDs, err := s.ListByUser(ctx, tenantID, userID)
	if err != nil {
		return err
	}

	if len(sessionIDs) == 0 {
		return nil
	}

	// Delete all sessions
	pipe := s.client.Pipeline()
	for _, sessionID := range sessionIDs {
		pipe.Del(ctx, s.sessionKey(sessionID))
	}
	pipe.Del(ctx, s.userSessionsKey(tenantID, userID))

	_, err = pipe.Exec(ctx)
	return err
}

func (s *sessionRedis) ListByUser(ctx context.Context, tenantID, userID string) ([]string, error) {
	return s.client.SMembers(ctx, s.userSessionsKey(tenantID, userID)).Result()
}

func (s *sessionRedis) Shutdown() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

func Service(cfg *Config) *sessionRedis {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})
	return &sessionRedis{
		client: client,
		prefix: cfg.Prefix,
	}
}

func ServiceFactory(params map[string]any) any {
	cfg := &Config{
		Addr:     utils.GetValueFromMap(params, "addr", "localhost:6379"),
		Password: utils.GetValueFromMap(params, "password", ""),
		DB:       utils.GetValueFromMap(params, "db", 0),
		PoolSize: utils.GetValueFromMap(params, "pool_size", 10),
		Prefix:   utils.GetValueFromMap(params, "prefix", "auth"),
	}
	return Service(cfg)
}

func Register() {
	lokstra_registry.RegisterServiceType(SERVICE_TYPE, ServiceFactory,
		nil)
}
