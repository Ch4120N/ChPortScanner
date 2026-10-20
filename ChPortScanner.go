/*
 * ChPortScanner - High-Speed Port Scanner for Restricted Lab Environments
 * Owner: Ch4120N
 * Version: 1.0.0
 *
 * Build with version metadata injection:
 *   go build -trimpath \
 *     -ldflags "-s -w \
 *       -X main.buildVersion=1.0.0 \
 *       -X main.buildTag=v1.0.0 \
 *       -X main.buildCommit=$(git rev-parse --short HEAD) \
 *       -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
 *     -o chportscanner .
 */

package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/Ch4120N/ChPortScanner/internal/config"
	"github.com/Ch4120N/ChPortScanner/internal/netutil"
	"github.com/Ch4120N/ChPortScanner/internal/scanner"
	"github.com/Ch4120N/ChPortScanner/internal/ui"
)


var (
	buildVersion = "1.0.0"    // e.g. "1.0.0"
	buildTag     = "dev"      // e.g. "v1.0.0"
	buildCommit  = "unknown"  // e.g. "a1b2c3d"
	buildDate    = "unknown"  // e.g. "2026-10-24T12:34:56Z"
)
