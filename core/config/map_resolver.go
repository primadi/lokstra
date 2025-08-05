package config

type MapResolver struct {
	inputMap map[string]string
}

// Resolve implements VariableResolver.
func (m *MapResolver) Resolve(source string, key string, defaultValue string) (string, bool) {
	if source != "MAP" {
		return "", false
	}
	value, exists := m.inputMap[key]
	if !exists || value == "" {
		return defaultValue, false
	}
	return value, true
}

var _ VariableResolver = (*MapResolver)(nil)

func NewMapResolver(inputMap map[string]string) *MapResolver {
	if inputMap == nil {
		inputMap = make(map[string]string)
	}
	return &MapResolver{inputMap: inputMap}
}
