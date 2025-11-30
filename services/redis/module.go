package redis

import (
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/redis/go-redis/v9"
)

const SERVICE_TYPE = "redis"

// Config represents the configuration for Redis client service.
type Config struct {
	Addr     string `json:"addr" yaml:"addr"`         // host:port address
	Password string `json:"password" yaml:"password"` // password
	DB       int    `json:"db" yaml:"db"`             // database number
	PoolSize int    `json:"pool_size" yaml:"pool_size"`
}

type redisService struct {
	client *redis.Client
}

func (r *redisService) Client() *redis.Client {
	return r.client
}

func (r *redisService) Shutdown() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

func Service(cfg *Config) *redisService {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})
	return &redisService{client: client}
}

func ServiceFactory(params map[string]any) any {
	cfg := &Config{
		Addr:     utils.GetValueFromMap(params, "addr", "localhost:6379"),
		Password: utils.GetValueFromMap(params, "password", ""),
		DB:       utils.GetValueFromMap(params, "db", 0),
		PoolSize: utils.GetValueFromMap(params, "pool_size", 10),
	}
	return Service(cfg)
}

func Register() {
	lokstra_registry.RegisterServiceType(SERVICE_TYPE, ServiceFactory)
}
