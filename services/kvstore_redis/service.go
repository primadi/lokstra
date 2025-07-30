package kvstore_redis

import (
	"context"
	"time"

	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	client *redis.Client
}

// Delete implements serviceapi.KvStore.
func (s *Service) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}

// DeleteKeys implements serviceapi.KvStore.
func (s *Service) DeleteKeys(ctx context.Context, keys ...string) error {
	return s.client.Del(ctx, keys...).Err()
}

// Get implements serviceapi.KvStore.
func (s *Service) Get(ctx context.Context, key string, dest any) error {
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// Keys implements serviceapi.KvStore.
func (s *Service) Keys(ctx context.Context, pattern string) ([]string, error) {
	return s.client.Keys(ctx, pattern).Result()
}

// Set implements serviceapi.KvStore.
func (s *Service) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, key, data, ttl).Err()
}

func New(client *redis.Client) *Service {
	return &Service{client: client}
}

var _ serviceapi.KvStore = (*Service)(nil)
