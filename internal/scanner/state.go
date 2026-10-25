package scanner

import (
	"encoding/json"
	"os"
	"sync"

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

var stateWriteMu sync.Mutex

func (s *Scanner) loadState(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.state)
}

func (s *Scanner) saveState() {
	stateWriteMu.Lock()
	defer stateWriteMu.Unlock()

	s.state.OpenCount = s.open.Load()
	data, _ := json.MarshalIndent(s.state, "", "  ")
	_ = os.WriteFile(s.cfg.StateFile, data, 0644)
}
