package rpc_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
)

func GetRemoteService[T any](baseUrl string) T {
	var t T
	ifaceType := reflect.TypeOf(&t).Elem()

	impl := reflect.New(ifaceType).Elem()
	numMethod := ifaceType.NumMethod()

	for i := 0; i < numMethod; i++ {
		m := ifaceType.Method(i)
		methodType := m.Type
		methodName := strings.ToLower(m.Name)

		fn := reflect.MakeFunc(methodType, func(args []reflect.Value) []reflect.Value {
			// 1. Prepare arguments
			in := make([]any, len(args))
			for i := range args {
				in[i] = args[i].Interface()
			}

			// 2. Encode msgpack
			data, err := msgpack.Marshal(in)
			if err != nil {
				return returnError(methodType, fmt.Errorf("encode error: %w", err))
			}

			// 3. Send HTTP request
			resp, err := http.Post(baseUrl+"/"+methodName,
				"application/octet-stream", bytes.NewReader(data))
			if err != nil {
				return returnError(methodType, fmt.Errorf("http error: %w", err))
			}
			defer resp.Body.Close()

			respBytes, _ := io.ReadAll(resp.Body)

			// 4. Cek status code
			if resp.StatusCode >= 400 {
				// decode JSON error response
				var errResp struct {
					Message string `json:"message"`
				}
				json.Unmarshal(respBytes, &errResp)
				if errResp.Message == "" {
					errResp.Message = string(respBytes)
				}
				return returnError(methodType, fmt.Errorf("remote error: %s", errResp.Message))
			}

			// 5. Decode result (msgpack)
			if methodType.NumOut() == 1 {
				// only one return value, no error
				return []reflect.Value{reflect.Zero(methodType.Out(0))}
			}

			outVal := reflect.New(methodType.Out(0))
			if err := msgpack.Unmarshal(respBytes, outVal.Interface()); err != nil {
				return returnError(methodType, fmt.Errorf("decode error: %w", err))
			}

			return []reflect.Value{
				outVal.Elem(),
				reflect.Zero(methodType.Out(1)), // error = nil
			}
		})

		impl.Field(i).Set(fn)
	}

	return impl.Interface().(T)
}

func returnError(methodType reflect.Type, err error) []reflect.Value {
	if methodType.NumOut() == 1 {
		return []reflect.Value{reflect.ValueOf(err)}
	}
	return []reflect.Value{
		reflect.Zero(methodType.Out(0)),
		reflect.ValueOf(err),
	}
}
