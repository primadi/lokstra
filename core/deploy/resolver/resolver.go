package resolver

import (
	"os"
	"strings"
	"sync"
)

// Resolver resolves configuration values from various sources
type Resolver interface {
	// Name returns the resolver name (e.g., "env", "consul", "aws-ssm")
	Name() string

	// Resolve resolves a key to its value
	// Returns the resolved value and whether it was found
	Resolve(key string) (string, bool)
}

var (
	// flagValues caches parsed flag values to avoid re-parsing
	flagValues map[string]string
	flagOnce   sync.Once
)

// parseFlags parses command-line flags once and caches the results
// Supports multiple formats:
//
//	-KEY=value    (single dash with equals)
//	--KEY=value   (double dash with equals)
//	-KEY value    (single dash with space)
//	--KEY value   (double dash with space)
func parseFlags() {
	flagValues = make(map[string]string)
	args := os.Args[1:]

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Skip non-flag arguments
		if !strings.HasPrefix(arg, "-") {
			continue
		}

		var key, value string
		var hasValue bool

		// Handle formats with equals sign: -KEY=value or --KEY=value
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) == 2 {
				key = cleanKey(parts[0]) // Remove - or -- prefix
				value = parts[1]
				hasValue = true
			}
		} else {
			// Handle formats with space: -KEY value or --KEY value
			key = cleanKey(arg) // Remove - or -- prefix

			// Check if next arg exists and is not a flag (value)
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				value = args[i+1]
				hasValue = true
				i++ // Skip next arg since we consumed it as value
			}
		}

		// Store the key-value pair if we found both
		if hasValue && key != "" {
			flagValues[key] = value
		}
	}
}

// cleanKey removes leading dashes from flag key
// --KEY -> KEY, -KEY -> KEY
func cleanKey(flagArg string) string {
	key := flagArg

	// Remove leading dashes
	if strings.HasPrefix(key, "--") {
		key = key[2:]
	} else if strings.HasPrefix(key, "-") {
		key = key[1:]
	}

	return strings.ToLower(key)
}

// getCommandLineParam extracts value from command-line arguments
// Uses flag parsing to properly handle -KEY=value format
func getCommandLineParam(key string) (string, bool) {
	flagOnce.Do(parseFlags)

	value, ok := flagValues[strings.ToLower(key)]
	return value, ok
}
