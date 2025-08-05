package utils

func CamelToSnake(s string) string {
	result := make([]rune, 0, len(s))
	for i, r := range s {
		if i > 0 && isUpper(r) {
			// If it's not the first character and it's uppercase,
			prev := rune(s[i-1])
			if !isUpper(prev) || (i+1 < len(s) && !isUpper(rune(s[i+1]))) {
				result = append(result, '_')
			}
		}
		result = append(result, toLower(r))
	}
	return string(result)
}

func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

func toLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + ('a' - 'A')
	}
	return r
}
