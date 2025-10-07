package config

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// VariableResolverService defines interface for resolving variable keys from sources like ENV, AWS, etc.
type VariableResolver interface {
	Resolve(source string, key string, defaultValue string) (string, bool)
}

// Registry of available variable resolvers.
var variableResolvers = map[string]VariableResolver{
	"ENV": &EnvResolver{},
}

func AddVariableResolver(name string, resolver VariableResolver) {
	if _, exists := variableResolvers[name]; exists {
		panic("Variable resolver already exists: " + name)
	}
	variableResolvers[name] = resolver
}

// expandVariables replaces placeholders in the form:
// - ${KEY}              -> ENV:KEY (no default)
// - ${KEY:DEFAULT}      -> ENV:KEY with default (DEFAULT can contain :)
// - ${@RESOLVER:KEY}    -> Custom resolver (ENV, AWS, VAULT, etc.)
// - ${@RESOLVER:KEY:DEFAULT} -> Custom resolver with default
//
// The @ prefix explicitly specifies a resolver source.
// Without @, ENV resolver is used by default.
//
// Two-Pass Expansion for CFG resolver:
// 1. PASS 1: Expand all resolvers EXCEPT CFG
// 2. Extract "configs" section and store in temporary registry
// 3. PASS 2: Expand CFG resolver using temporary registry
//
// This allows ${@CFG:config-name} to work even before the full config registry is built.
//
// Examples:
//
//	${PORT}                              -> ENV var PORT
//	${PORT:8080}                         -> ENV var PORT, default 8080
//	${BASE_URL:http://localhost}         -> ENV var BASE_URL, default http://localhost
//	${DSN:postgresql://localhost:5432}   -> ENV var DSN with complex default
//	${@ENV:API_KEY}                      -> Explicit ENV resolver
//	${@ENV:API_KEY:fallback}             -> Explicit ENV resolver with default
//	${@AWS:secret-name}                  -> AWS Secrets Manager
//	${@AWS:secret-name:local-secret}     -> AWS with fallback
//	${@VAULT:path/to/secret}             -> HashiCorp Vault
//	${@CFG:database.host}                -> From config registry (two-pass expansion)
func expandVariables(input string) string {
	// PHASE 1: Expand all resolvers EXCEPT CFG
	phase1Result := expandVariablesExcept(input, []string{"CFG"})

	// PHASE 2: Parse YAML to extract "configs" section, store in temp registry, then expand CFG
	phase2Result := expandCFGWithTempRegistry(phase1Result)

	return phase2Result
}

// expandVariablesOnly expands ONLY the specified resolvers
func expandVariablesOnly(input string, onlyResolvers []string) string {
	allowedMap := make(map[string]bool)
	for _, name := range onlyResolvers {
		allowedMap[name] = true
	}
	return expandVariablesInternal(input, allowedMap, false)
}

// expandVariablesExcept expands all resolvers EXCEPT the specified ones
func expandVariablesExcept(input string, exceptResolvers []string) string {
	if len(exceptResolvers) == 0 {
		return expandVariablesInternal(input, nil, false)
	}
	exceptMap := make(map[string]bool)
	for _, name := range exceptResolvers {
		exceptMap[name] = true
	}
	return expandVariablesInternal(input, exceptMap, true)
}

