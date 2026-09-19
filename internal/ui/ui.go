package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	Bold   = "\033[1m"
)

var (
	UseColor = true
	Quiet    = false
)

func Init(noColor, quiet bool) {
	if noColor {
		UseColor = false
	}
	Quiet = quiet
}

func Info(format string, a ...any) {
	if Quiet {
		return
	}
	msg := fmt.Sprintf(format, a...)
	if UseColor {
		fmt.Fprintf(os.Stderr, Cyan+"[i] "+Reset+msg+"\n")
	} else {
		fmt.Fprintf(os.Stderr, "[i] "+msg+"\n")
	}
}

func Success(format string, a ...any) {
	if UseColor {
		fmt.Fprintf(os.Stderr, Green+"[+] "+Reset+format+"\n", a...)
	} else {
		fmt.Fprintf(os.Stderr, "[+] "+format+"\n", a...)
	}
}

func Warn(format string, a ...any) {
	if UseColor {
		fmt.Fprintf(os.Stderr, Yellow+"[!] "+Reset+format+"\n", a...)
	} else {
		fmt.Fprintf(os.Stderr, "[!] "+format+"\n", a...)
	}
}

func Banner(name, version, owner string) {
	if Quiet {
		return
	}
	if !UseColor {
		fmt.Fprintf(os.Stderr, "== %s v%s by %s ==\n\n", name, version, owner)
		return
	}
	line := "══════════════════════════════════════════════════"
	fmt.Fprintf(os.Stderr, "%s%s╔%s╗%s\n", Bold, Cyan, line, Reset)
	fmt.Fprintf(os.Stderr, "%s%s║%s  %-46s%s%s║%s\n",
		Bold, Cyan, Reset, fmt.Sprintf("%s v%s", name, version), Bold, Cyan, Reset)
	fmt.Fprintf(os.Stderr, "%s%s║%s  %-46s%s%s║%s\n",
		Bold, Cyan, Reset, "Owner: "+owner, Bold, Cyan, Reset)
	fmt.Fprintf(os.Stderr, "%s%s╚%s╝%s\n\n", Bold, Cyan, line, Reset)
}

// ---- Progress bar ----
type Progress struct {
	done  *atomic.Uint64
	open  *atomic.Uint64
	total int64
	start time.Time
}

func NewProgress(done, open *atomic.Uint64, total int64, start time.Time) *Progress {
	return &Progress{done: done, open: open, total: total, start: start}
}

func (p *Progress) Run(ctx context.Context) {
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			d := p.done.Load()
			o := p.open.Load()
			elapsed := time.Since(p.start).Seconds()
			rate := float64(d) / elapsed
			percent := 0.0
			eta := ""
			if p.total > 0 {
				percent = float64(d) / float64(p.total) * 100
				if rate > 0 {
					rem := float64(p.total-int64(d)) / rate
					eta = fmt.Sprintf(" ETA:%s", time.Duration(rem*float64(time.Second)))
				}
			}
			barWidth := 20
			fill := int(percent / 100 * float64(barWidth))
			bar := make([]byte, barWidth)
			for i := 0; i < barWidth; i++ {
				if i < fill {
					bar[i] = '#'
				} else {
					bar[i] = '.'
				}
			}
			line := fmt.Sprintf("\r[%s] %6.2f%% | done:%d open:%d | %.0f pps%s",
				string(bar), percent, d, o, rate, eta)
			if UseColor {
				fmt.Fprintf(os.Stderr, "%s%s%s", Cyan, line, Reset)
			} else {
				fmt.Fprint(os.Stderr, line)
			}
		}
	}
}

