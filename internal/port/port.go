package port

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

var ServiceMap = map[uint16]string{
	20: "ftp-data", 21: "ftp", 22: "ssh", 23: "telnet", 25: "smtp", 53: "domain",
	67: "bootps", 68: "bootpc", 69: "tftp", 80: "http", 88: "kerberos",
	110: "pop3", 111: "rpcbind", 123: "ntp", 135: "msrpc", 137: "netbios-ns",
	138: "netbios-dgm", 139: "netbios-ssn", 143: "imap", 161: "snmp",
	162: "snmptrap", 179: "bgp", 389: "ldap", 443: "https", 445: "microsoft-ds",
	465: "smtps", 514: "syslog", 515: "printer", 548: "afp", 554: "rtsp",
	587: "submission", 631: "ipp", 636: "ldaps", 873: "rsync", 990: "ftps",
	993: "imaps", 995: "pop3s", 1433: "ms-sql-s", 1434: "ms-sql-m",
	1521: "oracle", 1723: "pptp", 2049: "nfs", 2082: "cpanel",
	2083: "cpanel-ssl", 2181: "zookeeper", 2375: "docker", 2376: "docker-tls",
	3000: "grafana", 3128: "squid", 3306: "mysql", 3389: "ms-wbt-server",
	5000: "upnp", 5432: "postgresql", 5900: "vnc", 5985: "wsman",
	5986: "wsmans", 6379: "redis", 6666: "irc", 8000: "http-alt",
	8080: "http-proxy", 8081: "blackice", 8443: "https-alt", 8888: "sun-answerbook",
	9000: "cslistener", 9090: "zeus-admin", 9200: "elasticsearch", 9300: "elasticsearch",
	27017: "mongodb", 28017: "mongodb-web", 50000: "sap",
}

var topPorts = []uint16{
	7, 9, 13, 21, 22, 23, 25, 26, 37, 53, 79, 80, 81, 88, 106, 110, 111, 113,
	119, 135, 139, 143, 144, 179, 199, 389, 427, 443, 444, 445, 465, 513, 514,
	515, 543, 544, 548, 554, 587, 631, 646, 873, 990, 993, 995, 1025, 1026,
	1027, 1028, 1029, 1110, 1433, 1720, 1723, 1755, 1900, 2000, 2001, 2049,
	2121, 2717, 3000, 3128, 3306, 3389, 3986, 4899, 5000, 5009, 5051, 5060,
	5101, 5190, 5357, 5432, 5631, 5666, 5800, 5900, 6000, 6001, 6646, 7070,
	8000, 8008, 8009, 8080, 8081, 8443, 8888, 9100, 9999, 10000, 32768,
	49152, 49153, 49154, 49155, 49156, 49157,
}

func init() {
	seen := map[uint16]bool{}
	deduped := topPorts[:0]
	for _, p := range topPorts {
		if !seen[p] {
			seen[p] = true
			deduped = append(deduped, p)
		}
	}
	topPorts = deduped
	sort.Slice(topPorts, func(i, j int) bool { return topPorts[i] < topPorts[j] })
}

func TopN(n int) []uint16 {
	if n <= 0 || n > len(topPorts) {
		n = len(topPorts)
	}
	return topPorts[:n]
}

func Parse(spec, exclude string, topN int) ([]uint16, error) {
	set := map[uint16]struct{}{}

	if topN > 0 {
		for _, p := range TopN(topN) {
			set[p] = struct{}{}
		}
	} else {
		if spec == "" {
			return nil, fmt.Errorf("no ports given")
		}
		for _, chunk := range strings.Split(spec, ",") {
			chunk = strings.TrimSpace(chunk)
			if chunk == "" {
				continue
			}
			if strings.Contains(chunk, "-") {
				b := strings.SplitN(chunk, "-", 2)
				lo, _ := strconv.Atoi(strings.TrimSpace(b[0]))
				hi, _ := strconv.Atoi(strings.TrimSpace(b[1]))
				if lo > hi {
					lo, hi = hi, lo
				}
				for p := lo; p <= hi; p++ {
					if p >= 1 && p <= 65535 {
						set[uint16(p)] = struct{}{}
					}
				}
			} else {
				p, err := strconv.Atoi(chunk)
				if err == nil && p >= 1 && p <= 65535 {
					set[uint16(p)] = struct{}{}
				}
			}
		}
	}

	if exclude != "" {
		for _, chunk := range strings.Split(exclude, ",") {
			chunk = strings.TrimSpace(chunk)
			if chunk == "" {
				continue
			}
			if strings.Contains(chunk, "-") {
				b := strings.SplitN(chunk, "-", 2)
				lo, _ := strconv.Atoi(strings.TrimSpace(b[0]))
