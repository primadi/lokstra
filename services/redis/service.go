package redis

import (
	"context"
	"fmt"
	"lokstra/common/iface"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisService struct {
	instanceName string
	client       *redis.Client
	config       map[string]any
}

func (r *RedisService) InstanceName() string {
	return r.instanceName
}

func (r *RedisService) GetConfig(key string) any {
	return r.config[key]
}

func (r *RedisService) GetClient() *redis.Client {
	return r.client
}

func (r *RedisService) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisService) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisService) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

func (r *RedisService) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

func (r *RedisService) Close() error {
	return r.client.Close()
}

func newRedisService(instanceName string, config map[string]any) (*RedisService, error) {
	addr, ok := config["addr"].(string)
	if !ok {
		return nil, fmt.Errorf("redis service requires 'addr' in config")
	}

	password := ""
	if p, ok := config["password"].(string); ok {
		password = p
	}

	db := 0
	if d, ok := config["db"].(int); ok {
		db = d
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisService{
		instanceName: instanceName,
		client:       client,
		config:       config,
	}, nil
}
