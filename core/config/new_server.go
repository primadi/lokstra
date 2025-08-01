package config

import (
	"fmt"
	"net/http"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/iface"
	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/core/server"
)

func NewServerFromConfig(regCtx iface.RegistrationContext, cfg *LokstraConfig) (*server.Server, error) {
	// 1. Start Modules
	if err := startModulesFromConfig(regCtx, cfg.Modules); err != nil {
		return nil, fmt.Errorf("start modules: %w", err)
	}

	// 2. Start Services
	if err := startServicesInOrder(regCtx, cfg.Services); err != nil {
		return nil, fmt.Errorf("start services: %w", err)
	}

	// 3. New Server
	server := server.NewServer(regCtx, cfg.Server.Name)
	for k, v := range cfg.Server.Settings {
		server.SetSetting(k, v)
	}

	// 4. New Apps
	if err := newAppsFromConfig(regCtx, server, cfg.Apps); err != nil {
		return nil, fmt.Errorf("start apps: %w", err)
	}

	return server, nil
}

func startServicesInOrder(regCtx iface.RegistrationContext,
	services []*ServiceConfig) error {
	depMap := map[string][]string{}
	inDegree := map[string]int{}
	serviceMap := map[string]*ServiceConfig{}

	for _, svc := range services {
		serviceMap[svc.Name] = svc
		for _, dep := range svc.DependsOn {
			depMap[dep] = append(depMap[dep], svc.Name)
			inDegree[svc.Name]++
		}
	}

	// Queue service without dependencies
	var queue []string
	for _, svc := range services {
		if inDegree[svc.Name] == 0 {
			queue = append(queue, svc.Name)
		}
	}

	var ordered []string
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		ordered = append(ordered, current)

		for _, dependent := range depMap[current] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	if len(ordered) != len(services) {
		return fmt.Errorf("cyclic dependency detected in services")
	}

	// Start services in order
	for _, name := range ordered {
		svc := serviceMap[name]
		if _, err := regCtx.CreateService(svc.Type, svc.Name, svc.Config); err != nil {
			return fmt.Errorf("start service %s: %w", name, err)
		}
	}

	return nil
}

func startModulesFromConfig(regCtx iface.RegistrationContext, modules []*ModuleConfig) error {
	for _, mod := range modules {
		if mod.Path != "" || mod.Entry != "" {
			newCtx := regCtx.NewPermissionContextFromConfig(mod.Settings, mod.Permissions)
			if err := newCtx.RegisterCompiledModuleWithFuncName(mod.Name, mod.Path, mod.Entry); err != nil {
				return fmt.Errorf("register module %s: %w", mod.Name, err)
			}
		}
	}
	return nil
}

func newAppsFromConfig(regCtx iface.RegistrationContext, server *server.Server, apps []*AppConfig) error {
	for _, ac := range apps {
		app := app.NewAppCustom(regCtx, ac.Name, ac.Address,
			ac.ListenerType, ac.RouterEngineType, ac.Settings)

		for _, mw := range ac.Middleware {
			if mw.Enabled {
				app.Use(midware.Named(mw.Name, mw.Config))
			}
		}

		for _, route := range ac.Routes {
			mw := make([]*midware.Execution, 0, len(route.Middleware))
			for _, m := range route.Middleware {
				if m.Enabled {
					mw = append(mw, midware.Named(m.Name, m.Config))
				}
			}
			if route.OverrideMiddleware {
				app.HandleOverrideMiddleware(route.Method, route.Path, route.Handler, utils.ToAnySlice(mw)...)
			} else {
				app.Handle(route.Method, route.Path, route.Handler, utils.ToAnySlice(mw)...)
			}
		}

		for _, static := range ac.MountStatic {
			app.MountStatic(static.Prefix, http.Dir(static.Folder))
		}

		for _, spa := range ac.MountSPA {
			app.MountSPA(spa.Prefix, spa.FallbackFile)
		}

		for _, proxy := range ac.MountReverseProxy {
			app.MountReverseProxy(proxy.Prefix, proxy.Target)
		}

		for _, group := range ac.Groups {
			buildGroup(regCtx, app, group)
		}

		server.AddApp(app)
	}

	return nil
}

func buildGroup(regCtx iface.RegistrationContext, parent router.Router, group GroupConfig) {
	gr := parent.Group(group.Prefix)

	for _, mw := range group.Middleware {
		if mw.Enabled {
			gr.Use(midware.Named(mw.Name, mw.Config))
		}
	}

	for _, route := range group.Routes {
		mw := make([]*midware.Execution, 0, len(route.Middleware))
		for _, m := range route.Middleware {
			if m.Enabled {
				mw = append(mw, midware.Named(m.Name, m.Config))
			}
		}
		if route.OverrideMiddleware {
			gr.HandleOverrideMiddleware(route.Method, route.Path, route.Handler, utils.ToAnySlice(mw)...)
		} else {
			gr.Handle(route.Method, route.Path, route.Handler, utils.ToAnySlice(mw)...)
		}
	}

	for _, static := range group.MountStatic {
		gr.MountStatic(static.Prefix, http.Dir(static.Folder))
	}

	for _, spa := range group.MountSpa {
		gr.MountSPA(spa.Prefix, spa.FallbackFile)
	}

	for _, proxy := range group.MountReverseProxy {
		gr.MountReverseProxy(proxy.Prefix, proxy.Target)
	}

	for _, subGroup := range group.Groups {
		buildGroup(regCtx, gr, subGroup)
	}

}
