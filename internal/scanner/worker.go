package scanner

import (
	"context"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Ch4120N/ChPortScanner/internal/netutil"
	"github.com/Ch4120N/ChPortScanner/internal/output"
	"github.com/Ch4120N/ChPortScanner/internal/port"
	"github.com/Ch4120N/ChPortScanner/internal/probe"
)

func (s *Scanner) worker(ctx context.Context) {
	defer s.wg.Done()

	timeout := time.Duration(s.cfg.Timeout) * time.Second
	if timeout <= 0 {
		timeout = time.Second
	}

	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: -1, // disable TCP keepalive — saves syscalls
		Control:   netutil.SetFastClose,
	}
	if s.cfg.SourceIP != "" {
		if ip := net.ParseIP(s.cfg.SourceIP); ip != nil {
			if s.cfg.UDP {
				dialer.LocalAddr = &net.UDPAddr{IP: ip}
			} else {
				dialer.LocalAddr = &net.TCPAddr{IP: ip}
			}
		}
	}

	network := "tcp"
	if s.cfg.UDP {
		network = "udp"
	}

