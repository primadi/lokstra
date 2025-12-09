package resolver

import (
	"fmt"
	"strings"
)

// ResolveSingleValue resolves a single value (not YAML content)
// Used for resolving configs.server before applying overrides
// Only resolves ${ENV:...} and ${@provider:...}, NOT ${@cfg:...}
func ResolveSingleValue(value string) string {
	if !strings.Contains(value, "${") {
		return value
	}

	// Find and replace all ${...} placeholders (except ${@cfg:...})
	result := value
	pos := 0
	for {
		start := strings.Index(result[pos:], "${")
		if start == -1 {
			break
		}
		start += pos

		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}
		end += start

		placeholder := result[start+2 : end]

		// Skip @cfg placeholders
		if strings.HasPrefix(placeholder, "@cfg:") {
			pos = end + 1
			continue
		}

		// Resolve using provider registry
		resolved := resolvePlaceholder(placeholder)

		// Replace placeholder with resolved value
		result = result[:start] + resolved + result[end+1:]

		// Update position to after the resolved value
		pos = start + len(resolved)
	}

	return result
}

// resolveYAMLBytesStep1 resolves all ${...} placeholders EXCEPT ${@cfg:...}
// This is STEP 1 of 2-step resolution process
// Resolves: ${ENV_VAR}, ${@env:VAR}, ${@aws-secret:key}, ${@vault:path}, etc.
// Skips: ${@cfg:KEY} (needs configs map from step 1 result)
func ResolveYAMLBytesStep1(data []byte) []byte {
	content := string(data)

	// Find and replace all ${...} placeholders (except ${@cfg:...})
	// Track position to avoid re-processing
	pos := 0
	for {
		start := strings.Index(content[pos:], "${")
		if start == -1 {
			break
		}
		start += pos

		end := strings.Index(content[start:], "}")
		if end == -1 {
			// Unclosed placeholder - leave as is
			break
		}
		end += start

		placeholder := content[start+2 : end]

		// Skip @cfg placeholders (will be resolved in step 2)
		if strings.HasPrefix(placeholder, "@cfg:") {
			// Move position past this placeholder and continue
			pos = end + 1
			continue
		}

		// Resolve using provider registry
		resolved := resolvePlaceholder(placeholder)

		// Replace placeholder with resolved value
		content = content[:start] + resolved + content[end+1:]

		// Update position to after the resolved value
		pos = start + len(resolved)
	}

	return []byte(content)
}

// resolveYAMLBytesStep2 resolves ${@cfg:...} placeholders using configs map
// This is STEP 2 of 2-step resolution process
//
// Format: ${@cfg:KEY} or ${@cfg:KEY:default}
//
// Supports nested keys using dot notation: email_smtp.host
//
// Examples:
//
//	${@cfg:db.host} -> key="db.host"
//	${@cfg:db.host:localhost} -> key="db.host", default="localhost"
//	${@cfg:server.port:3000} -> key="server.port", default="3000"
func ResolveYAMLBytesStep2(data []byte, configs map[string]any) []byte {
	content := string(data)

	// Find and replace all ${@cfg:...} placeholders
	for {
		start := strings.Index(content, "${@cfg:")
		if start == -1 {
			break
		}

		end := strings.Index(content[start:], "}")
		if end == -1 {
			// Unclosed placeholder - leave as is
			break
		}
		end += start

		// Extract key: ${@cfg:KEY} -> KEY
		key := content[start+7 : end] // 7 = len("${@cfg:")

		// Parse key:default format (split on FIRST ':')
		// This allows default values to contain ':' (e.g., URLs, DSNs)
		var configKey, defaultValue string
		if firstColon := strings.Index(key, ":"); firstColon != -1 {
			configKey = key[:firstColon]
			defaultValue = key[firstColon+1:]
		} else {
			configKey = key
		}

		// Lookup in configs - support nested keys with dot notation
		var resolved string
		val := getNestedConfig(configs, configKey)
		if val != nil {
			resolved = fmt.Sprintf("%v", val)
		} else if defaultValue != "" {
			// User provided explicit default value
			resolved = defaultValue
		} else {
			// Key not found and no explicit default - use empty string
			// This is safer than panic, allows app to continue
			resolved = ""
		}

		// Replace placeholder with resolved value
		content = content[:start] + resolved + content[end+1:]
	}

	return []byte(content)
}

