package resolver

import (
	"fmt"
	"strings"
)

// resolveYAMLBytesStep1 resolves all ${...} placeholders EXCEPT ${@cfg:...}
// This is STEP 1 of 2-step resolution process
// Resolves: ${ENV_VAR}, ${@env:VAR}, ${@aws-secret:key}, ${@vault:path}, etc.
// Skips: ${@cfg:KEY} (needs configs map from step 1 result)
func ResolveYAMLBytesStep1(data []byte) []byte {
	content := string(data)

	// Find and replace all ${...} placeholders (except ${@cfg:...})
	for {
		start := strings.Index(content, "${")
		if start == -1 {
			break
		}

		end := strings.Index(content[start:], "}")
		if end == -1 {
			// Unclosed placeholder - leave as is
			break
		}
		end += start

		placeholder := content[start+2 : end]

		// Skip @cfg placeholders (will be resolved in step 2)
		if strings.HasPrefix(placeholder, "@cfg:") {
			// Continue searching after this placeholder
			nextStart := strings.Index(content[end+1:], "${")
			if nextStart == -1 {
				break
			}
			content = content[:end+1] + content[end+1:]
			continue
		}

		// Resolve using provider registry
		resolved := resolvePlaceholder(placeholder)

		// Replace placeholder with resolved value
		content = content[:start] + resolved + content[end+1:]
	}

	return []byte(content)
}

// resolveYAMLBytesStep2 resolves ${@cfg:...} placeholders using configs map
// This is STEP 2 of 2-step resolution process
//
// Format: ${@cfg:KEY} or ${@cfg:KEY:default}
//
// IMPORTANT: Config keys CANNOT contain ':' character (use '.' for nesting)
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

		// Lookup in configs (case-insensitive)
		var resolved string
		if val, ok := configs[strings.ToLower(configKey)]; ok {
			resolved = fmt.Sprintf("%v", val)
		} else if val, ok := configs[configKey]; ok {
			resolved = fmt.Sprintf("%v", val)
		} else if defaultValue != "" {
			resolved = defaultValue
		} else {
			// Not found - keep original for debugging
			resolved = "${@cfg:" + key + "}"
		}

		// Replace placeholder with resolved value
		content = content[:start] + resolved + content[end+1:]
	}

	return []byte(content)
}

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

	// Use default value if provided
	if defaultValue != "" {
		return defaultValue
	}

	// Not found - return original placeholder for debugging
	return "${" + placeholder + "}"
}
