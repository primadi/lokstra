package router

import (
	"fmt"

	"github.com/primadi/lokstra/core/route"
)

func MergeYamlRoutes(yamlRoutes []route.Route, codeRoutes []route.Route) ([]route.Route, error) {
	merged := make([]route.Route, 0, len(codeRoutes))

	// index YAML routes by name
	yamlMap := make(map[string]route.Route)
	for _, yr := range yamlRoutes {
		yamlMap[yr.Name] = yr
	}

	// track matched routes
	matched := make(map[string]bool)

	for _, cr := range codeRoutes {
		yr, ok := yamlMap[cr.Name]
		if ok {
			// mark matched
			matched[cr.Name] = true

			// merge: prefer YAML for Path & Middleware
			if yr.Path != "" {
				cr.Path = yr.Path
			}
			if yr.Middleware != nil {
				cr.Middleware = yr.Middleware
			}
			if yr.Description != "" {
				cr.Description = yr.Description
			}
		}

		// Code route must have a handler
		if cr.Handler == nil {
			return nil, fmt.Errorf("route [%s] missing handler in code", cr.Name)
		}

		merged = append(merged, cr)
	}

	// cek ada yaml route yang tidak match di code
	for _, yr := range yamlRoutes {
		if !matched[yr.Name] {
			return nil, fmt.Errorf("invalid route [%s] defined in YAML but not in code", yr.Name)
		}
	}

	return merged, nil
}
