package iface

// Service defines the minimal interface all services must implement.
type Service interface {
	Name() string // Unique instance name
	Type() string // Service type identifier
	HasSetting    // Read-only settings access
}
