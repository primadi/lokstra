package engine

var engineRegistry = make(map[string]func() RouterEngine)

// RegisterEngine registers a new RouterEngine constructor with a name.
func RegisterEngine(name string, constructor func() RouterEngine) {
	engineRegistry[name] = constructor
}

// CreateEngine creates a new RouterEngine based on the engine type.
func CreateEngine(engineType string) RouterEngine {
	if constructor, exists := engineRegistry[engineType]; exists {
		return constructor()
	}
	panic("Unsupported engine type: " + engineType)
}

func init() {
	RegisterEngine("default", NewModernServeMux)  // Use modern Go 1.22+ version as default
	RegisterEngine("servemux", NewModernServeMux) // New modern version
}
