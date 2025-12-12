package logger

import (
	"os"
	"strings"
)

// LogLevel represents the logging level
type LogLevel int

const (
	LogLevelSilent LogLevel = iota
	LogLevelError
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
	LogLevelFromEnvi
)

var (
	activeBackend LoggerBackend = NewSlogBackend() // default slog
)

// LoggerBackend is the interface for logging backends.
// This allows replacing slog with zap, zerolog, etc in the future.
type LoggerBackend interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Panic(msg string, args ...any)
	SetLogLevel(level LogLevel)
	GetLogLevel() LogLevel
}

// SetBackend replaces the active logger backend
func SetBackend(backend LoggerBackend) {
	activeBackend = backend
}

// SetLogLevel sets the global log level
func SetLogLevel(level LogLevel) {
	if level == LogLevelFromEnvi {
		SetLogLevelFromEnv()
		return
	}
	activeBackend.SetLogLevel(level)
}

// GetLogLevel returns the current log level
func GetLogLevel() LogLevel {
	return activeBackend.GetLogLevel()
}

// SetLogLevelFromEnv sets log level from env var: LOKSTRA_LOG_LEVEL
func SetLogLevelFromEnv() {
	envLevel := strings.ToLower(os.Getenv("LOKSTRA_LOG_LEVEL"))
	switch envLevel {
	case "silent":
		SetLogLevel(LogLevelSilent)
	case "error":
		SetLogLevel(LogLevelError)
	case "warn", "warning":
		SetLogLevel(LogLevelWarn)
	case "info":
		SetLogLevel(LogLevelInfo)
	case "debug":
		SetLogLevel(LogLevelDebug)
	}
}

// Public wrapper functions (unchanged API)
func LogDebug(format string, args ...any)   { activeBackend.Debug(format, args...) }
func LogInfo(format string, args ...any)    { activeBackend.Info(format, args...) }
func LogWarn(format string, args ...any)    { activeBackend.Warn(format, args...) }
func LogWarning(format string, args ...any) { activeBackend.Warn(format, args...) }
func LogError(format string, args ...any)   { activeBackend.Error(format, args...) }
func LogPanic(format string, args ...any)   { activeBackend.Panic(format, args...) }
