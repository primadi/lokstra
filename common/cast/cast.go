package cast

import (
	"fmt"
	"reflect"
	"time"
)

// Attempt to convert any value to int
//   - return 0 and nil, if val is nil
//   - return 0 and error, if conversion failed
func ToInt(val any) (int, error) {
	if val == nil {
		return 0, nil // No content
	}

	switch v := val.(type) {
	case int:
		return v, nil
	case int8:
		return int(v), nil
	case int16:
		return int(v), nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case uint:
		return int(v), nil
	case uint8:
		return int(v), nil
	case uint16:
		return int(v), nil
	case uint32:
		return int(v), nil
	case uint64:
		return int(v), nil
	}
	return 0, fmt.Errorf("unexpected cast to int: %T", val)
}

// Attempt to convert any value to float64
//   - if val is string, try to parse it using t
func ToFloat64(val any) (float64, error) {
	if val == nil {
		return 0, nil // No content
	}

	switch v := val.(type) {
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	}
	return 0, fmt.Errorf("unexpected cast to float64: %T", val)
}

// Attempt to convert any value to time.Time
//   - return zero time and nil, if val is nil
//   - if val is string, try to parse it using time.DateTime, time.DateOnly, time.TimeOnly, time.RFC3339, time.RFC3339Nano
//   - if val is int64 or float64, treat it as Unix timestamp
func ToTime(val any) (time.Time, error) {
	if val == nil {
		return time.Time{}, nil // No content
	}
	switch v := val.(type) {
	case time.Time:
		return v, nil
	case string:
		// Try different time formats
		formats := []string{
			time.DateTime,    // "2006-01-02 15:04:05"
			time.DateOnly,    // "2006-01-02"
			time.TimeOnly,    // "15:04:05"
			time.RFC3339,     // "2006-01-02T15:04:05Z07:00"
			time.RFC3339Nano, // "2006-01-02T15:04:05.999999999Z07:00"

		}
		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("unable to parse time: %s", v)
	case int64:
		return time.Unix(v, 0), nil
	case float64:
		return time.Unix(int64(v), 0), nil
	default:
		return time.Time{}, fmt.Errorf("unexpected time type: %T", val)
	}
}

// Attempt to convert any value to type T
//   - return zero value of T and nil, if val is nil
//   - return zero value of T and error, if conversion failed
//   - support basic types (int, float64, string, bool), struct, pointer to struct, slice of struct
//   - for struct and slice of struct, use ToStruct for conversion
func ToType[T any](val any, strict bool) (T, error) {
	var zero T

	if val == nil {
		return zero, nil // No content
	}

	switch any(zero).(type) {
	case int:
		val, err := ToInt(val)
		if err != nil {
			return zero, err
		}
		return any(val).(T), nil
	case float64:
		val, err := ToFloat64(val)
		if err != nil {
			return zero, err
		}
		return any(val).(T), nil
	case time.Time:
		val, err := ToTime(val)
		if err != nil {
			return zero, err
		}
		return any(val).(T), nil
	}

	tType := reflect.TypeOf(zero)

	// handle pointer to struct
	if tType.Kind() == reflect.Pointer && tType.Elem().Kind() == reflect.Struct {
		ptr := reflect.New(tType.Elem()) // pointer to struct
		if err := ToStruct(val, ptr.Interface(), strict); err != nil {
			return zero, err
		}
		return ptr.Interface().(T), nil
	}

	// handle struct directly
	if tType.Kind() == reflect.Struct {
		ptr := reflect.New(tType) // Create a pointer to the struct type
		if err := ToStruct(val, ptr.Interface(), strict); err != nil {
			return zero, err
		}
		// return the dereferenced pointer as the type T
		return ptr.Elem().Interface().(T), nil
	}

	if tType.Kind() == reflect.Slice {
		slicePtr := reflect.New(tType).Interface()
		if err := ToStruct(val, slicePtr, strict); err != nil {
			return zero, err
		}
		return reflect.ValueOf(slicePtr).Elem().Interface().(T), nil
	}

	if m, ok := val.(T); ok {
		return m, nil
	}

	return zero, fmt.Errorf("unexpected cast to %T: %T", zero, val)
}

// Attempt to convert any slice to slice of type T
//   - return zero value of T and nil, if slice is nil
//   - return zero value of T and error, if conversion failed
//   - T must be a slice type
//   - each element in the input slice must be assignable to the element type of T
func SliceConvert[T any](slice any) (T, error) {
	var zero T
	if slice == nil {
		return zero, nil // No content
	}

	sliceVal := reflect.ValueOf(slice)
	if sliceVal.Kind() != reflect.Slice {
		return zero, fmt.Errorf("expected slice type, got %T", slice)
	}

	tType := reflect.TypeOf(zero)
	if tType.Kind() != reflect.Slice {
		return zero, fmt.Errorf("type parameter T must be a slice, got %T", zero)
	}

	elemType := tType.Elem()

	result := reflect.MakeSlice(tType, sliceVal.Len(), sliceVal.Len())
	for i := 0; i < sliceVal.Len(); i++ {
		item := sliceVal.Index(i).Interface()
		itemVal := reflect.ValueOf(item)
		if !itemVal.Type().AssignableTo(elemType) {
			return zero, fmt.Errorf("cannot assign value of type %T to %s at index %d", item, elemType, i)
		}

		result.Index(i).Set(itemVal)
	}

	return result.Interface().(T), nil
}

// Check if a value is empty
func IsEmpty(val any) bool {
	if val == nil {
		return true
	}

	v := reflect.ValueOf(val)

	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	case reflect.Pointer, reflect.Interface:
		return v.IsNil()
	case reflect.Struct:
		// Optionally: check if all fields in struct are empty
		for i := 0; i < v.NumField(); i++ {
			if !IsEmpty(v.Field(i).Interface()) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
