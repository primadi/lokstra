package annotation

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// parseFileAnnotations parses all annotations in a file
func parseFileAnnotations(path string) ([]*ParsedAnnotation, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var annotations []*ParsedAnnotation
	scanner := bufio.NewScanner(file)
	lineNum := 0
	var pendingAnnotations []*ParsedAnnotation

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Check for annotation
		if strings.HasPrefix(line, "//") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "//"))
			if strings.HasPrefix(line, "@") {
				ann, err := parseAnnotationLine(line, lineNum)
				if err != nil {
					return nil, fmt.Errorf("line %d: %w", lineNum, err)
				}
				if ann != nil {
					pendingAnnotations = append(pendingAnnotations, ann)
				}
			}
		} else if line != "" && !strings.HasPrefix(line, "//") {
			// Non-comment, non-empty line - attach pending annotations
			if len(pendingAnnotations) > 0 {
				// Extract target name (struct, func, field, etc.)
				targetName := extractTargetNameFromLine(line)
				for _, ann := range pendingAnnotations {
					ann.TargetName = targetName
					annotations = append(annotations, ann)
				}
				pendingAnnotations = nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return annotations, nil
}

// parseAnnotationLine parses a single annotation line
// Supports both formats:
//
//	@RouterService name="user-service", prefix="/api"
//	@RouterService "user-service", "/api"
func parseAnnotationLine(line string, lineNum int) (*ParsedAnnotation, error) {
	// Extract annotation name
	parts := strings.SplitN(line, "(", 2)
	if len(parts) == 1 {
		// No parentheses - might have args without parens or no args
		// @RouterService name="value"
		// @RouterService
		nameAndArgs := strings.TrimSpace(parts[0])
		spaceIdx := strings.Index(nameAndArgs, " ")

		if spaceIdx == -1 {
			// Just annotation name, no args
			return &ParsedAnnotation{
				Name: strings.TrimPrefix(nameAndArgs, "@"),
				Args: make(map[string]interface{}),
				Line: lineNum,
			}, nil
		}

		// Has args without parens
		name := strings.TrimPrefix(nameAndArgs[:spaceIdx], "@")
		argsStr := strings.TrimSpace(nameAndArgs[spaceIdx+1:])

		args, positional, err := parseAnnotationArgs(argsStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse args: %w", err)
		}

		return &ParsedAnnotation{
			Name:           name,
			Args:           args,
			PositionalArgs: positional,
			Line:           lineNum,
		}, nil
	}

	// Has parentheses
	name := strings.TrimSpace(strings.TrimPrefix(parts[0], "@"))
	argsStr := strings.TrimSpace(parts[1])

	// Remove closing paren
	if idx := strings.LastIndex(argsStr, ")"); idx != -1 {
		argsStr = argsStr[:idx]
	}

	args, positional, err := parseAnnotationArgs(argsStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse args: %w", err)
	}

	return &ParsedAnnotation{
		Name:           name,
		Args:           args,
		PositionalArgs: positional,
		Line:           lineNum,
	}, nil
}

// parseAnnotationArgs parses argument string and returns both named and positional args
// Examples:
//
//	name="user-service", prefix="/api"
//	"user-service", "/api", ["recovery", "logger"]
func parseAnnotationArgs(argsStr string) (map[string]interface{}, []interface{}, error) {
	if strings.TrimSpace(argsStr) == "" {
		return make(map[string]interface{}), nil, nil
	}

	namedArgs := make(map[string]interface{})
	var positionalArgs []interface{}

	// Split by comma, but respect quoted strings and arrays
	args := smartSplit(argsStr, ',')

	for i, arg := range args {
		arg = strings.TrimSpace(arg)
		if arg == "" {
			continue
		}

		// Check if it's named (key=value) or positional
		if strings.Contains(arg, "=") {
			// Named argument
			parts := strings.SplitN(arg, "=", 2)
			key := strings.TrimSpace(parts[0])
			valueStr := strings.TrimSpace(parts[1])

			value, err := parseArgValue(valueStr)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid value for %s: %w", key, err)
			}

			namedArgs[key] = value
		} else {
			// Positional argument
			value, err := parseArgValue(arg)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid positional arg %d: %w", i, err)
			}

			positionalArgs = append(positionalArgs, value)
		}
	}

	return namedArgs, positionalArgs, nil
}

