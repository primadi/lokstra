package lokstra_registry

import (
	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/config"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/core/server"
)

// Applies the configuration to the lokstra_registry
func RegisterConfig(c *config.Config) {
	// apply general configs
	for _, gc := range c.Configs {
		SetConfig(gc.Name, gc.Value)
	}

	// Apply services
	for _, svc := range c.Services {
		if svc.IsEnabled() {
			svcType := svc.Type
			if svcType == "" {
				svcType = svc.Name
			}
			RegisterLazyService(svc.Name, svcType,
				svc.Config, AllowOverride(true))
		}
	}
	// Apply middlewares
	for _, mw := range c.Middlewares {
		if mw.IsEnabled() {
			mwType := mw.Type
			if mwType == "" {
				mwType = mw.Name
			}
			RegisterMiddlewareName(mw.Name, mwType,
				mw.Config, AllowOverride(true))
		}
	}
	// Apply servers
	for _, srvConfig := range c.Servers {
		apps := make([]*app.App, 0, len(srvConfig.Apps))
		for i, appConfig := range srvConfig.Apps {
			routers := make([]router.Router, 0, len(appConfig.Routers))
			for _, routerName := range appConfig.Routers {
				r := GetRouter(routerName)
				if r == nil {
					panic("router " + routerName + " not found for app " + appConfig.Name)
				}
				routers = append(routers, r)
			}
			a := app.New(appConfig.GetName(i), appConfig.Addr, routers...)
			r := a.GetRouter()
			for r != nil {
				RegisterClientRouter(r.Name(), srvConfig.Name,
					srvConfig.BaseUrl, appConfig.Addr, 0)
				r = r.GetNextChain()
			}
			apps = append(apps, a)
		}
		srv := server.New(srvConfig.Name, apps...)
		srv.DeploymentID = srvConfig.DeploymentID
		srv.BaseUrl = srvConfig.BaseUrl
		RegisterServer(srvConfig.Name, srv, AllowOverride(true))
	}
}
