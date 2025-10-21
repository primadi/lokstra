package convention

import "strings"

// RPCConvention implements RPC routing convention
type RPCConvention struct{}

func (c *RPCConvention) Name() ConventionType {
	return RPC
}

func (c *RPCConvention) ResolveMethod(methodName string, resource string, resourcePlural string) (httpMethod string, pathTemplate string, found bool) {
	// All RPC methods are POST to /rpc/{method-name}
	return "POST", "/rpc/" + strings.ToLower(methodName), true
}

// Ensure RPCConvention implements Convention
var _ Convention = (*RPCConvention)(nil)
