package resolver

import (
	"fmt"
	"strings"
)

// Registry manages all available resolvers
type Registry struct {
	resolvers map[string]Resolver
}

// NewRegistry creates a new resolver registry with default resolvers
func NewRegistry() *Registry {
	r := &Registry{
		resolvers: make(map[string]Resolver),
	}

	// Register default resolver (environment variables)
	r.Register(NewEnvResolver())

	return r
}

// Register adds a resolver to the registry
func (r *Registry) Register(resolver Resolver) {
	r.resolvers[resolver.Name()] = resolver
}

// Get returns a resolver by name
func (r *Registry) Get(name string) Resolver {
	return r.resolvers[name]
}

// ResolveValue resolves a single value that may contain ${...} placeholders
// Supports multiple formats:
//   - Static value: "localhost"
//   - Env var: ${DB_HOST} or ${DB_HOST:default}
//   - Custom resolver: ${@consul:path:default} or ${@aws-ssm:/path}
//
// Special resolver @cfg:
//   - ${@cfg:KEY} - resolves from configs map (must be resolved in step 2)
//
// Resolution happens in 2 steps:
//  1. Resolve all ${...} except ${@cfg:...}
//  2. Resolve ${@cfg:...} (using step 1 results)
func (r *Registry) ResolveValue(value string, configs map[string]any) (any, error) {
	// Step 1: Resolve all non-@cfg placeholders
	step1Result, err := r.resolveStep1(value)
	if err != nil {
		return nil, err
	}

	// Step 2: Resolve @cfg placeholders using configs map
	step2Result, err := r.resolveStep2(step1Result, configs)
	if err != nil {
		return nil, err
	}

	return step2Result, nil
}

// Step 1: Resolve all ${...} except ${@cfg:...}
func (r *Registry) resolveStep1(value string) (string, error) {
	result := value

	// Find all ${...} placeholders
	for {
		start := strings.Index(result, "${")
		if start == -1 {
			break
		}

		end := strings.Index(result[start:], "}")
		if end == -1 {
			return "", fmt.Errorf("unclosed placeholder in: %s", value)
		}
		end += start

		placeholder := result[start+2 : end]

		// Skip @cfg placeholders (resolve in step 2)
		if strings.HasPrefix(placeholder, "@cfg:") {
			// Find next placeholder after this one
			nextStart := strings.Index(result[end+1:], "${")
			if nextStart == -1 {
				break
			}
			// Continue searching after current placeholder
			result = result[:end+1] + result[end+1:]
			continue
		}

		// Resolve this placeholder
		resolved, err := r.resolvePlaceholder(placeholder)
		if err != nil {
			return "", err
		}

		// Replace placeholder with resolved value
		result = result[:start] + resolved + result[end+1:]
	}

	return result, nil
}

// Step 2: Resolve ${@cfg:...} placeholders
func (r *Registry) resolveStep2(value string, configs map[string]any) (any, error) {
	result := value
	hasCfgPlaceholder := false

	// Find all ${@cfg:...} placeholders
	for {
		start := strings.Index(result, "${@cfg:")
		if start == -1 {
			break
		}

		hasCfgPlaceholder = true
		end := strings.Index(result[start:], "}")
		if end == -1 {
			return nil, fmt.Errorf("unclosed @cfg placeholder in: %s", value)
		}
		end += start

		// Extract key: ${@cfg:KEY} -> KEY
		key := result[start+7 : end] // 7 = len("${@cfg:")

		// Lookup in configs
		cfgValue, ok := configs[key]
		if !ok {
			return nil, fmt.Errorf("config key %s not found (referenced in ${@cfg:%s})", key, key)
		}

		// If the entire value is just this placeholder, return the actual type
		if start == 0 && end == len(result)-1 {
			return cfgValue, nil
		}

		// Otherwise, convert to string and continue replacing
		cfgStr := fmt.Sprintf("%v", cfgValue)
		result = result[:start] + cfgStr + result[end+1:]
	}

	// If no @cfg placeholder found, return as string
	if !hasCfgPlaceholder {
		return result, nil
	}

	return result, nil
}

// resolvePlaceholder resolves a single placeholder (without ${ })
// Formats:
//   - VAR_NAME or VAR_NAME:default -> environment variable (default resolver)
//   - @resolver:key or @resolver:key:default -> custom resolver
func (r *Registry) resolvePlaceholder(placeholder string) (string, error) {
	var resolverName string
	var key string
	var defaultValue string
	var hasDefault bool

	// Check if it's a custom resolver (@resolver:key)
	if strings.HasPrefix(placeholder, "@") {
		// Format: @resolver:key:default or @resolver:key
		parts := strings.SplitN(placeholder[1:], ":", 3)
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid resolver format: ${%s} (expected ${@resolver:key} or ${@resolver:key:default})", placeholder)
		}

		resolverName = parts[0]
		key = parts[1]
		if len(parts) == 3 {
			defaultValue = parts[2]
			hasDefault = true
		}
	} else {
		// Default resolver (environment variable)
		// Format: VAR_NAME:default or VAR_NAME
		parts := strings.SplitN(placeholder, ":", 2)
		resolverName = "env" // Default to env resolver
		key = parts[0]
		if len(parts) == 2 {
			defaultValue = parts[1]
			hasDefault = true
		}
	}

	// Get resolver
	resolver := r.Get(resolverName)
	if resolver == nil {
		return "", fmt.Errorf("resolver %s not found (in ${%s})", resolverName, placeholder)
	}

	// Resolve value
	value, ok := resolver.Resolve(key)
	if ok {
		return value, nil
	}

	// Use default if available
	if hasDefault {
		return defaultValue, nil
	}

	// Error: not found and no default
	return "", fmt.Errorf("failed to resolve ${%s}: key %s not found in resolver %s", placeholder, key, resolverName)
}
