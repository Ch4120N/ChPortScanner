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
