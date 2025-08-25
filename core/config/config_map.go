package config

import (
	"fmt"
	"reflect"

	"github.com/primadi/lokstra/common/cast"
	"github.com/primadi/lokstra/core/registration"
)

func AnyParamsToStruct[T any](regCtx registration.Context, raw any) (*T, error) {
	var cfg T
	if err := decodeConfig(regCtx, raw, &cfg); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	return &cfg, nil
}

func decodeConfig[T any](regCtx registration.Context, raw any, cfg *T) error {
	rawMap, ok := raw.(map[string]any)
	if !ok {
		return fmt.Errorf("raw must be map[string]any, got %T", raw)
	}

	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("cfg must be pointer to struct")
	}
	v = v.Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		fieldVal := v.Field(i)
		if !fieldVal.CanSet() {
			continue
		} // Skip unexported fields
		field := t.Field(i)

		if useSvcName, ok := field.Tag.Lookup("useService"); ok {
			val, err := regCtx.GetService(useSvcName)
			if err != nil {
				return fmt.Errorf("get service %s: %w", useSvcName, err)
			}
			fieldVal.Set(reflect.ValueOf(val))
			continue
		}

		if svcKey, ok := field.Tag.Lookup("service"); ok {
			svcName, exists := rawMap[svcKey]
			svcNameStr, _ := svcName.(string)

			if !exists || cast.IsEmpty(svcNameStr) {
				if defVal, ok := field.Tag.Lookup("default"); ok {
					svcNameStr = defVal
				} else {
					continue
				}
			}

			val, err := regCtx.GetService(svcNameStr)
			if err != nil {
				return fmt.Errorf("get service %s: %w", svcNameStr, err)
			}
			fieldVal.Set(reflect.ValueOf(val))
			continue
		}

		if cfgKey, ok := field.Tag.Lookup("config"); ok {
			if val, exists := rawMap[cfgKey]; exists {
				fieldVal.Set(reflect.ValueOf(val))
			}
		}
	}

	return nil
}
