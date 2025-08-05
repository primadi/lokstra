package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
		result := GetValueFromMap(testMap, "string_value", "default")
		assert.Equal(t, "hello", result)
	})

	t.Run("Get existing int value", func(t *testing.T) {
		result := GetValueFromMap(testMap, "int_value", 0)
		assert.Equal(t, 42, result)
	})

	t.Run("Get existing float value", func(t *testing.T) {
		result := GetValueFromMap(testMap, "float_value", 0.0)
		assert.Equal(t, 3.14, result)
	})

	t.Run("Get existing bool value", func(t *testing.T) {
		result := GetValueFromMap(testMap, "bool_value", false)
		assert.Equal(t, true, result)
	})

	t.Run("Get non-existent key returns default", func(t *testing.T) {
		result := GetValueFromMap(testMap, "non_existent", "default_value")
		assert.Equal(t, "default_value", result)
	})

	t.Run("Get value with wrong type returns default", func(t *testing.T) {
		result := GetValueFromMap(testMap, "string_value", 123) // Expecting int but value is string
		assert.Equal(t, 123, result)
	})

	t.Run("Get nil value returns default", func(t *testing.T) {
		result := GetValueFromMap(testMap, "nil_value", "default")
		assert.Equal(t, "default", result)
	})

	t.Run("Get pointer string value", func(t *testing.T) {
		result := GetValueFromMap(testMap, "pointer_string", "")
		assert.Equal(t, "pointer", result)
	})

	t.Run("Get pointer int value", func(t *testing.T) {
		result := GetValueFromMap(testMap, "pointer_int", 0)
		assert.Equal(t, 99, result)
	})

	t.Run("Generic type with custom struct", func(t *testing.T) {
		type CustomStruct struct {
			Name string
			Age  int
		}

		structMap := map[string]any{
			"custom": CustomStruct{Name: "John", Age: 30},
		}

		result := GetValueFromMap(structMap, "custom", CustomStruct{})
		assert.Equal(t, CustomStruct{Name: "John", Age: 30}, result)
	})
}

func TestGetDurationFromMap(t *testing.T) {
	t.Run("Get duration from string", func(t *testing.T) {
		testMap := map[string]any{
			"duration": "30s",
		}
		result := GetDurationFromMap(testMap, "duration", 5*time.Second)
		assert.Equal(t, 30*time.Second, result)
	})

	t.Run("Get duration from invalid string returns default", func(t *testing.T) {
		testMap := map[string]any{
			"duration": "invalid_duration",
		}
		result := GetDurationFromMap(testMap, "duration", 5*time.Second)
		assert.Equal(t, 5*time.Second, result)
	})

	t.Run("Get duration from float64", func(t *testing.T) {
		testMap := map[string]any{
			"duration": 30.5,
		}
		result := GetDurationFromMap(testMap, "duration", 5*time.Second)
		// Implementation truncates float64 to int, so 30.5 becomes 30
		assert.Equal(t, 30*time.Second, result)
	})

	t.Run("Get duration from int", func(t *testing.T) {
		testMap := map[string]any{
			"duration": 45,
		}
		result := GetDurationFromMap(testMap, "duration", 5*time.Second)
		assert.Equal(t, 45*time.Second, result)
	})

	t.Run("Get duration from int64", func(t *testing.T) {
		testMap := map[string]any{
			"duration": int64(60),
		}
		result := GetDurationFromMap(testMap, "duration", 5*time.Second)
		assert.Equal(t, 60*time.Second, result)
	})

	t.Run("Get duration from time.Duration", func(t *testing.T) {
		testMap := map[string]any{
			"duration": 2 * time.Minute,
		}
		result := GetDurationFromMap(testMap, "duration", 5*time.Second)
		assert.Equal(t, 2*time.Minute, result)
	})

	t.Run("Get duration from unsupported type returns default", func(t *testing.T) {
		testMap := map[string]any{
			"duration": []string{"not", "a", "duration"},
		}
		result := GetDurationFromMap(testMap, "duration", 5*time.Second)
		assert.Equal(t, 5*time.Second, result)
	})

	t.Run("Get duration from non-existent key returns default", func(t *testing.T) {
		testMap := map[string]any{}
		result := GetDurationFromMap(testMap, "non_existent", 10*time.Second)
		assert.Equal(t, 10*time.Second, result)
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
			result := GetDurationFromMap(testMap, "duration", 0)
			assert.Equal(t, tc.expected, result, "Failed for input: %s", tc.input)
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

		cloned := CloneMap(original)

		// Check all values are copied
		assert.Equal(t, original, cloned)

		// Check they are different instances
		assert.NotSame(t, original, cloned)

		// Modify original and ensure clone is unaffected
		original["key1"] = "modified"
		assert.Equal(t, "value1", cloned["key1"])
		assert.Equal(t, "modified", original["key1"])
	})

	t.Run("Clone int map", func(t *testing.T) {
		original := map[int]int{
			1: 10,
			2: 20,
			3: 30,
		}

		cloned := CloneMap(original)

		assert.Equal(t, original, cloned)
		assert.NotSame(t, original, cloned)

		original[1] = 999
		assert.Equal(t, 10, cloned[1])
		assert.Equal(t, 999, original[1])
	})

	t.Run("Clone empty map", func(t *testing.T) {
		original := map[string]int{}
		cloned := CloneMap(original)

		assert.Equal(t, original, cloned)
		assert.NotSame(t, original, cloned)
		assert.Len(t, cloned, 0)
	})

	t.Run("Clone nil map", func(t *testing.T) {
		var original map[string]int
		cloned := CloneMap(original)

		// CloneMap creates a new empty map even for nil input
		assert.NotEqual(t, original, cloned)
		assert.NotNil(t, cloned)
		assert.Len(t, cloned, 0)
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

		cloned := CloneMap(original)

		assert.Equal(t, original, cloned)
		assert.NotSame(t, original, cloned)

		// Modify original
		original["john"] = Person{Name: "Johnny", Age: 31}
		assert.Equal(t, Person{Name: "John", Age: 30}, cloned["john"])
	})

	t.Run("Clone map performance with large map", func(t *testing.T) {
		original := make(map[int]int, 1000)
		for i := 0; i < 1000; i++ {
			original[i] = i * 2
		}

		cloned := CloneMap(original)

		assert.Equal(t, len(original), len(cloned))
		assert.Equal(t, original[500], cloned[500])
	})
}
