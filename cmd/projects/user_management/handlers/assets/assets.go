package assets

import (
	_ "embed"
	"html/template"
)

// Embedded JavaScript files - these will be compiled into the binary
// Paths are relative to the location of this file (handlers/ directory)

//go:embed js/user-form-validation.js
var userFormValidationJS string

//go:embed js/table-enhancements.js
var tableEnhancementsJS string

//go:embed js/navigation-enhancements.js
var navigationEnhancementsJS string

// Main page JavaScript components
//
//go:embed js/main-page/navigation.js
var mainPageNavigationJS string

//go:embed js/main-page/sidebar.js
var mainPageSidebarJS string

//go:embed js/main-page/emergency-cleanup.js
var mainPageEmergencyCleanupJS string

//go:embed js/main-page/app.js
var mainPageAppJS string

// CSS files
//
//go:embed css/main-layout.css
var mainLayoutCSS string

// JavaScript registry for easy access
var EmbeddedScripts = map[string]string{
	"user-form-validation":        userFormValidationJS,
	"table-enhancements":          tableEnhancementsJS,
	"navigation-enhancements":     navigationEnhancementsJS,
	"main-page-navigation":        mainPageNavigationJS,
	"main-page-sidebar":           mainPageSidebarJS,
	"main-page-emergency-cleanup": mainPageEmergencyCleanupJS,
	"main-page-app":               mainPageAppJS,
}

// CSS registry for easy access
var EmbeddedStyles = map[string]string{
	"main-layout": mainLayoutCSS,
}

// GetEmbeddedScript returns the JavaScript content for a given script name
func GetEmbeddedScript(scriptName string) string {
	if script, exists := EmbeddedScripts[scriptName]; exists {
		return script
	}
	return ""
}

// GetEmbeddedStyle returns the CSS content for a given style name
func GetEmbeddedStyle(styleName string) string {
	if style, exists := EmbeddedStyles[styleName]; exists {
		return style
	}
	return ""
}

// GetInlineScript returns JavaScript wrapped in script tags for inline embedding
func GetInlineScript(scriptName string) template.HTML {
	script := GetEmbeddedScript(scriptName)
	if script == "" {
		return ""
	}

	return template.HTML("<script>" + script + "</script>")
}

// GetInlineStyle returns CSS wrapped in style tags for inline embedding
func GetInlineStyle(styleName string) template.HTML {
	style := GetEmbeddedStyle(styleName)
	if style == "" {
		return ""
	}

	return template.HTML("<style>" + style + "</style>")
}
