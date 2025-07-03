package logger_api

import "lokstra"

type LogLevel int

const ConfigKeyLogLevel = "log-level"

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
	LogLevelDisabled
)

func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "debug"
	case LogLevelInfo:
		return "info"
	case LogLevelWarn:
		return "warn"
	case LogLevelError:
		return "error"
	case LogLevelFatal:
		return "fatal"
	case LogLevelDisabled:
		return "disabled"
	default:
		return "unknown"
	}
}

func ParseLogLevelSafe(levelStr string) (LogLevel, bool) {
	switch levelStr {
	case "debug":
		return LogLevelDebug, true
	case "info":
		return LogLevelInfo, true
	case "warn", "warning":
		return LogLevelWarn, true
	case "error", "err":
		return LogLevelError, true
	case "fatal":
		return LogLevelFatal, true
	case "disabled", "off", "none":
		return LogLevelDisabled, true
	default:
		return LogLevelInfo, false // Default to Info if unknown
	}
}

// Logger defines logging interface used across services
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)

	GetLogLevel() LogLevel
	SetLogLevel(level LogLevel)
	WithField(key string, value any) Logger
}

// Field is a key-value pair for structured logging
type Field struct {
	Key   string
	Value any
}

func NewService(serviceType, name string, config map[string]any) (Logger, error) {
	service, err := lokstra.NewService[Logger](serviceType, name, config)
	if err != nil {
		return nil, err
	}
	return service, nil
}
