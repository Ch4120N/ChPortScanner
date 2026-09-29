//go:build !windows

package netutil

import (
	"syscall"
)

// SetFastClose applies socket options optimized for high-speed port scanning:
//   - SO_LINGER=0  → sends RST on close (no TIME_WAIT accumulation)
