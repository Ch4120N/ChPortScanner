package output

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"

	"github.com/Ch4120N/ChPortScanner/internal/config"
)

// Result is a single open-port finding.
// The output shows only IP and Port (+ optional banner).
type Result struct {
	IP     string `json:"ip"`
	Port   int    `json:"port"`
	Banner string `json:"banner,omitempty"`
}

type Writer interface {
	Write(Result)
	Close()
}

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
	bw := bufio.NewWriterSize(raw, 64*1024)
	base := &bufferedWriter{bw: bw, file: f, done: make(chan struct{})}
	go base.flushLoop()

	switch cfg.Format {
	case "json":
		return &jsonWriter{base: base}, nil
	case "jsonl":
		return &jsonlWriter{base: base}, nil
	case "csv":
		cw := csv.NewWriter(bw)
		_ = cw.Write([]string{"ip", "port", "banner"})
		return &csvWriter{base: base, csv: cw}, nil
	default:
		return &textWriter{base: base}, nil
	}
}

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

type textWriter struct{ base *bufferedWriter }

func (t *textWriter) Write(r Result) {
	line := r.IP + ":" + itoa(r.Port)
	if r.Banner != "" {
		line += " " + r.Banner
	}
	t.base.writeLine(line + "\n")
}

func (t *textWriter) Close() { t.base.close() }

type jsonWriter struct{ base *bufferedWriter }

func (j *jsonWriter) Write(r Result) {
	b, _ := json.MarshalIndent(r, "", "  ")
	j.base.writeLine(string(b) + "\n")
}

func (j *jsonWriter) Close() { j.base.close() }

type jsonlWriter struct{ base *bufferedWriter }

func (j *jsonlWriter) Write(r Result) {
	b, _ := json.Marshal(r)
	j.base.writeLine(string(b) + "\n")
}

func (j *jsonlWriter) Close() { j.base.close() }

type csvWriter struct {
	base *bufferedWriter
	csv  *csv.Writer
}

func (c *csvWriter) Write(r Result) {
	c.base.mu.Lock()
	_ = c.csv.Write([]string{r.IP, itoa(r.Port), r.Banner})
	c.csv.Flush()
	c.base.mu.Unlock()
}

func (c *csvWriter) Close() { c.base.close() }

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