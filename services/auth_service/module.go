package auth_service

import (
	"fmt"

	"github.com/primadi/lokstra/core/iface"
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/serviceapi"
)

const FACTORY_NAME = "auth_service.session.kvstore"

type module struct{}

// Description implements registration.Module.
func (m *module) Description() string {
	return "Auth Service Module provides authentication and session management services."
}

// Name implements registration.Module.
func (m *module) Name() string {
	return "auth_service"
}

// Register implements registration.Module.
func (m *module) Register(regCtx iface.RegistrationContext) error {
	factorySession := func(config any) (service.Service, error) {
		var kvStoreName string
		switch cfg := config.(type) {
		case string:
			kvStoreName = cfg
		case map[string]string:
			kvStoreName = cfg["kvstore"]
		default:
			return nil, service.ErrUnsupportedConfig(config)
		}
		if kvStoreName == "" {
			return nil, fmt.Errorf("kvstore name must be provided")
		}
		kvs, err := regCtx.GetService(kvStoreName)
		if err != nil {
			return nil, err
		}
		kvStore, ok := kvs.(serviceapi.KvStore)
		if !ok {
			return nil, service.ErrInvalidServiceType(kvStoreName, "serviceapi.KvStore")
		}
		return NewSession(kvStore), nil
	}

	regCtx.RegisterServiceFactory(FACTORY_NAME, factorySession)
	return nil
}

var _ registration.Module = (*module)(nil)

func GetModule() registration.Module {
	return &module{}
}
