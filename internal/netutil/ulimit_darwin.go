//go:build darwin

package netutil

import "syscall"

// RaiseUlimit tries to raise RLIMIT_NOFILE. On Darwin, Rlimit fields are uint64.
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
}
