package main

import (
	"fmt"
	"strings"
)

func Oldmain() {
	// Test extractPathParamNames directly
	routes := []string{
		"/departments/{dep}/users/{id}",
		"/api/{version}/items/{itemId}",
		"/users/:userId/posts/:postId",
	}

	for _, route := range routes {
		fmt.Printf("Route: %s\n", route)
		fmt.Printf("Params: %v\n\n", extractPathParamNames(route))
	}
}

// Copy the function to test it standalone
func extractPathParamNames(path string) []string {
	var names []string
	segments := splitPath(path)

	for _, seg := range segments {
		if after, ok := strings.CutPrefix(seg, "{"); ok {
			// Check for {param} style
			paramName := strings.TrimSuffix(after, "}")
			if paramName != "" {
				names = append(names, paramName)
			}
		} else if after, ok := strings.CutPrefix(seg, ":"); ok {
			// Check for :param style
			paramName := after
			if paramName != "" {
				names = append(names, paramName)
			}
		}
	}

	return names
}

func splitPath(path string) []string {
	if path == "" {
		return []string{}
	}
	if path[0] != '/' {
		path = "/" + path
	}
	segments := []string{}
	for _, seg := range strings.Split(path, "/") {
		if seg != "" {
			segments = append(segments, seg)
		}
	}
	return segments
}