// getNestedConfig retrieves a value from nested config map using dot notation
// Example: "email_smtp.host" -> configs["email_smtp"]["host"]
func getNestedConfig(configs map[string]any, key string) any {
	parts := strings.Split(key, ".")

	var current any = configs
	for i, part := range parts {
		// Try case-insensitive lookup at current level
		m, ok := current.(map[string]any)
		if !ok {
			return nil
		}

		// Try exact match first
		val, found := m[part]
		if !found {
			// Try lowercase
			val, found = m[strings.ToLower(part)]
			if !found {
				// Debug: show what keys are available
				if i == 0 {
					fmt.Printf("   Available keys at root: %v\n", getMapKeys(m))
				}
				return nil
			}
		}

		current = val
	}

	return current
}

// getMapKeys returns all keys from a map (for debugging)
func getMapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// getConfigKeysDebug recursively lists all available config keys for debugging
// func getConfigKeysDebug(configs map[string]any, prefix string) []string {
// 	var keys []string
// 	for k, v := range configs {
// 		fullKey := k
// 		if prefix != "" {
// 			fullKey = prefix + "." + k
// 		}
// 		keys = append(keys, fullKey)

// 		// If value is a map, recurse
// 		if nested, ok := v.(map[string]any); ok {
// 			nestedKeys := getConfigKeysDebug(nested, fullKey)
// 			keys = append(keys, nestedKeys...)
// 		}
// 	}
// 	return keys
// }

// func min(a, b int) int {
// 	if a < b {
// 		return a
// 	}
// 	return b
// }

// resolvePlaceholder resolves a placeholder using provider registry
// Formats supported:
//   - VAR_NAME -> @env provider (default)
//   - VAR_NAME:default -> @env provider with default
//   - @provider:key -> custom provider
//   - @provider:key:default -> custom provider with default
//
// IMPORTANT: Keys CANNOT contain ':' character (reserved as separator)
//
// Examples:
//
//	${DB_HOST} -> provider="env", key="DB_HOST"
//	${DB_HOST:localhost} -> provider="env", key="DB_HOST", default="localhost"
//	${@env:DB_HOST} -> explicit env provider
//	${@vault:secret/data/db} -> provider="vault", key="secret/data/db"
//	${@vault:secret/data/db:password} -> key="secret/data/db", default="password"
//	${@aws-secret:db/password:fallback} -> key="db/password", default="fallback"
func resolvePlaceholder(placeholder string) string {
	var providerName string
	var key string
	var defaultValue string

	// Check if it's a custom provider (@provider:key)
	if strings.HasPrefix(placeholder, "@") {
		// Format: @provider:key:default or @provider:key
		afterAt := placeholder[1:]
		firstColon := strings.Index(afterAt, ":")
		if firstColon == -1 {
			// Invalid format - no ':' after @provider
			return "${" + placeholder + "}"
		}

		providerName = afterAt[:firstColon]
		restAfterProvider := afterAt[firstColon+1:]

		// Parse key:default (split on FIRST ':')
		// This allows default values to contain ':' (e.g., URLs, DSNs)
		if firstColonInRest := strings.Index(restAfterProvider, ":"); firstColonInRest != -1 {
			key = restAfterProvider[:firstColonInRest]
			defaultValue = restAfterProvider[firstColonInRest+1:]
		} else {
			key = restAfterProvider
		}
	} else {
		// Default to @env provider
		// Format: VAR_NAME:default or VAR_NAME
		providerName = "env"

		// Parse key:default (split on FIRST ':')
		// This allows default values to contain ':' (e.g., URLs, DSNs)
		if firstColon := strings.Index(placeholder, ":"); firstColon != -1 {
			key = placeholder[:firstColon]
			defaultValue = placeholder[firstColon+1:]
		} else {
			key = placeholder
		}
	}

	// Get provider from registry
	provider := GetProvider(providerName)
	if provider == nil {
		// Provider not found - return original or default
		if defaultValue != "" {
			return defaultValue
		}
		return "${" + placeholder + "}"
	}

	// Resolve using provider
	if value, ok := provider.Resolve(key); ok {
		return value
	}

	return defaultValue
}
