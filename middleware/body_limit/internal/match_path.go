package internal

import (
	"path/filepath"
	"strings"
)

// MatchPath checks if a path matches a pattern
// Supports basic wildcard patterns with * and **
func MatchPath(requestPath, pattern string) bool {
	// Direct match
	if requestPath == pattern {
		return true
	}

	// Handle ** patterns (match any number of path segments)
	if strings.Contains(pattern, "**") {
		parts := strings.SplitN(pattern, "**", 2)
		if len(parts) == 2 {
			prefix := parts[0]
			suffix := parts[1]

			prefixMatch := prefix == "" || strings.HasPrefix(requestPath, prefix)
			suffixMatch := suffix == "" || strings.HasSuffix(requestPath, suffix)

			return prefixMatch && suffixMatch
		}
	}

	// Handle single * patterns (should not match path separators)
	if strings.Contains(pattern, "*") && !strings.Contains(pattern, "**") {
		// Split pattern by / to handle segments properly
		patternParts := strings.Split(pattern, "/")
		pathParts := strings.Split(requestPath, "/")

		// Must have same number of segments for single * match
		if len(patternParts) != len(pathParts) {
			return false
		}

		for i, patternPart := range patternParts {
			if patternPart == "*" {
				continue // * matches any single segment
			}
			if matched, err := filepath.Match(patternPart, pathParts[i]); err != nil || !matched {
				return false
			}
		}
		return true
	}

	// Fallback to filepath.Match for patterns without wildcards
	if matched, err := filepath.Match(pattern, requestPath); err == nil && matched {
		return true
	}

	return false
}
