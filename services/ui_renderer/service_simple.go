package ui_renderer

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"path/filepath"

	"github.com/primadi/lokstra/serviceapi"
)

// Config holds configuration for the UI renderer service
type Config struct {
	TemplateDir    string                 `yaml:"template_dir" json:"template_dir"`
	Theme          string                 `yaml:"theme" json:"theme"`
	CacheTemplates bool                   `yaml:"cache_templates" json:"cache_templates"`
	HotReload      bool                   `yaml:"hot_reload" json:"hot_reload"`
	MinifyOutput   bool                   `yaml:"minify_output" json:"minify_output"`
	ComponentDirs  []string               `yaml:"component_dirs" json:"component_dirs"`
	PrelineVersion string                 `yaml:"preline_version" json:"preline_version"`
	HTMXConfig     map[string]interface{} `yaml:"htmx_config" json:"htmx_config"`
	AlpineConfig   map[string]interface{} `yaml:"alpine_config" json:"alpine_config"`
	TailwindConfig map[string]interface{} `yaml:"tailwind_config" json:"tailwind_config"`
}

// UIRenderer implements the serviceapi.UIRenderer interface
type UIRenderer struct {
	config    *Config
	templates *template.Template
}

// NewService creates a new UI Renderer service instance
func NewService(config *Config) serviceapi.UIRenderer {
	service := &UIRenderer{
		config: config,
	}

	// Load templates
	if err := service.loadTemplates(); err != nil {
		// In production, you might want to handle this differently
		fmt.Printf("Warning: Failed to load templates: %v\n", err)
	}

	return service
}

// loadTemplates loads all template files
func (u *UIRenderer) loadTemplates() error {
	if u.config.TemplateDir == "" {
		u.config.TemplateDir = "./templates"
	}

	// Create template with helper functions
	tmpl := template.New("ui").Funcs(template.FuncMap{
		"renderHTMXAttrs": func(config interface{}) string {
			// Helper function to render HTMX attributes
			return ""
		},
		"renderAlpineAttrs": func(config interface{}) string {
			// Helper function to render Alpine.js attributes
			return ""
		},
		"formatDate": func(date interface{}) string {
			return fmt.Sprintf("%v", date)
		},
		"formatCurrency": func(amount interface{}) string {
			return fmt.Sprintf("$%v", amount)
		},
		"default": func(defaultVal, value interface{}) interface{} {
			if value == nil || value == "" {
				return defaultVal
			}
			return value
		},
	})

	// Load template files
	templatePattern := filepath.Join(u.config.TemplateDir, "*.html")
	var err error
	u.templates, err = tmpl.ParseGlob(templatePattern)
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	return nil
}

// RenderApp renders the main application layout
func (u *UIRenderer) RenderApp(ctx context.Context, config *serviceapi.AppConfig) (template.HTML, error) {
	if u.templates == nil {
		return template.HTML(u.renderFallbackApp(config)), nil
	}

	var buf bytes.Buffer
	if err := u.templates.ExecuteTemplate(&buf, "layout.html", config); err != nil {
		return template.HTML(u.renderFallbackApp(config)), nil
	}

	return template.HTML(buf.String()), nil
}

// RenderForm renders a form component
func (u *UIRenderer) RenderForm(ctx context.Context, config *serviceapi.FormConfig) (template.HTML, error) {
	if u.templates == nil {
		return template.HTML(u.renderFallbackForm(config)), nil
	}

	var buf bytes.Buffer
	if err := u.templates.ExecuteTemplate(&buf, "form", config); err != nil {
		return template.HTML(u.renderFallbackForm(config)), nil
	}

	return template.HTML(buf.String()), nil
}

// RenderList renders a list/table component
func (u *UIRenderer) RenderList(ctx context.Context, config *serviceapi.ListConfig, data interface{}) (template.HTML, error) {
	if u.templates == nil {
		return template.HTML(u.renderFallbackList(config)), nil
	}

	// Combine config and data for template
	templateData := map[string]interface{}{
		"Config": config,
		"Data":   data,
	}

	var buf bytes.Buffer
	if err := u.templates.ExecuteTemplate(&buf, "table", templateData); err != nil {
		return template.HTML(u.renderFallbackList(config)), nil
	}

	return template.HTML(buf.String()), nil
}

