package ui_renderer

import (
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/serviceapi"
)

const MODULE_NAME = "lokstra.ui_renderer"

type module struct{}

// serviceWrapper wraps UIRenderer to implement service.Service
type serviceWrapper struct {
	uiRenderer serviceapi.UIRenderer
}

func (sw *serviceWrapper) Start() error {
	return nil
}

func (sw *serviceWrapper) Stop() error {
	return nil
}

func (sw *serviceWrapper) Name() string {
	return MODULE_NAME
}

// Name implements registration.Module.
func (m *module) Name() string {
	return MODULE_NAME
}

// Register implements registration.Module.
func (m *module) Register(regCtx registration.Context) error {
	factory := func(config any) (service.Service, error) {
		uiConfig := &Config{
			TemplateDir:    "services/ui_renderer/templates",
			Theme:          "light",
			CacheTemplates: true,
		}

		switch v := config.(type) {
		case map[string]any:
			// Parse template directory
			if templateDir := utils.GetValueFromMap(v, "template_dir", "services/ui_renderer/templates"); templateDir != "" {
				uiConfig.TemplateDir = templateDir
			}

			// Parse theme
			if theme := utils.GetValueFromMap(v, "theme", "light"); theme != "" {
				uiConfig.Theme = theme
			}

			// Parse cache setting
			uiConfig.CacheTemplates = utils.GetValueFromMap(v, "cache_templates", true)

		case string:
			// Legacy support: just template directory as string
			if v != "" {
				uiConfig.TemplateDir = v
			}
		}

		// Create a wrapper that implements service.Service
		uiRenderer := NewService(uiConfig)
		return &serviceWrapper{uiRenderer: uiRenderer}, nil
	}

	regCtx.RegisterServiceFactory(m.Name(), factory)
	return nil
}

// Description implements service.Module.
func (m *module) Description() string {
	return "UI Renderer Service for Lokstra - Enables declarative UI development using YAML configuration with Preline, Tailwind, HTMX, and Alpine.js"
}

// GetModule returns the UI renderer module with serviceType "lokstra.ui_renderer".
func GetModule() registration.Module {
	return &module{}
}

var _ registration.Module = (*module)(nil)
