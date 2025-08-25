package defaults

import "github.com/primadi/lokstra/core/registration"

func RegisterAll(regCtx registration.Context) {
	RegisterAllHTTPRouters(regCtx)
	RegisterAllHTTPListeners(regCtx)
	RegisterAllAuthFlow(regCtx)
	RegisterAllServices(regCtx)
	RegisterAllMiddleware(regCtx)
}
