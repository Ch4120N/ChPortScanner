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


const (
	AppName  = "ChPortScanner"
	AppOwner = "Ch4120N"
)


func main() {
	// Parse configuration (handles -h/--help, -v/--version, all flags)
	cfg, err := config.Load(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if cfg.ShowHelp {
		ui.PrintHelp(os.Stdout, AppName, buildVersion, AppOwner)
		return
	}

	if cfg.ShowVersion {
		printVersion()
		return
	}
	
	ui.Init(cfg.NoColor, cfg.Quiet)
	ui.Banner(AppName, buildVersion, AppOwner)
	
	if cfg.IsUnix {
		if ok, old, newLimit := netutil.RaiseUlimit(65535); ok && old != newLimit {
			ui.Info("Raised file descriptor limit: %d → %d", old, newLimit)
		}
	}
	
	if cfg.Verbose {
		ui.Info("Runtime : %s/%s", runtime.GOOS, runtime.GOARCH)
		ui.Info("Go      : %s", runtime.Version())
		ui.Info("CPUs    : %d", runtime.NumCPU())
		if buildTag != "dev" {
			ui.Info("Build   : %s (%s) built %s", buildTag, buildCommit, buildDate)
		}
	}

	s := scanner.New(cfg)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Fprintf(os.Stderr, "\n%s[!] Interrupted — saving state and exiting...%s\n",
			ui.Red, ui.Reset)
		s.Stop()
		os.Exit(0)
	}()

	if err := s.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%sError: %v%s\n", ui.Red, err, ui.Reset)
		os.Exit(1)
	}
}

// printVersion prints the full version banner including build metadata.
// The dev fallback is shown when the binary was built without ldflags.
func printVersion() {
	// Colored output if terminal supports it
	if ui.UseColor {
		fmt.Printf("%s%s%s v%s%s by %s%s%s\n",
			ui.Bold, ui.Cyan, AppName, buildVersion, ui.Reset,
			ui.Green, AppOwner, ui.Reset)
	} else {
		fmt.Printf("%s v%s by %s\n", AppName, buildVersion, AppOwner)
	}
