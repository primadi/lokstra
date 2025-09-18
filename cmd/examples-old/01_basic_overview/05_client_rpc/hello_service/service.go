package hello_service

import (
	"time"

	"github.com/primadi/lokstra/modules/rpc_service"
)

type GreetingServiceClient struct {
	client *rpc_service.RpcClient
}

func NewGreetingServiceClient(baseURL string) *GreetingServiceClient {
	return &GreetingServiceClient{
		client: rpc_service.NewRpcClient(baseURL),
	}
}

// Return string, error
func (c *GreetingServiceClient) Hello(name string) (string, error) {
	return rpc_service.CallReturnType[string](c.client, "Hello", name)
}

// Return *User, error (interface implementation)
func (c *GreetingServiceClient) GetUser(id int) (UserIface, error) {
	return rpc_service.CallReturnType[*User](c.client, "GetUser", id)
}

// Return []*User, error
func (c *GreetingServiceClient) GetUsers(limit int) ([]UserIface, error) {
	return rpc_service.CallReturnSliceIface[[]*User, []UserIface](
		c.client, "GetUsers", limit)
}

// Return map, error
func (c *GreetingServiceClient) GetUserStats(id int) (map[string]any, error) {
	return rpc_service.CallReturnType[map[string]any](c.client, "GetUserStats", id)
}

// Return primitive types, error
func (c *GreetingServiceClient) GetUserCount() (int, error) {
	return rpc_service.CallReturnType[int](c.client, "GetUserCount")
}

func (c *GreetingServiceClient) GetUserActive(id int) (bool, error) {
	return rpc_service.CallReturnType[bool](c.client, "GetUserActive", id)
}

func (c *GreetingServiceClient) GetServerTime() (time.Time, error) {
	return rpc_service.CallReturnType[time.Time](c.client, "GetServerTime")
}

// Return any, error
func (c *GreetingServiceClient) GetDynamicData(dataType string) (any, error) {
	return rpc_service.CallReturnType[any](c.client, "GetDynamicData", dataType)
}

// Return only error (void operations)
func (c *GreetingServiceClient) DeleteUser(id int) error {
	return rpc_service.CallReturnVoid(c.client, "DeleteUser", id)
}

func (c *GreetingServiceClient) ClearCache() error {
	return rpc_service.CallReturnVoid(c.client, "ClearCache")
}

func (c *GreetingServiceClient) Ping() error {
	return rpc_service.CallReturnVoid(c.client, "Ping")
}

func (c *GreetingServiceClient) GetSystemInfo() (SystemInfo, error) {
	return rpc_service.CallReturnType[SystemInfo](c.client, "GetSystemInfo")
}

var _ GreetingService = (*GreetingServiceClient)(nil)
