package config

import "os"

// EnvResolver implements VariableResolverService for OS environment variables.
type EnvResolver struct{}

func (e *EnvResolver) Resolve(source, key, def string) (string, bool) {
	if source != "ENV" {
		return "", false
	}
	val := os.Getenv(key)
	if val == "" {
		return def, false
	}
	return val, true
}

var _ VariableResolver = (*EnvResolver)(nil)
