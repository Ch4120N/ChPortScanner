package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Ch4120N/ChPortScanner/internal/config"
	"github.com/Ch4120N/ChPortScanner/internal/output"
	"github.com/Ch4120N/ChPortScanner/internal/port"
	"github.com/Ch4120N/ChPortScanner/internal/target"
	"github.com/Ch4120N/ChPortScanner/internal/ui"
)

type Job struct {
	IP   string // pre-formatted string (avoid allocation in worker)
	Port uint16
}

type Scanner struct {
	cfg   *config.Config
	ports []uint16
	excl  []target.Spec
	state State

	total int64
	done  atomic.Uint64
	open  atomic.Uint64
	start time.Time

	jobCh    chan Job
