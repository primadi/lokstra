package router

import (
	"net/http"

	"github.com/primadi/lokstra/core/request"
)

type RequestHandler struct {
	Name        string
	Path        string
	Method      string
	HandlerFunc request.HandlerFunc
	HandlerRaw  http.HandlerFunc
	Middleware  []request.HandlerFunc
}
