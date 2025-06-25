package iface

// App represents an application instance served on a specific port.
type App interface {
	Name() string
	GetServer() Server
	HasSetting
}
