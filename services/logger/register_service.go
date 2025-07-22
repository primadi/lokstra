package logger

import (
	"fmt"

	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/serviceapi"
)

const FACTORY_NAME = "logger"

type LoggerServiceModule struct{}

// FactoryName implements service.ServiceModule.
func (l *LoggerServiceModule) FactoryName() string {
	return FACTORY_NAME
}

// Factory implements iface.ServiceModule.
func (l *LoggerServiceModule) Factory(serviceName string, config any) (service.Service, error) {
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
	return NewService(serviceName, level)
}

// Meta implements iface.ServiceModule.
func (l *LoggerServiceModule) Meta() *service.ServiceMeta {
	return &service.ServiceMeta{
		Description: "Logger service for Lokstra",
		Tags:        []string{"logging", "service"},
	}
}

// GetModule returns the logger service with serviceType "lokstra.logger".
func GetModule() service.ServiceModule {
	return &LoggerServiceModule{}
}

var _ service.ServiceModule = (*LoggerServiceModule)(nil)
