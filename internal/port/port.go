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
