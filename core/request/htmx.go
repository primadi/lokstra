package request

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/primadi/lokstra/common/htmx_fsmanager"
)

// Renders HTMX content
func (ctx *Context) HTMXString(html string, data map[string]any) error {
	if data == nil {
		return ctx.Response.WithHeader("Vary", "HX-Request").HTML(http.StatusOK, html)
	}

	tmpl, err := template.New("page").Parse(html)
	if err != nil {
		return fmt.Errorf("failed to parse layout template: %w", err)
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return ctx.Response.WithHeader("Vary", "HX-Request").
			HTML(http.StatusInternalServerError, "Failed to parse HTMX template string")
	}
	return ctx.Response.WithHeader("Vary", "HX-Request").HTML(http.StatusOK, buf.String())
}

// Renders HTMX FS Page content with status 200
func (ctx *Context) HTMXFSPage(pagePath string, data any, title string, description string) error {
	hfm := ctx.hfmContainer.GetHtmxFsManager()
	if hfm == nil {
		return ctx.Response.WithHeader("Vary", "HX-Request").
			HTML(http.StatusInternalServerError, "HTMX FS Manager not configured")
	}

	html, err := hfm.RenderPageWithRequest(ctx.Request, ctx.Writer,
		pagePath, data, hfm.GetStaticPrefix(), title, description)
	if err != nil {
		return ctx.HTML(http.StatusInternalServerError,
			"Failed to render HTMX FS page: "+err.Error())
	}

	return ctx.HTML(http.StatusOK, html)
}

// Renders HTMX Custom FS Page content with status 200
func (ctx *Context) HTMXCustomFSPage(pagePath string, data any,
	hfm *htmx_fsmanager.HtmxFsManager, title string, description string) error {
	if hfm == nil {
		return ctx.Response.WithHeader("Vary", "HX-Request").
			HTML(http.StatusInternalServerError, "HTMX FS Manager not configured")
	}
	hfm.SetFallback(ctx.hfmContainer.GetHtmxFsManager())

	html, err := hfm.RenderPageWithRequest(ctx.Request, ctx.Writer,
		pagePath, data, hfm.GetStaticPrefix(), title, description)
	if err != nil {
		return ctx.HTML(http.StatusInternalServerError, "Failed to render HTMX FS page")
	}

	return ctx.HTML(http.StatusOK, html)
}

func (ctx *Context) HTMXEvent(eventName, eventData any) error {
	eventValue := ""
	switch v := eventData.(type) {
	case string:
		eventValue = v
	case fmt.Stringer:
		eventValue = v.String()
	case map[string]any:
		parts := make([]string, 0, len(v))
		for k, val := range v {
			parts = append(parts, fmt.Sprintf(`"%s":"%v"`, k, val))
		}
		eventValue = fmt.Sprintf("{%s}", strings.Join(parts, ","))
	default:
		eventValue = fmt.Sprintf("%v", v)
	}
	return ctx.Response.WithHeader("Vary", "HX-Request").
		WithHeader("HX-Trigger", fmt.Sprintf(`{"%s":"%s"}`, eventName, eventValue)).
		OkNoContent()
}

func (ctx *Context) HTMXRedirect(url string) error {
	return ctx.Response.WithHeader("Vary", "HX-Request").
		WithHeader("HX-Redirect", url).
		OkNoContent()
}

func (ctx *Context) HTMXRefresh() error {
	return ctx.Response.WithHeader("Vary", "HX-Request").
		WithHeader("HX-Refresh", "true").
		OkNoContent()
}
