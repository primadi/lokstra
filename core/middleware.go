package core

// ComposeMiddleware chains middleware and final handler into one RequestHandler
// Example: ComposeMiddleware([mw1, mw2, mw3], handler) => finalHandler = mw1(mw2(mw3(handler)))
func ComposeMiddleware(mw []MiddlewareHandler, finalHandler RequestHandler) RequestHandler {
	handler := finalHandler
	for i := len(mw) - 1; i >= 0; i-- {
		handler = mw[i](handler)
	}
	return handler
}
