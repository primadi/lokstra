package utils

import "strconv"

func ParseInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	v, err := strconv.Atoi(s)
	if err != nil || v <= 0 {
		return defaultValue
	}
	return v
}
