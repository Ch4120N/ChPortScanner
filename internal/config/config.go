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
