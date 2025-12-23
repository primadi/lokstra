package cast_test

import (
	"testing"
	"time"

	"github.com/primadi/lokstra/common/cast"
)

func TestToStruct_DurationFieldAutoConvert(t *testing.T) {
	type ServerConfig struct {
		Host    string
		Port    int
		Timeout time.Duration
	}

	tests := []struct {
		name     string
		input    map[string]any
		expected ServerConfig
		wantErr  bool
	}{
		{
			name: "duration as int64 (nanoseconds)",
			input: map[string]any{
				"Host":    "localhost",
				"Port":    8080,
				"Timeout": int64(900000000000), // 15 minutes in nanoseconds
			},
			expected: ServerConfig{
				Host:    "localhost",
				Port:    8080,
				Timeout: 15 * time.Minute,
			},
		},
		{
			name: "duration as string - 15m",
			input: map[string]any{
				"Host":    "localhost",
				"Port":    8080,
				"Timeout": "15m",
			},
			expected: ServerConfig{
				Host:    "localhost",
				Port:    8080,
				Timeout: 15 * time.Minute,
			},
		},
		{
			name: "duration as string - 2h",
			input: map[string]any{
				"Host":    "localhost",
				"Port":    8080,
				"Timeout": "2h",
			},
			expected: ServerConfig{
				Host:    "localhost",
				Port:    8080,
				Timeout: 2 * time.Hour,
			},
		},
		{
			name: "duration as string - 30s",
			input: map[string]any{
				"Host":    "localhost",
				"Port":    8080,
				"Timeout": "30s",
			},
			expected: ServerConfig{
				Host:    "localhost",
				Port:    8080,
				Timeout: 30 * time.Second,
			},
		},
		{
			name: "duration as string - complex 1h30m",
			input: map[string]any{
				"Host":    "localhost",
				"Port":    8080,
				"Timeout": "1h30m",
			},
			expected: ServerConfig{
				Host:    "localhost",
				Port:    8080,
				Timeout: 90 * time.Minute,
			},
		},
		{
			name: "duration as time.Duration directly",
			input: map[string]any{
				"Host":    "localhost",
				"Port":    8080,
				"Timeout": 15 * time.Minute,
			},
			expected: ServerConfig{
				Host:    "localhost",
				Port:    8080,
				Timeout: 15 * time.Minute,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result ServerConfig
			err := cast.ToStruct(tt.input, &result, false)

			if (err != nil) != tt.wantErr {
				t.Errorf("ToStruct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result.Host != tt.expected.Host {
					t.Errorf("Host = %v, want %v", result.Host, tt.expected.Host)
				}
				if result.Port != tt.expected.Port {
					t.Errorf("Port = %v, want %v", result.Port, tt.expected.Port)
				}
				if result.Timeout != tt.expected.Timeout {
					t.Errorf("Timeout = %v, want %v", result.Timeout, tt.expected.Timeout)
				}
			}
		})
	}
}

func TestToStruct_NestedStructWithDuration(t *testing.T) {
	type RetryConfig struct {
		MaxRetries int
		Delay      time.Duration
	}

	type ServiceConfig struct {
		Name  string
		Retry RetryConfig
	}

	input := map[string]any{
		"Name": "my-service",
		"Retry": map[string]any{
			"MaxRetries": 3,
			"Delay":      "5s", // Duration as string
		},
	}

	var result ServiceConfig
	err := cast.ToStruct(input, &result, false)
	if err != nil {
		t.Fatalf("ToStruct() error = %v", err)
	}

	if result.Name != "my-service" {
		t.Errorf("Name = %v, want my-service", result.Name)
	}
	if result.Retry.MaxRetries != 3 {
		t.Errorf("Retry.MaxRetries = %v, want 3", result.Retry.MaxRetries)
	}
	if result.Retry.Delay != 5*time.Second {
		t.Errorf("Retry.Delay = %v, want 5s", result.Retry.Delay)
	}
}
