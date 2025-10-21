package router

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response"
)

var (
	typeOfContext      = reflect.TypeOf((*request.Context)(nil))
	typeOfError        = reflect.TypeOf((*error)(nil)).Elem()
	typeOfResponse     = reflect.TypeOf((*response.Response)(nil))
	typeOfApiHelper    = reflect.TypeOf((*response.ApiHelper)(nil))
	typeOfResponseVal  = reflect.TypeOf(response.Response{})
	typeOfApiHelperVal = reflect.TypeOf(response.ApiHelper{})
)

// handlerMetadata contains signature information extracted during registration
type handlerMetadata struct {
	hasContext       bool // Whether first parameter is *request.Context
	startParamIndex  int  // Index where non-Context parameters start (0 or 1)
	numIn            int  // Total number of input parameters
	numOut           int  // Total number of return values
	returnsResponse  bool // Whether first return is *response.Response or response.Response
	returnsApiHelper bool // Whether first return is *response.ApiHelper or response.ApiHelper
	isResponsePtr    bool // Whether returns *response.Response (vs response.Response)
	isApiHelperPtr   bool // Whether returns *response.ApiHelper (vs response.ApiHelper)
}

// paramExtractorFunc extracts a parameter from context
// Optimized: pathParamNames captured in closure, not passed per call
type paramExtractorFunc func(*request.Context) (reflect.Value, error)