// parseArgValue parses a single value (string, array, number, bool)
func parseArgValue(valueStr string) (interface{}, error) {
	valueStr = strings.TrimSpace(valueStr)

	// Array: ["item1", "item2"]
	if strings.HasPrefix(valueStr, "[") && strings.HasSuffix(valueStr, "]") {
		return parseArrayValue(valueStr)
	}

	// String: "value" or 'value'
	if (strings.HasPrefix(valueStr, "\"") && strings.HasSuffix(valueStr, "\"")) ||
		(strings.HasPrefix(valueStr, "'") && strings.HasSuffix(valueStr, "'")) {
		return valueStr[1 : len(valueStr)-1], nil
	}

	// Boolean
	if valueStr == "true" {
		return true, nil
	}
	if valueStr == "false" {
		return false, nil
	}

	// Number
	if num, err := strconv.Atoi(valueStr); err == nil {
		return num, nil
	}
	if num, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return num, nil
	}

	// Unquoted string (treat as string)
	return valueStr, nil
}

// parseArrayValue parses array syntax: ["item1", "item2"]
func parseArrayValue(arrayStr string) ([]string, error) {
	// Remove brackets
	arrayStr = strings.TrimSpace(arrayStr)
	if !strings.HasPrefix(arrayStr, "[") || !strings.HasSuffix(arrayStr, "]") {
		return nil, fmt.Errorf("invalid array syntax: %s", arrayStr)
	}

	content := arrayStr[1 : len(arrayStr)-1]
	if strings.TrimSpace(content) == "" {
		return []string{}, nil
	}

	items := smartSplit(content, ',')
	result := make([]string, 0, len(items))

	for _, item := range items {
		item = strings.TrimSpace(item)
		// Remove quotes if present
		if (strings.HasPrefix(item, "\"") && strings.HasSuffix(item, "\"")) ||
			(strings.HasPrefix(item, "'") && strings.HasSuffix(item, "'")) {
			item = item[1 : len(item)-1]
		}
		if item != "" {
			result = append(result, item)
		}
	}

	return result, nil
}

// smartSplit splits string by delimiter but respects quotes and brackets
func smartSplit(s string, delim rune) []string {
	var result []string
	var current strings.Builder
	inQuote := false
	inArray := 0
	quoteChar := rune(0)

	for _, ch := range s {
		switch {
		case (ch == '"' || ch == '\'') && quoteChar == 0:
			// Start quote
			inQuote = true
			quoteChar = ch
			current.WriteRune(ch)
		case ch == quoteChar:
			// End quote
			inQuote = false
			quoteChar = 0
			current.WriteRune(ch)
		case ch == '[' && !inQuote:
			inArray++
			current.WriteRune(ch)
		case ch == ']' && !inQuote:
			inArray--
			current.WriteRune(ch)
		case ch == delim && !inQuote && inArray == 0:
			// Split here
			result = append(result, current.String())
			current.Reset()
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

// extractTargetNameFromLine extracts the target name from a code line
func extractTargetNameFromLine(line string) string {
	// type StructName struct
	if strings.HasPrefix(line, "type ") {
		re := regexp.MustCompile(`type\s+(\w+)`)
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			return matches[1]
		}
	}

	// func (r *Receiver) MethodName(
	// func FunctionName(
	if strings.HasPrefix(line, "func ") {
		re := regexp.MustCompile(`func\s+(?:\([^)]+\)\s+)?(\w+)`)
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			return matches[1]
		}
	}

	// FieldName type
	re := regexp.MustCompile(`^\s*(\w+)\s+`)
	if matches := re.FindStringSubmatch(line); len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// readArgs reads arguments from ParsedAnnotation
// Supports both named and positional arguments
func (a *ParsedAnnotation) readArgs(expectedArgs ...string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// If we have named args, use them
	if len(a.Args) > 0 {
		// Validate all keys are expected
		for key := range a.Args {
			found := false
			for _, expected := range expectedArgs {
				if key == expected {
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("unexpected argument '%s' (expected: %v)", key, expectedArgs)
			}
		}

		return a.Args, nil
	}

	// Use positional args
	if len(a.PositionalArgs) > len(expectedArgs) {
		return nil, fmt.Errorf("too many arguments: got %d, expected max %d", len(a.PositionalArgs), len(expectedArgs))
	}

	for i, val := range a.PositionalArgs {
		result[expectedArgs[i]] = val
	}

	return result, nil
}
