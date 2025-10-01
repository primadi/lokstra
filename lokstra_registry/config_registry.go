package lokstra_registry

import (
	"fmt"
	"reflect"
)

// Global configuration registry
var configRegistry = make(map[string]any)

// RegisterConfig registers a configuration value with a name
func RegisterConfig(name string, value any, opts ...RegisterOption) {
	var options registerOptions
	for _, opt := range opts {
		opt.apply(&options)
	}
	if !options.allowOverride {
		if _, exists := configRegistry[name]; exists {
			panic("config " + name + " already registered")
		}
	}
	configRegistry[name] = value
}

// GetConfig retrieves a configuration value by name with type assertion and default value
func GetConfig[T any](name string, defaultValue T) T {
	if value, ok := configRegistry[name]; ok {
		if typedValue, ok := value.(T); ok {
			return typedValue
		}
		// Try to convert if types don't match exactly
		if converted, ok := convertValue[T](value); ok {
			return converted
		}
		// If conversion fails, return default and log warning
		fmt.Printf("Warning: config %s has wrong type, using default value\n", name)
	}
	return defaultValue
}

// SetConfig sets a configuration value (allows runtime changes)
func SetConfig(name string, value any) {
	configRegistry[name] = value
}

// GetConfigString is a convenience function for string configs
func GetConfigString(name, defaultValue string) string {
	return GetConfig(name, defaultValue)
}

// GetConfigInt is a convenience function for int configs
func GetConfigInt(name string, defaultValue int) int {
	return GetConfig(name, defaultValue)
}

// GetConfigBool is a convenience function for bool configs
func GetConfigBool(name string, defaultValue bool) bool {
	return GetConfig(name, defaultValue)
}

// ListConfigNames returns all registered configuration names
func ListConfigNames() []string {
	names := make([]string, 0, len(configRegistry))
	for name := range configRegistry {
		names = append(names, name)
	}
	return names
}

// convertValue attempts to convert between compatible types
func convertValue[T any](value any) (T, bool) {
	var zero T
	targetType := reflect.TypeOf(zero)
	sourceValue := reflect.ValueOf(value)

	// If types are the same, return directly
	if sourceValue.Type() == targetType {
		return value.(T), true
	}

	// Try conversion for compatible types
	if sourceValue.Type().ConvertibleTo(targetType) {
		converted := sourceValue.Convert(targetType)
		return converted.Interface().(T), true
	}

	// Special cases for string conversions
	if targetType.Kind() == reflect.String {
		return any(fmt.Sprintf("%v", value)).(T), true
	}

	return zero, false
}
