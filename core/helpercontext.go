package core

// contextHelperBuilder is a function that takes a *RequestContext and returns a modified *RequestContext.
// By default, it points to defaultContextHelper which returns the context unchanged.
var contextHelperBuilder = defaultContextHelper

// defaultContextHelper is the default implementation of contextHelperBuilder.
// It performs no modifications and simply returns the original context.
func defaultContextHelper(ctx *RequestContext) *RequestContext {
	return ctx
}

// RegisterHelperContext allows external packages to register a custom context helper function.
// This enables wrapping or extending *RequestContext for better DX (Developer Experience),
// such as adding helper methods or struct embedding.
//
// Important:
// HelperContext is intended only to improve DX when accessing RequestContext in handlers, middleware, etc.
// It should NOT change the actual behavior or lifecycle of RequestContext itself.
//
// For example, previously in a handler you might write:
//
//     func (ctx *core.RequestContext) error { }
//
// After using a helper, you might define:
//
//     func (ctx *myApp.CustomRequest) error {
//         ctx.GetTenant()
//         ctx.Logger()
//         // etc.
//     }
//
// This allows richer access to context-related utilities, while keeping the core behavior intact.
func RegisterHelperContext(builder func(*RequestContext) *RequestContext) {
	contextHelperBuilder = builder
}
