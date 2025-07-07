package core_service

import (
	"lokstra/common/iface"
	"net/http"
)

const ROUTER_ENGINE_NAME = "lokstra.router_engine"

const HTTPROUTER_ROUTER_ENGINE_NAME = ROUTER_ENGINE_NAME + ".httprouter"
const SERVEMUX_ROUTER_ENGINE_NAME = ROUTER_ENGINE_NAME + ".servemux"
const DEFAULT_ROUTER_ENGINE_NAME = HTTPROUTER_ROUTER_ENGINE_NAME

// RouterEngine defines the interface for a router engine that can handle HTTP methods,
// serve static files, single-page applications (SPA), and reverse proxies.
type RouterEngine interface {
	// HandleMethod registers a handler for a specific HTTP method and path.
	HandleMethod(method iface.HTTPMethod, path string, handler http.Handler)
	GetRouterEngineType() string

	ServeHTTP(w http.ResponseWriter, r *http.Request)
	ServeStatic(prefix string, folder http.Dir)
	ServeSPA(prefix string, indexFile string)
	ServeReverseProxy(prefix string, target string)
}
