package request_logger

import (
	"fmt"
	"log"
	"time"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/old_registry"
)

const REQUEST_LOGGER_TYPE = "request_logger"
const PARAMS_ENABLE_COLORS = "enable_colors"
const PARAMS_SKIP_PATHS = "skip_paths"

type Config struct {
	// EnableColors enables colored output for terminal
	EnableColors bool

	// SkipPaths is a list of paths to skip logging
	// Example: ["/health", "/metrics"]
	SkipPaths []string

	// CustomLogger is a custom logging function
	// If nil, uses default log.Printf
	CustomLogger func(format string, args ...any)
}

func DefaultConfig() *Config {
	return &Config{
		EnableColors: true,
		SkipPaths:    []string{},
		CustomLogger: nil,
	}
}

// ANSI color codes
const (
	colorReset = "\033[0m"
	// colorRed    = "\033[31m"
	// colorGreen  = "\033[32m"
	// colorYellow = "\033[33m"
	// colorBlue   = "\033[34m"
	colorCyan = "\033[36m"
	colorGray = "\033[90m"
)

// logs all incoming HTTP requests
func Middleware(cfg *Config) request.HandlerFunc {
	defConfig := DefaultConfig()
	if cfg.SkipPaths == nil {
		cfg.SkipPaths = defConfig.SkipPaths
	}
	if cfg.CustomLogger == nil {
		cfg.CustomLogger = log.Printf
	}

	return request.HandlerFunc(func(c *request.Context) error {
		// Check if path should be skipped
		requestPath := c.R.URL.Path
		for _, skipPath := range cfg.SkipPaths {
			if requestPath == skipPath {
				return c.Next()
			}
		}

		// Record start time
		start := time.Now()

		// Call next handler
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get status code from writer wrapper
		statusCode := c.W.StatusCode()
		if statusCode == 0 {
			statusCode = 200 // Default if not set
		}

		// Format and log request
		if cfg.EnableColors {
			msg := fmt.Sprintf("%s%s%s %s %s%d %s%s",
				colorCyan,
				c.R.Method,
				colorReset,
				c.R.URL.Path,
				colorGray,
				statusCode,
				formatDuration(duration),
				colorReset,
			)
			cfg.CustomLogger("%s", msg)
		} else {
			msg := fmt.Sprintf("[%s] %s - Status: %d - Duration: %s",
				c.R.Method,
				c.R.URL.Path,
				statusCode,
				formatDuration(duration),
			)
			cfg.CustomLogger("%s", msg)
		}

		return err
	})
}

func MiddlewareFactory(params map[string]any) request.HandlerFunc {
	defConfig := DefaultConfig()
	if params == nil {
		return Middleware(defConfig)
	}

	cfg := &Config{
		EnableColors: utils.GetValueFromMap(params, PARAMS_ENABLE_COLORS, defConfig.EnableColors),
		SkipPaths:    utils.GetValueFromMap(params, PARAMS_SKIP_PATHS, defConfig.SkipPaths),
		CustomLogger: nil, // Cannot be set via params
	}
	return Middleware(cfg)
}

func Register() {
	old_registry.RegisterMiddlewareFactory(REQUEST_LOGGER_TYPE, MiddlewareFactory,
		old_registry.AllowOverride(true))
}

// formatDuration formats duration for display
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return d.Round(time.Microsecond).String()
	}
	if d < time.Second {
		return d.Round(time.Millisecond).String()
	}
	return d.Round(10 * time.Millisecond).String()
}
