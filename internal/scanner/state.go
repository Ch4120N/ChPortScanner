package scanner

import (
	"encoding/json"
	"os"

	"github.com/Ch4120N/ChPortScanner/internal/target"
)

type State struct {
	Targets   []target.Spec `json:"targets"`
	PortSpec  string        `json:"port_spec"`
	CurSpec   int           `json:"cur_spec"`
	CurIP     string        `json:"cur_ip"`
	CurPort   int           `json:"cur_port"`
	Finished  bool          `json:"finished"`
	OpenCount uint64        `json:"open_count"`
	Format    string        `json:"format"`
	Output    string        `json:"output"`
	ShardIdx  int64         `json:"shard_start_idx"`
	ShardEnd  int64         `json:"shard_end_idx"`
}

func (s *Scanner) loadState(path string) error {
