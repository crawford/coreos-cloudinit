package config

import (
	"reflect"
	"testing"

	"github.com/coreos/coreos-cloudinit/system"
)

func TestFleetUnit(t *testing.T) {
	for _, tt := range []struct {
		cfg  FleetEnvironment
		root string
		uu   []system.Unit
		err  error
	}{
		{
			cfg:  FleetEnvironment{},
			root: "/",
		},
		{
			cfg: FleetEnvironment{
				PublicIP: "12.34.56.78",
			},
			uu: []system.Unit{
				{
					Name: "fleet.service",
					Content: `[Service]
Environment="FLEET_PUBLIC_IP=12.34.56.78"
`,
					Runtime: true,
					DropIn:  true,
				},
			},
			root: "/",
		},
	} {
		uu, err := tt.cfg.Units(tt.root)
		if tt.err != err {
			t.Errorf("bad error (%q): want %q, got %q", tt.cfg, tt.err, err)
		}
		if !reflect.DeepEqual(uu, tt.uu) {
			t.Errorf("bad units (%q): want %q, got %q", tt.cfg, tt.uu, uu)
		}
	}
}
