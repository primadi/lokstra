package logger

import "lokstra/core"

// LoggerFactory creates a LoggerService from config
func LoggerFactory(serviceType string, cfg map[string]any) (core.Service, error) {
	levelStr := "info"
	if v, ok := cfg["logLevel"].(string); ok {
		levelStr = v
	}

	instanceName, _ := cfg["name"].(string)
	if instanceName == "" {
		instanceName = serviceType
	}

	level := parseLogLevel(levelStr)
	return NewLoggerService(instanceName, level)
}
