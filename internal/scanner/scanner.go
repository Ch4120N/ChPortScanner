package scanner

import (
	"context"
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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Ch4120N/ChPortScanner/internal/config"
	"github.com/Ch4120N/ChPortScanner/internal/output"
	"github.com/Ch4120N/ChPortScanner/internal/port"
	"github.com/Ch4120N/ChPortScanner/internal/target"
	"github.com/Ch4120N/ChPortScanner/internal/ui"
)

type Job struct {
	IP   string
	Port uint32
}

type Scanner struct {
	cfg   *config.Config
	ports []uint32
	excl  []target.Spec
	state State

	total   int64
	done    atomic.Uint64
	open    atomic.Uint64
	retries atomic.Uint64
	start   time.Time

	jobCh    chan Job
	resultCh chan output.Result
	tokens   chan struct{}
	ticker   *time.Ticker

	wg            sync.WaitGroup
	cancel        context.CancelFunc
	stopOnce      sync.Once
	writer        output.Writer
	progress      *ui.Progress
	dashboardProg *tea.Program
	dashboardDone chan struct{}
	dashQuitOnce  sync.Once
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

	ports, warnings, err := port.Parse(s.cfg.Ports, s.cfg.ExcludePorts, s.cfg.TopPorts)
	if err != nil {
		return err
	}
	s.ports = ports
	for _, w := range warnings {
		ui.Warn("%s", w)
	}

	if s.cfg.Resume != "" {
		if err := s.loadState(s.cfg.Resume); err != nil {
			return fmt.Errorf("loading resume state: %w", err)
		}
		specs = s.state.Targets
		ports, _, err = port.Parse(s.state.PortSpec, "", 0)
		if err != nil {
			return fmt.Errorf("resume ports: %w", err)
		}
		s.ports = ports
	} else {
		portStrs := make([]string, len(ports))
		for i, p := range ports {
			portStrs[i] = strconv.FormatUint(uint64(p), 10)
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

	if s.cfg.Random {
		rand.Shuffle(len(s.state.Targets), func(i, j int) {
			s.state.Targets[i], s.state.Targets[j] = s.state.Targets[j], s.state.Targets[i]
		})
		rand.Shuffle(len(s.ports), func(i, j int) {
			s.ports[i], s.ports[j] = s.ports[j], s.ports[i]
		})
	}

	var totalIPs uint64
	for _, sp := range s.state.Targets {
		totalIPs += target.Count(sp)
	}
	if totalIPs == 0 || len(s.ports) == 0 {
		return fmt.Errorf("empty target or port set")
	}
	if totalIPs > math.MaxInt64/uint64(len(s.ports)) {
		s.total = math.MaxInt64
	} else {
		s.total = int64(totalIPs * uint64(len(s.ports)))
	}

	s.state.ShardIdx = 0
	s.state.ShardEnd = s.total
	if s.cfg.Shard != "" {
		parts := strings.SplitN(s.cfg.Shard, "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid shard (use X/Y)")
		}
		x, _ := strconv.Atoi(parts[0])
		y, _ := strconv.Atoi(parts[1])
		if y < 1 || x < 1 || x > y {
			return fmt.Errorf("invalid shard values")
		}
		chunk := s.total / int64(y)
		s.state.ShardIdx = int64(x-1) * chunk
		if x == y {
			s.state.ShardEnd = s.total
		} else {
			s.state.ShardEnd = int64(x) * chunk
		}
	}

	s.writer, err = output.NewWriter(s.cfg)
	if err != nil {
		return err
	}
	defer s.writer.Close()

	s.setupRateLimiter()

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	defer cancel()

	s.jobCh = make(chan Job, jobBufferSize(s.total, s.cfg))
	s.resultCh = make(chan output.Result, 512)

	s.start = time.Now()
	workers := computeWorkers(s.cfg, len(s.ports))

	// ─── TUI dashboard ────────────────────────────────────────────
	useTUI := !s.cfg.NoTUI && !s.cfg.Quiet && isTTY()
	if useTUI {
		dcfg := ui.DashboardConfig{
			Targets: targetDisplay(s.cfg),
			Output:  s.cfg.Output,
			Mode:    dashboardMode(s.cfg),
			Workers: workers,
			Banner:  s.cfg.Banner,
			Total:   s.total,
			Start:   s.start,
			Tried:   &s.done,
			Open:    &s.open,
			Retries: &s.retries,
		}
		dash := ui.NewDashboard(dcfg, ui.UseColor)
		// KEY FIX: q / esc / ctrl+c must stop the SCAN, not just the TUI.
		dash.OnQuit = func() {
			s.Stop()
		}
		s.dashboardProg = tea.NewProgram(dash, tea.WithAltScreen())
		s.dashboardDone = make(chan struct{})
		go func() {
			if _, err := s.dashboardProg.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "dashboard error: %v\n", err)
			}
			close(s.dashboardDone)
		}()
	} else if !s.cfg.Quiet {
		s.progress = ui.NewProgress(&s.done, &s.open, s.total, s.start)
		go s.progress.Run(ctx)
		ui.Info("Ports: %d | Workers: %d | Rate: %s | Timeout: %ds | Jobs: %d",
			len(s.ports), workers, rateLabel(s.cfg.Rate), s.cfg.Timeout, s.total)
	}

	go s.stateSaver(ctx)
	go s.resultCollector(ctx)

	s.wg.Add(1)
	go s.producer(ctx)
	for i := 0; i < workers; i++ {
		s.wg.Add(1)
		go s.worker(ctx)
	}

	// Wait for scan to finish (naturally or via Stop)
	s.wg.Wait()
	close(s.resultCh)
	time.Sleep(100 * time.Millisecond)

	// Ensure TUI is closed cleanly (handles both "user pressed q" and "scan finished")
	s.closeDashboard()

	s.state.Finished = true
	s.saveState()

	if s.cfg.ShowStats {
		s.printStats()
	}
	if !s.cfg.Quiet {
		open := s.open.Load()
		if open == 0 {
			ui.Warn("No open ports found across %d attempts.", s.done.Load())
		} else {
			ui.Success("Scan completed: %d open port(s)", open)
		}
		fmt.Fprintf(os.Stderr, "Results written to: %s\n", s.cfg.Output)
	}
	return nil
}

