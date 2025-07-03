package logger

import (
	"fmt"
	"lokstra/common/iface"
	"lokstra/serviceapi/logger_api"
)

// ServiceFactory creates a LoggerServiceFactory from config
// It expects a map with a "log-level" key that specifies the logging level.
// If "log-level" is not provided, it defaults to "info".
func ServiceFactory(cfg any) (iface.Service, error) {
	levelStr := "info"
	switch v := cfg.(type) {
	case map[string]any:
		if val, ok := v[logger_api.ConfigKeyLogLevel]; ok {
			if str, ok := val.(string); ok {
				levelStr = str
			}
		}
	case map[string]string:
		if v, ok := v[logger_api.ConfigKeyLogLevel]; ok {
			levelStr = v
		}
	case string:
		levelStr = v
	}

	level, ok := logger_api.ParseLogLevelSafe(levelStr)
	if !ok {
		fmt.Printf("Invalid log level '%s', defaulting to 'info'", levelStr)
	}
	return NewService(level), nil
}
