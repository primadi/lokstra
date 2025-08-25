package kvstore_redis

import (
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/service"
	"github.com/redis/go-redis/v9"
)

const REDIS_SERVICENAME_KEY = "redis"

type module struct{}

// Description implements registration.Module.
func (m *module) Description() string {
	return "Redis key-value store service"
}

// Name implements registration.Module.
func (m *module) Name() string {
	return "kvstore_redis"
}

// Register implements registration.Module.
func (m *module) Register(regCtx registration.Context) error {
	factory := func(config any) (service.Service, error) {
		var redisServiceName string
		switch cfg := config.(type) {
		case string:
			redisServiceName = cfg
		case map[string]string:
			redisServiceName = cfg[REDIS_SERVICENAME_KEY]
		default:
			return nil, service.ErrUnsupportedConfig(config)
		}
		rs, err := regCtx.GetService(redisServiceName)
		if err != nil {
			return nil, err
		}
		redisClient, ok := rs.(*redis.Client)
		if !ok {
			return nil, service.ErrInvalidServiceType(redisServiceName, "redis.Client")
		}
		return New(redisClient), nil
	}

	regCtx.RegisterServiceFactory(m.Name(), factory)
	return nil
}

var _ registration.Module = (*module)(nil)

func GetModule() registration.Module {
	return &module{}
}
