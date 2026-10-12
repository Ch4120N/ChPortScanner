package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Ch4120N/ChPortScanner/internal/config"
	"github.com/Ch4120N/ChPortScanner/internal/output"
	"github.com/Ch4120N/ChPortScanner/internal/port"
	"github.com/Ch4120N/ChPortScanner/internal/target"
	"github.com/Ch4120N/ChPortScanner/internal/ui"
)

type Job struct {
	IP   string // pre-formatted string (avoid allocation in worker)
	Port uint16
}

type Scanner struct {
	cfg   *config.Config
	ports []uint16
	excl  []target.Spec
	state State

	total int64
	done  atomic.Uint64
	open  atomic.Uint64
	start time.Time

	jobCh    chan Job
	resultCh chan output.Result
	tokens   chan struct{}
	ticker   *time.Ticker

	wg       sync.WaitGroup
	cancel   context.CancelFunc
	stopOnce sync.Once
	writer   output.Writer
	progress *ui.Progress
}

func New(cfg *config.Config) *Scanner {
	return &Scanner{cfg: cfg}
}

func (s *Scanner) Stop() {
	s.stopOnce.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
		if s.ticker != nil {
			s.ticker.Stop()
		}
		s.saveState()
	})
}

func (s *Scanner) Run() error {
	targetStr := s.cfg.Target
	if s.cfg.InputFile != "" {
		lines, err := target.ReadLines(s.cfg.InputFile)
		if err != nil {
			return fmt.Errorf("reading input file: %w", err)
		}
		if len(lines) > 0 {
			if targetStr != "" {
				targetStr += ","
			}
			targetStr += strings.Join(lines, ",")
		}
	}
	if targetStr == "" && s.cfg.Resume == "" {
		return fmt.Errorf("no targets specified (use -t or -iL)")
	}

	specs, err := target.Parse(targetStr)
	if err != nil {
		return err
	}

	ports, err := port.Parse(s.cfg.Ports, s.cfg.ExcludePorts, s.cfg.TopPorts)
	if err != nil {
		return err
	}
	s.ports = ports

	if s.cfg.Resume != "" {
		if err := s.loadState(s.cfg.Resume); err != nil {
			return fmt.Errorf("loading resume state: %w", err)
		}
		specs = s.state.Targets
		ports, err = port.Parse(s.state.PortSpec, "", 0)
		if err != nil {
			return fmt.Errorf("resume ports: %w", err)
		}
		s.ports = ports
	} else {
		portStrs := make([]string, len(ports))
		for i, p := range ports {
			portStrs[i] = strconv.Itoa(int(p))
		}
		s.state = State{
			Targets:  specs,
			PortSpec: strings.Join(portStrs, ","),
			Format:   s.cfg.Format,
			Output:   s.cfg.Output,
		}
	}

	if s.state.Finished {
		ui.Info("Scan already finished per state file")
		return nil
	}

	// Excludes
	exclLines, err := target.ReadLines(s.cfg.ExcludeFile)
	if err != nil && s.cfg.ExcludeFile != "" {
		return fmt.Errorf("reading excludes: %w", err)
	}
	if len(exclLines) > 0 {
		s.excl, err = target.Parse(strings.Join(exclLines, ","))
		if err != nil {
			return fmt.Errorf("parsing excludes: %w", err)
		}
	}

	// Randomize
	if s.cfg.Random {
