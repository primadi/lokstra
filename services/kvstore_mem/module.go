package kvstore_mem

import (
	"github.com/primadi/lokstra/core/iface"
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/service"
)

type module struct{}

// Description implements registration.Module.
func (m *module) Description() string {
	return "In-memory key-value store service"
}

// Name implements registration.Module.
func (m *module) Name() string {
	return "kvstore_mem"
}

// Register implements registration.Module.
func (m *module) Register(regCtx iface.RegistrationContext) error {
	regCtx.RegisterServiceFactory(m.Name(),
		func(_ any) (service.Service, error) {
			return New(), nil
		})
	return nil
}

var _ registration.Module = (*module)(nil)

func GetModule() registration.Module {
	return &module{}
}
