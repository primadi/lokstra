package logger

import (
	"fmt"
	"log/slog"
	"os"
)

type SlogBackend struct {
	logger *slog.Logger
	level  LogLevel
}

func NewSlogBackend() *SlogBackend {
	b := &SlogBackend{level: LogLevelInfo}
	b.rebuildLogger()
	return b
}

func (b *SlogBackend) SetLogLevel(level LogLevel) {
	// If user passes LogLevelFromEnvi directly → load env level
	if level == LogLevelFromEnvi {
		// Call global logic to resolve env variable
		SetLogLevelFromEnv()
		return
	}

	b.level = level
	b.rebuildLogger()
}

func (b *SlogBackend) GetLogLevel() LogLevel {
	return b.level
}

func (b *SlogBackend) Debug(format string, args ...any) {
	if b.level >= LogLevelDebug {
		b.logger.Debug(fmt.Sprintf(format, args...))
	}
}

func (b *SlogBackend) Info(format string, args ...any) {
	if b.level >= LogLevelInfo {
		b.logger.Info(fmt.Sprintf(format, args...))
	}
}

func (b *SlogBackend) Warn(format string, args ...any) {
	if b.level >= LogLevelWarn {
		b.logger.Warn(fmt.Sprintf(format, args...))
	}
}

func (b *SlogBackend) Error(format string, args ...any) {
	if b.level >= LogLevelError {
		b.logger.Error(fmt.Sprintf(format, args...))
	}
}

func (b *SlogBackend) Panic(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	b.logger.Error(msg)
	panic(msg)
}

func (b *SlogBackend) rebuildLogger() {
	handler := &ReadableHandler{
		Level: b.level,
		Out:   os.Stdout,
	}

	b.logger = slog.New(handler)
}
