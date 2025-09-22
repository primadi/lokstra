package htmx_fsmanager

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"regexp"
	"strings"
)

// regex to detect layout directive in HTML page comments
var layoutRegex = regexp.MustCompile(`<!--\s*layout:\s*([a-zA-Z0-9_\-./]+)\s*-->`)
var titleRegex = regexp.MustCompile(`<!--\s*title:\s*([a-zA-Z0-9_\-./]+)\s*-->`)

// regex to detect existing <title> tag in HTML layout
var titleLayoutRegex = regexp.MustCompile(`(?i)<title>.*?</title>`)

type IContainer interface {
	GetHtmxFsManager() *HtmxFsManager
}

var staticCounter = 0

type HtmxFsManager struct {
	layoutsFiles    []fs.FS
	pagesFiles      []fs.FS
	staticFiles     []fs.FS
	fallbackHfm     *HtmxFsManager
	scriptInjection *ScriptInjection
	staticCounter   int
}

// Create a new HtmxFsManager instance
func New() *HtmxFsManager {
	staticCounter++
	return &HtmxFsManager{
		staticCounter: staticCounter,
	}
}

func (fm *HtmxFsManager) GetStaticFiles() []fs.FS {
	return fm.staticFiles
}

func (fm *HtmxFsManager) GetStaticPrefix() string {
	switch fm.staticCounter {
	case 0:
		panic("HtmxFsManager not initialized properly")
	case 1:
		return "/static"
	default:
		return fmt.Sprintf("/static-%d", fm.staticCounter-1)
	}
}

// Set a fallback HtmxFsManager to be used when a file is not found in the current manager
func (fm *HtmxFsManager) SetFallback(fallback *HtmxFsManager) *HtmxFsManager {
	if fm.fallbackHfm != fallback {
		fm.fallbackHfm = fallback
	}
	return fm
}

// Set script injection for layout rendering
func (fm *HtmxFsManager) SetScriptInjection(si *ScriptInjection) *HtmxFsManager {
	fm.scriptInjection = si
	return fm
}

// Combine two HtmxFsManager instances by appending their file systems
func (fm *HtmxFsManager) Merge(other *HtmxFsManager) *HtmxFsManager {
	if other == nil {
		return fm
	}

	merged := &HtmxFsManager{
		layoutsFiles:    append(fm.layoutsFiles, other.layoutsFiles...),
		pagesFiles:      append(fm.pagesFiles, other.pagesFiles...),
		staticFiles:     append(fm.staticFiles, other.staticFiles...),
		scriptInjection: fm.scriptInjection, // Use current script injection, not merged
	}

	// Use other's script injection if current is nil
	if merged.scriptInjection == nil && other.scriptInjection != nil {
		merged.scriptInjection = other.scriptInjection
	}

	return merged
}

// Add file systems for Layouts to the manager
// Optional dir parameter to specify subdirectory within the fs
func (fm *HtmxFsManager) AddLayoutFiles(fsFiles fs.FS, dir ...string) *HtmxFsManager {
	if len(dir) > 0 {
		var err error
		fsFiles, err = fs.Sub(fsFiles, dir[0])
		if err != nil {
			return fm
		}
	}
	fm.layoutsFiles = append(fm.layoutsFiles, fsFiles)
	return fm
}

// Add file systems for Pages to the manager
// Optional dir parameter to specify subdirectory within the fs
func (fm *HtmxFsManager) AddPageFiles(fsFiles fs.FS, dir ...string) *HtmxFsManager {
	if len(dir) > 0 {
		var err error
		fsFiles, err = fs.Sub(fsFiles, dir[0])
		if err != nil {
			return fm
		}
	}
	fm.pagesFiles = append(fm.pagesFiles, fsFiles)
	return fm
}

// Add file systems for Static files to the manager
// Optional dir parameter to specify subdirectory within the fs
func (fm *HtmxFsManager) AddStaticFiles(fsFiles fs.FS, dir ...string) *HtmxFsManager {
	if len(dir) > 0 {
		var err error
		fsFiles, err = fs.Sub(fsFiles, dir[0])
		if err != nil {
			return fm
		}
	}
	fm.staticFiles = append(fm.staticFiles, fsFiles)
	return fm
}

// Reads a layout file by searching through the registered layout filesystems
func (fm *HtmxFsManager) ReadLayoutFile(name string) ([]byte, error) {
	var err error
	for _, fSys := range fm.layoutsFiles {
		var data []byte
		data, err = fs.ReadFile(fSys, name)
		if err == nil {
			return data, nil
		}
	}
	if fm.fallbackHfm != nil {
		return fm.fallbackHfm.ReadLayoutFile(name)
	}
	return nil, err
}

// Reads a page file by searching through the registered page filesystems
func (fm *HtmxFsManager) ReadPageFile(name string) ([]byte, error) {
	var err error
	for _, fSys := range fm.pagesFiles {
		var data []byte
		data, err = fs.ReadFile(fSys, name)
		if err == nil {
			return data, nil
		}
	}
	if fm.fallbackHfm != nil {
		return fm.fallbackHfm.ReadPageFile(name)
	}
	return nil, err
}

// Reads a static file by searching through the registered static filesystems
func (fm *HtmxFsManager) ReadStaticFile(name string) ([]byte, error) {
	var err error
	for _, fSys := range fm.staticFiles {
		var data []byte
		data, err = fs.ReadFile(fSys, name)
		if err == nil {
			return data, nil
		}
	}
	if fm.fallbackHfm != nil {
		return fm.fallbackHfm.ReadStaticFile(name)
	}
	return nil, err
}

