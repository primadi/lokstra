package engine

import "net/http"

type RouterEngine interface {
	http.Handler
	Handle(pattern string, h http.Handler)
}
