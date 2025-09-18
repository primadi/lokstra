package rpc_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/primadi/lokstra/common/cast"
	"github.com/vmihailenco/msgpack/v5"
)

// RpcClient is the core HTTP client for making RPC calls
type RpcClient struct {
	baseUrl string
	client  *http.Client
}

func NewRpcClient(baseUrl string) *RpcClient {
	return &RpcClient{
		baseUrl: baseUrl,
		client:  &http.Client{},
	}
}

func (c *RpcClient) Call(methodName string, args ...any) (any, error) {
	// Convert method name to lowercase for URL
	endpoint := c.baseUrl + "/" + strings.ToLower(methodName)

	// 1. Encode arguments as msgpack
	data, err := msgpack.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("encode error: %w", err)
	}

	// 2. Send HTTP request
	resp, err := c.client.Post(endpoint, "application/octet-stream", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	// 3. Read response
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response error: %w", err)
	}

	// 4. Check status code
	if resp.StatusCode >= 400 {
		// Decode JSON error response
		var errResp struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(respBytes, &errResp) != nil || errResp.Message == "" {
			errResp.Message = string(respBytes)
		}
		return nil, fmt.Errorf("remote error: %s", errResp.Message)
	}

	// 5. Decode result (msgpack)
	if len(respBytes) == 0 {
		// Empty response - return nil
		return nil, nil
	}

	var result any
	if err := msgpack.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return result, nil
}

// CallReturnType calls RpcClient method and returns result as type T
func CallReturnType[T any](c *RpcClient, methodName string, args ...any) (T, error) {
	var zero T

	result, err := c.Call(methodName, args...)
	if err != nil {
		return zero, err
	}

	if ret, ok := result.(T); ok {
		return ret, nil
	}

	return cast.ToType[T](result)
}

// CallReturnSliceIface calls RpcClient method and converts the result to a slice of type T2
func CallReturnSliceIface[TSliceStruct any, TSliceIface any](c *RpcClient,
	methodName string, args ...any) (TSliceIface, error) {

	ret, err := CallReturnType[TSliceStruct](c, methodName, args...)
	if err != nil {
		var zero TSliceIface
		return zero, fmt.Errorf("cast error: %w", err)
	}

	return cast.SliceConvert[TSliceIface](ret)
}

func CallReturnVoid(c *RpcClient, methodName string, args ...any) error {
	_, err := c.Call(methodName, args...)
	if err != nil {
		return fmt.Errorf("call error: %w", err)
	}
	return nil
}
