package slow_request_logger

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/middleware/slow_request_logger/internal"
)

const SLOW_REQUEST_LOGGER_TYPE = "slow_request_logger"
const PARAMS_THRESHOLD = "threshold"
const PARAMS_ENABLE_COLORS = "enable_colors"
const PARAMS_SKIP_PATHS = "skip_paths"

type Config struct {
	// Threshold is the minimum duration to consider a request as slow
	// Requests faster than this will not be logged
	Threshold time.Duration

	// EnableColors enables colored output for terminal
	EnableColors bool

	// SkipPaths is a list of paths to skip logging
	// Example: ["/health", "/metrics"]
	SkipPaths []string

	// CustomLogger is a custom logging function
	// If nil, uses default logger.LogInfo
	CustomLogger func(format string, args ...any)
}

func DefaultConfig() *Config {
	return &Config{
		Threshold:    500 * time.Millisecond, // 500ms default
		EnableColors: true,
		SkipPaths:    []string{},
		CustomLogger: nil,
	}
}

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

// logs only slow requests that exceed the threshold
func Middleware(cfg *Config) request.HandlerFunc {
	defConfig := DefaultConfig()
	if cfg.Threshold == 0 {
		cfg.Threshold = defConfig.Threshold
	}
	if cfg.SkipPaths == nil {
		cfg.SkipPaths = defConfig.SkipPaths
	}
	if cfg.CustomLogger == nil {
		cfg.CustomLogger = logger.LogInfo
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

		// Only log if request is slow
		if duration >= cfg.Threshold {
			// Get status code from writer wrapper
			statusCode := c.W.StatusCode()
			if statusCode == 0 {
				statusCode = 200 // Default if not set
			}

			// Determine color based on duration severity
			durationColor := colorYellow
			if duration >= cfg.Threshold*2 {
				durationColor = colorRed // Extra slow
			}

			// Format and log slow request
			if cfg.EnableColors {
				msg := fmt.Sprintf("%s[SLOW REQUEST]%s %s%s%s %s - Status: %d - Duration: %s%s%s (threshold: %s)",
					colorRed,
					colorReset,
					colorCyan,
					c.R.Method,
					colorReset,
					c.R.URL.Path,
					statusCode,
					durationColor,
					internal.FormatDuration(duration),
					colorReset,
					internal.FormatDuration(cfg.Threshold),
				)
				cfg.CustomLogger("%s", msg)
			} else {
				msg := fmt.Sprintf("[SLOW REQUEST] [%s] %s - Status: %d - Duration: %s (threshold: %s)",
					c.R.Method,
					c.R.URL.Path,
					statusCode,
					internal.FormatDuration(duration),
					internal.FormatDuration(cfg.Threshold),
				)
				cfg.CustomLogger("%s", msg)
			}
		}

		return err
	})
}

func MiddlewareFactory(params map[string]any) request.HandlerFunc {
	defConfig := DefaultConfig()
	if params == nil {
		return Middleware(defConfig)
	}

	// Handle threshold - could be int (milliseconds) or duration string
	threshold := defConfig.Threshold
	if thresholdVal, ok := params[PARAMS_THRESHOLD]; ok {
		switch v := thresholdVal.(type) {
		case int:
			threshold = time.Duration(v) * time.Millisecond
		case int64:
			threshold = time.Duration(v) * time.Millisecond
		case time.Duration:
			threshold = v
		case string:
			if d, err := time.ParseDuration(v); err == nil {
				threshold = d
			}
		}
	}

	cfg := &Config{
		Threshold:    threshold,
		EnableColors: utils.GetValueFromMap(params, PARAMS_ENABLE_COLORS, defConfig.EnableColors),
		SkipPaths:    utils.GetValueFromMap(params, PARAMS_SKIP_PATHS, defConfig.SkipPaths),
		CustomLogger: nil, // Cannot be set via params
	}
	return Middleware(cfg)
}

func Register() {
	lokstra_registry.RegisterMiddlewareFactory(SLOW_REQUEST_LOGGER_TYPE, MiddlewareFactory,
		lokstra_registry.AllowOverride(true))
}
