package probe

import (
	"crypto/tls"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
)

var (
	titleRe  = regexp.MustCompile(`(?is)<title>(.*?)</title>`)
	serverRe = regexp.MustCompile(`(?i)server:\s*(.+)`)
)

// HTTPTitle fetches the page title from a web port.
func HTTPTitle(ip string, port int, timeout time.Duration) string {
	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return ""
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))
	fmt.Fprintf(conn, "GET / HTTP/1.1\r\nHost: %s\r\nUser-Agent: ChPortScanner/1.0\r\nConnection: close\r\n\r\n", ip)
	buf := make([]byte, 4096)
	n, _ := conn.Read(buf)
	if n <= 0 {
		return ""
	}
	body := string(buf[:n])
	if m := titleRe.FindStringSubmatch(body); len(m) > 1 {
		t := strings.TrimSpace(m[1])
		if len(t) > 80 {
			t = t[:80] + "..."
		}
		return t
	}
	if m := serverRe.FindStringSubmatch(body); len(m) > 1 {
		return "server:" + strings.TrimSpace(m[1])
	}
	return ""
}

// TLSCertInfo returns CN + issuer + expiry.
func TLSCertInfo(ip string, port int, timeout time.Duration) string {
	dialer := &net.Dialer{Timeout: timeout}
	conn, err := tls.DialWithDialer(dialer, "tcp",
		fmt.Sprintf("%s:%d", ip, port),
		&tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return ""
	}
	defer conn.Close()
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return ""
	}
	c := certs[0]
	var parts []string
	if c.Subject.CommonName != "" {
		parts = append(parts, "CN="+c.Subject.CommonName)
	}
	if c.Issuer.CommonName != "" {
		parts = append(parts, "issuer="+c.Issuer.CommonName)
