//go:build windows

package netutil

import (
	"syscall"
)

// SetFastClose is a no-op on Windows (Windows lacks SO_LINGER=0 semantics
// for the connect-scan use case via the standard syscall package).
func SetFastClose(network, address string, c syscall.RawConn) error {
	return nil
}

// RaiseUlimit is unsupported on Windows.
func RaiseUlimit(target uint64) (bool, uint64, uint64) {
