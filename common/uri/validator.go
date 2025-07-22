package uri

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

func ValidateServiceURI(uri string) error {
	u, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("invalid URI: %w", err)
	}
	if u.Scheme != "lokstra" {
		return fmt.Errorf("invalid scheme: %s", u.Scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("missing serviceType (interface)")
	}
	instance := strings.Trim(u.Path, "/")
	if instance == "" {
		return fmt.Errorf("missing service instance name")
	}

	parts := strings.Split(u.Host, ".")
	var iface string
	switch len(parts) {
	case 1:
		iface = parts[0]
	case 2:
		// package.Interface
		iface = parts[1]
	default:
		return fmt.Errorf("invalid serviceType format: %s", u.Host)
	}

	if !isCamelCase(iface) {
		return fmt.Errorf("interface name must be CamelCase: %s", iface)
	}

	return nil
}

func isCamelCase(s string) bool {
	if s == "" || !unicode.IsUpper(rune(s[0])) {
		return false
	}
	return !strings.Contains(s, "_")
}
