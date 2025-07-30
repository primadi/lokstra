package redis

import (
	"fmt"

	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/service"
)

type module struct{}

// Description implements registration.Module.
func (m *module) Description() string {
	return "Redis service module"
}

// Name implements registration.Module.
func (m *module) Name() string {
	return "redis"
}

// Register implements registration.Module.
func (m *module) Register(regCtx registration.Context) error {
	factory := func(config any) (service.Service, error) {
		var dsn string
		var addr, db, password string

		switch cfg := config.(type) {
		case string:
			dsn = cfg
		case map[string]string:
			dsn = cfg["dsn"]
			addr = cfg["addr"]
			db = cfg["db"]
			password = cfg["password"]
		default:
			return nil, service.ErrUnsupportedConfig(config)
		}
		if dsn == "" {
			if password != "" {
				dsn = fmt.Sprintf("redis://:%s@%s/%s", password, addr, db)
			} else {
				dsn = fmt.Sprintf("redis://%s/%s", addr, db)
			}
		}
		return NewDsn(dsn), nil
	}

	regCtx.RegisterServiceFactory(m.Name(), factory)
	return nil
}

var _ registration.Module = (*module)(nil)

func GetModule() registration.Module {
	return &module{}
}
