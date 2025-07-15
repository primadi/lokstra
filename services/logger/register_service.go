package logger

import (
	"fmt"
	"lokstra/common/iface"
	"lokstra/serviceapi"
)

const NAME = "lokstra.logger"

type LoggerServiceModule struct{}

// Factory implements iface.ServiceModule.
func (l *LoggerServiceModule) Factory(config any) (iface.Service, error) {
	levelStr := "info"
	switch v := config.(type) {
	case map[string]any:
		if val, ok := v[serviceapi.ConfigKeyLogLevel]; ok {
			if str, ok := val.(string); ok {
				levelStr = str
			}
		}
	case map[string]string:
		if v, ok := v[serviceapi.ConfigKeyLogLevel]; ok {
			levelStr = v
		}
	case string:
		levelStr = v
	}

	level, ok := serviceapi.ParseLogLevelSafe(levelStr)
	if !ok {
		fmt.Printf("Invalid log level '%s', defaulting to 'info'", levelStr)
	}
	return NewService(level), nil
}

// Meta implements iface.ServiceModule.
func (l *LoggerServiceModule) Meta() *iface.ServiceMeta {
	return &iface.ServiceMeta{
		Description: "Logger service for Lokstra",
		Tags:        []string{"logging", "service"},
	}
}

// Name implements iface.ServiceModule.
func (l *LoggerServiceModule) Name() string {
	return NAME
}

var _ iface.ServiceModule = (*LoggerServiceModule)(nil)

// GetModule returns the logger service with serviceType "lokstra.logger".
func GetModule() iface.ServiceModule {
	return &LoggerServiceModule{}
}
