package router

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/primadi/lokstra/core/request"
)

var (
	typeOfContext = reflect.TypeOf((*request.Context)(nil))
	typeOfError   = reflect.TypeOf((*error)(nil)).Elem()
)

// handlerMetadata contains signature information extracted during registration
type handlerMetadata struct {
	hasContext      bool // Whether first parameter is *request.Context
	startParamIndex int  // Index where non-Context parameters start (0 or 1)
	numIn           int  // Total number of input parameters
	numOut          int  // Total number of return values
}

// paramExtractorFunc extracts a parameter from context
// Optimized: pathParamNames captured in closure, not passed per call
type paramExtractorFunc func(*request.Context) (reflect.Value, error)

func adaptSmart(path string, v any) request.HandlerFunc {
	fnVal := reflect.ValueOf(v)
	fnType := fnVal.Type()

	// Must be a function
	if fnType.Kind() != reflect.Func {
		panic(invalidHandlerMsg(path))
	}

	// Build metadata (called once during registration - performance doesn't matter)
	meta := buildHandlerMetadata(fnType, path)

	// Build parameter extractors
	extractors := makeParameterExtractors(fnType, meta.startParamIndex)

	// Pre-allocate args slice with exact capacity needed
	argsCapacity := meta.numIn

	// Return the optimized handler (THIS is called per request!)
	return func(ctx *request.Context) error {
		// OPTIMIZATION: Pre-allocated slice with exact capacity
		args := make([]reflect.Value, 0, argsCapacity)

		// Add context if needed
		// NOTE: reflect.ValueOf(ctx) is unavoidable but very fast
		if meta.hasContext {
			args = append(args, reflect.ValueOf(ctx))
		}

		// Extract remaining parameters using pre-compiled extractors
		// OPTIMIZATION: Extractors already have pathParamNames captured
		for _, extractor := range extractors {
			arg, err := extractor(ctx)
			if err != nil {
				// Return binding/validation error immediately
				return err
			}
			args = append(args, arg)
		}

		// Call the function
		// NOTE: fnVal.Call() is unavoidable reflection cost
		results := fnVal.Call(args)

		// Handle return values
		// OPTIMIZATION: Branch prediction helps here (most common case first)
		if meta.numOut == 2 {
			// Two return values: (data, error) - most common case
			if !results[1].IsNil() {
				return results[1].Interface().(error)
			}
			// Success - wrap data in response
			return ctx.Api.Ok(results[0].Interface())
		}

		// Only error return
		if !results[0].IsNil() {
			return results[0].Interface().(error)
		}
		// Success with no data - check if handler wrote response
		if meta.hasContext && ctx.Resp.WriterFunc != nil {
			return nil
		}
		// Send default success response
		return ctx.Api.Ok(nil)
	}
}

// buildHandlerMetadata analyzes function signature and builds parameter extractors
// buildHandlerMetadata extracts metadata about handler function signature
// OPTIMIZATION: Only builds metadata, doesn't create extractors (they need pathParamNames)
func buildHandlerMetadata(fnType reflect.Type, path string) *handlerMetadata {
	numIn := fnType.NumIn()
	numOut := fnType.NumOut()

	// Validate output
	if numOut < 1 || numOut > 2 {
		panic(invalidHandlerMsg(path))
	}

	// Last return value must be error
	if !fnType.Out(numOut - 1).Implements(typeOfError) {
		panic(invalidHandlerMsg(path))
	}

	// Detect if first param is *request.Context
	hasContext := false
	startParamIndex := 0
	if numIn > 0 && fnType.In(0) == typeOfContext {
		hasContext = true
		startParamIndex = 1
	}

	return &handlerMetadata{
		hasContext:      hasContext,
		startParamIndex: startParamIndex,
		numIn:           numIn,
		numOut:          numOut,
	}
}

// makeParameterExtractors creates optimized parameter extractors
// OPTIMIZATION: Only supports struct-based parameters (pointer or value)
// Direct path parameters (string, int) not supported - use struct with tags instead
func makeParameterExtractors(fnType reflect.Type, startParamIndex int) []paramExtractorFunc {
	numParams := fnType.NumIn() - startParamIndex
	extractors := make([]paramExtractorFunc, numParams)

	for i := 0; i < numParams; i++ {
		paramType := fnType.In(startParamIndex + i)

		if paramType.Kind() == reflect.Pointer && paramType.Elem().Kind() == reflect.Struct {
			// Struct pointer - use BindAll
			elemType := paramType.Elem()
			extractors[i] = func(ctx *request.Context) (reflect.Value, error) {
				paramPtr := reflect.New(elemType)
				if err := ctx.Req.BindAll(paramPtr.Interface()); err != nil {
					return reflect.Value{}, err
				}
				return paramPtr, nil
			}
		} else if paramType.Kind() == reflect.Struct {
			// Struct value - use BindAll
			extractors[i] = func(ctx *request.Context) (reflect.Value, error) {
				paramPtr := reflect.New(paramType)
				if err := ctx.Req.BindAll(paramPtr.Interface()); err != nil {
					return reflect.Value{}, err
				}
				return paramPtr.Elem(), nil
			}
		} else {
			// Only struct-based parameters are supported
			panic(fmt.Sprintf("Parameter type %v not supported. Use struct with tags instead.", paramType))
		}
	}

	return extractors
}

func invalidHandlerMsg(path string) string {
	return "Invalid handler type for path [" + path + "]. Supported signatures:\n" +
		"  - func(*Context) error\n" +
		"  - func(*Context) (data, error)\n" +
		"  - func(*Context, *Struct) error\n" +
		"  - func(*Context, *Struct) (data, error)\n" +
		"  - func(*Struct) error\n" +
		"  - func(*Struct) (data, error)\n" +
		"  - request.HandlerFunc\n" +
		"  - http.HandlerFunc\n" +
		"  - http.Handler\n" +
		"Note: Direct path parameters (string, int) not supported. Use struct with 'path' tags."
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
	case string:
		// return lokstra_registry.CreateMiddleware(v)
		return nil
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
