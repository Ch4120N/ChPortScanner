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
		rand.Shuffle(len(s.state.Targets), func(i, j int) {
			s.state.Targets[i], s.state.Targets[j] = s.state.Targets[j], s.state.Targets[i]
		})
		rand.Shuffle(len(s.ports), func(i, j int) {
			s.ports[i], s.ports[j] = s.ports[j], s.ports[i]
		})
		ui.Info("Scan order randomized")
	}

	// Totals
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

	// Shard
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
		ui.Info("Shard %s: jobs %d → %d", s.cfg.Shard, s.state.ShardIdx, s.state.ShardEnd)
	}

	// Writer
	s.writer, err = output.NewWriter(s.cfg)
	if err != nil {
		return err
	}
	defer s.writer.Close()

	// Rate limiter
	s.setupRateLimiter()

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	defer cancel()

	s.jobCh = make(chan Job, jobBufferSize(s.total, s.cfg))
	s.resultCh = make(chan output.Result, 512)

	s.start = time.Now()

	if !s.cfg.Quiet {
		s.progress = ui.NewProgress(&s.done, &s.open, s.total, s.start)
		go s.progress.Run(ctx)
	}

	go s.stateSaver(ctx)
	go s.resultCollector(ctx)

	// Worker count
	workers := computeWorkers(s.cfg, len(s.ports))
	ui.Info("Starting scan: %d workers | rate=%s | timeout=%ds | jobs=%d",
		workers, rateLabel(s.cfg.Rate), s.cfg.Timeout, s.total)

	s.wg.Add(1)
	go s.producer(ctx)

	for i := 0; i < workers; i++ {
		s.wg.Add(1)
		go s.worker(ctx)
	}

	s.wg.Wait()

	// Drain result channel
	close(s.resultCh)
	time.Sleep(100 * time.Millisecond)

	s.state.Finished = true
	s.saveState()

	if s.cfg.ShowStats {
		s.printStats()
	}
	if !s.cfg.Quiet {
		fmt.Fprintln(os.Stderr)
		ui.Success("Scan completed: %d open port(s)", s.open.Load())
	}
	return nil
}

func (s *Scanner) setupRateLimiter() {
	if s.cfg.Rate <= 0 {
		s.tokens = nil
		return
	}
	// Batch-fill: 10ms tick, fill rate/100 tokens per tick.
	// Keeps timer overhead low even at very high rates.
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
		}
	}
}

func (s *Scanner) printStats() {
	elapsed := time.Since(s.start).Seconds()
	done := s.done.Load()
	open := s.open.Load()
	rate := float64(done) / elapsed
	ui.Info("\n=== Final Stats ===")
	ui.Info("Scanned     : %d", done)
	ui.Info("Open        : %d", open)
	ui.Info("Elapsed     : %.2fs", elapsed)
	ui.Info("Avg rate    : %.0f pps", rate)
}

