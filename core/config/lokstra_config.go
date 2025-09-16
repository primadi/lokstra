package config

import (
	"fmt"
	"io/fs"
	"os"
	"plugin"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/core/server"
)

// StartAllModules starts all modules defined in the configuration.
func (cfg *LokstraConfig) StartAllModules(regCtx registration.Context) error {
	modules := cfg.Modules
	for _, mod := range modules {
		// 1. Register the module itself if path is provided
		if mod.Path != "" {
			newCtx := regCtx.NewPermissionContextFromConfig(mod.Settings, mod.Permissions)
			if err := newCtx.RegisterCompiledModuleWithFuncName(mod.Path, mod.Entry); err != nil {
				return fmt.Errorf("register module %s: %w", mod.Name, err)
			}
		}

		// 2. Check required services
		for _, serviceName := range mod.RequiredServices {
			if _, err := regCtx.GetService(serviceName); err != nil {
				return fmt.Errorf("module %s requires service %s which is not available: %w", mod.Name, serviceName, err)
			}
		}

		// 3. Create services defined in the module
		for _, serviceConfig := range mod.CreateServices {
			if err := createServiceFromConfig(regCtx, &serviceConfig); err != nil {
				return fmt.Errorf("module %s failed to create service %s: %w", mod.Name, serviceConfig.Name, err)
			}
		}

		// 4. Register service factories from the module
		if mod.Path != "" && len(mod.RegisterServiceFactories) > 0 {
			if err := callModuleMethods(mod.Path, mod.RegisterServiceFactories, regCtx, "service factory"); err != nil {
				return fmt.Errorf("module %s failed to register service factories: %w", mod.Name, err)
			}
		}

		// 5. Register handlers from the module
		if mod.Path != "" && len(mod.RegisterHandlers) > 0 {
			if err := callModuleMethods(mod.Path, mod.RegisterHandlers, regCtx, "handler"); err != nil {
				return fmt.Errorf("module %s failed to register handlers: %w", mod.Name, err)
			}
		}

		// 6. Register middleware from the module
		if mod.Path != "" && len(mod.RegisterMiddleware) > 0 {
			if err := callModuleMethods(mod.Path, mod.RegisterMiddleware, regCtx, "middleware"); err != nil {
				return fmt.Errorf("module %s failed to register middleware: %w", mod.Name, err)
			}
		}
	}
	return nil
}

// StartAllServices starts all services defined in the configuration, respecting dependencies.
func (cfg *LokstraConfig) StartAllServices(regCtx registration.Context) error {
	services := cfg.Services
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

// NewAllApps creates all applications defined in the configuration and adds them to the server.
func (cfg *LokstraConfig) NewAllApps(regCtx registration.Context, server *server.Server) error {
	apps := cfg.Apps
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
			tmpFS := make([]fs.FS, 0, len(static.Folder))
			for _, folder := range static.Folder {
				tmpFS = append(tmpFS, os.DirFS(folder))
			}
			app.MountStatic(static.Prefix, static.Spa, tmpFS...)
		}

		for _, htmx := range ac.MountHtmx {
			tmpFS := make([]fs.FS, 0, len(htmx.Sources))
			for _, folder := range htmx.Sources {
				tmpFS = append(tmpFS, os.DirFS(folder))
			}
			app.MountHtmx(htmx.Prefix, nil, tmpFS...)
		}

		for _, proxy := range ac.MountReverseProxy {
			app.MountReverseProxy(proxy.Prefix, proxy.Target, false)
		}

		for _, group := range ac.Groups {
			buildGroup(regCtx, app, group)
		}

		server.AddApp(app)
	}

	return nil
}

// NewServer creates a new Server instance with the given context and name.
func (cfg *LokstraConfig) NewServer(regCtx registration.Context) (*server.Server, error) {
	server := server.NewServer(regCtx, cfg.Server.Name)
	for k, v := range cfg.Server.Settings {
		server.SetSetting(k, v)
	}

	return server, nil
}

// NewServerFromConfig start all modules, start all services, new all apps, and
// creates a new server instance from the provided configuration.
func (cfg *LokstraConfig) NewServerFromConfig(regCtx registration.Context) (*server.Server, error) {
	return cfg.loadAndStartAll(regCtx, nil)
}

// LoadConfigToServer loads configuration into an existing server instance,
func (cfg *LokstraConfig) LoadConfigToServer(regCtx registration.Context,
	server *server.Server) (*server.Server, error) {
	return cfg.loadAndStartAll(regCtx, server)
}

func (cfg *LokstraConfig) loadAndStartAll(regCtx registration.Context,
	server *server.Server) (*server.Server, error) {
	if server == nil {
		var err error
		server, err = cfg.NewServer(regCtx)
		if err != nil {
			return nil, fmt.Errorf("create server: %w", err)
		}
	} else {
		server.SetSettingsIfAbsent(cfg.Server.Settings)
	}

	// 1. Start Modules
	if err := cfg.StartAllModules(regCtx); err != nil {
		return nil, fmt.Errorf("start modules: %w", err)
	}

	// 2. Start Services
	if err := cfg.StartAllServices(regCtx); err != nil {
		return nil, fmt.Errorf("start services: %w", err)
	}

	// 3. New Apps
	if err := cfg.NewAllApps(regCtx, server); err != nil {
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
		gr.MountHtmx(htmx.Prefix, nil, tmpFS...)
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
