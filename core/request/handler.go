package request

import "net/http"

type HandlerFunc func(c *Context) error

// ServeHTTP implements http.Handler.
func (h HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := NewContext(w, r, []HandlerFunc{h})
	c.FinalizeResponse(c.executeHandler())
}

type Handler struct {
	handlers []HandlerFunc
}

func NewHandler(h HandlerFunc, mw ...HandlerFunc) *Handler {
	return &Handler{
		handlers: append(mw, h),
	}
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := NewContext(w, r, h.handlers)
	c.FinalizeResponse(c.executeHandler())
}

var _ http.Handler = (*Handler)(nil)
var _ http.Handler = (*HandlerFunc)(nil)
