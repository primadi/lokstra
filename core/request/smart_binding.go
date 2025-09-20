package request

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/primadi/lokstra/common/json"
)

// BindBodySmart intelligently binds request body based on Content-Type header
// Supports both application/json and application/x-www-form-urlencoded
func (ctx *Context) BindBodySmart(v any) error {
	ctx.cacheRequestBody()
	if ctx.requestBodyErr != nil {
		return ctx.requestBodyErr
	}
	if len(ctx.rawRequestBody) == 0 {
		return nil // No body to bind
	}

	contentType := ctx.GetHeader("Content-Type")

	// Handle form-urlencoded content by delegating to bindFormURLEncoded
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		return ctx.bindFormURLEncoded(v)
	}

	// Default to JSON binding
	return jsonBodyDecoder.Unmarshal(ctx.rawRequestBody, v)
}

// BindAllSmart binds path, query, header parameters and body with smart content-type detection
func (ctx *Context) BindAllSmart(v any) error {
	// If v is pointer to map[string]any, perform map-merge binding
	t := reflect.TypeOf(v)
	if t != nil && t.Kind() == reflect.Pointer {
		elem := t.Elem()
		if elem.Kind() == reflect.Map && elem.Key().Kind() == reflect.String {
			// Prepare the map value
			rvMap := reflect.ValueOf(v).Elem()
			if !rvMap.IsValid() {
				return nil
			}
			if rvMap.IsNil() {
				rvMap.Set(reflect.MakeMap(rvMap.Type()))
			}

			// Merge query params
			query := ctx.Request.URL.Query()
			for k, vals := range query {
				if len(vals) > 1 {
					// store slice of strings
					arr := make([]any, len(vals))
					for i, vv := range vals {
						arr[i] = vv
					}
					rvMap.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(arr))
				} else if len(vals) == 1 {
					rvMap.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(vals[0]))
				}
			}

			// Merge headers (may override query)
			for k, vals := range ctx.Request.Header {
				if len(vals) > 1 {
					arr := make([]any, len(vals))
					for i, vv := range vals {
						arr[i] = vv
					}
					rvMap.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(arr))
				} else if len(vals) == 1 {
					rvMap.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(vals[0]))
				}
			}

			// Merge body (overrides previous keys) - reuse BindBodySmart for parsing
			if err := ctx.BindBodySmart(v); err != nil {
				return err
			}

			return nil
		}
	}

	// Default: struct-based binding (existing behavior)
	bindMeta := getOrBuildBindingMeta(reflect.TypeOf(v))
	rv := reflect.ValueOf(v).Elem()
	header := ctx.Request.Header
	query := ctx.Request.URL.Query()

	for _, fieldMeta := range bindMeta.Fields {
		switch fieldMeta.Tag {
		case "query":
			if err := ctx.bindQueryField(fieldMeta, rv, query); err != nil {
				return err
			}
		case "header":
			if err := ctx.bindHeaderField(fieldMeta, rv, header); err != nil {
				return err
			}
		default: //case "path":
			if err := ctx.bindPathField(fieldMeta, rv); err != nil {
				return err
			}
		}
	}

	if err := ctx.BindBodySmart(v); err != nil {
		return err
	}
	return nil
}

// bindFormURLEncoded binds URL-encoded form data to struct
func (ctx *Context) bindFormURLEncoded(v any) error {
	// Parse form data
	formData, err := url.ParseQuery(string(ctx.rawRequestBody))
	if err != nil {
		return err
	}

	// If v is pointer
	if t := reflect.TypeOf(v); t != nil && t.Kind() == reflect.Pointer {
		elem := t.Elem()
		// If v is pointer to map[string]any, fill the map directly
		if elem.Kind() == reflect.Map && elem.Key().Kind() == reflect.String {
			rvMap := reflect.ValueOf(v).Elem()
			if rvMap.IsNil() {
				rvMap.Set(reflect.MakeMap(rvMap.Type()))
			}
			for k, vals := range formData {
				if len(vals) > 1 {
					arr := make([]any, len(vals))
					for i, vv := range vals {
						arr[i] = vv
					}
					rvMap.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(arr))
				} else if len(vals) == 1 {
					rvMap.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(vals[0]))
				}
			}
			return nil
		}

		// If v is pointer to struct, marshal formData to JSON and unmarshal
		if elem.Kind() == reflect.Struct {
			// Convert formData to map[string]any
			m := make(map[string]any)
			for k, vals := range formData {
				if len(vals) > 1 {
					m[k] = vals
				} else if len(vals) == 1 {
					m[k] = vals[0]
				}
			}
			// Marshal to JSON
			b, err := json.Marshal(m)
			if err != nil {
				return err
			}
			// Unmarshal into v
			return jsonBodyDecoder.Unmarshal(b, v)
		}
	}

	// v is not pointer to struct or map
	return fmt.Errorf("bindFormURLEncoded: unsupported type %T", v)
}
