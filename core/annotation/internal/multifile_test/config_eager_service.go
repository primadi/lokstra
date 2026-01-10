package main

import (
	"github.com/primadi/lokstra/core/service"
)

// Test eager injection (non-Cached types)
type DBPool struct {
	DSN string
}

type RedisClient struct {
	Addr string
}

// @EndpointService name="config-eager-service", prefix="/api/v1/config"
type ConfigEagerService struct {
	// @Inject "user-repository"
	UserRepo *service.Cached[any] // Lazy (wrapped in service.Cached)

	// @Inject "database"
	DB *DBPool // Eager (direct injection, not wrapped)

	// @Inject "redis"
	Redis *RedisClient // Eager (direct injection)
}

// @Route "GET /status"
func (s *ConfigEagerService) GetStatus() (map[string]any, error) {
	return map[string]any{
		"db_dsn":     s.DB.DSN,     // Direct access, no .MustGet()
		"redis_addr": s.Redis.Addr, // Direct access, no .MustGet()
	}, nil
}
