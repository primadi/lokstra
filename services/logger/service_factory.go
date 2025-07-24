package logger

import (
	"fmt"

	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/serviceapi"
)

func ServiceFactory(serviceName string, config any) (service.Service, error) {
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
