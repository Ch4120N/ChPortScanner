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
