package dbpool_pg

import (
	"fmt"

	"github.com/primadi/lokstra/core/service"
)

const DSN_KEY = "dsn"

func ServiceFactory(config any) (service.Service, error) {
	var dsn string

	switch t := config.(type) {
	case string:
		dsn = t
	case map[string]any:
		if dk, ok := t[DSN_KEY].(string); ok {
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
