package config

import "fmt"

func normalizeMiddlewareConfig(input any) ([]MiddlewareConfig, error) {
	var result []MiddlewareConfig

	switch items := input.(type) {
	case []any:
		for _, item := range items {
			switch v := item.(type) {
			case string:
				result = append(result, MiddlewareConfig{
					Name:    v,
					Enabled: true,
				})
			case map[string]any:
				name, ok := v["name"].(string)
				if !ok || name == "" {
					return nil, fmt.Errorf("middleware entry missing 'name': %+v", v)
				}

				cfg := MiddlewareConfig{
					Name:    name,
					Enabled: true, // default
				}

				if enabled, ok := v["enabled"].(bool); ok {
					cfg.Enabled = enabled
				}

				if config, ok := v["config"].(map[string]any); ok {
					cfg.Config = config
				}

				result = append(result, cfg)
			default:
				return nil, fmt.Errorf("invalid middleware entry type: %T", v)
			}
		}
	default:
		return nil, fmt.Errorf("middleware config must be a list")
	}

	return result, nil
}
