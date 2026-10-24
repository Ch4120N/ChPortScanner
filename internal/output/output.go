package output

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/Ch4120N/ChPortScanner/internal/config"
	"github.com/Ch4120N/ChPortScanner/internal/ui"
)

// Result is a single scan finding.
type Result struct {
	IP       string    `json:"ip"`
	Port     int       `json:"port,omitempty"`
	Protocol string    `json:"proto,omitempty"`
	State    string    `json:"state,omitempty"`
	Service  string    `json:"service,omitempty"`
	Banner   string    `json:"banner,omitempty"`
	Title    string    `json:"title,omitempty"`
	TLSInfo  string    `json:"tls,omitempty"`
	Hostname string    `json:"hostname,omitempty"`
	Time     time.Time `json:"ts,omitempty"`
}

// Writer is the output abstraction used by the scanner.
type Writer interface {
	Write(Result)
	Close()
}

// NewWriter builds the correct Writer for the configured format.
func NewWriter(cfg *config.Config) (Writer, error) {
	var raw io.Writer = os.Stdout
	var f *os.File
	if cfg.Output != "" {
		var err error
		f, err = os.OpenFile(cfg.Output, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		raw = f
	}
	isFile := cfg.Output != ""
	bw := bufio.NewWriterSize(raw, 64*1024)

	base := &bufferedWriter{
		bw:   bw,
		file: f,
		done: make(chan struct{}),
	}
	go base.flushLoop()

	switch cfg.Format {
	case "json":
		return &jsonWriter{base: base}, nil
	case "jsonl":
		return &jsonlWriter{base: base}, nil
	case "csv":
		cw := csv.NewWriter(bw)
		_ = cw.Write([]string{"ip", "port", "proto", "state", "service", "banner", "title", "tls", "hostname"})
		return &csvWriter{base: base, csv: cw}, nil
	case "grep":
		return &grepWriter{base: base}, nil
	default:
		return &textWriter{base: base, useColor: ui.UseColor && !isFile}, nil
	}
}

// ─── buffered writer ───
type bufferedWriter struct {
	mu     sync.Mutex
	bw     *bufio.Writer
	file   *os.File
	done   chan struct{}
	closed sync.Once
}

func (b *bufferedWriter) flushLoop() {
	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-b.done:
			return
		case <-t.C:
			b.mu.Lock()
			_ = b.bw.Flush()
			b.mu.Unlock()
		}
	}
}

func (b *bufferedWriter) writeLine(s string) {
	b.mu.Lock()
	_, _ = b.bw.WriteString(s)
	b.mu.Unlock()
}

func (b *bufferedWriter) close() {
	b.closed.Do(func() {
		close(b.done)
		b.mu.Lock()
		_ = b.bw.Flush()
		b.mu.Unlock()
		if b.file != nil {
			_ = b.file.Close()
		}
	})
}

// ─── text ───
type textWriter struct {
	base     *bufferedWriter
	useColor bool
}

func (t *textWriter) Write(r Result) {
	line := r.IP + ":" + itoa(r.Port)
	if r.Service != "" {
		line += " (" + r.Service + ")"
	}
	if r.Title != "" {
		line += " [title:" + r.Title + "]"
	}
	if r.TLSInfo != "" {
		line += " [tls:" + r.TLSInfo + "]"
	}
	if r.Hostname != "" {
		line += " [rdns:" + r.Hostname + "]"
	}
	if r.Banner != "" {
		line += " - " + r.Banner
	}
	if t.useColor {
		t.base.writeLine(ui.Green + line + ui.Reset + "\n")
	} else {
		t.base.writeLine(line + "\n")
	}
}

func (t *textWriter) Close() { t.base.close() }

// ─── json ───
type jsonWriter struct{ base *bufferedWriter }

func (j *jsonWriter) Write(r Result) {
	r.Time = time.Now()
	b, _ := json.MarshalIndent(r, "", "  ")
	j.base.writeLine(string(b) + "\n")
}

func (j *jsonWriter) Close() { j.base.close() }

// ─── jsonl ───
type jsonlWriter struct{ base *bufferedWriter }

func (j *jsonlWriter) Write(r Result) {
	r.Time = time.Now()
	b, _ := json.Marshal(r)
	j.base.writeLine(string(b) + "\n")
}

func (j *jsonlWriter) Close() { j.base.close() }

// ─── csv ───
type csvWriter struct {
	base *bufferedWriter
	csv  *csv.Writer
}

func (c *csvWriter) Write(r Result) {
	c.base.mu.Lock()
	_ = c.csv.Write([]string{
		r.IP, itoa(r.Port), r.Protocol, r.State,
		r.Service, r.Banner, r.Title, r.TLSInfo, r.Hostname,
	})
	c.csv.Flush()
	c.base.mu.Unlock()
}

func (c *csvWriter) Close() { c.base.close() }

// ─── grep ───
type grepWriter struct{ base *bufferedWriter }

func (g *grepWriter) Write(r Result) {
	g.base.writeLine(fmt.Sprintf("Host: %s () Ports: %d/%s/%s//%s///\n",
		r.IP, r.Port, r.State, r.Protocol, r.Service))
}

func (g *grepWriter) Close() { g.base.close() }

// ─── helpers ───
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [12]byte
	i := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}