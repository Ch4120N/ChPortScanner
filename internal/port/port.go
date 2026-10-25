package port

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// topPorts is a static list of common ports. No service names attached.
var topPorts = []uint16{
	7, 9, 13, 21, 22, 23, 25, 26, 37, 53, 79, 80, 81, 88, 106, 110, 111, 113,
	119, 135, 139, 143, 144, 179, 199, 389, 427, 443, 444, 445, 465, 513, 514,
	515, 543, 544, 548, 554, 587, 631, 646, 873, 990, 993, 995, 1025, 1026,
	1027, 1028, 1029, 1110, 1433, 1720, 1723, 1755, 1900, 2000, 2001, 2049,
	2121, 2717, 3000, 3128, 3306, 3389, 3986, 4899, 5000, 5009, 5051, 5060,
	5101, 5190, 5357, 5432, 5631, 5666, 5800, 5900, 6000, 6001, 6646, 7070,
	8000, 8008, 8009, 8080, 8081, 8443, 8888, 9100, 9999, 10000, 32768,
	49152, 49153, 49154, 49155, 49156, 49157,
}

// Parse returns every port as uint32. Nothing is silently filtered.
//
// Ports above 65535 are preserved here; the scanner wraps them modulo
// 65536 before dialing (the TCP/UDP header only carries 16 bits).
//
// Return values:
//   ports    — deduplicated, sorted uint32 port list
//   warnings — human-readable notes about entries that couldn't be parsed
//   err      — only when zero ports remain after parsing/exclusion
func Parse(spec, exclude string, topN int) ([]uint32, []string, error) {
	set := map[uint32]struct{}{}
	var warnings []string

	if topN > 0 {
		n := topN
		if n > len(topPorts) {
			n = len(topPorts)
		}
		for i := 0; i < n; i++ {
			set[uint32(topPorts[i])] = struct{}{}
		}
	} else {
		if spec == "" {
			return nil, nil, fmt.Errorf("no ports given")
		}
		for _, chunk := range strings.Split(spec, ",") {
			chunk = strings.TrimSpace(chunk)
			if chunk == "" {
				continue
			}
			if strings.Contains(chunk, "-") {
				b := strings.SplitN(chunk, "-", 2)
				if len(b) != 2 {
					warnings = append(warnings, fmt.Sprintf("skipped invalid range %q", chunk))
					continue
				}
				lo, errLo := strconv.ParseUint(strings.TrimSpace(b[0]), 10, 32)
				hi, errHi := strconv.ParseUint(strings.TrimSpace(b[1]), 10, 32)
				if errLo != nil || errHi != nil {
					warnings = append(warnings, fmt.Sprintf("skipped invalid range %q", chunk))
					continue
				}
				if lo > hi {
					lo, hi = hi, lo
				}
				if hi-lo > 2_000_000 {
					warnings = append(warnings, fmt.Sprintf("range %q too large (>2M entries), skipped", chunk))
					continue
				}
				for p := lo; p <= hi; p++ {
					set[uint32(p)] = struct{}{}
				}
			} else {
				p, err := strconv.ParseUint(chunk, 10, 32)
				if err != nil {
					warnings = append(warnings, fmt.Sprintf("skipped invalid port %q", chunk))
					continue
				}
				set[uint32(p)] = struct{}{}
			}
		}
	}

	if exclude != "" {
		for _, chunk := range strings.Split(exclude, ",") {
			chunk = strings.TrimSpace(chunk)
			if chunk == "" {
				continue
			}
			if strings.Contains(chunk, "-") {
				b := strings.SplitN(chunk, "-", 2)
				if len(b) != 2 {
					continue
				}
				lo, errLo := strconv.ParseUint(strings.TrimSpace(b[0]), 10, 32)
				hi, errHi := strconv.ParseUint(strings.TrimSpace(b[1]), 10, 32)
				if errLo != nil || errHi != nil {
					continue
				}
				if lo > hi {
					lo, hi = hi, lo
				}
				for p := lo; p <= hi; p++ {
					delete(set, uint32(p))
				}
			} else {
				p, err := strconv.ParseUint(chunk, 10, 32)
				if err == nil {
					delete(set, uint32(p))
				}
			}
		}
	}

	if len(set) == 0 {
		return nil, warnings, fmt.Errorf("no ports after parsing")
	}

	out := make([]uint32, 0, len(set))
	for p := range set {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out, warnings, nil
}