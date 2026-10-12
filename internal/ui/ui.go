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
