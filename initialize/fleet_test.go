package initialize

import (
	"reflect"
	"testing"

	"github.com/coreos/coreos-cloudinit/config"
)

func TestFleetUnit(t *testing.T) {
	for _, tt := range []struct {
		cfg  config.FleetEnvironment
		root string
		uu   []config.Unit
		err  error
	}{
		{
			cfg:  config.FleetEnvironment{},
			root: "/",
		},
		{
			cfg: config.FleetEnvironment{
				PublicIP: "12.34.56.78",
			},
			uu: []config.Unit{
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
