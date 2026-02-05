package kvstore_redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/redis/go-redis/v9"
)

const SERVICE_TYPE = "kvrepository_redis"

var ErrKeyNotFound = errors.New("key not found")

var (
	mu         sync.Mutex
	poolClient = make(map[Config]*redis.Client)
)

// Config represents the configuration for Redis-based KvRepository service.
type Config struct {
	Addr     string `json:"addr" yaml:"addr"`         // host:port address
	Password string `json:"password" yaml:"password"` // password
	DB       int    `json:"db" yaml:"db"`             // database number
	PoolSize int    `json:"pool_size" yaml:"pool_size"`
	Prefix   string `json:"prefix" yaml:"prefix"` // key prefix for namespacing
}

type kvRepositoryRedis struct {
	client *redis.Client
	prefix string
}

var _ serviceapi.KvRepository = (*kvRepositoryRedis)(nil)

// GetPrefix implements [serviceapi.KvRepository].
func (k *kvRepositoryRedis) GetPrefix() string {
	return k.prefix
}

// SetPrefix implements [serviceapi.KvRepository].
func (k *kvRepositoryRedis) SetPrefix(prefix string) {
	k.prefix = prefix
}

func (k *kvRepositoryRedis) prefixKey(key string) string {
	if k.prefix != "" {
		return k.prefix + ":" + key
	}
	return key
}

func (k *kvRepositoryRedis) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return k.client.Set(ctx, k.prefixKey(key), data, ttl).Err()
}

func (k *kvRepositoryRedis) Get(ctx context.Context, key string, dest any) error {
	data, err := k.client.Get(ctx, k.prefixKey(key)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrKeyNotFound
		}
		return fmt.Errorf("redis get %q: %w", key, err)
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("unmarshal key %q: %w", key, err)
	}
	return nil
}

func (k *kvRepositoryRedis) Delete(ctx context.Context, key string) error {
	return k.client.Del(ctx, k.prefixKey(key)).Err()
}

func (k *kvRepositoryRedis) DeleteKeys(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = k.prefixKey(key)
	}
	return k.client.Del(ctx, prefixedKeys...).Err()
}

func (k *kvRepositoryRedis) Keys(ctx context.Context, pattern string) ([]string, error) {
	prefixedPattern := k.prefixKey(pattern)
	keys, err := k.client.Keys(ctx, prefixedPattern).Result()
	if err != nil {
		return nil, err
	}

	startItem := len(k.prefix)
	if startItem > 0 {
		startItem++ // to account for the colon
	}

	for i, key := range keys {
		if len(key) > startItem {
			keys[i] = key[startItem:]
		}
	}

	return keys, nil
}

func (k *kvRepositoryRedis) Shutdown() error {
	if k.client != nil {
		return k.client.Close()
	}
	return nil
}

func Service(cfg *Config) *kvRepositoryRedis {
	mu.Lock()
	client, exists := poolClient[*cfg]
	if !exists {
		client = redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
			PoolSize: cfg.PoolSize,
		})
		poolClient[*cfg] = client
	}
	mu.Unlock()

	return &kvRepositoryRedis{
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
		Prefix:   utils.GetValueFromMap(params, "prefix", "kv"),
	}
	return Service(cfg)
}

func Register() {
	lokstra_registry.RegisterServiceType(SERVICE_TYPE, ServiceFactory)
}
