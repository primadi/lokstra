package serviceapi

type LogLevel int

const ConfigKeyLogLevel = "log_level"
const ConfigKeyLogFormat = "log_format"
const ConfigKeyLogOutput = "log_output"

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

type LogFields = map[string]any

// Logger defines logging interface used across services
type Logger interface {
	Debugf(msg string, v ...any)
	Infof(msg string, v ...any)
	Warnf(msg string, v ...any)
	Errorf(msg string, v ...any)
	Fatalf(msg string, v ...any)

	GetLogLevel() LogLevel
	SetLogLevel(level LogLevel)
	WithField(key string, value any) Logger
	WithFields(fields LogFields) Logger

	SetFormat(format string)
	SetOutput(output string)
}
