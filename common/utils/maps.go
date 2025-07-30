package utils

import (
	"fmt"
	"maps"
	"time"
)

func GetValueFromMap[T any](settings map[string]any, key string, defaultValue T) T {
	if value, exists := settings[key]; exists {
		if typedValue, ok := value.(T); ok {
			return typedValue
		}
		if typedValue, ok := value.(*T); ok {
			return *typedValue
		}
	}
	return defaultValue
}

func GetDurationFromMap(settings map[string]any, key string, defaultValue time.Duration) time.Duration {
	if val, ok := settings[key]; ok {
		switch v := val.(type) {
		case string:
			if d, err := time.ParseDuration(v); err == nil {
				return d
			} else {
				fmt.Printf("Invalid duration string for key %q: %v\n", key, err)
			}
		case float64: // jika YAML sudah diparse jadi angka (ms, s, dst)
			return time.Duration(v) * time.Second
		case int:
			return time.Duration(v) * time.Second
		case int64:
			return time.Duration(v) * time.Second
		case time.Duration:
			return v
		default:
			fmt.Printf("Unsupported duration type for key %q: %T\n", key, val)
		}
	}
	return defaultValue
}

func CloneMap[K comparable, V any](original map[K]V) map[K]V {
	clone := make(map[K]V, len(original))
	maps.Copy(clone, original)
	return clone
}
