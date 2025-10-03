package lokstra_registry

import (
	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/config"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/core/server"
)

// Applies the configuration to the lokstra_registry
func RegisterConfig(c *config.Config) {
	// Apply general configs
	for _, gc := range c.Configs {
		SetConfig(gc.Name, gc.Value)
	}

	// Apply services
	for _, svc := range c.Services {
		if svc.IsEnabled() {
			RegisterLazyService(svc.Name, svc.Type,
				svc.Config, AllowOverride(true))
		}
	}
	// Apply middlewares
	for _, mw := range c.Middlewares {
		if mw.IsEnabled() {
			RegisterMiddlewareName(mw.Name, mw.Type,
				mw.Config, AllowOverride(true))
		}
	}
	// Apply servers
	for _, srvConfig := range c.Servers {
		apps := make([]*app.App, 0, len(srvConfig.Apps))
		for i, appConfig := range srvConfig.Apps {
			routers := make([]router.Router, 0, len(appConfig.Routers))
			for _, routerName := range appConfig.Routers {
				if r := GetRouter(routerName); r != nil {
					clonedRouter := r.Clone()
					routers = append(routers, clonedRouter)
					RegisterClientRouter(clonedRouter.Name(),
						srvConfig.Name, srvConfig.BaseUrl, appConfig.Addr)
				} else {
					panic("router " + routerName + " not found for app " + appConfig.Name)
				}
			}
			app := app.New(appConfig.GetName(i), appConfig.Addr, routers...)
			apps = append(apps, app)
		}
		srv := server.New(srvConfig.Name, apps...)
		srv.DeploymentID = srvConfig.DeploymentID
		srv.BaseUrl = srvConfig.BaseUrl
		RegisterServer(srvConfig.Name, srv, AllowOverride(true))
	}
}
