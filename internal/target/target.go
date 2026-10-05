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
