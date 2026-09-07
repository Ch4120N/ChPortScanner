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
		}
		if e := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); e != nil {
			sockErr = e
			return
		}
		if network == "tcp" || network == "tcp4" || network == "tcp6" {
			_ = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, syscall.TCP_NODELAY, 1)
		}
	})
	if err != nil {
		return err
	}
	return sockErr
}

// RaiseUlimit tries to raise the RLIMIT_NOFILE soft limit. Returns
// (success, oldLimit, newLimit). Silently returns (false, 0, 0) if unsupported.
func RaiseUlimit(target uint64) (bool, uint64, uint64) {
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		return false, 0, 0
	}
	old := rLimit.Cur
	if old >= target {
		return true, old, old
	}
	newCur := target
	if newCur > rLimit.Max {
		newCur = rLimit.Max
	}
	rLimit.Cur = newCur
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		return false, old, old
	}
	return true, old, newCur
