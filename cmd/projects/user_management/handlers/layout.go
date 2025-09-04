package handlers

import (
	"bytes"
	"fmt"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/cmd/projects/user_management/handlers/assets"
)

// PageContent holds all page data for consistent behavior
type PageContent struct {
	HTML            string            // Main content HTML
	Title           string            // Page title (for browser tab and meta)
	CurrentPage     string            // Current page identifier (for sidebar active state)
	Scripts         []string          // Page-specific external scripts
	Styles          []string          // Page-specific external styles
	CustomCSS       string            // Page-specific custom CSS
	EmbeddedScripts []string          // Embedded JavaScript file names (compiled into binary)
	MetaTags        map[string]string // Page-specific meta tags
	SidebarData     interface{}       // Custom sidebar data if needed
}

// PageContentFunc is a function that returns complete page content
type PageContentFunc func(*lokstra.Context) (*PageContent, error)

// RenderFullPage renders a complete HTML page with layout
func RenderFullPage(pageContent *PageContent) string {
	// Use template-based rendering
	return renderFullPageFromTemplate(pageContent)
}

// renderFullPageFromTemplate uses external template files
func renderFullPageFromTemplate(pageContent *PageContent) string {
	// Initialize templates if not done yet
	if baseLayoutTemplate == nil {
		if err := InitializeTemplates(); err != nil {
			// Log error and return error message
			fmt.Printf("Template initialization error: %v\n", err)
			return fmt.Sprintf("<html><body><h1>Template Error</h1><p>%v</p></body></html>", err)
		}
	}

	// Prepare template data
	templateData := prepareTemplateData(pageContent)

	// Render the complete page
	var buf bytes.Buffer
	err := baseLayoutTemplate.Execute(&buf, templateData)
	if err != nil {
		// Log error and return error message
		fmt.Printf("Template execution error: %v\n", err)
		return fmt.Sprintf("<html><body><h1>Template Execution Error</h1><p>%v</p></body></html>", err)
	}

	return buf.String()
}

// RenderPartialContent renders just the content for HTMX requests
// WITH page-specific assets for consistent behavior
func RenderPartialContent(pageContent *PageContent) string {
	content := pageContent.HTML

	// Add page-specific external scripts to content for HTMX consistency
	if len(pageContent.Scripts) > 0 {
		for _, script := range pageContent.Scripts {
			content = fmt.Sprintf(`<script src="%s"></script>
%s`, script, content)
		}
	}

	// Add embedded JavaScript scripts to content for HTMX consistency
	if len(pageContent.EmbeddedScripts) > 0 {
		for _, scriptName := range pageContent.EmbeddedScripts {
			embeddedScript := assets.GetEmbeddedScript(scriptName)
			if embeddedScript != "" {
				content = fmt.Sprintf(`<script>
// Embedded Script: %s
%s
</script>
%s`, scriptName, embeddedScript, content)
			}
		}
	}

	// Add page-specific styles to content for HTMX consistency
	if len(pageContent.Styles) > 0 || pageContent.CustomCSS != "" {
		stylesBlock := ""
		for _, style := range pageContent.Styles {
			stylesBlock += fmt.Sprintf(`<link rel="stylesheet" href="%s">
`, style)
		}
		if pageContent.CustomCSS != "" {
			stylesBlock += fmt.Sprintf(`<style>
%s
</style>
`, pageContent.CustomCSS)
		}
		content = stylesBlock + content
	}

	return content
}

// PageHandler creates a handler with consistent behavior for both full page and HTMX requests
func PageHandler(contentFunc PageContentFunc) lokstra.HandlerFunc {
	return func(c *lokstra.Context) error {
		// Get complete page content
		pageContent, err := contentFunc(c)
		if err != nil {
			return err
		}

		// Check if this is an HTMX request for partial content
		isHTMXRequest := c.GetHeader("HX-Request") == "true"

		if isHTMXRequest {
			// Return content WITH page-specific assets for consistency
			html := RenderPartialContent(pageContent)
			return c.HTML(html)
		}

		// Return full page for direct access
		fullPageHTML := RenderFullPage(pageContent)
		return c.HTML(fullPageHTML)
	}
}
