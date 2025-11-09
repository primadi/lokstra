package resolver

import "os"

// EnvResolver resolves from command-line parameters first, then environment variables
// Priority: 1. Command params (-KEY=value) 2. Environment variables ($KEY)
type EnvResolver struct{}

func NewEnvResolver() *EnvResolver {
	return &EnvResolver{}
}

func (e *EnvResolver) Name() string {
	return "env"
}

func (e *EnvResolver) Resolve(key string) (string, bool) {
	// Priority 1: Check command-line parameters first
	if value, ok := getCommandLineParam(key); ok {
		return value, true
	}

	// Priority 2: Check environment variables
	value, ok := os.LookupEnv(key)
	return value, ok
}
