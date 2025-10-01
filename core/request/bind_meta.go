package request

import (
	"reflect"
	"sync"
)

type bindFieldMeta struct {
	Field           reflect.StructField
	Index           []int
	Name            string // param name
	Tag             string // path/query/header
	IsSlice         bool
	IsUnmarshalJSON bool

	IsIndexedKeyValue bool
	IndexKey          []int
	IndexValue        []int
	IsMap             bool
}

type bindMeta struct {
	Type   reflect.Type
	Fields []bindFieldMeta
}

var bindMetaCache sync.Map // map[reflect.Type]*bindMeta

func getOrBuildBindMeta(t reflect.Type) *bindMeta {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return &bindMeta{Type: t, Fields: nil}
	}

	if bm, ok := bindMetaCache.Load(t); ok {
		return bm.(*bindMeta)
	}

	bm := &bindMeta{
		Type: t,
	}

	numField := t.NumField()
	for i := range numField {
		field := t.Field(i)

		tagType, paramName := parseBindingTag(field)
		if tagType == "" {
			continue
		}

		isIndexedKeyValue := false
		var indexKey, indexValue []int
		// Check if it's a slice of struct with Key and Value fields
		// we look for fields named "Key"/ "Field" and "Value"
		// used for indexed key-value pairs in query/header
		// e.g. ?filter[status]=active&filter[role]=admin
		// will be bound to []struct{ Key, Value string }{
		//   { Key: "status", Value: "active" },
		//   { Key: "role", Value: "admin" },
		// }
		if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Struct {
			structType := field.Type.Elem()

			for i := range structType.NumField() {
				f := structType.Field(i)
				switch f.Name {
				case "Key", "Field":
					indexKey = f.Index
				case "Value":
					indexValue = f.Index
				}
			}

			if len(indexKey) > 0 && len(indexValue) > 0 {
				isIndexedKeyValue = true
			}
		}

		// Check if it's a map[string]string
		// used for map binding in query/header
		// e.g. ?meta[foo]=bar&meta[baz]=qux
		// will be bound to map[string]string{
		//   "foo": "bar",
		//   "baz": "qux",
		// }
		isMap := false
		if field.Type.Kind() == reflect.Map &&
			field.Type.Key().Kind() == reflect.String &&
			field.Type.Elem().Kind() == reflect.String {
			isMap = true
		}

		fieldMeta := bindFieldMeta{
			Field:             field,
			Index:             field.Index,
			Name:              paramName,
			Tag:               tagType,
			IsSlice:           field.Type.Kind() == reflect.Slice,
			IsUnmarshalJSON:   implementsUnmarshalJSON(field.Type),
			IsIndexedKeyValue: isIndexedKeyValue,
			IndexKey:          indexKey,
			IndexValue:        indexValue,
			IsMap:             isMap,
		}

		bm.Fields = append(bm.Fields, fieldMeta)
	}

	actual, loaded := bindMetaCache.LoadOrStore(t, bm)
	if loaded {
		return actual.(*bindMeta)
	}
	return bm
}

func parseBindingTag(field reflect.StructField) (tagType, paramName string) {
	for _, key := range []string{"path", "query", "header"} {
		if val, ok := field.Tag.Lookup(key); ok && val != "" {
			return key, val
		}
	}
	return "", ""
}

// unmarshalJSONType represents the interface type for json.Unmarshaler
var unmarshalJSONType = reflect.TypeOf((*interface {
	UnmarshalJSON([]byte) error
})(nil)).Elem()

func implementsUnmarshalJSON(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Check if the type or its pointer implements UnmarshalJSON
	return t.Implements(unmarshalJSONType) || reflect.PointerTo(t).Implements(unmarshalJSONType)
}
