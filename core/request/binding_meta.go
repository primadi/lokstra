package request

import (
	"reflect"
	"sync"
)

type bindingFieldMeta struct {
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

type bindingMeta struct {
	Type   reflect.Type
	Fields []bindingFieldMeta
}

var bindingMetaCache sync.Map // map[reflect.Type]*bindingMeta

func getOrBuildBindingMeta(t reflect.Type) *bindingMeta {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if meta, ok := bindingMetaCache.Load(t); ok {
		return meta.(*bindingMeta)
	}

	meta := &bindingMeta{
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

		isMap := false
		if field.Type.Kind() == reflect.Map &&
			field.Type.Key().Kind() == reflect.String &&
			field.Type.Elem().Kind() == reflect.String {
			isMap = true
		}

		fieldMeta := bindingFieldMeta{
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

		meta.Fields = append(meta.Fields, fieldMeta)
	}

	actual, loaded := bindingMetaCache.LoadOrStore(t, meta)
	if loaded {
		return actual.(*bindingMeta)
	}
	return meta
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
