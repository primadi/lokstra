package config

import (
	"fmt"
	"net/http"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/core/server"
)

func NewServerFromConfig(regCtx registration.Context, cfg *LokstraConfig) (*server.Server, error) {
	if err := startAllServices(regCtx, cfg.Services); err != nil {
		return nil, fmt.Errorf("start services: %w", err)
	}

	if err := startModulesFromConfig(regCtx, cfg.Modules); err != nil {
		return nil, fmt.Errorf("start modules: %w", err)
	}

	server := server.NewServer(regCtx, cfg.Server.Name)
	for k, v := range cfg.Server.Settings {
		server.SetSetting(k, v)
	}

	if err := newAppsFromConfig(regCtx, server, cfg.Apps); err != nil {
		return nil, fmt.Errorf("start apps: %w", err)
	}

	return server, nil
}

func startAllServices(regCtx registration.Context, services []*ServiceConfig) error {
	for _, svc := range services {
		if _, err := regCtx.CreateService(svc.Type, svc.Name, svc.Config); err != nil {
			return fmt.Errorf("register service %s: %w", svc.Name, err)
		}
	}
	return nil
}

func startModulesFromConfig(regCtx registration.Context, modules []*ModuleConfig) error {
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

func newAppsFromConfig(regCtx registration.Context, server *server.Server, apps []*AppConfig) error {
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

func buildGroup(regCtx registration.Context, parent router.Router, group GroupConfig) {
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
