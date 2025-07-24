package helloservice

import (
	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/service"
)

type HelloModule struct{}

// Name implements registration.Module.
func (h *HelloModule) Name() string {
	return "hello_service"
}

// Register implements registration.Module.
func (h *HelloModule) Register(regCtx registration.Context) error {
	regCtx.RegisterServiceFactory(h.Name(), factory)
	return nil
}

func factory(serviceName string, _ any) (service.Service, error) {
	return NewGreetingService(serviceName).(service.Service), nil
}

// Description implements service.Module.
func (h *HelloModule) Description() string {
	return "Hello Service Module provides a simple greeting service."
}

// Tags implements service.Module.
func (h *HelloModule) Tags() []string {
	return []string{"greeting", "example", "service"}
}

var _ lokstra.Module = (*HelloModule)(nil)
