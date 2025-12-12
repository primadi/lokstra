package utils

import (
	"maps"
	"time"

	"github.com/primadi/lokstra/common/logger"
)

func GetValueFromMap[T any](settings map[string]any, key string, defaultValue T) T {
	// Special case: if T is time.Duration, delegate to GetDurationFromMap
	if _, ok := any(defaultValue).(time.Duration); ok {
		duration := GetDurationFromMap(settings, key, defaultValue)
		return any(duration).(T)
	}

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

func GetDurationFromMap(settings map[string]any, key string, defaultValue any) time.Duration {
	if val, ok := settings[key]; ok {
		switch v := val.(type) {
		case string:
			if d, err := time.ParseDuration(v); err == nil {
				return d
			} else {
				logger.LogInfo("Invalid duration string for key %q: %v\n", key, err)
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
			logger.LogInfo("Unsupported duration type for key %q: %T\n", key, val)
		}
	}

	switch v := defaultValue.(type) {
	case string:
		if d, err := time.ParseDuration(v); err == nil {
			return d
		} else {
			logger.LogInfo("Invalid duration string for key %q: %v\n", key, err)
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
		logger.LogInfo("Unsupported default duration type for key %q: %T\n",
			key, defaultValue)
	}

	return 0
}

func CloneMap[K comparable, V any](original map[K]V) map[K]V {
	clone := make(map[K]V, len(original))
	maps.Copy(clone, original)
	return clone
}
