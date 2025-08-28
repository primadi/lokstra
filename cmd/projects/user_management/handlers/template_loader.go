package handlers

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"github.com/primadi/lokstra/cmd/projects/user_management/handlers/assets"
)

// Embed all template files from the organized structure
//
//go:embed templates/layouts/*.html templates/pages/*.html templates/components/*.html templates/assets/scripts.html templates/assets/styles.html templates/assets/page-styles/*.css
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
	// Layout templates
	baseLayoutTemplate  *template.Template
	metaTagsTemplate    *template.Template
	sidebarHTMLTemplate *template.Template

	// Asset templates
	scriptsTemplate *template.Template
	stylesTemplate  *template.Template

	// Page content templates
	dashboardTemplate *template.Template
	usersTemplate     *template.Template
	userFormTemplate  *template.Template
	rolesTemplate     *template.Template
	settingsTemplate  *template.Template

	// Component templates
	formsTemplate  *template.Template
	tablesTemplate *template.Template
	commonTemplate *template.Template
)

// InitializeTemplates loads and parses all templates from the new organized structure
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

	// === LAYOUT TEMPLATES ===

	// Load base layout template
	baseLayoutContent, err := templateFS.ReadFile("templates/layouts/base.html")
	if err != nil {
		return fmt.Errorf("failed to read layouts/base.html: %w", err)
	}

	baseLayoutTemplate, err = template.New("base_layout").Funcs(funcMap).Parse(string(baseLayoutContent))
	if err != nil {
		return fmt.Errorf("failed to parse base_layout template: %w", err)
	}

	// Load meta tags template
	metaTagsContent, err := templateFS.ReadFile("templates/layouts/meta_tags.html")
	if err != nil {
		return fmt.Errorf("failed to read layouts/meta_tags.html: %w", err)
	}

	metaTagsTemplate, err = template.New("meta_tags").Funcs(funcMap).Parse(string(metaTagsContent))
	if err != nil {
		return fmt.Errorf("failed to parse meta_tags template: %w", err)
	}

	// Load sidebar template
	sidebarContent, err := templateFS.ReadFile("templates/layouts/sidebar.html")
	if err != nil {
		return fmt.Errorf("failed to read layouts/sidebar.html: %w", err)
	}

	sidebarHTMLTemplate, err = template.New("sidebar").Funcs(funcMap).Parse(string(sidebarContent))
	if err != nil {
		return fmt.Errorf("failed to parse sidebar template: %w", err)
	}

	// === ASSET TEMPLATES ===

	// Load scripts template
	scriptsContent, err := templateFS.ReadFile("templates/assets/scripts.html")
	if err != nil {
		return fmt.Errorf("failed to read assets/scripts.html: %w", err)
	}

	scriptsTemplate, err = template.New("scripts").Funcs(funcMap).Parse(string(scriptsContent))
	if err != nil {
		return fmt.Errorf("failed to parse scripts template: %w", err)
	}

	// Load styles template
	stylesContent, err := templateFS.ReadFile("templates/assets/styles.html")
	if err != nil {
		return fmt.Errorf("failed to read assets/styles.html: %w", err)
	}

	stylesTemplate, err = template.New("styles").Funcs(funcMap).Parse(string(stylesContent))
	if err != nil {
		return fmt.Errorf("failed to parse styles template: %w", err)
	} // === PAGE TEMPLATES ===

	// Load page content templates
	pageTemplates := map[string]**template.Template{
		"dashboard": &dashboardTemplate,
		"users":     &usersTemplate,
		"user-form": &userFormTemplate,
		"roles":     &rolesTemplate,
		"settings":  &settingsTemplate,
	}

	for templateName, templateVar := range pageTemplates {
		content, err := templateFS.ReadFile(fmt.Sprintf("templates/pages/%s.html", templateName))
		if err != nil {
			return fmt.Errorf("failed to read pages/%s.html: %w", templateName, err)
		}

		*templateVar, err = template.New(templateName).Funcs(funcMap).Parse(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse %s template: %w", templateName, err)
		}
	}

	// === COMPONENT TEMPLATES ===

	// Load component templates
	componentTemplates := map[string]**template.Template{
		"forms":  &formsTemplate,
		"tables": &tablesTemplate,
		"common": &commonTemplate,
	}

	for templateName, templateVar := range componentTemplates {
		content, err := templateFS.ReadFile(fmt.Sprintf("templates/components/%s.html", templateName))
		if err != nil {
			// Components are optional, so just log and continue
			fmt.Printf("Warning: failed to read components/%s.html: %v\n", templateName, err)
			continue
		}

		*templateVar, err = template.New(templateName).Funcs(funcMap).Parse(string(content))
		if err != nil {
			fmt.Printf("Warning: failed to parse %s component template: %v\n", templateName, err)
			continue
		}
	}

	fmt.Println("✅ All templates loaded successfully from organized structure!")
	return nil
}

// prepareTemplateData converts PageContent to TemplateData
func prepareTemplateData(pageContent *PageContent) *TemplateData {
	// Load page-specific CSS if available
	pageCSS := loadPageCSS(pageContent.CurrentPage)
	customCSS := pageContent.CustomCSS
	if pageCSS != "" {
		if customCSS != "" {
			customCSS = pageCSS + "\n" + customCSS
		} else {
			customCSS = pageCSS
		}
	}

	data := &TemplateData{
		HTML:            pageContent.HTML,
		Title:           pageContent.Title,
		SidebarHTML:     getSidebarHTML(pageContent.CurrentPage),
		PageMetaTags:    pageContent.MetaTags,
		ExternalScripts: pageContent.Scripts,
		ExternalStyles:  pageContent.Styles,
		CustomCSS:       customCSS,
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

// renderSidebar renders sidebar using template with menu data
func renderSidebar(menuData interface{}) string {
	var buf bytes.Buffer
	if err := sidebarHTMLTemplate.Execute(&buf, menuData); err != nil {
		// Fallback to basic sidebar if template rendering fails
		return `<div class="w-64 bg-gray-800 border-r border-gray-700">Template Error</div>`
	}
	return buf.String()
}

// renderPageContent renders page content using the appropriate template
func renderPageContent(templateName string, data interface{}) string {
	var tmpl *template.Template

	switch templateName {
	case "dashboard":
		tmpl = dashboardTemplate
	case "users":
		tmpl = usersTemplate
	case "user-form":
		tmpl = userFormTemplate
	case "roles":
		tmpl = rolesTemplate
	case "settings":
		tmpl = settingsTemplate
	default:
		return fmt.Sprintf(`<div class="text-red-400">Template "%s" not found</div>`, templateName)
	}

	if tmpl == nil {
		return fmt.Sprintf(`<div class="text-red-400">Template "%s" not initialized</div>`, templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Sprintf(`<div class="text-red-400">Template rendering error: %v</div>`, err)
	}
	return buf.String()
}

// loadPageCSS loads CSS content for a specific page from the organized structure
func loadPageCSS(pageName string) string {
	content, err := templateFS.ReadFile(fmt.Sprintf("templates/assets/page-styles/%s.css", pageName))
	if err != nil {
		return "" // No CSS file for this page
	}
	return string(content)
}
