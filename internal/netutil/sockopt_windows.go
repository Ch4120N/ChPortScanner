//go:build windows

package netutil

import (
	"syscall"
)

// SetFastClose is a no-op on Windows (Windows lacks SO_LINGER=0 semantics
