package logger

import (
	"fmt"

	"github.com/primadi/lokstra/common/utils"

	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/serviceapi"
)

const MODULE_NAME = "lokstra.logger"

type module struct{}

// Name implements registration.Module.
func (m *module) Name() string {
	return MODULE_NAME
}

// Register implements registration.Module.
func (m *module) Register(regCtx registration.Context) error {
	factory := func(config any) (service.Service, error) {
		loggerConfig := &LoggerConfig{
			Level:  serviceapi.LogLevelInfo,
			Format: "text",
			Output: "stdout",
		}

		switch v := config.(type) {
		case map[string]any:
			// Parse level
			levelStr := utils.GetValueFromMap(v, "level", "info")
			if level, ok := serviceapi.ParseLogLevelSafe(levelStr); ok {
				loggerConfig.Level = level
			} else {
				fmt.Printf("Invalid log level '%s', defaulting to 'info'\n", levelStr)
			}

			// Parse format
			if format := utils.GetValueFromMap(v, "format", "text"); format != "" {
				loggerConfig.Format = format
			}

			// Parse output
			if output := utils.GetValueFromMap(v, "output", "stdout"); output != "" {
				loggerConfig.Output = output
			}

		case string:
			// Legacy support: just level as string
			if level, ok := serviceapi.ParseLogLevelSafe(v); ok {
				loggerConfig.Level = level
			} else {
				fmt.Printf("Invalid log level '%s', defaulting to 'info'\n", v)
			}
		}

		return NewServiceWithConfig(loggerConfig)
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
