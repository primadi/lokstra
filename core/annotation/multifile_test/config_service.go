package main

import "github.com/primadi/lokstra/core/service"

// @RouterService name="config-service", prefix="/api/config"
type ConfigService struct {
	// @Inject "config-repo"
	ConfigRepo *service.Cached[interface{}]
}

// @Route "GET /version"
func (s *ConfigService) GetVersion() (string, error) {
	return "v1.0.0", nil
}

// @Route "POST /reload"
func (s *ConfigService) Reload() error {
	return nil
}

// @Route "GET /settings"
func (s *ConfigService) GetSettings() (map[string]string, error) {
	return map[string]string{"key": "value"}, nil
}
