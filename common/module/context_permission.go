package module

import "maps"

type PermissionRequest struct {
	WhitelistGetService []string

	AllowRegisterHandler    bool
	AllowRegisterMiddleware bool
	AllowRegisterService    bool

	ContextSettings map[string]any
}

type PermissionGranted struct {
	whitelistGetService []string

	allowRegisterHandler    bool
	allowRegisterMiddleware bool
	allowRegisterService    bool

	contextSettings map[string]any
}

func newPermissionGranted(req *PermissionRequest) *PermissionGranted {
	return &PermissionGranted{
		whitelistGetService: req.WhitelistGetService,

		allowRegisterHandler:    req.AllowRegisterHandler,
		allowRegisterMiddleware: req.AllowRegisterMiddleware,
		allowRegisterService:    req.AllowRegisterService,

		contextSettings: maps.Clone(req.ContextSettings),
	}
}

func (p *PermissionGranted) IsAllowedGetService(name string) bool {
	// allow full match or prefix match with wildcard
	for _, whitelisted := range p.whitelistGetService {
		if whitelisted == name {
			return true
		}
		lw := len(whitelisted)
		if lw > 0 && whitelisted[lw-1] == '*' {
			prefix := whitelisted[:lw-1]
			if len(name) >= len(prefix) && name[:len(prefix)] == prefix {
				return true
			}
		}
	}
	return false
}

func (p *PermissionGranted) IsAllowedRegisterHandler() bool {
	return p.allowRegisterHandler
}

func (p *PermissionGranted) IsAllowedRegisterMiddleware() bool {
	return p.allowRegisterMiddleware
}

func (p *PermissionGranted) IsAllowedRegisterService() bool {
	return p.allowRegisterService
}

func (p *PermissionGranted) GetContextSettings() map[string]any {
	return p.contextSettings
}
