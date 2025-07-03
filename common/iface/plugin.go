package iface

import "lokstra/common/permission"

type Plugin interface {
	Register(lic *permission.PermissionLicense)
}
