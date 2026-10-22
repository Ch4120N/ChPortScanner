package scanner

import (
	"context"
	"net/netip"

	"github.com/Ch4120N/ChPortScanner/internal/target"
)

func (s *Scanner) producer(ctx context.Context) {
