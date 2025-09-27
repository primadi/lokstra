package middleware

import (
	"github.com/primadi/lokstra/core/request"
)

type Handler struct {
	Name        string
	HandlerFunc request.HandlerFunc
	Params      map[string]any
}
