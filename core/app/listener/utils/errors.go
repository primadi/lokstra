package listener_utils

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"
)

// wrapListenError wraps network listen errors with more informative messages
func WrapListenError(addr string, err error) error {
	if err == nil {
		return nil
	}

	// Check if error message contains typical "address already in use" text
	errMsg := err.Error()
	if strings.Contains(errMsg, "bind: Only one usage of each socket address") ||
		strings.Contains(errMsg, "bind: address already in use") ||
		strings.Contains(errMsg, "listen tcp") && strings.Contains(errMsg, "bind:") {
		return fmt.Errorf("❌ Port already in use: %s\n"+
			"   Another process is already listening on this port.\n"+
			"   Please choose a different port or stop the other process.\n"+
			"   Original error: %w", addr, err)
	}

	// Additional check for syscall errors (Unix/Linux)
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		var syscallErr *os.SyscallError
		if errors.As(opErr.Err, &syscallErr) {
			if errors.Is(syscallErr.Err, syscall.EADDRINUSE) {
				return fmt.Errorf("❌ Port already in use: %s\n"+
					"   Another process is already listening on this port.\n"+
					"   Please choose a different port or stop the other process.\n"+
					"   Original error: %w", addr, err)
			}
		}
	}

	return fmt.Errorf("failed to listen on %s: %w", addr, err)
}
