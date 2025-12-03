package request_logger_test

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response/api_formatter"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/middleware/request_logger"
	"github.com/primadi/lokstra/middleware/request_logger/internal"
)

func TestRequestLogger(t *testing.T) {
	tests := []struct {
		name      string
		config    *request_logger.Config
		path      string
		method    string
		shouldLog bool
	}{
		{
			name: "log GET request",
			config: &request_logger.Config{
				EnableColors: false,
				SkipPaths:    []string{},
			},
			path:      "/api/test",
			method:    "GET",
			shouldLog: true,
		},
		{
			name: "skip health check path",
			config: &request_logger.Config{
				EnableColors: false,
				SkipPaths:    []string{"/health", "/metrics"},
			},
			path:      "/health",
			method:    "GET",
			shouldLog: false,
		},
		{
			name: "log POST request",
			config: &request_logger.Config{
				EnableColors: false,
				SkipPaths:    []string{},
			},
			path:      "/api/create",
			method:    "POST",
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

			// Add logger middleware
			r.Use(request_logger.Middleware(tt.config))

			// Add test handler
			switch tt.method {
			case "GET":
				r.GET(tt.path, func(c *request.Context) error {
					return c.Api.Ok("success")
				})
			case "POST":
				r.POST(tt.path, func(c *request.Context) error {
					return c.Api.Ok("success")
				})
			default:
				r.GET(tt.path, func(c *request.Context) error {
					return c.Api.Ok("success")
				})
			}

			// Create request
			req := httptest.NewRequest(tt.method, tt.path, nil)

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
				t.Logf("Log output: %s", logLine)

				// Verify log contains method and path
				if !strings.Contains(logLine, tt.method) {
					t.Errorf("Log should contain method %s", tt.method)
				}
				if !strings.Contains(logLine, tt.path) {
					t.Errorf("Log should contain path %s", tt.path)
				}
			}
		})
	}
}

func TestRequestLoggerWithColors(t *testing.T) {
	// Setup formatter
	api_formatter.SetGlobalFormatter(api_formatter.NewApiResponseFormatter())

	var logOutput []string
	cfg := &request_logger.Config{
		EnableColors: true,
		CustomLogger: func(format string, args ...any) {
			msg := fmt.Sprintf(format, args...)
			logOutput = append(logOutput, msg)
		},
	}

	r := router.New("test-router")
	r.Use(request_logger.Middleware(cfg))

	r.GET("/test", func(c *request.Context) error {
		return c.Api.Ok("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if len(logOutput) == 0 {
		t.Fatal("Expected log output")
	}

	// Verify ANSI color codes are present
	logLine := logOutput[0]
	if !strings.Contains(logLine, "\033[") {
		t.Error("Expected ANSI color codes in output")
	}

	t.Logf("Colored output: %s", logLine)
}

func TestRequestLoggerDuration(t *testing.T) {
	// Setup formatter
	api_formatter.SetGlobalFormatter(api_formatter.NewApiResponseFormatter())

	var logOutput []string
	cfg := &request_logger.Config{
		EnableColors: false,
		CustomLogger: func(format string, args ...any) {
			msg := fmt.Sprintf(format, args...)
			logOutput = append(logOutput, msg)
		},
	}

	r := router.New("test-router")
	r.Use(request_logger.Middleware(cfg))

	r.GET("/slow", func(c *request.Context) error {
		time.Sleep(10 * time.Millisecond) // Simulate slow request
		return c.Api.Ok("success")
	})

	req := httptest.NewRequest("GET", "/slow", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if len(logOutput) == 0 {
		t.Fatal("Expected log output")
	}

	logLine := logOutput[0]
	// Log should contain duration
	if !strings.Contains(logLine, "ms") && !strings.Contains(logLine, "µs") && !strings.Contains(logLine, "s") {
		t.Error("Log should contain duration")
	}

	t.Logf("Log with duration: %s", logLine)
}

func TestRequestLoggerFactory(t *testing.T) {
	// Test with nil params
	middleware1 := request_logger.MiddlewareFactory(nil)
	if middleware1 == nil {
		t.Error("Expected middleware with nil params")
	}

	// Test with custom params
	params := map[string]any{
		request_logger.PARAMS_ENABLE_COLORS: false,
		request_logger.PARAMS_SKIP_PATHS:    []string{"/health"},
	}
	middleware2 := request_logger.MiddlewareFactory(params)
	if middleware2 == nil {
		t.Error("Expected middleware with custom params")
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
		result := internal.FormatDuration(tt.duration)
		if !strings.Contains(result, tt.contains) {
			t.Errorf("formatDuration(%v) = %s, expected to contain %s", tt.duration, result, tt.contains)
		}
	}
}
