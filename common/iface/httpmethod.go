package iface

// HTTPMethod represents an HTTP method as a string.
// It is used to define the type of HTTP request methods supported by the router.
type HTTPMethod = string

const (
	GET    HTTPMethod = "GET"
	POST   HTTPMethod = "POST"
	PUT    HTTPMethod = "PUT"
	DELETE HTTPMethod = "DELETE"
	PATCH  HTTPMethod = "PATCH"
)
