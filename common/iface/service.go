package iface

import "errors"

var ErrServiceTypeMismatch = errors.New("service type mismatch")

// App interface defines the methods required for an application to start and stop.
type Service any
