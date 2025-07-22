package logger

import (
	"os"
	"time"

	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/serviceapi"

	"github.com/rs/zerolog"
)

// LoggerService implements iface.Service and Logger
type LoggerService struct {
	*service.BaseService
	logger *zerolog.Logger
}

// GetServiceUri implements service.Service.
func (l *LoggerService) GetServiceUri() string {
	return "lokstra://logger/" + l.GetServiceName()
}

func NewService(name string, level serviceapi.LogLevel) (*LoggerService, error) {
	zerolog.TimeFieldFormat = time.RFC3339
	zlogger := zerolog.New(os.Stdout).Level(toZerologLevel(level)).With().Timestamp().Logger()

	return &LoggerService{
		BaseService: service.NewBaseService(name),
		logger:      &zlogger,
	}, nil
}

// GetLogLevel implements serviceapi.Logger.
func (l *LoggerService) GetLogLevel() serviceapi.LogLevel {
	return fromZerologLevel(l.logger.GetLevel())
}

// SetLogLevel implements serviceapi.Logger.
func (l *LoggerService) SetLogLevel(level serviceapi.LogLevel) {
	newLogger := l.logger.Level(toZerologLevel(level))
	l.logger = &newLogger
}

// WithField implements serviceapi.Logger.
func (l *LoggerService) WithField(key string, value any) serviceapi.Logger {
	newLogger := l.logger.With().Interface(key, value).Logger()
	return &LoggerService{
		logger: &newLogger,
	}
}

// WithFields implements serviceapi.Logger.
func (l *LoggerService) WithFields(LogFields serviceapi.LogFields) serviceapi.Logger {
	lctx := l.logger.With()
	for key, value := range LogFields {
		lctx = lctx.Interface(key, value)
	}
	newLogger := lctx.Logger()
	return &LoggerService{
		logger: &newLogger,
	}
}

func toZerologLevel(lvl serviceapi.LogLevel) zerolog.Level {
	switch lvl {
	case serviceapi.LogLevelDebug:
		return zerolog.DebugLevel
	case serviceapi.LogLevelInfo:
		return zerolog.InfoLevel
	case serviceapi.LogLevelWarn:
		return zerolog.WarnLevel
	case serviceapi.LogLevelError:
		return zerolog.ErrorLevel
	case serviceapi.LogLevelFatal:
		return zerolog.FatalLevel
	case serviceapi.LogLevelDisabled:
		return zerolog.Disabled
	default:
		return zerolog.InfoLevel
	}
}

func fromZerologLevel(lvl zerolog.Level) serviceapi.LogLevel {
	switch lvl {
	case zerolog.DebugLevel:
		return serviceapi.LogLevelDebug
	case zerolog.InfoLevel:
		return serviceapi.LogLevelInfo
	case zerolog.WarnLevel:
		return serviceapi.LogLevelWarn
	case zerolog.ErrorLevel:
		return serviceapi.LogLevelError
	case zerolog.FatalLevel:
		return serviceapi.LogLevelFatal
	case zerolog.Disabled:
		return serviceapi.LogLevelDisabled
	default:
		return serviceapi.LogLevelInfo
	}
}

// Debug implements serviceapi.Logger.
func (l *LoggerService) Debugf(msg string, v ...any) {
	l.logger.Debug().Msgf(msg, v...)
}

// Info implements serviceapi.Logger.
func (l *LoggerService) Infof(msg string, v ...any) {
	l.logger.Info().Msgf(msg, v...)
}

// Warn implements serviceapi.Logger.
func (l *LoggerService) Warnf(msg string, v ...any) {
	l.logger.Warn().Msgf(msg, v...)
}

// Error implements serviceapi.Logger.
func (l *LoggerService) Errorf(msg string, v ...any) {
	l.logger.Error().Msgf(msg, v...)
}

// Fatal implements serviceapi.Logger.
func (l *LoggerService) Fatalf(msg string, v ...any) {
	l.logger.Fatal().Msgf(msg, v...)
}

var _ serviceapi.Logger = (*LoggerService)(nil)
var _ service.Service = (*LoggerService)(nil)
