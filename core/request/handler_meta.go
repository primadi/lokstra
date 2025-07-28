package request

// HandlerMeta represents a named handler.
// Can be a direct function or resolved later by name.
type HandlerMeta struct {
	Name        string
	HandlerFunc HandlerFunc
	Extension   any // Optional extension for the handler, currently used for *RpcServiceMeta
}
