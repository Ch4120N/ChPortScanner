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
