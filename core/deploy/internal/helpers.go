package internal

import "strings"

// ParseMiddlewareName parses middleware name and optional inline parameters
// Examples:
//   - "cors" -> ("cors", nil)
//   - "cors origin=http://localhost:3000" -> ("cors", {"origin": "http://localhost:3000"})
//   - With quotes: "mw-test param1="value1", param2="value2""
//   - Without quotes: "mw-test param1=value1, param2=value2"
//   - Mixed: "mw-test max=100, url="https://example.com""
//
// Output: ("mw-test", map[string]any{"param1": "value1", "param2": "value2"})
func ParseMiddlewareName(input string) (string, map[string]any) {
	input = strings.TrimSpace(input)

	// Find first space (separator between name and params)
	spaceIdx := strings.IndexAny(input, " \t")
	if spaceIdx == -1 {
		// No parameters - just name
		return input, nil
	}

	// Extract name and params string
	name := input[:spaceIdx]
	paramsStr := strings.TrimSpace(input[spaceIdx+1:])

	if paramsStr == "" {
		return name, nil
	}

	// Parse parameters
	params := make(map[string]any)
	var key, value string
	var inQuotes bool
	var quoteChar rune
	var buffer strings.Builder

	state := "key" // "key" or "value"

	for i, ch := range paramsStr {
		switch {
		case ch == '"' || ch == '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = ch
			} else if ch == quoteChar {
				inQuotes = false
			} else {
				buffer.WriteRune(ch)
			}

		case ch == '=' && !inQuotes && state == "key":
			key = strings.TrimSpace(buffer.String())
			buffer.Reset()
			state = "value"

		case ch == ',' && !inQuotes:
			// End of this parameter
			value = strings.TrimSpace(buffer.String())
			if key != "" {
				params[key] = value
			}
			buffer.Reset()
			key = ""
			value = ""
			state = "key"

		case i == len(paramsStr)-1:
			// Last character
			buffer.WriteRune(ch)
			value = strings.TrimSpace(buffer.String())
			if key != "" {
				params[key] = value
			}

		default:
			buffer.WriteRune(ch)
		}
	}

	// Handle case where there's no trailing comma
	if key != "" && state == "value" {
		value = strings.TrimSpace(buffer.String())
		params[key] = value
	}

	if len(params) == 0 {
		return name, nil
	}

	return name, params
}

// MergeConfig merges two config maps, with override taking precedence
func MergeConfig(base, override map[string]any) map[string]any {
	if base == nil && override == nil {
		return nil
	}
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	result := make(map[string]any, len(base)+len(override))

	// Copy base
	for k, v := range base {
		result[k] = v
	}

	// Override with new values
	for k, v := range override {
		result[k] = v
	}

	return result
}
