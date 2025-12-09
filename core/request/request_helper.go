package request

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/common/validator"
	"github.com/primadi/lokstra/core/response/api_formatter"
)

var (
	jsonDecoder = jsoniter.ConfigCompatibleWithStandardLibrary
)

func unmarshalBody(data []byte, v any) error {
	err := jsonDecoder.Unmarshal(data, v)
	if err == nil {
		return nil
	}

	// Create a more user-friendly error message for JSON parsing errors
	errMsg := err.Error()

	// Try to detect common JSON parsing errors and provide better messages
	userFriendlyMsg := "Invalid JSON format"
	if strings.Contains(errMsg, "expect { or n, but found") {
		userFriendlyMsg = "Invalid data type in request body. Expected an object but received a different type."
	} else if strings.Contains(errMsg, "expects \" or n, but found") {
		userFriendlyMsg = "Invalid data type in request body. Expected a string but received a different type."
	} else if strings.Contains(errMsg, "readObjectStart") {
		userFriendlyMsg = "Invalid array element format. Expected object notation but received a different type."
	}

	// Wrap JSON parsing error as validation error for better error handling
	return &ValidationError{
		FieldErrors: []api_formatter.FieldError{
			{
				Field:   "body",
				Code:    "INVALID_JSON",
				Message: userFriendlyMsg,
			},
		},
	}
}

// RequestHelper contains helper methods for request handling
type RequestHelper struct {
	ctx *Context

	// Request body caching
	rawRequestBody []byte
	requestBodyErr error
}

func newRequestHelper(ctx *Context) *RequestHelper {
	return &RequestHelper{ctx: ctx}
}

// QueryParam retrieves a query parameter by name, returning defaultValue if not present
func (h *RequestHelper) QueryParam(name string, defaultValue string) string {
	v := h.ctx.R.URL.Query().Get(name)
	if v == "" {
		return defaultValue
	}
	return v
}

// FormParam retrieves a form parameter by name, returning defaultValue if not present
func (h *RequestHelper) FormParam(name string, defaultValue string) string {
	v := h.ctx.R.FormValue(name)
	if v == "" {
		return defaultValue
	}
	return v
}

// PathParam retrieves a path parameter by name, returning defaultValue if not present
func (h *RequestHelper) PathParam(name string, defaultValue string) string {
	v := h.ctx.R.PathValue(name)
	if v == "" {
		return defaultValue
	}
	return v
}

// HeaderParam retrieves a header parameter by name, returning defaultValue if not present
func (h *RequestHelper) HeaderParam(name string, defaultValue string) string {
	v := h.ctx.R.Header.Get(name)
	if v == "" {
		return defaultValue
	}
	return v
}

// Multiple value parameter methods

// QueryParams retrieves all query parameter values by name
func (h *RequestHelper) QueryParams(name string) []string {
	return h.ctx.R.URL.Query()[name]
}

// FormParams retrieves all form parameter values by name
func (h *RequestHelper) FormParams(name string) []string {
	if err := h.ctx.R.ParseForm(); err != nil {
		return nil
	}
	return h.ctx.R.Form[name]
}

// HeaderValues retrieves all header values by name
func (h *RequestHelper) HeaderValues(name string) []string {
	return h.ctx.R.Header.Values(name)
}

// AllQueryParams retrieves all query parameters as map
func (h *RequestHelper) AllQueryParams() map[string][]string {
	return h.ctx.R.URL.Query()
}

// AllHeaders retrieves all headers as map
func (h *RequestHelper) AllHeaders() map[string][]string {
	return h.ctx.R.Header
}

// RawRequestBody returns the cached request body
func (h *RequestHelper) RawRequestBody() ([]byte, error) {
	h.cacheRequestBody()
	return h.rawRequestBody, h.requestBodyErr
}

// cacheRequestBody caches the request body for reuse
func (h *RequestHelper) cacheRequestBody() {
	if h.rawRequestBody != nil || h.requestBodyErr != nil {
		return // already cached
	}

	if h.ctx.R.Body == nil {
		return
	}

	body, err := io.ReadAll(h.ctx.R.Body)
	if err != nil {
		h.requestBodyErr = err
	} else {
		h.rawRequestBody = body
	}
}

// Helper methods for binding fields (moved from Context)

func (h *RequestHelper) bindPathField(fieldMeta bindFieldMeta, rv reflect.Value) error {
	rawValue := h.PathParam(fieldMeta.Name, "")
	rawValues := []string{rawValue}
	return convertAndSetField(rv.FieldByIndex(fieldMeta.Index), rawValues,
		fieldMeta.IsSlice, fieldMeta.IsUnmarshalJSON)
}