// RenderComponent renders individual components
func (u *UIRenderer) RenderComponent(ctx context.Context, componentType string, props map[string]interface{}) (template.HTML, error) {
	if u.templates == nil {
		return template.HTML(u.renderFallbackComponent(componentType, props)), nil
	}

	var buf bytes.Buffer
	if err := u.templates.ExecuteTemplate(&buf, componentType, props); err != nil {
		return template.HTML(u.renderFallbackComponent(componentType, props)), nil
	}

	return template.HTML(buf.String()), nil
}

// RenderBreadcrumb renders breadcrumb navigation
func (u *UIRenderer) RenderBreadcrumb(ctx context.Context, config *serviceapi.BreadcrumbConfig) (template.HTML, error) {
	if u.templates == nil {
		return template.HTML(u.renderFallbackBreadcrumb(config)), nil
	}

	var buf bytes.Buffer
	if err := u.templates.ExecuteTemplate(&buf, "breadcrumb", config); err != nil {
		return template.HTML(u.renderFallbackBreadcrumb(config)), nil
	}

	return template.HTML(buf.String()), nil
}

// RenderMenu renders menu navigation
func (u *UIRenderer) RenderMenu(ctx context.Context, config *serviceapi.MenuConfig) (template.HTML, error) {
	if u.templates == nil {
		return template.HTML(u.renderFallbackMenu(config)), nil
	}

	var buf bytes.Buffer
	if err := u.templates.ExecuteTemplate(&buf, "menu", config); err != nil {
		return template.HTML(u.renderFallbackMenu(config)), nil
	}

	return template.HTML(buf.String()), nil
}

// RenderField renders a single form field
func (u *UIRenderer) RenderField(ctx context.Context, field *serviceapi.FieldConfig) (template.HTML, error) {
	if u.templates == nil {
		return template.HTML(u.renderFallbackField(field)), nil
	}

	var buf bytes.Buffer
	if err := u.templates.ExecuteTemplate(&buf, "field", field); err != nil {
		return template.HTML(u.renderFallbackField(field)), nil
	}

	return template.HTML(buf.String()), nil
}

// RenderTable renders a data table
func (u *UIRenderer) RenderTable(ctx context.Context, config *serviceapi.TableConfig, data interface{}) (template.HTML, error) {
	if u.templates == nil {
		return template.HTML(u.renderFallbackTable(config)), nil
	}

	templateData := map[string]interface{}{
		"Config": config,
		"Data":   data,
	}

	var buf bytes.Buffer
	if err := u.templates.ExecuteTemplate(&buf, "table.html", templateData); err != nil {
		return template.HTML(u.renderFallbackTable(config)), nil
	}

	return template.HTML(buf.String()), nil
}

// RenderPagination renders pagination controls
func (u *UIRenderer) RenderPagination(ctx context.Context, config *serviceapi.PaginationConfig) (template.HTML, error) {
	if u.templates == nil {
		return template.HTML(u.renderFallbackPagination(config)), nil
	}

	var buf bytes.Buffer
	if err := u.templates.ExecuteTemplate(&buf, "pagination", config); err != nil {
		return template.HTML(u.renderFallbackPagination(config)), nil
	}

	return template.HTML(buf.String()), nil
}

// RenderModal renders modal dialog
func (u *UIRenderer) RenderModal(ctx context.Context, config *serviceapi.ModalConfig) (template.HTML, error) {
	if u.templates == nil {
		return template.HTML(u.renderFallbackModal(config)), nil
	}

	var buf bytes.Buffer
	if err := u.templates.ExecuteTemplate(&buf, "modal", config); err != nil {
		return template.HTML(u.renderFallbackModal(config)), nil
	}

	return template.HTML(buf.String()), nil
}

// RenderCard renders card component
func (u *UIRenderer) RenderCard(ctx context.Context, config *serviceapi.CardConfig) (template.HTML, error) {
	if u.templates == nil {
		return template.HTML(u.renderFallbackCard(config)), nil
	}

	var buf bytes.Buffer
	if err := u.templates.ExecuteTemplate(&buf, "card", config); err != nil {
		return template.HTML(u.renderFallbackCard(config)), nil
	}

	return template.HTML(buf.String()), nil
}

