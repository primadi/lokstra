package utils_test

import (
	"testing"
	"time"

	"github.com/primadi/lokstra/common/utils"
)

// assertEqual is a simple helper for testing equality
func assertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestGetValueFromMap(t *testing.T) {
	testMap := map[string]any{
		"string_value":   "hello",
		"int_value":      42,
		"float_value":    3.14,
		"bool_value":     true,
		"nil_value":      nil,
		"pointer_string": func() *string { s := "pointer"; return &s }(),
		"pointer_int":    func() *int { i := 99; return &i }(),
	}

	t.Run("Get existing string value", func(t *testing.T) {
		result := utils.GetValueFromMap(testMap, "string_value", "default")
		assertEqual(t, result, "hello")
	})

	t.Run("Get existing int value", func(t *testing.T) {
		result := utils.GetValueFromMap(testMap, "int_value", 0)
		assertEqual(t, result, 42)
	})

	t.Run("Get existing float value", func(t *testing.T) {
		result := utils.GetValueFromMap(testMap, "float_value", 0.0)
		assertEqual(t, result, 3.14)
	})

	t.Run("Get existing bool value", func(t *testing.T) {
		result := utils.GetValueFromMap(testMap, "bool_value", false)
		assertEqual(t, result, true)
	})

	t.Run("Get non-existent key returns default", func(t *testing.T) {
		result := utils.GetValueFromMap(testMap, "non_existent", "default_value")
		assertEqual(t, result, "default_value")
	})

	t.Run("Get value with wrong type returns default", func(t *testing.T) {
		result := utils.GetValueFromMap(testMap, "string_value", 123) // Expecting int but value is string
		assertEqual(t, result, 123)
	})

	t.Run("Get nil value returns default", func(t *testing.T) {
		result := utils.GetValueFromMap(testMap, "nil_value", "default")
		assertEqual(t, result, "default")
	})

	t.Run("Get pointer string value", func(t *testing.T) {
		result := utils.GetValueFromMap(testMap, "pointer_string", "")
		assertEqual(t, result, "pointer")
	})

	t.Run("Get pointer int value", func(t *testing.T) {
		result := utils.GetValueFromMap(testMap, "pointer_int", 0)
		assertEqual(t, result, 99)
	})

	t.Run("Generic type with custom struct", func(t *testing.T) {
		type CustomStruct struct {
			Name string
			Age  int
		}

		structMap := map[string]any{
			"custom": CustomStruct{Name: "John", Age: 30},
		}

		result := utils.GetValueFromMap(structMap, "custom", CustomStruct{})
		assertEqual(t, CustomStruct{Name: "John", Age: 30}, result)
	})
}

func TestGetDurationFromMap(t *testing.T) {
	t.Run("Get duration from string", func(t *testing.T) {
		testMap := map[string]any{
			"duration": "30s",
		}
		result := utils.GetDurationFromMap(testMap, "duration", 5*time.Second)
		assertEqual(t, result, 30*time.Second)
	})

	t.Run("Get duration from invalid string returns default", func(t *testing.T) {
		testMap := map[string]any{
			"duration": "invalid_duration",
		}
		result := utils.GetDurationFromMap(testMap, "duration", 5*time.Second)
		assertEqual(t, result, 5*time.Second)
	})

	t.Run("Get duration from float64", func(t *testing.T) {
		testMap := map[string]any{
			"duration": 30.5,
		}
		result := utils.GetDurationFromMap(testMap, "duration", 5*time.Second)
		// Implementation truncates float64 to int, so 30.5 becomes 30
		assertEqual(t, result, 30*time.Second)
	})

	t.Run("Get duration from int", func(t *testing.T) {
		testMap := map[string]any{
			"duration": 45,
		}
		result := utils.GetDurationFromMap(testMap, "duration", 5*time.Second)
		assertEqual(t, result, 45*time.Second)
	})

	t.Run("Get duration from int64", func(t *testing.T) {
		testMap := map[string]any{
			"duration": int64(60),
		}
		result := utils.GetDurationFromMap(testMap, "duration", 5*time.Second)
		assertEqual(t, result, 60*time.Second)
	})

	t.Run("Get duration from time.Duration", func(t *testing.T) {
		testMap := map[string]any{
			"duration": 2 * time.Minute,
		}
		result := utils.GetDurationFromMap(testMap, "duration", 5*time.Second)
		assertEqual(t, result, 2*time.Minute)
	})

	t.Run("Get duration from unsupported type returns default", func(t *testing.T) {
		testMap := map[string]any{
			"duration": []string{"not", "a", "duration"},
		}
		result := utils.GetDurationFromMap(testMap, "duration", 5*time.Second)
		assertEqual(t, result, 5*time.Second)
	})

	t.Run("Get duration from non-existent key returns default", func(t *testing.T) {
		testMap := map[string]any{}
		result := utils.GetDurationFromMap(testMap, "non_existent", 10*time.Second)
		assertEqual(t, result, 10*time.Second)
	})

	t.Run("Complex duration strings", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected time.Duration
		}{
			{"1h30m", 1*time.Hour + 30*time.Minute},
			{"500ms", 500 * time.Millisecond},
			{"2h45m30s", 2*time.Hour + 45*time.Minute + 30*time.Second},
			{"1m30s", 1*time.Minute + 30*time.Second},
		}

		for _, tc := range testCases {
			testMap := map[string]any{
				"duration": tc.input,
			}
			result := utils.GetDurationFromMap(testMap, "duration", 0)
			if result != tc.expected {
				t.Errorf("Failed for input: %s - got %v, want %v", tc.input, result, tc.expected)
			}
		}
	})
}

