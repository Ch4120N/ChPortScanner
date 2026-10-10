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
