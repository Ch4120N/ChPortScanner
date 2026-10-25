package ui

import (
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ═══════════════════════════════════════════════════════════════════
//  PUBLIC TYPES
// ═══════════════════════════════════════════════════════════════════

type DashboardConfig struct {
	Targets string
	Output  string
	Mode    string
	Workers int
	Banner  bool
	Total   int64
	Start   time.Time

	Tried   *atomic.Uint64
	Open    *atomic.Uint64
	Retries *atomic.Uint64
}

type FoundMsg struct {
	IP     string
	Port   int
	Banner string
}

type Dashboard struct {
	cfg DashboardConfig

	found []FoundMsg

	lastTried   uint64
	lastTime    time.Time
	currentRate float64

	width, height int
	ready         bool
	stopping      bool
	useColor      bool

	// Callback invoked on user quit (q/esc/ctrl+c)
	OnQuit func()

	// Styles
	artStyle, footerStyle, labelStyle, valueStyle lipgloss.Style
	valueBold, titleStyle, barFill, barEmpty       lipgloss.Style
	foundIP, foundBanner, helpStyle                lipgloss.Style
}

func NewDashboard(cfg DashboardConfig, useColor bool) *Dashboard {
	d := &Dashboard{cfg: cfg, useColor: useColor}
	d.setupStyles()
	d.lastTime = time.Now()
	return d
}

// ═══════════════════════════════════════════════════════════════════
//  MESSAGES
// ═══════════════════════════════════════════════════════════════════

type tickMsg time.Time
type clearFrameMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// ═══════════════════════════════════════════════════════════════════
//  BUBBLETEA MODEL
// ═══════════════════════════════════════════════════════════════════

func (m *Dashboard) Init() tea.Cmd {
	return tea.Batch(tea.SetWindowTitle("ChPortScanner"), tickCmd())
}

func (m *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		oldW, oldH := m.width, m.height
		m.width, m.height = msg.Width, msg.Height
		m.ready = true
		if oldW != msg.Width || oldH != msg.Height {
			// Hard reset — exit/re-enter alt-screen + clear + one blank frame
			return m, tea.Sequence(
				tea.ExitAltScreen,
				tea.ClearScreen,
				tea.EnterAltScreen,
			)
		}
		return m, nil

	case clearFrameMsg:
		// No-op — forces a re-render on the next tick
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			if !m.stopping {
				m.stopping = true
				if m.OnQuit != nil {
					m.OnQuit()
				}
			}
			return m, tea.Quit
		}

	case tickMsg:
		now := time.Now()
		tried := m.cfg.Tried.Load()
		dt := now.Sub(m.lastTime).Seconds()
		if dt > 0 {
			instant := float64(tried-m.lastTried) / dt
			if m.currentRate == 0 {
				m.currentRate = instant
			} else {
				m.currentRate = 0.4*instant + 0.6*m.currentRate
			}
			m.lastTried = tried
			m.lastTime = now
		}
		return m, tickCmd()

	case FoundMsg:
		m.found = append(m.found, msg)
		if len(m.found) > 500 {
			m.found = m.found[len(m.found)-500:]
		}
		return m, nil
	}
	return m, nil
}

// ═══════════════════════════════════════════════════════════════════
//  VIEW  —  EXACTLY m.height lines, every section returns []string
// ═══════════════════════════════════════════════════════════════════

