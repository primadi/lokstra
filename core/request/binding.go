package request

import (
	"net/http"
	"net/url"
	"reflect"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

var jsonBodyDecoder = jsoniter.Config{TagKey: "body"}.Froze()

func (ctx *Context) bindPathField(fieldMeta bindingFieldMeta, rv reflect.Value) error {
	rawValue := ctx.GetPathParam(fieldMeta.Name)
	rawValues := []string{rawValue}
	return convertAndSetField(rv.FieldByIndex(fieldMeta.Index), rawValues,
		fieldMeta.IsSlice, fieldMeta.IsUnmarshalJSON)
}

func (ctx *Context) BindPath(v any) error {
	bindMeta := getOrBuildBindingMeta(reflect.TypeOf(v))
	rv := reflect.ValueOf(v).Elem()

	for _, fieldMeta := range bindMeta.Fields {
		if fieldMeta.Tag != "path" {
			continue
		}

		if err := ctx.bindPathField(fieldMeta, rv); err != nil {
			return err
		}
	}

	return nil
}

func (ctx *Context) bindQueryField(fieldMeta bindingFieldMeta, rv reflect.Value, query url.Values) error {
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

func (ctx *Context) BindQuery(v any) error {
	bindMeta := getOrBuildBindingMeta(reflect.TypeOf(v))
	rv := reflect.ValueOf(v).Elem()
	query := ctx.Request.URL.Query()

	for _, fieldMeta := range bindMeta.Fields {
		if fieldMeta.Tag != "query" {
			continue
		}

		if err := ctx.bindQueryField(fieldMeta, rv, query); err != nil {
			return err
		}
	}

	return nil
}

func (ctx *Context) bindHeaderField(fieldMeta bindingFieldMeta, rv reflect.Value, header http.Header) error {
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

func (ctx *Context) BindHeader(v any) error {
	bindMeta := getOrBuildBindingMeta(reflect.TypeOf(v))
	rv := reflect.ValueOf(v).Elem()
	header := ctx.Request.Header

	for _, fieldMeta := range bindMeta.Fields {
		if fieldMeta.Tag != "header" {
			continue
		}

		if err := ctx.bindHeaderField(fieldMeta, rv, header); err != nil {
			return err
		}
	}

	return nil
}

func (ctx *Context) BindBody(v any) error {
	ctx.cacheRequestBody()
	if ctx.requestBodyErr != nil {
		return ctx.requestBodyErr
	}
	if len(ctx.rawRequestBody) == 0 {
		return nil // No body to bind
	}
	return jsonBodyDecoder.Unmarshal(ctx.rawRequestBody, v)
}

func (ctx *Context) BindAll(v any) error {
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

	if err := ctx.BindBody(v); err != nil {
		return err
	}
	return nil
}