// closeDashboard triggers TUI shutdown (idempotent) and waits for it to exit.
func (s *Scanner) closeDashboard() {
	if s.dashboardProg == nil {
		return
	}
	s.dashQuitOnce.Do(func() {
		s.dashboardProg.Quit()
	})
	select {
	case <-s.dashboardDone:
	case <-time.After(2 * time.Second):
		// Fallback so we never hang
	}
}

func (s *Scanner) setupRateLimiter() {
	if s.cfg.Rate <= 0 {
		s.tokens = nil
		return
	}
	rate := s.cfg.Rate
	s.tokens = make(chan struct{}, rate)
	batchSize := rate / 100
	interval := 10 * time.Millisecond
	if batchSize < 1 {
		batchSize = 1
		interval = time.Second / time.Duration(rate)
	}
	s.ticker = time.NewTicker(interval)
	go func() {
		for range s.ticker.C {
			for i := 0; i < batchSize; i++ {
				select {
				case s.tokens <- struct{}{}:
				default:
					break
				}
			}
		}
	}()
}

func (s *Scanner) stateSaver(ctx context.Context) {
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.saveState()
		}
	}
}

func (s *Scanner) resultCollector(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case res, ok := <-s.resultCh:
			if !ok {
				return
			}
			s.writer.Write(res)
			if s.dashboardProg != nil {
				s.dashboardProg.Send(ui.FoundMsg{
					IP:     res.IP,
					Port:   res.Port,
					Banner: res.Banner,
				})
			}
		}
	}
}

func (s *Scanner) printStats() {
	elapsed := time.Since(s.start).Seconds()
	done := s.done.Load()
	open := s.open.Load()
	rate := float64(done) / elapsed
	ui.Info("\n=== Final Stats ===")
	ui.Info("Attempts    : %d", done)
	ui.Info("Open        : %d", open)
	ui.Info("Elapsed     : %.2fs", elapsed)
	ui.Info("Avg rate    : %.0f pps", rate)
}

func computeWorkers(cfg *config.Config, numPorts int) int {
	if cfg.Workers > 0 {
		return cfg.Workers
	}
	cpus := runtime.NumCPU()
	var w int
	switch {
	case cfg.Lite || cpus <= 2:
		w = 200
	case cpus <= 4:
		w = 1000
	case cpus <= 8:
		w = 2500
	case cpus <= 16:
		w = 4000
	default:
		w = 6000
	}
	if cfg.Turbo {
		w *= 2
	}
	if w > 10000 {
		w = 10000
	}
	if w < 50 {
		w = 50
	}
	if cfg.Rate > 0 && w > cfg.Rate {
		w = cfg.Rate
	}
	if cfg.Rate == 0 && numPorts > 0 && w > 5000 {
		w = 5000
	}
	return w
}

func jobBufferSize(total int64, cfg *config.Config) int {
	if cfg.Lite {
		return 5000
	}
	if total <= 0 {
		return 10000
	}
	if total < 50000 {
		return int(total)
	}
	return 50000
}

func rateLabel(rate int) string {
	if rate <= 0 {
		return "unlimited"
	}
	return strconv.Itoa(rate)
}

func dashboardMode(cfg *config.Config) string {
	if cfg.Turbo {
		return "TURBO"
	}
	if cfg.Lite {
		return "LITE"
	}
	if cfg.Profile != "" {
		return strings.ToUpper(cfg.Profile)
	}
	if cfg.UDP {
		return "UDP"
	}
	return "NORMAL"
}

func targetDisplay(cfg *config.Config) string {
	if cfg.InputFile != "" && cfg.Target == "" {
		return cfg.InputFile
	}
	return cfg.Target
}

func isTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
