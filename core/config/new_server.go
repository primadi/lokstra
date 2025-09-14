package config

import (
	"fmt"
	"io/fs"
	"os"
	"plugin"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/core/server"
)

func LoadAllAndNewServer(regCtx registration.Context, cfg *LokstraConfig) (*server.Server, error) {
	// 1. Start Modules
	if err := cfg.StartModules(regCtx); err != nil {
		return nil, fmt.Errorf("start modules: %w", err)
	}

	// 2. Start Services
	if err := cfg.StartServices(regCtx); err != nil {
		return nil, fmt.Errorf("start services: %w", err)
	}

	// 3. New Server
	var server *server.Server
	var err error
	if server, err = cfg.NewServer(regCtx); err != nil {
		return nil, fmt.Errorf("create server: %w", err)
	}

	// 4. New Apps
	if err := cfg.NewApps(regCtx, server); err != nil {
		return nil, fmt.Errorf("start apps: %w", err)
	}

	return server, nil
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
		tmpFS := make([]fs.FS, 0, len(static.Folder))
		for _, folder := range static.Folder {
			tmpFS = append(tmpFS, os.DirFS(folder))
		}
		gr.MountStatic(static.Prefix, static.Spa, tmpFS...)
	}

	for _, htmx := range group.MountHtmx {
		tmpFS := make([]fs.FS, 0, len(htmx.Sources))
		for _, folder := range htmx.Sources {
			tmpFS = append(tmpFS, os.DirFS(folder))
		}
		gr.MountHtmx(htmx.Prefix, tmpFS...)
	}

	for _, proxy := range group.MountReverseProxy {
		gr.MountReverseProxy(proxy.Prefix, proxy.Target, false)
	}

	for _, subGroup := range group.Groups {
		buildGroup(regCtx, gr, subGroup)
	}

}

// createServiceFromConfig creates a service instance from ServiceConfig
func createServiceFromConfig(regCtx registration.Context, serviceConfig *ServiceConfig) error {
	_, err := regCtx.CreateService(serviceConfig.Type, serviceConfig.Name, serviceConfig.Config)
	if err != nil {
		return fmt.Errorf("failed to create service %s of type %s: %w", serviceConfig.Name, serviceConfig.Type, err)
	}
	return nil
}

// callModuleMethods calls specified methods from a plugin module
func callModuleMethods(pluginPath string, methodNames []string, regCtx registration.Context, methodType string) error {
	if len(methodNames) == 0 {
		return nil
	}

	// Open the plugin
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to open plugin %s: %w", pluginPath, err)
	}

	// Call each method
	for _, methodName := range methodNames {
		sym, err := p.Lookup(methodName)
		if err != nil {
			return fmt.Errorf("method %s not found in plugin %s: %w", methodName, pluginPath, err)
		}

		// Try different function signatures based on method type
		switch methodType {
		case "service factory":
			if factoryFunc, ok := sym.(func(registration.Context) error); ok {
				if err := factoryFunc(regCtx); err != nil {
					return fmt.Errorf("failed to execute service factory method %s: %w", methodName, err)
				}
			} else {
				return fmt.Errorf("method %s has invalid signature for service factory (expected: func(registration.RegistrationContext) error)", methodName)
			}

		case "handler":
			if handlerFunc, ok := sym.(func(registration.Context) error); ok {
				if err := handlerFunc(regCtx); err != nil {
					return fmt.Errorf("failed to execute handler method %s: %w", methodName, err)
				}
			} else {
				return fmt.Errorf("method %s has invalid signature for handler (expected: func(registration.RegistrationContext) error)", methodName)
			}

		case "middleware":
			if middlewareFunc, ok := sym.(func(registration.Context) error); ok {
				if err := middlewareFunc(regCtx); err != nil {
					return fmt.Errorf("failed to execute middleware method %s: %w", methodName, err)
				}
			} else {
				return fmt.Errorf("method %s has invalid signature for middleware (expected: func(registration.RegistrationContext) error)", methodName)
			}

		default:
			return fmt.Errorf("unknown method type: %s", methodType)
		}
	}

	return nil
}
