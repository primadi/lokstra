package registry

import (
	"lokstra/common/permission"
	"lokstra/core/request"
	"strings"
)

type HandlerRegister struct {
	Name        string
	HandlerFunc request.HandlerFunc
}

var handlers = make(map[string]*HandlerRegister)

// GetHandler retrieves a registered handler by its name.
// It checks if the handler is whitelisted in the provided permission license.
func GetHandler(name string, lic ...*permission.PermissionLicense) (*HandlerRegister, bool) {
	if permission.GlobalAccessLocked() {
		if len(lic) == 0 || lic[0] == nil || !lic[0].IsHandlerWhitelisted(name) {
			return nil, false
		}
	}

	if !strings.Contains(name, ".") {
		name = "main." + name
	}

	handler, exists := handlers[name]
	return handler, exists
}

// RegisterHandler registers a new handler with the given name.
func RegisterHandler(name string, handler request.HandlerFunc, lic ...*permission.PermissionLicense) {
	if permission.GlobalAccessLocked() {
		if len(lic) == 0 || lic[0] == nil || !strings.HasPrefix(name, lic[0].GetModuleName()+":") {
			panic("cannot register handler after global access is locked or handler is not created in the same module")
		}
	}

	if handler == nil {
		panic("handler cannot be nil")
	}
	if name == "" {
		panic("handler name cannot be empty")
	}

	if !strings.Contains(name, ".") {
		name = "main." + name
	}

	if _, exists := handlers[name]; exists {
		panic("handler with name '" + name + "' already exists")
	}

	info := &HandlerRegister{
		Name:        name,
		HandlerFunc: handler,
	}

	handlers[name] = info
}

// ResetHandlers clears all registered handlers.
// This is useful for testing or reinitializing the registry.
func ResetHandlers() {
	if permission.GlobalAccessLocked() {
		panic("cannot reset handler after global access is locked")
	}

	handlers = make(map[string]*HandlerRegister)
}
