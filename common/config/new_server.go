package config

import (
	"fmt"
	"lokstra/common/meta"
	"lokstra/common/module"
	"lokstra/common/utils"
	"lokstra/core/app"
	"lokstra/core/router"
	"lokstra/core/server"
	"net/http"
)

func NewServerFromConfig(ctx *module.RegistrationContextImpl, cfg *LokstraConfig) (*server.Server, error) {
	if err := startAllServices(ctx, cfg.Services); err != nil {
		return nil, fmt.Errorf("start services: %w", err)
	}

	if err := startModulesFromConfig(ctx, cfg.Modules); err != nil {
		return nil, fmt.Errorf("start modules: %w", err)
	}

	server := server.NewServer(ctx, cfg.Server.Name)
	for k, v := range cfg.Server.Settings {
		server.SetSetting(k, v)
	}

	if err := newAppsFromConfig(ctx, server, cfg.Apps); err != nil {
		return nil, fmt.Errorf("start apps: %w", err)
	}

	return server, nil
}

func startAllServices(ctx *module.RegistrationContextImpl, services []*ServiceConfig) error {
	for _, svc := range services {
		if _, err := ctx.CreateService(svc.Type, svc.Name, svc.Config); err != nil {
			return fmt.Errorf("register service %s: %w", svc.Name, err)
		}
	}
	return nil
}

func startModulesFromConfig(ctx *module.RegistrationContextImpl, modules []*ModuleConfig) error {
	for _, mod := range modules {
		if mod.Path != "" || mod.Entry != "" {
			newCtx := ctx.NewPermissionContextFromConfig(mod.Settings, mod.Permissions)
			if err := newCtx.RegisterPluginModuleWithEntry(mod.Name, mod.Path, mod.Entry); err != nil {
				return fmt.Errorf("register module %s: %w", mod.Name, err)
			}
		}
	}
	return nil
}

func newAppsFromConfig(ctx *module.RegistrationContextImpl, server *server.Server, apps []*AppConfig) error {
	for _, ac := range apps {
		app := app.NewAppCustom(ctx, ac.Name, ac.Address,
			ac.ListenerType, ac.RouterEngineType, ac.Settings)

		for _, mw := range ac.Middleware {
			if mw.Enabled {
				app.Use(meta.NamedMiddleware(mw.Name, mw.Config))
			}
		}

		for _, route := range ac.Routes {
			mw := make([]*meta.MiddlewareExecution, 0, len(route.Middleware))
			for _, m := range route.Middleware {
				if m.Enabled {
					mw = append(mw, meta.NamedMiddleware(m.Name, m.Config))
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
			buildGroup(ctx, app, group)
		}

		server.AddApp(app)
	}

	return nil
}

func buildGroup(ctx *module.RegistrationContextImpl, parent router.Router, group GroupConfig) {
	gr := parent.Group(group.Prefix)

	for _, mw := range group.Middleware {
		if mw.Enabled {
			gr.Use(meta.NamedMiddleware(mw.Name, mw.Config))
		}
	}

	for _, route := range group.Routes {
		mw := make([]*meta.MiddlewareExecution, 0, len(route.Middleware))
		for _, m := range route.Middleware {
			if m.Enabled {
				mw = append(mw, meta.NamedMiddleware(m.Name, m.Config))
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

	for _, spa := range group.MountSPA {
		gr.MountSPA(spa.Prefix, spa.FallbackFile)
	}

	for _, proxy := range group.MountReverseProxy {
		gr.MountReverseProxy(proxy.Prefix, proxy.Target)
	}

	for _, subGroup := range group.Groups {
		buildGroup(ctx, gr, subGroup)
	}

}
