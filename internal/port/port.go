package port

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

var ServiceMap = map[uint16]string{
	20: "ftp-data", 21: "ftp", 22: "ssh", 23: "telnet", 25: "smtp", 53: "domain",
