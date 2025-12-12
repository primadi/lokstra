package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

type ReadableHandler struct {
	Level LogLevel
	Out   *os.File
}

func (h *ReadableHandler) Enabled(_ context.Context, level slog.Level) bool {
	switch h.Level {
	case LogLevelSilent:
		return false
	case LogLevelError:
		return level >= slog.LevelError
	case LogLevelWarn:
		return level >= slog.LevelWarn
	case LogLevelInfo:
		return level >= slog.LevelInfo
	case LogLevelDebug:
		return level >= slog.LevelDebug
	}
	return true
}

func (h *ReadableHandler) Handle(_ context.Context, r slog.Record) error {
	timestamp := time.Now().Format("2006/01/02 15:04:05")

	level := "[" + strings.ToUpper(r.Level.String()) + "]"

	// message
	line := fmt.Sprintf("%s %s %s", timestamp, level, r.Message)

	// attributes → optional, currently ignored for simplicity
	// You may add key=value printing here if needed.

	_, err := fmt.Fprintln(h.Out, line)
	return err
}

func (h *ReadableHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *ReadableHandler) WithGroup(name string) slog.Handler {
	return h
}
