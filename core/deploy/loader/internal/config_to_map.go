package internal

import "github.com/primadi/lokstra/core/deploy/schema"

// converts DeployConfig to map[string]any for JSON schema validation
func ConfigToMap(config *schema.DeployConfig) map[string]any {
	result := make(map[string]any)

	if len(config.Configs) > 0 {
		result["configs"] = config.Configs
	}

	if len(config.ServiceDefinitions) > 0 {
		services := make(map[string]any)
		for name, svc := range config.ServiceDefinitions {
			svcMap := map[string]any{
				"type": svc.Type,
			}
			if len(svc.DependsOn) > 0 {
				svcMap["depends-on"] = svc.DependsOn
			}
			if len(svc.Config) > 0 {
				svcMap["config"] = svc.Config
			}
			services[name] = svcMap
		}
		result["service-definitions"] = services
	}

	if len(config.RouterDefinitions) > 0 {
		routers := make(map[string]any)
		for name, rtr := range config.RouterDefinitions {
			rtrMap := make(map[string]any)

			// Inline override fields
			if rtr.PathPrefix != "" {
				rtrMap["path-prefix"] = rtr.PathPrefix
			}
			if len(rtr.Middlewares) > 0 {
				rtrMap["middlewares"] = rtr.Middlewares
			}
			if len(rtr.Hidden) > 0 {
				rtrMap["hidden"] = rtr.Hidden
			}
			if len(rtr.Custom) > 0 {
				rtrMap["custom"] = rtr.Custom
			}

			routers[name] = rtrMap
		}
		result["router-definitions"] = routers // Renamed from "routers"
	}

	if len(config.Deployments) > 0 {
		deployments := make(map[string]any)
		for name, dep := range config.Deployments {
			depMap := make(map[string]any)

			if len(dep.ConfigOverrides) > 0 {
				depMap["config-overrides"] = dep.ConfigOverrides
			}

			if len(dep.Servers) > 0 {
				servers := make(map[string]any)
				for srvName, srv := range dep.Servers {
					srvMap := map[string]any{
						"base-url": srv.BaseURL,
					}

					if len(srv.Apps) > 0 {
						apps := make([]any, len(srv.Apps))
						for i, app := range srv.Apps {
							appMap := map[string]any{
								"addr": app.Addr,
							}
							if len(app.Routers) > 0 {
								appMap["routers"] = app.Routers
							}
							if len(app.PublishedServices) > 0 {
								appMap["published-services"] = app.PublishedServices
							}
							apps[i] = appMap
						}
						srvMap["apps"] = apps
					}

					servers[srvName] = srvMap
				}
				depMap["servers"] = servers
			}

			deployments[name] = depMap
		}
		result["deployments"] = deployments
	}

	return result
}
