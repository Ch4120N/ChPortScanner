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
