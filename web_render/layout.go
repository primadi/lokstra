package web_render

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/primadi/lokstra/core/request"
)

// PageOptions untuk opsi rendering halaman
type PageOptions struct {
	Title       string
	CurrentPage string
	Scripts     []string
	Styles      []string
	CustomCSS   string
	MetaTags    map[string]string
	SidebarData any
}

var mainLayoutPage = "base.html"

// MainLayoutPage struct untuk menyimpan nama layout utama
type MainLayoutPage struct {
	Name   string
	Loader *TemplateLoader
}

// NewMainLayoutPage: inisialisasi layout utama
func NewMainLayoutPage(name string) *MainLayoutPage {
	// Default loader: layouts/pages di bawah working dir
	loader := NewTemplateLoader(".")
	return &MainLayoutPage{Name: name, Loader: loader}
}

// NewMainLayoutPageWithLoader: inisialisasi layout utama dengan loader custom
func NewMainLayoutPageWithLoader(name string, loader *TemplateLoader) *MainLayoutPage {
	return &MainLayoutPage{Name: name, Loader: loader}
}

// Set layout utama
func (m *MainLayoutPage) Set(name string) {
	m.Name = name
}

// Get layout utama
func (m *MainLayoutPage) Get() string {
	return m.Name
}

// RenderPage: API utama untuk render halaman dengan layout dan data
// Menggunakan TemplateLoader override/fallback
func (m *MainLayoutPage) RenderPage(
	c *request.Context,
	templateName string,
	data any,
	opts *PageOptions,
) *PageContent {
	// use loader from struct
	loader := m.Loader
	contentHTML := ""
	// When fullLayout, load all templates needed for composition
	var tmpl *template.Template
	var err error

	isHTMX := false
	if c != nil {
		isHTMX = c.GetHeader("HX-Request") == "true"
	}
	fullLayout := true
	if isHTMX {
		fullLayout = false
	}
	if opts != nil && opts.MetaTags != nil {
		if v, ok := opts.MetaTags["full_layout"]; ok && v == "false" {
			fullLayout = false
		}
	}

	if fullLayout {
		// Load layout, page, and sidebar partial/component
		layoutPath := loader.LayoutDir + "/" + m.Name
		pagePath := loader.PageDir + "/" + templateName + ".html"
		sidebarPath := loader.LayoutDir + "/sidebar.html"
		// Add more partials/components as needed
		tmpl, err = template.ParseFiles(layoutPath, pagePath, sidebarPath)
	} else {
		// Only load the page template for partial/HTMX
		pagePath := loader.PageDir + "/" + templateName + ".html"
		tmpl, err = template.ParseFiles(pagePath)
	}
	if err == nil && tmpl != nil {
		var buf strings.Builder
		if fullLayout {
			if err := tmpl.ExecuteTemplate(&buf, m.Name, data); err == nil {
				contentHTML = buf.String()
			} else {
				contentHTML = "<div>Layout execution error: " + err.Error() + "</div>"
			}
		} else {
			if err := tmpl.ExecuteTemplate(&buf, templateName+".html", data); err == nil {
				contentHTML = buf.String()
			} else {
				contentHTML = "<div>Template execution error: " + err.Error() + "</div>"
			}
		}
	} else {
		contentHTML = "<div>Template not found: " + templateName + ".html</div>"
	}

	html := contentHTML
	if fullLayout {
		// Bisa override via opts.MetaTags["main_layout"]
		layoutName := m.Name
		if opts != nil && opts.MetaTags != nil {
			if v, ok := opts.MetaTags["main_layout"]; ok && v != "" {
				layoutName = v
			}
		}
		// Render mainLayout and inject contentHTML into {{.Content}}
		layoutData := struct {
			PageContent
			Content string
		}{
			PageContent: PageContent{
				Title:       opts.Title,
				CurrentPage: opts.CurrentPage,
				Scripts:     opts.Scripts,
				Styles:      opts.Styles,
				CustomCSS:   opts.CustomCSS,
				MetaTags:    opts.MetaTags,
				SidebarData: opts.SidebarData,
			},
			Content: contentHTML,
		}
		layoutTmpl, err := loader.Load(layoutName)
		if err == nil && layoutTmpl != nil {
			var buf strings.Builder
			if err := layoutTmpl.ExecuteTemplate(&buf, layoutName, layoutData); err == nil {
				html = buf.String()
			} else {
				html = "<div>Layout execution error: " + err.Error() + "</div>"
			}
		} else {
			html = "<div>Layout template not found: " + layoutName + "</div>" + contentHTML
		}
	}

	// Build PageContent
	return &PageContent{
		HTML:        html,
		Title:       opts.Title,
		CurrentPage: opts.CurrentPage,
		Scripts:     opts.Scripts,
		Styles:      opts.Styles,
		CustomCSS:   opts.CustomCSS,
		MetaTags:    opts.MetaTags,
		SidebarData: opts.SidebarData,
	}
}

// PageContent holds all page data for consistent behavior
// (moved from user_management/handlers)
type PageContent struct {
	HTML        string            // Main content HTML
	Title       string            // Page title (for browser tab and meta)
	CurrentPage string            // Current page identifier (for sidebar active state)
	Scripts     []string          // Page-specific external scripts
	Styles      []string          // Page-specific external styles
	CustomCSS   string            // Page-specific custom CSS
	MetaTags    map[string]string // Page-specific meta tags
	SidebarData any               // Custom sidebar data if needed
}

// PageContentFunc is a function that returns complete page content
// (moved from user_management/handlers)
type PageContentFunc func(*request.Context) (*PageContent, error)

// RenderFullPage renders a complete HTML page with layout
func RenderFullPage(pageContent *PageContent, renderTemplate func(*PageContent) (string, error)) string {
	result, err := renderTemplate(pageContent)
	if err != nil {
		fmt.Printf("Template execution error: %v\n", err)
		return fmt.Sprintf("<html><body><h1>Template Execution Error</h1><p>%v</p></body></html>", err)
	}
	return result
}

// RenderPartialContent renders just the content for HTMX requests
// WITH page-specific assets for consistent behavior
func RenderPartialContent(pageContent *PageContent) string {
	content := pageContent.HTML

	// Add page-specific external scripts to content for HTMX consistency
	if len(pageContent.Scripts) > 0 {
		for _, script := range pageContent.Scripts {
			content = fmt.Sprintf(`<script src="%s"></script>\n%s`, script, content)
		}
	}

	// Add page-specific styles to content for HTMX consistency
	if len(pageContent.Styles) > 0 || pageContent.CustomCSS != "" {
		stylesBlock := ""
		for _, style := range pageContent.Styles {
			stylesBlock += fmt.Sprintf(`<link rel="stylesheet" href="%s">\n`, style)
		}
		if pageContent.CustomCSS != "" {
			stylesBlock += fmt.Sprintf(`<style>\n%s\n</style>\n`, pageContent.CustomCSS)
		}
		content = stylesBlock + content
	}

	return content
}

// PageHandler creates a handler with consistent behavior for both full page and HTMX requests
func PageHandler(contentFunc PageContentFunc, renderTemplate func(*PageContent) (string, error)) func(*request.Context) error {
	return func(c *request.Context) error {
		pageContent, err := contentFunc(c)
		if err != nil {
			return err
		}
		isHTMXRequest := c.GetHeader("HX-Request") == "true"
		if isHTMXRequest {
			html := RenderPartialContent(pageContent)
			return c.HTML(html)
		}
		fullPageHTML := RenderFullPage(pageContent, renderTemplate)
		return c.HTML(fullPageHTML)
	}
}
