package redis

import (
	"fmt"
	"lokstra/common/iface"
	"lokstra/common/module"
)

type Registration struct{}

func (r *Registration) RegisterService(ctx module.RegistrationContext) {
	ctx.RegisterServiceFactory("redis", ServiceFactory)
}

func ServiceFactory(config any) (iface.Service, error) {
	configMap, ok := config.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("redis service requires configuration as map")
	}

	instanceName := "redis"
	if name, ok := configMap["instance_name"].(string); ok {
		instanceName = name
	}

	return newRedisService(instanceName, configMap)
}

func GetModule() iface.ServiceModule {
	return &ServiceModule{}
}

type ServiceModule struct{}

func (s *ServiceModule) Name() string {
	return "lokstra.redis"
}

func (s *ServiceModule) Factory(config any) (iface.Service, error) {
	return ServiceFactory(config)
}

func (s *ServiceModule) Meta() *iface.ServiceMeta {
	return &iface.ServiceMeta{
		Description: "Redis connection pool service for caching and data storage",
		Tags:        []string{"redis", "cache", "storage"},
	}
}
