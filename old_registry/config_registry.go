package old_registry

import (
	"fmt"
	"reflect"
	"sync"
)

// Global configuration registry
var configRegistry sync.Map

// GetConfig retrieves a configuration value by name with type assertion and default value
func GetConfig[T any](name string, defaultValue T) T {
	value, ok := configRegistry.Load(name)

	if ok {
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

// GetConfigValue retrieves a configuration value by name (non-generic version)
// Returns (value, found) for use with ConfigGetter interface
func GetConfigValue(name string) (any, bool) {
	value, ok := configRegistry.Load(name)
	return value, ok
}

// SetConfig sets a configuration value (allows runtime changes)
func SetConfig(name string, value any) {
	configRegistry.Store(name, value)
}

// ListConfigNames returns all registered configuration names
func ListConfigNames() []string {
	names := make([]string, 0)
	configRegistry.Range(func(key, value any) bool {
		names = append(names, key.(string))
		return true
	})
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

// ConfigRegistryGetter is an adapter to make lokstra_registry compatible with ConfigGetter interface
type ConfigRegistryGetter struct{}

// GetConfig implements ConfigGetter interface
func (g *ConfigRegistryGetter) GetConfig(key string) (any, bool) {
	return GetConfigValue(key)
}
