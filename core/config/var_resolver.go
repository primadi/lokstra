package config

import (
	"os"
	"strings"
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

// expandVariables replaces placeholders in the form ${KEY}, ${KEY:default}, ${ENV:KEY}, ${ENV:KEY:default}
func expandVariables(input string) string {
	return os.Expand(input, func(key string) string {
		parts := strings.SplitN(key, ":", 3)

		source := "ENV"
		k := ""
		def := ""

		switch len(parts) {
		case 1:
			k = parts[0]
		case 2:
			if _, isResolver := variableResolvers[parts[0]]; isResolver {
				source = parts[0]
				k = parts[1]
			} else {
				k = parts[0]
				def = parts[1]
			}
		case 3:
			source = parts[0]
			k = parts[1]
			def = parts[2]
		}

		if resolver, ok := variableResolvers[source]; ok {
			val, found := resolver.Resolve(source, k, def)
			if found && val != "" {
				return val
			}
		}
		return def
	})
}
