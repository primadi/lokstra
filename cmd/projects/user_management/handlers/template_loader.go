package handlers

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"github.com/primadi/lokstra/cmd/projects/user_management/handlers/assets"
)

//go:embed templates/*.html
var templateFS embed.FS

// TemplateData holds all data for template rendering
type TemplateData struct {
	HTML            string
	Title           string
	SidebarHTML     string
	MetaTags        string
	Scripts         string
	Styles          string
	PageMetaTags    map[string]string
	MainPageScripts []ScriptData
	ExternalScripts []string
	EmbeddedScripts []ScriptData
	MainLayoutCSS   string
	ExternalStyles  []string
	CustomCSS       string
}

// ScriptData holds script name and content
type ScriptData struct {
	Name    string
	Content string
}

var (
	baseLayoutTemplate *template.Template
	metaTagsTemplate   *template.Template
	scriptsTemplate    *template.Template
	stylesTemplate     *template.Template
)

// InitializeTemplates loads and parses all templates
func InitializeTemplates() error {
	var err error

	// Create template functions for safe content rendering
	funcMap := template.FuncMap{
		"safeJS": func(s string) template.JS {
			return template.JS(s)
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	}

	// Load base layout template
	baseLayoutContent, err := templateFS.ReadFile("templates/base_layout.html")
	if err != nil {
		return fmt.Errorf("failed to read base_layout.html: %w", err)
	}

	// Load meta tags template
	metaTagsContent, err := templateFS.ReadFile("templates/meta_tags.html")
	if err != nil {
		return fmt.Errorf("failed to read meta_tags.html: %w", err)
	}

	// Load scripts template
	scriptsContent, err := templateFS.ReadFile("templates/scripts.html")
	if err != nil {
		return fmt.Errorf("failed to read scripts.html: %w", err)
	}

	// Load styles template
	stylesContent, err := templateFS.ReadFile("templates/styles.html")
	if err != nil {
		return fmt.Errorf("failed to read styles.html: %w", err)
	}

	// Parse templates with function map
	baseLayoutTemplate, err = template.New("base_layout").Funcs(funcMap).Parse(string(baseLayoutContent))
	if err != nil {
		return fmt.Errorf("failed to parse base_layout template: %w", err)
	}

	metaTagsTemplate, err = template.New("meta_tags").Funcs(funcMap).Parse(string(metaTagsContent))
	if err != nil {
		return fmt.Errorf("failed to parse meta_tags template: %w", err)
	}

	scriptsTemplate, err = template.New("scripts").Funcs(funcMap).Parse(string(scriptsContent))
	if err != nil {
		return fmt.Errorf("failed to parse scripts template: %w", err)
	}

	stylesTemplate, err = template.New("styles").Funcs(funcMap).Parse(string(stylesContent))
	if err != nil {
		return fmt.Errorf("failed to parse styles template: %w", err)
	}

	return nil
}

// prepareTemplateData converts PageContent to TemplateData
func prepareTemplateData(pageContent *PageContent) *TemplateData {
	data := &TemplateData{
		HTML:            pageContent.HTML,
		Title:           pageContent.Title,
		SidebarHTML:     getSidebarHTML(pageContent.CurrentPage),
		PageMetaTags:    pageContent.MetaTags,
		ExternalScripts: pageContent.Scripts,
		ExternalStyles:  pageContent.Styles,
		CustomCSS:       pageContent.CustomCSS,
		MainLayoutCSS:   assets.GetEmbeddedStyle("main-layout"),
	}

	// Render meta tags
	data.MetaTags = renderMetaTags(data)

	// Render scripts
	data.Scripts = renderScripts(data, pageContent)

	// Render styles
	data.Styles = renderStyles(data)

	return data
}

// renderMetaTags renders meta tags using template
func renderMetaTags(data *TemplateData) string {
	var buf bytes.Buffer
	if err := metaTagsTemplate.Execute(&buf, data); err != nil {
		// Fallback to basic meta tags
		return `<meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0">`
	}
	return buf.String()
}

// renderScripts renders scripts using template
func renderScripts(data *TemplateData, pageContent *PageContent) string {
	// Prepare main page scripts
	mainPageScripts := []string{
		"main-page-navigation",
		"main-page-sidebar",
		"main-page-emergency-cleanup",
		"main-page-app",
	}

	for _, scriptName := range mainPageScripts {
		embeddedScript := assets.GetEmbeddedScript(scriptName)
		if embeddedScript != "" {
			data.MainPageScripts = append(data.MainPageScripts, ScriptData{
				Name:    scriptName,
				Content: embeddedScript,
			})
		}
	}

	// Prepare embedded scripts
	for _, scriptName := range pageContent.EmbeddedScripts {
		embeddedScript := assets.GetEmbeddedScript(scriptName)
		if embeddedScript != "" {
			data.EmbeddedScripts = append(data.EmbeddedScripts, ScriptData{
				Name:    scriptName,
				Content: embeddedScript,
			})
		}
	}

	var buf bytes.Buffer
	if err := scriptsTemplate.Execute(&buf, data); err != nil {
		// Fallback to basic scripts
		return `<script src="https://cdn.tailwindcss.com"></script><script src="https://unpkg.com/htmx.org@1.9.0"></script>`
	}
	return buf.String()
}

// renderStyles renders styles using template
func renderStyles(data *TemplateData) string {
	var buf bytes.Buffer
	if err := stylesTemplate.Execute(&buf, data); err != nil {
		// Fallback to basic styles
		return fmt.Sprintf(`<style>%s</style>`, data.MainLayoutCSS)
	}
	return buf.String()
}
