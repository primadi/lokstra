package meta

import (
	"github.com/primadi/lokstra/core/request"
)

// HandlerMeta represents a named handler.
// Can be a direct function or resolved later by name.
type HandlerMeta struct {
	Name        string
	HandlerFunc request.HandlerFunc
}
