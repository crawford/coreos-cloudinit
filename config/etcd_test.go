package config

import (
	"testing"
)

func TestEtcdEnvironment(t *testing.T) {
	for _, tt := range []struct {
		cfg EtcdEnvironment
		env string
	}{
		{
			EtcdEnvironment{
				Discovery:    "http://disco.example.com/foobar",
				PeerBindAddr: "127.0.0.1:7002",
			},
			`[Service]
Environment="ETCD_DISCOVERY=http://disco.example.com/foobar"
Environment="ETCD_PEER_BIND_ADDR=127.0.0.1:7002"
`,
		},
	} {
		if env := tt.cfg.String(); env != tt.env {
			t.Errorf("bad environment (%q): want %q, got %q", tt.cfg, tt.env, env)
		}
	}
}

func TestEtcdEnvironmentEmptyNoOp(t *testing.T) {
	ee := EtcdEnvironment{}
	uu, err := ee.Units("")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(uu) > 0 {
		t.Fatalf("Generated etcd units unexpectedly: %s", ee)
	}
}
