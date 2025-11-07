package deploy

import (
	"fmt"
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
)

var (
	currentLogLevel = LogLevelInfo // Default log level
)

// SetLogLevel sets the global log level for the deploy package
func SetLogLevel(level LogLevel) {
	currentLogLevel = level
}

// GetLogLevel returns the current log level
func GetLogLevel() LogLevel {
	return currentLogLevel
}

// SetLogLevelFromEnv sets log level from environment variable LOKSTRA_LOG_LEVEL
// Supported values: silent, error, warn, info, debug
func SetLogLevelFromEnv() {
	envLevel := strings.ToLower(os.Getenv("LOKSTRA_LOG_LEVEL"))
	switch envLevel {
	case "silent":
		currentLogLevel = LogLevelSilent
	case "error":
		currentLogLevel = LogLevelError
	case "warn", "warning":
		currentLogLevel = LogLevelWarn
	case "info":
		currentLogLevel = LogLevelInfo
	case "debug":
		currentLogLevel = LogLevelDebug
	}
}

// LogDebug prints debug messages if log level is Debug or higher
func LogDebug(format string, args ...any) {
	if currentLogLevel >= LogLevelDebug {
		fmt.Printf("🐛 "+format+"\n", args...)
	}
}

// LogInfo prints info messages if log level is Info or higher
func LogInfo(format string, args ...any) {
	if currentLogLevel >= LogLevelInfo {
		fmt.Printf("ℹ️  "+format+"\n", args...)
	}
}

// LogWarn prints warning messages if log level is Warn or higher
func LogWarn(format string, args ...any) {
	if currentLogLevel >= LogLevelWarn {
		fmt.Printf("⚠️  "+format+"\n", args...)
	}
}

// LogError prints error messages if log level is Error or higher
func LogError(format string, args ...any) {
	if currentLogLevel >= LogLevelError {
		fmt.Printf("❌ "+format+"\n", args...)
	}
}
