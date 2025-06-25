package iface

// Server provides access to server-wide information.
type Server interface {
	Name() string
	HasSetting
}
