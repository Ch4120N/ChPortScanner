//go:build !windows

package netutil

import (
	"syscall"
)

// SetFastClose applies socket options optimized for high-speed port scanning:
