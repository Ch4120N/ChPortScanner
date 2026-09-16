package target

import (
	"fmt"
	"math/big"
	"net"
	"net/netip"
	"os"
	"strings"
)

type Spec struct {
	Raw    string     `json:"raw"`
	Kind   string     `json:"kind"` // single | cidr | range
	IP     netip.Addr `json:"ip,omitempty"`
	Start  netip.Addr `json:"start,omitempty"`
	End    netip.Addr `json:"end,omitempty"`
	Prefix netip.Prefix `json:"prefix,omitempty"`
}

func Parse(raw string) ([]Spec, error) {
	if raw == "" {
		return nil, nil
	}
	parts := strings.FieldsFunc(raw, func(r rune) bool { return r == ',' || r == ' ' })
	out := make([]Spec, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		s, err := parseOne(p)
		if err != nil {
			return nil, fmt.Errorf("invalid target %q: %w", p, err)
		}
		out = append(out, s)
	}
	return out, nil
}

func parseOne(s string) (Spec, error) {
	if strings.Contains(s, "/") {
		pfx, err := netip.ParsePrefix(s)
		if err != nil {
			return Spec{}, err
		}
		return Spec{Raw: s, Kind: "cidr", Start: pfx.Masked().Addr(), End: lastInPrefix(pfx), Prefix: pfx}, nil
	}
	if strings.Contains(s, "-") {
		parts := strings.SplitN(s, "-", 2)
		start, err := netip.ParseAddr(strings.TrimSpace(parts[0]))
		if err != nil {
			return Spec{}, err
		}
		end, err := netip.ParseAddr(strings.TrimSpace(parts[1]))
		if err != nil {
			return Spec{}, err
		}
		if start.Compare(end) > 0 {
			start, end = end, start
		}
		return Spec{Raw: s, Kind: "range", Start: start, End: end}, nil
	}
	ip, err := netip.ParseAddr(s)
	if err != nil {
		return Spec{}, err
	}
	return Spec{Raw: s, Kind: "single", IP: ip}, nil
}

func lastInPrefix(p netip.Prefix) netip.Addr {
	if p.Addr().Is4() {
		ip := p.Masked().Addr().As4()
		mask := net.CIDRMask(p.Bits(), 32)
		for i := range ip {
			ip[i] |= ^mask[i]
		}
		return netip.AddrFrom4(ip)
	}
	ip16 := p.Masked().Addr().As16()
	mask16 := net.CIDRMask(p.Bits(), 128)
	for i := range ip16 {
		ip16[i] |= ^mask16[i]
	}