func (m *Dashboard) View() string {
	if !m.ready {
		return "Initializing..."
	}
	if m.height < 5 || m.width < 30 {
		return "Terminal too small — resize to view dashboard"
	}

	var lines []string

	if m.showBanner() {
		lines = append(lines, m.bannerLines()...)
	}
	lines = append(lines, m.configLines()...)
	lines = append(lines, m.progressLines()...)
	lines = append(lines, m.statsLines()...)

	// Available lines left for the found list
	used := len(lines) + 1 // +1 for footer
	remaining := m.height - used
	if remaining > 7 { // max 5 items + title + trailing blank
		remaining = 7
	}
	if remaining >= 3 && len(m.found) > 0 {
		maxItems := remaining - 2
		if maxItems > 5 {
			maxItems = 5
		}
		if maxItems > 0 {
			lines = append(lines, m.foundLines(maxItems)...)
		}
	}
	lines = append(lines, m.footerLines()...)

	// Force exactly m.height lines — no more, no less
	if len(lines) > m.height {
		lines = lines[:m.height]
	}
	for len(lines) < m.height {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

// ═══════════════════════════════════════════════════════════════════
//  SECTION RENDERERS  —  each returns one entry per display line
// ═══════════════════════════════════════════════════════════════════

func (m *Dashboard) bannerLines() []string {
	out := make([]string, 0, len(BannerArt)+2)
	for _, line := range BannerArt {
		out = append(out, m.artStyle.Render(line))
	}
	footer := fmt.Sprintf("%s v%s . Owned By %s", dashAppName, dashAppVersion, dashAppOwner)
	out = append(out, m.footerStyle.Render(footer))
	out = append(out, "")
	return out
}

func (m *Dashboard) configLines() []string {
	targets := truncate(m.cfg.Targets, maxInt(m.width-22, 15))
	output := truncate(m.cfg.Output, maxInt(m.width-22, 15))

	if m.width < 80 {
		return []string{
			m.kv("Targets", targets),
			m.kv("Output", output),
			m.kv("Mode", m.cfg.Mode),
			m.kv("Workers", strconv.Itoa(m.cfg.Workers)),
			"",
		}
	}

	colW := m.width / 2
	left := []string{
		m.kv("Targets", targets),
		m.kv("Output", output),
	}
	right := []string{
		m.kv("Mode", m.cfg.Mode),
		m.kv("Workers", strconv.Itoa(m.cfg.Workers)),
	}
	return []string{
		padRight(left[0], colW) + right[0],
		padRight(left[1], colW) + right[1],
		"",
	}
}

func (m *Dashboard) progressLines() []string {
	tried := m.cfg.Tried.Load()
	percent := 0.0
	if m.cfg.Total > 0 {
		percent = float64(tried) / float64(m.cfg.Total) * 100
	}
	if percent > 100 {
		percent = 100
	}

	barWidth := m.width - 11
	if barWidth < 20 {
		barWidth = 20
	}
	if barWidth > 100 {
		barWidth = 100
	}
	fill := int(float64(barWidth) * percent / 100)
	if fill < 0 {
		fill = 0
	}
	if fill > barWidth {
		fill = barWidth
	}

	fillStr := m.barFill.Render(strings.Repeat("#", fill))
	emptyStr := m.barEmpty.Render(strings.Repeat(".", barWidth-fill))
	pct := m.valueBold.Render(fmt.Sprintf("  %6.2f%%", percent))

	return []string{
		"[" + fillStr + emptyStr + "]" + pct,
		"",
	}
}

func (m *Dashboard) statsLines() []string {
	tried := m.cfg.Tried.Load()
	open := m.cfg.Open.Load()
	closed := uint64(0)
	if tried > open {
		closed = tried - open
	}
	retries := m.cfg.Retries.Load()

	elapsed := time.Since(m.cfg.Start)
	avg := 0.0
	if elapsed.Seconds() > 0 {
		avg = float64(tried) / elapsed.Seconds()
	}

	var remaining time.Duration
	if m.currentRate > 0 && int64(tried) < m.cfg.Total {
		sec := float64(m.cfg.Total-int64(tried)) / m.currentRate
		remaining = time.Duration(sec * float64(time.Second))
	}

	type row struct{ label, value string }
	left := []row{
		{"Tried", comma(tried) + " / " + comma(uint64(m.cfg.Total))},
		{"Open", comma(open)},
		{"Closed", comma(closed)},
		{"Retries", comma(retries)},
	}
	right := []row{
		{"Elapsed", dur(elapsed)},
		{"Remaining", dur(remaining)},
		{"Rate", rateFmt(m.currentRate)},
		{"Avg", rateFmt(avg)},
	}

	if m.width < 80 {
		out := make([]string, 0, 9)
		for i := 0; i < 4; i++ {
			out = append(out, m.kv(left[i].label, left[i].value))
			out = append(out, m.kv(right[i].label, right[i].value))
		}
		out = append(out, "")
		return out
	}

	colW := m.width / 2
	out := make([]string, 0, 5)
	for i := 0; i < 4; i++ {
		out = append(out, padRight(m.kv(left[i].label, left[i].value), colW)+m.kv(right[i].label, right[i].value))
	}
	out = append(out, "")
	return out
}

func (m *Dashboard) foundLines(max int) []string {
	start := 0
	if len(m.found) > max {
		start = len(m.found) - max
	}

	out := []string{m.titleStyle.Render("Found:")}
	for _, f := range m.found[start:] {
		ipPort := f.IP + ":" + strconv.Itoa(f.Port)
		line := "  >  " + m.foundIP.Render(ipPort)
		if f.Banner != "" && m.cfg.Banner {
			maxB := m.width - 30
			if maxB < 12 {
				maxB = 12
			}
			line += "  " + m.valueStyle.Render("|  ") + m.foundBanner.Render(truncate(f.Banner, maxB))
		}
		out = append(out, line)
	}
	out = append(out, "")
	return out
}

func (m *Dashboard) footerLines() []string {
	if m.stopping {
		return []string{m.helpStyle.Copy().Foreground(lipgloss.Color("214")).Bold(true).Render("Stopping... please wait")}
	}
	return []string{m.helpStyle.Render("q/esc: quit  •  ctrl+c: stop scan")}
}

// ═══════════════════════════════════════════════════════════════════
//  LAYOUT HELPERS
// ═══════════════════════════════════════════════════════════════════

func (m *Dashboard) showBanner() bool {
	return m.height >= 27 && m.width >= 74
}

func (m *Dashboard) setupStyles() {
	m.artStyle = lipgloss.NewStyle()
	m.footerStyle = lipgloss.NewStyle()
	m.labelStyle = lipgloss.NewStyle()
	m.valueStyle = lipgloss.NewStyle()
	m.valueBold = lipgloss.NewStyle()
	m.titleStyle = lipgloss.NewStyle().Bold(true)
	m.barFill = lipgloss.NewStyle()
	m.barEmpty = lipgloss.NewStyle()
	m.foundIP = lipgloss.NewStyle()
	m.foundBanner = lipgloss.NewStyle()
	m.helpStyle = lipgloss.NewStyle().Italic(true)

	if m.useColor {
		m.artStyle = m.artStyle.Bold(true).Foreground(lipgloss.Color("36"))
		m.footerStyle = m.footerStyle.Bold(true).Foreground(lipgloss.Color("36"))
		m.labelStyle = m.labelStyle.Bold(true).Foreground(lipgloss.Color("244"))
		m.valueStyle = m.valueStyle.Foreground(lipgloss.Color("252"))
		m.valueBold = m.valueBold.Bold(true).Foreground(lipgloss.Color("81"))
		m.titleStyle = m.titleStyle.Foreground(lipgloss.Color("201"))
		m.barFill = m.barFill.Foreground(lipgloss.Color("42"))
		m.barEmpty = m.barEmpty.Foreground(lipgloss.Color("240"))
		m.foundIP = m.foundIP.Foreground(lipgloss.Color("82"))
		m.foundBanner = m.foundBanner.Foreground(lipgloss.Color("244"))
		m.helpStyle = m.helpStyle.Foreground(lipgloss.Color("240"))
	}
}

// ═══════════════════════════════════════════════════════════════════
//  FORMATTERS
// ═══════════════════════════════════════════════════════════════════

func (m *Dashboard) kv(label, value string) string {
	lbl := fmt.Sprintf("%-10s", label+":")
	return m.labelStyle.Render(lbl) + "  " + m.valueStyle.Render(value)
}

func padRight(s string, width int) string {
	gap := width - lipgloss.Width(s)
	if gap <= 0 {
		return s
	}
	return s + strings.Repeat(" ", gap)
}

func truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if len(s) <= n {
		return s
	}
	if n < 4 {
		return s[:n]
	}
	return s[:n-3] + "..."
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func comma(n uint64) string {
	s := strconv.FormatUint(n, 10)
	if len(s) <= 3 {
		return s
	}
	var out []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, byte(c))
	}
	return string(out)
}

func dur(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	h := int(d.Hours())
	mi := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, mi, s)
}

func rateFmt(r float64) string {
	if r >= 1_000_000 {
		return fmt.Sprintf("%.2f Mpps", r/1_000_000)
	}
	if r >= 1000 {
		return fmt.Sprintf("%.1f kpps", r/1000)
	}
	return fmt.Sprintf("%.0f pps", r)
}
