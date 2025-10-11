package config

import (
	"regexp"
	"strings"
)

var commonSuffixes = []string{
	"Service",
	"Svc",
	"Impl",
	"Handler",
	"Server",
	"Controller",
}

// precompiled regex to remove trailing versioning like _v1, _v2
var versionSuffixRe = regexp.MustCompile(`(?i)_v\d+$`)

func extractResourceNameFromType(serviceType string) string {
	name := serviceType

	// 1. if there is an underscore, take the part before it
	if idx := strings.Index(name, "_"); idx != -1 {
		name = name[:idx]
	}

	// 2. Remove common suffixes (Service, Impl, Handler, etc.)
	for _, suffix := range commonSuffixes {
		if strings.HasSuffix(name, suffix) {
			name = strings.TrimSuffix(name, suffix)
			break
		}
	}

	// 3. Remove versioning like `_v1`, `_V2`
	name = versionSuffixRe.ReplaceAllString(name, "")

	// 4. Normalize casing (optional)
	return strings.ToLower(strings.TrimSpace(name))
}
