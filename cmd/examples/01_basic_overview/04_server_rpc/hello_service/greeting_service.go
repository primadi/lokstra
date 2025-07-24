package hello_service

import "github.com/primadi/lokstra/core/service"

type GreetingService interface {
	Hello(name string) (string, error)
}

type GreetingServiceImpl struct {
	*service.BaseService
}

// GetServiceUri implements service.Service.
func (h *GreetingServiceImpl) GetServiceUri() string {
	return "lokstra://helloservice.greeting_service/" + h.GetServiceName()
}

// Hello implements GreetingServiceImpl.
func (h *GreetingServiceImpl) Hello(name string) (string, error) {
	return "Hello, " + name + "!", nil
}

var _ GreetingService = (*GreetingServiceImpl)(nil)
var _ service.Service = (*GreetingServiceImpl)(nil)

func NewGreetingService(serviceName string) GreetingService {
	return &GreetingServiceImpl{
		BaseService: service.NewBaseService(serviceName),
	}
}
