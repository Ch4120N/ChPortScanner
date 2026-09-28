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
