package router_engine

import (
	"net/http"
	"regexp"
	"strings"
)

const HTTPROUTER_ROUTER_ENGINE_NAME = "coreservice.httprouter"
const SERVEMUX_ROUTER_ENGINE_NAME = "coreservice.servemux"

// paramPatternServeMux is a regular expression that matches path parameters in the format
// :param or *param. It captures both types of parameters:
var paramPatternServeMux = regexp.MustCompile(`:([a-zA-Z_][a-zA-Z0-9_]*)|\*([a-zA-Z_][a-zA-Z0-9_]*)`)

// ConvertToServeMuxParamPath converts a path with parameters (e.g., :id or *id)
// to a format suitable for http.ServeMux, which uses {param} syntax.
// It replaces :id with {id} and *id with {id...} to indicate a variadic parameter.
// For example, "/users/:id" becomes "/users/{id}" and "/files/*id" becomes "/files/{id...}".
// This is necessary because http.ServeMux does not support the :param or *param syntax.
// Instead, it uses {param} for single parameters and {param...} for variadic parameters.
func ConvertToServeMuxParamPath(path string) string {
	// Handle exact root match vs catch-all
	if path == "" {
		return "/" // Convert empty path to "/" for exact match
	}

	return paramPatternServeMux.ReplaceAllStringFunc(path, func(m string) string {
		match := paramPatternServeMux.FindStringSubmatch(m)
		if match[2] != "" {
			return "{" + match[2] + "...}" // *id -> {id...}
		}
		return "{" + match[1] + "}" // :id -> {id}
	})
}

var paramPatternHttpRouter = regexp.MustCompile(`\{([a-zA-Z_][a-zA-Z0-9_]*)\}|\{([a-zA-Z_][a-zA-Z0-9_]*)\.\.\.\}`)

// ConvertToHttpRouterParamPath converts a path with parameters (e.g., {id} or {id...})
// to a format suitable for http.Router, which uses :param and *param syntax.
func ConvertToHttpRouterParamPath(path string) string {
	converted := paramPatternHttpRouter.ReplaceAllStringFunc(path, func(m string) string {
		match := paramPatternHttpRouter.FindStringSubmatch(m)
		if match[2] != "" {
			return "*" + match[2] // {id...} -> *id
		}
		return ":" + match[1] // {id} -> :id
	})
	if path != "/" && strings.HasSuffix(path, "/") {
		converted += "*filepath"
	}

	if strings.Count(converted, "*") > 1 {
		panic("only one wildcard parameter is allowed in path: " + path)
	}

	// wildcard parameter must be at the end of the path
	parts := strings.Split(converted, "*")
	if len(parts) == 2 && strings.Contains(parts[1], "/") {
		panic("wildcard parameter must be at the end of the path: " + path)
	}

	return converted
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
	// Handle exact root match (empty string should only match root)
	if prefix == "" {
		return ""
	}

	// Handle catch-all root pattern
	if prefix == "/" {
		return "/"
	}

	return "/" + strings.Trim(prefix, "/")
}
