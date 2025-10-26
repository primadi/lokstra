package slow_request_logger

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response/api_formatter"
	"github.com/primadi/lokstra/core/router"
)

func TestSlowRequestLogger(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		path      string
		delay     time.Duration
		shouldLog bool
	}{
		{
			name: "log slow request",
			config: &Config{
				Threshold:    100 * time.Millisecond,
				EnableColors: false,
			},
			path:      "/api/slow",
			delay:     150 * time.Millisecond,
			shouldLog: true,
		},
		{
			name: "skip fast request",
			config: &Config{
				Threshold:    100 * time.Millisecond,
				EnableColors: false,
			},
			path:      "/api/fast",
			delay:     10 * time.Millisecond,
			shouldLog: false,
		},
		{
			name: "skip path even if slow",
			config: &Config{
				Threshold:    50 * time.Millisecond,
				EnableColors: false,
				SkipPaths:    []string{"/health"},
			},
			path:      "/health",
			delay:     100 * time.Millisecond,
			shouldLog: false,
		},
		{
			name: "log extra slow request",
			config: &Config{
				Threshold:    100 * time.Millisecond,
				EnableColors: false,
			},
			path:      "/api/extra-slow",
			delay:     250 * time.Millisecond, // 2.5x threshold
			shouldLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup formatter
			api_formatter.SetGlobalFormatter(api_formatter.NewApiResponseFormatter())

			// Capture logs
			var logOutput []string
			tt.config.CustomLogger = func(format string, args ...any) {
				msg := fmt.Sprintf(format, args...)
				logOutput = append(logOutput, msg)
			}

			// Create router
			r := router.New("test-router")

			// Add slow logger middleware
			r.Use(Middleware(tt.config))

			// Add test handler with delay
			r.GET(tt.path, func(c *request.Context) error {
				time.Sleep(tt.delay)
				return c.Api.Ok("success")
			})

			// Create request
			req := httptest.NewRequest("GET", tt.path, nil)

			// Record response
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// Check if logged
			logged := len(logOutput) > 0
			if logged != tt.shouldLog {
				t.Errorf("Expected logged=%v, got %v", tt.shouldLog, logged)
			}

			if tt.shouldLog && logged {
				logLine := logOutput[0]
				t.Logf("Slow request log: %s", logLine)

				// Verify log contains "SLOW REQUEST"
				if !strings.Contains(logLine, "SLOW REQUEST") {
					t.Error("Log should contain 'SLOW REQUEST'")
				}

				// Verify log contains path
				if !strings.Contains(logLine, tt.path) {
					t.Errorf("Log should contain path %s", tt.path)
				}

				// Verify log contains threshold
				if !strings.Contains(logLine, "threshold") {
					t.Error("Log should contain threshold information")
				}
			}
		})
	}
}

func TestSlowRequestLoggerWithColors(t *testing.T) {
	// Setup formatter
	api_formatter.SetGlobalFormatter(api_formatter.NewApiResponseFormatter())

	var logOutput []string
	cfg := &Config{
		Threshold:    50 * time.Millisecond,
		EnableColors: true,
		CustomLogger: func(format string, args ...any) {
			msg := fmt.Sprintf(format, args...)
			logOutput = append(logOutput, msg)
		},
	}

	r := router.New("test-router")
	r.Use(Middleware(cfg))

	r.GET("/test", func(c *request.Context) error {
		time.Sleep(60 * time.Millisecond) // Trigger slow log
		return c.Api.Ok("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if len(logOutput) == 0 {
		t.Fatal("Expected log output for slow request")
	}

	// Verify ANSI color codes are present
	logLine := logOutput[0]
	if !strings.Contains(logLine, "\033[") {
		t.Error("Expected ANSI color codes in output")
	}

	t.Logf("Colored slow request output: %s", logLine)
}

func TestSlowRequestLoggerThresholdBoundary(t *testing.T) {
	// Setup formatter
	api_formatter.SetGlobalFormatter(api_formatter.NewApiResponseFormatter())

	var logOutput []string
	cfg := &Config{
		Threshold:    100 * time.Millisecond,
		EnableColors: false,
		CustomLogger: func(format string, args ...any) {
			msg := fmt.Sprintf(format, args...)
			logOutput = append(logOutput, msg)
		},
	}

	r := router.New("test-router")
	r.Use(Middleware(cfg))

	// Test exactly at threshold
	r.GET("/boundary", func(c *request.Context) error {
		time.Sleep(100 * time.Millisecond) // Exactly at threshold
		return c.Api.Ok("success")
	})

	req := httptest.NewRequest("GET", "/boundary", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should log (>= threshold)
	if len(logOutput) == 0 {
		t.Error("Expected log at threshold boundary")
	}

	t.Logf("Boundary test log: %s", logOutput[0])
}

func TestSlowRequestLoggerFactory(t *testing.T) {
	// Test with nil params (uses defaults)
	middleware1 := MiddlewareFactory(nil)
	if middleware1 == nil {
		t.Error("Expected middleware with nil params")
	}

	// Test with int threshold (milliseconds)
	params1 := map[string]any{
		PARAMS_THRESHOLD:     1000, // 1000ms = 1s
		PARAMS_ENABLE_COLORS: false,
	}
	middleware2 := MiddlewareFactory(params1)
	if middleware2 == nil {
		t.Error("Expected middleware with int threshold")
	}

	// Test with string threshold
	params2 := map[string]any{
		PARAMS_THRESHOLD:  "2s",
		PARAMS_SKIP_PATHS: []string{"/health"},
	}
	middleware3 := MiddlewareFactory(params2)
	if middleware3 == nil {
		t.Error("Expected middleware with string threshold")
	}
}

func TestSlowRequestLoggerFastRequestsNotLogged(t *testing.T) {
	// Setup formatter
	api_formatter.SetGlobalFormatter(api_formatter.NewApiResponseFormatter())

	var logOutput []string
	cfg := &Config{
		Threshold:    500 * time.Millisecond,
		EnableColors: false,
		CustomLogger: func(format string, args ...any) {
			msg := fmt.Sprintf(format, args...)
			logOutput = append(logOutput, msg)
		},
	}

	r := router.New("test-router")
	r.Use(Middleware(cfg))

	// Fast handler
	r.GET("/fast", func(c *request.Context) error {
		return c.Api.Ok("success")
	})

	req := httptest.NewRequest("GET", "/fast", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should not log (fast request)
	if len(logOutput) > 0 {
		t.Errorf("Fast request should not be logged, but got: %s", logOutput[0])
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		contains string
	}{
		{100 * time.Microsecond, "µs"},
		{5 * time.Millisecond, "ms"},
		{2 * time.Second, "s"},
	}

	for _, tt := range tests {
		result := formatDuration(tt.duration)
		if !strings.Contains(result, tt.contains) {
			t.Errorf("formatDuration(%v) = %s, expected to contain %s", tt.duration, result, tt.contains)
		}
	}
}
