package standardservices

import "github.com/primadi/lokstra/core/iface"

func RegisterAll(regCtx iface.RegistrationContext) {
	RegisterAllHTTPRouters(regCtx)
	RegisterAllHTTPListeners(regCtx)
	RegisterAllAuthFlow(regCtx)
}
