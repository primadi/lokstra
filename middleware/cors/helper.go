package cors

import "strings"

func matchOrigin(whitelist []string, origin string) bool {
	for _, allowed := range whitelist {
		allowed = strings.TrimSpace(allowed)
		if allowed == "*" {
			return true
		}
		if suffix, ok := strings.CutPrefix(allowed, "*"); ok {
			if strings.HasSuffix(origin, suffix) {
				return true
			}
		}
		if prefix, ok := strings.CutSuffix(allowed, "*"); ok {
			if strings.HasPrefix(origin, prefix) {
				return true
			}
		}
		if origin == allowed {
			return true
		}
	}

	return false
}
