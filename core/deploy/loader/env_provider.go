package loader

import "os"

// envProvider implements environment variable resolution
type envProvider struct{}

func (p *envProvider) Name() string {
	return "env"
}

func (p *envProvider) Resolve(key string) (string, bool) {
	// Try environment variable first
	if value := os.Getenv(key); value != "" {
		return value, true
	}

	// Try command-line flag
	if value, ok := getCommandLineParam(key); ok {
		return value, true
	}

	return "", false
}
