package internal

import "strings"

func IsEnabled(param string, def bool) bool {
	if param == "" {
		return def
	}
	return strings.ToLower(param) != "false"
}
