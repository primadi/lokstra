package kvstore_redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/redis/go-redis/v9"
)

const SERVICE_TYPE = "kvstore_redis"

// Config represents the configuration for Redis-based KvStore service.
type Config struct {
	Addr     string `json:"addr" yaml:"addr"`         // host:port address
	Password string `json:"password" yaml:"password"` // password
	DB       int    `json:"db" yaml:"db"`             // database number
	PoolSize int    `json:"pool_size" yaml:"pool_size"`
	Prefix   string `json:"prefix" yaml:"prefix"` // key prefix for namespacing
}

type kvStoreRedis struct {
	client *redis.Client
	prefix string
}

var _ serviceapi.KvStore = (*kvStoreRedis)(nil)

func (k *kvStoreRedis) prefixKey(key string) string {
	if k.prefix != "" {
		return k.prefix + ":" + key
	}
	return key
}

func (k *kvStoreRedis) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return k.client.Set(ctx, k.prefixKey(key), data, ttl).Err()
}

func (k *kvStoreRedis) Get(ctx context.Context, key string, dest any) error {
	data, err := k.client.Get(ctx, k.prefixKey(key)).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (k *kvStoreRedis) Delete(ctx context.Context, key string) error {
	return k.client.Del(ctx, k.prefixKey(key)).Err()
}

func (k *kvStoreRedis) DeleteKeys(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = k.prefixKey(key)
	}
	return k.client.Del(ctx, prefixedKeys...).Err()
}

func (k *kvStoreRedis) Keys(ctx context.Context, pattern string) ([]string, error) {
	prefixedPattern := k.prefixKey(pattern)
	keys, err := k.client.Keys(ctx, prefixedPattern).Result()
	if err != nil {
		return nil, err
	}

	// Remove prefix from returned keys
	if k.prefix != "" {
		prefixLen := len(k.prefix) + 1 // +1 for the colon
		for i, key := range keys {
			if len(key) > prefixLen {
				keys[i] = key[prefixLen:]
			}
		}
	}

	return keys, nil
}

func (k *kvStoreRedis) Shutdown() error {
	if k.client != nil {
		return k.client.Close()
	}
	return nil
}

func Service(cfg *Config) *kvStoreRedis {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})
	return &kvStoreRedis{
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
