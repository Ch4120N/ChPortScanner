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

	retryDelay := time.Duration(s.cfg.RetryDelay) * time.Millisecond
	retries := s.cfg.Retries

	// Small scratch buffer for building addresses (avoids JoinHostPort allocation)
	var addrBuf [64]byte

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-s.jobCh:
			if !ok {
				return
			}

			// Rate limiting (only if configured)
			if s.tokens != nil {
				select {
				case <-s.tokens:
				case <-ctx.Done():
					return
				}
			}

			// Build "ip:port" without allocation
			addr := buildAddr(addrBuf[:], job.IP, job.Port)

			var open bool
			var conn net.Conn

			for attempt := 0; attempt <= retries; attempt++ {
				c, err := dialer.DialContext(ctx, network, addr)
				if err == nil {
					conn = c
					open = true
					break
				}
				if s.cfg.UDP && strings.Contains(err.Error(), "connection refused") {
					break
				}
				if attempt < retries {
					time.Sleep(retryDelay)
				}
			}

			if open && conn != nil {
				res := output.Result{
					IP:       job.IP,
					Port:     int(job.Port),
					Protocol: network,
					State:    "open",
				}
				if svc, ok := port.ServiceMap[job.Port]; ok {
					res.Service = svc
				}

				if s.cfg.Banner && network == "tcp" {
					res.Banner = grabBanner(conn, s.cfg.BannerTimeout)
				}
				conn.Close()

				// Optional probes
				if s.cfg.HTTPProbe {
					res.Title = probe.HTTPTitle(job.IP, int(job.Port), timeout)
				}
				if s.cfg.TLSCert {
					res.TLSInfo = probe.TLSCertInfo(job.IP, int(job.Port), timeout)
				}
				if s.cfg.RDNS {
					res.Hostname = probe.RDNS(job.IP)
				}

				select {
				case s.resultCh <- res:
				case <-ctx.Done():
					return
				}
				s.open.Add(1)
			}
			s.done.Add(1)
		}
	}
}

// buildAddr formats "ip:port" into dst without allocation.
func buildAddr(dst []byte, ip string, port uint16) string {
	n := copy(dst, ip)
	dst[n] = ':'
	n++
	n += appendUint16(dst[n:], port)
	return string(dst[:n])
}

func appendUint16(dst []byte, v uint16) int {
	if v == 0 {
		dst[0] = '0'
		return 1
	}
