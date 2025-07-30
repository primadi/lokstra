package logger

import (
	"fmt"

	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/serviceapi"
)

const FACTORY_NAME = "logger"

type module struct{}

// Name implements registration.Module.
func (m *module) Name() string {
	return FACTORY_NAME
}

// Register implements registration.Module.
func (m *module) Register(regCtx registration.Context) error {
	factory := func(config any) (service.Service, error) {
		levelStr := "info"

		switch v := config.(type) {
		case map[string]any:
			if val, ok := v[serviceapi.ConfigKeyLogLevel]; ok {
				if str, ok := val.(string); ok {
					levelStr = str
				}
			}
		case string:
			levelStr = v
		}

		level, ok := serviceapi.ParseLogLevelSafe(levelStr)
		if !ok {
			fmt.Printf("Invalid log level '%s', defaulting to 'info'", levelStr)
		}
		return NewService(level)
	}

	regCtx.RegisterServiceFactory(m.Name(), factory)
	return nil
}

// Description implements service.Module.
func (m *module) Description() string {
	return "Logger Service for Lokstra"
}

// GetModule returns the logger service with serviceType "lokstra.logger".
func GetModule() registration.Module {
	return &module{}
}

var _ registration.Module = (*module)(nil)
