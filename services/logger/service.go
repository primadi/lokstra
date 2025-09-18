package logger

import (
	"io"
	"os"
	"time"

	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/serviceapi"

	"github.com/rs/zerolog"
)

// LoggerService implements iface.Service and Logger
type LoggerService struct {
	logger *zerolog.Logger
	writer io.Writer
}

// LoggerConfig holds configuration for logger service
type LoggerConfig struct {
	Level  serviceapi.LogLevel
	Format string // "json", "text", "console"
	Output string // "stdout", "stderr", "file"
}

func NewService(level serviceapi.LogLevel) (*LoggerService, error) {
	return NewServiceWithConfig(&LoggerConfig{
		Level:  level,
		Format: "json",
		Output: "stdout",
	})
}

func NewServiceWithConfig(config *LoggerConfig) (*LoggerService, error) {
	zerolog.TimeFieldFormat = time.RFC3339

	// Determine output writer
	var writer io.Writer
	switch config.Output {
	case "stderr":
		writer = os.Stderr
	case "stdout":
		writer = os.Stdout
	default:
		writer = os.Stdout
	}

	// Configure format
	switch config.Format {
	case "text", "console":
		// Use ConsoleWriter for human-readable output
		consoleWriter := zerolog.ConsoleWriter{
			Out:        writer,
			TimeFormat: time.RFC3339,
		}
		writer = consoleWriter
	case "json":
		// Keep default JSON format
		// writer is already set above
	default:
		// Default to JSON format
		// writer is already set above
	}

	zlogger := zerolog.New(writer).Level(toZerologLevel(config.Level)).
		With().Timestamp().Logger()

	return &LoggerService{
		logger: &zlogger,
		writer: writer,
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

// SetFormat implements serviceapi.Logger.
func (l *LoggerService) SetFormat(format string) {
	// Get current level to maintain it
	currentLevel := l.logger.GetLevel()

	// Determine the base writer (extract from ConsoleWriter if needed)
	var baseWriter io.Writer
	if cw, ok := l.writer.(zerolog.ConsoleWriter); ok {
		baseWriter = cw.Out
	} else {
		baseWriter = l.writer
	}

	var writer io.Writer
	// Configure format based on the new format
	switch format {
	case "text", "console":
		// Use ConsoleWriter for human-readable output
		consoleWriter := zerolog.ConsoleWriter{
			Out:        baseWriter,
			TimeFormat: time.RFC3339,
			NoColor:    false, // Enable color for console format
		}
		writer = consoleWriter
	case "json":
		// Use base writer for pure JSON output (no console formatting)
		writer = baseWriter
	default:
		// Default to JSON format
		writer = baseWriter
	}

	// Create new logger with the new format but same level
	newLogger := zerolog.New(writer).Level(currentLevel).
		With().Timestamp().Logger()
	l.logger = &newLogger
	l.writer = writer
}

// SetOutput implements serviceapi.Logger.
func (l *LoggerService) SetOutput(output string) {
	var writer io.Writer
	switch output {
	case "stderr":
		writer = os.Stderr
	case "stdout":
		writer = os.Stdout
	default:
		writer = os.Stdout
	}
	l.writer = writer
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
