package dbpool_pg

import (
	"fmt"

	"github.com/primadi/lokstra/core/service"
)

const DSN_KEY = "dsn"

func ServiceFactory(serviceName string, config any) (service.Service, error) {
	var dsn string

	name := "default"
	switch t := config.(type) {
	case string:
		dsn = t
	case map[string]any:
		if nm, ok := t["name"].(string); ok {
			name = nm
		}
		if dk, ok := t[DSN_KEY].(string); ok {
			dsn = dk
		} else {
			return nil, fmt.Errorf("dbpool_pg requires a valid DSN in the configuration map")
		}
	case []string:
		if len(t) == 1 {
			dsn = t[0]
		} else if len(t) == 2 {
			name = t[0]
			dsn = t[1]
		} else {
			return nil, fmt.Errorf("dbpool_pg requires a valid DSN in the configuration slice")
		}
	default:
		return nil, fmt.Errorf("dbpool_pg requires a valid DSN in the configuration")
	}

	return NewPgxPostgresPool(name, dsn)
}