// PrintHelp writes the full help to w.
func PrintHelp(w io.Writer, name, version, owner string) {
	fmt.Fprintf(w, "%s%s %s%s — High-Speed Port Scanner by %s\n", Bold, name, version, Reset, owner)
	fmt.Fprintf(w, "Designed for restricted lab environments (no raw sockets, no root required)\n\n")
	fmt.Fprintf(w, "%sUSAGE%s\n  %s [options] -t <IP/CIDR/range>\n\n", Yellow, Reset, name)

	sections := []struct {
		Title string
		Flags [][2]string
	}{
		{"TARGETING", [][2]string{
			{"-t, --target", "Targets: IPs/CIDRs/ranges (comma-separated)"},
			{"-iL, --input", "Input file with targets"},
			{"--exclude", "File with IPs/CIDRs to exclude"},
			{"-sn", "Ping sweep only"},
		}},
		{"PORTS", [][2]string{
			{"-p, --ports", "Ports to scan (e.g. 80,443,1000-2000)"},
			{"--top-ports N", "Scan top N common ports"},
			{"--exclude-ports", "Ports to exclude"},
			{"--profile", "quick|default|full|stealth"},
		}},
		{"SCANNING", [][2]string{
			{"--rate N", "Packets per second (0 = unlimited)"},
			{"--workers N", "Concurrent workers (0 = auto)"},
			{"--timeout N", "Connection timeout (seconds)"},
			{"--retries N", "Retries for failed connections"},
			{"--retry-delay", "Retry delay (ms)"},
			{"-u, --udp", "UDP scan mode"},
			{"--random", "Randomize scan order"},
			{"--shard X/Y", "Shard (e.g. 1/4)"},
			{"-S", "Source IP to bind"},
			{"--lite", "Low-resource mode"},
			{"--turbo", "Maximum speed mode"},
		}},
		{"PROBING", [][2]string{
			{"--banner", "Enable banner grabbing"},
			{"--banner-timeout", "Banner read timeout (s)"},
			{"--http-probe", "Fetch HTTP page title"},
			{"--tls-cert", "Extract TLS certificate info"},
			{"--rdns", "Reverse DNS lookup"},
		}},
		{"OUTPUT", [][2]string{
			{"-o, --output", "Output file (default stdout)"},
			{"-oJ", "JSON output"},
			{"--format", "text|json|jsonl|csv|grep"},
			{"--no-color", "Disable colored output"},
			{"-q, --quiet", "Quiet mode"},
			{"-V, --verbose", "Verbose mode"},
			{"--stats", "Show final stats"},
		}},
		{"UTILITIES", [][2]string{
			{"--extract FILE", "Extract unique IPs from output file"},
			{"--extract-ports", "Keep IP:port pairs on extraction"},
			{"--resume", "Resume from state file"},
			{"--state", "State file path"},
			{"-c", "JSON config file"},
			{"-h, --help", "Show this help"},
			{"-v, --version", "Show version"},
		}},
	}

	for _, s := range sections {
		fmt.Fprintf(w, "%s%s%s\n", Yellow, s.Title, Reset)
		for _, f := range s.Flags {
			fmt.Fprintf(w, "  %-22s %s\n", f[0], f[1])
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintf(w, "%sPERFORMANCE NOTES%s\n", Yellow, Reset)
	fmt.Fprintf(w, "  • Uses TCP connect scan (no root required, works where masscan is blocked)\n")
	fmt.Fprintf(w, "  • Applies SO_LINGER=0 to prevent TIME_WAIT exhaustion\n")
	fmt.Fprintf(w, "  • Auto-tunes worker count based on CPU count\n")
	fmt.Fprintf(w, "  • Auto-raises ulimit -n on Unix for more concurrent sockets\n")
	fmt.Fprintf(w, "  • Use --lite on 2-core / <4GB RAM systems\n\n")

	fmt.Fprintf(w, "%sEXAMPLES%s\n", Yellow, Reset)
	fmt.Fprintf(w, "  %s -t 192.168.1.0/24 -p 22,80,443\n", name)
	fmt.Fprintf(w, "  %s -t 10.0.0.0/24 -p 1-10000 --turbo --stats\n", name)
	fmt.Fprintf(w, "  %s -iL targets.txt --top-ports 1000 --rdns -oJ\n", name)
	fmt.Fprintf(w, "  %s -t 192.168.0.0/16 -p 22,80 --shard 1/4 --lite\n", name)
	fmt.Fprintf(w, "  %s --resume scan.state\n", name)
