package cast_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/primadi/lokstra/common/cast"
)

// CustomTime is a struct that implements json.Unmarshaler
type CustomTime struct {
	time.Time
}

func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	// Custom parsing: support both RFC3339 and Unix timestamp
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		// Try parsing as RFC3339
		if t, err := time.Parse(time.RFC3339, str); err == nil {
			ct.Time = t
			return nil
		}
		// Try parsing as date only
		if t, err := time.Parse("2006-01-02", str); err == nil {
			ct.Time = t
			return nil
		}
	}

	// Try parsing as Unix timestamp
	var timestamp int64
	if err := json.Unmarshal(data, &timestamp); err == nil {
		ct.Time = time.Unix(timestamp, 0)
		return nil
	}

	return json.Unmarshal(data, &ct.Time)
}

type EventConfig struct {
	Name      string
	StartTime CustomTime
	Duration  time.Duration
}

func TestToStruct_WithUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected EventConfig
		wantErr  bool
	}{
		{
			name: "RFC3339 format",
			input: map[string]any{
				"Name":      "Meeting",
				"StartTime": "2024-12-25T10:00:00Z",
				"Duration":  3600000000000, // 1 hour in nanoseconds
			},
			expected: EventConfig{
				Name:      "Meeting",
				StartTime: CustomTime{time.Date(2024, 12, 25, 10, 0, 0, 0, time.UTC)},
				Duration:  time.Hour,
			},
		},
		{
			name: "Date only format",
			input: map[string]any{
				"Name":      "Birthday",
				"StartTime": "2024-12-25",
				"Duration":  86400000000000, // 24 hours in nanoseconds
			},
			expected: EventConfig{
				Name:      "Birthday",
				StartTime: CustomTime{time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC)},
				Duration:  24 * time.Hour,
			},
		},
		{
			name: "Unix timestamp",
			input: map[string]any{
				"Name":      "Launch",
				"StartTime": int64(1735128000), // 2024-12-25 10:00:00 UTC
				"Duration":  7200000000000,     // 2 hours in nanoseconds
			},
			expected: EventConfig{
				Name:      "Launch",
				StartTime: CustomTime{time.Unix(1735128000, 0)},
				Duration:  2 * time.Hour,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result EventConfig
			err := cast.ToStruct(tt.input, &result, false)

			if (err != nil) != tt.wantErr {
				t.Errorf("ToStruct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result.Name != tt.expected.Name {
					t.Errorf("Name = %v, want %v", result.Name, tt.expected.Name)
				}
				if !result.StartTime.Equal(tt.expected.StartTime.Time) {
					t.Errorf("StartTime = %v, want %v", result.StartTime.Time, tt.expected.StartTime.Time)
				}
				if result.Duration != tt.expected.Duration {
					t.Errorf("Duration = %v, want %v", result.Duration, tt.expected.Duration)
				}
			}
		})
	}
}

// Custom struct with validation
type Email struct {
	Address string
}

func (e *Email) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	// Simple validation
	if len(str) == 0 || !containsAt(str) {
		return json.Unmarshal([]byte(`"invalid@example.com"`), &e.Address)
	}

	e.Address = str
	return nil
}

func containsAt(s string) bool {
	for _, ch := range s {
		if ch == '@' {
			return true
		}
	}
	return false
}

type UserConfig struct {
	Name  string
	Email Email
}

func TestToStruct_WithValidation(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected UserConfig
	}{
		{
			name: "valid email",
			input: map[string]any{
				"Name":  "John",
				"Email": "john@example.com",
			},
			expected: UserConfig{
				Name:  "John",
				Email: Email{Address: "john@example.com"},
			},
		},
		{
			name: "invalid email - use default",
			input: map[string]any{
				"Name":  "Jane",
				"Email": "not-an-email",
			},
			expected: UserConfig{
				Name:  "Jane",
				Email: Email{Address: "invalid@example.com"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result UserConfig
			err := cast.ToStruct(tt.input, &result, false)

			if err != nil {
				t.Errorf("ToStruct() error = %v", err)
				return
			}

			if result.Name != tt.expected.Name {
				t.Errorf("Name = %v, want %v", result.Name, tt.expected.Name)
			}
			if result.Email.Address != tt.expected.Email.Address {
				t.Errorf("Email = %v, want %v", result.Email.Address, tt.expected.Email.Address)
			}
		})
	}
}
