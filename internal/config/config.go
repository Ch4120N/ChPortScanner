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
