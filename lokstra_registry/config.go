package lokstra_registry

import (
	"fmt"

	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/config"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/core/server"
)

// processServices handles both simple and layered service configurations
func processServices(services *config.ServicesConfig) error {
	if services.IsSimple() {
		// Simple mode: process all services in array order
		for _, svc := range services.Simple {
			if err := registerService(svc, nil); err != nil {
				return err
			}
		}
		return nil
	}

	// Layered mode: validate and process in layer order
	if err := config.ValidateLayeredServices(services); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Build dependency map for lazy injection
	depMap := buildDependencyMap(services)

	// Process services layer by layer
	for _, layerName := range services.Order {
		layerServices := services.Layered[layerName]
		for _, svc := range layerServices {
			if err := registerService(svc, depMap); err != nil {
				return fmt.Errorf("layer %s, service %s: %w", layerName, svc.Name, err)
			}
		}
	}

	return nil
}

// registerService registers a single service with optional dependency injection
func registerService(svc *config.Service, depMap map[string]*config.GenericLazyService) error {
	if !svc.IsEnabled() {
		return nil
	}

	svcType := svc.Type
	if svcType == "" {
		svcType = svc.Name
	}

	// Inject lazy dependencies into config
	cfg := svc.Config
	if cfg == nil {
		cfg = make(map[string]interface{})
	}

	// Clone config to avoid mutating original
	injectedCfg := make(map[string]interface{})
	for k, v := range cfg {
		injectedCfg[k] = v
	}

	// Replace service name references with GenericLazyService
	if depMap != nil {
		for _, depName := range svc.DependsOn {
			if lazySvc, ok := depMap[depName]; ok {
				// Find config keys that reference this dependency
				for k, v := range injectedCfg {
					if strVal, ok := v.(string); ok && strVal == depName {
						// Replace string reference with lazy service
						injectedCfg[k] = lazySvc
					}
				}
			}
		}
	}

	RegisterLazyService(svc.Name, svcType, injectedCfg, AllowOverride(true))
	return nil
}

// buildDependencyMap creates a map of service name -> GenericLazyService
func buildDependencyMap(services *config.ServicesConfig) map[string]*config.GenericLazyService {
	depMap := make(map[string]*config.GenericLazyService)

	// Add simple services to dependency map
	for _, svc := range services.Simple {
		if svc.IsEnabled() {
			depMap[svc.Name] = config.NewGenericLazyService(svc.Name)
		}
	}

	// Add layered services to dependency map as they're processed
	for _, layerName := range services.Order {
		for _, svc := range services.Layered[layerName] {
			if svc.IsEnabled() {
				depMap[svc.Name] = config.NewGenericLazyService(svc.Name)
			}
		}
	}

	return depMap
}

// Applies the configuration to the lokstra_registry
func RegisterConfig(c *config.Config) {
	// apply general configs
	for _, gc := range c.Configs {
		SetConfig(gc.Name, gc.Value)
	}

	// Apply services
	if err := processServices(&c.Services); err != nil {
		panic("failed to process services: " + err.Error())
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
			routers := make([]router.Router, 0, len(appConfig.Routers)+len(appConfig.RoutersWithPrefix))

			// Add routers without prefix
			for _, routerName := range appConfig.Routers {
				r := GetRouter(routerName)
				if r == nil {
					panic("router " + routerName + " not found for app " + appConfig.Name)
				}
				routers = append(routers, r)
			}

			// Add routers with custom prefix
			for _, rwp := range appConfig.RoutersWithPrefix {
				r := GetRouter(rwp.Name)
				if r == nil {
					panic("router " + rwp.Name + " not found for app " + appConfig.Name)
				}
				// Clone router to avoid side effects when used in multiple places
				clonedRouter := r.Clone()
				// Combine app prefix from config with router's existing prefix
				combinedPrefix := rwp.Prefix + clonedRouter.PathPrefix()
				clonedRouter.SetPathPrefix(combinedPrefix)
				routers = append(routers, clonedRouter)
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