func (h *RequestHelper) bindQueryField(fieldMeta bindFieldMeta, rv reflect.Value, query url.Values) error {
	field := rv.FieldByIndex(fieldMeta.Index)

	// Support array of struct {Key,Value} or {Field,Value}
	if fieldMeta.IsIndexedKeyValue {
		paramPrefix := fieldMeta.Name
		if err := parseIndexedParamValuesReflect(paramPrefix, query, field,
			fieldMeta.IndexKey, fieldMeta.IndexValue); err != nil {
			return err
		}
		return nil
	} else if fieldMeta.IsMap {
		paramPrefix := fieldMeta.Name + "["
		paramPrefixLen := len(paramPrefix)

		result := map[string]string{}

		for key, val := range query {
			if strings.HasPrefix(key, paramPrefix) && strings.HasSuffix(key, "]") {
				mapKey := key[paramPrefixLen : len(key)-1]
				if len(val) > 0 {
					result[mapKey] = val[0]
				}
			}
		}

		field.Set(reflect.ValueOf(result))
		return nil
	}

	// Normal slice
	values := query[fieldMeta.Name]
	var rawValues []string

	if fieldMeta.IsSlice {
		if len(values) > 1 {
			rawValues = values
		} else if len(values) == 1 {
			rawValues = splitCommaSeparated(values[0])
		} else {
			rawValues = nil
		}
	} else {
		if len(values) > 0 {
			rawValues = []string{values[0]}
		} else {
			rawValues = nil
		}
	}

	return convertAndSetField(field, rawValues, fieldMeta.IsSlice, fieldMeta.IsUnmarshalJSON)
}

func (h *RequestHelper) bindHeaderField(fieldMeta bindFieldMeta, rv reflect.Value, header http.Header) error {
	values := header.Values(fieldMeta.Name)
	if len(values) == 0 && !fieldMeta.IsSlice {
		return nil
	}

	rawValues := values
	if !fieldMeta.IsSlice && len(values) > 0 {
		rawValues = []string{values[0]}
	}

	return convertAndSetField(rv.FieldByIndex(fieldMeta.Index), rawValues,
		fieldMeta.IsSlice, fieldMeta.IsUnmarshalJSON)
}

// bindFormURLEncoded binds URL-encoded form data to struct
func (h *RequestHelper) bindFormURLEncoded(v any) error {
	// Parse form data
	formData, err := url.ParseQuery(string(h.rawRequestBody))
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
			return unmarshalBody(b, v)
		}
	}

	// v is not pointer to struct or map
	return fmt.Errorf("bindFormURLEncoded: unsupported type %T", v)
}

// Public binding methods

// BindPath binds path parameters to struct
func (h *RequestHelper) BindPath(v any) error {
	bm := getOrBuildBindMeta(reflect.TypeOf(v))
	rv := reflect.ValueOf(v).Elem()

	for _, fieldMeta := range bm.Fields {
		if fieldMeta.Tag != "path" {
			continue
		}

		if err := h.bindPathField(fieldMeta, rv); err != nil {
			return err
		}
	}

	// Validate after binding
	return h.validateStruct(v)
}

// BindQuery binds query parameters to struct
func (h *RequestHelper) BindQuery(v any) error {
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
			query := h.ctx.R.URL.Query()
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

			return nil
		}
	}

	// Default: struct-based binding (existing behavior)
	bm := getOrBuildBindMeta(reflect.TypeOf(v))
	rv := reflect.ValueOf(v).Elem()
	query := h.ctx.R.URL.Query()

	for _, fieldMeta := range bm.Fields {
		if fieldMeta.Tag != "query" {
			continue
		}

		if err := h.bindQueryField(fieldMeta, rv, query); err != nil {
			return err
		}
	}

	// Validate after binding
	return h.validateStruct(v)
}

// BindHeader binds header values to struct
func (h *RequestHelper) BindHeader(v any) error {
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

			// Merge headers (may override query)
			for k, vals := range h.ctx.R.Header {
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
	}

	// Default: struct-based binding (existing behavior)
	bm := getOrBuildBindMeta(reflect.TypeOf(v))
	rv := reflect.ValueOf(v).Elem()
	header := h.ctx.R.Header

	for _, fieldMeta := range bm.Fields {
		if fieldMeta.Tag != "header" {
			continue
		}

		if err := h.bindHeaderField(fieldMeta, rv, header); err != nil {
			return err
		}
	}

	// Validate after binding
	return h.validateStruct(v)
}

