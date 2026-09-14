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
