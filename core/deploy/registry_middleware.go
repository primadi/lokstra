package deploy

import (
	"fmt"
	"strings"

	"github.com/primadi/lokstra/core/request"
)

// RegisterMiddlewareType registers a middleware factory
// Supports optional AllowOverride option
func (g *GlobalRegistry) RegisterMiddlewareType(middlewareType string, factory MiddlewareFactory, opts ...MiddlewareTypeOption) {
	g.mu.Lock()
	defer g.mu.Unlock()

	var options middlewareTypeOptions
	for _, opt := range opts {
		opt(&options)
	}

	if !options.allowOverride {
		if _, exists := g.middlewareFactories[middlewareType]; exists {
			panic(fmt.Sprintf("middleware type %s already registered", middlewareType))
		}
	}

	g.middlewareFactories[middlewareType] = factory
}

// RegisterMiddlewareName registers a middleware entry by name, associating it with a type and config.
// This allows creating multiple middleware instances from the same factory with different configurations.
//
// Example:
//
//	g.RegisterMiddlewareType("logger", loggerFactory)
//	g.RegisterMiddlewareName("logger-debug", "logger", map[string]any{"level": "debug"})
//	g.RegisterMiddlewareName("logger-info", "logger", map[string]any{"level": "info"})
func (g *GlobalRegistry) RegisterMiddlewareName(name, middlewareType string, config map[string]any, opts ...MiddlewareNameOption) {
	var options middlewareNameOptions
	for _, opt := range opts {
		opt(&options)
	}

	if !options.allowOverride {
		if _, exists := g.middlewareEntries.Load(name); exists {
			panic(fmt.Sprintf("middleware name %s already registered", name))
		}
	}

	g.middlewareEntries.Store(name, &MiddlewareEntry{
		Type:   middlewareType,
		Config: config,
	})
}

// GetMiddlewareFactory returns the middleware factory
func (g *GlobalRegistry) GetMiddlewareFactory(middlewareType string) MiddlewareFactory {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.middlewareFactories[middlewareType]
}

// RegisterMiddleware registers a middleware instance by name (direct registration)
func (g *GlobalRegistry) RegisterMiddleware(name string, mw request.HandlerFunc) {
	if _, exists := g.middlewareInstances.Load(name); exists {
		panic(fmt.Sprintf("middleware %s already registered", name))
	}
	g.middlewareInstances.Store(name, mw)
}

// GetMiddleware retrieves a middleware instance by name
func (g *GlobalRegistry) GetMiddleware(name string) (request.HandlerFunc, bool) {
	if v, ok := g.middlewareInstances.Load(name); ok {
		return v.(request.HandlerFunc), true
	}
	return nil, false
}

// CreateMiddleware creates a middleware instance from definition
// Supports inline parameters syntax: "middleware-name param1="value1", param2="value2"
//
// Examples:
//   - "recovery" - Load middleware without params
//   - "cors" - Load from RegisterMiddlewareName if exists, or factory with nil config
//   - "rate-limit max=100, window="1m"" - Load factory with inline params
func (g *GlobalRegistry) CreateMiddleware(name string) request.HandlerFunc {
	// Parse name and extract inline parameters
	middlewareName, inlineParams := parseMiddlewareName(name)

	// Step 1: First check if already instantiated
	cacheKey := name // Use full name as cache key to support different params
	if mw, ok := g.middlewareInstances.Load(cacheKey); ok {
		return mw.(request.HandlerFunc)
	}

	// Step 2: Check if it's registered via RegisterMiddlewareName (factory pattern)
	if entryAny, ok := g.middlewareEntries.Load(middlewareName); ok {
		entry := entryAny.(*MiddlewareEntry)
		factory := g.GetMiddlewareFactory(entry.Type)
		if factory != nil {
			// Merge inline params with registered config (inline takes precedence)
			config := mergeConfig(entry.Config, inlineParams)
			mw := factory(config)
			if handlerFunc, ok := mw.(request.HandlerFunc); ok {
				// Cache it
				g.middlewareInstances.Store(cacheKey, handlerFunc)
				return handlerFunc
			}
		}
		return nil
	}

	// Step 3: If not found in entries, assume middlewareName is a factory type
	// Create directly from factory with inline params (or nil)
	factory := g.GetMiddlewareFactory(middlewareName)
	if factory != nil {
		var config map[string]any
		if len(inlineParams) > 0 {
			config = inlineParams
		}
		mw := factory(config)
		if handlerFunc, ok := mw.(request.HandlerFunc); ok {
			// Cache it
			g.middlewareInstances.Store(cacheKey, handlerFunc)
			return handlerFunc
		}
	}

	return nil
}

