package rpc_service

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"

	"github.com/vmihailenco/msgpack/v5"
)

type serviceMeta struct {
	Name    string
	Impl    service.Service
	Methods map[string]*methodMeta
}

type methodMeta struct {
	Name     string
	Func     reflect.Value
	ArgTypes []reflect.Type
	OutTypes []reflect.Type
}

var serviceMetaRegistry = make(map[string]*serviceMeta)

func registerServiceMeta(svc service.Service) (*serviceMeta, error) {
	serviceName := reflect.TypeOf(svc).Name()
	if _, exists := serviceMetaRegistry[serviceName]; exists {
		return nil, fmt.Errorf("rpc service %q already registered", serviceName)
	}

	t := reflect.TypeOf(svc)
	if t.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("rpc service %q must be a pointer", serviceName)
	}

	methods := map[string]*methodMeta{}
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

		methods[strings.ToLower(m.Name)] = &methodMeta{
			Name:     m.Name,
			Func:     m.Func,
			ArgTypes: argTypes,
			OutTypes: outTypes,
		}
	}

	svcMeta := &serviceMeta{
		Name:    serviceName,
		Impl:    svc,
		Methods: methods,
	}

	serviceMetaRegistry[serviceName] = svcMeta

	return svcMeta, nil
}

func getServiceMeta(svc service.Service) (*serviceMeta, error) {
	serviceName := reflect.TypeOf(svc).Name()
	svcMeta, ok := serviceMetaRegistry[serviceName]
	if !ok {
		var err error
		svcMeta, err = registerServiceMeta(svc)
		if err != nil {
			return nil, fmt.Errorf("failed to register service %q: %w", serviceName, err)
		}
	}
	return svcMeta, nil
}

func (mm *methodMeta) HandleRequest(svc *serviceMeta, ctx *request.Context) error {
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
		ctx.WithHeader("Content-Type", "application/octet-stream")
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

func (sm *serviceMeta) HandleRpcRequest(ctx *request.Context, methodName string) error {
	method := sm.Methods[strings.ToLower(methodName)]
	if method == nil {
		return ctx.ErrorBadRequest("method not found: " + methodName)
	}

	return method.HandleRequest(sm, ctx)
}

func HandleRpcRequest(ctx *request.Context, svc service.Service, methodName string) error {
	svcMeta, err := getServiceMeta(svc)
	if err != nil {
		return ctx.ErrorInternal("failed to get serviceMeta: " + err.Error())
	}

	return svcMeta.HandleRpcRequest(ctx, methodName)
}
