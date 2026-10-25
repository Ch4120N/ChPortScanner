package scanner

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/Ch4120N/ChPortScanner/internal/netutil"
	"github.com/Ch4120N/ChPortScanner/internal/output"
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
		KeepAlive: -1,
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

	var addrBuf [64]byte

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-s.jobCh:
			if !ok {
				return
			}

			if s.tokens != nil {
				select {
				case <-s.tokens:
				case <-ctx.Done():
					return
				}
			}

			// Wrap the port: TCP/UDP fields are 16 bits wide.
			// Any uint32 is sent modulo 65536 on the wire; the
			// original value is preserved in the result.
			wirePort := uint16(job.Port % 65536)
			addr := buildAddr(addrBuf[:], job.IP, wirePort)

			var conn net.Conn
			var open bool
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
					s.retries.Add(1)
					time.Sleep(retryDelay)
				}
			}

			if !open {
				s.done.Add(1)
				continue
			}

			res := output.Result{
				IP:   job.IP,
				Port: int(job.Port),
			}

			if s.cfg.Banner && network == "tcp" {
				res.Banner = grabBanner(conn, s.cfg.BannerTimeout)
			}
			conn.Close()

			if s.cfg.HTTPProbe {
				if t := probe.HTTPTitle(job.IP, int(wirePort), timeout); t != "" {
					res.Banner = joinBanner(res.Banner, t)
				}
			}
			if s.cfg.TLSCert {
				if t := probe.TLSCertInfo(job.IP, int(wirePort), timeout); t != "" {
					res.Banner = joinBanner(res.Banner, t)
				}
			}
			if s.cfg.RDNS {
				if host := probe.RDNS(job.IP); host != "" {
					res.Banner = joinBanner(res.Banner, "rdns:"+host)
				}
			}

			select {
			case s.resultCh <- res:
			case <-ctx.Done():
				return
			}
			s.open.Add(1)
			s.done.Add(1)
		}
	}
}

func joinBanner(a, b string) string {
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	return a + " " + b
}

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
	var tmp [5]byte
	i := len(tmp)
	for v > 0 {
		i--
		tmp[i] = byte('0' + v%10)
		v /= 10
	}
	return copy(dst, tmp[i:])
}

func grabBanner(conn net.Conn, timeoutSec int) string {
	_ = conn.SetReadDeadline(time.Now().Add(time.Duration(timeoutSec) * time.Second))
	buf := make([]byte, 512)
	n, _ := conn.Read(buf)
	if n <= 0 {
		return ""
	}
	b := strings.TrimSpace(string(buf[:n]))
	b = strings.ReplaceAll(b, "\n", "\\n")
	b = strings.ReplaceAll(b, "\r", "\\r")
	if len(b) > 200 {
		b = b[:200] + "..."
	}
	return b
}