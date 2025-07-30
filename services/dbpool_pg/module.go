package dbpool_pg

import (
	"fmt"

	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/service"
)

type module struct{}

// Name implements registration.Module.
func (m *module) Name() string {
	return "dbpool_pg"
}

// Register implements registration.Module.
func (m *module) Register(regCtx registration.Context) error {
	factory := func(config any) (service.Service, error) {
		var dsn string

		switch t := config.(type) {
		case string:
			dsn = t
		case map[string]any:
			if dk, ok := t["dsn"].(string); ok {
				dsn = dk
			} else {
				return nil, fmt.Errorf("dbpool_pg requires a valid DSN in the configuration map")
			}
		case []string:
			if len(t) == 1 {
				dsn = t[0]
			} else {
				return nil, fmt.Errorf("dbpool_pg requires a valid DSN in the configuration slice")
			}
		default:
			return nil, fmt.Errorf("dbpool_pg requires a valid DSN in the configuration")
		}

		return NewPgxPostgresPool(dsn)
	}

	regCtx.RegisterServiceFactory(m.Name(), factory)
	return nil
}

// Description implements service.Module.
func (m *module) Description() string {
	return "PostgreSQL Database Pool Service Module"
}

var _ registration.Module = (*module)(nil)

func GetModule() registration.Module {
	return &module{}
}
