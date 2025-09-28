package engine

import "net/http"

type RouterEngine interface {
	http.Handler
	Handle(method, path string, h http.Handler)
}
