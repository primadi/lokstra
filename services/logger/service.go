package logger

import (
	"lokstra/serviceapi/logger_api"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// LoggerService implements iface.Service and Logger
type LoggerService struct {
	logger *zerolog.Logger
}

func NewService(level logger_api.LogLevel) logger_api.Logger {
	zerolog.TimeFieldFormat = time.RFC3339
	zlogger := zerolog.New(os.Stdout).Level(toZerologLevel(level)).With().Timestamp().Logger()

	return &LoggerService{
		logger: &zlogger,
	}
}

// GetLogLevel implements logger_api.Logger.
func (l *LoggerService) GetLogLevel() logger_api.LogLevel {
	return fromZerologLevel(l.logger.GetLevel())
}

// SetLogLevel implements logger_api.Logger.
func (l *LoggerService) SetLogLevel(level logger_api.LogLevel) {
	newLogger := l.logger.Level(toZerologLevel(level))
	l.logger = &newLogger
}

// WithField implements logger_api.Logger.
func (l *LoggerService) WithField(key string, value any) logger_api.Logger {
	newLogger := l.logger.With().Interface(key, value).Logger()
	return &LoggerService{
		logger: &newLogger,
	}
}

func toZerologLevel(lvl logger_api.LogLevel) zerolog.Level {
	switch lvl {
	case logger_api.LogLevelDebug:
		return zerolog.DebugLevel
	case logger_api.LogLevelInfo:
		return zerolog.InfoLevel
	case logger_api.LogLevelWarn:
		return zerolog.WarnLevel
	case logger_api.LogLevelError:
		return zerolog.ErrorLevel
	case logger_api.LogLevelFatal:
		return zerolog.FatalLevel
	case logger_api.LogLevelDisabled:
		return zerolog.Disabled
	default:
		return zerolog.InfoLevel
	}
}

func fromZerologLevel(lvl zerolog.Level) logger_api.LogLevel {
	switch lvl {
	case zerolog.DebugLevel:
		return logger_api.LogLevelDebug
	case zerolog.InfoLevel:
		return logger_api.LogLevelInfo
	case zerolog.WarnLevel:
		return logger_api.LogLevelWarn
	case zerolog.ErrorLevel:
		return logger_api.LogLevelError
	case zerolog.FatalLevel:
		return logger_api.LogLevelFatal
	case zerolog.Disabled:
		return logger_api.LogLevelDisabled
	default:
		return logger_api.LogLevelInfo
	}
}

func logWithFields(ev *zerolog.Event, msg string, fields ...logger_api.Field) {
	for _, f := range fields {
		ev = ev.Interface(f.Key, f.Value)
	}
	ev.Msg(msg)
}

// Debug implements logger_api.Logger.
func (l *LoggerService) Debug(msg string, fields ...logger_api.Field) {
	logWithFields(l.logger.Debug(), msg, fields...)
}

// Info implements logger_api.Logger.
func (l *LoggerService) Info(msg string, fields ...logger_api.Field) {
	logWithFields(l.logger.Info(), msg, fields...)
}

// Warn implements logger_api.Logger.
func (l *LoggerService) Warn(msg string, fields ...logger_api.Field) {
	logWithFields(l.logger.Warn(), msg, fields...)
}

// Error implements logger_api.Logger.
func (l *LoggerService) Error(msg string, fields ...logger_api.Field) {
	logWithFields(l.logger.Error(), msg, fields...)
}

// Fatal implements logger_api.Logger.
func (l *LoggerService) Fatal(msg string, fields ...logger_api.Field) {
	logWithFields(l.logger.Fatal(), msg, fields...)
}

var _ logger_api.Logger = (*LoggerService)(nil)
