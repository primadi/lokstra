package core

import (
	"lokstra/common/permission"
	"lokstra/common/response/response_iface"
	"lokstra/core/request"
	"lokstra/core/router/listener"
	"lokstra/core/router/router_engine"
)

type OnResponseHook func(ctx *request.Context) *request.Context

type AppRuntimeSetting struct {
	// newHttpListenerFunc is the HTTP listener used by the application.
	newHttpListenerFunc func() listener.HttpListener
	// newRouterEngineFunc is the router engine used to handle HTTP requests.
	newRouterEngineFunc func() router_engine.RouterEngine

	// requestContextHelper is a function that takes a *RequestContext and returns a modified *RequestContext.
	// By default, it points to defaultContextHelper which returns the context unchanged.
	requestContextHelper func(*request.Context) *request.Context

	// responseFormatter is the formatter used to format responses.
	// It can be set to a custom formatter to change how responses are formatted.
	responseFormatter response_iface.ResponseFormatter
	// The default formatter is set to JSONFormatter in the init function.
	// responseTemplateFunc is a function that generates response templates.
	// It can be set to a custom function to change how templates are generated.
	responseTemplateFunc response_iface.TemplateFunc
	// onResponseHooks is a list of hooks that are called before the response is written.
	onResponseHooks []OnResponseHook
}

var globalRuntime = AppRuntimeSetting{
	newHttpListenerFunc:  listener.NewNetHttpListener, // Default to NetHttpListenerType
	newRouterEngineFunc:  router_engine.NewHTTPRouterEngine,
	requestContextHelper: defaultContextHelper,
	onResponseHooks:      []OnResponseHook{},
}

// GlobalRuntime is the global application runtime settings.
// It is initialized with default values and can be modified at runtime.
func GlobalRuntime() *AppRuntimeSetting {
	return &globalRuntime
}

// SetResponseFormatter sets the response formatter for the global runtime.
func (g *AppRuntimeSetting) SetResponseFormatter(formatter response_iface.ResponseFormatter) {
	if permission.GlobalAccessLocked() {
		panic("permission is locked, cannot modify settings")
	}
	g.responseFormatter = formatter
}

// SetNewRouterEngineFunc sets the function to create a new router engine.
func (g *AppRuntimeSetting) SetNewRouterEngineFunc(fn func() router_engine.RouterEngine) {
	if permission.GlobalAccessLocked() {
		panic("permission is locked, cannot modify settings")
	}
	g.newRouterEngineFunc = fn
}

// SetNewHttpListener sets the function to create a new HTTP listener.
func (g *AppRuntimeSetting) SetNewHttpListener(fn func() listener.HttpListener) {
	if permission.GlobalAccessLocked() {
		panic("permission is locked, cannot modify settings")
	}
	g.newHttpListenerFunc = fn
}

// SetResponseTemplateFunc sets the response template function for the global runtime.
func (g *AppRuntimeSetting) SetResponseTemplateFunc(fn response_iface.TemplateFunc) {
	if permission.GlobalAccessLocked() {
		panic("permission is locked, cannot modify settings")
	}
	g.responseTemplateFunc = fn
}

// AddOnResponseHook adds a new hook to be called before the response is written.
func (g *AppRuntimeSetting) AddOnResponseHook(hook OnResponseHook) {
	if permission.GlobalAccessLocked() {
		panic("permission is locked, cannot modify settings")
	}
	g.onResponseHooks = append(g.onResponseHooks, hook)
}

// defaultContextHelper is the default implementation of contextHelperBuilder.
// It performs no modifications and simply returns the original context.
func defaultContextHelper(ctx *request.Context) *request.Context {
	return ctx
}

// SetRequestContextHelper allows external packages to register a custom context helper function.
// This enables wrapping or extending *RequestContext for better DX (Developer Experience),
// such as adding helper methods or struct embedding.
//
// Important:
// HelperContext is intended only to improve DX when accessing RequestContext in handlers, middleware, etc.
// It should NOT change the actual behavior or lifecycle of RequestContext itself.
//
// For example, previously in a handler you might write:
//
//	func (ctx *core.RequestContext) error {
//		logger.Info(ctx.GetValue("tenant_id").(string))
//	 }
//
// After using a helper, you might define:
//
//	func (ctx *myApp.CustomRequest) error {
//	    logger.Info(ctx.GetTenantId())
//	}
//
// This allows richer access to context-related utilities, while keeping the core behavior intact.
func (g *AppRuntimeSetting) SetRequestContextHelper(builder func(*request.Context) *request.Context) {
	g.requestContextHelper = builder
}

func (g *AppRuntimeSetting) GetRequestContextHelper() func(*request.Context) *request.Context {
	return g.requestContextHelper
}

func (g *AppRuntimeSetting) GetResponseFormatter() response_iface.ResponseFormatter {
	return g.responseFormatter
}

func (g *AppRuntimeSetting) GetResponseTemplateFunc() response_iface.TemplateFunc {
	return g.responseTemplateFunc
}

func (g *AppRuntimeSetting) GetNewHttpListenerFunc() func() listener.HttpListener {
	return g.newHttpListenerFunc
}

func (g *AppRuntimeSetting) GetNewRouterEngineFunc() func() router_engine.RouterEngine {
	return g.newRouterEngineFunc
}
