package request

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

var formBodyDecoder = jsoniter.Config{TagKey: "form"}.Froze()

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

	contentType := ctx.Request.Header.Get("Content-Type")

	// Handle form-urlencoded content
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		return ctx.bindFormURLEncoded(v)
	}

	// Default to JSON binding (for application/json or unspecified)
	return jsonBodyDecoder.Unmarshal(ctx.rawRequestBody, v)
}

// BindAllSmart binds path, query, header parameters and body with smart content-type detection
func (ctx *Context) BindAllSmart(v any) error {
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

	// Debug: log parsed form data
	fmt.Printf("DEBUG: Parsed form data: %+v\n", formData)

	// Get binding metadata
	bindMeta := getOrBuildBindingMeta(reflect.TypeOf(v))
	rv := reflect.ValueOf(v).Elem()

	// Debug: log binding metadata
	fmt.Printf("DEBUG: Binding meta fields: %d\n", len(bindMeta.Fields))

	// Bind each field
	for _, fieldMeta := range bindMeta.Fields {
		// Skip fields that don't have body, form, or json tags
		if fieldMeta.Tag != "body" && !hasFormTag(fieldMeta.Field) && !hasJSONTag(fieldMeta.Field) {
			continue
		}

		fieldName := getFormFieldName(fieldMeta.Field)
		if fieldName == "" {
			fieldName = strings.ToLower(fieldMeta.Field.Name)
		}

		// Debug: log field processing
		fmt.Printf("DEBUG: Processing field %s (form name: %s, tag: %s)\n", fieldMeta.Field.Name, fieldName, fieldMeta.Tag)

		// Get form values
		values := formData[fieldName]
		if len(values) == 0 {
			fmt.Printf("DEBUG: No values found for field %s\n", fieldName)
			continue
		}

		fmt.Printf("DEBUG: Found values for %s: %+v\n", fieldName, values)

		// Set field value
		field := rv.FieldByIndex(fieldMeta.Index)
		if err := setFormFieldValue(field, values, fieldMeta.IsSlice); err != nil {
			return err
		}
	}

	return nil
}

// hasFormTag checks if field has form tag
func hasFormTag(field reflect.StructField) bool {
	_, ok := field.Tag.Lookup("form")
	return ok
}

// hasJSONTag checks if field has json tag
func hasJSONTag(field reflect.StructField) bool {
	_, ok := field.Tag.Lookup("json")
	return ok
}

// getFormFieldName gets the field name from form or json tag
func getFormFieldName(field reflect.StructField) string {
	// First try form tag
	if formTag := field.Tag.Get("form"); formTag != "" && formTag != "-" {
		return strings.Split(formTag, ",")[0]
	}

	// Then try json tag (for backward compatibility)
	if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
		return strings.Split(jsonTag, ",")[0]
	}

	// Finally try body tag
	if bodyTag := field.Tag.Get("body"); bodyTag != "" && bodyTag != "-" {
		return strings.Split(bodyTag, ",")[0]
	}

	return ""
}

// setFormFieldValue sets the field value from form values
func setFormFieldValue(field reflect.Value, values []string, isSlice bool) error {
	if !field.CanSet() {
		return nil
	}

	// Handle slice fields
	if isSlice && field.Kind() == reflect.Slice {
		elemType := field.Type().Elem()
		sliceValue := reflect.MakeSlice(field.Type(), len(values), len(values))

		for i, value := range values {
			elem := sliceValue.Index(i)
			if err := setSingleFieldValue(elem, value, elemType); err != nil {
				return err
			}
		}
		field.Set(sliceValue)
		return nil
	}

	// Handle single value (use first value from slice)
	value := ""
	if len(values) > 0 {
		value = values[0]
	}

	return setSingleFieldValue(field, value, field.Type())
}

// setSingleFieldValue sets a single field value based on its type
func setSingleFieldValue(field reflect.Value, value string, fieldType reflect.Type) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value == "" {
			return nil
		}
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if value == "" {
			return nil
		}
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		if value == "" {
			return nil
		}
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatVal)
	case reflect.Bool:
		switch value {
		case "", "false", "0":
			field.SetBool(false)
		case "true", "1", "on": // Handle HTML checkbox "on"
			field.SetBool(true)
		default:
			boolVal, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}
			field.SetBool(boolVal)
		}
	case reflect.Ptr:
		if value != "" {
			// Create new pointer to the appropriate type
			newVal := reflect.New(fieldType.Elem())
			if err := setSingleFieldValue(newVal.Elem(), value, fieldType.Elem()); err != nil {
				return err
			}
			field.Set(newVal)
		}
	default:
		// For complex types, try JSON unmarshaling
		if value != "" {
			var jsonValue any
			if err := json.Unmarshal([]byte(value), &jsonValue); err == nil {
				jsonBytes, _ := json.Marshal(jsonValue)
				newValue := reflect.New(fieldType).Interface()
				if err := json.Unmarshal(jsonBytes, newValue); err == nil {
					field.Set(reflect.ValueOf(newValue).Elem())
				}
			}
		}
	}

	return nil
}
