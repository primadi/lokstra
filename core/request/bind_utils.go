package request

import (
	"errors"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/primadi/lokstra/common/json"
)

// convertAndSetField converts raw values to the appropriate type and sets them on the field.
func convertAndSetField(field reflect.Value, rawValues []string, isSlice bool, isUnmarshalJSON bool) error {
	if !field.CanSet() {
		return errors.New("field cannot be set")
	}

	if isSlice {
		sliceVal := reflect.MakeSlice(field.Type(), len(rawValues), len(rawValues))
		for i, raw := range rawValues {
			elemField := sliceVal.Index(i)
			if err := setValue(elemField, raw, isUnmarshalJSON); err != nil {
				return err
			}
		}
		field.Set(sliceVal)
	} else {
		value := ""
		if len(rawValues) > 0 {
			value = rawValues[0]
		}
		if err := setValue(field, value, isUnmarshalJSON); err != nil {
			return err
		}
	}
	return nil
}

// setValue sets the value of a field based on its type and the provided raw string.
func setValue(field reflect.Value, raw string, isUnmarshalJSON bool) error {
	if isUnmarshalJSON {
		data, _ := json.Marshal(raw)
		return field.Addr().Interface().(interface {
			UnmarshalJSON([]byte) error
		}).UnmarshalJSON(data)
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(raw)
	case reflect.Bool:
		// Empty string defaults to false for bool
		if raw == "" {
			field.SetBool(false)
			return nil
		}
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return err
		}
		field.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Empty string defaults to 0 for integers
		if raw == "" {
			field.SetInt(0)
			return nil
		}
		i, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// Empty string defaults to 0 for unsigned integers
		if raw == "" {
			field.SetUint(0)
			return nil
		}
		u, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(u)
	case reflect.Float32, reflect.Float64:
		// Empty string defaults to 0.0 for floats
		if raw == "" {
			field.SetFloat(0.0)
			return nil
		}
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return err
		}
		field.SetFloat(f)
	default:
		return errors.New("unsupported field type")
	}

	return nil
}

// splitCommaSeparated splits a comma-separated string into a slice of strings, trimming whitespace.
func splitCommaSeparated(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// parseIndexedParamValuesReflect parses indexed parameters from a URL query into a slice of structs.
func parseIndexedParamValuesReflect(paramPrefix string, query url.Values, field reflect.Value, indexKey, indexValue []int) error {
	if field.Kind() != reflect.Slice {
		return errors.New("target field is not slice")
	}

	elemType := field.Type().Elem()
	sliceVal := reflect.MakeSlice(field.Type(), 0, 0)

	for key, values := range query {
		if strings.HasPrefix(key, paramPrefix+"[") && strings.HasSuffix(key, "]") {
			fieldName := key[len(paramPrefix)+1 : len(key)-1]

			for _, val := range values {
				elem := reflect.New(elemType).Elem()

				// Set Key
				keyField := elem.FieldByIndex(indexKey)
				if err := setValue(keyField, fieldName, false); err != nil {
					return err
				}

				// Set Value
				valueField := elem.FieldByIndex(indexValue)
				if err := setValue(valueField, val, false); err != nil {
					return err
				}

				// Append to slice
				sliceVal = reflect.Append(sliceVal, elem)
			}
		}
	}

	field.Set(sliceVal)
	return nil
}
