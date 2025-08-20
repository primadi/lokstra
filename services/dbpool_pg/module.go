package dbpool_pg

import (
	"context"
	"fmt"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/iface"
	"github.com/primadi/lokstra/core/service"
)

// for single service module, module name equals service name
const MODULE_NAME = "lokstra.dbpool_pg"

type module struct{}

// Name implements registration.Module.
func (m *module) Name() string {
	return MODULE_NAME
}

// Register implements registration.Module.
func (m *module) Register(regCtx iface.RegistrationContext) error {
	factory := func(config any) (service.Service, error) {
		var dsn string

		switch t := config.(type) {
		case string:
			dsn = t
		case map[string]any:
			if dk, ok := t["dsn"].(string); ok {
				dsn = dk
			} else {
				// build dsn from parameter
				host := utils.GetValueFromMap(t, "host", "localhost")
				port := utils.GetValueFromMap(t, "port", 5432)
				database := utils.GetValueFromMap(t, "database", "")
				username := utils.GetValueFromMap(t, "username", "")
				password := utils.GetValueFromMap(t, "password", "")

				minConn := utils.GetValueFromMap(t, "min_connections", 0)
				maxConn := utils.GetValueFromMap(t, "max_connections", 4)
				maxIdleTime := utils.GetDurationFromMap(t, "max_idle_time", "30m")
				maxLifetime := utils.GetDurationFromMap(t, "max_lifetime", "1h")
				sslMode := utils.GetValueFromMap(t, "sslmode", "disable")

				dsn = fmt.Sprintf(
					"postgres://%s:%s@%s:%d/%s?sslmode=%s&pool_min_conns=%d&pool_max_conns=%d&pool_max_conn_idle_time=%s&pool_max_conn_lifetime=%s",
					username, password, host, port, database,
					sslMode, minConn, maxConn, maxIdleTime, maxLifetime)
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

		return NewPgxPostgresPool(context.Background(), dsn)
	}

	regCtx.RegisterServiceFactory(m.Name(), factory)
	return nil
}

// Description implements service.Module.
func (m *module) Description() string {
	return "PostgreSQL Database Pool Service Module"
}

var _ iface.Module = (*module)(nil)

func GetModule() iface.Module {
	return &module{}
}
