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

