package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

func parseLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return LogLevelDebug
	case "info":
		return LogLevelInfo
	case "warn", "warning":
		return LogLevelWarn
	case "error":
		return LogLevelError
	case "fatal":
		return LogLevelFatal
	case "disabled", "off":
		return LogLevelDisabled
	default:
		return LogLevelInfo
	}
}

func toZerologLevel(lvl LogLevel) zerolog.Level {
	switch lvl {
	case LogLevelDebug:
		return zerolog.DebugLevel
	case LogLevelInfo:
		return zerolog.InfoLevel
	case LogLevelWarn:
		return zerolog.WarnLevel
	case LogLevelError:
		return zerolog.ErrorLevel
	case LogLevelFatal:
		return zerolog.FatalLevel
	case LogLevelDisabled:
		return zerolog.Disabled
	default:
		return zerolog.InfoLevel
	}
}

func NewLoggerService(instanceName string, level LogLevel) (*LoggerService, error) {
	zerolog.TimeFieldFormat = time.RFC3339
	zlogger := zerolog.New(os.Stdout).Level(toZerologLevel(level)).With().Timestamp().Logger()

	return &LoggerService{
		name:   instanceName,
		level:  level,
		logger: &zlogger,
	}, nil
}

// iface.Service
func (l *LoggerService) InstanceName() string { return l.name }
func (l *LoggerService) GetConfig(key string) any {
	if key == "level" {
		return l.level
	}
	return nil
}

// Logger methods
func logWithFields(ev *zerolog.Event, msg string, fields ...Field) {
	for _, f := range fields {
		ev = ev.Interface(f.Key, f.Value)
	}
	ev.Msg(msg)
}

func (l *LoggerService) Debug(msg string, fields ...Field) {
	logWithFields(l.logger.Debug(), msg, fields...)
}

func (l *LoggerService) Info(msg string, fields ...Field) {
	logWithFields(l.logger.Info(), msg, fields...)
}

func (l *LoggerService) Warn(msg string, fields ...Field) {
	logWithFields(l.logger.Warn(), msg, fields...)
}

func (l *LoggerService) Error(msg string, fields ...Field) {
	logWithFields(l.logger.Error(), msg, fields...)
}

func (l *LoggerService) Fatal(msg string, fields ...Field) {
	logWithFields(l.logger.Fatal(), msg, fields...)
}
