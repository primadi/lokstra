package email

import (
	"fmt"
	"lokstra/common/iface"
	"lokstra/common/module"
)

type Registration struct{}

func (r *Registration) RegisterService(ctx module.RegistrationContext) {
	ctx.RegisterServiceFactory("email", ServiceFactory)
}

func ServiceFactory(config any) (iface.Service, error) {
	configMap, ok := config.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("email service requires configuration as map")
	}

	instanceName := "email"
	if name, ok := configMap["instance_name"].(string); ok {
		instanceName = name
	}

	return newEmailService(instanceName, configMap)
}

func GetModule() iface.ServiceModule {
	return &ServiceModule{}
}

type ServiceModule struct{}

func (s *ServiceModule) Name() string {
	return "lokstra.email"
}

func (s *ServiceModule) Factory(config any) (iface.Service, error) {
	return ServiceFactory(config)
}

func (s *ServiceModule) Meta() *iface.ServiceMeta {
	return &iface.ServiceMeta{
		Description: "SMTP email sending service for notifications and communications",
		Tags:        []string{"email", "smtp", "notifications"},
	}
}