// use reflection to adapt various handler signatures to request.HandlerFunc
// OPTIMIZATION: Pre-compiles metadata and extractors during registration
// Supports handler signatures:
//   - func() error
//   - func() (any, error)
//   - func(any) error
//   - func(any) (any, error)
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
			// Two return values: (data/response, error)
			errResult := results[1]

			// Check if error is non-nil - error always takes precedence
			if !errResult.IsNil() {
				return errResult.Interface().(error)
			}

			// No error, check first return value type
			firstResult := results[0]

			// Case 1: Returns *response.Response or response.Response
			if meta.returnsResponse {
				// Check for nil pointer (only if it's a pointer type)
				if meta.isResponsePtr && firstResult.IsNil() {
					// Nil response pointer - send default success
					return ctx.Api.Ok(nil)
				}

				var resp *response.Response
				if meta.isResponsePtr {
					resp = firstResult.Interface().(*response.Response)
				} else {
					// response.Response value - get address
					respVal := firstResult.Interface().(response.Response)
					resp = &respVal
				}

				// Use the Response directly by copying it to ctx.Resp
				*ctx.Resp = *resp
				return nil
			}

			// Case 2: Returns *response.ApiHelper or response.ApiHelper
			if meta.returnsApiHelper {
				// Check for nil pointer (only if it's a pointer type)
				if meta.isApiHelperPtr && firstResult.IsNil() {
					// Nil ApiHelper pointer - send default success
					return ctx.Api.Ok(nil)
				}

				var apiHelper *response.ApiHelper
				if meta.isApiHelperPtr {
					apiHelper = firstResult.Interface().(*response.ApiHelper)
				} else {
					// response.ApiHelper value - get address
					apiHelperVal := firstResult.Interface().(response.ApiHelper)
					apiHelper = &apiHelperVal
				}

				// Extract Response from ApiHelper and copy to ctx.Resp
				*ctx.Resp = *apiHelper.Resp()
				return nil
			}

			// Case 3: Regular data return - wrap in API response
			return ctx.Api.Ok(firstResult.Interface())
		}

		// Single return value
		firstResult := results[0]

		// Check if it's an error return
		if firstResult.Type().Implements(typeOfError) {
			// Only error return
			if !firstResult.IsNil() {
				return firstResult.Interface().(error)
			}
			// Success with no data - check if handler wrote response
			if meta.hasContext && ctx.Resp.WriterFunc != nil {
				return nil
			}
			// Send default success response
			return ctx.Api.Ok(nil)
		}

		// Single non-error return: data, *Response, or *ApiHelper

		// Case 1: Returns *response.Response or response.Response
		if meta.returnsResponse {
			// Check for nil pointer (only if it's a pointer type)
			if meta.isResponsePtr && firstResult.IsNil() {
				// Nil response pointer - send default success
				return ctx.Api.Ok(nil)
			}

			var resp *response.Response
			if meta.isResponsePtr {
				resp = firstResult.Interface().(*response.Response)
			} else {
				// response.Response value - get address
				respVal := firstResult.Interface().(response.Response)
				resp = &respVal
			}

			// Use the Response directly by copying it to ctx.Resp
			*ctx.Resp = *resp
			return nil
		}

		// Case 2: Returns *response.ApiHelper or response.ApiHelper
		if meta.returnsApiHelper {
			// Check for nil pointer (only if it's a pointer type)
			if meta.isApiHelperPtr && firstResult.IsNil() {
				// Nil ApiHelper pointer - send default success
				return ctx.Api.Ok(nil)
			}

			var apiHelper *response.ApiHelper
			if meta.isApiHelperPtr {
				apiHelper = firstResult.Interface().(*response.ApiHelper)
			} else {
				// response.ApiHelper value - get address
				apiHelperVal := firstResult.Interface().(response.ApiHelper)
				apiHelper = &apiHelperVal
			}

			// Extract Response from ApiHelper and copy to ctx.Resp
			*ctx.Resp = *apiHelper.Resp()
			return nil
		}

		// Case 3: Regular data return - wrap in API response
		return ctx.Api.Ok(firstResult.Interface())
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

	// Check if last return value is error (for numOut == 2)
	// If numOut == 1, it can be either error OR any (response/data)
	hasErrorReturn := false
	if numOut == 2 {
		// Two returns: must be (data, error)
		if !fnType.Out(1).Implements(typeOfError) {
			panic(invalidHandlerMsg(path))
		}
		hasErrorReturn = true
	} else {
		// One return: check if it's error
		if fnType.Out(0).Implements(typeOfError) {
			hasErrorReturn = true
		}
	}

	// Detect if first param is *request.Context
	hasContext := false
	startParamIndex := 0
	if numIn > 0 && fnType.In(0) == typeOfContext {
		hasContext = true
		startParamIndex = 1
	}

	// Detect return type
	returnsResponse := false
	returnsApiHelper := false
	isResponsePtr := false
	isApiHelperPtr := false

	// Check first return value (or only return value if numOut == 1)
	if numOut > 0 && !hasErrorReturn {
		// numOut == 1 and not error, so it's a data/response return
		firstReturnType := fnType.Out(0)

		switch firstReturnType {
		case typeOfResponse:
			returnsResponse = true
			isResponsePtr = true
		case typeOfResponseVal:
			returnsResponse = true
			isResponsePtr = false
		case typeOfApiHelper:
			returnsApiHelper = true
			isApiHelperPtr = true
		case typeOfApiHelperVal:
			returnsApiHelper = true
			isApiHelperPtr = false
		}
	} else if numOut == 2 {
		// numOut == 2: (data, error) - check first return
		firstReturnType := fnType.Out(0)

		switch firstReturnType {
		case typeOfResponse:
			returnsResponse = true
			isResponsePtr = true
		case typeOfResponseVal:
			returnsResponse = true
			isResponsePtr = false
		case typeOfApiHelper:
			returnsApiHelper = true
			isApiHelperPtr = true
		case typeOfApiHelperVal:
			returnsApiHelper = true
			isApiHelperPtr = false
		}
	}

	return &handlerMetadata{
		hasContext:       hasContext,
		startParamIndex:  startParamIndex,
		numIn:            numIn,
		numOut:           numOut,
		returnsResponse:  returnsResponse,
		returnsApiHelper: returnsApiHelper,
		isResponsePtr:    isResponsePtr,
		isApiHelperPtr:   isApiHelperPtr,
	}
}

