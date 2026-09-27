package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
)

type Config struct {
	// Targeting
	Target      string
	InputFile   string
	ExcludeFile string
	PingSweep   bool

	// Ports
	Ports        string
	TopPorts     int
	ExcludePorts string
	Profile      string

	// Scanning
	Rate        int
	Workers     int
	Timeout     int
	Retries     int
	RetryDelay  int
	UDP         bool
	Random      bool
	Shard       string
	SourceIP    string

	// Probing
	Banner        bool
	BannerTimeout int
	HTTPProbe     bool
	TLSCert       bool
	RDNS          bool

	// Output
	Output    string
	JSONOut   bool
	Format    string
	NoColor   bool
	Quiet     bool
	Verbose   bool
	ShowStats bool

	// Utilities
	Extract      string
	ExtractPorts bool
	Resume       string
	StateFile    string
	ConfFile     string

	// Meta
	ShowHelp    bool
	ShowVersion bool

	// Internal
	Lite    bool
	Turbo   bool
	IsUnix  bool
	IsWin   bool
}

func New() *Config {
	return &Config{
		Ports:         "80,443",
		Rate:          0, // unlimited by default = fast
		Workers:       0, // auto-size
		Timeout:       1,
		Retries:       0,
		RetryDelay:    500,
		BannerTimeout: 2,
		Format:        "text",
		StateFile:     "scan.state",
		IsUnix:        runtime.GOOS != "windows",
		IsWin:         runtime.GOOS == "windows",
	}
}

func Load(args []string) (*Config, error) {
	c := New()
	fs := flag.NewFlagSet("chportscanner", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	// Meta
	fs.BoolVar(&c.ShowHelp, "h", false, "Show help")
	fs.BoolVar(&c.ShowHelp, "help", false, "Show help")
	fs.BoolVar(&c.ShowVersion, "v", false, "Show version")
	fs.BoolVar(&c.ShowVersion, "version", false, "Show version")

	// Targeting
	fs.StringVar(&c.Target, "t", "", "Targets: IPs/CIDRs/ranges (comma-separated)")
	fs.StringVar(&c.Target, "target", "", "Targets (long)")
	fs.StringVar(&c.InputFile, "iL", "", "Input file with targets")
	fs.StringVar(&c.InputFile, "input", "", "Input file (long)")
	fs.StringVar(&c.ExcludeFile, "exclude", "", "File with IPs/CIDRs to exclude")
	fs.BoolVar(&c.PingSweep, "sn", false, "Ping sweep only")

	// Ports
	fs.StringVar(&c.Ports, "p", "80,443", "Ports to scan")
	fs.StringVar(&c.Ports, "ports", "80,443", "Ports (long)")
	fs.IntVar(&c.TopPorts, "top-ports", 0, "Scan top N common ports")
	fs.StringVar(&c.ExcludePorts, "exclude-ports", "", "Ports to exclude")
	fs.StringVar(&c.Profile, "profile", "", "quick|default|full|stealth")

	// Scanning
	fs.IntVar(&c.Rate, "rate", 0, "Packets/sec (0 = unlimited)")
	fs.IntVar(&c.Workers, "workers", 0, "Concurrent workers (0 = auto)")
	fs.IntVar(&c.Timeout, "timeout", 1, "Connection timeout (s)")
	fs.IntVar(&c.Retries, "retries", 0, "Retries")
	fs.IntVar(&c.RetryDelay, "retry-delay", 500, "Retry delay (ms)")
	fs.BoolVar(&c.UDP, "udp", false, "UDP scan mode")
	fs.BoolVar(&c.Random, "random", false, "Randomize scan order")
	fs.StringVar(&c.Shard, "shard", "", "Shard X/Y")
	fs.StringVar(&c.SourceIP, "S", "", "Source IP to bind")
	fs.BoolVar(&c.Lite, "lite", false, "Low-resource mode (for low-end PCs)")
	fs.BoolVar(&c.Turbo, "turbo", false, "Maximum speed mode")

	// Probing
	fs.BoolVar(&c.Banner, "banner", false, "Enable banner grabbing")
	fs.IntVar(&c.BannerTimeout, "banner-timeout", 2, "Banner timeout (s)")
	fs.BoolVar(&c.HTTPProbe, "http-probe", false, "Fetch HTTP page title")
	fs.BoolVar(&c.TLSCert, "tls-cert", false, "Extract TLS cert info")
	fs.BoolVar(&c.RDNS, "rdns", false, "Reverse DNS lookup")

	// Output
	fs.StringVar(&c.Output, "o", "", "Output file")
	fs.StringVar(&c.Output, "output", "", "Output file (long)")
	fs.BoolVar(&c.JSONOut, "oJ", false, "JSON output")
	fs.StringVar(&c.Format, "format", "text", "text|json|jsonl|csv|grep")
	fs.BoolVar(&c.NoColor, "no-color", false, "Disable colors")
