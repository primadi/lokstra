package db_pool

import (
	"fmt"
	"lokstra/core"
)

type Config struct {
	Backend string `yaml:"backend"`
	DSN     string `yaml:"dsn"`
}

func Register() {
	core.RegisterNamedService("db_pool", func(instanceName string, settings map[string]any) (core.Service, error) {
		backend, ok := settings["backend"].(string)
		if !ok || backend == "" {
			return nil, fmt.Errorf("db_pool backend not specified or invalid")
		}
		dsn, ok := settings["dsn"].(string)
		if !ok || dsn == "" {
			return nil, fmt.Errorf("db_pool DSN not specified or invalid")
		}

		switch backend {
		case "postgres":
			return newPgxPostgresPool(instanceName, dsn)
		default:
			return nil, fmt.Errorf("unsupported db_pool backend: %s", backend)
		}
	})
}
