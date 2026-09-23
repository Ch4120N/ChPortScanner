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
