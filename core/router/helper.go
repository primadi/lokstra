package router

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response/api_formatter"
)

var (
	typeOfContext = reflect.TypeOf((*request.Context)(nil))
	typeOfError   = reflect.TypeOf((*error)(nil)).Elem()
)

func adaptSmart(path string, v any) request.HandlerFunc {
	fnVal := reflect.ValueOf(v)
	fnType := fnVal.Type()

	if fnType.Kind() == reflect.Func &&
		fnType.NumIn() == 2 &&
		fnType.NumOut() == 1 &&
		fnType.In(0) == typeOfContext &&
		fnType.Out(0) == typeOfError &&
		fnType.In(1).Kind() == reflect.Ptr &&
		fnType.In(1).Elem().Kind() == reflect.Struct {

		paramType := fnType.In(1)

		return func(ctx *request.Context) error {
			paramPtr := reflect.New(paramType.Elem()).Interface()
			if err := ctx.Req.BindAllSmart(paramPtr); err != nil {
				// return ctx.Resp.WithStatus(400).Json(map[string]string{"error": err.Error()})
				return ctx.Api.ValidationError("Binding failed", []api_formatter.FieldError{
					{
						Field:   "request",
						Code:    "BIND_ERROR",
						Message: err.Error(),
					},
				})
			}
			out := fnVal.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(paramPtr)})
			if !out[0].IsNil() {
				return out[0].Interface().(error)
			}
			return nil
		}
	}
	msg := "Invalid handler type for path [" + path +
		"], it must be request.HandlerFunc, http.HandlerFunc, http.Handler, or func(*Context, *T) error"
	fmt.Println(msg)
	panic(msg)
}

// adaptHandler converts various handler types to request.HandlerFunc.
// TODO: Consider dual-path architecture for pure http.Handler chains in future versions.
func adaptHandler(path string, h any) request.HandlerFunc {
	switch v := h.(type) {
	case func(*request.Context) error:
		return v // Zero-cost for Lokstra handlers
	case request.HandlerFunc:
		return v // Zero-cost for Lokstra handlers
	case http.HandlerFunc:
		// Lightweight wrapper for standard handlers
		return func(c *request.Context) error {
			v(c.W, c.R)
			return nil
		}
	case func(http.ResponseWriter, *http.Request):
		// Lightweight wrapper for standard handlers
		return func(c *request.Context) error {
			v(c.W, c.R)
			return nil
		}
	case http.Handler:
		// Lightweight wrapper for standard handlers
		return func(c *request.Context) error {
			v.ServeHTTP(c.W, c.R)
			return nil
		}
	default:
		return adaptSmart(path, v) // Smart binding for complex handlers
	}
}

func adaptMiddlewares(mw []any) []request.HandlerFunc {
	var adapted []request.HandlerFunc
	for _, m := range mw {
		adapted = append(adapted, adaptHandler("middleware", m))
	}
	return adapted
}

func cleanPath(p string) string {
	p = strings.Trim(p, "/")
	if p == "" {
		return ""
	}
	return "/" + p
}

func cleanPrefix(p string) string {
	p = strings.Trim(p, "/")
	if p == "" {
		return "/"
	}
	// For Go 1.22+ ServeMux prefix patterns, use {path...} wildcard
	return "/" + p + "/{path...}"
}

func normalizeGroupName(childName, childPath string) string {
	if len(childName) == 0 {
		childName = strings.ReplaceAll(strings.Trim(childPath, "/"), "/", ".")
	}
	return childName
}
