package router

import (
	"net/http"

	"github.com/primadi/lokstra/common/iface"
	"github.com/primadi/lokstra/common/meta"

	"github.com/valyala/fasthttp"
)

type Router interface {
	Prefix() string

	Use(any) Router
	Handle(method iface.HTTPMethod, path string, handler any, mw ...any) Router
	HandleOverrideMiddleware(method iface.HTTPMethod, path string, handler any, mw ...any) Router
	GET(path string, handler any, mw ...any) Router
	POST(path string, handler any, mw ...any) Router
	PUT(path string, handler any, mw ...any) Router
	PATCH(path string, handler any, mw ...any) Router
	DELETE(path string, handler any, mw ...any) Router

	WithOverrideMiddleware(enable bool) Router
	WithPrefix(prefix string) Router

	MountStatic(prefix string, folder http.Dir) Router
	MountSPA(prefix string, fallbackFile string) Router
	MountReverseProxy(prefix string, target string) Router

	Group(prefix string, mw ...any) Router
	GroupBlock(prefix string, fn func(gr Router)) Router

	RecurseAllHandler(callback func(rt *meta.RouteMeta))
	DumpRoutes()

	ServeHTTP(w http.ResponseWriter, r *http.Request)
	FastHttpHandler() fasthttp.RequestHandler
	OverrideMiddleware() Router
	GetMiddleware() []*meta.MiddlewareExecution
	LockMiddleware()

	GetMeta() *meta.RouterMeta
}

// func WrapParamHandler(handler any) RequestHandler {
// 	if h, ok := handler.(RequestHandler); ok {
// 		return h
// 	}

// 	val := reflect.ValueOf(handler)
// 	typ := val.Type()

// 	if typ.Kind() != reflect.Func {
// 		panic("handler must be a function")
// 	}

// 	if typ.NumIn() != 2 {
// 		panic("handler must have exactly 2 parameters")
// 	}

// 	if typ.In(0) != reflect.TypeOf(&RequestContext{}) {
// 		panic("first parameter must be *RequestContext")
// 	}

// 	paramType := typ.In(1)
// 	if paramType.Kind() != reflect.Ptr || paramType.Elem().Kind() != reflect.Struct {
// 		panic("second parameter must be pointer to a struct")
// 	}

// 	return func(ctx *RequestContext) error {
// 		paramPtr := reflect.New(paramType.Elem()).Interface()

// 		if err := ctx.BindAll(paramPtr); err != nil {
// 			return err
// 		}

// 		results := val.Call([]reflect.Value{
// 			reflect.ValueOf(ctx),
// 			reflect.ValueOf(paramPtr),
// 		})

// 		if len(results) != 1 {
// 			return errors.New("handler must return exactly one result of type error")
// 		}
// 		if err, ok := results[0].Interface().(error); ok {
// 			return ctx.FailBadRequest("invalid payload", err)
// 		}
// 		return nil
// 	}
// }

// func WrapGenericParamHandler[T any](handler func(*RequestContext, *T) error) RequestHandler {
// 	var zero T
// 	if reflect.TypeOf(zero).Kind() != reflect.Struct {
// 		panic("handler parameter T must be a struct")
// 	}

// 	return func(ctx *RequestContext) error {
// 		params := new(T)
// 		if err := ctx.BindAll(params); err != nil {
// 			return ctx.FailBadRequest("invalid payload", err)
// 		}
// 		return handler(ctx, params)
// 	}
// }
