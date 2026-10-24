//go:build unix && !linux && !darwin && !freebsd

package netutil

// RaiseUlimit is unsupported on this Unix variant — returns a no-op.
// The scanner still works; it simply cannot auto-raise ulimit here.
func RaiseUlimit(target uint64) (bool, uint64, uint64) {
	return false, 0, 0
}
