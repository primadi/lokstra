package logger

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/primadi/lokstra/serviceapi"
	"github.com/rs/zerolog"
)

func TestLoggerService_NewService(t *testing.T) {
	service, err := NewService(serviceapi.LogLevelInfo)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if service == nil {
		t.Fatal("Expected service to be created")
	}
	if service.GetLogLevel() != serviceapi.LogLevelInfo {
		t.Errorf("Expected log level Info, got %v", service.GetLogLevel())
	}
}

func TestLoggerService_NewServiceWithConfig(t *testing.T) {
	config := &LoggerConfig{
		Level:  serviceapi.LogLevelDebug,
		Format: "json",
		Output: "stdout",
	}

	service, err := NewServiceWithConfig(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if service == nil {
		t.Fatal("Expected service to be created")
	}
	if service.GetLogLevel() != serviceapi.LogLevelDebug {
		t.Errorf("Expected log level Debug, got %v", service.GetLogLevel())
	}
}

func TestLoggerService_SetLogLevel(t *testing.T) {
	service, _ := NewService(serviceapi.LogLevelInfo)

	service.SetLogLevel(serviceapi.LogLevelError)
	if service.GetLogLevel() != serviceapi.LogLevelError {
		t.Errorf("Expected log level Error, got %v", service.GetLogLevel())
	}
}

func TestLoggerService_SetFormat(t *testing.T) {
	// Capture original stdout
	originalStdout := os.Stdout
	defer func() { os.Stdout = originalStdout }()

	// Test JSON format
	t.Run("JSON format", func(t *testing.T) {
		// Create pipe to capture output
		r, w, _ := os.Pipe()
		os.Stdout = w

		config := &LoggerConfig{
			Level:  serviceapi.LogLevelInfo,
			Format: "console", // Start with console
			Output: "stdout",
		}
		service, _ := NewServiceWithConfig(config)

		// Change to JSON format
		service.SetFormat("json")

		// Log a message
		service.Infof("test message")

		// Close writer and read output
		w.Close()
		os.Stdout = originalStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// Verify JSON format (should contain JSON fields)
		var jsonLog map[string]any
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) > 0 {
			err := json.Unmarshal([]byte(lines[0]), &jsonLog)
			if err != nil {
				t.Errorf("Expected JSON format, but got error parsing: %v\nOutput: %s", err, output)
			} else {
				// Verify it has expected JSON fields
				if _, ok := jsonLog["level"]; !ok {
					t.Error("Expected 'level' field in JSON output")
				}
				if _, ok := jsonLog["time"]; !ok {
					t.Error("Expected 'time' field in JSON output")
				}
				if _, ok := jsonLog["message"]; !ok {
					t.Error("Expected 'message' field in JSON output")
				}
			}
		}
	})

	// Test Console format
	t.Run("Console format", func(t *testing.T) {
		// Create pipe to capture output
		r, w, _ := os.Pipe()
		os.Stdout = w

		config := &LoggerConfig{
			Level:  serviceapi.LogLevelInfo,
			Format: "json", // Start with JSON
			Output: "stdout",
		}
		service, _ := NewServiceWithConfig(config)

		// Change to console format
		service.SetFormat("console")

		// Log a message
		service.Infof("test console message")

		// Close writer and read output
		w.Close()
		os.Stdout = originalStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// Verify console format (should NOT be valid JSON)
		var jsonLog map[string]any
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) > 0 {
			err := json.Unmarshal([]byte(lines[0]), &jsonLog)
			if err == nil {
				t.Error("Expected console format (not JSON), but output is valid JSON")
			}
			// Console format should contain the message in a human-readable way
			if !strings.Contains(output, "test console message") {
				t.Errorf("Expected message in console output, got: %s", output)
			}
		}
	})
}

func TestLoggerService_WithField(t *testing.T) {
	service, _ := NewService(serviceapi.LogLevelInfo)

	newLogger := service.WithField("key", "value")
	if newLogger == nil {
		t.Fatal("Expected new logger to be created")
	}

	// Verify it's a different instance
	if newLogger == service {
		t.Error("Expected new logger instance, got same instance")
	}
}

func TestLoggerService_WithFields(t *testing.T) {
	service, _ := NewService(serviceapi.LogLevelInfo)

	fields := serviceapi.LogFields{
		"field1": "value1",
		"field2": 42,
		"field3": true,
	}

	newLogger := service.WithFields(fields)
	if newLogger == nil {
		t.Fatal("Expected new logger to be created")
	}

	// Verify it's a different instance
	if newLogger == service {
		t.Error("Expected new logger instance, got same instance")
	}
}

func TestLoggerService_LoggingMethods(t *testing.T) {
	// Test that all logging methods work without panic
	service, _ := NewService(serviceapi.LogLevelDebug)

	// These should not panic
	service.Debugf("debug message: %s", "test")
	service.Infof("info message: %s", "test")
	service.Warnf("warn message: %s", "test")
	service.Errorf("error message: %s", "test")
	// Note: Not testing Fatalf as it would exit the program
}

func TestLoggerService_LevelConversion(t *testing.T) {
	tests := []struct {
		apiLevel     serviceapi.LogLevel
		zerologLevel zerolog.Level
	}{
		{serviceapi.LogLevelDebug, zerolog.DebugLevel},
		{serviceapi.LogLevelInfo, zerolog.InfoLevel},
		{serviceapi.LogLevelWarn, zerolog.WarnLevel},
		{serviceapi.LogLevelError, zerolog.ErrorLevel},
		{serviceapi.LogLevelFatal, zerolog.FatalLevel},
		{serviceapi.LogLevelDisabled, zerolog.Disabled},
	}

	for _, test := range tests {
		t.Run(string(rune(test.apiLevel)), func(t *testing.T) {
			// Test toZerologLevel
			converted := toZerologLevel(test.apiLevel)
			if converted != test.zerologLevel {
				t.Errorf("toZerologLevel(%v) = %v, want %v", test.apiLevel, converted, test.zerologLevel)
			}

			// Test fromZerologLevel
			convertedBack := fromZerologLevel(test.zerologLevel)
			if convertedBack != test.apiLevel {
				t.Errorf("fromZerologLevel(%v) = %v, want %v", test.zerologLevel, convertedBack, test.apiLevel)
			}
		})
	}
}

func TestLoggerService_ConfigFormats(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{"JSON format", "json"},
		{"Text format", "text"},
		{"Console format", "console"},
		{"Default format", "unknown"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := &LoggerConfig{
				Level:  serviceapi.LogLevelInfo,
				Format: test.format,
				Output: "stdout",
			}

			service, err := NewServiceWithConfig(config)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
			if service == nil {
				t.Fatal("Expected service to be created")
			}
		})
	}
}

func TestLoggerService_ConfigOutputs(t *testing.T) {
	tests := []struct {
		name   string
		output string
	}{
		{"Stdout output", "stdout"},
		{"Stderr output", "stderr"},
		{"Default output", "unknown"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := &LoggerConfig{
				Level:  serviceapi.LogLevelInfo,
				Format: "json",
				Output: test.output,
			}

			service, err := NewServiceWithConfig(config)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
			if service == nil {
				t.Fatal("Expected service to be created")
			}
		})
	}
}