// expandVariablesInternal is the core expansion logic with filtering
// If filterMap is nil, expand all
// If isExceptMode=true, expand all EXCEPT those in filterMap
// If isExceptMode=false, expand ONLY those in filterMap
func expandVariablesInternal(input string, filterMap map[string]bool, isExceptMode bool) string {
	return os.Expand(input, func(key string) string {
		originalKey := key // Save original key for reconstructing placeholders
		var source, k, def string

		if strings.HasPrefix(key, "@") {
			// Format: @RESOLVER:KEY or @RESOLVER:KEY:DEFAULT
			key = key[1:] // Remove @ prefix
			firstColon := strings.Index(key, ":")

			if firstColon == -1 {
				// Invalid: @RESOLVER without KEY
				return ""
			}

			source = key[:firstColon]
			remainder := key[firstColon+1:]

			// Find second colon for default value
			secondColon := strings.Index(remainder, ":")
			if secondColon == -1 {
				// @RESOLVER:KEY (no default)
				k = remainder
				def = ""
			} else {
				// @RESOLVER:KEY:DEFAULT
				k = remainder[:secondColon]
				def = remainder[secondColon+1:] // Everything after second : is default
			}
		} else {
			// Format: KEY or KEY:DEFAULT (ENV resolver by default)
			source = "ENV"
			firstColon := strings.Index(key, ":")

			if firstColon == -1 {
				// ${KEY} - no default
				k = key
				def = ""
			} else {
				// ${KEY:DEFAULT} - everything after first : is default
				k = key[:firstColon]
				def = key[firstColon+1:]
			}
		}

		// Check if this resolver should be processed
		if filterMap != nil {
			if isExceptMode {
				// Except mode: skip if in filterMap
				if filterMap[source] {
					// Don't expand this resolver, return original placeholder
					return "${" + originalKey + "}"
				}
			} else {
				// Only mode: skip if NOT in filterMap
				if !filterMap[source] {
					// Don't expand this resolver, return original placeholder
					return "${" + originalKey + "}"
				}
			}
		}

		// Resolve using the specified resolver
		if resolver, ok := variableResolvers[source]; ok {
			val, found := resolver.Resolve(source, k, def)
			if found && val != "" {
				return val
			}
		}
		return def
	})
}

// expandCFGWithTempRegistry performs PHASE 2 expansion:
// 1. Parse YAML to extract "configs" section
// 2. Store configs in temporary registry (tempCFGResolver)
// 3. Expand CFG resolver placeholders
// 4. Return expanded result
func expandCFGWithTempRegistry(input string) string {
	// Try to parse YAML to extract configs
	var parsed map[string]any
	if err := yaml.Unmarshal([]byte(input), &parsed); err != nil {
		// If YAML parsing fails, just return input as-is (no CFG expansion)
		return input
	}

	// Extract "configs" section
	configsInterface, hasConfigs := parsed["configs"]
	if !hasConfigs {
		// No configs section, just return input as-is (no CFG to expand)
		return input
	}

	// Convert configs to map[string]ConfigEntry
	configsMap, ok := configsInterface.([]any)
	if !ok {
		// Invalid configs format, return input as-is
		return input
	}

	// Build temporary config registry
	tempRegistry := make(map[string]any)
	for _, item := range configsMap {
		configItem, ok := item.(map[string]any)
		if !ok {
			continue
		}

		name, hasName := configItem["name"].(string)
		value, hasValue := configItem["value"]

		if hasName && hasValue {
			tempRegistry[name] = value
		}
	}

	// If no configs to register, return input as-is
	if len(tempRegistry) == 0 {
		return input
	}

	// Create temporary CFG resolver
	tempResolver := &TempCFGResolver{
		configs: tempRegistry,
	}

	// Temporarily register CFG resolver
	oldResolver := variableResolvers["CFG"]
	variableResolvers["CFG"] = tempResolver
	defer func() {
		// Restore old resolver (or remove if didn't exist)
		if oldResolver != nil {
			variableResolvers["CFG"] = oldResolver
		} else {
			delete(variableResolvers, "CFG")
		}
	}()

	// Now expand CFG variables only
	return expandVariablesOnly(input, []string{"CFG"})
}

// TempCFGResolver is a temporary resolver for CFG variables during two-pass expansion
type TempCFGResolver struct {
	configs map[string]any
}

// Resolve implements VariableResolver interface
func (r *TempCFGResolver) Resolve(source string, key string, defaultValue string) (string, bool) {
	if value, ok := r.configs[key]; ok {
		// Convert value to string
		switch v := value.(type) {
		case string:
			return v, true
		case int, int64, float64, bool:
			return toString(v), true
		default:
			// Complex types - use YAML marshal
			if bytes, err := yaml.Marshal(v); err == nil {
				return strings.TrimSpace(string(bytes)), true
			}
		}
	}

	return defaultValue, false
}

// toString converts basic types to string
func toString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return formatInt(int64(v))
	case int64:
		return formatInt(v)
	case float64:
		return formatFloat(v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}

// formatInt converts int64 to string (simple implementation)
func formatInt(n int64) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	// Build string backwards
	var digits []byte
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}

	// Reverse
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}

	if negative {
		return "-" + string(digits)
	}
	return string(digits)
}

// formatFloat converts float64 to string (simple implementation)
func formatFloat(f float64) string {
	if f == 0 {
		return "0"
	}

	// Simple float formatting - integer part only for now
	// For more complex formatting, would need full implementation
	return formatInt(int64(f))
}
