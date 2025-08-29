package web_render

import (
	"embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
)

// TemplateLoader loads templates from a directory or embed.FS
// Usage: loader := NewTemplateLoader("./templates")
//
//	tmpl := loader.Load("base.html")
//
// TemplateLoader handles project-level asset loading and override logic.
// Usage: loader := NewTemplateLoader("./templates")
type TemplateLoader struct {
	LayoutDir string    // path to layouts (project)
	PageDir   string    // path to pages (project)
	EmbedFS   *embed.FS // optional, for embedded fallback
}

// FrameworkAssetLoader handles framework-level asset loading only.
type FrameworkAssetLoader struct {
	LayoutDir string    // path to layouts (framework)
	PageDir   string    // path to pages (framework)
	JsDir     string    // path to js (framework)
	CssDir    string    // path to css (framework)
	EmbedFS   *embed.FS // optional, for embedded fallback
}

// Create a new TemplateLoader for project assets.
func NewTemplateLoader(dir string) *TemplateLoader {
	return &TemplateLoader{
		LayoutDir: dir + "/layouts",
		PageDir:   dir + "/pages",
		EmbedFS:   nil,
	}
}

// Create a new FrameworkAssetLoader for framework assets.
func NewFrameworkAssetLoader(dir string) *FrameworkAssetLoader {
	return &FrameworkAssetLoader{
		LayoutDir: dir + "/layouts",
		PageDir:   dir + "/layouts", // fallback pages to layouts
		JsDir:     dir + "/static/js",
		CssDir:    dir + "/static/css",
		EmbedFS:   nil,
	}
}

// NewTemplateLoaderWithEmbed: loader dengan embed fallback
// Create a new TemplateLoader with embedded fallback.
func NewTemplateLoaderWithEmbed(dir string, fs *embed.FS) *TemplateLoader {
	return &TemplateLoader{
		LayoutDir: dir + "/layouts",
		PageDir:   dir + "/pages",
		EmbedFS:   fs,
	}
}

// Create a new FrameworkAssetLoader with embedded fallback.
func NewFrameworkAssetLoaderWithEmbed(dir string, fs *embed.FS) *FrameworkAssetLoader {
	return &FrameworkAssetLoader{
		LayoutDir: dir + "/layouts",
		PageDir:   dir + "/layouts",
		JsDir:     dir + "/static/js",
		CssDir:    dir + "/static/css",
		EmbedFS:   fs,
	}
}

// Load template from project assets only (layouts/pages). No framework fallback.
func (l *TemplateLoader) Load(name string) (*template.Template, error) {
	projectRoot, _ := os.Getwd()

	// 1. Check layout path Project
	layoutPath := filepath.Join(projectRoot, l.LayoutDir, name)
	fmt.Printf("[DEBUG] Trying layout path: %s\n", layoutPath)
	if _, err := os.Stat(layoutPath); err == nil {
		fmt.Printf("[DEBUG] Found template at layout path: %s\n", layoutPath)
		return template.ParseFiles(layoutPath)
	} else {
		fmt.Printf("[DEBUG] Layout path not found: %v\n", err)
	}

	// 2. Check page path Project
	pagePath := filepath.Join(projectRoot, l.PageDir, name)
	fmt.Printf("[DEBUG] Trying page path: %s\n", pagePath)
	if _, err := os.Stat(pagePath); err == nil {
		fmt.Printf("[DEBUG] Found template at page path: %s\n", pagePath)
		return template.ParseFiles(pagePath)
	} else {
		fmt.Printf("[DEBUG] Page path not found: %v\n", err)
	}
	if l.EmbedFS != nil {
		// 3. Check embed.FS layout dir fallback
		fmt.Printf("[DEBUG] Trying embed.FS for layout: %s\n", filepath.Join(l.LayoutDir, name))
		tmplBytes, err := l.EmbedFS.ReadFile(filepath.Join(l.LayoutDir, name))
		if err == nil {
			fmt.Printf("[DEBUG] Found template in embed.FS layout: %s\n", filepath.Join(l.LayoutDir, name))
			return template.New(name).Parse(string(tmplBytes))
		}
		fmt.Printf("[DEBUG] Embed.FS layout not found: %v\n", err)

		// 4. Check embed.FS page dir fallback
		fmt.Printf("[DEBUG] Trying embed.FS for page: %s\n", filepath.Join(l.PageDir, name))
		tmplBytes, err = l.EmbedFS.ReadFile(filepath.Join(l.PageDir, name))
		if err == nil {
			fmt.Printf("[DEBUG] Found template in embed.FS page: %s\n", filepath.Join(l.PageDir, name))
			return template.New(name).Parse(string(tmplBytes))
		}
		fmt.Printf("[DEBUG] Embed.FS page not found: %v\n", err)

		// 5. Check embed.FS root fallback
		fmt.Printf("[DEBUG] Trying embed.FS for root: %s\n", name)
		tmplBytes, err = l.EmbedFS.ReadFile(name)
		if err == nil {
			fmt.Printf("[DEBUG] Found template in embed.FS root: %s\n", name)
			return template.New(name).Parse(string(tmplBytes))
		}
		fmt.Printf("[DEBUG] Embed.FS root not found: %v\n", err)
	}
	fmt.Printf("[ERROR] Template %s not found in project or embed.FS\n", name)
	return nil, fmt.Errorf("template %s not found in project or embed.FS", name)
}
