package registry

import (
	"errors"
	"lokstra/common/iface"
	"strings"
)

var serviceInstances = make(map[string]iface.Service) // map of serviceType to Service instance

func NewService(serviceType, name string, config ...any) (iface.Service, error) {
	if !strings.Contains(serviceType, ".") {
		serviceType = "main." + serviceType
	}

	if factory, exists := serviceFactories[serviceType]; exists {
		var cfg any
		if len(config) == 0 {
			cfg = nil
		} else if len(config) == 1 {
			cfg = config[0]
		} else {
			cfg = config
		}

		service, err := factory(cfg)
		if err != nil {
			return nil, err
		}

		serviceName := serviceType + ":" + name
		if _, found := serviceInstances[serviceName]; found {
			return nil, errors.New("service with name '" + serviceName + "' already exists")
		}
		serviceInstances[serviceName] = service
		return service, nil
	}
	return nil, errors.New("service factory not found for type: " + serviceType)
}

func GetService(name string) iface.Service {
	if !strings.Contains(name, ".") {
		name = "main." + name
	}
	if service, exists := serviceInstances[name]; exists {
		return service
	}
	return nil
}
