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