// BindBody binds request body to struct
func (h *RequestHelper) BindBody(v any) error {
	h.cacheRequestBody()
	if h.requestBodyErr != nil {
		return h.requestBodyErr
	}
	if len(h.rawRequestBody) == 0 {
		return nil // No body to bind
	}

	// Check if v is a struct with wildcard fields
	t := reflect.TypeOf(v)
	if t != nil && t.Kind() == reflect.Pointer && t.Elem().Kind() == reflect.Struct {
		bm := getOrBuildBindMeta(t)

		// Check if any field has wildcard binding
		hasWildcard := false
		var wildcardField *bindFieldMeta
		for i := range bm.Fields {
			if bm.Fields[i].IsWildcard && bm.Fields[i].Tag == "json" {
				hasWildcard = true
				wildcardField = &bm.Fields[i]
				break
			}
		}

		if hasWildcard {
			// Bind wildcard field - unmarshal entire body into the map field
			rv := reflect.ValueOf(v).Elem()
			mapField := rv.FieldByIndex(wildcardField.Index)

			// Ensure the field is a map type
			if mapField.Kind() == reflect.Map {
				// Unmarshal body directly into the map
				mapPtr := reflect.New(mapField.Type())
				if err := jsonDecoder.Unmarshal(h.rawRequestBody, mapPtr.Interface()); err != nil {
					return &ValidationError{
						FieldErrors: []api_formatter.FieldError{
							{
								Field:   wildcardField.Field.Name,
								Code:    "INVALID_JSON",
								Message: "Failed to parse body as map: " + err.Error(),
							},
						},
					}
				}

				mapField.Set(mapPtr.Elem())
			}

			// Also bind other json/body fields normally
			if err := unmarshalBody(h.rawRequestBody, v); err != nil {
				return err
			}

			// Validate after binding
			return h.validateStruct(v)
		}
	}

	// Normal struct binding (no wildcard)
	if err := unmarshalBody(h.rawRequestBody, v); err != nil {
		return err
	}

	// Validate after binding
	return h.validateStruct(v)
}

// binds all request data (path, query, header, body) to struct
func (h *RequestHelper) BindAll(v any) error {
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
			query := h.ctx.R.URL.Query()
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
			for k, vals := range h.ctx.R.Header {
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

			// Merge body (overrides previous keys) - reuse BindBody for parsing
			if err := h.BindBody(v); err != nil {
				return err
			}

			return nil
		}
	}

	// Default: struct-based binding (existing behavior)
	bm := getOrBuildBindMeta(reflect.TypeOf(v))
	rv := reflect.ValueOf(v).Elem()
	header := h.ctx.R.Header
	query := h.ctx.R.URL.Query()

	for _, fieldMeta := range bm.Fields {
		// Skip wildcard fields - they will be handled by BindBody
		if fieldMeta.IsWildcard {
			continue
		}

		switch fieldMeta.Tag {
		case "query":
			if err := h.bindQueryField(fieldMeta, rv, query); err != nil {
				return err
			}
		case "header":
			if err := h.bindHeaderField(fieldMeta, rv, header); err != nil {
				return err
			}
		case "path":
			if err := h.bindPathField(fieldMeta, rv); err != nil {
				return err
			}
		// Skip json fields - they will be handled by BindBody
		case "json":
			continue
		}
	}

	if err := h.BindBody(v); err != nil {
		return err
	}

	// Validate after binding
	return h.validateStruct(v)
}

// binds request body with auto content-type detection
func (h *RequestHelper) BindBodyAuto(v any) error {
	h.cacheRequestBody()
	if h.requestBodyErr != nil {
		return h.requestBodyErr
	}
	if len(h.rawRequestBody) == 0 {
		return nil // No body to bind
	}

	contentType := h.ctx.R.Header.Get("Content-Type")

	// Handle form-urlencoded content by delegating to bindFormURLEncoded
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		return h.bindFormURLEncoded(v)
	}

	// Default to JSON binding
	return unmarshalBody(h.rawRequestBody, v)
}

// binds all request data with auto content-type detection
func (h *RequestHelper) BindAllAuto(v any) error {
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
			query := h.ctx.R.URL.Query()
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
			for k, vals := range h.ctx.R.Header {
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
			if err := h.BindBodyAuto(v); err != nil {
				return err
			}

			return nil
		}
	}

	// Default: struct-based binding (existing behavior)
	bindMeta := getOrBuildBindMeta(reflect.TypeOf(v))
	rv := reflect.ValueOf(v).Elem()
	header := h.ctx.R.Header
	query := h.ctx.R.URL.Query()

	for _, fieldMeta := range bindMeta.Fields {
		switch fieldMeta.Tag {
		case "query":
			if err := h.bindQueryField(fieldMeta, rv, query); err != nil {
				return err
			}
		case "header":
			if err := h.bindHeaderField(fieldMeta, rv, header); err != nil {
				return err
			}
		default: //case "path":
			if err := h.bindPathField(fieldMeta, rv); err != nil {
				return err
			}
		}
	}

	if err := h.BindBodyAuto(v); err != nil {
		return err
	}

	// Validate after binding
	return h.validateStruct(v)
}

// validateStruct validates a struct using validator.ValidateStruct
// Returns ValidationError if validation fails
func (h *RequestHelper) validateStruct(v any) error {
	fieldErrors, err := validator.ValidateStruct(v)
	if err != nil {
		// System error
		return err
	}

	if len(fieldErrors) > 0 {
		// Return ValidationError with formatted message
		return &ValidationError{
			FieldErrors: fieldErrors,
		}
	}

	return nil
}

// ValidationError represents validation errors from struct validation
type ValidationError struct {
	FieldErrors []api_formatter.FieldError
}

func (e *ValidationError) Error() string {
	if len(e.FieldErrors) == 0 {
		return "validation failed"
	}

	messages := make([]string, len(e.FieldErrors))
	for i, fe := range e.FieldErrors {
		messages[i] = fmt.Sprintf("%s: %s", fe.Field, fe.Message)
	}
	return strings.Join(messages, "; ")
}
