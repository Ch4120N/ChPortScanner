package scanner

import (
	"context"
	"net/netip"

	"github.com/Ch4120N/ChPortScanner/internal/target"
)

func (s *Scanner) producer(ctx context.Context) {
	defer s.wg.Done()
	defer close(s.jobCh)

	specs := s.state.Targets
	ports := s.ports
	specIdx := s.state.CurSpec
	portIdx := s.state.CurPort
	shardStart := s.state.ShardIdx
	shardEnd := s.state.ShardEnd
	count := int64(0)

	var curIP netip.Addr
	if s.state.CurIP != "" {
		if ip, err := netip.ParseAddr(s.state.CurIP); err == nil {
			curIP = ip
		}
	}
	if !curIP.IsValid() && specIdx < len(specs) {
		if specs[specIdx].Kind == "single" {
			curIP = specs[specIdx].IP
		} else {
			curIP = specs[specIdx].Start
		}
	}

	// Pre-size the string buffer used for ip.String() to minimize allocations.
	emit := func(ip netip.Addr) bool {
		if target.IsExcluded(ip, s.excl) {
			count += int64(len(ports))
			portIdx = 0
			return true
		}
		// Format IP once per host, reuse for all ports.
		ipStr := ip.String()
		for portIdx < len(ports) {
			if count >= shardEnd {
				return false
			}
			if count >= shardStart {
				select {
				case s.jobCh <- Job{IP: ipStr, Port: ports[portIdx]}:
				case <-ctx.Done():
					return false
				}
			}
			count++
			portIdx++
		}
		portIdx = 0
		s.state.CurIP = ipStr
		s.state.CurPort = 0
		return true
	}