func TestCloneMap(t *testing.T) {
	t.Run("Clone string map", func(t *testing.T) {
		original := map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		}

		cloned := utils.CloneMap(original)

		// Check all values are copied
		if len(cloned) != len(original) {
			t.Errorf("len(cloned) = %d, want %d", len(cloned), len(original))
		}
		for k, v := range original {
			if cloned[k] != v {
				t.Errorf("cloned[%q] = %v, want %v", k, cloned[k], v)
			}
		}

		// Check they are different instances (NotSame check removed - maps are reference types)

		// Modify original and ensure clone is unaffected
		original["key1"] = "modified"
		assertEqual(t, cloned["key1"], "value1")
		assertEqual(t, original["key1"], "modified")
	})

	t.Run("Clone int map", func(t *testing.T) {
		original := map[int]int{
			1: 10,
			2: 20,
			3: 30,
		}

		cloned := utils.CloneMap(original)

		if len(cloned) != len(original) {
			t.Errorf("len(cloned) = %d, want %d", len(cloned), len(original))
		}
		for k, v := range original {
			if cloned[k] != v {
				t.Errorf("cloned[%d] = %v, want %v", k, cloned[k], v)
			}
		}
		// NotSame check removed - maps are reference types

		original[1] = 999
		assertEqual(t, cloned[1], 10)
		assertEqual(t, original[1], 999)
	})

	t.Run("Clone empty map", func(t *testing.T) {
		original := map[string]int{}
		cloned := utils.CloneMap(original)

		if len(cloned) != len(original) {
			t.Errorf("len(cloned) = %d, want %d", len(cloned), len(original))
		}
		// NotSame check removed - maps are reference types
		if len(cloned) != 0 {
			t.Errorf("len(cloned) = %d, want 0", len(cloned))
		}
	})

	t.Run("Clone nil map", func(t *testing.T) {
		var original map[string]int
		cloned := utils.CloneMap(original)

		// utils.CloneMap creates a new empty map even for nil input
		// if original == nil && cloned == nil {
		// 	t.Error("cloned should not be nil when original is nil")
		// }
		if cloned == nil {
			t.Error("cloned should not be nil")
		}
		if len(cloned) != 0 {
			t.Errorf("len(cloned) = %d, want 0", len(cloned))
		}
	})

	t.Run("Clone complex value map", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}

		original := map[string]Person{
			"john": {Name: "John", Age: 30},
			"jane": {Name: "Jane", Age: 25},
		}

		cloned := utils.CloneMap(original)

		if len(cloned) != len(original) {
			t.Errorf("len(cloned) = %d, want %d", len(cloned), len(original))
		}
		for k, v := range original {
			if cloned[k] != v {
				t.Errorf("cloned[%q] = %v, want %v", k, cloned[k], v)
			}
		}
		// NotSame check removed - maps are reference types

		// Modify original
		original["john"] = Person{Name: "Johnny", Age: 31}
		assertEqual(t, Person{Name: "John", Age: 30}, cloned["john"])
	})

	t.Run("Clone map performance with large map", func(t *testing.T) {
		original := make(map[int]int, 1000)
		for i := 0; i < 1000; i++ {
			original[i] = i * 2
		}

		cloned := utils.CloneMap(original)

		if len(cloned) != len(original) {
			t.Errorf("len(cloned) = %d, want %d", len(cloned), len(original))
		}
		assertEqual(t, cloned[500], original[500])
	})
}