// ParseTemplate loads a template from a file
func (u *UIRenderer) ParseTemplate(templatePath string) error {
	if u.templates == nil {
		u.templates = template.New("ui")
	}

	_, err := u.templates.ParseFiles(templatePath)
	return err
}

// RenderTemplate renders a template by name with data
func (u *UIRenderer) RenderTemplate(ctx context.Context, templateName string, data interface{}) (template.HTML, error) {
	if u.templates == nil {
		return template.HTML(""), fmt.Errorf("no templates loaded")
	}

	var buf bytes.Buffer
	if err := u.templates.ExecuteTemplate(&buf, templateName, data); err != nil {
		return template.HTML(""), err
	}

	return template.HTML(buf.String()), nil
}

// Additional fallback methods

func (u *UIRenderer) renderFallbackApp(config *serviceapi.AppConfig) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>%s</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <script defer src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body class="bg-gray-50">
    <div class="min-h-screen">
        <header class="bg-white shadow">
            <div class="px-6 py-4">
                <h1 class="text-2xl font-bold text-gray-900">%s</h1>
            </div>
        </header>
        <main class="p-6">
            <div id="main-content">
                <!-- Content will be loaded here -->
            </div>
        </main>
    </div>
</body>
</html>`, config.Title, config.Title)
}

func (u *UIRenderer) renderFallbackForm(config *serviceapi.FormConfig) string {
	fieldsHTML := ""
	for _, field := range config.Fields {
		fieldsHTML += fmt.Sprintf(`
        <div class="mb-4">
            <label class="block text-sm font-medium text-gray-700 mb-2">%s</label>
            <input type="%s" name="%s" placeholder="%s" 
                   class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
                   %s>
        </div>`, field.Label, field.Type, field.Name, field.Placeholder,
			func() string {
				if field.Required {
					return "required"
				}
				return ""
			}())
	}

	return fmt.Sprintf(`
<div class="max-w-2xl mx-auto bg-white p-6 rounded-lg shadow">
    <h2 class="text-xl font-bold mb-6">%s</h2>
    <form action="%s" method="%s">
        %s
        <button type="submit" class="w-full bg-indigo-600 text-white py-2 px-4 rounded-md hover:bg-indigo-700">
            Submit
        </button>
    </form>
</div>`, config.Title, config.Action, config.Method, fieldsHTML)
}

func (u *UIRenderer) renderFallbackList(config *serviceapi.ListConfig) string {
	return fmt.Sprintf(`
<div class="bg-white rounded-lg shadow">
    <div class="p-6 border-b">
        <h2 class="text-xl font-bold">%s</h2>
        <p class="text-gray-600">%s</p>
    </div>
    <div class="p-6">
        <div id="list-content">
            <!-- List content will be loaded here -->
        </div>
    </div>
</div>`, config.Title, config.Description)
}

func (u *UIRenderer) renderFallbackComponent(componentType string, props map[string]interface{}) string {
	title, _ := props["title"].(string)
	return fmt.Sprintf(`
<div class="component component-%s bg-white p-4 rounded-lg shadow">
    <h3 class="font-medium">%s</h3>
    <div class="mt-2">
        <!-- %s component content -->
    </div>
</div>`, componentType, title, componentType)
}

func (u *UIRenderer) renderFallbackBreadcrumb(config *serviceapi.BreadcrumbConfig) string {
	items := ""
	for i, item := range config.Items {
		separator := ""
		if i > 0 {
			separator = `<span class="mx-2 text-gray-400">/</span>`
		}

		if item.URL != "" {
			items += fmt.Sprintf(`%s<a href="%s" class="text-indigo-600 hover:text-indigo-800">%s</a>`,
				separator, item.URL, item.Label)
		} else {
			items += fmt.Sprintf(`%s<span class="text-gray-900 font-medium">%s</span>`,
				separator, item.Label)
		}
	}

	return fmt.Sprintf(`
