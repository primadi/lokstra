package request

import (
	"net/http"
)

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
	// FIX: Create a NEW slice to prevent aliasing when append doesn't reallocate
	// This happens when len(mw) < cap(mw), causing multiple handlers to share
	// the same underlying array and overwriting each other
	handlers := make([]HandlerFunc, len(mw)+1)
	copy(handlers, mw)
	handlers[len(mw)] = h

	return &Handler{
		handlers: handlers,
	}
} // ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := NewContext(w, r, h.handlers)
	c.FinalizeResponse(c.executeHandler())
}

var _ http.Handler = (*Handler)(nil)
var _ http.Handler = (*HandlerFunc)(nil)
