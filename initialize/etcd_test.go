package initialize

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/coreos/coreos-cloudinit/system"
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

func TestEtcdEnvironmentWrittenToDisk(t *testing.T) {
	ee := EtcdEnvironment{
		Name:         "node001",
		Discovery:    "http://disco.example.com/foobar",
		PeerBindAddr: "127.0.0.1:7002",
	}
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	sd := system.NewUnitManager(dir)

	uu, err := ee.Units(dir)
	if err != nil {
		t.Fatalf("Generating etcd unit failed: %v", err)
	}
	if len(uu) != 1 {
		t.Fatalf("Expected 1 unit to be returned, got %d", len(uu))
	}
	u := uu[0]

	dst := u.Destination(dir)
	os.Stderr.WriteString("writing to " + dir + "\n")
	if err := sd.PlaceUnit(&u, dst); err != nil {
		t.Fatalf("Writing of EtcdEnvironment failed: %v", err)
	}

	fullPath := path.Join(dir, "run", "systemd", "system", "etcd.service.d", "20-cloudinit.conf")

	fi, err := os.Stat(fullPath)
	if err != nil {
		t.Fatalf("Unable to stat file: %v", err)
	}

	if fi.Mode() != os.FileMode(0644) {
		t.Errorf("File has incorrect mode: %v", fi.Mode())
	}

	contents, err := ioutil.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Unable to read expected file: %v", err)
	}

	expect := `[Service]
Environment="ETCD_DISCOVERY=http://disco.example.com/foobar"
Environment="ETCD_NAME=node001"
Environment="ETCD_PEER_BIND_ADDR=127.0.0.1:7002"
`
	if c := string(contents); c != expect {
		t.Fatalf("bad contents: want %q, got %q", expect, c)
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

func TestEtcdEnvironmentWrittenToDiskDefaultToMachineID(t *testing.T) {
	ee := EtcdEnvironment{
		Discovery: "http://disco.example.com/foobar",
	}
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	sd := system.NewUnitManager(dir)

	os.Mkdir(path.Join(dir, "etc"), os.FileMode(0755))
	err = ioutil.WriteFile(path.Join(dir, "etc", "machine-id"), []byte("node007"), os.FileMode(0444))
	if err != nil {
		t.Fatalf("Failed writing out /etc/machine-id: %v", err)
	}

	uu, err := ee.Units(dir)
	if err != nil {
		t.Fatalf("Generating etcd unit failed: %v", err)
	}
	if len(uu) == 0 {
		t.Fatalf("Returned empty etcd units unexpectedly")
	}
	u := uu[0]

	dst := u.Destination(dir)
	os.Stderr.WriteString("writing to " + dir + "\n")
	if err := sd.PlaceUnit(&u, dst); err != nil {
		t.Fatalf("Writing of EtcdEnvironment failed: %v", err)
	}

	fullPath := path.Join(dir, "run", "systemd", "system", "etcd.service.d", "20-cloudinit.conf")

	contents, err := ioutil.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Unable to read expected file: %v", err)
	}

	expect := `[Service]
Environment="ETCD_DISCOVERY=http://disco.example.com/foobar"
Environment="ETCD_NAME=node007"
`
	if c := string(contents); c != expect {
		t.Fatalf("bad contents: want %q, got %q", expect, c)
	}
}
