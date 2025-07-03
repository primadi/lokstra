package permission

import (
	"strings"
)

type PermissionRegistration struct {
	ModuleName string

	AllowRegisterHandler    bool
	AllowRegisterMiddleware bool
	AllowRegisterService    bool

	WhitelistGetHandler  []string
	WhitelistGetServices []string
}

type PermissionLicense struct {
	moduleName string

	allowRegisterHandler    bool
	allowRegisterMiddleware bool
	allowRegisterService    bool

	whitelistGetHandler  []string
	whitelistGetServices []string
}

// NewPermissionLicense creates a new PermissionLicense based on the provided PermissionRegistration.
func NewPermissionLicense(permissionReg *PermissionRegistration) *PermissionLicense {
	// if globalAccessLocked {
	// 	panic("Global access is locked, cannot create new permission")
	// }
	return &PermissionLicense{
		moduleName:              permissionReg.ModuleName,
		allowRegisterHandler:    permissionReg.AllowRegisterHandler,
		allowRegisterMiddleware: permissionReg.AllowRegisterMiddleware,
		allowRegisterService:    permissionReg.AllowRegisterService,
		whitelistGetHandler:     permissionReg.WhitelistGetHandler,
		whitelistGetServices:    permissionReg.WhitelistGetServices,
	}
}

func (p *PermissionLicense) GetModuleName() string {
	return p.moduleName
}

func (p *PermissionLicense) AllowRegisterHandler() bool {
	return p.allowRegisterHandler
}

func (p *PermissionLicense) AllowRegisterMiddleware() bool {
	return p.allowRegisterMiddleware
}

func (p *PermissionLicense) AllowRegisterService() bool {
	return p.allowRegisterService
}

// IsHandlerWhitelisted checks if a handler is whitelisted.
// It supports wildcards at the end of the handler name.
// For example, "user.*" will match "user.create", "user.update", etc.
func (p *PermissionLicense) IsHandlerWhitelisted(handlerName string) bool {
	return isWhiteListed(p.whitelistGetHandler, handlerName)
}

// IsServiceWhitelisted checks if a service is whitelisted.
// It supports wildcards at the end of the service name.
// For example, "service.*" will match "service.user", "service.order", etc.
func (p *PermissionLicense) IsServiceWhitelisted(serviceName string) bool {
	return isWhiteListed(p.whitelistGetServices, serviceName)
}

func isWhiteListed(wlArray []string, data string) bool {
	// If permissions are not locked, everything is allowed
	if !globalAccessLocked {
		return true
	}

	// Check if the data is in the whitelist array
	for _, wlItem := range wlArray {
		// If the item is an exact match
		if wlItem == data {
			return true
		}
		// If the item ends with a wildcard '*', check if the data starts with the prefix
		if strings.HasSuffix(wlItem, "*") {
			prefix := strings.TrimSuffix(wlItem, "*")
			if strings.HasPrefix(data, prefix) {
				return true
			}
		}
	}
	return false
}
