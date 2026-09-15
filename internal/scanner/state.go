package scanner

import (
	"encoding/json"
	"os"

	"github.com/Ch4120N/ChPortScanner/internal/target"
)

type State struct {
	Targets   []target.Spec `json:"targets"`
	PortSpec  string        `json:"port_spec"`
