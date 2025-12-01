package resolver

import (
	"os"
	"strings"
)

// envProvider implements environment variable resolution
type envProvider struct{}

func (p *envProvider) Name() string {
	return "env"
}

func (p *envProvider) Resolve(key string) (string, bool) {
	// Try command-line flag first (highest priority - explicit override)
	if value, ok := getCommandLineParam(key); ok {
		return value, true
	}

	// Then try environment variable (lower priority - deployment config)
	if value := os.Getenv(key); value != "" {
		return value, true
	}

	return "", false
}

// getCommandLineParam extracts value from command-line arguments
// Uses flag parsing to properly handle -KEY=value format
func getCommandLineParam(key string) (string, bool) {
	// Simple implementation - parse os.Args for -KEY=value or --KEY=value
	keyLower := strings.ToLower(key)

	for _, arg := range os.Args[1:] {
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) == 2 {
				argKey := strings.TrimPrefix(parts[0], "--")
				argKey = strings.TrimPrefix(argKey, "-")
				argKey = strings.ToLower(argKey)

				if argKey == keyLower {
					return parts[1], true
				}
			}
		}
	}

	return "", false
}
