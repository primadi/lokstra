package pg_dbpool

import (
	"fmt"
	"lokstra/common/iface"
	"lokstra/common/module"
)

const DSN_KEY = "dsn"

type Registration struct{}

// Register implements iface.Plugin.
func (r *Registration) Register(ctx module.RegistrationContext) error {
	ctx.RegisterServiceFactory("pg_dbpool", func(config any) (iface.Service, error) {
		var dsn string

		switch t := config.(type) {
		case string:
			dsn = t
		case map[string]any:
			if dk, ok := t[DSN_KEY].(string); ok {
				dsn = dk
			} else {
				return nil, fmt.Errorf("pg_dbpool requires a valid DSN in the configuration map")
			}
		default:
			return nil, fmt.Errorf("pg_dbpool requires a valid DSN as a string or a map with key '%s'", DSN_KEY)
		}

		return newPgxPostgresPool(dsn)
	})
	return nil
}
