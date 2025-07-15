package router_engine

import (
	"net/http"
	"regexp"
	"strings"
)

// paramPattern is a regular expression that matches path parameters in the format
// :param or *param. It captures both types of parameters:
var paramPattern = regexp.MustCompile(`:([a-zA-Z_][a-zA-Z0-9_]*)|\*([a-zA-Z_][a-zA-Z0-9_]*)`)

// ConvertToServeMuxParamPath converts a path with parameters (e.g., :id or *id)
// to a format suitable for http.ServeMux, which uses {param} syntax.
// It replaces :id with {id} and *id with {id...} to indicate a variadic parameter.
// For example, "/users/:id" becomes "/users/{id}" and "/files/*id" becomes "/files/{id...}".
// This is necessary because http.ServeMux does not support the :param or *param syntax.
// Instead, it uses {param} for single parameters and {param...} for variadic parameters.
func ConvertToServeMuxParamPath(path string) string {
	return paramPattern.ReplaceAllStringFunc(path, func(m string) string {
		match := paramPattern.FindStringSubmatch(m)
		if match[2] != "" {
			return "{" + match[2] + "...}" // *id -> {id...}
		}
		return "{" + match[1] + "}" // :id -> {id}
	})
}

// headFallbackWriter is a custom http.ResponseWriter that discards the body
// for HEAD requests. It implements the http.ResponseWriter interface.
type headFallbackWriter struct {
	http.ResponseWriter
}

// Write implements http.ResponseWriter interface, but discards the body
// since HEAD requests should not have a body.
func (w headFallbackWriter) Write(b []byte) (int, error) {
	// discard body (HEAD request dont have body)
	return len(b), nil
}

func cleanPrefix(prefix string) string {
	if prefix == "/" || prefix == "" {
		return "/"
	}

	return "/" + strings.Trim(prefix, "/")
}