<nav class="flex mb-4" aria-label="Breadcrumb">
    <ol class="flex items-center space-x-1">
        %s
    </ol>
</nav>`, items)
}

func (u *UIRenderer) renderFallbackMenu(config *serviceapi.MenuConfig) string {
	itemsHTML := ""
	for _, item := range config.Items {
		activeClass := ""
		if item.Active {
			activeClass = "bg-indigo-100 text-indigo-700"
		}

		itemsHTML += fmt.Sprintf(`
        <li>
            <a href="%s" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-100 %s">
                %s
            </a>
        </li>`, item.URL, activeClass, item.Label)
	}

	return fmt.Sprintf(`
<nav class="space-y-1">
    <ul>
        %s
    </ul>
</nav>`, itemsHTML)
}

func (u *UIRenderer) renderFallbackField(field *serviceapi.FieldConfig) string {
	required := ""
	if field.Required {
		required = "required"
	}

	return fmt.Sprintf(`
<div class="mb-4">
    <label class="block text-sm font-medium text-gray-700 mb-2">%s</label>
    <input type="%s" name="%s" placeholder="%s" 
           class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
           %s>
</div>`, field.Label, field.Type, field.Name, field.Placeholder, required)
}

func (u *UIRenderer) renderFallbackTable(config *serviceapi.TableConfig) string {
	headersHTML := ""
	for _, col := range config.Columns {
		headersHTML += fmt.Sprintf(`<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">%s</th>`, col.Label)
	}

	return fmt.Sprintf(`
<div class="overflow-hidden shadow ring-1 ring-black ring-opacity-5 md:rounded-lg">
    <table class="min-w-full divide-y divide-gray-300">
        <thead class="bg-gray-50">
            <tr>
                %s
            </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
            <!-- Table rows will be loaded here -->
        </tbody>
    </table>
</div>`, headersHTML)
}

func (u *UIRenderer) renderFallbackPagination(config *serviceapi.PaginationConfig) string {
	if !config.Enabled {
		return ""
	}

	return fmt.Sprintf(`
<div class="flex items-center justify-between border-t border-gray-200 bg-white px-4 py-3 sm:px-6">
    <div class="flex flex-1 justify-between sm:hidden">
        <a href="#" class="relative inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50">Previous</a>
        <a href="#" class="relative ml-3 inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50">Next</a>
    </div>
    <div class="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
        <div>
            <p class="text-sm text-gray-700">
                Showing results (Page size: %d)
            </p>
        </div>
        <div class="flex items-center space-x-2">
            <!-- Pagination buttons will be rendered here -->
        </div>
    </div>
</div>`, config.PageSize)
}

func (u *UIRenderer) renderFallbackModal(config *serviceapi.ModalConfig) string {
	return fmt.Sprintf(`
<div class="fixed inset-0 z-10 overflow-y-auto" aria-labelledby="modal-title" role="dialog" aria-modal="true">
    <div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" aria-hidden="true"></div>
        <div class="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
            <div class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
                <div class="mt-3 text-center sm:mt-0 sm:text-left">
                    <h3 class="text-lg leading-6 font-medium text-gray-900" id="modal-title">%s</h3>
                    <div class="mt-2">
                        <p class="text-sm text-gray-500">%s</p>
                    </div>
                </div>
            </div>
            <div class="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
                <button type="button" class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-indigo-600 text-base font-medium text-white hover:bg-indigo-700 focus:outline-none sm:ml-3 sm:w-auto sm:text-sm">
                    Confirm
                </button>
                <button type="button" class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm">
                    Cancel
                </button>
            </div>
        </div>
    </div>
</div>`, config.Title, config.Content)
}

func (u *UIRenderer) renderFallbackCard(config *serviceapi.CardConfig) string {
	return fmt.Sprintf(`
<div class="bg-white overflow-hidden shadow rounded-lg">
    <div class="px-4 py-5 sm:p-6">
        <h3 class="text-lg leading-6 font-medium text-gray-900">%s</h3>
        <div class="mt-2">
            <p class="text-sm text-gray-500">%s</p>
        </div>
        <div class="mt-3">
            %s
        </div>
    </div>
</div>`, config.Title, config.Description, config.Content)
}
