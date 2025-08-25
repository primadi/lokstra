package redis

import (
	"fmt"

	"github.com/primadi/lokstra/common/utils"

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

		switch cfg := config.(type) {
		case string:
			dsn = cfg
		case map[string]any:
			dsn = utils.GetValueFromMap(cfg, "dsn", "")
			if dsn == "" {
				addr := utils.GetValueFromMap(cfg, "addr", "")
				if addr == "" {
					host := utils.GetValueFromMap(cfg, "host", "localhost")
					port := utils.GetValueFromMap(cfg, "port", "6379")
					addr = fmt.Sprintf("%s:%s", host, port)
				}
				db := utils.GetValueFromMap(cfg, "db", 0)
				username := utils.GetValueFromMap(cfg, "username", "")
				password := utils.GetValueFromMap(cfg, "password", "")

				auth := ""
				if username != "" || password != "" {
					auth = fmt.Sprintf("%s:%s@", username, password)
				}
				dsn = fmt.Sprintf("redis://%s%s/%d", auth, addr, db)
			}
		default:
			return nil, service.ErrUnsupportedConfig(config)
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
