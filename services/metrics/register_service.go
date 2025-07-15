package metrics

import (
	"fmt"
	"lokstra/common/iface"
	"lokstra/common/module"
)

type Registration struct{}

func (r *Registration) RegisterService(ctx module.RegistrationContext) {
	ctx.RegisterServiceFactory("metrics", ServiceFactory)
}

func ServiceFactory(config any) (iface.Service, error) {
	configMap, ok := config.(map[string]any)
	if !ok {
		configMap = make(map[string]any)
	}

	instanceName := "metrics"
	if name, ok := configMap["instance_name"].(string); ok {
		instanceName = name
	}

	return newMetricsService(instanceName, configMap)
}

func GetModule() iface.ServiceModule {
	return &ServiceModule{}
}

type ServiceModule struct{}

func (s *ServiceModule) Name() string {
	return "lokstra.metrics"
}

func (s *ServiceModule) Factory(config any) (iface.Service, error) {
	return ServiceFactory(config)
}

func (s *ServiceModule) Meta() *iface.ServiceMeta {
	return &iface.ServiceMeta{
		Description: "Prometheus metrics collection service for monitoring and observability",
		Tags:        []string{"metrics", "prometheus", "monitoring"},
	}
}
