package rpc_service

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/primadi/lokstra/core/request"

	"github.com/vmihailenco/msgpack/v5"
)

type ServiceRegistry struct {
	services map[string]*ServiceMeta
}

type ServiceMeta struct {
	Name    string
	Impl    any
	Methods map[string]*MethodMeta
}

type MethodMeta struct {
	Name     string
	Func     reflect.Value
	ArgTypes []reflect.Type
	OutTypes []reflect.Type
}

var registry = &ServiceRegistry{
	services: make(map[string]*ServiceMeta),
}

func RegisterService(name string, impl any) error {
	if _, exists := registry.services[name]; exists {
		return fmt.Errorf("rpc service %q already registered", name)
	}

	t := reflect.TypeOf(impl)
	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("rpc service %q must be a pointer", name)
	}

	methods := map[string]*MethodMeta{}
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		if !m.IsExported() {
			continue
		}
		mType := m.Type

		switch mType.NumOut() {
		case 1:
			// single return value must be error
			if !mType.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
				continue
			}
		case 2:
			// two return values must be (result, error)
			if !mType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
				continue
			}
		default:
			continue // invalid number of return values
		}

		argTypes := []reflect.Type{}
		for j := 1; j < mType.NumIn(); j++ { // skip receiver
			argTypes = append(argTypes, mType.In(j))
		}

		outTypes := []reflect.Type{}
		for j := 0; j < mType.NumOut(); j++ {
			outTypes = append(outTypes, mType.Out(j))
		}

		methods[strings.ToLower(m.Name)] = &MethodMeta{
			Name:     m.Name,
			Func:     m.Func,
			ArgTypes: argTypes,
			OutTypes: outTypes,
		}
	}

	registry.services[name] = &ServiceMeta{
		Name:    name,
		Impl:    impl,
		Methods: methods,
	}

	return nil
}

func GetService(name string) (*ServiceMeta, error) {
	svc, ok := registry.services[name]
	if !ok {
		return nil, fmt.Errorf("rpc service %q not found", name)
	}
	return svc, nil
}

func (mm *MethodMeta) HandleRequest(svc *ServiceMeta, ctx *request.Context) error {
	// Step 1: Decode body (msgpack)
	body, err := ctx.GetRawBody()
	if err != nil {
		return ctx.ErrorBadRequest("invalid body")
	}

	var rawArgs []any
	if err := msgpack.Unmarshal(body, &rawArgs); err != nil {
		return ctx.ErrorBadRequest("invalid msgpack payload")
	}

	if len(rawArgs) != len(mm.ArgTypes) {
		return ctx.ErrorBadRequest("invalid argument count")
	}

	// Step 2: Convert to reflect.Value
	in := []reflect.Value{reflect.ValueOf(svc.Impl)}
	for i, arg := range rawArgs {
		argVal := reflect.ValueOf(arg)
		wantType := mm.ArgTypes[i]
		if !argVal.Type().AssignableTo(wantType) {
			argVal = argVal.Convert(wantType)
		}
		in = append(in, argVal)
	}

	// Step 3: Call actual method
	result := mm.Func.Call(in)

	// Step 4: Encode response
	switch len(result) {
	case 1:
		if !result[0].IsNil() {
			errVal := result[0].Interface().(error)
			return ctx.ErrorInternal(errVal.Error())
		}
		ctx.Headers.Set("Content-Type", "application/octet-stream")
		return ctx.Ok(nil) // No content
	case 2:
		if !result[1].IsNil() {
			errVal := result[1].Interface().(error)
			return ctx.ErrorInternal(errVal.Error())
		}
		respData, err := msgpack.Marshal(result[0].Interface())
		if err != nil {
			return ctx.ErrorInternal("encoding error")
		}

		return ctx.WriteRaw("application/octet-stream", 200, respData)
	default:
		return ctx.ErrorInternal("unexpected number of return values")
	}
}

type RpcService interface {
	HandleRequest(ctx *request.Context, serviceName, MethodName string) error
}
