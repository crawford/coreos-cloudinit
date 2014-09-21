package system

import (
	"github.com/coreos/coreos-cloudinit/config"
)

type UnitManager interface {
	PlaceUnit(unit *config.Unit, dst string) error
	EnableUnitFile(unit string, runtime bool) error
	RunUnitCommand(command, unit string) (string, error)
	DaemonReload() error
	MaskUnit(unit *config.Unit) error
	UnmaskUnit(unit *config.Unit) error
}