// getDirectiveContent extracts directive content from HTML comments
func getDirectiveContent(html string, regexFind *regexp.Regexp, defaultContent string) string {
	matches := regexFind.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}
	return defaultContent
}

// enhanceDataWithStaticPaths automatically injects static path variables into template data
func (fm *HtmxFsManager) enhanceDataWithStaticPaths(prefix string, data any) map[string]any {
	var result map[string]any

	// Convert data to map[string]any
	switch v := data.(type) {
	case map[string]any:
		result = v
	case nil:
		result = make(map[string]any)
	default:
		// For other types, wrap in a map under "data" key
		result = map[string]any{"data": v}
	}

	result["StaticPath"] = prefix
	result["StaticCSS"] = prefix + "/css"
	result["StaticJS"] = prefix + "/js"
	result["StaticImg"] = prefix + "/images"

	return result
}

func replaceOrGetInsertElement(html string, regexFind *regexp.Regexp, newElement string) string {
	if regexFind.MatchString(html) {
		// Replace existing element
		return regexFind.ReplaceAllString(html, newElement)
	}
	return newElement + "\n"
}

// Handles both partial and full page rendering based on request headers
func (fm *HtmxFsManager) RenderPageWithRequest(r *http.Request, w http.ResponseWriter,
	pagePath string, data any, prefix string,
	pageTitle, pageDesc string) (string, error) {

	// 1. Normalize page path (add .html if not present)
	normalizedPagePath := pagePath
	if !strings.HasSuffix(normalizedPagePath, ".html") {
		normalizedPagePath += ".html"
	}

	// 2. Read page file
	pageContent, err := fm.ReadPageFile(normalizedPagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read page file %s: %w", normalizedPagePath, err)
	}
	strPageContent := string(pageContent)

	// 3. Extract layout name from page content
	layoutName := getDirectiveContent(strPageContent, layoutRegex, "base.html")

	// 4. Set pageTitle if not provided
	if pageTitle == "" {
		pageTitle = getDirectiveContent(strPageContent, titleRegex, "Lokstra App")
	}

	// 5. Set currentLayout if not provided
	var currentLayout string = r.Header.Get("LS-Layout")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("LS-Layout", layoutName)
	w.Header().Set("Vary", "HX-Request")

	// 6. Check if partial render (HTMX) or full render (normal)
	isPartial := false
	if r != nil {
		isPartial = r.Header.Get("HX-Request") == "true" &&
			currentLayout == layoutName
	}

	tmpl := template.New("")

	// Enhance data with automatic static path injection
	enhancedData := fm.enhanceDataWithStaticPaths(prefix, data)

	if isPartial {
		w.Header().Set("LS-Title", pageTitle)
		// Partial render → only page template
		_, err = tmpl.Parse(strPageContent)
		if err != nil {
			return "", fmt.Errorf("failed to parse page template: %w", err)
		}

		// Execute page template only
		var buf strings.Builder
		err = tmpl.Execute(&buf, enhancedData)
		if err != nil {
			return "", fmt.Errorf("failed to execute page template: %w", err)
		}

		return buf.String(), nil
	}

	// 7. Full render → load layout + page
	layoutPath := layoutName
	if !strings.HasSuffix(layoutPath, ".html") {
		layoutPath += ".html"
	}

	layoutContent, err := fm.ReadLayoutFile(layoutPath)
	if err != nil {
		return "", fmt.Errorf("failed to read layout file %s: %w", layoutPath, err)
	}
	strLayoutContent := string(layoutContent)

	// 8. Apply script injection if configured
	si := fm.scriptInjection
	if si == nil {
		si = NewDefaultScriptInjection(true)
	}
	strLayoutContent = si.LoadInjectionScripts(strLayoutContent)

	// 9. Inject head elements (title, description, layout meta)
	headInjection := replaceOrGetInsertElement(strLayoutContent, titleLayoutRegex,
		fmt.Sprintf(`<title>%s</title>`, pageTitle))

	if pageDesc != "" {
		headInjection += fmt.Sprintf(`<meta name="description" content="%s">`, pageDesc) + "\n"
	}
	headInjection += fmt.Sprintf(`<meta name="ls-layout" content="%s">`, layoutName) + "\n"

	strLayoutContent = strings.Replace(strLayoutContent, "<head>",
		"<head>\n"+headInjection, 1)

	// 10. Create template and parse layout first
	_, err = tmpl.Parse(strLayoutContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse layout template: %w", err)
	}

	// 11. Parse page template using tmpl.New("page")
	_, err = tmpl.New("page").Parse(strPageContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse page template: %w", err)
	}

	// 12. Execute template with data
	var buf strings.Builder
	err = tmpl.Execute(&buf, enhancedData)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// GetPageLayout extracts layout name from page content
func (fm *HtmxFsManager) GetPageLayout(pagePath string) (string, error) {
	normalizedPagePath := pagePath
	if !strings.HasSuffix(normalizedPagePath, ".html") {
		normalizedPagePath += ".html"
	}

	pageContent, err := fm.ReadPageFile(normalizedPagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read page file %s: %w", normalizedPagePath, err)
	}

	return getDirectiveContent(string(pageContent), layoutRegex, "base.html"), nil
}

// HasStaticFiles checks if this HtmxFsManager has any static files registered
func (fm *HtmxFsManager) HasStaticFiles() bool {
	return len(fm.staticFiles) > 0
}