// makeParameterExtractors creates optimized parameter extractors
// OPTIMIZATION: Only supports struct-based parameters (pointer or value)
// Direct path parameters (string, int) not supported - use struct with tags instead
func makeParameterExtractors(fnType reflect.Type, startParamIndex int) []paramExtractorFunc {
	numParams := fnType.NumIn() - startParamIndex
	extractors := make([]paramExtractorFunc, numParams)

	for i := range numParams {
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
		"  - func(*Context) data\n" +
		"  - func(*Context) (*Response, error)\n" +
		"  - func(*Context) *Response\n" +
		"  - func(*Context) (*ApiHelper, error)\n" +
		"  - func(*Context) *ApiHelper\n" +
		"  - func(*Context, *Struct) error\n" +
		"  - func(*Context, *Struct) (data, error)\n" +
		"  - func(*Context, *Struct) data\n" +
		"  - func(*Context, *Struct) (*Response, error)\n" +
		"  - func(*Context, *Struct) *Response\n" +
		"  - func(*Context, *Struct) (*ApiHelper, error)\n" +
		"  - func(*Context, *Struct) *ApiHelper\n" +
		"  - func(*Struct) error\n" +
		"  - func(*Struct) (data, error)\n" +
		"  - func(*Struct) data\n" +
		"  - func(*Struct) (*Response, error)\n" +
		"  - func(*Struct) *Response\n" +
		"  - func(*Struct) (*ApiHelper, error)\n" +
		"  - func(*Struct) *ApiHelper\n" +
		"  - func() error\n" +
		"  - func() (data, error)\n" +
		"  - func() data\n" +
		"  - func() (*Response, error)\n" +
		"  - func() *Response\n" +
		"  - func() (*ApiHelper, error)\n" +
		"  - func() *ApiHelper\n" +
		"  - request.HandlerFunc\n" +
		"  - http.HandlerFunc\n" +
		"  - http.Handler\n" +
		"Note: Direct path parameters (string, int) not supported. Use struct with 'path' tags.\n" +
		"Note: Handlers can return data/Response/ApiHelper with or without error.\n" +
		"Note: *Response and *ApiHelper returns allow full control over response (status, headers, body)."
}

// adaptHandler converts various handler types to request.HandlerFunc.
// OPTIMIZATION: Fast paths for common signatures avoid reflection overhead.
// Performance tiers:
//   - Tier 0 (0ns): Direct function return, zero wrapper
//   - Tier 1 (~10ns): Lightweight wrapper, no reflection
//   - Tier 2 (~80ns): Smart adapter with reflection (fallback)
func adaptHandler(path string, h any) request.HandlerFunc {
	switch v := h.(type) {
	// ========================================================================
	// TIER 0: ZERO-COST HANDLERS (Direct return, no wrapper)
	// ========================================================================
	case func(*request.Context) error:
		return v // Direct function, no wrapper needed
	case request.HandlerFunc:
		return v // Already the right type

	// ========================================================================
	// TIER 1: FAST PATH - Common Patterns with *Context (No Reflection)
	// These cover ~80% of real-world use cases
	// ========================================================================

	// Pattern: func(*Context) (any, error)
	// Most common production pattern
	case func(*request.Context) (any, error):
		return func(c *request.Context) error {
			data, err := v(c)
			if err != nil {
				return err
			}
			return c.Api.Ok(data)
		}

	// Pattern: func(*Context) (*Response, error)
	// For handlers needing full response control
	case func(*request.Context) (*response.Response, error):
		return func(c *request.Context) error {
			resp, err := v(c)
			if err != nil {
				return err
			}
			if resp == nil {
				return c.Api.Ok(nil)
			}
			*c.Resp = *resp
			return nil
		}

	// Pattern: func(*Context) (*ApiHelper, error)
	// For REST API handlers with helper methods
	case func(*request.Context) (*response.ApiHelper, error):
		return func(c *request.Context) error {
			api, err := v(c)
			if err != nil {
				return err
			}
			if api == nil {
				return c.Api.Ok(nil)
			}
			*c.Resp = *api.Resp()
			return nil
		}

	// Pattern: func(*Context) any
	// Simple handlers without error (mock/test)
	case func(*request.Context) any:
		return func(c *request.Context) error {
			data := v(c)
			return c.Api.Ok(data)
		}

	// Pattern: func(*Context) *Response
	// Full control handlers without error
	case func(*request.Context) *response.Response:
		return func(c *request.Context) error {
			resp := v(c)
			if resp == nil {
				return c.Api.Ok(nil)
			}
			*c.Resp = *resp
			return nil
		}

	// Pattern: func(*Context) *ApiHelper
	// API helpers without error
	case func(*request.Context) *response.ApiHelper:
		return func(c *request.Context) error {
			api := v(c)
			if api == nil {
				return c.Api.Ok(nil)
			}
			*c.Resp = *api.Resp()
			return nil
		}

	// Pattern: func(*Context) (Response, error) - value return
	// Less common but supported for flexibility
	case func(*request.Context) (response.Response, error):
		return func(c *request.Context) error {
			resp, err := v(c)
			if err != nil {
				return err
			}
			*c.Resp = resp
			return nil
		}

	// Pattern: func(*Context) (ApiHelper, error) - value return
	case func(*request.Context) (response.ApiHelper, error):
		return func(c *request.Context) error {
			api, err := v(c)
			if err != nil {
				return err
			}
			*c.Resp = *api.Resp()
			return nil
		}

	// Pattern: func(*Context) Response - value return, no error
	case func(*request.Context) response.Response:
		return func(c *request.Context) error {
			resp := v(c)
			*c.Resp = resp
			return nil
		}

	// Pattern: func(*Context) ApiHelper - value return, no error
	case func(*request.Context) response.ApiHelper:
		return func(c *request.Context) error {
			api := v(c)
			*c.Resp = *api.Resp()
			return nil
		}

	// ========================================================================
	// TIER 1: FAST PATH - Simple handlers without Context
	// For stateless/mock handlers
	// ========================================================================

	// Pattern: func() (any, error)
	case func() (any, error):
		return func(c *request.Context) error {
			data, err := v()
			if err != nil {
				return err
			}
			return c.Api.Ok(data)
		}

	// Pattern: func() (*Response, error)
	case func() (*response.Response, error):
		return func(c *request.Context) error {
			resp, err := v()
			if err != nil {
				return err
			}
			if resp == nil {
				return c.Api.Ok(nil)
			}
			*c.Resp = *resp
			return nil
		}

	// Pattern: func() (*ApiHelper, error)
	case func() (*response.ApiHelper, error):
		return func(c *request.Context) error {
			api, err := v()
			if err != nil {
				return err
			}
			if api == nil {
				return c.Api.Ok(nil)
			}
			*c.Resp = *api.Resp()
			return nil
		}

	// Pattern: func() any
	case func() any:
		return func(c *request.Context) error {
			data := v()
			return c.Api.Ok(data)
		}

	// Pattern: func() *Response
	case func() *response.Response:
		return func(c *request.Context) error {
			resp := v()
			if resp == nil {
				return c.Api.Ok(nil)
			}
			*c.Resp = *resp
			return nil
		}

	// Pattern: func() *ApiHelper
	case func() *response.ApiHelper:
		return func(c *request.Context) error {
			api := v()
			if api == nil {
				return c.Api.Ok(nil)
			}
			*c.Resp = *api.Resp()
			return nil
		}

	// Pattern: func() error
	case func() error:
		return func(c *request.Context) error {
			return v()
		}

	// ========================================================================
	// TIER 1: COMPATIBILITY - Standard HTTP handlers
	// ========================================================================
	case http.HandlerFunc:
		return func(c *request.Context) error {
			v(c.W, c.R)
			return nil
		}

	case func(http.ResponseWriter, *http.Request):
		return func(c *request.Context) error {
			v(c.W, c.R)
			return nil
		}

	case http.Handler:
		return func(c *request.Context) error {
			v.ServeHTTP(c.W, c.R)
			return nil
		}

	// ========================================================================
	// TIER 2: SMART ADAPTER (Reflection-based fallback)
	// Handles:
	//   - func(*Struct) (any, error) - struct parameter binding
	//   - func(*Context, *Struct) (any, error) - context + struct
	//   - All other complex combinations
	// ========================================================================
	case string:
		// Middleware by name (future feature)
		return nil

	default:
		// Fallback to reflection-based adapter for complex signatures
		return adaptSmart(path, v)
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
	if p == "" || p == "/" {
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
