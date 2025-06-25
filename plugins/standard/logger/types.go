package logger

import "github.com/rs/zerolog"

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
	LogLevelDisabled
)

// Logger defines logging interface used across services
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
}

// Field is a key-value pair for structured logging
type Field struct {
	Key   string
	Value any
}

// LoggerService implements iface.Service and Logger
type LoggerService struct {
	name   string
	level  LogLevel
	logger *zerolog.Logger
}