// parseMiddlewareName parses middleware name and extracts inline parameters
// Supports both quoted and unquoted values:
//   - With quotes: "mw-test param1="value1", param2="value2""
//   - Without quotes: "mw-test param1=value1, param2=value2"
//   - Mixed: "mw-test max=100, url="https://example.com""
//
// Output: ("mw-test", map[string]any{"param1": "value1", "param2": "value2"})
func parseMiddlewareName(input string) (string, map[string]any) {
	input = strings.TrimSpace(input)

	// Find first space (separator between name and params)
	spaceIdx := strings.IndexAny(input, " \t")
	if spaceIdx == -1 {
		// No parameters
		return input, nil
	}

	name := strings.TrimSpace(input[:spaceIdx])
	paramsStr := strings.TrimSpace(input[spaceIdx+1:])

	if paramsStr == "" {
		return name, nil
	}

	// Parse parameters: key="value" or key=value
	params := make(map[string]any)

	// Simple state machine parser
	var key, value strings.Builder
	inQuotes := false
	inValue := false
	escaped := false

	for i := 0; i < len(paramsStr); i++ {
		ch := paramsStr[i]

		if escaped {
			if inValue {
				value.WriteByte(ch)
			} else {
				key.WriteByte(ch)
			}
			escaped = false
			continue
		}

		switch ch {
		case '\\':
			escaped = true
		case '"':
			inQuotes = !inQuotes
			// Don't include quotes in the value
		case '=':
			if !inQuotes && !inValue {
				inValue = true
			} else if inValue {
				value.WriteByte(ch)
			}
		case ',':
			if !inQuotes {
				// End of key-value pair
				if k := strings.TrimSpace(key.String()); k != "" {
					params[k] = strings.TrimSpace(value.String())
				}
				key.Reset()
				value.Reset()
				inValue = false
			} else {
				value.WriteByte(ch)
			}
		case ' ', '\t':
			if inQuotes {
				// Keep spaces inside quotes
				if inValue {
					value.WriteByte(ch)
				} else {
					key.WriteByte(ch)
				}
			} else if inValue {
				// Space outside quotes while in value:
				// Could be end of unquoted value or just whitespace before comma
				// Peek ahead to see if comma or end of string
				if i+1 < len(paramsStr) {
					nextCh := paramsStr[i+1]
					if nextCh == ',' || nextCh == ' ' || nextCh == '\t' {
						// This is end of unquoted value, save it
						if k := strings.TrimSpace(key.String()); k != "" {
							params[k] = strings.TrimSpace(value.String())
						}
						key.Reset()
						value.Reset()
						inValue = false
						// Skip remaining whitespace until comma
						for i+1 < len(paramsStr) && (paramsStr[i+1] == ' ' || paramsStr[i+1] == '\t') {
							i++
						}
					} else {
						// Space is part of unquoted value (unusual but allowed)
						value.WriteByte(ch)
					}
				} else {
					// End of string, save the pair
					if k := strings.TrimSpace(key.String()); k != "" {
						params[k] = strings.TrimSpace(value.String())
					}
				}
			}
			// Skip whitespace outside quotes and outside value
		default:
			if inValue {
				value.WriteByte(ch)
			} else {
				key.WriteByte(ch)
			}
		}
	}

	// Add last key-value pair
	if k := strings.TrimSpace(key.String()); k != "" {
		params[k] = strings.TrimSpace(value.String())
	}

	return name, params
}

// mergeConfig merges two config maps, with override taking precedence
func mergeConfig(base, override map[string]any) map[string]any {
	if base == nil && override == nil {
		return nil
	}
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	result := make(map[string]any, len(base)+len(override))
	for k, v := range base {
		result[k] = v
	}
	for k, v := range override {
		result[k] = v
	}
	return result
}
