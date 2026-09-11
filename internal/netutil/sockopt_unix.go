//go:build !windows

package netutil

import (
	"syscall"
)

// SetFastClose applies socket options optimized for high-speed port scanning:
//   - SO_LINGER=0  → sends RST on close (no TIME_WAIT accumulation)
//   - SO_REUSEADDR → allows quick rebinding
//   - TCP_NODELAY  → disables Nagle (irrelevant for connect, harmless)
func SetFastClose(network, address string, c syscall.RawConn) error {
	var sockErr error
	err := c.Control(func(fd uintptr) {
		linger := syscall.Linger{Onoff: 1, Linger: 0}
		if e := syscall.SetsockoptLinger(int(fd), syscall.SOL_SOCKET, syscall.SO_LINGER, &linger); e != nil {
			sockErr = e
			return
