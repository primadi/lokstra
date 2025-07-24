package listener

import (
	"testing"
	"time"
)

func TestConstants(t *testing.T) {
	// Test timeout key constants
	expectedKeys := map[string]string{
		"read_timeout":  READ_TIMEOUT_KEY,
		"write_timeout": WRITE_TIMEOUT_KEY,
		"idle_timeout":  IDLE_TIMEOUT_LEY,
	}

	for expected, actual := range expectedKeys {
		if actual != expected {
			t.Errorf("Expected constant %s, got %s", expected, actual)
		}
	}

	// Test default timeout values
	if DEFAULT_READ_TIMEOUT != 5*time.Minute {
		t.Errorf("DEFAULT_READ_TIMEOUT = %v, want %v", DEFAULT_READ_TIMEOUT, 5*time.Minute)
	}

	if DEFAULT_WRITE_TIMEOUT != 5*time.Minute {
		t.Errorf("DEFAULT_WRITE_TIMEOUT = %v, want %v", DEFAULT_WRITE_TIMEOUT, 5*time.Minute)
	}

	if DEFAULT_IDLE_TIMEOUT != 10*time.Minute {
		t.Errorf("DEFAULT_IDLE_TIMEOUT = %v, want %v", DEFAULT_IDLE_TIMEOUT, 10*time.Minute)
	}
}

func TestTimeoutConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{
			name:     "read_timeout_key",
			constant: READ_TIMEOUT_KEY,
			expected: "read_timeout",
		},
		{
			name:     "write_timeout_key",
			constant: WRITE_TIMEOUT_KEY,
			expected: "write_timeout",
		},
		{
			name:     "idle_timeout_key",
			constant: IDLE_TIMEOUT_LEY,
			expected: "idle_timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("Constant %s = %v, want %v", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestDefaultTimeouts(t *testing.T) {
	tests := []struct {
		name     string
		timeout  time.Duration
		expected time.Duration
	}{
		{
			name:     "default_read_timeout",
			timeout:  DEFAULT_READ_TIMEOUT,
			expected: 5 * time.Minute,
		},
		{
			name:     "default_write_timeout",
			timeout:  DEFAULT_WRITE_TIMEOUT,
			expected: 5 * time.Minute,
		},
		{
			name:     "default_idle_timeout",
			timeout:  DEFAULT_IDLE_TIMEOUT,
			expected: 10 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.timeout != tt.expected {
				t.Errorf("Timeout %s = %v, want %v", tt.name, tt.timeout, tt.expected)
			}
		})
	}
}
