package chi

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/router/engine"
)

// ChiRouter wraps go-chi router to implement RouterEngine interface
// Converts Go 1.22+ patterns like "/api/{path...}" to Chi patterns like "/api/*"
//
// Path parameters can be accessed using standard Go 1.22+ r.PathValue():
//
//	func handler(w http.ResponseWriter, r *http.Request) {
//	    id := r.PathValue("id")  // for route "/users/{id}"
//	}
type ChiRouter struct {
	mux          *chi.Mux
	allowMethods map[string]string // path -> pre-computed Allow header for OPTIONS
}

// NewChiRouter creates a new ChiRouter
func NewChiRouter() engine.RouterEngine {
	return &ChiRouter{
		mux:          chi.NewRouter(),
		allowMethods: make(map[string]string),
	}
}

func (c *ChiRouter) Handle(pattern string, h http.Handler) {
	method, path := parseMethodPath(pattern)

	// Convert Go 1.22+ wildcard patterns to Chi patterns
	chiPath := convertToChiPattern(path)

	// Wrap handler to support PathValue compatibility
	wrappedHandler := c.wrapHandlerForPathValue(h)

	if method == "ANY" {
		// Chi doesn't have "ANY", so register for all common methods
		methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
		for _, m := range methods {
			c.mux.Method(m, chiPath, wrappedHandler)
		}
		// Update Allow methods for this path
		c.addToAllowMethods(path, methods...)
		c.addToAllowMethods(path, "HEAD", "OPTIONS")

		// Auto-register HEAD and OPTIONS
		c.registerHeadAndOptions(path, chiPath)
	} else {
		c.mux.Method(method, chiPath, wrappedHandler)

		// Add method to Allow methods for this path
		c.addToAllowMethods(path, method)
		if method == "GET" {
			// Auto-register HEAD
			c.mux.Head(chiPath, func(w http.ResponseWriter, r *http.Request) {
				// Chi will call GET handler and discard body automatically
				wrappedHandler.ServeHTTP(w, r)
			})
			c.addToAllowMethods(path, "HEAD")
		}
		c.addToAllowMethods(path, "OPTIONS")

		// Auto-register OPTIONS
		c.registerOptions(path, chiPath)
	}
}

func (c *ChiRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.mux.ServeHTTP(w, r)
}

// wrapHandlerForPathValue wraps a handler to make Chi path parameters available via r.PathValue()
func (c *ChiRouter) wrapHandlerForPathValue(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get path parameters from Chi context
		rctx := chi.RouteContext(r.Context())
		if rctx != nil {
			// Set path values using Go 1.22+ SetPathValue so r.PathValue() works
			for i, key := range rctx.URLParams.Keys {
				r.SetPathValue(key, rctx.URLParams.Values[i])
			}
		}

		// Call original handler
		h.ServeHTTP(w, r)
	})
}

// addToAllowMethods adds methods to the Allow header for a path
func (c *ChiRouter) addToAllowMethods(path string, newMethods ...string) {
	currentAllow := c.allowMethods[path]
	var currentMethods []string
	if currentAllow != "" {
		currentMethods = strings.Split(currentAllow, ", ")
	}

	// Add new methods
	allMethods := utils.NewSliceAndAppend(currentMethods, newMethods...)

	// Remove duplicates and update
	uniqueMethods := removeDuplicates(allMethods)
	c.allowMethods[path] = strings.Join(uniqueMethods, ", ")
}

// registerHeadAndOptions registers HEAD and OPTIONS for ANY routes
func (c *ChiRouter) registerHeadAndOptions(path, chiPath string) {
	// Register HEAD (Chi will auto-handle by calling GET and discarding body)
	c.mux.Head(chiPath, func(w http.ResponseWriter, r *http.Request) {
		// Chi will automatically call the GET handler and discard body
		w.WriteHeader(http.StatusOK)
	})

	// Register OPTIONS
	c.registerOptions(path, chiPath)
}

// registerOptions registers OPTIONS handler that returns pre-computed Allow header
func (c *ChiRouter) registerOptions(path, chiPath string) {
	c.mux.Options(chiPath, func(w http.ResponseWriter, r *http.Request) {
		if allowHeader, exists := c.allowMethods[path]; exists {
			w.Header().Set("Allow", allowHeader)
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

// removeDuplicates removes duplicate strings from slice
func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	return result
}

// parseMethodPath parses pattern like "GET /path" into method and path
func parseMethodPath(pattern string) (method, path string) {
	parts := strings.SplitN(pattern, " ", 2)
	if len(parts) == 2 {
		switch parts[1] {
		case "":
			return parts[0], "/"
		case "/":
			return parts[0], "/*"
		default:
			return parts[0], parts[1]
		}
	}
	return "ANY", pattern
}

// converts Go 1.22+ patterns to Chi patterns
// "/api/{path...}" -> "/api/*"
// "/users/:id" -> "/users/{id}"
func convertToChiPattern(path string) string {
	// Convert {path...} wildcard to Chi's * wildcard
	if before, found := strings.CutSuffix(path, "/{path...}"); found {
		path = before + "/*"
	}

	// Convert :param to {param}
	if strings.Contains(path, ":") {
		parts := strings.Split(path, "/")
		for i := range parts {
			if prefix, found := strings.CutPrefix(parts[i], ":"); found {
				parts[i] = "{" + prefix + "}"
			}
		}
		return strings.Join(parts, "/")
	}

	return path
}

var _ engine.RouterEngine = (*ChiRouter)(nil)
