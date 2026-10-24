//go:build freebsd

package netutil

import "syscall"

// RaiseUlimit tries to raise RLIMIT_NOFILE.
// On FreeBSD, Rlimit fields are int64 (unlike Linux/Darwin which use uint64).
func RaiseUlimit(target uint64) (bool, uint64, uint64) {
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		return false, 0, 0
	}
	old := uint64(rLimit.Cur)
	if old >= target {
		return true, old, old
	}
	newCur := target
	if newCur > uint64(rLimit.Max) {
		newCur = uint64(rLimit.Max)
	}
	rLimit.Cur = int64(newCur)
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		return false, old, old
	}
	return true, old, newCur
}
