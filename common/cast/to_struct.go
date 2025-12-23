package cast

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/primadi/lokstra/common/json"
)

// cache: map[reflect.Type]map[string]int
var structCache sync.Map

func getFieldMap(t reflect.Type) map[string]int {
	if fm, ok := structCache.Load(t); ok {
		return fm.(map[string]int)
	}

	fm := make(map[string]int)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		name := field.Tag.Get("json")
		if name == "-" {
			continue // Skip fields with json:"-"
		}
		if name == "" {
			name = field.Name
		}
		fm[name] = i
	}
	structCache.Store(t, fm)
	return fm
}

// Convert map[string]any to struct
//   - structOut must be pointer to struct
//   - if strict is true, unknown fields will cause error
func ToStruct(val any, structOut any, strict bool) error {
	if structOut == nil {
		return fmt.Errorf("structOut cannot be nil")
	}

	v := reflect.ValueOf(structOut)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return fmt.Errorf("structOut must be a non-nil pointer to struct")
	}
	v = v.Elem()

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("structOut must point to a struct, got %T", structOut)
	}

	m, ok := val.(map[string]any)
	if !ok {
		return fmt.Errorf("val must be map[string]any, got %T", val)
	}

	t := v.Type()
	fieldMap := getFieldMap(t)

	for name, raw := range m {
		idx, exists := fieldMap[name]
		if !exists {
			if strict {
				return fmt.Errorf("unknown field: %s", name)
			}
			continue // skip if non-strict
		}

		field := v.Field(idx)
		if !field.CanSet() {
			continue
		}

		if err := assignValue(field, raw, strict); err != nil {
			return fmt.Errorf("field %s: %w", name, err)
		}
	}

	return nil
}

// assignValue → recursive assignment
func assignValue(field reflect.Value, raw any, strict bool) error {
	if raw == nil {
		return nil
	}

	rawVal := reflect.ValueOf(raw)
	ft := field.Type()

	// case 0: check if type implements json.Unmarshaler
	if field.CanAddr() {
		if unmarshaler, ok := field.Addr().Interface().(json.Unmarshaler); ok {
			// Convert raw to JSON bytes
			jsonBytes, err := json.Marshal(raw)
			if err != nil {
				return fmt.Errorf("failed to marshal for Unmarshaler: %w", err)
			}
			if err := unmarshaler.UnmarshalJSON(jsonBytes); err != nil {
				return fmt.Errorf("UnmarshalJSON failed: %w", err)
			}
			return nil
		}
	}

	// case 1: langsung assignable
	if rawVal.Type().AssignableTo(ft) {
		field.Set(rawVal)
		return nil
	}

	// case 2: convert primitive
	if converted, ok := tryConvert(rawVal, ft); ok {
		field.Set(converted)
		return nil
	}

	// case 3: nested struct
	if ft.Kind() == reflect.Struct {
		if m, ok := raw.(map[string]any); ok {
			ptr := reflect.New(ft)
			if err := ToStruct(m, ptr.Interface(), strict); err != nil {
				return err
			}
			field.Set(ptr.Elem())
			return nil
		}
	}

	// case 4: pointer to struct
	if ft.Kind() == reflect.Pointer && ft.Elem().Kind() == reflect.Struct {
		if m, ok := raw.(map[string]any); ok {
			ptr := reflect.New(ft.Elem())
			if err := ToStruct(m, ptr.Interface(), strict); err != nil {
				return err
			}
			field.Set(ptr)
			return nil
		}
	}

	// case 5: slice
	if ft.Kind() == reflect.Slice {
		rawSlice, ok := raw.([]any)
		if !ok {
			return fmt.Errorf("expected slice for field, got %T", raw)
		}

		slice := reflect.MakeSlice(ft, len(rawSlice), len(rawSlice))
		for i, elem := range rawSlice {
			if err := assignValue(slice.Index(i), elem, strict); err != nil {
				return fmt.Errorf("index %d: %w", i, err)
			}
		}
		field.Set(slice)
		return nil
	}

	return fmt.Errorf("cannot assign %T to %s", raw, ft)
}

// tryConvert: basic conversions
func tryConvert(val reflect.Value, targetType reflect.Type) (reflect.Value, bool) {
	switch targetType.Kind() {
	case reflect.Int:
		switch val.Kind() {
		case reflect.Float64:
			return reflect.ValueOf(int(val.Float())), true
		case reflect.String:
			i, err := strconv.Atoi(val.String())
			if err == nil {
				return reflect.ValueOf(i), true
			}
		}
	case reflect.String:
		return reflect.ValueOf(fmt.Sprintf("%v", val.Interface())), true
	case reflect.Float64:
		if val.Kind() == reflect.Int {
			return reflect.ValueOf(float64(val.Int())), true
		}
	}

	// Special case: time.Time
	if targetType.String() == "time.Time" {
		if val.Kind() == reflect.String {
			t, err := ToTime(val.String())
			if err == nil {
				return reflect.ValueOf(t), true
			}
		}
	}

	// Special case: time.Duration
	if targetType.String() == "time.Duration" {
		switch val.Kind() {
		case reflect.Int, reflect.Int64:
			return reflect.ValueOf(time.Duration(val.Int())), true
		case reflect.Float64:
			return reflect.ValueOf(time.Duration(val.Float())), true
		case reflect.String:
			d, err := time.ParseDuration(val.String())
			if err == nil {
				return reflect.ValueOf(d), true
			}
		}
	}

	return reflect.Value{}, false
}
