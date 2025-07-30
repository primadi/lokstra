package session_kvstore

import (
	"fmt"

	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/serviceapi"
)

type module struct{}

// Description implements registration.Module.
func (m *module) Description() string {
	return "Session key-value store service"
}

// Name implements registration.Module.
func (m *module) Name() string {
	return "session_kvstore"
}

// Register implements registration.Module.
func (m *module) Register(regCtx registration.Context) error {
	factory := func(config any) (service.Service, error) {
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
		return New(kvStore), nil
	}

	regCtx.RegisterServiceFactory(m.Name(), factory)
	return nil
}

var _ registration.Module = (*module)(nil)

func GetModule() registration.Module {
	return &module{}
}
